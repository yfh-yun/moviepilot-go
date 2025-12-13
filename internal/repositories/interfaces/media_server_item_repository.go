package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// MediaServerItemRepository 媒体服务器媒体条目仓储接口
type MediaServerItemRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, item *database.MediaServerItem) error
	GetByID(ctx context.Context, id uint) (*database.MediaServerItem, error)
	Update(ctx context.Context, item *database.MediaServerItem) error
	Delete(ctx context.Context, id uint) error

	// 查询方法
	GetByItemID(ctx context.Context, itemID string) (*database.MediaServerItem, error)
	ExistByTMDBID(ctx context.Context, tmdbID int, itemType string) (*database.MediaServerItem, error)
	ExistsByTitle(ctx context.Context, title, itemType string, year int) (*database.MediaServerItem, error)

	// 批量操作
	Empty(ctx context.Context, server *string) error
}
