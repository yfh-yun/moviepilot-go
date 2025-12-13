package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &UserRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *database.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user")
	}
	return err
}

// GetByID 根据ID获取用户
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*database.User, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user")
		if err == nil && hit {
			if user, ok := cachedValue.(*database.User); ok {
				return user, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var user database.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &user, 3600*time.Second, "user")
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*database.User, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user:username:%s", username)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user")
		if err == nil && hit {
			if user, ok := cachedValue.(*database.User); ok {
				return user, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var user database.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &user, 3600*time.Second, "user")
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*database.User, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user:email:%s", email)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user")
		if err == nil && hit {
			if user, ok := cachedValue.(*database.User); ok {
				return user, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var user database.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &user, 3600*time.Second, "user")
	}

	return &user, nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *database.User) error {
	err := r.db.WithContext(ctx).Save(user).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user")
	}
	return err
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.User{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user")
	}
	return err
}

// List 获取用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, params interfaces.ListUserParams) ([]*database.User, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("user:list:page:%d:page_size:%d:status:%s",
		params.Page, params.PageSize, params.Status)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Users []*database.User
				Total int64
			}); ok {
				return cacheData.Users, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var users []*database.User
	var total int64

	query := r.db.WithContext(ctx).Model(&database.User{})

	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, struct {
			Users []*database.User
			Total int64
		}{Users: users, Total: total}, 3600*time.Second, "user")
	}

	return users, total, err
}

// UpdatePassword 更新密码
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID, password string) error {
	err := r.db.WithContext(ctx).Model(&database.User{}).Where("id = ?", userID).Update("hashed_password", password).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user")
	}
	return err
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID string) error {
	// User模型中没有last_login_at字段，直接返回nil
	return nil
}

// HasAny 检查是否存在任何用户
func (r *UserRepositoryImpl) HasAny(ctx context.Context) (bool, error) {
	// 生成缓存键
	cacheKey := "user:has_any"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user")
		if err == nil && hit {
			if hasAny, ok := cachedValue.(bool); ok {
				return hasAny, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var count int64
	err := r.db.WithContext(ctx).Model(&database.User{}).Count(&count).Error
	if err != nil {
		return false, err
	}

	hasAny := count > 0

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, hasAny, 3600*time.Second, "user")
	}

	return hasAny, nil
}
