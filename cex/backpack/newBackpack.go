package backpack

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"
)

type Backpack struct {
	apiKey      string             // Backpack API公钥 (Base64编码的ED25519公钥)
	privateKey  ed25519.PrivateKey // ED25519私钥 (从Base64编码的seed生成)
	proxyURL    string             // 代理URL
	HttpClient  *http.Client       // HTTP客户端
	UrlWs       string             // WebSocket URL
	UrlRest     string             // REST API URL
	Exc         ExchangeInfo       // 合约交易对基础信息
	rateLimiter *rate.Limiter      // 速率限制器
}

// New 创建Backpack客户端
// apiKey: Base64编码的ED25519公钥
// privateKeySeed: Base64编码的ED25519私钥种子(32字节)
// proxyURL: 代理URL,支持http/https/socks5
func New(apiKey, privateKeySeed, proxyURL string) *Backpack {
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

	// 从Base64编码的seed生成ED25519私钥
	var privateKey ed25519.PrivateKey
	if privateKeySeed != "" {
		seed, err := base64.StdEncoding.DecodeString(privateKeySeed)
		if err == nil && len(seed) == ed25519.SeedSize {
			privateKey = ed25519.NewKeyFromSeed(seed)
		}
	}

	backpack := &Backpack{
		apiKey:      apiKey,
		privateKey:  privateKey,
		proxyURL:    proxyURL,
		HttpClient:  httpClient,
		UrlWs:       "wss://ws.backpack.exchange/",
		UrlRest:     "https://api.backpack.exchange",
		Exc:         ExchangeInfo{},
		rateLimiter: rate.NewLimiter(rate.Limit(20), 20), // 每秒20次请求
	}

	backpack.Exc = make(ExchangeInfo)
	return backpack
}
