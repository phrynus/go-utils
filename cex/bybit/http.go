package bybit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/phrynus/go-utils"
)

// HttpRequest 通用HTTP请求结构
type HttpRequest struct {
	bybit       *Bybit                 // Bybit客户端实例
	baseUrl     string                 // 基础URL
	apiUrl      string                 // 请求URL
	params      map[string]interface{} // 请求参数
	sign        bool                   // 是否需要签名
	isTimestamp bool                   // 是否需要时间戳
}

// SetParams 设置请求参数
func (h *HttpRequest) SetParams(params map[string]interface{}) *HttpRequest {
	h.params = params
	return h
}

// GetUrl 获取完整请求URL
func (h *HttpRequest) GetUrl() string {
	return h.baseUrl + h.apiUrl
}

// NeedSign 是否需要签名
func (h *HttpRequest) NeedSign() bool {
	return h.sign
}

// NeedTimestamp 是否需要时间戳
func (h *HttpRequest) NeedTimestamp() bool {
	return h.isTimestamp
}

// buildQueryString 构建查询字符串
func (h *HttpRequest) buildQueryString() string {
	if len(h.params) == 0 {
		return ""
	}

	values := url.Values{}
	for key, value := range h.params {
		switch v := value.(type) {
		case string:
			values.Add(key, v)
		case int:
			values.Add(key, strconv.Itoa(v))
		case int64:
			values.Add(key, strconv.FormatInt(v, 10))
		case float64:
			values.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			values.Add(key, strconv.FormatBool(v))
		default:
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}

	return values.Encode()
}

// generateSignature 生成HMAC SHA256签名 (Bybit V5)
func (h *HttpRequest) generateSignature(timestamp string, queryString string) string {
	// Bybit V5签名格式: timestamp + apiKey + recvWindow + queryString
	recvWindow := "5000"
	signaturePayload := timestamp + h.bybit.apiKey + recvWindow + queryString

	mac := hmac.New(sha256.New, []byte(h.bybit.secretKey))
	mac.Write([]byte(signaturePayload))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildFullUrl 构建完整的请求URL
func (h *HttpRequest) buildFullUrl(timestamp string) (string, string) {
	queryString := h.buildQueryString()
	signature := ""

	// 如果需要签名,生成签名
	if h.sign {
		signature = h.generateSignature(timestamp, queryString)
	}

	// 构建完整URL: baseUrl + apiUrl + queryString
	fullUrl := h.baseUrl + h.apiUrl
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	return fullUrl, signature
}

// doRequest 执行HTTP请求
func (h *HttpRequest) doRequest(ctx context.Context, method string) ([]byte, error) {
	// 等待速率限制器允许
	if err := h.bybit.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制器错误: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	fullUrl, signature := h.buildFullUrl(timestamp)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if h.bybit.apiKey != "" {
		req.Header.Set("X-BAPI-API-KEY", h.bybit.apiKey)
		req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
		req.Header.Set("X-BAPI-RECV-WINDOW", "5000")
		if h.sign {
			req.Header.Set("X-BAPI-SIGN", signature)
		}
	}

	// 发送请求
	resp, err := h.bybit.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败,状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Get 发送GET请求
func (h *HttpRequest) Get(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodGet)
}

// Post 发送POST请求
func (h *HttpRequest) Post(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodPost)
}

// GetJSON 发送GET请求并解析JSON响应
func (h *HttpRequest) GetJSON(ctx context.Context, v interface{}) error {
	body, err := h.Get(ctx)
	if err != nil {
		return err
	}

	// 先解析到临时结构检查错误码
	var result struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查业务错误码
	if result.RetCode != 0 {
		return fmt.Errorf("业务错误 [%d]: %s", result.RetCode, result.RetMsg)
	}

	return utils.SmartUnmarshal(body, v)
}

// PostJSON 发送POST请求并解析JSON响应
func (h *HttpRequest) PostJSON(ctx context.Context, v interface{}) error {
	body, err := h.Post(ctx)
	if err != nil {
		return err
	}

	// 先解析到临时结构检查错误码
	var result struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查业务错误码
	if result.RetCode != 0 {
		return fmt.Errorf("业务错误 [%d]: %s", result.RetCode, result.RetMsg)
	}

	return utils.SmartUnmarshal(body, v)
}
