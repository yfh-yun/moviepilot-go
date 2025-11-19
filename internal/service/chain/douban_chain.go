package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// DoubanChain 豆瓣处理链
type DoubanChain struct {
	cache      *cache.Cache
	logger     *logger.Logger
	mediaChain *MediaChain
}

// NewDoubanChain 创建豆瓣处理链
func NewDoubanChain(cache *cache.Cache, logger *logger.Logger, mediaChain *MediaChain) *DoubanChain {
	return &DoubanChain{
		cache:      cache,
		logger:     logger,
		mediaChain: mediaChain,
	}
}

// PersonDetail 根据人物ID查询豆瓣人物详情
func (dc *DoubanChain) PersonDetail(ctx context.Context, personID int) (*model.MediaPerson, error) {
	cacheKey := fmt.Sprintf("douban:person:%d", personID)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if person, ok := cached.(*model.MediaPerson); ok {
			return person, nil
		}
	}

	dc.logger.Info("开始获取豆瓣人物详情", "personID", personID)

	// 调用豆瓣API获取人物详情
	person, err := dc.fetchPersonDetail(ctx, personID)
	if err != nil {
		dc.logger.Error("获取豆瓣人物详情失败", "personID", personID, "error", err)
		return nil, err
	}

	// 缓存结果（24小时）
	dc.cache.Set(cacheKey, person, 24*time.Hour)

	dc.logger.Info("成功获取豆瓣人物详情", "personID", personID)
	return person, nil
}

// PersonCredits 根据人物ID查询人物参演作品
func (dc *DoubanChain) PersonCredits(ctx context.Context, personID int, page int) ([]*model.MediaInfo, error) {
	cacheKey := fmt.Sprintf("douban:person_credits:%d:%d", personID, page)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if mediaList, ok := cached.([]*model.MediaInfo); ok {
			return mediaList, nil
		}
	}

	dc.logger.Info("开始获取豆瓣人物参演作品", "personID", personID, "page", page)

	// 调用豆瓣API获取人物参演作品
	mediaList, err := dc.fetchPersonCredits(ctx, personID, page)
	if err != nil {
		dc.logger.Error("获取豆瓣人物参演作品失败", "personID", personID, "page", page, "error", err)
		return nil, err
	}

	// 缓存结果（12小时）
	dc.cache.Set(cacheKey, mediaList, 12*time.Hour)

	dc.logger.Info("成功获取豆瓣人物参演作品", "personID", personID, "page", page, "count", len(mediaList))
	return mediaList, nil
}

// MovieTop250 获取豆瓣电影TOP250
func (dc *DoubanChain) MovieTop250(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	cacheKey := fmt.Sprintf("douban:movie_top250:%d:%d", page, count)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if mediaList, ok := cached.([]*model.MediaInfo); ok {
			return mediaList, nil
		}
	}

	dc.logger.Info("开始获取豆瓣电影TOP250", "page", page, "count", count)

	// 调用豆瓣API获取TOP250
	mediaList, err := dc.fetchMovieTop250(ctx, page, count)
	if err != nil {
		dc.logger.Error("获取豆瓣电影TOP250失败", "page", page, "count", count, "error", err)
		return nil, err
	}

	// 缓存结果（6小时）
	dc.cache.Set(cacheKey, mediaList, 6*time.Hour)

	dc.logger.Info("成功获取豆瓣电影TOP250", "page", page, "count", count, "count", len(mediaList))
	return mediaList, nil
}

