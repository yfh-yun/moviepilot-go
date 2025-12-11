package media

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

// MetaInfoResult 内嵌元信息解析结果
type MetaInfoResult struct {
	TMDBID       int64
	DoubanID     string
	Type         MediaType
	BeginSeason  *int
	EndSeason    *int
	TotalSeason  *int
	BeginEpisode *int
	EndEpisode   *int
	TotalEpisode *int
}

// MetaService 元信息解析服务接口
type MetaService interface {
	MetaInfo(ctx context.Context, title, subtitle string, isFile bool, customWords []string) (*MetaBase, error)
	MetaInfoPath(ctx context.Context, path string) (*MetaBase, error)
	IsAnime(title string) bool
	FindMetaInfo(title string) (string, *MetaInfoResult)
}

// metaService 元信息解析服务实现
type metaService struct {
	wordsMatcher       *WordsMatcher
	releaseMatcher     *ReleaseGroupsMatcher
	customizationMatch *CustomizationMatcher
	streamingPlatforms *StreamingPlatforms
}

// NewMetaService 创建新的MetaService实例
func NewMetaService(deps MetaParserDeps) MetaService {
	return &metaService{
		wordsMatcher:       deps.WordsMatcher,
		releaseMatcher:     deps.ReleaseMatcher,
		customizationMatch: deps.CustomizationMatch,
		streamingPlatforms: deps.StreamingPlatforms,
	}
}

// MetaInfo 解析标题/种子名/文件名，返回MetaBase子类
func (ms *metaService) MetaInfo(ctx context.Context, title, subtitle string, isFile bool, customWords []string) (*MetaBase, error) {
	// 1. 记录原标题
	origTitle := title

	// 2. 预处理标题（WordsMatcher）
	cleanedTitle, applyWords := ms.wordsMatcher.Prepare(title, customWords)

	// 3. 内嵌元信息解析
	cleanedTitle, metainfo := ms.FindMetaInfo(cleanedTitle)

	// 4. 选择解析器：动漫 or 普通视频
	var metaObj *MetaBase
	if ms.IsAnime(cleanedTitle) {
		anime := NewMetaAnime(cleanedTitle, subtitle, isFile, MetaParserDeps{
			WordsMatcher:       ms.wordsMatcher,
			ReleaseMatcher:     ms.releaseMatcher,
			CustomizationMatch: ms.customizationMatch,
			StreamingPlatforms: ms.streamingPlatforms,
		})
		metaObj = anime.MetaBase
	} else {
		video := NewMetaVideo(cleanedTitle, subtitle, isFile, MetaParserDeps{
			WordsMatcher:       ms.wordsMatcher,
			ReleaseMatcher:     ms.releaseMatcher,
			CustomizationMatch: ms.customizationMatch,
			StreamingPlatforms: ms.streamingPlatforms,
		})
		metaObj = video.MetaBase
	}

	// 5. 补充字段（原标题、使用的识别词、tmdbid 等）
	metaObj.OrgString = origTitle
	metaObj.AppliedWords = applyWords

	if metainfo.TMDBID != 0 {
		metaObj.TMDBID = metainfo.TMDBID
	}
	if metainfo.DoubanID != "" {
		metaObj.DoubanID = metainfo.DoubanID
	}
	if metainfo.Type != "" {
		metaObj.Type = metainfo.Type
	}
	if metainfo.BeginSeason != nil {
		metaObj.BeginSeason = *metainfo.BeginSeason
	}
	if metainfo.EndSeason != nil {
		metaObj.EndSeason = *metainfo.EndSeason
	}
	if metainfo.TotalSeason != nil {
		metaObj.TotalSeason = *metainfo.TotalSeason
	}
	if metainfo.BeginEpisode != nil {
		metaObj.BeginEpisode = *metainfo.BeginEpisode
	}
	if metainfo.EndEpisode != nil {
		metaObj.EndEpisode = *metainfo.EndEpisode
	}
	if metainfo.TotalEpisode != nil {
		metaObj.TotalEpisode = *metainfo.TotalEpisode
	}

	return metaObj, nil
}

// MetaInfoPath 根据完整路径推导元信息
func (ms *metaService) MetaInfoPath(ctx context.Context, path string) (*MetaBase, error) {
	// 简化实现，后续可以扩展
	// 这里只取文件名作为标题
	fileName := path
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		fileName = parts[len(parts)-1]
	}

	return ms.MetaInfo(ctx, fileName, "", true, nil)
}

