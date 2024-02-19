package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const Namespace = "crypto_analyst"

var (
	Errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "errors",
		Help:      "The total errors",
	}, []string{"action", "source"})

	CoefficientDuration = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "coefficient_duration",
		Help:      "The total duration coefficient loaded in ms",
	}, []string{"metric"})

	PopularSymbols = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "popular_symbols",
		Help:      "Count popular symbols",
	})
	ChangePriceCalculateDuration = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "change_price_calculate_duration",
		Help:      "The total duration change price calculation in ms",
	})
	SavePriceDuration = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "save_price_duration",
		Help:      "The total duration save prices in ms",
	})
	SavePrices = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "save_price",
		Help:      "The total save prices",
	})
	SaveSnapshot = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "save_snapshot",
		Help:      "The total save snapshot",
	})

	SaveNewSymbolDuration = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "save_new_symbol_duration",
		Help:      "The total duration save new symbol in ms",
	})
	SaveNewSymbol = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "save_new_symbol",
		Help:      "The total save new symbol",
	})
)
