package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// PluginRepository 插件仓储接口
type PluginRepository interface {
	// Create 创建插件记录
	Create(plugin *model.Plugin) error

	// Update 更新插件记录
	Update(plugin *model.Plugin) error

	// Delete 删除插件记录
	Delete(pluginID string) error

	// GetByID 根据ID获取插件
	GetByID(pluginID string) (*model.Plugin, error)

	// GetByName 根据名称获取插件
	GetByName(name string) (*model.Plugin, error)

	// GetAll 获取所有插件
	GetAll() ([]*model.Plugin, error)

	// GetEnabled 获取已启用的插件
	GetEnabled() ([]*model.Plugin, error)

	// GetInstalled 获取已安装的插件
	GetInstalled() ([]*model.Plugin, error)

	// GetByType 根据类型获取插件
	GetByType(pluginType string) ([]*model.Plugin, error)

	// ListByPage 分页获取插件列表
	ListByPage(page, count int) ([]*model.Plugin, int64, error)

	// Search 搜索插件
	Search(keyword string, page, count int) ([]*model.Plugin, int64, error)

	// Count 统计插件数量
	Count() (int64, error)

	// CountEnabled 统计已启用的插件数量
	CountEnabled() (int64, error)

	// CountInstalled 统计已安装的插件数量
	CountInstalled() (int64, error)

	// UpdateConfig 更新插件配置
	UpdateConfig(pluginID, config string) error

	// UpdateStatus 更新插件状态
	UpdateStatus(pluginID string, enabled bool) error

	// UpdateInstallStatus 更新插件安装状态
	UpdateInstallStatus(pluginID string, installed bool) error
}
