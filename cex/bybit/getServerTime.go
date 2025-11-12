package bybit

import "context"

type Time struct {
	httpRequest *HttpRequest
}

// GetServerTime 创建服务器时间接口
func (b *Bybit) NewGetServerTime() *Time {
	return &Time{
		httpRequest: &HttpRequest{
			bybit:       b,                            // Bybit客户端实例
			baseUrl:     b.UrlRest,                    // 基础URL
			apiUrl:      "/v5/market/time",            // 请求URL
			sign:        false,                        // 是否签名
			isTimestamp: false,                        // 是否时间戳
			params:      make(map[string]interface{}), // 请求参数
		},
	}
}

func (t *Time) Do(ctx context.Context) (int64, error) {

	var res = new(struct {
		Result struct {
			TimeSecond int64 `json:"timeSecond"` // 服务器时间(秒)
			TimeNano   int64 `json:"timeNano"`   // 服务器时间(纳秒)
		} `json:"result"`
	})

	err := t.httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return 0, err
	}
	return res.Result.TimeSecond, nil
}
