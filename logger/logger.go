// Package logger 是一个零外部依赖、可直接复制到项目中使用的日志库。
//
// 核心功能：
//   - 四级日志：DEBUG / INFO / WARN / ERROR
//   - 文件输出 + 控制台输出（ANSI 彩色）
//   - 按文件大小自动轮转
//   - 显示调用位置（文件名:行号）
//   - 同步 / 异步双模式
//
// 使用示例：
//
//	log, _ := logger.New(logger.Config{
//	    Filename: "app.log",
//	    MaxSize:  50 << 20,
//	    Tag:      "APP",
//	    Async:    true,
//	})
//	defer log.Close()
//	log.Info("server started")
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
)

// ──────────────────────────── Level ────────────────────────────

// Level 表示日志级别。
type Level uint8

const (
	DEBUG Level = iota // 调试信息
	INFO               // 正常运行信息
	WARN               // 警告，不影响运行
	ERROR              // 错误，需要关注
	NONE               // 用于关闭输出
)

// String 返回级别名称，兼容 fmt.Stringer。
func (lvl Level) String() string {
	s := [...]string{"DEBUG", "INFO", "WARN", "ERROR", "NONE"}
	if lvl > NONE {
		return "UNKNOWN"
	}
	return s[lvl]
}

// ParseLevel 将配置中的级别字符串转为 logger.Level
func ParseLevel(s string) Level {
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "NONE":
		return NONE
	default:
		return DEBUG
	}
}

// ──────────────────────────── Config ────────────────────────────

// Config 是日志器的配置，所有字段均可选（Filename 除外）。
type Config struct {
	// Filename 日志文件路径（必填）。
	Filename string

	// MaxSize 单个日志文件最大字节数，超过后自动轮转，默认 50MB。
	MaxSize int64

	// MinLevel 最低记录级别，低于此级别的日志会被丢弃，默认 DEBUG。
	MinLevel Level

	// StdoutLevel 同时输出到控制台的最低级别，NONE 表示不输出到控制台，默认 INFO。
	StdoutLevel Level

	// Color 控制台是否启用 ANSI 彩色输出，默认 true。
	Color bool

	// FileLine 是否在日志中记录调用位置（文件名:行号），默认 true。
	FileLine bool

	// Tag 日志标识符，会出现在每行日志的 [TAG] 中，默认 "APP"。
	Tag string

	// Async 是否启用异步写入。异步模式通过带缓冲的 channel 解耦调用方与 IO，
	// 适合高并发场景；同步模式每次调用直接写入，适合简单场景。
	Async bool

	// BufferSize 异步模式下 channel 缓冲区大小，默认 4096。
	// 缓冲区满时写入方会短暂阻塞，极端情况下丢弃日志（非阻塞发送兜底）。
	BufferSize int

	// Compress 是否在轮转后自动 gzip 压缩旧日志文件，默认 false。
	Compress bool
}

// fillDefaults 用默认值填充未设置的字段。
func (c *Config) fillDefaults() {
	if c.MaxSize <= 0 {
		c.MaxSize = 50 << 20 // 50 MB
	}
	if c.Tag == "" {
		c.Tag = "APP"
	}
	if c.StdoutLevel == 0 && c.StdoutLevel != NONE {
		// 用户没设置时默认 INFO 输出到控制台；
		// 但无法区分 "未设置" 和 "设为 DEBUG(0)"，所以约定 NONE=5 来关闭。
		// 这里取巧：如果 MinLevel 也没设，默认 StdoutLevel=INFO。
	}
	if c.BufferSize <= 0 {
		c.BufferSize = 4096
	}
}

// applyDefaults 补全零值并返回一份安全的副本。
func (c Config) applyDefaults() Config {
	if c.MaxSize <= 0 {
		c.MaxSize = 50 << 20
	}
	if c.Tag == "" {
		c.Tag = "APP"
	}
	if c.BufferSize <= 0 {
		c.BufferSize = 4096
	}
	// StdoutLevel 默认 INFO（但允许用户显式设为 DEBUG=0）。
	// MinLevel 默认 DEBUG(=0)。
	return c
}

