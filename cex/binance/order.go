package binance

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Order 下单请求构建器
type Order struct {
	client      *Binance
	httpRequest *HttpRequest
	lastErr     error
}

// OrderResponse 下单响应结构（按示例字段选择常用字段）
type OrderResponse struct {
	ClientOrderId           string  `json:"clientOrderId"`
	CumQty                  float64 `json:"cumQty,string"`
	CumQuote                float64 `json:"cumQuote,string"`
	ExecutedQty             float64 `json:"executedQty,string"`
	OrderId                 int64   `json:"orderId"`
	AvgPrice                float64 `json:"avgPrice,string"`
	OrigQty                 float64 `json:"origQty,string"`
	Price                   float64 `json:"price,string"`
	ReduceOnly              bool    `json:"reduceOnly"`
	Side                    string  `json:"side"`
	PositionSide            string  `json:"positionSide"`
	Status                  string  `json:"status"`
	StopPrice               float64 `json:"stopPrice,string"`
	ClosePosition           bool    `json:"closePosition"`
	Symbol                  string  `json:"symbol"`
	TimeInForce             string  `json:"timeInForce"`
	Type                    string  `json:"type"`
	OrigType                string  `json:"origType"`
	UpdateTime              int64   `json:"updateTime"`
	WorkingType             string  `json:"workingType"`
	PriceProtect            bool    `json:"priceProtect"`
	PriceMatch              string  `json:"priceMatch"`
	SelfTradePreventionMode string  `json:"selfTradePreventionMode"`
	GoodTillDate            int64   `json:"goodTillDate,omitempty"`
}

// NewOrder 创建下单请求
func (b *Binance) NewOrder() *Order {
	return &Order{
		client: b,
		httpRequest: &HttpRequest{
			binance:     b,
			baseUrl:     b.UrlRest,
			apiUrl:      "/fapi/v1/order",
			sign:        true,
			isTimestamp: true,
			params:      make(map[string]interface{}),
		},
	}
}

// Test 切换到测试下单端点（/fapi/v1/order/test）
func (o *Order) Test() *Order {
	o.httpRequest.apiUrl = "/fapi/v1/order/test"
	return o
}

// Symbol 设置交易对
func (o *Order) Symbol(symbol string) *Order {
	// 如果之前已有错误，直接返回链式对象（保留第一次错误）
	if o.lastErr != nil {
		return o
	}
	if strings.TrimSpace(symbol) == "" {
		o.lastErr = fmt.Errorf("symbol 不能为空")
		return o
	}
	o.httpRequest.params["symbol"] = symbol
	return o
}

// Side 设置买卖方向，允许 BUY 或 SELL（不区分大小写）
func (o *Order) Side(side string) *Order {

	if o.lastErr != nil {
		return o
	}
	s := strings.ToUpper(strings.TrimSpace(side))
	if s != "BUY" && s != "SELL" {
		o.lastErr = fmt.Errorf("side 必须为 BUY 或 SELL")
		return o
	}
	o.httpRequest.params["side"] = s
	return o
}

// PositionSide 设置持仓方向，允许 LONG、SHORT 或 BOTH
func (o *Order) PositionSide(positionSide string) *Order {

	if o.lastErr != nil {
		return o
	}
	ps := strings.ToUpper(strings.TrimSpace(positionSide))
	if ps != "LONG" && ps != "SHORT" && ps != "BOTH" {
		o.lastErr = fmt.Errorf("positionSide 必须为 LONG、SHORT 或 BOTH")
		return o
	}
	o.httpRequest.params["positionSide"] = ps
	return o
}

// Type 设置订单类型（LIMIT/ MARKET/ STOP 等）
// 支持: LIMIT, MARKET, STOP, TAKE_PROFIT, STOP_MARKET, TAKE_PROFIT_MARKET, TRAILING_STOP_MARKET
func (o *Order) Type(t string) *Order {

	if o.lastErr != nil {
		return o
	}
	tt := strings.ToUpper(strings.TrimSpace(t))
	allowedTypes := map[string]bool{
		"LIMIT": true, "MARKET": true, "STOP": true, "TAKE_PROFIT": true,
		"STOP_MARKET": true, "TAKE_PROFIT_MARKET": true, "TRAILING_STOP_MARKET": true,
	}
	if !allowedTypes[tt] {
		o.lastErr = fmt.Errorf("type 值不合法: %s", tt)
		return o
	}
	o.httpRequest.params["type"] = tt
	return o
}

