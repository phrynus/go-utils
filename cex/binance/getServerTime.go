package binance

import "context"

type Time struct {
	httpRequest *HttpRequest
}

// GetServerTime 创建服务器时间接口
func (b *Binance) NewGetServerTime() *Time {
	return &Time{
		httpRequest: &HttpRequest{
			binance:     b,                            // 币安客户端实例
			baseUrl:     b.UrlRest,                    // 基础URL
			apiUrl:      "/fapi/v1/time",              // 请求URL
			sign:        false,                        // 是否签名
			isTimestamp: false,                        // 是否时间戳
			params:      make(map[string]interface{}), // 请求参数
		},
	}
}

func (t *Time) Do(ctx context.Context) (int64, error) {
	var res = new(struct {
		ServerTime int64 `json:"serverTime"` // 服务器时间
	})

	err := t.httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return 0, err
	}
	return res.ServerTime, nil
}
