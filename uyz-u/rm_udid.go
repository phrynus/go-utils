package user

import (
	"context"
	"errors"
	"strconv"
)

// RmUDID 构建并执行解绑设备 API 请求
type RmUDID struct {
	client *Client
	req    RmUDIDRequest
}

// RmUDIDRequest 携带解绑设备端点的信息
type RmUDIDRequest struct {
	Token string `json:"token"`
	UDID  string `json:"udid"`
}

// NewRmUDID 解绑设备
func (c *Client) NewRmUDID() *RmUDID {
	return &RmUDID{client: c}
}

// UDID 设置设备标识符
func (r *RmUDID) UDID(udid string) *RmUDID {
	r.req.UDID = udid
	return r
}

// Do 发送请求，可选择性地覆盖 context
func (r *RmUDID) Do(ctx ...context.Context) (bool, error) {
	if r.client == nil {
		return false, errNilClient
	}
	token, err := r.client.GetToken()
	if err != nil {
		return false, err
	}
	r.req.Token = token
	if r.req.UDID == "" {
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
	res, err := r.client.SecurePost(callCtx, "rmUdid", r.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
