package ta

import (
	"fmt"
	"math"
)

// TaCCI 表示商品通道指标(Commodity Channel Index)的计算结果
// 说明：
//
//	CCI是由Donald Lambert开发的技术分析工具：
//	1. 测量价格是否偏离其统计平均水平
//	2. 用于识别周期性的超买超卖水平
//	3. 可以预测价格趋势的反转
//	特点：
//	- 取值范围理论上无限，但通常在±100之间波动
//	- +100以上为超买区
//	- -100以下为超卖区
//	- 数值的绝对值越大，价格偏离度越高
type TaCCI struct {
	Values []float64 `json:"values"` // CCI值序列
}

// CalculateCCI 计算商品通道指标
// 说明：
//
//	计算步骤：
//	1. 计算典型价格(TP)：
//	   TP = (最高价 + 最低价 + 收盘价) / 3
//	2. 计算TP的N周期简单移动平均(SMA)
//	3. 计算平均偏差(MD)：
//	   MD = TP与其SMA的差值的N周期平均
//	4. 计算CCI：
//	   CCI = (TP - SMA) / (0.015 * MD)
//	其中0.015是Lambert选择的常数
//
// 参数：
//   - klineData: K线数据
//   - period: 计算周期，通常为20
//
// 返回值：
//   - *TaCCI: 包含CCI计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	cci, err := CalculateCCI(klineData, 20)
func CalculateCCI(klineData KlineDatas, period int) (*TaCCI, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	typicalPrice, cci := slices[0], slices[1]

	for i := 0; i < length; i++ {
		typicalPrice[i] = (klineData[i].High + klineData[i].Low + klineData[i].Close) / 3
	}

	for i := period - 1; i < length; i++ {

		var sumTP float64
		for j := i - period + 1; j <= i; j++ {
			sumTP += typicalPrice[j]
		}
		smaTP := sumTP / float64(period)

		var sumAbsDev float64
		for j := i - period + 1; j <= i; j++ {
			sumAbsDev += math.Abs(typicalPrice[j] - smaTP)
		}
		meanDeviation := sumAbsDev / float64(period)

		if meanDeviation != 0 {
			cci[i] = (typicalPrice[i] - smaTP) / (0.015 * meanDeviation)
		}
	}

	return &TaCCI{
		Values: cci,
	}, nil
}

// CCI 为K线数据计算商品通道指标
// 说明：
//
//	对当前K线数据计算CCI指标
//
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaCCI: 包含CCI计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) CCI(period int) (*TaCCI, error) {
	return CalculateCCI(*k, period)
}

// CCI_ 获取最新的CCI值
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - float64: 最新的CCI值
func (k *KlineDatas) CCI_(period int) float64 {
	cci, err := k.CCI(period)
	if err != nil {
		return 0
	}
	return cci.Value()
}

// Value 获取最新的CCI值
// 说明：
//
//	返回最新的CCI值
//	使用建议：
//	- CCI > +100 考虑卖出（超买）
//	- CCI < -100 考虑买入（超卖）
//	- CCI从超买区下穿+100，卖出信号增强
//	- CCI从超卖区上穿-100，买入信号增强
//	- CCI的背离可以预示趋势反转
//
// 返回值：
//   - float64: 最新的CCI值
func (t *TaCCI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
