package user

import (
	"context"
	"errors"
)

// MessageList 构建并执行留言列表 API 请求
type MessageList struct {
	client *Client
	req    MessageListRequest
}

// MessageListRequest 携带留言列表端点的信息
type MessageListRequest struct {
	Token string `json:"token"`
	Page  int    `json:"pg"`
}

// MessageListData 留言列表数据
type MessageListData struct {
	CurrentPage int           `json:"currentPage"` // 当前页码
	DataTotal   int           `json:"dataTotal"`   // 数据总数
	List        []MessageItem `json:"list"`        // 留言列表
	PageTotal   int           `json:"pageTotal"`   // 总页数
}

// MessageItem 留言项
type MessageItem struct {
	ID       int    `json:"id"`        // 留言ID
	Title    string `json:"title"`     // 标题
	Time     int64  `json:"time"`      // 创建时间
	LastTime int64  `json:"last_time"` // 最后回复时间
	State    int    `json:"state"`     // 状态
}

// NewMessageList 留言列表
func (c *Client) NewMessageList() *MessageList {
	return &MessageList{client: c}
}

// Page 设置页码
func (m *MessageList) Page(page int) *MessageList {
	m.req.Page = page
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *MessageList) Do(ctx ...context.Context) (MessageListData, error) {
	if m.client == nil {
		return MessageListData{}, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return MessageListData{}, err
	}
	m.req.Token = token
	if m.req.Page == 0 {
		return MessageListData{}, errors.New("页码是必需的")
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
	var payload MessageListData
	if _, err := m.client.SecurePost(callCtx, "messageList", m.req, &payload); err != nil {
		return MessageListData{}, err
	}
	return payload, nil
}
