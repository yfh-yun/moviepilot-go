package utils

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// SystemHelper 系统工具类
type SystemHelper struct {
	systemFlagFile string
	eventHandlers  map[string]func(event interface{})
	mutex          sync.RWMutex
}

// SystemEvent 系统事件
type SystemEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ConfigChangeEventData 配置变更事件数据
type ConfigChangeEventData struct {
	Key   string      `json:"key"`
	Old   interface{} `json:"old"`
	New   interface{} `json:"new"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string    `json:"os"`
	Arch         string    `json:"arch"`
	Hostname     string    `json:"hostname"`
	CPUCount     int       `json:"cpu_count"`
	MemoryTotal  uint64    `json:"memory_total"`
	MemoryUsed   uint64    `json:"memory_used"`
	DiskTotal    uint64    `json:"disk_total"`
	DiskUsed     uint64    `json:"disk_used"`
	Uptime       time.Time `json:"uptime"`
	IsDocker     bool      `json:"is_docker"`
	IsFrozen     bool      `json:"is_frozen"`
}

// NewSystemHelper 创建系统助手实例
func NewSystemHelper() *SystemHelper {
	return &SystemHelper{
		systemFlagFile: "/var/log/nginx/__moviepilot__",
		eventHandlers:  make(map[string]func(event interface{})),
	}
}

// CanRestart 判断是否可以内部重启
func (sh *SystemHelper) CanRestart() bool {
	// 检查Docker socket是否存在
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return true
	}

	// 检查Docker客户端API
	dockerClientAPI := os.Getenv("DOCKER_CLIENT_API")
	if dockerClientAPI != "" && dockerClientAPI != "tcp://127.0.0.1:38379" {
		return true
	}

	return false
}

// Restart 重启系统
func (sh *SystemHelper) Restart() error {
	if !sh.CanRestart() {
		return fmt.Errorf("system restart is not supported")
	}

	// 发送重启信号
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	// 发送SIGTERM信号
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send restart signal: %v", err)
	}

	return nil
}

// Shutdown 关闭系统
func (sh *SystemHelper) Shutdown() error {
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	// 发送SIGINT信号
	if err := process.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("failed to send shutdown signal: %v", err)
	}

	return nil
}

// GetSystemInfo 获取系统信息
func (sh *SystemHelper) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
		IsDocker: IsDockerEnvironment(),
		IsFrozen: IsFrozenEnvironment(),
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// 获取内存信息
	if memInfo, err := sh.getMemoryInfo(); err == nil {
		info.MemoryTotal = memInfo.Total
		info.MemoryUsed = memInfo.Used
	}

	// 获取磁盘信息
	if diskInfo, err := sh.getDiskInfo(); err == nil {
		info.DiskTotal = diskInfo.Total
		info.DiskUsed = diskInfo.Used
	}

	// 获取启动时间
	if uptime, err := sh.getUptime(); err == nil {
		info.Uptime = uptime
	}

	return info, nil
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// getMemoryInfo 获取内存信息
func (sh *SystemHelper) getMemoryInfo() (*MemoryInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("memory info not supported on %s", runtime.GOOS)
	}

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	var memTotal, memAvailable uint64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memTotal = val * 1024 // 转换为字节
			}
		case "MemAvailable:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memAvailable = val * 1024 // 转换为字节
			}
		}
	}

	if memTotal == 0 {
		return nil, fmt.Errorf("failed to parse memory info")
	}

	return &MemoryInfo{
		Total: memTotal,
		Used:  memTotal - memAvailable,
		Free:  memAvailable,
	}, nil
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// getDiskInfo 获取磁盘信息
func (sh *SystemHelper) getDiskInfo() (*DiskInfo, error) {
	// 获取当前目录的磁盘信息
	stat := syscall.Statfs_t{}
	if err := syscall.Statfs(".", &stat); err != nil {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return &DiskInfo{
		Total: total,
		Used:  used,
		Free:  free,
	}, nil
}

// getUptime 获取系统启动时间
func (sh *SystemHelper) getUptime() (time.Time, error) {
	if runtime.GOOS != "linux" {
		return time.Time{}, fmt.Errorf("uptime not supported on %s", runtime.GOOS)
	}

	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read /proc/uptime: %v", err)
	}

	parts := strings.Fields(string(data))
	if len(parts) < 1 {
		return time.Time{}, fmt.Errorf("invalid uptime format")
	}

	seconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse uptime: %v", err)
	}

	uptime := time.Duration(seconds) * time.Second
	return time.Now().Add(-uptime), nil
}

// RegisterEventHandler 注册事件处理器
func (sh *SystemHelper) RegisterEventHandler(eventType string, handler func(event interface{})) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.eventHandlers[eventType] = handler
}

// UnregisterEventHandler 注销事件处理器
func (sh *SystemHelper) UnregisterEventHandler(eventType string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	delete(sh.eventHandlers, eventType)
}

// TriggerEvent 触发事件
func (sh *SystemHelper) TriggerEvent(event *SystemEvent) {
	sh.mutex.RLock()
	handler, exists := sh.eventHandlers[event.Type]
	sh.mutex.RUnlock()

	if exists {
		handler(event)
	}
}

// HandleConfigChanged 处理配置变更事件
func (sh *SystemHelper) HandleConfigChanged(eventData *ConfigChangeEventData) {
	// 检查是否为日志相关配置变更
	logKeys := []string{
		"DEBUG",
		"LOG_LEVEL",
		"LOG_MAX_FILE_SIZE",
		"LOG_BACKUP_COUNT",
		"LOG_FILE_FORMAT",
		"LOG_CONSOLE_FORMAT",
	}

	for _, key := range logKeys {
		if eventData.Key == key {
			// 触发日志更新事件
			event := &SystemEvent{
				Type:      "log_config_changed",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"config_change": eventData,
				},
			}
			sh.TriggerEvent(event)
			break
		}
	}
}

// SetupSignalHandlers 设置信号处理器
func (sh *SystemHelper) SetupSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				event := &SystemEvent{
					Type:      "shutdown",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"signal": sig.String(),
					},
				}
				sh.TriggerEvent(event)
			case syscall.SIGHUP:
				event := &SystemEvent{
					Type:      "reload",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"signal": sig.String(),
					},
				}
				sh.TriggerEvent(event)
			}
		}
	}()
}

// GetEnvironmentInfo 获取环境信息
func (sh *SystemHelper) GetEnvironmentInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// 环境变量
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}
	info["environment"] = envVars

	// 运行时信息
	info["runtime"] = map[string]interface{}{
		"go_version": runtime.Version(),
		"go_os":      runtime.GOOS,
		"go_arch":    runtime.GOARCH,
		"cpu_count":  runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
	}

	// 进程信息
	info["process"] = map[string]interface{}{
		"pid":  os.Getpid(),
		"ppid": os.Getppid(),
		"uid":  os.Getuid(),
		"gid":  os.Getgid(),
	}

	return info
}

// ExecuteCommand 执行系统命令
func (sh *SystemHelper) ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// ExecuteCommandWithTimeout 执行带超时的系统命令
func (sh *SystemHelper) ExecuteCommandWithTimeout(timeout time.Duration, command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", timeout)
		}
		return "", fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// GetProcessInfo 获取进程信息
func (sh *SystemHelper) GetProcessInfo(pid int) (*ProcessInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("process info not supported on %s", runtime.GOOS)
	}

	// 读取进程状态文件
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read process stat: %v", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 24 {
		return nil, fmt.Errorf("invalid process stat format")
	}

	// 解析进程信息
	info := &ProcessInfo{
		PID: pid,
	}

	// 进程名称（可能包含空格和括号）
	if len(fields) > 1 {
		name := fields[1]
		if len(name) > 2 && name[0] == '(' && name[len(name)-1] == ')' {
			info.Name = name[1 : len(name)-1]
		} else {
			info.Name = name
		}
	}

	// 状态
	if len(fields) > 2 {
		info.Status = fields[2]
	}

	// 父进程ID
	if len(fields) > 3 {
		if ppid, err := strconv.Atoi(fields[3]); err == nil {
			info.PPID = ppid
		}
	}

	// 启动时间
	if len(fields) > 21 {
		if starttime, err := strconv.ParseUint(fields[21], 10, 64); err == nil {
			// 转换为实际时间
			info.StartTime = time.Unix(int64(starttime)/100, 0)
		}
	}

	return info, nil
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID       int       `json:"pid"`
	PPID      int       `json:"ppid"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
}

