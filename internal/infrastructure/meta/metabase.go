package meta

import (
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"time"
)

// MetaBase 元数据基础类
type MetaBase struct {
	// 基础信息
	Title         string            `json:"title"`          // 标题
	OriginalTitle string            `json:"original_title"` // 原始标题
	Year          int               `json:"year"`           // 年份
	MediaType     string            `json:"media_type"`     // 媒体类型
	Overview      string            `json:"overview"`       // 简介
	PosterURL     string            `json:"poster_url"`     // 海报URL
	BackdropURL   string            `json:"backdrop_url"`   // 背景图URL
	
	// 解析相关
	Name          string            `json:"name"`           // 原始名称
	ParseStatus   string            `json:"parse_status"`   // 解析状态
	Confidence    float64           `json:"confidence"`     // 置信度
	
	// 媒体信息
	Resolution    string            `json:"resolution"`     // 分辨率
	VideoCodec    string            `json:"video_codec"`    // 视频编码
	AudioCodec    string            `json:"audio_codec"`    // 音频编码
	AudioChannels string            `json:"audio_channels"` // 音频声道
	
	// 文件信息
	FileSize      int64             `json:"file_size"`      // 文件大小
	Ext           string            `json:"ext"`            // 文件扩展名
	
	// 额外信息
	Tags          []string          `json:"tags"`           // 标签
	Language      string            `json:"language"`       // 语言
	Country       string            `json:"country"`        // 国家
	Region        string            `json:"region"`         // 区域
	ReleaseGroup  string            `json:"release_group"`  // 发布组
	
	// 时间戳
	CreatedAt     time.Time         `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time         `json:"updated_at"`     // 更新时间
}

// NewMetaBase 创建基础元数据实例
func NewMetaBase(name string) *MetaBase {
	return &MetaBase{
		Name:        name,
		MediaType:   MediaTypeOther,
		ParseStatus: ParseStatusFailed,
		Resolution:  ResolutionUnknown,
		VideoCodec:  VideoCodecUnknown,
		AudioCodec:  AudioCodecUnknown,
		Tags:        make([]string, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// 基础信息相关方法

// GetTitle 获取标题
func (m *MetaBase) GetTitle() string {
	return m.Title
}

// SetTitle 设置标题
func (m *MetaBase) SetTitle(title string) {
	m.Title = title
	m.UpdatedAt = time.Now()
}

// GetOriginalTitle 获取原始标题
func (m *MetaBase) GetOriginalTitle() string {
	return m.OriginalTitle
}

// SetOriginalTitle 设置原始标题
func (m *MetaBase) SetOriginalTitle(title string) {
	m.OriginalTitle = title
	m.UpdatedAt = time.Now()
}

// GetYear 获取年份
func (m *MetaBase) GetYear() int {
	return m.Year
}

// SetYear 设置年份
func (m *MetaBase) SetYear(year int) {
	m.Year = year
	m.UpdatedAt = time.Now()
}

// GetMediaType 获取媒体类型
func (m *MetaBase) GetMediaType() string {
	return m.MediaType
}

// SetMediaType 设置媒体类型
func (m *MetaBase) SetMediaType(mediaType string) {
	m.MediaType = mediaType
	m.UpdatedAt = time.Now()
}

// GetOverview 获取简介
func (m *MetaBase) GetOverview() string {
	return m.Overview
}

// SetOverview 设置简介
func (m *MetaBase) SetOverview(overview string) {
	m.Overview = overview
	m.UpdatedAt = time.Now()
}

// GetPosterURL 获取海报URL
func (m *MetaBase) GetPosterURL() string {
	return m.PosterURL
}

// SetPosterURL 设置海报URL
func (m *MetaBase) SetPosterURL(url string) {
	m.PosterURL = url
	m.UpdatedAt = time.Now()
}

// GetBackdropURL 获取背景图URL
func (m *MetaBase) GetBackdropURL() string {
	return m.BackdropURL
}

// SetBackdropURL 设置背景图URL
func (m *MetaBase) SetBackdropURL(url string) {
	m.BackdropURL = url
	m.UpdatedAt = time.Now()
}

// 解析相关方法

// GetName 获取原始名称
func (m *MetaBase) GetName() string {
	return m.Name
}

// SetName 设置原始名称
func (m *MetaBase) SetName(name string) {
	m.Name = name
	m.UpdatedAt = time.Now()
}

// GetParseStatus 获取解析状态
func (m *MetaBase) GetParseStatus() string {
	return m.ParseStatus
}

// SetParseStatus 设置解析状态
func (m *MetaBase) SetParseStatus(status string) {
	m.ParseStatus = status
	m.UpdatedAt = time.Now()
}

// GetConfidence 获取置信度
func (m *MetaBase) GetConfidence() float64 {
	return m.Confidence
}

// SetConfidence 设置置信度
func (m *MetaBase) SetConfidence(confidence float64) {
	m.Confidence = confidence
	m.UpdatedAt = time.Now()
}

// 媒体信息相关方法

// GetResolution 获取分辨率
func (m *MetaBase) GetResolution() string {
	return m.Resolution
}

// SetResolution 设置分辨率
func (m *MetaBase) SetResolution(resolution string) {
	m.Resolution = resolution
	m.UpdatedAt = time.Now()
}

// GetVideoCodec 获取视频编码
func (m *MetaBase) GetVideoCodec() string {
	return m.VideoCodec
}

// SetVideoCodec 设置视频编码
func (m *MetaBase) SetVideoCodec(codec string) {
	m.VideoCodec = codec
	m.UpdatedAt = time.Now()
}

// GetAudioCodec 获取音频编码
func (m *MetaBase) GetAudioCodec() string {
	return m.AudioCodec
}

// SetAudioCodec 设置音频编码
func (m *MetaBase) SetAudioCodec(codec string) {
	m.AudioCodec = codec
	m.UpdatedAt = time.Now()
}

// GetAudioChannels 获取音频声道
func (m *MetaBase) GetAudioChannels() string {
	return m.AudioChannels
}

// SetAudioChannels 设置音频声道
func (m *MetaBase) SetAudioChannels(channels string) {
	m.AudioChannels = channels
	m.UpdatedAt = time.Now()
}

// 文件信息相关方法

// GetFileSize 获取文件大小
func (m *MetaBase) GetFileSize() int64 {
	return m.FileSize
}

// SetFileSize 设置文件大小
func (m *MetaBase) SetFileSize(size int64) {
	m.FileSize = size
	m.UpdatedAt = time.Now()
}

// GetExt 获取文件扩展名
func (m *MetaBase) GetExt() string {
	return m.Ext
}

// SetExt 设置文件扩展名
func (m *MetaBase) SetExt(ext string) {
	m.Ext = ext
	m.UpdatedAt = time.Now()
}

// 额外信息相关方法

// GetTags 获取标签
func (m *MetaBase) GetTags() []string {
	return m.Tags
}

// AddTag 添加标签
func (m *MetaBase) AddTag(tag string) {
	for _, t := range m.Tags {
		if t == tag {
			return // 避免重复
		}
	}
	m.Tags = append(m.Tags, tag)
	m.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (m *MetaBase) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)
			m.UpdatedAt = time.Now()
			return
		}
	}
}

// GetLanguage 获取语言
func (m *MetaBase) GetLanguage() string {
	return m.Language
}

// SetLanguage 设置语言
func (m *MetaBase) SetLanguage(language string) {
	m.Language = language
	m.UpdatedAt = time.Now()
}

// GetCountry 获取国家
func (m *MetaBase) GetCountry() string {
	return m.Country
}

// SetCountry 设置国家
func (m *MetaBase) SetCountry(country string) {
	m.Country = country
	m.UpdatedAt = time.Now()
}

// GetRegion 获取区域
func (m *MetaBase) GetRegion() string {
	return m.Region
}

// SetRegion 设置区域
func (m *MetaBase) SetRegion(region string) {
	m.Region = region
	m.UpdatedAt = time.Now()
}

// GetReleaseGroup 获取发布组
func (m *MetaBase) GetReleaseGroup() string {
	return m.ReleaseGroup
}

// SetReleaseGroup 设置发布组
func (m *MetaBase) SetReleaseGroup(group string) {
	m.ReleaseGroup = group
	m.UpdatedAt = time.Now()
}

// 通用方法

// IsValid 判断元数据是否有效
func (m *MetaBase) IsValid() bool {
	return m.ParseStatus == ParseStatusSuccess || m.ParseStatus == ParseStatusPartially
}

// ToString 转换为字符串
func (m *MetaBase) ToString() string {
	return fmt.Sprintf("%s (%d) [%s]", m.Title, m.Year, m.MediaType)
}

// Clone 克隆元数据
func (m *MetaBase) Clone() MetaInfo {
	clone := *m
	// 深拷贝切片
	clone.Tags = make([]string, len(m.Tags))
	copy(clone.Tags, m.Tags)
	return &clone
}

// ParseYearFromName 从名称中解析年份
func (m *MetaBase) ParseYearFromName() int {
	// 匹配4位数字年份（1900-2099）
	yearRegex := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	matches := yearRegex.FindStringSubmatch(m.Name)
	if len(matches) > 0 {
		year, err := strconv.Atoi(matches[0])
		if err == nil && year >= 1900 && year <= time.Now().Year()+1 {
			return year
		}
	}
	return 0
}

// ParseYearFromTitle 从标题中解析年份
func (m *MetaBase) ParseYearFromTitle() int {
	// 匹配括号中的年份，如 "Movie (2023)"
	yearRegex := regexp.MustCompile(`\((19|20)\d{2}\)`)
	matches := yearRegex.FindStringSubmatch(m.Title)
	if len(matches) > 0 {
		yearStr := matches[1] + matches[0][3:5] // 提取年份数字
		year, err := strconv.Atoi(yearStr)
		if err == nil && year >= 1900 && year <= time.Now().Year()+1 {
			return year
		}
	}
	return 0
}

// CleanupName 清理名称
func (m *MetaBase) CleanupName() string {
	name := m.Name
	
	// 移除文件扩展名
	extIndex := strings.LastIndex(name, ".")
	if extIndex > 0 {
		name = name[:extIndex]
	}
	
	// 移除常见的额外信息
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[.*?\]`),      // 方括号内容
		regexp.MustCompile(`\(.*?\)`),      // 圆括号内容
		regexp.MustCompile(`\{.*?\}`),      // 花括号内容
		regexp.MustCompile(`\s+\d{3,4}p\s*`), // 分辨率
		regexp.MustCompile(`\s+[xh]26[45]\s*`), // 编码
		regexp.MustCompile(`\s+hevc\s*`),     // HEVC编码
		regexp.MustCompile(`\s+aac\s*`),      // 音频编码
		regexp.MustCompile(`\s+dts\s*`),      // DTS音频
		regexp.MustCompile(`\s+flac\s*`),     // FLAC音频
		regexp.MustCompile(`\s+\d+\.\d+ch\s*`), // 音频声道
		regexp.MustCompile(`\s+web-?dl\s*`),   // Web-DL标记
		regexp.MustCompile(`\s+bdrip\s*`),     // BDrip标记
		regexp.MustCompile(`\s+bluray\s*`),    // BluRay标记
		regexp.MustCompile(`\s+hdrip\s*`),     // HDRip标记
		regexp.MustCompile(`\s+dvdrip\s*`),    // DVDRip标记
	}
	
	for _, pattern := range patterns {
		name = pattern.ReplaceAllString(name, "")
	}
	
	// 移除多余的空格
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	
	return name
}

