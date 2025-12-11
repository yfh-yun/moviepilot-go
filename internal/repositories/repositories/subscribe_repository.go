package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// SubscribeRepositoryImpl 订阅仓储实现
type SubscribeRepositoryImpl struct {
	db *gorm.DB
}

// NewSubscribeRepository 创建订阅仓储实例
func NewSubscribeRepository(db *gorm.DB) interfaces.SubscribeRepository {
	return &SubscribeRepositoryImpl{db: db}
}

// Create 创建订阅
func (r *SubscribeRepositoryImpl) Create(ctx context.Context, subscribe *database.Subscribe) error {
	return r.db.WithContext(ctx).Create(subscribe).Error
}

// GetByID 根据ID获取订阅
func (r *SubscribeRepositoryImpl) GetByID(ctx context.Context, id string) (*database.Subscribe, error) {
	var subscribe database.Subscribe
	err := r.db.WithContext(ctx).First(&subscribe, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// Update 更新订阅
func (r *SubscribeRepositoryImpl) Update(ctx context.Context, subscribe *database.Subscribe) error {
	return r.db.WithContext(ctx).Save(subscribe).Error
}

// Delete 删除订阅
func (r *SubscribeRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&database.Subscribe{}, "id = ?", id).Error
}

// List 获取订阅列表
func (r *SubscribeRepositoryImpl) List(ctx context.Context, params interfaces.ListSubscribeParams) ([]*database.Subscribe, int64, error) {
	var subscribes []*database.Subscribe
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Subscribe{})

	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&subscribes).Error

	return subscribes, total, err
}

// GetByUserID 根据用户ID获取订阅列表
func (r *SubscribeRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subscribes).Error
	return subscribes, err
}

// GetActiveSubscriptions 获取活跃订阅
func (r *SubscribeRepositoryImpl) GetActiveSubscriptions(ctx context.Context) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&subscribes).Error
	return subscribes, err
}

// Exists 检查订阅是否存在
func (r *SubscribeRepositoryImpl) Exists(ctx context.Context, tmdbID *int, doubanID *string, season *int) (bool, error) {
	query := r.db.WithContext(ctx).Model(&database.Subscribe{})

	// 构建查询条件
	if tmdbID != nil {
		query = query.Where("tmdb_id = ?", *tmdbID)
	}
	if doubanID != nil && *doubanID != "" {
		query = query.Where("douban_id = ?", *doubanID)
	}
	if season != nil {
		query = query.Where("season = ?", *season)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// GetByTMDBID 根据TMDB ID获取订阅
func (r *SubscribeRepositoryImpl) GetByTMDBID(ctx context.Context, tmdbID int, season *int) (*database.Subscribe, error) {
	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("tmdb_id = ?", tmdbID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByDoubanID 根据豆瓣ID获取订阅
func (r *SubscribeRepositoryImpl) GetByDoubanID(ctx context.Context, doubanID string, season *int) (*database.Subscribe, error) {
	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("douban_id = ?", doubanID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// ListByState 根据状态查询订阅
func (r *SubscribeRepositoryImpl) ListByState(ctx context.Context, state string) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", state).
		Order("last_update DESC").
		Find(&subscribes).Error
	return subscribes, err
}

// ListActive 查询活跃订阅（状态为R的订阅）
func (r *SubscribeRepositoryImpl) ListActive(ctx context.Context) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", "R").
		Order("last_update DESC").
		Find(&subscribes).Error
	return subscribes, err
}

// Statistics 统计订阅信息
func (r *SubscribeRepositoryImpl) Statistics(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 统计总数
	var total int64
	if err := r.db.WithContext(ctx).Model(&database.Subscribe{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	var stateStats []struct {
		State string
		Count int64
	}
	err := r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("state, COUNT(*) as count").
		Group("state").
		Scan(&stateStats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range stateStats {
		stats["state_"+stat.State] = stat.Count
	}

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	err = r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range typeStats {
		stats["type_"+stat.Type] = stat.Count
	}

	return stats, nil
}

// ListByType 根据类型查询订阅
func (r *SubscribeRepositoryImpl) ListByType(ctx context.Context, mediaType string) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("type = ?", mediaType).
		Order("last_update DESC").
		Find(&subscribes).Error
	return subscribes, err
}

// UpdateState 更新订阅状态
func (r *SubscribeRepositoryImpl) UpdateState(ctx context.Context, id uint, state string) error {
	return r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Where("id = ?", id).
		Update("state", state).Error
}
