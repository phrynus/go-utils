package binance

import (
	"context"
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
	// wsTimeout Binance 服务器在10分钟内没有收到pong会断开连接
	wsTimeout = 10 * time.Minute
	// wsTimeoutCheckInterval 超时检查间隔（应该比 wsTimeout 短得多）
	wsTimeoutCheckInterval = 30 * time.Second
	// wsPongWriteTimeout pong 消息写入超时
	wsPongWriteTimeout = 10 * time.Second
	// wsReconnectInitialDelay 初始重连延迟
	wsReconnectInitialDelay = 1 * time.Second
	// wsReconnectMaxDelay 最大重连延迟
	wsReconnectMaxDelay = 60 * time.Second
	// wsReconnectMaxAttempts 最大重连次数（0 表示无限重连）
	wsReconnectMaxAttempts = 0
)

// wsServe 底层WebSocket连接（带自动重连）
func (b *Binance) wsServe(
	ctx context.Context,
	path string,
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
			connDoneC, connStopC, err := b.wsServeOnce(ctx, path, messageHandler, errHandler)
			if err != nil {
				if errHandler != nil {
					errHandler(fmt.Errorf("WebSocket 连接失败: %w，%v 后重试", err, reconnectDelay))
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
						errHandler(fmt.Errorf("WebSocket 连接断开，%v 后重连", reconnectDelay))
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
func (b *Binance) wsServeOnce(
	ctx context.Context,
	path string,
	messageHandler func([]byte),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	doneC = make(chan struct{})
	stopC = make(chan struct{})

	base := strings.TrimRight(b.UrlWs, "/")
	path = strings.TrimLeft(path, "/")
	fullURL := base + "/" + path

	dialer := &websocket.Dialer{
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	if b.proxyURL != "" {
		proxyURLParsed, perr := url.Parse(b.proxyURL)
		if perr == nil {
			scheme := strings.ToLower(proxyURLParsed.Scheme)
			if scheme == "socks" || scheme == "socks5" {
				d, err := proxy.SOCKS5("tcp", proxyURLParsed.Host, nil, proxy.Direct)
				if err != nil {
					return nil, nil, fmt.Errorf("创建 SOCKS5 代理失败: %w", err)
				}
				dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return d.Dial(network, addr)
				}
			} else {
				dialer.Proxy = http.ProxyURL(proxyURLParsed)
			}
		}
	}

	conn, _, err := dialer.DialContext(ctx, fullURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("WebSocket 连接失败: %w", err)
	}

	// 设置读取限制
	conn.SetReadLimit(wsReadLimit)

	// 启动保活机制：响应服务器的协议层 ping
	keepAliveCtx, keepAliveCancel := context.WithCancel(context.Background())
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
	// Binance 服务器会主动发送协议层 ping，客户端需要回复 pong
	conn.SetPingHandler(func(pingData string) error {
		// 使用服务器的 ping payload 回复 pong
		err := conn.WriteControl(
			websocket.PongMessage,
			[]byte(pingData),
			time.Now().Add(wsPongWriteTimeout),
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
			messageType, message, err := conn.ReadMessage()
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

			// 优化：快速检查是否是应用层 ping（避免不必要的 JSON 解析）
			if messageType == websocket.TextMessage {
				// 快速检查：ping 消息通常很短，先检查长度和前缀
				if len(message) < 100 {
					var pingMsg map[string]interface{}
					if err := json.Unmarshal(message, &pingMsg); err == nil {
						// 检查是否是应用层 ping 消息
						if pingValue, ok := pingMsg["ping"]; ok {
							// 立即回复应用层 pong，payload 与 ping 一致
							pongMsg := map[string]interface{}{
								"pong": pingValue,
							}
							pongData, err := json.Marshal(pongMsg)
							if err == nil {
								if err := conn.WriteMessage(websocket.TextMessage, pongData); err != nil {
									if errHandler != nil {
										errHandler(fmt.Errorf("发送应用层 pong 消息失败: %w", err))
									}
									return
								}
								// 应用层 pong 消息已发送，跳过 messageHandler
								atomic.StoreInt64(&lastResponse, time.Now().Unix())
								continue
							}
						}
					}
				}
			}

			if messageHandler != nil {
				messageHandler(message)
			}
		}
	}()

	return doneC, stopC, nil
}

// wsServeJSON JSON消息处理（带自动重连）
func (b *Binance) wsServeJSON(
	ctx context.Context,
	path string,
	newEvent func() interface{},
	eventHandler func(interface{}, []byte),
	errHandler func(error),
) (doneC chan struct{}, stopC chan struct{}, err error) {
	return b.wsServe(ctx, path, func(msg []byte) {

		evt := newEvent()
		if err := utils.SmartUnmarshal(msg, evt); err != nil {
			if errHandler != nil {
				errHandler(fmt.Errorf("解析 WebSocket JSON 消息失败: %w", err))
			}
			return
		}
		if eventHandler != nil {
			eventHandler(evt, msg)
		}
	}, errHandler)
}
