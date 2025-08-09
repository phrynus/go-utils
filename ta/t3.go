package ta

import (
	"fmt"
)

// TaT3 表示Tillson T3移动平均线的计算结果
// 说明：
//
//	T3是Tim Tillson开发的一种改进型移动平均线，具有以下特点：
//	1. 通过多重EMA计算减少滞后性
//	2. 使用体积因子(Volume Factor)调整平滑度
//	3. 比传统移动平均更好地跟踪趋势
//	4. 对价格转折点的反应更快
//	特点：
//	- 平滑度高，噪音少
//	- 转向延迟较小
//	- 可通过参数调整灵敏度
//	- 计算复杂但效果优异
type TaT3 struct {
	Values []float64 `json:"values"` // T3移动平均线的值序列
	Period int       `json:"period"` // 计算周期
	VFact  float64   `json:"vfact"`  // 体积因子，用于调整平滑度
}

// CalculateT3 计算Tillson T3移动平均线
// 说明：
//
//	计算步骤：
//	1. 计算六重指数移动平均(EMA1-EMA6)
//	2. 使用体积因子计算加权系数(c1-c4)
//	3. 将加权系数应用于EMA3-EMA6得到最终T3值
//	计算公式：
//	T3 = c1*ema6 + c2*ema5 + c3*ema4 + c4*ema3
//	其中：
//	- c1 = -a³
//	- c2 = 3a² + 3a³
//	- c3 = -6a² - 3a - 3a³
//	- c4 = 1 + 3a + a³ + 3a²
//	(a为体积因子vfact)
//
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，通常为5-30
//   - vfact: 体积因子，范围0-1，通常为0.7
//   - 接近0时更平滑但滞后更大
//   - 接近1时响应更快但可能产生更多噪音
//
// 返回值：
//   - *TaT3: 包含T3计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	t3, err := CalculateT3(prices, 10, 0.7)
func CalculateT3(prices []float64, period int, vfact float64) (*TaT3, error) {
	if len(prices) < period*6 {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 7)
	ema1, ema2, ema3, ema4, ema5, ema6, t3 := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5], slices[6]

	k := 2.0 / float64(period+1)
	ema1[0] = prices[0]
	for i := 1; i < length; i++ {
		ema1[i] = prices[i]*k + ema1[i-1]*(1-k)
	}

	for i := 1; i < length; i++ {
		ema2[i] = ema1[i]*k + ema2[i-1]*(1-k)
		ema3[i] = ema2[i]*k + ema3[i-1]*(1-k)
		ema4[i] = ema3[i]*k + ema4[i-1]*(1-k)
		ema5[i] = ema4[i]*k + ema5[i-1]*(1-k)
		ema6[i] = ema5[i]*k + ema6[i-1]*(1-k)
	}

	b := vfact
	c1 := -b * b * b
	c2 := 3*b*b + 3*b*b*b
	c3 := -6*b*b - 3*b - 3*b*b*b
	c4 := 1 + 3*b + b*b*b + 3*b*b

	for i := period * 6; i < length; i++ {
		t3[i] = c1*ema6[i] + c2*ema5[i] + c3*ema4[i] + c4*ema3[i]
	}

	return &TaT3{
		Values: t3,
		Period: period,
		VFact:  vfact,
	}, nil
}

// T3 为K线数据计算Tillson T3移动平均线
// 说明：
//
//	基于K线数据中指定的价格类型计算T3移动平均线
//
// 参数：
//   - period: 计算周期
//   - vfact: 体积因子
//   - source: 价格数据来源，可以是"close"、"open"、"high"、"low"等
//
// 返回值：
//   - *TaT3: 包含T3计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) T3(period int, vfact float64, source string) (*TaT3, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateT3(prices, period, vfact)
}

// T3_ 获取最新的T3值
// 参数：
//   - period: 计算周期
//   - vfact: 体积因子
//   - source: 价格数据来源
//
// 返回值：
//   - float64: 最新的T3值
func (k *KlineDatas) T3_(period int, vfact float64, source string) float64 {
	t3, err := k.T3(period, vfact, source)
	if err != nil {
		return 0
	}
	return t3.Value()
}

// Value 获取最新的T3值
// 说明：
//
//	返回T3移动平均线的最新值
//	使用建议：
//	- 可作为趋势方向的参考
//	- 价格在T3线上方视为上涨趋势
//	- 价格在T3线下方视为下跌趋势
//	- T3线的斜率可反映趋势强度
//
// 返回值：
//   - float64: 最新的T3值
func (t *TaT3) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
