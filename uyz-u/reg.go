package user

import (
	"context"
	"errors"
	"strconv"
)

// Register 构建并执行注册 API 请求
type Register struct {
	client *Client
	req    RegisterRequest
}

// RegisterRequest 携带注册端点的信息
type RegisterRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	UDID     string `json:"udid"`
	InvID    int    `json:"invid,omitempty"`
	Code     string `json:"code,omitempty"`
	Time     int64  `json:"time"`
}

// NewRegister 注册
func (c *Client) NewRegister() *Register {
	return &Register{client: c}
}

// Account 设置账户标识符
func (r *Register) Account(account string) *Register {
	r.req.Account = account
	return r
}

// Password 设置密码
func (r *Register) Password(password string) *Register {
	r.req.Password = password
	return r
}

// UDID 设置设备标识符
func (r *Register) UDID(udid string) *Register {
	r.req.UDID = udid
	return r
}

// InvID 设置邀请人ID
func (r *Register) InvID(invID int) *Register {
	r.req.InvID = invID
	return r
}

// Code 设置验证码
func (r *Register) Code(code string) *Register {
	r.req.Code = code
	return r
}

// Do 发送请求，可选择性地覆盖 context
func (r *Register) Do(ctx ...context.Context) (bool, error) {
	if r.client == nil {
		return false, errNilClient
	}
	if r.req.Account == "" {
		return false, errors.New("账户是必需的")
	}
	if r.req.Password == "" {
		return false, errors.New("密码是必需的")
	}
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
	res, err := r.client.SecurePost(callCtx, "reg", r.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
