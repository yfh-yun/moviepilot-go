package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// TransferHistoryRepository 转移历史仓储接口
type TransferHistoryRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, history *database.TransferHistory) error
	GetByID(ctx context.Context, id uint) (*database.TransferHistory, error)
	Update(ctx context.Context, history *database.TransferHistory) error
	Delete(ctx context.Context, id uint) error

	// 查询方法
	GetByTitle(ctx context.Context, title string) ([]*database.TransferHistory, error)
	GetBySrc(ctx context.Context, src string, storage string) (*database.TransferHistory, error)
	GetByDest(ctx context.Context, dest string) (*database.TransferHistory, error)
	ListByHash(ctx context.Context, downloadHash string) ([]*database.TransferHistory, error)

	// 分页查询
	ListByPage(ctx context.Context, params ListTransferHistoryParams) ([]*database.TransferHistory, int64, error)

	// 统计方法
	Statistics(ctx context.Context, days int) ([]TransferStatistic, error)

	// 复杂查询
	ListByConditions(ctx context.Context, params TransferQueryParams) ([]*database.TransferHistory, error)
}

// ListTransferHistoryParams 分页查询参数
type ListTransferHistoryParams struct {
	Page     int
	PageSize int
	Status   *bool  // 转移状态
	Type     string // 类型
	Title    string // 标题 (模糊查询)
}

// TransferStatistic 转移统计结果
type TransferStatistic struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// TransferQueryParams 复杂查询参数
type TransferQueryParams struct {
	Title    *string
	Year     *string
	Type     *string
	Seasons  *string
	Episodes *string
	TMDBID   *int
	Dest     *string
}
