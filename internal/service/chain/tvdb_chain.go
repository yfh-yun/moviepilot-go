package chain

import (
	"context"
	"errors"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// TvdbChain TVDB处理链
type TvdbChain struct {
	cache        *cache.Cache
	logger       *logger.Logger
	mediaService *service.MediaService
}

// NewTvdbChain 创建TVDB处理链实例
func NewTvdbChain(cache *cache.Cache, logger *logger.Logger, mediaService *service.MediaService) *TvdbChain {
	return &TvdbChain{
		cache:        cache,
		logger:       logger,
		mediaService: mediaService,
	}
}

// TvdbSearch 搜索TVDB
func (c *TvdbChain) TvdbSearch(ctx context.Context, query, year string, page int) ([]*model.TvdbMediaInfo, error) {
	c.logger.Info("搜索TVDB", "query", query, "year", year, "page", page)

	// 参数验证
	if query == "" {
		return nil, errors.New("搜索关键词不能为空")
	}

	// 搜索TVDB
	result, err := c.mediaService.SearchTvdb(ctx, query, year, page)
	if err != nil {
		c.logger.Error("搜索TVDB失败", "error", err)
		return nil, err
	}

	c.logger.Info("搜索TVDB成功", "count", len(result))
	return result, nil
}

// TvdbInfo 获取TVDB媒体信息
func (c *TvdbChain) TvdbInfo(ctx context.Context, tvdbID int) (*model.TvdbMediaInfo, error) {
	c.logger.Info("获取TVDB媒体信息", "tvdbID", tvdbID)

	// 获取TVDB信息
	info, err := c.mediaService.GetTvdbInfo(ctx, tvdbID)
	if err != nil {
		c.logger.Error("获取TVDB媒体信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB媒体信息成功", "tvdbID", tvdbID)
	return info, nil
}

// TvdbEpisodes 获取TVDB剧集列表
func (c *TvdbChain) TvdbEpisodes(ctx context.Context, tvdbID, season int) ([]*model.TvdbEpisode, error) {
	c.logger.Info("获取TVDB剧集列表", "tvdbID", tvdbID, "season", season)

	// 获取剧集列表
	episodes, err := c.mediaService.GetTvdbEpisodes(ctx, tvdbID, season)
	if err != nil {
		c.logger.Error("获取TVDB剧集列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB剧集列表成功", "count", len(episodes))
	return episodes, nil
}

// TvdbSeries 获取TVDB系列信息
func (c *TvdbChain) TvdbSeries(ctx context.Context, tvdbID int) (*model.TvdbSeries, error) {
	c.logger.Info("获取TVDB系列信息", "tvdbID", tvdbID)

	// 获取系列信息
	series, err := c.mediaService.GetTvdbSeries(ctx, tvdbID)
	if err != nil {
		c.logger.Error("获取TVDB系列信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB系列信息成功", "tvdbID", tvdbID)
	return series, nil
}

// TvdbActors 获取TVDB演员信息
func (c *TvdbChain) TvdbActors(ctx context.Context, tvdbID int) ([]*model.TvdbActor, error) {
	c.logger.Info("获取TVDB演员信息", "tvdbID", tvdbID)

	// 获取演员信息
	actors, err := c.mediaService.GetTvdbActors(ctx, tvdbID)
	if err != nil {
		c.logger.Error("获取TVDB演员信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB演员信息成功", "count", len(actors))
	return actors, nil
}

// TvdbSeasons 获取TVDB季信息
func (c *TvdbChain) TvdbSeasons(ctx context.Context, tvdbID int) ([]*model.TvdbSeason, error) {
	c.logger.Info("获取TVDB季信息", "tvdbID", tvdbID)

	// 获取季信息
	seasons, err := c.mediaService.GetTvdbSeasons(ctx, tvdbID)
	if err != nil {
		c.logger.Error("获取TVDB季信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB季信息成功", "count", len(seasons))
	return seasons, nil
}

// MatchTmdbToTvdb 匹配TMDB到TVDB
func (c *TvdbChain) MatchTmdbToTvdb(ctx context.Context, tmdbID int, mediaType model.MediaType) (*model.TvdbMediaInfo, error) {
	c.logger.Info("匹配TMDB到TVDB", "tmdbID", tmdbID, "mediaType", mediaType)

	// 匹配ID
	match, err := c.mediaService.MatchTmdbToTvdb(ctx, tmdbID, mediaType)
	if err != nil {
		c.logger.Error("匹配TMDB到TVDB失败", "error", err)
		return nil, err
	}

	c.logger.Info("匹配TMDB到TVDB成功", "tvdbID", match.TvdbID)
	return match, nil
}

// TvdbImages 获取TVDB图片
func (c *TvdbChain) TvdbImages(ctx context.Context, tvdbID int) ([]*model.TvdbImage, error) {
	c.logger.Info("获取TVDB图片", "tvdbID", tvdbID)

	// 获取图片
	images, err := c.mediaService.GetTvdbImages(ctx, tvdbID)
	if err != nil {
		c.logger.Error("获取TVDB图片失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB图片成功", "count", len(images))
	return images, nil
}

// TvdbUpdates 获取TVDB更新
func (c *TvdbChain) TvdbUpdates(ctx context.Context, sinceTime string) ([]*model.TvdbUpdate, error) {
	c.logger.Info("获取TVDB更新", "sinceTime", sinceTime)

	// 获取更新
	updates, err := c.mediaService.GetTvdbUpdates(ctx, sinceTime)
	if err != nil {
		c.logger.Error("获取TVDB更新失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取TVDB更新成功", "count", len(updates))
	return updates, nil
}

// GetTvdbStats 获取TVDB统计信息
func (c *TvdbChain) GetTvdbStats(ctx context.Context) (*model.TvdbStats, error) {
	c.logger.Info("获取TVDB统计信息")

	stats, err := c.mediaService.GetTvdbStats(ctx)
	if err != nil {
		c.logger.Error("获取TVDB统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}
