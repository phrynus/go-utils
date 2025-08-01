package logger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// 日志级别常量定义
const (
	INFO  = iota // 信息级别：用于记录正常的业务流程信息
	DEBUG        // 调试级别：用于记录调试信息，帮助开发人员排查问题
	WARN         // 警告级别：用于记录可能的问题或异常情况，但不影响系统正常运行
	ERROR        // 错误级别：用于记录严重错误，会导致程序退出
)

// LogConfig 日志配置结构体
// 说明：
//
//	用于配置日志记录器的行为，包括：
//	- 日志文件的位置和大小限制
//	- 输出格式和样式
//	- 日志级别的控制
type LogConfig struct {
	Filename     string       // 日志文件名（包含路径）
	LogDir       string       // 日志归档目录，用于存储轮转后的日志文件
	MaxSize      int          // 单个日志文件的最大大小（KB），超过后会触发日志轮转
	StdoutLevels map[int]bool // 控制哪些级别的日志需要同时输出到控制台
	ColorOutput  bool         // 是否在控制台使用彩色输出
	ShowFileLine bool         // 是否在日志中显示代码文件名和行号
}

// Logger 日志记录器结构体
// 说明：
//
//	实现了一个功能完整的日志记录系统，特点：
//	1. 支持日志文件轮转和压缩
//	2. 支持多种日志级别
//	3. 支持控制台彩色输出
//	4. 使用缓冲区提高写入性能
//	5. 支持并发安全的日志记录
type Logger struct {
	config        LogConfig            // 日志配置信息
	file          *os.File             // 当前日志文件句柄
	currentSize   int64                // 当前日志文件大小
	mux           sync.Mutex           // 互斥锁，保证并发安全
	colorMap      map[int]*color.Color // 日志级别对应的颜色映射
	stdoutLevels  map[int]bool         // 控制台输出级别配置
	buffer        *bytes.Buffer        // 写入缓冲区
	flushInterval time.Duration        // 缓冲区刷新间隔
}

// 日志级别名称映射
var levelNames = []string{
	"INFO",
	"DEBUG",
	"WARN",
	"ERROR",
}

// NewLogger 创建新的日志记录器
// 说明：
//
//	根据提供的配置创建并初始化一个新的日志记录器
//	同时启动后台的缓冲区刷新守护进程
//
// 参数：
//   - config: 日志配置信息
//
// 返回值：
//   - *Logger: 日志记录器实例
//   - error: 初始化过程中的错误
//
// 示例：
//
//	config := LogConfig{
//	  Filename: "/var/log/app.log",
//	  MaxSize: 1024,  // 1MB
//	  ColorOutput: true,
//	}
//	logger, err := NewLogger(config)
func NewLogger(config LogConfig) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(config.Filename), 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(config.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	colorMap := map[int]*color.Color{
		INFO:  color.BgRGB(39, 174, 96).AddRGB(255, 255, 255),
		DEBUG: color.BgRGB(55, 66, 250).AddRGB(255, 255, 255),
		WARN:  color.BgRGB(255, 128, 0).AddRGB(255, 255, 255),
		ERROR: color.BgRGB(231, 76, 60).AddRGB(255, 255, 255),
		4:     color.RGB(99, 99, 99),
	}

	logger := &Logger{
		config:        config,
		file:          file,
		currentSize:   info.Size(),
		colorMap:      colorMap,
		stdoutLevels:  config.StdoutLevels,
		buffer:        bytes.NewBuffer(nil),
		flushInterval: time.Second,
	}

	go logger.flushDaemon()

	return logger, nil
}

// flushDaemon 日志刷新守护进程
// 说明：
//
//	定期将缓冲区中的日志内容写入文件
//	这是一个需要在后台持续运行的goroutine
func (l *Logger) flushDaemon() {
	for range time.Tick(l.flushInterval) {
		l.mux.Lock()
		l.flushLocked()
		l.mux.Unlock()
	}
}

// flushLocked 将缓冲区内容写入文件
// 说明：
//
//	在已获得锁的情况下，执行以下操作：
//	1. 检查缓冲区是否有内容
//	2. 将内容写入文件
//	3. 更新文件大小
//	4. 必要时触发日志轮转
//
// 返回值：
//   - error: 写入过程中的错误
func (l *Logger) flushLocked() error {
	if l.buffer.Len() == 0 {
		return nil
	}

	content := l.buffer.String()
	n, err := l.file.WriteString(content)
	if err != nil {
		return err
	}
	l.currentSize += int64(n)
	l.buffer.Reset()

	if l.currentSize > int64(l.config.MaxSize)*1024 {
		if err := l.rotateFileLocked(); err != nil {
			return fmt.Errorf("rotate file failed: %v", err)
		}
	}

	return nil
}