// IsProcessRunning 检查进程是否运行
func (sh *SystemHelper) IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送信号0检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// KillProcess 杀死进程
func (sh *SystemHelper) KillProcess(pid int, signal syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	return process.Signal(signal)
}

// GetSystemLoad 获取系统负载
func (sh *SystemHelper) GetSystemLoad() (*SystemLoad, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("system load not supported on %s", runtime.GOOS)
	}

	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/loadavg: %v", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid loadavg format")
	}

	load := &SystemLoad{}
	if load1, err := strconv.ParseFloat(fields[0], 64); err == nil {
		load.Load1 = load1
	}
	if load5, err := strconv.ParseFloat(fields[1], 64); err == nil {
		load.Load5 = load5
	}
	if load15, err := strconv.ParseFloat(fields[2], 64); err == nil {
		load.Load15 = load15
	}

	return load, nil
}

// SystemLoad 系统负载
type SystemLoad struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// GetNetworkInterfaces 获取网络接口信息
func (sh *SystemHelper) GetNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var addresses []string
		for _, addr := range addrs {
			addresses = append(addresses, addr.String())
		}

		result = append(result, NetworkInterface{
			Name:      iface.Name,
			Index:     iface.Index,
			MTU:       iface.MTU,
			Flags:     iface.Flags.String(),
			Addresses: addresses,
		})
	}

	return result, nil
}

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name      string   `json:"name"`
	Index     int      `json:"index"`
	MTU       int      `json:"mtu"`
	Flags     string   `json:"flags"`
	Addresses []string `json:"addresses"`
}

