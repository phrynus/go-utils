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
	"strings"
	"sync"
	"time"
)

var ipEndpoints = []string{
	"https://v6r.ipip.net/",          // 0.438 真
	"https://icanhazip.com/",         // 0.070
	"https://checkip.amazonaws.com/", // 0.160
	"https://api.ipify.org/",         // 0.264
	"https://ipinfo.io/ip",           // 0.462
	"https://api.ip.sb/ip",           // 0.921

}

var Ip string = "0.0.0.0"
var ipMutex sync.RWMutex

// ========== 系统信息 ==========

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	CPUCores     int    `json:"cpu_cores"`
	GoVersion    string `json:"go_version"`
	Hostname     string `json:"hostname"`
	TempDir      string `json:"temp_dir"`
	HomeDir      string `json:"home_dir"`
	WorkingDir   string `json:"working_dir"`
	LocalIP      string `json:"local_ip"`
	OutboundIP   string `json:"outbound_ip"`
	ComputerName string `json:"computer_name"`
	CPUId        string `json:"cpu_id"`
	BaseboardId  string `json:"baseboard_id"`
	MemoryId     string `json:"memory_id"`
	MachineCode  string `json:"machine_code"`
	HttpProxy    string `json:"http_proxy"`
	HttpsProxy   string `json:"https_proxy"`
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

	// 获取代理信息
	httpProxy, httpsProxy := GetProxy()

	return &SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		Hostname:     hostname,
		TempDir:      os.TempDir(),
		HomeDir:      homeDir,
		WorkingDir:   workingDir,
		LocalIP:      GetLocalIP(),
		OutboundIP:   GetOutboundIP(),
		ComputerName: GetComputerName(),
		CPUId:        GetCpuId(),
		BaseboardId:  GetBaseboardId(),
		MemoryId:     GetMemoryId(),
		MachineCode:  GetMachineCode(),
		HttpProxy:    httpProxy,
		HttpsProxy:   httpsProxy,
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

// isValidIP 验证IP地址格式是否有效
func isValidIP(ip string) bool {
	return net.ParseIP(strings.TrimSpace(ip)) != nil
}

// GetOutboundIP 获取对外通信的IP地址
// 如果已缓存有效IP，直接返回；否则尝试多个服务获取IP并缓存结果
func GetOutboundIP() string {
	// 先检查是否已有缓存的有效IP
	ipMutex.RLock()
	if Ip != "0.0.0.0" && isValidIP(Ip) {
		ipMutex.RUnlock()
		return Ip
	}
	ipMutex.RUnlock()

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// 尝试每个IP检测服务
	for _, endpoint := range ipEndpoints {
		resp, err := client.Get(endpoint)
		if err != nil {
			continue // 尝试下一个服务
		}

		// 读取响应体内容
		ip, err := io.ReadAll(resp.Body)
		resp.Body.Close() // 立即关闭响应体

		if err != nil {
			continue // 尝试下一个服务
		}

		// 验证获取到的IP地址
		ipStr := strings.TrimSpace(string(ip))
		if isValidIP(ipStr) {
			// 缓存有效的IP地址
			ipMutex.Lock()
			Ip = ipStr
			ipMutex.Unlock()
			return ipStr
		}
	}

	// 所有服务都失败，返回默认值
	return "0.0.0.0"
}

// ResetIPCache 重置IP缓存，强制下次调用GetOutboundIP时重新获取
func ResetIPCache() {
	ipMutex.Lock()
	Ip = "0.0.0.0"
	ipMutex.Unlock()
}

