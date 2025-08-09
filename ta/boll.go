package ta

import (
	"fmt"
	"math"
)

// TaBoll 表示布林带(Bollinger Bands)的计算结果
// 说明：
//
//	布林带是由John Bollinger开发的技术分析工具，包含三条轨道线：
//	1. 中轨：N周期简单移动平均线(SMA)
//	2. 上轨：中轨 + K倍标准差
//	3. 下轨：中轨 - K倍标准差
//	特点：
//	- 价格通常在上下轨道之间波动
//	- 轨道宽度反映市场波动性
//	- 可用于判断超买超卖和趋势强度
type TaBoll struct {
	Upper []float64 `json:"upper"` // 上轨线序列
	Mid   []float64 `json:"mid"`   // 中轨线序列（移动平均线）
	Lower []float64 `json:"lower"` // 下轨线序列
}

// CalculateBoll 计算布林带指标
// 说明：
//
//	计算步骤：
//	1. 计算中轨（简单移动平均线）
//	2. 计算标准差：
//	   - 计算每个价格与移动平均线的差值
//	   - 计算差值的平方和
//	   - 计算平方和的均值并开方
//	3. 计算上下轨：
//	   - 上轨 = 中轨 + K倍标准差
//	   - 下轨 = 中轨 - K倍标准差
//
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，通常为20
//   - stdDev: 标准差倍数，通常为2
//
// 返回值：
//   - *TaBoll: 包含布林带计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	boll, err := CalculateBoll(prices, 20, 2.0)
func CalculateBoll(prices []float64, period int, stdDev float64) (*TaBoll, error) {

	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	upper, mid, lower := slices[0], slices[1], slices[2]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}

	mid[period-1] = sum / float64(period)

	for i := period; i < length; i++ {

		sum = sum - prices[i-period] + prices[i]

		mid[i] = sum / float64(period)
	}

	for i := period - 1; i < length; i++ {

		var sumSquares float64

		for j := 0; j < period; j++ {
			diff := prices[i-j] - mid[i]
			sumSquares += diff * diff
		}

		sd := math.Sqrt(sumSquares / float64(period))

		band := sd * stdDev

		upper[i] = mid[i] + band

		lower[i] = mid[i] - band
	}

	return &TaBoll{
		Upper: upper,
		Mid:   mid,
		Lower: lower,
	}, nil
}

// Boll 为K线数据计算布林带指标
// 说明：
//
//	对指定价格类型计算布林带指标
//
// 参数：
//   - period: 计算周期
//   - stdDev: 标准差倍数
//   - source: 价格类型，支持"open"、"high"、"low"、"close"等
//
// 返回值：
//   - *TaBoll: 包含布林带计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) Boll(period int, stdDev float64, source string) (*TaBoll, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateBoll(prices, period, stdDev)
}

// Boll_ 获取最新的布林带值
// 参数：
//   - period: 计算周期
//   - stdDev: 标准差倍数
//   - source: 价格类型
//
// 返回值：
//   - float64: 最新的上轨值
//   - float64: 最新的中轨值
//   - float64: 最新的下轨值
func (k *KlineDatas) Boll_(period int, stdDev float64, source string) (float64, float64, float64) {
	boll, err := k.Boll(period, stdDev, source)
	if err != nil {
		return 0, 0, 0
	}
	return boll.Value()
}

// Value 获取最新的布林带值
// 说明：
//
//	返回最新的上中下轨值
//	使用建议：
//	- 价格突破上轨可能表示超买
//	- 价格突破下轨可能表示超卖
//	- 轨道收窄预示行情即将剧烈波动
//	- 轨道变宽表示波动性增加
//
// 返回值：
//   - upper: 上轨值
//   - mid: 中轨值
//   - lower: 下轨值
func (t *TaBoll) Value() (upper, mid, lower float64) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Mid[lastIndex], t.Lower[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
