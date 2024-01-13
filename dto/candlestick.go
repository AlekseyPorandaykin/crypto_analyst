package dto

import "time"

type Candlestick struct {
	Symbol       string    `json:"symbol" db:"symbol"`
	Exchange     string    `json:"exchange" db:"exchange"`
	OpenTime     time.Time `json:"open_time" db:"open_time"`
	CloseTime    time.Time `json:"close_time" db:"close_time"`
	OpenPrice    float64   `json:"open_price" db:"open_price"`
	HighPrice    float64   `json:"high_price" db:"high_price"`
	LowPrice     float64   `json:"low_price" db:"low_price"`
	ClosePrice   float64   `json:"close_price" db:"close_price"`
	Volume       float64   `json:"volume" db:"volume"`
	NumberTrades int       `json:"number_trades" db:"number_trades"`
	Interval     string    `json:"candle_interval" db:"candle_interval"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
