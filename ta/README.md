# 技术分析指标库 (Technical Analysis Library)

这是一个用Go语言实现的技术分析指标库，提供了常用的技术分析指标计算功能。

## 功能特性

- ✅ 支持多种技术分析指标
- ✅ 兼容 `go-binance` 库K线数据结构
- ✅ 支持自动识别多种K线数据格式
- ✅ 高性能并发处理（大数据量时自动启用）
- ✅ 支持动态添加K线数据

## 安装

```bash
go get github.com/phrynus/go-utils/ta
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/adshao/go-binance/v2/futures"
    "github.com/phrynus/go-utils/ta"
)

func main() {
    // 获取币安K线数据
    client := futures.NewClient("", "")
    binanceKline, err := client.NewKlinesService().
        Limit(1000).
        Symbol("BTCUSDT").
        Interval("1h").
        Do(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // 转换为工具库格式
    // 第二个参数 true 表示排除最后一根K线（通常未完成）
    kline, err := ta.NewKlineDatas(binanceKline, true)
    if err != nil {
        log.Fatal(err)
    }

    // 计算技术指标
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
}
```

### 动态添加K线数据

```go
// 从WebSocket接收新的K线数据并添加
wsKline := &binance.WsKline{
    StartTime: 1234567890000,
    Open:      "50000",
    High:      "51000",
    Low:       "49000",
    Close:     "50500",
    Volume:    "1000",
}

err := kline.Add(wsKline)
if err != nil {
    log.Fatal(err)
}
```

### 提取价格序列

```go
// 提取收盘价序列
closePrices, err := kline.ExtractSlice("close")
if err != nil {
    log.Fatal(err)
}

// 支持的类型：open, high, low, close, volume
openPrices, _ := kline.ExtractSlice("open")
highPrices, _ := kline.ExtractSlice("high")
lowPrices, _ := kline.ExtractSlice("low")
volume, _ := kline.ExtractSlice("volume")
```

## 支持的指标

### 趋势指标

#### SMA (简单移动平均线)

```go
sma, err := kline.SMA(period, "close")
// period: 周期，例如 20, 50, 200
```

#### EMA (指数移动平均线)

```go
ema, err := kline.EMA(period, "close")
// period: 周期，例如 12, 26, 50
```

#### RMA (移动平均)

```go
rma, err := kline.RMA(period, "close")
```

#### T3 (Tillson T3移动平均线)

```go
t3, err := kline.T3(period, "close")
```

#### JingZheMA (惊蛰均线)

```go
jzma, err := kline.JingZheMA(period, factor)
// period: 周期
// factor: 因子
cond1, cond2, cond3, cond4, cond5 := jzma.Value()
directionNum := jzma.DirectionNum()
```

### 动量指标

#### MACD (移动平均趋势指标)

```go
macd, err := kline.MACD("close", fastPeriod, slowPeriod, signalPeriod)
// fastPeriod: 快线周期，默认 12
// slowPeriod: 慢线周期，默认 26
// signalPeriod: 信号线周期，默认 9
```

#### RSI (相对强弱指标)

```go
rsi, err := kline.RSI(period, "close")
// period: 周期，常用 14
```

#### Stochastic RSI (随机相对强弱指标)

```go
stochRsi, err := kline.StochRSI(period, smoothK, smoothD)
```

#### KDJ (随机指标)

```go
kdj, err := kline.KDJ(period, kPeriod, dPeriod)
```

#### Williams %R (威廉指标)

```go
williamsR, err := kline.WilliamsR(period)
```

#### CCI (商品通道指标)

```go
cci, err := kline.CCI(period)
```

#### CMF (钱德动量指标)

```go
cmf, err := kline.CMF(period)
```

#### DPO (偏离价格振荡器)

```go
dpo, err := kline.DPO(period)
```

#### ADX (平均趋向指标)

```go
adx, err := kline.ADX(period)
// 检测DI线交叉
isCrossOver := adx.CrossOver()
```

### 波动率指标

#### ATR (平均真实波幅)

```go
atr, err := kline.ATR(period)
// 计算ATR相对于当前价格的百分比
percent := atr.Percent()
```

#### BOLL (布林带)

```go
boll, err := kline.BOLL(period, stdDev)
// period: 周期，常用 20
// stdDev: 标准差倍数，常用 2
```

#### VR (波动率比率指标)

```go
vr, err := kline.VR(period)
```

### 成交量指标

#### OBV (能量潮指标)

```go
obv, err := kline.OBV()
```

### 趋势指标（高级）

#### SuperTrend (超级趋势指标)

```go
superTrend, err := kline.SuperTrend(period, multiplier)
```

#### SuperTrendPivot (基于轴点的超级趋势指标)

```go
superTrendPivot, err := kline.SuperTrendPivot(period, multiplier)
```

#### SuperTrendPivotHl2 (基于HL2的超级趋势指标)

```go
superTrendPivotHl2, err := kline.SuperTrendPivotHl2(period, multiplier)
```

## 项目结构

- `ta.go`: 核心数据结构和通用工具函数
- `adx.go`: ADX (平均趋向指标)
  - `CrossOver()`: 检测DI线的交叉信号
- `atr.go`: ATR (平均真实波幅)
  - `Percent()`: 计算ATR相对于当前价格的百分比
