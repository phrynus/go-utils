package user

import (
	"context"
	"errors"
	"strconv"
)

// CloudFunction 构建并执行云函数 API 请求
type CloudFunction struct {
	client *Client
	req    CloudFunctionRequest
}

// NewCloudFunction 云函数
func (c *Client) NewCloudFunction() *CloudFunction {
	return &CloudFunction{client: c}
}

// Name 设置云函数名称
func (c *CloudFunction) Name(name string) *CloudFunction {
	c.req.Name = name
	return c
}

// Param 设置云函数参数
func (c *CloudFunction) Param(param string) *CloudFunction {
	c.req.Param = param
	return c
}

// Do 发送请求，可选择性地覆盖 context
func (c *CloudFunction) Do(ctx ...context.Context) (bool, error) {
	if c.client == nil {
		return false, errNilClient
	}
	// token 是可选的，如果有则从 client 获取
	if token, err := c.client.GetToken(); err == nil {
		c.req.Token = token
	}
	if c.req.Name == "" {
		return false, errors.New("云函数名称是必需的")
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
	res, err := c.client.SecurePost(callCtx, "cloudFunction", c.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
