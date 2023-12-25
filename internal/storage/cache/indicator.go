package cache

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

// TODO каждый индикатор храниться в соем кэше, у каждого свой ключ
type Indicator struct {
	ma8 *lru.Cache[string, float64]
}

func NewIndicator() (*Indicator, error) {
	l, err := lru.New[string, float64](2 << 10)
	if err != nil {
		return nil, err
	}
	return &Indicator{
		ma8: l,
	}, nil
}

func (i *Indicator) Save() {
	i.ma8.Add("S", 12)
}
