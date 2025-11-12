package gate

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"
)

type Gate struct {
	apiKey      string        // Gate API密钥
	secretKey   string        // Gate API密钥
	proxyURL    string        // 代理URL
	HttpClient  *http.Client  // HTTP客户端
	UrlWs       string        // WebSocket URL
	UrlRest     string        // REST API URL
	Exc         ExchangeInfo  // 合约交易对基础信息
	rateLimiter *rate.Limiter // 速率限制器
}

func New(apiKey, secretKey, proxyURL string) *Gate {
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

	gate := &Gate{
		apiKey:      apiKey,
		secretKey:   secretKey,
		proxyURL:    proxyURL,
		HttpClient:  httpClient,
		UrlWs:       "wss://ws.gate.io/v4",
		UrlRest:     "https://api.gateio.ws/api/v4",
		Exc:         ExchangeInfo{},
		rateLimiter: rate.NewLimiter(rate.Limit(20), 20), // 每秒20次请求
	}
	gate.Exc = make(ExchangeInfo)
	return gate
}
