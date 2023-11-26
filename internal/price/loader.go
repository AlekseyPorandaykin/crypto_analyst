package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/AlekseyPorandaykin/crypto_loader/api/grpc/client"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"time"
)

type Loader struct {
	loader *client.Loader
	repo   *repositories.PriceRepository
}

func NewLoader(loader *client.Loader, repo *repositories.PriceRepository) *Loader {
	return &Loader{loader: loader, repo: repo}
}

func (l *Loader) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch := <-l.loader.Batch():
			start := time.Now()
			sourcePrices := batch.Prices()
			prices := make([]*domain.SymbolPrice, 0, len(sourcePrices))
			for _, item := range sourcePrices {
				prices = append(prices, &domain.SymbolPrice{
					Exchange: item.Exchange,
					Symbol:   item.Symbol,
					Price:    item.Price,
					Date:     item.Date,
				})
			}
			err := backoff.Retry(func() error {
				return l.repo.SavePrices(ctx, prices)
			}, backoff.NewExponentialBackOff())
			if err != nil {
				zap.L().Error("error save prices", zap.Error(err))
			}
			metric.SavePriceDuration.Add(float64(time.Since(start).Milliseconds()))
			metric.SavePrices.Add(float64(len(prices)))
		}
	}
}
