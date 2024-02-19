package calculation

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/db"
	"github.com/cenkalti/backoff/v4"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type exchangePrices map[time.Time]map[string]float64

type PriceChange struct {
	symbolRepo       *db.Symbols
	priceRepo        *db.PriceRepository
	priceChangesRepo *db.PriceChanges
}

func NewChangeCalculator(
	priceRepo *db.PriceRepository,
	priceChangesRepo *db.PriceChanges,
	symbolRepo *db.Symbols) *PriceChange {
	return &PriceChange{
		priceRepo: priceRepo, priceChangesRepo: priceChangesRepo, symbolRepo: symbolRepo,
	}
}

func (p *PriceChange) Run(ctx context.Context, d time.Duration) error {
	errCh := make(chan error)
	if err := p.execute(ctx); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
			case <-ticker.C:
				if err := p.execute(ctx); err != nil {
					errCh <- err
				}
				ticker.Reset(d)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
			case <-ticker.C:
				if err := p.priceChangesRepo.DeleteOldRows(ctx, time.Now().Add(-7*24*time.Hour)); err != nil {
					errCh <- errors.Wrap(err, "error delete old rows price_changes")
				}
			}
		}
	}()
	return <-errCh
}

func (p *PriceChange) execute(ctx context.Context) error {
	defer func(start time.Time) {
		metric.ChangePriceCalculateDuration.Add(float64(time.Since(start).Milliseconds()))
	}(time.Now())
	return backoff.Retry(func() error {
		return p.calculate(ctx)
	}, backoff.NewExponentialBackOff())
}

func (p *PriceChange) calculate(ctx context.Context) error {
	symbols, err := p.symbolRepo.PopularSymbols(ctx, 3)
	if err != nil {
		return errors.Wrap(err, "get all symbols")
	}
	metric.PopularSymbols.Add(float64(len(symbols)))
	for _, symbol := range symbols {
		p.calculatePriceChanges(ctx, symbol)
	}
	if err := p.priceRepo.ClearOldPrices(ctx, time.Now().In(time.UTC).Add(-1*time.Hour)); err != nil {
		zap.L().Error("error clear old symbolPrice", zap.Error(err))
	}
	return nil
}

func (p *PriceChange) calculatePriceChanges(ctx context.Context, symbol string) {
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
		if len(data) == 0 && from.Before(time.Now()) {
			from = from.Add(24 * time.Hour)
			continue
		}
		for date := range data {
			if date.After(to) {
				to = date
			}
		}
		if len(data) == 0 || len(keys) == 0 {
			break
		}

		coefficients := p.priceChanges(data, keys, symbol)

		if len(coefficients) > 0 {
			errSave := backoff.Retry(func() error {
				return p.priceChangesRepo.Save(ctx, coefficients)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				zap.L().Error("save CoefficientOfChange", zap.Error(errSave))
				break
			}
		}

		if err := p.priceRepo.DeletePrices(ctx, symbol, from, to.Add(-5*time.Minute)); err != nil {
			zap.L().Error("error delete symbolPrice", zap.Error(err))
			break
		}
		if !to.After(from) {
			break
		}
		from = to
	}

	if err := p.priceRepo.DeleteOldPrices(ctx, symbol, from.Add(-15*time.Minute)); err != nil {
		zap.L().Error("error delete old symbolPrice", zap.Error(err))
	}
}

func (p *PriceChange) loadSymbolData(ctx context.Context, symbol string, from, to time.Time) (exchangePrices, []time.Time, error) {
	data := make(exchangePrices)
	keys := make([]time.Time, 0, 100)
	symbolPrices, err := p.priceRepo.SymbolPrices(ctx, symbol, from, to)
	if err != nil {
		return nil, nil, err
	}
	for _, symbolPrice := range symbolPrices {
		key := domain.ToDatetimeWithoutSec(symbolPrice.Date)

		if data[key] == nil {
			data[key] = make(map[string]float64)
			keys = append(keys, key)
		}

		data[key][symbolPrice.Exchange] = symbolPrice.Price
	}
	return data, keys, nil
}
func (p *PriceChange) priceChanges(data exchangePrices, keys []time.Time, symbol string) []domain.PriceChange {
	prevValues := make(map[string]float64)
	result := make([]domain.PriceChange, 0, len(keys))
	for _, key := range keys {
		for exchange, val := range data[key] {
			if _, ok := prevValues[exchange]; ok {
				var coefficientOfChanges int64
				currentPrice := mathutil.RoundToFloat(val, 10)
				prevPrice := mathutil.RoundToFloat(prevValues[exchange], 10)
				if currentPrice > 0 && prevPrice > 0 {
					coefficientOfChanges = int64((currentPrice - prevPrice) / currentPrice * 10000)
				}
				if coefficientOfChanges > 100_000 || coefficientOfChanges < -100_000 {
					coefficientOfChanges = 0
				}
				result = append(result, domain.PriceChange{
					Date:                key,
					Symbol:              symbol,
					Exchange:            exchange,
					CoefficientOfChange: coefficientOfChanges,
					Price:               val,
					PrevPrice:           prevValues[exchange],
					CreatedAt:           time.Now().In(time.UTC),
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

func (p *PriceChange) lastUpdateDatetime(ctx context.Context, symbol string) time.Time {
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
