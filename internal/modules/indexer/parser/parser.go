// Package parser 索引器Parser包
package parser

import (
	"context"
	"fmt"
	"strings"

	"goquery.org/document"
)

// Parser 种子信息解析器接口
type Parser interface {
	// Name 获取Parser名称
	Name() string

	// IsSupported 检查是否支持该站点
	IsSupported(url string) bool

	// ParseTorrentPage 解析种子列表页
	ParseTorrentPage(ctx context.Context, htmlContent string, page int) ([]*TorrentInfo, error)

	// ParseTorrentDetail 解析种子详情页
	ParseTorrentDetail(ctx context.Context, htmlContent string) (*TorrentDetail, error)

	// ParseSearchResult 解析搜索结果页
	ParseSearchResult(ctx context.Context, htmlContent string, keyword string, page int) ([]*TorrentInfo, error)

	// ParseUserPage 解析用户页
	ParseUserPage(ctx context.Context, htmlContent string, userID string) (*UserInfo, error)

	// GetSiteInfo 获取站点信息
	GetSiteInfo() *SiteInfo
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	ID              string            `json:"id"`              // 种子ID
	Title           string            `json:"title"`           // 标题
	Category        string            `json:"category"`        // 分类
	SubCategory     string            `json:"sub_category"`     // 子分类
	URL             string            `json:"url"`             // 详情页URL
	DownloadURL     string            `json:"download_url"`     // 下载URL
	Size            int64             `json:"size"`            // 大小
	Seeders         int               `json:"seeders"`         // 做种数
	Leechers        int               `json:"leechers"`        // 下载数
	Downloads       int               `json:"downloads"`       // 下载次数
	UploadedAt      string            `json:"uploaded_at"`      // 上传时间
	TTL             string            `json:"ttl"`             // 剩余时间
	Promotional     bool              `json:"promotional"`     // 是否推广
	FreeLeech       bool              `json:"free_leech"`       // 是否免费
	HDR             bool              `json:"hdr"`             // 是否HDR
	Dubbed          bool              `json:"dubbed"`          // 是否配音
	Subtitled       bool              `json:"subtitled"`       // 是否字幕
	UploadFactor    float64           `json:"upload_factor"`    // 上传倍率
	DownloadFactor  float64           `json:"download_factor"`  // 下载倍率
	Comments        int               `json:"comments"`        // 评论数
	IMDBID          string            `json:"imdb_id"`         // IMDB ID
	TMDBID          string            `json:"tmdb_id"`         // TMDB ID
	ReleaseGroup    string            `json:"release_group"`    // 发布组
	MediaInfo       *MediaInfo        `json:"media_info"`       // 媒体信息
	Tags            []string          `json:"tags"`            // 标签
	Meta            map[string]string `json:"meta"`            // 元数据
}

// TorrentDetail 种子详情
type TorrentDetail struct {
	ID              string               `json:"id"`              // 种子ID
	Title           string               `json:"title"`           // 标题
	Description     string               `json:"description"`     // 描述
	Category        string               `json:"category"`        // 分类
	SubCategory     string               `json:"sub_category"`     // 子分类
	URL             string               `json:"url"`             // 详情页URL
	DownloadURL     string               `json:"download_url"`     // 下载URL
	MagnetURL       string               `json:"magnet_url"`       // 磁力链接
	Size            int64                `json:"size"`            // 大小
	Seeders         int                  `json:"seeders"`         // 做种数
	Leechers        int                  `json:"leechers"`        // 下载数
	Downloads       int                  `json:"downloads"`       // 下载次数
	UploadedAt      string               `json:"uploaded_at"`      // 上传时间
	TTL             string               `json:"ttl"`             // 剩余时间
	Promotional     bool                 `json:"promotional"`     // 是否推广
	FreeLeech       bool                 `json:"free_leech"`       // 是否免费
	HDR             bool                 `json:"hdr"`             // 是否HDR
	Dubbed          bool                 `json:"dubbed"`          // 是否配音
	Subtitled       bool                 `json:"subtitled"`       // 是否字幕
	UploadFactor    float64              `json:"upload_factor"`    // 上传倍率
	DownloadFactor  float64              `json:"download_factor"`  // 下载倍率
	Comments        int                  `json:"comments"`        // 评论数
	IMDBID          string               `json:"imdb_id"`         // IMDB ID
	TMDBID          string               `json:"tmdb_id"`         // TMDB ID
	TVDBID          string               `json:"tvdb_id"`         // TVDB ID
	ReleaseGroup    string               `json:"release_group"`    // 发布组
	MediaInfo       *MediaInfo           `json:"media_info"`       // 媒体信息
	Tags            []string             `json:"tags"`            // 标签
	Meta            map[string]string    `json:"meta"`            // 元数据
	Files           []*TorrentFile       `json:"files"`           // 文件列表
	Screenshots      []string             `json:"screenshots"`      // 截图
	TechnicalInfo   *TechnicalInfo       `json:"technical_info"`   // 技术信息
	CommentsList    []*Comment           `json:"comments_list"`    // 评论列表
}

