package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// MediaServerRepository 媒体服务器仓储接口
type MediaServerRepository interface {
	// Create 创建媒体服务器
	Create(mediaServer *model.MediaServer) error
	
	// GetByID 根据ID获取媒体服务器
	GetByID(id uint) (*model.MediaServer, error)
	
	// GetByName 根据名称获取媒体服务器
	GetByName(name string) (*model.MediaServer, error)
	
	// GetByType 根据类型获取媒体服务器
	GetByType(serverType string) ([]*model.MediaServer, error)
	
	// GetActive 获取活跃的媒体服务器
	GetActive() ([]*model.MediaServer, error)
	
	// TestConnection 测试媒体服务器连接
	TestConnection(id uint) error
	
	// SyncLibraries 同步媒体库
	SyncLibraries(id uint) error
	
	// Update 更新媒体服务器
	Update(mediaServer *model.MediaServer) error
	
	// UpdateStatus 更新状态
	UpdateStatus(id uint, isActive bool) error
	
	// Delete 删除媒体服务器
	Delete(id uint) error
	
	// List 分页获取媒体服务器列表
	List(offset, limit int) ([]*model.MediaServer, int64, error)
	
	// Count 统计媒体服务器数量
	Count() (int64, error)
	
	// GetLibraries 获取媒体库列表
	GetLibraries(id uint) ([]interface{}, error)
}