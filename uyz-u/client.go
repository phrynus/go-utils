package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/phrynus/go-utils/uyz-u/crypto"
)

// APIResponse 镜像平台返回的顶层结构
type APIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time int64  `json:"time"`
	Data string `json:"data,omitempty"`
	Sign string `json:"sign,omitempty"`
}

// KamiTopupRequest 将卡密充值到当前账户
type KamiTopupRequest struct {
	Token    string `json:"token"`
	Kami     string `json:"kami"`
	Password string `json:"password,omitempty"`
	Time     int64  `json:"time"`
}

// FenRequest 验证用户是否可以消费积分
type FenRequest struct {
	Token   string `json:"token"`
	FenID   int    `json:"fenid"`
	FenMark string `json:"fenmark,omitempty"`
	Time    int64  `json:"time"`
}

// HeartbeatRequest 向服务器发送心跳以保持会话活跃
type HeartbeatRequest struct {
	Token string `json:"token"`
	Time  int64  `json:"time"`
}

// LogoutRequest 终止当前会话
type LogoutRequest struct {
	Token string `json:"token"`
	Time  int64  `json:"time"`
}

// CloudFunctionRequest 调用在控制台中创建的自定义函数
type CloudFunctionRequest struct {
	Token string `json:"token,omitempty"`
	Name  string `json:"name"`
	Param string `json:"param,omitempty"`
	Time  int64  `json:"time"`
}

// Client 提供用于调用 uverif API 的类型化辅助方法
type Client struct {
	cfg     ClientConfig
	baseURL string
	http    *http.Client
	token   string
	tokenMu sync.RWMutex
}

// Config 控制 SDK 如何与 uverif 后端通信
type ClientConfig struct {
	BaseURL          string         // 例如: https://uverif.xxx/api/user
	AppID            int            // 应用 ID
	AppKey           string         // 用于 MD5 签名
	Version          string         // 客户端语义版本，例如: "1.0.0"
	VersionIndex     string         // 例如: "web"
	ClientPrivateKey string         // PEM 格式的私钥，用于解密 payload
	ServerPublicKey  string         // PEM 格式的公钥，用于加密 payload
	HTTPTimeout      time.Duration  // 可选；为零时使用默认值
	EncryptionMode   EncryptionMode // AES/DES/RC4/RSA/none
	EncodingMode     EncodingMode   // 对称模式的编码方式：base64 或 hex
	SymmetricKey     string         // AES/DES/RC4 的共享密钥
	DisableSignature bool           // 为 true 时，省略 MD5 签名
	ProxyURL         string         // 代理 URL 可使用 utils.GetProxy() 获取代理URL
}

// EncryptionMode 枚举支持的 payload 保护策略
type EncryptionMode string

// 支持的加密模式
const (
	EncryptionRSA  EncryptionMode = "rsa"
	EncryptionAES  EncryptionMode = "aes"
	EncryptionDES  EncryptionMode = "des"
	EncryptionRC4  EncryptionMode = "rc4"
	EncryptionNone EncryptionMode = "none"
)

// EncodingMode 选择对称密文在传输时的编码方式
type EncodingMode string

const (
	EncodingBase64 EncodingMode = "base64"
	EncodingHex    EncodingMode = "hex"
)

func normalizeConfig(cfg ClientConfig) (ClientConfig, error) {
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return ClientConfig{}, err
	}
	return cfg, nil
}

func (cfg *ClientConfig) applyDefaults() {
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}
	if cfg.VersionIndex == "" {
		cfg.VersionIndex = "web"
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if cfg.EncryptionMode == "" {
		cfg.EncryptionMode = EncryptionNone
	}
	if cfg.EncodingMode == "" {
		cfg.EncodingMode = EncodingBase64
	}

}

