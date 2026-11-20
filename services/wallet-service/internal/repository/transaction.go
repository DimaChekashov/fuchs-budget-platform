package repository

import (
	"context"
	"encoding/json"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Create(ctx context.Context, tx pgx.Tx, t *domain.Transaction, key *string) error {
	query := `
  INSERT INTO transactions (wallet_id, type, amount, description, idempotency_key)
  VALUES ($1, $2, $3::text::numeric, $4, $5)
  RETURNING id, created_at
 `
	return tx.QueryRow(ctx, query, t.WalletID, t.Type, string(t.Amount), t.Description, key).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *TransactionRepository) GetByKey(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, key string) (*domain.Transaction, error) {
	query := `
  SELECT id, wallet_id, type, amount::text, COALESCE(description, ''), created_at
  FROM transactions WHERE wallet_id = $1 AND idempotency_key = $2
 `
	return scanTransaction(tx.QueryRow(ctx, query, walletID, key))
}

func (r *TransactionRepository) GetByWalletID(ctx context.Context, walletID uuid.UUID) ([]*domain.Transaction, error) {
	query := `
  SELECT id, wallet_id, type, amount::text, COALESCE(description, ''), created_at
  FROM transactions WHERE wallet_id = $1
  ORDER BY created_at DESC
 `
	rows, err := r.pool.Query(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*domain.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func scanTransaction(row pgx.Row) (*domain.Transaction, error) {
	t := &domain.Transaction{}
	var amount string
	if err := row.Scan(&t.ID, &t.WalletID, &t.Type, &amount, &t.Description, &t.CreatedAt); err != nil {
		return nil, err
	}
	t.Amount = json.Number(amount)
	return t, nil
}
