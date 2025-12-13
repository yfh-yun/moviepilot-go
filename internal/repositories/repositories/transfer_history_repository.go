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

// TransferHistoryRepositoryImpl 转移历史仓储实现
type TransferHistoryRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewTransferHistoryRepository 创建转移历史仓储实例
func NewTransferHistoryRepository(db *gorm.DB) interfaces.TransferHistoryRepository {
	return &TransferHistoryRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建转移历史
func (r *TransferHistoryRepositoryImpl) Create(ctx context.Context, history *database.TransferHistory) error {
	err := r.db.WithContext(ctx).Create(history).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("transfer_history")
	}
	return err
}

// GetByID 根据ID获取转移历史
func (r *TransferHistoryRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.TransferHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("transfer_history:id:%d", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if history, ok := cachedValue.(*database.TransferHistory); ok {
				return history, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var history database.TransferHistory
	err := r.db.WithContext(ctx).First(&history, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &history, 3600*time.Second, "transfer_history")
	}

	return &history, nil
}

// Update 更新转移历史
func (r *TransferHistoryRepositoryImpl) Update(ctx context.Context, history *database.TransferHistory) error {
	err := r.db.WithContext(ctx).Save(history).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("transfer_history")
	}
	return err
}

// Delete 删除转移历史
func (r *TransferHistoryRepositoryImpl) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&database.TransferHistory{}, id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("transfer_history")
	}
	return err
}

// GetByTitle 根据标题查询转移历史
func (r *TransferHistoryRepositoryImpl) GetByTitle(ctx context.Context, title string) ([]*database.TransferHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("transfer_history:title:%s", title)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if histories, ok := cachedValue.([]*database.TransferHistory); ok {
				return histories, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.TransferHistory
	err := r.db.WithContext(ctx).
		Where("title LIKE ?", "%"+title+"%").
		Or("src LIKE ?", "%"+title+"%").
		Or("dest LIKE ?", "%"+title+"%").
		Order("date DESC").
		Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, histories, 3600*time.Second, "transfer_history")
	}

	return histories, nil
}

// GetBySrc 根据源路径查询转移历史
func (r *TransferHistoryRepositoryImpl) GetBySrc(ctx context.Context, src string, storage string) (*database.TransferHistory, error) {
	// 生成缓存键，包含storage参数
	var cacheKey string
	if storage != "" {
		cacheKey = fmt.Sprintf("transfer_history:src:%s:storage:%s", src, storage)
	} else {
		cacheKey = fmt.Sprintf("transfer_history:src:%s", src)
	}

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if history, ok := cachedValue.(*database.TransferHistory); ok {
				return history, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var history database.TransferHistory
	query := r.db.WithContext(ctx).Where("src = ?", src)
	if storage != "" {
		query = query.Where("src_storage = ?", storage)
	}
	err := query.First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &history, 3600*time.Second, "transfer_history")
	}

	return &history, nil
}

// GetByDest 根据目标路径查询转移历史
func (r *TransferHistoryRepositoryImpl) GetByDest(ctx context.Context, dest string) (*database.TransferHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("transfer_history:dest:%s", dest)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if history, ok := cachedValue.(*database.TransferHistory); ok {
				return history, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var history database.TransferHistory
	err := r.db.WithContext(ctx).Where("dest = ?", dest).First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &history, 3600*time.Second, "transfer_history")
	}

	return &history, nil
}

// ListByHash 根据下载hash查询转移历史
func (r *TransferHistoryRepositoryImpl) ListByHash(ctx context.Context, downloadHash string) ([]*database.TransferHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("transfer_history:hash:%s", downloadHash)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if histories, ok := cachedValue.([]*database.TransferHistory); ok {
				return histories, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.TransferHistory
	err := r.db.WithContext(ctx).
		Where("download_hash = ?", downloadHash).
		Order("date DESC").
		Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, histories, 3600*time.Second, "transfer_history")
	}

	return histories, nil
}

