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
	"sync/atomic"
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
	PHRYNUS      string       // 日志标识符，用于标识日志来源
}

// logEntry 表示一个日志条目
type logEntry struct {
	level     int
	message   string
	fileLine  string
	timestamp time.Time
	dateStr   string // 预格式化的日期字符串
	timeStr   string // 预格式化的时间字符串
	phrynus   string // 日志标识符
}

// Logger 日志记录器结构体
// 说明：
//
//	实现了一个高性能的日志记录系统，特点：
//	1. 支持日志文件轮转和压缩
//	2. 支持多种日志级别
//	3. 支持控制台彩色输出
//	4. 使用异步写入和缓冲区提高性能
//	5. 支持并发安全的日志记录
//	6. 使用对象池减少内存分配
//	7. 优化内存对齐减少填充字节
type Logger struct {
	// 8字节对齐的指针字段
	file      *os.File       // 当前日志文件句柄
	buffer    *bytes.Buffer  // 写入缓冲区
	logChan   chan *logEntry // 日志条目通道
	flushChan chan struct{}  // 刷新信号通道
	closeChan chan struct{}  // 关闭信号通道

	// 大结构体字段
	config      LogConfig       // 日志配置信息
	colorMap    [5]*color.Color // 日志级别对应的颜色映射（数组访问更快）
	bufferPool  sync.Pool       // 缓冲区对象池
	builderPool sync.Pool       // 字符串构建器对象池

	// 8字节对齐的数值字段
	currentSize   int64         // 当前日志文件大小
	flushInterval time.Duration // 缓冲区刷新间隔

	// 4字节对齐的字段
	mux      sync.Mutex // 互斥锁，保证并发安全
	isClosed int32      // 关闭状态标记（原子操作）

	// 较小的字段
	stdoutLevels map[int]bool // 控制台输出级别配置
	phrynus      string       // 日志标识符

	// 父子关系管理（用于级联关闭）
	parent   *Logger              // 父logger
	children map[*Logger]struct{} // 子logger集合
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

	colorMap := [5]*color.Color{
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
		phrynus:       config.PHRYNUS,
		logChan:       make(chan *logEntry, 50000), // 缓冲50k个日志条目，减少阻塞
		flushChan:     make(chan struct{}, 1),
		closeChan:     make(chan struct{}),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 256)) // 预分配256字节容量
			},
		},
		builderPool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
		isClosed: 0,
		children: make(map[*Logger]struct{}), // 初始化子logger集合
	}

	go logger.asyncWriter()
	go logger.flushDaemon()

	return logger, nil
}

// asyncWriter 异步日志写入器
// 说明：
//
//	从通道中读取日志条目并异步处理：
//	1. 格式化日志消息
//	2. 写入缓冲区
//	3. 处理控制台输出
//	4. 触发缓冲区刷新
//	5. 收到nil结束标记时优雅退出
func (l *Logger) asyncWriter() {
	for {
		select {
		case entry := <-l.logChan:
			if entry == nil {
				// 收到nil表示关闭，立即退出
				return
			}
			l.processLogEntry(entry)
		case <-l.closeChan:
			// 收到关闭信号，立即退出
			return
		}
	}
}

// processLogEntry 处理单个日志条目
func (l *Logger) processLogEntry(entry *logEntry) {
	l.mux.Lock()
	defer l.mux.Unlock()

	// 使用对象池获取缓冲区
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	// 格式化日志条目
	l.formatLogEntry(buf, entry)

	// 写入主缓冲区
	l.buffer.Write(buf.Bytes())

	// 处理控制台输出
	if l.stdoutLevels[entry.level] {
		l.writeToConsole(entry)
	}

	// 根据条件触发刷新
	if l.shouldFlush(entry.level) {
		if err := l.flushLocked(); err != nil {
			fmt.Printf("flush failed: %v\n", err)
		}
	}

	// 错误级别直接退出
	if entry.level == ERROR {
		os.Exit(1)
	}
}

