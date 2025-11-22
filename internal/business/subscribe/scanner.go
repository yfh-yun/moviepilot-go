package subscribe

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repository/interfaces"
	"moviepilot-go/pkg/rss"
)

// Scanner 订阅扫描器
type Scanner struct {
	parser  *rss.Parser
	matcher *Matcher
	repo    interfaces.SubscribeRepository
	logger  *zap.Logger
}

// NewScanner 创建扫描器
func NewScanner(parser *rss.Parser, matcher *Matcher, repo interfaces.SubscribeRepository, logger *zap.Logger) *Scanner {
	return &Scanner{
		parser:  parser,
		matcher: matcher,
		repo:    repo,
		logger:  logger,
	}
}

// RSSSource RSS 订阅源
type RSSSource struct {
	Name     string
	URL      string
	Enabled  bool
	Interval time.Duration
}

// ScanResult 扫描结果
type ScanResult struct {
	Subscribe *models.Subscribe
	Matches   []MatchResult
	Error     error
}

// ScanAll 扫描所有活跃订阅
func (s *Scanner) ScanAll(ctx context.Context, sources []RSSSource) ([]ScanResult, error) {
	// 获取所有活跃订阅
	subscribes, err := s.repo.GetActiveSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscribes: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("scanning subscribes",
			zap.Int("count", len(subscribes)),
			zap.Int("sources", len(sources)))
	}

	// 获取所有 RSS 源的内容
	allTorrents := make([]*rss.TorrentInfo, 0)
	for _, source := range sources {
		if !source.Enabled {
			continue
		}

		torrents, err := s.fetchSource(ctx, source)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to fetch RSS source",
					zap.String("source", source.Name),
					zap.Error(err))
			}
			continue
		}

		allTorrents = append(allTorrents, torrents...)
	}

	if s.logger != nil {
		s.logger.Info("fetched torrents", zap.Int("count", len(allTorrents)))
	}

	// 扫描每个订阅
	results := make([]ScanResult, 0, len(subscribes))
	for _, sub := range subscribes {
		result := s.ScanSubscribe(ctx, sub, allTorrents)
		results = append(results, result)
	}

	return results, nil
}

// ScanSubscribe 扫描单个订阅
func (s *Scanner) ScanSubscribe(ctx context.Context, subscribe *models.Subscribe, torrents []*rss.TorrentInfo) ScanResult {
	result := ScanResult{
		Subscribe: subscribe,
		Matches:   make([]MatchResult, 0),
	}

	// 构建匹配规则
	rule := BuildMatchRule(subscribe)

	// 匹配所有 Torrent
	for _, torrent := range torrents {
		// 检查上下文
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		default:
		}

		// 基本标题匹配
		if !s.matchTitle(torrent, subscribe) {
			continue
		}

		// 详细匹配
		matched, score := s.matcher.Match(torrent, rule)

		matchResult := MatchResult{
			Torrent:   torrent,
			Subscribe: subscribe,
			Score:     score,
			Matched:   matched,
		}

		if matched {
			matchResult.Reason = "matched"
			result.Matches = append(result.Matches, matchResult)
		}
	}

	if s.logger != nil && len(result.Matches) > 0 {
		s.logger.Info("subscribe matched",
			zap.String("name", subscribe.Name),
			zap.Int("matches", len(result.Matches)))
	}

	return result
}

// fetchSource 获取 RSS 源内容
func (s *Scanner) fetchSource(ctx context.Context, source RSSSource) ([]*rss.TorrentInfo, error) {
	// 解析 RSS
	feed, err := s.parser.ParseURL(ctx, source.URL)
	if err != nil {
		return nil, err
	}

	// 提取 Torrent 信息
	torrents := make([]*rss.TorrentInfo, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		torrent, err := rss.ExtractTorrentInfo(item)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to extract torrent info",
					zap.String("title", item.Title),
					zap.Error(err))
			}
			continue
		}

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// matchTitle 基本标题匹配
func (s *Scanner) matchTitle(torrent *rss.TorrentInfo, subscribe *models.Subscribe) bool {
	// 简单的标题包含匹配
	torrentTitle := strings.ToLower(torrent.MediaTitle)
	subscribeName := strings.ToLower(subscribe.Name)

	// 检查是否包含订阅名称
	if strings.Contains(torrentTitle, subscribeName) {
		return true
	}

	// 检查关键词
	if subscribe.Keyword != "" {
		keywords := strings.Split(subscribe.Keyword, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(strings.ToLower(keyword))
			if strings.Contains(torrentTitle, keyword) {
				return true
			}
		}
	}

	return false
}

// GetBestMatch 获取最佳匹配
func (s *Scanner) GetBestMatch(result ScanResult) *MatchResult {
	if len(result.Matches) == 0 {
		return nil
	}

	// 找到评分最高的
	var best *MatchResult
	for i := range result.Matches {
		if best == nil || result.Matches[i].Score.Total > best.Score.Total {
			best = &result.Matches[i]
		}
	}

	return best
}
