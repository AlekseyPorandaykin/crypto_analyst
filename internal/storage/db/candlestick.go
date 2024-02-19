package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
	"github.com/jmoiron/sqlx"
)

type Candlestick struct {
	db *sqlx.DB
}

func NewCandlestick(db *sqlx.DB) *Candlestick {
	return &Candlestick{db: db}
}

func (repo *Candlestick) Save(ctx context.Context, data []dto.Candlestick) error {
	var values []string

	if len(data) == 0 {
		return nil
	}
	for _, item := range data {
		values = append(
			values,
			fmt.Sprintf(
				"('%s','%s','%s','%s',%f,%f,%f,%f,%f, %d, '%s','%s')",
				item.Symbol,
				item.Exchange,
				item.OpenTime.In(time.UTC).Format(DatetimeFormat),
				item.CloseTime.In(time.UTC).Format(DatetimeFormat),
				item.OpenPrice,
				item.HighPrice,
				item.LowPrice,
				item.ClosePrice,
				item.Volume,
				item.NumberTrades,
				item.Interval,
				item.CreatedAt.In(time.UTC).Format(DatetimeFormat),
			),
		)
	}
	query := fmt.Sprintf(
		`
INSERT INTO 
    crypto_analyst.candlesticks(symbol, exchange, open_time, close_time, open_price, high_price, low_price,close_price, volume, number_trades, candle_interval, created_at)
VALUES %s ON CONFLICT (symbol, exchange, open_time, close_time, candle_interval) DO NOTHING
`,
		strings.Join(values, ", "),
	)
	_, err := repo.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (repo *Candlestick) Candlesticks(ctx context.Context, exchange, symbol string, from, to time.Time) ([]dto.Candlestick, error) {
	var (
		query = `
SELECT symbol,
       exchange,
       open_time,
       close_time,
       open_price,
       high_price,
       low_price,
       close_price,
       volume,
       number_trades,
       candle_interval,
       created_at
FROM crypto_analyst.candlesticks
WHERE exchange = $1 
  AND symbol = $2
AND created_at>= $3 AND created_at<=$4
ORDER BY close_time
`
		result []dto.Candlestick
	)
	if err := repo.db.SelectContext(ctx, &result, query, exchange, symbol, from, to); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Candlestick) LastCandlestick(ctx context.Context, exchange, symbol, interval string) (*dto.Candlestick, error) {
	var (
		query = `
SELECT symbol,
       exchange,
       open_time,
       close_time,
       open_price,
       high_price,
       low_price,
       close_price,
       volume,
       number_trades,
       candle_interval,
       created_at
FROM crypto_analyst.candlesticks
WHERE exchange = $1 
  AND symbol = $2
  AND candle_interval = $3
ORDER BY close_time
LIMIT 1
`
		result dto.Candlestick
	)
	if err := repo.db.GetContext(ctx, &result, query, exchange, symbol, interval); err != nil {
		return nil, err
	}
	return &result, nil
}
