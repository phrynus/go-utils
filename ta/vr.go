package ta

import (
	"fmt"
	"math"
)

// TaVolatilityRatio 表示波动率比率(Volatility Ratio)指标的计算结果
// 说明：
//
//	波动率比率(VR)是一个用于衡量市场波动性变化的技术指标：
//	1. 通过比较短期和长期的真实波幅(TR)来衡量波动强度
//	2. 可以帮助识别市场波动的扩大和收缩
//	3. 用于判断市场是否即将发生重要变化
//	特点：
//	- VR > 1 表示短期波动高于长期波动，市场活跃度增加
//	- VR < 1 表示短期波动低于长期波动，市场活跃度降低
//	- VR 的突变通常预示着市场即将发生重要变化
type TaVolatilityRatio struct {
	Values []float64 `json:"values"` // 波动率比率的时间序列
	Period int       `json:"period"` // 计算使用的长周期
}

// CalculateVolatilityRatio 计算波动率比率指标
// 说明：
//
//	计算步骤：
//	1. 计算每个周期的真实波幅(TR)：
//	   TR = max(high-low, |high-prevClose|, |low-prevClose|)
//	2. 分别计算短期和长期的平均TR
//	3. 计算短期TR与长期TR的比率
//	使用场景：
//	- 识别市场波动性的变化
//	- 预测可能的趋势变化
//	- 辅助判断交易时机
//
// 参数：
//   - klineData: K线数据
//   - shortPeriod: 短周期，通常为3-7
//   - longPeriod: 长周期，通常为14-21
//
// 返回值：
//   - *TaVolatilityRatio: 包含VR计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	vr, err := CalculateVolatilityRatio(klineData, 5, 14)
func CalculateVolatilityRatio(klineData KlineDatas, shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	if len(klineData) < longPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	slices := preallocateSlices(length, 2)
	trueRange, ratio := slices[0], slices[1]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	for i := longPeriod; i < length; i++ {
		var shortTR float64
		for j := i - shortPeriod + 1; j <= i; j++ {
			shortTR += trueRange[j]
		}
		shortTR /= float64(shortPeriod)

		var longTR float64
		for j := i - longPeriod + 1; j <= i; j++ {
			longTR += trueRange[j]
		}
		longTR /= float64(longPeriod)

		if longTR != 0 {
			ratio[i] = shortTR / longTR
		} else {
			ratio[i] = 1.0
		}
	}

	return &TaVolatilityRatio{
		Values: ratio,
		Period: longPeriod,
	}, nil
}

// VolatilityRatio 为K线数据计算波动率比率指标
// 说明：
//
//	对当前K线数据计算VR指标
//
// 参数：
//   - shortPeriod: 短周期
//   - longPeriod: 长周期
//
// 返回值：
//   - *TaVolatilityRatio: 包含VR计算结果的结构体指针
//   - error: 计算过程中的错误
func (k KlineDatas) VolatilityRatio(shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	return CalculateVolatilityRatio(k, shortPeriod, longPeriod)
}

// Value 获取最新的波动率比率值
// 说明：
//
//	返回VR指标的最新值
//	使用建议：
//	- VR > 1.2 表示波动显著扩大，可能即将发生趋势变化
//	- VR < 0.8 表示波动显著收缩，可能处于盘整阶段
//	- VR的快速变化预示市场状态的转换
//
// 返回值：
//   - float64: 最新的VR值，如果没有数据返回0
func (vr *TaVolatilityRatio) Value() float64 {
	if len(vr.Values) == 0 {
		return 0
	}
	return vr.Values[len(vr.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
