package user

import (
	"context"
	"errors"
)

// QQLoginQuery 构建并执行QQ登录状态查询 API 请求
type QQLoginQuery struct {
	client *Client
	req    QQLoginQueryRequest
}

// QQLoginQueryRequest 携带QQ登录状态查询端点的信息
type QQLoginQueryRequest struct {
	UUID string `json:"uuid"`
	Time int64  `json:"time"`
}

// QQLoginQueryData QQ登录状态查询数据
type QQLoginQueryData struct {
	Token string   `json:"token"` // 令牌
	State string   `json:"state"` // 状态
	Info  UserInfo `json:"info"`  // 用户信息
}

// NewQQLoginQuery QQ登录状态查询
func (c *Client) NewQQLoginQuery() *QQLoginQuery {
	return &QQLoginQuery{client: c}
}

// UUID 设置QQ登录标识
func (q *QQLoginQuery) UUID(uuid string) *QQLoginQuery {
	q.req.UUID = uuid
	return q
}

// Do 发送请求，可选择性地覆盖 context
func (q *QQLoginQuery) Do(ctx ...context.Context) (QQLoginQueryData, error) {
	if q.client == nil {
		return QQLoginQueryData{}, errNilClient
	}
	if q.req.UUID == "" {
		return QQLoginQueryData{}, errors.New("QQ登录标识是必需的")
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
	var payload QQLoginQueryData
	if _, err := q.client.SecurePost(callCtx, "qqloginQuery", q.req, &payload); err != nil {
		return QQLoginQueryData{}, err
	}
	// 保存 token 到客户端
	if payload.Token != "" {
		q.client.SetToken(payload.Token)
	}
	return payload, nil
}






