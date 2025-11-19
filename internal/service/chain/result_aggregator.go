package chain

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// NewResultAggregator 创建结果聚合器
func NewResultAggregator(logger *zap.Logger) *ResultAggregator {
	return &ResultAggregator{
		logger:  logger,
		weights: map[string]float64{
			"tmdb":     1.0,  // TMDB 最权威
			"imdb":     0.9,  // IMDB 次权威
			"tvdb":     0.8,  // TVDB 电视剧
			"douban":   0.7,  // 豆瓣中文内容
			"bangumi":  0.6,  // 番组计划动漫
		},
		confidenceCalculator: &ConfidenceCalculator{
			titleMatcher:   utils.NewTitleMatcher(),
			yearComparator: utils.NewYearComparator(),
			ratingAnalyzer: utils.NewRatingAnalyzer(),
		},
	}
}

// AggregateResults 聚合刮削结果
func (r *ResultAggregator) AggregateResults(results []*ScrapeResult, req *SearchRequest) (*model.MediaInfo, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("没有可用的刮削结果")
	}

	r.logger.Debug("开始聚合刮削结果",
		zap.Int("total_results", len(results)),
		zap.String("search_title", req.Title))

	// 1. 过滤有效结果
	validResults := r.filterValidResults(results)
	if len(validResults) == 0 {
		return nil, fmt.Errorf("没有有效的刮削结果")
	}

	// 2. 按置信度排序
	sort.Slice(validResults, func(i, j int) bool {
		return validResults[i].Confidence > validResults[j].Confidence
	})

	// 3. 选择最佳结果
	bestResult := validResults[0]
	if bestResult.Data == nil {
		return nil, fmt.Errorf("最佳结果为空")
	}

	// 4. 聚合其他数据源的信息
	aggregated := r.aggregateAdditionalInfo(bestResult.Data, validResults[1:], req)

	r.logger.Info("结果聚合完成",
		zap.String("best_source", bestResult.Source),
		zap.Float64("confidence", bestResult.Confidence),
		zap.String("title", aggregated.Title))

	return aggregated, nil
}

// AggregateIMDBResults 聚合IMDB搜索结果
func (r *ResultAggregator) AggregateIMDBResults(results []*ScrapeResult, imdbID string) (*model.MediaInfo, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("没有IMDB搜索结果")
	}

	validResults := r.filterValidResults(results)
	if len(validResults) == 0 {
		return nil, fmt.Errorf("没有有效的IMDB搜索结果")
	}

	// 选择权重最高的数据源
	sort.Slice(validResults, func(i, j int) bool {
		weightI := r.weights[validResults[i].Source]
		weightJ := r.weights[validResults[j].Source]
		return weightI > weightJ
	})

	aggregated := r.aggregateAdditionalInfo(validResults[0].Data, validResults[1:], nil)
	aggregated.IMDBID = &imdbID

	return aggregated, nil
}

// AggregateTMDBResults 聚合TMDB搜索结果
func (r *ResultAggregator) AggregateTMDBResults(results []*ScrapeResult, tmdbID int) (*model.MediaInfo, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("没有TMDB搜索结果")
	}

	validResults := r.filterValidResults(results)
	if len(validResults) == 0 {
		return nil, fmt.Errorf("没有有效的TMDB搜索结果")
	}

	// 选择权重最高的数据源
	sort.Slice(validResults, func(i, j int) bool {
		weightI := r.weights[validResults[i].Source]
		weightJ := r.weights[validResults[j].Source]
		return weightI > weightJ
	})

	aggregated := r.aggregateAdditionalInfo(validResults[0].Data, validResults[1:], nil)
	aggregated.TMDBID = &tmdbID

	return aggregated, nil
}

// filterValidResults 过滤有效结果
func (r *ResultAggregator) filterValidResults(results []*ScrapeResult) []*ScrapeResult {
	var valid []*ScrapeResult
	for _, result := range results {
		if result.Data != nil && result.Error == nil {
			valid = append(valid, result)
		}
	}
	return valid
}

// aggregateAdditionalInfo 聚合额外信息
func (r *ResultAggregator) aggregateAdditionalInfo(base *model.MediaInfo, others []*ScrapeResult, req *SearchRequest) *model.MediaInfo {
	aggregated := *base // 复制基础信息

	// 聚合其他数据源的信息
	for _, result := range others {
		if result.Data == nil {
			continue
		}

		// 补充缺失的ID
		if aggregated.TMDBID == nil && result.Data.TMDBID != nil {
			aggregated.TMDBID = result.Data.TMDBID
		}
		if aggregated.IMDBID == nil && result.Data.IMDBID != nil {
			aggregated.IMDBID = result.Data.IMDBID
		}
		if aggregated.TVDBID == nil && result.Data.TVDBID != nil {
			aggregated.TVDBID = result.Data.TVDBID
		}
		if aggregated.DoubanID == nil && result.Data.DoubanID != nil {
			aggregated.DoubanID = result.Data.DoubanID
		}
		if aggregated.BangumiID == nil && result.Data.BangumiID != nil {
			aggregated.BangumiID = result.Data.BangumiID
		}

		// 选择更好的标题和描述
		if r.isBetterTitle(result.Data.Title, aggregated.Title, req) {
			aggregated.Title = result.Data.Title
		}
		if r.isBetterDescription(result.Data.Description, aggregated.Description) {
			aggregated.Description = result.Data.Description
		}

		// 选择更高的评分
		if result.Data.Vote != nil && (aggregated.Vote == nil || *result.Data.Vote > *aggregated.Vote) {
			aggregated.Vote = result.Data.Vote
		}

		// 合并类型和国家信息
		aggregated.Genres = r.mergeGenres(aggregated.Genres, result.Data.Genres)
		aggregated.Countries = r.mergeCountries(aggregated.Countries, result.Data.Countries)

		// 选择更好的海报和背景图
		if r.isBetterImage(result.Data.Poster, aggregated.Poster) {
			aggregated.Poster = result.Data.Poster
		}
		if r.isBetterImage(result.Data.Backdrop, aggregated.Backdrop) {
			aggregated.Backdrop = result.Data.Backdrop
		}

		// 补充年份信息
		if result.Data.Year != nil && (aggregated.Year == nil || *result.Data.Year > 1900) {
			aggregated.Year = result.Data.Year
		}
	}

	// 设置置信度
	aggregated.Confidence = r.calculateOverallConfidence(&aggregated, req)

	return &aggregated
}

