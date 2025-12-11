package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SubscribeHistoryRepository 订阅历史仓储接口
type SubscribeHistoryRepository interface {
	// Create 创建订阅历史
	Create(ctx context.Context, history *database.SubscribeHistory) error

	// GetByID 根据ID获取订阅历史
	GetByID(ctx context.Context, id string) (*database.SubscribeHistory, error)

	// GetBySubscribeID 根据订阅ID获取历史
	GetBySubscribeID(ctx context.Context, subscribeID string, params ListSubscribeHistoryParams) ([]*database.SubscribeHistory, int64, error)

	// Update 更新订阅历史
	Update(ctx context.Context, history *database.SubscribeHistory) error

	// Delete 删除订阅历史
	Delete(ctx context.Context, id string) error

	// List 获取订阅历史列表
	List(ctx context.Context, params ListSubscribeHistoryParams) ([]*database.SubscribeHistory, int64, error)

	// GetLatestBySubscribeID 获取订阅的最新历史
	GetLatestBySubscribeID(ctx context.Context, subscribeID string) (*database.SubscribeHistory, error)
}

// ListSubscribeHistoryParams 订阅历史列表查询参数
type ListSubscribeHistoryParams struct {
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	SubscribeID string `json:"subscribe_id"`
	Status      string `json:"status"`
	DateFrom    string `json:"date_from"`
	DateTo      string `json:"date_to"`
}
