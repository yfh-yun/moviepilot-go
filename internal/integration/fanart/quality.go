package fanart

import (
	"sort"
)

// QualityFilter 质量过滤器
type QualityFilter struct {
	// 最小质量阈值
	MinQuality int
	// 优先语言
	PreferredLanguage string
	// 允许的语言列表
	AllowedLanguages []string
}

// NewQualityFilter 创建质量过滤器
func NewQualityFilter(minQuality int, preferredLang string, allowedLangs []string) *QualityFilter {
	return &QualityFilter{
		MinQuality:        minQuality,
		PreferredLanguage: preferredLang,
		AllowedLanguages:  allowedLangs,
	}
}

// FilterImages 过滤图片
func (f *QualityFilter) FilterImages(images []Image) []Image {
	if len(images) == 0 {
		return images
	}

	// 分组图片
	grouped := f.groupImagesByLanguage(images)

	// 优选指定语言
	if preferred, exists := grouped[f.PreferredLanguage]; exists {
		filtered := f.filterByQuality(preferred)
		if len(filtered) > 0 {
			return f.sortByQuality(filtered)
		}
	}

	// 如果没有指定语言的图片，选择最高质量的图片
	var allImages []Image
	for _, langImages := range grouped {
		if f.isLanguageAllowed(langImages[0].Lang) {
			allImages = append(allImages, f.filterByQuality(langImages)...)
		}
	}

	return f.sortByQuality(allImages)
}

// groupImagesByLanguage 按语言分组图片
func (f *QualityFilter) groupImagesByLanguage(images []Image) map[string][]Image {
	grouped := make(map[string][]Image)

	for _, img := range images {
		lang := img.Lang
		if lang == "" {
			lang = "en" // 默认英语
		}

		grouped[lang] = append(grouped[lang], img)
	}

	return grouped
}

// filterByQuality 按质量过滤图片
func (f *QualityFilter) filterByQuality(images []Image) []Image {
	var filtered []Image

	for _, img := range images {
		likes := f.parseLikes(img.Likes)
		if likes >= f.MinQuality {
			filtered = append(filtered, img)
		}
	}

	return filtered
}

// sortByQuality 按质量排序图片
func (f *QualityFilter) sortByQuality(images []Image) []Image {
	sort.Slice(images, func(i, j int) bool {
		likesI := f.parseLikes(images[i].Likes)
		likesJ := f.parseLikes(images[j].Likes)
		return likesI > likesJ
	})

	return images
}

// isLanguageAllowed 检查语言是否允许
func (f *QualityFilter) isLanguageAllowed(lang string) bool {
	if len(f.AllowedLanguages) == 0 {
		return true
	}

	for _, allowed := range f.AllowedLanguages {
		if allowed == lang {
			return true
		}
	}

	return false
}

// parseLikes 解析likes字符串为数字
func (f *QualityFilter) parseLikes(likesStr string) int {
	likes := 0

	// 简单的likes解析
	for _, char := range likesStr {
		if char >= '0' && char <= '9' {
			likes = likes*10 + int(char-'0')
		}
	}

	return likes
}

// GetBestImage 获取最佳图片
func (f *QualityFilter) GetBestImage(images []Image) *Image {
	if len(images) == 0 {
		return nil
	}

	filtered := f.FilterImages(images)
	if len(filtered) > 0 {
		return &filtered[0]
	}

	return nil
}

// GetBestMoviePoster 获取最佳电影海报
func (f *QualityFilter) GetBestMoviePoster(movie *MovieImages) *Image {
	return f.GetBestImage(movie.Posters)
}

// GetBestTVPoster 获取最佳电视剧海报
func (f *QualityFilter) GetBestTVPoster(tv *TVImages) *Image {
	return f.GetBestImage(tv.Posters)
}

// GetBestSeasonPoster 获取最佳季海报
func (f *QualityFilter) GetBestSeasonPoster(tv *TVImages, season int) *Image {
	seasonKey := fmt.Sprintf("%d", season)
	if seasonData, exists := tv.Seasons[seasonKey]; exists {
		return f.GetBestImage(seasonData.Posters)
	}
	return nil
}
