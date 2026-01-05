package gate

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phrynus/go-utils"
	"golang.org/x/net/proxy"
)

const (
	// wsReadLimit WebSocket 读取限制（字节）
	wsReadLimit = 655350
	// wsTimeout Gate.io 服务器在30秒内没有收到消息会断开连接，这里设置为90秒作为安全缓冲
	wsTimeout = 90 * time.Second
	// wsTimeoutCheckInterval 超时检查间隔（应该比 wsTimeout 短得多）
	wsTimeoutCheckInterval = 30 * time.Second
	// wsPingInterval 应用层 ping 发送间隔
	wsPingInterval = 30 * time.Second
	// wsReconnectInitialDelay 初始重连延迟
	wsReconnectInitialDelay = 1 * time.Second
	// wsReconnectMaxDelay 最大重连延迟
	wsReconnectMaxDelay = 60 * time.Second
	// wsReconnectMaxAttempts 最大重连次数（0 表示无限重连）
	wsReconnectMaxAttempts = 0
)

// wsError WebSocket 错误对象
type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// wsRequest WebSocket 请求结构
type wsRequest struct {
	ID      int64         `json:"id,omitempty"`      // 可选请求ID
	Time    int64         `json:"time"`              // 时间戳(秒)
	Channel string        `json:"channel"`           // 频道名, 如 futures.order_book / futures.orders
	Event   string        `json:"event"`             // 事件, subscribe / unsubscribe / update 等
	Payload []interface{} `json:"payload,omitempty"` // 负载
	Auth    *wsAuth       `json:"auth,omitempty"`    // 鉴权信息(私有频道必需)
}

// wsAuth WebSocket 鉴权字段
type wsAuth struct {
	Method string `json:"method"` // 目前仅支持 api_key
	Key    string `json:"KEY"`    // API Key
	Sign   string `json:"SIGN"`   // 签名
}

// wsSign 生成 WebSocket 鉴权签名:
// s = fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, timestamp)
// sign = HMAC_SHA512(s, secret)
func (g *Gate) wsSign(channel, event string, timestamp int64) *wsAuth {
	if g.apiKey == "" || g.secretKey == "" {
		return nil
	}

	msg := fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, timestamp)
	mac := hmac.New(sha512.New, []byte(g.secretKey))
	mac.Write([]byte(msg))

	return &wsAuth{
		Method: "api_key",
		Key:    g.apiKey,
		Sign:   hex.EncodeToString(mac.Sum(nil)),
	}
}

// newWsDialer 构建带代理的 Dialer
func (g *Gate) newWsDialer() *websocket.Dialer {
	dialer := &websocket.Dialer{
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	if g.proxyURL != "" {
		proxyURLParsed, err := url.Parse(g.proxyURL)
		if err == nil {
			scheme := strings.ToLower(proxyURLParsed.Scheme)
			if scheme == "socks" || scheme == "socks5" {
				d, err := proxy.SOCKS5("tcp", proxyURLParsed.Host, nil, proxy.Direct)
				if err == nil {
					dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						return d.Dial(network, addr)
					}
				}
			} else {
				dialer.Proxy = http.ProxyURL(proxyURLParsed)
			}
		}
	}

	return dialer
}

// wsServe Gate WebSocket 底层连接封装（带自动重连）
// onConnected: 连接建立后立即回调, 用于发送订阅/鉴权请求
func (g *Gate) wsServe(
	ctx context.Context,
	onConnected func(*websocket.Conn) error,
	messageHandler func([]byte),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	doneC = make(chan struct{})
	stopC = make(chan struct{})

	var reconnectAttempts int64
	reconnectDelay := wsReconnectInitialDelay

	// 重连循环
	go func() {
		defer close(doneC)

		for {
			// 检查停止条件
			select {
			case <-stopC:
				return
			case <-ctx.Done():
				return
			default:
			}

			// 检查重连次数限制
			if wsReconnectMaxAttempts > 0 {
				if atomic.LoadInt64(&reconnectAttempts) >= int64(wsReconnectMaxAttempts) {
					if errHandler != nil {
						errHandler(fmt.Errorf("达到最大重连次数 %d，停止重连", wsReconnectMaxAttempts))
					}
					return
				}
			}

			// 建立单次连接
			connDoneC, connStopC, err := g.wsServeOnce(ctx, onConnected, messageHandler, errHandler)
			if err != nil {
				if errHandler != nil {
					errHandler(fmt.Errorf("Gate WebSocket 连接失败: %w，%v 后重试", err, reconnectDelay))
				}
				select {
				case <-stopC:
					return
				case <-ctx.Done():
					return
				case <-time.After(reconnectDelay):
					reconnectDelay *= 2
					if reconnectDelay > wsReconnectMaxDelay {
						reconnectDelay = wsReconnectMaxDelay
					}
					atomic.AddInt64(&reconnectAttempts, 1)
					continue
				}
			}

			// 连接成功，重置计数
			atomic.StoreInt64(&reconnectAttempts, 0)
			reconnectDelay = wsReconnectInitialDelay

			// 监听停止信号
			var closeConnStopOnce sync.Once
			go func() {
				<-stopC
				closeConnStopOnce.Do(func() { close(connStopC) })
			}()

			// 等待连接结束
			select {
			case <-connDoneC:
				// 连接断开，准备重连
				select {
				case <-stopC:
					return
				case <-ctx.Done():
					return
				default:
					if errHandler != nil {
						errHandler(fmt.Errorf("Gate WebSocket 连接断开，%v 后重连", reconnectDelay))
					}
					select {
					case <-stopC:
						return
					case <-ctx.Done():
						return
					case <-time.After(reconnectDelay):
						reconnectDelay *= 2
						if reconnectDelay > wsReconnectMaxDelay {
							reconnectDelay = wsReconnectMaxDelay
						}
						atomic.AddInt64(&reconnectAttempts, 1)
					}
				}
			case <-stopC:
				closeConnStopOnce.Do(func() { close(connStopC) })
				return
			case <-ctx.Done():
				closeConnStopOnce.Do(func() { close(connStopC) })
				return
			}
		}
	}()

	return doneC, stopC, nil
}

