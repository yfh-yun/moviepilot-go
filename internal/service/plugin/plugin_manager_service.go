package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

// PluginManagerService 插件管理器服务
type PluginManagerService struct {
	manager *plugin.HybridPluginManager
	logger  *logger.Logger
	config  *plugin.Config
	mutex   sync.RWMutex
}

// NewPluginManagerService 创建插件管理器服务
func NewPluginManagerService(config *plugin.Config, log *logger.Logger) (service.PluginManagerService, error) {
	// 创建混合插件管理器
	manager, err := plugin.NewHybridPluginManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create hybrid plugin manager: %w", err)
	}

	service := &PluginManagerService{
		manager: manager,
		logger:  log,
		config:  config,
	}

	// 启动健康监控
	go service.manager.MonitorPluginHealth(context.Background())

	return service, nil
}

// GetManager 获取插件管理器
func (s *PluginManagerService) GetManager() *plugin.HybridPluginManager {
	return s.manager
}

// LoadPlugin 加载插件
func (s *PluginManagerService) LoadPlugin(pluginPath string, pluginType plugin.PluginType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debug("Loading plugin", "path", pluginPath, "type", pluginType)
	
	if err := s.manager.LoadPlugin(pluginPath, pluginType); err != nil {
		s.logger.Error("Failed to load plugin", "path", pluginPath, "type", pluginType, "error", err)
		return err
	}

	s.logger.Info("Plugin loaded successfully", "path", pluginPath, "type", pluginType)
	return nil
}

// InitializePlugin 初始化插件
func (s *PluginManagerService) InitializePlugin(pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debug("Initializing plugin", "pluginId", pluginID)
	
	if err := s.manager.InitializePlugin(pluginID); err != nil {
		s.logger.Error("Failed to initialize plugin", "pluginId", pluginID, "error", err)
		return err
	}

	s.logger.Info("Plugin initialized successfully", "pluginId", pluginID)
	return nil
}

// StartPlugin 启动插件
func (s *PluginManagerService) StartPlugin(pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debug("Starting plugin", "pluginId", pluginID)
	
	if err := s.manager.StartPlugin(pluginID); err != nil {
		s.logger.Error("Failed to start plugin", "pluginId", pluginID, "error", err)
		return err
	}

	s.logger.Info("Plugin started successfully", "pluginId", pluginID)
	return nil
}

// StopPlugin 停止插件
func (s *PluginManagerService) StopPlugin(pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debug("Stopping plugin", "pluginId", pluginID)
	
	if err := s.manager.StopPlugin(pluginID); err != nil {
		s.logger.Error("Failed to stop plugin", "pluginId", pluginID, "error", err)
		return err
	}

	s.logger.Info("Plugin stopped successfully", "pluginId", pluginID)
	return nil
}

// UnloadPlugin 卸载插件
func (s *PluginManagerService) UnloadPlugin(pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debug("Unloading plugin", "pluginId", pluginID)
	
	if err := s.manager.UnloadPlugin(pluginID); err != nil {
		s.logger.Error("Failed to unload plugin", "pluginId", pluginID, "error", err)
		return err
	}

	s.logger.Info("Plugin unloaded successfully", "pluginId", pluginID)
	return nil
}

// CallPluginMethod 调用插件方法
func (s *PluginManagerService) CallPluginMethod(pluginID, method string, args ...interface{}) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Calling plugin method", "pluginId", pluginID, "method", method)
	
	result, err := s.manager.CallPluginMethod(pluginID, method, args...)
	if err != nil {
		s.logger.Error("Failed to call plugin method", "pluginId", pluginID, "method", method, "error", err)
		return nil, err
	}

	s.logger.Debug("Plugin method called successfully", "pluginId", pluginID, "method", method)
	return result, nil
}

// GetPluginInfo 获取插件信息
func (s *PluginManagerService) GetPluginInfo(pluginID string) (*plugin.PluginInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.manager.GetPluginInfo(pluginID)
}

// ListPlugins 列出所有插件
func (s *PluginManagerService) ListPlugins() []*plugin.PluginInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.manager.ListPlugins()
}

// PublishEvent 发布事件
func (s *PluginManagerService) PublishEvent(event plugin.Event) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Publishing event", "type", event.Type, "source", event.Source)
	s.manager.PublishEvent(event)
}

// RestartPlugin 重启插件
func (s *PluginManagerService) RestartPlugin(pluginID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("Restarting plugin", "pluginId", pluginID)

	// 停止插件
	if err := s.manager.StopPlugin(pluginID); err != nil {
		s.logger.Error("Failed to stop plugin during restart", "pluginId", pluginID, "error", err)
		return err
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 启动插件
	if err := s.manager.StartPlugin(pluginID); err != nil {
		s.logger.Error("Failed to start plugin during restart", "pluginId", pluginID, "error", err)
		return err
	}

	s.logger.Info("Plugin restarted successfully", "pluginId", pluginID)
	return nil
}

// GetPluginStatus 获取插件状态
func (s *PluginManagerService) GetPluginStatus(pluginID string) (plugin.PluginState, error) {
	info, err := s.GetPluginInfo(pluginID)
	if err != nil {
		return plugin.StateError, err
	}

	return plugin.PluginState(info.State), nil
}

// IsPluginRunning 检查插件是否正在运行
func (s *PluginManagerService) IsPluginRunning(pluginID string) bool {
	state, err := s.GetPluginStatus(pluginID)
	if err != nil {
		return false
	}

	return state == plugin.StateRunning
}

// GetRunningPlugins 获取所有正在运行的插件
func (s *PluginManagerService) GetRunningPlugins() []*plugin.PluginInfo {
	plugins := s.ListPlugins()
	var runningPlugins []*plugin.PluginInfo

	for _, p := range plugins {
		if p.State == string(plugin.StateRunning) {
			runningPlugins = append(runningPlugins, p)
		}
	}

	return runningPlugins
}

// StopAllPlugins 停止所有插件
func (s *PluginManagerService) StopAllPlugins() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	plugins := s.manager.ListPlugins()
	var errors []error

	for _, p := range plugins {
		if p.GetState() == plugin.StateRunning {
			if err := s.manager.StopPlugin(p.ID()); err != nil {
				s.logger.Error("Failed to stop plugin", "pluginId", p.ID(), "error", err)
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some plugins: %v", errors)
	}

	s.logger.Info("All plugins stopped successfully")
	return nil
}

// Shutdown 关闭插件管理器服务
func (s *PluginManagerService) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down plugin manager service")

	// 停止所有插件
	if err := s.StopAllPlugins(); err != nil {
		s.logger.Error("Failed to stop all plugins during shutdown", "error", err)
		return err
	}

	s.logger.Info("Plugin manager service shutdown completed")
	return nil
}