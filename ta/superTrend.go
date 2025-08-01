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
	Upper      []float64 `json:"upper"`      // 上轨线的时间序列
	Lower      []float64 `json:"lower"`      // 下轨线的时间序列
	Trend      []bool    `json:"trend"`      // 趋势方向，true表示上涨趋势，false表示下跌趋势
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

	slices := preallocateSlices(length, 2)
	upperBand, lowerBand := slices[0], slices[1]
	trend := make([]bool, length)

	for i := period; i < length; i++ {
		midpoint := (klineData[i].High + klineData[i].Low) / 2
		atrValue := atr.Values[i]
		upperBand[i] = midpoint + multiplier*atrValue
		lowerBand[i] = midpoint - multiplier*atrValue
	}

	trend[period] = klineData[period].Close > lowerBand[period]

	for i := period + 1; i < length; i++ {
		if trend[i-1] {
			if klineData[i].Close < lowerBand[i] {
				trend[i] = false
				upperBand[i] = upperBand[i-1]
			} else {
				trend[i] = true
				lowerBand[i] = math.Max(lowerBand[i], lowerBand[i-1])
			}
		} else {
			if klineData[i].Close > upperBand[i] {
				trend[i] = true
				lowerBand[i] = lowerBand[i-1]
			} else {
				trend[i] = false
				upperBand[i] = math.Min(upperBand[i], upperBand[i-1])
			}
		}
	}

	return &TaSuperTrend{
		Upper:      upperBand,
		Lower:      lowerBand,
		Trend:      trend,
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
func (t *TaSuperTrend) Value() (upper, lower float64, isUpTrend bool) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
