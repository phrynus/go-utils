package user

import (
	"context"
	"errors"
	"strconv"
)

// MessageEnd 构建并执行结束留言 API 请求
type MessageEnd struct {
	client *Client
	req    MessageEndRequest
}

// MessageEndRequest 携带结束留言端点的信息
type MessageEndRequest struct {
	Token string `json:"token"`
	MID   int    `json:"mid"`
}

// NewMessageEnd 结束留言
func (c *Client) NewMessageEnd() *MessageEnd {
	return &MessageEnd{client: c}
}

// MID 设置留言ID
func (m *MessageEnd) MID(mid int) *MessageEnd {
	m.req.MID = mid
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *MessageEnd) Do(ctx ...context.Context) (bool, error) {
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
	res, err := m.client.SecurePost(callCtx, "messageEnd", m.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
