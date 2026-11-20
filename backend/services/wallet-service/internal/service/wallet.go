package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrWrongOwner          = errors.New("wallet does not belong to user")
	ErrInvalidAmount       = errors.New("amount must be a positive number with at most two decimal places and fit NUMERIC(18,2)")
	ErrInvalidKey          = errors.New("invalid idempotency key")
	ErrIdempotencyConflict = errors.New("idempotency key already used for a different operation")
	ErrBalanceLimit        = errors.New("balance exceeds the supported limit")
	ErrInvalidName         = errors.New("wallet name must contain 1 to 100 characters")
	ErrCurrencyPrecision   = errors.New("amount exceeds the wallet currency precision")
)

type WalletService struct {
	pool            *pgxpool.Pool
	walletRepo      *repository.WalletRepository
	transactionRepo *repository.TransactionRepository
	checkCurrency   CurrencyCheck
}

func NewWalletService(
	pool *pgxpool.Pool,
	walletRepo *repository.WalletRepository,
	transactionRepo *repository.TransactionRepository,
	checkCurrency CurrencyCheck,
) *WalletService {
	return &WalletService{
		pool:            pool,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		checkCurrency:   checkCurrency,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID, name, currency, authorization string) (*domain.Wallet, error) {
	precision := 2
	if s.checkCurrency != nil {
		currency = strings.ToUpper(strings.TrimSpace(currency))
		value, err := s.checkCurrency(ctx, currency, authorization)
		if err != nil {
			return nil, err
		}
		precision = value
	}
	wallet := &domain.Wallet{UserID: userID, Name: name, Currency: currency, CurrencyPrecision: precision}
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return wallet, nil
}

func (s *WalletService) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

func (s *WalletService) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetOwned(ctx, walletID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	return wallet, err
}

func (s *WalletService) RenameWallet(ctx context.Context, walletID, userID uuid.UUID, name string) (*domain.Wallet, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsRune(name, 0) || utf8.RuneCountInString(name) > 100 {
		return nil, ErrInvalidName
	}
	wallet, err := s.walletRepo.Rename(ctx, walletID, userID, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	return wallet, err
}

func (s *WalletService) GetHistory(ctx context.Context, walletID, userID uuid.UUID, offset int) ([]*domain.Transaction, error) {
	if _, err := s.GetWallet(ctx, walletID, userID); err != nil {
		return nil, err
	}
	return s.transactionRepo.GetHistory(ctx, walletID, userID, offset)
}

type OperationInput struct {
	WalletID       uuid.UUID
	UserID         uuid.UUID
	Amount         json.RawMessage
	Description    string
	IdempotencyKey *string
}

func (s *WalletService) Deposit(ctx context.Context, input OperationInput) (*domain.Transaction, error) {
	return s.record(ctx, input, domain.TransactionDeposit)
}

func (s *WalletService) Withdraw(ctx context.Context, input OperationInput) (*domain.Transaction, error) {
	return s.record(ctx, input, domain.TransactionWithdrawal)
}

func (s *WalletService) record(ctx context.Context, input OperationInput, kind domain.TransactionType) (*domain.Transaction, error) {
	amount, err := parseAmount(input.Amount)
	if err != nil {
		return nil, err
	}
	if input.IdempotencyKey != nil {
		key := *input.IdempotencyKey
		if len(key) == 0 || len(key) > 128 || strings.ContainsFunc(key, func(r rune) bool { return r < 33 || r > 126 || r == ',' }) {
			return nil, ErrInvalidKey
		}
	}

	operation := &domain.Transaction{
		WalletID:    input.WalletID,
		Type:        kind,
		Amount:      amount,
		Description: input.Description,
	}
	err = pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		precision, err := s.lockWallet(ctx, tx, input)
		if err != nil {
			return err
		}
		fraction := string(amount)[len(amount)-2:]
		if len(strings.TrimRight(fraction, "0")) > precision {
			return ErrCurrencyPrecision
		}
		if input.IdempotencyKey != nil {
			previous, err := s.transactionRepo.GetByKey(ctx, tx, input.WalletID, *input.IdempotencyKey)
			switch {
			case err == nil:
				if previous.Type != kind || previous.Amount != amount || previous.Description != input.Description {
					return ErrIdempotencyConflict
				}
				operation = previous
				return nil
			case !errors.Is(err, pgx.ErrNoRows):
				return fmt.Errorf("read idempotency key: %w", err)
			}
		}

		delta := string(amount)
		if kind == domain.TransactionWithdrawal {
			delta = "-" + delta
		}
		updated, err := s.walletRepo.UpdateBalance(ctx, tx, input.WalletID, delta)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "22003" {
				return ErrBalanceLimit
			}
			return fmt.Errorf("update balance: %w", err)
		}
		if !updated {
			return ErrInsufficientFunds
		}
		if err := s.transactionRepo.Create(ctx, tx, operation, input.IdempotencyKey); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return operation, nil
}

func (s *WalletService) lockWallet(ctx context.Context, tx pgx.Tx, input OperationInput) (int, error) {
	precision, err := s.walletRepo.LockOwned(ctx, tx, input.WalletID, input.UserID)
	if !errors.Is(err, pgx.ErrNoRows) {
		return precision, err
	}
	exists, err := s.walletRepo.Exists(ctx, tx, input.WalletID)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrWrongOwner
	}
	return 0, ErrWalletNotFound
}

func parseAmount(raw json.RawMessage) (json.Number, error) {
	value := strings.TrimSpace(string(raw))
	if len(value) == 0 || len(value) > 128 || value[0] < '0' || value[0] > '9' || !json.Valid([]byte(value)) {
		return "", ErrInvalidAmount
	}
	// Bound the exponent before passing it to big.Rat, which can allocate based on it.
	if exponentIndex := strings.IndexAny(value, "eE"); exponentIndex >= 0 {
		exponent, err := strconv.Atoi(value[exponentIndex+1:])
		if err != nil || exponent < -128 || exponent > 128 {
			return "", ErrInvalidAmount
		}
	}
	amount, ok := new(big.Rat).SetString(value)
	if !ok || amount.Sign() <= 0 {
		return "", ErrInvalidAmount
	}
	cents := new(big.Rat).Mul(amount, big.NewRat(100, 1))
	if !cents.IsInt() || cents.Num().Cmp(big.NewInt(999999999999999999)) > 0 {
		return "", ErrInvalidAmount
	}
	return json.Number(amount.FloatString(2)), nil
}