// log 核心日志记录函数
// 说明：
//
//	实现了日志记录的核心逻辑：
//	1. 格式化日志消息
//	2. 添加时间戳和级别标识
//	3. 可选添加文件名和行号
//	4. 支持控制台彩色输出
//	5. 触发缓冲区刷新
//
// 参数：
//   - level: 日志级别
//   - format: 格式化字符串
//   - args: 格式化参数
func (l *Logger) log(level int, format string, args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()

	var msg string
	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	now := time.Now()

	var fileLine string
	if l.config.ShowFileLine {
		_, file, line, ok := runtime.Caller(2) // 2表示调用栈的深度，跳过log和Info/Debug等函数
		if ok {
			file = filepath.Base(file) // 只取短文件名
			fileLine = fmt.Sprintf("%s:%d ", file, line)
		}
	}

	logEntry := fmt.Sprintf("[PHRYNUS][%s %s][%s] %s%s\n",
		now.Format("2006/01/02"),
		now.Format("15:04:05.000"),
		levelNames[level],
		fileLine,
		msg)

	l.buffer.WriteString(logEntry)

	if l.stdoutLevels[level] {
		consoleOutput := logEntry
		if l.config.ColorOutput {
			if c, ok := l.colorMap[level]; ok {
				codeLevel := fmt.Sprintf("[%s]", levelNames[level])
				title := fmt.Sprintf("[%s]", now.Format("15:04:05.000"))
				fileLineStr := ""
				if l.config.ShowFileLine {
					fileLineStr = fileLine
				}
				consoleOutput = fmt.Sprintf("%s%s %s%s\n",
					l.colorMap[4].Sprint(title),
					c.Sprint(codeLevel),
					fileLineStr,
					msg)
			}
		}
		fmt.Print(consoleOutput)
	}

	if l.buffer.Len() >= 1*1024 || level == ERROR || level == WARN { // 达到10KB阈值或是错误/警告级别时触发刷新
		if err := l.flushLocked(); err != nil {
			fmt.Printf("flush failed: %v\n", err)
		}
	}

	if level == ERROR {
		os.Exit(1)
	}
}

// rotateFileLocked 日志文件轮转
// 说明：
//
//	在已获得锁的情况下，执行日志文件轮转：
//	1. 关闭当前日志文件
//	2. 将当前日志文件重命名为带时间戳的归档文件
//	3. 创建新的日志文件
//	4. 异步压缩归档的日志文件
//
// 返回值：
//   - error: 轮转过程中的错误
func (l *Logger) rotateFileLocked() error {
	if err := l.file.Close(); err != nil {
		return err
	}

	baseName := filepath.Base(l.config.Filename)
	timeStamp := time.Now().Format("20060102150405")
	backupName := fmt.Sprintf("%s.%s.log",
		strings.TrimSuffix(baseName, filepath.Ext(baseName)),
		timeStamp)

	backupPath := filepath.Join(l.config.LogDir, backupName)
	if err := os.MkdirAll(l.config.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	if err := os.Rename(l.config.Filename, backupPath); err != nil {
		return fmt.Errorf("failed to rename log file: %v", err)
	}

	file, err := os.OpenFile(l.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %v", err)
	}
	l.file = file
	l.currentSize = 0

	go func() {
		if err := compressLog(backupPath); err != nil {
			fmt.Printf("压缩日志失败: %v\n", err)
		}
	}()

	return nil
}

// compressLog 压缩日志文件
// 说明：
//
//	将指定的日志文件压缩为gzip格式：
//	1. 读取源文件内容
//	2. 创建gzip压缩文件
//	3. 压缩完成后删除源文件
//
// 参数：
//   - srcPath: 要压缩的日志文件路径
//
// 返回值：
//   - error: 压缩过程中的错误
func compressLog(srcPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(srcPath + ".gz")
	if err != nil {
		return err
	}
	defer destFile.Close()

	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return err
	}

	srcFile.Close()

	if err := os.Remove(srcPath); err != nil {
		return err
	}

	return nil
}

// Close 关闭日志记录器
// 说明：
//
//	安全地关闭日志记录器：
//	1. 刷新剩余的缓冲区内容
//	2. 关闭日志文件
//
// 返回值：
//   - error: 关闭过程中的错误
func (l *Logger) Close() error {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.flushLocked()
}

// 以下是各个日志级别的记录方法
// 说明：
//   提供了两组方法：
//   1. 普通方法：直接记录参数
//   2. f后缀方法：支持格式化字符串

// Info 记录信息级别日志
func (l *Logger) Info(args ...interface{}) { l.log(INFO, "", args...) }

// Debug 记录调试级别日志
func (l *Logger) Debug(args ...interface{}) { l.log(DEBUG, "", args...) }

// Warn 记录警告级别日志
func (l *Logger) Warn(args ...interface{}) { l.log(WARN, "", args...) }

// Error 记录错误级别日志
// 注意：调用此方法会导致程序退出
func (l *Logger) Error(args ...interface{}) { l.log(ERROR, "", args...) }

// Infof 记录带格式的信息级别日志
func (l *Logger) Infof(format string, args ...interface{}) { l.log(INFO, format, args...) }

// Debugf 记录带格式的调试级别日志
func (l *Logger) Debugf(format string, args ...interface{}) { l.log(DEBUG, format, args...) }

// Warnf 记录带格式的警告级别日志
func (l *Logger) Warnf(format string, args ...interface{}) { l.log(WARN, format, args...) }

// Errorf 记录带格式的错误级别日志
// 注意：调用此方法会导致程序退出
func (l *Logger) Errorf(format string, args ...interface{}) { l.log(ERROR, format, args...) }
