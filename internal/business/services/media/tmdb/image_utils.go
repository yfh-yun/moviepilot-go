package tmdb

import (
	"path/filepath"
	"strings"
)

// GetBestImage 获取最佳质量的图片
func GetBestImage(images []Image, preferLanguage string) *Image {
	if len(images) == 0 {
		return nil
	}

	var bestImage *Image
	var bestScore float64 = 0

	for _, img := range images {
		score := calculateImageScore(img, preferLanguage)
		if score > bestScore {
			bestScore = score
			bestImage = &img
		}
	}

	return bestImage
}

// calculateImageScore 计算图片质量分数
func calculateImageScore(img Image, preferLanguage string) float64 {
	var score float64 = 0

	// 分辨率权重 (30%)
	if img.Width > 1920 {
		score += 30
	} else if img.Width > 1280 {
		score += 25
	} else if img.Width > 720 {
		score += 20
	} else {
		score += 10
	}

	// 宽高比权重 (20%)
	if img.AspectRatio > 1.7 && img.AspectRatio < 2.0 {
		score += 20
	} else if img.AspectRatio > 1.3 && img.AspectRatio < 2.4 {
		score += 15
	} else {
		score += 5
	}

	// 评分权重 (20%)
	score += img.VoteAverage * 4

	// 投票数权重 (15%)
	if img.VoteCount > 100 {
		score += 15
	} else if img.VoteCount > 50 {
		score += 10
	} else if img.VoteCount > 10 {
		score += 5
	}

	// 语言权重 (15%)
	if preferLanguage != "" {
		if img.ISO6391 == preferLanguage {
			score += 15
		} else if img.ISO6391 == "" || img.ISO6391 == "xx" {
			score += 10
		}
	} else {
		if img.ISO6391 == "" || img.ISO6391 == "xx" {
			score += 15
		}
	}

	return score
}

// GetImageExtension 获取图片文件扩展名
func GetImageExtension(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return ".jpg" // 默认扩展名
	}
	return ext
}

// FilterImagesBySize 按尺寸过滤图片
func FilterImagesBySize(images []Image, minWidth, minHeight int) []Image {
	if len(images) == 0 {
		return images
	}

	var filtered []Image
	for _, img := range images {
		if img.Width >= minWidth && img.Height >= minHeight {
			filtered = append(filtered, img)
		}
	}

	return filtered
}

// FilterImagesByAspectRatio 按宽高比过滤图片
func FilterImagesByAspectRatio(images []Image, minRatio, maxRatio float64) []Image {
	if len(images) == 0 {
		return images
	}

	var filtered []Image
	for _, img := range images {
		if img.AspectRatio >= minRatio && img.AspectRatio <= maxRatio {
			filtered = append(filtered, img)
		}
	}

	return filtered
}

// FilterImagesByLanguage 按语言过滤图片
func FilterImagesByLanguage(images []Image, languages []string) []Image {
	if len(images) == 0 || len(languages) == 0 {
		return images
	}

	var filtered []Image
	langSet := make(map[string]bool)
	for _, lang := range languages {
		langSet[strings.ToLower(lang)] = true
	}

	for _, img := range images {
		if img.ISO6391 == "" || langSet[strings.ToLower(img.ISO6391)] {
			filtered = append(filtered, img)
		}
	}

	return filtered
}
