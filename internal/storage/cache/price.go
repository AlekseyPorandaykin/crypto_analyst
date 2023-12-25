package cache

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage"
	"github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

var _ storage.PriceStorage = (*Price)(nil)

type Price struct {
	lastPrices  map[string][]domain.SymbolPrice
	muLastPrice sync.Mutex

	prices        map[string]*lru.Cache[time.Time, domain.SymbolPrice]
	muCachePrices sync.Mutex
}

func NewPrice() *Price {
	return &Price{
		lastPrices: make(map[string][]domain.SymbolPrice, 1000),
		prices:     make(map[string]*lru.Cache[time.Time, domain.SymbolPrice], 1000),
	}
}

func (c *Price) SavePrices(ctx context.Context, prices []*domain.SymbolPrice) error {
	c.muCachePrices.Lock()
	defer c.muCachePrices.Unlock()
	newPrices := make(map[string][]domain.SymbolPrice, 1000)
	for _, price := range prices {
		key := fmt.Sprintf("%s-%s", price.Exchange, price.Symbol)
		if _, has := c.prices[key]; !has {
			lruCache, err := lru.New[time.Time, domain.SymbolPrice](1500)
			if err != nil {
				return err
			}
			c.prices[key] = lruCache
		}
		newPrices[price.Symbol] = append(newPrices[price.Symbol], *price)
		c.prices[key].Add(domain.ToDatetimeWithoutSec(price.Date), *price)
	}
	c.muLastPrice.Lock()
	c.lastPrices = newPrices
	c.muLastPrice.Unlock()
	return nil
}

func (c *Price) Prices(ctx context.Context, symbol string) ([]domain.SymbolPrice, error) {
	c.muLastPrice.Lock()
	defer c.muLastPrice.Unlock()
	return c.lastPrices[symbol], nil
}

func (c *Price) ExchangeSymbolPrices(ctx context.Context, exchange, symbol string) ([]domain.SymbolPrice, error) {
	c.muCachePrices.Lock()
	lruCache := c.prices[fmt.Sprintf("%s-%s", exchange, symbol)]
	c.muCachePrices.Unlock()
	if lruCache == nil {
		return nil, nil
	}

	return lruCache.Values(), nil
}
