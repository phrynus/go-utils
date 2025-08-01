package ta

import (
	"fmt"
)

// TaRMA 相对移动平均线(Relative Moving Average)指标结果结构体
// 说明：
//
//	存储RMA计算结果及相关参数，提供多种基于RMA的技术分析方法
//
// 字段：
//   - Values: RMA计算结果数组 ([]float64)
//   - Period: 计算周期 (int)
type TaRMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateRMA 计算相对移动平均线(RMA)
// 参数：
//   - prices: 价格数据数组 ([]float64)
//   - period: 计算周期 (int)
//
// 返回值：
//   - *TaRMA: RMA计算结果对象
//   - error: 错误信息，若输入数据不足则返回错误
//
// 说明/注意事项：
//
//	RMA是一种平滑的移动平均线，与EMA类似但计算方式略有不同
//	计算公式：RMA[i] = alpha * price[i] + (1-alpha) * RMA[i-1]，其中alpha=1/period
//	要求输入价格数据长度不小于计算周期
//
// 示例：
//
//	prices := []float64{10.0, 11.0, 12.0, 13.0, 14.0}
//	rma, err := CalculateRMA(prices, 3)
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

// RMA 从K线数据中提取指定源数据并计算RMA
// 参数：
//   - period: 计算周期 (int)
//   - source: 数据源字段名 (string)
//
// 返回值：
//   - *TaRMA: RMA计算结果对象
//   - error: 错误信息
//
// 说明/注意事项：
//
//	该方法是KlineDatas结构体的成员方法，用于便捷地从K线数据计算RMA
//	支持的数据源字段由ExtractSlice方法决定
func (k *KlineDatas) RMA(period int, source string) (*TaRMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRMA(prices, period)
}

// Value 获取RMA的最新值
// 返回值：
//   - float64: RMA数组中的最后一个值
func (t *TaRMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
