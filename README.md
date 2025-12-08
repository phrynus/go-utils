# go-utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL--3.0-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v1.4.0-orange.svg)](https://github.com/phrynus/go-utils/releases)

Go 语言工具库，提供技术分析指标、日志记录、钉钉机器人、飞书机器人、用户API客户端等功能。

## 功能模块

- **[ta](./ta/)** - 技术分析指标库，提供多种技术指标计算（MACD、RSI、KDJ、布林带等）
- **[logger](./logger/)** - 日志记录器，支持日志轮转、压缩、彩色输出、多级别日志
- **[dingtalk](./dingtalk/)** - 钉钉机器人客户端，支持发送文本、Markdown、链接、ActionCard、FeedCard等消息
- **[feishu](./feishu/)** - 飞书机器人客户端，支持发送文本、富文本、图片、分享群名片、消息卡片等
- **[uyz-u](./uyz-u/)** - 用户API客户端，支持加密通信、签名验证、登录、支付等功能

## 安装

```bash
go get github.com/phrynus/go-utils
```

## 快速开始

### 技术分析指标

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
    kline, err := ta.NewKlineDatas(binanceKline, true)
    if err != nil {
        log.Fatal(err)
    }

    // 计算技术指标
    macd, _ := kline.MACD("close", 12, 26, 9)
    rsi, _ := kline.RSI(14, "close")
    atr, _ := kline.ATR(14)
}
```

更多示例请查看 [ta/README.md](./ta/README.md)

### 日志记录

```go
package main

import (
    "fmt"
    "github.com/phrynus/go-utils/logger"
)

func main() {
    // 创建日志记录器
    log, err := logger.NewLogger(logger.LogConfig{
        Filename: "main.log", // log filename
        LogDir:   "logs",     // log directory
        MaxSize:  50 * 1024,  // KB
        StdoutLevels: map[int]bool{
            logger.INFO:  true,
            logger.DEBUG: false,
            logger.WARN:  true,
            logger.ERROR: true,
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
            fmt.Printf("关闭日志记录器失败: %v\n", err)
        }
    }()
    
    // 使用日志
    log.Info("这是一条信息日志")
    log.Debugf("调试信息: %s", "value")
    log.Warn("警告信息")
}
```

更多示例请查看 [logger/README.md](./logger/README.md)

### 钉钉机器人

```go
package main

import "github.com/phrynus/go-utils/dingtalk"

func main() {
    dt := dingtalk.NewDingtalk("your_access_token").WithSecret("your_secret")
    
    // 发送文本消息
    err := dt.SendText("Hello, DingTalk!", nil)
    
    // 发送Markdown消息
    title := "系统通知"
    text := "## 系统维护通知\n请提前做好准备！"
    at := &dingtalk.AtMeta{IsAtAll: true}
    
    err = dt.SendMarkdown(title, text, at)
}
```

更多示例请查看 [dingtalk/README.md](./dingtalk/README.md) 或 [example/dingtalk.go](./example/dingtalk.go)

### 飞书机器人

```go
package main

import "github.com/phrynus/go-utils/feishu"

func main() {
    // 创建飞书客户端（使用webhook URL和密钥）
    fs := feishu.NewFeiShu("https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id").
        WithSecret("your_secret")
    
    // 发送文本消息
    err := fs.SendText("Hello, FeiShu! 这是一条测试消息。")
    
    // 发送富文本消息
    post := &feishu.Post{
        ZhCn: &feishu.PostDetail{
            Title: "系统监控告警",
            Content: [][]feishu.PostElem{
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

更多示例请查看 [feishu/README.md](./feishu/README.md) 或 [example/feishu.go](./example/feishu.go)

### 用户API客户端

```go
package main

import (
    "context"
    "log"
    user "github.com/phrynus/go-utils/uyz-u"
)

func main() {
    client, err := user.New(user.ClientConfig{
        BaseURL:          "https://example.com/api/user",
        AppID:            1003,
        AppKey:           "your_app_key",
        Version:          "1.0.0",
        VersionIndex:     "web",
        ClientPrivateKey: rsaClientPrivateKey,
        ServerPublicKey:  rsaServerPublicKey,
        EncryptionMode:   user.EncryptionRSA,
        DisableSignature: false,
    })
    if err != nil {
        log.Fatalf("create client: %v", err)
    }

    // 登录
    login, err := client.NewLogin().
        Account("username").
        Password("password").
        UDID("device-id").
        Do(context.Background())
    if err != nil {
        log.Fatalf("login failed: %v", err)
    }
    
    fmt.Println(login)
}
```

更多示例请查看 [uyz-u/README.md](./uyz-u/README.md) 或 [example/uyz-u.go](./example/uyz-u.go)

## 示例代码

完整的示例代码请查看 [example](./example/) 目录：

- [ta.go](./example/ta.go) - 技术分析指标使用示例
- [dingtalk.go](./example/dingtalk.go) - 钉钉机器人使用示例
- [feishu.go](./example/feishu.go) - 飞书机器人使用示例
- [uyz-u.go](./example/uyz-u.go) - 用户API客户端使用示例

## 版本发布

```bash
git tag v1.3.2
git push origin --tags
```

## 许可证

本项目采用 GNU Affero General Public License v3.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。数字货币和合约交易具有高风险，请谨慎使用。
