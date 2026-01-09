package logger

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// GormLogger GORM日志适配器，将GORM日志对接到自定义日志系统
type GormLogger struct {
	LogLevel                  gormLogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	logger                    *Logger // GORM专用子日志实例
}

// NewGormLogger 创建GORM日志适配器
// 说明：
//
//	创建一个GORM日志适配器，将GORM的日志输出到自定义日志系统
//	会自动克隆一个带"GORM"标识的子logger实例
//
// 返回值：
//   - *GormLogger: GORM日志适配器实例
//
// 示例：
//
//	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
//	    Logger: myLogger.NewGormLogger(),
//	})
func (l *Logger) NewGormLogger() *GormLogger {
	return &GormLogger{
		LogLevel:                  gormLogger.Info,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      false,
		logger:                    l.Clone("GORM", false),
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	newLogger.logger = l.logger // 共享同一个子日志实例
	return &newLogger
}

// Info 打印Info级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		l.logger.Debugf(msg, data...)
	}
}

// Warn 打印Warn级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		l.logger.Warnf(msg, data...)
	}
}

// Error 打印Error级别日志（实际使用Warn级别）
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		l.logger.Warnf(msg, data...)
	}
}

// Trace 打印SQL执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 获取调用文件和行号
	fileInfo := l.getCallerInfo()

	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// 错误日志（使用Warn级别）
		if rows == -1 {
			l.logger.Warnf("%s [%.3fms] [rows:-] %s | %v",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				sql,
				err,
			)
		} else {
			l.logger.Warnf("%s [%.3fms] [rows:%v] %s | %v",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				sql,
				err,
			)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		// 慢查询日志
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.logger.Warnf("%s [%.3fms] [rows:-] %s | %s",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				slowLog,
				sql,
			)
		} else {
			l.logger.Warnf("%s [%.3fms] [rows:%v] %s | %s",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				slowLog,
				sql,
			)
		}
	case l.LogLevel == gormLogger.Info:
		// 普通SQL日志
		if rows == -1 {
			l.logger.Debugf("%s [%.3fms] [rows:-] %s",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				sql,
			)
		} else {
			l.logger.Debugf("%s [%.3fms] [rows:%v] %s",
				fileInfo,
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				sql,
			)
		}
	}
}

// getCallerInfo 获取调用者的文件和行号信息
func (l *GormLogger) getCallerInfo() string {
	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// 跳过 gorm 内部文件
		if strings.Contains(file, "gorm.io/gorm") ||
			strings.Contains(file, "gorm.io/driver") ||
			strings.Contains(file, "/logger/gorm_adapter.go") ||
			strings.Contains(file, "\\logger\\gorm_adapter.go") {
			continue
		}
		// 提取文件名（不包含完整路径）
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		} else if idx := strings.LastIndex(file, "\\"); idx >= 0 {
			file = file[idx+1:]
		}
		return fmt.Sprintf("%s:%d", file, line)
	}
	return "unknown"
}
