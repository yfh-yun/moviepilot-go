package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
)

// ModulesService 模块服务
type ModulesService struct {
	config *config.AppConfig
	logger *logger.LoggerManager
}

// NewModulesService 创建模块服务实例
func NewModulesService() *ModulesService {
	return &ModulesService{
		config: config.GetConfig(),
		logger: logger.GetLoggerManager(),
	}
}

// StartFrontend 启动前端服务
func (m *ModulesService) StartFrontend() {
	// 仅Windows可执行文件支持内嵌nginx
	if !m.isFrozen() || runtime.GOOS != "windows" {
		return
	}

	// 临时Nginx目录
	nginxPath := filepath.Join(m.config.GetRootPath(), "nginx")
	if _, err := os.Stat(nginxPath); os.IsNotExist(err) {
		return
	}

	// 配置目录下的Nginx目录
	runNginxDir := filepath.Join(m.config.GetConfigPath(), "nginx")
	if _, err := os.Stat(runNginxDir); os.IsNotExist(err) {
		// 移动到配置目�?		m.move(nginxPath, runNginxDir)
	}

	// 启动Nginx
	cmd := exec.Command("nginx.exe")
	cmd.Dir = runNginxDir
	cmd.Start()
}

// StopFrontend 停止前端服务
func (m *ModulesService) StopFrontend() {
	// 仅Windows可执行文件支持内嵌nginx
	if !m.isFrozen() || runtime.GOOS != "windows" {
		return
	}

	// 停止Nginx进程
	cmd := exec.Command("taskkill", "/f", "/im", "nginx.exe")
	cmd.Run()
}

// ClearTemp 清理临时文件和图片缓�?func (m *ModulesService) ClearTemp() {
	// 清理临时目录中指定天数前的文�?	m.clear(m.config.GetTempPath(), m.config.TempFileDays)
	// 清理图片缓存目录中指定天数前的文�?	imageCachePath := filepath.Join(m.config.GetCachePath(), "images")
	m.clear(imageCachePath, m.config.GlobalImageCacheDays)
}

// UserAuth 用户认证检�?func (m *ModulesService) UserAuth() {
	// TODO: 实现用户认证检查逻辑
	// 这需要SitesHelper、SystemConfigOper等组件的Go版本实现
	m.logger.Info("用户认证检�?..")
}

// CheckAuth 检查认证状�?func (m *ModulesService) CheckAuth() {
	// TODO: 实现认证状态检查逻辑
	// 这需要SitesHelper等组件的Go版本实现
	m.logger.Info("检查认证状�?..")
}

// StopModules 服务关闭
func (m *ModulesService) StopModules() {
	m.logger.Info("停止模块服务...")

	// 停止模块
	// ModuleManager().stop()

	// 停止事件消费
	// EventManager().stop()

	// 停止虚拟显示
	// DisplayHelper().stop()

	// 停止线程�?	// ThreadHelper().shutdown()

	// 停止消息服务
	// stop_message()

	// 关闭Redis缓存连接
	// RedisHelper().close()
	// AsyncRedisHelper().close()

	// 停止数据库连�?	// close_database()

	// 停止前端服务
	m.StopFrontend()

	// 清理临时文件
	m.ClearTemp()

	m.logger.Info("模块服务已停�?)
}

// InitModules 启动模块
func (m *ModulesService) InitModules() {
	m.logger.Info("初始化模块服�?..")

	// 虚拟显示
	// DisplayHelper()

	// DoH
	// DohHelper()

	// 站点管理
	// SitesHelper()

	// 资源包检�?	// ResourceHelper()

	// 用户认证
	m.UserAuth()

	// 加载模块
	// ModuleManager()

	// 启动事件消费
	// EventManager().start()

	// 初始化订阅分�?	// SubscribeHelper()

	// 启动前端服务
	m.StartFrontend()

	// 检查认证状�?	m.CheckAuth()

	m.logger.Info("模块服务初始化完�?)
}

// isFrozen 检查是否为可执行文件（类似Python的SystemUtils.is_frozen()�?func (m *ModulesService) isFrozen() bool {
	// 在Go中，通常通过构建标签或环境变量来判断
	// 这里简化处理，返回false
	// 可以通过检查是否是编译后的可执行文件来判断
	execPath, _ := os.Executable()
	// 如果是编译后的可执行文件，通常没有 .go 扩展�?	return filepath.Ext(execPath) != ".go"
}

// move 移动文件或目录（类似Python的SystemUtils.move�?func (m *ModulesService) move(src, dst string) error {
	return os.Rename(src, dst)
}

// clear 清理指定目录中指定天数前的文件（类似Python的SystemUtils.clear�?func (m *ModulesService) clear(path string, days int) {
	// 检查目录是否存�?	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	// 确保days有默认�?	if days <= 0 {
		days = 3 // 默认3�?	}

	// 遍历目录中的文件
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算文件修改时间与当前时间的差�?		diff := time.Since(info.ModTime())
		if diff > time.Duration(days)*24*time.Hour {
			// 删除文件
			os.Remove(filePath)
		}

		return nil
	})
}
