package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/client/loader"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type Loader struct {
	loader *loader.Loader
	repo   *repositories.PriceRepository
}

func NewLoader(loader *loader.Loader, repo *repositories.PriceRepository) *Loader {
	return &Loader{loader: loader, repo: repo}
}

func (l *Loader) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch := <-l.loader.Batch():
			err := backoff.Retry(func() error {
				return l.repo.SavePrices(ctx, batch.Prices())
			}, backoff.NewExponentialBackOff())
			if err != nil {
				zap.L().Error("error save prices", zap.Error(err))
			}
		}
	}
}
