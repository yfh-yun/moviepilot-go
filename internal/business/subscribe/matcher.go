package subscribe

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/rss"
)

// Matcher 订阅匹配器
type Matcher struct {
	logger *zap.Logger
}

// NewMatcher 创建匹配器
func NewMatcher(logger *zap.Logger) *Matcher {
	return &Matcher{
		logger: logger,
	}
}

// MatchRule 匹配规则
type MatchRule struct {
	Subscribe *models.Subscribe

	// 质量优先级 (从高到低)
	QualityPriority []string

	// 来源优先级 (从高到低)
	SourcePriority []string

	// 包含关键词
	IncludeKeywords []string

	// 排除关键词
	ExcludeKeywords []string

	// 文件大小限制
	MinSize int64
	MaxSize int64
}

// MatchScore 匹配评分
type MatchScore struct {
	Total        int
	QualityScore int
	SourceScore  int
	SizeScore    int
	SeedScore    int
	KeywordScore int
}

// MatchResult 匹配结果
type MatchResult struct {
	Torrent   *rss.TorrentInfo
	Subscribe *models.Subscribe
	Score     MatchScore
	Matched   bool
	Reason    string
}

// Match 匹配单个 Torrent
func (m *Matcher) Match(torrent *rss.TorrentInfo, rule MatchRule) (bool, MatchScore) {
	score := MatchScore{}

	// 1. 检查排除关键词
	for _, keyword := range rule.ExcludeKeywords {
		if !torrent.ExcludesKeyword(keyword) {
			if m.logger != nil {
				m.logger.Debug("torrent excluded by keyword",
					zap.String("torrent", torrent.Title),
					zap.String("keyword", keyword))
			}
			return false, score
		}
	}

	// 2. 检查包含关键词
	if len(rule.IncludeKeywords) > 0 {
		matched := false
		for _, keyword := range rule.IncludeKeywords {
			if torrent.ContainsKeyword(keyword) {
				matched = true
				score.KeywordScore += 10
				break
			}
		}
		if !matched {
			return false, score
		}
	}

	// 3. 检查季集匹配 (电视剧)
	if rule.Subscribe.Type == "tv" {
		if rule.Subscribe.Season != nil && torrent.Season != *rule.Subscribe.Season {
			return false, score
		}

		// 检查集数范围
		if rule.Subscribe.StartEpisode != nil && torrent.Episode < *rule.Subscribe.StartEpisode {
			return false, score
		}
	}

	// 4. 检查年份匹配 (电影)
	if rule.Subscribe.Type == "movie" && rule.Subscribe.Year != nil {
		if torrent.Year > 0 && torrent.Year != parseYear(*rule.Subscribe.Year) {
			return false, score
		}
	}

	// 5. 计算质量评分
	score.QualityScore = m.calculateQualityScore(torrent, rule.QualityPriority)

	// 6. 计算来源评分
	score.SourceScore = m.calculateSourceScore(torrent, rule.SourcePriority)

	// 7. 计算文件大小评分
	score.SizeScore = m.calculateSizeScore(torrent, rule.MinSize, rule.MaxSize)
	if score.SizeScore < 0 {
		return false, score
	}

	// 8. 计算做种数评分
	score.SeedScore = m.calculateSeedScore(torrent)

	// 9. 计算总分
	score.Total = score.QualityScore + score.SourceScore + score.SizeScore +
		score.SeedScore + score.KeywordScore

	if m.logger != nil {
		m.logger.Debug("torrent matched",
			zap.String("torrent", torrent.Title),
			zap.Int("score", score.Total))
	}

	return true, score
}

