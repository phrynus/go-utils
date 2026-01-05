package binance

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"
)

// Binance 币安合约客户端
type Binance struct {
	apiKey      string        // API Key
	secretKey   string        // Secret Key
	proxyURL    string        // 代理地址
	HttpClient  *http.Client  // HTTP客户端
	UrlWs       string        // WS地址
	UrlRest     string        // REST地址
	Exc         ExchangeInfo  // 交易对信息
	ExcMu       sync.RWMutex  // Exc 并发保护
	Balance     *Balance      // 账户余额
	rateLimiter *rate.Limiter // 请求限流器
}

// New 创建 Binance 客户端
func New(apiKey, secretKey, proxyURL string) *Binance {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略TLS证书验证
			},
			MaxIdleConns:        100,              // 最大空闲连接数
			MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
			IdleConnTimeout:     60 * time.Second, // 空闲连接超时时间
		},
	}

	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err == nil {
			scheme := strings.ToLower(proxyURLParsed.Scheme)
			if scheme == "socks" || scheme == "socks5" {
				dialer, err := proxy.SOCKS5("tcp", proxyURLParsed.Host, nil, proxy.Direct)
				if err == nil {
					httpClient.Transport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						return dialer.Dial(network, addr)
					}
				}
			} else {
				httpClient.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURLParsed)
			}
		}
	}

	binance := &Binance{
		apiKey:      apiKey,
		secretKey:   secretKey,
		proxyURL:    proxyURL,
		HttpClient:  httpClient,
		UrlWs:       "wss://fstream.binance.com/ws",
		UrlRest:     "https://fapi.binance.com",
		Exc:         ExchangeInfo{},
		rateLimiter: rate.NewLimiter(rate.Limit(20), 20),
	}

	binance.Exc = make(ExchangeInfo)
	return binance
}
