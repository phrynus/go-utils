package gate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WsUserTradesEvent Gate 用户私有成交频道事件
type WsUserTradesEvent struct {
	Time    int64                `json:"time"`    // 响应时间(秒)
	TimeMs  int64                `json:"time_ms"` // 响应时间(毫秒)
	Channel string               `json:"channel"` // 频道名, futures.usertrades
	Event   string               `json:"event"`   // 事件, subscribe / unsubscribe / update
	Error   *wsError             `json:"error"`   // 错误信息
	Result  []WsUserTradesResult `json:"result"`  // 成交数据数组
}

// WsUserTradesResult 用户私有成交数据
type WsUserTradesResult struct {
	ID           string  `json:"id"`             // 交易 ID
	CreateTime   int64   `json:"create_time"`    // 创建时间
	CreateTimeMs int64   `json:"create_time_ms"` // 创建时间（以毫秒为单位）
	Contract     string  `json:"contract"`       // 合约名称
	OrderID      string  `json:"order_id"`       // 订单 ID
	Price        string  `json:"price"`          // 交易价格
	Size         string  `json:"size"`           // 交易数量
	Role         string  `json:"role"`           // 用户角色 (maker/taker)
	Text         string  `json:"text"`           // 订单自定义信息
	Fee          float64 `json:"fee"`            // 手续费
	PointFee     float64 `json:"point_fee"`      // 点卡手续费
}

// WsUserTrades 订阅 Gate 用户私有成交频道
// userID: 用户 ID
// contract: 合约名称，如果为空字符串则订阅所有市场（使用 "!all"）
// tradesHandler: 成交更新回调
// errHandler: 错误回调
func (g *Gate) WsUserTrades(
	ctx context.Context,
	userID string,
	contract string,
	tradesHandler func([]WsUserTradesResult),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(userID) == "" {
		return nil, nil, fmt.Errorf("userID 不能为空")
	}

	// 构建 payload
	var payload []interface{}
	if strings.TrimSpace(contract) == "" {
		// 订阅所有市场
		payload = []interface{}{userID, "!all"}
	} else {
		payload = []interface{}{userID, contract}
	}

	onConnected := func(conn *websocket.Conn) error {
		now := time.Now().Unix()
		req := &wsRequest{
			Time:    now,
			Channel: "futures.usertrades",
			Event:   "subscribe",
			Payload: payload,
			Auth:    g.wsSign("futures.usertrades", "subscribe", now),
		}
		if err := conn.WriteJSON(req); err != nil {
			return fmt.Errorf("订阅用户成交频道失败: %w", err)
		}
		return nil
	}

	doneC, stopC, err = g.wsServeJSON(
		ctx,
		onConnected,
		func() interface{} { return &WsUserTradesEvent{} },
		func(evt interface{}, _ []byte) {
			event := evt.(*WsUserTradesEvent)

			// 处理订阅响应
			if event.Event == "subscribe" {
				if event.Error != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("订阅用户成交频道失败: code=%d, message=%s",
							event.Error.Code, event.Error.Message))
					}
					return
				}
				// 订阅成功，继续等待更新
				return
			}

			// 处理成交更新
			if event.Event == "update" && len(event.Result) > 0 {
				if tradesHandler != nil {
					tradesHandler(event.Result)
				}
			}
		},
		errHandler,
	)

	return doneC, stopC, err
}
