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

type ChangeCalculator struct {
	priceRepo        *repositories.PriceRepository
	priceChangesRepo *repositories.PriceChanges
}

func NewChangeCalculator(
	priceRepo *repositories.PriceRepository,
	priceChangesRepo *repositories.PriceChanges) *ChangeCalculator {
	return &ChangeCalculator{priceRepo: priceRepo, priceChangesRepo: priceChangesRepo}
}

func (p *ChangeCalculator) Run(ctx context.Context, d time.Duration) error {
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

func (p *ChangeCalculator) execute(ctx context.Context) error {
	defer func(start time.Time) {
		zap.L().Info("ChangeCalculator analysis calculated", zap.String("execute_sec", time.Since(start).String()))
	}(time.Now())
	return backoff.Retry(func() error {
		return p.calculate(ctx)
	}, backoff.NewExponentialBackOff())
}

func (p *ChangeCalculator) calculate(ctx context.Context) error {
	symbols, err := p.priceRepo.PopularSymbols(ctx, 3)
	if err != nil {
		return errors.Wrap(err, "get all symbols")
	}
	for _, symbol := range symbols {
		zap.L().Debug("run avg_coefficient", zap.String("symbolPriceChanges", symbol))
		p.runAvgCoefficient(ctx, symbol)
	}
	if err := p.priceRepo.ClearOldPrices(ctx, time.Now().In(time.UTC).Add(-1*time.Hour)); err != nil {
		zap.L().Error("error clear old prices", zap.Error(err))
	}
	return nil
}

func (p *ChangeCalculator) runAvgCoefficient(ctx context.Context, symbol string) {
	from := p.lastUpdateDatetime(ctx, symbol)
	if from.IsZero() {
		return
	}
	to := from
	for {
		data, keys, err := p.loadSymbolData(ctx, symbol, from.Add(-5*time.Minute), from.Add(24*time.Hour))
		if err != nil {
			zap.L().Error(
				"load symbolPriceChanges data",
				zap.Error(err),
				zap.String("symbolPriceChanges", symbol),
				zap.Time("from", from),
			)
		}
		for date := range data {
			if date.After(to) {
				to = date
			}
		}
		if len(data) == 0 || len(keys) == 0 {
			break
		}

		coefficients := p.calculateAvgCoefficient(data, keys, symbol)

		zap.L().Debug(
			"save avg_coefficient",
			zap.String("symbolPriceChanges", symbol),
			zap.Int("count", len(coefficients)),
		)
		if err := p.priceChangesRepo.Save(ctx, coefficients); err != nil {
			zap.L().Error("save avg_coefficient", zap.Error(err))
			break
		}

		if err := p.priceRepo.DeletePrices(ctx, symbol, from, to.Add(-5*time.Minute)); err != nil {
			zap.L().Error("error delete prices", zap.Error(err))
			break
		}
		if !to.After(from) {
			break
		}
		from = to
	}

	if err := p.priceRepo.DeleteOldPrices(ctx, symbol, from.Add(-15*time.Minute)); err != nil {
		zap.L().Error("error delete old prices", zap.Error(err))
	}
}

func (p *ChangeCalculator) loadSymbolData(ctx context.Context, symbol string, from, to time.Time) (exchangePrices, []time.Time, error) {
	data := make(exchangePrices)
	keys := make([]time.Time, 0, 100)
	symbolPrices, err := p.priceRepo.SymbolPrices(ctx, symbol, from, to)
	if err != nil {
		return nil, nil, err
	}
	zap.L().Debug("loaded symbolPriceChanges prices",
		zap.String("symbolPriceChanges", symbol),
		zap.Time("from", from),
		zap.Time("to", to),
		zap.Int("count", len(symbolPrices)),
	)
	for _, symbolPrice := range symbolPrices {
		key := domain.ToDatetimeWithoutSec(symbolPrice.Date)

		if data[key] == nil {
			data[key] = make(map[string]float64)
			keys = append(keys, key)
		}

		data[key][symbolPrice.Exchange] = float64(symbolPrice.Price)
	}
	return data, keys, nil
}
func (p *ChangeCalculator) calculateAvgCoefficient(data exchangePrices, keys []time.Time, symbol string) []domain.PriceChange {
	prevValues := make(map[string]float64)
	result := make([]domain.PriceChange, 0, len(keys))
	for _, key := range keys {
		for exchange, val := range data[key] {
			if _, ok := prevValues[exchange]; ok {
				r := int((val - prevValues[exchange]) / val * 10000)
				result = append(result, domain.PriceChange{
					Date:      key,
					Symbol:    symbol,
					Exchange:  exchange,
					AfgValue:  int64(r),
					Price:     val,
					PrevPrice: prevValues[exchange],
					CreatedAt: time.Now().In(time.UTC),
				})
			}
		}
		prevValues = data[key]
	}
	if len(result) == 0 {
		return nil
	}

	return result
}

func (p *ChangeCalculator) lastUpdateDatetime(ctx context.Context, symbol string) time.Time {
	lastDatetime, err := p.priceChangesRepo.LastDatetimeSymbolRow(ctx, symbol)
	if err != nil {
		zap.L().Error("init last update avg coefficient datetime", zap.Error(err))
	}
	if !lastDatetime.IsZero() {
		return lastDatetime.In(time.UTC)
	}

	firstDatetime, err := p.priceRepo.FirstDatetime(ctx, symbol)
	if err != nil {
		zap.L().Error("init first datetime", zap.Error(err))
	}
	if !firstDatetime.IsZero() {
		return firstDatetime.In(time.UTC)
	}

	return time.Now().Add(-10 * 24 * time.Hour).In(time.UTC)
}
