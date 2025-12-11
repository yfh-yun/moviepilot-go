package torrent

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"
)

// TorrentCacheItem 种子缓存项
type TorrentCacheItem struct {
	Hash          string    `json:"hash"`
	Domain        string    `json:"domain"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Size          int64     `json:"size"`
	Pubdate       time.Time `json:"pubdate"`
	SiteName      string    `json:"site_name"`
	MediaName     string    `json:"media_name"`
	MediaYear     string    `json:"media_year"`
	MediaType     string    `json:"media_type"`
	SeasonEpisode string    `json:"season_episode"`
	ResourceTerm  string    `json:"resource_term"`
	Enclosure     string    `json:"enclosure"`
	PageURL       string    `json:"page_url"`
	PosterPath    string    `json:"poster_path"`
	BackdropPath  string    `json:"backdrop_path"`
}

// TorrentContext 种子上下文
type TorrentContext struct {
	TorrentInfo *TorrentInfo    `json:"torrent_info"`
	MetaInfo    *MetaInfo       `json:"meta_info"`
	MediaInfo   *database.Media `json:"media_info"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Title       string    `json:"title"`
	Site        string    `json:"site"`
	Description string    `json:"description"`
	Enclosure   string    `json:"enclosure"`
	PageURL     string    `json:"page_url"`
	Size        int64     `json:"size"`
	PubDate     time.Time `json:"pub_date"`
	Seeders     int       `json:"seeders"`
	Leechers    int       `json:"leechers"`
}

// MetaInfo 元数据信息
type MetaInfo struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Year       int    `json:"year"`
	Season     int    `json:"season"`
	Episodes   []int  `json:"episodes"`
	Type       string `json:"type"`
	Resolution string `json:"resolution"`
	Quality    string `json:"quality"`
	Effect     string `json:"effect"`
	VideoCodec string `json:"video_codec"`
	AudioCodec string `json:"audio_codec"`
}

// Service 种子服务接口
type Service interface {
	// GetTorrentsCache 获取种子缓存
	GetTorrentsCache(ctx context.Context, mode string) (map[string][]TorrentContext, error)
	// DeleteTorrentCache 删除指定种子缓存
	DeleteTorrentCache(ctx context.Context, domain string, torrentHash string) error
	// ClearTorrentsCache 清理所有种子缓存
	ClearTorrentsCache(ctx context.Context) error
	// RefreshTorrentsCache 刷新种子缓存
	RefreshTorrentsCache(ctx context.Context) (map[string][]TorrentContext, error)
	// ReidentifyTorrent 重新识别种子
	ReidentifyTorrent(ctx context.Context, domain string, torrentHash string, tmdbID *int, doubanID *string) (*TorrentCacheItem, error)
	// GetFormattedCache 获取格式化的种子缓存
	GetFormattedCache(ctx context.Context, mode string) (*GetTorrentCacheResponse, error)
}

// torrentService 种子服务实现
type torrentService struct {
	*base.ServiceBase
	cache  cache.Cache
	logger *zap.Logger
}

// NewTorrentService 创建种子服务实例
func NewTorrentService(cache cache.Cache) Service {
	return &torrentService{
		ServiceBase: base.NewServiceBase(),
		cache:       cache,
		logger:      logger.GetLogger(),
	}
}

// Name 获取服务名称
func (s *torrentService) Name() string {
	return "TorrentService"
}

// GetTorrentsCache 获取种子缓存
func (s *torrentService) GetTorrentsCache(ctx context.Context, mode string) (map[string][]TorrentContext, error) {
	s.logger.Info("获取种子缓存", zap.String("mode", mode))

	// 构建缓存键
	cacheKey := fmt.Sprintf("torrents:%s", mode)

	// 从缓存中获取种子数据
	var torrentsMap map[string][]TorrentContext
	err := s.cache.GetJSON(ctx, cacheKey, &torrentsMap)
	if err != nil {
		s.logger.Warn("从缓存获取种子数据失败", zap.Error(err))
		// 返回空映射而非错误，避免缓存未命中时失败
		return make(map[string][]TorrentContext), nil
	}

	return torrentsMap, nil
}

