package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// MediaServerRepositoryImpl 媒体服务器仓储实现
type MediaServerRepositoryImpl struct {
	db *gorm.DB
}

// NewMediaServerRepository 创建媒体服务器仓储实例
func NewMediaServerRepository(db *gorm.DB) interfaces.MediaServerRepository {
	return &MediaServerRepositoryImpl{db: db}
}

// Create 创建媒体服务器
func (r *MediaServerRepositoryImpl) Create(ctx context.Context, server *database.MediaServer) error {
	return r.db.WithContext(ctx).Create(server).Error
}

// GetByID 根据ID获取媒体服务器
func (r *MediaServerRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.MediaServer, error) {
	var server database.MediaServer
	err := r.db.WithContext(ctx).First(&server, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// GetByName 根据名称获取媒体服务器
func (r *MediaServerRepositoryImpl) GetByName(ctx context.Context, name string) (*database.MediaServer, error) {
	var server database.MediaServer
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&server).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// Update 更新媒体服务器
func (r *MediaServerRepositoryImpl) Update(ctx context.Context, server *database.MediaServer) error {
	return r.db.WithContext(ctx).Save(server).Error
}

// Delete 删除媒体服务器
func (r *MediaServerRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.MediaServer{}, id).Error
}

// List 获取媒体服务器列表
func (r *MediaServerRepositoryImpl) List(ctx context.Context, params interfaces.ListMediaServerParams) ([]*database.MediaServer, error) {
	var servers []*database.MediaServer
	query := r.db.WithContext(ctx).Model(&database.MediaServer{})

	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	err := query.Find(&servers).Error
	return servers, err
}

// ListActive 获取所有启用的媒体服务器
func (r *MediaServerRepositoryImpl) ListActive(ctx context.Context) ([]*database.MediaServer, error) {
	var servers []*database.MediaServer
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&servers).Error
	return servers, err
}

// ListByType 根据类型获取媒体服务器列表
func (r *MediaServerRepositoryImpl) ListByType(ctx context.Context, serverType string) ([]*database.MediaServer, error) {
	var servers []*database.MediaServer
	err := r.db.WithContext(ctx).
		Where("type = ?", serverType).
		Find(&servers).Error
	return servers, err
}

// UpdateAPIKey 更新API密钥
func (r *MediaServerRepositoryImpl) UpdateAPIKey(ctx context.Context, id uint, apiKey string) error {
	return r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("id = ?", id).
		Update("api_key", apiKey).Error
}

// UpdateSyncLibs 更新同步库配置
func (r *MediaServerRepositoryImpl) UpdateSyncLibs(ctx context.Context, id uint, libs string) error {
	return r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("id = ?", id).
		Update("sync_libs", libs).Error
}

// UpdateSettings 更新设置
func (r *MediaServerRepositoryImpl) UpdateSettings(ctx context.Context, id uint, settings string) error {
	return r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("id = ?", id).
		Update("settings", settings).Error
}

// SetActive 设置启用状态
func (r *MediaServerRepositoryImpl) SetActive(ctx context.Context, id uint, active bool) error {
	return r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}

// Exists 检查媒体服务器是否存在
func (r *MediaServerRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// ExistsByType 检查指定类型的媒体服务器是否存在
func (r *MediaServerRepositoryImpl) ExistsByType(ctx context.Context, serverType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.MediaServer{}).
		Where("type = ?", serverType).
		Count(&count).Error
	return count > 0, err
}
