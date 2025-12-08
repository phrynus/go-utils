# 飞书机器人客户端

飞书自定义机器人 Go 语言客户端，支持发送多种类型的消息。

## 功能特性

- ✅ 支持文本消息
- ✅ 支持富文本消息（Post）
- ✅ 支持图片消息
- ✅ 支持分享群名片
- ✅ 支持消息卡片（Interactive）
- ✅ 支持签名验证（安全设置）
- ✅ 提供富文本元素辅助函数

## 安装

```bash
go get github.com/phrynus/go-utils/feishu
```

## 快速开始

### 创建客户端

```go
import "github.com/phrynus/go-utils/feishu"

// 创建客户端（仅使用 webhook URL）
fs := feishu.NewFeiShu("https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id")

// 创建客户端（使用 webhook URL 和 secret，支持签名验证）
fs := feishu.NewFeiShu("https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id").
    WithSecret("your_secret")
```

### 发送文本消息

```go
err := fs.SendText("Hello, FeiShu! 这是一条测试消息。")
```

### 发送富文本消息（简单版）

```go
post := &feishu.Post{
    ZhCn: &feishu.PostDetail{
        Title: "项目发布通知",
        Content: [][]feishu.PostElem{
            {
                {Tag: "text", Text: "项目名称: "},
                {Tag: "text", Text: "go-utils"},
            },
            {
                {Tag: "text", Text: "版本号: "},
                {Tag: "text", Text: "v1.0.0"},
            },
        },
    },
}

err := fs.SendPost(post)
```

### 发送富文本消息（包含链接、@用户等元素）

```go
post := &feishu.Post{
    ZhCn: &feishu.PostDetail{
        Title: "系统监控告警",
        Content: [][]feishu.PostElem{
            // 第一段
            {
                {Tag: "text", Text: "告警时间: 2024-01-01 14:30:00\n"},
            },
            // 第二段
            {
                {Tag: "text", Text: "告警级别: "},
                {Tag: "text", Text: "严重"},
            },
            // 第三段（带链接）
            {
                {Tag: "text", Text: "查看详情: "},
                {Tag: "a", Text: "点击这里", Href: "https://monitor.example.com/alert/123"},
            },
            // 第四段（@所有人）
            {
                {Tag: "at", UserId: "all"},
                {Tag: "text", Text: " 请及时处理！"},
            },
        },
    },
}

err := fs.SendPost(post)
```

### 使用辅助函数创建富文本消息

```go
post := &feishu.Post{
    ZhCn: &feishu.PostDetail{
        Title: "使用辅助函数示例",
        Content: [][]feishu.PostElem{
            {
                feishu.NewTextElem("这是一条使用辅助函数创建的消息"),
            },
            {
                feishu.NewTextElem("点击链接: "),
                feishu.NewLinkElem("访问官网", "https://www.feishu.cn"),
            },
            {
                feishu.NewTextElem("通知用户: "),
                feishu.NewAtElem("user_id_123"),
            },
            {
                feishu.NewAtAllElem(),
                feishu.NewTextElem(" 请所有人注意！"),
            },
        },
    },
}

err := fs.SendPost(post)
```

### 发送图片消息

```go
// imageKey 需要通过飞书图片上传接口获取
err := fs.SendImage("image_key_here")
```

### 发送分享群名片

```go
err := fs.SendShareChat("chat_id_here")
```

### 发送消息卡片

```go
// card 需要符合飞书消息卡片的格式
card := map[string]interface{}{
    "config": map[string]interface{}{
        "wide_screen_mode": true,
    },
    "elements": []map[string]interface{}{
        {
            "tag": "div",
            "text": map[string]interface{}{
                "tag": "lark_md",
                "content": "这是一条消息卡片",
            },
        },
    },
}

err := fs.SendInteractive(card)
```

### 自定义消息

```go
msg := &feishu.Message{
    MsgType: feishu.MsgTypeText,
    Content: &feishu.TextContent{
        Text: "自定义消息内容",
    },
}

err := fs.Send(msg)
```

## 消息类型

### 文本消息 (text)

- `TextContent`: 文本消息内容

### 富文本消息 (post)

- `Post`: 富文本消息结构
- `PostDetail`: 富文本详情（支持中英文）
- `PostElem`: 富文本元素
  - `text`: 文本
  - `a`: 超链接
  - `at`: @用户
  - `img`: 图片

### 图片消息 (image)

- `ImageContent`: 图片消息内容
- 需要先通过飞书图片上传接口获取 `image_key`

### 分享群名片 (share_chat)

- `ShareChatContent`: 分享群名片内容
- 需要提供群聊的 `chat_id`

### 消息卡片 (interactive)

- 支持自定义消息卡片格式
- 需要符合飞书消息卡片的 JSON 结构

## 富文本元素辅助函数

为了方便创建富文本元素，提供了以下辅助函数：

- `NewTextElem(text string) PostElem`: 创建文本元素
- `NewTextElemWithUnescape(text string, unEscape bool) PostElem`: 创建文本元素（带 unescape）
- `NewLinkElem(text, href string) PostElem`: 创建超链接元素
- `NewAtElem(userId string) PostElem`: 创建@用户元素
- `NewAtElemWithName(userId, userName string) PostElem`: 创建@用户元素（带用户名）
- `NewAtAllElem() PostElem`: 创建@所有人元素
- `NewImageElem(imageKey string) PostElem`: 创建图片元素

## 消息发送频率限制

每个机器人单个群组消息发送频率限制为 50 QPS。

## 完整示例

完整示例请查看 [example/feishu.go](../example/feishu.go)

## API 参考

### 类型定义

- `FeiShu`: 飞书客户端
- `Message`: 消息结构体
- `TextContent`: 文本消息内容
- `Post`: 富文本消息
- `PostDetail`: 富文本详情
- `PostElem`: 富文本元素
- `ImageContent`: 图片消息内容
- `ShareChatContent`: 分享群名片内容
- `ResponseMeta`: 响应信息

### 方法

- `NewFeiShu(webhookURL string) *FeiShu`: 创建飞书客户端
- `WithSecret(secret string) *FeiShu`: 设置签名密钥（链式调用）
- `Send(msg *Message) error`: 发送消息
- `SendText(text string) error`: 发送文本消息
- `SendPost(post *Post) error`: 发送富文本消息
- `SendImage(imageKey string) error`: 发送图片消息
- `SendShareChat(chatId string) error`: 发送分享群名片
- `SendInteractive(card any) error`: 发送消息卡片

## 获取 Webhook URL

1. 在飞书群聊中添加自定义机器人
2. 在机器人设置中获取 Webhook 地址
3. 完整的 Webhook URL 格式：`https://open.feishu.cn/open-apis/bot/v2/hook/{webhook_id}`
4. 如果启用了安全设置，需要配置 `secret`

## 注意事项

- 富文本消息支持中英文两种语言（`ZhCn` 和 `EnUs`）
- 富文本内容是一个二维数组，第一维是段落，第二维是段落内的元素
- 如果配置了签名验证，必须使用 `WithSecret()` 方法设置密钥
- 图片消息需要先通过飞书图片上传接口获取 `image_key`
- 建议在生产环境中使用签名验证以提高安全性

