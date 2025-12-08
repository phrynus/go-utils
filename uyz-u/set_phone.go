package user

import (
	"context"
	"errors"
	"strconv"
)

// SetPhone 构建并执行绑定手机号 API 请求
type SetPhone struct {
	client *Client
	req    SetPhoneRequest
}

// SetPhoneRequest 携带绑定手机号端点的信息
type SetPhoneRequest struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// NewSetPhone 绑定手机号
func (c *Client) NewSetPhone() *SetPhone {
	return &SetPhone{client: c}
}

// Phone 设置手机号
func (s *SetPhone) Phone(phone string) *SetPhone {
	s.req.Phone = phone
	return s
}

// Code 设置验证码
func (s *SetPhone) Code(code string) *SetPhone {
	s.req.Code = code
	return s
}

// Do 发送请求，可选择性地覆盖 context
func (s *SetPhone) Do(ctx ...context.Context) (bool, error) {
	if s.client == nil {
		return false, errNilClient
	}
	token, err := s.client.GetToken()
	if err != nil {
		return false, err
	}
	s.req.Token = token
	if s.req.Phone == "" {
		return false, errors.New("手机号是必需的")
	}
	if s.req.Code == "" {
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
	res, err := s.client.SecurePost(callCtx, "setPhone", s.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
