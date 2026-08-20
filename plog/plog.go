package plog

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// level 表示日志级别。
type level uint8

const (
	DEBUG level = iota // 调试信息
	INFO               // 正常运行信息
	WARN               // 警告，不影响运行
	ERROR              // 错误，需要关注
	NONE               // 用于关闭输出
)

// levelNames 级别名称表，包级常量避免每次调用重新分配。
var levelNames = [...]string{"DEBUG", "INFO", "WARN", "ERROR", "NONE"}

// String 返回级别名称，兼容 fmt.Stringer。
func (lvl level) String() string {
	if lvl > NONE {
		return "UNKNOWN"
	}
	return levelNames[lvl]
}

// Config 是日志器的配置，所有字段均可选（Filename 除外）。
type Config struct {
	// Filename 日志文件路径（必填）。
	Filename string

	// MaxSize 单个日志文件最大字节数，超过后自动轮转，默认 50MB。
	MaxSize int64

	// MinLevel 最低记录级别，低于此级别的日志会被丢弃，默认 DEBUG。
	MinLevel level

	// StdoutLevel 同时输出到控制台的最低级别，NONE 表示不输出到控制台，默认 INFO。
	StdoutLevel level

	// Color 控制台是否启用 ANSI 彩色输出，默认 true。
	Color bool

	// FileLine 是否在日志中记录调用位置（文件名:行号），默认 true。
	FileLine bool

	// Tag 日志标识符，会出现在每行日志的 [TAG] 中，默认 "MAIN"。
	Tag string

	// BufferSize 异步模式下 channel 缓冲区大小，默认 4096。
	// 缓冲区满时写入方会短暂阻塞，极端情况下丢弃日志（非阻塞发送兜底）。
	BufferSize int

	// Compress 是否在轮转后自动 gzip 压缩旧日志文件，默认 false。
	Compress bool

	// MaxBackups 保留的压缩日志文件最大数量，超过部分会被自动删除，默认 5。
	// 仅在 Compress 为 true 时生效。
	MaxBackups int
}

// ──────────────────────────── Log ────────────────────────────

// Log 是日志记录器实例，并发安全。
//
// 通过 New() 创建，使用完毕必须调用 Close() 以确保缓冲区落盘。
// 通过 Sub() 可创建共享底层资源的子日志器（不同 tag），子日志器无需 Close。
type Log struct {
	cfg    Config
	mu     sync.Mutex
	file   *os.File
	buf    bytes.Buffer // 文件写入缓冲区
	cbuf   bytes.Buffer // 控制台输出缓冲区（writeConsole 复用，避免每行分配）
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
	level level
	msg   string
	file  string // 调用文件名（不含路径）
	line  int    // 调用行号
	ts    time.Time
	tag   string
}

// ──────────────────────────── 构造函数 ────────────────────────────

// New 根据配置创建一个 Log 并打开日志文件。
// 文件所在目录会自动创建。
func New(cfg Config) (*Log, error) {
	// 应用文档约定的默认值。
	// 注意：Go 零值无法区分“未设置”与显式传入 0/false，这里统一按文档默认覆盖，
	// 因此无法通过零值显式关闭 Color / FileLine，也无法把 StdoutLevel 设为 DEBUG。
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 4096
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 50 << 20 // 50MB
	}
	if cfg.StdoutLevel == 0 {
		cfg.StdoutLevel = INFO
	}
	if !cfg.Color {
		cfg.Color = true
	}
	if !cfg.FileLine {
		cfg.FileLine = true
	}
	if cfg.Tag == "" {
		cfg.Tag = "MAIN"
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 5
	}

	dir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("log: mkdir %s: %w", dir, err)
	}

	f, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("log: open %s: %w", cfg.Filename, err)
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("log: stat %s: %w", cfg.Filename, err)
	}

	l := &Log{
		cfg:  cfg,
		file: f,
		size: fi.Size(),
		pool: sync.Pool{New: func() any { return &logEntry{} }},
	}

	l.ch = make(chan *logEntry, cfg.BufferSize)
	l.done = make(chan struct{})
	go l.run()

	return l, nil
}

// ──────────────────────────── 公开方法 ────────────────────────────

// Debug 记录 DEBUG 级别日志。
func (l *Log) Debug(args ...any) { l.log(DEBUG, "", args...) }

