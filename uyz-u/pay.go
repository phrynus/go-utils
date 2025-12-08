package user

import (
	"context"
	"errors"
)

// Pay 构建并执行在线支付 API 请求
type Pay struct {
	client *Client
	req    PayRequest
}

// PayRequest 携带在线支付端点的信息
type PayRequest struct {
	Account string `json:"account,omitempty"`
	Token   string `json:"token,omitempty"`
	GID     int    `json:"gid"`
	Type    string `json:"type,omitempty"`
	Mode    string `json:"mode,omitempty"`
	Time    int64  `json:"time"`
}

// PayData 支付数据
type PayData struct {
	Money   float64 `json:"money"`    // 金额
	Name    string  `json:"name"`     // 商品名称
	OrderNo string  `json:"order_no"` // 订单号
	PayURL  string  `json:"pay_url"`  // 支付地址
}

// NewPay 在线支付
func (c *Client) NewPay() *Pay {
	return &Pay{client: c}
}

// Account 设置充值账号（和token字段二选一）
func (p *Pay) Account(account string) *Pay {
	p.req.Account = account
	return p
}

// GID 设置商品ID
func (p *Pay) GID(gid int) *Pay {
	p.req.GID = gid
	return p
}

// Type 设置支付类型：wx=微信，ali=支付宝
func (p *Pay) Type(typ string) *Pay {
	p.req.Type = typ
	return p
}

// Mode 设置支付模式：h5(仅微信支持),app,qr(微信Native支付,支付宝当面付)
func (p *Pay) Mode(mode string) *Pay {
	p.req.Mode = mode
	return p
}

// Do 发送请求，可选择性地覆盖 context
func (p *Pay) Do(ctx ...context.Context) (PayData, error) {
	if p.client == nil {
		return PayData{}, errNilClient
	}
	// 如果没有设置 account，则从 client 获取 token
	if p.req.Account == "" {
		token, err := p.client.GetToken()
		if err != nil {
			return PayData{}, err
		}
		p.req.Token = token
	}
	if p.req.GID == 0 {
		return PayData{}, errors.New("商品ID是必需的")
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
	var payload PayData
	if _, err := p.client.SecurePost(callCtx, "pay", p.req, &payload); err != nil {
		return PayData{}, err
	}
	return payload, nil
}
