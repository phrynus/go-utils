package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"
)

// ========== 系统信息 ==========

// SystemInfo 系统信息
type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPUCores   int    `json:"cpu_cores"`
	GoVersion  string `json:"go_version"`
	Hostname   string `json:"hostname"`
	TempDir    string `json:"temp_dir"`
	HomeDir    string `json:"home_dir"`
	WorkingDir string `json:"working_dir"`
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() (*SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "unknown"
	}

	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "unknown"
	}

	return &SystemInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		CPUCores:   runtime.NumCPU(),
		GoVersion:  runtime.Version(),
		Hostname:   hostname,
		TempDir:    os.TempDir(),
		HomeDir:    homeDir,
		WorkingDir: workingDir,
	}, nil
}

// GetLocalIP 获取本机IP地址
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"

	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// GetOutboundIP 获取对外通信的IP地址
func GetOutboundIP() string {

	resp, err := http.Get("https://v6r.ipip.net/") // https://api.ipify.org/ https://v6r.ipip.net/
	if err != nil {
		return "0.0.0.0"
	}
	defer resp.Body.Close() // 确保关闭响应体

	// 读取响应体内容
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0.0.0.0"
	}
	return string(ip)

}

// 获取CPU ID
func GetCpuId() string {
	cmd := exec.Command("wmic", "cpu", "get", "ProcessorID")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Println(string(out))
	str := string(out)
	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	str = reg.ReplaceAllString(str, "")
	return str[11:]
}

// 获取主板 ID
func GetBaseboardId() string {
	cmd := exec.Command("wmic", "baseboard", "get", "serialnumber")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	str := string(out)
	reg := regexp.MustCompile(`\s+`)
	str = reg.ReplaceAllString(str, "")
	return str[12:]
}

// 获取内存 ID
func GetMemoryId() string {
	cmd := exec.Command("wmic", "memorychip", "get", "serialnumber")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	str := string(out)
	reg := regexp.MustCompile(`\s+`)
	str = reg.ReplaceAllString(str, "")
	return str[12:]
}

// 取机器码UUID
func GetMachineCode() string {
	cmd := exec.Command("wmic", "csproduct", "get", "uuid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	str := string(out)
	reg := regexp.MustCompile(`\s+`)
	str = reg.ReplaceAllString(str, "")
	return str[4:]
}

// 取当前电脑代理
func GetProxy() (string, string) {
	httpProxy := os.Getenv("http_proxy")
	if httpProxy == "" {
		httpProxy = os.Getenv("HTTP_PROXY")
	}

	httpsProxy := os.Getenv("https_proxy")
	if httpsProxy == "" {
		httpsProxy = os.Getenv("HTTPS_PROXY")
	}

	return httpProxy, httpsProxy
}

// ========== 网络工具 ==========

// DownloadFile 下载文件
func DownloadFile(url, filepath string) error {
	// 创建目标文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 发起HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}

// CheckPort 检查端口是否可用
func CheckPort(host string, port int) bool {
	timeout := time.Second * 2
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetAvailablePort 获取一个可用的端口
func GetAvailablePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}
