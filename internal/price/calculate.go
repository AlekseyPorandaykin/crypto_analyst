package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Calculate struct {
	symbolRepo       *repositories.Symbols
	priceChangesRepo *repositories.PriceChanges
}

type Report struct {
	Headers []string `json:"headers"`
	Data    []interface{}
}

func NewCalculate(
	symbolRepo *repositories.Symbols,
	priceChangesRepo *repositories.PriceChanges,
) *Calculate {
	return &Calculate{symbolRepo: symbolRepo, priceChangesRepo: priceChangesRepo}
}

func (c *Calculate) ReportPriceChanges(ctx context.Context, symbol string, from, to time.Time) ([][]string, error) {
	rows, err := c.toPriceChangeReport(ctx,
		c.symbolPriceChanges(ctx, symbol, from, to),
		from,
		to)
	if err != nil {
		return nil, errors.Wrap(err, "generate data")
	}
	return rows, nil
}

func (c *Calculate) symbolPriceChanges(
	ctx context.Context, symbol string, from, end time.Time,
) map[time.Time]map[string]domain.PriceChange {
	data := make(map[time.Time]map[string]domain.PriceChange)
	prices, err := c.priceChangesRepo.List(ctx, symbol, from, end)
	if err != nil {
		zap.L().Error("error get list priceChanges", zap.Error(err))
		return data
	}
	for _, price := range prices {
		key := domain.ToDatetimeWithoutSec(price.Date)
		if _, has := data[key]; !has {
			data[key] = make(map[string]domain.PriceChange)
		}
		data[key][price.Exchange] = price
	}

	return data
}

func (c *Calculate) toPriceChangeReport(
	ctx context.Context, data map[time.Time]map[string]domain.PriceChange, from, to time.Time,
) ([][]string, error) {
	exchanges, err := c.priceChangesRepo.Exchanges(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get exchanges")
	}
	dates := c.dates(from, to)
	result := make([][]string, 0, len(dates)+1)
	headers := make([]string, 0, len(exchanges)+1)
	headers = append(headers, "Dates")
	for _, exchange := range exchanges {
		headers = append(headers, exchange)
	}
	result = append(result, headers)
	for _, date := range dates {
		rows := make([]string, 0, len(exchanges))
		rows = append(rows, date.Format(time.DateTime))
		for _, exchange := range exchanges {
			if _, has := data[date]; !has {
				rows = append(rows, "")
				continue
			}
			price := data[date][exchange]
			rows = append(rows, strconv.Itoa(int(price.CoefficientOfChange)))
		}
		result = append(result, rows)
	}
	return result, err

}

func (c *Calculate) dates(from, to time.Time) []time.Time {
	res := make([]time.Time, 0, 100000)
	datetime := from
	for {
		if to.Before(datetime) {
			break
		}
		res = append(res, domain.ToDatetimeWithoutSec(datetime))
		datetime = datetime.Add(time.Minute)
	}

	return res
}
