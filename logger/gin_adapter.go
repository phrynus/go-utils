package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger Gin日志中间件，将请求日志接入自定义日志系统
// 说明：
//
//	创建一个Gin中间件，用于记录HTTP请求信息
//	会自动克隆一个带"GIN"标识的子logger实例
//
// 返回值:
//   - gin.HandlerFunc: Gin中间件函数
//
// 示例：
//
//	r := gin.New()
//	r.Use(myLogger.GinLogger())
func (l *Logger) GinLogger() gin.HandlerFunc {
	ginLogger := l.Clone("GIN", false)

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 请求信息
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 根据状态码选择日志级别
		logFunc := ginLogger.Debugf
		if statusCode >= 500 {
			logFunc = ginLogger.Warnf
		} else if statusCode >= 400 {
			logFunc = ginLogger.Warnf
		}

		// 记录日志
		if errorMessage != "" {
			logFunc("[%d] %s %s | %v | %s | %s",
				statusCode,
				method,
				path,
				latency,
				clientIP,
				errorMessage,
			)
		} else {
			logFunc("[%d] %s %s | %v | %s",
				statusCode,
				method,
				path,
				latency,
				clientIP,
			)
		}
	}
}

// GinRecovery 自定义恢复中间件，捕获panic并记录日志
// 说明：
//
//	创建一个Gin中间件，用于捕获panic并记录到日志系统
//	当panic发生时，会返回500错误给客户端
//
// 返回值:
//   - gin.HandlerFunc: Gin中间件函数
//
// 示例：
//
//	r := gin.New()
//	r.Use(myLogger.GinRecovery())
func (l *Logger) GinRecovery() gin.HandlerFunc {
	ginLogger := l.Clone("GIN", false)

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				ginLogger.Warnf("Panic recovered: %v | %s %s | IP: %s",
					err,
					c.Request.Method,
					c.Request.URL.Path,
					c.ClientIP(),
				)

				// 返回500错误
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": fmt.Sprintf("服务器内部错误: %v", err),
				})
			}
		}()

		c.Next()
	}
}
