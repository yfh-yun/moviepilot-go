package core

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// PluginManager 插件管理器接�?type PluginManager interface {
	Sync() []interface{}
	InstallPluginMissingDependencies() []interface{}
	InitConfig()
	Start()
	Stop()
	StopMonitor()
}

// PluginAPI 注册插件API的接�?type PluginAPI interface {
	RegisterPluginAPI()
}

// PluginsInitializer 插件初始化器
type PluginsInitializer struct {
	logger        *zap.Logger
	pluginManager PluginManager
	pluginAPI     PluginAPI
}

// NewPluginsInitializer 创建新的插件初始化器
func NewPluginsInitializer(logger *zap.Logger, pluginManager PluginManager, pluginAPI PluginAPI) *PluginsInitializer {
	return &PluginsInitializer{
		logger:        logger,
		pluginManager: pluginManager,
		pluginAPI:     pluginAPI,
	}
}

// SyncPlugins 初始化安装插件，并动态注册后台任务及API
func (pi *PluginsInitializer) SyncPlugins() bool {
	pi.logger.Info("开始同步插�?..")

	defer func() {
		if r := recover(); r != nil {
			pi.logger.Error("插件初始化过程中出现异常", zap.Any("error", r))
		}
	}()

	// 执行插件同步任务
	syncResult := pi.executeTask(pi.pluginManager.Sync, "插件同步到本�?)
	
	// 执行缺失依赖项安装任�?	resolvedDependencies := pi.executeTask(pi.pluginManager.InstallPluginMissingDependencies, "缺失依赖项安�?)

	// 判断是否需要进行插件初始化
	if len(syncResult) == 0 && len(resolvedDependencies) == 0 {
		pi.logger.Debug("没有新的插件同步到本地或缺失依赖项需要安�?)
		return false
	}

	// 继续执行后续的插件初始化步骤
	pi.logger.Info("正在重新初始化插�?)
	
	// 重新初始化插�?	pi.pluginManager.InitConfig()
	
	// 重新注册插件API
	pi.RegisterPluginAPI()
	
	pi.logger.Info("所有插件初始化完成")
	return true
}

// executeTask 执行后台任务
func (pi *PluginsInitializer) executeTask(taskFunc func() []interface{}, taskName string) []interface{} {
	defer func() {
		if r := recover(); r != nil {
			pi.logger.Error(fmt.Sprintf("%s 时发生错�?, taskName), zap.Any("error", r))
		}
	}()

	result := taskFunc()
	if len(result) > 0 {
		pi.logger.Debug(fmt.Sprintf("%s 已完成，共处�?%d 个项�?, taskName, len(result)))
	} else {
		pi.logger.Debug(fmt.Sprintf("没有新的 %s 需要处�?, taskName))
	}
	
	return result
}

// RegisterPluginAPI 插件启动后注册插件API
func (pi *PluginsInitializer) RegisterPluginAPI() {
	// TODO: 实现插件API注册逻辑
	// 这里需要根据Go项目的API结构进行实现
	pi.logger.Info("注册插件API")
	if pi.pluginAPI != nil {
		pi.pluginAPI.RegisterPluginAPI()
	}
}

// InitPlugins 初始化插�?func (pi *PluginsInitializer) InitPlugins() {
	pi.logger.Info("初始化插�?)
	
	// 启动插件管理�?	pi.pluginManager.Start()
	
	// 注册插件API
	pi.RegisterPluginAPI()
}

// StopPlugins 停止插件
func (pi *PluginsInitializer) StopPlugins() {
	pi.logger.Info("停止插件")
	
	defer func() {
		if r := recover(); r != nil {
			pi.logger.Error("停止插件时发生错�?, zap.Any("error", r))
		}
	}()

	// 停止插件管理�?	pi.pluginManager.Stop()
	pi.pluginManager.StopMonitor()
}

// Async version functions

// AsyncSyncPlugins 异步初始化安装插件，并动态注册后台任务及API
func (pi *PluginsInitializer) AsyncSyncPlugins() chan bool {
	resultChan := make(chan bool, 1)
	
	go func() {
		result := pi.SyncPlugins()
		resultChan <- result
	}()
	
	return resultChan
}

// AsyncExecuteTask 异步执行后台任务
func (pi *PluginsInitializer) AsyncExecuteTask(taskFunc func() []interface{}, taskName string) chan []interface{} {
	resultChan := make(chan []interface{}, 1)
	
	go func() {
		result := pi.executeTask(taskFunc, taskName)
		resultChan <- result
	}()
	
	return resultChan
}
