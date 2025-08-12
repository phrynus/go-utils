package ta

import (
	"fmt"
)

// TaJingZheMA 表示惊蛰均线指标的计算结果
// 说明：
//
//	惊蛰均线是基于EMA和OBV组合的复合指标：
//	1. 结合价格与成交量变化进行趋势判断
//	2. 通过多周期EMA和OBV信号提供交易参考
//	3. 用于捕捉市场趋势转换的关键点
//	特点：
//	- 综合考量价格和成交量的变化
//	- 提供多种条件信号线
//	- 可用于趋势跟踪和趋势反转判断
//	- 适合中长期交易策略
type TaJingZheMA struct {
	Cond1Values []float64 `json:"cond1_values"` // 条件1结果序列
	Cond2Values []float64 `json:"cond2_values"` // 条件2结果序列
	Cond3Values []float64 `json:"cond3_values"` // 条件3结果序列
	Cond4Values []float64 `json:"cond4_values"` // 条件4结果序列
	Cond5Values []float64 `json:"cond5_values"` // 条件5结果序列
	Period1     int       `json:"period1"`      // 主要周期
	Period2     int       `json:"period2"`      // OBV周期
}

// CalculateJingZheMA 计算惊蛰均线指标
// 说明：
//
//	计算步骤：
//	1. 计算三个周期的价格EMA指标
//	2. 计算OBV指标及其两个不同周期的EMA
//	3. 根据不同条件生成5个条件信号线
//	应用场景：
//	- 趋势确认和跟踪
//	- 发现价量背离
//	- 捕捉趋势转折点
//
// 参数：
//   - prices: 价格序列（通常使用收盘价）
//   - volumes: 成交量序列
//   - period1: 主要计算周期
//   - period2: OBV的EMA计算周期
//
// 返回值：
//   - *TaJingZheMA: 包含惊蛰均线计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	jzma, err := CalculateJingZheMA(prices, volumes, 25, 6)
func CalculateJingZheMA(prices, volumes []float64, period1, period2 int) (*TaJingZheMA, error) {
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("价格和成交量数据长度不一致")
	}

	if len(prices) < period1*3 {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 计算三个周期的EMA
	ema1, err := CalculateEMA(prices, period1)
	if err != nil {
		return nil, err
	}

	ema2, err := CalculateEMA(prices, period1*2)
	if err != nil {
		return nil, err
	}

	ema3, err := CalculateEMA(prices, period1*3)
	if err != nil {
		return nil, err
	}

	// 计算OBV及其EMA
	obv, err := CalculateOBV(prices, volumes)
	if err != nil {
		return nil, err
	}

	obv1, err := CalculateEMA(obv.Values, period1)
	if err != nil {
		return nil, err
	}

	obv2, err := CalculateEMA(obv.Values, period2)
	if err != nil {
		return nil, err
	}

	length := len(prices)
	cond1 := make([]float64, length)
	cond2 := make([]float64, length)
	cond3 := make([]float64, length)
	cond4 := make([]float64, length)
	cond5 := make([]float64, length)

	// 初始化条件值
	var upMa1, upMa2, upMa3, upMa4, upMa5 float64 = 0, 0, 0, 0, 0

	// 计算所有条件信号
	for i := 1; i < length; i++ {
		// 确保索引在有效范围内
		emaIdx1 := i
		emaIdx2 := i
		emaIdx3 := i
		obv1Idx := i
		obv2Idx := i

		if emaIdx1 >= len(ema1.Values) {
			emaIdx1 = len(ema1.Values) - 1
		}
		if emaIdx2 >= len(ema2.Values) {
			emaIdx2 = len(ema2.Values) - 1
		}
		if emaIdx3 >= len(ema3.Values) {
			emaIdx3 = len(ema3.Values) - 1
		}
		if obv1Idx >= len(obv1.Values) {
			obv1Idx = len(obv1.Values) - 1
		}
		if obv2Idx >= len(obv2.Values) {
			obv2Idx = len(obv2.Values) - 1
		}

		// 条件1: 收盘价 > EMA1
		if prices[i] > ema1.Values[emaIdx1] {
			cond1[i] = ema1.Values[emaIdx1]
			upMa1 = ema1.Values[emaIdx1]
		} else {
			cond1[i] = upMa1
		}

		// 条件2: OBV[1] > OBV[2]
		if obv2Idx > 1 && obv2.Values[obv2Idx-1] > obv2.Values[obv2Idx-2] {
			cond2[i] = ema1.Values[emaIdx1]
			upMa2 = ema1.Values[emaIdx1]
		} else {
			cond2[i] = upMa2
		}

		// 条件3: 收盘价 > EMA2 且 OBV1[1] > OBV1[2]
		if prices[i] > ema2.Values[emaIdx2] &&
			obv1Idx > 1 && obv1.Values[obv1Idx-1] > obv1.Values[obv1Idx-2] {
			cond3[i] = ema2.Values[emaIdx2]
			upMa3 = ema2.Values[emaIdx2]
		} else {
			cond3[i] = upMa3
		}

		// 条件4: 收盘价 > EMA2
		if prices[i] > ema2.Values[emaIdx2] {
			cond4[i] = ema2.Values[emaIdx2]
			upMa4 = ema2.Values[emaIdx2]
		} else {
			cond4[i] = upMa4
		}

		// 条件5: 收盘价 > EMA3 且 OBV1[1] > OBV1[2]
		if prices[i] > ema3.Values[emaIdx3] &&
			obv1Idx > 1 && obv1.Values[obv1Idx-1] > obv1.Values[obv1Idx-2] {
			cond5[i] = ema3.Values[emaIdx3]
			upMa5 = ema3.Values[emaIdx3]
		} else {
			cond5[i] = upMa5
		}
	}

	return &TaJingZheMA{
		Cond1Values: cond1,
		Cond2Values: cond2,
		Cond3Values: cond3,
		Cond4Values: cond4,
		Cond5Values: cond5,
		Period1:     period1,
		Period2:     period2,
	}, nil
}

