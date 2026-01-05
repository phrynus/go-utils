package user

import (
	"context"
	"errors"
)

// WXLoginWeb 构建并执行微信登录 API 请求
type WXLoginWeb struct {
	client *Client
	req    WXLoginWebRequest
}

// WXLoginWebRequest 携带微信登录端点的信息
type WXLoginWebRequest struct {
	UDID  string `json:"udid"`
	InvID int    `json:"invid,omitempty"`
	Time  int64  `json:"time"`
}

// WXLoginWebData 微信登录数据
type WXLoginWebData struct {
	URL  string `json:"url"`  // 登录URL
	UUID string `json:"uuid"` // 登录标识
}

// NewWXLoginWeb 微信登录
func (c *Client) NewWXLoginWeb() *WXLoginWeb {
	return &WXLoginWeb{client: c}
}

// UDID 设置机器码
func (w *WXLoginWeb) UDID(udid string) *WXLoginWeb {
	w.req.UDID = udid
	return w
}

// InvID 设置邀请人ID
func (w *WXLoginWeb) InvID(invID int) *WXLoginWeb {
	w.req.InvID = invID
	return w
}

// Do 发送请求，可选择性地覆盖 context
func (w *WXLoginWeb) Do(ctx ...context.Context) (WXLoginWebData, error) {
	if w.client == nil {
		return WXLoginWebData{}, errNilClient
	}
	if w.req.UDID == "" {
		return WXLoginWebData{}, errors.New("机器码是必需的")
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
	var payload WXLoginWebData
	if _, err := w.client.SecurePost(callCtx, "wxloginWeb", w.req, &payload); err != nil {
		return WXLoginWebData{}, err
	}
	return payload, nil
}









