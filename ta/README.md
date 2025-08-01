# 技术分析指标库 (Technical Analysis Library)

这是一个用Go语言实现的技术分析指标库，提供了常用的技术分析指标计算功能。

兼容`go-binance`库K线数据结构。

## 项目结构

- ta.go : 核心数据结构和通用工具函数

- adx.go : ADX (平均趋向指标)
  - CrossOver 检测DI线的交叉信号
- atr.go : ATR(平均真实波幅)
  - Percent 计算ATR相对于当前价格的百分比
- boll.go : BOLL (布林带)
- cci.go : CCI (商品通道指标)
- cmf.go : CMF (钱德动量指标)
- ema.go : EMA (指数移动平均线)
- kdj.go : KDJ (随机指标)
- macd.go : MACD (移动平均趋势指标)
- obv.go : OBV (能量潮指标)
- rma.go : RMA (移动平均)
- rsi.go : RSI (相对强弱指标)
- sma.go : SMA (简单移动平均线)
- stochRsi.go : Stochastic RSI (随机相对强弱指标)
- superTrend.go : SuperTrend (超级趋势指标)
- superTrendPivot.go : SuperTrendPivot (基于轴点的超级趋势指标)
- superTrendPivotHl2.go : SuperTrendPivotHl2 (基于HL2的超级趋势指标)
- t3.go : T3 (Tillson T3移动平均线)
- vr.go : VR (波动率比率指标)
- williamsR.go : Williams %R(威廉指标)

## 使用示例

```go

binanceKline, err := binance.NewFuturesKlinesService().  
    Limit(1000).
    Symbol("BTCUSDT").
    Interval("1H").
    Do(context.Background())
if err != nil {
    log.Fatal(err)
}


kline, err := ta.NewKlineDatas(binanceKline, true)
if err != nil {
    log.Fatal(err)
}


macd, err := kline.MACD("close", 12, 26, 9)
if err != nil {
    log.Fatal(err)
}

rsi, err := kline.RSI(14, "close")
if err != nil {
    log.Fatal(err)
}

atr, err := kline.ATR(14)
if err != nil {
    log.Fatal(err)
}
```

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。