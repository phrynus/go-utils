package ta

import (
	"fmt"
)

// TaEMA 指数移动平均线（EMA）计算结果的结构体
// 说明：
//
//	该结构体用于存储 EMA 计算的结果，包含计算得到的 EMA 值数组和计算使用的周期数。
//
// 字段：
//   - Values: 存储 EMA 计算结果的浮点数数组 (float64 类型)
//   - Period: 计算 EMA 时使用的周期数 (int 类型)
type TaEMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateEMA 计算指数移动平均线（EMA）
// 参数：
//   - prices: 输入的价格数据数组 (float64 类型)
//   - period: 计算 EMA 时使用的周期数 (int 类型)
//
// 返回值：
//   - *TaEMA: 存储 EMA 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	当输入的价格数据长度小于指定的周期数时，会返回错误。
//	该函数使用标准的 EMA 计算公式进行计算。
//
// 示例：
//
//	prices := []float64{10, 11, 12, 13, 14}
//	period := 3
//	ema, err := CalculateEMA(prices, period)
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

// EMA 从 KlineDatas 中提取数据并计算 EMA
// 参数：
//   - period: 计算 EMA 时使用的周期数 (int 类型)
//   - source: 提取数据的源字段名 (string 类型)
//
// 返回值：
//   - *TaEMA: 存储 EMA 计算结果的结构体指针
//   - error: 提取数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法依赖 KlineDatas 的 ExtractSlice 方法提取数据。
//
// 示例：
//
//	klineData := KlineDatas{...}
//	period := 3
//	source := "close"
//	ema, err := klineData.EMA(period, source)
func (k *KlineDatas) EMA(period int, source string) (*TaEMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateEMA(prices, period)
}

// Value 获取 TaEMA 结构体中最后一个 EMA 值
// 返回值：
//   - float64: TaEMA 结构体中最后一个 EMA 值
//
// 说明/注意事项：
//
//	若 TaEMA 结构体的 Values 数组为空，可能会导致越界错误。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11, 12}, Period: 3}
//	value := ema.Value()
func (t *TaEMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
