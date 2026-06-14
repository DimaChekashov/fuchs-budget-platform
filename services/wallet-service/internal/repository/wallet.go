package repository

import (
	"context"
	"fmt"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) Create(ctx context.Context, w *domain.Wallet) error {
	query := `
		INSERT INTO wallets (user_id, name, currency)
		VALUES ($1, $2, $3)
		RETURNING id, balance, created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query, w.UserID, w.Name, w.Currency).
		Scan(&w.ID, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
}

func (r *WalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, name, currency, balance, created_at, updated_at
		FROM wallets WHERE id = $1
	`
	w := &domain.Wallet{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&w.ID, &w.UserID, &w.Name, &w.Currency, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	return w, nil
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	query := `
		SELECT id, user_id, name, currency, balance, created_at, updated_at
		FROM wallets WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*domain.Wallet
	for rows.Next() {
		w := &domain.Wallet{}
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Currency, &w.Balance, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	return wallets, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	query := `
		UPDATE wallets
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, amount, id)
	return err
}
