package user

import (
	"context"
	"errors"
	"strconv"
)

// QQBindSDK 构建并执行QQ SDK绑定 API 请求
type QQBindSDK struct {
	client *Client
	req    QQBindSDKRequest
}

// QQBindSDKRequest 携带QQ SDK绑定端点的信息
type QQBindSDKRequest struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
}

// NewQQBindSDK QQ SDK绑定
func (c *Client) NewQQBindSDK() *QQBindSDK {
	return &QQBindSDK{client: c}
}

// AccessToken 设置QQ互联SDK返回的access_token
func (q *QQBindSDK) AccessToken(accessToken string) *QQBindSDK {
	q.req.AccessToken = accessToken
	return q
}

// OpenID 设置QQ OpenID
func (q *QQBindSDK) OpenID(openID string) *QQBindSDK {
	q.req.OpenID = openID
	return q
}

// Do 发送请求，可选择性地覆盖 context
func (q *QQBindSDK) Do(ctx ...context.Context) (bool, error) {
	if q.client == nil {
		return false, errNilClient
	}
	token, err := q.client.GetToken()
	if err != nil {
		return false, err
	}
	q.req.Token = token
	if q.req.AccessToken == "" {
		return false, errors.New("access_token是必需的")
	}
	if q.req.OpenID == "" {
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
	res, err := q.client.SecurePost(callCtx, "qqBindSDK", q.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}






