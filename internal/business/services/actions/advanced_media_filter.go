// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// AdvancedMediaFilter 高级媒体过滤器
// 提供智能媒体内容过滤和分类功能
type AdvancedMediaFilter struct {
	mediaRepo      repository.MediaRepository
	categoryRepo   repository.CategoryRepository
	tagRepo        repository.TagRepository
	logger         *zap.Logger
}

// NewAdvancedMediaFilter 创建高级媒体过滤器实例
func NewAdvancedMediaFilter(
	mediaRepo repository.MediaRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) *AdvancedMediaFilter {
	return &AdvancedMediaFilter{
		mediaRepo:      mediaRepo,
		categoryRepo:   categoryRepo,
		tagRepo:        tagRepo,
		logger:         logger.NewLogger("advanced_media_filter"),
	}
}

// FilterMedia 智能媒体过滤
func (f *AdvancedMediaFilter) FilterMedia(ctx context.Context, media *model.Media, filterRules *MediaFilterRules) (*model.Media, error) {
	f.logger.Info("开始智能媒体过滤", 
		zap.String("media_title", media.Title),
		zap.String("media_type", string(media.Type)))

	// 1. 基础信息过滤
	if err := f.filterBasicInfo(ctx, media, filterRules); err != nil {
		return nil, fmt.Errorf("基础信息过滤失败: %w", err)
	}

	// 2. 内容质量评估
	qualityScore, err := f.evaluateContentQuality(ctx, media)
	if err != nil {
		return nil, fmt.Errorf("内容质量评估失败: %w", err)
	}

	// 3. 智能分类
	if err := f.intelligentClassification(ctx, media); err != nil {
		return nil, fmt.Errorf("智能分类失败: %w", err)
	}

	// 4. 标签优化
	if err := f.optimizeTags(ctx, media); err != nil {
		return nil, fmt.Errorf("标签优化失败: %w", err)
	}

	// 设置质量评分
	media.QualityScore = qualityScore
	media.FilteredAt = time.Now()

	f.logger.Info("媒体过滤完成", 
		zap.String("media_title", media.Title),
		zap.Float64("quality_score", qualityScore),
		zap.Int("category_count", len(media.Categories)))

	return media, nil
}

// filterBasicInfo 基础信息过滤
func (f *AdvancedMediaFilter) filterBasicInfo(ctx context.Context, media *model.Media, rules *MediaFilterRules) error {
	// 标题长度检查
	if len(strings.TrimSpace(media.Title)) == 0 {
		return fmt.Errorf("标题不能为空")
	}

	// 发布日期检查
	if media.ReleaseDate.IsZero() {
		return fmt.Errorf("发布日期无效")
	}

	// 文件大小检查
	if media.FileSize <= 0 {
		return fmt.Errorf("文件大小无效")
	}

	// 重复内容检查
	existing, err := f.mediaRepo.FindByTitleAndType(ctx, media.Title, media.Type)
	if err == nil && existing != nil {
		return fmt.Errorf("重复的媒体内容: %s", media.Title)
	}

	return nil
}

// evaluateContentQuality 内容质量评估
func (f *AdvancedMediaFilter) evaluateContentQuality(ctx context.Context, media *model.Media) (float64, error) {
	var score float64

	// 1. 标题质量评分（0-25分）
	titleScore := f.evaluateTitleQuality(media.Title)
	score += titleScore

	// 2. 描述质量评分（0-25分）
	descriptionScore := f.evaluateDescriptionQuality(media.Description)
	score += descriptionScore

	// 3. 元数据完整度评分（0-30分）
	metadataScore := f.evaluateMetadataCompleteness(media)
	score += metadataScore

	// 4. 文件质量评分（0-20分）
	fileScore := f.evaluateFileQuality(media)
	score += fileScore

	return score, nil
}

// evaluateTitleQuality 标题质量评估
func (f *AdvancedMediaFilter) evaluateTitleQuality(title string) float64 {
	var score float64

	// 长度适中（5-50字符）
	titleLength := len(strings.TrimSpace(title))
	if titleLength >= 5 && titleLength <= 50 {
		score += 10
	} else if titleLength >= 3 && titleLength <= 100 {
		score += 5
	}

	// 不包含特殊字符
	if !strings.ContainsAny(title, "<>[]{}|\\^~`")
		score += 5
	}

	// 包含有意义的关键词
	meaningfulKeywords := []string{"电影", "电视剧", "纪录片", "动漫", "音乐", "游戏"}
	for _, keyword := range meaningfulKeywords {
		if strings.Contains(title, keyword) {
			score += 10
			break
		}
	}

	return score
}

// evaluateDescriptionQuality 描述质量评估
func (f *AdvancedMediaFilter) evaluateDescriptionQuality(description string) float64 {
	if len(strings.TrimSpace(description)) == 0 {
		return 0
	}

	var score float64
	descLength := len(strings.TrimSpace(description))

	// 描述长度适中（50-500字符）
	if descLength >= 50 && descLength <= 500 {
		score += 15
	} else if descLength >= 20 && descLength <= 1000 {
		score += 10
	}

	// 描述内容丰富度
	if strings.Contains(description, "主演") || strings.Contains(description, "导演") || 
		strings.Contains(description, "剧情") || strings.Contains(description, "简介") {
		score += 10
	}

	return score
}

