package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/repository"
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
)

type WalletService struct {
	pool       *pgxpool.Pool
	walletRepo *repository.WalletRepository
	txnRepo    *repository.TransactionRepository
}

func NewWalletService(pool *pgxpool.Pool, walletRepo *repository.WalletRepository, txnRepo *repository.TransactionRepository) *WalletService {
	return &WalletService{pool: pool, walletRepo: walletRepo, txnRepo: txnRepo}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID, name, currency string) (*domain.Wallet, error) {
	w := &domain.Wallet{UserID: userID, Name: name, Currency: currency}
	if err := s.walletRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (s *WalletService) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
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
		if err := s.lockWallet(ctx, tx, input); err != nil {
			return err
		}
		if input.IdempotencyKey != nil {
			previous, err := s.txnRepo.GetByKey(ctx, tx, input.WalletID, *input.IdempotencyKey)
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
		if err := s.txnRepo.Create(ctx, tx, operation, input.IdempotencyKey); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return operation, nil
}

func (s *WalletService) lockWallet(ctx context.Context, tx pgx.Tx, input OperationInput) error {
	err := s.walletRepo.LockOwned(ctx, tx, input.WalletID, input.UserID)
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	exists, err := s.walletRepo.Exists(ctx, tx, input.WalletID)
	if err != nil {
		return err
	}
	if exists {
		return ErrWrongOwner
	}
	return ErrWalletNotFound
}

func parseAmount(raw json.RawMessage) (json.Number, error) {
	value := strings.TrimSpace(string(raw))
	if len(value) == 0 || len(value) > 128 || value[0] < '0' || value[0] > '9' || !json.Valid([]byte(value)) {
		return "", ErrInvalidAmount
	}
	// Bound the exponent before passing it to big.Rat, which can allocate based on it.
	if i := strings.IndexAny(value, "eE"); i >= 0 {
		exponent, err := strconv.Atoi(value[i+1:])
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
