package ta

// TaMacd 表示移动平均趋同/背离指标(Moving Average Convergence/Divergence)的计算结果
// 说明：
//
//	MACD是由Gerald Appel开发的趋势跟踪指标：
//	1. DIF(差离值)：短期与长期EMA的差值
//	2. DEA(信号线)：DIF的移动平均
//	3. MACD柱：DIF与DEA的差值
//	特点：
//	- 能够同时显示趋势方向和强度
//	- 可以识别市场的超买超卖状态
//	- 适合中长期趋势交易
//	- 能够发现趋势的背离现象
type TaMacd struct {
	Macd         []float64 `json:"macd"`          // MACD柱状值序列
	Dif          []float64 `json:"dif"`           // 差离值序列
	Dea          []float64 `json:"dea"`           // 信号线序列
	ShortPeriod  int       `json:"short_period"`  // 短期EMA周期
	LongPeriod   int       `json:"long_period"`   // 长期EMA周期
	SignalPeriod int       `json:"signal_period"` // 信号线周期
}

// CalculateMACD 计算MACD指标
// 说明：
//
//	计算步骤：
//	1. 计算短期和长期EMA：
//	   - 短期EMA（通常12日）
//	   - 长期EMA（通常26日）
//	2. 计算DIF：
//	   DIF = 短期EMA - 长期EMA
//	3. 计算DEA：
//	   DEA = DIF的N日EMA（通常9日）
//	4. 计算MACD柱：
//	   MACD = 2 * (DIF - DEA)
//	使用场景：
//	- 判断趋势方向和强度
//	- 寻找买卖点
//	- 发现趋势背离
//
// 参数：
//   - prices: 价格序列
//   - shortPeriod: 短期EMA周期，通常为12
//   - longPeriod: 长期EMA周期，通常为26
//   - signalPeriod: 信号线周期，通常为9
//
// 返回值：
//   - *TaMacd: 包含MACD计算结果的结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 示例：
//
//	macd, err := CalculateMACD(prices, 12, 26, 9)
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

// MACD 为K线数据计算MACD指标
// 说明：
//
//	对指定价格类型计算MACD指标
//
// 参数：
//   - source: 价格类型，支持"open"、"high"、"low"、"close"等
//   - shortPeriod: 短期EMA周期
//   - longPeriod: 长期EMA周期
//   - signalPeriod: 信号线周期
//
// 返回值：
//   - *TaMacd: 包含MACD计算结果的结构体指针
//   - error: 计算过程中的错误
func (k *KlineDatas) MACD(source string, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
}

// Value 获取最新的MACD值
// 说明：
//
//	返回最新的MACD、DIF和DEA值
//	使用建议：
//	- DIF上穿DEA，形成金叉，买入信号
//	- DIF下穿DEA，形成死叉，卖出信号
//	- MACD柱由负变正，趋势向上转换
//	- MACD柱由正变负，趋势向下转换
//	- DIF与价格的背离预示趋势可能反转：
//	  * 顶背离：价格创新高但DIF未创新高
//	  * 底背离：价格创新低但DIF未创新低
//
// 返回值：
//   - macd: 最新的MACD柱值
//   - dif: 最新的DIF值
//   - dea: 最新的DEA值
func (t *TaMacd) Value() (macd, dif, dea float64) {
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex], t.Dif[lastIndex], t.Dea[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
