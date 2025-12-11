package media

import (
	"regexp"
	"strconv"
	"strings"
)

// MetaVideo 视频元信息解析
type MetaVideo struct {
	*MetaBase
	// 解析过程内部状态可以用未导出字段保存
}

// NewMetaVideo 创建新的MetaVideo实例
func NewMetaVideo(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaVideo {
	meta := &MetaVideo{
		MetaBase: NewMetaBase(title, title, subtitle),
	}

	// 初始化副标题
	meta.InitSubtitle(subtitle)

	// 设置媒体类型
	meta.Type = MediaTypeTV

	// 解析标题
	meta.parseTitle(title)

	return meta
}

// parseTitle 解析视频标题
func (mv *MetaVideo) parseTitle(title string) {
	if title == "" {
		return
	}

	// 解析年份
	mv.parseYear(title)

	// 解析季集
	mv.parseSeasonEpisode(title)

	// 解析分辨率
	mv.parseResolution(title)

	// 解析编码
	mv.parseEncode(title)

	// 解析资源类型
	mv.parseResourceType(title)
}

// parseYear 解析年份
func (mv *MetaVideo) parseYear(title string) {
	// 匹配格式：2023, (2023), [2023]
	// 使用正则表达式匹配4位数字的年份，范围在1900-2100之间
	re := regexp.MustCompile(`(?i)(?:^|\D)(19[0-9]{2}|20[0-9]{2}|2100)(?:$|\D)`)
	matches := re.FindStringSubmatch(title)
	if len(matches) > 1 {
		year, err := strconv.Atoi(matches[1])
		if err == nil {
			mv.Year = year
		}
	}
}

// parseSeasonEpisode 解析季集
func (mv *MetaVideo) parseSeasonEpisode(title string) {
	// 匹配格式：S01E01, S01-E01, S01.E01, 1x01, 第1季第1集等

	// 1. 匹配 SxxExx 格式（S01E01, S01-E01, S01.E01）
	reSxxExx := regexp.MustCompile(`(?i)(?:^|\D)(S(\d{1,3}))[\s\-\.]?(E(\d{1,3})(?:[\s\-\.](E)?(\d{1,3}))?)(?:$|\D)`)
	matches := reSxxExx.FindStringSubmatch(title)
	if len(matches) > 4 {
		season, _ := strconv.Atoi(matches[2])
		episode, _ := strconv.Atoi(matches[4])
		endEpisode := episode
		if matches[6] != "" {
			endEpisode, _ = strconv.Atoi(matches[6])
		}
		mv.SetSeason(season)
		mv.SetEpisodes(episode, endEpisode)
		return
	}

	// 2. 匹配 xxExx 格式（01E01, 01-E01）
	rexxExx := regexp.MustCompile(`(?i)(?:^|\D)(\d{1,3})[\s\-\.]?E(\d{1,3})(?:[\s\-\.]E(\d{1,3}))?(?:$|\D)`)
	matches = rexxExx.FindStringSubmatch(title)
	if len(matches) > 2 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		endEpisode := episode
		if matches[3] != "" {
			endEpisode, _ = strconv.Atoi(matches[3])
		}
		mv.SetSeason(season)
		mv.SetEpisodes(episode, endEpisode)
		return
	}

	// 3. 匹配 xxxx 格式（1x01, 1x01-10）
	rexxxx := regexp.MustCompile(`(?i)(?:^|\D)(\d{1,3})x(\d{1,3})(?:[\s\-](\d{1,3}))?(?:$|\D)`)
	matches = rexxxx.FindStringSubmatch(title)
	if len(matches) > 2 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		endEpisode := episode
		if matches[3] != "" {
			endEpisode, _ = strconv.Atoi(matches[3])
		}
		mv.SetSeason(season)
		mv.SetEpisodes(episode, endEpisode)
		return
	}

	// 4. 匹配中文格式：第1季第1集, 第1季第1-10集
	reChinese := regexp.MustCompile(`(?i)(?:^|\D)第(\d{1,3})季[\s\-]*第(\d{1,3})(?:[\s\-]*[集话](?:[\s\-]*第)?(\d{1,3})[集话])?(?:$|\D)`)
	matches = reChinese.FindStringSubmatch(title)
	if len(matches) > 2 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		endEpisode := episode
		if matches[3] != "" {
			endEpisode, _ = strconv.Atoi(matches[3])
		}
		mv.SetSeason(season)
		mv.SetEpisodes(episode, endEpisode)
		return
	}

	// 5. 匹配只包含季的格式：S01, 第1季
	reOnlySeason := regexp.MustCompile(`(?i)(?:^|\D)(?:S(\d{1,3})|第(\d{1,3})季)(?:$|\D)`)
	matches = reOnlySeason.FindStringSubmatch(title)
	if len(matches) > 2 {
		season := 0
		if matches[1] != "" {
			season, _ = strconv.Atoi(matches[1])
		} else if matches[2] != "" {
			season, _ = strconv.Atoi(matches[2])
		}
		if season > 0 {
			mv.SetSeason(season)
		}
		return
	}

	// 6. 匹配只包含集的格式：E01, 第1集
	reOnlyEpisode := regexp.MustCompile(`(?i)(?:^|\D)(?:E(\d{1,3})|第(\d{1,3})[集话])(?:$|\D)`)
	matches = reOnlyEpisode.FindStringSubmatch(title)
	if len(matches) > 2 {
		episode := 0
		if matches[1] != "" {
			episode, _ = strconv.Atoi(matches[1])
		} else if matches[2] != "" {
			episode, _ = strconv.Atoi(matches[2])
		}
		if episode > 0 {
			mv.SetEpisode(episode)
		}
		return
	}
}