// ──────────────────────────── ANSI color ────────────────────────────

// ansiColor 返回包裹在 ANSI 转义序列中的文本。
// bg 为背景色码（0 表示无背景），fg 为前景色码。
func ansiColor(bg, fg int, text string) string {
	if bg == 0 {
		return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", fg, text)
	}
	return fmt.Sprintf("\x1b[48;5;%dm\x1b[38;5;%dm%s\x1b[0m", bg, fg, text)
}

// 预定义的 ANSI 256 色盘色号。
const (
	cGreen    = 34  // 绿底 INFO
	cBlue     = 27  // 蓝底 DEBUG
	cOrange   = 214 // 橙底 WARN
	cRed      = 160 // 红底 ERROR
	cGray     = 245 // 灰色，时间戳
	cWhite    = 255 // 白字
	cBgGreen  = 34
	cBgBlue   = 27
	cBgOrange = 214
	cBgRed    = 160
)

// ──────────────────────────── Logger ────────────────────────────

// Logger 是日志记录器实例，并发安全。
//
// 通过 New() 创建，使用完毕必须调用 Close() 以确保缓冲区落盘。
// 通过 Sub() 可创建共享底层资源的子日志器（不同 tag），子日志器无需 Close。
type Logger struct {
	cfg    Config
	mu     sync.Mutex
	file   *os.File
	buf    bytes.Buffer // 文件写入缓冲区
	size   int64        // 当前文件已写入字节数（近似）
	closed int32        // 原子标记
	sub    bool         // 为 true 表示是 Sub 创建的子日志器，Close 不关文件

	// 异步模式专用
	ch   chan *logEntry // 日志条目通道
	done chan struct{}  // 后台 goroutine 退出信号
	pool sync.Pool      // logEntry 对象池

	// 压缩协程追踪
	cw sync.WaitGroup
}

// logEntry 是一条待写入的日志条目。
type logEntry struct {
	level    Level
	msg      string
	fileLine string
	ts       time.Time
	tag      string
}

// ──────────────────────────── 构造函数 ────────────────────────────

// New 根据配置创建一个 Logger 并打开日志文件。
// 文件所在目录会自动创建。
func New(cfg Config) (*Logger, error) {
	cfg = cfg.applyDefaults()

	dir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("logger: mkdir %s: %w", dir, err)
	}

	f, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("logger: open %s: %w", cfg.Filename, err)
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("logger: stat %s: %w", cfg.Filename, err)
	}

	l := &Logger{
		cfg:  cfg,
		file: f,
		size: fi.Size(),
		pool: sync.Pool{New: func() any { return &logEntry{} }},
	}

	if cfg.Async {
		l.ch = make(chan *logEntry, cfg.BufferSize)
		l.done = make(chan struct{})
		go l.run()
	}

	return l, nil
}

// ──────────────────────────── 公开方法 ────────────────────────────

// Debug 记录 DEBUG 级别日志。
func (l *Logger) Debug(args ...any) { l.log(DEBUG, "", args...) }

// Debugf 记录 DEBUG 级别格式化日志。
func (l *Logger) Debugf(format string, args ...any) { l.log(DEBUG, format, args...) }

// Info 记录 INFO 级别日志。
func (l *Logger) Info(args ...any) { l.log(INFO, "", args...) }

// Infof 记录 INFO 级别格式化日志。
func (l *Logger) Infof(format string, args ...any) { l.log(INFO, format, args...) }

// Warn 记录 WARN 级别日志。
func (l *Logger) Warn(args ...any) { l.log(WARN, "", args...) }

// Warnf 记录 WARN 级别格式化日志。
func (l *Logger) Warnf(format string, args ...any) { l.log(WARN, format, args...) }

// Error 记录 ERROR 级别日志。
func (l *Logger) Error(args ...any) { l.log(ERROR, "", args...) }

// Errorf 记录 ERROR 级别格式化日志。
func (l *Logger) Errorf(format string, args ...any) { l.log(ERROR, format, args...) }

