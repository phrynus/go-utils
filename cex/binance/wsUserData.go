package binance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/phrynus/go-utils"
)

// UserDataEventType 用户数据事件类型
type UserDataEventType string

// TimeInForceType 有效方式类型
type TimeInForceType string

// WsUserDataEvent 用户数据事件
type WsUserDataEvent struct {
	Event         UserDataEventType   `json:"e"` // 事件类型
	Time          int64               `json:"E"` // 事件时间
	AccountUpdate WsAccountUpdateList // 账户更新
	BalanceUpdate WsBalanceUpdate     // 余额更新
	OrderUpdate   WsOrderUpdate       // 订单更新
	OCOUpdate     WsOCOUpdate         // OCO更新
	TradeLite     WsTradeLite         // 成交事件
}

// WsAccountUpdateList 账户更新
type WsAccountUpdateList struct {
	AccountUpdateTime int64             `json:"u"` // 更新时间
	WsAccountUpdates  []WsAccountUpdate `json:"B"` // 资产列表
}

// WsAccountUpdate 资产余额
type WsAccountUpdate struct {
	Asset  string  `json:"a"`        // 资产
	Free   float64 `json:"f,string"` // 可用
	Locked float64 `json:"l,string"` // 冻结
}

// WsBalanceUpdate 余额变更
type WsBalanceUpdate struct {
	Asset           string  `json:"a"`        // 资产
	Change          float64 `json:"d,string"` // 变动量
	TransactionTime int64   `json:"T"`        // 变动时间
}

// WsOrderUpdate 订单更新
type WsOrderUpdate struct {
	Symbol        string          `json:"s"` // 交易对
	ClientOrderId string          `json:"c"` // 客户订单ID
	Side          string          `json:"S"` // 方向
	Type          string          `json:"o"` // 类型
	TimeInForce   TimeInForceType `json:"f"` // 有效方式

	Volume        float64 `json:"q,string"` // 订单量
	Price         float64 `json:"p,string"` // 订单价
	StopPrice     float64 `json:"P,string"` // 触发价
	IceBergVolume float64 `json:"F,string"` // 冰山量

	OrderListId       int64  `json:"g"` // 列表ID
	OrigCustomOrderId string `json:"C"` // 原始订单ID

	ExecutionType string `json:"x"` // 执行类型
	Status        string `json:"X"` // 状态
	RejectReason  string `json:"r"` // 拒绝原因

	Id           int64   `json:"i"`        // 订单ID
	LatestVolume float64 `json:"l,string"` // 最新成交量
	FilledVolume float64 `json:"z,string"` // 累计成交量
	LatestPrice  float64 `json:"L,string"` // 最新成交价

	FeeAsset string  `json:"N"`        // 手续费资产
	FeeCost  float64 `json:"n,string"` // 手续费

	TransactionTime int64 `json:"T"` // 成交时间
	TradeId         int64 `json:"t"` // 成交ID

	IgnoreI       int64 `json:"I"` // 保留
	IsInOrderBook bool  `json:"w"` // 在盘口中
	IsMaker       bool  `json:"m"` // 是否maker
	IgnoreM       bool  `json:"M"` // 保留

	CreateTime              int64   `json:"O"`        // 创建时间
	FilledQuoteVolume       float64 `json:"Z,string"` // 成交金额
	LatestQuoteVolume       float64 `json:"Y,string"` // 最新成交金额
	QuoteVolume             float64 `json:"Q,string"` // 订单金额
	SelfTradePreventionMode string  `json:"V"`        // 自成交保护

	TrailingDelta              int64   `json:"d"`         // 追踪幅度
	TrailingTime               int64   `json:"D"`         // 追踪时间
	StrategyId                 int64   `json:"j"`         // 策略ID
	StrategyType               int64   `json:"J"`         // 策略类型
	PreventedMatchId           int64   `json:"v"`         // STP阻止ID
	PreventedQuantity          float64 `json:"A,string"`  // STP阻止量
	LastPreventedQuantity      float64 `json:"B,string"`  // 最新STP阻止量
	TradeGroupId               int64   `json:"u"`         // 交易分组ID
	CounterOrderId             int64   `json:"U"`         // 对手订单ID
	CounterSymbol              string  `json:"Cs"`        // 对手交易对
	PreventedExecutionQuantity float64 `json:"pl,string"` // STP阻止执行量
	PreventedExecutionPrice    float64 `json:"pL,string"` // STP阻止执行价
	PreventedExecutionQuoteQty float64 `json:"pY,string"` // STP阻止执行金额
	WorkingTime                int64   `json:"W"`         // 工作时间
	MatchType                  string  `json:"b"`         // 撮合类型
	AllocationId               int64   `json:"a"`         // 分配ID
	WorkingFloor               string  `json:"k"`         // 交易场所
	UsedSor                    bool    `json:"uS"`        // 使用SOR
}

// WsOCOUpdate OCO更新
type WsOCOUpdate struct {
	Symbol          string         `json:"s"` // 交易对
	OrderListId     int64          `json:"g"` // 列表ID
	ContingencyType string         `json:"c"` // 触发类型
	ListStatusType  string         `json:"l"` // 列表状态
	ListOrderStatus string         `json:"L"` // 订单状态
	RejectReason    string         `json:"r"` // 拒绝原因
	ClientOrderId   string         `json:"C"` // 客户订单ID
	TransactionTime int64          `json:"T"` // 时间
	Orders          WsOCOOrderList `json:"O"` // 子订单列表
}

// WsOCOOrderList OCO子订单
type WsOCOOrderList struct {
	WsOCOOrders []WsOCOOrder `json:"O"` // 列表
}

