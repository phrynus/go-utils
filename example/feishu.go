package main

import (
	"fmt"
	"log"

	"github.com/phrynus/go-utils/feishu"
)

// 注意：以下示例中的webhookURL和secret需要替换为真实的值
const (
	webhookURL = "https://open.feishu.cn/open-apis/bot/v2/hook/23c7d1c6-e52c-405e-a2b3-7ae10cc05e05"
	fsSecret   = "sOWtlqNTRJiCqZlsf2VFQe" // 替换为真实的密钥
)

var fs = feishu.NewFeiShu(webhookURL).WithSecret(fsSecret)

func TestFeiShu() {
	ExampleFsSendText()
	ExampleFsSendPost()
	ExampleFsSendPostWithElements()
	ExampleFsSendPostWithHelpers()
}

func ExampleFsSendText() {
	// 发送简单文本消息

	err := fs.SendText("Hello, FeiShu! 这是一条测试消息。")
	if err != nil {
		log.Printf("发送文本消息失败: %v", err)
		return
	}
	fmt.Println("文本消息发送成功")
}

func ExampleFsSendPost() {
	// 发送富文本消息（简单版）

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
	if err != nil {
		log.Printf("发送富文本消息失败: %v", err)
		return
	}
	fmt.Println("富文本消息发送成功")
}

func ExampleFsSendPostWithElements() {
	// 发送富文本消息（包含链接、@用户等元素）

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
				// 第三段
				{
					{Tag: "text", Text: "告警内容: CPU使用率超过90%"},
				},
				// 第四段（带链接）
				{
					{Tag: "text", Text: "查看详情: "},
					{Tag: "a", Text: "点击这里", Href: "https://monitor.example.com/alert/123"},
				},
			},
		},
	}

	err := fs.SendPost(post)
	if err != nil {
		log.Printf("发送富文本消息失败: %v", err)
		return
	}
	fmt.Println("富文本消息（带元素）发送成功")
}

func ExampleFsSendPostWithHelpers() {
	// 使用辅助函数发送富文本消息（更简洁的方式）

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
			},
		},
	}

	err := fs.SendPost(post)
	if err != nil {
		log.Printf("发送富文本消息失败: %v", err)
		return
	}
	fmt.Println("富文本消息（使用辅助函数）发送成功")
}
