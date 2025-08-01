package ta

import (
	"fmt"
)

// TaWilliamsR 表示威廉指标(Williams %R)的计算结果
// 说明：
//
//	Williams %R是一个动量指标，用于判断市场是否处于超买或超卖状态：
//	1. 测量收盘价在最近N个周期内的高低价范围中的相对位置
//	2. 取值范围为0到-100
//	3. 与随机指标KD相似，但计算方法略有不同
//	特点：
//	- 0到-20区域表示超买
//	- -80到-100区域表示超卖
//	- 指标与价格的背离可能预示趋势反转
//	- 相比其他超买超卖指标反应更快速
type TaWilliamsR struct {
	Values []float64 `json:"values"` // Williams %R值的时间序列
	Period int       `json:"period"` // 计算周期
}

// CalculateWilliamsR 计算威廉指标
// 说明：
//
//	计算公式：
//	W%R = (最高价 - 收盘价) / (最高价 - 最低价) * -100
//	计算步骤：
//	1. 找出周期内的最高价和最低价
//	2. 计算当前收盘价在这个范围内的相对位置
//	3. 将结果乘以-100得到最终值
//	使用场景：
//	- 判断市场超买超卖状态
//	- 寻找潜在的趋势反转点
//	- 与其他指标配合确认交易信号
//
// 参数：
//   - high: 最高价序列
//   - low: 最低价序列
//   - close: 收盘价序列
//   - period: 计算周期，通常为14
//
// 返回值：
//   - *TaWilliamsR: 包含计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	wr, err := CalculateWilliamsR(high, low, close, 14)
func CalculateWilliamsR(high, low, close []float64, period int) (*TaWilliamsR, error) {
	if len(high) < period || len(low) < period || len(close) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 1)
	wr := slices[0]

	for i := period - 1; i < length; i++ {

		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < period; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		if highestHigh != lowestLow {
			wr[i] = ((highestHigh - close[i]) / (highestHigh - lowestLow)) * -100
		} else {
			wr[i] = -50
		}
	}

	return &TaWilliamsR{
		Values: wr,
		Period: period,
	}, nil
}

// WilliamsR 为K线数据计算威廉指标
// 说明：
//
//	从K线数据中提取价格序列并计算Williams %R
//
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaWilliamsR: 包含计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) WilliamsR(period int) (*TaWilliamsR, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	return CalculateWilliamsR(high, low, close, period)
}

// WilliamsR_ 快速计算最新的威廉指标值
// 说明：
//
//	计算最近period*14个K线的Williams %R值
//	这是一个简化版本，主要用于快速获取最新值
//
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - float64: 最新的Williams %R值，计算失败返回0
func (k *KlineDatas) WilliamsR_(period int) float64 {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	wr, err := _k.WilliamsR(period)
	if err != nil {
		return 0
	}
	return wr.Value()
}

// Value 获取最新的威廉指标值
// 说明：
//
//	返回Williams %R的最新值
//	使用建议：
//	- 当指标从超买区域(-20以上)向下穿越时，考虑卖出
//	- 当指标从超卖区域(-80以下)向上穿越时，考虑买入
//	- 与价格趋势一起使用，不要单独作为交易信号
//	- 可以通过调整周期来改变指标的敏感度
//
// 返回值：
//   - float64: 最新的Williams %R值
func (t *TaWilliamsR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
