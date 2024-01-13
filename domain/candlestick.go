package domain

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
	"github.com/shopspring/decimal"
	"time"
)

const (
	FourHourInterval = "4h"
	OneHourInterval  = "1h"
)

var ListIntervals = []string{FourHourInterval, OneHourInterval}

type Candlestick struct {
	Symbol       string
	Exchange     string
	OpenTime     time.Time
	CloseTime    time.Time
	OpenPrice    decimal.Decimal
	HighPrice    decimal.Decimal
	LowPrice     decimal.Decimal
	ClosePrice   decimal.Decimal
	Volume       decimal.Decimal
	NumberTrades int
	Interval     string
	CreatedAt    time.Time
}

type CandlestickStorage interface {
	CandlestickSaver
	CandlestickLoader
}

type CandlestickSaver interface {
	Save(ctx context.Context, data []dto.Candlestick) error
}

type CandlestickLoader interface {
	Candlesticks(ctx context.Context, exchange, symbol string, from, to time.Time) ([]dto.Candlestick, error)
	LastCandlestick(ctx context.Context, exchange, symbol, interval string) (*dto.Candlestick, error)
}
