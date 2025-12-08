# U验证 用户API客户端

U验证 用户API客户端，支持加密通信、签名验证、登录、支付等功能。

## 功能特性

- ✅ 支持多种加密模式（RSA、AES、DES、RC4、None）
- ✅ 支持MD5签名验证
- ✅ 支持自动Token管理
- ✅ 支持代理配置
- ✅ 提供链式API调用
- ✅ 支持用户登录、注册
- ✅ 支持商品列表、支付
- ✅ 支持用户信息管理
- ✅ 支持消息管理
- ✅ 支持设备管理

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
        BaseURL:          "https://example.com/api/user",
        AppID:            1003,
        AppKey:           "your_app_key",
        Version:          "1.0.0",
        VersionIndex:     "web",
        ClientPrivateKey: rsaClientPrivateKey,
        ServerPublicKey:  rsaServerPublicKey,
        EncryptionMode:   user.EncryptionRSA,
        DisableSignature: false,
        // EncodingMode:     user.EncodingHex, // 用于 AES/DES/RC4
        // SymmetricKey:     "your_symmetric_key", // 用于 AES/DES/RC4
        // ProxyURL:         "http://proxy.example.com:8080", // 可选代理
    })
    if err != nil {
        log.Fatalf("create client: %v", err)
    }
}
```

### 用户登录

```go
login, err := client.NewLogin().
    Account("username").
    Password("password").
    UDID("device-id").
    Do(context.Background())
if err != nil {
    log.Fatalf("login failed: %v", err)
}

// Token 会自动保存到客户端
fmt.Printf("登录成功，Token: %s\n", login.Token)
fmt.Printf("用户信息: %+v\n", login.Info)
```

### 获取用户信息

```go
info, err := client.NewInfo().Do(context.Background())
if err != nil {
    log.Fatalf("get info failed: %v", err)
}

fmt.Printf("用户ID: %d\n", info.UID)
fmt.Printf("账号: %s\n", info.AcctNo)
fmt.Printf("邮箱: %s\n", info.Email)
fmt.Printf("手机: %s\n", info.Phone)
fmt.Printf("VIP到期时间: %s\n", info.VipExpDate)
```

### 获取商品列表

```go
goods, err := client.NewGoods().Page(1).Do(context.Background())
if err != nil {
    log.Fatalf("get goods failed: %v", err)
}

fmt.Printf("商品总数: %d\n", goods.DataTotal)
fmt.Printf("当前页: %d/%d\n", goods.CurrentPage, goods.PageTotal)
for _, item := range goods.List {
    fmt.Printf("商品: %s, 价格: %.2f\n", item.Name, item.Money)
}
```

### 在线支付

```go
pay, err := client.NewPay().
    GID(14).                    // 商品ID
    Type("ali").                // 支付类型：wx=微信，ali=支付宝
    Mode("qr").                 // 支付模式：h5, app, qr
    Do(context.Background())
if err != nil {
    log.Fatalf("pay failed: %v", err)
}

fmt.Printf("订单号: %s\n", pay.OrderNo)
fmt.Printf("金额: %.2f\n", pay.Money)
fmt.Printf("支付地址: %s\n", pay.PayURL)
```

### VIP验证

```go
isVIP, err := client.NewVIP().Do(context.Background())
if err != nil {
    log.Fatalf("check VIP failed: %v", err)
}

if isVIP {
    fmt.Println("用户是VIP")
} else {
    fmt.Println("用户不是VIP")
}
```

### 获取配置

```go
config, err := client.NewGetConfig().Do(context.Background())
if err != nil {
    log.Fatalf("get config failed: %v", err)
}

if config.Notice != nil {
    fmt.Printf("公告: %s\n", config.Notice.Content)
}

if config.Version != nil {
    fmt.Printf("当前版本: %s\n", config.Version.Current)
    fmt.Printf("最新版本: %s\n", config.Version.Latest)
}
```

## 配置说明

### ClientConfig 结构体

```go
type ClientConfig struct {
    BaseURL          string         // API基础URL，例如: https://example.com/api/user
    AppID            int            // 应用ID
    AppKey           string         // 用于MD5签名的应用密钥
    Version          string         // 客户端语义版本，例如: "1.0.0"
    VersionIndex     string         // 版本索引，例如: "web"
    ClientPrivateKey string         // PEM格式的私钥，用于解密响应数据
    ServerPublicKey  string         // PEM格式的公钥，用于加密请求数据
    HTTPTimeout      time.Duration  // HTTP超时时间，默认为10秒
    EncryptionMode   EncryptionMode // 加密模式：RSA/AES/DES/RC4/None
    EncodingMode     EncodingMode   // 对称加密的编码方式：base64或hex
    SymmetricKey     string         // AES/DES/RC4的共享密钥
    DisableSignature bool           // 为true时，省略MD5签名
    ProxyURL         string         // 代理URL
}
```

### 加密模式

#### RSA 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://example.com/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    ClientPrivateKey: rsaClientPrivateKey,
    ServerPublicKey:  rsaServerPublicKey,
    EncryptionMode:   user.EncryptionRSA,
    DisableSignature: false,
})
```

