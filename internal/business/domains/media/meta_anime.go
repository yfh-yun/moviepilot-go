package media

import (
	"regexp"
	"strconv"
	"strings"
)

// MetaAnime 动漫元信息解析
type MetaAnime struct {
	*MetaBase
	deps MetaParserDeps
}

// NewMetaAnime 创建新的MetaAnime实例
func NewMetaAnime(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaAnime {
	meta := &MetaAnime{
		MetaBase: NewMetaBase(title, title, subtitle, isFile),
		deps:     deps,
	}

	// 设置媒体类型
	meta.Type = MediaTypeAnime

	// 解析标题
	meta.parseAnimeInfo(title, subtitle, isFile)

	return meta
}

// parseAnimeInfo 解析动漫信息
func (ma *MetaAnime) parseAnimeInfo(title, subtitle string, isFile bool) {
	if title == "" {
		return
	}

	// 调用第三方模块识别动漫
	originalTitle := title
	title = ma.prepareTitle(title)

	// 解析标题，集成类似anitopy的解析逻辑
	ma.parseTitle(title, originalTitle)

	// 解析副标题，只要季和集
	ma.InitSubtitle(ma.OrgString)
	if subtitle != "" {
		ma.InitSubtitle(subtitle)
	}

	// 确保媒体类型
	if ma.Type == "" {
		ma.Type = MediaTypeAnime
	}
}

// prepareTitle 预处理标题
func (ma *MetaAnime) prepareTitle(title string) string {
	if title == "" {
		return title
	}

	// 所有【】换成[]
	title = strings.ReplaceAll(title, "【", "[")
	title = strings.ReplaceAll(title, "】", "]")
	title = strings.TrimSpace(title)

	// 截掉xx番剧漫
	match := regexp.MustCompile(`新番|月?番|[日美国][漫剧]`).FindStringIndex(title)
	if match != nil && match[1] < len(title)-1 {
		title = regexp.MustCompile(`.*番.|.*[日美国][漫剧].`).ReplaceAllString(title, "")
	} else if match != nil {
		// 截掉末尾的番剧标识
		if strings.Contains(title, "[") {
			title = title[:strings.LastIndex(title, "[")]
		}
	}

	// 截掉分类
	firstItem := title
	if strings.Contains(title, "]") {
		firstItem = strings.Split(title, "]")[0]
	}

	// 检查分类关键词
	classifyKeywords := []string{
		"动漫", "动画", "漫画", "纪录片", "电影", "视频", "连续剧", "剧集", "日", "美", "韩", "中", "港", "台", "海外", "亚洲", "华语", "大陆", "综艺", "原盘", "高清",
		"TV", "Animation", "Movie", "Documentar", "Anime",
	}
	for _, keyword := range classifyKeywords {
		if strings.Contains(firstItem, keyword) {
			title = strings.TrimSpace(regexp.MustCompile(`^[^]]*]`).ReplaceAllString(title, ""))
			break
		}
	}

	// 去掉大小
	title = regexp.MustCompile(`[0-9.]+\s*[MGT]i?B`).ReplaceAllString(title, "")

	// 将TVxx改为xx
	title = regexp.MustCompile(`\[TV\s+(\d{1,4})`).ReplaceAllString(title, "[$1")

	// 将4K转为2160p
	title = regexp.MustCompile(`\[4k\]`).ReplaceAllString(strings.ToLower(title), "2160p")

	// 处理/分隔的中英文标题
	names := strings.Split(title, "]")
	if len(names) > 1 && !strings.Contains(title, "- ") {
		titles := []string{}
		for _, name := range names {
			if name == "" {
				continue
			}
			leftChar := ""
			if strings.HasPrefix(name, "[") {
				leftChar = "["
				name = name[1:]
			}
			if strings.Contains(name, "/") {
				splitNames := strings.Split(name, "/")
				if splitNames[len(splitNames)-1] != "" {
					titles = append(titles, leftChar+strings.TrimSpace(splitNames[len(splitNames)-1]))
				} else {
					titles = append(titles, leftChar+strings.TrimSpace(splitNames[0]))
				}
			} else if name != "" {
				if ma.isChinese(name) && !ma.isAllChinese(name) {
					if !regexp.MustCompile(`\[\d+`).MatchString(name) {
						name = strings.TrimSpace(regexp.MustCompile(`[\d|#:：\-()（）\p{Han}]`).ReplaceAllString(name, ""))
					}
					if name == "" || strings.TrimSpace(name) == "" {
						continue
					}
				}
				if name == "[" {
					titles = append(titles, "")
				} else {
					titles = append(titles, leftChar+strings.TrimSpace(name))
				}
			}
		}
		title = strings.Join(titles, "]")
	}

	return title
}

// parseTitle 解析动漫标题
func (ma *MetaAnime) parseTitle(title, originalTitle string) {
	if title == "" {
		return
	}

	// 动漫特殊关键词
	animeNoWords := []string{"CHS&CHT", "MP4", "GB MP4", "WEB-DL"}
	nameNoStringRe := regexp.MustCompile(`S\d{2}\s*\-\s*S\d{2}|S\d{2}|\s+S\d{1,2}|EP?\d{2,4}\s*\-\s*EP?\d{2,4}|EP?\d{2,4}|\s+EP?\d{1,4}|\s+GB`)

	// 使用正则提取动漫标题
	name := ""

	// 尝试从[]中提取名称
	bracketMatch := regexp.MustCompile(`\[(.+?)\]`).FindStringSubmatch(title)
	if bracketMatch != nil && bracketMatch[1] != "" {
		name = bracketMatch[1]
	}

	// 检查名称是否有效
	if name == "" || ma.containsAny(name, animeNoWords) || (len(name) < 5 && !ma.isChinese(name)) {
		// 尝试添加[ANIME]前缀后再提取
		tempTitle := "[ANIME]" + title
		bracketMatch := regexp.MustCompile(`\[(.+?)\]`).FindStringSubmatch(tempTitle)
		if bracketMatch != nil && bracketMatch[1] != "" {
			name = bracketMatch[1]
		}
	}

	// 再次检查名称是否有效
	if name == "" || ma.containsAny(name, animeNoWords) || (len(name) < 5 && !ma.isChinese(name)) {
		// 尝试从第一个[]中提取
		bracketMatch := regexp.MustCompile(`\[(.+?)\]`).FindStringSubmatch(title)
		if bracketMatch != nil && bracketMatch[1] != "" {
			name = bracketMatch[1]
		}
	}

	// 拆份中英文名称
	if name != "" {
		splitFlag := true

		// 按/拆分中英文
		if strings.Contains(name, "/") {
			names := strings.Split(name, "/")
			if ma.isChinese(names[0]) {
				ma.CnName = names[0]
				if len(names) > 1 {
					ma.EnName = names[1]
				}
				splitFlag = false
			} else if ma.isChinese(names[len(names)-1]) {
				ma.CnName = names[len(names)-1]
				if len(names) > 1 {
					ma.EnName = names[0]
				}
				splitFlag = false
			} else {
				name = names[len(names)-1]
			}
		}

		// 拆分中英文
		if splitFlag {
			lastwordType := ""
			for _, word := range strings.Split(name, " ") {
				if word == "" {
					continue
				}
				if strings.HasSuffix(word, "]") {
					word = word[:len(word)-1]
				}
				if ma.isDigit(word) {
					if lastwordType == "cn" {
						ma.CnName = ma.concat(ma.CnName, word)
					} else if lastwordType == "en" {
						ma.EnName = ma.concat(ma.EnName, word)
					}
				} else if ma.isChinese(word) {
					ma.CnName = ma.concat(ma.CnName, word)
					lastwordType = "cn"
				} else {
					ma.EnName = ma.concat(ma.EnName, word)
					lastwordType = "en"
				}
			}
		}

		// 清理名称
		if ma.CnName != "" {
			ma.CnName = strings.TrimSpace(nameNoStringRe.ReplaceAllString(ma.CnName, ""))
		}
		if ma.EnName != "" {
			ma.EnName = strings.TrimSpace(nameNoStringRe.ReplaceAllString(ma.EnName, ""))
			ma.EnName = ma.strTitle(ma.EnName)
			ma.Title = ma.strTitle(ma.EnName)
		}
	}

	// 解析年份
	yearMatch := regexp.MustCompile(`(19\d{2}|20\d{2})`).FindStringSubmatch(title)
	if yearMatch != nil {
		ma.Year = yearMatch[1]
	}

	// 解析季号
	seasonMatch := regexp.MustCompile(`(?i)S(\d{1,2})`).FindStringSubmatch(title)
	if seasonMatch != nil {
		season, _ := strconv.Atoi(seasonMatch[1])
		ma.SetSeason(season)
		ma.Type = MediaTypeAnime
	}

	// 解析集号
	episodeMatch := regexp.MustCompile(`(?i)EP?(\d{1,4})`).FindStringSubmatch(title)
	if episodeMatch != nil {
		episode, _ := strconv.Atoi(episodeMatch[1])
		ma.SetEpisode(episode)
		ma.Type = MediaTypeAnime
	}

	// 解析分辨率
	resolutionMatch := regexp.MustCompile(`(?i)(\d{3,4}p|4k|2160p|1080p|720p|480p)`).FindStringSubmatch(title)
	if resolutionMatch != nil {
		ma.ResourcePix = ResourcePix(strings.ToLower(resolutionMatch[1]))
	}

	// 解析视频编码
	videoEncodeMatch := regexp.MustCompile(`(?i)(x264|h264|avc|x265|h265|hevc|vp9|av1|mpeg-4|mp4|mpeg-2|mp2)`).FindStringSubmatch(title)
	if videoEncodeMatch != nil {
		ma.VideoEncode = strings.ToUpper(videoEncodeMatch[1])
	}

	// 解析音频编码
	audioEncodeMatch := regexp.MustCompile(`(?i)(aac|ac3|dd|dts|dts-hd|truehd|flac)`).FindStringSubmatch(title)
	if audioEncodeMatch != nil {
		ma.AudioEncode = strings.ToUpper(audioEncodeMatch[1])
	}

	// 资源类型
	if ma.Type == "" {
		ma.Type = MediaTypeAnime
	}

	// 制作组/字幕组
	if ma.deps.ReleaseMatcher != nil {
		releaseGroup := ma.deps.ReleaseMatcher.Match(originalTitle, "")
		if releaseGroup != "" {
			ma.ResourceTeam = releaseGroup
		}
	}

	// 自定义占位符
	if ma.deps.CustomizationMatch != nil {
		customization := ma.deps.CustomizationMatch.Match(originalTitle)
		if customization != "" {
			ma.Customization = customization
		}
	}
}

// isChinese 检查字符串是否包含中文
func (ma *MetaAnime) isChinese(s string) bool {
	return regexp.MustCompile(`\p{Han}`).MatchString(s)
}

// isAllChinese 检查字符串是否全部为中文
func (ma *MetaAnime) isAllChinese(s string) bool {
	return regexp.MustCompile(`^\p{Han}+$`).MatchString(s)
}

// isDigit 检查字符串是否为数字
func (ma *MetaAnime) isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// containsAny 检查字符串是否包含任何指定的子串
func (ma *MetaAnime) containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// concat 连接字符串，处理空值
func (ma *MetaAnime) concat(a, b string) string {
	if a == "" {
		return b
	}
	return a + " " + b
}

// strTitle 将字符串转换为首字母大写，其他小写
func (ma *MetaAnime) strTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.Title(strings.ToLower(s))
}
