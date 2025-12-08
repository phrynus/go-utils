package user

import (
	"context"
)

// Info 构建并执行获取信息 API 请求
type Info struct {
	client *Client
	req    InfoRequest
}

// InfoRequest 携带获取信息端点的信息
type InfoRequest struct {
	Token string `json:"token"`
	Time  int64  `json:"time"`
}

// InfoData 用户信息数据
type InfoData struct {
	AcctNo     string                 `json:"acctno"`     // 账号
	Email      string                 `json:"email"`      // 邮箱
	Extend     map[string]interface{} `json:"extend"`     // 扩展信息
	Fen        int                    `json:"fen"`        // 积分
	InvID      int                    `json:"invID"`      // 邀请ID
	Name       string                 `json:"name"`       // 昵称
	Phone      string                 `json:"phone"`      // 手机号
	Pic        string                 `json:"pic"`        // 头像
	UID        int                    `json:"uid"`        // 用户ID
	VipExpDate string                 `json:"vipExpDate"` // VIP到期日期
	VipExpTime int                    `json:"vipExpTime"` // VIP到期时间戳
}

// NewInfo 获取信息
func (c *Client) NewInfo() *Info {
	return &Info{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (i *Info) Do(ctx ...context.Context) (InfoData, error) {
	if i.client == nil {
		return InfoData{}, errNilClient
	}
	token, err := i.client.GetToken()
	if err != nil {
		return InfoData{}, err
	}
	i.req.Token = token
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
	var payload InfoData
	if _, err := i.client.SecurePost(callCtx, "info", i.req, &payload); err != nil {
		return InfoData{}, err
	}
	return payload, nil
}
