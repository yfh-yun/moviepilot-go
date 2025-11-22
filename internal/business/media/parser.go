package media

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FileMetadata 文件元数据
type FileMetadata struct {
	Title       string   // 标题
	Year        int      // 年份
	Season      int      // 季号
	Episode     int      // 集号
	EpisodeTitle string  // 集标题
	Resolution  string   // 分辨率
	Source      string   // 来源
	Codec       string   // 编码
	Audio       string   // 音频
	Subtitle    string   // 字幕
	Group       string   // 发布组
	Extension   string   // 文件扩展名
	IsAnime     bool     // 是否动漫
	Type        string   // 媒体类型 movie/tv/anime
}

// ParseFileName 解析文件名
func ParseFileName(filePath string) *FileMetadata {
	filename := filepath.Base(filePath)
	
	// 移除扩展名
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	
	metadata := &FileMetadata{
		Extension: strings.ToLower(ext),
	}
	
	// 解析基本信息
	metadata = parseBasicInfo(name, metadata)
	
	// 解析分辨率
	metadata = parseResolution(name, metadata)
	
	// 解析来源
	metadata = parseSource(name, metadata)
	
	// 解析编码
	metadata = parseCodec(name, metadata)
	
	// 解析音频
	metadata = parseAudio(name, metadata)
	
	// 解析字幕
	metadata = parseSubtitle(name, metadata)
	
	// 解析发布组
	metadata = parseGroup(name, metadata)
	
	// 判断媒体类型
	metadata.Type = DetermineMediaType(metadata)
	
	return metadata
}

// parseBasicInfo 解析基本信息（标题、年份、季、集）
func parseBasicInfo(name string, metadata *FileMetadata) *FileMetadata {
	// 去除常见的发布组标记
	cleanName := name
	
	// 移除方括号内容（通常为发布组）
	cleanName = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(cleanName, "")
	
	// 移除圆括号内容（通常为年份等，但我们要保留年份）
	
	// 提取年份
	yearRegex := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if matches := yearRegex.FindStringSubmatch(cleanName); len(matches) > 0 {
		if year, err := strconv.Atoi(matches[0]); err == nil {
			metadata.Year = year
		}
	}
	
	// 提取季集信息（多种格式）
	seasonEpisodePatterns := []string{
		`S(\d{1,2})E(\d{1,3})`,           // S01E01
		`Season\s*(\d{1,2})\s*Episode\s*(\d{1,3})`, // Season 1 Episode 1
		`(\d{1,2})x(\d{1,3})`,           // 1x01
		`第(\d{1,2})季.*?第(\d{1,3})集`,   // 第1季第1集
		`EP(\d{1,3})`,                   // EP01 (单集)
		`(\d{1,3})`,                     // 纯数字（集号）
	}
	
	for _, pattern := range seasonEpisodePatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(cleanName); len(matches) >= 3 {
			if season, err := strconv.Atoi(matches[1]); err == nil {
				metadata.Season = season
			}
			if episode, err := strconv.Atoi(matches[2]); err == nil {
				metadata.Episode = episode
			}
			break
		} else if len(matches) == 2 {
			// 单集情况
			if episode, err := strconv.Atoi(matches[1]); err == nil {
				metadata.Episode = episode
			}
			break
		}
	}
	
	// 提取标题
	title := cleanName
	
	// 移除季集信息
	title = regexp.MustCompile(`(?i)(S\d{1,2}E\d{1,3}|Season\s*\d+\s*Episode\s*\d+|\d{1,2}x\d{1,3}|第\d+季.*?第\d+集|EP\d+|\d{1,3})`).ReplaceAllString(title, "")
	
	// 移除年份
	title = yearRegex.ReplaceAllString(title, "")
	
	// 移除分辨率、来源、编码等信息
	title = regexp.MustCompile(`(?i)(1080p|720p|480p|2160p|4K|BluRay|BDRip|DVDRip|WEB-DL|HDTV|HDCAM|x264|x265|HEVC|AAC|AC3|DTS|MP3)`).ReplaceAllString(title, "")
	
	// 移除方括号和圆括号内容
	title = regexp.MustCompile(`[\[\(].*?[\]\)]`).ReplaceAllString(title, "")
	
	// 清理分隔符
	title = regexp.MustCompile(`[._\-\+\s]+`).ReplaceAllString(title, " ")
	
	// 去除首尾空格和特殊字符
	title = strings.TrimSpace(title)
	title = strings.Trim(title, " -_.+")
	
	metadata.Title = title
	
	// 判断是否动漫
	metadata.IsAnime = isAnime(title, name)
	
	return metadata
}

// parseResolution 解析分辨率
func parseResolution(name string, metadata *FileMetadata) *FileMetadata {
	resolutions := map[string]string{
		"2160p": "2160p",
		"4K":    "2160p",
		"1080p": "1080p",
		"720p":  "720p",
		"480p":  "480p",
		"360p":  "360p",
	}
	
	for pattern, res := range resolutions {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(name) {
			metadata.Resolution = res
			break
		}
	}
	
	return metadata
}

// parseSource 解析来源
func parseSource(name string, metadata *FileMetadata) *FileMetadata {
	sources := map[string]string{
		"BluRay":   "BluRay",
		"BDRip":    "BluRay",
		"WEB-DL":   "WEB-DL",
		"WEBDL":    "WEB-DL",
		"WEB":      "WEB",
		"HDTV":     "HDTV",
		"DVD":      "DVD",
		"DVDRip":   "DVD",
		"HDCAM":    "CAM",
		"CAM":      "CAM",
		"TS":       "CAM",
		"TeleSync": "CAM",
	}
	
	for pattern, source := range sources {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(name) {
			metadata.Source = source
			break
		}
	}
	
	return metadata
}

