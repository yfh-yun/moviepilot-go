package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// PluginDataRepository 插件数据仓储接口
type PluginDataRepository interface {
	// Create 创建插件数据
	Create(pluginData *model.PluginData) error
	
	// GetByID 根据ID获取插件数据
	GetByID(id uint) (*model.PluginData, error)
	
	// GetByPluginKey 根据插件Key获取数据
	GetByPluginKey(pluginKey string) ([]*model.PluginData, error)
	
	// GetByKey 根据数据Key获取数据
	GetByKey(key string) ([]*model.PluginData, error)
	
	// GetByPluginKeyAndDataKey 根据插件Key和数据Key获取数据
	GetByPluginKeyAndDataKey(pluginKey, dataKey string) (*model.PluginData, error)
	
	// GetByUserID 根据用户ID获取插件数据
	GetByUserID(userID string) ([]*model.PluginData, error)
	
	// Update 更新插件数据
	Update(pluginData *model.PluginData) error
	
	// UpdateByPluginKeyAndDataKey 根据插件Key和数据Key更新数据
	UpdateByPluginKeyAndDataKey(pluginKey, dataKey, dataValue string) error
	
	// Delete 删除插件数据
	Delete(id uint) error
	
	// DeleteByPluginKey 根据插件Key删除数据
	DeleteByPluginKey(pluginKey string) error
	
	// DeleteByUserID 根据用户ID删除数据
	DeleteByUserID(userID string) error
	
	// ClearData 清空插件数据
	ClearData(pluginKey string) error
	
	// List 分页获取插件数据列表
	List(offset, limit int) ([]*model.PluginData, int64, error)
	
	// Count 统计插件数据数量
	Count() (int64, error)
	
	// CountByPluginKey 根据插件Key统计数量
	CountByPluginKey(pluginKey string) (int64, error)
}