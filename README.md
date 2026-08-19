# go-utils

Go 语言常用工具库集，提供类型转换、Web 响应处理、系统信息、技术指标、消息推送等实用功能。

## 安装

```bash
go get github.com/phrynus/go-utils
```

## 模块概览

| 模块 | 路径 | 说明 |
|------|------|------|
| HTTP响应 | `res.go` | Gin统一响应 |
| 系统工具 | `system.go` | 系统信息、机器码 |
| 环境变量 | `env.go` | 环境变量加载 |
| Unknown | `unknown` | 任意类型智能转换 |
| 钉钉 | `dingtalk` | 钉钉消息推送 |
| 飞书 | `feishu` | 飞书消息推送 |
| 日志 | `log` | 高性能日志库 |
| 技术指标 | `ta` | K线技术分析指标 |
| 加密 | `crypto` | 多种加密、UUID、MD5 |
| U验证平台SDK | `uyz-u` | Uverif API客户端 |

---

## unknown - 任意类型智能转换

- ✅ 20+ 种类型转换（Bool、Int/Uint/Float 系列、Complex、Duration、Time、String、Bytes、Array、Map、Struct、指针、通道、函数）
- ✅ 智能类型推断（自动识别 string/int/float/bool/json.Number 等多种输入类型）
- ✅ 零值安全（所有转换方法均提供 Must 系列，支持默认值参数，nil 值不 panic）
- ✅ 丰富的时间解析（RFC3339、RFC1123、紧凑格式、时间戳、JSON Date 等 30+ 种格式）
- ✅ 嵌套 Map 访问（支持 `user.profile.name` 形式的点号路径）
- ✅ SmartUnmarshal（JSON 字符串或 Map 自动填充到任意目标结构体，支持多种 key 匹配策略）

👉 详细文档：[unknown/README.md](unknown/README.md)

---

## crypto - 加密工具集

- ✅ 密码哈希（bcrypt）
- ✅ UUID v7 生成
- ✅ MD5 签名

👉 详细文档：[crypto/README.md](crypto/README.md)

---

## dingtalk - 钉钉消息推送

- ✅ 文本消息
- ✅ Markdown消息
- ✅ 链接消息
- ✅ ActionCard消息（独立跳转/整体跳转）
- ✅ FeedCard消息
- ✅ @用户功能
- ✅ 签名验证

👉 详细文档：[dingtalk/README.md](dingtalk/README.md)

---

## feishu - 飞书消息推送

- ✅ 文本消息
- ✅ 富文本消息（Post）
- ✅ 图片消息
- ✅ 分享群名片
- ✅ 消息卡片（Interactive）
- ✅ 签名验证
- ✅ 富文本元素辅助函数

👉 详细文档：[feishu/README.md](feishu/README.md)

---

## log - 高性能日志库

- ✅ 多种日志级别（INFO、DEBUG、WARN、ERROR）
- ✅ 日志文件轮转（按大小）
- ✅ 自动压缩归档文件（gzip）
- ✅ 控制台彩色输出
- ✅ 显示文件名和行号
- ✅ 缓冲区提高性能
- ✅ 并发安全
- ✅ Log克隆和父子关系管理
- ✅ 主Log关闭时级联关闭所有子Log

👉 详细文档：[log/README.md](log/README.md)

---

## ta - 技术分析指标库

- ✅ 20+种技术分析指标
- ✅ 兼容 go-binance 库K线数据结构
- ✅ 自动识别多种K线数据格式（结构体/数组）
- ✅ 高性能并发处理
- ✅ 支持动态添加K线数据

**支持指标**：EMA、SMA、MACD、RSI、KDJ、BOLL、ATR、ADX、CCI、OBV、DPO、Stochastic RSI、SuperTrend、T3、VR、Williams %R、RMA、CMF、JingZheMA

👉 详细文档：[ta/README.md](ta/README.md)

---

## crypto - 加密工具集

- ✅ AES-128/192/256 对称加密（CBC 模式，PKCS7 填充）
- ✅ DES 对称加密（CBC 模式，PKCS7 填充）
- ✅ RC4 流加密
- ✅ RSA 非对称加密（PKCS#1 v1.5，支持自动分块）
- ✅ bcrypt 密码哈希
- ✅ MD5 散列
- ✅ UUID v7 生成（时间有序）
- ✅ Base64 / Hex 统一编码解码

👉 详细文档：[crypto/README.md](crypto/README.md)

## uyz-u - U验证平台 API 客户端

- ✅ 多种加密模式（RSA、AES、DES、RC4、None）
- ✅ 完整的验签系统
- ✅ 自动Token管理
- ✅ 代理配置支持
- ✅ 链式API调用

**功能模块**：用户认证、多平台登录（QQ/微信）、用户信息、账号绑定、设备管理、商品和支付、消息管理、云函数

👉 详细文档：[uyz-u/README.md](uyz-u/README.md)

---

## 许可证

[AGPL-3.0](LICENSE)
