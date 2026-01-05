// 获取合约账号
package gate

import "context"

// GetBalance 获取合约账号请求
type GetBalance struct {
	gate        *Gate
	httpRequest *HttpRequest
}

// Balance 账户信息
type Balance struct {
	Total                  float64 `json:"total,string"`                    // 钱包余额
	UnrealisedPnl          float64 `json:"unrealised_pnl,string"`           // 未实现盈亏
	PositionMargin         float64 `json:"position_margin,string"`          // 已废弃
	OrderMargin            float64 `json:"order_margin,string"`             // 订单保证金
	Available              float64 `json:"available,string"`                // 可用余额
	Point                  float64 `json:"point,string"`                    // 点卡数额
	Currency               string  `json:"currency"`                        // 结算币种
	InDualMode             bool    `json:"in_dual_mode"`                    // 双向持仓
	PositionMode           string  `json:"position_mode"`                   // 持仓模式
	EnableCredit           bool    `json:"enable_credit"`                   // 统一账户模式
	PositionInitialMargin  float64 `json:"position_initial_margin,string"`  // 下寸保证金
	MaintenanceMargin      float64 `json:"maintenance_margin,string"`       // 维持保证金
	Bonus                  float64 `json:"bonus,string"`                    // 体验金
	EnableEvolvedClassic   bool    `json:"enable_evolved_classic"`          // 已废弃
	CrossOrderMargin       float64 `json:"cross_order_margin,string"`       // 全仓订单保证金
	CrossInitialMargin     float64 `json:"cross_initial_margin,string"`     // 全仓初始保证金
	CrossMaintenanceMargin float64 `json:"cross_maintenance_margin,string"` // 全仓维持保证金
	CrossUnrealisedPnl     float64 `json:"cross_unrealised_pnl,string"`     // 全仓未实现盈亏
	CrossAvailable         float64 `json:"cross_available,string"`          // 全仓可用余额
	CrossMarginBalance     float64 `json:"cross_margin_balance,string"`     // 全仓保证金余额
	CrossMmr               float64 `json:"cross_mmr,string"`                // 全仓维持保证金率
	CrossImr               float64 `json:"cross_imr,string"`                // 全仓初始保证金率
	IsolatedPositionMargin float64 `json:"isolated_position_margin,string"` // 逐仓保证金
	EnableNewDualMode      bool    `json:"enable_new_dual_mode"`            // 已废弃
	MarginMode             int     `json:"margin_mode"`                     // 保证金模式
	EnableTieredMm         bool    `json:"enable_tiered_mm"`                // 梯度式维持保证金
	History                struct {
		Dnw         float64 `json:"dnw,string"`          // 累计转入转出
		Pnl         float64 `json:"pnl,string"`          // 累计盈亏
		Fee         float64 `json:"fee,string"`          // 累计手续费
		Refr        float64 `json:"refr,string"`         // 累计推荐返佣
		Fund        float64 `json:"fund,string"`         // 累计资金费用
		PointDnw    float64 `json:"point_dnw,string"`    // 累计点卡转入转出
		PointFee    float64 `json:"point_fee,string"`    // 累计点卡手续费
		PointRefr   float64 `json:"point_refr,string"`   // 累计点卡返佣
		BonusDnw    float64 `json:"bonus_dnw,string"`    // 累计体验金转入转出
		BonusOffset float64 `json:"bonus_offset,string"` // 累计体验金抵扣
	} `json:"history"`
}

// NewGetBalance 创建获取合约账号请求
func (g *Gate) NewGetBalance() *GetBalance {

	return &GetBalance{
		gate: g,
		httpRequest: &HttpRequest{
			gate:        g,
			baseUrl:     g.UrlRest,
			apiUrl:      "/futures/usdt/accounts",
			sign:        true, // 需要签名
			isTimestamp: true, // 需要时间戳
			params:      make(map[string]interface{}),
		},
	}
}

// Do 执行请求并返回合约账号信息
func (t *GetBalance) Do(ctx context.Context) (*Balance, error) {
	balance := new(Balance)
	if err := t.httpRequest.GetJSON(ctx, balance); err != nil {
		return nil, err
	}
	return balance, nil
}
