package ta

import (
	"fmt"
	"math"
)

type TaSuperTrendPivot struct {
	Upper       []float64 `json:"upper"`
	Lower       []float64 `json:"lower"`
	Trend       []int     `json:"trend"`
	PivotPeriod int       `json:"pivot_period"`
	Factor      float64   `json:"factor"`
	AtrPeriod   int       `json:"atr_period"`
}

func FindPivotHighPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].High > klineData[index].High {
			return math.NaN()
		}
	}
	return klineData[index].High
}

func FindPivotLowPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].Low < klineData[index].Low {
			return math.NaN()
		}
	}
	return klineData[index].Low
}

func CalculateSuperTrendPivot(klineData KlineDatas, pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {

	dataLen := len(klineData)
	if dataLen < pivotPeriod || dataLen < atrPeriod {
		return nil, fmt.Errorf("计算数据不足: 数据长度%d, 需要轴点周期%d和ATR周期%d", dataLen, pivotPeriod, atrPeriod)
	}

	trendUp := make([]float64, dataLen)
	trendDown := make([]float64, dataLen)
	trend := make([]int, dataLen)

	atr, err := klineData.ATR(atrPeriod)
	if err != nil {
		return nil, fmt.Errorf("计算ATR失败: %v", err)
	}

	var center float64
	var centerCount int

	for i := pivotPeriod; i < dataLen; i++ {

		pivotHigh := FindPivotHighPoint(klineData, i, pivotPeriod)
		pivotLow := FindPivotLowPoint(klineData, i, pivotPeriod)

		if !math.IsNaN(pivotHigh) || !math.IsNaN(pivotLow) {
			newCenter := 0.0
			if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {

				newCenter = (pivotHigh + pivotLow) / 2
			} else if !math.IsNaN(pivotHigh) {
				newCenter = pivotHigh
			} else {
				newCenter = pivotLow
			}

			if centerCount == 0 {
				center = newCenter
			} else {
				center = (center*2 + newCenter) / 3
			}
			centerCount++
		}

		if centerCount == 0 {
			center = (klineData[i].High + klineData[i].Low) / 2
		}

		band := factor * atr.Values[i]
		upperBand := center + band
		lowerBand := center - band

		if i > 0 {

			if klineData[i-1].Close > trendUp[i-1] {
				trendUp[i] = math.Max(lowerBand, trendUp[i-1])
			} else {
				trendUp[i] = lowerBand
			}

			if klineData[i-1].Close < trendDown[i-1] {
				trendDown[i] = math.Min(upperBand, trendDown[i-1])
			} else {
				trendDown[i] = upperBand
			}

			if klineData[i].Close > trendDown[i-1] {
				trend[i] = 1
			} else if klineData[i].Close < trendUp[i-1] {
				trend[i] = -1
			} else {
				trend[i] = trend[i-1]
			}
		} else {

			trendUp[i] = lowerBand
			trendDown[i] = upperBand
			trend[i] = 0
		}
	}

	return &TaSuperTrendPivot{
		Upper:       trendDown,
		Lower:       trendUp,
		Trend:       trend,
		PivotPeriod: pivotPeriod,
		Factor:      factor,
		AtrPeriod:   atrPeriod,
	}, nil
}

func (k *KlineDatas) SuperTrendPivot(pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {
	return CalculateSuperTrendPivot(*k, pivotPeriod, factor, atrPeriod)
}
func (t *TaSuperTrendPivot) Value() (upper, lower float64, trend int) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
