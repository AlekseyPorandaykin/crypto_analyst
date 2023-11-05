package domain

import (
	"time"
)

type SymbolPrice struct {
	Exchange string    `json:"exchange" db:"exchange"`
	Symbol   string    `json:"symbol" db:"symbol"`
	Price    float32   `json:"price" db:"price"`
	Date     time.Time `json:"date" db:"datetime"`
}

type PriceChange struct {
	Date      time.Time `json:"date" db:"datetime"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Exchange  string    `json:"exchange" db:"exchange"`
	AfgValue  int64     `json:"afgValue" db:"afg_value"`
	Price     float64   `json:"price" db:"price"`
	PrevPrice float64   `json:"prevPrice" db:"prev_price"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
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
		val.Location(),
	)
}
