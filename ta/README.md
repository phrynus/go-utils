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

## 注意事项

- **数据量要求**: 建议提供至少指标周期2-3倍的历史数据

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。
