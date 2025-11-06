package main

import (
	"fmt"

	"github.com/phrynus/go-utils"
)

func TestSystem() {
	systemInfo, err := utils.GetSystemInfo()
	if err != nil {
		fmt.Println("获取系统信息失败:", err)
		return
	}

	// 打印完整的系统信息
	fmt.Println("=== 完整系统信息 ===")
	fmt.Printf("操作系统: %s\n", systemInfo.OS)
	fmt.Printf("架构: %s\n", systemInfo.Arch)
	fmt.Printf("CPU核心数: %d\n", systemInfo.CPUCores)
	fmt.Printf("Go版本: %s\n", systemInfo.GoVersion)
	fmt.Printf("主机名: %s\n", systemInfo.Hostname)
	fmt.Printf("临时目录: %s\n", systemInfo.TempDir)
	fmt.Printf("用户目录: %s\n", systemInfo.HomeDir)
	fmt.Printf("工作目录: %s\n", systemInfo.WorkingDir)
	fmt.Printf("本机IP: %s\n", systemInfo.LocalIP)
	fmt.Printf("外网IP: %s\n", systemInfo.OutboundIP)
	fmt.Printf("电脑名称: %s\n", systemInfo.ComputerName)
	fmt.Printf("CPU ID: %s\n", systemInfo.CPUId)
	fmt.Printf("主板ID: %s\n", systemInfo.BaseboardId)
	fmt.Printf("内存ID: %s\n", systemInfo.MemoryId)
	fmt.Printf("机器码: %s\n", systemInfo.MachineCode)
	fmt.Printf("HTTP代理: %s\n", systemInfo.HttpProxy)
	fmt.Printf("HTTPS代理: %s\n", systemInfo.HttpsProxy)

	// 测试其他网络功能
	fmt.Println("\n=== 网络功能测试 ===")
	port := utils.GetAvailablePort()
	fmt.Printf("可用端口: %d\n", port)

}
