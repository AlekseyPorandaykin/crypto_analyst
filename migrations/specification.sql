CREATE TABLE IF NOT EXISTS prices
(
    price      double precision NOT NULL,
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   TIMESTAMP        NOT NULL,
    updated_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON prices (price, symbol, exchange, datetime);

CREATE INDEX exchange_idx ON prices (exchange);
CREATE INDEX symbol_idx ON prices (symbol);

alter table prices
    owner to crypto_analyst;

CREATE TABLE IF NOT EXISTS price_changes
(
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   VARCHAR(50)      NOT NULL,
    afg_value  INTEGER          NOT NULL DEFAULT 0,
    price      double precision NOT NULL DEFAULT 0,
    prev_price double precision NOT NULL DEFAULT 0,
    created_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON price_changes (symbol, exchange, datetime);

CREATE INDEX price_changes_exchange_idx ON price_changes (exchange);
CREATE INDEX price_changes_symbol_idx ON price_changes (symbol);
CREATE INDEX price_changes_datetime_idx ON price_changes (datetime);

alter table price_changes
    owner to crypto_analyst;