package plog

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// FiberLog 返回一个 Fiber 中间件，将 HTTP 请求日志记入 Log。
//
// 使用示例：
//
//	app := fiber.New()
//	app.Use(log.FiberLog())
func (l *Log) FiberLog() func(c fiber.Ctx) error {
	sub := l.Sub("FIBER")

	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		code := c.Response().StatusCode()
		ip := c.IP()
		method := c.Method()
		path := c.Path()

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		switch {
		case errMsg != "":
			sub.Warnf("[%d] %s %s | %v | %s | %s", code, method, path, latency, ip, errMsg)
		case code >= 500:
			sub.Errorf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		case code >= 400:
			sub.Warnf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		default:
			sub.Debugf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		}

		return err
	}
}

// FiberRecovery 返回一个 Fiber 中间件，捕获 panic 并记录日志。
//
// 使用示例：
//
//	app.Use(log.FiberRecovery())
func (l *Log) FiberRecovery() func(c fiber.Ctx) error {
	sub := l.Sub("FIBER")

	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				sub.Errorf("PANIC: %v | %s %s | IP: %s", r, c.Method(), c.Path(), c.IP())
				_ = c.Status(500).JSON(fiber.Map{
					"code":    500,
					"message": "internal server error",
				})
			}
		}()
		return c.Next()
	}
}
