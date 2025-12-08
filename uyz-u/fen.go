package user

import (
	"context"
	"errors"
	"strconv"
)

// Fen 构建并执行积分验证 API 请求
type Fen struct {
	client *Client
	req    FenRequest
}

// NewFen 积分验证
func (c *Client) NewFen() *Fen {
	return &Fen{client: c}
}

// FenID 设置积分事件ID
func (f *Fen) FenID(fenID int) *Fen {
	f.req.FenID = fenID
	return f
}

// FenMark 设置积分事件标记
func (f *Fen) FenMark(fenMark string) *Fen {
	f.req.FenMark = fenMark
	return f
}

// Do 发送请求，可选择性地覆盖 context
func (f *Fen) Do(ctx ...context.Context) (bool, error) {
	if f.client == nil {
		return false, errNilClient
	}
	token, err := f.client.GetToken()
	if err != nil {
		return false, err
	}
	f.req.Token = token
	if f.req.FenID == 0 {
		return false, errors.New("积分事件ID是必需的")
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
	res, err := f.client.SecurePost(callCtx, "fen", f.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
