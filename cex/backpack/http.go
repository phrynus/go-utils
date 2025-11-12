package backpack

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/phrynus/go-utils"
)

// InstructionType 指令类型
type InstructionType string

const (
	InstructionAccountQuery            InstructionType = "accountQuery"
	InstructionBalanceQuery            InstructionType = "balanceQuery"
	InstructionBorrowLendExecute       InstructionType = "borrowLendExecute"
	InstructionBorrowHistoryQueryAll   InstructionType = "borrowHistoryQueryAll"
	InstructionCollateralQuery         InstructionType = "collateralQuery"
	InstructionDepositAddressQuery     InstructionType = "depositAddressQuery"
	InstructionDepositQueryAll         InstructionType = "depositQueryAll"
	InstructionFillHistoryQueryAll     InstructionType = "fillHistoryQueryAll"
	InstructionFundingHistoryQueryAll  InstructionType = "fundingHistoryQueryAll"
	InstructionInterestHistoryQueryAll InstructionType = "interestHistoryQueryAll"
	InstructionOrderCancel             InstructionType = "orderCancel"
	InstructionOrderCancelAll          InstructionType = "orderCancelAll"
	InstructionOrderExecute            InstructionType = "orderExecute"
	InstructionOrderHistoryQueryAll    InstructionType = "orderHistoryQueryAll"
	InstructionOrderQuery              InstructionType = "orderQuery"
	InstructionOrderQueryAll           InstructionType = "orderQueryAll"
	InstructionPnlHistoryQueryAll      InstructionType = "pnlHistoryQueryAll"
	InstructionPositionQuery           InstructionType = "positionQuery"
	InstructionQuoteSubmit             InstructionType = "quoteSubmit"
	InstructionStrategyCancel          InstructionType = "strategyCancel"
	InstructionStrategyCancelAll       InstructionType = "strategyCancelAll"
	InstructionStrategyCreate          InstructionType = "strategyCreate"
	InstructionStrategyHistoryQueryAll InstructionType = "strategyHistoryQueryAll"
	InstructionStrategyQuery           InstructionType = "strategyQuery"
	InstructionStrategyQueryAll        InstructionType = "strategyQueryAll"
	InstructionWithdraw                InstructionType = "withdraw"
	InstructionWithdrawalQueryAll      InstructionType = "withdrawalQueryAll"
)

// HttpRequest 通用HTTP请求结构
type HttpRequest struct {
	backpack    *Backpack              // Backpack客户端实例
	baseUrl     string                 // 基础URL
	apiUrl      string                 // 请求URL
	params      map[string]interface{} // 请求参数(用于GET请求的query参数)
	body        interface{}            // 请求体(用于POST请求)
	sign        bool                   // 是否需要签名
	instruction InstructionType        // 指令类型
	window      int64                  // 请求有效时间窗口(毫秒),默认5000,最大60000
}

// SetParams 设置请求参数(用于GET请求)
func (h *HttpRequest) SetParams(params map[string]interface{}) *HttpRequest {
	h.params = params
	return h
}

// SetBody 设置请求体(用于POST请求)
func (h *HttpRequest) SetBody(body interface{}) *HttpRequest {
	h.body = body
	return h
}

