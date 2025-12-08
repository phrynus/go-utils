package user

import (
	"context"
	"errors"
	"strconv"
)

// SetEmail 构建并执行绑定邮箱 API 请求
type SetEmail struct {
	client *Client
	req    SetEmailRequest
}

// SetEmailRequest 携带绑定邮箱端点的信息
type SetEmailRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

// NewSetEmail 绑定邮箱
func (c *Client) NewSetEmail() *SetEmail {
	return &SetEmail{client: c}
}

// Email 设置邮箱
func (s *SetEmail) Email(email string) *SetEmail {
	s.req.Email = email
	return s
}

// Code 设置验证码
func (s *SetEmail) Code(code string) *SetEmail {
	s.req.Code = code
	return s
}

// Do 发送请求，可选择性地覆盖 context
func (s *SetEmail) Do(ctx ...context.Context) (bool, error) {
	if s.client == nil {
		return false, errNilClient
	}
	token, err := s.client.GetToken()
	if err != nil {
		return false, err
	}
	s.req.Token = token
	if s.req.Email == "" {
		return false, errors.New("邮箱是必需的")
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
	res, err := s.client.SecurePost(callCtx, "setEmail", s.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
