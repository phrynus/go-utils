package user

import (
	"context"
	"errors"
	"strconv"
)

// KamiTopup 构建并执行卡密充值 API 请求
type KamiTopup struct {
	client *Client
	req    KamiTopupRequest
}

// NewKamiTopup 卡密充值
func (c *Client) NewKamiTopup() *KamiTopup {
	return &KamiTopup{client: c}
}

// Kami 设置卡密
func (k *KamiTopup) Kami(kami string) *KamiTopup {
	k.req.Kami = kami
	return k
}

// Password 设置卡密密码（仅卡密版应用有效）
func (k *KamiTopup) Password(password string) *KamiTopup {
	k.req.Password = password
	return k
}

// Do 发送请求，可选择性地覆盖 context
func (k *KamiTopup) Do(ctx ...context.Context) (bool, error) {
	if k.client == nil {
		return false, errNilClient
	}
	token, err := k.client.GetToken()
	if err != nil {
		return false, err
	}
	k.req.Token = token
	if k.req.Kami == "" {
		return false, errors.New("卡密是必需的")
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
	res, err := k.client.SecurePost(callCtx, "kamiTopup", k.req, nil)
	if err != nil {
		return false, err
	}
	if res.Code != 0 {
		return false, errors.New(strconv.Itoa(res.Code) + ":" + res.Msg)
	}
	return true, nil
}
