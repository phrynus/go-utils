package ta

import (
	"fmt"
	"math"
)

// TaADX 平均趋向指标(ADX)的计算结果结构体
// 说明：
//
//	ADX指标用于衡量市场趋势的强度，不论趋势是上涨还是下跌
//
// 字段：
//   - ADX: ADX指标值数组，表示趋势强度
//   - PlusDI: +DI指标值数组，表示上升趋势的强度
//   - MinusDI: -DI指标值数组，表示下降趋势的强度
//   - Period: 计算周期
type TaADX struct {
	ADX     []float64 `json:"adx"`
	PlusDI  []float64 `json:"plus_di"`
	MinusDI []float64 `json:"minus_di"`
	Period  int       `json:"period"`
}

// CalculateADX 计算给定K线数据的ADX、+DI和-DI指标
// 参数：
//   - klineData: K线数据数组，包含OHLC价格数据
//   - period: 计算周期
//
// 返回值：
//   - *TaADX: 包含计算结果的TaADX结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 说明：
//
//	该函数实现了ADX指标的完整计算过程，包括：
//	1. 计算+DM和-DM
//	2. 计算真实波幅(TR)
//	3. 计算平滑后的+DI和-DI
//	4. 最终计算ADX值
//	计算过程采用Wilder平滑方法
//
// 示例：
//
//	adx, err := CalculateADX(klineData, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("ADX: %v, +DI: %v, -DI: %v\n", adx.ADX[len(adx.ADX)-1], adx.PlusDI[len(adx.PlusDI)-1], adx.MinusDI[len(adx.MinusDI)-1])
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

// ADX 计算K线数据的ADX指标
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaADX: ADX计算结果
//   - error: 计算过程中的错误
func (k *KlineDatas) ADX(period int) (*TaADX, error) {
	return CalculateADX(*k, period)
}

// Value 获取最新的ADX、+DI和-DI值
// 返回值：
//   - adx: 最新的ADX值
//   - plusDI: 最新的+DI值
//   - minusDI: 最新的-DI值
func (t *TaADX) Value() (adx, plusDI, minusDI float64) {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex], t.PlusDI[lastIndex], t.MinusDI[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// CrossOver 判断金叉、死叉
// 返回值：
//   - Int: 1 表示金叉，-1 表示死叉，0 表示无交叉
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