// MovieShowing 获取正在上映的电影
func (dc *DoubanChain) MovieShowing(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	cacheKey := fmt.Sprintf("douban:movie_showing:%d:%d", page, count)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if mediaList, ok := cached.([]*model.MediaInfo); ok {
			return mediaList, nil
		}
	}

	dc.logger.Info("开始获取正在上映的电影", "page", page, "count", count)

	// 调用豆瓣API获取正在上映的电影
	mediaList, err := dc.fetchMovieShowing(ctx, page, count)
	if err != nil {
		dc.logger.Error("获取正在上映的电影失败", "page", page, "count", count, "error", err)
		return nil, err
	}

	// 缓存结果（1小时）
	dc.cache.Set(cacheKey, mediaList, time.Hour)

	dc.logger.Info("成功获取正在上映的电影", "page", page, "count", count, "count", len(mediaList))
	return mediaList, nil
}

// TVWeeklyChinese 获取本周中国剧集榜
func (dc *DoubanChain) TVWeeklyChinese(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	cacheKey := fmt.Sprintf("douban:tv_weekly_chinese:%d:%d", page, count)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if mediaList, ok := cached.([]*model.MediaInfo); ok {
			return mediaList, nil
		}
	}

	dc.logger.Info("开始获取本周中国剧集榜", "page", page, "count", count)

	// 调用豆瓣API获取本周中国剧集榜
	mediaList, err := dc.fetchTVWeeklyChinese(ctx, page, count)
	if err != nil {
		dc.logger.Error("获取本周中国剧集榜失败", "page", page, "count", count, "error", err)
		return nil, err
	}

	// 缓存结果（6小时）
	dc.cache.Set(cacheKey, mediaList, 6*time.Hour)

	dc.logger.Info("成功获取本周中国剧集榜", "page", page, "count", count, "count", len(mediaList))
	return mediaList, nil
}

// TVWeeklyGlobal 获取本周全球剧集榜
func (dc *DoubanChain) TVWeeklyGlobal(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	cacheKey := fmt.Sprintf("douban:tv_weekly_global:%d:%d", page, count)

	// 尝试从缓存获取
	if cached, found := dc.cache.Get(cacheKey); found {
		if mediaList, ok := cached.([]*model.MediaInfo); ok {
			return mediaList, nil
		}
	}

	dc.logger.Info("开始获取本周全球剧集榜", "page", page, "count", count)

	// 调用豆瓣API获取本周全球剧集榜
	mediaList, err := dc.fetchTVWeeklyGlobal(ctx, page, count)
	if err != nil {
		dc.logger.Error("获取本周全球剧集榜失败", "page", page, "count", count, "error", err)
		return nil, err
	}

	// 缓存结果（6小时）
	dc.cache.Set(cacheKey, mediaList, 6*time.Hour)

	dc.logger.Info("成功获取本周全球剧集榜", "page", page, "count", count, "count", len(mediaList))
	return mediaList, nil
}

// Private methods for actual API calls

func (dc *DoubanChain) fetchPersonDetail(ctx context.Context, personID int) (*model.MediaPerson, error) {
	// TODO: 实现豆瓣API调用获取人物详情
	// 这里需要实现实际的豆瓣API调用逻辑
	return &model.MediaPerson{}, nil
}

func (dc *DoubanChain) fetchPersonCredits(ctx context.Context, personID int, page int) ([]*model.MediaInfo, error) {
	// TODO: 实现豆瓣API调用获取人物参演作品
	return []*model.MediaInfo{}, nil
}

func (dc *DoubanChain) fetchMovieTop250(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	// TODO: 实现豆瓣API调用获取TOP250
	return []*model.MediaInfo{}, nil
}

func (dc *DoubanChain) fetchMovieShowing(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	// TODO: 实现豆瓣API调用获取正在上映的电影
	return []*model.MediaInfo{}, nil
}

func (dc *DoubanChain) fetchTVWeeklyChinese(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	// TODO: 实现豆瓣API调用获取本周中国剧集榜
	return []*model.MediaInfo{}, nil
}

func (dc *DoubanChain) fetchTVWeeklyGlobal(ctx context.Context, page int, count int) ([]*model.MediaInfo, error) {
	// TODO: 实现豆瓣API调用获取本周全球剧集榜
	return []*model.MediaInfo{}, nil
}