// TorrentFile 种子文件
type TorrentFile struct {
	Path      string `json:"path"`      // 文件路径
	Size      int64  `json:"size"`      // 文件大小
	Extension string `json:"extension"` // 文件扩展名
}

// Comment 评论
type Comment struct {
	ID        string `json:"id"`        // 评论ID
	UserID    string `json:"user_id"`  // 用户ID
	Username  string `json:"username"`  // 用户名
	Avatar    string `json:"avatar"`    // 头像
	Content   string `json:"content"`   // 评论内容
	CreatedAt string `json:"created_at"` // 创建时间
	ReplyTo   string `json:"reply_to"`  // 回复评论ID
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Type         string   `json:"type"`         // 媒体类型
	Title        string   `json:"title"`        // 媒体标题
	OriginalTitle string   `json:"original_title"` // 原标题
	Year         int      `json:"year"`         // 年份
	Season       int      `json:"season"`       // 季数
	Episode      int      `json:"episode"`      // 集数
	IMDBID       string   `json:"imdb_id"`     // IMDB ID
	TMDBID       string   `json:"tmdb_id"`     // TMDB ID
	TVDBID       string   `json:"tvdb_id"`     // TVDB ID
	Overview     string   `json:"overview"`     // 简介
	Genres       []string `json:"genres"`       // 类型
	Poster       string   `json:"poster"`       // 海报
	Backdrop     string   `json:"backdrop"`     // 背景图
	Rating       float64  `json:"rating"`       // 评分
	Runtime      int      `json:"runtime"`      // 时长
	ReleaseDate  string   `json:"release_date"`  // 发布日期
	Status       string   `json:"status"`       // 状态
	Network      string   `json:"network"`      // 电视网
	Language     string   `json:"language"`     // 语言
	Country      string   `json:"country"`      // 国家
	Director     []string `json:"director"`     // 导演
	Writer       []string `json:"writer"`       // 编剧
	Actors       []string `json:"actors"`       // 演员
	Subtitles    []string `json:"subtitles"`    // 字幕
}

// TechnicalInfo 技术信息
type TechnicalInfo struct {
	Container    string `json:"container"`    // 容器格式
	VideoCodec   string `json:"video_codec"`  // 视频编码
	VideoBitrate int64  `json:"video_bitrate"` // 视频码率
	Resolution   string `json:"resolution"`   // 分辨率
	FrameRate    string `json:"frame_rate"`   // 帧率
	Aspect       string `json:"aspect"`       // 宽高比
	AudioCodec   string `json:"audio_codec"`  // 音频编码
	AudioBitrate int64  `json:"audio_bitrate"` // 音频码率
	Channels     string `json:"channels"`     // 声道数
	SampleRate   string `json:"sample_rate"`   // 采样率
	Subtitle     string `json:"subtitle"`     // 字幕格式
	Source       string `json:"source"`       // 视频来源
	Encoding     string `json:"encoding"`     // 编码设置
}

// UserInfo 用户信息
type UserInfo struct {
	ID            string   `json:"id"`            // 用户ID
	Username      string   `json:"username"`      // 用户名
	Email         string   `json:"email"`         // 邮箱
	Avatar        string   `json:"avatar"`        // 头像
	Class         string   `json:"class"`         // 用户等级
	Rank          string   `json:"rank"`          // 用户等级
	Uploaded      int64    `json:"uploaded"`      // 上传量
	Downloaded    int64    `json:"downloaded"`    // 下载量
	Ratio         float64  `json:"ratio"`         // 分享率
	SeedingPoints float64  `json:"seeding_points"` // 做种积分
	BonusPoints   float64  `json:"bonus_points"`   // 魔力值
	Invites       int      `json:"invites"`       // 邀请数量
	JoinedAt      string   `json:"joined_at"`     // 加入时间
	LastActive    string   `json:"last_active"`   // 最后活跃时间
	UploadSpeed   int64    `json:"upload_speed"`   // 上传速度
	DownloadSpeed int64    `json:"download_speed"` // 下载速度
	Passkey       string   `json:"passkey"`       // Passkey
	Signature     string   `json:"signature"`     // 签名
	ProfileURL    string   `json:"profile_url"`   // 个人资料URL
	TorrentsCount int      `json:"torrents_count"` // 种子数量
	PerfectCount  int      `json:"perfect_count"`  // 完美种数量
	Meta          map[string]string `json:"meta"` // 元数据
}

// SiteInfo 站点信息
type SiteInfo struct {
	Name         string `json:"name"`         // 站点名称
	Domain       string `json:"domain"`       // 站点域名
	Charset      string `json:"charset"`      // 字符编码
	Timezone     string `json:"timezone"`     // 时区
	Description  string `json:"description"`  // 描述
	Features     []string `json:"features"`     // 特性
	Tags         []string `json:"tags"`         // 标签
	Categories   []string `json:"categories"`   // 分类
	Resolutions  []string `json:"resolutions"`  // 分辨率
	VideoCodecs  []string `json:"video_codecs"`  // 视频编码
	AudioCodecs  []string `json:"audio_codecs"`  // 音频编码
	Containers   []string `json:"containers"`    // 容器格式
	Sources      []string `json:"sources"`       // 视频来源
}

