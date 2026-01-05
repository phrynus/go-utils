package gate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WsDepthEvent Gate 深度全量更新频道事件
type WsDepthEvent struct {
	Time    int64          `json:"time"`    // 响应时间(秒)
	TimeMs  int64          `json:"time_ms"` // 响应时间(毫秒)
	Channel string         `json:"channel"` // 频道名, futures.order_book
	Event   string         `json:"event"`   // 事件, subscribe / all
	Error   *wsError       `json:"error"`   // 错误信息
	Result  *WsDepthResult `json:"result"`  // 深度数据
}

// WsDepthOrderBookEntry 深度订单簿条目
type WsDepthOrderBookEntry struct {
	Price  string `json:"p"` // 档位价格
	Amount string `json:"s"` // 档位数量
}

// WsDepthResult 深度数据结果
type WsDepthResult struct {
	// 订阅响应
	Status string `json:"status,omitempty"` // success / fail

	// 全量深度更新数据
	Timestamp int64                   `json:"t,omitempty"`        // 深度生成时间戳(毫秒)
	Contract  string                  `json:"contract,omitempty"` // 合约名称
	ID        int64                   `json:"id,omitempty"`       // 深度 ID
	Asks      []WsDepthOrderBookEntry `json:"asks,omitempty"`     // 卖方深度档位列表
	Bids      []WsDepthOrderBookEntry `json:"bids,omitempty"`     // 买方深度档位列表
	Level     string                  `json:"l,omitempty"`        // 深度层级
}

// WsDepth 订阅 Gate 深度全量更新频道
// symbol: 合约名称, 如 BTC_USD
// level: 深度层级, 100, 50, 20, 10, 5, 1
// depthHandler: 深度更新回调
// errHandler: 错误回调
func (g *Gate) WsDepth(
	ctx context.Context,
	symbol string,
	level int,
	depthHandler func(*WsDepthResult),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, nil, fmt.Errorf("symbol 不能为空")
	}

	validLevels := []int{100, 50, 20, 10, 5, 1}
	isValidLevel := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		return nil, nil, fmt.Errorf("level 必须是 100, 50, 20, 10, 5, 1 中的一个")
	}

	onConnected := func(conn *websocket.Conn) error {
		now := time.Now().Unix()
		req := &wsRequest{
			Time:    now,
			Channel: "futures.order_book",
			Event:   "subscribe",
			Payload: []interface{}{symbol, fmt.Sprintf("%d", level), "0"},
		}
		if err := conn.WriteJSON(req); err != nil {
			return fmt.Errorf("订阅深度频道失败: %w", err)
		}
		return nil
	}
	doneC, stopC, err = g.wsServeJSON(
		ctx,
		onConnected,
		func() interface{} { return &WsDepthEvent{} },
		func(evt interface{}, _ []byte) {
			event := evt.(*WsDepthEvent)

			// 处理订阅响应
			if event.Event == "subscribe" {
				if event.Error != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("订阅深度频道失败: code=%d, message=%s",
							event.Error.Code, event.Error.Message))
					}
					return
				}
				if event.Result != nil && event.Result.Status == "fail" {
					if errHandler != nil {
						errHandler(fmt.Errorf("订阅深度频道失败: status=fail"))
					}
					return
				}
				// 订阅成功，继续等待更新
				return
			}

			// 处理全量深度更新
			if event.Event == "all" && event.Result != nil {
				if depthHandler != nil {
					depthHandler(event.Result)
				}
			}
		},
		errHandler,
	)

	return doneC, stopC, err
}
