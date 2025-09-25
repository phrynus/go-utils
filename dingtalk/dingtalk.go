package dingtalk

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

const (
	CodeOK    = 0
	MessageOK = "ok"

	sendURL = "https://oapi.dingtalk.com/robot/send?access_token="

	// 消息类型常量
	MsgTypeText       = "text"       // 文本
	MsgTypeLink       = "link"       // 链接
	MsgTypeMarkdown   = "markdown"   // markdown
	MsgTypeFeedCard   = "feedCard"   // FeedCard
	MsgTypeActionCard = "actionCard" // ActionCard
)

var ErrRequest = errors.New("dingtalk request failed")

func (t ResponseMeta) String() string {
	return fmt.Sprintf("errcode: %v, errmsg: %s", t.ErrorCode, t.ErrorMessage)
}

// Succeed 操作是否成功
func (t ResponseMeta) Succeed() bool {
	return t.ErrorCode == CodeOK
}

// NewDingtalk 创建新的钉钉客户端
func NewDingtalk(accessToken string) *DingTalk {
	return &DingTalk{
		accessToken: accessToken,
	}
}

// WithSecret 设置密钥（支持链式调用）
func (d *DingTalk) WithSecret(secret string) *DingTalk {
	d.secret = secret
	return d
}

// ValidateMsgType 验证消息类型
func ValidateMsgType(v string) error {
	switch v {
	case MsgTypeText, MsgTypeLink, MsgTypeMarkdown, MsgTypeActionCard, MsgTypeFeedCard:
	default:
		return fmt.Errorf("%s not in [%q %q %q %q %q]", v,
			MsgTypeText, MsgTypeLink, MsgTypeMarkdown, MsgTypeActionCard, MsgTypeFeedCard)
	}
	return nil
}

// Message 钉钉自定义机器人消息
type Message struct {
	MsgType    string        `json:"msgtype"`              // 消息类型
	Text       *TextMeta     `json:"text,omitempty"`       // 文本消息
	Markdown   *MarkdownMeta `json:"markdown,omitempty"`   // markdown消息
	At         *AtMeta       `json:"at,omitempty"`         // @
	Link       *LinkMeta     `json:"link,omitempty"`       // 链接
	ActionCard any           `json:"actionCard,omitempty"` // ActionCard
	FeedCard   *FeedCardMeta `json:"feedCard,omitempty"`   // FeedCard
}

// Send 发送钉钉自定义机器人消息
//
// 消息发送频率限制
// 每个机器人每分钟最多发送20条消息到群里，如果超过20条，会限流10分钟
// 如果你有大量发消息的场景（譬如系统监控报警）可以将这些信息进行整合，通过markdown消息以摘要的形式发送到群里。
func (d *DingTalk) Send(msg *Message) error {
	u := sendURL + url.QueryEscape(d.accessToken)
	if d.secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sign, err := Sign(timestamp, d.secret)
		if err != nil {
			return fmt.Errorf("sign failed: %w", err)
		}
		u = u + "&timestamp=" + timestamp + "&sign=" + url.QueryEscape(sign)
	}
	var resp ResponseMeta
	headers, err := PostJSON(u, msg, &resp)
	if err != nil {
		if headers == nil {
			return err
		}
		return fmt.Errorf("%w, %s", err, resp.String())
	}
	if !resp.Succeed() {
		return fmt.Errorf("%s", resp.String())
	}
	return nil
}

// SendText 发送文本消息（支持链式调用）
func (d *DingTalk) SendText(content string, at *AtMeta) error {
	msg := &Message{
		MsgType: MsgTypeText,
		Text: &TextMeta{
			Content: content,
		},
		At: at,
	}
	return d.Send(msg)
}

// SendMarkdown 发送markdown消息（支持链式调用）
func (d *DingTalk) SendMarkdown(title, text string, at *AtMeta) error {
	msg := &Message{
		MsgType: MsgTypeMarkdown,
		Markdown: &MarkdownMeta{
			Title: title,
			Text:  text,
		},
		At: at,
	}
	return d.Send(msg)
}

// SendLink 发送链接消息（支持链式调用）
func (d *DingTalk) SendLink(link *LinkMeta) error {
	msg := &Message{
		MsgType: MsgTypeLink,
		Link:    link,
	}
	return d.Send(msg)
}

// SendActionCard 发送ActionCard消息（支持链式调用）
func (d *DingTalk) SendActionCard(actionCard *ActionCardMeta) error {
	msg := &Message{
		MsgType:    MsgTypeActionCard,
		ActionCard: actionCard,
	}
	return d.Send(msg)
}

// SendFeedCard 发送FeedCard消息（支持链式调用）
func (d *DingTalk) SendFeedCard(feedCard *FeedCardMeta) error {
	msg := &Message{
		MsgType:  MsgTypeFeedCard,
		FeedCard: feedCard,
	}
	return d.Send(msg)
}
