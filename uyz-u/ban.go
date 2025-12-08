package user

import (
	"context"
	"errors"
	"strconv"
)

// Ban 构建并执行账户禁用 API 请求
type Ban struct {
	client *Client
	req    BanRequest
}

// BanRequest 携带账户禁用端点的信息
type BanRequest struct {
	Token   string `json:"token"`
	Second  int    `json:"second,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewBan 账户禁用
func (c *Client) NewBan() *Ban {
	return &Ban{client: c}
}

// Second 设置禁用时间（秒，默认60，最大值2592000）
func (b *Ban) Second(second int) *Ban {
	b.req.Second = second
	return b
}

// Message 设置禁用提示
func (b *Ban) Message(message string) *Ban {
	b.req.Message = message
	return b
}

// Do 发送请求，可选择性地覆盖 context
func (b *Ban) Do(ctx ...context.Context) (bool, error) {
	if b.client == nil {
		return false, errNilClient
	}
	token, err := b.client.GetToken()
	if err != nil {
		return false, err
	}
	b.req.Token = token
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
	res, err := b.client.SecurePost(callCtx, "ban", b.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
