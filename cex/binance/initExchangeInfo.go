package binance

import (
	"context"

	"github.com/phrynus/go-utils"
)

type ExchangeInfo map[string]*Symbol

// "symbol": "BLZUSDT",  // 交易对
//  			"pair": "BLZUSDT",  // 标的交易对
//  			"contractType": "PERPETUAL",	// 合约类型
//  			"deliveryDate": 4133404800000,  // 交割日期
//  			"onboardDate": 1598252400000,	  // 上线日期
//  			"status": "TRADING",  // 交易对状态
//  			"maintMarginPercent": "2.5000",  // 请忽略
//  			"requiredMarginPercent": "5.0000", // 请忽略
//  			"baseAsset": "BLZ",  // 标的资产
//  			"quoteAsset": "USDT", // 报价资产
//  			"marginAsset": "USDT", // 保证金资产
//  			"pricePrecision": 5,  // 价格小数点位数(仅作为系统精度使用，注意同tickSize 区分）
//  			"quantityPrecision": 0,  // 数量小数点位数(仅作为系统精度使用，注意同stepSize 区分）
//  			"baseAssetPrecision": 8,  // 标的资产精度
//  			"quotePrecision": 8,  // 报价资产精度
//  			"underlyingType": "COIN",
//  			"underlyingSubType": ["STORAGE"],
//  			"settlePlan": 0,
//  			"triggerProtect": "0.15", // 开启"priceProtect"的条件订单的触发阈值

// [
// 	{
// 		"filterType": "PRICE_FILTER", // 价格限制
// 		"maxPrice": "300", // 价格上限, 最大价格
// 		"minPrice": "0.0001", // 价格下限, 最小价格
// 		"tickSize": "0.0001" // 订单最小价格间隔
// 	},
// 	{
// 		"filterType": "LOT_SIZE", // 数量限制
// 		"maxQty": "10000000", // 数量上限, 最大数量
// 		"minQty": "1", // 数量下限, 最小数量
// 		"stepSize": "1" // 订单最小数量间隔
// 	},
// 	{
// 		"filterType": "MARKET_LOT_SIZE", // 市价订单数量限制
// 		"maxQty": "590119", // 数量上限, 最大数量
// 		"minQty": "1", // 数量下限, 最小数量
// 		"stepSize": "1" // 允许的步进值
// 	},
// 	{
// 		"filterType": "MAX_NUM_ORDERS", // 最多订单数限制
// 		"limit": 200
// 	},
// 	{
// 		"filterType": "MAX_NUM_ALGO_ORDERS", // 最多条件订单数限制
// 		"limit": 10
// 	},
// 	{
// 		"filterType": "MIN_NOTIONAL",  // 最小名义价值
// 		"notional": "5.0",
// 	},
// 	{
// 		"filterType": "PERCENT_PRICE", // 价格比限制
// 		"multiplierUp": "1.1500", // 价格上限百分比
// 		"multiplierDown": "0.8500", // 价格下限百分比
// 		"multiplierDecimal": "4"
// 	},
// ]

// PriceFilter 价格限制
type PriceFilter struct {
	MaxPrice float64 `json:"maxPrice,string"` // 价格上限
	MinPrice float64 `json:"minPrice,string"` // 价格下限
	TickSize float64 `json:"tickSize,string"` // 订单最小价格间隔
}

// LotSizeFilter 数量限制
type LotSizeFilter struct {
	MaxQty   float64 `json:"maxQty,string"`   // 数量上限
	MinQty   float64 `json:"minQty,string"`   // 数量下限
	StepSize float64 `json:"stepSize,string"` // 订单最小数量间隔
}

// MarketLotSizeFilter 市价订单数量限制
type MarketLotSizeFilter struct {
	MaxQty   float64 `json:"maxQty,string"`   // 数量上限
	MinQty   float64 `json:"minQty,string"`   // 数量下限
	StepSize float64 `json:"stepSize,string"` // 允许的步进值
}

// MaxNumOrdersFilter 最多订单数限制
type MaxNumOrdersFilter struct {
	Limit int `json:"limit"` // 最多订单数
}

// MaxNumAlgoOrdersFilter 最多条件订单数限制
type MaxNumAlgoOrdersFilter struct {
	Limit int `json:"limit"` // 最多条件订单数
}

// MinNotionalFilter 最小名义价值
type MinNotionalFilter struct {
	Notional float64 `json:"notional,string"` // 最小名义价值
}

// PercentPriceFilter 价格比限制
type PercentPriceFilter struct {
	MultiplierUp      float64 `json:"multiplierUp,string"`      // 价格上限百分比
	MultiplierDown    float64 `json:"multiplierDown,string"`    // 价格下限百分比
	MultiplierDecimal int     `json:"multiplierDecimal,string"` // 价格百分比小数位
}

// Filters 过滤器集合
type Filters struct {
	PriceFilter         *PriceFilter            `json:"-"` // 价格限制
	LotSizeFilter       *LotSizeFilter          `json:"-"` // 数量限制
	MarketLotSizeFilter *MarketLotSizeFilter    `json:"-"` // 市价订单数量限制
	MaxNumOrders        *MaxNumOrdersFilter     `json:"-"` // 最多订单数限制
	MaxNumAlgoOrders    *MaxNumAlgoOrdersFilter `json:"-"` // 最多条件订单数限制
	MinNotional         *MinNotionalFilter      `json:"-"` // 最小名义价值
	PercentPrice        *PercentPriceFilter     `json:"-"` // 价格比限制
}