// parseCodec 解析编码
func parseCodec(name string, metadata *FileMetadata) *FileMetadata {
	codecs := map[string]string{
		"x264":   "x264",
		"h264":   "x264",
		"AVC":    "x264",
		"x265":   "x265",
		"h265":   "x265",
		"HEVC":   "x265",
		"XVID":   "XviD",
		"DIVX":   "DivX",
		"VC-1":   "VC-1",
		"AV1":    "AV1",
	}
	
	for pattern, codec := range codecs {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(name) {
			metadata.Codec = codec
			break
		}
	}
	
	return metadata
}

// parseAudio 解析音频
func parseAudio(name string, metadata *FileMetadata) *FileMetadata {
	audios := map[string]string{
		"AAC":    "AAC",
		"AC3":    "AC3",
		"DTS":    "DTS",
		"DTSHD":  "DTS-HD",
		"DTS-HD": "DTS-HD",
		"TrueHD": "TrueHD",
		"FLAC":   "FLAC",
		"MP3":    "MP3",
		"Opus":   "Opus",
	}
	
	for pattern, audio := range audios {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(name) {
			metadata.Audio = audio
			break
		}
	}
	
	return metadata
}

// parseSubtitle 解析字幕
func parseSubtitle(name string, metadata *FileMetadata) *FileMetadata {
	if regexp.MustCompile(`(?i)(硬字幕|内嵌字幕|内嵌|Hardcoded|HC)`).MatchString(name) {
		metadata.Subtitle = "Hardcoded"
	} else if regexp.MustCompile(`(?i)(中字|中英字幕|Chinese|CHS|CHT)`).MatchString(name) {
		metadata.Subtitle = "Chinese"
	} else if regexp.MustCompile(`(?i)(英字幕|English)`).MatchString(name) {
		metadata.Subtitle = "English"
	}
	
	return metadata
}

// parseGroup 解析发布组
func parseGroup(name string, metadata *FileMetadata) *FileMetadata {
	// 方括号内的发布组
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	if matches := re.FindStringSubmatch(name); len(matches) > 1 {
		metadata.Group = matches[1]
		return metadata
	}
	
	// 常见的发布组标识
	groupPatterns := []string{
		`-(\w+)$`,           // 末尾的-组名
		`_([A-Za-z0-9]+)$`,  // 末尾的_组名
		`\(([^\)]+)\)$`,      // 圆括号内的组名
	}
	
	for _, pattern := range groupPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(name); len(matches) > 1 {
			group := strings.TrimSpace(matches[1])
			if len(group) > 0 && len(group) < 20 { // 合理的组名长度
				metadata.Group = group
				break
			}
		}
	}
	
	return metadata
}

// isAnime 判断是否为动漫
func isAnime(title, filename string) bool {
	animeIndicators := []string{
		"Anime", "动漫", "动画", "Manga", "漫画",
		"SubsPlease", "HorribleSubs", "Erai-raws",
	}
	
	lowerTitle := strings.ToLower(title)
	lowerFilename := strings.ToLower(filename)
	
	for _, indicator := range animeIndicators {
		if strings.Contains(lowerTitle, strings.ToLower(indicator)) ||
		   strings.Contains(lowerFilename, strings.ToLower(indicator)) {
			return true
		}
	}
	
	// 检查文件名格式（动漫常见格式）
	animePatterns := []string{
		`\[\w+\]\s*[\w\s-]+\s*-\s*\d+`, // [Group] Title - 01
		`\d{2,3}\s*[\w\s-]+`,           // 01 Title
	}
	
	for _, pattern := range animePatterns {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(filename) {
			return true
		}
	}
	
	return false
}

// DetermineMediaType 确定媒体类型
func DetermineMediaType(metadata *FileMetadata) string {
	if metadata.IsAnime {
		return "anime"
	}
	
	if metadata.Season > 0 || metadata.Episode > 0 {
		return "tv"
	}
	
	return "movie"
}

// SanitizeForSearch 清理标题用于搜索
func SanitizeForSearch(title string) string {
	// 移除特殊字符
	clean := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(title, " ")
	
	// 替换多个空格为单个空格
	clean = regexp.MustCompile(`\s+`).ReplaceAllString(clean, " ")
	
	// 去除首尾空格
	clean = strings.TrimSpace(clean)
	
	return clean
}

// ExtractYearFromTitle 从标题中提取年份
func ExtractYearFromTitle(title string) int {
	yearRegex := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if matches := yearRegex.FindStringSubmatch(title); len(matches) > 0 {
		if year, err := strconv.Atoi(matches[0]); err == nil {
			return year
		}
	}
	return 0
}

// IsVideoFile 判断是否为视频文件
func IsVideoFile(filename string) bool {
	videoExts := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".3gp":  true,
		".mpg":  true,
		".mpeg": true,
		".ts":   true,
		".m2ts": true,
	}
	
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExts[ext]
}

// IsSubtitleFile 判断是否为字幕文件
func IsSubtitleFile(filename string) bool {
	subtitleExts := map[string]bool{
		".srt":  true,
		".ass":  true,
		".ssa":  true,
		".sub":  true,
		".vtt":  true,
	}
	
	ext := strings.ToLower(filepath.Ext(filename))
	return subtitleExts[ext]
}