package plugin

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/plugin"
)

// Service 插件服务接口
type Service interface {
	// ListPlugins 列出所有插件
	ListPlugins(ctx context.Context) ([]plugin.PluginInfo, error)

	// GetPlugin 获取插件信息
	GetPlugin(ctx context.Context, id string) (*plugin.PluginInfo, error)

	// EnablePlugin 启用插件
	EnablePlugin(ctx context.Context, id string) error

	// DisablePlugin 禁用插件
	DisablePlugin(ctx context.Context, id string) error

	// ConfigurePlugin 配置插件
	ConfigurePlugin(ctx context.Context, id string, config map[string]any) error

	// ReloadPlugin 重载插件
	ReloadPlugin(ctx context.Context, id string) error

	// GetPluginConfig 获取插件配置
	GetPluginConfig(ctx context.Context, id string) (map[string]any, error)
}

// service 插件服务实现
type service struct {
	manager plugin.Manager
	logger  *zap.Logger
}

// NewService 创建插件服务
func NewService(manager plugin.Manager) Service {
	return &service{
		manager: manager,
		logger:  logger.GetLogger(),
	}
}

// ListPlugins 列出所有插件
func (s *service) ListPlugins(ctx context.Context) ([]plugin.PluginInfo, error) {
	s.logger.Info("列出所有插件")

	pluginInfos, err := s.manager.GetPluginsInfo(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]plugin.PluginInfo, 0, len(pluginInfos))
	for _, p := range pluginInfos {
		if p != nil {
			infos = append(infos, *p)
		}
	}

	return infos, nil
}

// GetPlugin 获取插件信息
func (s *service) GetPlugin(ctx context.Context, id string) (*plugin.PluginInfo, error) {
	s.logger.Info("获取插件信息", zap.String("id", id))

	return s.manager.GetPluginInfo(ctx, id)
}

// EnablePlugin 启用插件
func (s *service) EnablePlugin(ctx context.Context, id string) error {
	s.logger.Info("启用插件", zap.String("id", id))

	return s.manager.EnablePlugin(ctx, id)
}

// DisablePlugin 禁用插件
func (s *service) DisablePlugin(ctx context.Context, id string) error {
	s.logger.Info("禁用插件", zap.String("id", id))

	return s.manager.DisablePlugin(ctx, id)
}

// ConfigurePlugin 配置插件
func (s *service) ConfigurePlugin(ctx context.Context, id string, config map[string]any) error {
	s.logger.Info("配置插件", zap.String("id", id))

	return s.manager.SetPluginConfig(ctx, id, config)
}

// ReloadPlugin 重载插件
func (s *service) ReloadPlugin(ctx context.Context, id string) error {
	s.logger.Info("重载插件", zap.String("id", id))

	// 停止插件
	if err := s.DisablePlugin(ctx, id); err != nil {
		s.logger.Warn("停止插件失败", zap.Error(err))
	}

	// 启动插件
	if err := s.EnablePlugin(ctx, id); err != nil {
		return fmt.Errorf("启动插件失败: %w", err)
	}

	return nil
}

// GetPluginConfig 获取插件配置
func (s *service) GetPluginConfig(ctx context.Context, id string) (map[string]any, error) {
	s.logger.Info("获取插件配置", zap.String("id", id))

	return s.manager.GetPluginConfig(ctx, id)
}
