# go-utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL--3.0-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v1.2.1-orange.svg)](https://github.com/phrynus/go-utils/releases)

Go 语言工具库，提供技术分析指标、日志记录、钉钉机器人功能。


## 安装

```bash
go get github.com/phrynus/go-utils
```

## 使用示例

### 技术分析指标

```go
package main

import (
    "context"
    "log"
    
    "github.com/adshao/go-binance/v2/futures"
    "github.com/phrynus/go-utils"
)

func main() {
    // 获取币安K线数据
    client := binance.NewFuturesClient("", "")
    binanceKline, err := client.NewKlinesService().
        Limit(1000).
        Symbol("BTCUSDT").
        Interval("1h").
        Do(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // 转换为工具库格式
    kline, err := utils.NewKlineDatas(binanceKline, true)
    if err != nil {
        log.Fatal(err)
    }

    // 计算技术指标
    macd, _ := kline.MACD("close", 12, 26, 9)
    rsi, _ := kline.RSI(14, "close")
    atr, _ := kline.ATR(14)
}
```

### 日志记录

```go
package main

import "github.com/phrynus/go-utils"

func main() {
    config := &utils.LogConfig{
        Level: utils.INFO,
        Color: true,
    }
    logger := utils.NewLogger(config)
    
    logger.Info("应用程序启动")
    logger.Error("发生错误")
}
```

### 钉钉机器人

```go
package main

import "github.com/phrynus/go-utils"

func main() {
    dt := utils.NewDingtalk("your_access_token").WithSecret("your_secret")
    
    // 发送文本消息
    err := dt.SendText("Hello, DingTalk!", nil)
    
    // 发送Markdown消息
    title := "系统通知"
    text := "## 系统维护通知\n请提前做好准备！"
    at := &utils.AtMeta{IsAtAll: true}
    
    err = dt.SendMarkdown(title, text, at)
}
```

## 版本发布

```bash
git tag v1.2.2
git push origin --tags
```

## 许可证

本项目采用 GNU Affero General Public License v3.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。数字货币和合约交易具有高风险，请谨慎使用。
