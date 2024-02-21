package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
)

type SymbolPrice struct {
	Exchange string    `json:"exchange" db:"exchange"`
	Symbol   string    `json:"symbol" db:"symbol"`
	Price    float64   `json:"price" db:"price"`
	Date     time.Time `json:"date" db:"datetime"`
}

type PriceChange struct {
	Date                time.Time `json:"date" db:"datetime"`
	Symbol              string    `json:"symbol" db:"symbol"`
	Exchange            string    `json:"exchange" db:"exchange"`
	CoefficientOfChange int64     `json:"coefficient_change" db:"coefficient_change"`
	Price               float64   `json:"price" db:"price"`
	PrevPrice           float64   `json:"prev_price" db:"prev_price"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

func (p PriceChange) PriceString() string {
	return fmt.Sprintf("%.6f", p.Price)
}
func (p PriceChange) PrevPriceString() string {
	return fmt.Sprintf("%.6f", p.PrevPrice)
}

func ToDatetimeWithoutSec(val time.Time) time.Time {
	return time.Date(
		val.Year(),
		val.Month(),
		val.Day(),
		val.Hour(),
		val.Minute(),
		0,
		0,
		time.UTC,
	)
}

func ToDatetimeWithoutMin(val time.Time) time.Time {
	return time.Date(
		val.Year(),
		val.Month(),
		val.Day(),
		val.Hour(),
		0,
		0,
		0,
		time.UTC,
	)
}

func ToDatetimeWithoutHour(val time.Time) time.Time {
	return time.Date(
		val.Year(),
		val.Month(),
		val.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func ToDatetimeWeek(val time.Time) time.Time {
	endOfWeek := datetime.EndOfWeek(val)
	return time.Date(
		endOfWeek.Year(),
		endOfWeek.Month(),
		endOfWeek.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func ToDatetimeWithoutDay(val time.Time) time.Time {
	return time.Date(
		val.Year(),
		val.Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

type PriceStorage interface {
	PriceSaver
	PriceLoader
}

type PriceSaver interface {
	SavePrices(ctx context.Context, prices []*SymbolPrice) error
}

type PriceLoader interface {
	Prices(ctx context.Context, symbol string) ([]SymbolPrice, error)
}

type PriceChangeLoader interface {
	Changes(ctx context.Context, exchange, symbol string, from, to time.Time) ([]PriceChange, error)
}