// ExtractLanguageFromName 从名称中提取语言
func (m *MetaBase) ExtractLanguageFromName() string {
	name := strings.ToLower(m.Name)
	
	languagePatterns := map[string]string{
		"chinese": "Chinese",
		"中字":    "Chinese",
		"国语":    "Chinese",
		"双语":    "Chinese-English",
		"双语字幕":  "Chinese-English",
		"eng":     "English",
		"英文":    "English",
		"japanese": "Japanese",
		"日语":    "Japanese",
		"korean":  "Korean",
		"韩语":    "Korean",
		"multi":   "Multi",
		"多语言":   "Multi",
	}
	
	for pattern, language := range languagePatterns {
		if strings.Contains(name, pattern) {
			return language
		}
	}
	
	return "Unknown"
}

// ExtractResolution 从名称中提取分辨率
func (m *MetaBase) ExtractResolution() string {
	name := strings.ToLower(m.Name)
	
	resolutions := []struct {
		Pattern    string
		Resolution string
	}{
		{"8k", Resolution8K},
		{"4k", Resolution4K},
		{"2160p", Resolution4K},
		{"4096p", Resolution4K},
		{"2k", Resolution2K},
		{"1440p", Resolution2K},
		{"1080p", Resolution1080p},
		{"720p", Resolution720p},
		{"480p", Resolution480p},
	}
	
	for _, res := range resolutions {
		if strings.Contains(name, res.Pattern) {
			return res.Resolution
		}
	}
	
	return ResolutionUnknown
}

