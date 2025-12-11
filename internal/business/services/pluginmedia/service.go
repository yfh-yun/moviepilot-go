package pluginmedia

import (
	"context"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string
	Category string
	Sites    []string
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Title       string
	URL         string
	Seeders     int
	Leechers    int
	Size        int64
	PublishDate string
	Site        string
	Category    string
}

// RecognizeRequest 识别请求
type RecognizeRequest struct {
	Path string
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title    string
	Year     int
	Type     string
	TMDBID   int
	Season   int
	Episode  int
	Category string
}

// Service 接口
type Service interface {
	SearchTorrents(ctx context.Context, req SearchRequest) ([]*TorrentInfo, error)
	RecognizeMedia(ctx context.Context, req RecognizeRequest) (*MediaInfo, error)
}

type service struct {
	log *zap.Logger
}

func NewService() Service {
	return &service{
		log: logger.GetLogger().With(zap.String("service", "pluginmedia")),
	}
}

func (s *service) SearchTorrents(ctx context.Context, req SearchRequest) ([]*TorrentInfo, error) {
	s.log.Info("SearchTorrents called", zap.String("keyword", req.Keyword))
	return []*TorrentInfo{}, nil
}

func (s *service) RecognizeMedia(ctx context.Context, req RecognizeRequest) (*MediaInfo, error) {
	s.log.Info("RecognizeMedia called", zap.String("path", req.Path))
	return &MediaInfo{}, nil
}