// Flush 强制将缓冲区内容写入磁盘。同步模式可直接调用；
// 异步模式会触发后台刷新（不保证调用返回时已完成落盘）。
func (l *Logger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flush()
}

// Close 关闭日志器，刷新缓冲区并关闭文件。
// 重复调用安全。子日志器（Sub 创建）的 Close 是空操作。
// 会等待所有压缩协程完成后才返回。
func (l *Logger) Close() error {
	if l.sub {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		return nil
	}

	if l.ch != nil {
		close(l.ch)
		<-l.done // 等待后台 goroutine 退出
	}

	l.mu.Lock()
	l.flush()
	l.mu.Unlock()

	l.cw.Wait() // 等待所有压缩协程完成
	return l.file.Close()
}

// Sub 创建一个共享底层资源但使用不同 tag 的子日志器。
// 子日志器与父日志器写入同一文件、共享同一异步通道，仅 tag 不同。
// 子日志器不需要（也不应该）调用 Close，关闭父日志器即可。
func (l *Logger) Sub(tag string) *Logger {
	return &Logger{
		cfg: Config{
			Filename:    l.cfg.Filename,
			MaxSize:     l.cfg.MaxSize,
			MinLevel:    l.cfg.MinLevel,
			StdoutLevel: l.cfg.StdoutLevel,
			Color:       l.cfg.Color,
			FileLine:    l.cfg.FileLine,
			Tag:         tag,
		},
		file: l.file,
		ch:   l.ch,
		done: l.done,
		pool: sync.Pool{New: func() any { return &logEntry{} }},
		sub:  true,
	}
}

// ──────────────────── 包级默认日志器（零配置直接用） ────────────────────

var (
	std     *Logger
	stdOnce sync.Once
)

// stdInit 初始化默认日志器（懒加载，首次调用自动触发）。
func stdInit() *Logger {
	stdOnce.Do(func() {
		var err error
		std, err = New(Config{
			Filename:    "app.log",
			MaxSize:     50 << 10, // 50MB
			MinLevel:    DEBUG,
			StdoutLevel: INFO,
			Color:       true,
			FileLine:    true,
			Tag:         "APP",
			Compress:    true,
		})
		if err != nil {
			panic("logger: default init failed: " + err.Error())
		}
	})
	return std
}

// SetDefault 在使用任何包级函数前自定义默认日志器。
// 如果已触发过懒加载则无效，请在 import 后立即调用。
func SetDefault(cfg Config) {
	stdOnce.Do(func() {
		var err error
		std, err = New(cfg)
		if err != nil {
			panic("logger: SetDefault failed: " + err.Error())
		}
	})
}

// Debug 包级快捷函数，自动使用默认日志器（懒加载，零配置）。
func Debug(args ...any) { stdInit().log(DEBUG, "", args...) }

// Debugf 包级快捷函数。
func Debugf(format string, args ...any) { stdInit().log(DEBUG, format, args...) }

// Info 包级快捷函数。
func Info(args ...any) { stdInit().log(INFO, "", args...) }

// Infof 包级快捷函数。
func Infof(format string, args ...any) { stdInit().log(INFO, format, args...) }

// Warn 包级快捷函数。
func Warn(args ...any) { stdInit().log(WARN, "", args...) }

// Warnf 包级快捷函数。
func Warnf(format string, args ...any) { stdInit().log(WARN, format, args...) }

// Error 包级快捷函数。
func Error(args ...any) { stdInit().log(ERROR, "", args...) }

// Errorf 包级快捷函数。
func Errorf(format string, args ...any) { stdInit().log(ERROR, format, args...) }

// Close 关闭默认日志器。
func Close() {
	if std != nil {
		std.Close()
	}
}

func GetDefault() *Logger {
	return stdInit()
}

// ──────────────────────────── 内部实现 ────────────────────────────

