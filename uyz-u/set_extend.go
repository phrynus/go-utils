package user

import (
	"context"
	"errors"
	"strconv"
)

// SetExtend 构建并执行设置扩展信息 API 请求
type SetExtend struct {
	client *Client
	req    SetExtendRequest
}

// SetExtendRequest 携带设置扩展信息端点的信息
type SetExtendRequest struct {
	Token string `json:"token"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Time  int64  `json:"time"`
}

// NewSetExtend 设置扩展信息
func (c *Client) NewSetExtend() *SetExtend {
	return &SetExtend{client: c}
}

// Key 设置键
func (s *SetExtend) Key(key string) *SetExtend {
	s.req.Key = key
	return s
}

// Value 设置值
func (s *SetExtend) Value(value string) *SetExtend {
	s.req.Value = value
	return s
}

// Do 发送请求，可选择性地覆盖 context
func (s *SetExtend) Do(ctx ...context.Context) (bool, error) {
	if s.client == nil {
		return false, errNilClient
	}
	token, err := s.client.GetToken()
	if err != nil {
		return false, err
	}
	s.req.Token = token
	if s.req.Key == "" {
		return false, errors.New("键是必需的")
	}
	if s.req.Value == "" {
		return false, errors.New("值是必需的")
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
	res, err := s.client.SecurePost(callCtx, "setExtend", s.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
