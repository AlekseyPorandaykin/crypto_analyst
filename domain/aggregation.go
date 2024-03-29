package domain

import (
	"fmt"
	"time"
)

type MetricAggregationPrice string

const (
	ChangeCoefficientOnHour MetricAggregationPrice = "ChangeCoefficientOnHour"
	ChangeCoefficientOnDay  MetricAggregationPrice = "ChangeCoefficientOnDay"
	ChangeCoefficientOnWeek MetricAggregationPrice = "ChangeCoefficientOnWeek"

	IndicatorChangeOnHour MetricAggregationPrice = "IndicatorChangeOnHour"
	IndicatorChangeOnDay  MetricAggregationPrice = "IndicatorChangeOnDay"
	IndicatorChangeOnWeek MetricAggregationPrice = "IndicatorChangeOnWeek"
)

type PriceAggregation struct {
	Symbol    string                 `json:"symbol" db:"symbol"`
	Exchange  string                 `json:"exchange" db:"exchange"`
	Metric    MetricAggregationPrice `json:"metric" db:"metric"`
	Key       string                 `json:"key" db:"key"`
	Value     string                 `json:"value" db:"value"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

func (pa PriceAggregation) UniqKey() string {
	return fmt.Sprintf("%s-%s-%s-%s", pa.Symbol, pa.Exchange, pa.Metric, pa.Key)
}