// WsOCOOrder OCO子订单
type WsOCOOrder struct {
	Symbol        string `json:"s"` // 交易对
	OrderId       int64  `json:"i"` // 订单ID
	ClientOrderId string `json:"c"` // 客户订单ID
}

// WsTradeLite 精简成交
type WsTradeLite struct {
	EventType     UserDataEventType `json:"e"`        // 事件类型
	EventTime     int64             `json:"E"`        // 事件时间
	TradeTime     int64             `json:"T"`        // 交易时间
	Symbol        string            `json:"s"`        // 交易对
	OrigQty       float64           `json:"q,string"` // 订单量
	OrigPrice     float64           `json:"p,string"` // 订单价
	IsMaker       bool              `json:"m"`        // 是否maker
	ClientOrderId string            `json:"c"`        // 客户订单ID
	Side          string            `json:"S"`        // 方向
	LastPrice     float64           `json:"L,string"` // 最新成交价
	LastQty       float64           `json:"l,string"` // 最新成交量
	TradeId       int64             `json:"t"`        // 成交ID
	OrderId       int64             `json:"i"`        // 订单ID
}

// WsUserData 用户数据流 WebSocket
func (b *Binance) WsUserData(
	ctx context.Context,
	listenKey string,
	userDataHandler func(*WsUserDataEvent),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	if strings.TrimSpace(listenKey) == "" {
		return nil, nil, fmt.Errorf("listenKey 不能为空")
	}

	// 使用通用 wsServeJSON，先解析通用字段 e/E，再根据事件类型做一次精细解析
	return b.wsServeJSON(
		ctx,
		listenKey,
		func() interface{} {
			return &struct {
				Event UserDataEventType `json:"e"`
				Time  int64             `json:"E"`
			}{}
		},
		func(evt interface{}, raw []byte) {
			if userDataHandler == nil {
				return
			}

			base := evt.(*struct {
				Event UserDataEventType `json:"e"`
				Time  int64             `json:"E"`
			})

			event := &WsUserDataEvent{
				Event: base.Event,
				Time:  base.Time,
			}

			// 根据事件类型解析对应 payload
			switch base.Event {
			case "ACCOUNT_UPDATE":
				// 顶层的 "a" 字段中包含账户更新列表
				var payload struct {
					Account WsAccountUpdateList `json:"a"`
				}
				if err := utils.SmartUnmarshal(raw, &payload); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("解析 ACCOUNT_UPDATE 事件失败: %w", err))
					}
					return
				}
				event.AccountUpdate = payload.Account

			case "BALANCE_UPDATE":
				// 余额更新事件字段直接在顶层
				var payload struct {
					WsBalanceUpdate
				}
				if err := utils.SmartUnmarshal(raw, &payload); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("解析 BALANCE_UPDATE 事件失败: %w", err))
					}
					return
				}
				event.BalanceUpdate = payload.WsBalanceUpdate

			case "ORDER_TRADE_UPDATE":
				// 订单更新事件在 "o" 字段中
				var payload struct {
					Order WsOrderUpdate `json:"o"`
				}
				if err := utils.SmartUnmarshal(raw, &payload); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("解析 ORDER_TRADE_UPDATE 事件失败: %w", err))
					}
					return
				}
				event.OrderUpdate = payload.Order

			case "LIST_STATUS":
				// OCO 列表状态更新事件结构即为 WsOCOUpdate
				var payload WsOCOUpdate
				if err := utils.SmartUnmarshal(raw, &payload); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("解析 LIST_STATUS 事件失败: %w", err))
					}
					return
				}
				event.OCOUpdate = payload

			case "TRADE_LITE":
				// 精简成交事件, 结构紧凑, 直接映射到 WsTradeLite
				var payload WsTradeLite
				if err := utils.SmartUnmarshal(raw, &payload); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("解析 TRADE_LITE 事件失败: %w", err))
					}
					return
				}
				event.TradeLite = payload
			}

			userDataHandler(event)
		},
		errHandler,
	)
}

// StartUserDataStream 启动并维护用户数据流
func (b *Binance) WsUserDataStream(
	ctx context.Context,
	userDataHandler func(*WsUserDataEvent),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	// 先创建 listenKey
	listenKey, err := b.CreateListenKey(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 用上层 ctx 派生子 ctx, 方便统一取消
	wsCtx, cancelWs := context.WithCancel(ctx)

	wsDoneC, wsStopC, err := b.WsUserData(wsCtx, listenKey, userDataHandler, errHandler)
	if err != nil {
		cancelWs()
		return nil, nil, fmt.Errorf("启动用户数据 WebSocket 失败: %w", err)
	}

	doneC = make(chan struct{})
	stopC = make(chan struct{})

	// 保活 goroutine
	go func() {
		// 官方要求 60 分钟内至少调用一次; 这里取 30 分钟安全值
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-wsCtx.Done():
				return
			case <-stopC:
				return
			case <-ticker.C:
				// 使用 Background 防止因上层 ctx 过期导致无法保活
				if err := b.KeepAliveListenKey(context.Background(), listenKey); err != nil {
					if errHandler != nil {
						errHandler(fmt.Errorf("listenKey 保活失败: %w", err))
					}
				}
			}
		}
	}()

	// 统一收尾 goroutine
	go func() {
		defer close(doneC)
		defer cancelWs()

		select {
		case <-stopC:
			// 主动停止: 先关 WS, 再删除 listenKey
			wsStopC <- struct{}{}
			<-wsDoneC
			_ = b.CloseListenKey(context.Background(), listenKey)
		case <-wsDoneC:
			// WS 自然结束: 也尝试删除 listenKey
			_ = b.CloseListenKey(context.Background(), listenKey)
		}
	}()

	return doneC, stopC, nil
}
