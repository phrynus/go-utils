# U验证 用户API客户端 v3.3.18

U验证 用户API客户端，支持加密通信、签名验证、登录、支付等功能。

## 功能特性

- ✅ 支持多种加密模式（RSA、AES、DES、RC4、None）
- ✅ 支持完整的验签系统
- ✅ 支持自动Token管理
- ✅ 支持代理配置
- ✅ 提供链式API调用

## 安装

```bash
go get github.com/phrynus/go-utils/uyz-u
```

## 快速开始

### 创建客户端

```go
package main

import (
    "context"
    "log"
    "time"
    user "github.com/phrynus/go-utils/uyz-u"
)

const rsaClientPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
...你的私钥...
-----END RSA PRIVATE KEY-----`

const rsaServerPublicKey = `-----BEGIN PUBLIC KEY-----
...服务器公钥...
-----END PUBLIC KEY-----`

func main() {
    client, err := user.New(user.ClientConfig{
        BaseURL:          "https://uverif.xxx/api/user",
        AppID:            1003,
        AppKey:           "your_app_key",
        Version:          "1.0.0",
        VersionIndex:     "web",
        ClientPrivateKey: rsaClientPrivateKey,
        ServerPublicKey:  rsaServerPublicKey,
        HTTPTimeout:      10 * time.Second,
        EncryptionMode:   user.EncryptionRSA,
        EncodingMode:     user.EncodingBase64,
        DisableSignature: false,
    })
    if err != nil {
        log.Fatalf("create client: %v", err)
    }

    // 登录
    login, err := client.NewLogin().
        Account("username").
        Password("password").
        UDID("device-id").
        Do(context.Background())
    if err != nil {
        log.Fatalf("login failed: %v", err)
    }
    fmt.Printf("登录成功，Token: %s\n", login.Token)

    // 获取用户信息
    info, err := client.NewInfo().Do(context.Background())
    if err != nil {
        log.Fatalf("get info failed: %v", err)
    }
    fmt.Printf("用户ID: %d\n", info.UID)

    // 获取商品列表
    goods, err := client.NewGoods().Page(1).Do(context.Background())
    if err != nil {
        log.Fatalf("get goods failed: %v", err)
    }
    fmt.Printf("商品总数: %d\n", goods.DataTotal)

    // 在线支付
    pay, err := client.NewPay().
        GID(14).
        Type("ali").  // wx=微信，ali=支付宝
        Mode("qr").   // h5, app, qr
        Do(context.Background())
    if err != nil {
        log.Fatalf("pay failed: %v", err)
    }
    fmt.Printf("订单号: %s\n", pay.OrderNo)
}
```

## 配置说明

### ClientConfig 结构体

```go
type ClientConfig struct {
    BaseURL          string         // 例如: https://uverif.xxx/api/user
    AppID            int            // 应用 ID
    AppKey           string         // 用于 MD5 签名
    Version          string         // 客户端语义版本，例如: "1.0.0" (默认: "1.0.0")
    VersionIndex     string         // 例如: "web" (默认: "web")
    ClientPrivateKey string         // PEM 格式的私钥，用于解密 payload (RSA模式必需)
    ServerPublicKey  string         // PEM 格式的公钥，用于加密 payload (RSA模式必需)
    HTTPTimeout      time.Duration  // HTTP超时时间 (默认: 10秒)
    EncryptionMode   EncryptionMode // AES/DES/RC4/RSA/none (默认: none)
    EncodingMode     EncodingMode   // 对称模式的编码方式：base64 或 hex (默认: base64)
    SymmetricKey     string         // AES/DES/RC4 的共享密钥 (对称加密模式必需)
    DisableSignature bool           // 为 true 时，省略 MD5 签名 (默认: false)
    ProxyURL         string         // 代理 URL 也可使用 utils.GetProxy() 获取代理URL
}
```

### 加密模式示例

#### RSA 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://uverif.xxx/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    ClientPrivateKey: rsaClientPrivateKey,
    ServerPublicKey:  rsaServerPublicKey,
    EncryptionMode:   user.EncryptionRSA,
})
```

#### AES/DES/RC4 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:        "https://uverif.xxx/api/user",
    AppID:          1003,
    AppKey:         "your_app_key",
    EncryptionMode: user.EncryptionAES,  // 或 EncryptionDES, EncryptionRC4
    EncodingMode:   user.EncodingBase64,  // 或 EncodingHex
    SymmetricKey:   "your_symmetric_key",
})
```

#### 无加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:        "https://uverif.xxx/api/user",
    AppID:          1003,
    AppKey:         "your_app_key",
    EncryptionMode: user.EncryptionNone,
})
```

