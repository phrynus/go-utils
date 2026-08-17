# logger

零外部依赖、单文件复制即用的 Go 日志库。核心 `logger.go` 只用标准库，适配器按需拷贝。

## 为什么用它

- 不想 `go get` 一个带几十个间接依赖的日志库
- 需要文件轮转 + gzip 压缩 + 彩色控制台，但不想折腾配置
- 不同模块需要不同 `[TAG]` 前缀，方便 grep
- 高并发项目需要异步写入不阻塞业务
- 轮转时保证缓冲数据不丢失

## 文件清单

```
logger.go           ← 核心（必须），零外部依赖
adapter_fiber.go    ← Fiber 中间件（需要 fiber/v3）
adapter_gin.go      ← Gin 中间件（需要 gin）
adapter_gorm.go     ← GORM SQL 日志（需要 gorm）
```

## 快速开始

```go
package main

import "yourproject/logger"

func main() {
    log, err := logger.New(logger.Config{
        Filename: "app.log",       // 必填
        Tag:      "APP",
        Async:    true,            // 高并发场景推荐
        Compress: true,            // 轮转后自动 gzip
    })
    if err != nil {
        panic(err)
    }
    defer log.Close()

    log.Info("服务启动")
    log.Infof("监听 :%d", 8080)
    log.Debugf("配置项: %+v", cfg)
    log.Warn("响应时间偏长")
    log.Errorf("连接失败: %v", err)
}
```

运行本项目自带的 demo（日志在 `logs/app.log`）：

```bash
go run ./demo/
```

测试套件（项目内 `logger_test/` 目录，不自动删除）：

```bash
go run ./cmd/test/
```

## 配置项

| 字段          | 类型     | 默认              | 说明                                                        |
| ------------- | -------- | ----------------- | ----------------------------------------------------------- |
| `Filename`    | `string` | **必填**          | 日志文件路径，目录自动创建                                  |
| `MaxSize`     | `int64`  | `52428800` (50MB) | 单文件最大字节数，超过自动轮转                              |
| `MinLevel`    | `Level`  | `DEBUG`           | 低于此级别的日志不记录                                      |
| `StdoutLevel` | `Level`  | `INFO`            | 同时输出到控制台的最低级别，`NONE` 关闭                     |
| `Color`       | `bool`   | `true`            | 控制台 ANSI 彩色（DEBUG 蓝 / INFO 绿 / WARN 橙 / ERROR 红） |
| `FileLine`    | `bool`   | `true`            | 记录调用位置 `文件:行号`                                    |
| `Tag`         | `string` | `"APP"`           | 日志标识，出现在每行的 `[TAG]` 中                           |
| `Async`       | `bool`   | `false`           | `true` 启用异步写入                                         |
| `BufferSize`  | `int`    | `4096`            | 异步 channel 容量                                           |
| `Compress`    | `bool`   | `false`           | `true` 轮转后自动 gzip 压缩旧文件                           |

## API

### 创建与关闭

```go
log, err := logger.New(cfg)   // 创建
log.Close()                   // 关闭（刷新缓冲 → 关闭文件）
log.Flush()                   // 强制刷新缓冲到磁盘
```

### 日志方法

```go
log.Debug(args...)    log.Debugf(format, args...)
log.Info(args...)     log.Infof(format, args...)
log.Warn(args...)     log.Warnf(format, args...)
log.Error(args...)    log.Errorf(format, args...)
```

### 子日志器

```go
db  := log.Sub("DB")    // 共享文件 + 通道，仅 tag 不同
api := log.Sub("API")

db.Info("连接成功")       // [DB][...][INFO] 连接成功
api.Infof("端口 %d", 80)  // [API][...][INFO] 端口 80

db.Close()  // 空操作，不关文件
api.Close() // 空操作
log.Close() // 父日志器关闭才真正关文件
```

### 日志级别常量

| 常量           | 值  | 用途                           |
| -------------- | --- | ------------------------------ |
| `logger.DEBUG` | `0` | 调试，生产可关                 |
| `logger.INFO`  | `1` | 正常流程                       |
| `logger.WARN`  | `2` | 警告                           |
| `logger.ERROR` | `3` | 错误                           |
| `logger.NONE`  | `4` | 关闭输出（用于 `StdoutLevel`） |

## 日志格式

**文件输出：**

```
[TAG][2006/01/02 15:04:05.000][LEVEL] 文件:行号 消息体
```

示例：

```
[APP][2026/06/04 14:12:09.004][INFO] main.go:32 服务启动
[DB][2026/06/04 14:12:09.004][WARN] repo.go:56 慢查询 [1234.500ms]: SELECT ...
[API][2026/06/04 14:12:09.004][ERROR] handler.go:37 请求超时: context deadline exceeded
```

**控制台输出**（ANSI 彩色，可通过 `Color: false` 关闭）：

```
14:12:09.004 [INFO] main.go:32 服务启动
14:12:09.004 [WARN] repo.go:56 慢查询 [1234.500ms]: SELECT ...
14:12:09.004 [ERROR] handler.go:37 请求超时: context deadline exceeded
```

颜色：DEBUG 蓝底 / INFO 绿底 / WARN 橙底 / ERROR 红底 / 时间戳灰色。

## 文件轮转与压缩

- 当日志文件大小 ≥ `MaxSize` 时自动触发轮转
- 当前文件重命名为 `文件名.20060102.150405.扩展名`
- 若 `Compress: true`，异步 gzip 压缩旧文件为 `.gz` 并删除原文件
- 轮转前防御性检查缓冲 + `file.Sync()` 强制刷盘，**保证不丢数据**

## 适配器

### Fiber（需复制 `adapter_fiber.go`）

```go
log, _ := logger.New(logger.Config{Filename: "app.log"})
defer log.Close()

app := fiber.New()
app.Use(log.FiberLogger())    // 请求日志（按状态码自动选级别）
app.Use(log.FiberRecovery())  // panic 恢复
```

### Gin（需复制 `adapter_gin.go`）

```go
r := gin.New()
r.Use(log.GinLogger())
r.Use(log.GinRecovery())
```

### GORM（需复制 `adapter_gorm.go`）

```go
db, err := gorm.Open(dialector, &gorm.Config{
    Logger: log.NewGormLogger(),
})
// 自动记录 SQL 执行时间、慢查询、错误
```

## 设计原则

- **零依赖核心** — `logger.go` 只用标准库，ANSI 颜色手写，无 `fatih/color`
- **按需复制** — 用哪个框架拷哪个适配器，不用的不拷
- **无全局状态** — 没有包级 `log.Info()`，手动 `New()` 管理实例
- **同步/异步可选** — 小项目同步直写，高并发异步不阻塞
- **轮转不丢数据** — `rotate()` 防御性检查缓冲 + `file.Sync()` 强制刷盘
- **配置自包含** — 不依赖外部 config 结构体，一个 `Config` 搞定
