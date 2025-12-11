package tvdb

import (
	"context"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/dto"
)

// TvdbService TVDB服务
// 原TvdbChain，负责TVDB API调用
type TvdbService struct {
	*base.ServiceBase
}

// NewTvdbService 创建TvdbService实例
func NewTvdbService() *TvdbService {
	return &TvdbService{
		ServiceBase: base.NewServiceBase(),
	}
}

// Initialize 初始化服务
func (s *TvdbService) Initialize() error {
	return nil
}

// Name 获取服务名称
func (s *TvdbService) Name() string {
	return "TvdbService"
}

// Close 关闭服务
func (s *TvdbService) Close() error {
	return nil
}

// GetSeriesDetail 获取剧集详情
func (s *TvdbService) GetSeriesDetail(ctx context.Context, tvdbID int) (*dto.MediaInfo, error) {
	// TODO: 实现获取剧集详情逻辑
	return nil, nil
}

// Search 搜索
func (s *TvdbService) Search(ctx context.Context, keyword string) ([]*dto.MediaInfo, error) {
	// TODO: 实现搜索逻辑
	return nil, nil
}
