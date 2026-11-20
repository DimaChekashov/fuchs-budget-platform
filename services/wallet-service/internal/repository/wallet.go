package repository

import (
	"context"
	"encoding/json"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		RETURNING id, balance::text, created_at, updated_at
	`
	var balance string
	err := r.pool.QueryRow(ctx, query, w.UserID, w.Name, w.Currency).
		Scan(&w.ID, &balance, &w.CreatedAt, &w.UpdatedAt)
	w.Balance = json.Number(balance)
	return err
}

func (r *WalletRepository) LockOwned(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID) error {
	var lockedID uuid.UUID
	return tx.QueryRow(ctx, `SELECT id FROM wallets WHERE id = $1 AND user_id = $2 FOR UPDATE`, id, userID).
		Scan(&lockedID)
}

func (r *WalletRepository) Exists(ctx context.Context, tx pgx.Tx, id uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM wallets WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	query := `
		SELECT id, user_id, name, currency, balance::text, created_at, updated_at
		FROM wallets WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*domain.Wallet
	for rows.Next() {
		w := &domain.Wallet{}
		var balance string
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Currency, &balance, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		w.Balance = json.Number(balance)
		wallets = append(wallets, w)
	}
	return wallets, rows.Err()
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta string) (bool, error) {
	query := `
		UPDATE wallets
		SET balance = balance + $1::text::numeric, updated_at = NOW()
		WHERE id = $2 AND ($1::text::numeric > 0 OR balance + $1::text::numeric >= 0)
	`
	tag, err := tx.Exec(ctx, query, delta, id)
	return tag.RowsAffected() == 1, err
}
