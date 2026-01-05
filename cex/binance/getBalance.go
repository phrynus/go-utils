// 获取账户余额
package binance

import "context"

// GetBalance 余额查询请求
type GetBalance struct {
	client      *Binance
	httpRequest *HttpRequest
}

// Balance 账户余额列表
type Balance []struct {
	AccountAlias       string  `json:"accountAlias"`              // 账户别名
	Asset              string  `json:"asset"`                     // 资产
	Balance            float64 `json:"balance,string"`            // 余额
	CrossWalletBalance float64 `json:"crossWalletBalance,string"` // 全仓钱包余额
	CrossUnPnl         float64 `json:"crossUnPnl,string"`         // 全仓未实现盈亏
	AvailableBalance   float64 `json:"availableBalance,string"`   // 可用余额
	MaxWithdrawAmount  float64 `json:"maxWithdrawAmount,string"`  // 最大提取额
	MarginAvailable    bool    `json:"marginAvailable"`           // 可用于联合保证金
	UpdateTime         int64   `json:"updateTime"`                // 更新时间
}

// NewGetBalance 创建余额查询请求
func (b *Binance) NewGetBalance() *GetBalance {
	return &GetBalance{
		client: b,
		httpRequest: &HttpRequest{
			binance:     b,
			baseUrl:     b.UrlRest,
			apiUrl:      "/fapi/v3/balance",
			sign:        true,
			isTimestamp: true,
			params:      make(map[string]interface{}),
		},
	}
}

// Do 执行请求并缓存余额
func (t *GetBalance) Do(ctx context.Context) (*Balance, error) {
	balance := new(Balance)
	if err := t.httpRequest.GetJSON(ctx, balance); err != nil {
		return &Balance{}, err
	}
	t.client.Balance = balance
	return balance, nil
}
