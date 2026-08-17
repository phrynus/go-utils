package main

import (
	"time"

	"logger"
)

func main() {
	defer logger.Close()

	// 零配置，直接用！自动写入 app.log，控制台彩色输出
	logger.Info("═══ 服务启动 ═══")
	logger.Infof("监听端口: %d", 8080)
	logger.Debugf("配置项: %s=%v", "max_conn", 100)
	logger.Warnf("响应偏慢: %v", 1.2*float64(time.Second))
	logger.Errorf("启动失败: %v", "port already in use")

	// 循环1w次
	for i := 0; i < 10000; i++ {
		logger.Debugf("循环日志 #%d", i+1)
	}
}