// ListByPage 分页查询转移历史
func (r *TransferHistoryRepositoryImpl) ListByPage(ctx context.Context, params interfaces.ListTransferHistoryParams) ([]*database.TransferHistory, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	statusStr := "nil"
	if params.Status != nil {
		statusStr = fmt.Sprintf("%v", *params.Status)
	}
	cacheKey := fmt.Sprintf("transfer_history:page:%d:page_size:%d:status:%s:type:%s:title:%s",
		params.Page, params.PageSize, statusStr, params.Type, params.Title)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Histories []*database.TransferHistory
				Total     int64
			}); ok {
				return cacheData.Histories, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.TransferHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&database.TransferHistory{})

	// 添加过滤条件
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Title != "" {
		query = query.Where("title LIKE ?", "%"+params.Title+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.
		Offset(offset).
		Limit(params.PageSize).
		Order("date DESC").
		Find(&histories).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Histories []*database.TransferHistory
			Total     int64
		}{
			Histories: histories,
			Total:     total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "transfer_history")
	}

	return histories, total, err
}

// Statistics 统计最近N天的转移历史数量
func (r *TransferHistoryRepositoryImpl) Statistics(ctx context.Context, days int) ([]interfaces.TransferStatistic, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("transfer_history:statistics:%d", days)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if results, ok := cachedValue.([]interfaces.TransferStatistic); ok {
				return results, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var results []interfaces.TransferStatistic

	// 计算起始日期
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// 使用原生 SQL 进行统计
	err := r.db.WithContext(ctx).Raw(`
		SELECT DATE(date) as date, COUNT(*) as count
		FROM transfer_histories
		WHERE date >= ?
		GROUP BY DATE(date)
		ORDER BY date DESC
	`, startDate).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, results, 3600*time.Second, "transfer_history")
	}

	return results, nil
}

// ListByConditions 根据复杂条件查询转移历史
func (r *TransferHistoryRepositoryImpl) ListByConditions(ctx context.Context, params interfaces.TransferQueryParams) ([]*database.TransferHistory, error) {
	// 生成缓存键，包含所有条件参数
	title := "nil"
	year := "nil"
	mtype := "nil"
	seasons := "nil"
	episodes := "nil"
	tmdbid := "nil"
	dest := "nil"

	if params.Title != nil {
		title = *params.Title
	}
	if params.Year != nil {
		year = *params.Year
	}
	if params.Type != nil {
		mtype = *params.Type
	}
	if params.Seasons != nil {
		seasons = *params.Seasons
	}
	if params.Episodes != nil {
		episodes = *params.Episodes
	}
	if params.TMDBID != nil {
		tmdbid = fmt.Sprintf("%d", *params.TMDBID)
	}
	if params.Dest != nil {
		dest = *params.Dest
	}

	cacheKey := fmt.Sprintf("transfer_history:conditions:title:%s:year:%s:type:%s:seasons:%s:episodes:%s:tmdbid:%s:dest:%s",
		title, year, mtype, seasons, episodes, tmdbid, dest)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "transfer_history")
		if err == nil && hit {
			if histories, ok := cachedValue.([]*database.TransferHistory); ok {
				return histories, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.TransferHistory

	query := r.db.WithContext(ctx).Model(&database.TransferHistory{})

	// 动态添加查询条件
	if params.Title != nil && *params.Title != "" {
		query = query.Where("title LIKE ?", "%"+*params.Title+"%")
	}
	if params.Year != nil && *params.Year != "" {
		query = query.Where("year = ?", *params.Year)
	}
	if params.Type != nil && *params.Type != "" {
		query = query.Where("type = ?", *params.Type)
	}
	if params.Seasons != nil && *params.Seasons != "" {
		query = query.Where("seasons = ?", *params.Seasons)
	}
	if params.Episodes != nil && *params.Episodes != "" {
		query = query.Where("episodes = ?", *params.Episodes)
	}
	if params.TMDBID != nil {
		query = query.Where("tmdb_id = ?", *params.TMDBID)
	}
	if params.Dest != nil && *params.Dest != "" {
		query = query.Where("dest = ?", *params.Dest)
	}

	err := query.Order("date DESC").Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, histories, 3600*time.Second, "transfer_history")
	}

	return histories, nil
}
