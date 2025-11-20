package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
)

var (
	// ErrPluginNotFound 插件不存在
	ErrPluginNotFound = errors.New("插件不存在")
	// ErrPluginAlreadyInstalled 插件已安装
	ErrPluginAlreadyInstalled = errors.New("插件已安装")
	// ErrPluginNotInstalled 插件未安装
	ErrPluginNotInstalled = errors.New("插件未安装")
	// ErrPluginAlreadyEnabled 插件已启用
	ErrPluginAlreadyEnabled = errors.New("插件已启用")
	// ErrPluginAlreadyDisabled 插件已禁用
	ErrPluginAlreadyDisabled = errors.New("插件已禁用")
)

// PluginService 插件服务实现
type PluginService struct {
	pluginRepo interfaces.PluginRepository
	logger     *logger.Logger
	basePath   string
}

// NewPluginService 创建插件服务
func NewPluginService(
	pluginRepo interfaces.PluginRepository,
	log *logger.Logger,
	basePath string,
) service.PluginService {
	return &PluginService{
		pluginRepo: pluginRepo,
		logger:     log,
		basePath:   basePath,
	}
}

// PluginManifest 插件清单
type PluginManifest struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Repository   string                 `json:"repository"`
	Type         string                 `json:"type"`
	Entry        string                 `json:"entry"`
	Config       map[string]interface{} `json:"config"`
	Permissions  []string               `json:"permissions"`
	Dependencies map[string]string      `json:"dependencies"`
}

// InstallPlugin 安装插件
func (s *PluginService) InstallPlugin(ctx context.Context, pluginID string) error {
	s.logger.Info("安装插件", "pluginID", pluginID)

	// 检查插件是否已安装
	existingPlugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil && err.Error() != "record not found" {
		s.logger.Error("检查插件状态失败", "pluginID", pluginID, "error", err)
		return err
	}

	if existingPlugin != nil {
		return ErrPluginAlreadyInstalled
	}

	// 这里实现插件下载和安装逻辑
	// 目前先创建空插件记录
	plugin := &models.Plugin{
		ID:          pluginID,
		Name:        pluginID,
		Version:     "1.0.0",
		Description: "自动安装的插件",
		Author:      "System",
		Enabled:     false,
		Installed:   true,
		InstallPath: filepath.Join(s.basePath, "plugins", pluginID),
		Config:      "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建插件目录
	pluginDir := plugin.InstallPath
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		s.logger.Error("创建插件目录失败", "path", pluginDir, "error", err)
		return err
	}

	// 保存插件记录
	if err := s.pluginRepo.Create(plugin); err != nil {
		s.logger.Error("保存插件记录失败", "pluginID", pluginID, "error", err)
		// 清理创建的目录
		os.RemoveAll(pluginDir)
		return err
	}

	s.logger.Info("插件安装成功", "pluginID", pluginID, "path", pluginDir)
	return nil
}

// UninstallPlugin 卸载插件
func (s *PluginService) UninstallPlugin(ctx context.Context, pluginID string) error {
	s.logger.Info("卸载插件", "pluginID", pluginID)

	// 获取插件
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return err
	}

	if plugin == nil {
		return ErrPluginNotFound
	}

	if !plugin.Installed {
		return ErrPluginNotInstalled
	}

	// 删除插件目录
	if plugin.InstallPath != "" {
		if err := os.RemoveAll(plugin.InstallPath); err != nil {
			s.logger.Error("删除插件目录失败", "path", plugin.InstallPath, "error", err)
			// 继续删除记录，不中断
		}
	}

	// 删除插件记录
	if err := s.pluginRepo.Delete(pluginID); err != nil {
		s.logger.Error("删除插件记录失败", "pluginID", pluginID, "error", err)
		return err
	}

	s.logger.Info("插件卸载成功", "pluginID", pluginID)
	return nil
}

// EnablePlugin 启用插件
func (s *PluginService) EnablePlugin(ctx context.Context, pluginID string) error {
	s.logger.Info("启用插件", "pluginID", pluginID)

	// 获取插件
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return err
	}

	if plugin == nil {
		return ErrPluginNotFound
	}

	if !plugin.Installed {
		return ErrPluginNotInstalled
	}

	if plugin.Enabled {
		return ErrPluginAlreadyEnabled
	}

	// 更新插件状态
	plugin.Enabled = true
	plugin.UpdatedAt = time.Now()

	if err := s.pluginRepo.Update(plugin); err != nil {
		s.logger.Error("更新插件状态失败", "pluginID", pluginID, "error", err)
		return err
	}

	s.logger.Info("插件启用成功", "pluginID", pluginID)
	return nil
}

