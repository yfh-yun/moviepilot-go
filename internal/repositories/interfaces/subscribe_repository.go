package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SubscribeRepository 订阅仓储接口
type SubscribeRepository interface {
	// Create 创建订阅
	Create(ctx context.Context, subscribe *database.Subscribe) error

	// GetByID 根据ID获取订阅
	GetByID(ctx context.Context, id string) (*database.Subscribe, error)

	// Update 更新订阅
	Update(ctx context.Context, subscribe *database.Subscribe) error

	// Delete 删除订阅
	Delete(ctx context.Context, id string) error

	// List 获取订阅列表
	List(ctx context.Context, params ListSubscribeParams) ([]*database.Subscribe, int64, error)

	// GetByUserID 根据用户ID获取订阅列表
	GetByUserID(ctx context.Context, userID string) ([]*database.Subscribe, error)

	// GetActiveSubscriptions 获取活跃订阅
	GetActiveSubscriptions(ctx context.Context) ([]*database.Subscribe, error)

	// 新增方法
	// Exists 检查订阅是否存在
	Exists(ctx context.Context, tmdbID *int, doubanID *string, season *int) (bool, error)

	// GetByTMDBID 根据TMDB ID获取订阅
	GetByTMDBID(ctx context.Context, tmdbID int, season *int) (*database.Subscribe, error)

	// GetByDoubanID 根据豆瓣ID获取订阅
	GetByDoubanID(ctx context.Context, doubanID string, season *int) (*database.Subscribe, error)

	// ListByState 根据状态查询订阅
	ListByState(ctx context.Context, state string) ([]*database.Subscribe, error)

	// ListActive 查询活跃订阅（状态为R的订阅）
	ListActive(ctx context.Context) ([]*database.Subscribe, error)

	// Statistics 统计订阅信息
	Statistics(ctx context.Context) (map[string]int64, error)

	// ListByType 根据类型查询订阅
	ListByType(ctx context.Context, mediaType string) ([]*database.Subscribe, error)

	// UpdateState 更新订阅状态
	UpdateState(ctx context.Context, id uint, state string) error
}

// ListSubscribeParams 订阅列表查询参数
type ListSubscribeParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	UserID   string `json:"user_id"`
}
