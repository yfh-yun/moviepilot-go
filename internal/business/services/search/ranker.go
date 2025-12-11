package search

import (
	"sort"
	"strings"
	"time"
)

// Ranker 搜索结果排序器
type Ranker interface {
	// Rank 对搜索结果排序
	Rank(results []*SearchResult, query *SearchQuery) []*RankedResult

	// CalculateScore 计算单个结果的分数
	CalculateScore(result *SearchResult, query *SearchQuery) float64
}

// ranker 排序器实现
type ranker struct {
	weights ScoreWeights
}

// NewRanker 创建排序器
func NewRanker(weights ScoreWeights) Ranker {
	return &ranker{
		weights: weights,
	}
}

// ScoreWeights 评分权重
type ScoreWeights struct {
	TitleMatch      float64 // 标题匹配度权重
	Quality         float64 // 质量权重
	Seeders         float64 // 做种数权重
	Size            float64 // 大小权重
	PublishTime     float64 // 发布时间权重
	SiteReliability float64 // 站点可靠性权重
}

// DefaultScoreWeights 默认权重
func DefaultScoreWeights() ScoreWeights {
	return ScoreWeights{
		TitleMatch:      0.30,
		Quality:         0.25,
		Seeders:         0.20,
		Size:            0.10,
		PublishTime:     0.10,
		SiteReliability: 0.05,
	}
}

// RankedResult 排序后的结果
type RankedResult struct {
	*SearchResult
	Score       float64            `json:"score"`
	ScoreDetail map[string]float64 `json:"score_detail"`
	Rank        int                `json:"rank"`
}

// SearchResult 搜索结果（简化版）
type SearchResult struct {
	Title       string    `json:"title"`
	Size        int64     `json:"size"`
	Seeders     int       `json:"seeders"`
	Leechers    int       `json:"leechers"`
	PublishTime time.Time `json:"publish_time"`
	SiteName    string    `json:"site_name"`
	DownloadURL string    `json:"download_url"`
	DetailURL   string    `json:"detail_url"`
}

// SearchQuery 搜索查询（简化版）
type SearchQuery struct {
	Keyword    string
	Type       string // movie, tv
	Quality    string
	Resolution string
	MinSize    int64
	MaxSize    int64
}

// Rank 对搜索结果排序
func (r *ranker) Rank(results []*SearchResult, query *SearchQuery) []*RankedResult {
	ranked := make([]*RankedResult, 0, len(results))

	for _, result := range results {
		score := r.CalculateScore(result, query)

		rankedResult := &RankedResult{
			SearchResult: result,
			Score:        score,
			ScoreDetail:  r.calculateScoreDetail(result, query),
		}

		ranked = append(ranked, rankedResult)
	}

	// 按分数排序
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	// 设置排名
	for i, result := range ranked {
		result.Rank = i + 1
	}

	return ranked
}

// CalculateScore 计算单个结果的分数
func (r *ranker) CalculateScore(result *SearchResult, query *SearchQuery) float64 {
	score := 0.0

	// 标题匹配度
	titleScore := r.calculateTitleMatchScore(result.Title, query.Keyword)
	score += titleScore * r.weights.TitleMatch

	// 质量分数
	qualityScore := r.calculateQualityScore(result.Title, query.Quality, query.Resolution)
	score += qualityScore * r.weights.Quality

	// 做种数分数
	seedersScore := r.calculateSeedersScore(result.Seeders)
	score += seedersScore * r.weights.Seeders

	// 大小分数
	sizeScore := r.calculateSizeScore(result.Size, query.MinSize, query.MaxSize)
	score += sizeScore * r.weights.Size

	// 发布时间分数
	timeScore := r.calculateTimeScore(result.PublishTime)
	score += timeScore * r.weights.PublishTime

	// 站点可靠性分数
	siteScore := r.calculateSiteScore(result.SiteName)
	score += siteScore * r.weights.SiteReliability

	return score
}

// calculateScoreDetail 计算详细分数
func (r *ranker) calculateScoreDetail(result *SearchResult, query *SearchQuery) map[string]float64 {
	return map[string]float64{
		"title_match":      r.calculateTitleMatchScore(result.Title, query.Keyword),
		"quality":          r.calculateQualityScore(result.Title, query.Quality, query.Resolution),
		"seeders":          r.calculateSeedersScore(result.Seeders),
		"size":             r.calculateSizeScore(result.Size, query.MinSize, query.MaxSize),
		"publish_time":     r.calculateTimeScore(result.PublishTime),
		"site_reliability": r.calculateSiteScore(result.SiteName),
	}
}

