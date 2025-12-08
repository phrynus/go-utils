package user

import (
	"context"
	"errors"
	"strconv"
)

// ModifyPwd 构建并执行修改密码 API 请求
type ModifyPwd struct {
	client *Client
	req    ModifyPwdRequest
}

// ModifyPwdRequest 携带修改密码端点的信息
type ModifyPwdRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// NewModifyPwd 修改密码
func (c *Client) NewModifyPwd() *ModifyPwd {
	return &ModifyPwd{client: c}
}

// Password 设置新密码
func (m *ModifyPwd) Password(password string) *ModifyPwd {
	m.req.Password = password
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *ModifyPwd) Do(ctx ...context.Context) (bool, error) {
	if m.client == nil {
		return false, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return false, err
	}
	m.req.Token = token
	if m.req.Password == "" {
		return false, errors.New("新密码是必需的")
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
	res, err := m.client.SecurePost(callCtx, "modifyPwd", m.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