// DeleteTorrentCache 删除指定种子缓存
func (s *torrentService) DeleteTorrentCache(ctx context.Context, domain string, torrentHash string) error {
	s.logger.Info("删除指定种子缓存", zap.String("domain", domain), zap.String("hash", torrentHash))

	// 获取所有模式的缓存
	modes := []string{"spider", "rss"}
	for _, mode := range modes {
		// 从缓存中获取种子数据
		cacheKey := fmt.Sprintf("torrents:%s", mode)
		var torrentsMap map[string][]TorrentContext
		err := s.cache.GetJSON(ctx, cacheKey, &torrentsMap)
		if err != nil {
			continue
		}

		// 检查站点是否存在
		if _, ok := torrentsMap[domain]; !ok {
			continue
		}

		// 查找并删除指定种子
		originalCount := len(torrentsMap[domain])
		var updatedTorrents []TorrentContext
		for _, torrent := range torrentsMap[domain] {
			// 计算哈希值
			computedHash := s.calculateTorrentHash(torrent.TorrentInfo.Title, torrent.TorrentInfo.Description)
			if computedHash != torrentHash {
				updatedTorrents = append(updatedTorrents, torrent)
			}
		}

		// 如果有变化，保存更新后的缓存
		if len(updatedTorrents) != originalCount {
			torrentsMap[domain] = updatedTorrents
			if err := s.cache.SetJSON(ctx, cacheKey, torrentsMap, 0); err != nil {
				s.logger.Error("保存更新后的缓存失败", zap.Error(err), zap.String("mode", mode))
				return err
			}
			s.logger.Info("种子缓存删除成功", zap.String("domain", domain), zap.String("hash", torrentHash), zap.String("mode", mode))
			return nil // 只需要在一个模式中找到并删除即可
		}
	}

	return fmt.Errorf("未找到指定的种子")
}

// ClearTorrentsCache 清理所有种子缓存
func (s *torrentService) ClearTorrentsCache(ctx context.Context) error {
	s.logger.Info("清理所有种子缓存")

	// 获取所有缓存键并删除
	modes := []string{"spider", "rss"}
	for _, mode := range modes {
		cacheKey := fmt.Sprintf("torrents:%s", mode)
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			s.logger.Error("删除种子缓存失败", zap.Error(err), zap.String("mode", mode))
			return err
		}
	}

	return nil
}

// RefreshTorrentsCache 刷新种子缓存
func (s *torrentService) RefreshTorrentsCache(ctx context.Context) (map[string][]TorrentContext, error) {
	s.logger.Info("刷新种子缓存")

	// TODO: 实现实际的种子缓存刷新逻辑
	// 1. 调用种子爬虫或RSS解析器获取最新种子
	// 2. 更新缓存
	// 3. 返回刷新后的种子数据

	// 目前返回空映射
	return make(map[string][]TorrentContext), nil
}