// TimeInForce 设置 timeInForce，支持 GTC/IOC/FOK/GTD（可选）
func (o *Order) TimeInForce(tif string) *Order {

	if o.lastErr != nil {
		return o
	}
	tf := strings.ToUpper(strings.TrimSpace(tif))
	if tf != "GTC" && tf != "IOC" && tf != "FOK" && tf != "GTD" {
		o.lastErr = fmt.Errorf("timeInForce 值不合法: %s", tf)
		return o
	}
	o.httpRequest.params["timeInForce"] = tf
	return o
}

// ReduceOnly 设置是否仅减仓（可选）
func (o *Order) ReduceOnly(reduce bool) *Order {

	if o.lastErr != nil {
		return o
	}
	o.httpRequest.params["reduceOnly"] = reduce
	return o
}

// Quantity 设置数量，必须大于 0
func (o *Order) Quantity(q float64) *Order {

	if o.lastErr != nil {
		return o
	}
	if q <= 0 {
		o.lastErr = fmt.Errorf("quantity 必须大于 0")
		return o
	}
	o.httpRequest.params["quantity"] = q
	return o
}

// Price 设置价格，必须大于等于 0（0 在某些类型下有意义）
func (o *Order) Price(p float64) *Order {

	if o.lastErr != nil {
		return o
	}
	if p < 0 {
		o.lastErr = fmt.Errorf("price 不能为负数")
		return o
	}
	o.httpRequest.params["price"] = p
	return o
}

// NewClientOrderId 设置用户自定义订单号，长度与格式受限（可选）
func (o *Order) NewClientOrderId(id string) *Order {

	if o.lastErr != nil {
		return o
	}
	if id == "" {
		// 允许空，由系统生成；不作为错误
		o.httpRequest.params["newClientOrderId"] = id
		return o
	}
	if len(id) < 1 || len(id) > 36 {
		o.lastErr = fmt.Errorf("newClientOrderId 长度必须在1到36之间")
		return o
	}
	re := regexp.MustCompile(`^[\.A-Z\:/a-z0-9_-]{1,36}$`)
	if !re.MatchString(id) {
		o.lastErr = fmt.Errorf("newClientOrderId 格式不合法")
		return o
	}
	o.httpRequest.params["newClientOrderId"] = id
	return o
}

// NewOrderRespType 设置返回类型，允许 ACK 或 RESULT（可选）
func (o *Order) NewOrderRespType(rt string) *Order {

	if o.lastErr != nil {
		return o
	}
	if rt == "" {
		o.httpRequest.params["newOrderRespType"] = rt
		return o
	}
	r := strings.ToUpper(strings.TrimSpace(rt))
	if r != "ACK" && r != "RESULT" {
		o.lastErr = fmt.Errorf("newOrderRespType 必须为 ACK 或 RESULT")
		return o
	}
	o.httpRequest.params["newOrderRespType"] = r
	return o
}

// PriceMatch 设置盘口价格下单模式（与 price 互斥）
func (o *Order) PriceMatch(pm string) *Order {

	if o.lastErr != nil {
		return o
	}
	if pm == "" {
		return o
	}
	p := strings.ToUpper(strings.TrimSpace(pm))
	allowed := map[string]bool{
		"OPPONENT": true, "OPPONENT_5": true, "OPPONENT_10": true, "OPPONENT_20": true,
		"QUEUE": true, "QUEUE_5": true, "QUEUE_10": true, "QUEUE_20": true,
	}
	if !allowed[p] {
		o.lastErr = fmt.Errorf("priceMatch 值不合法: %s", p)
		return o
	}
	o.httpRequest.params["priceMatch"] = p
	return o
}

