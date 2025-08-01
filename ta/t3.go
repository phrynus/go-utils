package ta

import (
	"fmt"
)

type TaT3 struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
	VFact  float64   `json:"vfact"`
}

func CalculateT3(prices []float64, period int, vfact float64) (*TaT3, error) {
	if len(prices) < period*6 {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 7)
	ema1, ema2, ema3, ema4, ema5, ema6, t3 := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5], slices[6]

	k := 2.0 / float64(period+1)
	ema1[0] = prices[0]
	for i := 1; i < length; i++ {
		ema1[i] = prices[i]*k + ema1[i-1]*(1-k)
	}

	for i := 1; i < length; i++ {
		ema2[i] = ema1[i]*k + ema2[i-1]*(1-k)
		ema3[i] = ema2[i]*k + ema3[i-1]*(1-k)
		ema4[i] = ema3[i]*k + ema4[i-1]*(1-k)
		ema5[i] = ema4[i]*k + ema5[i-1]*(1-k)
		ema6[i] = ema5[i]*k + ema6[i-1]*(1-k)
	}

	b := vfact
	c1 := -b * b * b
	c2 := 3*b*b + 3*b*b*b
	c3 := -6*b*b - 3*b - 3*b*b*b
	c4 := 1 + 3*b + b*b*b + 3*b*b

	for i := period * 6; i < length; i++ {
		t3[i] = c1*ema6[i] + c2*ema5[i] + c3*ema4[i] + c4*ema3[i]
	}

	return &TaT3{
		Values: t3,
		Period: period,
		VFact:  vfact,
	}, nil
}

func (k *KlineDatas) T3(period int, vfact float64, source string) (*TaT3, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateT3(prices, period, vfact)
}

func (t *TaT3) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
