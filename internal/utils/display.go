package utils

import (
	"os"
	"os/exec"
	"sync"

	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// DisplayHelper 虚拟显示助手结构�?type DisplayHelper struct {
	display *exec.Cmd
	mutex   sync.Mutex
}

// displayHelperInstance DisplayHelper单例实例
var displayHelperInstance *DisplayHelper
var displayHelperOnce sync.Once

// NewDisplayHelper 创建DisplayHelper单例实例
func NewDisplayHelper() *DisplayHelper {
	displayHelperOnce.Do(func() {
		displayHelperInstance = &DisplayHelper{}
		displayHelperInstance.init()
	})
	return displayHelperInstance
}

// init 初始化虚拟显�?func (dh *DisplayHelper) init() {
	/*
		初始化虚拟显�?	*/
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	// 检查是否为Docker环境
	if !dh.isDocker() {
		return
	}

	// 尝试启动虚拟显示
	displayEnv := os.Getenv("DISPLAY")
	if displayEnv == "" {
		displayEnv = ":0"
	}

	// 启动Xvfb虚拟显示服务�?	// 注意：这需要系统中已安装Xvfb
	dh.display = exec.Command("Xvfb", displayEnv, "-screen", "0", "1024x768x24")
	
	err := dh.display.Start()
	if err != nil {
		logger.GetLoggerManager().Error("DisplayHelper init error", zap.Error(err))
		dh.display = nil
		return
	}

	// 设置环境变量
	os.Setenv("DISPLAY", displayEnv)
	logger.GetLoggerManager().Info("虚拟显示已启�?, zap.String("display", displayEnv))
}

// isDocker 判断是否为Docker环境
func (dh *DisplayHelper) isDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}

// Stop 停止虚拟显示
func (dh *DisplayHelper) Stop() {
	/*
		停止虚拟显示
	*/
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	if dh.display != nil {
		logger.GetLoggerManager().Info("正在停止虚拟显示...")
		err := dh.display.Process.Kill()
		if err != nil {
			logger.GetLoggerManager().Error("停止虚拟显示失败", zap.Error(err))
		} else {
			logger.GetLoggerManager().Info("虚拟显示已停�?)
		}
		dh.display = nil
	}
}
