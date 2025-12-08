package user

import (
	"context"
	"errors"
	"strconv"
)

// Logout 构建并执行退出 API 请求
type Logout struct {
	client *Client
	req    LogoutRequest
}

// NewLogout 退出
func (c *Client) NewLogout() *Logout {
	return &Logout{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (l *Logout) Do(ctx ...context.Context) (bool, error) {
	if l.client == nil {
		return false, errNilClient
	}
	token, err := l.client.GetToken()
	if err != nil {
		return false, err
	}
	l.req.Token = token
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
	res, err := l.client.SecurePost(callCtx, "logout", l.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	// 清除 token
	l.client.ClearToken()
	return true, nil
}
