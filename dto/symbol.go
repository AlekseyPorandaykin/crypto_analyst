package dto

type ExchangeSymbol struct {
	Symbol   string `db:"symbol"`
	Exchange string `db:"exchange"`
}
