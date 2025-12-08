package user

import (
	"context"
	"errors"
	"strconv"
)

// BindUDID 构建并执行绑定设备 API 请求
type BindUDID struct {
	client *Client
	req    BindUDIDRequest
}

// BindUDIDRequest 携带绑定设备端点的信息
type BindUDIDRequest struct {
	Token string `json:"token"`
	UDID  string `json:"udid"`
}

// NewBindUDID 绑定设备
func (c *Client) NewBindUDID() *BindUDID {
	return &BindUDID{client: c}
}

// UDID 设置设备标识符
func (b *BindUDID) UDID(udid string) *BindUDID {
	b.req.UDID = udid
	return b
}

// Do 发送请求，可选择性地覆盖 context
func (b *BindUDID) Do(ctx ...context.Context) (bool, error) {
	if b.client == nil {
		return false, errNilClient
	}
	token, err := b.client.GetToken()
	if err != nil {
		return false, err
	}
	b.req.Token = token
	if b.req.UDID == "" {
		return false, errors.New("设备标识符是必需的")
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
	res, err := b.client.SecurePost(callCtx, "bindUdid", b.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