func (cfg ClientConfig) validate() error {
	if cfg.BaseURL == "" {
		return errors.New("基础 URL 是必需的")
	}
	if cfg.AppID == 0 {
		return errors.New("应用 ID 是必需的")
	}
	if !cfg.DisableSignature && cfg.AppKey == "" {
		return errors.New("启用签名时需要应用密钥")
	}
	if err := validateProtection(cfg); err != nil {
		return err
	}
	return nil
}

func validateProtection(cfg ClientConfig) error {
	switch cfg.EncryptionMode {
	case EncryptionRSA, EncryptionAES, EncryptionDES, EncryptionRC4, EncryptionNone:
	default:
		return fmt.Errorf("不支持的加密模式 %q", cfg.EncryptionMode)
	}
	if requiresSymmetricKey(cfg.EncryptionMode) && cfg.SymmetricKey == "" {
		return errors.New("所选加密模式需要对称密钥")
	}
	if usesEncoding(cfg.EncryptionMode) {
		switch cfg.EncodingMode {
		case EncodingBase64, EncodingHex:
		default:
			return fmt.Errorf("不支持的编码模式 %q", cfg.EncodingMode)
		}
	}
	return nil
}

func requiresSymmetricKey(mode EncryptionMode) bool {
	return mode == EncryptionAES || mode == EncryptionDES || mode == EncryptionRC4
}

func usesEncoding(mode EncryptionMode) bool {
	return requiresSymmetricKey(mode)
}