#### AES 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://example.com/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    EncryptionMode:   user.EncryptionAES,
    EncodingMode:     user.EncodingBase64, // 或 user.EncodingHex
    SymmetricKey:     "your_symmetric_key",
    DisableSignature: false,
})
```

#### DES 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://example.com/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    EncryptionMode:   user.EncryptionDES,
    EncodingMode:     user.EncodingBase64,
    SymmetricKey:     "your_symmetric_key",
    DisableSignature: false,
})
```

#### RC4 加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://example.com/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    EncryptionMode:   user.EncryptionRC4,
    EncodingMode:     user.EncodingBase64,
    SymmetricKey:     "your_symmetric_key",
    DisableSignature: false,
})
```

#### 无加密

```go
client, err := user.New(user.ClientConfig{
    BaseURL:          "https://example.com/api/user",
    AppID:            1003,
    AppKey:           "your_app_key",
    EncryptionMode:   user.EncryptionNone,
    DisableSignature: false,
})
```

## API 方法

### 用户相关

#### 登录

```go
login, err := client.NewLogin().
    Account("username").
    Password("password").
    UDID("device-id").
    Do(context.Background())
```

#### 注册

```go
reg, err := client.NewReg().
    Account("username").
    Password("password").
    UDID("device-id").
    Do(context.Background())
```

#### 获取用户信息

```go
info, err := client.NewInfo().Do(context.Background())
```

#### 修改密码

```go
err := client.NewModifyPwd().
    OldPassword("old_password").
    NewPassword("new_password").
    Do(context.Background())
```

#### 重置密码

```go
err := client.NewResetPwd().
    Account("username").
    Code("verification_code").
    NewPassword("new_password").
    Do(context.Background())
```

### 商品和支付

#### 获取商品列表

```go
goods, err := client.NewGoods().Page(1).Do(context.Background())
```

#### 在线支付

```go
pay, err := client.NewPay().
    GID(14).
    Type("ali").
    Mode("qr").
    Do(context.Background())
```

#### 卡密充值

```go
err := client.NewKamiTopup().
    Kami("card_code").
    Password("card_password").
    Do(context.Background())
```

#### 订单列表

```go
orders, err := client.NewOrderList().
    Page(1).
    Do(context.Background())
```

#### 订单查询

```go
order, err := client.NewOrderQuery().
    OrderNo("order_number").
    Do(context.Background())
```

### VIP相关

#### VIP验证

```go
isVIP, err := client.NewVIP().Do(context.Background())
```

### 用户信息管理

#### 设置账号

```go
err := client.NewSetAcctno().
    AcctNo("new_account").
    Do(context.Background())
```

#### 设置邮箱

```go
err := client.NewSetEmail().
    Email("new_email@example.com").
    Code("verification_code").
    Do(context.Background())
```

#### 设置手机

```go
err := client.NewSetPhone().
    Phone("13800138000").
    Code("verification_code").
    Do(context.Background())
```

#### 移除邮箱

```go
err := client.NewRmEmail().
    Code("verification_code").
    Do(context.Background())
```

#### 移除手机

```go
err := client.NewRmPhone().
    Code("verification_code").
    Do(context.Background())
```

#### 设置扩展信息

```go
err := client.NewSetExtend().
    Key("key").
    Value("value").
    Do(context.Background())
```

#### 修改头像

```go
err := client.NewModifyPic().
    Pic("base64_encoded_image").
    Do(context.Background())
```

### 设备管理

#### 绑定设备

```go
err := client.NewBindUdid().
    UDID("device-id").
    Do(context.Background())
```

#### 移除设备

```go
err := client.NewRmUdid().
    UDID("device-id").
    Do(context.Background())
```

#### 获取设备列表

```go
udids, err := client.NewGetUdid().Do(context.Background())
```

### 消息管理

#### 添加消息

```go
err := client.NewMessageAdd().
    Title("消息标题").
    Content("消息内容").
    Do(context.Background())
```

#### 消息列表

```go
messages, err := client.NewMessageList().
    Page(1).
    Do(context.Background())
```

#### 消息详情

```go
message, err := client.NewMessageContent().
    ID(123).
    Do(context.Background())