// parseResolution 解析分辨率
func (mv *MetaVideo) parseResolution(title string) {
	// 匹配格式：4K, 2160P, 1080P, 720P, 480P等

	// 优先级：4K > 2160P > 1080P > 720P > 480P
	resolutions := []struct {
		pattern string
		pix     ResourcePix
	}{
		{`(?i)(?:^|\D)(4K|2160[Pp])(?:$|\D)`, ResourcePix4K},
		{`(?i)(?:^|\D)(1080[Pp]|FHD)(?:$|\D)`, ResourcePix1080P},
		{`(?i)(?:^|\D)(720[Pp]|HD)(?:$|\D)`, ResourcePix720P},
		{`(?i)(?:^|\D)(480[Pp]|SD)(?:$|\D)`, ResourcePix480P},
	}

	for _, res := range resolutions {
		re := regexp.MustCompile(res.pattern)
		if re.MatchString(title) {
			mv.ResourcePix = res.pix
			return
		}
	}

	// 默认未知分辨率
	mv.ResourcePix = ResourcePixUnknown
}

// parseEncode 解析编码
func (mv *MetaVideo) parseEncode(title string) {
	// 解析视频编码和音频编码

	// 视频编码：x264, H.264, AVC, x265, H.265, HEVC, VP9, AV1等
	videoEncodes := []string{
		`(?i)(?:^|\D)(AV1)(?:$|\D)`,
		`(?i)(?:^|\D)(VP9)(?:$|\D)`,
		`(?i)(?:^|\D)(HEVC|H\.265|x265)(?:$|\D)`,
		`(?i)(?:^|\D)(H\.264|AVC|x264)(?:$|\D)`,
		`(?i)(?:^|\D)(MPEG-4|MP4)(?:$|\D)`,
		`(?i)(?:^|\D)(MPEG-2|MP2)(?:$|\D)`,
	}

	for _, pattern := range videoEncodes {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) > 1 {
			mv.VideoEncode = strings.ToUpper(matches[1])
			break
		}
	}

	// 音频编码：AAC, AC3, DTS, DTS-HD, TrueHD, FLAC等
	audioEncodes := []string{
		`(?i)(?:^|\D)(FLAC)(?:$|\D)`,
		`(?i)(?:^|\D)(TrueHD)(?:$|\D)`,
		`(?i)(?:^|\D)(DTS-HD)(?:$|\D)`,
		`(?i)(?:^|\D)(DTS)(?:$|\D)`,
		`(?i)(?:^|\D)(AC3|DD)(?:$|\D)`,
		`(?i)(?:^|\D)(AAC)(?:$|\D)`,
	}

	for _, pattern := range audioEncodes {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) > 1 {
			mv.AudioEncode = strings.ToUpper(matches[1])
			break
		}
	}
}

// parseResourceType 解析资源类型
func (mv *MetaVideo) parseResourceType(title string) {
	// 解析资源效果：BluRay, HDTV, WEB, DVD, SD等

	// 优先级：BluRay > HDTV > WEB > DVD > SD
	resourceEffects := []struct {
		pattern string
		effect  ResourceEffect
	}{
		{`(?i)(?:^|\D)(BluRay|BD|Blue-ray)(?:$|\D)`, ResourceEffectBluray},
		{`(?i)(?:^|\D)(HDTV)(?:$|\D)`, ResourceEffectHDTV},
		{`(?i)(?:^|\D)(WEB|WEB-DL|WEBRip|WebRip)(?:$|\D)`, ResourceEffectWEB},
		{`(?i)(?:^|\D)(DVD)(?:$|\D)`, ResourceEffectDVD},
		{`(?i)(?:^|\D)(SD)(?:$|\D)`, ResourceEffectSD},
	}

	for _, res := range resourceEffects {
		re := regexp.MustCompile(res.pattern)
		if re.MatchString(title) {
			mv.ResourceEffect = res.effect
			break
		}
	}

	// 默认未知资源效果
	if mv.ResourceEffect == "" {
		mv.ResourceEffect = ResourceEffectUnknown
	}

	// 解析资源类型：Movie, TV, Anime等
	// 这里简化处理，根据媒体类型设置资源类型
	switch mv.Type {
	case MediaTypeMovie:
		mv.ResourceType = ResourceTypeMovie
	case MediaTypeTV:
		mv.ResourceType = ResourceTypeTV
	case MediaTypeAnime:
		mv.ResourceType = ResourceTypeAnime
	default:
		mv.ResourceType = ResourceTypeOther
	}
}