// evaluateMetadataCompleteness 元数据完整度评估
func (f *AdvancedMediaFilter) evaluateMetadataCompleteness(media *model.Media) float64 {
	var score float64

	// 基础元数据完整性
	if media.ReleaseDate.Year() > 1900 {
		score += 5
	}
	if media.Duration > 0 {
		score += 5
	}
	if media.FileSize > 0 {
		score += 5
	}

	// 扩展元数据
	if len(media.Cast) > 0 {
		score += 5
	}
	if len(media.Directors) > 0 {
		score += 5
	}
	if len(media.Genres) > 0 {
		score += 5
	}

	return score
}

// evaluateFileQuality 文件质量评估
func (f *AdvancedMediaFilter) evaluateFileQuality(media *model.Media) float64 {
	var score float64

	// 文件格式支持
	supportedFormats := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv"}
	for _, format := range supportedFormats {
		if strings.HasSuffix(strings.ToLower(media.FilePath), format) {
			score += 10
			break
		}
	}

	// 文件大小合理性
	if media.FileSize > 10*1024*1024 && media.FileSize < 10*1024*1024*1024 {
		score += 10
	}

	return score
}

// intelligentClassification 智能分类
func (f *AdvancedMediaFilter) intelligentClassification(ctx context.Context, media *model.Media) error {
	// 基于标题和描述的智能分类
	categories, err := f.classifyByContent(ctx, media)
	if err != nil {
		return err
	}

	// 基于文件属性的分类
	fileCategories, err := f.classifyByFileAttributes(ctx, media)
	if err != nil {
		return err
	}

	// 合并分类结果
	allCategories := append(categories, fileCategories...)
	media.Categories = f.deduplicateCategories(allCategories)

	return nil
}

// classifyByContent 基于内容分类
func (f *AdvancedMediaFilter) classifyByContent(ctx context.Context, media *model.Media) ([]model.Category, error) {
	var categories []model.Category

	content := strings.ToLower(media.Title + " " + media.Description)

	// 影视类型识别
	if strings.Contains(content, "电影") || strings.Contains(content, "movie") {
		if category, err := f.categoryRepo.FindByName(ctx, "电影"); err == nil && category != nil {
			categories = append(categories, *category)
		}
	}

	if strings.Contains(content, "电视剧") || strings.Contains(content, "tv series") {
		if category, err := f.categoryRepo.FindByName(ctx, "电视剧"); err == nil && category != nil {
			categories = append(categories, *category)
		}
	}

	// 其他类型识别逻辑...

	return categories, nil
}

// classifyByFileAttributes 基于文件属性分类
func (f *AdvancedMediaFilter) classifyByFileAttributes(ctx context.Context, media *model.Media) ([]model.Category, error) {
	var categories []model.Category

	// 根据文件扩展名分类
	fileExt := strings.ToLower(media.FilePath[strings.LastIndex(media.FilePath, "."):])
	switch fileExt {
	case ".mp4", ".mkv", ".avi":
		if category, err := f.categoryRepo.FindByName(ctx, "视频"); err == nil && category != nil {
			categories = append(categories, *category)
		}
	case ".mp3", ".flac", ".wav":
		if category, err := f.categoryRepo.FindByName(ctx, "音频"); err == nil && category != nil {
			categories = append(categories, *category)
		}
	}

	return categories, nil
}

// optimizeTags 标签优化
func (f *AdvancedMediaFilter) optimizeTags(ctx context.Context, media *model.Media) error {
	// 提取关键词作为标签
	keywords := f.extractKeywords(media.Title + " " + media.Description)
	
	var tags []model.Tag
	for _, keyword := range keywords {
		if tag, err := f.tagRepo.FindByName(ctx, keyword); err == nil && tag != nil {
			tags = append(tags, *tag)
		} else {
			// 创建新标签
			newTag := &model.Tag{
				Name:      keyword,
				CreatedAt: time.Now(),
			}
			if createdTag, err := f.tagRepo.Create(ctx, newTag); err == nil {
				tags = append(tags, *createdTag)
			}
		}
	}

	media.Tags = tags
	return nil
}

// extractKeywords 提取关键词
func (f *AdvancedMediaFilter) extractKeywords(text string) []string {
	// 简单的关键词提取逻辑
	words := strings.Fields(text)
	var keywords []string

	// 常见停用词
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true,
		"有": true, "和": true, "就": true, "不": true, "人": true,
		"都": true, "一": true, "一个": true, "上": true, "也": true,
		"很": true, "到": true, "说": true, "要": true, "去": true,
	}

	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) >= 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// deduplicateCategories 去重分类
func (f *AdvancedMediaFilter) deduplicateCategories(categories []model.Category) []model.Category {
	seen := make(map[string]bool)
	var result []model.Category

	for _, category := range categories {
		if !seen[category.Name] {
			seen[category.Name] = true
			result = append(result, category)
		}
	}

	return result
}

// MediaFilterRules 媒体过滤规则
type MediaFilterRules struct {
	MinTitleLength     int     `json:"min_title_length"`
	MaxTitleLength     int     `json:"max_title_length"`
	MinQualityScore    float64 `json:"min_quality_score"`
	RequireDescription bool    `json:"require_description"`
	AllowedFormats     []string `json:"allowed_formats"`
	MaxFileSize        int64   `json:"max_file_size"`
	MinFileSize        int64   `json:"min_file_size"`
}