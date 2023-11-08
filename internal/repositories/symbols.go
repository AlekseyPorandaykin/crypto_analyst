package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Symbols struct {
	db *sqlx.DB
}

func NewSymbols(db *sqlx.DB) *Symbols {
	return &Symbols{db: db}
}

func (repo *Symbols) List(ctx context.Context) ([]string, error) {
	var (
		query   = `SELECT DISTINCT symbol FROM crypto_analyst.prices`
		symbols []string
	)
	if err := repo.db.SelectContext(ctx, &symbols, query); err != nil {
		return nil, err
	}
	return symbols, nil
}

func (repo *Symbols) PopularSymbols(ctx context.Context, limit int) ([]string, error) {
	var (
		query = `
SELECT 
    symbol, count(exchange) 
FROM (
	SELECT DISTINCT symbol, exchange FROM crypto_analyst.prices
	 ) as temp_table 
GROUP BY symbol
`
		symbols []string
	)
	rows, err := repo.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			symbol string
			count  int
		)
		if err := rows.Scan(&symbol, &count); err != nil {
			return nil, err
		}
		if count >= limit {
			symbols = append(symbols, symbol)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return symbols, nil
}
