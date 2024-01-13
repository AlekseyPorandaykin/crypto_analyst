package domain

import "context"

const (
	USDT = "USDT"
	BTC  = "BTC"
)

const (
	BTCUSDT = "BTCUSDT"
	BTCUSDC = "BTCUSDC"
	ETHUSDT = "ETHUSDT"
	LTCUSDT = "LTCUSDT"
	ETHUSDC = "ETHUSDC"
	SOLUSDT = "SOLUSDT"
)

var SubscribeSymbols = []string{BTCUSDT, ETHUSDT, LTCUSDT, SOLUSDT}

var MainSymbolPairs = map[string]bool{
	BTCUSDT: true,
	ETHUSDT: true,
}

var PopularSymbols = map[string]int{
	BTCUSDT: 100,
	ETHUSDT: 99,
	BTCUSDC: 98,
	LTCUSDT: 97,
	ETHUSDC: 96,
	SOLUSDT: 95,
}

type SymbolStorage interface {
	List(ctx context.Context) ([]string, error)
}
