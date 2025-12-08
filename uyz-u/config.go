package user

import (
	"context"
)

// GetConfig 构建并执行获取配置 API 请求
type GetConfig struct {
	client *Client
}

// ConfigData 配置数据
type ConfigData struct {
	Extend map[string]any `json:"extend"` // 扩展信息
	Notice *struct {
		ID      int    `json:"id"`      // 公告ID
		Visit   int    `json:"visit"`   // 访问次数
		Content string `json:"content"` // 公告内容
		Time    int64  `json:"time"`    // 发布时间
	} `json:"notice"` // 公告信息
	Version *struct {
		Current string `json:"current"` // 当前版本
		Latest  string `json:"latest"`  // 最新版本
		Content string `json:"content"` // 更新内容
		URL     string `json:"url"`     // 下载地址
		Number  int    `json:"number"`  // 版本号
	} `json:"version"` // 版本信息
}

// NewGetConfig 获取配置
func (c *Client) NewGetConfig() *GetConfig {
	return &GetConfig{client: c}
}

// Do 发送请求，可选择性地覆盖 context
func (g *GetConfig) Do(ctx ...context.Context) (ConfigData, error) {
	if g.client == nil {
		return ConfigData{}, errNilClient
	}
	var callCtx context.Context
	for _, candidate := range ctx {
		if candidate != nil {
			callCtx = candidate
			break
		}
	}
	if callCtx == nil {
		callCtx = context.Background()
	}
	var payload ConfigData
	if err := g.client.RawGet(callCtx, "ini", &payload); err != nil {
		return ConfigData{}, err
	}
	return payload, nil
}
