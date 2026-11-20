-- +goose Up
ALTER TABLE transactions ADD COLUMN idempotency_key VARCHAR(128);
CREATE UNIQUE INDEX transactions_wallet_idempotency_key_idx
    ON transactions (wallet_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- +goose Down
DROP INDEX transactions_wallet_idempotency_key_idx;
ALTER TABLE transactions DROP COLUMN idempotency_key;
