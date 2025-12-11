package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *database.User) error

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id string) (*database.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*database.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*database.User, error)

	// Update 更新用户
	Update(ctx context.Context, user *database.User) error

	// Delete 删除用户
	Delete(ctx context.Context, id string) error

	// List 获取用户列表
	List(ctx context.Context, params ListUserParams) ([]*database.User, int64, error)

	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, userID, password string) error

	// UpdateLastLogin 更新最后登录时间
	UpdateLastLogin(ctx context.Context, userID string) error

	// HasAny 检查是否存在任何用户
	HasAny(ctx context.Context) (bool, error)
}

// ListUserParams 用户列表查询参数
type ListUserParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	Role     string `json:"role"`
}
