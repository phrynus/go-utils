package ta

import (
	"fmt"
	"math"
)

type TaSuperTrend struct {
	Upper      []float64 `json:"upper"`
	Lower      []float64 `json:"lower"`
	Trend      []bool    `json:"trend"`
	Period     int       `json:"period"`
	Multiplier float64   `json:"multiplier"`
}

func CalculateSuperTrend(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrend, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := klineData.ATR(period)
	if err != nil {
		return nil, err
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	upperBand, lowerBand := slices[0], slices[1]
	trend := make([]bool, length)

	for i := period; i < length; i++ {
		midpoint := (klineData[i].High + klineData[i].Low) / 2
		atrValue := atr.Values[i]
		upperBand[i] = midpoint + multiplier*atrValue
		lowerBand[i] = midpoint - multiplier*atrValue
	}

	trend[period] = klineData[period].Close > lowerBand[period]

	for i := period + 1; i < length; i++ {
		if trend[i-1] {
			if klineData[i].Close < lowerBand[i] {
				trend[i] = false
				upperBand[i] = upperBand[i-1]
			} else {
				trend[i] = true
				lowerBand[i] = math.Max(lowerBand[i], lowerBand[i-1])
			}
		} else {
			if klineData[i].Close > upperBand[i] {
				trend[i] = true
				lowerBand[i] = lowerBand[i-1]
			} else {
				trend[i] = false
				upperBand[i] = math.Min(upperBand[i], upperBand[i-1])
			}
		}
	}

	return &TaSuperTrend{
		Upper:      upperBand,
		Lower:      lowerBand,
		Trend:      trend,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

func (k *KlineDatas) SuperTrend(period int, multiplier float64) (*TaSuperTrend, error) {
	return CalculateSuperTrend(*k, period, multiplier)
}

func (t *TaSuperTrend) Value() (upper, lower float64, isUpTrend bool) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
