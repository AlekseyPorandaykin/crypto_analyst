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

CREATE INDEX prices_exchange_idx ON crypto_analyst.prices (exchange);
CREATE INDEX prices_symbol_idx ON crypto_analyst.prices (symbol);

alter table crypto_analyst.prices
    owner to crypto_app;

CREATE TABLE IF NOT EXISTS crypto_analyst.price_changes
(
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   VARCHAR(50)      NOT NULL,
    afg_value  BIGINT           NOT NULL DEFAULT 0,
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


ALTER TABLE crypto_analyst.price_changes
    RENAME COLUMN afg_value TO coefficient_change;


CREATE TABLE IF NOT EXISTS crypto_analyst.price_aggregation
(
    symbol     VARCHAR(50)  NOT NULL,
    exchange   VARCHAR(50)  NOT NULL DEFAULT '',
    metric     VARCHAR(50)  NOT NULL,
    key        VARCHAR(50)  NOT NULL,
    value      VARCHAR(250) NOT NULL,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX price_aggregation_exchange_idx ON crypto_analyst.price_aggregation (exchange);
CREATE INDEX price_aggregation_symbol_idx ON crypto_analyst.price_aggregation (symbol);
CREATE INDEX price_aggregation_name_idx ON crypto_analyst.price_aggregation (metric);

CREATE UNIQUE INDEX ON crypto_analyst.price_aggregation (symbol, exchange, metric, key);

alter table crypto_analyst.price_aggregation
    owner to crypto_app;

CREATE TABLE IF NOT EXISTS crypto_analyst.candlesticks
(
    symbol          VARCHAR(50)      NOT NULL,
    exchange        VARCHAR(50)      NOT NULL DEFAULT '',
    open_time       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    close_time      TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    open_price      double precision NOT NULL DEFAULT 0,
    high_price      double precision NOT NULL DEFAULT 0,
    low_price       double precision NOT NULL DEFAULT 0,
    close_price     double precision NOT NULL DEFAULT 0,
    volume          double precision NOT NULL DEFAULT 0,
    number_trades   INT              NOT NULL DEFAULT 0,
    candle_interval VARCHAR(10)      NOT NULL,
    created_at      TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON crypto_analyst.candlesticks (symbol, exchange, open_time, close_time, candle_interval);
CREATE INDEX candlesticks_exchange_idx ON crypto_analyst.candlesticks (exchange);
CREATE INDEX candlesticks_symbol_idx ON crypto_analyst.candlesticks (symbol);
CREATE INDEX candlesticks_interval_idx ON crypto_analyst.candlesticks (candle_interval);


CREATE TABLE IF NOT EXISTS crypto_analyst.new_symbols
(
    price      double precision NOT NULL,
    symbol     VARCHAR(50)      NOT NULL,
    exchange   VARCHAR(50)      NOT NULL,
    datetime   TIMESTAMP        NOT NULL,
    updated_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON crypto_analyst.new_symbols (price, symbol, exchange, datetime);

CREATE INDEX new_symbols_exchange_idx ON crypto_analyst.new_symbols (exchange);
CREATE INDEX new_symbols_symbol_idx ON crypto_analyst.new_symbols (symbol);
CREATE UNIQUE INDEX ON crypto_analyst.new_symbols (symbol, exchange);

alter table crypto_analyst.new_symbols
    owner to crypto_app;