// calculateTitleMatchScore 计算标题匹配分数
func (r *ranker) calculateTitleMatchScore(title, keyword string) float64 {
	if keyword == "" {
		return 1.0
	}

	titleLower := strings.ToLower(title)
	keywordLower := strings.ToLower(keyword)

	// 完全匹配
	if titleLower == keywordLower {
		return 1.0
	}

	// 包含完整关键词
	if strings.Contains(titleLower, keywordLower) {
		return 0.8
	}

	// 关键词分词匹配
	keywords := strings.Fields(keywordLower)
	matchCount := 0
	for _, kw := range keywords {
		if strings.Contains(titleLower, kw) {
			matchCount++
		}
	}

	if len(keywords) > 0 {
		return float64(matchCount) / float64(len(keywords)) * 0.6
	}

	return 0.0
}

// calculateQualityScore 计算质量分数
func (r *ranker) calculateQualityScore(title, quality, resolution string) float64 {
	score := 0.0
	titleLower := strings.ToLower(title)

	// 分辨率评分
	if strings.Contains(titleLower, "2160p") || strings.Contains(titleLower, "4k") {
		score += 1.0
	} else if strings.Contains(titleLower, "1080p") {
		score += 0.8
	} else if strings.Contains(titleLower, "720p") {
		score += 0.6
	} else if strings.Contains(titleLower, "480p") {
		score += 0.4
	}

	// 编码评分
	if strings.Contains(titleLower, "h265") || strings.Contains(titleLower, "hevc") {
		score += 0.3
	} else if strings.Contains(titleLower, "h264") || strings.Contains(titleLower, "avc") {
		score += 0.2
	}

	// 来源评分
	if strings.Contains(titleLower, "bluray") || strings.Contains(titleLower, "blu-ray") {
		score += 0.4
	} else if strings.Contains(titleLower, "web-dl") || strings.Contains(titleLower, "webdl") {
		score += 0.3
	} else if strings.Contains(titleLower, "webrip") {
		score += 0.2
	}

	// 音频评分
	if strings.Contains(titleLower, "atmos") {
		score += 0.2
	} else if strings.Contains(titleLower, "dts") || strings.Contains(titleLower, "truehd") {
		score += 0.15
	} else if strings.Contains(titleLower, "ac3") {
		score += 0.1
	}

	// 归一化到 0-1
	return score / 1.9
}

// calculateSeedersScore 计算做种数分数
func (r *ranker) calculateSeedersScore(seeders int) float64 {
	if seeders >= 100 {
		return 1.0
	} else if seeders >= 50 {
		return 0.8
	} else if seeders >= 20 {
		return 0.6
	} else if seeders >= 10 {
		return 0.4
	} else if seeders >= 5 {
		return 0.2
	}
	return 0.0
}

// calculateSizeScore 计算大小分数
func (r *ranker) calculateSizeScore(size, minSize, maxSize int64) float64 {
	// 如果没有限制，返回中等分数
	if minSize == 0 && maxSize == 0 {
		return 0.5
	}

	// 在范围内
	if (minSize == 0 || size >= minSize) && (maxSize == 0 || size <= maxSize) {
		return 1.0
	}

	// 超出范围
	if maxSize > 0 && size > maxSize {
		// 超出越多，分数越低
		excess := float64(size-maxSize) / float64(maxSize)
		if excess > 1.0 {
			return 0.0
		}
		return 1.0 - excess
	}

	if minSize > 0 && size < minSize {
		// 不足越多，分数越低
		deficit := float64(minSize-size) / float64(minSize)
		if deficit > 1.0 {
			return 0.0
		}
		return 1.0 - deficit
	}

	return 0.5
}

// calculateTimeScore 计算发布时间分数
func (r *ranker) calculateTimeScore(publishTime time.Time) float64 {
	if publishTime.IsZero() {
		return 0.5
	}

	age := time.Since(publishTime)

	// 1天内
	if age < 24*time.Hour {
		return 1.0
	}
	// 3天内
	if age < 3*24*time.Hour {
		return 0.8
	}
	// 7天内
	if age < 7*24*time.Hour {
		return 0.6
	}
	// 30天内
	if age < 30*24*time.Hour {
		return 0.4
	}
	// 90天内
	if age < 90*24*time.Hour {
		return 0.2
	}

	return 0.0
}

// calculateSiteScore 计算站点分数
func (r *ranker) calculateSiteScore(siteName string) float64 {
	// TODO: 从数据库获取站点可靠性评分
	// 这里使用简单的规则

	// 知名站点
	knownSites := map[string]float64{
		"mteam":      1.0,
		"hdchina":    1.0,
		"chdbits":    1.0,
		"ourbits":    0.9,
		"hdhome":     0.9,
		"hdsky":      0.9,
		"totheglory": 0.8,
	}

	siteNameLower := strings.ToLower(siteName)
	if score, ok := knownSites[siteNameLower]; ok {
		return score
	}

	return 0.5 // 默认分数
}
