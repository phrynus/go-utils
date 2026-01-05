# 技术分析指标库

用Go语言实现的技术分析指标库，提供常用的技术分析指标计算功能。

## 功能特性

- ✅ 支持20+种技术分析指标
- ✅ 兼容 `go-binance` 库K线数据结构
- ✅ 自动识别多种K线数据格式（结构体/数组）
- ✅ 高性能并发处理（大数据量时自动启用）
- ✅ 支持动态添加K线数据

## 安装

```bash
go get github.com/phrynus/go-utils/ta
```

## 支持的指标

核心指标文件：

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

自动识别多种K线数据格式，支持结构体和数组两种格式：

### 字段识别

- **时间**: `StartTime`, `OpenTime`, `Time`, `Timestamp` 等
- **价格**: `Open`, `High`, `Low`, `Close`, `Volume` 等
- **支持类型**: `float64`, `int64`, 字符串数字，`interface{}` 包装类型



### 数组格式
通过数字索引指定字段位置，默认顺序：`[时间, 开盘价, 最高价, 最低价, 收盘价, 成交量]`

## 快速开始

完整示例请查看 [example/ta.go](../example/ta.go)

```go
package main

import (
    "fmt"
    "github.com/phrynus/go-utils/ta"
)

func main() {
    // 从币安或其他数据源获取K线数据
    // klines := 获取的K线数据

    // 转换为工具库格式（自动识别字段）
    klineDatas, err := ta.NewKlineDatas(klines, true) // true表示排除最后一根未完成K线
    if err != nil {
        fmt.Printf("转换失败: %v", err)
        return
    }

    // 计算指标
    ema, _ := klineDatas.EMA(25, "close")
    rsi, _ := klineDatas.RSI(14, "close")
    macd, _ := klineDatas.MACD("close", 12, 26, 9)

    // 获取最新值
    fmt.Printf("EMA(25): %.2f\n", ema.Value())
    fmt.Printf("RSI(14): %.2f\n", rsi.Value())
    fmt.Printf("MACD: %+v\n", macd.Value())

    // 动态添加新K线
    newKline := []interface{}{timestamp, open, high, low, close, volume}
    klineDatas.Add(newKline)
}
```


## API 概览



### 核心类型

- `KlineDatas`: K线数据集合
- `FieldNames`: 自定义字段名称配置



### 核心方法
- `NewKlineDatas(klines interface{}, excludeLast bool, customFields ...*FieldNames) (KlineDatas, error)`: 创建K线数据集合
- `Add(kline interface{}, customFields ...*FieldNames) error`: 添加新K线数据
- `ExtractSlice(priceType string) ([]float64, error)`: 提取价格序列（支持 "open", "high", "low", "close", "volume"）

所有指标方法都返回对应的指标对象，通过 `Value()` 方法获取最新值。

## 性能优化

- 当K线数据量超过 1000 根时，自动启用并发处理
- 使用字段缓存机制，避免重复反射操作
- 支持多种数据格式的自动识别和转换

## 注意事项

- **数据量要求**: 建议提供至少指标周期2-3倍的历史数据
- **最后一根K线**: 使用 `excludeLast=true` 排除可能未完成的最后一根K线
- **价格类型**: 支持 `"open"`, `"high"`, `"low"`, `"close"`, `"volume"`
- **错误处理**: 所有指标方法都可能返回错误，务必检查
- **自定义字段**: 可通过 `FieldNames` 配置自定义字段名称，`Add` 方法需使用相同配置
- **数组格式**: 通过数字索引指定字段位置，默认顺序为 `[时间, 开盘, 最高, 最低, 收盘, 成交量]`

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。
