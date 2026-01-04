# 日志记录器

功能完整的日志记录系统，支持日志轮转、压缩、彩色输出、多级别日志等功能。

## 功能特性

- ✅ 支持多种日志级别（INFO、DEBUG、WARN、ERROR）
- ✅ 支持日志文件轮转（按大小）
- ✅ 支持日志文件自动压缩（gzip）
- ✅ 支持控制台彩色输出
- ✅ 支持显示文件名和行号
- ✅ 支持控制台输出级别控制
- ✅ 使用缓冲区提高写入性能
- ✅ 支持并发安全的日志记录
- ✅ 自动刷新缓冲区
- ✅ 支持Logger克隆和父子关系管理
- ✅ 支持主Logger关闭时级联关闭所有子Logger

## 安装

```bash
go get github.com/phrynus/go-utils/logger
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/phrynus/go-utils/logger"
)

func main() {
    // 创建日志记录器
    log, err := logger.NewLogger(logger.LogConfig{
        Filename: "main.log", // 日志文件名
        LogDir:   "logs",     // 日志归档目录
        MaxSize:  50 * 1024,  // 单个日志文件最大大小（KB）
        StdoutLevels: map[int]bool{
            logger.INFO:  true,
            logger.DEBUG: false,
            logger.WARN:  true,
            logger.ERROR: true,
        },
        ColorOutput:  true,  // 控制台彩色输出
        ShowFileLine: true,  // 显示文件名和行号
    })
    if err != nil {
        panic(err)
    }
    
    // 使用 defer 确保程序退出时关闭日志
    defer func() {
        if err := log.Close(); err != nil {
            fmt.Printf("关闭日志记录器失败: %v\n", err)
        }
    }()
    
    // 使用日志
    log.Info("这是一条信息日志")
    log.Debugf("调试信息: %s", "value")
    log.Warn("警告信息")
    log.Errorf("错误信息: %s", "error details")
}
```

### 日志级别

```go
// INFO - 信息级别：用于记录正常的业务流程信息
log.Info("用户登录成功")
log.Infof("用户 %s 登录成功", username)

// DEBUG - 调试级别：用于记录调试信息，帮助开发人员排查问题
log.Debug("开始处理请求")
log.Debugf("请求参数: %+v", params)

// WARN - 警告级别：用于记录可能的问题或异常情况，但不影响系统正常运行
log.Warn("连接超时，正在重试")
log.Warnf("连接超时，已重试 %d 次", retryCount)

// ERROR - 错误级别：用于记录严重错误，会导致程序退出
log.Error("数据库连接失败")
log.Errorf("数据库连接失败: %v", err)
```

**注意**：调用 `Error()` 或 `Errorf()` 方法会导致程序退出（`os.Exit(1)`）。

### 配置说明

#### LogConfig 结构体

```go
type LogConfig struct {
    Filename     string       // 日志文件名（包含路径）
    LogDir       string       // 日志归档目录，用于存储轮转后的日志文件
    MaxSize      int          // 单个日志文件的最大大小（KB），超过后会触发日志轮转
    StdoutLevels map[int]bool // 控制哪些级别的日志需要同时输出到控制台
    ColorOutput  bool         // 是否在控制台使用彩色输出
    ShowFileLine bool         // 是否在日志中显示代码文件名和行号
}
```

#### 配置示例

```go
// 完整配置示例
config := logger.LogConfig{
    Filename: "logs/app.log",  // 日志文件路径
    LogDir:   "logs/archive",   // 归档目录
    MaxSize:  100 * 1024,       // 100MB
    StdoutLevels: map[int]bool{
        logger.INFO:  true,   // INFO 级别输出到控制台
        logger.DEBUG: false,   // DEBUG 级别不输出到控制台
        logger.WARN:  true,    // WARN 级别输出到控制台
        logger.ERROR: true,    // ERROR 级别输出到控制台
    },
    ColorOutput:  true,   // 启用彩色输出
    ShowFileLine: true,   // 显示文件名和行号
}

log, err := logger.NewLogger(config)
```

### Logger克隆和父子关系

Logger支持克隆功能，可以创建具有不同标识符的子Logger，所有Logger共享同一个异步处理系统以提高效率。

#### 克隆Logger

```go
// 创建主Logger
log, _ := logger.NewLogger(config)

// 克隆子Logger，具有不同的标识符
childLog := log.Clone("CHILD")

// 子Logger再克隆
grandChildLog := childLog.Clone("GRANDCHILD")

// 每个Logger都有独立的标识符
log.Info("主Logger消息")           // [MAIN] 主Logger消息
childLog.Info("子Logger消息")       // [CHILD] 子Logger消息
grandChildLog.Info("孙Logger消息")   // [GRANDCHILD] 孙Logger消息
```

#### 关闭行为

- **主Logger关闭**：级联关闭所有子Logger
- **子Logger关闭**：只关闭自己，不影响其他Logger

```go
// 关闭子Logger，不影响其他Logger
childLog.Close()  // 只关闭CHILD

// 关闭主Logger，自动关闭所有相关Logger
log.Close()  // 关闭MAIN、GRANDCHILD
```

#### 使用场景

- 微服务架构中不同模块使用不同的日志标识符
- 多租户应用中不同租户使用独立的日志标识符
- 复杂的应用中按功能模块划分日志来源

