package user

import (
	"context"
	"errors"
	"strconv"
)

// MessageAdd 构建并执行新增留言 API 请求
type MessageAdd struct {
	client *Client
	req    MessageAddRequest
}

// MessageAddRequest 携带新增留言端点的信息
type MessageAddRequest struct {
	Token   string   `json:"token"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	File    []string `json:"file,omitempty"`
}

// NewMessageAdd 新增留言
func (c *Client) NewMessageAdd() *MessageAdd {
	return &MessageAdd{client: c}
}

// Title 设置标题
func (m *MessageAdd) Title(title string) *MessageAdd {
	m.req.Title = title
	return m
}

// Content 设置内容
func (m *MessageAdd) Content(content string) *MessageAdd {
	m.req.Content = content
	return m
}

// File 设置文件列表（可空，调用上传文件接口）
func (m *MessageAdd) File(files []string) *MessageAdd {
	m.req.File = files
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *MessageAdd) Do(ctx ...context.Context) (bool, error) {
	if m.client == nil {
		return false, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return false, err
	}
	m.req.Token = token
	if m.req.Title == "" {
		return false, errors.New("标题是必需的")
	}
	if m.req.Content == "" {
		return false, errors.New("内容是必需的")
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
	res, err := m.client.SecurePost(callCtx, "messageAdd", m.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
