# go-utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL--3.0-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v1.2.1-orange.svg)](https://github.com/phrynus/go-utils/releases)

Go 语言工具库，提供技术分析指标、日志记录、钉钉机器人、飞书机器人功能。


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
    // 创建日志记录器
    log, err = utils.NewLogger(utils.LogConfig{
        Filename: "main.log", // log filename
        LogDir:   "logs",     // log directory
        MaxSize:  50 * 1024,  // KB
        StdoutLevels: map[int]bool{
        utils.INFO:  true,
        utils.DEBUG: false,
        utils.WARN:  true,
        utils.ERROR: true,
        },
        ColorOutput:  true,
        ShowFileLine: true,
    })
    if err != nil {
        panic(err)
    }
    
    // 使用 defer 确保程序退出时关闭日志
    defer func() {
        if err := log.Close(); err != nil {
            // 处理关闭错误
            fmt.Printf("关闭日志记录器失败: %v\n", err)
        }
    }()
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

### 飞书机器人

```go
package main

import "github.com/phrynus/go-utils"

func main() {
    // 创建飞书客户端（使用webhook URL和密钥）
    fs := utils.NewFeiShu("https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id").
        WithSecret("your_secret")
    
    // 发送文本消息
    err := fs.SendText("Hello, FeiShu! 这是一条测试消息。")
    
    // 发送富文本消息
    post := &utils.FsPost{
        ZhCn: &utils.FsPostDetail{
            Title: "系统监控告警",
            Content: [][]utils.FsPostElem{
                {
                    {Tag: "text", Text: "告警时间: 2024-01-01 14:30:00"},
                },
                {
                    {Tag: "text", Text: "告警级别: "},
                    {Tag: "text", Text: "严重"},
                },
                {
                    {Tag: "text", Text: "查看详情: "},
                    {Tag: "a", Text: "点击这里", Href: "https://example.com"},
                },
                {
                    {Tag: "at", UserId: "all"}, // @所有人
                    {Tag: "text", Text: " 请及时处理！"},
                },
            },
        },
    }
    
    err = fs.SendPost(post)
}
```

## 版本发布

```bash
git tag v1.3.0
git push origin --tags
```

## 许可证

本项目采用 GNU Affero General Public License v3.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。数字货币和合约交易具有高风险，请谨慎使用。