```

#### 回复消息

```go
err := client.NewMessageReply().
    ID(123).
    Content("回复内容").
    Do(context.Background())
```

#### 结束消息

```go
err := client.NewMessageEnd().
    ID(123).
    Do(context.Background())
```

### 其他功能

#### 获取验证码

```go
err := client.NewGetCode().
    Type("email").  // 或 "phone"
    Target("email@example.com").
    Do(context.Background())
```

#### 验证积分

```go
err := client.NewFen().
    FenID(1).
    FenMark("mark").
    Do(context.Background())
```

#### 心跳

```go
err := client.NewHeartbeat().Do(context.Background())
```

#### 登出

```go
err := client.NewLogout().Do(context.Background())
```

#### 上传文件

```go
err := client.NewUpload().
    File("base64_encoded_file").
    Do(context.Background())
```

#### 云函数

```go
result, err := client.NewCloudFunction().
    Name("function_name").
    Param("function_params").
    Do(context.Background())
```

#### 获取配置

```go
config, err := client.NewGetConfig().Do(context.Background())
```

## Token 管理

客户端会自动管理 Token：

- 登录成功后，Token 会自动保存
- 需要 Token 的接口会自动使用保存的 Token
- 可以手动设置 Token：`client.SetToken("token")`
- 可以获取 Token：`token, err := client.GetToken()`
- 可以清除 Token：`client.ClearToken()`

## 完整示例

完整示例请查看 [example/uyz-u.go](../example/uyz-u.go)

```go
package main

import (
    "context"
    "fmt"
    "log"
    user "github.com/phrynus/go-utils/uyz-u"
)

func main() {
    // 创建客户端
    client, err := user.New(user.ClientConfig{
        BaseURL:          "https://example.com/api/user",
        AppID:            1003,
        AppKey:           "your_app_key",
        Version:          "1.0.0",
        VersionIndex:     "web",
        ClientPrivateKey: rsaClientPrivateKey,
        ServerPublicKey:  rsaServerPublicKey,
        EncryptionMode:   user.EncryptionRSA,
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
    fmt.Println(login)

    // 获取用户信息
    info, err := client.NewInfo().Do(context.Background())
    if err != nil {
        log.Fatalf("get info failed: %v", err)
    }
    fmt.Println(info)

    // 获取商品列表
    goods, err := client.NewGoods().Page(1).Do(context.Background())
    if err != nil {
        log.Fatalf("get goods failed: %v", err)
    }
    fmt.Println(goods)

    // 支付
    pay, err := client.NewPay().
        GID(14).
        Type("ali").
        Mode("qr").
        Do(context.Background())
    if err != nil {
        log.Fatalf("pay failed: %v", err)
    }
    fmt.Println(pay)
}
```

## API 参考

### 类型定义

- `Client`: 客户端结构体
- `ClientConfig`: 客户端配置
- `EncryptionMode`: 加密模式枚举
- `EncodingMode`: 编码模式枚举

### 加密模式常量

- `EncryptionRSA`: RSA 加密
- `EncryptionAES`: AES 加密
- `EncryptionDES`: DES 加密
- `EncryptionRC4`: RC4 加密
- `EncryptionNone`: 无加密

### 编码模式常量

- `EncodingBase64`: Base64 编码
- `EncodingHex`: Hex 编码

### 方法

- `New(cfg ClientConfig) (*Client, error)`: 创建客户端
- `SetToken(token string)`: 设置 Token
- `GetToken() (string, error)`: 获取 Token
- `ClearToken()`: 清除 Token
- `SecurePost(ctx context.Context, action string, body any, out any) (APIResponse, error)`: 加密 POST 请求
- `RawGet(ctx context.Context, action string, out any) error`: 普通 GET 请求
- `DecryptResponse(data string, out any) error`: 解密响应数据

## 注意事项

1. **加密模式**：根据服务器要求选择合适的加密模式。RSA 需要公钥和私钥，AES/DES/RC4 需要共享密钥。

2. **签名验证**：默认启用 MD5 签名验证，如果服务器不需要签名，可以设置 `DisableSignature: true`。

3. **Token 管理**：登录成功后 Token 会自动保存，后续请求会自动使用。如果需要手动管理 Token，可以使用 `SetToken()`、`GetToken()` 和 `ClearToken()` 方法。

4. **错误处理**：所有 API 方法都可能返回错误，请务必检查错误。

5. **Context**：所有 API 方法都支持传入 `context.Context`，用于控制请求超时和取消。

6. **代理配置**：如果需要使用代理，可以在配置中设置 `ProxyURL`。

7. **超时设置**：默认 HTTP 超时时间为 10 秒，可以通过 `HTTPTimeout` 配置项修改。

