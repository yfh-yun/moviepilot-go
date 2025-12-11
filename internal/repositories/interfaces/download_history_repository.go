package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// DownloadHistoryRepository 下载历史仓储接口
type DownloadHistoryRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, history *database.DownloadHistory) error
	GetByID(ctx context.Context, id uint) (*database.DownloadHistory, error)
	Update(ctx context.Context, history *database.DownloadHistory) error
	Delete(ctx context.Context, id uint) error

	// 查询方法
	GetByPath(ctx context.Context, path string) (*database.DownloadHistory, error)
	GetByHash(ctx context.Context, hash string) (*database.DownloadHistory, error)
	GetByMediaID(ctx context.Context, tmdbID *int, doubanID *string) ([]*database.DownloadHistory, error)

	// 分页查询
	ListByPage(ctx context.Context, params ListDownloadHistoryParams) ([]*database.DownloadHistory, int64, error)

	// 文件管理
	AddFiles(ctx context.Context, files []*database.DownloadFile) error
	GetFilesByHash(ctx context.Context, hash string, state *int) ([]*database.DownloadFile, error)
	GetFileByFullPath(ctx context.Context, fullPath string) (*database.DownloadFile, error)
	GetFilesByFullPath(ctx context.Context, fullPath string) ([]*database.DownloadFile, error)
	GetFilesBySavePath(ctx context.Context, savePath string) ([]*database.DownloadFile, error)
	UpdateFileState(ctx context.Context, fullPath string, state int) error
	TruncateFiles(ctx context.Context) error
}

// ListDownloadHistoryParams 分页查询参数
type ListDownloadHistoryParams struct {
	Page     int
	PageSize int
	Type     string // 类型
	State    string // 状态
	Title    string // 标题 (模糊查询)
}
