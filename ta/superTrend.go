package ta

import (
	"fmt"
	"math"
)

// TaSuperTrend 表示超级趋势指标(SuperTrend)的计算结果
// 说明：
//
//	SuperTrend是一个趋势跟踪指标，结合了ATR和中轨线的概念：
//	1. 利用ATR来衡量市场波动性
//	2. 根据中轨线和ATR计算上下轨
//	3. 通过价格与轨道的关系判断趋势
//	特点：
//	- 能有效识别市场趋势
//	- 自动调整对市场波动的敏感度
//	- 提供明确的趋势反转信号
//	- 可作为止损止盈的参考位置
type TaSuperTrend struct {
	Values     []float64 `json:"values"`     // 指标值序列，上涨趋势时为下轨，下跌趋势时为上轨
	Trend      []int     `json:"direction"`  // 趋势方向：1表示上涨，-1表示下跌，0表示初始状态
	Upper      []float64 `json:"upper_band"` // 上轨线序列
	Lower      []float64 `json:"lower_band"` // 下轨线序列
	Period     int       `json:"period"`     // ATR计算周期
	Multiplier float64   `json:"multiplier"` // ATR乘数，用于调整轨道宽度
}

// CalculateSuperTrend 计算给定K线数据的超级趋势指标
// 说明：
//
//	计算步骤：
//	1. 计算ATR值作为波动参考
//	2. 计算中轨线 = (最高价 + 最低价) / 2
//	3. 计算上轨线 = 中轨线 + multiplier * ATR
//	4. 计算下轨线 = 中轨线 - multiplier * ATR
//	5. 根据收盘价与轨道的位置关系确定趋势
//	趋势判断规则：
//	- 当收盘价上穿上轨线时，趋势转为上涨
//	- 当收盘价下穿下轨线时，趋势转为下跌
//
// 参数：
//   - klineData: K线数据
//   - period: ATR计算周期，通常为7-14
//   - multiplier: ATR乘数，通常为2-3，越大轨道越宽
//
// 返回值：
//   - *TaSuperTrend: 包含SuperTrend计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	superTrend, err := CalculateSuperTrend(klineData, 10, 3.0)
func CalculateSuperTrend(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrend, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := klineData.ATR(period)
	if err != nil {
		return nil, err
	}

	length := len(klineData)

	slices := preallocateSlices(length, 3)
	upperBand, lowerBand, values := slices[0], slices[1], slices[2]
	trend := make([]int, length)

	for i := period; i < length; i++ {
		midpoint := (klineData[i].High + klineData[i].Low) / 2
		atrValue := atr.Values[i]
		upperBand[i] = midpoint + multiplier*atrValue
		lowerBand[i] = midpoint - multiplier*atrValue
	}

	if klineData[period].Close > lowerBand[period] {
		trend[period] = 1
	} else {
		trend[period] = -1
	}
	if trend[period] == 1 {
		values[period] = lowerBand[period]
	} else {
		values[period] = upperBand[period]
	}

	for i := period + 1; i < length; i++ {
		if trend[i-1] == 1 {
			if klineData[i].Close < lowerBand[i] {
				trend[i] = -1
				upperBand[i] = upperBand[i-1]
			} else {
				trend[i] = 1
				lowerBand[i] = math.Max(lowerBand[i], lowerBand[i-1])
			}
		} else {
			if klineData[i].Close > upperBand[i] {
				trend[i] = 1
				lowerBand[i] = lowerBand[i-1]
			} else {
				trend[i] = -1
				upperBand[i] = math.Min(upperBand[i], upperBand[i-1])
			}
		}

		if trend[i] == 1 {
			values[i] = lowerBand[i]
		} else if trend[i] == -1 {
			values[i] = upperBand[i]
		}
	}

	return &TaSuperTrend{
		Values:     values,
		Trend:      trend,
		Upper:      upperBand,
		Lower:      lowerBand,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

// SuperTrend 为K线数据计算超级趋势指标
// 说明：
//
//	对当前K线数据计算SuperTrend指标
//
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - *TaSuperTrend: 包含SuperTrend计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) SuperTrend(period int, multiplier float64) (*TaSuperTrend, error) {
	return CalculateSuperTrend(*k, period, multiplier)
}

// SuperTrend_ 获取最新的SuperTrend指标值
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - float64: 最新的上轨线值
//   - float64: 最新的下轨线值
//   - bool: 当前趋势方向，true表示上涨趋势，false表示下跌趋势
func (k *KlineDatas) SuperTrend_(period int, multiplier float64) (float64, float64, int) {
	superTrend, err := k.SuperTrend(period, multiplier)
	if err != nil {
		return 0, 0, 0
	}
	return superTrend.Value()
}

// Value 获取最新的SuperTrend指标值
// 说明：
//
//	返回最新的上轨线、下轨线值和趋势方向
//	使用建议：
//	- 上涨趋势时，下轨线可作为止损位
//	- 下跌趋势时，上轨线可作为止损位
//	- 趋势反转时及时调整持仓方向
//
// 返回值：
//   - upper: 最新的上轨线值
//   - lower: 最新的下轨线值
//   - isUpTrend: 当前趋势方向，true表示上涨趋势，false表示下跌趋势
func (t *TaSuperTrend) Value() (upper, lower float64, trend int) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
