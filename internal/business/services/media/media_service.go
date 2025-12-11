package media

import (
	"context"
	"strconv"
	"sync"

	"moviepilot-go/internal/business/services/base"
	mediadto "moviepilot-go/internal/models/dto/media"
)

// MediaService 媒体服务
// 原MediaChain，负责媒体信息识别、匹配、刮削等功能
type MediaService struct {
	*base.ServiceBase
}

var (
	mediaServiceInstance *MediaService
	mediaServiceOnce     sync.Once
)

// GetMediaService 获取MediaService单例
func GetMediaService() *MediaService {
	mediaServiceOnce.Do(func() {
		mediaServiceInstance = &MediaService{
			ServiceBase: base.NewServiceBase(),
		}
	})
	return mediaServiceInstance
}

// NewMediaService 创建MediaService实例（用于测试）
func NewMediaService() *MediaService {
	return &MediaService{
		ServiceBase: base.NewServiceBase(),
	}
}

// Initialize 初始化服务
func (s *MediaService) Initialize() error {
	// TODO: 初始化逻辑
	return nil
}

// Name 获取服务名称
func (s *MediaService) Name() string {
	return "MediaService"
}

// Close 关闭服务
func (s *MediaService) Close() error {
	// TODO: 清理资源
	return nil
}

// RecognizeMedia 识别媒体信息
// 从文件名或路径中识别媒体信息
func (s *MediaService) RecognizeMedia(ctx context.Context, title string) (*mediadto.MediaInfo, error) {
	// TODO: 实现媒体识别逻辑
	// 1. 调用模块进行识别
	// 2. 解析标题、年份、季集等信息
	// 3. 返回MediaInfo
	return &mediadto.MediaInfo{
		Source:     "themoviedb",
		Type:       "movie",
		Title:      title,
		Year:       "0",
		PosterPath: "",
	}, nil
}

// MatchMedia 匹配媒体信息
// 根据识别结果匹配TMDB/豆瓣等数据源
func (s *MediaService) MatchMedia(ctx context.Context, metaInfo *mediadto.MetaInfo) (*mediadto.MediaInfo, error) {
	// TODO: 实现媒体匹配逻辑
	// 1. 根据标题和年份搜索
	// 2. 匹配最佳结果
	// 3. 返回MediaInfo
	return &mediadto.MediaInfo{
		Source:     "themoviedb",
		Type:       "movie",
		Title:      metaInfo.Name,
		Year:       metaInfo.Year,
		PosterPath: "",
	}, nil
}

// GetMediaInfo 获取媒体详细信息
// 根据TMDB ID等获取完整的媒体信息
func (s *MediaService) GetMediaInfo(ctx context.Context, tmdbID int, mediaType string) (*mediadto.MediaInfo, error) {
	// TODO: 实现获取媒体信息逻辑
	// 1. 调用TMDB API
	// 2. 获取详细信息
	// 3. 返回MediaInfo
	return &mediadto.MediaInfo{
		Source:     "themoviedb",
		Type:       mediaType,
		Title:      "",
		Year:       "0",
		TmdbID:     &tmdbID,
		PosterPath: "",
	}, nil
}

// ScrapeMetadata 刮削媒体元数据
// 下载海报、背景图等
func (s *MediaService) ScrapeMetadata(ctx context.Context, mediaInfo *mediadto.MediaInfo, targetPath string) error {
	// TODO: 实现刮削逻辑
	// 1. 下载海报
	// 2. 下载背景图
	// 3. 生成NFO文件
	return nil
}

// GetSeasonEpisodes 获取季集信息
func (s *MediaService) GetSeasonEpisodes(ctx context.Context, tmdbID int, season int) ([]*mediadto.MediaSeason, error) {
	// TODO: 实现获取季集信息
	return []*mediadto.MediaSeason{
		{
			SeasonNumber: season,
			Name:         "第 " + strconv.Itoa(season) + " 季",
			EpisodeCount: 0,
		},
	}, nil
}

// SearchMedia 搜索媒体
func (s *MediaService) SearchMedia(ctx context.Context, keyword string, mediaType string) ([]*mediadto.MediaInfo, error) {
	// TODO: 实现媒体搜索
	return []*mediadto.MediaInfo{
		{
			Source:     "themoviedb",
			Type:       mediaType,
			Title:      keyword,
			Year:       "0",
			PosterPath: "",
		},
	}, nil
}
