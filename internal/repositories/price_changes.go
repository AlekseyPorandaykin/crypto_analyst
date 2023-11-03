package repositories

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

const DatetimeFormat = "2006-01-02 15:04:05"

type PriceChanges struct {
	db *sqlx.DB
}

func NewPriceChanges(db *sqlx.DB) *PriceChanges {
	return &PriceChanges{
		db: db,
	}
}

func (repo *PriceChanges) Save(ctx context.Context, data []domain.AfgCoefficient) error {
	var (
		query = `
INSERT INTO 
    price_changes(symbol, exchange, datetime, afg_value, price, prev_price, created_at) 
VALUES %s ON CONFLICT (symbol, exchange, datetime) DO NOTHING
`
		params []string
	)
	for _, item := range data {
		params = append(
			params,
			fmt.Sprintf(
				"('%s', '%s', '%s', %d, %f, %f, '%s')",
				item.Symbol,
				item.Exchange,
				item.Date.Format(DatetimeFormat),
				item.AfgValue,
				item.Price,
				item.PrevPrice,
				item.CreatedAt.Format(DatetimeFormat),
			))
	}
	if _, err := repo.db.ExecContext(ctx, fmt.Sprintf(query, strings.Join(params, ","))); err != nil {
		return err
	}
	return nil
}

func (repo *PriceChanges) LastDatetimeRow(ctx context.Context) (time.Time, error) {
	var (
		query    = `SELECT coalesce(max(created_at), NOW() - INTERVAL '1 year') FROM price_changes`
		datetime time.Time
	)
	if err := repo.db.GetContext(ctx, &datetime, query); err != nil {
		return time.Time{}, err
	}
	return datetime, nil
}