// Debugf 记录 DEBUG 级别格式化日志。
func (l *Log) Debugf(format string, args ...any) { l.log(DEBUG, format, args...) }

// Info 记录 INFO 级别日志。
func (l *Log) Info(args ...any) { l.log(INFO, "", args...) }

// Infof 记录 INFO 级别格式化日志。
func (l *Log) Infof(format string, args ...any) { l.log(INFO, format, args...) }

// Warn 记录 WARN 级别日志。
func (l *Log) Warn(args ...any) { l.log(WARN, "", args...) }

// Warnf 记录 WARN 级别格式化日志。
func (l *Log) Warnf(format string, args ...any) { l.log(WARN, format, args...) }

// Error 记录 ERROR 级别日志。
func (l *Log) Error(args ...any) { l.log(ERROR, "", args...) }

// Errorf 记录 ERROR 级别格式化日志。
func (l *Log) Errorf(format string, args ...any) { l.log(ERROR, format, args...) }

// Flush 强制将缓冲区内容写入磁盘。同步模式可直接调用；
// 异步模式会触发后台刷新（不保证调用返回时已完成落盘）。
func (l *Log) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flush()
}

// Close 关闭日志器，刷新缓冲区并关闭文件。
// 重复调用安全。子日志器（Sub 创建）的 Close 是空操作。
// 会等待所有压缩协程完成后才返回。
func (l *Log) Close() error {
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

	l.cw.Wait()         // 等待所有压缩协程完成
	l.pruneOldBackups() // 最终清理：此时所有 .gz 均已关闭，可确定删除过期备份
	return l.file.Close()
}

