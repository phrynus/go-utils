package user

import (
	"context"
	"errors"
	"strconv"
)

// RmEmail 构建并执行解绑邮箱 API 请求
type RmEmail struct {
	client *Client
	req    RmEmailRequest
}

// RmEmailRequest 携带解绑邮箱端点的信息
type RmEmailRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

// NewRmEmail 解绑邮箱
func (c *Client) NewRmEmail() *RmEmail {
	return &RmEmail{client: c}
}

// Email 设置邮箱
func (r *RmEmail) Email(email string) *RmEmail {
	r.req.Email = email
	return r
}

// Code 设置验证码
func (r *RmEmail) Code(code string) *RmEmail {
	r.req.Code = code
	return r
}

// Do 发送请求，可选择性地覆盖 context
func (r *RmEmail) Do(ctx ...context.Context) (bool, error) {
	if r.client == nil {
		return false, errNilClient
	}
	token, err := r.client.GetToken()
	if err != nil {
		return false, err
	}
	r.req.Token = token
	if r.req.Email == "" {
		return false, errors.New("邮箱是必需的")
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
	res, err := r.client.SecurePost(callCtx, "rmEmail", r.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