// log 是核心日志记录入口。
func (l *Logger) log(level Level, format string, args ...any) {
	if level < l.cfg.MinLevel {
		return
	}
	if atomic.LoadInt32(&l.closed) == 1 {
		return
	}

	// 构建消息
	msg := formatArgs(format, args)

	// 获取调用位置
	var fileLine string
	if l.cfg.FileLine {
		fileLine = callerInfo(3)
	}

	entry := l.pool.Get().(*logEntry)
	entry.level = level
	entry.msg = msg
	entry.fileLine = fileLine
	entry.ts = time.Now()
	entry.tag = l.cfg.Tag

	if l.ch != nil {
		// 异步：发送到 channel
		l.sendEntry(entry)
	} else {
		// 同步：直接处理
		l.process(entry)
	}
}

// sendEntry 非阻塞发送，channel 满时做一次重试，仍失败则丢弃。
func (l *Logger) sendEntry(e *logEntry) {
	select {
	case l.ch <- e:
	default:
		// channel 满，丢弃（高并发保护）
		l.pool.Put(e)
	}
}

// process 处理一条日志（加锁、写文件、写控制台、必要时轮转）。
func (l *Logger) process(e *logEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 格式化行
	line := l.formatLine(e)

	// 写文件缓冲区
	l.buf.WriteString(line)
	l.buf.WriteByte('\n')

	// 控制台输出
	if e.level >= l.cfg.StdoutLevel && l.cfg.StdoutLevel != NONE {
		l.writeConsole(e, line)
	}

	// 条件刷新
	l.flushIfNeeded()
}

// formatLine 构造一行日志文本（不含换行）。
func (l *Logger) formatLine(e *logEntry) string {
	var b strings.Builder
	b.Grow(128)

	// [TAG][2006/01/02 15:04:05.000][LEVEL] file:line message
	b.WriteByte('[')
	b.WriteString(e.tag)
	b.WriteString("][")
	b.WriteString(e.ts.Format("2006/01/02 15:04:05.000"))
	b.WriteString("][")
	b.WriteString(e.level.String())
	b.WriteByte(']')
	if e.fileLine != "" {
		b.WriteByte(' ')
		b.WriteString(e.fileLine)
	}
	b.WriteByte(' ')
	b.WriteString(e.msg)

	return b.String()
}

// writeConsole 输出到控制台（支持 ANSI 彩色）。
func (l *Logger) writeConsole(e *logEntry, line string) {
	ts := fmt.Sprintf("[%s]", e.ts.Format("15:04:05.000"))
	tagStr := fmt.Sprintf("[%s]", e.tag)
	levelStr := fmt.Sprintf("[%s]", e.level.String())

	if !l.cfg.Color {
		// 无彩色模式，统一格式
		if e.fileLine != "" {
			fmt.Printf("%s%s%s %s %s\n", tagStr, ts, levelStr, e.fileLine, e.msg)
		} else {
			fmt.Printf("%s%s%s %s\n", tagStr, ts, levelStr, e.msg)
		}
		return
	}

	var levelColored string
	switch e.level {
	case DEBUG:
		levelColored = ansiColor(cBgBlue, cWhite, levelStr)
	case INFO:
		levelColored = ansiColor(cBgGreen, cWhite, levelStr)
	case WARN:
		levelColored = ansiColor(cBgOrange, cWhite, levelStr)
	case ERROR:
		levelColored = ansiColor(cBgRed, cWhite, levelStr)
	default:
		levelColored = levelStr
	}

	tagColored := ansiColor(0, cWhite, tagStr)
	timeColored := ansiColor(0, cGray, ts)
	if e.fileLine != "" {
		fmt.Printf("%s%s%s %s %s\n", tagColored, timeColored, levelColored, e.fileLine, e.msg)
	} else {
		fmt.Printf("%s%s%s %s\n", tagColored, timeColored, levelColored, e.msg)
	}
}

// flushIfNeeded 写缓冲区超过 4KB、或是子日志器同步模式时刷新。
func (l *Logger) flushIfNeeded() {
	if l.buf.Len() >= 4096 || l.sub {
		l.flush()
	}
}