// Sub 创建一个共享底层资源但使用不同 tag 的子日志器。
// 子日志器与父日志器写入同一文件、共享同一异步通道，仅 tag 不同。
// 子日志器不需要（也不应该）调用 Close，关闭父日志器即可。
func (l *Log) Sub(tag string) *Log {
	return &Log{
		cfg: Config{
			Filename:    l.cfg.Filename,
			MaxSize:     l.cfg.MaxSize,
			MinLevel:    l.cfg.MinLevel,
			StdoutLevel: l.cfg.StdoutLevel,
			Color:       l.cfg.Color,
			FileLine:    l.cfg.FileLine,
			Tag:         tag,
			MaxBackups:  l.cfg.MaxBackups,
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
	std     *Log
	stdOnce sync.Once
)

// stdInit 初始化默认日志器（懒加载，首次调用自动触发）。
func stdInit() *Log {
	stdOnce.Do(func() {
		var err error
		std, err = New(Config{
			Filename:    "app.log",
			MaxSize:     50 << 10, // 50MB
			MinLevel:    DEBUG,
			StdoutLevel: INFO,
			Color:       true,
			FileLine:    true,
			Tag:         "MAIN",
			Compress:    true,
			MaxBackups:  5,
		})
		if err != nil {
			panic("log: default init failed: " + err.Error())
		}
	})
	return std
}

func SetDefault(cfg Config) {
	stdOnce.Do(func() {
		var err error
		std, err = New(cfg)
		if err != nil {
			panic("log: SetDefault failed: " + err.Error())
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

// Sub 创建一个共享底层资源但使用不同 tag 的子日志器。
func Sub(tag string) *Log {
	return stdInit().Sub(tag)
}

// ──────────────────────────── 异步 goroutine ────────────────────────────

// run 是异步模式的后台写入协程。
func (l *Log) run() {
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

// process 处理一条日志（加锁、写文件、写控制台、必要时轮转）。
func (l *Log) process(e *logEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 直接写入文件缓冲区，避免构造中间字符串
	l.formatTo(&l.buf, e)
	l.buf.WriteByte('\n')

	// 控制台输出
	if e.level >= l.cfg.StdoutLevel && l.cfg.StdoutLevel != NONE {
		l.writeConsole(e)
	}

	// 条件刷新 写缓冲区超过 4KB、或是子日志器同步模式时刷新。
	if l.buf.Len() >= 4096 || l.sub {
		l.flush()
	}
}

// formatTo 将一条日志格式化为 [TAG][2006/01/02 15:04:05.000][LEVEL] file:line message 写入 buf（不含换行）。
func (l *Log) formatTo(b *bytes.Buffer, e *logEntry) {
	b.Grow(128)

	b.WriteByte('[')
	b.WriteString(e.tag)
	b.WriteString("][")
	appendTime(b, e.ts, true)
	b.WriteString("][")
	b.WriteString(levelNames[e.level])
	b.WriteByte(']')
	if e.file != "" {
		b.WriteByte(' ')
		b.WriteString(e.file)
		b.WriteByte(':')
		appendInt(b, e.line, 0)
	}
	b.WriteByte(' ')
	b.WriteString(e.msg)
}

// appendTime 将时间写入 buf，避免 t.Format 的字符串分配。
// withDate 为 true 时输出完整日期，否则只输出时分秒。
func appendTime(b *bytes.Buffer, t time.Time, withDate bool) {
	if withDate {
		y, m, d := t.Date()
		appendInt(b, y, 4)
		b.WriteByte('/')
		appendInt(b, int(m), 2)
		b.WriteByte('/')
		appendInt(b, d, 2)
		b.WriteByte(' ')
	}
	hh, mm, ss := t.Clock()
	appendInt(b, hh, 2)
	b.WriteByte(':')
	appendInt(b, mm, 2)
	b.WriteByte(':')
	appendInt(b, ss, 2)
	b.WriteByte('.')
	appendInt(b, t.Nanosecond()/1e6, 3)
}

// appendInt 以固定最小宽度写入十进制整数，高位补零（width<=0 时不补零）。
func appendInt(b *bytes.Buffer, v, width int) {
	var tmp [16]byte
	i := len(tmp)
	neg := v < 0
	if neg {
		v = -v
	}
	for v > 0 {
		i--
		tmp[i] = byte('0' + v%10)
		v /= 10
	}
	if i == len(tmp) { // v == 0
		i--
		tmp[i] = '0'
	}
	for i > len(tmp)-width {
		i--
		tmp[i] = '0'
	}
	if neg {
		i--
		tmp[i] = '-'
	}
	b.Write(tmp[i:])
}

// writeConsole 输出到控制台（支持 ANSI 彩色）。
// 复用 l.cbuf 避免每行多次分配；调用方必须持有 l.mu。
func (l *Log) writeConsole(e *logEntry) {
	b := &l.cbuf
	b.Reset()
	b.Grow(96)

	if !l.cfg.Color {
		// 无彩色模式，与文件格式一致
		b.WriteByte('[')
		b.WriteString(e.tag)
		b.WriteString("][")
		appendTime(b, e.ts, false)
		b.WriteString("][")
		b.WriteString(levelNames[e.level])
		b.WriteByte(']')
		if e.file != "" {
			b.WriteByte(' ')
			b.WriteString(e.file)
			b.WriteByte(':')
			appendInt(b, e.line, 0)
		}
		b.WriteByte(' ')
		b.WriteString(e.msg)
		b.WriteByte('\n')
		os.Stdout.Write(b.Bytes())
		return
	}

	// 彩色模式：[TAG][time][LEVEL] file:line message
	appendAnsi(b, 0, cWhite)
	b.WriteByte('[')
	b.WriteString(e.tag)
	b.WriteByte(']')
	b.WriteString(ansiReset)

	appendAnsi(b, 0, cGray)
	b.WriteByte('[')
	appendTime(b, e.ts, false)
	b.WriteByte(']')
	b.WriteString(ansiReset)

	var bg int
	switch e.level {
	case DEBUG:
		bg = cBgBlue
	case INFO:
		bg = cBgGreen
	case WARN:
		bg = cBgOrange
	case ERROR:
		bg = cBgRed
	}
	appendAnsi(b, bg, cWhite)
	b.WriteByte('[')
	b.WriteString(levelNames[e.level])
	b.WriteByte(']')
	b.WriteString(ansiReset)

	if e.file != "" {
		b.WriteByte(' ')
		b.WriteString(e.file)
		b.WriteByte(':')
		appendInt(b, e.line, 0)
	}
	b.WriteByte(' ')
	b.WriteString(e.msg)
	b.WriteByte('\n')

	os.Stdout.Write(b.Bytes())
}

// flush 将缓冲区写入文件，必要时轮转。
// 调用方必须持有 l.mu。
func (l *Log) flush() {
	if l.buf.Len() == 0 {
		return
	}
	n, err := l.file.Write(l.buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "log: flush error: %v\n", err)
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
func (l *Log) rotate() {
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
		fmt.Fprintf(os.Stderr, "log: rotate rename: %v\n", err)
	} else {
		renamed = true
	}

	f, err := os.OpenFile(l.cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log: rotate open: %v\n", err)
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
			l.pruneOldBackups()
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

// pruneOldBackups 清理 log 目录中的旧压缩日志，仅保留最近 MaxBackups 份。
// 压缩文件名按时间戳生成，字典序即时间序，最旧的排在最前。
func (l *Log) pruneOldBackups() {
	if l.cfg.MaxBackups <= 0 {
		return
	}

	// 多个压缩协程可能并发触发清理，串行化避免重复删除同一个文件。
	l.mu.Lock()
	defer l.mu.Unlock()

	dir := filepath.Join(filepath.Dir(l.cfg.Filename), "log")
	matches, err := filepath.Glob(filepath.Join(dir, "*.gz"))
	if err != nil {
		return
	}
	sort.Strings(matches)
	if len(matches) <= l.cfg.MaxBackups {
		return
	}

	// 删除最旧的 (len - MaxBackups) 份。
	// 删除失败通常是 Windows 下文件正被其他压缩协程占用（瞬时），忽略即可：
	// 后续 prune 会重试，Close 时还有一次最终清理兜底。
	for _, m := range matches[:len(matches)-l.cfg.MaxBackups] {
		_ = os.Remove(m)
	}
}

// sendEntry 非阻塞发送，channel 满时做一次重试，仍失败则丢弃。
func (l *Log) sendEntry(e *logEntry) {
	select {
	case l.ch <- e:
	default:
		// channel 满，丢弃（高并发保护）
		l.pool.Put(e)
	}
}

// log 是核心日志记录入口。
func (l *Log) log(level level, format string, args ...any) {
	if level < l.cfg.MinLevel {
		return
	}
	if atomic.LoadInt32(&l.closed) == 1 {
		return
	}

	// 构建消息
	msg := fmt.Sprintf(format, args...)
	if format == "" {
		msg = strings.TrimSpace(fmt.Sprint(args...))
	}

	entry := l.pool.Get().(*logEntry)
	entry.level = level
	entry.msg = msg
	entry.file = ""
	entry.line = 0
	entry.ts = time.Now()
	entry.tag = l.cfg.Tag

	// 获取调用位置（文件与行号分开存，延迟到写盘时再拼接，避免此处产生字符串分配）
	if l.cfg.FileLine {
		file, line := callerLocation()
		if file != "" {
			entry.file = file
			entry.line = line
		}
	}

	l.sendEntry(entry)
}

// logEntryNames 是 plog 包内所有日志入口的名称（方法或包级函数），
// 用于在定位真实调用方时跳过内部帧（兼容编译器内联导致的栈帧偏移）。
var logEntryNames = map[string]bool{
	"Debug": true, "Debugf": true, "Info": true, "Infof": true,
	"Warn": true, "Warnf": true, "Error": true, "Errorf": true,
	"Sub": true, "Close": true, "Flush": true,
	"SetDefault": true, "stdInit": true,
}

// callerLocation 返回真实调用方的文件名（不含路径）与行号。
// 跳过本包内部的日志入口，避免包装方法被内联后 runtime.Caller 深度偏移。
func callerLocation() (string, int) {
	var pcs [16]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		fr, more := frames.Next()
		if !more {
			return "", 0
		}
		// 跳过 *Log 的方法（log / Debug / Info / ...）
		if strings.Contains(fr.Function, ".(*Log).") {
			continue
		}
		// 跳过包级快捷函数
		if i := strings.LastIndexByte(fr.Function, '.'); i >= 0 && logEntryNames[fr.Function[i+1:]] {
			continue
		}
		return filepath.Base(fr.File), fr.Line
	}
}

// ──────────────────────────── ANSI color ────────────────────────────

// ansiReset 重置 ANSI 样式。
const ansiReset = "\x1b[0m"

// appendAnsi 将 ANSI 颜色转义前缀写入 buf（无分配）。
// bg 为背景色码（0 表示无背景），fg 为前景色码。
func appendAnsi(b *bytes.Buffer, bg, fg int) {
	b.WriteString("\x1b[")
	if bg == 0 {
		b.WriteString("38;5;")
	} else {
		b.WriteString("48;5;")
		appendInt(b, bg, 0)
		b.WriteString(";38;5;")
	}
	appendInt(b, fg, 0)
	b.WriteByte('m')
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