## API方法

### 用户认证

- `client.NewLogin().Account().Password().UDID().Do()` - 登录
- `client.NewRegister().Account().Password().UDID().InvID().Code().Do()` - 注册
- `client.NewLogout().Do()` - 登出
- `client.NewSignIn().Do()` - 签到

### QQ登录

- `client.NewQQLoginQuery().UUID().Do()` - QQ登录状态查询
- `client.NewQQLoginWeb().UDID().InvID().Do()` - QQ网页登录
- `client.NewQQLoginSDK().AccessToken().OpenID().UDID().InvID().Do()` - QQ SDK登录
- `client.NewQQBindSDK().AccessToken().OpenID().Do()` - QQ SDK绑定

### 微信登录

- `client.NewWXLoginQuery().UUID().Do()` - 微信登录状态查询
- `client.NewWXLoginWeb().UDID().InvID().Do()` - 微信登录
- `client.NewWXLoginSDK().AccessToken().OpenID().UDID().InvID().Do()` - 微信SDK登录
- `client.NewWXBindSDK().AccessToken().OpenID().Do()` - 微信SDK绑定

### 用户信息

- `client.NewInfo().Do()` - 获取用户信息
- `client.NewModifyPwd().Password().Do()` - 修改密码
- `client.NewResetPwd().Account().Password().Code().Do()` - 重置密码
- `client.NewSetAcctno().Acctno().Do()` - 设置账号
- `client.NewModifyPic().File().Do()` - 上传头像

### 账号绑定

- `client.NewSetEmail().Email().Code().Do()` - 绑定邮箱
- `client.NewSetPhone().Phone().Code().Do()` - 绑定手机
- `client.NewRmEmail().Email().Code().Do()` - 解绑邮箱
- `client.NewRmPhone().Phone().Code().Do()` - 解绑手机

### 设备管理

- `client.NewBindUDID().UDID().Do()` - 绑定设备
- `client.NewRmUDID().UDID().Do()` - 解绑设备
- `client.NewGetUDID().Do()` - 获取已绑定设备列表

### 商品和支付

- `client.NewGoods().Page().Do()` - 获取商品列表
- `client.NewPay().GID().Type().Mode().Do()` - 在线支付
- `client.NewKamiTopup().Kami().Password().Do()` - 卡密充值
- `client.NewOrderList().Page().Do()` - 订单列表
- `client.NewOrderQuery().Order().Do()` - 订单查询

### 消息管理

- `client.NewMessageAdd().Title().Content().File().Do()` - 新增留言
- `client.NewMessageList().Page().Do()` - 留言列表
- `client.NewMessageContent().MID().Do()` - 获取留言内容
- `client.NewMessageReply().MID().Content().File().Do()` - 回复留言
- `client.NewMessageEnd().MID().Do()` - 结束留言

### 其他功能

- `client.NewVIP().Do()` - VIP验证
- `client.NewFen().FenID().FenMark().Do()` - 积分验证
- `client.NewGetCode().Account().Type().Do()` - 获取验证码
- `client.NewSetExtend().Key().Value().Do()` - 设置扩展信息
- `client.NewUpload().File().Do()` - 上传文件
- `client.NewCloudFunction().Name().Param().Do()` - 云函数
- `client.NewGetConfig().Do()` - 获取配置
- `client.NewHeartbeat().Do()` - 心跳
- `client.NewBan().Second().Message().Do()` - 账户禁用

## Token 管理

客户端会自动管理 Token：

- 登录成功后，Token 会自动保存
- 需要 Token 的接口会自动使用保存的 Token
- 手动管理：`client.SetToken("token")`、`client.GetToken()`、`client.ClearToken()`

## 注意事项

1. **加密模式**：根据服务器要求选择合适的加密模式。RSA 需要公钥和私钥，AES/DES/RC4 需要共享密钥。

2. **签名验证**：默认启用 MD5 签名验证，如果服务器不需要签名，可以设置 `DisableSignature: true`。

3. **Token 管理**：登录成功后 Token 会自动保存，后续请求会自动使用。

4. **错误处理**：所有 API 方法都可能返回错误，请务必检查错误。

5. **Context**：所有 API 方法都支持传入 `context.Context`，用于控制请求超时和取消。

6. **超时设置**：默认 HTTP 超时时间为 10 秒，可以通过 `HTTPTimeout` 配置项修改。
