package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWrongOwner        = errors.New("wallet does not belong to user")
)

type WalletService struct {
	walletRepo *repository.WalletRepository
	txnRepo    *repository.TransactionRepository
}

func NewWalletService(walletRepo *repository.WalletRepository, txnRepo *repository.TransactionRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo, txnRepo: txnRepo}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID, name, currency string) (*domain.Wallet, error) {
	w := &domain.Wallet{
		UserID:   userID,
		Name:     name,
		Currency: currency,
	}
	if err := s.walletRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (s *WalletService) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

type DepositInput struct {
	WalletID    uuid.UUID
	UserID      uuid.UUID
	Amount      float64
	Description string
}

func (s *WalletService) Deposit(ctx context.Context, input DepositInput) (*domain.Transaction, error) {
	wallet, err := s.walletRepo.GetByID(ctx, input.WalletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	if wallet.UserID != input.UserID {
		return nil, ErrWrongOwner
	}

	txn := &domain.Transaction{
		WalletID:    input.WalletID,
		Type:        domain.TransactionDeposit,
		Amount:      input.Amount,
		Description: input.Description,
	}
	if err := s.txnRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := s.walletRepo.UpdateBalance(ctx, input.WalletID, input.Amount); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	return txn, nil
}

type WithdrawInput struct {
	WalletID    uuid.UUID
	UserID      uuid.UUID
	Amount      float64
	Description string
}

func (s *WalletService) Withdraw(ctx context.Context, input WithdrawInput) (*domain.Transaction, error) {
	wallet, err := s.walletRepo.GetByID(ctx, input.WalletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	if wallet.UserID != input.UserID {
		return nil, ErrWrongOwner
	}
	if wallet.Balance < input.Amount {
		return nil, ErrInsufficientFunds
	}

	txn := &domain.Transaction{
		WalletID:    input.WalletID,
		Type:        domain.TransactionWithdrawal,
		Amount:      input.Amount,
		Description: input.Description,
	}
	if err := s.txnRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := s.walletRepo.UpdateBalance(ctx, input.WalletID, -input.Amount); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	return txn, nil
}
