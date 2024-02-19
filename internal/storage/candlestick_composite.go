package storage

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
	"github.com/pkg/errors"
)

type CandlestickComposite struct {
	fastStorage     domain.CandlestickStorage
	longTermStorage domain.CandlestickStorage
}

func NewCandlestickComposite(fastStorage, longTermStorage domain.CandlestickStorage) *CandlestickComposite {
	return &CandlestickComposite{fastStorage: fastStorage, longTermStorage: longTermStorage}
}

func (c *CandlestickComposite) Save(ctx context.Context, data []dto.Candlestick) error {
	if len(data) == 0 {
		return nil
	}
	if err := c.fastStorage.Save(ctx, data); err != nil {
		return errors.Wrap(err, "error save in fastStorage")
	}
	if err := c.longTermStorage.Save(ctx, data); err != nil {
		return errors.Wrap(err, "error save in longTermStorage")
	}
	return nil
}

func (c *CandlestickComposite) Candlesticks(ctx context.Context, exchange, symbol string, from, to time.Time) ([]dto.Candlestick, error) {
	var (
		candlesticks []dto.Candlestick
		err          error
	)
	candlesticks, err = c.fastStorage.Candlesticks(ctx, exchange, symbol, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error get candlesticks from fastStorage")
	}
	if len(candlesticks) > 0 {
		return candlesticks, nil
	}
	candlesticks, err = c.longTermStorage.Candlesticks(ctx, exchange, symbol, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error get candlesticks from longTermStorage")
	}
	return candlesticks, nil
}

func (c *CandlestickComposite) LastCandlestick(ctx context.Context, exchange, symbol, interval string) (*dto.Candlestick, error) {
	var (
		candlestick *dto.Candlestick
		err         error
	)
	candlestick, err = c.fastStorage.LastCandlestick(ctx, exchange, symbol, interval)
	if err != nil {
		return nil, errors.Wrap(err, "error get candlesticks from fastStorage")
	}
	if candlestick != nil {
		return candlestick, nil
	}
	candlestick, err = c.longTermStorage.LastCandlestick(ctx, exchange, symbol, interval)
	if err != nil {
		return nil, errors.Wrap(err, "error get candlesticks from longTermStorage")
	}
	return candlestick, nil
}
