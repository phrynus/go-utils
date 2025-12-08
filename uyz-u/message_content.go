package user

import (
	"context"
	"errors"
)

// MessageContent 构建并执行获取留言内容 API 请求
type MessageContent struct {
	client *Client
	req    MessageContentRequest
}

// MessageContentRequest 携带获取留言内容端点的信息
type MessageContentRequest struct {
	Token string `json:"token"`
	MID   int    `json:"mid"`
}

// MessageContentItem 留言内容项
type MessageContentItem struct {
	ID      int      `json:"id"`      // 内容ID
	UG      string   `json:"ug"`      // 用户组
	Content string   `json:"content"` // 内容
	Time    int64    `json:"time"`    // 时间
	State   int      `json:"state"`   // 状态
	User    string   `json:"user"`    // 用户名
	File    []string `json:"file"`    // 文件列表
	Avatars string   `json:"avatars"` // 头像
}

// NewMessageContent 获取留言内容
func (c *Client) NewMessageContent() *MessageContent {
	return &MessageContent{client: c}
}

// MID 设置留言ID
func (m *MessageContent) MID(mid int) *MessageContent {
	m.req.MID = mid
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *MessageContent) Do(ctx ...context.Context) ([]MessageContentItem, error) {
	if m.client == nil {
		return nil, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return nil, err
	}
	m.req.Token = token
	if m.req.MID == 0 {
		return nil, errors.New("留言ID是必需的")
	}
	var callCtx context.Context
	for _, candidate := range ctx {
		if candidate != nil {
			callCtx = candidate
			break
		}
	}
	if callCtx == nil {
		callCtx = context.Background()
	}
	var payload []MessageContentItem
	if _, err := m.client.SecurePost(callCtx, "messageContent", m.req, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