### 日志轮转

当日志文件大小超过 `MaxSize` 时，会自动触发日志轮转：

1. 当前日志文件会被重命名为带时间戳的归档文件
2. 归档文件会被移动到 `LogDir` 目录
3. 创建新的日志文件继续记录
4. 归档文件会自动压缩为 `.gz` 格式

示例：
- 原始文件：`main.log`
- 归档文件：`logs/main.20240101120000.log.gz`

### 日志格式

日志格式如下：

```
[PHRYNUS][2006/01/02 15:04:05.000][LEVEL] filename.go:123 message
```

- `[PHRYNUS]`: 固定前缀
- `[2006/01/02 15:04:05.000]`: 时间戳（日期 时间.毫秒）
- `[LEVEL]`: 日志级别（INFO、DEBUG、WARN、ERROR）
- `filename.go:123`: 文件名和行号（如果启用了 `ShowFileLine`）
- `message`: 日志消息

### 控制台输出

如果启用了 `ColorOutput`，控制台输出会使用不同的颜色：

- **INFO**: 绿色背景
- **DEBUG**: 蓝色背景
- **WARN**: 橙色背景
- **ERROR**: 红色背景

### 性能优化

- 使用缓冲区批量写入，减少 I/O 操作
- 自动刷新缓冲区（默认每秒刷新一次）
- 达到 1KB 阈值或错误/警告级别时立即刷新
- 支持并发安全的日志记录

### 完整示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/phrynus/go-utils/logger"
)

func main() {
    // 创建日志记录器
    log, err := logger.NewLogger(logger.LogConfig{
        Filename: "logs/app.log",
        LogDir:   "logs/archive",
        MaxSize:  10 * 1024, // 10MB
        StdoutLevels: map[int]bool{
            logger.INFO:  true,
            logger.DEBUG: true,
            logger.WARN:  true,
            logger.ERROR: true,
        },
        ColorOutput:  true,
        ShowFileLine: true,
    })
    if err != nil {
        panic(err)
    }

    defer log.Close()

    // 克隆子Logger用于不同模块
    dbLog := log.Clone("DATABASE")
    apiLog := log.Clone("API")

    // 模拟业务逻辑
    log.Info("应用程序启动")
    dbLog.Info("数据库连接初始化")
    apiLog.Info("API服务器启动")

    for i := 0; i < 5; i++ {
        log.Debugf("处理任务 %d", i)
        dbLog.Debugf("查询用户数据 %d", i)
        apiLog.Debugf("处理API请求 %d", i)
        time.Sleep(100 * time.Millisecond)
    }

    dbLog.Warn("数据库连接池使用率较高")
    apiLog.Warn("API响应时间增加")
    log.Warn("系统负载较高")

    // 子Logger可以独立关闭
    dbLog.Close()  // 只关闭数据库Logger

    apiLog.Info("API服务器继续运行")
    log.Info("应用程序核心功能正常")

    // 主Logger关闭时会自动关闭剩余的子Logger
    // log.Close() 在defer中自动调用
}
```

## API 参考

### 常量

- `logger.INFO`: 信息级别（0）
- `logger.DEBUG`: 调试级别（1）
- `logger.WARN`: 警告级别（2）
- `logger.ERROR`: 错误级别（3）

### 类型

- `LogConfig`: 日志配置结构体
- `Logger`: 日志记录器结构体

### 方法

- `NewLogger(config LogConfig) (*Logger, error)`: 创建新的日志记录器
- `Clone(phrynus string) *Logger`: 克隆日志记录器，创建具有新标识符的子Logger
- `Close() error`: 关闭日志记录器，刷新缓冲区并关闭文件
- `Info(args ...interface{})`: 记录信息级别日志
- `Debug(args ...interface{})`: 记录调试级别日志
- `Warn(args ...interface{})`: 记录警告级别日志
- `Error(args ...interface{})`: 记录错误级别日志（会导致程序退出）
- `Infof(format string, args ...interface{})`: 记录带格式的信息级别日志
- `Debugf(format string, args ...interface{})`: 记录带格式的调试级别日志
- `Warnf(format string, args ...interface{})`: 记录带格式的警告级别日志
- `Errorf(format string, args ...interface{})`: 记录带格式的错误级别日志（会导致程序退出）

## 注意事项

1. **程序退出**：调用 `Error()` 或 `Errorf()` 会导致程序立即退出（`os.Exit(1)`），请谨慎使用。

2. **资源清理**：建议使用 `defer log.Close()` 确保程序退出时正确关闭日志记录器。

3. **Logger克隆**：克隆的Logger共享异步处理系统以提高效率，但关闭行为有特殊规则。

4. **关闭行为**：
   - 主Logger关闭时会自动关闭所有子Logger
   - 子Logger关闭时不会影响其他Logger
   - 建议在应用关闭时只关闭主Logger

5. **日志轮转**：确保 `LogDir` 目录有写入权限，否则轮转可能失败。

6. **并发安全**：日志记录器是并发安全的，可以在多个 goroutine 中同时使用。

7. **性能考虑**：日志写入使用缓冲区，可能会有短暂的延迟。如果需要立即写入，可以考虑在关键位置手动触发刷新。

8. **文件大小**：`MaxSize` 的单位是 KB，例如 `50 * 1024` 表示 50MB。

