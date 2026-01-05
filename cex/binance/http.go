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

// HttpRequest 通用HTTP请求
type HttpRequest struct {
	binance     *Binance               // 客户端
	baseUrl     string                 // 基础URL
	apiUrl      string                 // API路径
	params      map[string]interface{} // 请求参数
	sign        bool                   // 是否签名
	isTimestamp bool                   // 是否带时间戳
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
// internal use only
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

// generateSignature 生成HMAC-SHA256签名
// internal use only
func (h *HttpRequest) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(h.binance.secretKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildFullUrl 构建完整URL
// internal use only
func (h *HttpRequest) buildFullUrl() string {
	queryString := h.buildQueryString()

	if h.isTimestamp || h.sign {
		if queryString != "" {
			queryString += "&"
		}
		queryString += fmt.Sprintf("timestamp=%d", time.Now().UnixMilli())
	}

	if h.sign {
		signature := h.generateSignature(queryString)
		queryString += fmt.Sprintf("&signature=%s", signature)
	}
	fullUrl := h.baseUrl + h.apiUrl
	if queryString != "" {
		fullUrl += "?" + queryString
	}

	return fullUrl
}

// doRequest 执行HTTP请求
// internal use only
func (h *HttpRequest) doRequest(ctx context.Context, method string) ([]byte, error) {
	if err := h.binance.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("限流错误: %w", err)
	}

	fullUrl := h.buildFullUrl()

	req, err := http.NewRequestWithContext(ctx, method, fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if h.binance.apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", h.binance.apiKey)
	}

	resp, err := h.binance.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

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

// GetJSON GET请求+JSON解析
func (h *HttpRequest) GetJSON(ctx context.Context, v interface{}) error {
	body, err := h.Get(ctx)
	if err != nil {
		return err
	}

	var errorResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Code != 0 {
		return fmt.Errorf("业务错误 [%d]: %s", errorResp.Code, errorResp.Msg)
	}

	return utils.SmartUnmarshal(body, v)
}

// PostJSON POST请求+JSON解析
func (h *HttpRequest) PostJSON(ctx context.Context, v interface{}) error {
	body, err := h.Post(ctx)
	if err != nil {
		return err
	}

	var errorResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Code != 0 {
		return fmt.Errorf("业务错误 [%d]: %s", errorResp.Code, errorResp.Msg)
	}

	return utils.SmartUnmarshal(body, v)
}
