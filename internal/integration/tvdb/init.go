package tvdb

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/infrastructure/config"
)

// TVDBService TVDB集成服务
type TVDBService struct {
	client *Client
	logger *logger.Logger
}

// NewTVDBService 创建TVDB服务
func NewTVDBService(cfg *config.Config) (*TVDBService, error) {
	if cfg.TVDB.APIKey == "" {
		return nil, fmt.Errorf("TVDB API密钥未配置")
	}

	client := NewClient(cfg)

	// 测试连接
	if err := client.testConnection(); err != nil {
		return nil, fmt.Errorf("TVDB连接测试失败: %w", err)
	}

	return &TVDBService{
		client: client,
		logger: logger.NewLogger("tvdb_service"),
	}, nil
}

// testConnection 测试连接
func (c *Client) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用一个已知的剧集进行测试（如The Expanse）
	_, err := c.SearchSeries(ctx, "The Expanse")
	if err != nil {
		// 如果搜索失败，可能是认证问题，尝试重新认证
		authManager := NewAuthManager(c)
		if authErr := authManager.ForceRefresh(); authErr != nil {
			return fmt.Errorf("连接测试失败: %w, 认证错误: %v", err, authErr)
		}

		// 重试搜索
		_, err = c.SearchSeries(ctx, "The Expanse")
		if err != nil {
			return fmt.Errorf("重试连接测试失败: %w", err)
		}
	}

	return nil
}

// GetSeriesWithEpisodes 获取剧集信息（包含剧集列表）
func (s *TVDBService) GetSeriesWithEpisodes(ctx context.Context, seriesID int) (*Series, error) {
	s.logger.Debugf("获取剧集信息: %d", seriesID)

	series, err := s.client.GetSeries(ctx, seriesID)
	if err != nil {
		s.logger.Errorf("获取剧集信息失败: %v", err)
		return nil, err
	}

	// 如果剧集不包含剧集列表，获取并添加
	if len(series.Episodes) == 0 {
		episodes, err := s.client.GetEpisodes(ctx, seriesID)
		if err != nil {
			s.logger.Warnf("获取剧集列表失败: %v", err)
		} else {
			series.Episodes = episodes
		}
	}

	s.logger.Debugf("成功获取剧集信息: %s (ID: %d)", series.Name, seriesID)
	return series, nil
}

// SearchSeriesWithDetails 搜索剧集并获取详细信息
func (s *TVDBService) SearchSeriesWithDetails(ctx context.Context, query string) ([]*Series, error) {
	s.logger.Debugf("搜索剧集: %s", query)

	results, err := s.client.SearchSeries(ctx, query)
	if err != nil {
		s.logger.Errorf("搜索剧集失败: %v", err)
		return nil, err
	}

	var detailedResults []*Series

	// 为每个搜索结果获取详细信息
	for _, result := range results {
		series, err := s.GetSeriesWithEpisodes(ctx, result.ID)
		if err != nil {
			s.logger.Warnf("获取剧集详细信息失败: %d - %v", result.ID, err)
			// 使用基本信息
			detailedResults = append(detailedResults, &result)
		} else {
			detailedResults = append(detailedResults, series)
		}
	}

	s.logger.Debugf("搜索完成，找到 %d 个结果", len(detailedResults))
	return detailedResults, nil
}

// GetEpisodesBySeason 按季节获取剧集
func (s *TVDBService) GetEpisodesBySeason(ctx context.Context, seriesID, season int) ([]Episode, error) {
	s.logger.Debugf("获取剧集季节信息: %d 第 %d 季", seriesID, season)

	episodes, err := s.client.GetEpisodes(ctx, seriesID)
	if err != nil {
		s.logger.Errorf("获取剧集列表失败: %v", err)
		return nil, err
	}

	var seasonEpisodes []Episode
	for _, episode := range episodes {
		if episode.Season == season {
			seasonEpisodes = append(seasonEpisodes, episode)
		}
	}

	s.logger.Debugf("成功获取第 %d 季剧集: %d 集", season, len(seasonEpisodes))
	return seasonEpisodes, nil
}

// GetImageURL 获取图片URL
func (s *TVDBService) GetImageURL(imagePath string) string {
	if imagePath == "" {
		return ""
	}

	// 根据TVDB API文档构建图片URL
	return fmt.Sprintf("https://artworks.thetvdb.com%s", imagePath)
}

// ClearCache 清空缓存
func (s *TVDBService) ClearCache() {
	s.client.ClearCache()
	s.logger.Info("TVDB缓存已清空")
}

// GetCacheStats 获取缓存统计
func (s *TVDBService) GetCacheStats() (seriesCount, episodesCount, searchCount int) {
	return s.client.GetCacheStats()
}

// HealthCheck 健康检查
func (s *TVDBService) HealthCheck(ctx context.Context) error {
	s.logger.Debug("执行TVDB健康检查")

	// 测试搜索功能
	_, err := s.SearchSeries(ctx, "test")
	if err != nil {
		s.logger.Errorf("TVDB健康检查失败: %v", err)
		return fmt.Errorf("TVDB健康检查失败: %w", err)
	}

	s.logger.Debug("TVDB健康检查通过")
	return nil
}

// IsAvailable 检查服务是否可用
func (s *TVDBService) IsAvailable() bool {
	authManager := NewAuthManager(s.client)
	return authManager.IsTokenValid()
}
