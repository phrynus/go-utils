package bitget

import "context"

type Time struct {
	httpRequest *HttpRequest
}

// GetServerTime 创建服务器时间接口
func (b *Bitget) NewGetServerTime() *Time {
	return &Time{
		httpRequest: &HttpRequest{
			bitget:      b,                            // Bitget客户端实例
			baseUrl:     b.UrlRest,                    // 基础URL
			apiUrl:      "/api/v2/public/time",        // 请求URL
			sign:        false,                        // 是否签名
			isTimestamp: false,                        // 是否时间戳
			params:      make(map[string]interface{}), // 请求参数
		},
	}
}

func (t *Time) Do(ctx context.Context) (int64, error) {

	var res = new(struct {
		Data struct {
			ServerTime int64 `json:"serverTime"` // 返回码
		} `json:"data"`
	})

	err := t.httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return 0, err
	}
	return res.Data.ServerTime, nil
}
