package media

import (
	"regexp"
	"strconv"
	"strings"
)

// MetaAnime 动漫元信息解析
type MetaAnime struct {
	*MetaBase
	// 解析过程内部状态可以用未导出字段保存
}

// NewMetaAnime 创建新的MetaAnime实例
func NewMetaAnime(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaAnime {
	meta := &MetaAnime{
		MetaBase: NewMetaBase(title, title, subtitle),
	}

	// 初始化副标题
	meta.InitSubtitle(subtitle)

	// 设置媒体类型
	meta.Type = MediaTypeAnime

	// 解析标题
	meta.parseTitle(title)

	return meta
}

// parseTitle 解析动漫标题
func (ma *MetaAnime) parseTitle(title string) {
	if title == "" {
		return
	}

	// 解析年份
	ma.parseYear(title)

	// 解析季集
	ma.parseSeasonEpisode(title)

	// 解析分辨率
	ma.parseResolution(title)

	// 解析编码
	ma.parseEncode(title)

	// 解析资源类型
	ma.parseResourceType(title)
}

// parseYear 解析年份
func (ma *MetaAnime) parseYear(title string) {
	// 匹配格式：2023, (2023), [2023], 【2023】
	// 使用正则表达式匹配4位数字的年份，范围在1900-2100之间
	re := regexp.MustCompile(`(?i)(?:^|\D|\[|\(|（)(19[0-9]{2}|20[0-9]{2}|2100)(?:$|\D|\]|\)|）)`)
	matches := re.FindStringSubmatch(title)
	if len(matches) > 1 {
		year, err := strconv.Atoi(matches[1])
		if err == nil {
			ma.Year = year
		}
	}
}

// parseSeasonEpisode 解析季集
func (ma *MetaAnime) parseSeasonEpisode(title string) {
	// 匹配动漫特有的季集格式：- 01, - 01-10, 第1季第1集等

	// 1. 匹配 "动漫名称 - 01" 或 "动漫名称 - 01-10" 格式
	reDash := regexp.MustCompile(`(?i)(?:^|\D)(?:-\s*)(\d{1,3})(?:\s*-\s*(\d{1,3}))?(?:$|\D|\[|\]|【|】)`)
	matches := reDash.FindStringSubmatch(title)
	if len(matches) > 1 {
		episode, _ := strconv.Atoi(matches[1])
		endEpisode := episode
		if matches[2] != "" {
			endEpisode, _ = strconv.Atoi(matches[2])
		}
		ma.SetEpisode(episode) // 动漫默认第1季
		ma.SetEpisodes(episode, endEpisode)
		return
	}

	// 2. 匹配中文格式：第1季第1集, 第1季第1-10集
	reChinese := regexp.MustCompile(`(?i)(?:^|\D)第(\d{1,3})季[\s\-]*第(\d{1,3})(?:[\s\-]*[集话](?:[\s\-]*第)?(\d{1,3})[集话])?(?:$|\D)`)
	matches = reChinese.FindStringSubmatch(title)
	if len(matches) > 2 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		endEpisode := episode
		if matches[3] != "" {
			endEpisode, _ = strconv.Atoi(matches[3])
		}
		ma.SetSeason(season)
		ma.SetEpisodes(episode, endEpisode)
		return
	}

	// 3. 匹配只包含集的格式：第1集, 第1-10集
	reOnlyEpisode := regexp.MustCompile(`(?i)(?:^|\D)第(\d{1,3})[集话](?:[\s\-]*第)?(\d{1,3})?[集话]?(?:$|\D)`)
	matches = reOnlyEpisode.FindStringSubmatch(title)
	if len(matches) > 1 {
		episode, _ := strconv.Atoi(matches[1])
		endEpisode := episode
		if matches[2] != "" {
			endEpisode, _ = strconv.Atoi(matches[2])
		}
		ma.SetEpisode(episode) // 动漫默认第1季
		ma.SetEpisodes(episode, endEpisode)
		return
	}
}

