package repository

import (
	"context"
	"fmt"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	query := `
		INSERT INTO transactions (wallet_id, type, amount, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, t.WalletID, t.Type, t.Amount, t.Description).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *TransactionRepository) GetByWalletID(ctx context.Context, walletID uuid.UUID) ([]*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, description, created_at
		FROM transactions WHERE wallet_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, walletID)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	defer rows.Close()

	var txns []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}
