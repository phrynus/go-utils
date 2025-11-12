package bybit

import (
	"context"
)

type ExchangeInfo map[string]*Symbol

type Symbol struct {
	Symbol             string         `json:"symbol"`             // 合约名称
	ContractType       string         `json:"contractType"`       // 合约类型
	Status             string         `json:"status"`             // 合约状态
	BaseCoin           string         `json:"baseCoin"`           // 交易币种
	QuoteCoin          string         `json:"quoteCoin"`          // 报价币种
	LaunchTime         int64          `json:"launchTime"`         // 发布时间 (ms)
	DeliveryTime       int64          `json:"deliveryTime"`       // 交割时间 (ms)
	DeliveryFeeRate    float64        `json:"deliveryFeeRate"`    // 交割费率. 仅对交割合約有效
	PriceScale         int            `json:"priceScale"`         // 价格精度
	UnifiedMarginTrade bool           `json:"unifiedMarginTrade"` // 是否支持统一保证金交易
	FundingInterval    int            `json:"fundingInterval"`    // 资金费率结算周期 (分钟)
	SettleCoin         string         `json:"settleCoin"`         // 结算币种
	CopyTrading        string         `json:"copyTrading"`        // 当前交易对是否支持带单交易
	UpperFundingRate   float64        `json:"upperFundingRate"`   // 资金费率上限
	LowerFundingRate   float64        `json:"lowerFundingRate"`   // 资金费率下限
	IsPreListing       bool           `json:"isPreListing"`       // 是否为盘前合约
	PreListingInfo     string         `json:"preListingInfo"`     // 如果isPreListing=false, preListingInfo=null
	SymbolType         string         `json:"symbolType"`         // 交易对所属区域
	RiskParameters     RiskParameters `json:"riskParameters"`     // 风险参数
	LotSizeFilter      LotSizeFilter  `json:"lotSizeFilter"`      // 订单数量属性
	PriceFilter        PriceFilter    `json:"priceFilter"`        // 价格属性
	LeverageFilter     LeverageFilter `json:"leverageFilter"`     // 杠杆属性
}

type RiskParameters struct {
	PriceLimitRatioX float64 `json:"priceLimitRatioX"` // 参数X
	PriceLimitRatioY float64 `json:"priceLimitRatioY"` // 参数Y
}

type LotSizeFilter struct {
	MaxOrderQty         float64 `json:"maxOrderQty"`         // 单笔限价或PostOnly单最大下单量
	MinOrderQty         float64 `json:"minOrderQty"`         // 单笔订单最小下单量
	QtyStep             float64 `json:"qtyStep"`             // 修改下单量的步长
	PostOnlyMaxOrderQty float64 `json:"postOnlyMaxOrderQty"` // 单笔订单最小下单量
	MaxMktOrderQty      float64 `json:"maxMktOrderQty"`      // 单笔市价单最大下单量
	MinNotionalValue    float64 `json:"minNotionalValue"`    // 订单最小名义价值
}

type PriceFilter struct {
	MinPrice float64 `json:"minPrice"` // 订单最小价格
	MaxPrice float64 `json:"maxPrice"` // 订单最大价格
	TickSize float64 `json:"tickSize"` // 修改价格的步长
}

type LeverageFilter struct {
	MinLeverage  int `json:"minLeverage"`  // 最小杠杆
	MaxLeverage  int `json:"maxLeverage"`  // 最大杠杆
	LeverageStep int `json:"leverageStep"` // 修改杠杆的步长
}

func (b *Bybit) InitExchangeInfo(ctx context.Context) error {
	var httpRequest = &HttpRequest{
		bybit:       b,                             // Bybit客户端实例
		baseUrl:     b.UrlRest,                     // 基础URL
		apiUrl:      "/v5/market/instruments-info", // 请求URL
		sign:        false,                         // 是否签名
		isTimestamp: false,                         // 是否时间戳
		params:      make(map[string]interface{}),  // 请求参数
	}
	httpRequest.params["category"] = "linear" // spot 现货,linear ,inverse,option

	var res = new(struct {
		Result struct {
			List []Symbol `json:"list"`
		} `json:"result"`
	})

	err := httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return err
	}

	for _, symbol := range res.Result.List {
		if symbol.Status == "Trading" {
			b.Exc[symbol.Symbol] = &symbol
		}
	}

	return nil
}
