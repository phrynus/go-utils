package ta

import (
	"fmt"
)

// TaOBV 用于存储 OBV 指标计算结果的结构体
// 说明：
//
//	该结构体用于存储 OBV 指标的计算结果，`Values` 字段保存了每个时间点的 OBV 值。
//
// 字段：
//   - Values: 存储 OBV 指标值的切片 (float64 类型)
type TaOBV struct {
	Values []float64 `json:"values"`
}

// CalculateOBV 计算 OBV 指标值
// 参数：
//   - prices: 价格数据切片 (float64 类型)
//   - volumes: 成交量数据切片 (float64 类型)
//
// 返回值：
//   - *TaOBV: 存储 OBV 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//   - 输入的 `prices` 和 `volumes` 切片长度必须一致，否则会返回错误。
//   - 输入数据长度至少为 2，否则会返回错误。
//
// 示例：
//
//	obv, err := CalculateOBV(prices, volumes)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateOBV(prices, volumes []float64) (*TaOBV, error) {
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(prices) < 2 {
		return nil, fmt.Errorf("计算数据不足")
	}

	obv := make([]float64, len(prices))
	obv[0] = volumes[0]

	for i := 1; i < len(prices); i++ {
		if prices[i] > prices[i-1] {
			obv[i] = obv[i-1] + volumes[i]
		} else if prices[i] < prices[i-1] {
			obv[i] = obv[i-1] - volumes[i]
		} else {
			obv[i] = obv[i-1]
		}
	}

	return &TaOBV{
		Values: obv,
	}, nil
}

// OBV 从 KlineDatas 中提取收盘价和成交量数据并计算 OBV 指标值
// 参数：
//   - source: 数据源标识 (string 类型)
//
// 返回值：
//   - *TaOBV: 存储 OBV 指标计算结果的结构体指针
//   - error: 提取数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//   - 该方法会从 `KlineDatas` 中提取 `close` 和 `volume` 数据进行计算。
//   - 若提取数据过程中出现错误，会返回相应的错误信息。
func (k *KlineDatas) OBV(source string) (*TaOBV, error) {
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateOBV(close, volume)
}

// Value 获取 TaOBV 结构体中 OBV 指标的最后一个值
// 返回值：
//   - float64: OBV 指标的最后一个值
//
// 说明/注意事项：
//   - 若 `Values` 切片为空，可能会导致数组越界错误。
func (t *TaOBV) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
