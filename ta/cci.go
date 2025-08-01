package ta

import (
	"fmt"
	"math"
)

// TaCCI 计算商品通道指数（CCI）的结构体
// 说明：
//
//	该结构体用于存储 CCI 计算结果，其中 Values 字段存储每个时间点的 CCI 值
//
// 字段：
//   - Values: 存储 CCI 计算结果的切片 (float64 类型)
type TaCCI struct {
	Values []float64 `json:"values"`
}

// CalculateCCI 根据 K 线数据计算商品通道指数（CCI）
// 参数：
//   - klineData: K 线数据切片 (KlineDatas 类型)
//   - period: 计算周期 (int 类型)
//
// 返回值：
//   - *TaCCI: 存储 CCI 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	当输入的 K 线数据长度小于计算周期时，会返回错误。该函数首先计算典型价格，然后计算简单移动平均和平均绝对偏差，最后计算 CCI 值。
//
// 示例：
//
//	result, err := CalculateCCI(klineData, 20)
//	if err != nil {
//	    // 处理错误
//	}
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

// CCI 为 KlineDatas 结构体添加的 CCI 计算方法
// 参数：
//   - period: 计算周期 (int 类型)
//
// 返回值：
//   - *TaCCI: 存储 CCI 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法调用 CalculateCCI 函数进行 CCI 计算。
//
// 示例：
//
//	result, err := k.CCI(20)
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) CCI(period int) (*TaCCI, error) {
	return CalculateCCI(*k, period)
}

// Value 获取 TaCCI 结构体中最后一个 CCI 值
// 返回值：
//   - float64: 最后一个 CCI 值
//
// 说明/注意事项：
//
//	该方法用于获取最新的 CCI 值。
//
// 示例：
//
//	value := t.Value()
func (t *TaCCI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
