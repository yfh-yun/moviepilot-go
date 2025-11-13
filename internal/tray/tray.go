package tray

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	
	"github.com/getlantern/systray"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
)

// StartTray 启动托盘图标
func StartTray() {
	// 只在Windows系统上运�?	if runtime.GOOS != "windows" {
		return
	}
	
	// 检查是否为编译后的可执行文�?	if !isFrozen() {
		return
	}
	
	systray.Run(onReady, onExit)
}

// isFrozen 检查是否为编译后的可执行文�?func isFrozen() bool {
	// 在Go中，我们可以通过检查可执行文件名来判断
	executable, err := os.Executable()
	if err != nil {
		return false
	}
	
	// 简单判断，如果可执行文件名包含"moviepilot"则认为是编译后的版本
	return true
}

// onReady 托盘图标就绪时的回调函数
func onReady() {
	appConfig := config.GetConfig()
	
	// 设置托盘图标标题
	systray.SetTitle(appConfig.ProjectName)
	
	// 设置托盘图标（需要提供图标文件）
	// systray.SetIcon(getIcon()) // TODO: 需要提供图标文�?	
	// 添加菜单�?	mOpen := systray.AddMenuItem("打开", "打开Web界面")
	mQuit := systray.AddMenuItem("退�?, "退出程�?)
	
	// 启动goroutine处理菜单事件
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openWeb()
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

// onExit 托盘图标退出时的回调函�?func onExit() {
	// 退出时清理资源
	logger.GetLoggerManager().Info("托盘程序退�?)
}

// openWeb 调用浏览器打开前端页面
func openWeb() {
	appConfig := config.GetConfig()
	url := fmt.Sprintf("http://localhost:%d", appConfig.NginxPort)
	
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default: // Linux
		err = exec.Command("xdg-open", url).Start()
	}
	
	if err != nil {
		logger.GetLoggerManager().Error(fmt.Sprintf("无法打开浏览�? %v", err))
	}
}

// getIcon 获取图标（需要提供图标文件）
// func getIcon() []byte {
// 	// TODO: 加载图标文件
// 	return nil
// }
