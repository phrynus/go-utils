package ta

import (
	"fmt"
)

type TaStochRSI struct {
	K           []float64 `json:"k"`
	D           []float64 `json:"d"`
	RsiPeriod   int       `json:"rsi_period"`
	StochPeriod int       `json:"stoch_period"`
	KPeriod     int       `json:"k_period"`
	DPeriod     int       `json:"d_period"`
}

func CalculateStochRSI(prices []float64, rsiPeriod, stochPeriod, kPeriod, dPeriod int) (*TaStochRSI, error) {
	if len(prices) < rsiPeriod+stochPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	rsi, err := CalculateRSI(prices, rsiPeriod)
	if err != nil {
		return nil, err
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	stochRsi, k, d := slices[0], slices[1], slices[2]

	for i := stochPeriod - 1; i < length; i++ {

		var highestRsi, lowestRsi = rsi.Values[i], rsi.Values[i]
		for j := 0; j < stochPeriod; j++ {
			idx := i - j
			if rsi.Values[idx] > highestRsi {
				highestRsi = rsi.Values[idx]
			}
			if rsi.Values[idx] < lowestRsi {
				lowestRsi = rsi.Values[idx]
			}
		}

		if highestRsi != lowestRsi {
			stochRsi[i] = (rsi.Values[i] - lowestRsi) / (highestRsi - lowestRsi) * 100
		} else {
			stochRsi[i] = 50
		}
	}

	var sumK float64
	for i := 0; i < kPeriod && i < length; i++ {
		sumK += stochRsi[i]
	}
	k[kPeriod-1] = sumK / float64(kPeriod)

	for i := kPeriod; i < length; i++ {
		sumK = sumK - stochRsi[i-kPeriod] + stochRsi[i]
		k[i] = sumK / float64(kPeriod)
	}

	var sumD float64
	for i := 0; i < dPeriod && i < length; i++ {
		sumD += k[i]
	}
	d[dPeriod-1] = sumD / float64(dPeriod)

	for i := dPeriod; i < length; i++ {
		sumD = sumD - k[i-dPeriod] + k[i]
		d[i] = sumD / float64(dPeriod)
	}

	return &TaStochRSI{
		K:           k,
		D:           d,
		RsiPeriod:   rsiPeriod,
		StochPeriod: stochPeriod,
		KPeriod:     kPeriod,
		DPeriod:     dPeriod,
	}, nil
}

func (k *KlineDatas) StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (*TaStochRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateStochRSI(prices, rsiPeriod, stochPeriod, kPeriod, dPeriod)
}

func (t *TaStochRSI) Value() (kValue, dValue float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
