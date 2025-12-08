package user

import (
	"context"
	"errors"
	"strconv"
)

// ResetPwd 构建并执行重置密码 API 请求
type ResetPwd struct {
	client *Client
	req    ResetPwdRequest
}

// ResetPwdRequest 携带重置密码端点的信息
type ResetPwdRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// NewResetPwd 重置密码
func (c *Client) NewResetPwd() *ResetPwd {
	return &ResetPwd{client: c}
}

// Account 设置账号
func (r *ResetPwd) Account(account string) *ResetPwd {
	r.req.Account = account
	return r
}

// Password 设置新密码
func (r *ResetPwd) Password(password string) *ResetPwd {
	r.req.Password = password
	return r
}

// Code 设置验证码
func (r *ResetPwd) Code(code string) *ResetPwd {
	r.req.Code = code
	return r
}

// Do 发送请求，可选择性地覆盖 context
func (r *ResetPwd) Do(ctx ...context.Context) (bool, error) {
	if r.client == nil {
		return false, errNilClient
	}
	if r.req.Account == "" {
		return false, errors.New("账号是必需的")
	}
	if r.req.Password == "" {
		return false, errors.New("新密码是必需的")
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
	res, err := r.client.SecurePost(callCtx, "resetPwd", r.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
