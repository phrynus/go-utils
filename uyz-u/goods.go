package user

import (
	"context"
	"errors"
)

// Goods 构建并执行商品列表 API 请求
type Goods struct {
	client *Client
	req    GoodsRequest
}

// GoodsRequest 携带商品列表端点的信息
type GoodsRequest struct {
	Page int `json:"pg"`
}

// GoodsListData 商品列表数据
type GoodsListData struct {
	CurrentPage int         `json:"currentPage"` // 当前页码
	DataTotal   int         `json:"dataTotal"`   // 数据总数
	List        []GoodsItem `json:"list"`        // 商品列表
	PageTotal   int         `json:"pageTotal"`   // 总页数
}

// GoodsItem 商品项
type GoodsItem struct {
	ID    int     `json:"id"`    // 商品ID
	Name  string  `json:"name"`  // 商品名称
	Type  string  `json:"type"`  // 商品类型
	Money float64 `json:"money"` // 价格
	Blurb string  `json:"blurb"` // 简介
}

// NewGoods 商品列表
func (c *Client) NewGoods() *Goods {
	return &Goods{client: c}
}

// Page 设置页码
func (g *Goods) Page(page int) *Goods {
	g.req.Page = page
	return g
}

// Do 发送请求，可选择性地覆盖 context
func (g *Goods) Do(ctx ...context.Context) (GoodsListData, error) {
	if g.client == nil {
		return GoodsListData{}, errNilClient
	}
	if g.req.Page == 0 {
		return GoodsListData{}, errors.New("页码是必需的")
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
	var payload GoodsListData
	if _, err := g.client.SecurePost(callCtx, "goods", g.req, &payload); err != nil {
		return GoodsListData{}, err
	}
	return payload, nil
}
