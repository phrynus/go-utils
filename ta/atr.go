package ta

import (
	"fmt"
	"math"
)

// TaATR 平均真实波动范围（ATR）计算结果结构体
// 说明：
//
//	用于存储 ATR 计算过程中的相关数据，包括 ATR 值、计算周期和真实波动范围
//
// 字段：
//   - Values: 每个时间点的 ATR 值切片
//   - Period: ATR 计算所使用的周期
//   - TrueRange: 每个时间点的真实波动范围切片
type TaATR struct {
	Values    []float64 `json:"values"`
	Period    int       `json:"period"`
	TrueRange []float64 `json:"true_range"`
}

// CalculateATR 计算给定 K 线数据的平均真实波动范围（ATR）
// 参数：
//   - klineData: K 线数据切片，包含每个时间点的高、低、收盘价等信息
//   - period: ATR 计算所使用的周期
//
// 返回值：
//   - *TaATR: 包含 ATR 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误，若计算数据不足则返回错误
//
// 说明/注意事项：
//
//	计算 ATR 时，需要至少 period 个 K 线数据。
//	真实波动范围（TR）的计算基于当前时间点的最高价、最低价和上一个时间点的收盘价。
//	初始 ATR 值为前 period 个 TR 的平均值，后续 ATR 值使用平滑公式计算。
//
// 示例：
//
//	klineData := ...
//	atr, err := CalculateATR(klineData, 14)
//	if err != nil {
//	    log.Fatal(err)
//	}
func CalculateATR(klineData KlineDatas, period int) (*TaATR, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	trueRange, atr := slices[0], slices[1]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	var sumTR float64
	for i := 1; i <= period; i++ {
		sumTR += trueRange[i]
	}
	atr[period] = sumTR / float64(period)

	for i := period + 1; i < length; i++ {
		atr[i] = (atr[i-1]*(float64(period)-1) + trueRange[i]) / float64(period)
	}

	return &TaATR{
		Values:    atr,
		Period:    period,
		TrueRange: trueRange,
	}, nil
}

// ATR 计算 K 线数据的平均真实波动范围（ATR）
// 参数：
//   - period: ATR 计算所使用的周期
//
// 返回值：
//   - *TaATR: 包含 ATR 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误，若计算数据不足则返回错误
//
// 说明/注意事项：
//
//	该方法是对 CalculateATR 函数的封装，直接作用于 KlineDatas 结构体实例。
//
// 示例：
//
//	klineData := ...
//	atr, err := klineData.ATR(14)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (k *KlineDatas) ATR(period int) (*TaATR, error) {
	return CalculateATR(*k, period)
}

// Value 返回 TaATR 结构体中最新的 ATR 值
// 返回值：
//   - float64: 最新的 ATR 值
//
// 说明/注意事项：
//
//	若 Values 切片为空，可能会引发越界错误，使用前需确保数据有效。
//
// 示例：
//
//	atr := ...
//	latestATR := atr.Value()
func (t *TaATR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// Percent 计算最新的 ATR 值相对于当前价格的百分比
// 参数：
//   - currentPrice: 当前价格
//
// 返回值：
//   - float64: ATR 占当前价格的百分比值（以小数形式返回，例如 0.05 表示 5%）
//
// 说明/注意事项：
//
//	该方法将最新的 ATR 值除以当前价格，得到百分比。
//	返回值为小数形式，需要乘以 100 才是实际的百分比数值。
//
// 示例：
//
//	atr := ...
//	percent := atr.Percent(100.0) // 如果 ATR 值为 5，则返回 0.05，表示 5%
func (t *TaATR) Percent(currentPrice float64) float64 {
	if currentPrice <= 0 {
		return 0
	}
	return t.Value() / currentPrice
}