// SelfTradePreventionMode 设置自成交保护模式
func (o *Order) SelfTradePreventionMode(mode string) *Order {

	if o.lastErr != nil {
		return o
	}
	if mode == "" {
		return o
	}
	m := strings.ToUpper(strings.TrimSpace(mode))
	if m != "EXPIRE_TAKER" && m != "EXPIRE_MAKER" && m != "EXPIRE_BOTH" {
		o.lastErr = fmt.Errorf("selfTradePreventionMode 值不合法: %s", m)
		return o
	}
	o.httpRequest.params["selfTradePreventionMode"] = m
	return o
}

// GoodTillDate 设置 GTD 的自动取消时间（毫秒），仅在 timeInForce 为 GTD 时使用
func (o *Order) GoodTillDate(ts int64) *Order {

	if o.lastErr != nil {
		return o
	}
	if ts <= 0 {
		o.lastErr = fmt.Errorf("goodTillDate 必须为正整数（毫秒时间戳）")
		return o
	}
	nowMs := time.Now().UnixMilli()
	if ts <= nowMs+600*1000 {
		o.lastErr = fmt.Errorf("goodTillDate 必须大于当前时间 + 600s")
		return o
	}
	if ts >= int64(253402300799000) {
		o.lastErr = fmt.Errorf("goodTillDate 超出允许的最大值")
		return o
	}
	o.httpRequest.params["goodTillDate"] = ts
	return o
}

// RecvWindow 设置请求超时时间窗口（可选）
func (o *Order) RecvWindow(rw int64) *Order {

	if o.lastErr != nil {
		return o
	}
	if rw <= 0 {
		o.httpRequest.params["recvWindow"] = rw
		return o
	}
	o.httpRequest.params["recvWindow"] = rw
	return o
}

// validate 参数校验
func (o *Order) validate() error {
	params := o.httpRequest.params
	// 如果 setter 中已记录错误，优先返回该错误
	if o.lastErr != nil {
		return o.lastErr
	}

	// symbol 必填
	symbol, ok := params["symbol"].(string)
	if !ok || strings.TrimSpace(symbol) == "" {
		return fmt.Errorf("symbol 不能为空")
	}

	// side 必填
	side, ok := params["side"].(string)
	if !ok || (side != "BUY" && side != "SELL") {
		return fmt.Errorf("side 必须为 BUY 或 SELL")
	}

	// type 必填
	typ, ok := params["type"].(string)
	if !ok || typ == "" {
		return fmt.Errorf("type 不能为空")
	}

	// priceMatch 与 price 不能同时传（互斥）
	if _, hasPrice := params["price"]; hasPrice {
		if _, hasPM := params["priceMatch"]; hasPM {
			return fmt.Errorf("priceMatch 不能与 price 同时传")
		}
	}

	// 根据 type 校验强制参数（部分校验已在 setter 执行，这里做组合校验）
	switch typ {
	case "LIMIT":
		if _, ok := params["timeInForce"].(string); !ok {
			params["timeInForce"] = "GTC"
		}
		if _, ok := params["quantity"]; !ok {
			return fmt.Errorf("LIMIT 类型必须指定 quantity")
		}
		if _, ok := params["price"]; !ok {
			return fmt.Errorf("LIMIT 类型必须指定 price")
		}
	case "MARKET":
		if _, ok := params["quantity"]; !ok {
			return fmt.Errorf("MARKET 类型必须指定 quantity")
		}
	}

	// 当 timeInForce 为 GTD 时，goodTillDate 必填（setter 已校验具体值）
	if tif, ok := params["timeInForce"].(string); ok && tif == "GTD" {
		if _, ok := params["goodTillDate"].(int64); !ok {
			return fmt.Errorf("timeInForce 为 GTD 时必须传 goodTillDate")
		}
	}

	return nil
}

// Do 发送下单请求
func (o *Order) Do(ctx context.Context) (*OrderResponse, error) {
	// 参数校验
	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("下单参数校验失败: %w", err)
	}

	res := new(OrderResponse)
	if err := o.httpRequest.PostJSON(ctx, res); err != nil {
		return nil, fmt.Errorf("下单失败: %w", err)
	}
	return res, nil
}