// GetComputerName 获取电脑名称/主机名
func GetComputerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// 获取CPU ID
func GetCpuId() string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("wmic", "cpu", "get", "ProcessorID")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ""
		}
		str := string(out)
		reg := regexp.MustCompile(`\s+`)
		str = reg.ReplaceAllString(str, "")
		if len(str) > 11 {
			return str[11:]
		}
		return ""
	case "linux":
		// 方法1: 尝试从 /proc/cpuinfo 读取更有用的CPU信息
		if content, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := strings.Split(string(content), "\n")

			// 优先查找 CPU serial number
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "Serial") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						serial := strings.TrimSpace(parts[1])
						if serial != "" && serial != "0000000000000000" {
							return serial
						}
					}
				}
			}

			// 查找 CPU ID 或 processor id (通常在ARM系统中)
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "cpu serial") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						serial := strings.TrimSpace(parts[1])
						if serial != "" {
							return serial
						}
					}
				}
			}

			// 查找 model name 作为备选方案
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "model name") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						model := strings.TrimSpace(parts[1])
						if model != "" {
							// 使用model name的hash作为唯一标识
							return fmt.Sprintf("%x", strings.ReplaceAll(model, " ", ""))[:16]
						}
					}
				}
			}
		}

		// 方法2: 尝试从 dmidecode 获取处理器信息
		if cmd := exec.Command("dmidecode", "-t", "processor"); cmd != nil {
			if out, err := cmd.CombinedOutput(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), "id:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							id := strings.TrimSpace(parts[1])
							if id != "" && id != "Not Specified" {
								return strings.ReplaceAll(id, " ", "")
							}
						}
					}
				}
			}
		}

		// 方法3: 尝试读取 /sys/devices/virtual/dmi/id/processor_*
		if files, err := os.ReadDir("/sys/devices/virtual/dmi/id/"); err == nil {
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "processor_") {
					if content, err := os.ReadFile("/sys/devices/virtual/dmi/id/" + file.Name()); err == nil {
						id := strings.TrimSpace(string(content))
						if id != "" && id != "Not Specified" {
							return id
						}
					}
				}
			}
		}

		return ""
	default:
		return ""
	}
}

// 获取主板 ID
func GetBaseboardId() string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("wmic", "baseboard", "get", "serialnumber")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ""
		}
		str := string(out)
		reg := regexp.MustCompile(`\s+`)
		str = reg.ReplaceAllString(str, "")
		if len(str) > 12 {
			return str[12:]
		}
		return ""
	case "linux":
		// 方法1: 直接读取 /sys/class/dmi/id/board_serial
		if content, err := os.ReadFile("/sys/class/dmi/id/board_serial"); err == nil {
			serial := strings.TrimSpace(string(content))
			if serial != "" && serial != "ToBeFilledByO.E.M." && serial != "Not Specified" {
				return serial
			}
		}

		// 方法2: 读取 /sys/devices/virtual/dmi/id/board_serial
		if content, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_serial"); err == nil {
			serial := strings.TrimSpace(string(content))
			if serial != "" && serial != "ToBeFilledByO.E.M." && serial != "Not Specified" {
				return serial
			}
		}

		// 方法3: 尝试使用 dmidecode 获取主板序列号
		if cmd := exec.Command("dmidecode", "-s", "baseboard-serial-number"); cmd != nil {
			if out, err := cmd.CombinedOutput(); err == nil {
				serial := strings.TrimSpace(string(out))
				if serial != "" && serial != "ToBeFilledByO.E.M." && serial != "Not Specified" {
					return serial
				}
			}
		}

		// 方法4: 尝试从 dmidecode 完整输出中获取主板信息
		if cmd := exec.Command("dmidecode", "-t", "baseboard"); cmd != nil {
			if out, err := cmd.CombinedOutput(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "Serial Number:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							serial := strings.TrimSpace(parts[1])
							if serial != "" && serial != "ToBeFilledByO.E.M." && serial != "Not Specified" {
								return serial
							}
						}
					}
				}
			}
		}

		// 方法5: 尝试读取主板产品名称作为备选
		if content, err := os.ReadFile("/sys/class/dmi/id/board_name"); err == nil {
			name := strings.TrimSpace(string(content))
			if name != "" && name != "ToBeFilledByO.E.M." && name != "Not Specified" {
				// 结合厂商信息
				if vendor, err := os.ReadFile("/sys/class/dmi/id/board_vendor"); err == nil {
					vendorName := strings.TrimSpace(string(vendor))
					if vendorName != "" && vendorName != "ToBeFilledByO.E.M." {
						return fmt.Sprintf("%s-%s", vendorName, name)
					}
				}
				return name
			}
		}

		return ""
	default:
		return ""
	}
}

