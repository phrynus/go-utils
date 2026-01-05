// 获取合约交易对基础信息
package binance

import (
	"context"
	"sync"
	"time"

	"github.com/phrynus/go-utils"
)

type ExchangeInfo map[string]*Symbol

var ExcMu sync.RWMutex

// PriceFilter 价格过滤器
type PriceFilter struct {
	MaxPrice float64 `json:"maxPrice,string"` // 上限
	MinPrice float64 `json:"minPrice,string"` // 下限
	TickSize float64 `json:"tickSize,string"` // 最小间隔
}

// LotSizeFilter 数量过滤器
type LotSizeFilter struct {
	MaxQty   float64 `json:"maxQty,string"`   // 上限
	MinQty   float64 `json:"minQty,string"`   // 下限
	StepSize float64 `json:"stepSize,string"` // 最小间隔
}

// MarketLotSizeFilter 市价单数量过滤器
type MarketLotSizeFilter struct {
	MaxQty   float64 `json:"maxQty,string"`   // 上限
	MinQty   float64 `json:"minQty,string"`   // 下限
	StepSize float64 `json:"stepSize,string"` // 步进值
}

// MaxNumOrdersFilter 单合约最多订单数
type MaxNumOrdersFilter struct {
	Limit int `json:"limit"` // 限制
}

// MaxNumAlgoOrdersFilter 单合约最多条件订单数
type MaxNumAlgoOrdersFilter struct {
	Limit int `json:"limit"` // 限制
}

// MinNotionalFilter 最小名义价值
type MinNotionalFilter struct {
	Notional float64 `json:"notional,string"` // 金额
}

// PercentPriceFilter 价格比过滤器
type PercentPriceFilter struct {
	MultiplierUp      float64 `json:"multiplierUp,string"`      // 上限倍数
	MultiplierDown    float64 `json:"multiplierDown,string"`    // 下限倍数
	MultiplierDecimal int     `json:"multiplierDecimal,string"` // 小数位
}

// Filters 过滤器集合
type Filters struct {
	PriceFilter         *PriceFilter            `json:"-"` // 价格
	LotSizeFilter       *LotSizeFilter          `json:"-"` // 数量
	MarketLotSizeFilter *MarketLotSizeFilter    `json:"-"` // 市价数量
	MaxNumOrders        *MaxNumOrdersFilter     `json:"-"` // 最多订单
	MaxNumAlgoOrders    *MaxNumAlgoOrdersFilter `json:"-"` // 最多条件单
	MinNotional         *MinNotionalFilter      `json:"-"` // 最小金额
	PercentPrice        *PercentPriceFilter     `json:"-"` // 价格比
}

// Symbol 合约交易对
type Symbol struct {
	Symbol                string   `json:"symbol"`                // 交易对
	Pair                  string   `json:"pair"`                  // 标的交易对
	DeliveryDate          int64    `json:"deliveryDate"`          // 交割日期
	OnboardDate           int64    `json:"onboardDate"`           // 上线日期
	Status                string   `json:"status"`                // 状态
	MaintMarginPercent    float64  `json:"maintMarginPercent"`    // 维持保证金率
	RequiredMarginPercent float64  `json:"requiredMarginPercent"` // 初始保证金率
	PricePrecision        int      `json:"pricePrecision"`        // 价格精度
	QuantityPrecision     int      `json:"quantityPrecision"`     // 数量精度
	BaseAssetPrecision    int      `json:"baseAssetPrecision"`    // 基础资产精度
	QuotePrecision        int      `json:"quotePrecision"`        // 报价精度
	UnderlyingType        string   `json:"underlyingType"`        // 底层类型
	UnderlyingSubType     []string `json:"underlyingSubType"`     // 底层子类型
	SettlePlan            int64    `json:"settlePlan"`            // 结算计划
	TriggerProtect        float64  `json:"triggerProtect"`        // 触发保护
	FiltersRaw            []Filter `json:"filters"`               // 原始过滤器
	Filters               Filters  `json:"-"`                     // 过滤器
	QuoteAsset            string   `json:"quoteAsset"`            // 计价资产
	MarginAsset           string   `json:"marginAsset"`           // 保证金资产
	BaseAsset             string   `json:"baseAsset"`             // 基础资产
	LiquidationFee        float64  `json:"liquidationFee"`        // 强平费率
	MarketTakeBound       float64  `json:"marketTakeBound"`       // 市价偏离比例
	ContractType          string   `json:"contractType"`          // 合约类型
	OrderType             []string `json:"orderType"`             // 订单类型
	TimeInForce           []string `json:"timeInForce"`           // 有效方式
}

// Filter 通用过滤器原始字段
type Filter map[string]interface{}

// ParseFilters 将原始 filters 解析到 Filters
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

// InitExchangeInfo 初始化交易对信息
func (b *Binance) InitExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	var httpRequest = &HttpRequest{
		binance:     b,
		baseUrl:     b.UrlRest,
		apiUrl:      "/fapi/v1/exchangeInfo",
		sign:        false,
		isTimestamp: false,
		params:      make(map[string]interface{}),
	}
	var res = new(struct {
		Symbols []Symbol `json:"symbols"`
	})

	err := httpRequest.GetJSON(ctx, res)
	if err != nil {
		return &ExchangeInfo{}, err
	}

	// 初次加载并安全写入 b.Exc
	ExcMu.Lock()
	for _, symbol := range res.Symbols {
		if err := symbol.ParseFilters(); err != nil {
			continue
		}
		if symbol.Status == "TRADING" {
			s := symbol
			b.Exc[s.Symbol] = &s
		}
	}
	ExcMu.Unlock()

	// 启动后台每小时自动更新（由传入 ctx 控制取消）
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var httpRequest2 = &HttpRequest{
					binance:     b,
					baseUrl:     b.UrlRest,
					apiUrl:      "/fapi/v1/exchangeInfo",
					sign:        false,
					isTimestamp: false,
					params:      make(map[string]interface{}),
				}
				var res2 = new(struct {
					Symbols []Symbol `json:"symbols"`
				})
				if err := httpRequest2.GetJSON(ctx, res2); err != nil {
					continue
				}
				ExcMu.Lock()
				for _, symbol := range res2.Symbols {
					if err := symbol.ParseFilters(); err != nil {
						continue
					}
					if symbol.Status == "TRADING" {
						s := symbol
						b.Exc[s.Symbol] = &s
					}
				}
				ExcMu.Unlock()
			}
		}
	}()
	return &b.Exc, nil
}
