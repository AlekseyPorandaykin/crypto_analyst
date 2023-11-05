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

func (repo *PriceChanges) Save(ctx context.Context, data []domain.PriceChange) error {
	var (
		query = `
INSERT INTO 
    crypto_analyst.price_changes(symbol, exchange, datetime, afg_value, price, prev_price, created_at) 
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
		query    = `SELECT coalesce(max(created_at), NOW()) FROM crypto_analyst.price_changes`
		datetime time.Time
	)
	if err := repo.db.GetContext(ctx, &datetime, query); err != nil {
		return time.Time{}, err
	}
	return datetime, nil
}

func (repo *PriceChanges) LastDatetimeSymbolRow(ctx context.Context, symbol string) (time.Time, error) {
	var (
		query    = `SELECT max(created_at) FROM crypto_analyst.price_changes WHERE symbol = $1`
		datetime time.Time
	)
	if err := repo.db.GetContext(ctx, &datetime, query, symbol); err != nil {
		return time.Time{}, err
	}
	return datetime, nil
}

func (repo *PriceChanges) FirstDatetimeRow(ctx context.Context) (time.Time, error) {
	var (
		query    = `SELECT TO_TIMESTAMP(datetime, 'YYYY/MM/DD/HH24:MI:ss') FROM crypto_analyst.price_changes ORDER BY datetime LIMIT 1`
		datetime time.Time
	)
	if err := repo.db.GetContext(ctx, &datetime, query); err != nil {
		return time.Time{}, err
	}
	datetime.In(time.Now().Location())
	return datetime, nil
}

func (repo *PriceChanges) List(ctx context.Context, symbol string, from, to time.Time) ([]domain.PriceChange, error) {
	var (
		query = `
SELECT symbol,
       exchange,
       TO_TIMESTAMP(datetime, 'YYYY/MM/DD/HH24:MI:ss') as datetime,
       afg_value,
       price,
       prev_price,
       created_at
FROM crypto_analyst.price_changes
WHERE symbol = $1
  AND created_at::DATE BETWEEN $2 AND $3
`
		res []domain.PriceChange
	)
	if err := repo.db.SelectContext(ctx, &res, query, symbol, from, to); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *PriceChanges) Exchanges(ctx context.Context) ([]string, error) {
	var (
		query = `SELECT DISTINCT exchange FROM crypto_analyst.price_changes ORDER BY exchange`
		res   []string
	)
	if err := repo.db.SelectContext(ctx, &res, query); err != nil {
		return nil, err
	}
	return res, nil
}