// BaseParser 基础解析器
type BaseParser struct {
	name     string
	domain   string
	siteInfo *SiteInfo
}

// NewBaseParser 创建基础解析器
func NewBaseParser(name, domain string) *BaseParser {
	return &BaseParser{
		name:   name,
		domain: domain,
		siteInfo: &SiteInfo{
			Name:    name,
			Domain:  domain,
			Charset: "UTF-8",
			Tags:    []string{},
		},
	}
}

// Name 获取Parser名称
func (p *BaseParser) Name() string {
	return p.name
}

// GetSiteInfo 获取站点信息
func (p *BaseParser) GetSiteInfo() *SiteInfo {
	return p.siteInfo
}

// IsSupported 检查是否支持该站点
func (p *BaseParser) IsSupported(url string) bool {
	return strings.Contains(strings.ToLower(url), strings.ToLower(p.domain))
}

// CleanText 清理文本
func (p *BaseParser) CleanText(text string) string {
	// 移除多余的空格和换行
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	
	// 移除多个连续空格
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	
	return text
}

// ExtractNumber 从字符串中提取数字
func (p *BaseParser) ExtractNumber(text string) int {
	numStr := ""
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else if numStr != "" {
			break
		}
	}
	
	if numStr == "" {
		return 0
	}
	
	result := 0
	fmt.Sscanf(numStr, "%d", &result)
	return result
}

// ExtractSize 从字符串中提取文件大小
func (p *BaseParser) ExtractSize(text string) int64 {
	text = strings.ToUpper(strings.TrimSpace(text))
	
	var size float64
	var unit string
	
	// 尝试解析 "数字+单位" 格式
	n, err := fmt.Sscanf(text, "%f%s", &size, &unit)
	if n < 2 || err != nil {
		// 尝试纯数字格式
		if size, err = fmt.Sscanf(text, "%f", &size); err != nil {
			return 0
		}
		return int64(size)
	}
	
	multiplier := int64(1)
	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "KIB":
		multiplier = 1024
	case "MIB":
		multiplier = 1024 * 1024
	case "GIB":
		multiplier = 1024 * 1024 * 1024
	case "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	
	return int64(size * float64(multiplier))
}

// ParseRating 解析评分
func (p *BaseParser) ParseRating(text string) float64 {
	text = strings.TrimSpace(strings.ReplaceAll(text, ",", "."))
	
	var rating float64
	_, err := fmt.Sscanf(text, "%f", &rating)
	if err != nil {
		return 0.0
	}
	
	// 限制评分范围
	if rating > 10.0 {
		rating = 10.0
	} else if rating < 0.0 {
		rating = 0.0
	}
	
	return rating
}

// ParseTime 解析时间
func (p *BaseParser) ParseTime(text string) string {
	// 简单的时间解析，实际实现需要更复杂的逻辑
	text = strings.TrimSpace(text)
	
	// 处理相对时间
	if strings.Contains(text, "前") {
		return text // 直接返回相对时间
	}
	
	// 处理标准时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}
	
	for _, format := range formats {
		if _, err := fmt.Parse(format, text); err == nil {
			return text
		}
	}
	
	return text
}

// GetAttr 获取元素属性
func (p *BaseParser) GetAttr(s *goquery.Selection, attrName string) string {
	if s.Length() == 0 {
		return ""
	}
	
	if attr, exists := s.Attr(attrName); exists {
		return attr
	}
	
	return ""
}

// GetText 获取元素文本
func (p *BaseParser) GetText(s *goquery.Selection) string {
	if s.Length() == 0 {
		return ""
	}
	
	return p.CleanText(s.Text())
}

// GetHTML 获取元素HTML
func (p *BaseParser) GetHTML(s *goquery.Selection) string {
	if s.Length() == 0 {
		return ""
	}
	
	html, err := s.Html()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(html)
}

// ParseTags 解析标签
func (p *BaseParser) ParseTags(tagStr string) []string {
	if tagStr == "" {
		return []string{}
	}
	
	// 支持多种分隔符
	delimiters := []string{",", ";", " ", "|", "/"}
	
	for _, delim := range delimiters {
		if strings.Contains(tagStr, delim) {
			tags := strings.Split(tagStr, delim)
			result := make([]string, 0, len(tags))
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					result = append(result, tag)
				}
			}
			return result
		}
	}
	
	return []string{strings.TrimSpace(tagStr)}
}

// Default implementations - 默认实现返回错误
func (p *BaseParser) ParseTorrentPage(ctx context.Context, htmlContent string, page int) ([]*TorrentInfo, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (p *BaseParser) ParseTorrentDetail(ctx context.Context, htmlContent string) (*TorrentDetail, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (p *BaseParser) ParseSearchResult(ctx context.Context, htmlContent string, keyword string, page int) ([]*TorrentInfo, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (p *BaseParser) ParseUserPage(ctx context.Context, htmlContent string, userID string) (*UserInfo, error) {
	return nil, fmt.Errorf("method not implemented")
}