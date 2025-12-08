package user

import (
	"context"
	"errors"
)

// OrderList 构建并执行订单列表 API 请求
type OrderList struct {
	client *Client
	req    OrderListRequest
}

// OrderListRequest 携带订单列表端点的信息
type OrderListRequest struct {
	Token string `json:"token"`
	Page  int    `json:"pg"`
}

// OrderListData 订单列表数据
type OrderListData struct {
	CurrentPage int         `json:"currentPage"` // 当前页码
	DataTotal   int         `json:"dataTotal"`   // 数据总数
	List        []OrderItem `json:"list"`        // 订单列表
	PageTotal   int         `json:"pageTotal"`   // 总页数
}

// OrderItem 订单项
type OrderItem struct {
	OrderNo string  `json:"order_no"` // 订单号
	TradeNo string  `json:"trade_no"` // 交易号
	Name    string  `json:"name"`     // 商品名称
	Payment string  `json:"payment"`  // 支付方式
	Money   float64 `json:"money"`    // 金额
	AddTime int64   `json:"add_time"` // 创建时间
	EndTime int64   `json:"end_time"` // 完成时间
	State   int     `json:"state"`    // 状态
}

// NewOrderList 订单列表
func (c *Client) NewOrderList() *OrderList {
	return &OrderList{client: c}
}

// Page 设置页码
func (o *OrderList) Page(page int) *OrderList {
	o.req.Page = page
	return o
}

// Do 发送请求，可选择性地覆盖 context
func (o *OrderList) Do(ctx ...context.Context) (OrderListData, error) {
	if o.client == nil {
		return OrderListData{}, errNilClient
	}
	token, err := o.client.GetToken()
	if err != nil {
		return OrderListData{}, err
	}
	o.req.Token = token
	if o.req.Page == 0 {
		return OrderListData{}, errors.New("页码是必需的")
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
	var payload OrderListData
	if _, err := o.client.SecurePost(callCtx, "order", o.req, &payload); err != nil {
		return OrderListData{}, err
	}
	return payload, nil
}
