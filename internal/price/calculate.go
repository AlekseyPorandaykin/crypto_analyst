package price

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

type Calculate struct {
	symbolRepo       *repositories.Symbols
	priceChangesRepo *repositories.PriceChanges
}

func NewCalculate(symbolRepo *repositories.Symbols, priceChangesRepo *repositories.PriceChanges) *Calculate {
	return &Calculate{symbolRepo: symbolRepo, priceChangesRepo: priceChangesRepo}
}

func (c *Calculate) ReportAvg(ctx context.Context, from, to time.Time) error {
	//symbols, err := c.symbolRepo.List(ctx)
	//if err != nil {
	//	return errors.Wrap(err, "read list symbols")
	//}
	firstDate, err := c.fistDatetime(ctx)
	if err != nil {
		return errors.Wrap(err, "get fists date")
	}
	rows, err := c.toAvgReport(ctx, c.symbolPriceChanges(ctx, "BTCUSDT", firstDate, firstDate.Add(24*time.Hour)), from, to)
	if err != nil {
		return errors.Wrap(err, "generate data")
	}
	return c.saveCsv(rows)
}

func (c *Calculate) fistDatetime(ctx context.Context) (time.Time, error) {
	res, err := c.priceChangesRepo.LastDatetimeRow(ctx)
	if err != nil {
		zap.L().Error("error get LastDatetimeRow", zap.Error(err))
		return time.Time{}, errors.Wrap(err, "error get LastDatetimeRow")
	}
	res = res.In(time.UTC)
	return res, nil
}

func (c *Calculate) symbolPriceChanges(
	ctx context.Context, symbol string, from, end time.Time,
) map[time.Time]map[string]domain.PriceChange {
	data := make(map[time.Time]map[string]domain.PriceChange)
	for {
		if end.Before(from) {
			break
		}
		to := from.Add(24 * time.Hour)
		prices, err := c.priceChangesRepo.List(ctx, symbol, from, to)
		if err != nil {
			zap.L().Error("error get list priceChanges", zap.Error(err))
			break
		}
		for _, price := range prices {
			key := domain.ToDatetimeWithoutSec(price.Date)
			if _, has := data[key]; !has {
				data[key] = make(map[string]domain.PriceChange)
			}
			data[key][price.Exchange] = price
		}

		from = to
	}

	return data
}

func (c *Calculate) toAvgReport(
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
			rows = append(rows, strconv.Itoa(int(price.AfgValue)))
		}
		result = append(result, rows)
	}
	return result, err

}

func (c *Calculate) toPriceReport(
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
			rows = append(rows, fmt.Sprintf("%f.4", price.Price))
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

func (c *Calculate) saveCsv(data [][]string) error {
	file, err := os.Create("/Users/alexey.porandaikin/Projects/go/projects/crypto_analyst/storage/reports/avg.csv")
	if err != nil {
		return err
	}

	w := csv.NewWriter(file)
	if err := w.WriteAll(data); err != nil {
		return err
	}

	return nil
}
