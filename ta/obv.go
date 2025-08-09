package ta

import (
	"fmt"
)

// TaOBV 表示能量潮指标(On Balance Volume)的计算结果
// 说明：
//
//	OBV是由Joe Granville开发的成交量指标：
//	1. 通过成交量来确认价格趋势
//	2. 判断成交量是否支撑价格走势
//	3. 可以预测价格突破的可能性
//	特点：
//	- 将成交量的变化与价格方向相结合
//	- 可以发现量价关系的背离
//	- 帮助判断趋势的强弱
//	- 适合寻找主力资金进出的迹象
type TaOBV struct {
	Values []float64 `json:"values"` // OBV值序列
}

// CalculateOBV 计算能量潮指标
// 说明：
//
//	计算规则：
//	1. 当收盘价上涨时：
//	   当日OBV = 前一日OBV + 当日成交量
//	2. 当收盘价下跌时：
//	   当日OBV = 前一日OBV - 当日成交量
//	3. 当收盘价不变时：
//	   当日OBV = 前一日OBV
//	使用场景：
//	- 判断量价配合程度
//	- 预测价格突破方向
//	- 发现趋势的强弱
//	- 识别主力资金动向
//
// 参数：
//   - prices: 价格序列（通常使用收盘价）
//   - volumes: 成交量序列
//
// 返回值：
//   - *TaOBV: 包含OBV计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	obv, err := CalculateOBV(prices, volumes)
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

// OBV 为K线数据计算能量潮指标
// 说明：
//
//	使用收盘价和成交量计算OBV指标
//
// 返回值：
//   - *TaOBV: 包含OBV计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) OBV() (*TaOBV, error) {
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

// OBV_ 获取最新的OBV值
// 返回值：
//   - float64: 最新的OBV值
func (k *KlineDatas) OBV_() float64 {
	obv, err := k.OBV()
	if err != nil {
		return 0
	}
	return obv.Value()
}

// Value 获取最新的OBV值
// 说明：
//
//	返回最新的OBV值
//	使用建议：
//	- OBV上升，表明买方力量占优
//	- OBV下降，表明卖方力量占优
//	- OBV与价格同向变动，趋势更可靠
//	- 背离信号：
//	  * 顶背离：价格创新高但OBV未创新高，可能即将下跌
//	  * 底背离：价格创新低但OBV未创新低，可能即将上涨
//
// 返回值：
//   - float64: 最新的OBV值
func (t *TaOBV) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
