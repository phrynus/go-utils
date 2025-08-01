package ta

import (
	"fmt"
	"math"
)

type TaRSI struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
	Gains  []float64 `json:"gains"`
	Losses []float64 `json:"losses"`
}

func CalculateRSI(prices []float64, period int) (*TaRSI, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	rsi, gains, losses := slices[0], slices[1], slices[2]

	for i := 1; i < length; i++ {
		change := prices[i] - prices[i-1]
		gains[i] = math.Max(0, change)
		losses[i] = math.Max(0, -change)
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	for i := period; i < length; i++ {
		if i > period {
			avgGain = (avgGain*(float64(period)-1) + gains[i]) / float64(period)
			avgLoss = (avgLoss*(float64(period)-1) + losses[i]) / float64(period)
		}

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return &TaRSI{
		Values: rsi,
		Period: period,
		Gains:  gains,
		Losses: losses,
	}, nil
}

func (k *KlineDatas) RSI(period int, source string) (*TaRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRSI(prices, period)
}

func (t *TaRSI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
