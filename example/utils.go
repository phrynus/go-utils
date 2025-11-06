package main

import (
	"fmt"

	"github.com/phrynus/go-utils"
)

func TestSystem() {
	systemInfo, err := utils.GetSystemInfo()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(systemInfo)

	localIp := utils.GetLocalIP()
	fmt.Println(localIp)

	outboundIp := utils.GetOutboundIP()
	fmt.Println(outboundIp)

	cpuId := utils.GetCpuId()
	fmt.Println(cpuId)

	baseboardId := utils.GetBaseboardId()
	fmt.Println(baseboardId)

	memoryId := utils.GetMemoryId()
	fmt.Println(memoryId)

	machineCode := utils.GetMachineCode()
	fmt.Println(machineCode)

	httpProxy, httpsProxy := utils.GetProxy()
	fmt.Println(httpProxy)
	fmt.Println(httpsProxy)

	port := utils.GetAvailablePort()
	fmt.Println(port)
}
