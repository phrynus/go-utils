package ta

import (
	"fmt"
)

// TaStochRSI 表示随机相对强弱指标(Stochastic RSI)的计算结果
// 说明：
//
//	StochRSI是一个复合指标，结合了RSI和随机指标的特点：
//	1. 首先计算RSI
//	2. 然后对RSI值应用随机指标的计算方法
//	3. 最后计算快线K值和慢线D值
//	特点：
//	- 取值范围0-100
//	- 相比传统RSI更敏感
//	- 可以更早发现超买超卖区域
//	- 80以上为超买区，20以下为超卖区
type TaStochRSI struct {
	K           []float64 `json:"k"`            // K值（快线）的时间序列
	D           []float64 `json:"d"`            // D值（慢线）的时间序列
	RsiPeriod   int       `json:"rsi_period"`   // RSI计算周期
	StochPeriod int       `json:"stoch_period"` // 随机指标计算周期
	KPeriod     int       `json:"k_period"`     // K值平滑周期
	DPeriod     int       `json:"d_period"`     // D值平滑周期
}

// CalculateStochRSI 计算给定价格序列的随机RSI指标
// 说明：
//
//	计算步骤：
//	1. 计算指定周期的RSI值
//	2. 计算RSI的随机值：(当前RSI - 最低RSI) / (最高RSI - 最低RSI) * 100
//	3. 对随机值进行平滑处理得到K值
//	4. 对K值进行平滑处理得到D值
//
// 参数：
//   - prices: 价格时间序列
//   - rsiPeriod: RSI计算周期，通常为14
//   - stochPeriod: 随机值计算周期，通常为14
//   - kPeriod: K值平滑周期，通常为3
//   - dPeriod: D值平滑周期，通常为3
//
// 返回值：
//   - *TaStochRSI: 包含StochRSI计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	prices := []float64{10, 10.5, 10.3, 10.2, 10.4, 10.3, 10.7}
//	stochRsi, err := CalculateStochRSI(prices, 14, 14, 3, 3)
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

// StochRSI 为K线数据计算随机RSI指标
// 说明：
//
//	基于K线数据中指定的价格类型（如收盘价、开盘价等）计算StochRSI指标
//
// 参数：
//   - rsiPeriod: RSI计算周期
//   - stochPeriod: 随机值计算周期
//   - kPeriod: K值平滑周期
//   - dPeriod: D值平滑周期
//   - source: 价格数据来源，可以是"close"、"open"、"high"、"low"等
//
// 返回值：
//   - *TaStochRSI: 包含StochRSI计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (*TaStochRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateStochRSI(prices, rsiPeriod, stochPeriod, kPeriod, dPeriod)
}

// StochRSI_ 获取最新的StochRSI的K值和D值
// 参数：
//   - rsiPeriod: RSI计算周期
//   - stochPeriod: 随机值计算周期
//   - kPeriod: K值平滑周期
//   - dPeriod: D值平滑周期
//   - source: 价格数据来源，可以是"close"、"open"、"high"、"low"等
//
// 返回值：
//   - float64: 最新的K值
//   - float64: 最新的D值
func (k *KlineDatas) StochRSI_(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (float64, float64) {
	stochRsi, err := k.StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod, source)
	if err != nil {
		return 0, 0
	}
	return stochRsi.Value()
}

// Value 获取最新的StochRSI的K值和D值
// 说明：
//
//	返回StochRSI的最新K值（快线）和D值（慢线）
//	交易信号：
//	- K线从下方穿过D线是买入信号
//	- K线从上方穿过D线是卖出信号
//	- K和D同时在超买区（80以上）或超卖区（20以下）时信号更强
//
// 返回值：
//   - kValue: 最新的K值
//   - dValue: 最新的D值
func (t *TaStochRSI) Value() (kValue, dValue float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
