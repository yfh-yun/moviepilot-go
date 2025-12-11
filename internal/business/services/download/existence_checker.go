package download

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
)

// ExistenceChecker 媒体库存在性检查器
type ExistenceChecker struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewExistenceChecker 创建存在性检查器
func NewExistenceChecker(db *gorm.DB, logger *zap.Logger) *ExistenceChecker {
	return &ExistenceChecker{
		db:     db,
		logger: logger,
	}
}

// GetNoExistsInfo 检查媒体库，查询是否存在
// 对于剧集同时返回不存在的季集信息
// 返回：当前媒体是否缺失，各标题总的季集和缺失的季集
func (c *ExistenceChecker) GetNoExistsInfo(ctx context.Context, meta *dto.MetaInfo, mediaInfo *dto.MediaInfo) (bool, map[int][]int, map[int][]int, error) {
	c.logger.Info("检查媒体库存在性",
		zap.String("title", mediaInfo.Title),
		zap.String("type", mediaInfo.Type),
	)

	// 总的季集信息
	totalSeasons := make(map[int][]int)
	// 缺失的季集信息
	lackSeasons := make(map[int][]int)

	// 如果是电影
	if mediaInfo.Type == "电影" || mediaInfo.Type == "movie" {
		exists, err := c.checkMovieExists(ctx, mediaInfo)
		if err != nil {
			return false, totalSeasons, lackSeasons, err
		}
		return !exists, totalSeasons, lackSeasons, nil
	}

	// 如果是电视剧
	if mediaInfo.Type == "电视剧" || mediaInfo.Type == "tv" {
		return c.checkTVExists(ctx, meta, mediaInfo)
	}

	c.logger.Warn("未知的媒体类型", zap.String("type", mediaInfo.Type))
	return true, totalSeasons, lackSeasons, nil
}

// checkMovieExists 检查电影是否存在
func (c *ExistenceChecker) checkMovieExists(ctx context.Context, mediaInfo *dto.MediaInfo) (bool, error) {
	c.logger.Debug("检查电影是否存在", zap.String("title", mediaInfo.Title))

	var count int64

	// 通过TMDB ID查询
	if mediaInfo.TmdbID != nil {
		err := c.db.WithContext(ctx).
			Model(&database.MediaServerItem{}).
			Where("tmdb_id = ? AND type = ?", *mediaInfo.TmdbID, "movie").
			Count(&count).Error

		if err != nil {
			c.logger.Error("查询电影失败", zap.Error(err))
			return false, err
		}

		if count > 0 {
			c.logger.Info("电影已存在", zap.String("title", mediaInfo.Title))
			return true, nil
		}
	}

	// 通过标题和年份查询
	query := c.db.WithContext(ctx).
		Model(&database.MediaServerItem{}).
		Where("title = ? AND type = ?", mediaInfo.Title, "movie")

	if mediaInfo.Year != "" {
		query = query.Where("year = ?", mediaInfo.Year)
	}

	err := query.Count(&count).Error
	if err != nil {
		c.logger.Error("查询电影失败", zap.Error(err))
		return false, err
	}

	exists := count > 0
	c.logger.Info("电影存在性检查完成",
		zap.String("title", mediaInfo.Title),
		zap.Bool("exists", exists),
	)

	return exists, nil
}

// checkTVExists 检查电视剧是否存在
func (c *ExistenceChecker) checkTVExists(ctx context.Context, meta *dto.MetaInfo, mediaInfo *dto.MediaInfo) (bool, map[int][]int, map[int][]int, error) {
	c.logger.Debug("检查电视剧是否存在", zap.String("title", mediaInfo.Title))

	totalSeasons := make(map[int][]int)
	lackSeasons := make(map[int][]int)

	// 获取媒体信息中的季集信息
	if mediaInfo.Seasons != nil {
		for season, episodes := range mediaInfo.Seasons {
			totalSeasons[season] = episodes
		}
	}

	// 如果元数据中有季信息
	if meta != nil && meta.BeginSeason != nil {
		season := *meta.BeginSeason

		// 获取该季的总集数
		var totalEpisodes []int
		if episodes, ok := mediaInfo.Seasons[season]; ok {
			totalEpisodes = episodes
		} else if mediaInfo.NumberOfEpisodes > 0 {
			// 如果没有具体的集数信息，使用总集数
			for i := 1; i <= mediaInfo.NumberOfEpisodes; i++ {
				totalEpisodes = append(totalEpisodes, i)
			}
		}

		if len(totalEpisodes) > 0 {
			totalSeasons[season] = totalEpisodes

			// 查询已存在的集
			existingEpisodes, err := c.getExistingEpisodes(ctx, mediaInfo, season)
			if err != nil {
				return false, totalSeasons, lackSeasons, err
			}

			// 计算缺失的集
			lackEpisodes := c.calculateLackEpisodes(totalEpisodes, existingEpisodes)
			if len(lackEpisodes) > 0 {
				lackSeasons[season] = lackEpisodes
			}
		}
	}

	// 判断是否有缺失
	hasLack := len(lackSeasons) > 0

	c.logger.Info("电视剧存在性检查完成",
		zap.String("title", mediaInfo.Title),
		zap.Bool("has_lack", hasLack),
		zap.Int("lack_season_count", len(lackSeasons)),
	)

	return hasLack, totalSeasons, lackSeasons, nil
}

// getExistingEpisodes 获取已存在的集
func (c *ExistenceChecker) getExistingEpisodes(ctx context.Context, mediaInfo *dto.MediaInfo, season int) ([]int, error) {
	c.logger.Debug("查询已存在的集",
		zap.String("title", mediaInfo.Title),
		zap.Int("season", season),
	)

	var items []database.MediaServerItem

	query := c.db.WithContext(ctx).
		Model(&database.MediaServerItem{}).
		Where("type = ? AND season = ?", "episode", season)

	// 通过TMDB ID查询
	if mediaInfo.TmdbID != nil {
		query = query.Where("tmdb_id = ?", *mediaInfo.TmdbID)
	} else {
		// 通过标题查询
		query = query.Where("title = ?", mediaInfo.Title)
	}

	err := query.Find(&items).Error
	if err != nil {
		c.logger.Error("查询已存在的集失败", zap.Error(err))
		return nil, err
	}

	// 提取集数
	episodes := make([]int, 0)
	for _, item := range items {
		if item.Episode != nil {
			episodes = append(episodes, *item.Episode)
		}
	}

	c.logger.Debug("已存在的集",
		zap.String("title", mediaInfo.Title),
		zap.Int("season", season),
		zap.Ints("episodes", episodes),
	)

	return episodes, nil
}

// calculateLackEpisodes 计算缺失的集
func (c *ExistenceChecker) calculateLackEpisodes(total []int, existing []int) []int {
	lack := make([]int, 0)

	existingMap := make(map[int]bool)
	for _, ep := range existing {
		existingMap[ep] = true
	}

	for _, ep := range total {
		if !existingMap[ep] {
			lack = append(lack, ep)
		}
	}

	return lack
}

// CheckDownloadExists 检查下载是否已存在
func (c *ExistenceChecker) CheckDownloadExists(ctx context.Context, hash string) (bool, error) {
	c.logger.Debug("检查下载是否存在", zap.String("hash", hash))

	var count int64
	err := c.db.WithContext(ctx).
		Model(&database.Download{}).
		Where("hash = ?", hash).
		Count(&count).Error

	if err != nil {
		c.logger.Error("查询下载失败", zap.Error(err))
		return false, err
	}

	return count > 0, nil
}
