package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	// Create 创建系统配置
	Create(config *model.SystemConfig) error
	
	// GetByID 根据ID获取系统配置
	GetByID(id uint) (*model.SystemConfig, error)
	
	// GetByKey 根据Key获取配置
	GetByKey(key string) (*model.SystemConfig, error)
	
	// GetByKeys 根据Keys批量获取配置
	GetByKeys(keys []string) ([]*model.SystemConfig, error)
	
	// GetAll 获取所有配置
	GetAll() ([]*model.SystemConfig, error)
	
	// Update 更新系统配置
	Update(config *model.SystemConfig) error
	
	// UpdateByKey 根据Key更新配置值
	UpdateByKey(key, value string) error
	
	// UpdateByKeys 批量更新配置
	UpdateByKeys(configs map[string]string) error
	
	// Delete 删除系统配置
	Delete(id uint) error
	
	// DeleteByKey 根据Key删除配置
	DeleteByKey(key string) error
	
	// List 分页获取系统配置列表
	List(offset, limit int) ([]*model.SystemConfig, int64, error)
	
	// Count 统计系统配置数量
	Count() (int64, error)
	
	// GetConfigsByType 根据类型获取配置
	GetConfigsByType(configType string) ([]*model.SystemConfig, error)
	
	// SetDefaultValue 设置默认值
	SetDefaultValue(key, value, configType, remark string) error
}