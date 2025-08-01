package ta

import (
	"fmt"
)

// TaRMA 表示平滑移动平均线(Running Moving Average)的计算结果
// 说明：
//
//	RMA是一种特殊的指数移动平均线：
//	1. 使用Wilder平滑法计算
//	2. 相比EMA具有更强的平滑效果
//	3. 常用于RSI等技术指标的计算
//	特点：
//	- 平滑度高于简单移动平均线
//	- 对异常值的敏感度较低
//	- 计算过程中不会丢失历史信息
//	- 适合用于波动较大的市场
type TaRMA struct {
	Values []float64 `json:"values"` // RMA值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateRMA 计算平滑移动平均线
// 说明：
//
//	计算步骤：
//	1. 设定平滑系数 alpha = 1/period
//	2. 第一个值直接使用原始数据
//	3. 后续值使用公式：
//	   RMA = alpha * 当前值 + (1 - alpha) * 前一期RMA
//	应用场景：
//	- 用于计算RSI等技术指标
//	- 平滑价格波动
//	- 识别中长期趋势
//	- 过滤市场噪音
//
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，常用值：
//   - RSI计算时通常为14
//   - 趋势跟踪时可选20-60
//
// 返回值：
//   - *TaRMA: 包含RMA计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	rma, err := CalculateRMA(prices, 14)
func CalculateRMA(prices []float64, period int) (*TaRMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	rma := slices[0]

	alpha := 1.0 / float64(period)
	rma[0] = prices[0]

	for i := 1; i < length; i++ {
		rma[i] = alpha*prices[i] + (1-alpha)*rma[i-1]
	}

	return &TaRMA{
		Values: rma,
		Period: period,
	}, nil
}

// RMA 为K线数据计算平滑移动平均线
// 说明：
//
//	对指定价格类型计算RMA指标
//
// 参数：
//   - period: 计算周期
//   - source: 价格类型，支持"open"、"high"、"low"、"close"等
//
// 返回值：
//   - *TaRMA: 包含RMA计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) RMA(period int, source string) (*TaRMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRMA(prices, period)
}

// Value 获取最新的RMA值
// 说明：
//
//	返回最新的RMA值
//	使用建议：
//	- 可作为动态支撑位和阻力位
//	- 价格上穿RMA视为上升趋势确立
//	- 价格下穿RMA视为下降趋势确立
//	- 与其他指标配合使用效果更好
//
// 返回值：
//   - float64: 最新的RMA值
func (t *TaRMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
