package ta

import (
	"fmt"
)

type TaSuperTrendPivotHl2 struct {
	Values     []float64 `json:"values"`
	Direction  []int     `json:"direction"`
	UpperBand  []float64 `json:"upper_band"`
	LowerBand  []float64 `json:"lower_band"`
	Period     int       `json:"period"`
	Multiplier float64   `json:"multiplier"`
}

func CalculateSuperTrendPivotHl2(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	length := len(klineData)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := CalculateATR(klineData, period)
	if err != nil {
		return nil, err
	}

	slices := preallocateSlices(length, 4)
	values, direction, upperBand, lowerBand := slices[0], make([]int, length), slices[2], slices[3]

	for i := 0; i < length; i++ {

		hl2 := (klineData[i].High + klineData[i].Low) / 2

		if i < period {

			upperBand[i] = hl2 + multiplier*atr.Values[i]
			lowerBand[i] = hl2 - multiplier*atr.Values[i]
			direction[i] = 0
			values[i] = hl2
			continue
		}

		basicUpperBand := hl2 + multiplier*atr.Values[i]
		basicLowerBand := hl2 - multiplier*atr.Values[i]

		if basicLowerBand > lowerBand[i-1] || klineData[i-1].Close < lowerBand[i-1] {
			lowerBand[i] = basicLowerBand
		} else {
			lowerBand[i] = lowerBand[i-1]
		}

		if basicUpperBand < upperBand[i-1] || klineData[i-1].Close > upperBand[i-1] {
			upperBand[i] = basicUpperBand
		} else {
			upperBand[i] = upperBand[i-1]
		}

		if direction[i-1] <= 0 {
			if klineData[i].Close > upperBand[i] {
				direction[i] = 1
			} else {
				direction[i] = -1
			}
		} else {
			if klineData[i].Close < lowerBand[i] {
				direction[i] = -1
			} else {
				direction[i] = 1
			}
		}

		if direction[i] == 1 {
			values[i] = lowerBand[i]
		} else {
			values[i] = upperBand[i]
		}
	}

	return &TaSuperTrendPivotHl2{
		Values:     values,
		Direction:  direction,
		UpperBand:  upperBand,
		LowerBand:  lowerBand,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

func (k *KlineDatas) SuperTrendPivotHl2(period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	return CalculateSuperTrendPivotHl2(*k, period, multiplier)
}

func (t *TaSuperTrendPivotHl2) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
