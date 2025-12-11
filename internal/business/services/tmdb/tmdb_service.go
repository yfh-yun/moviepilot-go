package tmdb

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/pkg/logger"
)

// TmdbService TMDB服务
// 原TmdbChain，负责TheMovieDB API调用
type TmdbService struct {
	*base.ServiceBase
	logger  *zap.Logger
	apiKey  string
	baseURL string
}

// NewTmdbService 创建TmdbService实例
func NewTmdbService(apiKey string) *TmdbService {
	return &TmdbService{
		ServiceBase: base.NewServiceBase(),
		logger:      logger.GetLogger(),
		apiKey:      apiKey,
		baseURL:     "https://api.themoviedb.org/3",
	}
}

// Initialize 初始化服务
func (s *TmdbService) Initialize() error {
	return nil
}

// Name 获取服务名称
func (s *TmdbService) Name() string {
	return "TmdbService"
}

// Close 关闭服务
func (s *TmdbService) Close() error {
	return nil
}

// Discover 发现媒体
func (s *TmdbService) Discover(ctx context.Context, mediaType types.MediaType, params map[string]any) ([]*dto.MediaInfo, error) {
	s.logger.Info("发现媒体", zap.String("media_type", string(mediaType)))

	if s.apiKey == "" {
		return []*dto.MediaInfo{}, fmt.Errorf("TMDB API Key未配置")
	}

	// 构建请求路径
	path := fmt.Sprintf("/discover/%s?api_key=%s", mediaType, s.apiKey)
	if page, ok := params["page"].(int); ok {
		path += fmt.Sprintf("&page=%d", page)
	}

	// TODO: 发送HTTP请求并解析响应
	// 当前返回空列表，待实现HTTP客户端调用
	return []*dto.MediaInfo{}, nil
}

// Trending 获取热门
func (s *TmdbService) Trending(ctx context.Context, page int) ([]*dto.MediaInfo, error) {
	s.logger.Info("获取热门", zap.Int("page", page))
	// TODO: 调用 /trending/all/week API
	return []*dto.MediaInfo{}, nil
}

// GetMovieDetail 获取电影详情
func (s *TmdbService) GetMovieDetail(ctx context.Context, tmdbID int) (*dto.MediaInfo, error) {
	s.logger.Info("获取电影详情", zap.Int("tmdb_id", tmdbID))
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的TMDB ID")
	}
	// TODO: 调用 /movie/{movie_id} API
	return &dto.MediaInfo{}, nil
}

// GetTVDetail 获取电视剧详情
func (s *TmdbService) GetTVDetail(ctx context.Context, tmdbID int) (*dto.MediaInfo, error) {
	s.logger.Info("获取电视剧详情", zap.Int("tmdb_id", tmdbID))
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的TMDB ID")
	}
	// TODO: 调用 /tv/{tv_id} API
	return &dto.MediaInfo{}, nil
}

// GetSeasonDetail 获取季详情
func (s *TmdbService) GetSeasonDetail(ctx context.Context, tmdbID, season int) (*dto.TmdbSeason, error) {
	// TODO: 实现获取季详情逻辑
	return nil, nil
}

// GetEpisodeDetail 获取集详情
func (s *TmdbService) GetEpisodeDetail(ctx context.Context, tmdbID, season, episode int) (*dto.TmdbEpisode, error) {
	// TODO: 实现获取集详情逻辑
	return nil, nil
}

// Search 搜索
func (s *TmdbService) Search(ctx context.Context, keyword string, mediaType types.MediaType) ([]*dto.MediaInfo, error) {
	s.logger.Info("搜索媒体",
		zap.String("keyword", keyword),
		zap.String("media_type", string(mediaType)))
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	if s.apiKey == "" {
		return []*dto.MediaInfo{}, fmt.Errorf("TMDB API Key未配置")
	}

	// 构建搜索路径
	searchType := "multi"
	if mediaType == types.MediaTypeMovie {
		searchType = "movie"
	} else if mediaType == types.MediaTypeTV {
		searchType = "tv"
	}

	_ = fmt.Sprintf("/search/%s?api_key=%s&query=%s&language=zh-CN", searchType, s.apiKey, keyword)

	// TODO: 发送HTTP请求并解析响应
	return []*dto.MediaInfo{}, nil
}

