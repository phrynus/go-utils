package ta

import (
	"fmt"
	"math"
)

// TaATR 表示平均真实波幅(Average True Range)的计算结果
// 说明：
//
//	ATR是衡量市场波动性的重要指标，由Welles Wilder开发：
//	1. 不测量价格方向，只测量价格波动幅度
//	2. 考虑了价格跳空的影响
//	3. 可用于判断市场波动状态和设置止损位
//	特点：
//	- ATR值越大，表示市场波动越剧烈
//	- ATR值越小，表示市场波动越平缓
//	- 常用于确定止损距离和仓位大小
type TaATR struct {
	Values    []float64 `json:"values"`     // ATR值序列
	Period    int       `json:"period"`     // 计算周期
	TrueRange []float64 `json:"true_range"` // 真实波幅序列
}

// CalculateATR 计算平均真实波幅
// 说明：
//
//	计算步骤：
//	1. 计算真实波幅(TR)：
//	   TR = max(
//	       当日最高价 - 当日最低价,
//	       |当日最高价 - 前日收盘价|,
//	       |当日最低价 - 前日收盘价|
//	   )
//	2. 计算ATR：
//	   第一个ATR = 前period日TR的简单平均
//	   之后的ATR = (前一日ATR * (period-1) + 当日TR) / period
//
// 参数：
//   - klineData: K线数据
//   - period: 计算周期，通常为14
//
// 返回值：
//   - *TaATR: 包含ATR计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	atr, err := CalculateATR(klineData, 14)
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

// ATR 为K线数据计算平均真实波幅
// 说明：
//
//	对当前K线数据计算ATR指标
//
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaATR: 包含ATR计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) ATR(period int) (*TaATR, error) {
	return CalculateATR(*k, period)
}

// ATR_ 获取最新的ATR值
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - float64: 最新的ATR值
func (k *KlineDatas) ATR_(period int) float64 {
	atr, err := k.ATR(period)
	if err != nil {
		return 0
	}
	return atr.Value()
}

// Value 获取最新的ATR值
// 说明：
//
//	返回最新的ATR值
//	使用建议：
//	- 可用于设置动态止损位置
//	- 可用于调整交易仓位大小
//	- 可用于判断市场波动状态
//
// 返回值：
//   - float64: 最新的ATR值
func (t *TaATR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// Percent 计算ATR相对于当前价格的百分比
// 说明：
//
//	计算ATR值占当前价格的百分比，用于：
//	1. 评估价格波动的相对幅度
//	2. 设置百分比止损位置
//	3. 对不同价位的品种进行波动性比较
//
// 参数：
//   - currentPrice: 当前价格
//
// 返回值：
//   - float64: ATR占当前价格的百分比
//     注意：如果当前价格小于等于0，返回0
func (t *TaATR) Percent(currentPrice float64) float64 {
	if currentPrice <= 0 {
		return 0
	}
	return t.Value() / currentPrice * 100
}
