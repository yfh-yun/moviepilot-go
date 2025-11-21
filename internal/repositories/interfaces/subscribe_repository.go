package interfaces

import (
	"context"
	"moviepilot-go/internal/models"
)

// SubscribeRepository 订阅仓储接口
type SubscribeRepository interface {
	// Create 创建订阅
	Create(ctx context.Context, subscribe *models.Subscribe) error
	
	// GetByID 根据ID获取订阅
	GetByID(ctx context.Context, id string) (*models.Subscribe, error)
	
	// Update 更新订阅
	Update(ctx context.Context, subscribe *models.Subscribe) error
	
	// Delete 删除订阅
	Delete(ctx context.Context, id string) error
	
	// List 获取订阅列表
	List(ctx context.Context, params ListSubscribeParams) ([]*models.Subscribe, int64, error)
	
	// GetByUserID 根据用户ID获取订阅列表
	GetByUserID(ctx context.Context, userID string) ([]*models.Subscribe, error)
	
	// GetActiveSubscriptions 获取活跃订阅
	GetActiveSubscriptions(ctx context.Context) ([]*models.Subscribe, error)
}

// ListSubscribeParams 订阅列表查询参数
type ListSubscribeParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	UserID   string `json:"user_id"`
}