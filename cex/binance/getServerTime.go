// 获取服务器时间
package binance

import "context"

type Time struct {
	httpRequest *HttpRequest
}

// NewGetServerTime 创建服务器时间查询
func (b *Binance) NewGetServerTime() *Time {
	return &Time{
		httpRequest: &HttpRequest{
			binance:     b,
			baseUrl:     b.UrlRest,
			apiUrl:      "/fapi/v1/time",
			sign:        false,
			isTimestamp: false,
			params:      make(map[string]interface{}),
		},
	}
}

// Do 执行请求
func (t *Time) Do(ctx context.Context) (int64, error) {
	var res = new(struct {
		ServerTime int64 `json:"serverTime"`
	})

	err := t.httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return 0, err
	}
	return res.ServerTime, nil
}
