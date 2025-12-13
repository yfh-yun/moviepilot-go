// Package plugin MoviePilot插件系统核心包
package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"
)

// Manager 插件管理器接口
type Manager interface {
	// LoadPlugins 加载插件
	LoadPlugins(ctx context.Context) error
	// GetPlugin 获取插件
	GetPlugin(ctx context.Context, pluginID string) (Plugin, error)
	// GetPlugins 获取所有插件
	GetPlugins(ctx context.Context) ([]Plugin, error)
	// GetPluginInfo 获取插件信息
	GetPluginInfo(ctx context.Context, pluginID string) (*PluginInfo, error)
	// GetPluginsInfo 获取所有插件信息
	GetPluginsInfo(ctx context.Context) ([]*PluginInfo, error)
	// EnablePlugin 启用插件
	EnablePlugin(ctx context.Context, pluginID string) error
	// DisablePlugin 禁用插件
	DisablePlugin(ctx context.Context, pluginID string) error
	// InitPlugin 初始化插件
	InitPlugin(ctx context.Context, pluginID string, cfg map[string]any) error
	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error
	// SetPluginConfig 设置插件配置
	SetPluginConfig(ctx context.Context, pluginID string, cfg map[string]any) error
	// GetPluginConfig 获取插件配置
	GetPluginConfig(ctx context.Context, pluginID string) (map[string]any, error)
	// ExecuteCommand 执行插件命令
	ExecuteCommand(ctx context.Context, pluginID string, cmd Command, args map[string]any) error
	// GetOnlinePlugins 获取所有在线插件信息
	GetOnlinePlugins(ctx context.Context, force bool) ([]*PluginInfo, error)
	// AsyncGetOnlinePlugins 异步获取所有在线插件信息
	AsyncGetOnlinePlugins(ctx context.Context, force bool) ([]*PluginInfo, error)
	// ClearOnlinePluginsCache 清除在线插件缓存
	ClearOnlinePluginsCache() error
	// Close 关闭插件管理器
	Close(ctx context.Context) error
}

// PluginManager 插件管理器实现
type PluginManager struct {
	plugins      map[string]Plugin
	configStore  ConfigStore
	dataStore    DataStore
	eventManager EventManager
	logger       *zap.Logger
	mutex        sync.RWMutex
	running      bool
	// 缓存字段
	onlinePluginsCache      *cache.CachedFunction // 在线插件缓存
	asyncOnlinePluginsCache *cache.CachedFunction // 异步在线插件缓存
	cacheBackend            cache.CacheBackend    // 缓存后端
}

// NewPluginManager 创建插件管理器实例
func NewPluginManager(configStore ConfigStore, dataStore DataStore, eventManager EventManager) Manager {
	pm := &PluginManager{
		plugins:      make(map[string]Plugin),
		configStore:  configStore,
		dataStore:    dataStore,
		eventManager: eventManager,
		logger:       logger.GetLogger(),
		running:      false,
	}

	// 初始化缓存后端
	pm.initCache()

	return pm
}

// LoadPlugins 加载插件
func (pm *PluginManager) LoadPlugins(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return errors.New("plugin manager already running")
	}

	pm.logger.Info("开始加载插件")

	// TODO: 实现插件加载逻辑
	// 1. 扫描插件目录
	// 2. 加载Go原生插件
	// 3. 连接Python插件服务
	// 4. 初始化插件

	pm.running = true
	pm.logger.Info("插件加载完成")

	return nil
}

// GetPlugin 获取插件
func (pm *PluginManager) GetPlugin(ctx context.Context, pluginID string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return plugin, nil
}

