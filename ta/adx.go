package ta

import (
	"fmt"
	"math"
)

// TaADX 表示平均趋向指标(Average Directional Index)的计算结果
// 说明：
//
//	ADX是一个趋势强度指标，由Welles Wilder开发，包含三个指标：
//	1. ADX: 趋势强度指标，不分方向
//	2. +DI: 上升趋向指标
//	3. -DI: 下降趋向指标
//	特点：
//	- ADX > 25 表示存在强趋势
//	- ADX < 20 表示趋势较弱
//	- ADX上升表示趋势增强
//	- ADX下降表示趋势减弱
type TaADX struct {
	ADX     []float64 `json:"adx"`      // ADX值序列
	PlusDI  []float64 `json:"plus_di"`  // +DI值序列
	MinusDI []float64 `json:"minus_di"` // -DI值序列
	Period  int       `json:"period"`   // 计算周期
}

// CalculateADX 计算平均趋向指标
// 说明：
//
//	计算步骤：
//	1. 计算+DM和-DM：
//	   +DM = 当日最高价 - 前日最高价 (如果为正且大于-DM)
//	   -DM = 前日最低价 - 当日最低价 (如果为正且大于+DM)
//	2. 计算真实波幅TR：
//	   TR = max(high-low, |high-prevClose|, |low-prevClose|)
//	3. 计算平滑后的+DI和-DI：
//	   +DI = 100 * 平滑(+DM) / 平滑(TR)
//	   -DI = 100 * 平滑(-DM) / 平滑(TR)
//	4. 计算ADX：
//	   DX = 100 * |+DI - -DI| / (+DI + -DI)
//	   ADX = 平滑(DX)
//
// 参数：
//   - klineData: K线数据
//   - period: 计算周期，通常为14
//
// 返回值：
//   - *TaADX: 包含ADX计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	adx, err := CalculateADX(klineData, 14)
func CalculateADX(klineData KlineDatas, period int) (*TaADX, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 6)
	plusDM, minusDM, trueRange, plusDI, minusDI, adx := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevHigh := klineData[i-1].High
		prevLow := klineData[i-1].Low

		upMove := high - prevHigh
		downMove := prevLow - low

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}

		tr1 := high - low
		tr2 := math.Abs(high - klineData[i-1].Close)
		tr3 := math.Abs(low - klineData[i-1].Close)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	var smoothPlusDM, smoothMinusDM, smoothTR float64

	for i := 1; i <= period; i++ {
		smoothPlusDM += plusDM[i]
		smoothMinusDM += minusDM[i]
		smoothTR += trueRange[i]
	}

	if smoothTR > 0 {
		plusDI[period] = 100 * smoothPlusDM / smoothTR
		minusDI[period] = 100 * smoothMinusDM / smoothTR
	}

	for i := period + 1; i < length; i++ {

		smoothPlusDM = smoothPlusDM - (smoothPlusDM / float64(period)) + plusDM[i]
		smoothMinusDM = smoothMinusDM - (smoothMinusDM / float64(period)) + minusDM[i]
		smoothTR = smoothTR - (smoothTR / float64(period)) + trueRange[i]

		if smoothTR > 0 {
			plusDI[i] = 100 * smoothPlusDM / smoothTR
			minusDI[i] = 100 * smoothMinusDM / smoothTR
		}

		diSum := math.Abs(plusDI[i] - minusDI[i])
		diDiff := plusDI[i] + minusDI[i]
		if diDiff > 0 {
			adx[i] = 100 * diSum / diDiff
		}
	}

	var smoothADX float64
	for i := period * 2; i < length; i++ {
		if i == period*2 {

			for j := period; j <= i; j++ {
				smoothADX += adx[j]
			}
			adx[i] = smoothADX / float64(period+1)
		} else {

			adx[i] = (adx[i-1]*float64(period-1) + adx[i]) / float64(period)
		}
	}

	return &TaADX{
		ADX:     adx,
		PlusDI:  plusDI,
		MinusDI: minusDI,
		Period:  period,
	}, nil
}

// ADX 为K线数据计算平均趋向指标
// 说明：
//
//	对当前K线数据计算ADX指标
//
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaADX: 包含ADX计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) ADX(period int) (*TaADX, error) {
	return CalculateADX(*k, period)
}

// Value 获取最新的ADX、+DI和-DI值
// 说明：
//
//	返回最新的三个指标值
//	使用建议：
//	- ADX用于判断趋势强度
//	- +DI和-DI用于判断趋势方向
//	- +DI > -DI 表示上升趋势
//	- -DI > +DI 表示下降趋势
//
// 返回值：
//   - adx: 趋势强度值
//   - plusDI: 上升趋向值
//   - minusDI: 下降趋向值
func (t *TaADX) Value() (adx, plusDI, minusDI float64) {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex], t.PlusDI[lastIndex], t.MinusDI[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// CrossOver 检测DI线的交叉信号
// 说明：
//
//	检测+DI和-DI的交叉情况，用于产生交易信号：
//	- +DI上穿-DI为买入信号
//	- +DI下穿-DI为卖出信号
//	使用建议：
//	- 结合ADX > 25时的信号更可靠
//	- 可以作为趋势交易的入场信号
//	- 建议与其他指标配合使用
//
// 返回值：
//   - 1: +DI上穿-DI（买入信号）
//   - -1: +DI下穿-DI（卖出信号）
//   - 0: 无交叉信号
func (t *TaADX) CrossOver() int {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return 0
	}
	lastIndex := len(t.PlusDI) - 1
	if t.PlusDI[lastIndex-1] < t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] > t.MinusDI[lastIndex] {
		return 1
	} else if t.PlusDI[lastIndex-1] > t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] < t.MinusDI[lastIndex] {
		return -1
	} else {
		return 0
	}
}
