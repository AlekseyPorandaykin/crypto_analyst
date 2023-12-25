package storage

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Composite struct {
	fastStorage     PriceStorage
	longTermStorage PriceStorage
}

func NewComposite(fastStorage, longTermStorage PriceStorage) *Composite {
	return &Composite{fastStorage: fastStorage, longTermStorage: longTermStorage}
}

func (c *Composite) SavePrices(ctx context.Context, prices []*domain.SymbolPrice) error {
	if err := c.fastStorage.SavePrices(ctx, prices); err != nil {
		return errors.Wrap(err, "error save in fastStorage")
	}

	if err := c.longTermStorage.SavePrices(ctx, prices); err != nil {
		return errors.Wrap(err, "error save in longTermStorage")
	}

	return nil
}

func (c *Composite) Prices(ctx context.Context, symbol string) ([]domain.SymbolPrice, error) {
	var (
		prices []domain.SymbolPrice
		err    error
	)
	prices, err = c.fastStorage.Prices(ctx, symbol)
	if err != nil {
		zap.L().Error("error get prices from fastStorage")
	}
	if len(prices) > 0 {
		return prices, nil
	}
	prices, err = c.longTermStorage.Prices(ctx, symbol)
	if err != nil {
		zap.L().Error("error get prices from longTermStorage")
		return nil, err
	}
	return prices, nil
}
