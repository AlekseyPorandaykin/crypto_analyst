package loader

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/metric"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/db"
	"github.com/AlekseyPorandaykin/crypto_loader/api/http/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type Price struct {
	client     *client.Client
	symbolRepo *db.Symbols
	priceRepo  *db.PriceRepository

	exchangeSymbols map[string]map[string]bool
	muSymbols       sync.Mutex
}

func NewPrice(client *client.Client, symbolRepo *db.Symbols, priceRepo *db.PriceRepository) *Price {
	return &Price{
		client:          client,
		symbolRepo:      symbolRepo,
		priceRepo:       priceRepo,
		exchangeSymbols: make(map[string]map[string]bool),
	}
}
func (p *Price) Run(ctx context.Context) error {
	errCh := make(chan error)
	for _, ex := range domain.ListExchanges {
		ex := ex
		go func(exchange string) {
			if err := p.loadExchangePrices(ctx, ex); err != nil {
				errCh <- err
			}
		}(ex)
	}
	go func() {
		if err := p.loadSymbols(ctx); err != nil {
			errCh <- err
			return
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.loadSymbols(ctx); err != nil {
					errCh <- err
					return
				}
			}
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

func (p *Price) loadExchangePrices(ctx context.Context, exchange string) error {
	ticker := time.NewTicker(time.Minute / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if len(p.exchangeSymbols) == 0 {
				continue
			}
			var sourcePrices []client.PriceResponse
			err := backoff.Retry(func() error {
				var err error
				sourcePrices, err = p.client.ExchangePrices(ctx, exchange)
				if err != nil {
					return err
				}
				return nil
			}, backoff.NewExponentialBackOff())
			if err != nil {
				return errors.Wrap(err, "error load exchange price")
			}
			prices := make([]domain.SymbolPrice, 0, len(sourcePrices))
			for _, item := range sourcePrices {
				if _, has := p.exchangeSymbols[item.Exchange]; has && p.exchangeSymbols[item.Exchange][item.Symbol] {
					continue
				}
				price, err := strconv.ParseFloat(item.Price, 64)
				if err != nil {
					zap.L().Error("error parse price", zap.Error(err))
					continue
				}
				prices = append(prices, domain.SymbolPrice{
					Exchange: item.Exchange,
					Symbol:   item.Symbol,
					Price:    price,
					Date:     item.Date,
				})
			}
			if len(prices) == 0 {
				continue
			}
			start := time.Now()
			errSave := backoff.Retry(func() error {
				return p.priceRepo.AddNewSymbol(ctx, prices)
			}, backoff.NewExponentialBackOff())
			if errSave != nil {
				return errors.Wrap(errSave, "save new symbols")
			}
			metric.SaveNewSymbolDuration.Add(float64(time.Since(start).Milliseconds()))
			metric.SaveNewSymbol.Add(float64(len(prices)))
		}
	}
}

func (p *Price) loadSymbols(ctx context.Context) error {
	exSymbols, err := p.symbolRepo.ExchangeSymbols(ctx)
	if err != nil {
		return err
	}
	p.muSymbols.Lock()
	defer p.muSymbols.Unlock()
	for _, exSymbol := range exSymbols {
		if _, has := p.exchangeSymbols[exSymbol.Exchange]; !has {
			p.exchangeSymbols[exSymbol.Exchange] = make(map[string]bool)
		}
		p.exchangeSymbols[exSymbol.Exchange][exSymbol.Symbol] = true
	}

	return nil
}
