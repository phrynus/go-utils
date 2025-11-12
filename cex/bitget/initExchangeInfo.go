package bitget

import (
	"context"
)

type ExchangeInfo map[string]*Symbol

type Symbol struct {
	Symbol               string   `json:"symbol"`               // 产品id
	SymbolName           string   `json:"symbolName"`           // 产品名称，可能为空
	SymbolDisplayName    string   `json:"symbolDisplayName"`    // 产品展示名称，可能为空
	BaseCoin             string   `json:"baseCoin"`             // 左币
	QuoteCoin            string   `json:"quoteCoin"`            // 右币
	BaseCoinDisplayName  string   `json:"baseCoinDisplayName"`  // 左币展示名称
	QuoteCoinDisplayName string   `json:"quoteCoinDisplayName"` // 右币展示名称
	BuyLimitPriceRatio   float64  `json:"buyLimitPriceRatio"`   // 买价限价比例
	SellLimitPriceRatio  float64  `json:"sellLimitPriceRatio"`  // 卖价限价比例
	FeeRateUpRatio       float64  `json:"feeRateUpRatio"`       // 手续费上浮比例
	MakerFeeRate         float64  `json:"makerFeeRate"`         // market手续费率
	TakerFeeRate         float64  `json:"takerFeeRate"`         // taker手续费率
	OpenCostUpRatio      float64  `json:"openCostUpRatio"`      // 开仓成本上浮比例
	SupportMarginCoins   []string `json:"supportMarginCoins"`   // 支持保证金币种
	MinTradeNum          float64  `json:"minTradeNum"`          // 最小开单数量(左币)
	PriceEndStep         float64  `json:"priceEndStep"`         // 价格步长
	VolumePlace          int      `json:"volumePlace"`          // 数量小数位
	PricePlace           int      `json:"pricePlace"`           // 价格小数位
	SizeMultiplier       float64  `json:"sizeMultiplier"`       // 数量乘数 下单数量要大于 minTradeNum 并且满足 sizeMultiplier 的倍数
	SymbolType           string   `json:"symbolType"`           // 合约类型 perpetual 永续 delivery交割
	SymbolStatus         string   `json:"symbolStatus"`         // Symbol Status
	OffTime              int64    `json:"offTime"`              // 下架时间, '-1' 表示正常
	LimitOpenTime        int64    `json:"limitOpenTime"`        // 限制开仓时间, '-1' 表示正常; 其它值表示symbol正在/计划维护，指定时间后禁止交易
}

func (b *Bitget) InitExchangeInfo(ctx context.Context) error {
	var httpRequest = &HttpRequest{
		bitget:      b,                              // Bitget客户端实例
		baseUrl:     b.UrlRest,                      // 基础URL
		apiUrl:      "/api/mix/v1/market/contracts", // 请求URL
		sign:        false,                          // 是否签名
		isTimestamp: false,                          // 是否时间戳
		params:      make(map[string]interface{}),   // 请求参数
	}
	httpRequest.params["productType"] = "umcbl"

	var res = new(struct {
		Data []Symbol `json:"data"`
	})

	err := httpRequest.GetJSON(ctx, res)
	if err != nil {
		return err
	}
	for _, symbol := range res.Data {
		if symbol.SymbolStatus == "normal" && symbol.OffTime == -1 && symbol.LimitOpenTime == -1 {
			b.Exc[symbol.SymbolName] = &symbol
		}
	}

	return nil
}