// parseResolution 解析分辨率
func (ma *MetaAnime) parseResolution(title string) {
	// 匹配格式：4K, 2160P, 1080P, 720P, 480P等

	// 优先级：4K > 2160P > 1080P > 720P > 480P
	resolutions := []struct {
		pattern string
		pix     ResourcePix
	}{
		{`(?i)(?:^|\D|\[|\]|【|】)(4K|2160[Pp])(?:$|\D|\[|\]|【|】)`, ResourcePix4K},
		{`(?i)(?:^|\D|\[|\]|【|】)(1080[Pp]|FHD)(?:$|\D|\[|\]|【|】)`, ResourcePix1080P},
		{`(?i)(?:^|\D|\[|\]|【|】)(720[Pp]|HD)(?:$|\D|\[|\]|【|】)`, ResourcePix720P},
		{`(?i)(?:^|\D|\[|\]|【|】)(480[Pp]|SD)(?:$|\D|\[|\]|【|】)`, ResourcePix480P},
	}

	for _, res := range resolutions {
		re := regexp.MustCompile(res.pattern)
		if re.MatchString(title) {
			ma.ResourcePix = res.pix
			return
		}
	}

	// 默认未知分辨率
	ma.ResourcePix = ResourcePixUnknown
}

// parseEncode 解析编码
func (ma *MetaAnime) parseEncode(title string) {
	// 解析视频编码和音频编码

	// 视频编码：x264, H.264, AVC, x265, H.265, HEVC, VP9, AV1等
	videoEncodes := []string{
		`(?i)(?:^|\D|\[|\]|【|】)(AV1)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(VP9)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(HEVC|H\.265|x265)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(H\.264|AVC|x264)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(MPEG-4|MP4)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(MPEG-2|MP2)(?:$|\D|\[|\]|【|】)`,
	}

	for _, pattern := range videoEncodes {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) > 1 {
			ma.VideoEncode = strings.ToUpper(matches[1])
			break
		}
	}

	// 音频编码：AAC, AC3, DTS, DTS-HD, TrueHD, FLAC等
	audioEncodes := []string{
		`(?i)(?:^|\D|\[|\]|【|】)(FLAC)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(TrueHD)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(DTS-HD)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(DTS)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(AC3|DD)(?:$|\D|\[|\]|【|】)`,
		`(?i)(?:^|\D|\[|\]|【|】)(AAC)(?:$|\D|\[|\]|【|】)`,
	}

	for _, pattern := range audioEncodes {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) > 1 {
			ma.AudioEncode = strings.ToUpper(matches[1])
			break
		}
	}
}

// parseResourceType 解析资源类型
func (ma *MetaAnime) parseResourceType(title string) {
	// 解析资源效果：BluRay, HDTV, WEB, DVD, SD等

	// 优先级：BluRay > HDTV > WEB > DVD > SD
	resourceEffects := []struct {
		pattern string
		effect  ResourceEffect
	}{
		{`(?i)(?:^|\D|\[|\]|【|】)(BluRay|BD|Blue-ray)(?:$|\D|\[|\]|【|】)`, ResourceEffectBluray},
		{`(?i)(?:^|\D|\[|\]|【|】)(HDTV)(?:$|\D|\[|\]|【|】)`, ResourceEffectHDTV},
		{`(?i)(?:^|\D|\[|\]|【|】)(WEB|WEB-DL|WEBRip|WebRip)(?:$|\D|\[|\]|【|】)`, ResourceEffectWEB},
		{`(?i)(?:^|\D|\[|\]|【|】)(DVD)(?:$|\D|\[|\]|【|】)`, ResourceEffectDVD},
		{`(?i)(?:^|\D|\[|\]|【|】)(SD)(?:$|\D|\[|\]|【|】)`, ResourceEffectSD},
	}

	for _, res := range resourceEffects {
		re := regexp.MustCompile(res.pattern)
		if re.MatchString(title) {
			ma.ResourceEffect = res.effect
			break
		}
	}

	// 默认未知资源效果
	if ma.ResourceEffect == "" {
		ma.ResourceEffect = ResourceEffectUnknown
	}

	// 动漫资源类型
	ma.ResourceType = ResourceTypeAnime
}
