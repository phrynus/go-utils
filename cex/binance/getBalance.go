package binance

import "context"

type GetBalance struct {
	httpRequest *HttpRequest
	Balance     Balance
}
type Balance struct{}

func (b *Binance) NewGetBalance() *GetBalance {
	return &GetBalance{
		httpRequest: &HttpRequest{
			binance:     b,                            // 币安客户端实例
			baseUrl:     b.UrlRest,                    // 基础URL
			apiUrl:      "/fapi/v1/time",              // 请求URL
			sign:        true,                         // 是否签名
			isTimestamp: true,                         // 是否时间戳
			params:      make(map[string]interface{}), // 请求参数
		},
	}
}

func (t *GetBalance) Do(ctx context.Context) (*GetBalance, error) {

	return t, nil
}
