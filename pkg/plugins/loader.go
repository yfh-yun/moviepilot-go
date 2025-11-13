package plugins

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"plugin"
	"strings"
)

// PluginLoader 插件加载�?type PluginLoader struct {
	// 已加载的插件映射
	loadedPlugins map[string]Plugin
}

// NewPluginLoader 创建一个新的插件加载器实例
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		loadedPlugins: make(map[string]Plugin),
	}
}

// LoadPlugin 从指定路径加载插�?func (pl *PluginLoader) LoadPlugin(pluginPath string) (Plugin, error) {
	// 检查插件是否已经加�?	pluginName := filepath.Base(pluginPath)
	if strings.HasSuffix(pluginName, ".so") {
		pluginName = pluginName[:len(pluginName)-3]
	}
	
	if existingPlugin, exists := pl.loadedPlugins[pluginName]; exists {
		return existingPlugin, nil
	}
	
	// 打开插件
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %v", err)
	}
	
	// 查找插件构造函�?	symbol, err := plug.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("failed to find plugin constructor: %v", err)
	}
	
	// 类型断言为插件构造函�?	constructor, ok := symbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin constructor has wrong signature")
	}
	
	// 创建插件实例
	pluginInstance := constructor()
	
	// 保存到已加载插件映射�?	pl.loadedPlugins[pluginName] = pluginInstance
	
	return pluginInstance, nil
}

// LoadPluginsFromDirectory 从指定目录加载所有插�?func (pl *PluginLoader) LoadPluginsFromDirectory(dirPath string) (map[string]Plugin, error) {
	// 读取目录内容
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %v", err)
	}
	
	// 存储加载的插�?	plugins := make(map[string]Plugin)
	
	// 遍历目录中的文件
	for _, file := range files {
		// 只处�?so文件（编译后的Go插件�?		if !file.IsDir() && strings.HasSuffix(file.Name(), ".so") {
			pluginPath := filepath.Join(dirPath, file.Name())
			
			// 加载插件
			pluginInstance, err := pl.LoadPlugin(pluginPath)
			if err != nil {
				fmt.Printf("Warning: failed to load plugin %s: %v\n", file.Name(), err)
				continue
			}
			
			// 获取插件名称
			pluginName := pluginInstance.GetName()
			if pluginName == "" {
				pluginName = strings.TrimSuffix(file.Name(), ".so")
			}
			
			// 添加到结果映射中
			plugins[pluginName] = pluginInstance
		}
	}
	
	return plugins, nil
}

// GetPlugin 获取已加载的插件
func (pl *PluginLoader) GetPlugin(pluginName string) (Plugin, bool) {
	pluginInstance, exists := pl.loadedPlugins[pluginName]
	return pluginInstance, exists
}

// GetAllPlugins 获取所有已加载的插�?func (pl *PluginLoader) GetAllPlugins() map[string]Plugin {
	return pl.loadedPlugins
}

// UnloadPlugin 卸载插件
func (pl *PluginLoader) UnloadPlugin(pluginName string) {
	if pluginInstance, exists := pl.loadedPlugins[pluginName]; exists {
		// 停止插件服务
		pluginInstance.StopService()
		
		// 关闭插件
		pluginInstance.Close()
		
		// 从映射中删除
		delete(pl.loadedPlugins, pluginName)
	}
}

// UnloadAllPlugins 卸载所有插�?func (pl *PluginLoader) UnloadAllPlugins() {
	for pluginName := range pl.loadedPlugins {
		pl.UnloadPlugin(pluginName)
	}
}
