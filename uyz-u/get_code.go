package user

import (
	"context"
	"errors"
	"strconv"
)

// GetCode 构建并执行获取验证码 API 请求
type GetCode struct {
	client *Client
	req    GetCodeRequest
}

// GetCodeRequest 携带获取验证码端点的信息
type GetCodeRequest struct {
	Account string `json:"account"`
	Type    string `json:"type"`
	Time    int64  `json:"time"`
}

// NewGetCode 获取验证码
func (c *Client) NewGetCode() *GetCode {
	return &GetCode{client: c}
}

// Account 设置邮箱或手机号
func (g *GetCode) Account(account string) *GetCode {
	g.req.Account = account
	return g
}

// Type 设置验证码类型：reg=注册,rmpwd=重置密码,bind=绑定账号通用,rmsn=解绑机器码,rmemail=解绑邮箱,rmphone=解绑手机
func (g *GetCode) Type(typ string) *GetCode {
	g.req.Type = typ
	return g
}

// Do 发送请求，可选择性地覆盖 context
func (g *GetCode) Do(ctx ...context.Context) (bool, error) {
	if g.client == nil {
		return false, errNilClient
	}
	if g.req.Account == "" {
		return false, errors.New("账户是必需的")
	}
	if g.req.Type == "" {
		return false, errors.New("验证码类型是必需的")
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
	res, err := g.client.SecurePost(callCtx, "getCode", g.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