// formatLogEntry 格式化日志条目到缓冲区
func (l *Logger) formatLogEntry(buf *bytes.Buffer, entry *logEntry) {
	// 预估容量并分配缓冲区，避免多次扩容
	buf.Grow(100 + len(entry.phrynus) + len(entry.dateStr) + len(entry.timeStr) + len(levelNames[entry.level]) + len(entry.fileLine) + len(entry.message))

	buf.WriteString("[")
	buf.WriteString(entry.phrynus)
	buf.WriteString("][")
	buf.WriteString(entry.dateStr)
	buf.WriteString(" ")
	buf.WriteString(entry.timeStr)
	buf.WriteString("][")
	buf.WriteString(levelNames[entry.level])
	buf.WriteString("] ")
	if entry.fileLine != "" {
		buf.WriteString(entry.fileLine)
	}
	buf.WriteString(entry.message)
	buf.WriteString("\n")
}

// writeToConsole 输出到控制台（带彩色）
func (l *Logger) writeToConsole(entry *logEntry) {
	if !l.config.ColorOutput {
		fmt.Print(entry.message)
		return
	}

	if entry.level >= 0 && entry.level < len(l.colorMap) && l.colorMap[entry.level] != nil {
		codeLevel := fmt.Sprintf("[%s]", levelNames[entry.level])
		title := fmt.Sprintf("[%s]", entry.timeStr)
		fileLineStr := entry.fileLine

		fmt.Printf("%s%s %s%s\n",
			l.colorMap[4].Sprint(title),
			l.colorMap[entry.level].Sprint(codeLevel),
			fileLineStr,
			entry.message)
	}
}

// shouldFlush 判断是否应该刷新缓冲区
func (l *Logger) shouldFlush(level int) bool {
	return l.buffer.Len() >= 4096 || level == ERROR || level == WARN
}

