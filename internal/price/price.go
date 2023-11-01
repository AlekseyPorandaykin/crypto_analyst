package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type exchangePrices map[time.Time]map[string]float64

type Price struct {
	priceRepo        *repositories.PriceRepository
	priceChangesRepo *repositories.PriceChanges
}

func NewPrice(
	priceRepo *repositories.PriceRepository,
	priceChangesRepo *repositories.PriceChanges) *Price {
	return &Price{priceRepo: priceRepo, priceChangesRepo: priceChangesRepo}
}

func (p *Price) Run(ctx context.Context, d time.Duration) error {
	if err := p.execute(ctx); err != nil {
		return err
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.execute(ctx); err != nil {
				return err
			}
			ticker.Reset(d)
		}
	}
}

func (p *Price) execute(ctx context.Context) error {
	defer func(start time.Time) {
		zap.L().Info("Price analysis calculated", zap.String("execute_sec", time.Since(start).String()))
	}(time.Now())
	return backoff.Retry(func() error {
		return p.calculate(ctx)
	}, backoff.NewExponentialBackOff())
}

func (p *Price) calculate(ctx context.Context) error {
	firstDatetime := p.lastUpdateDatetime(ctx)
	zap.L().Debug("load first datetime", zap.Time("datetime", firstDatetime))
	symbols, err := p.priceRepo.PopularSymbols(ctx, 3)
	if err != nil {
		return errors.Wrap(err, "get all symbols")
	}
	for _, symbol := range symbols {
		zap.L().Debug("run avg_coefficient", zap.String("symbol", symbol))
		p.runAvgCoefficient(ctx, symbol, firstDatetime)
	}
	if err := p.clearOldPrices(ctx); err != nil {
		zap.L().Error("error clear old prices", zap.Error(err))
	}

	return nil
}

func (p *Price) runAvgCoefficient(ctx context.Context, symbol string, from time.Time) {
	for {
		to := from.Add(24 * time.Hour)
		data, keys, err := p.loadSymbolData(ctx, symbol, from, to)
		if err != nil {
			zap.L().Error(
				"load symbol data",
				zap.Error(err),
				zap.String("symbol", symbol),
				zap.Time("from", from),
				zap.Time("to", to),
			)
		}
		if len(data) == 0 || len(keys) == 0 {
			return
		}
		p.calculateAvgCoefficient(ctx, data, keys, symbol)
		from = to
	}
}

func (p *Price) loadSymbolData(ctx context.Context, symbol string, from, to time.Time) (exchangePrices, []time.Time, error) {
	data := make(exchangePrices)
	keys := make([]time.Time, 0, 100)
	symbolPrices, err := p.priceRepo.SymbolPrices(ctx, symbol, from, to)
	if err != nil {
		return nil, nil, err
	}
	zap.L().Debug("loaded symbol prices",
		zap.String("symbol", symbol),
		zap.Time("from", from),
		zap.Time("to", to),
		zap.Int("count", len(symbolPrices)),
	)
	for _, symbolPrice := range symbolPrices {
		key := time.Date(
			symbolPrice.Date.Year(),
			symbolPrice.Date.Month(),
			symbolPrice.Date.Day(),
			symbolPrice.Date.Hour(),
			symbolPrice.Date.Minute(),
			0,
			0,
			symbolPrice.Date.Location(),
		)

		if data[key] == nil {
			data[key] = make(map[string]float64)
			keys = append(keys, key)
		}

		data[key][symbolPrice.Exchange] = float64(symbolPrice.Price)
	}
	return data, keys, nil
}
func (p *Price) calculateAvgCoefficient(ctx context.Context, data exchangePrices, keys []time.Time, symbol string) {
	prevValues := make(map[string]float64)
	result := make([]domain.AfgCoefficient, 0, len(keys))
	for _, key := range keys {
		for exchange, val := range data[key] {
			if _, ok := prevValues[exchange]; ok {
				r := int((val - prevValues[exchange]) / val * 10000)
				result = append(result, domain.AfgCoefficient{
					Date:      key,
					Symbol:    symbol,
					Exchange:  exchange,
					AfgValue:  int64(r),
					Price:     val,
					PrevPrice: prevValues[exchange],
					CreatedAt: time.Now(),
				})
			}
		}
		prevValues = data[key]
	}
	if len(result) == 0 {
		return
	}
	if err := p.priceChangesRepo.Save(ctx, result); err != nil {
		zap.L().Error("save avg_coefficient", zap.Error(err))
	}
	zap.L().Debug("save avg_coefficient", zap.String("symbol", symbol), zap.Int("count", len(result)))
}

func (p *Price) lastUpdateDatetime(ctx context.Context) time.Time {
	lastDatetime, err := p.priceChangesRepo.LastDatetimeRow(ctx)
	if err != nil {
		zap.L().Error("init last update avg coefficient datetime", zap.Error(err))
	}
	if !lastDatetime.IsZero() {
		return lastDatetime
	}

	firstDatetime, err := p.priceRepo.FirstDatetime(ctx)
	if err != nil {
		zap.L().Error("init first datetime", zap.Error(err))
	}
	if !firstDatetime.IsZero() {
		return firstDatetime
	}

	return time.Now().Add(-10 * 24 * time.Hour)
}

func (p *Price) clearOldPrices(ctx context.Context) error {
	lastDatetime, err := p.priceChangesRepo.LastDatetimeRow(ctx)
	if err != nil {
		return errors.Wrap(err, "get last datetime")
	}
	if err := p.priceRepo.DeleteOldData(ctx, lastDatetime); err != nil {
		return errors.Wrap(err, "delete old prices")
	}
	return nil
}