// ExtractVideoCodec 从名称中提取视频编码
func (m *MetaBase) ExtractVideoCodec() string {
	name := strings.ToLower(m.Name)
	
	codecs := []struct {
		Pattern string
		Codec   string
	}{
		{"av1", VideoCodecAV1},
		{"vp9", VideoCodecVP9},
		{"h265", VideoCodecH265},
		{"hevc", VideoCodecH265},
		{"x265", VideoCodecX265},
		{"h264", VideoCodecH264},
		{"avc", VideoCodecH264},
		{"x264", VideoCodecX264},
	}
	
	for _, codec := range codecs {
		if strings.Contains(name, codec.Pattern) {
			return codec.Codec
		}
	}
	
	return VideoCodecUnknown
}

// ExtractAudioCodec 从名称中提取音频编码
func (m *MetaBase) ExtractAudioCodec() string {
	name := strings.ToLower(m.Name)
	
	codecs := []struct {
		Pattern string
		Codec   string
	}{
		{"dts:x", AudioCodecDTSX},
		{"dts:x", AudioCodecDTSX},
		{"atmos", AudioCodecATMOS},
		{"truehd", AudioCodecTrueHD},
		{"dts-hd", AudioCodecDTS},
		{"dts", AudioCodecDTS},
		{"flac", AudioCodecFLAC},
		{"aac", AudioCodecAAC},
		{"mp3", AudioCodecMP3},
	}
	
	for _, codec := range codecs {
		if strings.Contains(name, codec.Pattern) {
			return codec.Codec
		}
	}
	
	return AudioCodecUnknown
}

