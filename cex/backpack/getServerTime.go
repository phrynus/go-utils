package backpack

import (
	"context"

	"github.com/phrynus/go-utils"
)

// GetServerTimeRequest 获取服务器时间请求
type GetServerTimeRequest struct {
	request *HttpRequest
}

// NewGetServerTime 创建获取服务器时间请求
func (b *Backpack) NewGetServerTime() *GetServerTimeRequest {
	return &GetServerTimeRequest{
		request: &HttpRequest{
			backpack: b,
			baseUrl:  b.UrlRest,
			apiUrl:   "/api/v1/time",               // 请求URL
			sign:     false,                        // 是否需要签名
			params:   make(map[string]interface{}), // 请求参数
			window:   5000,                         // 请求有效时间窗口
		},
	}
}

// Do 执行请求
func (r *GetServerTimeRequest) Do(ctx context.Context) (int64, error) {
	res, err := r.request.Get(ctx)
	if err != nil {
		return 0, err
	}
	return utils.ToInt64(string(res)), nil
}