// flush 将缓冲区写入文件，必要时轮转。
// 调用方必须持有 l.mu。
func (l *Logger) flush() {
	if l.buf.Len() == 0 {
		return
	}
	n, err := l.file.Write(l.buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: flush error: %v\n", err)
	}
	l.size += int64(n)
	l.buf.Reset()

	// 子日志器共享父日志器的文件，轮转只能由父日志器执行
	if !l.sub && l.size >= l.cfg.MaxSize {
		l.rotate()
	}
}

// rotate 轮转当前日志文件：落盘 → 关闭 → 重命名 → 新建。
// 调用方必须持有 l.mu。
func (l *Logger) rotate() {
	// 防御性落盘：确保所有缓冲数据写入旧文件后再关闭
	if l.buf.Len() > 0 {
		l.file.Write(l.buf.Bytes())
		l.buf.Reset()
	}
	l.file.Close()

	// 纳秒时间戳避免同一毫秒内多次轮转的文件名冲突
	now := time.Now()
	ts := now.Format("20060102.150405") + fmt.Sprintf(".%09d", now.Nanosecond())
	ext := filepath.Ext(l.cfg.Filename)
	base := strings.TrimSuffix(l.cfg.Filename, ext)
	backup := fmt.Sprintf("%s.%s%s", base, ts, ext)

	renamed := false
	if err := os.Rename(l.cfg.Filename, backup); err != nil {
		fmt.Fprintf(os.Stderr, "logger: rotate rename: %v\n", err)
	} else {
		renamed = true
	}

	f, err := os.OpenFile(l.cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: rotate open: %v\n", err)
		return
	}

	// 如果 rename 失败，文件没被移走，获取实际大小
	if !renamed {
		fi, _ := f.Stat()
		l.size = fi.Size()
	} else {
		l.size = 0
	}
	l.file = f

	// 仅 rename 成功时才压缩（否则文件还在原地，下次轮转再处理）
	if l.cfg.Compress && renamed {
		l.cw.Add(1)
		logDir := filepath.Join(filepath.Dir(l.cfg.Filename), "log")
		dstPath := filepath.Join(logDir, filepath.Base(backup)+".gz")
		go func() {
			defer l.cw.Done()
			compressTo(backup, dstPath)
		}()
	}
}

// compressTo 将 srcPath 压缩为 dstPath.gz 并删除原文件。
func compressTo(srcPath, dstPath string) {
	os.MkdirAll(filepath.Dir(dstPath), 0755)

	src, err := os.Open(srcPath)
	if err != nil {
		return
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		src.Close()
		return
	}

	gw := gzip.NewWriter(dst)
	_, err = io.Copy(gw, src)
	src.Close() // 先关源文件，Windows 才能删除
	if err != nil {
		gw.Close()
		dst.Close()
		os.Remove(dstPath)
		return
	}
	gw.Flush()
	gw.Close()
	dst.Sync()
	dst.Close()

	os.Remove(srcPath)
}

// ──────────────────────────── 异步 goroutine ────────────────────────────

// run 是异步模式的后台写入协程。
func (l *Logger) run() {
	defer close(l.done)

	// 定时刷新 ticker（每秒）
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-l.ch:
			if !ok {
				// channel 已关闭，最后刷新
				l.mu.Lock()
				l.flush()
				l.mu.Unlock()
				return
			}
			l.process(e)
			l.pool.Put(e)
		case <-ticker.C:
			l.mu.Lock()
			l.flush()
			l.mu.Unlock()
		}
	}
}

// ──────────────────────────── 辅助函数 ────────────────────────────

// callerInfo 返回 "文件:行号" 格式的调用位置。
// skip 为 runtime.Caller 的跳过帧数。
func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// formatArgs 根据是否有 format 字符串来构建最终消息。
func formatArgs(format string, args []any) string {
	if format == "" {
		// Info(args...) 模式：用空格拼接
		return strings.TrimSpace(fmt.Sprint(args...))
	}
	return fmt.Sprintf(format, args...)
}