// SelectBest 从多个 Torrent 中选择最佳
func (m *Matcher) SelectBest(torrents []*rss.TorrentInfo, rule MatchRule) *rss.TorrentInfo {
	if len(torrents) == 0 {
		return nil
	}

	var bestTorrent *rss.TorrentInfo
	var bestScore MatchScore

	for _, torrent := range torrents {
		matched, score := m.Match(torrent, rule)
		if !matched {
			continue
		}

		if bestTorrent == nil || score.Total > bestScore.Total {
			bestTorrent = torrent
			bestScore = score
		}
	}

	if bestTorrent != nil && m.logger != nil {
		m.logger.Info("best torrent selected",
			zap.String("torrent", bestTorrent.Title),
			zap.Int("score", bestScore.Total))
	}

	return bestTorrent
}

// calculateQualityScore 计算质量评分
func (m *Matcher) calculateQualityScore(torrent *rss.TorrentInfo, priority []string) int {
	if len(priority) == 0 {
		return 50 // 默认分数
	}

	for i, quality := range priority {
		if strings.EqualFold(torrent.Quality, quality) {
			// 优先级越高，分数越高
			return 100 - (i * 10)
		}
	}

	return 0
}

// calculateSourceScore 计算来源评分
func (m *Matcher) calculateSourceScore(torrent *rss.TorrentInfo, priority []string) int {
	if len(priority) == 0 {
		return 50 // 默认分数
	}

	for i, source := range priority {
		if strings.Contains(strings.ToLower(torrent.Source), strings.ToLower(source)) {
			return 100 - (i * 10)
		}
	}

	return 0
}

// calculateSizeScore 计算文件大小评分
func (m *Matcher) calculateSizeScore(torrent *rss.TorrentInfo, minSize, maxSize int64) int {
	if torrent.Size == 0 {
		return 50 // 未知大小，给默认分
	}

	// 检查大小限制
	if minSize > 0 && torrent.Size < minSize {
		return -1 // 不匹配
	}
	if maxSize > 0 && torrent.Size > maxSize {
		return -1 // 不匹配
	}

	// 大小在合理范围内
	return 50
}

// calculateSeedScore 计算做种数评分
func (m *Matcher) calculateSeedScore(torrent *rss.TorrentInfo) int {
	if torrent.Seeders == 0 {
		return 0
	}

	// 做种数越多，分数越高，但有上限
	score := torrent.Seeders * 2
	if score > 100 {
		score = 100
	}

	return score
}

// parseYear 解析年份字符串
func parseYear(yearStr string) int {
	var year int
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err == nil {
		return year
	}
	return 0
}

// BuildMatchRule 从订阅构建匹配规则
func BuildMatchRule(subscribe *models.Subscribe) MatchRule {
	rule := MatchRule{
		Subscribe: subscribe,
	}

	// 解析质量优先级
	if subscribe.Quality != "" {
		rule.QualityPriority = strings.Split(subscribe.Quality, ",")
		for i := range rule.QualityPriority {
			rule.QualityPriority[i] = strings.TrimSpace(rule.QualityPriority[i])
		}
	} else {
		// 默认质量优先级
		rule.QualityPriority = []string{"2160p", "1080p", "720p"}
	}

	// 解析来源优先级
	if subscribe.Resolution != "" {
		rule.SourcePriority = strings.Split(subscribe.Resolution, ",")
		for i := range rule.SourcePriority {
			rule.SourcePriority[i] = strings.TrimSpace(rule.SourcePriority[i])
		}
	} else {
		// 默认来源优先级
		rule.SourcePriority = []string{"BluRay", "WEB-DL", "WEBRip"}
	}

	// 解析包含关键词
	if subscribe.Include != "" {
		rule.IncludeKeywords = strings.Split(subscribe.Include, ",")
		for i := range rule.IncludeKeywords {
			rule.IncludeKeywords[i] = strings.TrimSpace(rule.IncludeKeywords[i])
		}
	}

	// 解析排除关键词
	if subscribe.Exclude != "" {
		rule.ExcludeKeywords = strings.Split(subscribe.Exclude, ",")
		for i := range rule.ExcludeKeywords {
			rule.ExcludeKeywords[i] = strings.TrimSpace(rule.ExcludeKeywords[i])
		}
	}

	return rule
}
