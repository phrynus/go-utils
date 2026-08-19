package plog

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinLog 返回一个 Gin 中间件，将 HTTP 请求日志记入 Log。
//
// 使用示例：
//
//	r := gin.New()
//	r.Use(log.GinLog())
func (l *Log) GinLog() gin.HandlerFunc {
	sub := l.Sub("GIN")

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		code := c.Writer.Status()
		ip := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		switch {
		case errMsg != "":
			sub.Warnf("[%d] %s %s | %v | %s | %s", code, method, path, latency, ip, errMsg)
		case code >= 500:
			sub.Errorf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		case code >= 400:
			sub.Warnf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		default:
			sub.Infof("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		}
	}
}

// GinRecovery 返回一个 Gin 中间件，捕获 panic 并记录日志。
//
// 使用示例：
//
//	r.Use(log.GinRecovery())
func (l *Log) GinRecovery() gin.HandlerFunc {
	sub := l.Sub("GIN")

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				sub.Errorf("PANIC: %v | %s %s | IP: %s", r, c.Request.Method, c.Request.URL.Path, c.ClientIP())
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
