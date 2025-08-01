package ta

import (
	"fmt"
)

// TaCMF 用于计算资金流量指标（Chaikin Money Flow）的结构体
// 说明：
//
//	该结构体存储了计算得到的 CMF 值以及计算时使用的周期
//
// 字段：
//   - Values: 存储计算得到的 CMF 值的切片 (float64 类型)
//   - Period: 计算 CMF 时使用的周期 (int 类型)
type TaCMF struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateCMF 计算资金流量指标（Chaikin Money Flow）
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - close: 收盘价数组
//   - volume: 成交量数组
//   - period: 计算周期
//
// 返回值：
//   - *TaCMF: 存储计算结果的 TaCMF 结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	输入的 high、low、close 和 volume 数组长度必须一致
//	输入数据长度必须大于等于计算周期
//
// 示例：
//
//	cmf, err := CalculateCMF(highPrices, lowPrices, closePrices, volumes, 20)
//	if err != nil {
//	    // 处理错误
//	}
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

// CMF 从 KlineDatas 结构体中提取数据并计算资金流量指标（Chaikin Money Flow）
// 参数：
//   - period: 计算周期
//   - source: 数据源标识
//
// 返回值：
//   - *TaCMF: 存储计算结果的 TaCMF 结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法会调用 KlineDatas 的 ExtractSlice 方法提取数据
//	提取数据过程中可能会出现错误
//
// 示例：
//
//	cmf, err := klineDatas.CMF(20, "source")
//	if err != nil {
//	    // 处理错误
//	}
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

// Value 获取 TaCMF 结构体中最后一个 CMF 值
// 返回值：
//   - float64: 最后一个 CMF 值
//
// 示例：
//
//	value := cmf.Value()
func (t *TaCMF) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