// DisablePlugin 禁用插件
func (s *PluginService) DisablePlugin(ctx context.Context, pluginID string) error {
	s.logger.Info("禁用插件", "pluginID", pluginID)

	// 获取插件
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return err
	}

	if plugin == nil {
		return ErrPluginNotFound
	}

	if !plugin.Installed {
		return ErrPluginNotInstalled
	}

	if !plugin.Enabled {
		return ErrPluginAlreadyDisabled
	}

	// 更新插件状态
	plugin.Enabled = false
	plugin.UpdatedAt = time.Now()

	if err := s.pluginRepo.Update(plugin); err != nil {
		s.logger.Error("更新插件状态失败", "pluginID", pluginID, "error", err)
		return err
	}

	s.logger.Info("插件禁用成功", "pluginID", pluginID)
	return nil
}

// GetPlugin 获取插件信息
func (s *PluginService) GetPlugin(ctx context.Context, pluginID string) (*models.Plugin, error) {
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return nil, err
	}

	if plugin == nil {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// ListPlugins 列出插件
func (s *PluginService) ListPlugins(ctx context.Context, enabledOnly bool) ([]*models.Plugin, error) {
	var plugins []*models.Plugin
	var err error

	if enabledOnly {
		plugins, err = s.pluginRepo.GetEnabled()
	} else {
		plugins, err = s.pluginRepo.GetAll()
	}

	if err != nil {
		s.logger.Error("获取插件列表失败", "enabledOnly", enabledOnly, "error", err)
		return nil, err
	}

	s.logger.Debug("获取插件列表成功", "count", len(plugins), "enabledOnly", enabledOnly)
	return plugins, nil
}

// UpdatePluginConfig 更新插件配置
func (s *PluginService) UpdatePluginConfig(ctx context.Context, pluginID string, config map[string]interface{}) error {
	s.logger.Info("更新插件配置", "pluginID", pluginID)

	// 获取插件
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return err
	}

	if plugin == nil {
		return ErrPluginNotFound
	}

	// 转换为JSON字符串
	configJSON, err := json.Marshal(config)
	if err != nil {
		s.logger.Error("序列化配置失败", "pluginID", pluginID, "error", err)
		return err
	}

	// 更新配置
	plugin.Config = string(configJSON)
	plugin.UpdatedAt = time.Now()

	if err := s.pluginRepo.Update(plugin); err != nil {
		s.logger.Error("更新插件配置失败", "pluginID", pluginID, "error", err)
		return err
	}

	s.logger.Info("插件配置更新成功", "pluginID", pluginID)
	return nil
}

// GetPluginConfig 获取插件配置
func (s *PluginService) GetPluginConfig(ctx context.Context, pluginID string) (map[string]interface{}, error) {
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return nil, err
	}

	if plugin == nil {
		return nil, ErrPluginNotFound
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(plugin.Config), &config); err != nil {
		s.logger.Error("解析插件配置失败", "pluginID", pluginID, "error", err)
		return nil, err
	}

	return config, nil
}

// ExecutePlugin 执行插件
func (s *PluginService) ExecutePlugin(ctx context.Context, pluginID string, data map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("执行插件", "pluginID", pluginID)

	// 获取插件
	plugin, err := s.pluginRepo.GetByID(pluginID)
	if err != nil {
		s.logger.Error("获取插件失败", "pluginID", pluginID, "error", err)
		return nil, err
	}

	if plugin == nil {
		return nil, ErrPluginNotFound
	}

	if !plugin.Installed {
		return nil, ErrPluginNotInstalled
	}

	if !plugin.Enabled {
		return nil, fmt.Errorf("插件未启用: %s", pluginID)
	}

	// 这里实现插件执行逻辑
	// 目前先返回空的执行结果
	result := map[string]interface{}{
		"plugin_id": pluginID,
		"status":    "executed",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.logger.Info("插件执行完成", "pluginID", pluginID)
	return result, nil
}
