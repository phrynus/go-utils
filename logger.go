package logger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	INFO  = iota // 信息
	DEBUG        // 调试
	WARN         // 警告
	ERROR        // 错误
)

// LogConfig 日志配置结构体
type LogConfig struct {
	Filename     string       // 日志文件名
	LogDir       string       // 日志目录
	MaxSize      int          // 日志文件最大大小(MB)
	StdoutLevels map[int]bool // 输出到标准输出的日志级别
	ColorOutput  bool         // 是否启用彩色输出
}

// Logger 日志记录器结构体
type Logger struct {
	config        LogConfig            // 日志配置
	file          *os.File             // 日志文件
	currentSize   int64                // 当前日志文件大小
	mux           sync.Mutex           // 互斥锁
	colorMap      map[int]*color.Color // 颜色映射
	stdoutLevels  map[int]bool         // 标准输出级别
	buffer        *bytes.Buffer        // 缓冲区
	flushInterval time.Duration        // 刷新间隔
}

var levelNames = []string{
	"INFO",
	"DEBUG",
	"WARN",
	"ERROR",
}

// NewLogger 创建新的日志记录器
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

// flushDaemon 刷新守护进程
func (l *Logger) flushDaemon() {
	for range time.Tick(l.flushInterval) {
		l.mux.Lock()
		l.flushLocked()
		l.mux.Unlock()
	}
}

// flushLocked 刷新缓冲区内容到文件(已加锁)
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

// log 记录日志
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
	logEntry := fmt.Sprintf("[PHRYNUS][%s %s][%s] %s\n",
		now.Format("2006/01/02"),
		now.Format("15:04:05.000"),
		levelNames[level],
		msg)

	l.buffer.WriteString(logEntry)

	if l.stdoutLevels[level] {
		consoleOutput := logEntry
		if l.config.ColorOutput {
			if c, ok := l.colorMap[level]; ok {
				codeLevel := fmt.Sprintf("[%s]", levelNames[level])
				title := fmt.Sprintf("[%s]", now.Format("15:04:05.000"))
				consoleOutput = fmt.Sprintf("%s%s %s\n",
					l.colorMap[4].Sprint(title),
					c.Sprint(codeLevel),
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

// rotateFileLocked 轮转日志文件(已加锁)
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
func (l *Logger) Close() error {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.flushLocked()
}

// Info 记录信息级别日志
func (l *Logger) Info(args ...interface{}) { l.log(INFO, "", args...) }

// Debug 记录调试级别日志
func (l *Logger) Debug(args ...interface{}) { l.log(DEBUG, "", args...) }

// Warn 记录警告级别日志
func (l *Logger) Warn(args ...interface{}) { l.log(WARN, "", args...) }

// Error 记录错误级别日志
func (l *Logger) Error(args ...interface{}) { l.log(ERROR, "", args...) }

// Infof 记录带格式的信息级别日志
func (l *Logger) Infof(format string, args ...interface{}) { l.log(INFO, format, args...) }

// Debugf 记录带格式的调试级别日志
func (l *Logger) Debugf(format string, args ...interface{}) { l.log(DEBUG, format, args...) }

// Warnf 记录带格式的警告级别日志
func (l *Logger) Warnf(format string, args ...interface{}) { l.log(WARN, format, args...) }

// Errorf 记录带格式的错误级别日志
func (l *Logger) Errorf(format string, args ...interface{}) { l.log(ERROR, format, args...) }