// GetPlugins 获取所有插件
func (pm *PluginManager) GetPlugins(ctx context.Context) ([]Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetPluginInfo 获取插件信息
func (pm *PluginManager) GetPluginInfo(ctx context.Context, pluginID string) (*PluginInfo, error) {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	config, err := plugin.Config(ctx)
	if err != nil {
		pm.logger.Error("获取插件配置失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		config = make(map[string]any)
	}

	commands, err := plugin.Commands(ctx)
	if err != nil {
		pm.logger.Error("获取插件命令失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		commands = make([]Command, 0)
	}

	apis, err := plugin.APIs(ctx)
	if err != nil {
		pm.logger.Error("获取插件API失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		apis = make([]API, 0)
	}

	services, err := plugin.Services(ctx)
	if err != nil {
		pm.logger.Error("获取插件服务失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		services = make([]Service, 0)
	}

	actions, err := plugin.Actions(ctx)
	if err != nil {
		pm.logger.Error("获取插件动作失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		actions = make([]ActionGroup, 0)
	}

	dashboardMeta, err := plugin.DashboardMeta(ctx)
	if err != nil {
		pm.logger.Error("获取插件仪表盘元信息失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		dashboardMeta = make([]DashboardMeta, 0)
	}

	state := plugin.State(ctx)

	return &PluginInfo{
		ID:            plugin.ID(),
		Name:          plugin.Name(),
		Version:       plugin.Version(),
		Description:   plugin.Description(),
		Icon:          plugin.Icon(),
		Author:        plugin.Author(),
		AuthorURL:     plugin.AuthorURL(),
		Label:         plugin.Label(),
		State:         string(state),
		HasPage:       len(dashboardMeta) > 0,
		HasUpdate:     false,
		IsLocal:       true,
		Config:        config,
		Commands:      commands,
		APIs:          apis,
		Services:      services,
		Actions:       actions,
		DashboardMeta: dashboardMeta,
	}, nil
}

// GetPluginsInfo 获取所有插件信息
func (pm *PluginManager) GetPluginsInfo(ctx context.Context) ([]*PluginInfo, error) {
	plugins, err := pm.GetPlugins(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]*PluginInfo, 0, len(plugins))
	for _, plugin := range plugins {
		info, err := pm.GetPluginInfo(ctx, plugin.ID())
		if err != nil {
			pm.logger.Error("获取插件信息失败",
				zap.String("plugin_id", plugin.ID()),
				zap.Error(err),
			)
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// EnablePlugin 启用插件
func (pm *PluginManager) EnablePlugin(ctx context.Context, pluginID string) error {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	if err := plugin.SetState(ctx, StateEnabled); err != nil {
		pm.logger.Error("启用插件失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return err
	}

	pm.logger.Info("插件已启用",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", plugin.Name()),
	)

	return nil
}

// DisablePlugin 禁用插件
func (pm *PluginManager) DisablePlugin(ctx context.Context, pluginID string) error {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	if err := plugin.SetState(ctx, StateDisabled); err != nil {
		pm.logger.Error("禁用插件失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return err
	}

	pm.logger.Info("插件已禁用",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", plugin.Name()),
	)

	return nil
}

// InitPlugin 初始化插件
func (pm *PluginManager) InitPlugin(ctx context.Context, pluginID string, cfg map[string]any) error {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	if err := plugin.Init(ctx, cfg); err != nil {
		pm.logger.Error("初始化插件失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return err
	}

	pm.logger.Info("插件已初始化",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", plugin.Name()),
	)

	return nil
}

// StopPlugin 停止插件
func (pm *PluginManager) StopPlugin(ctx context.Context, pluginID string) error {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	if err := plugin.Stop(ctx); err != nil {
		pm.logger.Error("停止插件失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return err
	}

	pm.logger.Info("插件已停止",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", plugin.Name()),
	)

	return nil
}

// SetPluginConfig 设置插件配置
func (pm *PluginManager) SetPluginConfig(ctx context.Context, pluginID string, cfg map[string]any) error {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	if err := plugin.SetConfig(ctx, cfg); err != nil {
		pm.logger.Error("设置插件配置失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return err
	}

	pm.logger.Info("插件配置已更新",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", plugin.Name()),
	)

	return nil
}

// GetPluginConfig 获取插件配置
func (pm *PluginManager) GetPluginConfig(ctx context.Context, pluginID string) (map[string]any, error) {
	plugin, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	config, err := plugin.Config(ctx)
	if err != nil {
		pm.logger.Error("获取插件配置失败",
			zap.String("plugin_id", pluginID),
			zap.Error(err),
		)
		return nil, err
	}

	return config, nil
}

// ExecuteCommand 执行插件命令
func (pm *PluginManager) ExecuteCommand(ctx context.Context, pluginID string, cmd Command, args map[string]any) error {
	_, err := pm.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	// TODO: 实现插件命令执行逻辑
	pm.logger.Info("执行插件命令",
		zap.String("plugin_id", pluginID),
		zap.String("command", cmd.Cmd),
		zap.Any("args", args),
	)

	return nil
}

// initCache 初始化缓存
func (pm *PluginManager) initCache() {
	// 初始化缓存后端（使用内存缓存，与Python版本兼容）
	pm.cacheBackend = cache.Cache("ttl", 1, 1800)
}

// GetOnlinePlugins 获取所有在线插件信息
// 原Python: @cached(maxsize=1, ttl=1800) 装饰的get_online_plugins方法
func (pm *PluginManager) GetOnlinePlugins(ctx context.Context, force bool) ([]*PluginInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 缓存键
	cacheKey := "online_plugins"
	region := "plugin.online"

	// 如果不是强制刷新，尝试从缓存获取
	if !force {
		cachedValue, hit, err := pm.cacheBackend.Get(cacheKey, region)
		if err == nil && hit {
			if plugins, ok := cachedValue.([]*PluginInfo); ok && plugins != nil {
				pm.logger.Debug("从缓存获取在线插件信息", zap.Int("count", len(plugins)))
				return plugins, nil
			}
		}
	}

	// 缓存未命中或强制刷新，获取在线插件
	pm.logger.Info("从插件市场获取在线插件信息")

	// TODO: 实现从插件市场获取插件的逻辑
	// 1. 调用PluginHelper获取在线插件
	// 2. 处理插件信息
	// 3. 返回插件列表

	// 模拟返回空列表（实际实现时替换为真实逻辑）
	plugins := []*PluginInfo{}

	// 将结果存入缓存
	err := pm.cacheBackend.Set(cacheKey, plugins, 1800*time.Second, region)
	if err != nil {
		pm.logger.Error("缓存在线插件信息失败", zap.Error(err))
	}

	return plugins, nil
}

// AsyncGetOnlinePlugins 异步获取所有在线插件信息
// 原Python: @cached(maxsize=1, ttl=1800) 装饰的async_get_online_plugins方法
func (pm *PluginManager) AsyncGetOnlinePlugins(ctx context.Context, force bool) ([]*PluginInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 缓存键
	cacheKey := "async_online_plugins"
	region := "plugin.online"

	// 如果不是强制刷新，尝试从缓存获取
	if !force {
		cachedValue, hit, err := pm.cacheBackend.Get(cacheKey, region)
		if err == nil && hit {
			if plugins, ok := cachedValue.([]*PluginInfo); ok && plugins != nil {
				pm.logger.Debug("从缓存获取异步在线插件信息", zap.Int("count", len(plugins)))
				return plugins, nil
			}
		}
	}

	// 缓存未命中或强制刷新，异步获取在线插件
	pm.logger.Info("异步从插件市场获取在线插件信息")

	// TODO: 实现异步从插件市场获取插件的逻辑
	// 1. 异步调用PluginHelper获取在线插件
	// 2. 处理插件信息
	// 3. 返回插件列表

	// 模拟返回空列表（实际实现时替换为真实逻辑）
	plugins := []*PluginInfo{}

	// 将结果存入缓存
	err := pm.cacheBackend.Set(cacheKey, plugins, 1800*time.Second, region)
	if err != nil {
		pm.logger.Error("缓存异步在线插件信息失败", zap.Error(err))
	}

	return plugins, nil
}

// ClearOnlinePluginsCache 清除在线插件缓存
func (pm *PluginManager) ClearOnlinePluginsCache() error {
	pm.logger.Info("清除在线插件缓存")
	return pm.cacheBackend.Clear("plugin.online")
}

// Close 关闭插件管理器
func (pm *PluginManager) Close(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return nil
	}

	// 停止所有插件
	for pluginID, plugin := range pm.plugins {
		if err := plugin.Stop(ctx); err != nil {
			pm.logger.Error("停止插件失败",
				zap.String("plugin_id", pluginID),
				zap.Error(err),
			)
			continue
		}
	}

	// 关闭缓存后端
	if pm.cacheBackend != nil {
		pm.cacheBackend.Close()
	}

	pm.plugins = make(map[string]Plugin)
	pm.running = false

	pm.logger.Info("插件管理器已关闭")

	return nil
}
