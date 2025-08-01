package ta

import (
	"fmt"
)

// TaKDJ 表示 KDJ 指标的计算结果结构体
// 说明：
//
//	该结构体用于存储 KDJ 指标的计算结果，包含 K、D、J 三条线的值
//
// 字段：
//   - K: K 线的值数组 (float64 类型)
//   - D: D 线的值数组 (float64 类型)
//   - J: J 线的值数组 (float64 类型)
type TaKDJ struct {
	K []float64 `json:"k"`
	D []float64 `json:"d"`
	J []float64 `json:"j"`
}

// CalculateKDJ 计算 KDJ 指标
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - close: 收盘价数组
//   - rsvPeriod: RSV 的计算周期
//   - kPeriod: K 线的计算周期
//   - dPeriod: D 线的计算周期
//
// 返回值：
//   - *TaKDJ: 包含 KDJ 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	输入的 high、low、close 数组长度必须不小于 rsvPeriod，否则将返回错误
//
// 示例：
//
//	high := []float64{...}
//	low := []float64{...}
//	close := []float64{...}
//	kdj, err := CalculateKDJ(high, low, close, 9, 3, 3)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateKDJ(high, low, close []float64, rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
	if len(high) < rsvPeriod || len(low) < rsvPeriod || len(close) < rsvPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 4)
	rsv, k, d, j := slices[0], slices[1], slices[2], slices[3]

	for i := rsvPeriod - 1; i < length; i++ {

		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < rsvPeriod; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		if highestHigh != lowestLow {
			rsv[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
		} else {
			rsv[i] = 50
		}
	}

	k[rsvPeriod-1] = rsv[rsvPeriod-1]
	d[rsvPeriod-1] = rsv[rsvPeriod-1]
	j[rsvPeriod-1] = rsv[rsvPeriod-1]

	for i := rsvPeriod; i < length; i++ {

		k[i] = (2.0*k[i-1] + rsv[i]) / 3.0

		d[i] = (2.0*d[i-1] + k[i]) / 3.0

		j[i] = 3.0*k[i] - 2.0*d[i]
	}

	return &TaKDJ{
		K: k,
		D: d,
		J: j,
	}, nil
}

// KDJ 计算 K 线数据的 KDJ 指标
// 参数：
//   - rsvPeriod: RSV 的计算周期
//   - kPeriod: K 线的计算周期
//   - dPeriod: D 线的计算周期
//
// 返回值：
//   - *TaKDJ: 包含 KDJ 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法会从 KlineDatas 中提取 high、low、close 数据进行计算
//
// 示例：
//
//	klineData := &KlineDatas{...}
//	kdj, err := klineData.KDJ(9, 3, 3)
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) KDJ(rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
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
	return CalculateKDJ(high, low, close, rsvPeriod, kPeriod, dPeriod)
}

// Value 获取 TaKDJ 结构体中 K、D、J 线的最后一个值
// 返回值：
//   - k: K 线的最后一个值
//   - d: D 线的最后一个值
//   - j: J 线的最后一个值
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	k, d, j := kdj.Value()
func (t *TaKDJ) Value() (k, d, j float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex], t.J[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
