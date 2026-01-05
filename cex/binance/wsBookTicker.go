package binance

import (
	"context"
	"fmt"
	"strings"
)

// WsBookTickerEvent 最优挂单
type WsBookTickerEvent struct {
	Event           string `json:"e"` // 事件类型
	UpdateID        int64  `json:"u"` // 更新ID
	EventTime       int64  `json:"E"` // 事件时间
	TransactionTime int64  `json:"T"` // 撮合时间
	Symbol          string `json:"s"` // 交易对
	BestBidPrice    string `json:"b"` // 买单价
	BestBidQty      string `json:"B"` // 买单量
	BestAskPrice    string `json:"a"` // 卖单价
	BestAskQty      string `json:"A"` // 卖单量
}

// WsBookTicker 订阅最优挂单
func (b *Binance) WsBookTicker(
	ctx context.Context,
	symbol string,
	wsBookTickerHandler func(*WsBookTickerEvent),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, nil, fmt.Errorf("symbol 不能为空")
	}

	stream := strings.ToLower(symbol) + "@bookTicker"

	return b.wsServeJSON(
		ctx,
		stream,
		func() interface{} { return &WsBookTickerEvent{} },
		func(evt interface{}, _ []byte) {
			if wsBookTickerHandler == nil {
				return
			}
			wsBookTickerHandler(evt.(*WsBookTickerEvent))
		},
		errHandler,
	)
}
