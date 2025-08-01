package ta

import (
	"fmt"
)

// TaEMA 表示指数移动平均线(Exponential Moving Average)的计算结果
// 说明：
//
//	EMA是一种赋予近期数据更高权重的移动平均线：
//	1. 比简单移动平均线(SMA)对价格变化更敏感
//	2. 能更快反映价格趋势的变化
//	3. 广泛用于其他技术指标的计算（如MACD）
//	特点：
//	- 对最新数据的反应更快
//	- 减少了滞后性
//	- 平滑度介于SMA和WMA之间
//	- 适合中短期趋势跟踪
type TaEMA struct {
	Values []float64 `json:"values"` // EMA值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateEMA 计算指数移动平均线
// 说明：
//
//	计算步骤：
//	1. 计算第一个EMA值（使用SMA作为起始值）：
//	   首个EMA = 前N日价格的算术平均值
//	2. 计算后续EMA值：
//	   EMA = 当日价格 * K + 前一日EMA * (1-K)
//	   其中 K = 2/(N+1)，N为周期数
//	应用场景：
//	- 判断价格趋势
//	- 支撑位和阻力位分析
//	- 用于构建其他技术指标
//
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，常用值：
//   - 短期：5、10、12
//   - 中期：20、26
//   - 长期：50、100、200
//
// 返回值：
//   - *TaEMA: 包含EMA计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	ema, err := CalculateEMA(prices, 20)
func CalculateEMA(prices []float64, period int) (*TaEMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	result := slices[0]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	multiplier := 2.0 / float64(period+1)
	oneMinusMultiplier := 1.0 - multiplier

	for i := period; i < length; i++ {
		result[i] = prices[i]*multiplier + result[i-1]*oneMinusMultiplier
	}

	return &TaEMA{
		Values: result,
		Period: period,
	}, nil
}

// EMA 为K线数据计算指数移动平均线
// 说明：
//
//	对指定价格类型计算EMA指标
//
// 参数：
//   - period: 计算周期
//   - source: 价格类型，支持"open"、"high"、"low"、"close"等
//
// 返回值：
//   - *TaEMA: 包含EMA计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) EMA(period int, source string) (*TaEMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateEMA(prices, period)
}

// Value 获取最新的EMA值
// 说明：
//
//	返回最新的EMA值
//	使用建议：
//	- 价格上穿EMA视为买入信号
//	- 价格下穿EMA视为卖出信号
//	- 多条EMA的交叉可产生交易信号
//	- 短期EMA上穿长期EMA为金叉
//	- 短期EMA下穿长期EMA为死叉
//
// 返回值：
//   - float64: 最新的EMA值
func (t *TaEMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
