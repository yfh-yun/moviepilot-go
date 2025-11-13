package plugins

import (
	"fmt"
	"sort"
	"sync"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/sandbox"
)

// PluginManager 插件管理�?type PluginManager struct {
	// 插件加载�?	loader *PluginLoader
	
	// 插件映射
	plugins map[string]Plugin
	
	// 插件互斥�?	mutex sync.RWMutex
	
	// 插件服务（用于管理Python插件容器�?	pluginService *sandbox.PluginService
	
	// 已启用的插件列表
	enabledPlugins []string
}

// NewPluginManager 创建一个新的插件管理器实例
func NewPluginManager() *PluginManager {
	return &PluginManager{
		loader:         NewPluginLoader(),
		plugins:        make(map[string]Plugin),
		pluginService:  sandbox.NewPluginService(),
		enabledPlugins: make([]string, 0),
	}
}

// Initialize 初始化插件管理器
func (pm *PluginManager) Initialize() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// 初始化插件服�?	if err := pm.pluginService.Start(); err != nil {
		return fmt.Errorf("failed to start plugin service: %v", err)
	}
	
	// 加载所有插�?	if err := pm.loadAllPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %v", err)
	}
	
	// 初始化已启用的插�?	pm.initEnabledPlugins()
	
	return nil
}

// loadAllPlugins 加载所有插�?func (pm *PluginManager) loadAllPlugins() error {
	// 加载Go插件
	goPlugins, err := pm.loader.LoadPluginsFromDirectory(config.Settings.PluginPath)
	if err != nil {
		return fmt.Errorf("failed to load Go plugins: %v", err)
	}
	
	// 将加载的插件添加到插件映射中
	for name, pluginInstance := range goPlugins {
		pm.plugins[name] = pluginInstance
	}
	
	// TODO: 加载Python插件（通过插件服务�?	
	return nil
}

// initEnabledPlugins 初始化已启用的插�?func (pm *PluginManager) initEnabledPlugins() {
	// 从配置中获取已启用的插件列表
	// 这里只是一个示例实现，实际应该从系统配置中读取
	pm.enabledPlugins = make([]string, 0)
	
	// 初始化每个已启用的插�?	for _, pluginName := range pm.enabledPlugins {
		if pluginInstance, exists := pm.plugins[pluginName]; exists {
			// 获取插件配置
			config := pluginInstance.GetConfig(nil)
			
			// 初始化插�?			pluginInstance.InitPlugin(config.(map[string]interface{}))
		}
	}
}

// GetPlugin 获取插件实例
func (pm *PluginManager) GetPlugin(pluginName string) (Plugin, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	pluginInstance, exists := pm.plugins[pluginName]
	return pluginInstance, exists
}

// GetAllPlugins 获取所有插�?func (pm *PluginManager) GetAllPlugins() map[string]Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	// 创建副本以避免外部修�?	plugins := make(map[string]Plugin)
	for name, pluginInstance := range pm.plugins {
		plugins[name] = pluginInstance
	}
	
	return plugins
}

// GetEnabledPlugins 获取已启用的插件列表
func (pm *PluginManager) GetEnabledPlugins() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	// 创建副本
	enabled := make([]string, len(pm.enabledPlugins))
	copy(enabled, pm.enabledPlugins)
	
	return enabled
}

// EnablePlugin 启用插件
func (pm *PluginManager) EnablePlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// 检查插件是否存�?	pluginInstance, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}
	
	// 检查插件是否已经启�?	for _, name := range pm.enabledPlugins {
		if name == pluginName {
			return fmt.Errorf("plugin already enabled: %s", pluginName)
		}
	}
	
	// 获取插件配置
	config := pluginInstance.GetConfig(nil)
	
	// 初始化插�?	pluginInstance.InitPlugin(config.(map[string]interface{}))
	
	// 添加到已启用插件列表
	pm.enabledPlugins = append(pm.enabledPlugins, pluginName)
	
	return nil
}

// DisablePlugin 禁用插件
func (pm *PluginManager) DisablePlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// 检查插件是否存�?	pluginInstance, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}
	
	// 查找插件在已启用列表中的位置
	index := -1
	for i, name := range pm.enabledPlugins {
		if name == pluginName {
			index = i
			break
		}
	}
	
	// 如果插件未启用，直接返回
	if index == -1 {
		return fmt.Errorf("plugin not enabled: %s", pluginName)
	}
	
	// 停止插件服务
	pluginInstance.StopService()
	
	// 从已启用插件列表中移�?	pm.enabledPlugins = append(pm.enabledPlugins[:index], pm.enabledPlugins[index+1:]...)
	
	return nil
}

// GetPluginCommands 获取所有插件的命令
func (pm *PluginManager) GetPluginCommands() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	var commands []map[string]interface{}
	
	// 遍历所有已启用的插�?	for _, pluginName := range pm.enabledPlugins {
		if pluginInstance, exists := pm.plugins[pluginName]; exists {
			pluginCommands := pluginInstance.GetCommand()
			if pluginCommands != nil {
				commands = append(commands, pluginCommands...)
			}
		}
	}
	
	return commands
}

// GetPluginAPIs 获取所有插件的API
func (pm *PluginManager) GetPluginAPIs() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	var apis []map[string]interface{}
	
	// 遍历所有已启用的插�?	for _, pluginName := range pm.enabledPlugins {
		if pluginInstance, exists := pm.plugins[pluginName]; exists {
			pluginAPIs := pluginInstance.GetAPI()
			if pluginAPIs != nil {
				apis = append(apis, pluginAPIs...)
			}
		}
	}
	
	return apis
}

// GetPluginServices 获取所有插件的服务
func (pm *PluginManager) GetPluginServices() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	var services []map[string]interface{}
	
	// 遍历所有已启用的插�?	for _, pluginName := range pm.enabledPlugins {
		if pluginInstance, exists := pm.plugins[pluginName]; exists {
			pluginServices := pluginInstance.GetService()
			if pluginServices != nil {
				services = append(services, pluginServices...)
			}
		}
	}
	
	return services
}

// GetSortedPlugins 获取按顺序排序的插件列表
func (pm *PluginManager) GetSortedPlugins() []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	// 创建插件切片
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, pluginInstance := range pm.plugins {
		plugins = append(plugins, pluginInstance)
	}
	
	// 按插件顺序排�?	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].GetOrder() < plugins[j].GetOrder()
	})
	
	return plugins
}

// LoadPythonPlugin 加载Python插件到独立容器中
func (pm *PluginManager) LoadPythonPlugin(pluginPath string) error {
	// 使用插件服务加载Python插件
	_, err := pm.pluginService.LoadPlugin(pluginPath)
	return err
}

// ExecutePythonPlugin 执行Python插件中的操作
func (pm *PluginManager) ExecutePythonPlugin(containerID, action string, params map[string]interface{}) (interface{}, error) {
	response, err := pm.pluginService.ExecutePlugin(containerID, action, params)
	if err != nil {
		return nil, err
	}
	
	if !response.Success {
		return nil, fmt.Errorf("plugin execution failed: %s", response.Message)
	}
	
	return response.Data, nil
}

// UnloadPythonPlugin 卸载Python插件容器
func (pm *PluginManager) UnloadPythonPlugin(containerID string) error {
	return pm.pluginService.UnloadPlugin(containerID)
}

// Shutdown 关闭插件管理�?func (pm *PluginManager) Shutdown() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// 停止插件服务
	if err := pm.pluginService.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin service: %v", err)
	}
	
	// 卸载所有插�?	pm.loader.UnloadAllPlugins()
	
	return nil
}
