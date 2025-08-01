package ta

import (
	"fmt"
)

type TaSMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateSMA(prices []float64, period int) (*TaSMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	sma := slices[0]

	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	sma[period-1] = sum / float64(period)

	for i := period; i < length; i++ {
		sum += prices[i] - prices[i-period]
		sma[i] = sum / float64(period)
	}

	return &TaSMA{
		Values: sma,
		Period: period,
	}, nil
}

func (k *KlineDatas) SMA(period int, source string) (*TaSMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateSMA(prices, period)
}

func (t *TaSMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
