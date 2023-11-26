package metric

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "errors",
		Help:      "The total errors",
	}, []string{"action", "source"})

	CoefficientDuration = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "coefficient_duration",
		Help:      "The total duration coefficient loaded in ms",
	}, []string{"metric"})

	PopularSymbols = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "popular_symbols",
		Help:      "Count popular symbols",
	})
	ChangePriceCalculateDuration = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "change_price_calculate_duration",
		Help:      "The total duration change price calculation in ms",
	})
	SavePriceDuration = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "save_price_duration",
		Help:      "The total duration save prices in ms",
	})
	SavePrices = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "crypto_analyst",
		Name:      "save_price",
		Help:      "The total save prices",
	})
)
