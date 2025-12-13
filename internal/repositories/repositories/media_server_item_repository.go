package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// MediaServerItemRepositoryImpl 媒体服务器项目仓储实现
type MediaServerItemRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewMediaServerItemRepository 创建媒体服务器项目仓储实例
func NewMediaServerItemRepository(db *gorm.DB) interfaces.MediaServerItemRepository {
	// 初始化缓存，使用TTL缓存，1000个条目，3600秒过期时间
	cacheBackend := cache.Cache("ttl", 1000, 3600)
	return &MediaServerItemRepositoryImpl{
		db:    db,
		cache: cacheBackend,
	}
}

// Create 创建媒体服务器项目
func (r *MediaServerItemRepositoryImpl) Create(ctx context.Context, item *database.MediaServerItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetByID 根据ID获取媒体服务器项目
func (r *MediaServerItemRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.MediaServerItem, error) {
	var item database.MediaServerItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// Update 更新媒体服务器项目
func (r *MediaServerItemRepositoryImpl) Update(ctx context.Context, item *database.MediaServerItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete 删除媒体服务器项目
func (r *MediaServerItemRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.MediaServerItem{}, id).Error
}

// GetByItemID 根据ItemID获取媒体服务器项目
func (r *MediaServerItemRepositoryImpl) GetByItemID(ctx context.Context, itemID string) (*database.MediaServerItem, error) {
	// 边界条件处理：空ItemID
	if itemID == "" {
		return nil, nil
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("item_id:%s", itemID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "mediaserveritem")
		if err == nil && hit {
			if item, ok := cachedValue.(*database.MediaServerItem); ok {
				return item, nil
			}
		}
		// 缓存操作失败不影响正常流程
	}

	// 缓存未命中，查询数据库
	var item database.MediaServerItem
	err := r.db.WithContext(ctx).Where("item_id = ?", itemID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		r.cache.Set(cacheKey, &item, 0, "mediaserveritem")
	}

	return &item, nil
}

// ExistByTMDBID 根据TMDBID和类型检查媒体服务器项目是否存在
func (r *MediaServerItemRepositoryImpl) ExistByTMDBID(ctx context.Context, tmdbID int, itemType string) (*database.MediaServerItem, error) {
	// 边界条件处理：无效TMDBID或空类型
	if tmdbID <= 0 || itemType == "" {
		return nil, nil
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("tmdbid:%d:type:%s", tmdbID, itemType)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "mediaserveritem")
		if err == nil && hit {
			if item, ok := cachedValue.(*database.MediaServerItem); ok {
				return item, nil
			}
		}
		// 缓存操作失败不影响正常流程
	}

	// 缓存未命中，查询数据库
	var item database.MediaServerItem
	err := r.db.WithContext(ctx).Where("tmdbid = ? AND type = ?", tmdbID, itemType).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		r.cache.Set(cacheKey, &item, 0, "mediaserveritem")
	}

	return &item, nil
}

// ExistsByTitle 根据标题、类型和年份检查媒体服务器项目是否存在
func (r *MediaServerItemRepositoryImpl) ExistsByTitle(ctx context.Context, title, itemType string, year int) (*database.MediaServerItem, error) {
	// 边界条件处理：空标题或空类型
	if title == "" || itemType == "" {
		return nil, nil
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("title:%s:type:%s:year:%d", title, itemType, year)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "mediaserveritem")
		if err == nil && hit {
			if item, ok := cachedValue.(*database.MediaServerItem); ok {
				return item, nil
			}
		}
		// 缓存操作失败不影响正常流程
	}

	// 缓存未命中，查询数据库
	var item database.MediaServerItem
	err := r.db.WithContext(ctx).Where("title = ? AND type = ? AND year = ?", title, itemType, year).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		r.cache.Set(cacheKey, &item, 0, "mediaserveritem")
	}

	return &item, nil
}

// Empty 清空媒体服务器项目
func (r *MediaServerItemRepositoryImpl) Empty(ctx context.Context, server *string) error {
	query := r.db.WithContext(ctx).Model(&database.MediaServerItem{})
	if server != nil {
		query = query.Where("server = ?", *server)
	}

	// 执行数据库删除
	err := query.Delete(&database.MediaServerItem{}).Error
	if err != nil {
		return err
	}

	// 清空相关缓存
	if r.cache != nil {
		r.cache.Clear("mediaserveritem")
	}

	return nil
}
