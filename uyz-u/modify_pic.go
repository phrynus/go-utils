package user

import (
	"context"
	"errors"
	"strconv"
)

// ModifyPic 构建并执行上传头像 API 请求
type ModifyPic struct {
	client *Client
	req    ModifyPicRequest
}

// ModifyPicRequest 携带上传头像端点的信息
type ModifyPicRequest struct {
	Token string `json:"token"`
	File  string `json:"file"`
}

// NewModifyPic 上传头像
func (c *Client) NewModifyPic() *ModifyPic {
	return &ModifyPic{client: c}
}

// File 设置文件路径（通过上传文件接口获取）
func (m *ModifyPic) File(file string) *ModifyPic {
	m.req.File = file
	return m
}

// Do 发送请求，可选择性地覆盖 context
func (m *ModifyPic) Do(ctx ...context.Context) (bool, error) {
	if m.client == nil {
		return false, errNilClient
	}
	token, err := m.client.GetToken()
	if err != nil {
		return false, err
	}
	m.req.Token = token
	if m.req.File == "" {
		return false, errors.New("文件是必需的")
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
	res, err := m.client.SecurePost(callCtx, "modifyPic", m.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
