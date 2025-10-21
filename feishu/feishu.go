package feishu

import (
	"errors"
	"fmt"
	"time"
)

const (
	CodeOK    = 0
	MessageOK = "ok"

	// 消息类型常量
	MsgTypeText        = "text"        // 文本
	MsgTypePost        = "post"        // 富文本
	MsgTypeImage       = "image"       // 图片
	MsgTypeShareChat   = "share_chat"  // 分享群名片
	MsgTypeInteractive = "interactive" // 消息卡片
)

var ErrRequest = errors.New("feishu request failed")

func (r ResponseMeta) String() string {
	return fmt.Sprintf("code: %v, msg: %s", r.Code, r.Msg)
}

// Succeed 操作是否成功
func (r ResponseMeta) Succeed() bool {
	return r.Code == CodeOK
}

// NewFeiShu 创建新的飞书客户端
// webhookURL: 完整的webhook地址
func NewFeiShu(webhookURL string) *FeiShu {
	return &FeiShu{
		webhookURL: webhookURL,
	}
}

// WithSecret 设置密钥（支持链式调用）
func (f *FeiShu) WithSecret(secret string) *FeiShu {
	f.secret = secret
	return f
}

// ValidateMsgType 验证消息类型
func ValidateMsgType(v string) error {
	switch v {
	case MsgTypeText, MsgTypePost, MsgTypeImage, MsgTypeShareChat, MsgTypeInteractive:
	default:
		return fmt.Errorf("%s not in [%q %q %q %q %q]", v,
			MsgTypeText, MsgTypePost, MsgTypeImage, MsgTypeShareChat, MsgTypeInteractive)
	}
	return nil
}

// Message 飞书自定义机器人消息
type Message struct {
	Timestamp string `json:"timestamp,omitempty"` // 时间戳（秒）
	Sign      string `json:"sign,omitempty"`      // 签名
	MsgType   string `json:"msg_type"`            // 消息类型
	Content   any    `json:"content"`             // 消息内容
}

// Send 发送飞书自定义机器人消息
//
// 消息发送频率限制
// 每个机器人单个群组消息发送频率限制为 50 QPS
func (f *FeiShu) Send(msg *Message) error {
	// 如果配置了密钥，则添加签名
	if f.secret != "" {
		timestamp := time.Now().Unix()
		sign, err := GenSign(f.secret, timestamp)
		if err != nil {
			return fmt.Errorf("sign failed: %w", err)
		}
		msg.Timestamp = fmt.Sprintf("%d", timestamp)
		msg.Sign = sign
	}

	var resp ResponseMeta
	headers, err := PostJSON(f.webhookURL, msg, &resp)
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

// SendText 发送文本消息
func (f *FeiShu) SendText(text string) error {
	msg := &Message{
		MsgType: MsgTypeText,
		Content: &TextContent{
			Text: text,
		},
	}
	return f.Send(msg)
}

// SendPost 发送富文本消息
func (f *FeiShu) SendPost(post *Post) error {
	msg := &Message{
		MsgType: MsgTypePost,
		Content: &PostContent{
			Post: post,
		},
	}
	return f.Send(msg)
}

// SendImage 发送图片消息
func (f *FeiShu) SendImage(imageKey string) error {
	msg := &Message{
		MsgType: MsgTypeImage,
		Content: &ImageContent{
			ImageKey: imageKey,
		},
	}
	return f.Send(msg)
}

// SendShareChat 发送分享群名片消息
func (f *FeiShu) SendShareChat(chatId string) error {
	msg := &Message{
		MsgType: MsgTypeShareChat,
		Content: &ShareChatContent{
			ChatId: chatId,
		},
	}
	return f.Send(msg)
}

// SendInteractive 发送消息卡片
func (f *FeiShu) SendInteractive(card any) error {
	msg := &Message{
		MsgType: MsgTypeInteractive,
		Content: card,
	}
	return f.Send(msg)
}
