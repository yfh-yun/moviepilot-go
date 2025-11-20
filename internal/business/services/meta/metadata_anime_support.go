package meta

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/meta"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/repositories"
)

// AnimeMetadataService 动漫元数据服务扩展
type AnimeMetadataService struct {
	*MetadataService
	animeParser *meta.MetaAnime
}

// NewAnimeMetadataService 创建动漫元数据服务
func NewAnimeMetadataService(baseService *MetadataService, systemConfigRepo repositories.SystemConfigRepository) *AnimeMetadataService {
	customizationMatcher := meta.NewCustomizationMatcher(systemConfigRepo)
	releaseGroupMatcher := meta.NewReleaseGroupsMatcher(systemConfigRepo)
	animeParser := meta.NewMetaAnime(systemConfigRepo, customizationMatcher, releaseGroupMatcher)

	return &AnimeMetadataService{
		MetadataService: baseService,
		animeParser:     animeParser,
	}
}

// ParseAnimeFile 解析动漫文件（重写基础方法以支持动漫）
func (ams *AnimeMetadataService) ParseAnimeFile(ctx context.Context, filePath string) (*ParseResult, error) {
	// 检查缓存
	if cached, found := ams.cache.Get(filePath); found {
		return cached, nil
	}

	// 解析文件名
	parseResult := &ParseResult{
		FileInfo:   ams.getFileInfo(filePath),
		Confidence: 0.0,
	}

	// 提取文件名信息
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ext)

	// 检查是否为动漫文件
	if ams.isAnimeFile(fileName) {
		// 使用动漫解析器解析
		if err := ams.animeParser.Parse(ctx, fileName, "", true); err != nil {
			parseResult.ParseErrors = append(parseResult.ParseErrors, err.Error())
		} else {
			// 转换动漫解析结果为MediaInfo
			mediaInfo := ams.convertAnimeToMediaInfo(ams.animeParser, filePath)
			parseResult.MediaInfo = mediaInfo
			parseResult.Confidence += 0.6 // 动漫识别置信度更高
		}
	} else {
		// 使用基础解析器
		mediaInfo, err := ams.extractFromFilename(filePath)
		if err != nil {
			parseResult.ParseErrors = append(parseResult.ParseErrors, err.Error())
		} else {
			parseResult.MediaInfo = mediaInfo
			parseResult.Confidence += 0.3
		}
	}

	// 尝试从文件内容解析
	if contentInfo, err := ams.extractFromContent(filePath); err == nil {
		ams.mergeMediaInfo(parseResult.MediaInfo, contentInfo)
		parseResult.Confidence += 0.2
	}

	// 如果置信度足够高，尝试在线查询
	if parseResult.Confidence >= 0.5 && parseResult.MediaInfo != nil {
		if onlineInfo, err := ams.queryOnlineAnime(ctx, parseResult.MediaInfo); err == nil {
			ams.mergeMediaInfo(parseResult.MediaInfo, onlineInfo)
			parseResult.Confidence += 0.2
		}
	}

	// 缓存结果
	ams.cache.Set(filePath, parseResult)

	return parseResult, nil
}

// isAnimeFile 检查是否为动漫文件
func (ams *AnimeMetadataService) isAnimeFile(fileName string) bool {
	animePatterns := []string{
		".*[Aa][Nn][Ii][Mm][Ee].*",
		".*[\\[\\(][Aa][Nn][Ii][Mm][Ee][\\]\\)].*",
		".*第?\\d+[集话話].*",
		".*EP?\\d+.*",
		".*\\[\\d{2,4}\\].*",
		".*[Ss]\\d{2}[Ee]\\d{2}.*", // 也要支持季集格式
	}

	for _, pattern := range animePatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}

	return false
}

