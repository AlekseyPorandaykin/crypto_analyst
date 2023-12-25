package storage

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
)

type PriceStorage interface {
	PriceSaver
	PriceLoader
}

type PriceSaver interface {
	SavePrices(ctx context.Context, prices []*domain.SymbolPrice) error
}

type PriceLoader interface {
	Prices(ctx context.Context, symbol string) ([]domain.SymbolPrice, error)
}

type IndicatorLoader interface {
	Indicators(ctx context.Context, symbol string)
}