// 获取内存 ID
func GetMemoryId() string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("wmic", "memorychip", "get", "serialnumber")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ""
		}
		str := string(out)
		reg := regexp.MustCompile(`\s+`)
		str = reg.ReplaceAllString(str, "")
		if len(str) > 12 {
			return str[12:]
		}
		return ""
	case "linux":
		// 方法1: 尝试使用 dmidecode 获取内存序列号
		if cmd := exec.Command("dmidecode", "-t", "memory"); cmd != nil {
			if out, err := cmd.CombinedOutput(); err == nil {
				lines := strings.Split(string(out), "\n")
				var serials []string

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "Serial Number:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							serial := strings.TrimSpace(parts[1])
							if serial != "" && serial != "ToBeFilledByO.E.M." &&
								serial != "Not Specified" && serial != "NO DIMM" &&
								serial != "Unknown" && serial != "0000000000000000" {
								serials = append(serials, serial)
							}
						}
					}
				}

				// 返回第一个有效的内存序列号
				if len(serials) > 0 {
					return serials[0]
				}
			}
		}

		// 方法2: 尝试从 /proc/meminfo 获取内存信息作为备选
		if content, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(content), "\n")
			var memTotal string
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						memTotal = parts[1]
						break
					}
				}
			}

			// 如果找到内存总量，结合系统信息生成唯一标识
			if memTotal != "" {
				if hostname, err := os.Hostname(); err == nil {
					return fmt.Sprintf("MEM-%s-%s", memTotal, hostname)
				}
				return fmt.Sprintf("MEM-%s", memTotal)
			}
		}

		// 方法3: 尝试读取 /sys/devices/system/memory/ 下的信息
		if files, err := os.ReadDir("/sys/devices/system/memory/"); err == nil {
			var memoryBlocks []string
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "memory") {
					memoryBlocks = append(memoryBlocks, file.Name())
				}
			}

			if len(memoryBlocks) > 0 {
				// 使用内存块数量和第一个块名称生成标识
				return fmt.Sprintf("MEMBLK-%d-%s", len(memoryBlocks), memoryBlocks[0])
			}
		}

		return ""
	default:
		return ""
	}
}

// 取机器码UUID
func GetMachineCode() string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("wmic", "csproduct", "get", "uuid")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ""
		}
		str := string(out)
		reg := regexp.MustCompile(`\s+`)
		str = reg.ReplaceAllString(str, "")
		if len(str) > 4 {
			return str[4:]
		}
		return ""
	case "linux":
		// 方法1: 读取 /sys/devices/virtual/dmi/id/product_uuid
		if content, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_uuid"); err == nil {
			uuid := strings.TrimSpace(string(content))
			if uuid != "" && uuid != "00000000-0000-0000-0000-000000000000" {
				return uuid
			}
		}

		// 方法2: 读取 /sys/class/dmi/id/product_uuid
		if content, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
			uuid := strings.TrimSpace(string(content))
			if uuid != "" && uuid != "00000000-0000-0000-0000-000000000000" {
				return uuid
			}
		}

		// 方法3: 使用 dmidecode 获取系统UUID
		if cmd := exec.Command("dmidecode", "-s", "system-uuid"); cmd != nil {
			if out, err := cmd.CombinedOutput(); err == nil {
				uuid := strings.TrimSpace(string(out))
				if uuid != "" && uuid != "00000000-0000-0000-0000-000000000000" &&
					uuid != "Not Specified" && uuid != "To Be Filled By O.E.M." {
					return uuid
				}
			}
		}

		// 方法4: 读取 /etc/machine-id
		if content, err := os.ReadFile("/etc/machine-id"); err == nil {
			machineId := strings.TrimSpace(string(content))
			if machineId != "" && len(machineId) >= 16 {
				return machineId
			}
		}

		// 方法5: 读取 /var/lib/dbus/machine-id
		if content, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			machineId := strings.TrimSpace(string(content))
			if machineId != "" && len(machineId) >= 16 {
				return machineId
			}
		}

		// 方法6: 尝试从多个DMI信息组合生成唯一标识
		var identifiers []string

		// 读取系统制造商
		if content, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
			vendor := strings.TrimSpace(string(content))
			if vendor != "" && vendor != "To Be Filled By O.E.M." {
				identifiers = append(identifiers, vendor)
			}
		}

		// 读取产品名称
		if content, err := os.ReadFile("/sys/class/dmi/id/product_name"); err == nil {
			product := strings.TrimSpace(string(content))
			if product != "" && product != "To Be Filled By O.E.M." {
				identifiers = append(identifiers, product)
			}
		}

		// 读取产品序列号
		if content, err := os.ReadFile("/sys/class/dmi/id/product_serial"); err == nil {
			serial := strings.TrimSpace(string(content))
			if serial != "" && serial != "To Be Filled By O.E.M." && serial != "Not Specified" {
				identifiers = append(identifiers, serial)
			}
		}

		if len(identifiers) > 0 {
			combined := strings.Join(identifiers, "-")
			return fmt.Sprintf("COMBINED-%x", combined)[:32]
		}

		return ""
	default:
		return ""
	}
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
