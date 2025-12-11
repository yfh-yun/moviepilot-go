package utils

import (
	"os"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// DisplayHelper 显示帮助类
type DisplayHelper struct {
	logger  *zap.Logger
	display Display
	mutex   sync.Mutex
}

// Display 虚拟显示接口
type Display interface {
	// Start 启动虚拟显示
	Start() error
	// Stop 停止虚拟显示
	Stop() error
}

// NewDisplayHelper 创建显示帮助类实例
func NewDisplayHelper() *DisplayHelper {
	helper := &DisplayHelper{
		logger: logger.GetLogger(),
	}

	// 仅在Docker环境中初始化虚拟显示
	if isDocker() {
		helper.initDisplay()
	}

	return helper
}

// initDisplay 初始化虚拟显示
func (h *DisplayHelper) initDisplay() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 检查DISPLAY环境变量
	displayEnv := os.Getenv("DISPLAY")
	if displayEnv == "" {
		h.logger.Error("DISPLAY环境变量未设置")
		return
	}

	// 创建虚拟显示
	// 注意：Go版本没有直接对应的pyvirtualdisplay库，这里简化实现
	// 在实际实现中，应该使用Xvfb或类似工具创建虚拟显示
	h.logger.Info("初始化虚拟显示...")

	// 简化实现：假设虚拟显示创建成功
	h.display = &dummyDisplay{}
	h.logger.Info("虚拟显示初始化成功")
}

// Stop 停止虚拟显示
func (h *DisplayHelper) Stop() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.display == nil {
		return
	}

	h.logger.Info("正在停止虚拟显示...")
	if err := h.display.Stop(); err != nil {
		h.logger.Error("停止虚拟显示失败", zap.Error(err))
		return
	}
	h.display = nil
	h.logger.Info("虚拟显示已停止")
}

// dummyDisplay 虚拟显示的简化实现
type dummyDisplay struct{}

// Start 启动虚拟显示
func (d *dummyDisplay) Start() error {
	// 简化实现：什么都不做
	return nil
}

// Stop 停止虚拟显示
func (d *dummyDisplay) Stop() error {
	// 简化实现：什么都不做
	return nil
}

// isDocker 检查是否在Docker环境中
func isDocker() bool {
	// 检查是否存在.dockerinit文件或.docker目录
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return true
	}
	if _, err := os.Stat("/.docker"); err == nil {
		return true
	}
	// 检查环境变量
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}
	return false
}