// ReidentifyTorrent 重新识别种子
func (s *torrentService) ReidentifyTorrent(ctx context.Context, domain string, torrentHash string, tmdbID *int, doubanID *string) (*TorrentCacheItem, error) {
	s.logger.Info("重新识别种子", zap.String("domain", domain), zap.String("hash", torrentHash), zap.Intp("tmdb_id", tmdbID), zap.Stringp("douban_id", doubanID))

	// 获取所有模式的缓存
	modes := []string{"spider", "rss"}
	for _, mode := range modes {
		// 从缓存中获取种子数据
		cacheKey := fmt.Sprintf("torrents:%s", mode)
		var torrentsMap map[string][]TorrentContext
		err := s.cache.GetJSON(ctx, cacheKey, &torrentsMap)
		if err != nil {
			continue
		}

		// 检查站点是否存在
		if _, ok := torrentsMap[domain]; !ok {
			continue
		}

		// 查找指定种子
		for i, torrent := range torrentsMap[domain] {
			// 计算哈希值
			computedHash := s.calculateTorrentHash(torrent.TorrentInfo.Title, torrent.TorrentInfo.Description)
			if computedHash == torrentHash {
				// TODO: 实现实际的媒体重新识别逻辑
				// 1. 调用媒体识别服务重新识别
				// 2. 更新种子上下文的媒体信息
				// 3. 保存更新后的缓存

				// 更新种子上下文
				// 这里只是示例，实际需要调用媒体识别服务
				yearStr := fmt.Sprintf("%d", torrent.MetaInfo.Year)
				torrentsMap[domain][i].MediaInfo = &database.Media{
					Title: torrent.MetaInfo.Title,
					Year:  &yearStr,
					Type:  torrent.MetaInfo.Type,
				}

				// 保存更新后的缓存
				if err := s.cache.SetJSON(ctx, cacheKey, torrentsMap, 0); err != nil {
					s.logger.Error("保存更新后的缓存失败", zap.Error(err), zap.String("mode", mode))
					return nil, err
				}

				// 转换为前端需要的格式
				cacheItem := s.convertToCacheItem(domain, &torrentsMap[domain][i])
				return cacheItem, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到指定的种子")
}

// calculateTorrentHash 计算种子哈希值
func (s *torrentService) calculateTorrentHash(title, description string) string {
	data := fmt.Sprintf("%s%s", title, description)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// convertToCacheItem 转换为前端需要的缓存项格式
func (s *torrentService) convertToCacheItem(domain string, torrent *TorrentContext) *TorrentCacheItem {
	// 构建season_episode和resource_term
	seasonEpisode := ""
	if torrent.MetaInfo != nil {
		if torrent.MetaInfo.Season > 0 {
			seasonEpisode = fmt.Sprintf("S%02d", torrent.MetaInfo.Season)
			if len(torrent.MetaInfo.Episodes) > 0 {
				episodeStrs := make([]string, len(torrent.MetaInfo.Episodes))
				for i, ep := range torrent.MetaInfo.Episodes {
					episodeStrs[i] = fmt.Sprintf("E%02d", ep)
				}
				seasonEpisode += fmt.Sprintf("%s", strings.Join(episodeStrs, "+"))
			}
		}
	}

	resourceTerm := ""
	if torrent.MetaInfo != nil {
		terms := []string{}
		if torrent.MetaInfo.Resolution != "" {
			terms = append(terms, torrent.MetaInfo.Resolution)
		}
		if torrent.MetaInfo.Quality != "" {
			terms = append(terms, torrent.MetaInfo.Quality)
		}
		if torrent.MetaInfo.Effect != "" {
			terms = append(terms, torrent.MetaInfo.Effect)
		}
		if torrent.MetaInfo.VideoCodec != "" {
			terms = append(terms, torrent.MetaInfo.VideoCodec)
		}
		if torrent.MetaInfo.AudioCodec != "" {
			terms = append(terms, torrent.MetaInfo.AudioCodec)
		}
		resourceTerm = strings.Join(terms, ".")
	}

	// 媒体名称和年份
	mediaName := ""
	mediaYear := ""
	mediaType := ""
	if torrent.MediaInfo != nil {
		mediaName = torrent.MediaInfo.Title
		mediaYear = fmt.Sprintf("%d", torrent.MediaInfo.Year)
		mediaType = torrent.MediaInfo.Type
	}

	return &TorrentCacheItem{
		Hash:          s.calculateTorrentHash(torrent.TorrentInfo.Title, torrent.TorrentInfo.Description),
		Domain:        domain,
		Title:         torrent.TorrentInfo.Title,
		Description:   torrent.TorrentInfo.Description,
		Size:          torrent.TorrentInfo.Size,
		Pubdate:       torrent.TorrentInfo.PubDate,
		SiteName:      torrent.TorrentInfo.Site,
		MediaName:     mediaName,
		MediaYear:     mediaYear,
		MediaType:     mediaType,
		SeasonEpisode: seasonEpisode,
		ResourceTerm:  resourceTerm,
		Enclosure:     torrent.TorrentInfo.Enclosure,
		PageURL:       torrent.TorrentInfo.PageURL,
		PosterPath:    "", // TODO: 获取媒体海报
		BackdropPath:  "", // TODO: 获取媒体背景图
	}
}

// GetTorrentCacheResponse 获取种子缓存响应
type GetTorrentCacheResponse struct {
	Count int                `json:"count"`
	Sites int                `json:"sites"`
	Data  []TorrentCacheItem `json:"data"`
}

// GetFormattedCache 获取格式化的种子缓存
func (s *torrentService) GetFormattedCache(ctx context.Context, mode string) (*GetTorrentCacheResponse, error) {
	s.logger.Info("获取格式化的种子缓存", zap.String("mode", mode))

	// 获取种子缓存
	torrentsMap, err := s.GetTorrentsCache(ctx, mode)
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	var torrentItems []TorrentCacheItem
	for domain, torrents := range torrentsMap {
		for _, torrent := range torrents {
			item := s.convertToCacheItem(domain, &torrent)
			torrentItems = append(torrentItems, *item)
		}
	}

	// 统计信息
	torrentCount := len(torrentItems)
	siteCount := len(torrentsMap)

	return &GetTorrentCacheResponse{
		Count: torrentCount,
		Sites: siteCount,
		Data:  torrentItems,
	}, nil
}
