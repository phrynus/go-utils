package user

import (
	"context"
	"errors"
)

// QQLoginSDK 构建并执行QQ SDK登录 API 请求
type QQLoginSDK struct {
	client *Client
	req    QQLoginSDKRequest
}

// QQLoginSDKRequest 携带QQ SDK登录端点的信息
type QQLoginSDKRequest struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UDID        string `json:"udid"`
	InvID       int    `json:"invid,omitempty"`
	Time        int64  `json:"time"`
}

// QQLoginSDKData QQ SDK登录数据
type QQLoginSDKData struct {
	Token string   `json:"token"` // 令牌
	State string   `json:"state"` // 状态
	Info  UserInfo `json:"info"`  // 用户信息
}

// NewQQLoginSDK QQ SDK登录
func (c *Client) NewQQLoginSDK() *QQLoginSDK {
	return &QQLoginSDK{client: c}
}

// AccessToken 设置QQ互联SDK返回的access_token
func (q *QQLoginSDK) AccessToken(accessToken string) *QQLoginSDK {
	q.req.AccessToken = accessToken
	return q
}

// OpenID 设置QQ OpenID
func (q *QQLoginSDK) OpenID(openID string) *QQLoginSDK {
	q.req.OpenID = openID
	return q
}

// UDID 设置机器码
func (q *QQLoginSDK) UDID(udid string) *QQLoginSDK {
	q.req.UDID = udid
	return q
}

// InvID 设置邀请人ID
func (q *QQLoginSDK) InvID(invID int) *QQLoginSDK {
	q.req.InvID = invID
	return q
}

// Do 发送请求，可选择性地覆盖 context
func (q *QQLoginSDK) Do(ctx ...context.Context) (QQLoginSDKData, error) {
	if q.client == nil {
		return QQLoginSDKData{}, errNilClient
	}
	if q.req.AccessToken == "" {
		return QQLoginSDKData{}, errors.New("access_token是必需的")
	}
	if q.req.OpenID == "" {
		return QQLoginSDKData{}, errors.New("openid是必需的")
	}
	if q.req.UDID == "" {
		return QQLoginSDKData{}, errors.New("机器码是必需的")
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
	var payload QQLoginSDKData
	if _, err := q.client.SecurePost(callCtx, "qqloginSDK", q.req, &payload); err != nil {
		return QQLoginSDKData{}, err
	}
	// 保存 token 到客户端
	if payload.Token != "" {
		q.client.SetToken(payload.Token)
	}
	return payload, nil
}









