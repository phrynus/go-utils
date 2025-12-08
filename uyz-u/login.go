package user

import (
	"context"
	"errors"
)

var errNilClient = errors.New("客户端为空")

// Login 构建并执行登录 API 请求
type Login struct {
	client *Client
	req    LoginRequest
}

// LoginRequest 携带登录端点的凭证信息
type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password,omitempty"`
	UDID     string `json:"udid"`
	Time     int64  `json:"time"`
}

// LoginData 登录数据
type LoginData struct {
	Token string   `json:"token"` // 令牌
	State string   `json:"state"` // 状态
	Info  UserInfo `json:"info"`  // 用户信息
}

// UserInfo 用户信息
type UserInfo struct {
	AcctNo     string                 `json:"acctno"`     // 账号
	Email      string                 `json:"email"`      // 邮箱
	Extend     map[string]interface{} `json:"extend"`     // 扩展信息
	Fen        int                    `json:"fen"`        // 积分
	InvCode    string                 `json:"invCode"`    // 邀请码
	InvID      int                    `json:"invID"`      // 邀请ID
	Name       string                 `json:"name"`       // 昵称
	OpenQQ     string                 `json:"openQQ"`     // QQ号
	OpenWx     string                 `json:"openWx"`     // 微信号
	Phone      string                 `json:"phone"`      // 手机号
	Pic        string                 `json:"pic"`        // 头像
	UID        int                    `json:"uid"`        // 用户ID
	VipExpDate string                 `json:"vipExpDate"` // VIP到期日期
	VipExpTime int                    `json:"vipExpTime"` // VIP到期时间戳
}

// NewLogin 登录
func (c *Client) NewLogin() *Login {
	return &Login{client: c}
}

// Account 设置账户标识符
func (l *Login) Account(account string) *Login {
	l.req.Account = account
	return l
}

// Password 设置可选的密码字段
func (l *Login) Password(password string) *Login {
	l.req.Password = password
	return l
}

// UDID 设置设备标识符
func (l *Login) UDID(udid string) *Login {
	l.req.UDID = udid
	return l
}

// Do 发送请求，可选择性地覆盖 context
func (l *Login) Do(ctx ...context.Context) (LoginData, error) {
	if l.client == nil {
		return LoginData{}, errNilClient
	}
	if l.req.Account == "" {
		return LoginData{}, errors.New("账户是必需的")
	}
	if l.req.UDID == "" {
		return LoginData{}, errors.New("设备标识符是必需的")
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
	var payload LoginData
	if _, err := l.client.SecurePost(callCtx, "logon", l.req, &payload); err != nil {
		return LoginData{}, err
	}
	// 保存 token 到客户端
	if payload.Token != "" {
		l.client.SetToken(payload.Token)
	}
	return payload, nil
}
