package ta

import (
	"fmt"
)

// TaCMF 表示钱德动量指标(Chaikin Money Flow)的计算结果
// 说明：
//
//	CMF是由Marc Chaikin开发的技术分析工具：
//	1. 结合了价格和成交量的动量指标
//	2. 用于衡量资金流入和流出的强度
//	3. 可以预测价格趋势的持续性
//	特点：
//	- 取值范围在-1到+1之间
//	- 正值表示资金流入（看多）
//	- 负值表示资金流出（看空）
//	- 绝对值越大表示资金流动越强烈
type TaCMF struct {
	Values []float64 `json:"values"` // CMF值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateCMF 计算钱德动量指标
// 说明：
//
//	计算步骤：
//	1. 计算资金流量乘数(MFM)：
//	   MFM = ((收盘价-最低价)-(最高价-收盘价))/(最高价-最低价)
//	2. 计算资金流量(MFV)：
//	   MFV = MFM * 成交量
//	3. 计算CMF：
//	   CMF = N周期MFV之和 / N周期成交量之和
//	使用场景：
//	- 判断主力资金流向
//	- 预测价格趋势持续性
//	- 识别潜在的趋势反转
//
// 参数：
//   - high: 最高价序列
//   - low: 最低价序列
//   - close: 收盘价序列
//   - volume: 成交量序列
//   - period: 计算周期，通常为20或21
//
// 返回值：
//   - *TaCMF: 包含CMF计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	cmf, err := CalculateCMF(high, low, close, volume, 20)
func CalculateCMF(high, low, close, volume []float64, period int) (*TaCMF, error) {
	if len(high) != len(low) || len(high) != len(close) || len(high) != len(volume) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(high) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	mfv := make([]float64, length)
	cmf := make([]float64, length)

	for i := 0; i < length; i++ {
		if high[i] == low[i] {
			mfv[i] = 0
		} else {
			mfm := ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
			mfv[i] = mfm * volume[i]
		}
	}

	for i := period - 1; i < length; i++ {
		sumMFV := 0.0
		sumVolume := 0.0
		for j := 0; j < period; j++ {
			sumMFV += mfv[i-j]
			sumVolume += volume[i-j]
		}
		if sumVolume != 0 {
			cmf[i] = sumMFV / sumVolume
		}
	}

	return &TaCMF{
		Values: cmf,
		Period: period,
	}, nil
}

// CMF 为K线数据计算钱德动量指标
// 说明：
//
//	对当前K线数据计算CMF指标
//
// 参数：
//   - period: 计算周期
//   - source: 价格类型（此参数在CMF计算中实际未使用，保留是为了接口一致性）
//
// 返回值：
//   - *TaCMF: 包含CMF计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) CMF(period int, source string) (*TaCMF, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateCMF(high, low, close, volume, period)
}

// Value 获取最新的CMF值
// 说明：
//
//	返回最新的CMF值
//	使用建议：
//	- CMF > 0.25 表示强势买入信号
//	- CMF < -0.25 表示强势卖出信号
//	- CMF在0线附近徘徊表示盘整
//	- CMF与价格的背离可能预示趋势反转
//	- 连续3个周期保持同向为较强信号
//
// 返回值：
//   - float64: 最新的CMF值
func (t *TaCMF) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
