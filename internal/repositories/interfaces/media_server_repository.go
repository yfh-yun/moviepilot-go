package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// MediaServerRepository 媒体服务器仓储接口
type MediaServerRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, server *database.MediaServer) error
	GetByID(ctx context.Context, id uint) (*database.MediaServer, error)
	GetByName(ctx context.Context, name string) (*database.MediaServer, error)
	Update(ctx context.Context, server *database.MediaServer) error
	Delete(ctx context.Context, id uint) error

	// 查询方法
	List(ctx context.Context, params ListMediaServerParams) ([]*database.MediaServer, error)
	ListActive(ctx context.Context) ([]*database.MediaServer, error)
	ListByType(ctx context.Context, serverType string) ([]*database.MediaServer, error)

	// 配置管理
	UpdateAPIKey(ctx context.Context, id uint, apiKey string) error
	UpdateSyncLibs(ctx context.Context, id uint, libs string) error
	UpdateSettings(ctx context.Context, id uint, settings string) error

	// 状态管理
	SetActive(ctx context.Context, id uint, active bool) error

	// 检查方法
	Exists(ctx context.Context, name string) (bool, error)
	ExistsByType(ctx context.Context, serverType string) (bool, error)
}

// ListMediaServerParams 媒体服务器查询参数
type ListMediaServerParams struct {
	Type     string // 服务器类型
	IsActive *bool  // 是否启用
}
