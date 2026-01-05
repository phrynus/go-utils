package gate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WsBookTickerEvent Gate 最佳买卖价频道事件
type WsBookTickerEvent struct {
	Time    int64               `json:"time"`    // 响应时间(秒)
	TimeMs  int64               `json:"time_ms"` // 响应时间(毫秒)
	Channel string              `json:"channel"` // 频道名, futures.book_ticker
	Event   string              `json:"event"`   // 事件, subscribe / update
	Error   *wsError            `json:"error"`   // 错误信息
	Result  *WsBookTickerResult `json:"result"`  // 最佳买卖价数据
}

// WsBookTickerResult 最佳买卖价数据结果
type WsBookTickerResult struct {
	// 订阅响应
	Status string `json:"status,omitempty"` // success / fail

	// 最佳买卖价更新数据
	Timestamp   int64  `json:"t,omitempty"` // 最佳买卖价行情生成的时间戳(毫秒)
	UpdateID    int64  `json:"u,omitempty"` // 深度的 ID
	Symbol      string `json:"s,omitempty"` // 合约名称
	BestBid     string `json:"b,omitempty"` // 最佳买方的价格，如果没有买方，则为空串
	BestBidQty  string `json:"B,omitempty"` // 最佳买方的数量，如果没有买方，则为 0
	BestAsk     string `json:"a,omitempty"` // 最佳卖方的价格，如果没有卖方，则为空串
	BestAskQty  string `json:"A,omitempty"` // 最佳卖方的数量，如果没有卖方，则为 0
}

// WsBookTicker 订阅 Gate 最佳买卖价频道
// symbol: 合约名称, 如 BTC_USDT
// bookTickerHandler: 最佳买卖价更新回调
// errHandler: 错误回调
func (g *Gate) WsBookTicker(
	ctx context.Context,
	symbol string,
	bookTickerHandler func(*WsBookTickerResult),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, nil, fmt.Errorf("symbol 不能为空")
	}

	onConnected := func(conn *websocket.Conn) error {
		now := time.Now().Unix()
		req := &wsRequest{
			Time:    now,
			Channel: "futures.book_ticker",
			Event:   "subscribe",
			Payload: []interface{}{symbol},
		}
		if err := conn.WriteJSON(req); err != nil {
			return fmt.Errorf("订阅最佳买卖价频道失败: %w", err)
		}
		return nil
	}

	doneC, stopC, err = g.wsServeJSON(
		ctx,
		onConnected,
		func() interface{} { return &WsBookTickerEvent{} },
		func(evt interface{}, _ []byte) {
			event := evt.(*WsBookTickerEvent)

			// 处理订阅响应
			if event.Event == "subscribe" {
				if event.Error != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("订阅最佳买卖价频道失败: code=%d, message=%s",
							event.Error.Code, event.Error.Message))
					}
					return
				}
				if event.Result != nil && event.Result.Status == "fail" {
					if errHandler != nil {
						errHandler(fmt.Errorf("订阅最佳买卖价频道失败: status=fail"))
					}
					return
				}
				// 订阅成功，继续等待更新
				return
			}

			// 处理最佳买卖价更新
			if event.Event == "update" && event.Result != nil {
				if bookTickerHandler != nil {
					bookTickerHandler(event.Result)
				}
			}
		},
		errHandler,
	)

	return doneC, stopC, err
}
