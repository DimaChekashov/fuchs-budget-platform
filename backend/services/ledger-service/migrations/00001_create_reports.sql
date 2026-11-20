-- +goose Up
CREATE TABLE processed_operations (
    operation_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX processed_operations_user_time_idx ON processed_operations (user_id, processed_at DESC);

CREATE TABLE daily_totals (
    user_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    currency VARCHAR(10) NOT NULL,
    day DATE NOT NULL,
    income NUMERIC NOT NULL,
    expenses NUMERIC NOT NULL,
    PRIMARY KEY (user_id, wallet_id, currency, day)
);
CREATE INDEX daily_totals_user_day_idx ON daily_totals (user_id, day);

CREATE TABLE rejected_events (
    topic TEXT NOT NULL,
    partition INT NOT NULL,
    kafka_offset BIGINT NOT NULL,
    reason TEXT NOT NULL,
    PRIMARY KEY (topic, partition, kafka_offset)
);

-- +goose Down
DROP TABLE rejected_events;
DROP TABLE daily_totals;
DROP TABLE processed_operations;
