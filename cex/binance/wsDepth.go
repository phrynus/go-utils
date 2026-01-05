package binance

import (
	"context"
	"fmt"
	"strings"
)

// WsDepthEvent 深度行情事件
type WsDepthEvent struct {
	Event             string      `json:"e"`  // 事件类型
	EventTime         int64       `json:"E"`  // 事件时间
	TransactionTime   int64       `json:"T"`  // 交易时间
	Symbol            string      `json:"s"`  // 交易对
	Pair              string      `json:"ps"` // 标的交易对
	FirstUpdateID     int64       `json:"U"`  // 首个ID
	FinalUpdateID     int64       `json:"u"`  // 末尾ID
	PrevFinalUpdateID int64       `json:"pu"` // 前一末尾ID
	Bids              [][]float64 `json:"b"`  // 买单
	Asks              [][]float64 `json:"a"`  // 卖单
}

// WsDepth 订阅深度流
func (b *Binance) WsDepth(
	ctx context.Context,
	symbol string,
	levels string,
	wsDepthHandler func(*WsDepthEvent),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, nil, fmt.Errorf("symbol 不能为空")
	}

	stream := strings.ToLower(symbol) + "@depth" + levels + "@100ms"

	return b.wsServeJSON(
		ctx,
		stream,
		func() interface{} { return &WsDepthEvent{} },
		func(evt interface{}, _ []byte) {
			if wsDepthHandler == nil {
				return
			}
			wsDepthHandler(evt.(*WsDepthEvent))
		},
		errHandler,
	)
}