// convertAnimeToMediaInfo 将动漫解析结果转换为MediaInfo
func (ams *AnimeMetadataService) convertAnimeToMediaInfo(animeParser *meta.MetaAnime, filePath string) *MediaInfo {
	mediaInfo := &MediaInfo{
		Type:      MediaTypeAnime,
		VideoPath: filePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 设置标题
	if animeParser.CnName != "" {
		mediaInfo.Title = animeParser.CnName
		mediaInfo.OriginalTitle = animeParser.CnName
		mediaInfo.Metadata["cnName"] = animeParser.CnName
	}

	if animeParser.EnName != "" {
		if mediaInfo.Title == "" {
			mediaInfo.Title = animeParser.EnName
			mediaInfo.OriginalTitle = animeParser.EnName
		}
		mediaInfo.Metadata["enName"] = animeParser.EnName
	}

	// 设置季集信息
	if animeParser.Season > 0 {
		mediaInfo.Metadata["season"] = animeParser.Season
	}

	if animeParser.Episode != "" {
		mediaInfo.Metadata["episode"] = animeParser.Episode
	}

	// 设置发布组和自定义匹配
	if animeParser.ReleaseGroup != "" {
		mediaInfo.Metadata["releaseGroup"] = animeParser.ReleaseGroup
	}

	if animeParser.Customization != "" {
		mediaInfo.Metadata["customization"] = animeParser.Customization
	}

	// 设置类型信息
	mediaInfo.Metadata["type"] = "anime"
	if animeParser.Type != "" {
		mediaInfo.Metadata["subtype"] = animeParser.Type
	}

	return mediaInfo
}

// queryOnlineAnime 在线查询动漫信息
func (ams *AnimeMetadataService) queryOnlineAnime(ctx context.Context, mediaInfo *MediaInfo) (*MediaInfo, error) {
	// 优先使用中文名查询，如果没有则使用英文名
	title := mediaInfo.Title
	if cnName, ok := mediaInfo.Metadata["cnName"].(string); ok && cnName != "" {
		title = cnName
	}

	// 尝试从TMDB查询动漫（TMDB有动漫分类）
	movieInfo, err := ams.queryMovie(ctx, title, mediaInfo.Year)
	if err == nil {
		// 检查是否为动漫类型
		if ams.isAnimeMediaType(movieInfo) {
			movieInfo.Type = MediaTypeAnime
			return movieInfo, nil
		}
	}

	// 如果TMDB没有找到，尝试TV API（动漫通常以TV形式存储）
	tvInfo, err := ams.queryTV(ctx, title, mediaInfo.Year)
	if err == nil {
		tvInfo.Type = MediaTypeAnime
		return tvInfo, nil
	}

	return nil, fmt.Errorf("未找到动漫信息: %s", title)
}

// isAnimeMediaType 检查TMDB返回的媒体是否为动漫类型
func (ams *AnimeMetadataService) isAnimeMediaType(mediaInfo *MediaInfo) bool {
	// 检查类型关键词
	animeKeywords := []string{
		"animation", "anime", "cartoon", "animated",
		"动画", "动漫", "卡通",
	}

	// 检查类型
	for _, genre := range mediaInfo.Genres {
		genreLower := strings.ToLower(genre)
		for _, keyword := range animeKeywords {
			if strings.Contains(genreLower, keyword) {
				return true
			}
		}
	}

	// 检查标题中是否包含动漫关键词
	titleLower := strings.ToLower(mediaInfo.Title)
	for _, keyword := range animeKeywords {
		if strings.Contains(titleLower, keyword) {
			return true
		}
	}

	return false
}

// EnhancedPatterns 扩展的模式匹配
func (ams *AnimeMetadataService) EnhancedPatterns() []struct {
	Pattern string
	Type    MediaType
} {
	return []struct {
		Pattern string
		Type    MediaType
	}{
		// 基础模式
		{".+[Ss]\\d{2}[Ee]\\d{2}.*", MediaTypeEpisode},
		{".+\\d{4}\\.\\d{2}\\.\\d{2}.*", MediaTypeMovie},
		{".+\\.S\\d{2}E\\d{2}\\..*", MediaTypeEpisode},
		{".+\\.\\d{4}\\..*", MediaTypeMovie},

		// 动漫模式（更高优先级）
		{".*[\\[\\(][Aa][Nn][Ii][Mm][Ee][\\]\\)].*", MediaTypeAnime},
		{".*第?\\d+[集話話].*", MediaTypeAnime},
		{".*EP?\\d+.*", MediaTypeAnime},
		{".*\\[\\d{2,4}\\].*", MediaTypeAnime},
		{".*[Aa][Nn][Ii][Mm][Ee].*", MediaTypeAnime},

		// 动漫特殊格式
		{".*\\[\\d{2,4}\\].*\\[\\d{3,4}[Pp]\\].*", MediaTypeAnime}, // [1080P]
		{".*S\\d+.*EP?\\d+.*", MediaTypeAnime},                     // S01E01格式
		{".*第\\d+季.*第\\d+集.*", MediaTypeAnime},                     // 第1季第1集格式
	}
}
