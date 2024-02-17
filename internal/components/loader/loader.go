package loader

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_loader/api/http/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	_ "github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const DefaultLoadPriceDuration = 1 * time.Minute

type Loader struct {
	client             *client.Client
	priceStorage       domain.PriceSaver
	candlestickStorage domain.CandlestickStorage
	price              *Price
}

func NewLoader(
	client *client.Client,
	priceStorage domain.PriceSaver,
	candlestickStorage domain.CandlestickStorage,
	price *Price,
) *Loader {
	return &Loader{client: client, priceStorage: priceStorage, candlestickStorage: candlestickStorage, price: price}
}

func (l *Loader) Run(ctx context.Context) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error)
	go func() {
		if err := l.loadPrices(childCtx); err != nil {
			errCh <- err
		}
	}()
	go func() {
		if err := l.loadSymbolSnapshot(childCtx); err != nil {
			errCh <- err
		}
	}()
	go func() {
		if err := l.loadCandlesticks(childCtx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := l.price.Run(childCtx); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		}
	}
}

func (l *Loader) loadPrices(ctx context.Context) error {
	ticker := time.NewTicker(DefaultLoadPriceDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var sourcePrices []client.PriceResponse
			err := backoff.Retry(func() error {
				var err error
				sourcePrices, err = l.client.AllSymbolPrices(ctx)
				if err != nil {
					return err
				}
				return nil
			}, backoff.NewExponentialBackOff())
			if err != nil {
				return errors.Wrap(err, "error load price")
			}
			prices := make([]*domain.SymbolPrice, 0, len(sourcePrices))
			for _, item := range sourcePrices {
				price, err := strconv.ParseFloat(item.Price, 64)
				if err != nil {
					zap.L().Error("error parse price", zap.Error(err))
					continue
				}
				prices = append(prices, &domain.SymbolPrice{
					Exchange: item.Exchange,
					Symbol:   item.Symbol,
					Price:    price,
					Date:     item.Date,
				})
			}
			start := time.Now()
			errSave := backoff.Retry(func() error {
				return l.priceStorage.SavePrices(ctx, prices)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				zap.L().Error("error save symbolPrice", zap.Error(errSave))
			}
			metric.SavePriceDuration.Add(float64(time.Since(start).Milliseconds()))
			metric.SavePrices.Add(float64(len(prices)))
		}
	}
}

func (l *Loader) loadSymbolSnapshot(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for _, symbol := range domain.SubscribeSymbols {
				var candlesticks []dto.Candlestick
				err := backoff.Retry(func() error {
					var err error
					resp, err := l.client.SymbolSnapshot(ctx, domain.BinanceExchange, symbol)
					if err != nil {
						return errors.Wrap(err, "error get symbolSnapshot")
					}
					candlesticks = toCandlestick(
						resp.Symbol,
						resp.Exchange,
						resp.Candlestick4H,
						resp.Candlestick1H,
					)

					return nil
				}, backoff.NewExponentialBackOff())
				if err != nil {
					zap.L().Error("error load SymbolSnapshot", zap.Error(err))
					continue
				}
				errSave := backoff.Retry(func() error {
					return l.candlestickStorage.Save(ctx, candlesticks)
				}, backoff.NewExponentialBackOff())
				if errSave != nil {
					zap.L().Error("error save snapshot", zap.Error(errSave))
					continue
				}
				metric.SaveSnapshot.Inc()
			}
		}
	}
}

func (l *Loader) loadCandlesticks(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, symbol := range domain.SubscribeSymbols {
				for _, interval := range domain.ListIntervals {
					if err := l.fetchCandlesticks(ctx, domain.BinanceExchange, symbol, interval); err != nil {
						zap.L().Error("error fetch candlesticks", zap.Error(err))
					}
				}
			}
		}
	}
}

func (l *Loader) fetchCandlesticks(ctx context.Context, exchange, symbol, interval string) error {
	candlesticks := make([]dto.Candlestick, 0, 1000)
	err := backoff.Retry(func() error {
		resp, err := l.client.Candlesticks(ctx, exchange, symbol, interval)
		if err != nil {
			return err
		}
		data := make([]client.SymbolSnapshotCandlestick, 0, len(resp))
		now := time.Now().In(time.UTC)
		for _, item := range resp {
			if item.CloseTime.After(now) {
				continue
			}
			data = append(data, item)
		}
		candlesticks = toCandlestick(symbol, exchange, data...)
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return errors.Wrap(err, "error get candlesticks")
	}
	errSave := backoff.Retry(func() error {
		return l.candlestickStorage.Save(ctx, candlesticks)
	}, backoff.NewExponentialBackOff())
	if errSave != nil {
		return errors.Wrap(errSave, "error save candlesticks")
	}
	return nil
}

func toCandlestick(
	symbol, exchange string, data ...client.SymbolSnapshotCandlestick,
) []dto.Candlestick {
	candlesticks := make([]dto.Candlestick, 0, len(data))
	for _, item := range data {
		if item.OpenTime.IsZero() || item.CloseTime.IsZero() || item.OpenPrice == 0 || item.ClosePrice == 0 {
			continue
		}
		candlesticks = append(candlesticks, dto.Candlestick{
			Symbol:       symbol,
			Exchange:     exchange,
			OpenTime:     item.OpenTime,
			CloseTime:    item.CloseTime,
			OpenPrice:    item.OpenPrice,
			HighPrice:    item.HighPrice,
			LowPrice:     item.LowPrice,
			ClosePrice:   item.ClosePrice,
			Volume:       item.Volume,
			NumberTrades: item.NumberTrades,
			Interval:     item.Interval,
			CreatedAt:    item.CreatedAt,
		})
	}
	return candlesticks
}