// isBetterTitle 判断是否是更好的标题
func (r *ResultAggregator) isBetterTitle(newTitle, currentTitle string, req *SearchRequest) bool {
	if currentTitle == "" {
		return true
	}
	if newTitle == "" {
		return false
	}

	if req != nil {
		// 基于搜索请求的标题相似度
		newSim := utils.CalculateTitleSimilarity(newTitle, req.Title)
		currentSim := utils.CalculateTitleSimilarity(currentTitle, req.Title)
		return newSim > currentSim
	}

	// 优先选择非英文标题（如果当前是英文）
	if r.isEnglishTitle(currentTitle) && !r.isEnglishTitle(newTitle) {
		return true
	}

	return len(newTitle) > len(currentTitle)
}

// isBetterDescription 判断是否是更好的描述
func (r *ResultAggregator) isBetterDescription(newDesc, currentDesc string) bool {
	if currentDesc == "" {
		return newDesc != ""
	}
	if newDesc == "" {
		return false
	}

	// 优先选择更长的描述（通常更详细）
	return len(newDesc) > len(currentDesc)
}

// isBetterImage 判断是否是更好的图片
func (r *ResultAggregator) isBetterImage(newImage, currentImage string) bool {
	if currentImage == "" {
		return newImage != ""
	}
	if newImage == "" {
		return false
	}

	// 优先选择高分辨率图片
	if strings.Contains(newImage, "original") || strings.Contains(newImage, "1280") {
		return true
	}
	if strings.Contains(currentImage, "original") || strings.Contains(currentImage, "1280") {
		return false
	}

	return len(newImage) > len(currentImage)
}

// mergeGenres 合并类型信息
func (r *ResultAggregator) mergeGenres(current, new string) string {
	if current == "" {
		return new
	}
	if new == "" {
		return current
	}

	currentGenres := parseJSONArray(current)
	newGenres := parseJSONArray(new)

	// 去重合并
	genreSet := make(map[string]bool)
	for _, genre := range currentGenres {
		genreSet[genre] = true
	}
	for _, genre := range newGenres {
		genreSet[genre] = true
	}

	// 转换为数组格式
	var mergedGenres []string
	for genre := range genreSet {
		mergedGenres = append(mergedGenres, genre)
	}

	return formatJSONArray(mergedGenres)
}

// mergeCountries 合并国家信息
func (r *ResultAggregator) mergeCountries(current, new string) string {
	if current == "" {
		return new
	}
	if new == "" {
		return current
	}

	currentCountries := parseJSONArray(current)
	newCountries := parseJSONArray(new)

	countrySet := make(map[string]bool)
	for _, country := range currentCountries {
		countrySet[country] = true
	}
	for _, country := range newCountries {
		countrySet[country] = true
	}

	var mergedCountries []string
	for country := range countrySet {
		mergedCountries = append(mergedCountries, country)
	}

	return formatJSONArray(mergedCountries)
}

// calculateOverallConfidence 计算整体置信度
func (r *ResultAggregator) calculateOverallConfidence(media *model.MediaInfo, req *SearchRequest) float64 {
	var score float64 = 0.0

	// 基础分数（有核心信息）
	if media.Title != "" {
		score += 0.3
	}
	if media.TMDBID != nil {
		score += 0.3
	}
	if media.IMDBID != nil {
		score += 0.2
	}
	if media.Year != nil {
		score += 0.1
	}
	if media.Description != "" {
		score += 0.05
	}
	if media.Poster != "" {
		score += 0.05
	}

	// 数据源完整性加分
	idCount := 0
	if media.TMDBID != nil {
		idCount++
	}
	if media.IMDBID != nil {
		idCount++
	}
	if media.TVDBID != nil {
		idCount++
	}
	if media.DoubanID != nil {
		idCount++
	}
	if media.BangumiID != nil {
		idCount++
	}

	score += float64(idCount) * 0.1

	return min(score, 1.0)
}

// isEnglishTitle 判断是否是英文标题
func (r *ResultAggregator) isEnglishTitle(title string) bool {
	for _, r := range title {
		if r > 127 {
			return false // 包含非ASCII字符
		}
	}
	return true
}

// parseJSONArray 解析JSON数组
func parseJSONArray(jsonStr string) []string {
	// 简单的JSON数组解析（实际项目中应该使用标准库）
	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "[") || !strings.HasSuffix(jsonStr, "]") {
		return []string{}
	}

	content := jsonStr[1 : len(jsonStr)-1]
	if content == "" {
		return []string{}
	}

	var items []string
	for _, item := range strings.Split(content, ",") {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, `"`)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

// formatJSONArray 格式化JSON数组
func formatJSONArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	var quotedItems []string
	for _, item := range items {
		quotedItems = append(quotedItems, `"`+item+`"`)
	}

	return "[" + strings.Join(quotedItems, ",") + "]"
}

// min 计算最小值
func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}