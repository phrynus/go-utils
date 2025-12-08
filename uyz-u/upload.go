package user

import (
	"context"
	"errors"
)

// Upload 构建并执行上传文件 API 请求
type Upload struct {
	client *Client
	req    UploadRequest
}

// UploadRequest 携带上传文件端点的信息
type UploadRequest struct {
	Token string `json:"token"`
	File  string `json:"file"`
}

// UploadData 上传数据
type UploadData struct {
	URL string `json:"url"` // 文件地址
}

// NewUpload 上传文件
func (c *Client) NewUpload() *Upload {
	return &Upload{client: c}
}

// File 设置文件路径
func (u *Upload) File(file string) *Upload {
	u.req.File = file
	return u
}

// Do 发送请求，可选择性地覆盖 context
func (u *Upload) Do(ctx ...context.Context) (UploadData, error) {
	if u.client == nil {
		return UploadData{}, errNilClient
	}
	token, err := u.client.GetToken()
	if err != nil {
		return UploadData{}, err
	}
	u.req.Token = token
	if u.req.File == "" {
		return UploadData{}, errors.New("文件是必需的")
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
	var payload UploadData
	if _, err := u.client.SecurePost(callCtx, "upload", u.req, &payload); err != nil {
		return UploadData{}, err
	}
	return payload, nil
}