// Symbol market symbol
type Symbol struct {
	Symbol                string   `json:"symbol"`                // 合约交易对
	Pair                  string   `json:"pair"`                  // 交易对
	DeliveryDate          int64    `json:"deliveryDate"`          // 交割日期
	OnboardDate           int64    `json:"onboardDate"`           // 上线日期
	Status                string   `json:"status"`                // 状态
	MaintMarginPercent    float64  `json:"maintMarginPercent"`    // 维持保证金率
	RequiredMarginPercent float64  `json:"requiredMarginPercent"` // 所需保证金率
	PricePrecision        int      `json:"pricePrecision"`        // 价格精度
	QuantityPrecision     int      `json:"quantityPrecision"`     // 数量精度
	BaseAssetPrecision    int      `json:"baseAssetPrecision"`    // 基础资产精度
	QuotePrecision        int      `json:"quotePrecision"`        // 报价精度
	UnderlyingType        string   `json:"underlyingType"`        // 底层类型
	UnderlyingSubType     []string `json:"underlyingSubType"`     // 底层子类型
	SettlePlan            int64    `json:"settlePlan"`            // 结算计划
	TriggerProtect        float64  `json:"triggerProtect"`        // 触发保护
	FiltersRaw            []Filter `json:"filters"`               // 原始过滤器数据
	Filters               Filters  `json:"-"`                     // 规范化的过滤器
	QuoteAsset            string   `json:"quoteAsset"`            // 报价资产
	MarginAsset           string   `json:"marginAsset"`           // 保证金资产
	BaseAsset             string   `json:"baseAsset"`             // 基础资产
	LiquidationFee        float64  `json:"liquidationFee"`        // 强平费率
	MarketTakeBound       float64  `json:"marketTakeBound"`       // 市价吃单(相对于标记价格)允许可造成的最大价格偏离比例
	ContractType          string   `json:"contractType"`          // 合约类型
	OrderType             []string `json:"orderType"`             // 订单类型
	TimeInForce           []string `json:"timeInForce"`           // 有效方式
}

// Filter 通用过滤器，用于解析 JSON
type Filter map[string]interface{}

// ParseFilters 解析过滤器数据，将原始 filters 转换为结构化数据
func (s *Symbol) ParseFilters() error {
	for _, filter := range s.FiltersRaw {
		filterType, ok := filter["filterType"].(string)
		if !ok {
			continue
		}

		switch filterType {
		case "PRICE_FILTER":
			s.Filters.PriceFilter = &PriceFilter{
				MaxPrice: utils.ToFloat64(filter["maxPrice"]),
				MinPrice: utils.ToFloat64(filter["minPrice"]),
				TickSize: utils.ToFloat64(filter["tickSize"]),
			}
		case "LOT_SIZE":
			s.Filters.LotSizeFilter = &LotSizeFilter{
				MaxQty:   utils.ToFloat64(filter["maxQty"]),
				MinQty:   utils.ToFloat64(filter["minQty"]),
				StepSize: utils.ToFloat64(filter["stepSize"]),
			}
		case "MARKET_LOT_SIZE":
			s.Filters.MarketLotSizeFilter = &MarketLotSizeFilter{
				MaxQty:   utils.ToFloat64(filter["maxQty"]),
				MinQty:   utils.ToFloat64(filter["minQty"]),
				StepSize: utils.ToFloat64(filter["stepSize"]),
			}
		case "MAX_NUM_ORDERS":
			s.Filters.MaxNumOrders = &MaxNumOrdersFilter{
				Limit: utils.ToInt(filter["limit"]),
			}
		case "MAX_NUM_ALGO_ORDERS":
			s.Filters.MaxNumAlgoOrders = &MaxNumAlgoOrdersFilter{
				Limit: utils.ToInt(filter["limit"]),
			}
		case "MIN_NOTIONAL":
			s.Filters.MinNotional = &MinNotionalFilter{
				Notional: utils.ToFloat64(filter["notional"]),
			}
		case "PERCENT_PRICE":
			s.Filters.PercentPrice = &PercentPriceFilter{
				MultiplierUp:      utils.ToFloat64(filter["multiplierUp"]),
				MultiplierDown:    utils.ToFloat64(filter["multiplierDown"]),
				MultiplierDecimal: utils.ToInt(filter["multiplierDecimal"]),
			}
		}
	}
	s.FiltersRaw = nil
	return nil
}

// InitContracts 初始化合约交易对基础信息
func (b *Binance) InitExchangeInfo(ctx context.Context) error {
	var httpRequest = &HttpRequest{
		binance:     b,                            // Binance客户端实例
		baseUrl:     b.UrlRest,                    // 基础URL
		apiUrl:      "/fapi/v1/exchangeInfo",      // 请求URL
		sign:        false,                        // 是否签名
		isTimestamp: false,                        // 是否时间戳
		params:      make(map[string]interface{}), // 请求参数
	}
	var res = new(struct {
		Symbols []Symbol `json:"symbols"`
	})

	err := httpRequest.GetJSON(ctx, res)
	if err != nil {
		return err
	}

	for _, symbol := range res.Symbols {
		// 解析过滤器
		if err := symbol.ParseFilters(); err != nil {
			continue // 忽略解析错误的交易对
		}
		// 只保留可以正常交易的合约
		if symbol.Status == "TRADING" {
			b.Exc[symbol.Symbol] = &symbol
		}
	}
	return nil
}
