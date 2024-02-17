package db

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type PriceRepository struct {
	db *sqlx.DB
}

func NewPriceRepository(db *sqlx.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

func (repo *PriceRepository) FirstDatetime(ctx context.Context, symbol string) (time.Time, error) {
	var (
		query     = `SELECT min(datetime) FROM crypto_analyst.prices WHERE symbol = $1 `
		firstDate time.Time
	)
	if err := repo.db.GetContext(ctx, &firstDate, query, symbol); err != nil {
		return time.Time{}, err
	}
	return firstDate, nil
}

func (repo *PriceRepository) SymbolPrices(ctx context.Context, symbol string, from, to time.Time) ([]domain.SymbolPrice, error) {
	var (
		query = `
SELECT 
    price, symbol, exchange, datetime 
FROM crypto_analyst.prices 
WHERE symbol = $1 
  AND (datetime BETWEEN $2 AND $3)  
ORDER BY  datetime ASC
`
		result []domain.SymbolPrice
	)
	if err := repo.db.SelectContext(ctx, &result, query, symbol, from, to); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *PriceRepository) DeleteOldPrices(ctx context.Context, symbol string, to time.Time) error {
	var query = `
DELETE FROM crypto_analyst.prices WHERE symbol = $1 
  AND datetime < $2
`
	_, err := repo.db.ExecContext(ctx, query, symbol, to)

	return err
}

func (repo *PriceRepository) ClearOldPrices(ctx context.Context, to time.Time) error {
	var query = `
DELETE FROM crypto_analyst.prices WHERE  datetime < $1
`
	_, err := repo.db.ExecContext(ctx, query, to)

	return err
}

func (repo *PriceRepository) DeletePrices(ctx context.Context, symbol string, from, to time.Time) error {
	var query = `
DELETE FROM crypto_analyst.prices WHERE symbol = $1 
  AND (datetime BETWEEN $2 AND $3)  
`
	_, err := repo.db.ExecContext(ctx, query, symbol, from, to)

	return err
}

func (repo *PriceRepository) SavePrices(ctx context.Context, prices []*domain.SymbolPrice) error {
	var (
		values []string
	)

	if len(prices) == 0 {
		return nil
	}
	for _, price := range prices {
		values = append(
			values,
			fmt.Sprintf(
				"(%f,'%s', '%s','%s')",
				price.Price, price.Symbol, price.Exchange, price.Date.Format(DatetimeFormat)),
		)
	}
	query := fmt.Sprintf(
		"INSERT INTO crypto_analyst.prices(price, symbol,exchange,datetime) VALUES %s ON CONFLICT (price, symbol, exchange, datetime) DO NOTHING",
		strings.Join(values, ", "),
	)
	_, err := repo.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PriceRepository) Prices(ctx context.Context, symbol string) ([]domain.SymbolPrice, error) {
	var (
		query = `
SELECT price, symbol, exchange, datetime 
FROM crypto_analyst.prices
WHERE updated_at = (SELECT updated_at
                    FROM crypto_analyst.prices
                    ORDER BY updated_at DESC
                    LIMIT 1)
AND symbol = $1 
ORDER BY exchange ASC
`
		result []domain.SymbolPrice
	)
	if err := repo.db.SelectContext(ctx, &result, query, symbol); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *PriceRepository) AddNewSymbol(ctx context.Context, prices []domain.SymbolPrice) error {
	var (
		values []string
	)

	if len(prices) == 0 {
		return nil
	}
	for _, price := range prices {
		values = append(
			values,
			fmt.Sprintf(
				"(%f,'%s', '%s','%s')",
				price.Price, price.Symbol, price.Exchange, price.Date.Format(DatetimeFormat)),
		)
	}
	query := fmt.Sprintf(
		"INSERT INTO crypto_analyst.new_symbols(price, symbol,exchange,datetime) VALUES %s ON CONFLICT (symbol, exchange) DO NOTHING",
		strings.Join(values, ", "),
	)
	_, err := repo.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PriceRepository) NewSymbols(ctx context.Context, from time.Time) ([]domain.SymbolPrice, error) {
	var (
		query = `
SELECT price, symbol, exchange, datetime
FROM crypto_analyst.new_symbols
WHERE updated_at >= $1
`
		symbols []domain.SymbolPrice
	)
	if err := repo.db.SelectContext(ctx, &symbols, query, from); err != nil {
		return nil, err
	}
	return symbols, nil
}