// GetCredits 获取演职员表
func (s *TmdbService) GetCredits(ctx context.Context, tmdbID int, mediaType types.MediaType) ([]*dto.MediaPerson, error) {
	s.logger.Info("获取演职员表",
		zap.Int("tmdb_id", tmdbID),
		zap.String("media_type", string(mediaType)))
	// TODO: 调用 /movie/{id}/credits 或 /tv/{id}/credits API
	return []*dto.MediaPerson{}, nil
}

// GetRecommendations 获取推荐
func (s *TmdbService) GetRecommendations(ctx context.Context, tmdbID int, mediaType types.MediaType) ([]*dto.MediaInfo, error) {
	s.logger.Info("获取推荐",
		zap.Int("tmdb_id", tmdbID),
		zap.String("media_type", string(mediaType)))
	// TODO: 调用 /movie/{id}/recommendations 或 /tv/{id}/recommendations API
	return []*dto.MediaInfo{}, nil
}

// GetSeasons 获取所有季信息
func (s *TmdbService) GetSeasons(ctx context.Context, tmdbID int) ([]*dto.TmdbSeason, error) {
	s.logger.Info("获取所有季信息", zap.Int("tmdb_id", tmdbID))
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的TMDB ID")
	}
	// TODO: 调用 /tv/{tv_id} API 并提取 seasons 字段
	return []*dto.TmdbSeason{}, nil
}

// GetSimilar 获取相似媒体
func (s *TmdbService) GetSimilar(ctx context.Context, tmdbID int, mediaType types.MediaType) ([]*dto.MediaInfo, error) {
	s.logger.Info("获取相似媒体",
		zap.Int("tmdb_id", tmdbID),
		zap.String("media_type", string(mediaType)))
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的TMDB ID")
	}
	// TODO: 调用 /movie/{id}/similar 或 /tv/{id}/similar API
	return []*dto.MediaInfo{}, nil
}

// GetCollection 获取系列合集
func (s *TmdbService) GetCollection(ctx context.Context, collectionID int) ([]*dto.MediaInfo, error) {
	s.logger.Info("获取系列合集", zap.Int("collection_id", collectionID))
	if collectionID <= 0 {
		return nil, fmt.Errorf("无效的合集ID")
	}
	// TODO: 调用 /collection/{collection_id} API
	return []*dto.MediaInfo{}, nil
}

// GetPersonDetail 获取人物详情
func (s *TmdbService) GetPersonDetail(ctx context.Context, personID int) (*dto.MediaPerson, error) {
	s.logger.Info("获取人物详情", zap.Int("person_id", personID))
	if personID <= 0 {
		return nil, fmt.Errorf("无效的人物ID")
	}
	// TODO: 调用 /person/{person_id} API
	return &dto.MediaPerson{}, nil
}

// GetPersonCredits 获取人物参演作品
func (s *TmdbService) GetPersonCredits(ctx context.Context, personID int, page int) ([]*dto.MediaInfo, error) {
	s.logger.Info("获取人物参演作品",
		zap.Int("person_id", personID),
		zap.Int("page", page))
	if personID <= 0 {
		return nil, fmt.Errorf("无效的人物ID")
	}
	// TODO: 调用 /person/{person_id}/combined_credits API
	return []*dto.MediaInfo{}, nil
}

// GetEpisodes 获取季的所有集
func (s *TmdbService) GetEpisodes(ctx context.Context, tmdbID int, season int, episodeGroup string) ([]*dto.TmdbEpisode, error) {
	s.logger.Info("获取季的所有集",
		zap.Int("tmdb_id", tmdbID),
		zap.Int("season", season),
		zap.String("episode_group", episodeGroup))
	if tmdbID <= 0 || season < 0 {
		return nil, fmt.Errorf("无效的TMDB ID或季号")
	}
	// TODO: 调用 /tv/{tv_id}/season/{season_number} API
	return []*dto.TmdbEpisode{}, nil
}
