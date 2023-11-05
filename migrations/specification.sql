CREATE SCHEMA IF NOT EXISTS crypto_analyst;

CREATE TABLE IF NOT EXISTS crypto_analyst.prices
(
    price      double precision NOT NULL,
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   TIMESTAMP        NOT NULL,
    updated_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON crypto_analyst.prices (price, symbol, exchange, datetime);

CREATE INDEX exchange_idx ON crypto_analyst.prices (exchange);
CREATE INDEX symbol_idx ON crypto_analyst.prices (symbol);

alter table crypto_analyst.prices
    owner to crypto_app;

CREATE TABLE IF NOT EXISTS crypto_analyst.price_changes
(
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   VARCHAR(50)      NOT NULL,
    afg_value  BIGINT          NOT NULL DEFAULT 0,
    price      double precision NOT NULL DEFAULT 0,
    prev_price double precision NOT NULL DEFAULT 0,
    created_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON crypto_analyst.price_changes (symbol, exchange, datetime);

CREATE INDEX price_changes_exchange_idx ON crypto_analyst.price_changes (exchange);
CREATE INDEX price_changes_symbol_idx ON crypto_analyst.price_changes (symbol);
CREATE INDEX price_changes_datetime_idx ON crypto_analyst.price_changes (datetime);

alter table crypto_analyst.price_changes
    owner to crypto_app;