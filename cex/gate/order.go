// Package gate 提供 Gate 交易所 REST/WebSocket 客户端的封装与辅助工具。
// 本文件实现了 Gate 永续合约的下单请求构建器和响应结构，便于以链式方法构造下单参数并发送请求。
// Order 提供一组链式 Setter 用于设置合约、数量、价格等参数，最终通过 Do 方法发送下单请求并返回响应。
package gate

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Order 是 Gate 永续合约的下单请求构建器。
type Order struct {
	gate        *Gate
	httpRequest *HttpRequest
	lastErr     error
}

// OrderResponse 包含 Gate 下单返回的常用字段。
type OrderResponse struct {
	Id            int64   `json:"id,omitempty"`                 // 合约订单 ID
	User          int64   `json:"user,omitempty"`               // 用户 ID
	CreateTime    float64 `json:"create_time,omitempty"`        // 订单创建时间（number,double）
	UpdateTime    float64 `json:"update_time,omitempty"`        // 订单更新时间（number,double）
	FinishTime    float64 `json:"finish_time,omitempty"`        // 订单结束时间（number,double）
	FinishAs      string  `json:"finish_as,omitempty"`          // 结束方式
	Status        string  `json:"status,omitempty"`             // 订单状态
	Contract      string  `json:"contract,omitempty"`           // 合约标识
	Size          float64 `json:"size,string,omitempty"`        // 交易数量（API 返回 string，这里解析为 float64）
	Iceberg       float64 `json:"iceberg,string,omitempty"`     // 冰山委托显示数量
	Price         float64 `json:"price,string,omitempty"`       // 委托价
	Close         bool    `json:"close,omitempty"`              // 请求时的平仓标记（只写）
	IsClose       bool    `json:"is_close,omitempty"`           // 是否为平仓委托（只读）
	ReduceOnly    bool    `json:"reduce_only,omitempty"`        // 请求时的只减仓标记（只写）
	IsReduceOnly  bool    `json:"is_reduce_only,omitempty"`     // 是否为只减仓委托（只读）
	IsLiq         bool    `json:"is_liq,omitempty"`             // 是否为强制平仓委托（只读）
	Tif           string  `json:"tif,omitempty"`                // time in force
	Left          float64 `json:"left,string,omitempty"`        // 未成交数量
	FillPrice     float64 `json:"fill_price,string,omitempty"`  // 成交价
	Text          string  `json:"text,omitempty"`               // 用户自定义信息
	Tkfr          float64 `json:"tkfr,string,omitempty"`        // 吃单费率
	Mkfr          float64 `json:"mkfr,string,omitempty"`        // 做单费率
	Refu          int64   `json:"refu,omitempty"`               // 推荐人用户 ID
	AutoSize      string  `json:"auto_size,omitempty"`          // 双仓平仓方向
	StpId         int64   `json:"stp_id,omitempty"`             // STP 用户组 ID
	StpAct        string  `json:"stp_act,omitempty"`            // 自成交限制策略
	AmendText     string  `json:"amend_text,omitempty"`         // 修改备注
	Pid           int64   `json:"pid,omitempty"`                // 仓位 ID
	OrderValue    float64 `json:"order_value,string,omitempty"` // 委托价值
	TradeValue    float64 `json:"trade_value,string,omitempty"` // 成交价值
	FinishMessage string  `json:"finish_message,omitempty"`     // 可选的结束信息（有些接口返回）
}

// NewOrder 创建下单请求，默认 settle 为 usdt
func (g *Gate) NewOrder(settle string) *Order {
	f := &Order{
		gate: g,
		httpRequest: &HttpRequest{
			gate:        g,
			baseUrl:     g.UrlRest,
			apiUrl:      "/futures/" + settle + "/orders", // 默认结算币种 usdt，可通过 SetSettle 修改
			sign:        true,
			isTimestamp: true,
			params:      make(map[string]interface{}),
		},
	}
	return f
}

// Contract 设置合约标识（必填）
func (f *Order) Contract(contract string) *Order {
	if f.lastErr != nil {
		return f
	}
	if strings.TrimSpace(contract) == "" {
		f.lastErr = fmt.Errorf("contract 不能为空")
		return f
	}
	f.httpRequest.params["contract"] = contract
	return f
}

// Size 设置合约张数（必填）: 为字符串或数值都可，库会序列化为 string 不能为小数
func (f *Order) Size(size interface{}) *Order {
	if f.lastErr != nil {
		return f
	}
	if size == nil {
		f.lastErr = fmt.Errorf("size 不能为空")
		return f
	}
	f.httpRequest.params["size"] = size
	return f
}

// Price 设置委托价格（必填，市价可传 0 并配合 tif=ioc）
func (f *Order) Price(price interface{}) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["price"] = price
	return f
}

// Iceberg 隐藏数量（可选）
func (f *Order) Iceberg(iceberg interface{}) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["iceberg"] = iceberg
	return f
}

// Close 设置为平仓（size 应该为 0）
func (f *Order) Close(close bool) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["close"] = close
	return f
}

// ReduceOnly 设置是否仅减仓
func (f *Order) ReduceOnly(reduce bool) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["reduce_only"] = reduce
	return f
}

