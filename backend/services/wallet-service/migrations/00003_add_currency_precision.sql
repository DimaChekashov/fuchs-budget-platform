-- +goose Up
ALTER TABLE wallets ADD COLUMN currency_precision SMALLINT NOT NULL DEFAULT 2
    CHECK (currency_precision BETWEEN 0 AND 2);

-- +goose Down
ALTER TABLE wallets DROP COLUMN currency_precision;
