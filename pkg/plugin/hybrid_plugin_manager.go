package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeNative PluginType = "native" // Go原生插件
	PluginTypeScript PluginType = "script" // Python脚本插件
	PluginTypeWeb    PluginType = "web"    // WebAssembly插件
)

// PluginState 插件状态
type PluginState string

const (
	StateUnloaded    PluginState = "unloaded"
	StateLoaded      PluginState = "loaded"
	StateInitialized PluginState = "initialized"
	StateRunning     PluginState = "running"
	StateStopped     PluginState = "stopped"
	StateError       PluginState = "error"
)

// Plugin 统一插件接口
type Plugin interface {
	ID() string
	Name() string
	Version() string
	Type() PluginType
	Description() string

	// 生命周期方法
	Initialize(config map[string]interface{}) error
	Start() error
	Stop() error
	Destroy() error
	GetState() PluginState

	// 事件处理
	HandleEvent(event Event) error

	// 配置和API
	GetConfigForm() *ConfigForm
	GetAPIRoutes() []APIRoute
	GetCommands() []Command
	GetServices() []Service
}

// HybridPluginManager 混合插件管理器
type HybridPluginManager struct {
	goManager      *NativePluginManager    // Go原生插件管理器
	pythonManager  *PythonPluginManager    // Python插件管理器
	pluginRegistry map[string]*PluginEntry // 插件注册表
	eventBus       EventBus                // 事件总线
	configManager  *configManager          // 配置管理器
	logger         Logger                  // 日志器
	mutex          sync.RWMutex            // 读写锁
}

// PluginEntry 插件条目
type PluginEntry struct {
	Plugin    Plugin
	Type      PluginType
	Config    map[string]interface{}
	State     PluginState
	Metadata  PluginMetadata
	LastError error
	LoadTime  time.Time
}



// NewHybridPluginManager 创建混合插件管理器
func NewHybridPluginManager(config *Config) (*HybridPluginManager, error) {
	// 创建Go原生插件管理器
	goManager, err := NewNativePluginManager(config.Plugins.Native.Path)
	if err != nil {
		return nil, err
	}

	// 创建Python插件管理器
	pythonManager, err := NewPythonPluginManager(
		config.Plugins.Python.Host,
		config.Plugins.Python.Port,
		config.Plugins.Python.Timeout,
	)
	if err != nil {
		return nil, err
	}

	return &HybridPluginManager{
		goManager:      goManager,
		pythonManager:  pythonManager,
		pluginRegistry: make(map[string]*PluginEntry),
		eventBus:       NewEventBus(),
		configManager:  NewConfigManager(),
		logger:         NewLogger(),
	}, nil
}

// LoadPlugin 加载插件
func (h *HybridPluginManager) LoadPlugin(pluginPath string, pluginType PluginType) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var plugin Plugin
	var err error

	switch pluginType {
	case PluginTypeNative:
		plugin, err = h.goManager.LoadPlugin(pluginPath)

	case PluginTypeScript:
		plugin, err = h.pythonManager.LoadPlugin(pluginPath)

	default:
		return fmt.Errorf("unsupported plugin type: %s", pluginType)
	}

	if err != nil {
		return err
	}

	// 获取插件配置
	config, err := h.configManager.GetPluginConfig(plugin.ID())
	if err != nil {
		h.logger.Warn("Failed to get plugin config", "plugin", plugin.ID(), "error", err)
		config = make(map[string]interface{})
	}

	// 创建插件条目
	entry := &PluginEntry{
		Plugin:   plugin,
		Type:     pluginType,
		Config:   config,
		State:    StateLoaded,
		Metadata: PluginMetadata{},
		LoadTime: time.Now(),
	}

	// 注册插件
	h.pluginRegistry[plugin.ID()] = entry
	h.logger.Info("Plugin loaded", "id", plugin.ID(), "type", pluginType, "path", pluginPath)

	return nil
}

// InitializePlugin 初始化插件
func (h *HybridPluginManager) InitializePlugin(pluginID string) error {
	h.mutex.RLock()
	entry, exists := h.pluginRegistry[pluginID]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 初始化插件
	if err := entry.Plugin.Initialize(entry.Config); err != nil {
		entry.State = StateError
		entry.LastError = err
		return fmt.Errorf("failed to initialize plugin %s: %w", pluginID, err)
	}

	entry.State = StateInitialized

	// 注册事件处理器
	h.registerEventHandlers(pluginID)

	return nil
}

// StartPlugin 启动插件
func (h *HybridPluginManager) StartPlugin(pluginID string) error {
	h.mutex.RLock()
	entry, exists := h.pluginRegistry[pluginID]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	if entry.State != StateInitialized {
		return fmt.Errorf("plugin %s not initialized", pluginID)
	}

	// 启动插件
	if err := entry.Plugin.Start(); err != nil {
		entry.State = StateError
		entry.LastError = err
		return fmt.Errorf("failed to start plugin %s: %w", pluginID, err)
	}

	entry.State = StateRunning
	h.logger.Info("Plugin started", "id", pluginID)

	return nil
}

// StopPlugin 停止插件
func (h *HybridPluginManager) StopPlugin(pluginID string) error {
	h.mutex.RLock()
	entry, exists := h.pluginRegistry[pluginID]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	if err := entry.Plugin.Stop(); err != nil {
		entry.LastError = err
		return fmt.Errorf("failed to stop plugin %s: %w", pluginID, err)
	}

	entry.State = StateStopped
	h.logger.Info("Plugin stopped", "id", pluginID)

	return nil
}

