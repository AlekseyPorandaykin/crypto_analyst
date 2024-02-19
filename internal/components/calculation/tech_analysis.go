package calculation

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"time"
)

type TechAnalysis struct {
	candlestickLoader domain.CandlestickLoader
}

func NewTechAnalysis(candlestickLoader domain.CandlestickLoader) *TechAnalysis {
	return &TechAnalysis{candlestickLoader: candlestickLoader}
}

func (ta *TechAnalysis) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := ta.loadCandlesticks(ctx); err != nil {
				return err
			}
		}
	}
}

func (ta *TechAnalysis) loadCandlesticks(ctx context.Context) error {
	series := techan.NewTimeSeries()
	candles, err := ta.candlestickLoader.Candlesticks(ctx, domain.BinanceExchange, "BTCUSDT", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return err
	}
	if len(candles) == 0 {
		return nil
	}
	for _, item := range candles {
		if item.Interval != domain.OneHourInterval {
			continue
		}
		candle := techan.NewCandle(techan.NewTimePeriod(item.OpenTime, time.Hour))
		candle.OpenPrice = big.NewDecimal(item.OpenPrice)
		candle.ClosePrice = big.NewDecimal(item.ClosePrice)
		candle.MaxPrice = big.NewDecimal(item.HighPrice)
		candle.MinPrice = big.NewDecimal(item.LowPrice)
		series.AddCandle(candle)
	}
	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, 10) // Create an exponential moving average with a window of 10
	trendlineIndicator := techan.NewTrendlineIndicator(closePrices, 10)
	fmt.Println("movingAverage", movingAverage.Calculate(0).FormattedString(2))
	fmt.Println("trendlineIndicator", trendlineIndicator.Calculate(10))

	return nil
}