// JingZheMA 为K线数据计算惊蛰均线指标
// 说明：
//
//	使用K线数据计算惊蛰均线指标
//	包含多个条件信号线
//
// 参数：
//   - period1: 主要计算周期（默认25）
//   - period2: OBV的EMA计算周期（默认6）
//
// 返回值：
//   - *TaJingZheMA: 包含惊蛰均线计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) JingZheMA(period1, period2 int) (*TaJingZheMA, error) {
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}

	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}

	return CalculateJingZheMA(close, volume, period1, period2)
}

// JingZheMA_ 获取最新的惊蛰均线信号值
// 参数：
//   - period1: 主要计算周期（默认25）
//   - period2: OBV的EMA计算周期（默认6）
//
// 返回值：
//   - float64: 最新的条件1信号值
//   - float64: 最新的条件2信号值
//   - float64: 最新的条件3信号值
//   - float64: 最新的条件4信号值
//   - float64: 最新的条件5信号值
func (k *KlineDatas) JingZheMA_(period1, period2 int) (float64, float64, float64, float64, float64) {
	ma, err := k.JingZheMA(period1, period2)
	if err != nil {
		return 0, 0, 0, 0, 0
	}
	return ma.Value()
}

// Value1 获取条件1最新信号值
// 说明：
//
//	返回条件1的最新信号值（收盘价 > EMA1时的EMA1值或保持前值）
//
// 返回值：
//   - float64: 最新的条件1信号值
//   - float64: 最新的条件2信号值
//   - float64: 最新的条件3信号值
//   - float64: 最新的条件4信号值
//   - float64: 最新的条件5信号值
func (t *TaJingZheMA) Value() (float64, float64, float64, float64, float64) {
	lastIndex := len(t.Cond1Values) - 1
	return t.Cond1Values[lastIndex], t.Cond2Values[lastIndex], t.Cond3Values[lastIndex], t.Cond4Values[lastIndex], t.Cond5Values[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