// IsAnime 判断一个名称是否更像“动漫（番剧）”而不是普通电影/剧集
func (ms *metaService) IsAnime(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}

	// 正则规则
	var (
		reAnimeStyle1 = regexp.MustCompile(`【[+0-9XVPI-]+】`)
		reAnimeStyle2 = regexp.MustCompile(`\s+-\s+[\dv]{1,4}\s+`)
		reTvPattern   = regexp.MustCompile(`S\d{2}\s*-\s*S\d{2}|S\d{2}|\s+S\d{1,2}|EP?\d{2,4}\s*-\s*EP?\d{2,4}|EP?\d{2,4}|\s+EP?\d{1,4}`)
		reAnimeStyle3 = regexp.MustCompile(`\[[+0-9XVPI-]+\]`)
		reAnimeStyle4 = regexp.MustCompile(`【.*?】.*?【`)
		reAnimeStyle5 = regexp.MustCompile(`\[.*?\].*?\[`)
	)

	if reAnimeStyle1.MatchString(name) || reAnimeStyle4.MatchString(name) {
		return true
	}
	if reAnimeStyle2.MatchString(name) {
		return true
	}
	if reTvPattern.MatchString(name) {
		return false
	}
	if reAnimeStyle3.MatchString(name) || reAnimeStyle5.MatchString(name) {
		return true
	}
	return false
}

// FindMetaInfo 从标题中剥离内嵌的“元信息标签”
func (ms *metaService) FindMetaInfo(title string) (string, *MetaInfoResult) {
	result := &MetaInfoResult{}
	if title == "" {
		return title, result
	}

	// 处理 {[ ... ]} 格式
	reCustomTag := regexp.MustCompile(`\{\[[\W\w]+?\]\}`)
	matches := reCustomTag.FindAllStringSubmatch(title, -1)
	for _, match := range matches {
		if len(match) > 0 {
			tag := match[0]
			// 从标题中删除标签
			title = strings.ReplaceAll(title, tag, "")

			// 解析标签内容
			tagContent := strings.Trim(tag, "{[]}")
			ms.parseMetaTag(tagContent, result)
		}
	}

	// 处理 Emby 风格标签
	reEmbyTag := regexp.MustCompile(`\[(tmdbid|tmdb)[=-](\d+)\]|\{(tmdbid|tmdb)[=-](\d+)\}`)
	matches = reEmbyTag.FindAllStringSubmatch(title, -1)
	for _, match := range matches {
		if len(match) > 2 {
			title = strings.ReplaceAll(title, match[0], "")
			// 提取 tmdbid
			if len(match) > 4 && match[4] != "" {
				tmdbid, _ := strconv.ParseInt(match[4], 10, 64)
				result.TMDBID = tmdbid
			} else if match[2] != "" {
				tmdbid, _ := strconv.ParseInt(match[2], 10, 64)
				result.TMDBID = tmdbid
			}
		}
	}

	// 计算 total_season / total_episode
	ms.calculateTotal(result)

	return strings.TrimSpace(title), result
}

// parseMetaTag 解析元信息标签内容
func (ms *metaService) parseMetaTag(tagContent string, result *MetaInfoResult) {
	parts := strings.Split(tagContent, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 解析 key=value
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch strings.ToLower(key) {
		case "tmdbid":
			tmdbid, _ := strconv.ParseInt(value, 10, 64)
			result.TMDBID = tmdbid
		case "doubanid":
			result.DoubanID = value
		case "type":
			switch strings.ToLower(value) {
			case "movies", "movie":
				result.Type = MediaTypeMovie
			case "tv", "series":
				result.Type = MediaTypeTV
			case "anime":
				result.Type = MediaTypeAnime
			}
		case "s":
			// 解析季范围，如 "1", "1-2"
			ms.parseRange(value, &result.BeginSeason, &result.EndSeason)
		case "e":
			// 解析集范围，如 "1", "1-10"
			ms.parseRange(value, &result.BeginEpisode, &result.EndEpisode)
		}
	}
}

// parseRange 解析范围字符串，如 "1", "1-2"
func (ms *metaService) parseRange(rangeStr string, begin, end **int) {
	if rangeStr == "" {
		return
	}

	parts := strings.Split(rangeStr, "-")
	if len(parts) == 1 {
		// 单个值
		val, err := strconv.Atoi(parts[0])
		if err == nil {
			*begin = &val
			*end = &val
		}
	} else if len(parts) == 2 {
		// 范围值
		beginVal, err1 := strconv.Atoi(parts[0])
		endVal, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			// 确保 begin <= end
			if beginVal > endVal {
				beginVal, endVal = endVal, beginVal
			}
			*begin = &beginVal
			*end = &endVal
		}
	}
}

// calculateTotal 计算总季数和总集数
func (ms *metaService) calculateTotal(result *MetaInfoResult) {
	// 计算总季数
	if result.BeginSeason != nil && result.EndSeason != nil {
		total := *result.EndSeason - *result.BeginSeason + 1
		result.TotalSeason = &total
	} else if result.BeginSeason != nil {
		one := 1
		result.TotalSeason = &one
	}

	// 计算总集数
	if result.BeginEpisode != nil && result.EndEpisode != nil {
		total := *result.EndEpisode - *result.BeginEpisode + 1
		result.TotalEpisode = &total
	} else if result.BeginEpisode != nil {
		one := 1
		result.TotalEpisode = &one
	}
}
