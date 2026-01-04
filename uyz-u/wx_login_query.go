package user

import (
	"context"
	"errors"
)

// WXLoginQuery 构建并执行微信登录状态查询 API 请求
type WXLoginQuery struct {
	client *Client
	req    WXLoginQueryRequest
}

// WXLoginQueryRequest 携带微信登录状态查询端点的信息
type WXLoginQueryRequest struct {
	UUID string `json:"uuid"`
	Time int64  `json:"time"`
}

// WXLoginQueryData 微信登录状态查询数据
type WXLoginQueryData struct {
	Token string   `json:"token"` // 令牌
	State string   `json:"state"` // 状态
	Info  UserInfo `json:"info"`  // 用户信息
}

// NewWXLoginQuery 微信登录状态查询
func (c *Client) NewWXLoginQuery() *WXLoginQuery {
	return &WXLoginQuery{client: c}
}

// UUID 设置微信登录标识
func (w *WXLoginQuery) UUID(uuid string) *WXLoginQuery {
	w.req.UUID = uuid
	return w
}

// Do 发送请求，可选择性地覆盖 context
func (w *WXLoginQuery) Do(ctx ...context.Context) (WXLoginQueryData, error) {
	if w.client == nil {
		return WXLoginQueryData{}, errNilClient
	}
	if w.req.UUID == "" {
		return WXLoginQueryData{}, errors.New("微信登录标识是必需的")
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
	var payload WXLoginQueryData
	if _, err := w.client.SecurePost(callCtx, "wxloginQuery", w.req, &payload); err != nil {
		return WXLoginQueryData{}, err
	}
	// 保存 token 到客户端
	if payload.Token != "" {
		w.client.SetToken(payload.Token)
	}
	return payload, nil
}