// flushDaemon 日志刷新守护进程
// 说明：
//
//	定期将缓冲区中的日志内容写入文件
//	支持批量刷新：当收到刷新信号时，会等待一小段时间来合并多个刷新请求
//	这是一个需要在后台持续运行的goroutine
func (l *Logger) flushDaemon() {
	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	var flushPending bool
	var flushTimer <-chan time.Time

	for {
		select {
		case <-ticker.C:
			l.mux.Lock()
			if l.buffer.Len() > 0 {
				l.flushLocked()
			}
			l.mux.Unlock()
		case <-l.flushChan:
			if !flushPending {
				// 第一次收到刷新信号，设置批量刷新定时器
				flushPending = true
				flushTimer = time.After(10 * time.Millisecond) // 10ms批量窗口
			}
		case <-flushTimer:
			// 批量刷新时间窗口结束，执行刷新
			l.mux.Lock()
			l.flushLocked()
			l.mux.Unlock()
			flushPending = false
			flushTimer = nil
		case <-l.closeChan:
			// 退出前最后一次刷新
			if flushPending {
				l.mux.Lock()
				l.flushLocked()
				l.mux.Unlock()
			}
			return
		}
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

// getFileInfo 获取实时的文件信息
func (l *Logger) getFileInfo() string {
	_, file, line, ok := runtime.Caller(3) // 3表示调用栈的深度
	if ok {
		return fmt.Sprintf("%s:%d ", filepath.Base(file), line)
	}
	return ""
}

// log 核心日志记录函数
// 说明：
//
//	实现了高性能的日志记录核心逻辑：
//	1. 格式化日志消息
//	2. 异步写入到通道
//	3. 使用缓存减少runtime.Caller调用
//
// 参数：
//   - level: 日志级别
//   - format: 格式化字符串
//   - args: 格式化参数
func (l *Logger) log(level int, format string, args ...interface{}) {
	// 检查是否已关闭
	if atomic.LoadInt32(&l.isClosed) == 1 {
		return
	}

	// 使用对象池获取字符串构建器
	builder := l.builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		l.builderPool.Put(builder)
	}()

	// 格式化消息
	if format == "" {
		fmt.Fprint(builder, args...)
	} else {
		fmt.Fprintf(builder, format, args...)
	}
	msg := builder.String()

	// 获取文件行号信息（实时）
	fileLine := ""
	if l.config.ShowFileLine {
		fileLine = l.getFileInfo()
	}

	// 获取当前时间并预格式化
	now := time.Now()

	// 创建日志条目
	entry := &logEntry{
		level:     level,
		message:   msg,
		fileLine:  fileLine,
		timestamp: now,
		dateStr:   now.Format("2006/01/02"),
		timeStr:   now.Format("15:04:05.000"),
		phrynus:   l.phrynus,
	}

	// 安全地发送到通道，使用recover处理已关闭通道的情况
	defer func() {
		if r := recover(); r != nil {
			// 通道已关闭，静默丢弃日志
			return
		}
	}()

	// 非阻塞发送到通道，如果通道满则触发刷新并重试
	select {
	case l.logChan <- entry:
		// 成功发送
	default:
		// 通道满时，触发刷新以腾出空间
		select {
		case l.flushChan <- struct{}{}:
		default:
		}

		// 短暂等待刷新完成，然后重试发送
		select {
		case l.logChan <- entry:
			// 重试成功
		default:
			// 如果仍然无法发送，说明系统过载，静默丢弃
			// 这在极高并发情况下是正常的保护措施
		}
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
//	1. 标记为已关闭状态
//	2. 如果是主logger，级联关闭所有子logger
//	3. 如果是子logger，只关闭自己并从父logger中移除
//	4. 关闭通道让异步goroutine退出
//	5. 等待异步操作完成
//	6. 刷新剩余的缓冲区内容
//	7. 关闭日志文件（仅主logger）
//
// 返回值：
//   - error: 关闭过程中的错误
func (l *Logger) Close() error {
	// 标记为已关闭
	if !atomic.CompareAndSwapInt32(&l.isClosed, 0, 1) {
		return nil // 已经关闭
	}

	// 如果是主logger（没有parent），级联关闭所有子logger
	if l.parent == nil {
		// 复制子logger列表，避免在遍历时修改
		l.mux.Lock()
		children := make([]*Logger, 0, len(l.children))
		for child := range l.children {
			children = append(children, child)
		}
		l.mux.Unlock()

		// 关闭所有子logger
		for _, child := range children {
			child.Close()
		}
	} else {
		// 如果是子logger，从父logger中移除自己
		l.parent.mux.Lock()
		delete(l.parent.children, l)
		l.parent.mux.Unlock()
		// 子logger不关闭共享资源
		return nil
	}

	// 安全地关闭通道，避免重复关闭
	defer func() {
		if r := recover(); r != nil {
			// 忽略通道重复关闭的panic
		}
	}()

	// 关闭通道，让异步goroutine自然退出
	close(l.logChan)
	close(l.closeChan)

	// 等待异步goroutine完成
	time.Sleep(200 * time.Millisecond)

	l.mux.Lock()
	defer l.mux.Unlock()

	// 最后一次刷新
	if err := l.flushLocked(); err != nil {
		return err
	}

	// 关闭文件（仅主logger）
	if l.file != nil {
		return l.file.Close()
	}

	return nil
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

// Clone 复制Logger实例并更换PHRYNUS标识符
// 说明：
//
//	创建一个新的Logger实例，复制原有配置但使用新的PHRYNUS标识符
//	主要用于在同一个应用中创建多个具有不同标识符的日志记录器
//	注意：克隆的Logger实例会共享同一个异步写入系统（通道和goroutine）以提高效率
//
// 参数：
//   - newPHRYNUS: 新的日志标识符
//
// 返回值：
//   - *Logger: 新的日志记录器实例
func (l *Logger) Clone(newPHRYNUS string) *Logger {
	l.mux.Lock()
	defer l.mux.Unlock()

	// 复制配置并更新PHRYNUS
	newConfig := l.config
	newConfig.PHRYNUS = newPHRYNUS

	// 创建新的Logger实例，共享异步写入系统
	newLogger := &Logger{
		config:        newConfig,
		file:          l.file, // 共享同一个文件句柄
		currentSize:   l.currentSize,
		colorMap:      l.colorMap,     // 共享颜色映射
		stdoutLevels:  l.stdoutLevels, // 共享输出级别配置
		buffer:        bytes.NewBuffer(nil),
		flushInterval: l.flushInterval,
		phrynus:       newPHRYNUS,
		logChan:       l.logChan,   // 共享同一个日志通道
		flushChan:     l.flushChan, // 共享同一个刷新通道
		closeChan:     l.closeChan, // 共享同一个关闭通道
		bufferPool: sync.Pool{ // 独立的对象池，避免并发竞争
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 256))
			},
		},
		builderPool: sync.Pool{ // 独立的对象池，避免并发竞争
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
		isClosed: l.isClosed,                 // 共享关闭状态
		parent:   l,                          // 设置父logger
		children: make(map[*Logger]struct{}), // 初始化子logger集合
	}

	// 将新logger添加到父logger的子logger集合中
	l.children[newLogger] = struct{}{}

	return newLogger
}
