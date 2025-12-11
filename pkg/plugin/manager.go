package plugin

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// HybridPluginManager 混合插件管理器（Go原生 + Python插件）
type HybridPluginManager struct {
	logger      *zap.Logger
	plugins     map[string]Plugin // 所有已注册插件（Go + Python代理）
	running     map[string]Plugin // 当前启用的插件
	configStore ConfigStore       // 插件配置存储
	dataStore   DataStore         // 插件数据存储
	pluginInfos []*PluginInfo     // 插件信息缓存
}

// NewHybridPluginManager 创建混合插件管理器
func NewHybridPluginManager(configStore ConfigStore, dataStore DataStore) *HybridPluginManager {
	return &HybridPluginManager{
		logger:      logger.GetLogger(),
		plugins:     make(map[string]Plugin),
		running:     make(map[string]Plugin),
		configStore: configStore,
		dataStore:   dataStore,
		pluginInfos: make([]*PluginInfo, 0),
	}
}

// Register 注册插件
func (m *HybridPluginManager) Register(p Plugin) {
	id := p.ID()
	m.plugins[id] = p
	m.logger.Info("插件已注册", zap.String("id", id), zap.String("name", p.Name()))
	m.refreshPluginInfos()
}

// Unregister 注销插件
func (m *HybridPluginManager) Unregister(id string) error {
	if _, exists := m.plugins[id]; !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	// 停止插件
	if err := m.StopPlugin(id); err != nil {
		m.logger.Warn("停止插件失败", zap.Error(err))
	}

	delete(m.plugins, id)
	m.logger.Info("插件已注销", zap.String("id", id))
	m.refreshPluginInfos()
	return nil
}

// StartPlugin 启动插件
func (m *HybridPluginManager) StartPlugin(id string) error {
	p, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	// 获取插件配置
	cfg, err := m.configStore.Get(context.Background(), id)
	if err != nil {
		m.logger.Error("获取插件配置失败", zap.String("id", id), zap.Error(err))
		cfg = make(map[string]any)
	}

	// 初始化插件
	if err := p.Init(context.Background(), cfg); err != nil {
		return fmt.Errorf("初始化插件失败: %w", err)
	}

	// 设置插件状态为启用
	if err := p.SetState(context.Background(), StateEnabled); err != nil {
		m.logger.Error("设置插件状态失败", zap.String("id", id), zap.Error(err))
		return err
	}

	m.running[id] = p
	m.logger.Info("插件已启动", zap.String("id", id), zap.String("name", p.Name()))
	m.refreshPluginInfos()
	return nil
}

// StopPlugin 停止插件
func (m *HybridPluginManager) StopPlugin(id string) error {
	p, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	// 设置插件状态为禁用
	if err := p.SetState(context.Background(), StateDisabled); err != nil {
		m.logger.Error("设置插件状态失败", zap.String("id", id), zap.Error(err))
	}

	// 停止插件
	if err := p.Stop(context.Background()); err != nil {
		m.logger.Error("停止插件失败", zap.String("id", id), zap.Error(err))
	}

	// 从运行中插件列表中移除
	delete(m.running, id)
	m.logger.Info("插件已停止", zap.String("id", id), zap.String("name", p.Name()))
	m.refreshPluginInfos()
	return nil
}

// ReloadPlugin 重载插件
func (m *HybridPluginManager) ReloadPlugin(id string) error {
	m.logger.Info("重载插件", zap.String("id", id))

	// 停止插件
	if err := m.StopPlugin(id); err != nil {
		m.logger.Warn("停止插件失败", zap.Error(err))
	}

	// 启动插件
	if err := m.StartPlugin(id); err != nil {
		return fmt.Errorf("启动插件失败: %w", err)
	}

	m.logger.Info("插件已重载", zap.String("id", id))
	return nil
}

// StartAll 启动所有插件
func (m *HybridPluginManager) StartAll() error {
	m.logger.Info("启动所有插件")

	for id := range m.plugins {
		if err := m.StartPlugin(id); err != nil {
			m.logger.Error("启动插件失败", zap.String("id", id), zap.Error(err))
		}
	}

	return nil
}

// StopAll 停止所有插件
func (m *HybridPluginManager) StopAll() error {
	m.logger.Info("停止所有插件")

	for id := range m.running {
		if err := m.StopPlugin(id); err != nil {
			m.logger.Error("停止插件失败", zap.String("id", id), zap.Error(err))
		}
	}

	return nil
}

