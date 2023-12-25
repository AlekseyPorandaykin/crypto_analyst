package domain

import (
	"github.com/duke-git/lancet/v2/datetime"
	"time"
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
