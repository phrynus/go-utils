package gate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/phrynus/go-utils"
)

// HttpRequest 通用HTTP请求结构
type HttpRequest struct {
	gate        *Gate                  // Gate客户端实例
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

// buildQueryString 构建查询字符串，返回用于签名的（未编码）和用于URL的（已编码）
// Gate.io要求: 没有使用URL编码的Request Parameters，顺序要和实际请求一致
func (h *HttpRequest) buildQueryString() (queryStringForSignature string, queryStringForUrl string) {
	if len(h.params) == 0 {
		return "", ""
	}

	// 对键进行排序以保证顺序一致（Gate.io要求顺序一致）
	keys := make([]string, 0, len(h.params))
	for key := range h.params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var partsForSignature []string
	values := url.Values{}

	for _, key := range keys {
		value := h.params[key]
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case int:
			valueStr = strconv.Itoa(v)
		case int64:
			valueStr = strconv.FormatInt(v, 10)
		case float64:
			valueStr = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			valueStr = strconv.FormatBool(v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		// 用于签名的未编码字符串
		partsForSignature = append(partsForSignature, key+"="+valueStr)
		// 用于URL的已编码值
		values.Add(key, valueStr)
	}

	queryStringForSignature = strings.Join(partsForSignature, "&")
	queryStringForUrl = values.Encode()
	return
}

// calculateBodyHash 计算请求体的SHA512哈希值
// 如果没有请求体，返回空字符串的SHA512哈希值
func (h *HttpRequest) calculateBodyHash(body []byte) string {
	if len(body) == 0 {
		// 空字符串的SHA512哈希值
		hash := sha512.Sum512([]byte(""))
		return hex.EncodeToString(hash[:])
	}
	hash := sha512.Sum512(body)
	return hex.EncodeToString(hash[:])
}

// generateSignature 生成HMAC SHA512签名 (Gate.io V4)
func (h *HttpRequest) generateSignature(method, requestPath, queryString, bodyHash, timestamp string) string {
	// Gate.io V4签名格式: method + "\n" + requestPath + "\n" + queryString + "\n" + bodyHash + "\n" + timestamp
	signaturePayload := method + "\n" + requestPath + "\n" + queryString + "\n" + bodyHash + "\n" + timestamp

	mac := hmac.New(sha512.New, []byte(h.gate.secretKey))
	mac.Write([]byte(signaturePayload))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildFullUrl 构建完整的请求URL和签名
func (h *HttpRequest) buildFullUrl(timestamp string, method string, body []byte) (string, string) {
	// 获取用于签名的（未编码）和用于URL的（已编码）查询字符串
	queryStringForSignature, queryStringForUrl := h.buildQueryString()

	// 构建完整URL: baseUrl + apiUrl + queryString
	fullUrl := h.baseUrl + h.apiUrl
	if queryStringForUrl != "" {
		fullUrl += "?" + queryStringForUrl
	}

	signature := ""
	// 如果需要签名,生成签名
	if h.sign {
		// 计算body的SHA512哈希值
		bodyHash := h.calculateBodyHash(body)
		// Gate.io V4 要求 Request URL 包含 /api/v4 前缀
		// 从 baseUrl 中提取路径部分，通常是 /api/v4
		requestPath := "/api/v4" + h.apiUrl
		// 使用未编码的查询字符串进行签名
		signature = h.generateSignature(method, requestPath, queryStringForSignature, bodyHash, timestamp)
	}

	return fullUrl, signature
}

// doRequest 执行HTTP请求
func (h *HttpRequest) doRequest(ctx context.Context, method string, requestBody []byte) ([]byte, error) {
	// 等待速率限制器允许
	if err := h.gate.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制器错误: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fullUrl, signature := h.buildFullUrl(timestamp, method, requestBody)

	// 创建请求体
	var bodyReader io.Reader
	if len(requestBody) > 0 {
		bodyReader = bytes.NewReader(requestBody)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullUrl, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if h.gate.apiKey != "" {
		req.Header.Set("KEY", h.gate.apiKey)
		req.Header.Set("Timestamp", timestamp)
		if h.sign {
			req.Header.Set("SIGN", signature)
		}
	}

	// 发送请求
	resp, err := h.gate.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码：Gate 在创建资源（如下单）时返回 201 Created，也应视为成功
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("请求失败,状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Get 发送GET请求
func (h *HttpRequest) Get(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodGet, nil)
}

// Post 发送POST请求
func (h *HttpRequest) Post(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodPost, nil)
}

// PostWithBody 发送带请求体的POST请求
func (h *HttpRequest) PostWithBody(ctx context.Context, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}
	return h.doRequest(ctx, http.MethodPost, bodyBytes)
}

// GetJSON 发送GET请求并解析JSON响应
func (h *HttpRequest) GetJSON(ctx context.Context, v interface{}) error {
	body, err := h.Get(ctx)
	if err != nil {
		return err
	}

	// Gate.io错误响应格式检查
	var errorResp struct {
		Label   string `json:"label"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		// 如果有label和message字段,说明是错误响应
		if errorResp.Label != "" {
			return fmt.Errorf("业务错误 [%s]: %s", errorResp.Label, errorResp.Message)
		}
	}

	return utils.SmartUnmarshal(body, v)
}

// PostJSON 发送POST请求并解析JSON响应
func (h *HttpRequest) PostJSON(ctx context.Context, v interface{}) error {
	body, err := h.Post(ctx)
	if err != nil {
		return err
	}

	// Gate.io错误响应格式检查
	var errorResp struct {
		Label   string `json:"label"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		// 如果有label和message字段,说明是错误响应
		if errorResp.Label != "" {
			return fmt.Errorf("业务错误 [%s]: %s", errorResp.Label, errorResp.Message)
		}
	}

	return utils.SmartUnmarshal(body, v)
}

// PostJSONWithBody 发送带请求体的POST请求并解析JSON响应
func (h *HttpRequest) PostJSONWithBody(ctx context.Context, requestBody interface{}, responseBody interface{}) error {
	body, err := h.PostWithBody(ctx, requestBody)
	if err != nil {
		return err
	}

	// Gate.io错误响应格式检查
	var errorResp struct {
		Label   string `json:"label"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		// 如果有label和message字段,说明是错误响应
		if errorResp.Label != "" {
			return fmt.Errorf("业务错误 [%s]: %s", errorResp.Label, errorResp.Message)
		}
	}

	return utils.SmartUnmarshal(body, responseBody)
}
