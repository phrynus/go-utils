package gate

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
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

// generateSignature 生成HMAC SHA512签名 (Gate.io V4)
func (h *HttpRequest) generateSignature(method, requestPath, queryString, bodyHash, timestamp string) string {
	// Gate.io V4签名格式: method + "\n" + requestPath + "\n" + queryString + "\n" + bodyHash + "\n" + timestamp
	signaturePayload := method + "\n" + requestPath + "\n" + queryString + "\n" + bodyHash + "\n" + timestamp

	mac := hmac.New(sha512.New, []byte(h.gate.secretKey))
	mac.Write([]byte(signaturePayload))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildFullUrl 构建完整的请求URL
func (h *HttpRequest) buildFullUrl(timestamp string, method string) (string, string) {
	queryString := h.buildQueryString()

	// 构建完整URL: baseUrl + apiUrl + queryString
	fullUrl := h.baseUrl + h.apiUrl
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	signature := ""
	// 如果需要签名,生成签名
	if h.sign {
		bodyHash := "" // GET请求body为空
		signature = h.generateSignature(method, h.apiUrl, queryString, bodyHash, timestamp)
	}

	return fullUrl, signature
}

// doRequest 执行HTTP请求
func (h *HttpRequest) doRequest(ctx context.Context, method string) ([]byte, error) {
	// 等待速率限制器允许
	if err := h.gate.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制器错误: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fullUrl, signature := h.buildFullUrl(timestamp, method)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullUrl, nil)
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