- `boll.go`: BOLL (布林带)
- `cci.go`: CCI (商品通道指标)
- `cmf.go`: CMF (钱德动量指标)
- `dpo.go`: DPO (偏离价格振荡器)
- `ema.go`: EMA (指数移动平均线)
- `jingzheMA.go`: JingZheMA (惊蛰均线)
- `kdj.go`: KDJ (随机指标)
- `kline.go`: K线数据操作方法
- `macd.go`: MACD (移动平均趋势指标)
- `obv.go`: OBV (能量潮指标)
- `rma.go`: RMA (移动平均)
- `rsi.go`: RSI (相对强弱指标)
- `sma.go`: SMA (简单移动平均线)
- `stochRsi.go`: Stochastic RSI (随机相对强弱指标)
- `superTrend.go`: SuperTrend (超级趋势指标)
- `superTrendPivot.go`: SuperTrendPivot (基于轴点的超级趋势指标)
- `superTrendPivotHl2.go`: SuperTrendPivotHl2 (基于HL2的超级趋势指标)
- `t3.go`: T3 (Tillson T3移动平均线)
- `vr.go`: VR (波动率比率指标)
- `williamsR.go`: Williams %R (威廉指标)

## 数据格式支持

库支持多种K线数据格式，会自动识别以下字段名：

- **时间字段**: `StartTime`, `OpenTime`, `Time`, `t`, `T`, `Timestamp`, `OpenAt`, `EventTime`
- **开盘价**: `Open`, `OpenPrice`, `O`, `o`
- **最高价**: `High`, `HighPrice`, `H`, `h`
- **最低价**: `Low`, `LowPrice`, `L`, `l`
- **收盘价**: `Close`, `ClosePrice`, `C`, `c`
- **成交量**: `Volume`, `Vol`, `V`, `v`, `Amount`, `Quantity`

支持的数据类型：

- 时间字段：`int64`（毫秒）或字符串格式的时间戳
- 价格字段：`float64` 或字符串格式的数字

## 完整示例

完整示例请查看 [example/ta.go](../example/ta.go)

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/adshao/go-binance/v2/futures"
    "github.com/phrynus/go-utils/ta"
)

func main() {
    // 创建币安客户端
    futuresClient := futures.NewClient(
        os.Getenv("BINANCE_API_KEY"),
        os.Getenv("BINANCE_SECRET_KEY"),
    )

    // 获取K线数据
    klines, err := futuresClient.NewKlinesService().
        Symbol("ETHUSDT").
        Interval("30m").
        Limit(1000).
        Do(context.Background())
    if err != nil {
        fmt.Printf("获取K线失败: %v\n", err)
        return
    }

    // 转换为工具库格式
    klineDatas, err := ta.NewKlineDatas(klines, true)
    if err != nil {
        fmt.Printf("转换K线数据失败: %v", err)
        return
    }

    // 计算多个指标
    ema, _ := klineDatas.EMA(25, "close")
    rsi, _ := klineDatas.RSI(14, "close")
    macd, _ := klineDatas.MACD("close", 12, 26, 9)
    atr, _ := klineDatas.ATR(14)
    obv, _ := klineDatas.OBV()
    
    // 惊蛰均线
    jingzhema, _ := klineDatas.JingZheMA(25, 6)
    cond1, cond2, cond3, cond4, cond5 := jingzhema.Value()
    directionNum := jingzhema.DirectionNum()

    // 输出结果
    lastTime := time.Unix(0, klineDatas[len(klineDatas)-1].StartTime*int64(time.Millisecond))
    fmt.Printf("最后一根K线时间: %s\n", lastTime.Format("2006-01-02 15:04:05"))
    fmt.Printf("EMA(25): %v\n", ema.Value())
    fmt.Printf("RSI(14): %v\n", rsi.Value())
    fmt.Printf("MACD: %v\n", macd.Value())
    fmt.Printf("ATR(14): %v\n", atr.Value())
    fmt.Printf("OBV: %v\n", obv.Value())
    fmt.Printf("JingZheMA: %v, %v, %v, %v, %v\n", cond1, cond2, cond3, cond4, cond5)
    fmt.Printf("JingZheMA方向数: %v\n", directionNum)
}
```

## API 参考

### 核心类型

- `KlineData`: K线数据结构
- `KlineDatas`: K线数据切片类型

### 核心方法

- `NewKlineDatas(klines interface{}, excludeLast bool) (KlineDatas, error)`: 创建K线数据集合
- `Add(wsKline interface{}) error`: 添加新的K线数据
- `ExtractSlice(priceType string) ([]float64, error)`: 提取价格序列

### 指标方法

所有指标方法都返回对应的指标对象，可以通过 `Value()` 方法获取最新值。

## 性能优化

- 当K线数据量超过 1000 根时，自动启用并发处理
- 使用字段缓存机制，避免重复反射操作
- 支持多种数据格式的自动识别和转换

## 注意事项

1. **数据量要求**：大多数指标需要足够的历史数据才能计算，建议至少提供指标周期的 2-3 倍数据量。

2. **最后一根K线**：`NewKlineDatas` 的第二个参数 `excludeLast` 用于排除最后一根K线，因为最后一根K线可能还未完成。

3. **价格类型**：提取价格序列时，`priceType` 参数支持：`"open"`, `"high"`, `"low"`, `"close"`, `"volume"`。

4. **错误处理**：所有指标计算方法都可能返回错误，请务必检查错误。

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。
