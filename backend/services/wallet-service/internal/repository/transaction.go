package repository

import (
	"context"
	"encoding/json"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/domain"
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

func (r *TransactionRepository) Create(ctx context.Context, tx pgx.Tx, transaction *domain.Transaction, idempotencyKey *string) error {
	query := `
  INSERT INTO transactions (wallet_id, type, amount, description, idempotency_key)
  VALUES ($1, $2, $3::text::numeric, $4, $5)
  RETURNING id, created_at
 `
	return tx.QueryRow(ctx, query,
		transaction.WalletID, transaction.Type, string(transaction.Amount),
		transaction.Description, idempotencyKey,
	).Scan(&transaction.ID, &transaction.CreatedAt)
}

func (r *TransactionRepository) GetByKey(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, idempotencyKey string) (*domain.Transaction, error) {
	query := `
  SELECT id, wallet_id, type, amount::text, COALESCE(description, ''), created_at
  FROM transactions WHERE wallet_id = $1 AND idempotency_key = $2
 `
	return scanTransaction(tx.QueryRow(ctx, query, walletID, idempotencyKey))
}

func (r *TransactionRepository) GetHistory(ctx context.Context, walletID, userID uuid.UUID, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT t.id, t.wallet_id, t.type, t.amount::text, COALESCE(t.description, ''), t.created_at
		FROM transactions t JOIN wallets w ON w.id = t.wallet_id
		WHERE t.wallet_id = $1 AND w.user_id = $2
		ORDER BY t.created_at DESC, t.id DESC LIMIT 50 OFFSET $3
 `
	rows, err := r.pool.Query(ctx, query, walletID, userID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, rows.Err()
}

func scanTransaction(row pgx.Row) (*domain.Transaction, error) {
	transaction := &domain.Transaction{}
	var amount string
	if err := row.Scan(
		&transaction.ID, &transaction.WalletID, &transaction.Type,
		&amount, &transaction.Description, &transaction.CreatedAt,
	); err != nil {
		return nil, err
	}
	transaction.Amount = json.Number(amount)
	return transaction, nil
}