// SetWindow 设置请求有效时间窗口(毫秒)
func (h *HttpRequest) SetWindow(window int64) *HttpRequest {
	h.window = window
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

// buildQueryString 构建查询字符串(按字母顺序排序)
func (h *HttpRequest) buildQueryString(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	// 获取所有key并排序
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 按排序后的key构建查询字符串
	values := url.Values{}
	for _, key := range keys {
		value := params[key]
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

// buildSignatureString 构建签名字符串
func (h *HttpRequest) buildSignatureString(queryString string, timestamp int64) string {
	// 设置默认窗口
	window := h.window
	if window == 0 {
		window = 5000 // 默认5000毫秒
	}
	if window > 60000 {
		window = 60000 // 最大60000毫秒
	}

	// 构建签名字符串: instruction=<instruction>&<params>&timestamp=<timestamp>&window=<window>
	signatureString := fmt.Sprintf("instruction=%s", h.instruction)
	if queryString != "" {
		signatureString += "&" + queryString
	}
	signatureString += fmt.Sprintf("&timestamp=%d&window=%d", timestamp, window)

	return signatureString
}

// buildBatchOrderSignatureString 构建批量订单签名字符串
// 批量订单需要特殊处理,每个订单都要添加instruction前缀
func (h *HttpRequest) buildBatchOrderSignatureString(orders []map[string]interface{}, timestamp int64) string {
	window := h.window
	if window == 0 {
		window = 5000
	}
	if window > 60000 {
		window = 60000
	}

	var signatureParts []string
	for _, order := range orders {
		// 每个订单都添加instruction前缀
		orderQuery := h.buildQueryString(order)
		signatureParts = append(signatureParts, fmt.Sprintf("instruction=%s&%s", h.instruction, orderQuery))
	}

	// 连接所有订单的查询字符串,并添加时间戳和窗口
	signatureString := ""
	for i, part := range signatureParts {
		if i > 0 {
			signatureString += "&"
		}
		signatureString += part
	}
	signatureString += fmt.Sprintf("&timestamp=%d&window=%d", timestamp, window)

	return signatureString
}

// generateSignature 生成ED25519签名
func (h *HttpRequest) generateSignature(message string) string {
	signature := ed25519.Sign(h.backpack.privateKey, []byte(message))
	return base64.StdEncoding.EncodeToString(signature)
}

// doRequest 执行HTTP请求
func (h *HttpRequest) doRequest(ctx context.Context, method string) ([]byte, error) {
	// 等待速率限制器允许
	if err := h.backpack.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制器错误: %w", err)
	}

	timestamp := time.Now().UnixMilli()
	window := h.window
	if window == 0 {
		window = 5000
	}
	if window > 60000 {
		window = 60000
	}

	var req *http.Request
	var err error

	if method == http.MethodGet {
		// GET请求:参数在URL中
		queryString := h.buildQueryString(h.params)
		fullUrl := h.baseUrl + h.apiUrl
		if queryString != "" {
			fullUrl += "?" + queryString
		}

		req, err = http.NewRequestWithContext(ctx, method, fullUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		// 如果需要签名
		if h.sign {
			signatureString := h.buildSignatureString(queryString, timestamp)
			signature := h.generateSignature(signatureString)

			req.Header.Set("X-API-Key", h.backpack.apiKey)
			req.Header.Set("X-Signature", signature)
			req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
			req.Header.Set("X-Window", strconv.FormatInt(window, 10))
		}
	} else {
		// POST/PUT/DELETE请求:参数在body中
		var bodyBytes []byte
		var queryString string

		// 检查是否是批量订单
		if orders, ok := h.body.([]map[string]interface{}); ok && h.instruction == InstructionOrderExecute {
			// 批量订单特殊处理
			bodyBytes, err = json.Marshal(orders)
			if err != nil {
				return nil, fmt.Errorf("序列化请求体失败: %w", err)
			}

			if h.sign {
				signatureString := h.buildBatchOrderSignatureString(orders, timestamp)
				signature := h.generateSignature(signatureString)

				req, err = http.NewRequestWithContext(ctx, method, h.baseUrl+h.apiUrl, bytes.NewReader(bodyBytes))
				if err != nil {
					return nil, fmt.Errorf("创建请求失败: %w", err)
				}

				req.Header.Set("X-API-Key", h.backpack.apiKey)
				req.Header.Set("X-Signature", signature)
				req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
				req.Header.Set("X-Window", strconv.FormatInt(window, 10))
			}
		} else {
			// 普通POST请求
			if h.body != nil {
				bodyBytes, err = json.Marshal(h.body)
				if err != nil {
					return nil, fmt.Errorf("序列化请求体失败: %w", err)
				}

				// 将body转换为map以便构建签名字符串
				var bodyMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
					queryString = h.buildQueryString(bodyMap)
				}
			}

			req, err = http.NewRequestWithContext(ctx, method, h.baseUrl+h.apiUrl, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, fmt.Errorf("创建请求失败: %w", err)
			}

			// 如果需要签名
			if h.sign {
				signatureString := h.buildSignatureString(queryString, timestamp)
				signature := h.generateSignature(signatureString)

				req.Header.Set("X-API-Key", h.backpack.apiKey)
				req.Header.Set("X-Signature", signature)
				req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
				req.Header.Set("X-Window", strconv.FormatInt(window, 10))
			}
		}
	}

	// 设置通用请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := h.backpack.HttpClient.Do(req)
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

	// 检查是否有错误响应
	var errorResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		if errorResp.Error != "" {
			return fmt.Errorf("业务错误: %s", errorResp.Error)
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

	// 检查是否有错误响应
	var errorResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		if errorResp.Error != "" {
			return fmt.Errorf("业务错误: %s", errorResp.Error)
		}
	}

	return utils.SmartUnmarshal(body, v)
}
