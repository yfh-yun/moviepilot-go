package media

import (
	"regexp"
	"strconv"
	"strings"
)

// MetaVideo 视频元信息解析
type MetaVideo struct {
	*MetaBase
	// 控制标位区
	stopNameFlag   bool
	stopCnNameFlag bool
	lastToken      string
	lastTokenType  string
	continueFlag   bool
	unknownNameStr string
	source         string
	effect         []string
	index          int
	// 正则式区（使用编译后的正则表达式以提高性能）
	seasonRe        *regexp.Regexp
	episodeRe       *regexp.Regexp
	partRe          *regexp.Regexp
	romanNumerals   *regexp.Regexp
	sourceRe        *regexp.Regexp
	effectRe        *regexp.Regexp
	resourcesTypeRe *regexp.Regexp
	nameNoBeginRe   *regexp.Regexp
	nameNoChineseRe *regexp.Regexp
	nameNostringRe  *regexp.Regexp
	resourcesPixRe  *regexp.Regexp
	resourcesPixRe2 *regexp.Regexp
	videoEncodeRe   *regexp.Regexp
	audioEncodeRe   *regexp.Regexp
}

// init 初始化正则表达式
func (mv *MetaVideo) initRegex() {
	mv.seasonRe = regexp.MustCompile(`S(\d{3})|^S(\d{1,3})$|S(\d{1,3})E`)
	mv.episodeRe = regexp.MustCompile(`EP?(\d{2,4})$|^EP?(\d{1,4})$|^S\d{1,2}EP?(\d{1,4})$|S\d{2}EP?(\d{2,4})`)
	mv.partRe = regexp.MustCompile(`(^PART[0-9ABI]{0,2}$|^CD[0-9]{0,2}$|^DVD[0-9]{0,2}$|^DISK[0-9]{0,2}$|^DISC[0-9]{0,2}$)`)
	mv.romanNumerals = regexp.MustCompile(`^M*(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`) // Go不支持前瞻断言，去掉(?=[MDCLXVI])
	mv.sourceRe = regexp.MustCompile(`^BLURAY$|^HDTV$|^UHDTV$|^HDDVD$|^WEBRIP$|^DVDRIP$|^BDRIP$|^BLU$|^WEB$|^BD$|^HDRip$|^REMUX$|^UHD$`)
	mv.effectRe = regexp.MustCompile(`^SDR$|^HDR\d*$|^DOLBY$|^DOVI$|^DV$|^3D$|^REPACK$|^HLG$|^HDR10(\+|Plus)$|^EDR$|^HQ$`)
	mv.resourcesTypeRe = regexp.MustCompile(`(^BLURAY$|^HDTV$|^UHDTV$|^HDDVD$|^WEBRIP$|^DVDRIP$|^BDRIP$|^BLU$|^WEB$|^BD$|^HDRip$|^REMUX$|^UHD$)|(^SDR$|^HDR\d*$|^DOLBY$|^DOVI$|^DV$|^3D$|^REPACK$|^HLG$|^HDR10(\+|Plus)$|^EDR$|^HQ$)`)
	mv.nameNoBeginRe = regexp.MustCompile(`^[\[【].+?[\]】]`)
	mv.nameNoChineseRe = regexp.MustCompile(`.*版|.*字幕`)
	mv.nameNostringRe = regexp.MustCompile(`^PTS|^JADE|^AOD|^CHC|^[A-Z]{1,4}TV[\-0-9UVHDK]*` +
		`|HBO$|\s+HBO|\d{1,2}th|\d{1,2}bit|NETFLIX|AMAZON|IMAX|^3D|\s+3D|^BBC\s+|\s+BBC|BBC$|DISNEY\+?|XXX|\s+DC$` +
		`|[第\s共]+[0-9一二三四五六七八九十\-\s]+季` +
		`|[第\s共]+[0-9一二三四五六七八九十百零\-\s]+[集话話]` +
		`|连载|日剧|美剧|电视剧|动画片|动漫|欧美|西德|日韩|超高清|高清|无水印|下载|蓝光|翡翠台|梦幻天堂·龙网|★?\d*月?新番` +
		`|最终季|合集|[多中国英葡法俄日韩德意西印泰台港粤双文语简繁体特效内封官译外挂]+字幕|版本|出品|台版|港版|\w+字幕组|\w+字幕社` +
		`|未删减版|UNCUT$|UNRATE$|WITH EXTRAS$|RERIP$|SUBBED$|PROPER$|REPACK$|SEASON$|EPISODE$|Complete$|Extended$|Extended Version$` +
		`|S\d{2}\s*\-\s*S\d{2}|S\d{2}|\s+S\d{1,2}|EP?\d{2,4}\s*\-\s*EP?\d{2,4}|EP?\d{2,4}|\s+EP?\d{1,4}` +
		`|CD[\s.]*[1-9]|DVD[\s.]*[1-9]|DISK[\s.]*[1-9]|DISC[\s.]*[1-9]` +
		`|[248]K|\d{3,4}[PIX]+` +
		`|CD[\s.]*[1-9]|DVD[\s.]*[1-9]|DISK[\s.]*[1-9]|DISC[\s.]*[1-9]|\s+GB`)
	mv.resourcesPixRe = regexp.MustCompile(`^[SBUHD]*(\d{3,4}[PI]+)|\d{3,4}X(\d{3,4})`)
	mv.resourcesPixRe2 = regexp.MustCompile(`(^[248]+K)`)
	mv.videoEncodeRe = regexp.MustCompile(`^(H26[45])$|^(x26[45])$|^AVC$|^HEVC$|^VC\d?$|^MPEG\d?$|^Xvid$|^DivX$|^AV1$|^HDR\d*$|^AVS(\+|[23])$`)
	mv.audioEncodeRe = regexp.MustCompile(`^DTS\d?$|^DTSHD$|^DTSHDMA$|^Atmos$|^TrueHD\d?$|^AC3$|^\dAudios?$|^DDP\d?$|^DD\+\d?$|^DD\d?$|^LPCM\d?$|^AAC\d?$|^FLAC\d?$|^HD\d?$|^MA\d?$|^HR\d?$|^Opus\d?$|^Vorbis\d?$|^AV[3S]A$`)
}