// GetSystemFlagFile 获取系统标志文件
func (sh *SystemHelper) GetSystemFlagFile() string {
	return sh.systemFlagFile
}

// SetSystemFlagFile 设置系统标志文件
func (sh *SystemHelper) SetSystemFlagFile(file string) {
	sh.systemFlagFile = file
}

// CheckSystemFlag 检查系统标志
func (sh *SystemHelper) CheckSystemFlag() bool {
	if sh.systemFlagFile == "" {
		return false
	}

	_, err := os.Stat(sh.systemFlagFile)
	return err == nil
}

// CreateSystemFlag 创建系统标志
func (sh *SystemHelper) CreateSystemFlag() error {
	if sh.systemFlagFile == "" {
		return fmt.Errorf("system flag file not set")
	}

	// 确保目录存在
	dir := filepath.Dir(sh.systemFlagFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 创建标志文件
	file, err := os.Create(sh.systemFlagFile)
	if err != nil {
		return fmt.Errorf("failed to create flag file: %v", err)
	}
	defer file.Close()

	return nil
}

// RemoveSystemFlag 移除系统标志
func (sh *SystemHelper) RemoveSystemFlag() error {
	if sh.systemFlagFile == "" {
		return fmt.Errorf("system flag file not set")
	}

	return os.Remove(sh.systemFlagFile)
}

// IsFrozenEnvironment 检查是否在冻结环境中运行
func IsFrozenEnvironment() bool {
	// 检查冻结标志文件
	if _, err := os.Stat("/tmp/moviepilot_frozen"); err == nil {
		return true
	}

	// 检查环境变量
	if os.Getenv("MOVIEPILOT_FROZEN") == "true" {
		return true
	}

	return false
}