// ExtractAudioChannels 从名称中提取音频声道
func (m *MetaBase) ExtractAudioChannels() string {
	name := strings.ToLower(m.Name)
	
	channels := []struct {
		Pattern  string
		Channels string
	}{
		{"7.1", "7.1"},
		{"5.1", "5.1"},
		{"2.0", "2.0"},
		{"2.1", "2.1"},
		{"6ch", "5.1"},
		{"8ch", "7.1"},
	}
	
	for _, ch := range channels {
		if strings.Contains(name, ch.Pattern) {
			return ch.Channels
		}
	}
	
	return ""
}

// SetDefaultValues 设置默认值
func (m *MetaBase) SetDefaultValues() {
	// 如果年份未设置，尝试从名称或标题中解析
	if m.Year == 0 {
		year := m.ParseYearFromName()
		if year == 0 {
			year = m.ParseYearFromTitle()
		}
		m.Year = year
	}
	
	// 如果未设置语言，尝试从名称中提取
	if m.Language == "" {
		m.Language = m.ExtractLanguageFromName()
	}
	
	// 如果未设置分辨率，尝试从名称中提取
	if m.Resolution == ResolutionUnknown {
		m.Resolution = m.ExtractResolution()
	}
	
	// 如果未设置视频编码，尝试从名称中提取
	if m.VideoCodec == VideoCodecUnknown {
		m.VideoCodec = m.ExtractVideoCodec()
	}
	
	// 如果未设置音频编码，尝试从名称中提取
	if m.AudioCodec == AudioCodecUnknown {
		m.AudioCodec = m.ExtractAudioCodec()
	}
	
	// 如果未设置音频声道，尝试从名称中提取
	if m.AudioChannels == "" {
		m.AudioChannels = m.ExtractAudioChannels()
	}
	
	// 如果标题未设置，尝试清理名称作为标题
	if m.Title == "" {
		m.Title = m.CleanupName()
	}
	
	// 更新解析状态
	if m.Title != "" {
		m.ParseStatus = ParseStatusSuccess
		m.Confidence = 1.0
	} else {
		m.ParseStatus = ParseStatusFailed
		m.Confidence = 0.0
	}
}

// FormatSize 格式化文件大小
func (m *MetaBase) FormatSize() string {
	size := m.FileSize
	if size < 0 {
		return "未知"
	}
	
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	
	return fmt.Sprintf("%.2f %s", float64(m.FileSize)/(1<<(10*uint(i))), units[i])
}

// GetMediaInfoSummary 获取媒体信息摘要
func (m *MetaBase) GetMediaInfoSummary() string {
	parts := []string{}
	
	if m.Resolution != ResolutionUnknown {
		parts = append(parts, m.Resolution)
	}
	
	if m.VideoCodec != VideoCodecUnknown {
		parts = append(parts, m.VideoCodec)
	}
	
	if m.AudioCodec != AudioCodecUnknown {
		audioInfo := m.AudioCodec
		if m.AudioChannels != "" {
			audioInfo += " " + m.AudioChannels
		}
		parts = append(parts, audioInfo)
	}
	
	if m.FileSize > 0 {
		parts = append(parts, m.FormatSize())
	}
	
	return strings.Join(parts, " ")
}