// wsServeOnce 单次 WebSocket 连接（内部使用）
func (g *Gate) wsServeOnce(
	ctx context.Context,
	onConnected func(*websocket.Conn) error,
	messageHandler func([]byte),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	doneC = make(chan struct{})
	stopC = make(chan struct{})

	// Gate UrlWs 一般为完整 ws 入口, 如:
	//  - 现货: wss://api.gateio.ws/ws/v4/
	//  - 永续: wss://fx-ws.gateio.ws/v4/ws/usdt
	fullURL := strings.TrimRight(g.UrlWs, "/") + "/"

	dialer := g.newWsDialer()

	conn, _, err := dialer.DialContext(ctx, fullURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Gate WebSocket 连接失败: %w", err)
	}

	// 设置读取限制
	conn.SetReadLimit(wsReadLimit)

	// 启动保活机制：跟踪连接活跃状态
	keepAliveCtx, keepAliveCancel := context.WithCancel(context.Background())
	defer func() {
		// 确保在所有错误路径上都能取消 context
		if err != nil {
			keepAliveCancel()
		}
	}()

	var lastResponse int64
	atomic.StoreInt64(&lastResponse, time.Now().Unix())

	// 使用 sync.Once 确保连接只关闭一次
	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			conn.Close()
		})
	}

	// 设置 ping handler：响应服务器发送的协议层 ping
	// Gate.io 服务器会主动发送协议层 ping，客户端需要回复 pong
	conn.SetPingHandler(func(pingData string) error {
		// 使用服务器的 ping payload 回复 pong
		err := conn.WriteControl(
			websocket.PongMessage,
			[]byte(pingData),
			time.Now().Add(10*time.Second),
		)
		if err != nil {
			return err
		}
		atomic.StoreInt64(&lastResponse, time.Now().Unix())
		return nil
	})

	// 设置 pong handler：跟踪协议层 pong（如果服务器发送）
	conn.SetPongHandler(func(pongData string) error {
		atomic.StoreInt64(&lastResponse, time.Now().Unix())
		return nil
	})

	// 连接建立后执行订阅/鉴权
	if onConnected != nil {
		if err := onConnected(conn); err != nil {
			closeConn()
			return nil, nil, fmt.Errorf("Gate WebSocket 订阅/鉴权失败: %w", err)
		}
	}

	// 定期发送应用层 ping 保持连接（futures.ping）
	// Gate.io 服务器会发送协议层 ping，但我们也发送应用层 ping 作为额外保活
	pingTicker := time.NewTicker(wsPingInterval)
	go func() {
		defer pingTicker.Stop()
		for {
			select {
			case <-stopC:
				return
			case <-keepAliveCtx.Done():
				return
			case <-pingTicker.C:
				// 发送应用层 ping 消息
				pingMsg := map[string]interface{}{
					"time":    time.Now().Unix(),
					"channel": "futures.ping",
				}
				pingData, err := json.Marshal(pingMsg)
				if err == nil {
					if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
						return
					}
					// 更新最后响应时间（发送 ping 也表示连接活跃）
					atomic.StoreInt64(&lastResponse, time.Now().Unix())
				}
			}
		}
	}()

	// 保活 goroutine：监控超时（更频繁地检查）
	go func() {
		ticker := time.NewTicker(wsTimeoutCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-keepAliveCtx.Done():
				return
			case <-ticker.C:
				// 检查是否超时
				lastRespTime := time.Unix(atomic.LoadInt64(&lastResponse), 0)
				if time.Since(lastRespTime) > wsTimeout {
					// 超时，关闭连接
					closeConn()
					return
				}
			}
		}
	}()

	// 主动关闭通道
	go func() {
		<-stopC
		keepAliveCancel()
		pingTicker.Stop()
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "normal closure"))
		closeConn()
	}()

	// 读取循环
	go func() {
		defer close(doneC)
		defer keepAliveCancel()
		defer closeConn()

		// 监听 context 取消，在单独的 goroutine 中处理
		go func() {
			select {
			case <-ctx.Done():
				// context 被取消，关闭连接
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context canceled"))
				closeConn()
			case <-stopC:
				// stopC 已处理，这里不需要额外操作
			}
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				// 检查 context 是否已取消
				if ctx.Err() != nil {
					return
				}
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				if errHandler != nil {
					errHandler(err)
				}
				return
			}

			// 更新最后响应时间（收到任何消息都表示连接活跃）
			atomic.StoreInt64(&lastResponse, time.Now().Unix())

			if messageHandler != nil {
				messageHandler(message)
			}
		}
	}()

	return doneC, stopC, nil
}

// wsServeJSON JSON 消息处理封装（带自动重连）
// 自动处理协议层 ping/pong（通过 SetPongHandler）
func (g *Gate) wsServeJSON(
	ctx context.Context,
	onConnected func(*websocket.Conn) error,
	newEvent func() interface{},
	eventHandler func(interface{}, []byte),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	return g.wsServe(ctx, onConnected, func(msg []byte) {
		evt := newEvent()
		if err := utils.SmartUnmarshal(msg, evt); err != nil {
			if errHandler != nil {
				errHandler(fmt.Errorf("解析 Gate WebSocket JSON 消息失败: %w", err))
			}
			return
		}
		if eventHandler != nil {
			eventHandler(evt, msg)
		}
	}, errHandler)
}
