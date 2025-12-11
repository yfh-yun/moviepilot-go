package download

import (
	"context"

	"moviepilot-go/internal/models/database"
)

// Repository 下载仓储接口
type Repository interface {
	// Create 创建下载记录
	Create(ctx context.Context, download *database.Download) error

	// Update 更新下载记录
	Update(ctx context.Context, download *database.Download) error

	// Delete 删除下载记录
	Delete(ctx context.Context, id uint) error

	// GetByID 根据 ID 获取下载
	GetByID(ctx context.Context, id uint) (*database.Download, error)

	// GetByHash 根据 Hash 获取下载
	GetByHash(ctx context.Context, hash string) (*database.Download, error)

	// List 获取下载列表
	List(ctx context.Context, opts ListOptions) ([]*database.Download, error)

	// ListByStatus 根据状态获取下载列表
	ListByStatus(ctx context.Context, statuses []string) ([]*database.Download, error)

	// ListBySubscribe 根据订阅获取下载列表
	ListBySubscribe(ctx context.Context, subscribeID uint) ([]*database.Download, error)
}

// ListOptions 列表查询选项
type ListOptions struct {
	Page       int
	PageSize   int
	Status     string
	Downloader string
	Category   string
}
