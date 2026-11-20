-- +goose Up
CREATE TABLE wallet_operation_outbox (
    operation_id UUID PRIMARY KEY REFERENCES transactions(id),
    wallet_id UUID NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);
CREATE INDEX wallet_operation_outbox_pending_idx
    ON wallet_operation_outbox (created_at, operation_id)
    WHERE published_at IS NULL;

-- +goose StatementBegin
CREATE FUNCTION wallet_operation_payload(t transactions, owner UUID, currency_code VARCHAR)
RETURNS JSONB LANGUAGE SQL STABLE AS $$
    SELECT jsonb_build_object(
        'version', 1, 'operation_id', t.id,
        'wallet_id', t.wallet_id, 'user_id', owner, 'currency', currency_code,
        'type', t.type, 'amount', t.amount::text, 'recorded_at', t.created_at
    );
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION enqueue_wallet_operation()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO wallet_operation_outbox (operation_id, wallet_id, payload)
    SELECT NEW.id, NEW.wallet_id, wallet_operation_payload(NEW, w.user_id, w.currency)
    FROM wallets w WHERE w.id = NEW.wallet_id;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER enqueue_wallet_operation
AFTER INSERT ON transactions
FOR EACH ROW WHEN (NEW.type IN ('deposit', 'withdrawal'))
EXECUTE FUNCTION enqueue_wallet_operation();

INSERT INTO wallet_operation_outbox (operation_id, wallet_id, payload, created_at)
SELECT t.id, t.wallet_id, wallet_operation_payload(t, w.user_id, w.currency), t.created_at
FROM transactions t JOIN wallets w ON w.id = t.wallet_id
WHERE t.type IN ('deposit', 'withdrawal')
ON CONFLICT (operation_id) DO NOTHING;

-- +goose Down
DROP TRIGGER enqueue_wallet_operation ON transactions;
DROP FUNCTION enqueue_wallet_operation();
DROP FUNCTION wallet_operation_payload(transactions, UUID, VARCHAR);
DROP TABLE wallet_operation_outbox;
