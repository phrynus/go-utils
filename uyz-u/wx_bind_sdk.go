package user

import (
	"context"
	"errors"
	"strconv"
)

// WXBindSDK 构建并执行微信 SDK绑定 API 请求
type WXBindSDK struct {
	client *Client
	req    WXBindSDKRequest
}

// WXBindSDKRequest 携带微信 SDK绑定端点的信息
type WXBindSDKRequest struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
}

// NewWXBindSDK 微信 SDK绑定
func (c *Client) NewWXBindSDK() *WXBindSDK {
	return &WXBindSDK{client: c}
}

// AccessToken 设置微信互联SDK返回的access_token
func (w *WXBindSDK) AccessToken(accessToken string) *WXBindSDK {
	w.req.AccessToken = accessToken
	return w
}

// OpenID 设置微信 OpenID
func (w *WXBindSDK) OpenID(openID string) *WXBindSDK {
	w.req.OpenID = openID
	return w
}

// Do 发送请求，可选择性地覆盖 context
func (w *WXBindSDK) Do(ctx ...context.Context) (bool, error) {
	if w.client == nil {
		return false, errNilClient
	}
	token, err := w.client.GetToken()
	if err != nil {
		return false, err
	}
	w.req.Token = token
	if w.req.AccessToken == "" {
		return false, errors.New("access_token是必需的")
	}
	if w.req.OpenID == "" {
		return false, errors.New("openid是必需的")
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
	res, err := w.client.SecurePost(callCtx, "wxBindSDK", w.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}