// GetPlugin 获取插件
func (m *HybridPluginManager) GetPlugin(id string) (Plugin, error) {
	p, exists := m.plugins[id]
	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}
	return p, nil
}

// GetRunningPlugin 获取运行中的插件
func (m *HybridPluginManager) GetRunningPlugin(id string) (Plugin, error) {
	p, exists := m.running[id]
	if !exists {
		return nil, fmt.Errorf("插件未运行: %s", id)
	}
	return p, nil
}

// ListPlugins 列出所有插件信息
func (m *HybridPluginManager) ListPlugins() []*PluginInfo {
	return m.pluginInfos
}

// GetPluginInfo 获取插件信息
func (m *HybridPluginManager) GetPluginInfo(id string) (*PluginInfo, error) {
	for _, info := range m.pluginInfos {
		if info.ID == id {
			return info, nil
		}
	}
	return nil, fmt.Errorf("插件不存在: %s", id)
}

// refreshPluginInfos 刷新插件信息缓存
func (m *HybridPluginManager) refreshPluginInfos() {
	ctx := context.Background()
	infos := make([]*PluginInfo, 0, len(m.plugins))

	for id, p := range m.plugins {
		// 获取插件配置
		cfg, err := p.Config(ctx)
		if err != nil {
			m.logger.Error("获取插件配置失败", zap.String("id", id), zap.Error(err))
			cfg = make(map[string]any)
		}

		// 获取插件命令
		commands, err := p.Commands(ctx)
		if err != nil {
			m.logger.Error("获取插件命令失败", zap.String("id", id), zap.Error(err))
			commands = make([]Command, 0)
		}

		// 获取插件API
		apis, err := p.APIs(ctx)
		if err != nil {
			m.logger.Error("获取插件API失败", zap.String("id", id), zap.Error(err))
			apis = make([]API, 0)
		}

		// 获取插件服务
		services, err := p.Services(ctx)
		if err != nil {
			m.logger.Error("获取插件服务失败", zap.String("id", id), zap.Error(err))
			services = make([]Service, 0)
		}

		// 获取插件动作
		actions, err := p.Actions(ctx)
		if err != nil {
			m.logger.Error("获取插件动作失败", zap.String("id", id), zap.Error(err))
			actions = make([]ActionGroup, 0)
		}

		// 获取仪表盘元信息
		dashboardMeta, err := p.DashboardMeta(ctx)
		if err != nil {
			m.logger.Error("获取仪表盘元信息失败", zap.String("id", id), zap.Error(err))
			dashboardMeta = make([]DashboardMeta, 0)
		}

		// 构建插件信息
		info := &PluginInfo{
			ID:            id,
			Name:          p.Name(),
			Version:       p.Version(),
			Description:   p.Description(),
			Icon:          p.Icon(),
			Author:        p.Author(),
			AuthorURL:     p.AuthorURL(),
			Label:         p.Label(),
			State:         string(p.State(ctx)),
			HasPage:       false, // 默认值，可根据插件实际情况调整
			HasUpdate:     false,
			IsLocal:       true, // 默认值，可根据插件实际情况调整
			RepoURL:       "",
			InstallCount:  0,
			Config:        cfg,
			Commands:      commands,
			APIs:          apis,
			Services:      services,
			Actions:       actions,
			DashboardMeta: dashboardMeta,
		}

		infos = append(infos, info)
	}

	m.pluginInfos = infos
}

// ConfigurePlugin 配置插件
func (m *HybridPluginManager) ConfigurePlugin(id string, config map[string]any) error {
	p, err := m.GetPlugin(id)
	if err != nil {
		return err
	}

	// 设置插件配置
	if err := p.SetConfig(context.Background(), config); err != nil {
		return fmt.Errorf("设置插件配置失败: %w", err)
	}

	// 保存配置到存储
	if err := m.configStore.Set(context.Background(), id, config); err != nil {
		return fmt.Errorf("保存插件配置失败: %w", err)
	}

	m.logger.Info("插件配置已更新", zap.String("id", id))
	m.refreshPluginInfos()
	return nil
}

// GetPluginConfig 获取插件配置
func (m *HybridPluginManager) GetPluginConfig(id string) (map[string]any, error) {
	p, err := m.GetPlugin(id)
	if err != nil {
		return nil, err
	}

	return p.Config(context.Background())
}

// EnablePlugin 启用插件
func (m *HybridPluginManager) EnablePlugin(id string) error {
	return m.StartPlugin(id)
}

// DisablePlugin 禁用插件
func (m *HybridPluginManager) DisablePlugin(id string) error {
	return m.StopPlugin(id)
}
