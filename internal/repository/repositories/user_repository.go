package repositories

import (
	"context"
	"errors"
	"time"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repository/interfaces"
	"moviepilot-go/internal/models"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// List 获取用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, params interfaces.ListUserParams) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{})
	
	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&users).Error
	
	return users, total, err
}

// UpdatePassword 更新密码
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID, password string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", password).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", time.Now()).Error
}