// UnloadPlugin 卸载插件
func (h *HybridPluginManager) UnloadPlugin(pluginID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	entry, exists := h.pluginRegistry[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止插件
	if entry.State == StateRunning {
		if err := entry.Plugin.Stop(); err != nil {
			h.logger.Error("Failed to stop plugin before unloading", "id", pluginID, "error", err)
		}
	}

	// 销毁插件
	if err := entry.Plugin.Destroy(); err != nil {
		h.logger.Error("Failed to destroy plugin", "id", pluginID, "error", err)
	}

	// 从注册表中移除
	delete(h.pluginRegistry, pluginID)
	h.logger.Info("Plugin unloaded", "id", pluginID)

	return nil
}

// CallPluginMethod 调用插件方法
func (h *HybridPluginManager) CallPluginMethod(pluginID, method string, args ...interface{}) (interface{}, error) {
	h.mutex.RLock()
	entry, exists := h.pluginRegistry[pluginID]
	h.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 根据插件类型调用不同的方法
	switch entry.Type {
	case PluginTypeNative:
		return h.callNativePluginMethod(entry.Plugin, method, args...)

	case PluginTypeScript:
		return h.callPythonPluginMethod(pluginID, method, args...)

	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", entry.Type)
	}
}

// callNativePluginMethod 调用原生插件方法
func (h *HybridPluginManager) callNativePluginMethod(plugin Plugin, method string, _ ...interface{}) (interface{}, error) {
	// 使用反射调用方法
	// 这里简化处理，实际实现需要复杂的反射逻辑
	h.logger.Debug("Calling native plugin method", "plugin", plugin.ID(), "method", method)

	// 模拟调用
	return nil, nil
}

// callPythonPluginMethod 调用Python插件方法
func (h *HybridPluginManager) callPythonPluginMethod(pluginID, method string, args ...interface{}) (interface{}, error) {
	// 构建请求
	request := map[string]interface{}{
		"plugin_id": pluginID,
		"method":    method,
		"args":      args,
		"timestamp": time.Now().Unix(),
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// 发送HTTP请求到Python插件容器
	url := fmt.Sprintf("%s/plugin/%s/%s", h.pythonManager.baseURL, pluginID, method)
	resp, err := h.pythonManager.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plugin call failed: %s", response["error"])
	}

	return response["result"], nil
}

// registerEventHandlers 注册事件处理器
func (h *HybridPluginManager) registerEventHandlers(pluginID string) {
	entry := h.pluginRegistry[pluginID]

	// 如果插件实现了EventHandler接口，则注册事件处理器
	if handler, ok := entry.Plugin.(EventHandler); ok {
		// 获取插件订阅的事件类型
		eventTypes := handler.GetEventSubscriptions()

		for _, eventType := range eventTypes {
			h.eventBus.Subscribe(eventType, handler)
		}
	}
}

// GetPluginInfo 获取插件信息
func (h *HybridPluginManager) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	h.mutex.RLock()
	entry, exists := h.pluginRegistry[pluginID]
	h.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return &PluginInfo{
		ID:          entry.Plugin.ID(),
		Name:        entry.Plugin.Name(),
		Version:     entry.Plugin.Version(),
		Type:        string(entry.Type),
		Description: entry.Plugin.Description(),
		State:       string(entry.State),
		Config:      entry.Config,
		Metadata:    entry.Metadata,
		LastError:   entry.LastError,
		LoadTime:    entry.LoadTime,
	}, nil
}

// ListPlugins 列出所有插件
func (h *HybridPluginManager) ListPlugins() []*PluginInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var plugins []*PluginInfo

	for _, entry := range h.pluginRegistry {
		info := &PluginInfo{
			ID:          entry.Plugin.ID(),
			Name:        entry.Plugin.Name(),
			Version:     entry.Plugin.Version(),
			Type:        string(entry.Type),
			Description: entry.Plugin.Description(),
			State:       string(entry.State),
			Config:      entry.Config,
			Metadata:    entry.Metadata,
			LastError:   entry.LastError,
			LoadTime:    entry.LoadTime,
		}
		plugins = append(plugins, info)
	}

	return plugins
}

// MonitorPluginHealth 监控插件健康状态
func (h *HybridPluginManager) MonitorPluginHealth(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkPluginHealth()
		}
	}
}

// checkPluginHealth 检查插件健康状态
func (h *HybridPluginManager) checkPluginHealth() {
	h.mutex.RLock()
	plugins := make([]*PluginEntry, 0, len(h.pluginRegistry))
	for _, entry := range h.pluginRegistry {
		plugins = append(plugins, entry)
	}
	h.mutex.RUnlock()

	for _, entry := range plugins {
		if entry.Type == PluginTypeScript {
			// 检查Python插件健康状态
			if err := h.checkPythonPluginHealth(entry.Plugin.ID()); err != nil {
				entry.State = StateError
				entry.LastError = err
				h.logger.Error("Python plugin health check failed", "id", entry.Plugin.ID(), "error", err)
			}
		}
	}
}

// checkPythonPluginHealth 检查Python插件健康状态
func (h *HybridPluginManager) checkPythonPluginHealth(pluginID string) error {
	url := fmt.Sprintf("%s/health/%s", h.pythonManager.baseURL, pluginID)
	resp, err := h.pythonManager.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// PublishEvent 发布事件（公开方法）
func (h *HybridPluginManager) PublishEvent(event Event) {
	h.eventBus.Publish(event)
}