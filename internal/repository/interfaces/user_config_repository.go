package interfaces

import (
	"context"
	"moviepilot-go/internal/models"
)

// UserConfigRepository 用户配置仓储接口
type UserConfigRepository interface {
	// Create 创建用户配置
	Create(ctx context.Context, config *models.UserConfig) error
	
	// GetByID 根据ID获取用户配置
	GetByID(ctx context.Context, id string) (*models.UserConfig, error)
	
	// GetByUserID 根据用户ID获取配置
	GetByUserID(ctx context.Context, userID string) (*models.UserConfig, error)
	
	// Update 更新用户配置
	Update(ctx context.Context, config *models.UserConfig) error
	
	// Delete 删除用户配置
	Delete(ctx context.Context, id string) error
	
	// List 获取用户配置列表
	List(ctx context.Context, params ListUserConfigParams) ([]*models.UserConfig, int64, error)
}

// ListUserConfigParams 用户配置列表查询参数
type ListUserConfigParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	UserID   string `json:"user_id"`
	Key      string `json:"key"`
}