// New 验证配置并返回一个可用的客户端
func New(cfg ClientConfig) (*Client, error) {
	var err error
	if cfg, err = normalizeConfig(cfg); err != nil {
		return nil, err
	}

	proxy, err := url.Parse(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{}
	if cfg.ProxyURL != "" {
		transport.Proxy = http.ProxyURL(proxy)
	}
	c := &Client{
		cfg:     cfg,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http: &http.Client{
			Timeout:   cfg.HTTPTimeout,
			Transport: transport,
		},
	}
	return c, nil
}

func (c *Client) buildURL(action string) string {
	return fmt.Sprintf("%s/%d/%s/%s/%s", c.baseURL, c.cfg.AppID, c.cfg.VersionIndex, c.cfg.Version, action)
}

func (c *Client) get(ctx context.Context, action string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(action), nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) postJSON(ctx context.Context, action string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(action), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 500 {
		return fmt.Errorf("远程服务器错误: %s", res.Status)
	}
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("请求失败: %s (%s)", res.Status, strings.TrimSpace(string(body)))
	}
	if out == nil {
		io.Copy(io.Discard, res.Body)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// buildSecurePayload 将 payload 进行 JSON 编码、加密，并可选地签名
func (c *Client) buildSecurePayload(payload any) (map[string]string, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 在加密模式下，如果payload没有Time字段，则添加一个
	if c.cfg.EncryptionMode != EncryptionNone {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			return nil, err
		}
		data["time"] = time.Now().Unix()
		jsonBytes, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	encrypted, err := c.encryptPayload(jsonBytes)
	if err != nil {
		return nil, err
	}
	body := map[string]string{
		"data": encrypted,
	}

	if !c.cfg.DisableSignature {
		body["sign"] = crypto.MD5Hex(string(jsonBytes) + c.cfg.AppKey)
	}
	return body, nil
}

func (c *Client) encryptPayload(plain []byte) (string, error) {
	switch c.cfg.EncryptionMode {
	case EncryptionNone:
		return string(plain), nil
	case EncryptionRSA:
		if c.cfg.ServerPublicKey == "" {
			return "", errors.New("RSA 加密需要服务器公钥")
		}
		return crypto.RSAEncrypt(c.cfg.ServerPublicKey, string(plain))
	case EncryptionAES:
		return crypto.AesEncrypt(c.cfg.SymmetricKey, string(plain), string(c.cfg.EncodingMode))
	case EncryptionDES:
		return crypto.DesEncrypt(c.cfg.SymmetricKey, string(plain), string(c.cfg.EncodingMode))
	case EncryptionRC4:
		return crypto.Rc4Encrypt(c.cfg.SymmetricKey, string(plain), string(c.cfg.EncodingMode))
	default:
		return "", fmt.Errorf("不支持的加密模式 %q", c.cfg.EncryptionMode)
	}
}

func (c *Client) decryptPayload(data string) (string, error) {
	if data == "" {
		return "", nil
	}
	switch c.cfg.EncryptionMode {
	case EncryptionNone:
		return data, nil
	case EncryptionRSA:
		if c.cfg.ClientPrivateKey == "" {
			return "", errors.New("解密响应数据需要客户端私钥")
		}
		return crypto.RSADecrypt(c.cfg.ClientPrivateKey, data)
	case EncryptionAES:
		return crypto.AesDecrypt(c.cfg.SymmetricKey, data, string(c.cfg.EncodingMode))
	case EncryptionDES:
		return crypto.DesDecrypt(c.cfg.SymmetricKey, data, string(c.cfg.EncodingMode))
	case EncryptionRC4:
		return crypto.Rc4Decrypt(c.cfg.SymmetricKey, data, string(c.cfg.EncodingMode))
	default:
		return "", fmt.Errorf("不支持的加密模式 %q", c.cfg.EncryptionMode)
	}
}

func (c *Client) decryptResponse(data string, out any) error {
	if out == nil || data == "" {
		return nil
	}
	plain, err := c.decryptPayload(data)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(plain), out)
}

// verifyResponseSignature 验证响应数据的签名是否正确
func (c *Client) verifyResponseSignature(resp APIResponse) error {
	if c.cfg.DisableSignature {
		return nil // 如果禁用了签名，跳过验证
	}
	if resp.Sign == "" {
		return nil // 如果没有签名，跳过验证
	}
	if c.cfg.AppKey == "" {
		return errors.New("验证签名需要应用密钥")
	}

	// 计算签名：MD5(JSON字符串 + AppKey)
	expectedSign := crypto.MD5Hex(strconv.Itoa(resp.Code) + strconv.FormatInt(resp.Time, 10) + c.cfg.AppKey)
	// 比较签名（不区分大小写）
	if !strings.EqualFold(expectedSign, resp.Sign) {
		return fmt.Errorf("签名验证失败: 期望 %s, 实际 %s", expectedSign, resp.Sign)
	}

	return nil
}

// SecurePost 为高级构建器提供加密 POST 功能
func (c *Client) SecurePost(ctx context.Context, action string, body any, out any) (APIResponse, error) {
	return c.securePost(ctx, action, body, out)
}

// RawGet 对指定的 action 执行普通 GET 请求
func (c *Client) RawGet(ctx context.Context, action string, out any) error {
	return c.get(ctx, action, out)
}

// DecryptResponse 帮助包解码加密的响应数据
func (c *Client) DecryptResponse(data string, out any) error {
	return c.decryptResponse(data, out)
}

// SetToken 设置客户端 token
func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
}

// GetToken 获取客户端 token，如果不存在则返回错误
func (c *Client) GetToken() (string, error) {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	if c.token == "" {
		return "", errors.New("未登录，请先调用登录接口")
	}
	return c.token, nil
}

// ClearToken 清除客户端 token
func (c *Client) ClearToken() {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = ""
}

// securePost 加密 payload，发送它，并可选地解密响应数据
func (c *Client) securePost(ctx context.Context, action string, body any, out any) (APIResponse, error) {
	payload, err := c.buildSecurePayload(body)
	if err != nil {
		return APIResponse{}, err
	}

	var resp APIResponse
	if err := c.postJSON(ctx, action, payload, &resp); err != nil {
		return APIResponse{}, err
	}

	// 如果响应包含签名，验证签名
	if err := c.verifyResponseSignature(resp); err != nil {
		return APIResponse{}, err
	}
	if err := c.decryptResponse(resp.Data, out); err != nil {
		return APIResponse{}, err
	}
	return resp, nil
}
