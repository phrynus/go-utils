package user

import (
	"context"
	"errors"
)

// OrderQuery 构建并执行订单查询 API 请求
type OrderQuery struct {
	client *Client
	req    OrderQueryRequest
}

// OrderQueryRequest 携带订单查询端点的信息
type OrderQueryRequest struct {
	Order string `json:"order"`
}

// OrderQueryData 订单查询数据
type OrderQueryData struct {
	OrderNo string  `json:"order_no"` // 订单号
	TradeNo string  `json:"trade_no"` // 交易号
	Name    string  `json:"name"`     // 商品名称
	Payment string  `json:"payment"`  // 支付方式
	Money   float64 `json:"money"`    // 金额
	AddTime int64   `json:"add_time"` // 创建时间
	EndTime int64   `json:"end_time"` // 完成时间
	State   int     `json:"state"`    // 状态
}

// NewOrderQuery 订单查询
func (c *Client) NewOrderQuery() *OrderQuery {
	return &OrderQuery{client: c}
}

// Order 设置订单号
func (o *OrderQuery) Order(order string) *OrderQuery {
	o.req.Order = order
	return o
}

// Do 发送请求，可选择性地覆盖 context
func (o *OrderQuery) Do(ctx ...context.Context) (OrderQueryData, error) {
	if o.client == nil {
		return OrderQueryData{}, errNilClient
	}
	if o.req.Order == "" {
		return OrderQueryData{}, errors.New("订单号是必需的")
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
	var payload OrderQueryData
	if _, err := o.client.SecurePost(callCtx, "orderQuery", o.req, &payload); err != nil {
		return OrderQueryData{}, err
	}
	return payload, nil
}