// NewMetaVideo 创建新的MetaVideo实例
func NewMetaVideo(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaVideo {
	meta := &MetaVideo{
		MetaBase:     NewMetaBase(title, title, subtitle, isFile),
		continueFlag: true,
		effect:       []string{},
	}

	// 初始化正则表达式
	meta.initRegex()

	if title == "" {
		return meta
	}

	originalTitle := title
	meta.source = ""
	meta.effect = []string{}
	meta.index = 0

	// 判断是否纯数字命名
	if isFile && title != "" && isAllDigit(title) && len(title) < 5 {
		ep, _ := strconv.Atoi(title)
		meta.BeginEpisode = &ep
		meta.Type = MediaTypeTV
		return meta
	}

	// 全名为Season xx 及 Sxx 直接返回
	seasonFullRe := regexp.MustCompile(`^Season\s+(\d{1,3})$|^S(\d{1,3})$`)
	seasonFullRes := seasonFullRe.FindStringSubmatch(title)
	if seasonFullRes != nil {
		meta.Type = MediaTypeTV
		seasonStr := ""
		if seasonFullRes[1] != "" {
			seasonStr = seasonFullRes[1]
		} else if seasonFullRes[2] != "" {
			seasonStr = seasonFullRes[2]
		}
		if seasonStr != "" {
			season, _ := strconv.Atoi(seasonStr)
			meta.BeginSeason = &season
			meta.TotalSeason = 1
		}
		return meta
	}

	// 去掉名称中第1个[]的内容
	title = meta.nameNoBeginRe.ReplaceAllString(title, "")

	// 把xxxx-xxxx年份换成前一个年份，常出现在季集上
	title = regexp.MustCompile(`([\s.]+)(\d{4})-(\d{4})`).ReplaceAllString(title, "$1$2")

	// 把大小去掉
	title = regexp.MustCompile(`[0-9.]+\s*[MGT]i?B`).ReplaceAllString(title, "")

	// 把年月日去掉
	title = regexp.MustCompile(`\d{4}[\s._-]\d{1,2}[\s._-]\d{1,2}`).ReplaceAllString(title, "")

	// 拆分tokens（简化实现，使用空格分割）
	tokens := strings.Fields(title)

	// 解析名称、年份、季、集、资源类型、分辨率等
	for i, token := range tokens {
		meta.index = i + 1 // 更新当前处理的token索引
		// Part
		meta.initPart(token)
		// 标题
		if meta.continueFlag {
			meta.initName(token)
		}
		// 年份
		if meta.continueFlag {
			meta.initYear(token)
		}
		// 分辨率
		if meta.continueFlag {
			meta.initResourcePix(token)
		}
		// 季
		if meta.continueFlag {
			meta.initSeason(token)
		}
		// 集
		if meta.continueFlag {
			meta.initEpisode(token)
		}
		// 资源类型
		if meta.continueFlag {
			meta.initResourceType(token)
		}
		// 流媒体平台（简化实现，暂不实现）
		if meta.continueFlag {
			meta.initWebSource(token)
		}
		// 视频编码
		if meta.continueFlag {
			meta.initVideoEncode(token)
		}
		// 音频编码
		if meta.continueFlag {
			meta.initAudioEncode(token)
		}
		// 重置continueFlag
		meta.continueFlag = true
	}

	// 合成质量
	if len(meta.effect) > 0 {
		// 反转效果数组
		for i, j := 0, len(meta.effect)-1; i < j; i, j = i+1, j-1 {
			meta.effect[i], meta.effect[j] = meta.effect[j], meta.effect[i]
		}
		meta.ResourceEffect = ResourceEffect(strings.Join(meta.effect, " "))
	}
	if meta.source != "" {
		meta.ResourceType = ResourceType(meta.source)
		// 提取原盘DIY
		if strings.Contains(string(meta.ResourceType), "BluRay") {
			diyRe := regexp.MustCompile(`D[Ii]Y`)
			if (meta.Subtitle != "" && diyRe.MatchString(meta.Subtitle)) ||
				regexp.MustCompile(`-D[Ii]Y@`).MatchString(originalTitle) {
				meta.ResourceType = ResourceType(string(meta.ResourceType) + " DIY")
			}
		}
	}

	// 解析副标题，只要季和集
	meta.InitSubtitle(meta.OrgString)
	if meta.Subtitle != "" {
		meta.InitSubtitle(meta.Subtitle)
	}

	// 去掉名字中不需要的干扰字符，过短的纯数字不要
	meta.CnName = meta.fixName(meta.CnName)
	meta.EnName = strings.Title(meta.fixName(meta.EnName))

	// 处理part
	if meta.Part != "" && strings.ToUpper(meta.Part) == "PART" {
		meta.Part = ""
	}

	return meta
}

// isAllDigit 判断字符串是否全为数字
func isAllDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// initName 识别名称
func (mv *MetaVideo) initName(token string) {
	if token == "" {
		return
	}

	// 回收标题
	if mv.unknownNameStr != "" {
		if mv.CnName == "" {
			if mv.EnName == "" {
				mv.EnName = mv.unknownNameStr
			} else if mv.unknownNameStr != mv.Year {
				mv.EnName = mv.EnName + " " + mv.unknownNameStr
			}
			mv.lastTokenType = "enname"
		}
		mv.unknownNameStr = ""
	}

	if mv.stopNameFlag {
		return
	}

	if strings.ToUpper(token) == "AKA" {
		mv.continueFlag = false
		mv.stopNameFlag = true
		return
	}

	nameSeWords := []string{"共", "第", "季", "集", "话", "話", "期"}
	for _, word := range nameSeWords {
		if token == word {
			mv.lastTokenType = "name_se_words"
			return
		}
	}

	if isChinese(token) {
		// 含有中文，直接做为标题（连着的数字或者英文会保留），且不再取用后面出现的中文
		mv.lastTokenType = "cnname"
		if mv.CnName == "" {
			mv.CnName = token
		} else if !mv.stopCnNameFlag {
			nameMovieWords := []string{"剧场版", "劇場版", "电影版", "電影版"}
			containsMovieWord := false
			for _, word := range nameMovieWords {
				if strings.Contains(token, word) {
					containsMovieWord = true
					break
				}
			}

			if containsMovieWord ||
				(!mv.nameNoChineseRe.MatchString(token) &&
					!strings.ContainsAny(token, strings.Join(nameSeWords, ""))) {
				mv.CnName = mv.CnName + " " + token
			}
			mv.stopCnNameFlag = true
		}
	} else {
		isRomanDigit := mv.romanNumerals.MatchString(token)
		if isAllDigit(token) || isRomanDigit {
			// 第季集后面的不要
			if mv.lastTokenType == "name_se_words" {
				return
			}

			if mv.Name() != "" {
				// 名字后面以 0 开头的不要，极有可能是集
				if strings.HasPrefix(token, "0") {
					return
				}

				// 中文名后面跟的数字不是年份的极有可能是集
				if !isRomanDigit && mv.lastTokenType == "cnname" {
					num, err := strconv.Atoi(token)
					if err == nil && num < 1900 {
						return
					}
				}

				if (isAllDigit(token) && len(token) < 4) || isRomanDigit {
					// 4位以下的数字或者罗马数字，拼装到已有标题中
					if mv.lastTokenType == "cnname" {
						mv.CnName = mv.CnName + " " + token
					} else if mv.lastTokenType == "enname" {
						mv.EnName = mv.EnName + " " + token
					}
					mv.continueFlag = false
				} else if isAllDigit(token) && len(token) == 4 {
					// 4位数字，可能是年份，也可能真的是标题的一部分，也有可能是集
					if mv.unknownNameStr == "" {
						mv.unknownNameStr = token
					}
				}
			} else {
				// 名字未出现前的第一个数字，记下来
				if mv.unknownNameStr == "" {
					mv.unknownNameStr = token
				}
			}
		} else if mv.seasonRe.MatchString(token) {
			// 季的处理
			if mv.EnName != "" && strings.HasSuffix(strings.ToUpper(mv.EnName), "SEASON") {
				// 如果匹配到季，英文名结尾为Season，说明Season属于标题，不应在后续作为干扰词去除
				mv.EnName += " "
			}
			mv.stopNameFlag = true
			return
		} else if mv.episodeRe.MatchString(token) ||
			mv.resourcesTypeRe.MatchString(token) ||
			mv.resourcesPixRe.MatchString(token) {
			// 集、来源、版本等不要
			mv.stopNameFlag = true
			return
		} else {
			// 英文或者英文+数字，拼装起来
			if mv.EnName != "" {
				mv.EnName = mv.EnName + " " + token
			} else {
				mv.EnName = token
			}
			mv.lastTokenType = "enname"
		}
	}
}

// isChinese 判断字符串是否包含中文
func isChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// initPart 识别Part
func (mv *MetaVideo) initPart(token string) {
	if mv.Name() == "" {
		return
	}

	if mv.Year == "" && mv.BeginSeason == nil && mv.BeginEpisode == nil &&
		mv.ResourcePix == "" && mv.ResourceType == "" {
		return
	}

	reRes := mv.partRe.FindStringSubmatch(token)
	if reRes != nil {
		if mv.Part == "" {
			mv.Part = reRes[1]
		}
		// 简化实现，不处理下一个token
		mv.lastTokenType = "part"
		mv.continueFlag = false
	}
}

// initYear 识别年份
func (mv *MetaVideo) initYear(token string) {
	if mv.Name() == "" {
		return
	}

	if !isAllDigit(token) {
		return
	}

	if len(token) != 4 {
		return
	}

	year, err := strconv.Atoi(token)
	if err != nil || year < 1900 || year > 2050 {
		return
	}

	if mv.Year != "" {
		if mv.EnName != "" {
			mv.EnName = mv.EnName + " " + mv.Year
		} else if mv.CnName != "" {
			mv.CnName = mv.CnName + " " + mv.Year
		}
	} else if mv.EnName != "" && strings.HasSuffix(strings.ToUpper(mv.EnName), "SEASON") {
		// 如果匹配到年，且英文名结尾为Season，说明Season属于标题，不应在后续作为干扰词去除
		mv.EnName += " "
	}

	mv.Year = token
	mv.lastTokenType = "year"
	mv.continueFlag = false
	mv.stopNameFlag = true
}

// initResourcePix 识别分辨率
func (mv *MetaVideo) initResourcePix(token string) {
	if mv.Name() == "" {
		return
	}

	// 将token转换为大写，以便不区分大小写匹配
	tokenUpper := strings.ToUpper(token)
	reRes := mv.resourcesPixRe.FindAllStringSubmatch(tokenUpper, -1)
	if reRes != nil {
		mv.lastTokenType = "pix"
		mv.continueFlag = false
		mv.stopNameFlag = true
		var resourcePix string

		for _, pixs := range reRes {
			for i, pix_i := range pixs {
				if i > 0 && pix_i != "" {
					resourcePix = pix_i
					break
				}
			}
			if resourcePix != "" && mv.ResourcePix == "" {
				mv.ResourcePix = ResourcePix(strings.ToLower(resourcePix))
				break
			}
		}

		if mv.ResourcePix != "" && isAllDigit(string(mv.ResourcePix)) {
			mv.ResourcePix = ResourcePix(string(mv.ResourcePix) + "p")
		}
	} else {
		reRes := mv.resourcesPixRe2.FindStringSubmatch(tokenUpper)
		if reRes != nil {
			mv.lastTokenType = "pix"
			mv.continueFlag = false
			mv.stopNameFlag = true
			if mv.ResourcePix == "" && len(reRes) > 1 {
				mv.ResourcePix = ResourcePix(strings.ToLower(reRes[1]))
			}
		} else {
			// 特殊处理常见分辨率格式，如 "1080p", "720p", "4K" 等
			commonResolutions := map[string]ResourcePix{
				"1080P": ResourcePix1080P,
				"720P":  ResourcePix720P,
				"480P":  ResourcePix480P,
				"4K":    ResourcePix4K,
			}
			for res, pix := range commonResolutions {
				if tokenUpper == res {
					mv.ResourcePix = pix
					mv.lastTokenType = "pix"
					mv.continueFlag = false
					mv.stopNameFlag = true
					break
				}
			}
		}
	}
}

// initSeason 识别季
func (mv *MetaVideo) initSeason(token string) {
	reRes := mv.seasonRe.FindAllStringSubmatch(token, -1)
	if reRes != nil {
		mv.lastTokenType = "season"
		mv.Type = MediaTypeTV
		mv.stopNameFlag = true
		mv.continueFlag = true

		for _, se := range reRes {
			var se_t string
			for i, se_i := range se {
				if i > 0 && se_i != "" && isAllDigit(se_i) {
					se_t = se_i
					break
				}
			}

			if se_t != "" {
				season, _ := strconv.Atoi(se_t)
				if mv.BeginSeason == nil {
					mv.BeginSeason = &season
					mv.TotalSeason = 1
				} else {
					if season > *mv.BeginSeason {
						mv.EndSeason = &season
						mv.TotalSeason = *mv.EndSeason - *mv.BeginSeason + 1
						// 简化实现，不处理isfile情况
					}
				}
			}
		}
	} else if isAllDigit(token) {
		if mv.lastTokenType == "SEASON" && mv.BeginSeason == nil && len(token) < 3 {
			season, _ := strconv.Atoi(token)
			mv.BeginSeason = &season
			mv.TotalSeason = 1
			mv.lastTokenType = "season"
			mv.stopNameFlag = true
			mv.continueFlag = false
			mv.Type = MediaTypeTV
		} else if mv.Type == MediaTypeTV && mv.BeginSeason == nil {
			season := 1
			mv.BeginSeason = &season
			mv.TotalSeason = 1
		}
	} else if strings.ToUpper(token) == "SEASON" && mv.BeginSeason == nil {
		mv.lastTokenType = "SEASON"
	} else if mv.Type == MediaTypeTV && mv.BeginSeason == nil {
		season := 1
		mv.BeginSeason = &season
		mv.TotalSeason = 1
	}
}

// initEpisode 识别集
func (mv *MetaVideo) initEpisode(token string) {
	reRes := mv.episodeRe.FindAllStringSubmatch(token, -1)
	if reRes != nil {
		mv.lastTokenType = "episode"
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.Type = MediaTypeTV

		for _, se := range reRes {
			var se_t string
			for i, se_i := range se {
				if i > 0 && se_i != "" && isAllDigit(se_i) {
					se_t = se_i
					break
				}
			}

			if se_t != "" {
				episode, _ := strconv.Atoi(se_t)
				if mv.BeginEpisode == nil {
					mv.BeginEpisode = &episode
					mv.TotalEpisode = 1
				} else {
					if episode > *mv.BeginEpisode {
						mv.EndEpisode = &episode
						mv.TotalEpisode = *mv.EndEpisode - *mv.BeginEpisode + 1
						// 简化实现，不处理isfile情况
					}
				}
			}
		}
	} else if isAllDigit(token) {
		if mv.BeginEpisode != nil && mv.EndEpisode == nil && len(token) < 5 {
			episode, err := strconv.Atoi(token)
			if err == nil && episode > *mv.BeginEpisode {
				mv.EndEpisode = &episode
				mv.TotalEpisode = *mv.EndEpisode - *mv.BeginEpisode + 1
				// 简化实现，不处理isfile情况
				mv.continueFlag = false
				mv.Type = MediaTypeTV
			}
		} else if mv.BeginEpisode == nil && len(token) > 1 && len(token) < 4 &&
			mv.lastTokenType != "year" && mv.lastTokenType != "videoencode" &&
			token != mv.unknownNameStr {
			episode, _ := strconv.Atoi(token)
			mv.BeginEpisode = &episode
			mv.TotalEpisode = 1
			mv.lastTokenType = "episode"
			mv.continueFlag = false
			mv.stopNameFlag = true
			mv.Type = MediaTypeTV
		} else if mv.lastTokenType == "EPISODE" && mv.BeginEpisode == nil && len(token) < 5 {
			episode, _ := strconv.Atoi(token)
			mv.BeginEpisode = &episode
			mv.TotalEpisode = 1
			mv.lastTokenType = "episode"
			mv.continueFlag = false
			mv.stopNameFlag = true
			mv.Type = MediaTypeTV
		}
	} else if strings.ToUpper(token) == "EPISODE" {
		mv.lastTokenType = "EPISODE"
	}
	// 只有集信息的电视剧，自动设置第1季
	if mv.Type == MediaTypeTV && mv.BeginEpisode != nil && mv.BeginSeason == nil {
		season := 1
		mv.BeginSeason = &season
		mv.TotalSeason = 1
	}
}

// initResourceType 识别资源类型
func (mv *MetaVideo) initResourceType(token string) {
	if mv.Name() == "" {
		return
	}

	// 将token转换为大写，以便不区分大小写匹配
	tokenUpper := strings.ToUpper(token)

	if tokenUpper == "DL" && mv.lastTokenType == "source" && mv.lastToken == "WEB" {
		mv.source = "WEB-DL"
		mv.continueFlag = false
		return
	} else if tokenUpper == "RAY" && mv.lastTokenType == "source" && mv.lastToken == "BLU" {
		// UHD BluRay组合
		if mv.source == "UHD" {
			mv.source = "UHD BluRay"
		} else {
			mv.source = "BluRay"
		}
		mv.continueFlag = false
		return
	} else if tokenUpper == "WEBDL" {
		mv.source = "WEB-DL"
		mv.continueFlag = false
		return
	}

	// UHD REMUX组合
	if tokenUpper == "REMUX" && mv.source == "BluRay" {
		mv.source = "BluRay REMUX"
		mv.continueFlag = false
		return
	} else if tokenUpper == "BLURAY" && mv.source == "UHD" {
		mv.source = "UHD BluRay"
		mv.continueFlag = false
		return
	}

	// 特殊处理常见资源类型，如 "BluRay", "HDTV", "WEB-DL" 等
	commonSourceTypes := map[string]string{
		"BLURAY": "BluRay",
		"HDTV":   "HDTV",
		"WEB":    "WEB",
		"WEB-DL": "WEB-DL",
		"WEBRIP": "WEBRip",
		"DVDRIP": "DVDRip",
		"BDRIP":  "BDRip",
		"DVD":    "DVD",
	}

	// 首先尝试直接匹配常见资源类型
	if source, ok := commonSourceTypes[tokenUpper]; ok {
		mv.lastTokenType = "source"
		mv.continueFlag = false
		mv.stopNameFlag = true
		if mv.source == "" {
			mv.source = source
			mv.lastToken = strings.ToUpper(mv.source)
		}
		return
	}

	// 然后尝试使用正则表达式匹配
	sourceRes := mv.sourceRe.FindStringSubmatch(tokenUpper)
	if sourceRes != nil {
		mv.lastTokenType = "source"
		mv.continueFlag = false
		mv.stopNameFlag = true
		if mv.source == "" {
			// 检查切片长度，确保安全访问
			if len(sourceRes) > 1 && sourceRes[1] != "" {
				mv.source = sourceRes[1]
				mv.lastToken = strings.ToUpper(mv.source)
			}
		}
		return
	}

	// 使用大写token匹配效果
	effectRes := mv.effectRe.FindStringSubmatch(tokenUpper)
	if effectRes != nil {
		mv.lastTokenType = "effect"
		mv.continueFlag = false
		mv.stopNameFlag = true
		// 检查切片长度，确保安全访问
		if len(effectRes) > 1 {
			effect := effectRes[1]
			// 检查effect是否已存在
			found := false
			for _, e := range mv.effect {
				if e == effect {
					found = true
					break
				}
			}
			if !found {
				mv.effect = append(mv.effect, effect)
			}
			mv.lastToken = strings.ToUpper(effect)
		}
	}
}

// initWebSource 识别流媒体平台（简化实现）
func (mv *MetaVideo) initWebSource(token string) {
	// 简化实现，暂不实现流媒体平台识别
}

// initVideoEncode 识别视频编码
func (mv *MetaVideo) initVideoEncode(token string) {
	if mv.Name() == "" {
		return
	}

	// 将token转换为大写，以便不区分大小写匹配
	tokenUpper := strings.ToUpper(token)

	// 特殊处理常见视频编码，如 "x264", "x265", "H264" 等
	commonVideoEncodes := map[string]string{
		"X264": "x264",
		"H264": "H264",
		"AVC":  "AVC",
		"X265": "x265",
		"H265": "H265",
		"HEVC": "HEVC",
		"AV1":  "AV1",
	}

	// 首先尝试直接匹配常见视频编码
	if encode, ok := commonVideoEncodes[tokenUpper]; ok {
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.lastTokenType = "videoencode"
		if mv.VideoEncode == "" {
			mv.VideoEncode = encode
			mv.lastToken = encode
		}
		return
	}

	if mv.Year == "" && mv.ResourcePix == "" && mv.ResourceType == "" &&
		mv.BeginSeason == nil && mv.BeginEpisode == nil {
		return
	}

	reRes := mv.videoEncodeRe.FindStringSubmatch(tokenUpper)
	if reRes != nil {
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.lastTokenType = "videoencode"
		if mv.VideoEncode == "" {
			var encode string
			// 检查切片长度，确保安全访问
			if len(reRes) > 1 {
				encode = strings.ToUpper(reRes[1])
			}
			mv.VideoEncode = encode
			mv.lastToken = encode
		} else if mv.VideoEncode == "10bit" {
			if len(reRes) > 1 {
				mv.VideoEncode = reRes[1] + " 10bit"
				mv.lastToken = strings.ToUpper(reRes[1])
			}
		}
	} else if tokenUpper == "H" || tokenUpper == "X" {
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.lastTokenType = "videoencode"
		if tokenUpper == "H" {
			mv.lastToken = "H"
		} else {
			mv.lastToken = "x"
		}
	} else if token == "264" || token == "265" {
		if mv.lastTokenType == "videoencode" && (mv.lastToken == "H" || mv.lastToken == "x") {
			mv.VideoEncode = mv.lastToken + token
		}
	} else if isAllDigit(token) {
		if mv.lastTokenType == "videoencode" && (strings.HasPrefix(mv.lastToken, "VC") || strings.HasPrefix(mv.lastToken, "MPEG")) {
			mv.VideoEncode = mv.lastToken + token
		}
	} else if tokenUpper == "10BIT" {
		mv.lastTokenType = "videoencode"
		if mv.VideoEncode == "" {
			mv.VideoEncode = "10bit"
		} else {
			mv.VideoEncode = mv.VideoEncode + " 10bit"
		}
	}
}

// initAudioEncode 识别音频编码
func (mv *MetaVideo) initAudioEncode(token string) {
	if mv.Name() == "" {
		return
	}

	// 将token转换为大写，以便不区分大小写匹配
	tokenUpper := strings.ToUpper(token)

	// 特殊处理常见音频编码，如 "DTS", "AC3", "AAC" 等
	commonAudioEncodes := map[string]string{
		"DTS":    "DTS",
		"AC3":    "AC3",
		"AAC":    "AAC",
		"FLAC":   "FLAC",
		"ATMOS":  "Atmos",
		"TRUEHD": "TrueHD",
		"DDP":    "DDP",
		"LPCM":   "LPCM",
	}

	// 首先尝试直接匹配常见音频编码
	if encode, ok := commonAudioEncodes[tokenUpper]; ok {
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.lastTokenType = "audioencode"
		if mv.AudioEncode == "" {
			mv.AudioEncode = encode
			mv.lastToken = tokenUpper
		} else {
			if strings.ToUpper(mv.AudioEncode) == "DTS" {
				mv.AudioEncode = mv.AudioEncode + "-" + encode
			} else {
				mv.AudioEncode = mv.AudioEncode + " " + encode
			}
		}
		return
	}

	if mv.Year == "" && mv.ResourcePix == "" && mv.ResourceType == "" &&
		mv.BeginSeason == nil && mv.BeginEpisode == nil {
		return
	}

	reRes := mv.audioEncodeRe.FindStringSubmatch(tokenUpper)
	if reRes != nil {
		mv.continueFlag = false
		mv.stopNameFlag = true
		mv.lastTokenType = "audioencode"
		// 检查切片长度，确保安全访问
		if len(reRes) > 1 {
			mv.lastToken = strings.ToUpper(reRes[1])
			if mv.AudioEncode == "" {
				mv.AudioEncode = reRes[1]
			} else {
				if strings.ToUpper(mv.AudioEncode) == "DTS" {
					mv.AudioEncode = mv.AudioEncode + "-" + reRes[1]
				} else {
					mv.AudioEncode = mv.AudioEncode + " " + reRes[1]
				}
			}
		}
	} else if isAllDigit(token) {
		if mv.lastTokenType == "audioencode" {
			if mv.AudioEncode != "" {
				if isAllDigit(mv.lastToken) {
					mv.AudioEncode = mv.AudioEncode + "." + token
				} else if len(mv.AudioEncode) > 0 && isAllDigit(string(mv.AudioEncode[len(mv.AudioEncode)-1])) {
					mv.AudioEncode = mv.AudioEncode[:len(mv.AudioEncode)-1] + " " + string(mv.AudioEncode[len(mv.AudioEncode)-1]) + "." + token
				} else {
					mv.AudioEncode = mv.AudioEncode + " " + token
				}
			}
			mv.lastToken = token
		}
	}
}

// fixName 去掉名字中不需要的干扰字符
func (mv *MetaVideo) fixName(name string) string {
	if name == "" {
		return name
	}

	// 去掉名字中不需要的干扰字符
	name = mv.nameNostringRe.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	// 过短的纯数字不要
	if isAllDigit(name) {
		num, _ := strconv.Atoi(name)
		if num < 1800 && mv.Year == "" && mv.BeginSeason == nil &&
			mv.ResourcePix == "" && mv.ResourceType == "" &&
			mv.VideoEncode == "" && mv.AudioEncode == "" {
			if mv.BeginEpisode == nil {
				mv.BeginEpisode = &num
				return ""
			} else if mv.IsInEpisode(num) && mv.BeginSeason == nil {
				return ""
			}
		}
	}

	return name
}

// getTitleFromDescription 从描述中提取标题（简化实现）
func (mv *MetaVideo) getTitleFromDescription(description string) string {
	// 简化实现，暂不实现
	return ""
}

// isPinyin 判断是否拼音（简化实现）
func (mv *MetaVideo) isPinyin(nameStr string) bool {
	// 简化实现，暂不实现
	return false
}
