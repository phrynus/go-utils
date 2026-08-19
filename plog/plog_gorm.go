package plog

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

// GormLog 将 GORM SQL 日志桥接到本 Log。
type GormLog struct {
	LogLevel                  gormLog.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ShowCaller                bool // 是否在 SQL 日志中显示调用位置，继承父日志器的 FileLine
	log                       *Log // 子日志器
}

// NewGormLog 创建 GORM 日志适配器。
//
// 使用示例：
//
//	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
//	    Log: myLog.NewGormLog(),
//	})
func (l *Log) NewGormLog() *GormLog {
	sub := l.Sub("GORM")
	sub.cfg.FileLine = false // 关闭 log 自身 FileLine，由适配器自行控制 caller 显示
	return &GormLog{
		LogLevel:                  gormLog.Info,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		ShowCaller:                l.cfg.FileLine, // 继承父日志器的 FileLine 设置
		log:                       sub,
	}
}

// LogMode 设置日志级别（实现 gorm.Log.Interface）。
func (g *GormLog) LogMode(level gormLog.LogLevel) gormLog.Interface {
	n := *g
	n.LogLevel = level
	return &n
}

// Info 实现 gorm.Log.Interface。
func (g *GormLog) Info(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Info {
		g.log.Debugf(msg, data...)
	}
}

// Warn 实现 gorm.Log.Interface。
func (g *GormLog) Warn(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Warn {
		g.log.Warnf(msg, data...)
	}
}

// Error 实现 gorm.Log.Interface。
func (g *GormLog) Error(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Error {
		g.log.Errorf(msg, data...)
	}
}

// Trace 实现 gorm.Log.Interface，记录 SQL 执行详情。
func (g *GormLog) Trace(_ context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	if g.LogLevel <= gormLog.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	var caller string
	if g.ShowCaller {
		caller = g.callerInfo() + " "
	}

	switch {
	case err != nil && g.LogLevel >= gormLog.Error &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || !g.IgnoreRecordNotFoundError):
		g.log.Warnf("%s[%.3fms] [rows:%d] %s | %v", caller, ms(elapsed), rows, sql, err)

	case elapsed > g.SlowThreshold && g.SlowThreshold > 0 && g.LogLevel >= gormLog.Warn:
		g.log.Warnf("%sSLOW(>=%v) [%.3fms] [rows:%d] %s", caller, g.SlowThreshold, ms(elapsed), rows, sql)

	case g.LogLevel >= gormLog.Info:
		g.log.Debugf("%s[%.3fms] [rows:%d] %s", caller, ms(elapsed), rows, sql)
	}
}

// callerInfo 返回调用 GORM 的业务代码位置。
func (g *GormLog) callerInfo() string {
	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.Contains(file, "gorm.io/") ||
			strings.Contains(file, "gorm_adapter.go") {
			continue
		}
		short := file
		if idx := strings.LastIndexByte(file, '/'); idx >= 0 {
			short = file[idx+1:]
		} else if idx := strings.LastIndexByte(file, '\\'); idx >= 0 {
			short = file[idx+1:]
		}
		return fmt.Sprintf("%s:%d", short, line)
	}
	return ""
}

// ms 返回 float64 毫秒值。
func ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
