package ta

import (
	"fmt"
)

// TaSMA 表示简单移动平均线(Simple Moving Average)的计算结果
// 说明：
//
//	SMA是最基础的移动平均线指标，用于平滑价格数据，帮助识别价格趋势。
//	计算方法是取特定周期内的价格平均值。
//	特点：
//	- 对所有数据点赋予相同权重
//	- 能有效过滤价格噪音
//	- 滞后性较强，适合确认中长期趋势
type TaSMA struct {
	Values []float64 `json:"values"` // SMA值的时间序列
	Period int       `json:"period"` // 计算SMA使用的周期
}

// CalculateSMA 计算给定价格序列的简单移动平均线
// 说明：
//
//	使用滑动窗口方法计算SMA，步骤如下：
//	1. 计算第一个SMA值（前period个价格的平均值）
//	2. 通过减去窗口最左侧价格并加入最新价格来更新移动窗口
//	3. 持续计算直至处理完所有数据
//
// 参数：
//   - prices: 价格时间序列
//   - period: SMA计算周期，如5日均线period=5
//
// 返回值：
//   - *TaSMA: 包含SMA计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	prices := []float64{10, 11, 12, 13, 14, 15, 16}
//	sma, err := CalculateSMA(prices, 5)
func CalculateSMA(prices []float64, period int) (*TaSMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	sma := slices[0]

	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	sma[period-1] = sum / float64(period)

	for i := period; i < length; i++ {
		sum += prices[i] - prices[i-period]
		sma[i] = sum / float64(period)
	}

	return &TaSMA{
		Values: sma,
		Period: period,
	}, nil
}

// SMA 为K线数据计算简单移动平均线
// 说明：
//
//	基于K线数据中指定的价格类型（如收盘价、开盘价等）计算SMA指标
//
// 参数：
//   - period: SMA计算周期
//   - source: 价格数据来源，可以是"close"、"open"、"high"、"low"等
//
// 返回值：
//   - *TaSMA: 包含SMA计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) SMA(period int, source string) (*TaSMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateSMA(prices, period)
}

// Value 获取最新的SMA值
// 说明：
//
//	返回SMA时间序列中的最后一个值，即最新的均线值
//
// 返回值：
//   - float64: 最新的SMA值
func (t *TaSMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
