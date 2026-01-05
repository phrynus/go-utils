// 获取合约交易对基础信息
package gate

import (
	"context"
	"time"
)

type ExchangeInfo map[string]*Symbol

type Symbol struct {
	Name                  string  `json:"name"`                    // 合约标识
	Type                  string  `json:"type"`                    // 合约类型, inverse - 反向合约, direct - 正向合约
	QuantoMultiplier      float64 `json:"quanto_multiplier"`       // 每张数量
	RefDiscountRate       float64 `json:"ref_discount_rate"`       // 被推荐人享受交易费率折扣
	OrderPriceDeviate     float64 `json:"order_price_deviate"`     // 下单价与当前标记价格允许的正负偏移量
	MaintenanceRate       float64 `json:"maintenance_rate"`        // 维持保证金比例
	MarkType              string  `json:"mark_type"`               // 价格标记方式, internal - 内盘成交价格, index - 外部指数价格
	LastPrice             float64 `json:"last_price"`              // 上一次成交价格
	MarkPrice             float64 `json:"mark_price"`              // 当前标记价格
	IndexPrice            float64 `json:"index_price"`             // 当前指数价格
	FundingRateIndicative float64 `json:"funding_rate_indicative"` // 预测资金费率
	MarkPriceRound        float64 `json:"mark_price_round"`        // 标记、强平等价格最小单位
	FundingOffset         int     `json:"funding_offset"`          // 资金费率偏移量
	InDelisting           bool    `json:"in_delisting"`            // 是否处于下线过渡期/下线状态
	RiskLimitBase         float64 `json:"risk_limit_base"`         // 基础风险限额,已废弃
	InterestRate          float64 `json:"interest_rate"`           // 利率
	OrderPriceRound       float64 `json:"order_price_round"`       // 委托价格最小单位
	OrderSizeMin          int64   `json:"order_size_min"`          // 最小下单数量
	RefRebateRate         float64 `json:"ref_rebate_rate"`         // 推荐人享受交易费率返佣比例
	FundingInterval       int     `json:"funding_interval"`        // 资金费率应用间隔 (秒)
	RiskLimitStep         float64 `json:"risk_limit_step"`         // 风险限额调整步长,已废弃
	LeverageMin           float64 `json:"leverage_min"`            // 最小杠杆
	LeverageMax           float64 `json:"leverage_max"`            // 最大杠杆
	CrossLeverageDefault  float64 `json:"cross_leverage_default"`  // 默认逐仓杠杆
	RiskLimitMax          float64 `json:"risk_limit_max"`          // 合约允许的最大风险限额,已废弃
	MakerFeeRate          float64 `json:"maker_fee_rate"`          // 挂单成交的手续费率，负数代表返佣
	TakerFeeRate          float64 `json:"taker_fee_rate"`          // 吃单成交的手续费率
	FundingRate           float64 `json:"funding_rate"`            // 当前资金费率
	OrderSizeMax          int64   `json:"order_size_max"`          // 最大下单数量
	FundingNextApply      int64   `json:"funding_next_apply"`      // 下次资金费率应用时间
	ShortUsers            int     `json:"short_users"`             // 做空用户数
	ConfigChangeTime      int64   `json:"config_change_time"`      // 配置变更时间
	LongUsers             int     `json:"long_users"`              // 做多用户数
	FundingImpactValue    float64 `json:"funding_impact_value"`    // 资金费率影响值
	OrdersLimit           int     `json:"orders_limit"`            // 最多挂单数量
	TradeID               int64   `json:"trade_id"`                // 最近成交ID
	TradeSize             int64   `json:"trade_size"`              // 历史累计成交量
	OrderbookID           int64   `json:"orderbook_id"`            // 订单簿ID
	PositionSize          int64   `json:"position_size"`           // 当前做多用户持有仓位总和
	EnableBonus           bool    `json:"enable_bonus"`            // 是否支持体验金
	EnableCredit          bool    `json:"enable_credit"`           // 是否支持统一账户
	CreateTime            int64   `json:"create_time"`             // 合约创建时间
	FundingCapRatio       float64 `json:"funding_cap_ratio"`       // 资金费率上限的系数
	VoucherLeverage       float64 `json:"voucher_leverage"`        // 优惠券杠杆
	Status                string  `json:"status"`                  // 合约状态: prelaunch, trading, delisting, delisted, circuit_breaker
	IsPreMarket           bool    `json:"is_pre_market"`           // 是否为预市
	LaunchTime            int64   `json:"launch_time"`             // 合约开盘时间
	EnableCircuitBreaker  bool    `json:"enable_circuit_breaker"`  // 是否启用熔断
	FundingRateLimit      float64 `json:"funding_rate_limit"`      // 资金费率上限
	DelistingTime         int64   `json:"delisting_time"`          // 合约进入只减仓状态时间
	DelistedTime          int64   `json:"delisted_time"`           // 合约下架时间
}

func (g *Gate) InitExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	var httpRequest = &HttpRequest{
		gate:        g,                            // Gate客户端实例
		baseUrl:     g.UrlRest,                    // 基础URL
		apiUrl:      "/futures/usdt/contracts",    // 请求URL
		sign:        false,                        // 是否签名
		isTimestamp: false,                        // 是否时间戳
		params:      make(map[string]interface{}), // 请求参数
	}

	var res = make([]Symbol, 0)

	err := httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return &ExchangeInfo{}, err
	}

	// 初次加载并安全写入 g.Exc
	g.ExcMu.Lock()
	for _, symbol := range res {
		if symbol.Status == "trading" {
			s := symbol
			g.Exc[s.Name] = &s
		}
	}
	g.ExcMu.Unlock()

	// 启动后台每小时自动更新（由传入 ctx 控制取消）
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var res2 = make([]Symbol, 0)
				var httpRequest2 = &HttpRequest{
					gate:        g,
					baseUrl:     g.UrlRest,
					apiUrl:      "/futures/usdt/contracts",
					sign:        false,
					isTimestamp: false,
					params:      make(map[string]interface{}),
				}
				if err := httpRequest2.GetJSON(ctx, &res2); err != nil {
					// 获取失败则本轮跳过
					continue
				}
				g.ExcMu.Lock()
				for _, symbol := range res2 {
					if symbol.Status == "trading" {
						s := symbol
						g.Exc[s.Name] = &s
					}
				}
				g.ExcMu.Unlock()
			}
		}
	}()

	return &g.Exc, nil
}
