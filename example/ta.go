package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/phrynus/go-utils/ta"
)

func TestTa() {
	// 设置API密钥（可以从环境变量中获取，也可以直接硬编码）
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	// 如果没有设置环境变量，请提示用户
	if apiKey == "" || secretKey == "" {
		fmt.Println("=========== 使用公共API ===========")
	}

	// 创建合约客户端
	futuresClient := futures.NewClient(apiKey, secretKey)

	// 测试合约K线获取
	testFuturesKline(futuresClient)
}

// testFuturesKline 测试获取合约K线并计算技术指标
func testFuturesKline(client *futures.Client) {
	fmt.Println("======= 获取合约市场K线数据 =======\n")

	// 设置参数
	symbol := "ETHUSDT"
	interval := "30m"
	limit := 1000

	// 获取K线数据
	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(context.Background())

	if err != nil {
		fmt.Printf("获取合约K线失败: %v\n", err)
		return
	}

	fmt.Printf("成功获取 %s %s 周期K线 %d 根\n", symbol, interval, len(klines))

	// 转换为ta库的KlineDatas格式
	klineDatas, err := ta.NewKlineDatas(klines, true)
	if err != nil {
		fmt.Printf("转换K线数据失败: %v", err)
		return
	}

	ks, _ := klineDatas.Keep(5)
	fmt.Println("================")
	for _, k := range ks {
		fmt.Println(k)
	}
	fmt.Println("================")
	fmt.Println(klineDatas.GetLastN(1, "c"))
	fmt.Println(klineDatas.GetLastN(0, "c"))
	fmt.Println("================")

	// 计算技术指标
	calculateIndicators(klineDatas, symbol, interval)
}

// calculateIndicators 计算各种技术指标并打印结果
func calculateIndicators(klineDatas ta.KlineDatas, symbol, interval string) {
	// 获取最近的K线时间
	lastTime := time.Unix(0, klineDatas[len(klineDatas)-1].StartTime*int64(time.Millisecond))
	fmt.Printf("最后一根K线开始时间: %s\n\n", lastTime.Format("2006-01-02 15:04:05"))

	ema, err := klineDatas.EMA(25, "close")
	if err != nil {
		fmt.Printf("计算EMA失败: %v", err)
		return
	}

	ema2, err := klineDatas.EMA(25*2, "close")
	if err != nil {
		fmt.Printf("计算EMA失败: %v", err)
		return
	}

	ema3, err := klineDatas.EMA(25*3, "close")
	if err != nil {
		fmt.Printf("计算EMA失败: %v", err)
		return
	}

	fmt.Printf("ema: %v\n", ema.Value())
	fmt.Printf("ema2: %v\n", ema2.Value())
	fmt.Printf("ema3: %v\n", ema3.Value())

	obv, err := klineDatas.OBV()
	if err != nil {
		fmt.Printf("计算OBV失败: %v", err)
		return
	}
	fmt.Printf("obv: %v\n", obv.Value())

	jingzhema, err := klineDatas.JingZheMA(25, 6)
	if err != nil {
		fmt.Printf("计算JingZheMA失败: %v", err)
		return
	}
	cond1, cond2, cond3, cond4, cond5 := jingzhema.Value()
	fmt.Printf("JingZheMA：%v, %v, %v, %v, %v\n", cond1, cond2, cond3, cond4, cond5)

	directionNum := jingzhema.DirectionNum()
	fmt.Printf("JingZheMA方向数：%v\n", directionNum)
}
