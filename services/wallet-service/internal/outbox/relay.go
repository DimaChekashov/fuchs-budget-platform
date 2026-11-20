package outbox

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Publish func(context.Context, uuid.UUID, []byte) error

type Relay struct {
	Pool    *pgxpool.Pool
	Publish Publish
}

func (r Relay) Next(ctx context.Context) (bool, error) {
	var operationID, walletID uuid.UUID
	var payload []byte
	err := pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
    SELECT operation_id, wallet_id, payload FROM wallet_operation_outbox
    WHERE published_at IS NULL ORDER BY created_at, operation_id
    LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&operationID, &walletID, &payload)
		if err != nil {
			return err
		}
		if err := r.Publish(ctx, walletID, payload); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE wallet_operation_outbox SET published_at = NOW() WHERE operation_id = $1`, operationID)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r Relay) Run(ctx context.Context) {
	for ctx.Err() == nil {
		operation, cancel := context.WithTimeout(ctx, 10*time.Second)
		processed, err := r.Next(operation)
		cancel()
		if err != nil && ctx.Err() == nil {
			log.Printf("publish wallet event; retrying: %v", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
