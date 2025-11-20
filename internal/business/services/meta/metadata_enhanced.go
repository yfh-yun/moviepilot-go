package meta

import (
	"context"
	"path/filepath"
	"strings"
)

// EnhancedExtractFromFilename 增强的文件名提取（支持动漫）
func (ms *MetadataService) EnhancedExtractFromFilename(filePath string) (*MediaInfo, error) {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ext)

	// 首先检查是否为动漫文件
	if animeInfo := ms.checkAnimePatterns(fileName); animeInfo != nil {
		return animeInfo, nil
	}

	// 如果不是动漫，使用基础方法
	return ms.extractFromFilename(filePath)
}

// checkAnimePatterns 检查动漫模式
func (ms *MetadataService) checkAnimePatterns(fileName string) *MediaInfo {
	animePatterns := []struct {
		pattern string
		handler func(string) *MediaInfo
	}{
		// 动漫标记模式
		{"[\\[\\(][Aa][Nn][Ii][Mm][Ee][\\]\\)]", ms.parseAnimeWithMarker},
		// 集数模式
		{"第?\\d+[集话話]", ms.parseAnimeWithEpisode},
		// EP模式
		{"EP?\\d+", ms.parseAnimeWithEP},
		// 数字括号模式
		{"\\[\\d{2,4}\\]", ms.parseAnimeWithNumber},
		// ANIME关键字模式
		{"[Aa][Nn][Ii][Mm][Ee]", ms.parseAnimeWithKeyword},
	}

	for _, p := range animePatterns {
		if matched, _ := filepath.Match(p.pattern, fileName); matched {
			return p.handler(fileName)
		}
	}

	return nil
}

// parseAnimeWithMarker 解析带标记的动漫文件
func (ms *MetadataService) parseAnimeWithMarker(fileName string) *MediaInfo {
	// 移除动漫标记
	cleanName := regexp.MustCompile(`[\\[\\(][Aa][Nn][Ii][Mm][Ee][\\]\\)]`).ReplaceAllString(fileName, "")
	
	return &MediaInfo{
		Title:     ms.cleanTitle(cleanName),
		Type:       MediaTypeAnime,
		Metadata: map[string]interface{}{
			"animeType": "marked",
			"original":  fileName,
		},
	}
}

// parseAnimeWithEpisode 解析带集数的动漫文件
func (ms *MetadataService) parseAnimeWithEpisode(fileName string) *MediaInfo {
	// 提取集数
	episodeRe := regexp.MustCompile(`第?(\d+)[集话話]`)
	episodeMatch := episodeRe.FindStringSubmatch(fileName)
	
	mediaInfo := &MediaInfo{
		Type:     MediaTypeAnime,
		Metadata: make(map[string]interface{}),
	}
	
	if len(episodeMatch) > 1 {
		mediaInfo.Metadata["episode"] = episodeMatch[1]
		// 移除集数信息获取标题
		cleanName := episodeRe.ReplaceAllString(fileName, "")
		mediaInfo.Title = ms.cleanTitle(cleanName)
	} else {
		mediaInfo.Title = ms.cleanTitle(fileName)
	}
	
	return mediaInfo
}

// parseAnimeWithEP 解析EP格式的动漫文件
func (ms *MetadataService) parseAnimeWithEP(fileName string) *MediaInfo {
	// 提取EP号数
	episodeRe := regexp.MustCompile(`EP?(\d+)`)
	episodeMatch := episodeRe.FindStringSubmatch(fileName)
	
	mediaInfo := &MediaInfo{
		Type:     MediaTypeAnime,
		Metadata: make(map[string]interface{}),
	}
	
	if len(episodeMatch) > 1 {
		mediaInfo.Metadata["episode"] = episodeMatch[1]
		// 移除EP信息获取标题
		cleanName := episodeRe.ReplaceAllString(fileName, "")
		mediaInfo.Title = ms.cleanTitle(cleanName)
	} else {
		mediaInfo.Title = ms.cleanTitle(fileName)
	}
	
	return mediaInfo
}

// parseAnimeWithNumber 解析带数字括号的动漫文件
func (ms *MetadataService) parseAnimeWithNumber(fileName string) *MediaInfo {
	// 提取数字
	numberRe := regexp.MustCompile(`\\[(\d{2,4})\\]`)
	numberMatch := numberRe.FindStringSubmatch(fileName)
	
	mediaInfo := &MediaInfo{
		Type:     MediaTypeAnime,
		Metadata: make(map[string]interface{}),
	}
	
	if len(numberMatch) > 1 {
		mediaInfo.Metadata["episode"] = numberMatch[1]
		// 移除数字信息获取标题
		cleanName := numberRe.ReplaceAllString(fileName, "")
		mediaInfo.Title = ms.cleanTitle(cleanName)
	} else {
		mediaInfo.Title = ms.cleanTitle(fileName)
	}
	
	return mediaInfo
}

// parseAnimeWithKeyword 解析带ANIME关键字的动漫文件
func (ms *MetadataService) parseAnimeWithKeyword(fileName string) *MediaInfo {
	// 移除ANIME关键字
	cleanName := regexp.MustCompile(`[Aa][Nn][Ii][Mm][Ee]`).ReplaceAllString(fileName, "")
	
	return &MediaInfo{
		Title:     ms.cleanTitle(cleanName),
		Type:       MediaTypeAnime,
		Metadata: map[string]interface{}{
			"animeType": "keyword",
			"original":  fileName,
		},
	}
}

// SupportAnimeMediaType 检查是否支持动漫媒体类型
func (ms *MetadataService) SupportAnimeMediaType(mediaType MediaType) bool {
	switch mediaType {
	case MediaTypeAnime, MediaTypeEpisode:
		return true
	default:
		return false
	}
}