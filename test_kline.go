package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/phrynus/go-utils/ta"
)

func main() {
	// 设置API密钥（可以从环境变量中获取，也可以直接硬编码）
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	// 如果没有设置环境变量，请提示用户
	if apiKey == "" || secretKey == "" {
		log.Println("警告: 未设置 BINANCE_API_KEY 或 BINANCE_SECRET_KEY 环境变量")
		log.Println("继续使用公共API（有请求限制）...")
	}

	// 创建合约客户端
	futuresClient := futures.NewClient(apiKey, secretKey)

	// 测试合约K线获取
	testFuturesKline(futuresClient)
}

// testFuturesKline 测试获取合约K线并计算技术指标
func testFuturesKline(client *futures.Client) {
	fmt.Println("\n======= 获取合约市场K线数据 =======")

	// 设置参数
	symbol := "BTCUSDT"
	interval := "15m"
	limit := 600

	// 获取K线数据
	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(context.Background())

	if err != nil {
		log.Fatalf("获取合约K线失败: %v", err)
	}

	fmt.Printf("成功获取 %s %s 周期K线 %d 根\n", symbol, interval, len(klines))

	// 转换为ta库的KlineDatas格式
	klineDatas, err := ta.NewKlineDatas(klines, false)
	if err != nil {
		log.Fatalf("转换K线数据失败: %v", err)
	}

	// 计算技术指标
	calculateIndicators(klineDatas, symbol, interval)
}

// calculateIndicators 计算各种技术指标并打印结果
func calculateIndicators(klineDatas ta.KlineDatas, symbol, interval string) {
	// 获取最近的K线时间
	lastTime := time.Unix(0, klineDatas[len(klineDatas)-1].StartTime*int64(time.Millisecond))
	fmt.Printf("最后一根K线时间: %s\n", lastTime.Format("2006-01-02 15:04:05"))

	dpo, err := klineDatas.DPO("close", 15, 19, 11, 3)
	if err != nil {
		log.Fatalf("计算DPO失败: %v", err)
	}
	// func (t *ta.TaDpo) Value() (short float64, long float64, diff float64, high float64, low float64, mid float64)
	short, long, diff, high, low, mid := dpo.Value()
	fmt.Printf("DPO: %v, %v, %v, %v, %v, %v\n", short, long, diff, high, low, mid)
}
