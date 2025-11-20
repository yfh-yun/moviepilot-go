package chain

import (
	"context"
	"math/rand"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// TmdbChain TMDB处理链
type TmdbChain struct {
	cache  *cache.Cache
	logger *logger.Logger
}

// NewTmdbChain 创建TMDB处理链实例
func NewTmdbChain(cache *cache.Cache, logger *logger.Logger) *TmdbChain {
	return &TmdbChain{
		cache:  cache,
		logger: logger,
	}
}

// TmdbDiscover 发现TMDB电影、剧集
func (c *TmdbChain) TmdbDiscover(ctx context.Context, mtype model.MediaType, sortBy, withGenres, withOriginalLanguage, withKeywords, withWatchProviders string, voteAverage float64, voteCount int, releaseDate string, page int) ([]*model.MediaInfo, error) {
	c.logger.Info("执行TMDB发现模块",
		"mediaType", mtype,
		"sortBy", sortBy,
		"page", page)

	result, err := service.NewMediaService(c.logger).DiscoverMedia(ctx, model.DiscoverParams{
		MediaType:            mtype,
		SortBy:               sortBy,
		WithGenres:           withGenres,
		WithOriginalLanguage: withOriginalLanguage,
		WithKeywords:         withKeywords,
		WithWatchProviders:   withWatchProviders,
		VoteAverage:          voteAverage,
		VoteCount:            voteCount,
		ReleaseDate:          releaseDate,
		Page:                 page,
	})

	if err != nil {
		c.logger.Error("TMDB发现失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TmdbTrending 获取TMDB流行趋势
func (c *TmdbChain) TmdbTrending(ctx context.Context, page int) ([]*model.MediaInfo, error) {
	c.logger.Info("获取TMDB流行趋势", "page", page)

	result, err := service.NewMediaService(c.logger).GetTrendingMedia(ctx, page)
	if err != nil {
		c.logger.Error("获取TMDB流行趋势失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TmdbCollection 根据合集ID查询集合
func (c *TmdbChain) TmdbCollection(ctx context.Context, collectionID int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询TMDB集合", "collectionID", collectionID)

	result, err := service.NewMediaService(c.logger).GetCollection(ctx, collectionID)
	if err != nil {
		c.logger.Error("查询TMDB集合失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TmdbSeasons 根据TMDBID查询所有季信息
func (c *TmdbChain) TmdbSeasons(ctx context.Context, tmdbID int) ([]*model.TmdbSeason, error) {
	c.logger.Info("查询TMDB季信息", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetSeasons(ctx, tmdbID)
	if err != nil {
		c.logger.Error("查询TMDB季信息失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TmdbEpisodes 根据TMDBID查询某季的所有集信息
func (c *TmdbChain) TmdbEpisodes(ctx context.Context, tmdbID, season int, episodeGroup string) ([]*model.TmdbEpisode, error) {
	c.logger.Info("查询TMDB集信息", "tmdbID", tmdbID, "season", season)

	result, err := service.NewMediaService(c.logger).GetEpisodes(ctx, tmdbID, season, episodeGroup)
	if err != nil {
		c.logger.Error("查询TMDB集信息失败", "error", err)
		return nil, err
	}

	return result, nil
}

// MovieSimilar 根据TMDBID查询类似电影
func (c *TmdbChain) MovieSimilar(ctx context.Context, tmdbID int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询类似电影", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetSimilarMovies(ctx, tmdbID)
	if err != nil {
		c.logger.Error("查询类似电影失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TvSimilar 根据TMDBID查询类似电视剧
func (c *TmdbChain) TvSimilar(ctx context.Context, tmdbID int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询类似电视剧", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetSimilarTVShows(ctx, tmdbID)
	if err != nil {
		c.logger.Error("查询类似电视剧失败", "error", err)
		return nil, err
	}

	return result, nil
}

// MovieRecommend 根据TMDBID查询推荐电影
func (c *TmdbChain) MovieRecommend(ctx context.Context, tmdbID int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询推荐电影", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetRecommendedMovies(ctx, tmdbID)
	if err != nil {
		c.logger.Error("查询推荐电影失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TvRecommend 根据TMDBID查询推荐电视剧
func (c *TmdbChain) TvRecommend(ctx context.Context, tmdbID int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询推荐电视剧", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetRecommendedTVShows(ctx, tmdbID)
	if err != nil {
		c.logger.Error("查询推荐电视剧失败", "error", err)
		return nil, err
	}

	return result, nil
}

// MovieCredits 根据TMDBID查询电影演职人员
func (c *TmdbChain) MovieCredits(ctx context.Context, tmdbID, page int) ([]*model.MediaPerson, error) {
	c.logger.Info("查询电影演职人员", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetMovieCredits(ctx, tmdbID, page)
	if err != nil {
		c.logger.Error("查询电影演职人员失败", "error", err)
		return nil, err
	}

	return result, nil
}

// TvCredits 根据TMDBID查询电视剧演职人员
func (c *TmdbChain) TvCredits(ctx context.Context, tmdbID, page int) ([]*model.MediaPerson, error) {
	c.logger.Info("查询电视剧演职人员", "tmdbID", tmdbID)

	result, err := service.NewMediaService(c.logger).GetTVCredits(ctx, tmdbID, page)
	if err != nil {
		c.logger.Error("查询电视剧演职人员失败", "error", err)
		return nil, err
	}

	return result, nil
}

// PersonDetail 根据人物ID查询演职员详情
func (c *TmdbChain) PersonDetail(ctx context.Context, personID int) (*model.MediaPerson, error) {
	c.logger.Info("查询演职员详情", "personID", personID)

	result, err := service.NewMediaService(c.logger).GetPersonDetail(ctx, personID)
	if err != nil {
		c.logger.Error("查询演职员详情失败", "error", err)
		return nil, err
	}

	return result, nil
}

// PersonCredits 根据人物ID查询人物参演作品
func (c *TmdbChain) PersonCredits(ctx context.Context, personID, page int) ([]*model.MediaInfo, error) {
	c.logger.Info("查询人物参演作品", "personID", personID)

	result, err := service.NewMediaService(c.logger).GetPersonCredits(ctx, personID, page)
	if err != nil {
		c.logger.Error("查询人物参演作品失败", "error", err)
		return nil, err
	}

	return result, nil
}

// GetRandomWallpaper 获取随机壁纸
func (c *TmdbChain) GetRandomWallpaper(ctx context.Context) (string, error) {
	c.logger.Info("获取随机壁纸")

	// 获取流行趋势
	infos, err := c.TmdbTrending(ctx, 1)
	if err != nil {
		return "", err
	}

	// 随机选择
	if len(infos) > 0 {
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < len(infos); i++ {
			info := infos[rand.Intn(len(infos))]
			if info != nil && info.BackdropPath != "" {
				return info.BackdropPath, nil
			}
		}
	}

	return "", nil
}

// GetTrendingWallpapers 获取所有流行壁纸
func (c *TmdbChain) GetTrendingWallpapers(ctx context.Context, num int) ([]string, error) {
	c.logger.Info("获取流行壁纸", "num", num)

	infos, err := c.TmdbTrending(ctx, 1)
	if err != nil {
		return nil, err
	}

	var wallpapers []string
	for _, info := range infos {
		if info != nil && info.BackdropPath != "" {
			wallpapers = append(wallpapers, info.BackdropPath)
			if len(wallpapers) >= num {
				break
			}
		}
	}

	return wallpapers, nil
}
