package ta

import (
	"fmt"
	"math"
)

// TaSuperTrendPivot 表示基于轴点的超级趋势指标计算结果
// 说明：
//
//	SuperTrendPivot是SuperTrend的改进版本，通过寻找价格轴点来优化趋势线：
//	1. 使用价格轴点（高点和低点）来确定中心线
//	2. 结合ATR动态调整轨道宽度
//	3. 提供更准确的趋势转向信号
//	特点：
//	- 相比传统SuperTrend更好地识别关键价格水平
//	- 减少假突破信号
//	- 趋势跟踪能力更强
//	- 适合中长期趋势交易
type TaSuperTrendPivot struct {
	Upper       []float64 `json:"upper"`        // 上轨线的时间序列
	Lower       []float64 `json:"lower"`        // 下轨线的时间序列
	Trend       []int     `json:"trend"`        // 趋势方向：1表示上涨，-1表示下跌，0表示初始状态
	PivotPeriod int       `json:"pivot_period"` // 寻找轴点的周期范围
	Factor      float64   `json:"factor"`       // ATR乘数，用于调整轨道宽度
	AtrPeriod   int       `json:"atr_period"`   // ATR计算周期
}

// FindPivotHighPoint 在指定周期范围内寻找高点轴点
// 说明：
//
//	在给定K线位置的前后period个周期内寻找是否存在高点轴点
//	当前位置的高点必须是指定范围内的最高点才能被认定为轴点
//
// 参数：
//   - klineData: K线数据
//   - index: 当前检查的位置
//   - period: 前后查找的周期范围
//
// 返回值：
//   - float64: 如果找到轴点返回该点的高点价格，否则返回NaN
func FindPivotHighPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].High > klineData[index].High {
			return math.NaN()
		}
	}
	return klineData[index].High
}

// FindPivotLowPoint 在指定周期范围内寻找低点轴点
// 说明：
//
//	在给定K线位置的前后period个周期内寻找是否存在低点轴点
//	当前位置的低点必须是指定范围内的最低点才能被认定为轴点
//
// 参数：
//   - klineData: K线数据
//   - index: 当前检查的位置
//   - period: 前后查找的周期范围
//
// 返回值：
//   - float64: 如果找到轴点返回该点的低点价格，否则返回NaN
func FindPivotLowPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].Low < klineData[index].Low {
			return math.NaN()
		}
	}
	return klineData[index].Low
}

// CalculateSuperTrendPivot 计算基于轴点的超级趋势指标
// 说明：
//
//	计算步骤：
//	1. 在每个位置寻找高点和低点轴点
//	2. 根据轴点计算中心线：
//	   - 如果同时存在高低点，取其平均值
//	   - 如果只有一个轴点，使用该轴点
//	   - 如果没有轴点，使用当前K线的中点
//	3. 使用加权平均方式平滑中心线
//	4. 结合ATR计算上下轨
//	5. 根据收盘价与轨道的关系确定趋势
//
// 参数：
//   - klineData: K线数据
//   - pivotPeriod: 寻找轴点的周期范围，通常为5-10
//   - factor: ATR乘数，通常为2-3，越大轨道越宽
//   - atrPeriod: ATR计算周期，通常为14-21
//
// 返回值：
//   - *TaSuperTrendPivot: 包含计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	pivot, err := CalculateSuperTrendPivot(klineData, 7, 2.5, 14)
func CalculateSuperTrendPivot(klineData KlineDatas, pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {

	dataLen := len(klineData)
	if dataLen < pivotPeriod || dataLen < atrPeriod {
		return nil, fmt.Errorf("计算数据不足: 数据长度%d, 需要轴点周期%d和ATR周期%d", dataLen, pivotPeriod, atrPeriod)
	}

	trendUp := make([]float64, dataLen)
	trendDown := make([]float64, dataLen)
	trend := make([]int, dataLen)

	atr, err := klineData.ATR(atrPeriod)
	if err != nil {
		return nil, fmt.Errorf("计算ATR失败: %v", err)
	}

	var center float64
	var centerCount int

	for i := pivotPeriod; i < dataLen; i++ {

		pivotHigh := FindPivotHighPoint(klineData, i, pivotPeriod)
		pivotLow := FindPivotLowPoint(klineData, i, pivotPeriod)

		if !math.IsNaN(pivotHigh) || !math.IsNaN(pivotLow) {
			newCenter := 0.0
			if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {

				newCenter = (pivotHigh + pivotLow) / 2
			} else if !math.IsNaN(pivotHigh) {
				newCenter = pivotHigh
			} else {
				newCenter = pivotLow
			}

			if centerCount == 0 {
				center = newCenter
			} else {
				center = (center*2 + newCenter) / 3
			}
			centerCount++
		}

		if centerCount == 0 {
			center = (klineData[i].High + klineData[i].Low) / 2
		}

		band := factor * atr.Values[i]
		upperBand := center + band
		lowerBand := center - band

		if i > 0 {

			if klineData[i-1].Close > trendUp[i-1] {
				trendUp[i] = math.Max(lowerBand, trendUp[i-1])
			} else {
				trendUp[i] = lowerBand
			}

			if klineData[i-1].Close < trendDown[i-1] {
				trendDown[i] = math.Min(upperBand, trendDown[i-1])
			} else {
				trendDown[i] = upperBand
			}

			if klineData[i].Close > trendDown[i-1] {
				trend[i] = 1
			} else if klineData[i].Close < trendUp[i-1] {
				trend[i] = -1
			} else {
				trend[i] = trend[i-1]
			}
		} else {

			trendUp[i] = lowerBand
			trendDown[i] = upperBand
			trend[i] = 0
		}
	}

	return &TaSuperTrendPivot{
		Upper:       trendDown,
		Lower:       trendUp,
		Trend:       trend,
		PivotPeriod: pivotPeriod,
		Factor:      factor,
		AtrPeriod:   atrPeriod,
	}, nil
}

// SuperTrendPivot 为K线数据计算基于轴点的超级趋势指标
// 说明：
//
//	对当前K线数据计算SuperTrendPivot指标
//
// 参数：
//   - pivotPeriod: 寻找轴点的周期范围
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - *TaSuperTrendPivot: 包含计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) SuperTrendPivot(pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {
	return CalculateSuperTrendPivot(*k, pivotPeriod, factor, atrPeriod)
}

// Value 获取最新的SuperTrendPivot指标值
// 说明：
//
//	返回最新的上轨线、下轨线值和趋势方向
//	使用建议：
//	- 趋势为1时表示上涨趋势，考虑做多
//	- 趋势为-1时表示下跌趋势，考虑做空
//	- 上涨趋势时下轨可作为止损位
//	- 下跌趋势时上轨可作为止损位
//
// 返回值：
//   - upper: 最新的上轨线值
//   - lower: 最新的下轨线值
//   - trend: 当前趋势方向(1: 上涨, -1: 下跌, 0: 初始状态)
func (t *TaSuperTrendPivot) Value() (upper, lower float64, trend int) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
