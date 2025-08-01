package ta

import (
	"fmt"
)

// TaKDJ 表示随机指标(Stochastic Oscillator)的计算结果
// 说明：
//
//	KDJ是一个超买超卖指标，也叫随机指标：
//	1. K值：快速线，对价格变化反应最敏感
//	2. D值：慢速线，是K值的移动平均
//	3. J值：方向线，反映K值与D值的偏离程度
//	特点：
//	- 取值范围一般在0-100之间（J线可能超出）
//	- 80以上为超买区
//	- 20以下为超卖区
//	- 常用于预测价格走势反转
type TaKDJ struct {
	K []float64 `json:"k"` // K值序列（快速线）
	D []float64 `json:"d"` // D值序列（慢速线）
	J []float64 `json:"j"` // J值序列（方向线）
}

// CalculateKDJ 计算随机指标
// 说明：
//
//	计算步骤：
//	1. 计算RSV值：
//	   RSV = (收盘价 - N日最低价) / (N日最高价 - N日最低价) * 100
//	2. 计算K值：
//	   当日K = (2 * 前日K + 当日RSV) / 3
//	3. 计算D值：
//	   当日D = (2 * 前日D + 当日K) / 3
//	4. 计算J值：
//	   J = 3 * K - 2 * D
//	应用场景：
//	- 判断超买超卖
//	- 预测趋势反转
//	- 寻找背离信号
//
// 参数：
//   - high: 最高价序列
//   - low: 最低价序列
//   - close: 收盘价序列
//   - rsvPeriod: RSV计算周期，通常为9
//   - kPeriod: K值计算周期，通常为3
//   - dPeriod: D值计算周期，通常为3
//
// 返回值：
//   - *TaKDJ: 包含KDJ计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	kdj, err := CalculateKDJ(high, low, close, 9, 3, 3)
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

// KDJ 为K线数据计算随机指标
// 说明：
//
//	对当前K线数据计算KDJ指标
//
// 参数：
//   - rsvPeriod: RSV计算周期
//   - kPeriod: K值计算周期
//   - dPeriod: D值计算周期
//
// 返回值：
//   - *TaKDJ: 包含KDJ计算结果的结构体指针
//   - error: 计算过程中的错误
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

// Value 获取最新的KDJ值
// 说明：
//
//	返回最新的K、D、J三个值
//	使用建议：
//	- K值和D值在80以上，超买信号
//	- K值和D值在20以下，超卖信号
//	- K线上穿D线，金叉买入信号
//	- K线下穿D线，死叉卖出信号
//	- J值大于100或小于0时，反转信号增强
//	- KDJ三线同向，趋势信号最强
//
// 返回值：
//   - k: 最新的K值
//   - d: 最新的D值
//   - j: 最新的J值
func (t *TaKDJ) Value() (k, d, j float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex], t.J[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
