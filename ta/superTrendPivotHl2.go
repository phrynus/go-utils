package ta

import (
	"fmt"
)

// TaSuperTrendPivotHl2 表示基于HL2的超级趋势指标计算结果
// 说明：
//
//	SuperTrendPivotHl2是SuperTrend的另一个变种，使用HL2(最高价和最低价的平均值)作为中心价格：
//	1. 使用HL2代替传统的中轨计算方法
//	2. 结合ATR动态调整轨道宽度
//	3. 提供更平滑的趋势跟踪效果
//	特点：
//	- 使用HL2降低价格波动的影响
//	- 趋势转换更加平滑
//	- 假突破信号较少
//	- 适合波动较大的市场
type TaSuperTrendPivotHl2 struct {
	Values     []float64 `json:"values"`     // 指标值序列，上涨趋势时为下轨，下跌趋势时为上轨
	Direction  []int     `json:"direction"`  // 趋势方向：1表示上涨，-1表示下跌，0表示初始状态
	UpperBand  []float64 `json:"upper_band"` // 上轨线序列
	LowerBand  []float64 `json:"lower_band"` // 下轨线序列
	Period     int       `json:"period"`     // ATR计算周期
	Multiplier float64   `json:"multiplier"` // ATR乘数，用于调整轨道宽度
}

// CalculateSuperTrendPivotHl2 计算基于HL2的超级趋势指标
// 说明：
//
//	计算步骤：
//	1. 计算HL2 = (最高价 + 最低价) / 2
//	2. 计算ATR值作为波动参考
//	3. 计算基础轨道：
//	   - 上轨 = HL2 + multiplier * ATR
//	   - 下轨 = HL2 - multiplier * ATR
//	4. 根据价格穿越情况调整最终轨道
//	5. 确定趋势方向和指标值
//	趋势判断规则：
//	- 当收盘价上穿上轨时，趋势转为上涨
//	- 当收盘价下穿下轨时，趋势转为下跌
//
// 参数：
//   - klineData: K线数据
//   - period: ATR计算周期，通常为10-21
//   - multiplier: ATR乘数，通常为2-3，越大轨道越宽
//
// 返回值：
//   - *TaSuperTrendPivotHl2: 包含计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	hl2, err := CalculateSuperTrendPivotHl2(klineData, 14, 2.0)
func CalculateSuperTrendPivotHl2(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	length := len(klineData)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := CalculateATR(klineData, period)
	if err != nil {
		return nil, err
	}

	slices := preallocateSlices(length, 4)
	values, direction, upperBand, lowerBand := slices[0], make([]int, length), slices[2], slices[3]

	for i := 0; i < length; i++ {

		hl2 := (klineData[i].High + klineData[i].Low) / 2

		if i < period {

			upperBand[i] = hl2 + multiplier*atr.Values[i]
			lowerBand[i] = hl2 - multiplier*atr.Values[i]
			direction[i] = 0
			values[i] = hl2
			continue
		}

		basicUpperBand := hl2 + multiplier*atr.Values[i]
		basicLowerBand := hl2 - multiplier*atr.Values[i]

		if basicLowerBand > lowerBand[i-1] || klineData[i-1].Close < lowerBand[i-1] {
			lowerBand[i] = basicLowerBand
		} else {
			lowerBand[i] = lowerBand[i-1]
		}

		if basicUpperBand < upperBand[i-1] || klineData[i-1].Close > upperBand[i-1] {
			upperBand[i] = basicUpperBand
		} else {
			upperBand[i] = upperBand[i-1]
		}

		if direction[i-1] <= 0 {
			if klineData[i].Close > upperBand[i] {
				direction[i] = 1
			} else {
				direction[i] = -1
			}
		} else {
			if klineData[i].Close < lowerBand[i] {
				direction[i] = -1
			} else {
				direction[i] = 1
			}
		}

		if direction[i] == 1 {
			values[i] = lowerBand[i]
		} else {
			values[i] = upperBand[i]
		}
	}

	return &TaSuperTrendPivotHl2{
		Values:     values,
		Direction:  direction,
		UpperBand:  upperBand,
		LowerBand:  lowerBand,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

// SuperTrendPivotHl2 为K线数据计算基于HL2的超级趋势指标
// 说明：
//
//	对当前K线数据计算SuperTrendPivotHl2指标
//
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - *TaSuperTrendPivotHl2: 包含计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) SuperTrendPivotHl2(period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	return CalculateSuperTrendPivotHl2(*k, period, multiplier)
}

// Value 获取最新的SuperTrendPivotHl2指标值
// 说明：
//
//	返回最新的指标值，该值代表当前的趋势线：
//	- 上涨趋势时返回下轨值（支撑位）
//	- 下跌趋势时返回上轨值（压力位）
//	使用建议：
//	- 可将返回值作为趋势跟踪的参考线
//	- 价格站上该线视为看多信号
//	- 价格跌破该线视为看空信号
//
// 返回值：
//   - float64: 最新的指标值
func (t *TaSuperTrendPivotHl2) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
