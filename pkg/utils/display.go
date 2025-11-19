package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// DisplayHelper 虚拟显示管理助手
type DisplayHelper struct {
	display     *VirtualDisplay
	isDocker    bool
	displaySize DisplaySize
	mutex       sync.RWMutex
}

// DisplaySize 显示尺寸
type DisplaySize struct {
	Width  int
	Height int
}

// VirtualDisplay 虚拟显示接口
type VirtualDisplay interface {
	Start() error
	Stop() error
	IsRunning() bool
}

// XvfbDisplay Xvfb虚拟显示实现
type XvfbDisplay struct {
	size      DisplaySize
	display   string
	running   bool
	mutex     sync.RWMutex
	cmd       *exec.Cmd
}

// NewDisplayHelper 创建显示助手实例
func NewDisplayHelper() *DisplayHelper {
	return &DisplayHelper{
		isDocker:    IsDockerEnvironment(),
		displaySize: DisplaySize{Width: 1024, Height: 768},
	}
}

// Start 启动虚拟显示
func (dh *DisplayHelper) Start() error {
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	if !dh.isDocker {
		return fmt.Errorf("not in docker environment, virtual display not needed")
	}

	if dh.display != nil && dh.display.IsRunning() {
		return fmt.Errorf("virtual display is already running")
	}

	display, err := dh.createVirtualDisplay()
	if err != nil {
		return fmt.Errorf("failed to create virtual display: %v", err)
	}

	if err := display.Start(); err != nil {
		return fmt.Errorf("failed to start virtual display: %v", err)
	}

	dh.display = display
	return nil
}

// Stop 停止虚拟显示
func (dh *DisplayHelper) Stop() error {
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	if dh.display == nil {
		return fmt.Errorf("virtual display is not running")
	}

	if err := dh.display.Stop(); err != nil {
		return fmt.Errorf("failed to stop virtual display: %v", err)
	}

	dh.display = nil
	return nil
}

// IsRunning 检查虚拟显示是否运行中
func (dh *DisplayHelper) IsRunning() bool {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	return dh.display != nil && dh.display.IsRunning()
}

// SetDisplaySize 设置显示尺寸
func (dh *DisplayHelper) SetDisplaySize(width, height int) {
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	dh.displaySize = DisplaySize{Width: width, Height: height}
}

// GetDisplaySize 获取显示尺寸
func (dh *DisplayHelper) GetDisplaySize() DisplaySize {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	return dh.displaySize
}

// createVirtualDisplay 创建虚拟显示实例
func (dh *DisplayHelper) createVirtualDisplay() (VirtualDisplay, error) {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":99"
	}

	return &XvfbDisplay{
		size:    dh.displaySize,
		display: display,
	}, nil
}

// Start 启动Xvfb虚拟显示
func (xd *XvfbDisplay) Start() error {
	xd.mutex.Lock()
	defer xd.mutex.Unlock()

	if xd.running {
		return fmt.Errorf("xvfb display is already running")
	}

	// 检查Xvfb是否可用
	if _, err := exec.LookPath("Xvfb"); err != nil {
		return fmt.Errorf("Xvfb not found: %v", err)
	}

	// 构建Xvfb命令
	args := []string{
		xd.display,
		"-screen", "0",
		fmt.Sprintf("%dx%dx24", xd.size.Width, xd.size.Height),
		"-ac",
		"+extension", "GLX",
		"+render", "-noreset",
	}

	xd.cmd = exec.Command("Xvfb", args...)
	
	if err := xd.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Xvfb: %v", err)
	}

	xd.running = true
	return nil
}

// Stop 停止Xvfb虚拟显示
func (xd *XvfbDisplay) Stop() error {
	xd.mutex.Lock()
	defer xd.mutex.Unlock()

	if !xd.running {
		return fmt.Errorf("xvfb display is not running")
	}

	if xd.cmd != nil && xd.cmd.Process != nil {
		if err := xd.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill Xvfb process: %v", err)
		}
	}

	xd.running = false
	xd.cmd = nil
	return nil
}

// IsRunning 检查Xvfb是否运行中
func (xd *XvfbDisplay) IsRunning() bool {
	xd.mutex.RLock()
	defer xd.mutex.RUnlock()

	return xd.running
}

// IsDockerEnvironment 检查是否在Docker环境中
func IsDockerEnvironment() bool {
	// 检查/.dockerenv文件
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 检查cgroup信息
	if runtime.GOOS == "linux" {
		cgroupPath := "/proc/1/cgroup"
		if data, err := os.ReadFile(cgroupPath); err == nil {
			content := string(data)
			if contains(content, []string{"docker", "containerd"}) {
				return true
			}
		}
	}

	return false
}

// contains 检查字符串是否包含任一关键词
func contains(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

// GetDisplayEnv 获取显示环境变量
func (dh *DisplayHelper) GetDisplayEnv() string {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":99"
	}
	return display
}

// SetDisplayEnv 设置显示环境变量
func (dh *DisplayHelper) SetDisplayEnv(display string) error {
	if display == "" {
		display = ":99"
	}
	return os.Setenv("DISPLAY", display)
}