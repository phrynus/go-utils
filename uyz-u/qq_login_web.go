package user

import (
	"context"
	"errors"
)

// QQLoginWeb 构建并执行QQ网页登录 API 请求
type QQLoginWeb struct {
	client *Client
	req    QQLoginWebRequest
}

// QQLoginWebRequest 携带QQ网页登录端点的信息
type QQLoginWebRequest struct {
	UDID  string `json:"udid"`
	InvID int    `json:"invid,omitempty"`
	Time  int64  `json:"time"`
}

// QQLoginWebData QQ网页登录数据
type QQLoginWebData struct {
	URL  string `json:"url"`  // 登录URL
	UUID string `json:"uuid"` // 登录标识
}

// NewQQLoginWeb QQ网页登录
func (c *Client) NewQQLoginWeb() *QQLoginWeb {
	return &QQLoginWeb{client: c}
}

// UDID 设置机器码
func (q *QQLoginWeb) UDID(udid string) *QQLoginWeb {
	q.req.UDID = udid
	return q
}

// InvID 设置邀请人ID
func (q *QQLoginWeb) InvID(invID int) *QQLoginWeb {
	q.req.InvID = invID
	return q
}

// Do 发送请求，可选择性地覆盖 context
func (q *QQLoginWeb) Do(ctx ...context.Context) (QQLoginWebData, error) {
	if q.client == nil {
		return QQLoginWebData{}, errNilClient
	}
	if q.req.UDID == "" {
		return QQLoginWebData{}, errors.New("机器码是必需的")
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
	var payload QQLoginWebData
	if _, err := q.client.SecurePost(callCtx, "qqloginWeb", q.req, &payload); err != nil {
		return QQLoginWebData{}, err
	}
	return payload, nil
}








