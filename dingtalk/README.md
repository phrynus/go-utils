# 钉钉机器人客户端

钉钉自定义机器人 Go 语言客户端，支持发送多种类型的消息。

## 功能特性

- ✅ 支持文本消息
- ✅ 支持 Markdown 消息
- ✅ 支持链接消息
- ✅ 支持 ActionCard 消息（独立跳转/整体跳转）
- ✅ 支持 FeedCard 消息
- ✅ 支持 @用户功能
- ✅ 支持签名验证（安全设置）

## 安装

```bash
go get github.com/phrynus/go-utils/dingtalk
```

## 快速开始

### 创建客户端

```go
import "github.com/phrynus/go-utils/dingtalk"

// 创建客户端（仅使用 access_token）
dt := dingtalk.NewDingtalk("your_access_token")

// 创建客户端（使用 access_token 和 secret，支持签名验证）
dt := dingtalk.NewDingtalk("your_access_token").WithSecret("your_secret")
```

### 发送文本消息

```go
// 简单文本消息
err := dt.SendText("Hello, DingTalk!", nil)

// 带@功能的文本消息
at := &dingtalk.AtMeta{
    AtMobiles: []string{"18156274316"}, // @指定手机号
    IsAtAll:   false,                   // 不@所有人
}
err := dt.SendText("紧急通知：系统维护将在今晚进行！", at)

// @所有人
at := &dingtalk.AtMeta{
    IsAtAll: true,
}
err := dt.SendText("重要通知：请所有人注意！", at)
```

### 发送 Markdown 消息

```go
title := "项目发布通知"
text := `## 项目发布通知

### 发布信息
- **项目名称**: go-utils
- **版本号**: v1.0.0
- **发布时间**: 2024-01-01 10:00:00

### 更新内容
1. 新增钉钉机器人功能
2. 优化HTTP客户端
3. 修复已知问题

> 如有问题，请及时反馈！`

at := &dingtalk.AtMeta{
    IsAtAll: false,
}

err := dt.SendMarkdown(title, text, at)
```

### 发送链接消息

```go
link := &dingtalk.LinkMeta{
    Title:      "Go语言官方网站",
    Text:       "Go是Google开发的一种静态强类型、编译型语言。",
    MessageUrl: "https://golang.org",
    PicUrl:     "https://golang.org/lib/godoc/images/go-logo-blue.svg",
}

err := dt.SendLink(link)
```

### 发送 ActionCard 消息

```go
// 独立跳转 ActionCard
actionCard := &dingtalk.ActionCardMeta{
    Title: "系统监控告警",
    Text: `## 系统监控告警

**告警时间**: 2024-01-01 14:30:00
**告警级别**: 严重
**告警内容**: CPU使用率超过90%

请相关人员及时处理！`,
    Btns: []dingtalk.ActionCardBtnMeta{
        {
            Title:     "查看详情",
            ActionURL: "https://monitor.example.com/alert/123",
        },
        {
            Title:     "处理告警",
            ActionURL: "https://monitor.example.com/handle/123",
        },
    },
    BtnOrientation: "1", // 按钮横向排列，"0" 为纵向排列
}

err := dt.SendActionCard(actionCard)
```

### 发送 FeedCard 消息

```go
feedCard := &dingtalk.FeedCardMeta{
    Links: []dingtalk.FeedCardLinkMeta{
        {
            Title:      "Go 1.21 发布",
            MessageURL: "https://golang.org/doc/go1.21",
            PicURL:     "https://golang.org/lib/godoc/images/go-logo-blue.svg",
        },
        {
            Title:      "Docker 最佳实践",
            MessageURL: "https://docs.docker.com/develop/dev-best-practices/",
            PicURL:     "https://www.docker.com/sites/default/files/d8/2019-07/vertical-logo-monochromatic.png",
        },
    },
}

err := dt.SendFeedCard(feedCard)
```

### 自定义消息

```go
msg := &dingtalk.Message{
    MsgType: dingtalk.MsgTypeText,
    Text: &dingtalk.TextMeta{
        Content: "自定义消息内容",
    },
    At: &dingtalk.AtMeta{
        IsAtAll: true,
    },
}

err := dt.Send(msg)
```

## 消息类型

### 文本消息 (text)

- `TextMeta`: 文本消息内容
- `AtMeta`: @用户配置

### Markdown 消息 (markdown)

- `MarkdownMeta`: Markdown 消息内容
- 支持 Markdown 语法
- 支持 @用户

### 链接消息 (link)

- `LinkMeta`: 链接消息内容
- 包含标题、文本、跳转链接、图片

### ActionCard 消息 (actionCard)

- `ActionCardMeta`: 独立跳转 ActionCard
- `SingleActionCardMeta`: 整体跳转 ActionCard
- 支持多个按钮

### FeedCard 消息 (feedCard)

- `FeedCardMeta`: FeedCard 消息内容
- 支持多个链接卡片

## 消息发送频率限制

每个机器人每分钟最多发送 20 条消息到群里，如果超过 20 条，会限流 10 分钟。

如果你有大量发消息的场景（譬如系统监控报警）可以将这些信息进行整合，通过 markdown 消息以摘要的形式发送到群里。

## 完整示例

完整示例请查看 [example/dingtalk.go](../example/dingtalk.go)

## API 参考

### 类型定义

- `DingTalk`: 钉钉客户端
- `Message`: 消息结构体
- `TextMeta`: 文本消息
- `MarkdownMeta`: Markdown 消息
- `LinkMeta`: 链接消息
- `ActionCardMeta`: ActionCard 消息
- `FeedCardMeta`: FeedCard 消息
- `AtMeta`: @用户配置
- `ResponseMeta`: 响应信息

### 方法

- `NewDingtalk(accessToken string) *DingTalk`: 创建钉钉客户端
- `WithSecret(secret string) *DingTalk`: 设置签名密钥（链式调用）
- `Send(msg *Message) error`: 发送消息
- `SendText(content string, at *AtMeta) error`: 发送文本消息
- `SendMarkdown(title, text string, at *AtMeta) error`: 发送 Markdown 消息
- `SendLink(link *LinkMeta) error`: 发送链接消息
- `SendActionCard(actionCard *ActionCardMeta) error`: 发送 ActionCard 消息
- `SendFeedCard(feedCard *FeedCardMeta) error`: 发送 FeedCard 消息

## 获取 Access Token

1. 在钉钉群聊中添加自定义机器人
2. 在机器人设置中获取 Webhook 地址
3. 从 Webhook 地址中提取 `access_token` 参数
4. 如果启用了安全设置，需要配置 `secret`

## 注意事项

- 所有消息类型都支持 @用户功能
- 如果配置了签名验证，必须使用 `WithSecret()` 方法设置密钥
- 消息内容需要符合钉钉的格式要求
- 建议在生产环境中使用签名验证以提高安全性

