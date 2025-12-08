package user

import (
	"context"
	"errors"
	"strconv"
)

// SignIn 构建并执行签到 API 请求
type SignIn struct {
	client *Client
	req    SignInRequest
}

// SignInRequest 携带签到端点的信息
type SignInRequest struct {
	Token string `json:"token"`
	Time  int64  `json:"time"`
}

// NewSignIn 签到
func (c *Client) NewSignIn() *SignIn {
	return &SignIn{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (s *SignIn) Do(ctx ...context.Context) (bool, error) {
	if s.client == nil {
		return false, errNilClient
	}
	token, err := s.client.GetToken()
	if err != nil {
		return false, err
	}
	s.req.Token = token
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
	res, err := s.client.SecurePost(callCtx, "signIn", s.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
