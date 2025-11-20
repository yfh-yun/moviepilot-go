package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(user *model.User) error

	// Update 更新用户
	Update(user *model.User) error

	// Delete 删除用户
	Delete(id uint) error

	// GetByID 根据ID获取用户
	GetByID(id uint) (*model.User, error)

	// FindByUsername 根据用户名获取用户
	FindByUsername(username string) (*model.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(email string) (*model.User, error)

	// List 获取用户列表
	List(offset, limit int) ([]*model.User, int64, error)

	// UpdateOTP 更新OTP设置
	UpdateOTP(id uint, isOTP bool, secret string) error

	// UpdatePassword 更新密码
	UpdatePassword(id uint, passwordHash string) error

	// UpdateAvatar 更新头像
	UpdateAvatar(id uint, avatar string) error

	// UpdateSettings 更新用户设置
	UpdateSettings(id uint, settings string) error

	// UpdatePermissions 更新用户权限
	UpdatePermissions(id uint, permissions string) error

	// Count 统计用户数量
	Count() (int64, error)

	// Exists 检查用户名是否存在
	Exists(username string) (bool, error)

	// ExistsEmail 检查邮箱是否存在
	ExistsEmail(email string) (bool, error)
}
