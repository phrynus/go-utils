package user

import (
	"context"
	"errors"
	"strconv"
)

// Heartbeat 构建并执行心跳 API 请求
type Heartbeat struct {
	client *Client
	req    HeartbeatRequest
}

// NewHeartbeat 心跳
func (c *Client) NewHeartbeat() *Heartbeat {
	return &Heartbeat{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (h *Heartbeat) Do(ctx ...context.Context) (bool, error) {
	if h.client == nil {
		return false, errNilClient
	}
	token, err := h.client.GetToken()
	if err != nil {
		return false, err
	}
	h.req.Token = token
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
	res, err := h.client.SecurePost(callCtx, "heartbeat", h.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
