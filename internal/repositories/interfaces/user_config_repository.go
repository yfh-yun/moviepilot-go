package interfaces

import "github.com/yfh-yun/moviepilot-go/internal/models"

// UserConfigRepository 用户配置仓储接口
type UserConfigRepository interface {
	// 基础CRUD
	Create(config *model.UserConfig) error
	GetByID(id uint) (*model.UserConfig, error)
	Update(config *model.UserConfig) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.UserConfig, int64, error)
	
	// 按用户和键获取
	GetByKey(username, key string) (*model.UserConfig, error)
	
	// 按用户获取
	GetByUsername(username string) ([]*model.UserConfig, error)
	
	// 按键获取
	GetByKeyOnly(key string) ([]*model.UserConfig, error)
	
	// 设置配置
	SetConfig(username, key, value string) error
	
	// 删除配置
	DeleteByKey(username, key string) error
	DeleteByUsername(username string) error
	
	// 批量操作
	BatchSet(username string, configs map[string]string) error
}