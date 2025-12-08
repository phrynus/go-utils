package user

import (
	"context"
	"errors"
	"strconv"
)

// MessageReply 构建并执行回复留言 API 请求
type MessageReply struct {
	client *Client
	req    MessageReplyRequest
}

// MessageReplyRequest 携带回复留言端点的信息
type MessageReplyRequest struct {
	Token   string   `json:"token"`
	MID     int      `json:"mid"`
	Content string   `json:"content"`
	File    []string `json:"file,omitempty"`
}

// NewMessageReply 回复留言
func (c *Client) NewMessageReply() *MessageReply {
	return &MessageReply{client: c}
}

// MID 设置留言ID
func (m *MessageReply) MID(mid int) *MessageReply {
	m.req.MID = mid
	return m
}

// Content 设置回复内容
func (m *MessageReply) Content(content string) *MessageReply {
	m.req.Content = content
	return m
}

// File 设置文件列表（可空，调用上传文件接口）
func (m *MessageReply) File(files []string) *MessageReply {
	m.req.File = files
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *MessageReply) Do(ctx ...context.Context) (bool, error) {
	if m.client == nil {
		return false, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return false, err
	}
	m.req.Token = token
	if m.req.MID == 0 {
		return false, errors.New("留言ID是必需的")
	}
	if m.req.Content == "" {
		return false, errors.New("回复内容是必需的")
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
	res, err := m.client.SecurePost(callCtx, "messageReply", m.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
