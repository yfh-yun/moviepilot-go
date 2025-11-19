package meta

import (
	"strings"
	"regexp"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// 媒体类型常量
const (
	MediaTypeMovie       = "movie"       // 电影
	MediaTypeTVSeries    = "tv"         // 电视剧
	MediaTypeAnime       = "anime"      // 动漫
	MediaTypeDocumentary = "documentary" // 纪录片
	MediaTypeVariety     = "variety"    // 综艺
	MediaTypeOther       = "other"      // 其他
)

// 解析状态常量
const (
	ParseStatusSuccess       = "success"       // 解析成功
	ParseStatusPartially     = "partially"     // 部分解析
	ParseStatusFailed        = "failed"        // 解析失败
	ParseStatusUncertain     = "uncertain"     // 不确定
)

// 区域常量
const (
	RegionCN = "CN" // 中国大陆
	RegionHK = "HK" // 中国香港
	RegionTW = "TW" // 中国台湾
	RegionUS = "US" // 美国
	RegionJP = "JP" // 日本
	RegionKR = "KR" // 韩国
	RegionUK = "UK" // 英国
	RegionEU = "EU" // 欧洲
	RegionOther = "OTHER" // 其他
)

// 分辨率常量
const (
	ResolutionUnknown = "unknown" // 未知
	Resolution480p    = "480p"    // 480p
	Resolution720p    = "720p"    // 720p
	Resolution1080p   = "1080p"   // 1080p
	Resolution2K      = "2160p"   // 2K
	Resolution4K      = "4096p"   // 4K
	Resolution8K      = "8192p"   // 8K
)

// 视频编码常量
const (
	VideoCodecUnknown = "unknown" // 未知
	VideoCodecH264    = "h264"    // H.264/AVC
	VideoCodecH265    = "h265"    // H.265/HEVC
	VideoCodecVP9     = "vp9"     // VP9
	VideoCodecAV1     = "av1"     // AV1
	VideoCodecX264    = "x264"    // X264
	VideoCodecX265    = "x265"    // X265
)

// 音频编码常量
const (
	AudioCodecUnknown = "unknown" // 未知
	AudioCodecAAC     = "aac"     // AAC
	AudioCodecMP3     = "mp3"     // MP3
	AudioCodecFLAC    = "flac"    // FLAC
	AudioCodecDTS     = "dts"     // DTS
	AudioCodecTrueHD  = "truehd"  // TrueHD
	AudioCodecDTSX    = "dtsx"    // DTS:X
	AudioCodecATMOS   = "atmos"   // Dolby Atmos
)

// 字幕类型常量
const (
	SubtitleTypeUnknown = "unknown" // 未知
	SubtitleTypeSRT     = "srt"     // SRT
	SubtitleTypeASS     = "ass"     // ASS
	SubtitleTypeSSA     = "ssa"     // SSA
	SubtitleTypeVTT     = "vtt"     // WebVTT
)

// MetaInfo 元数据接口
type MetaInfo interface {
	// 基础信息
	GetTitle() string
	SetTitle(title string)
	GetOriginalTitle() string
	SetOriginalTitle(title string)
	GetYear() int
	SetYear(year int)
	GetMediaType() string
	SetMediaType(mediaType string)
	GetOverview() string
	SetOverview(overview string)
	GetPosterURL() string
	SetPosterURL(url string)
	GetBackdropURL() string
	SetBackdropURL(url string)

	// 解析相关
	GetName() string
	SetName(name string)
	GetParseStatus() string
	SetParseStatus(status string)
	GetConfidence() float64
	SetConfidence(confidence float64)

	// 媒体信息
	GetResolution() string
	SetResolution(resolution string)
	GetVideoCodec() string
	SetVideoCodec(codec string)
	GetAudioCodec() string
	SetAudioCodec(codec string)
	GetAudioChannels() string
	SetAudioChannels(channels string)

	// 文件信息
	GetFileSize() int64
	SetFileSize(size int64)
	GetExt() string
	SetExt(ext string)

	// 其他方法
	IsValid() bool
	ToString() string
	Clone() MetaInfo
}

// MetaManager 元数据管理器
type MetaManager struct {
	logger         *logger.Logger
	animeKeywords  map[string]bool
	ignoreKeywords map[string]bool
	releaseGroups  []ReleaseGroup
}

// NewMetaManager 创建元数据管理器
func NewMetaManager(log *logger.Logger) *MetaManager {
	mm := &MetaManager{
		logger: log,
	}
	mm.initKeywords()
	mm.initReleaseGroups()
	log.Info("Meta manager initialized")
	return mm
}

// initKeywords 初始化关键词
func (mm *MetaManager) initKeywords() {
	// 初始化动漫关键词
	mm.animeKeywords = map[string]bool{
		"アニメ": true,
		"Anime": true,
		"BDrip": true,
		"WEB-DL": true,
		"1080p": true, // 这些会被进一步验证
		"720p": true,
		"2160p": true,
		"动画": true,
		"アニ": true,
		"OVA": true,
		"OAD": true,
		"SP": true,
		"剧场版": true,
		"劇場版": true,
	}

	// 初始化忽略关键词
	mm.ignoreKeywords = map[string]bool{
		"Sample": true,
		"sample": true,
		"SAMPLE": true,
		"Trailer": true,
		"trailer": true,
		"TRAILER": true,
		"Preview": true,
		"preview": true,
		"PREVIEW": true,
		"Extras": true,
		"extras": true,
		"EXTRA": true,
	}
}

// initReleaseGroups 初始化发布组
func (mm *MetaManager) initReleaseGroups() {
	mm.releaseGroups = []ReleaseGroup{
		{Name: "CMCT", Tags: []string{"CMCT"}},
		{Name: "FRDS", Tags: []string{"FRDS"}},
		{Name: "HDChina", Tags: []string{"HDChina"}},
		{Name: "CHD", Tags: []string{"CHD"}},
		{Name: "HDC", Tags: []string{"HDC"}},
		{Name: "HDTime", Tags: []string{"HDTime"}},
		{Name: "LemonHD", Tags: []string{"LemonHD"}},
		{Name: "TBS", Tags: []string{"TBS"}},
		{Name: "SBS", Tags: []string{"SBS"}},
		{Name: "TVN", Tags: []string{"TVN"}},
	}
}

// CreateMetaInfo 创建元数据实例
func (mm *MetaManager) CreateMetaInfo(name string, mediaType string) MetaInfo {
	switch mediaType {
	case MediaTypeAnime:
		return NewMetaAnime(name)
	case MediaTypeTVSeries:
		return NewMetaVideo(name)
	case MediaTypeMovie:
		return NewMetaVideo(name)
	default:
		return NewMetaVideo(name)
	}
}

// DetectMediaType 检测媒体类型
func (mm *MetaManager) DetectMediaType(name string) string {
	name = strings.ToLower(name)

	// 检查是否为动漫
	for keyword := range mm.animeKeywords {
		if strings.Contains(name, strings.ToLower(keyword)) {
			// 进一步确认，避免误判
			if mm.isAnime(name) {
				return MediaTypeAnime
			}
		}
	}

	// 检查是否为剧集（包含SxxExx格式）
	seriesRegex := regexp.MustCompile(`s\d{2,}e\d{2,}`)
	if seriesRegex.MatchString(name) {
		return MediaTypeTVSeries
	}

	// 检查是否为纪录片
	docRegex := regexp.MustCompile(`doc(u|umentary)?`)
	if docRegex.MatchString(name) {
		return MediaTypeDocumentary
	}

	// 默认为电影
	return MediaTypeMovie
}

// isAnime 判断是否为动漫
func (mm *MetaManager) isAnime(name string) bool {
	// 动漫特有的格式和关键词
	animePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\[\w+\]`), // 如[ANi]、[FFF]
		regexp.MustCompile(`\(\w+raw\)`), // 如(RAW)
		regexp.MustCompile(`\d{4}年\d{1,2}月`), // 日本日期格式
		regexp.MustCompile(`[ア-ン]+`), // 片假名
	}

	for _, pattern := range animePatterns {
		if pattern.MatchString(name) {
			return true
		}
	}

	// 检查是否包含常见动漫关键词
	animeKeywords := []string{
		"动画", "动漫", "アニメ", "オリジナル",
		"bdrip", "webrip", "ova", "oad",
	}

	for _, keyword := range animeKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}

	return false
}

// ParseFilename 解析文件名
func (mm *MetaManager) ParseFilename(filename string) MetaInfo {
	// 移除文件扩展名
	extIndex := strings.LastIndex(filename, ".")
	ext := ""
	if extIndex > 0 {
		ext = filename[extIndex+1:]
		filename = filename[:extIndex]
	}

	// 检测媒体类型
	mediaType := mm.DetectMediaType(filename)

	// 创建对应类型的元数据实例
	meta := mm.CreateMetaInfo(filename, mediaType)
	meta.SetExt(ext)

	// 解析分辨率
	resolution := mm.ParseResolution(filename)
	meta.SetResolution(resolution)

	// 解析编码信息
	videoCodec := mm.ParseVideoCodec(filename)
	audioCodec := mm.ParseAudioCodec(filename)
	meta.SetVideoCodec(videoCodec)
	meta.SetAudioCodec(audioCodec)

	// 解析音频声道
	audioChannels := mm.ParseAudioChannels(filename)
	meta.SetAudioChannels(audioChannels)

	return meta
}

// ParseResolution 解析分辨率
func (mm *MetaManager) ParseResolution(name string) string {
	patterns := map[string]string{
		"8k":      Resolution8K,
		"4k":      Resolution4K,
		"2160p":   Resolution4K,
		"4096p":   Resolution4K,
		"2k":      Resolution2K,
		"1440p":   Resolution2K,
		"1080p":   Resolution1080p,
		"720p":    Resolution720p,
		"480p":    Resolution480p,
	}

	name = strings.ToLower(name)
	for pattern, resolution := range patterns {
		if strings.Contains(name, pattern) {
			return resolution
		}
	}

	return ResolutionUnknown
}

// ParseVideoCodec 解析视频编码
func (mm *MetaManager) ParseVideoCodec(name string) string {
	patterns := map[string]string{
		"av1":     VideoCodecAV1,
		"vp9":     VideoCodecVP9,
		"h265":    VideoCodecH265,
		"hevc":    VideoCodecH265,
		"x265":    VideoCodecX265,
		"h264":    VideoCodecH264,
		"avc":     VideoCodecH264,
		"x264":    VideoCodecX264,
	}

	name = strings.ToLower(name)
	for pattern, codec := range patterns {
		if strings.Contains(name, pattern) {
			return codec
		}
	}

	return VideoCodecUnknown
}

// ParseAudioCodec 解析音频编码
func (mm *MetaManager) ParseAudioCodec(name string) string {
	patterns := map[string]string{
		"dts:x":   AudioCodecDTSX,
		"dts:x":   AudioCodecDTSX,
		"atmos":   AudioCodecATMOS,
		"truehd":  AudioCodecTrueHD,
		"dts-hd":  AudioCodecDTS,
		"dts":     AudioCodecDTS,
		"flac":    AudioCodecFLAC,
		"aac":     AudioCodecAAC,
		"mp3":     AudioCodecMP3,
	}

	name = strings.ToLower(name)
	for pattern, codec := range patterns {
		if strings.Contains(name, pattern) {
			return codec
		}
	}

	return AudioCodecUnknown
}

// ParseAudioChannels 解析音频声道
func (mm *MetaManager) ParseAudioChannels(name string) string {
	patterns := map[string]string{
		"7.1":  "7.1",
		"5.1":  "5.1",
		"2.0":  "2.0",
		"2.1":  "2.1",
		"6ch":  "5.1",
		"8ch":  "7.1",
	}

	name = strings.ToLower(name)
	for pattern, channels := range patterns {
		if strings.Contains(name, pattern) {
			return channels
		}
	}

	return ""
}

// IsIgnoreFile 判断是否忽略文件
func (mm *MetaManager) IsIgnoreFile(name string) bool {
	name = strings.ToLower(name)
	for keyword := range mm.ignoreKeywords {
		if strings.Contains(name, strings.ToLower(keyword)) {
			return true
		}
	}

	// 忽略样例文件
	samplePatterns := []*regexp.Regexp{
		regexp.MustCompile(`sample\d*\.`),
		regexp.MustCompile(`\bsample\b`),
	}

	for _, pattern := range samplePatterns {
		if pattern.MatchString(name) {
			return true
		}
	}

	return false
}

// GetReleaseGroup 获取发布组
func (mm *MetaManager) GetReleaseGroup(name string) string {
	name = strings.ToLower(name)
	for _, group := range mm.releaseGroups {
		for _, tag := range group.Tags {
			if strings.Contains(name, strings.ToLower(tag)) {
				return group.Name
			}
		}
	}
	return ""
}

// GetRegionFromTitle 从标题获取区域信息
func (mm *MetaManager) GetRegionFromTitle(title string) string {
	title = strings.ToLower(title)

	regionPatterns := map[string]string{
		"中字": RegionCN,
		"国语": RegionCN,
		"大陆": RegionCN,
		"中国": RegionCN,
		"香港": RegionHK,
		"港剧": RegionHK,
		"台剧": RegionTW,
		"台湾": RegionTW,
		"日語": RegionJP,
		"日语": RegionJP,
		"日本": RegionJP,
		"韓語": RegionKR,
		"韩语": RegionKR,
		"韩国": RegionKR,
		"英語": RegionUS,
		"英语": RegionUS,
		"美国": RegionUS,
		"uk": RegionUK,
		"british": RegionUK,
	}

	for pattern, region := range regionPatterns {
		if strings.Contains(title, pattern) {
			return region
		}
	}

	return RegionOther
}

// CleanupTitle 清理标题
func (mm *MetaManager) CleanupTitle(title string) string {
	// 移除常见的额外信息
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[.*?\]`),
		regexp.MustCompile(`\(.*?\)`),
		regexp.MustCompile(`\{.*?\}`),
		regexp.MustCompile(`\s+\d{3,4}p\s*`),
		regexp.MustCompile(`\s+[xh]26[45]\s*`),
		regexp.MustCompile(`\s+hevc\s*`),
		regexp.MustCompile(`\s+aac\s*`),
		regexp.MustCompile(`\s+dts\s*`),
		regexp.MustCompile(`\s+flac\s*`),
		regexp.MustCompile(`\s+\d+\.\d+ch\s*`),
		regexp.MustCompile(`\s+web-?dl\s*`),
		regexp.MustCompile(`\s+bdrip\s*`),
		regexp.MustCompile(`\s+bluray\s*`),
	}

	result := title
	for _, pattern := range patterns {
		result = pattern.ReplaceAllString(result, "")
	}

	// 移除多余的空格
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	return result
}