// TIF 设置 time in force（gtc,ioc,poc,fok）
// gtc: 撤销前有效
func (f *Order) TIF(tif string) *Order {
	if f.lastErr != nil {
		return f
	}
	if tif == "" {
		return f
	}
	tf := strings.ToLower(strings.TrimSpace(tif))
	switch tf {
	case "gtc", "ioc", "poc", "fok":
		f.httpRequest.params["tif"] = tf
	default:
		f.lastErr = fmt.Errorf("tif 值不合法: %s", tif)
	}
	return f
}

// Text 用户自定义信息（必须以 t- 开头）
func (f *Order) Text(text string) *Order {
	if f.lastErr != nil {
		return f
	}
	if text == "" {
		return f
	}
	// 简单校验：以 t- 开头且长度限制（不计 t- 不超过 28 字节）
	if !strings.HasPrefix(text, "t-") {
		f.lastErr = fmt.Errorf("text 必须以 t- 开头")
		return f
	}
	if len(text)-2 > 28 {
		f.lastErr = fmt.Errorf("text 长度不合法（超过 28 字节，去掉 t- 后计数）")
		return f
	}
	f.httpRequest.params["text"] = text
	return f
}

// AutoSize 设置双仓模式平仓方向（close_long/close_short）
func (f *Order) AutoSize(auto string) *Order {
	if f.lastErr != nil {
		return f
	}
	if auto == "" {
		return f
	}
	a := strings.ToLower(strings.TrimSpace(auto))
	if a != "close_long" && a != "close_short" {
		f.lastErr = fmt.Errorf("auto_size 值不合法: %s", auto)
		return f
	}
	f.httpRequest.params["auto_size"] = a
	return f
}

// StpAct 自成交限制（co/cn/cb/-）
func (f *Order) StpAct(act string) *Order {
	if f.lastErr != nil {
		return f
	}
	if act == "" {
		return f
	}
	a := strings.ToLower(strings.TrimSpace(act))
	switch a {
	case "co", "cn", "cb", "-":
		f.httpRequest.params["stp_act"] = a
	default:
		f.lastErr = fmt.Errorf("stp_act 值不合法: %s", act)
	}
	return f
}

// PID 设置仓位ID（可选）
func (f *Order) PID(pid int64) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["pid"] = pid
	return f
}

// OrderValue 设置委托价值（可选）
func (f *Order) OrderValue(v interface{}) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["order_value"] = v
	return f
}

// TradeValue 设置成交价值（可选）
func (f *Order) TradeValue(v interface{}) *Order {
	if f.lastErr != nil {
		return f
	}
	f.httpRequest.params["trade_value"] = v
	return f
}

// validate 参数校验
func (f *Order) validate() error {
	if f.lastErr != nil {
		return f.lastErr
	}
	params := f.httpRequest.params
	// contract 必填
	if c, ok := params["contract"].(string); !ok || strings.TrimSpace(c) == "" {
		return fmt.Errorf("contract 不能为空")
	}
	// size 必填（可以为 0，当 close 为 true 时）
	if _, ok := params["size"]; !ok {
		return fmt.Errorf("size 不能为空")
	}
	// price 必填（price 为 0 时需要 tif=ioc 表示市价立即成交）
	if _, ok := params["price"]; !ok {
		return fmt.Errorf("price 不能为空")
	}
	// tif 校验（若存在）
	if tif, ok := params["tif"].(string); ok {
		t := strings.ToLower(strings.TrimSpace(tif))
		if t != "gtc" && t != "ioc" && t != "poc" && t != "fok" {
			return fmt.Errorf("tif 值不合法: %s", tif)
		}
		// 如果 price == 0, tif 必须为 ioc（市价）
		if p, ok := params["price"]; ok {
			if str, ok2 := p.(string); ok2 && (str == "0" || str == "0.0") && t != "ioc" {
				return fmt.Errorf("price 为 0 时，tif 必须为 ioc")
			}
		}
	}

	// close=true 时 size 应为 0
	if closeFlag, ok := params["close"].(bool); ok && closeFlag {
		// 检查 size 是否为 0（接受数字或字符串）
		if s, ok := params["size"]; ok {
			switch v := s.(type) {
			case int:
				if v != 0 {
					return fmt.Errorf("close 为 true 时，size 必须为 0")
				}
			case int64:
				if v != 0 {
					return fmt.Errorf("close 为 true 时，size 必须为 0")
				}
			case float64:
				if v != 0 {
					return fmt.Errorf("close 为 true 时，size 必须为 0")
				}
			case string:
				if v != "0" && v != "0.0" {
					return fmt.Errorf("close 为 true 时，size 必须为 0")
				}
			default:
				return fmt.Errorf("无法识别的 size 类型")
			}
		}
	}

	return nil
}

// Do 发送下单请求并返回响应
func (f *Order) Do(ctx context.Context) (*OrderResponse, error) {
	// 校验参数
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("下单参数校验失败: %w", err)
	}

	// 设置请求体为 params（Gate HTTP 客户端会按需序列化）
	res := new(OrderResponse)
	if err := f.httpRequest.PostJSONWithBody(ctx, f.httpRequest.params, res); err != nil {
		return nil, fmt.Errorf("下单失败: %w", err)
	}

	// 若返回时间字段为空，尝试补充（以 float64 表示）
	if res.CreateTime == 0 {
		res.CreateTime = float64(time.Now().Unix())
	}
	return res, nil
}
