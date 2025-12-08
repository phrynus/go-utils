package user

import (
	"context"
)

// GetUDID 构建并执行获取已绑定设备列表 API 请求
type GetUDID struct {
	client *Client
	req    GetUDIDRequest
}

// GetUDIDRequest 携带获取已绑定设备列表端点的信息
type GetUDIDRequest struct {
	Token string `json:"token"`
}

// DeviceItem 设备项
type DeviceItem struct {
	Time int64  `json:"time"` // 绑定时间
	UDID string `json:"udid"` // 设备标识
}

// NewGetUDID 获取已绑定设备列表
func (c *Client) NewGetUDID() *GetUDID {
	return &GetUDID{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (g *GetUDID) Do(ctx ...context.Context) ([]DeviceItem, error) {
	if g.client == nil {
		return nil, errNilClient
	}
	token, err := g.client.GetToken()
	if err != nil {
		return nil, err
	}
	g.req.Token = token
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
	var payload []DeviceItem
	if _, err := g.client.SecurePost(callCtx, "getUdid", g.req, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
