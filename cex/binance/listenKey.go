package binance

import (
	"context"
	"fmt"
	"strings"
)

// CreateListenKey 创建listenKey
func (b *Binance) CreateListenKey(ctx context.Context) (string, error) {
	req := &HttpRequest{
		binance:     b,
		baseUrl:     b.UrlRest,
		apiUrl:      "/fapi/v1/listenKey",
		sign:        false,
		isTimestamp: false,
		params:      make(map[string]interface{}),
	}

	var res struct {
		ListenKey string `json:"listenKey"`
	}
	if err := req.PostJSON(ctx, &res); err != nil {
		return "", fmt.Errorf("创建 listenKey 失败: %w", err)
	}
	if strings.TrimSpace(res.ListenKey) == "" {
		return "", fmt.Errorf("创建 listenKey 失败: 返回为空")
	}
	return res.ListenKey, nil
}

// KeepAliveListenKey 保活listenKey
func (b *Binance) KeepAliveListenKey(ctx context.Context, listenKey string) error {
	if strings.TrimSpace(listenKey) == "" {
		return fmt.Errorf("listenKey 不能为空")
	}

	req := &HttpRequest{
		binance:     b,
		baseUrl:     b.UrlRest,
		apiUrl:      "/fapi/v1/listenKey",
		sign:        false,
		isTimestamp: false,
		params:      map[string]interface{}{"listenKey": listenKey},
	}

	if _, err := req.Put(ctx); err != nil {
		return fmt.Errorf("保活 listenKey 失败: %w", err)
	}
	return nil
}

// CloseListenKey 关闭listenKey
func (b *Binance) CloseListenKey(ctx context.Context, listenKey string) error {
	if strings.TrimSpace(listenKey) == "" {
		return fmt.Errorf("listenKey 不能为空")
	}

	req := &HttpRequest{
		binance:     b,
		baseUrl:     b.UrlRest,
		apiUrl:      "/fapi/v1/listenKey",
		sign:        false,
		isTimestamp: false,
		params:      map[string]interface{}{"listenKey": listenKey},
	}

	if _, err := req.Delete(ctx); err != nil {
		return fmt.Errorf("关闭 listenKey 失败: %w", err)
	}
	return nil
}
