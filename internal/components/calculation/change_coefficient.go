package calculation

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/db"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ExchangePriceChanges map[string][]domain.PriceChange

type ChangeCoefficient struct {
	priceChangesRepo *db.PriceChanges
	symbolsRepo      *db.Symbols
	repo             *db.Aggregation
}

func NewChangeCoefficient(
	priceChangesRepo *db.PriceChanges,
	repo *db.Aggregation,
	symbolsRepo *db.Symbols,
) *ChangeCoefficient {
	return &ChangeCoefficient{priceChangesRepo: priceChangesRepo, repo: repo, symbolsRepo: symbolsRepo}
}

func (s *ChangeCoefficient) Run(ctx context.Context, d time.Duration) {
	changeCoefficientMetrics := []domain.MetricAggregationPrice{
		domain.ChangeCoefficientOnHour,
		domain.ChangeCoefficientOnDay,
		domain.ChangeCoefficientOnWeek,
	}
	for _, metric := range changeCoefficientMetrics {
		go func(metric domain.MetricAggregationPrice) {
			ticker := time.NewTicker(d)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := s.executeChangeCoefficient(ctx, metric); err != nil {
						zap.L().Error(
							"error execute change coefficient",
							zap.Error(err),
							zap.String("metric", string(metric)),
						)
					}
					ticker.Reset(d)
				}
			}
		}(metric)
	}
	indicatorChangesMetrics := []domain.MetricAggregationPrice{
		domain.IndicatorChangeOnHour,
		domain.IndicatorChangeOnHour,
		domain.IndicatorChangeOnWeek,
	}
	for _, metric := range indicatorChangesMetrics {
		go func(metric domain.MetricAggregationPrice) {
			ticker := time.NewTicker(d)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := s.executeIndicatorChanges(ctx, metric); err != nil {
						zap.L().Error(
							"error execute change coefficient",
							zap.Error(err),
							zap.String("metric", string(metric)),
						)
					}
					ticker.Reset(d)
				}
			}
		}(metric)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.repo.DeleteOldRows(ctx, time.Now().Add(-7*24*time.Hour)); err != nil {
					zap.L().Error(
						"error delete old change coefficient",
						zap.Error(err),
					)
				}
			}
		}
	}()
}

