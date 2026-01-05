package binance

import (
	"context"
	"fmt"
	"strings"
)

// WsMiniTickerEvent 24h精简ticker
type WsMiniTickerEvent struct {
	Event       string `json:"e"` // 事件类型
	EventTime   int64  `json:"E"` // 事件时间
	Symbol      string `json:"s"` // 交易对
	LastPrice   string `json:"c"` // 最新价
	OpenPrice   string `json:"o"` // 开盘价
	HighPrice   string `json:"h"` // 高价
	LowPrice    string `json:"l"` // 低价
	Volume      string `json:"v"` // 成交量
	QuoteVolume string `json:"q"` // 成交额
}

// WsMiniTicker 订阅24h精简ticker
func (b *Binance) WsMiniTicker(
	ctx context.Context,
	symbol string,
	wsMiniTickerHandler func(*WsMiniTickerEvent),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, nil, fmt.Errorf("symbol 不能为空")
	}

	stream := strings.ToLower(symbol) + "@miniTicker"

	return b.wsServeJSON(
		ctx,
		stream,
		func() interface{} { return &WsMiniTickerEvent{} },
		func(evt interface{}, _ []byte) {
			if wsMiniTickerHandler == nil {
				return
			}
			wsMiniTickerHandler(evt.(*WsMiniTickerEvent))
		},
		errHandler,
	)
}
