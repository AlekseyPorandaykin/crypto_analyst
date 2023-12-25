package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage"
	"github.com/AlekseyPorandaykin/crypto_loader/api/http/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	_ "github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const DefaultLoadPriceDurationSec = 60 * time.Second

type Loader struct {
	client  *client.Client
	storage storage.PriceSaver
}

func NewLoader(client *client.Client, storage storage.PriceSaver) *Loader {
	return &Loader{client: client, storage: storage}
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
	ticker := time.NewTicker(DefaultLoadPriceDurationSec)
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
				return l.storage.SavePrices(ctx, prices)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				zap.L().Error("error save prices", zap.Error(errSave))
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
			var snapshot client.SymbolSnapshotResponse
			err := backoff.Retry(func() error {
				var err error
				snapshot, err = l.client.SymbolSnapshot(ctx, "binance", "BTCUSDT")
				if err != nil {
					return err
				}
				return nil
			}, backoff.NewExponentialBackOff())
			if err != nil {
				return errors.Wrap(err, "error load SymbolSnapshot")
			}
			_ = snapshot
			metric.SaveSnapshot.Inc()
		}
	}
}
