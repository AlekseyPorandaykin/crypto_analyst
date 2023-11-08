package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
)

type Aggregation struct {
	db *sqlx.DB
}

func NewAggregation(db *sqlx.DB) *Aggregation {
	return &Aggregation{db: db}
}

func (repo *Aggregation) LastRow(ctx context.Context, metric, symbol string) (*domain.PriceAggregation, error) {
	var (
		query = `
SELECT symbol, exchange, metric, key, value, updated_at
FROM crypto_analyst.price_aggregation
WHERE metric = $1 AND symbol = $2
ORDER BY updated_at DESC
LIMIT 1;
`
		dest = domain.PriceAggregation{}
	)
	if err := repo.db.GetContext(ctx, &dest, query, metric, symbol); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &dest, nil
}

func (repo *Aggregation) Save(ctx context.Context, data ...domain.PriceAggregation) error {
	var (
		query = `
INSERT INTO crypto_analyst.price_aggregation(symbol, exchange, metric, key, value, updated_at) 
VALUES %s ON CONFLICT (symbol, exchange, metric, key) 
DO  UPDATE SET value = EXCLUDED.value, updated_at=EXCLUDED.updated_at ;
`
		params []string
	)
	for _, item := range data {
		params = append(params, fmt.Sprintf(
			"('%s', '%s', '%s', '%s', '%s', '%s')",
			item.Symbol,
			item.Exchange,
			item.Metric,
			item.Key,
			item.Value,
			item.UpdatedAt.Format(DatetimeFormat),
		))
	}
	if _, err := repo.db.ExecContext(ctx, fmt.Sprintf(query, strings.Join(params, ","))); err != nil {
		return err
	}
	return nil
}
