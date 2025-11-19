package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// NativePluginManager Go原生插件管理器
type NativePluginManager struct {
	pluginPath string
	plugins    map[string]*NativePlugin
	mutex      sync.RWMutex
	logger     Logger
}

// NativePlugin 原生插件包装器
type NativePlugin struct {
	pluginPath string
	instance   Plugin
	state      PluginState
	logger     Logger
}

// NewNativePluginManager 创建原生插件管理器
func NewNativePluginManager(pluginPath string) (*NativePluginManager, error) {
	// 确保插件目录存在
	if err := os.MkdirAll(pluginPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	return &NativePluginManager{
		pluginPath: pluginPath,
		plugins:    make(map[string]*NativePlugin),
		logger:     NewLogger(),
	}, nil
}

// LoadPlugin 加载原生插件
func (npm *NativePluginManager) LoadPlugin(pluginPath string) (Plugin, error) {
	npm.mutex.Lock()
	defer npm.mutex.Unlock()

	// 解析绝对路径
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", absPath)
	}

	// 加载.so文件
	goPlugin, err := plugin.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// 查找NewPlugin函数
	newPluginSymbol, err := goPlugin.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export NewPlugin function: %w", err)
	}

	// 类型断言
	newPlugin, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("NewPlugin has wrong signature")
	}

	// 创建插件实例
	instance := newPlugin()

	// 创建包装器
	wrapper := &NativePlugin{
		pluginPath: absPath,
		instance:   instance,
		state:      StateLoaded,
		logger:     npm.logger,
	}

	// 存储插件
	npm.plugins[instance.ID()] = wrapper

	npm.logger.Info("Native plugin loaded", "id", instance.ID(), "path", absPath)

	return wrapper, nil
}

// UnloadPlugin 卸载原生插件
func (npm *NativePluginManager) UnloadPlugin(pluginID string) error {
	npm.mutex.Lock()
	defer npm.mutex.Unlock()

	wrapper, exists := npm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止插件
	if wrapper.state == StateRunning {
		if err := wrapper.Stop(); err != nil {
			npm.logger.Error("Failed to stop plugin before unloading", "id", pluginID, "error", err)
		}
	}

	// 销毁插件
	if err := wrapper.Destroy(); err != nil {
		npm.logger.Error("Failed to destroy plugin", "id", pluginID, "error", err)
	}

	// 从内存中移除
	delete(npm.plugins, pluginID)

	npm.logger.Info("Native plugin unloaded", "id", pluginID)

	return nil
}

// GetPlugin 获取插件
func (npm *NativePluginManager) GetPlugin(pluginID string) (Plugin, error) {
	npm.mutex.RLock()
	wrapper, exists := npm.plugins[pluginID]
	npm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return wrapper, nil
}

// ListPlugins 列出所有插件
func (npm *NativePluginManager) ListPlugins() []Plugin {
	npm.mutex.RLock()
	defer npm.mutex.RUnlock()

	var plugins []Plugin
	for _, wrapper := range npm.plugins {
		plugins = append(plugins, wrapper)
	}

	return plugins
}

// 实现Plugin接口的方法
func (np *NativePlugin) ID() string {
	return np.instance.ID()
}

func (np *NativePlugin) Name() string {
	return np.instance.Name()
}

func (np *NativePlugin) Version() string {
	return np.instance.Version()
}

func (np *NativePlugin) Type() PluginType {
	return PluginTypeNative
}

func (np *NativePlugin) Description() string {
	return np.instance.Description()
}

func (np *NativePlugin) Initialize(config map[string]interface{}) error {
	if np.state != StateLoaded {
		return fmt.Errorf("plugin not in loaded state")
	}

	if err := np.instance.Initialize(config); err != nil {
		np.state = StateError
		return err
	}

	np.state = StateInitialized
	np.logger.Debug("Native plugin initialized", "id", np.ID())

	return nil
}

func (np *NativePlugin) Start() error {
	if np.state != StateInitialized {
		return fmt.Errorf("plugin not initialized")
	}

	if err := np.instance.Start(); err != nil {
		np.state = StateError
		return err
	}

	np.state = StateRunning
	np.logger.Info("Native plugin started", "id", np.ID())

	return nil
}

func (np *NativePlugin) Stop() error {
	if np.state != StateRunning {
		return fmt.Errorf("plugin not running")
	}

	if err := np.instance.Stop(); err != nil {
		return err
	}

	np.state = StateStopped
	np.logger.Info("Native plugin stopped", "id", np.ID())

	return nil
}

func (np *NativePlugin) Destroy() error {
	if err := np.instance.Destroy(); err != nil {
		return err
	}

	np.state = StateUnloaded
	np.logger.Debug("Native plugin destroyed", "id", np.ID())

	return nil
}

func (np *NativePlugin) GetState() PluginState {
	return np.state
}

func (np *NativePlugin) HandleEvent(event Event) error {
	if handler, ok := np.instance.(EventHandler); ok {
		return handler.HandleEvent(event)
	}
	return nil
}

func (np *NativePlugin) GetConfigForm() *ConfigForm {
	return np.instance.GetConfigForm()
}

func (np *NativePlugin) GetAPIRoutes() []APIRoute {
	return np.instance.GetAPIRoutes()
}

func (np *NativePlugin) GetCommands() []Command {
	return np.instance.GetCommands()
}

func (np *NativePlugin) GetServices() []Service {
	return np.instance.GetServices()
}