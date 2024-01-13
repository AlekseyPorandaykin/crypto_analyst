package db

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

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
    crypto_analyst.price_changes(symbol, exchange, datetime, coefficient_change, price, prev_price, created_at) 
VALUES %s ON CONFLICT (symbol, exchange, datetime) DO NOTHING
`
		params []string
	)
	if len(data) == 0 {
		return nil
	}
	for _, item := range data {
		params = append(
			params,
			fmt.Sprintf(
				"('%s', '%s', '%s', %d, %f, %f, '%s')",
				item.Symbol,
				item.Exchange,
				item.Date.Format(DatetimeFormat),
				item.CoefficientOfChange,
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
       coefficient_change,
       price,
       prev_price,
       created_at
FROM crypto_analyst.price_changes
WHERE symbol = $1
  AND created_at>= $2 AND created_at<=$3
`
		res []domain.PriceChange
	)
	if err := repo.db.SelectContext(ctx, &res, query, symbol, from, to); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *PriceChanges) Changes(ctx context.Context, exchange, symbol string, from, to time.Time) ([]domain.PriceChange, error) {
	var (
		query = `
SELECT symbol,
       exchange,
       TO_TIMESTAMP(datetime, 'YYYY/MM/DD/HH24:MI:ss') as datetime,
       coefficient_change,
       price,
       prev_price,
       created_at
FROM crypto_analyst.price_changes
WHERE exchange  = $1  AND symbol = $2
  AND created_at>= $3 AND created_at<=$4
ORDER BY datetime DESC 
`
		res []domain.PriceChange
	)
	rows, err := repo.db.QueryxContext(ctx, query, exchange, symbol, from, to)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		pc := domain.PriceChange{}
		if err := rows.StructScan(&pc); err != nil {
			return nil, err
		}
		pc.Date = pc.Date.In(time.UTC)
		res = append(res, pc)
	}
	return res, nil
}

func (repo *PriceChanges) DeleteOldRows(ctx context.Context, to time.Time) error {
	var query = `
DELETE FROM crypto_analyst.price_changes WHERE  datetime < $1
`
	_, err := repo.db.ExecContext(ctx, query, to)

	return err
}
