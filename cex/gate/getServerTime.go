package gate

import "context"

type Time struct {
	httpRequest *HttpRequest
}

// GetServerTime 创建服务器时间接口
func (g *Gate) NewGetServerTime() *Time {
	return &Time{
		httpRequest: &HttpRequest{
			gate:        g,                            // Gate客户端实例
			baseUrl:     g.UrlRest,                    // 基础URL
			apiUrl:      "/spot/time",                 // 请求URL
			sign:        false,                        // 是否签名
			isTimestamp: false,                        // 是否时间戳
			params:      make(map[string]interface{}), // 请求参数
		},
	}
}

func (t *Time) Do(ctx context.Context) (int64, error) {
	var res = new(struct {
		ServerTime int64 `json:"server_time"` // 服务器时间(秒)
	})

	err := t.httpRequest.GetJSON(ctx, &res)
	if err != nil {
		return 0, err
	}
	return res.ServerTime, nil
}
