package loader

import "github.com/AlekseyPorandaykin/crypto_analyst/domain"

type Batch struct {
	prices []*domain.SymbolPrice
}

func NewBatch(size int) Batch {
	return Batch{prices: make([]*domain.SymbolPrice, 0, size)}
}

func (b *Batch) Append(price *domain.SymbolPrice) {
	b.prices = append(b.prices, price)
}

func (b *Batch) Prices() []*domain.SymbolPrice {
	return b.prices
}
