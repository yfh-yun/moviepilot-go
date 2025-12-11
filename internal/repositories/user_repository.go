package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *models.User) error
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id uint) (*models.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Update 更新用户
	Update(ctx context.Context, user *models.User) error
	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id uint) error
	// List 获取用户列表（分页）
	List(ctx context.Context, page, limit int) ([]*models.User, int64, error)
	// GetWithRoles 获取用户及其角色
	GetWithRoles(ctx context.Context, id uint) (*models.User, error)
	// UpdateLastLogin 更新最后登录信息
	UpdateLastLogin(ctx context.Context, id uint, ip string) error
	// Count 统计用户数量
	Count(ctx context.Context) (int64, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在: id=%d", id)
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在: username=%s", username)
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在: email=%s", email)
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// List 获取用户列表（分页）
func (r *userRepository) List(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 计算总数
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetWithRoles 获取用户及其角色
func (r *userRepository) GetWithRoles(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&user, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在: id=%d", id)
		}
		return nil, err
	}

	return &user, nil
}

// UpdateLastLogin 更新最后登录信息
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint, ip string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at": gorm.Expr("NOW()"),
			"last_login_ip": ip,
		}).Error
}

// Count 统计用户数量
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}
