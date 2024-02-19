package cache

import (
	"context"
	"sync"
	"time"

	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
)

const LengthCandlestickData = 100

type Candlestick struct {
	lastData map[string]map[string]map[string]dto.Candlestick
	mu       sync.Mutex
}

func NewCandlestick() *Candlestick {
	return &Candlestick{
		lastData: make(map[string]map[string]map[string]dto.Candlestick),
	}
}

func (s *Candlestick) Save(ctx context.Context, data []dto.Candlestick) error {
	if len(data) == 0 {
		return nil
	}
	s.mu.Lock()
	for _, item := range data {
		if s.lastData[item.Exchange] == nil {
			s.lastData[item.Exchange] = make(map[string]map[string]dto.Candlestick)
		}
		if s.lastData[item.Exchange][item.Symbol] == nil {
			s.lastData[item.Exchange][item.Symbol] = make(map[string]dto.Candlestick)
		}
		prevVal := s.lastData[item.Exchange][item.Symbol][item.Interval]
		if prevVal.CloseTime.Before(item.CloseTime) {
			s.lastData[item.Exchange][item.Symbol][item.Interval] = item
		}
	}
	s.mu.Unlock()
	return nil
}

func (s *Candlestick) Candlesticks(ctx context.Context, exchange, symbol string, from, to time.Time) ([]dto.Candlestick, error) {
	return nil, nil
}

func (s *Candlestick) LastCandlestick(ctx context.Context, exchange, symbol, interval string) (*dto.Candlestick, error) {
	s.mu.Lock()
	lastData := s.lastData[exchange]
	defer s.mu.Unlock()
	symbolData := lastData[symbol]
	if symbolData == nil {
		return nil, nil
	}
	for key, val := range symbolData {
		if key == interval {
			return &val, nil
		}
	}
	return nil, nil
}
