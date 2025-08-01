package ta

import (
	"fmt"
	"math"
)

type TaVolatilityRatio struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateVolatilityRatio(klineData KlineDatas, shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	if len(klineData) < longPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	slices := preallocateSlices(length, 2)
	trueRange, ratio := slices[0], slices[1]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	for i := longPeriod; i < length; i++ {
		var shortTR float64
		for j := i - shortPeriod + 1; j <= i; j++ {
			shortTR += trueRange[j]
		}
		shortTR /= float64(shortPeriod)

		var longTR float64
		for j := i - longPeriod + 1; j <= i; j++ {
			longTR += trueRange[j]
		}
		longTR /= float64(longPeriod)

		if longTR != 0 {
			ratio[i] = shortTR / longTR
		} else {
			ratio[i] = 1.0
		}
	}

	return &TaVolatilityRatio{
		Values: ratio,
		Period: longPeriod,
	}, nil
}

func (k KlineDatas) VolatilityRatio(shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	return CalculateVolatilityRatio(k, shortPeriod, longPeriod)
}

func (vr *TaVolatilityRatio) Value() float64 {
	if len(vr.Values) == 0 {
		return 0
	}
	return vr.Values[len(vr.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
