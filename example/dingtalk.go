package main

import (
	"fmt"
	"log"

	"github.com/phrynus/go-utils/dingtalk"
)

// 注意：以下示例中的accessToken和secret需要替换为真实的值
const (
	accessToken = "dc477bf40c7be545e31429d962c3c28946a6da114c7a9c75a4bea55f4d7a18a8"
	secret      = "SECd07ddca536b5385f14b22b819e9519ac144bd79263c96abcc5a64700762afb0d"
)

var dt = dingtalk.NewDingtalk(accessToken).WithSecret(secret)

func TestDingtalk() {
	ExampleSendText()
	ExampleSendTextWithAt()
	ExampleSendMarkdown()
	ExampleSendLink()
	ExampleSendActionCard()
	ExampleSendFeedCard()
}

func ExampleSendText() {
	// 使用新的链式操作发送简单文本消息

	err := dt.SendText("Hello, DingTalk!", nil)
	if err != nil {
		log.Printf("发送文本消息失败: %v", err)
		return
	}
	fmt.Println("文本消息发送成功")
}

func ExampleSendTextWithAt() {
	// 发送带@功能的文本消息

	at := &dingtalk.AtMeta{
		AtMobiles: []string{"18156274316"}, // @指定手机号
		IsAtAll:   false,                   // 不@所有人
	}

	err := dt.SendText("紧急通知：系统维护将在今晚进行！", at)
	if err != nil {
		log.Printf("发送@消息失败: %v", err)
		return
	}
	fmt.Println("@消息发送成功")
}

func ExampleSendMarkdown() {
	// 发送Markdown消息

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
		IsAtAll: false, // @所有人
	}

	err := dt.SendMarkdown(title, text, at)
	if err != nil {
		log.Printf("发送Markdown消息失败: %v", err)
		return
	}
	fmt.Println("Markdown消息发送成功")
}

func ExampleSendLink() {
	// 发送链接消息

	link := &dingtalk.LinkMeta{
		Title:      "Go语言官方网站",
		Text:       "Go是Google开发的一种静态强类型、编译型语言。Go语言语法与C相近，但功能上有内存安全、垃圾回收、结构化类型、CSP并发等。",
		MessageUrl: "https://golang.org",
		PicUrl:     "https://golang.org/lib/godoc/images/go-logo-blue.svg",
	}

	err := dt.SendLink(link)
	if err != nil {
		log.Printf("发送链接消息失败: %v", err)
		return
	}
	fmt.Println("链接消息发送成功")
}

func ExampleSendActionCard() {
	// 发送独立跳转ActionCard消息

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
		BtnOrientation: "1", // 按钮横向排列
	}

	err := dt.SendActionCard(actionCard)
	if err != nil {
		log.Printf("发送ActionCard消息失败: %v", err)
		return
	}
	fmt.Println("ActionCard消息发送成功")
}

func ExampleSendFeedCard() {
	// 发送FeedCard消息

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
			{
				Title:      "Kubernetes 入门教程",
				MessageURL: "https://kubernetes.io/docs/tutorials/",
				PicURL:     "https://kubernetes.io/images/kubernetes-horizontal-color.png",
			},
		},
	}

	err := dt.SendFeedCard(feedCard)
	if err != nil {
		log.Printf("发送FeedCard消息失败: %v", err)
		return
	}
	fmt.Println("FeedCard消息发送成功")
}
