package ta

// TaMacd 用于计算和存储 MACD 指标相关数据的结构体
// 说明：
//
//	该结构体存储了 MACD 指标的计算结果，包括 MACD 线、DIF 线、DEA 线以及计算所需的周期参数
//
// 字段：
//   - Macd: MACD 线的数据数组 (float64 类型)
//   - Dif: DIF 线的数据数组 (float64 类型)
//   - Dea: DEA 线的数据数组 (float64 类型)
//   - ShortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - LongPeriod: 长期 EMA 计算的周期 (int 类型)
//   - SignalPeriod: 信号线计算的周期 (int 类型)
type TaMacd struct {
	Macd         []float64 `json:"macd"`
	Dif          []float64 `json:"dif"`
	Dea          []float64 `json:"dea"`
	ShortPeriod  int       `json:"short_period"`
	LongPeriod   int       `json:"long_period"`
	SignalPeriod int       `json:"signal_period"`
}

// CalculateMACD 根据给定的价格数据和周期参数计算 MACD 指标
// 参数：
//   - prices: 价格数据数组 (float64 类型)
//   - shortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - longPeriod: 长期 EMA 计算的周期 (int 类型)
//   - signalPeriod: 信号线计算的周期 (int 类型)
//
// 返回值：
//   - *TaMacd: 存储 MACD 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该函数依赖 CalculateEMA 函数计算 EMA 值，若计算 EMA 时出错将直接返回错误
//
// 示例：
//
//	macdResult, err := CalculateMACD(prices, 12, 26, 9)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateMACD(prices []float64, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {

	shortEMA, err := CalculateEMA(prices, shortPeriod)
	if err != nil {
		return nil, err
	}
	longEMA, err := CalculateEMA(prices, longPeriod)
	if err != nil {
		return nil, err
	}

	dif := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		if i < longPeriod-1 {
			dif[i] = 0
		} else {
			dif[i] = shortEMA.Values[i] - longEMA.Values[i]
		}
	}

	dea, err := CalculateEMA(dif, signalPeriod)
	if err != nil {
		return nil, err
	}

	macd := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		macd[i] = 2 * (dif[i] - dea.Values[i]) / 2
	}
	return &TaMacd{
		Macd:         macd,
		Dif:          dif,
		Dea:          dea.Values,
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		SignalPeriod: signalPeriod,
	}, nil
}

// MACD 从 KlineDatas 中提取指定来源的价格数据并计算 MACD 指标
// 参数：
//   - source: 价格数据的来源 (string 类型)
//   - shortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - longPeriod: 长期 EMA 计算的周期 (int 类型)
//   - signalPeriod: 信号线计算的周期 (int 类型)
//
// 返回值：
//   - *TaMacd: 存储 MACD 指标计算结果的结构体指针
//   - error: 提取价格数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法依赖 KlineDatas 的 ExtractSlice 方法提取价格数据，若提取失败将直接返回错误
//
// 示例：
//
//	macdResult, err := k.MACD("close", 12, 26, 9)
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) MACD(source string, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
}

// Value 获取 TaMacd 结构体中 MACD、DIF 和 DEA 线的最后一个值
// 参数：无
// 返回值：
//   - macd: MACD 线的最后一个值 (float64 类型)
//   - dif: DIF 线的最后一个值 (float64 类型)
//   - dea: DEA 线的最后一个值 (float64 类型)
//
// 说明/注意事项：无
// 示例：
//
//	macd, dif, dea := t.Value()
func (t *TaMacd) Value() (macd, dif, dea float64) {
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex], t.Dif[lastIndex], t.Dea[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
