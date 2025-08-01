package ta

import (
	"fmt"
)

type TaWilliamsR struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateWilliamsR(high, low, close []float64, period int) (*TaWilliamsR, error) {
	if len(high) < period || len(low) < period || len(close) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 1)
	wr := slices[0]

	for i := period - 1; i < length; i++ {

		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < period; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		if highestHigh != lowestLow {
			wr[i] = ((highestHigh - close[i]) / (highestHigh - lowestLow)) * -100
		} else {
			wr[i] = -50
		}
	}

	return &TaWilliamsR{
		Values: wr,
		Period: period,
	}, nil
}

func (k *KlineDatas) WilliamsR(period int) (*TaWilliamsR, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	return CalculateWilliamsR(high, low, close, period)
}

func (k *KlineDatas) WilliamsR_(period int) float64 {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	wr, err := _k.WilliamsR(period)
	if err != nil {
		return 0
	}
	return wr.Value()
}

func (t *TaWilliamsR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
