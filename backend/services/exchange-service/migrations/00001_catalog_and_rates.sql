-- +goose Up
CREATE TABLE currencies (
    code VARCHAR(10) PRIMARY KEY,
    precision SMALLINT NOT NULL CHECK (precision BETWEEN 0 AND 2)
);

INSERT INTO currencies (code, precision)
VALUES ('RUB', 2), ('USD', 2), ('EUR', 2), ('CNY', 2), ('GBP', 2), ('JPY', 0);

CREATE TABLE user_currencies (
    user_id UUID NOT NULL,
    code VARCHAR(10) NOT NULL,
    precision SMALLINT NOT NULL CHECK (precision BETWEEN 0 AND 2),
    PRIMARY KEY (user_id, code)
);

CREATE TABLE exchange_rates (
    base VARCHAR(10) NOT NULL REFERENCES currencies(code),
    quote VARCHAR(10) NOT NULL REFERENCES currencies(code),
    observed_on DATE NOT NULL,
    rate NUMERIC NOT NULL CHECK (rate > 0),
    source TEXT NOT NULL,
    PRIMARY KEY (base, quote, observed_on, source),
    CHECK (base <> quote)
);

-- +goose Down
DROP TABLE exchange_rates;
DROP TABLE user_currencies;
DROP TABLE currencies;
