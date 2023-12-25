package domain

type IndicatorSnapshot struct {
	MA8  float64
	MA40 float64

	EMA float64
	WMA float64
	SMA float64

	RingLow  float64
	RingHigh float64

	PinBar float64
}

type SymbolsIndicatorList map[string][]IndicatorSnapshot
