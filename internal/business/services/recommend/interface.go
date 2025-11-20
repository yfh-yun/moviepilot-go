// Package recommend 推荐服务
package recommend

import (
	"context"
)

// Service 推荐服务接口
type Service interface {
	// GetSource 获取推荐源
	GetSource(ctx context.Context) ([]string, error)
	
	// BangumiCalendar 获取番组日历
	BangumiCalendar(ctx context.Context) (interface{}, error)
	
	// DoubanShowing 获取豆瓣正在上映的电影
	DoubanShowing(ctx context.Context) (interface{}, error)
	
	// DoubanMovies 获取豆瓣电影列表
	DoubanMovies(ctx context.Context, page, limit int) (interface{}, error)
	
	// DoubanTVs 获取豆瓣电视剧列表
	DoubanTVs(ctx context.Context, page, limit int) (interface{}, error)
	
	// DoubanMovieTop250 获取豆瓣电影Top250
	DoubanMovieTop250(ctx context.Context) (interface{}, error)
	
	// DoubanTVChinese 获取豆瓣华语电视剧
	DoubanTVChinese(ctx context.Context, page, limit int) (interface{}, error)
	
	// DoubanTVGlobal 获取豆瓣海外电视剧
	DoubanTVGlobal(ctx context.Context, page, limit int) (interface{}, error)
	
	// DoubanTVAnimation 获取豆瓣动画剧集
	DoubanTVAnimation(ctx context.Context, page, limit int) (interface{}, error)
	
	// DoubanMovieHot 获取豆瓣热门电影
	DoubanMovieHot(ctx context.Context) (interface{}, error)
	
	// DoubanTVHot 获取豆瓣热门电视剧
	DoubanTVHot(ctx context.Context) (interface{}, error)
	
	// TMDbMovies 获取TMDb电影列表
	TMDbMovies(ctx context.Context, page, limit int) (interface{}, error)
	
	// TMDbTVs 获取TMDb电视剧列表
	TMDbTVs(ctx context.Context, page, limit int) (interface{}, error)
	
	// TMDbTrending 获取TMDb热门趋势
	TMDbTrending(ctx context.Context) (interface{}, error)
	
	// GetSimilarMedia 获取相似媒体
	GetSimilarMedia(ctx context.Context, id string) (interface{}, error)
	
	// GetPreferenceBasedRecommendations 获取基于偏好的推荐
	GetPreferenceBasedRecommendations(ctx context.Context) (interface{}, error)
}