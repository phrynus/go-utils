package ta

import (
	"fmt"
	"math"
)

// TaRSI 表示相对强弱指标(RSI)的计算结果
// 说明：
//
//	RSI是一个动量指标，用于衡量价格变动的强度。它通过计算一定时期内价格上涨和下跌的相对强度来判断市场的超买或超卖状态。
//	RSI的取值范围是0-100：
//	- RSI > 70 通常被认为是超买状态
//	- RSI < 30 通常被认为是超卖状态
type TaRSI struct {
	Values []float64 `json:"values"` // RSI值的时间序列
	Period int       `json:"period"` // 计算RSI使用的周期
	Gains  []float64 `json:"gains"`  // 价格上涨幅度的时间序列
	Losses []float64 `json:"losses"` // 价格下跌幅度的时间序列
}

// CalculateRSI 计算给定价格序列的RSI指标
// 说明：
//
//	使用Wilder的RSI计算方法，包括以下步骤：
//	1. 计算每日价格变动的涨跌幅
//	2. 计算初始平均涨幅和跌幅
//	3. 使用平滑移动平均计算后续的平均涨幅和跌幅
//	4. 根据公式 RSI = 100 - (100 / (1 + RS)) 计算RSI值，其中RS = 平均涨幅/平均跌幅
//
// 参数：
//   - prices: 价格时间序列
//   - period: RSI计算周期
//
// 返回值：
//   - *TaRSI: 包含RSI计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	prices := []float64{10, 10.5, 10.3, 10.2, 10.4, 10.3, 10.7}
//	rsi, err := CalculateRSI(prices, 5)
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

// RSI 为K线数据计算RSI指标
// 说明：
//
//	基于K线数据中指定的价格类型（如收盘价、开盘价等）计算RSI指标
//
// 参数：
//   - period: RSI计算周期
//   - source: 价格数据来源，可以是"close"、"open"、"high"、"low"等
//
// 返回值：
//   - *TaRSI: 包含RSI计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) RSI(period int, source string) (*TaRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRSI(prices, period)
}

// Value 获取最新的RSI值
// 说明：
//
//	返回RSI时间序列中的最后一个值，即最新的RSI值
//
// 返回值：
//   - float64: 最新的RSI值
func (t *TaRSI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
