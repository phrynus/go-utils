package ta

import (
	"fmt"
	"math"
)

// TaDpo 表示偏离价格振荡器(Detrended Price Oscillator)的计算结果
// 说明：
//
//	DPO是一种趋势跟踪指标，用于识别价格周期：
//	1. DPO短周期：收盘价与短周期SMA的差值
//	2. DPO长周期：收盘价与长周期SMA的差值
//	3. DPO差值：短周期DPO减去长周期DPO
//	4. DPO高点：差值的X周期最高点
//	5. DPO低点：差值的X周期最低点
//	6. DPO中点：高点和低点的平均值
type TaDpo struct {
	Short        []float64 `json:"short"`         // 短周期DPO序列
	Long         []float64 `json:"long"`          // 长周期DPO序列
	Diff         []float64 `json:"diff"`          // DPO差值序列
	High         []float64 `json:"high"`          // 差值的X周期最高点
	Low          []float64 `json:"low"`           // 差值的X周期最低点
	Mid          []float64 `json:"mid"`           // 高点和低点的平均值
	ShortPeriod  int       `json:"short_period"`  // 短周期
	LongPeriod   int       `json:"long_period"`   // 长周期
	XPeriod      int       `json:"x_period"`      // 差值长度周期
	SmoothPeriod int       `json:"smooth_period"` // 平滑周期
}

// CalculateDPO 计算DPO指标
// 说明：
//
//	计算步骤：
//	1. 计算偏移量：
//	   - 短周期偏移量 = floor(短周期 / 2) + 1
//	   - 长周期偏移量 = floor(长周期 / 2) + 1
//	2. 计算原始DPO值：
//	   - 短周期DPO = 收盘价 - 短周期SMA[短周期偏移量]
//	   - 长周期DPO = 收盘价 - 长周期SMA[长周期偏移量]
//	3. 对DPO值进行平滑处理
//	4. 计算DPO差值：短周期DPO - 长周期DPO
//	5. 计算差值的X周期最高点和最低点
//	6. 对最高点和最低点进行平滑处理
//	7. 计算中间值：(高点 + 低点) / 2
//
// 参数：
//   - prices: 价格序列
//   - shortPeriod: 短周期，默认15
//   - longPeriod: 长周期，默认19
//   - xPeriod: 差值长度周期，默认11
//   - smoothPeriod: 平滑周期，默认3
//
// 返回值：
//   - *TaDpo: 包含DPO计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	dpo, err := CalculateDPO(prices, 15, 19, 11, 3)
func CalculateDPO(prices []float64, shortPeriod, longPeriod, xPeriod, smoothPeriod int) (*TaDpo, error) {
	length := len(prices)
	if length < longPeriod {
		return nil, fmt.Errorf("价格数据长度(%d)小于长周期(%d)", length, longPeriod)
	}

	// 计算偏移量
	offsetShort := int(math.Floor(float64(shortPeriod)/2.0)) + 1
	offsetLong := int(math.Floor(float64(longPeriod)/2.0)) + 1

	// 计算短周期和长周期SMA
	shortSMA, err := CalculateSMA(prices, shortPeriod)
	if err != nil {
		return nil, err
	}

	longSMA, err := CalculateSMA(prices, longPeriod)
	if err != nil {
		return nil, err
	}

	// 计算原始DPO值
	dpoShortRaw := make([]float64, length)
	dpoLongRaw := make([]float64, length)
	for i := 0; i < length; i++ {
		if i < shortPeriod+offsetShort-1 {
			dpoShortRaw[i] = 0
		} else {
			dpoShortRaw[i] = prices[i] - shortSMA.Values[i-offsetShort]
		}

		if i < longPeriod+offsetLong-1 {
			dpoLongRaw[i] = 0
		} else {
			dpoLongRaw[i] = prices[i] - longSMA.Values[i-offsetLong]
		}
	}

	// 对DPO值进行平滑处理
	dpoShort, err := CalculateSMA(dpoShortRaw, smoothPeriod)
	if err != nil {
		return nil, err
	}

	dpoLong, err := CalculateSMA(dpoLongRaw, smoothPeriod)
	if err != nil {
		return nil, err
	}

	// 计算DPO差值
	dpoDiff := make([]float64, length)
	for i := 0; i < length; i++ {
		dpoDiff[i] = dpoShort.Values[i] - dpoLong.Values[i]
	}

	// 计算差值的X周期最高点和最低点
	dpoDiffHigh := make([]float64, length)
	dpoDiffLow := make([]float64, length)
	for i := 0; i < length; i++ {
		if i < xPeriod-1 {
			dpoDiffHigh[i] = 0
			dpoDiffLow[i] = 0
		} else {
			// 查找X周期内的最高值和最低值
			high := dpoDiff[i]
			low := dpoDiff[i]
			for j := 0; j < xPeriod; j++ {
				if i-j >= 0 {
					if dpoDiff[i-j] > high {
						high = dpoDiff[i-j]
					}
					if dpoDiff[i-j] < low {
						low = dpoDiff[i-j]
					}
				}
			}
			dpoDiffHigh[i] = high
			dpoDiffLow[i] = low
		}
	}

	// 对最高点和最低点进行平滑处理
	dpoDiffHighSmooth, err := CalculateSMA(dpoDiffHigh, smoothPeriod)
	if err != nil {
		return nil, err
	}

	dpoDiffLowSmooth, err := CalculateSMA(dpoDiffLow, smoothPeriod)
	if err != nil {
		return nil, err
	}

	// 计算中间值
	dpoDiffMid := make([]float64, length)
	for i := 0; i < length; i++ {
		dpoDiffMid[i] = (dpoDiffHighSmooth.Values[i] + dpoDiffLowSmooth.Values[i]) / 2
	}

	return &TaDpo{
		Short:        dpoShort.Values,
		Long:         dpoLong.Values,
		Diff:         dpoDiff,
		High:         dpoDiffHighSmooth.Values,
		Low:          dpoDiffLowSmooth.Values,
		Mid:          dpoDiffMid,
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		XPeriod:      xPeriod,
		SmoothPeriod: smoothPeriod,
	}, nil
}

// DPO 为K线数据计算DPO指标
// 说明：
//
//	对指定价格类型计算DPO指标
//
// 参数：
//   - source: 价格类型，支持"open"、"high"、"low"、"close"等
//   - shortPeriod: 短周期
//   - longPeriod: 长周期
//   - xPeriod: 差值长度周期
//   - smoothPeriod: 平滑周期
//
// 返回值：
//   - *TaDpo: 包含DPO计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) DPO(source string, shortPeriod, longPeriod, xPeriod, smoothPeriod int) (*TaDpo, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateDPO(prices, shortPeriod, longPeriod, xPeriod, smoothPeriod)
}

// Value 获取最新的DPO值
// 说明：
//
//	返回最新的DPO短周期、长周期、差值、高点、低点和中点值
//	使用建议：
//	- 差值由负变正，买入信号
//	- 差值由正变负，卖出信号
//	- 差值位于高点，超买状态
//	- 差值位于低点，超卖状态
//
// 返回值：
//   - short: 最新的短周期DPO值
//   - long: 最新的长周期DPO值
//   - diff: 最新的差值
//   - high: 最新的高点
//   - low: 最新的低点
//   - mid: 最新的中点
func (t *TaDpo) Value() (short, long, diff, high, low, mid float64) {
	lastIndex := len(t.Diff) - 1
	return t.Short[lastIndex], t.Long[lastIndex], t.Diff[lastIndex], t.High[lastIndex], t.Low[lastIndex], t.Mid[lastIndex]
}
