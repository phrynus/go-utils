package user

import (
	"context"
	"errors"
)

// WXLoginSDK 构建并执行微信 SDK登录 API 请求
type WXLoginSDK struct {
	client *Client
	req    WXLoginSDKRequest
}

// WXLoginSDKRequest 携带微信 SDK登录端点的信息
type WXLoginSDKRequest struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UDID        string `json:"udid"`
	InvID       int    `json:"invid,omitempty"`
	Time        int64  `json:"time"`
}

// WXLoginSDKData 微信 SDK登录数据
type WXLoginSDKData struct {
	Token string   `json:"token"` // 令牌
	State string   `json:"state"` // 状态
	Info  UserInfo `json:"info"`  // 用户信息
}

// NewWXLoginSDK 微信 SDK登录
func (c *Client) NewWXLoginSDK() *WXLoginSDK {
	return &WXLoginSDK{client: c}
}

// AccessToken 设置微信互联SDK返回的access_token
func (w *WXLoginSDK) AccessToken(accessToken string) *WXLoginSDK {
	w.req.AccessToken = accessToken
	return w
}

// OpenID 设置微信 OpenID
func (w *WXLoginSDK) OpenID(openID string) *WXLoginSDK {
	w.req.OpenID = openID
	return w
}

// UDID 设置机器码
func (w *WXLoginSDK) UDID(udid string) *WXLoginSDK {
	w.req.UDID = udid
	return w
}

// InvID 设置邀请人ID
func (w *WXLoginSDK) InvID(invID int) *WXLoginSDK {
	w.req.InvID = invID
	return w
}

// Do 发送请求，可选择性地覆盖 context
func (w *WXLoginSDK) Do(ctx ...context.Context) (WXLoginSDKData, error) {
	if w.client == nil {
		return WXLoginSDKData{}, errNilClient
	}
	if w.req.AccessToken == "" {
		return WXLoginSDKData{}, errors.New("access_token是必需的")
	}
	if w.req.OpenID == "" {
		return WXLoginSDKData{}, errors.New("openid是必需的")
	}
	if w.req.UDID == "" {
		return WXLoginSDKData{}, errors.New("机器码是必需的")
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
	var payload WXLoginSDKData
	if _, err := w.client.SecurePost(callCtx, "wxloginSDK", w.req, &payload); err != nil {
		return WXLoginSDKData{}, err
	}
	// 保存 token 到客户端
	if payload.Token != "" {
		w.client.SetToken(payload.Token)
	}
	return payload, nil
}







