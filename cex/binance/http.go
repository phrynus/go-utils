package binance

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
	binance     *Binance               // Binance客户端实例
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

// generateSignature 生成HMAC SHA256签名
func (h *HttpRequest) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(h.binance.secretKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildFullUrl 构建完整的请求URL
func (h *HttpRequest) buildFullUrl() string {
	queryString := h.buildQueryString()

	// 如果需要时间戳,添加timestamp参数
	if h.isTimestamp || h.sign {
		if queryString != "" {
			queryString += "&"
		}
		queryString += fmt.Sprintf("timestamp=%d", time.Now().UnixMilli())
	}

	// 如果需要签名,添加signature参数
	if h.sign {
		signature := h.generateSignature(queryString)
		queryString += fmt.Sprintf("&signature=%s", signature)
	}

	// 构建完整URL: baseUrl + apiUrl + queryString
	fullUrl := h.baseUrl + h.apiUrl
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	return fullUrl
}

// doRequest 执行HTTP请求
func (h *HttpRequest) doRequest(ctx context.Context, method string) ([]byte, error) {
	// 等待速率限制器允许
	if err := h.binance.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制器错误: %w", err)
	}

	fullUrl := h.buildFullUrl()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if h.binance.apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", h.binance.apiKey)
	}

	// 发送请求
	resp, err := h.binance.HttpClient.Do(req)
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

// Put 发送PUT请求
func (h *HttpRequest) Put(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodPut)
}

// Delete 发送DELETE请求
func (h *HttpRequest) Delete(ctx context.Context) ([]byte, error) {
	return h.doRequest(ctx, http.MethodDelete)
}

// GetJSON 发送GET请求并解析JSON响应
func (h *HttpRequest) GetJSON(ctx context.Context, v interface{}) error {
	body, err := h.Get(ctx)
	if err != nil {
		return err
	}

	// 先检查是否有错误响应
	var errorResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		// 如果有code字段且不为0,说明是错误响应
		if errorResp.Code != 0 {
			return fmt.Errorf("业务错误 [%d]: %s", errorResp.Code, errorResp.Msg)
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

	// 先检查是否有错误响应
	var errorResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		// 如果有code字段且不为0,说明是错误响应
		if errorResp.Code != 0 {
			return fmt.Errorf("业务错误 [%d]: %s", errorResp.Code, errorResp.Msg)
		}
	}

	return utils.SmartUnmarshal(body, v)
}
