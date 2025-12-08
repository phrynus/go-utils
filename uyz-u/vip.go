package user

import (
	"context"
)

// VIP 构建并执行会员验证 API 请求
type VIP struct {
	client *Client
	req    VIPRequest
}

// VIPRequest 携带会员验证端点的信息
type VIPRequest struct {
	Token string `json:"token"`
}

// NewVIP 会员验证
func (c *Client) NewVIP() *VIP {
	return &VIP{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (v *VIP) Do(ctx ...context.Context) (bool, error) {
	if v.client == nil {
		return false, errNilClient
	}
	token, err := v.client.GetToken()
	if err != nil {
		return false, err
	}
	v.req.Token = token
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

	res, err := v.client.SecurePost(callCtx, "vip", v.req, nil)

	return res.Code == 0, err
}
