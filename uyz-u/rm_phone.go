package user

import (
	"context"
	"errors"
	"strconv"
)

// RmPhone 构建并执行解绑手机号 API 请求
type RmPhone struct {
	client *Client
	req    RmPhoneRequest
}

// RmPhoneRequest 携带解绑手机号端点的信息
type RmPhoneRequest struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// NewRmPhone 解绑手机号
func (c *Client) NewRmPhone() *RmPhone {
	return &RmPhone{client: c}
}

// Phone 设置手机号
func (r *RmPhone) Phone(phone string) *RmPhone {
	r.req.Phone = phone
	return r
}

// Code 设置验证码
func (r *RmPhone) Code(code string) *RmPhone {
	r.req.Code = code
	return r
}

// Do 发送请求，可选择性地覆盖 context
func (r *RmPhone) Do(ctx ...context.Context) (bool, error) {
	if r.client == nil {
		return false, errNilClient
	}
	token, err := r.client.GetToken()
	if err != nil {
		return false, err
	}
	r.req.Token = token
	if r.req.Phone == "" {
		return false, errors.New("手机号是必需的")
	}
	if r.req.Code == "" {
		return false, errors.New("验证码是必需的")
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
	res, err := r.client.SecurePost(callCtx, "rmPhone", r.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
