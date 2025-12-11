package repositories

import (
	"context"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/entity"
)

// SubscribeShareRepository 订阅分享数据仓库接口
type SubscribeShareRepository interface {
	// Create 创建分享
	Create(ctx context.Context, share *entity.SubscribeShare) error

	// Delete 删除分享
	Delete(ctx context.Context, id uint) error

	// GetByID 根据ID获取分享
	GetByID(ctx context.Context, id uint) (*entity.SubscribeShare, error)

	// GetAll 获取所有分享
	GetAll(ctx context.Context, offset, limit int) ([]*entity.SubscribeShare, error)

	// GetByUser 根据用户获取分享
	GetByUser(ctx context.Context, userID string) ([]*entity.SubscribeShare, error)

	// IncrementForkCount 增加复用次数
	IncrementForkCount(ctx context.Context, id uint) error

	// GetPopular 获取热门分享
	GetPopular(ctx context.Context, limit int) ([]*entity.SubscribeShare, error)
}

// UserFollowRepository 用户关注数据仓库接口
type UserFollowRepository interface {
	// Create 创建关注
	Create(ctx context.Context, follow *entity.UserFollow) error

	// Delete 删除关注
	Delete(ctx context.Context, userID, followUID string) error

	// GetFollowedUsers 获取关注的用户列表
	GetFollowedUsers(ctx context.Context, userID string) ([]string, error)

	// IsFollowing 检查是否关注
	IsFollowing(ctx context.Context, userID, followUID string) (bool, error)
}

// subscribeShareRepository 订阅分享仓库实现
type subscribeShareRepository struct {
	db *gorm.DB
}

// NewSubscribeShareRepository 创建订阅分享仓库
func NewSubscribeShareRepository(db *gorm.DB) SubscribeShareRepository {
	return &subscribeShareRepository{db: db}
}

// Create 创建分享
func (r *subscribeShareRepository) Create(ctx context.Context, share *entity.SubscribeShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// Delete 删除分享
func (r *subscribeShareRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.SubscribeShare{}, id).Error
}

// GetByID 根据ID获取分享
func (r *subscribeShareRepository) GetByID(ctx context.Context, id uint) (*entity.SubscribeShare, error) {
	var share entity.SubscribeShare
	err := r.db.WithContext(ctx).First(&share, id).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetAll 获取所有分享
func (r *subscribeShareRepository) GetAll(ctx context.Context, offset, limit int) ([]*entity.SubscribeShare, error) {
	var shares []*entity.SubscribeShare
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&shares).Error
	return shares, err
}

// GetByUser 根据用户获取分享
func (r *subscribeShareRepository) GetByUser(ctx context.Context, userID string) ([]*entity.SubscribeShare, error) {
	var shares []*entity.SubscribeShare
	err := r.db.WithContext(ctx).
		Where("share_user = ?", userID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// IncrementForkCount 增加复用次数
func (r *subscribeShareRepository) IncrementForkCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.SubscribeShare{}).
		Where("id = ?", id).
		UpdateColumn("fork_count", gorm.Expr("fork_count + ?", 1)).Error
}

// GetPopular 获取热门分享
func (r *subscribeShareRepository) GetPopular(ctx context.Context, limit int) ([]*entity.SubscribeShare, error) {
	var shares []*entity.SubscribeShare
	err := r.db.WithContext(ctx).
		Order("fork_count DESC, like_count DESC").
		Limit(limit).
		Find(&shares).Error
	return shares, err
}

// userFollowRepository 用户关注仓库实现
type userFollowRepository struct {
	db *gorm.DB
}

// NewUserFollowRepository 创建用户关注仓库
func NewUserFollowRepository(db *gorm.DB) UserFollowRepository {
	return &userFollowRepository{db: db}
}

// Create 创建关注
func (r *userFollowRepository) Create(ctx context.Context, follow *entity.UserFollow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

// Delete 删除关注
func (r *userFollowRepository) Delete(ctx context.Context, userID, followUID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND follow_uid = ?", userID, followUID).
		Delete(&entity.UserFollow{}).Error
}

// GetFollowedUsers 获取关注的用户列表
func (r *userFollowRepository) GetFollowedUsers(ctx context.Context, userID string) ([]string, error) {
	var follows []*entity.UserFollow
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&follows).Error
	if err != nil {
		return nil, err
	}

	uids := make([]string, 0, len(follows))
	for _, follow := range follows {
		uids = append(uids, follow.FollowUID)
	}
	return uids, nil
}

// IsFollowing 检查是否关注
func (r *userFollowRepository) IsFollowing(ctx context.Context, userID, followUID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.UserFollow{}).
		Where("user_id = ? AND follow_uid = ?", userID, followUID).
		Count(&count).Error
	return count > 0, err
}
