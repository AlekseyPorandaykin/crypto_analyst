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
		query   = `SELECT DISTINCT crypto_analyst.symbol FROM price_changes`
		symbols []string
	)
	if err := repo.db.SelectContext(ctx, &symbols, query); err != nil {
		return nil, err
	}
	return symbols, nil
}
