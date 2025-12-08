package user

import (
	"context"
	"errors"
	"strconv"
)

// SetAcctno 构建并执行设置账号 API 请求
type SetAcctno struct {
	client *Client
	req    SetAcctnoRequest
}

// SetAcctnoRequest 携带设置账号端点的信息
type SetAcctnoRequest struct {
	Token  string `json:"token"`
	Acctno string `json:"acctno"`
}

// NewSetAcctno 设置账号
func (c *Client) NewSetAcctno() *SetAcctno {
	return &SetAcctno{client: c}
}

// Acctno 设置账号
func (s *SetAcctno) Acctno(acctno string) *SetAcctno {
	s.req.Acctno = acctno
	return s
}

// Do 发送请求，可选择性地覆盖 context
func (s *SetAcctno) Do(ctx ...context.Context) (bool, error) {
	if s.client == nil {
		return false, errNilClient
	}
	token, err := s.client.GetToken()
	if err != nil {
		return false, err
	}
	s.req.Token = token
	if s.req.Acctno == "" {
		return false, errors.New("账号是必需的")
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
	res, err := s.client.SecurePost(callCtx, "setAcctno", s.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
