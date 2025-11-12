package backpack

import (
	"context"
	"fmt"
)

type ExchangeInfo map[string]*Symbol

// Symbol 交易对信息
type Symbol struct {
	Symbol      string  `json:"symbol"`      // 交易对符号，例如 "BTC_USDC"
	BaseSymbol  string  `json:"baseSymbol"`  // 基础货币符号，例如 "BTC"
	QuoteSymbol string  `json:"quoteSymbol"` // 计价货币符号，例如 "USDC"
	MarketType  string  `json:"marketType"`  // 市场类型：SPOT(现货) 或 PERP(永续合约)
	Filters     Filters `json:"filters"`     // 交易过滤器规则

	// 以下字段仅适用于永续合约市场
	ImfFunction           *MarginFunction `json:"imfFunction"`           // 初始保证金函数
	MmfFunction           *MarginFunction `json:"mmfFunction"`           // 维持保证金函数
	FundingInterval       int             `json:"fundingInterval"`       // 资金费率结算间隔（秒）
	FundingRateUpperBound float64         `json:"fundingRateUpperBound"` // 资金费率上限
	FundingRateLowerBound float64         `json:"fundingRateLowerBound"` // 资金费率下限
	OpenInterestLimit     float64         `json:"openInterestLimit"`     // 持仓量限制

	OrderBookState string `json:"orderBookState"` // 订单簿状态：Open(开放)、Closed(关闭)等
	CreatedAt      string `json:"createdAt"`      // 创建时间
	Visible        bool   `json:"visible"`        // 是否可见
}

// Filters 交易过滤器
type Filters struct {
	Price    PriceFilter    `json:"price"`    // 价格过滤器
	Quantity QuantityFilter `json:"quantity"` // 数量过滤器
}

// PriceFilter 价格过滤器
type PriceFilter struct {
	MinPrice float64 `json:"minPrice"` // 最小价格
	MaxPrice float64 `json:"maxPrice"` // 最大价格
	TickSize float64 `json:"tickSize"` // 价格步长（最小价格变动单位）

	// 以下字段仅适用于永续合约市场
	MaxMultiplier               float64      `json:"maxMultiplier"`               // 最大价格乘数
	MinMultiplier               float64      `json:"minMultiplier"`               // 最小价格乘数
	MaxImpactMultiplier         float64      `json:"maxImpactMultiplier"`         // 最大冲击价格乘数
	MinImpactMultiplier         float64      `json:"minImpactMultiplier"`         // 最小冲击价格乘数
	MeanMarkPriceBand           *PriceBand   `json:"meanMarkPriceBand"`           // 平均标记价格波动带
	MeanPremiumBand             *PremiumBand `json:"meanPremiumBand"`             // 平均溢价波动带
	BorrowEntryFeeMaxMultiplier float64      `json:"borrowEntryFeeMaxMultiplier"` // 借贷入场费最大乘数
	BorrowEntryFeeMinMultiplier float64      `json:"borrowEntryFeeMinMultiplier"` // 借贷入场费最小乘数
}

// PriceBand 价格波动带
type PriceBand struct {
	MaxMultiplier float64 `json:"maxMultiplier"` // 最大乘数
	MinMultiplier float64 `json:"minMultiplier"` // 最小乘数
}

// PremiumBand 溢价波动带
type PremiumBand struct {
	TolerancePct float64 `json:"tolerancePct"` // 容忍百分比
}

// QuantityFilter 数量过滤器
type QuantityFilter struct {
	MinQuantity float64 `json:"minQuantity"` // 最小下单数量
	MaxQuantity float64 `json:"maxQuantity"` // 最大下单数量
	StepSize    float64 `json:"stepSize"`    // 数量步长（最小数量变动单位）
}

// MarginFunction 保证金函数
type MarginFunction struct {
	Type   string `json:"type"`   // 函数类型，例如 "sqrt"（平方根）
	Base   string `json:"base"`   // 基础值
	Factor string `json:"factor"` // 因子
}

// InitExchangeInfo 初始化交易对基础信息
func (b *Backpack) InitExchangeInfo(ctx context.Context) error {
	var httpRequest = &HttpRequest{
		backpack: b,
		baseUrl:  b.UrlRest,
		apiUrl:   "/api/v1/markets",            // 请求URL
		sign:     false,                        // 是否需要签名
		params:   make(map[string]interface{}), // 请求参数
		window:   5000,                         // 请求有效时间窗口
	}

	var res = new([]Symbol)

	err := httpRequest.GetJSON(ctx, res)
	if err != nil {
		return fmt.Errorf("获取市场信息失败: %w", err)
	}

	for _, symbol := range *res {

		if symbol.MarketType == "PERP" && symbol.Visible && symbol.OrderBookState == "Open" {
			b.Exc[symbol.BaseSymbol+symbol.QuoteSymbol] = &symbol
		}

	}

	return nil
}
