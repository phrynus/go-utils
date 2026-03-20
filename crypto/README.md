# crypto - 加密工具集

Go 常用加密工具集，涵盖对称加密、非对称加密、密码哈希、散列及 ID 生成，开箱即用，无需额外配置。

## 功能特性

- ✅ AES-128/192/256 对称加密（CBC 模式，PKCS7 填充）
- ✅ DES 对称加密（CBC 模式，PKCS7 填充）
- ✅ RC4 流加密
- ✅ RSA 非对称加密（PKCS#1 v1.5，支持自动分块）
- ✅ bcrypt 密码哈希
- ✅ MD5 散列
- ✅ UUID v7 生成（时间有序）
- ✅ Base64 / Hex 统一编码解码

## 安装

```bash
go get github.com/phrynus/go-utils/crypto
```

## 快速开始

### 对称加密：AES / DES

AES 和 DES 接口完全一致，仅密钥长度不同：

```go
import "github.com/phrynus/go-utils/crypto"

// AES 加密（AES-128: 16字节密钥, AES-192: 24字节, AES-256: 32字节）
ciphertext, err := crypto.AesEncrypt("1234567890abcdef", "hello world", "base64")
plaintext, err := crypto.AesDecrypt("1234567890abcdef", ciphertext, "base64")

// DES 加密（8字节密钥）
ciphertext, err := crypto.DesEncrypt("12345678", "hello", "hex")
plaintext, err := crypto.DesDecrypt("12345678", ciphertext, "hex")
```

`mode` 参数支持 `"base64"`（默认）和 `"hex"`，加密解密须保持一致。

> ⚠️ DES 因密钥长度过短已不安全，仅用于兼容遗留系统。

### 流加密：RC4

```go
// 加密
ciphertext, err := crypto.Rc4Encrypt("key", "plaintext", "base64")

// 解密
plaintext, err := crypto.Rc4Decrypt("key", ciphertext, "base64")
```

> ⚠️ RC4 存在偏差攻击等安全弱点，谨慎用于新系统。

### 非对称加密：RSA

#### 生成密钥对（OpenSSL）

```bash
# 生成私钥
openssl genrsa -out private.pem 2048

# 导出公钥
openssl rsa -in private.pem -pubout -out public.pem
```

#### 加密与解密

```go
// 使用公钥加密
ciphertext, err := crypto.RSAEncrypt(publicKeyPEM, "data")

// 使用私钥解密
plaintext, err := crypto.RSADecrypt(privateKeyPEM, ciphertext)
```

- 自动分块处理变长数据，无需手动管理 padding
- 公钥 PEM 类型为 `PUBLIC KEY`，私钥为 `RSA PRIVATE KEY`
- 输出为 Base64 编码字符串

### 密码哈希：bcrypt

```go
// 生成哈希
hash, err := crypto.HashPassword("mypassword")

// 验证密码
err := crypto.VerifyPassword("mypassword", hash)
// err == nil 表示验证通过
```

内部使用 `bcrypt.DefaultCost`，每次生成的哈希随机不同（含 salt），必须用 `VerifyPassword` 验证，不能直接字符串比较。

### 散列：MD5

```go
// 返回 32 位十六进制小写字符串
h := crypto.MD5("hello")
// h := "5d41402abc4b2a76b9719d911017c592"
```

> ⚠️ MD5 存在碰撞攻击风险，不适用于安全敏感场景，仅推荐用于数据完整性校验。

### ID 生成：UUID v7

```go
// 生成时间有序 UUID，适合作为数据库主键或日志关联 ID
id := crypto.UUID()
// id := "0191c7e4-5b4a-7c8a-9b3e-2f1d0c9a8b7e"
```

## 完整示例

完整示例请查看 [example/main.go](../example/main.go)

## API 参考

### 公开类型

| 类型 | 说明 |
|------|------|
| 无外部导出类型 | 所有方法均为包级函数，直接调用 |

### 方法

| 方法 | 说明 |
|------|------|
| `AesEncrypt(key, data, mode string) (string, error)` | AES CBC 加密，mode 支持 `"base64"` / `"hex"` |
| `AesDecrypt(key, data, mode string) (string, error)` | AES CBC 解密 |
| `DesEncrypt(key, data, mode string) (string, error)` | DES CBC 加密 |
| `DesDecrypt(key, data, mode string) (string, error)` | DES CBC 解密 |
| `Rc4Encrypt(key, data, mode string) (string, error)` | RC4 流加密 |
| `Rc4Decrypt(key, data, mode string) (string, error)` | RC4 流解密 |
| `RSAEncrypt(publicKey, data string) (string, error)` | RSA PKCS#1 v1.5 加密（公钥 Base64 输出） |
| `RSADecrypt(privateKey, data string) (string, error)` | RSA PKCS#1 v1.5 解密 |
| `HashPassword(password string) (string, error)` | bcrypt 密码哈希 |
| `VerifyPassword(password, hash string) error` | bcrypt 密码验证 |
| `MD5(s string) string` | MD5 散列（十六进制小写） |
| `UUID() string` | UUID v7 生成 |

## 注意事项

- AES / DES 密钥长度必须严格匹配，否则解密会报错
- RSA 加密数据量受密钥长度限制（2048位密钥每次最多加密 245 字节），工具内部自动分块处理
- bcrypt 每次生成的哈希值不同，验证必须用 `VerifyPassword`，不可字符串相等判断
- UUID v7 基于时间戳生成，适合需要时间有序 ID 的场景