func (s *ChangeCoefficient) executeChangeCoefficient(
	ctx context.Context, m domain.MetricAggregationPrice,
) error {
	defer func(start time.Time) {
		metric.CoefficientDuration.WithLabelValues(string(m)).Add(float64(time.Since(start).Milliseconds()))
	}(time.Now())
	symbols, err := s.symbolsRepo.List(ctx)
	if err != nil {
		return errors.Wrap(err, "get all symbols")
	}
	for _, symbol := range symbols {
		from := s.lastTimeUpdateMetric(ctx, m, symbol)
		priceChanges, err := s.listPriceChanges(ctx, symbol, from)
		if err != nil {
			return errors.Wrap(err, "get price changes")
		}
		prices := s.changeCoefficient(priceChanges, symbol, m)
		if len(prices) > 0 {
			errSave := backoff.Retry(func() error {
				return s.repo.Save(ctx, prices...)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				return errors.Wrap(errSave, "save change_coefficient")
			}
		}
	}

	return nil
}

func (s *ChangeCoefficient) executeIndicatorChanges(
	ctx context.Context, m domain.MetricAggregationPrice,
) error {
	defer func(start time.Time) {
		metric.CoefficientDuration.WithLabelValues(string(m)).Add(float64(time.Since(start).Milliseconds()))
	}(time.Now())
	symbols, err := s.symbolsRepo.List(ctx)
	if err != nil {
		return errors.Wrap(err, "get all symbols")
	}
	for _, symbol := range symbols {
		from := s.lastTimeUpdateMetric(ctx, m, symbol)
		priceChanges, err := s.listPriceChanges(ctx, symbol, from)
		if err != nil {
			return errors.Wrap(err, "get price changes")
		}
		prices := s.indicatorChanges(priceChanges, symbol, m)
		if len(prices) > 0 {
			errSave := backoff.Retry(func() error {
				return s.repo.Save(ctx, prices...)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				return errors.Wrap(errSave, "save indicator_changes")
			}
		}
	}

	return nil
}

func (s *ChangeCoefficient) listPriceChanges(ctx context.Context, symbol string, from *time.Time) (ExchangePriceChanges, error) {
	res := make(ExchangePriceChanges)
	if from == nil {
		firstDate, err := s.priceChangesRepo.FirstDatetimeRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get first datetime price")
		}
		from = &firstDate
	}
	priceChanges, err := s.priceChangesRepo.List(ctx, symbol, *from, (*from).Add(24*time.Hour))
	if err != nil {
		return nil, errors.Wrap(err, "get price changes")
	}
	for _, item := range priceChanges {
		res[item.Exchange] = append(res[item.Exchange], item)
	}
	return res, nil
}

func (s *ChangeCoefficient) changeCoefficient(
	data map[string][]domain.PriceChange, symbol string, metric domain.MetricAggregationPrice,
) []domain.PriceAggregation {
	res := make([]domain.PriceAggregation, 0, len(data))
	for exchange, priceChanges := range data {
		timeCoefficients := make(map[time.Time][]float64)
		for _, priceChange := range priceChanges {
			key := changeCoefficientKeyByMetric(metric, priceChange.Date)
			if key.IsZero() {
				continue
			}
			timeCoefficients[key] = append(timeCoefficients[key], float64(priceChange.CoefficientOfChange))
		}
		for key, item := range timeCoefficients {
			res = append(res, domain.PriceAggregation{
				Symbol:    symbol,
				Exchange:  exchange,
				Metric:    metric,
				Key:       key.Format(time.DateTime),
				Value:     fmt.Sprintf("%.2f", avgValue(item)),
				UpdatedAt: time.Now().In(time.UTC),
			})
		}
	}
	return res
}

func (s *ChangeCoefficient) indicatorChanges(
	data map[string][]domain.PriceChange, symbol string, metric domain.MetricAggregationPrice,
) []domain.PriceAggregation {
	res := make([]domain.PriceAggregation, 0, len(data))
	dedublication := make(map[string]bool)
	for exchange, priceChanges := range data {
		timeCoefficients := make(map[time.Time][]float64)
		for _, priceChange := range priceChanges {
			key := changeCoefficientKeyByMetric(metric, priceChange.Date)
			if key.IsZero() {
				continue
			}
			timeCoefficients[key] = append(timeCoefficients[key], math.Abs(float64(priceChange.CoefficientOfChange)))
		}
		for key, item := range timeCoefficients {
			pa := domain.PriceAggregation{
				Symbol:    symbol,
				Exchange:  exchange,
				Metric:    metric,
				Key:       key.Format(time.DateTime),
				Value:     fmt.Sprintf("%.2f", avgValue(item)),
				UpdatedAt: time.Now().In(time.UTC),
			}
			if dedublication[pa.UniqKey()] {
				continue
			}
			res = append(res, pa)
			dedublication[pa.UniqKey()] = true
		}
	}
	return res
}

func changeCoefficientKeyByMetric(metric domain.MetricAggregationPrice, val time.Time) time.Time {
	switch metric {
	case domain.ChangeCoefficientOnHour, domain.IndicatorChangeOnHour:
		return domain.ToDatetimeWithoutMin(val)
	case domain.ChangeCoefficientOnDay, domain.IndicatorChangeOnDay:
		return domain.ToDatetimeWithoutDay(val)
	case domain.ChangeCoefficientOnWeek, domain.IndicatorChangeOnWeek:
		return domain.ToDatetimeWeek(val)
	}
	return time.Time{}
}

func (s *ChangeCoefficient) lastTimeUpdateMetric(
	ctx context.Context, metric domain.MetricAggregationPrice, symbol string,
) *time.Time {
	priceAggr, err := s.repo.LastRow(ctx, string(metric), symbol)
	if err != nil {
		zap.L().Error("get last row", zap.Error(err))
		return nil
	}
	if priceAggr == nil {
		return nil
	}
	t, err := time.Parse(time.DateTime, priceAggr.Key)
	if err != nil {
		zap.L().Error("error parse key", zap.Error(err))
		return nil
	}
	t.In(time.UTC)
	return &t
}

func avgValue(data []float64) float64 {
	var sum float64
	for _, item := range data {
		sum += item
	}
	if sum == 0 {
		return 0
	}
	return sum / float64(len(data))
}
