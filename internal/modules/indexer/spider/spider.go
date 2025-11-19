// Package spider 索引器Spider包
package spider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/helper"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// Spider 索引器Spider接口
type Spider interface {
	// Name 获取Spider名称
	Name() string

	// SiteURL 获取站点URL
	SiteURL() string

	// IsSupported 检查是否支持该站点
	IsSupported(url string) bool

	// Search 搜索种子
	Search(ctx context.Context, keyword string, filters *SearchFilters) ([]*TorrentItem, error)

	// GetUserTorrents 获取用户种子列表
	GetUserTorrents(ctx context.Context, userID string) ([]*TorrentItem, error)

	// GetTorrentDetails 获取种子详情
	GetTorrentDetails(ctx context.Context, torrentID string) (*TorrentDetail, error)

	// DownloadTorrent 下载种子文件
	DownloadTorrent(ctx context.Context, torrentID string) ([]byte, error)

	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context) (*UserInfo, error)
}

// SearchFilters 搜索过滤器
type SearchFilters struct {
	Category     string            `json:"category"`     // 分类：电影、电视剧、动漫等
	Resolution   string            `json:"resolution"`   // 分辨率：1080p、4K等
	VideoCodec    string            `json:"video_codec"`  // 视频编码：x264、x265等
	AudioCodec    string            `json:"audio_codec"`  // 音频编码：AAC、DTS等
	Source        string            `json:"source"`       // 来源：BluRay、Web等
	Container     string            `json:"container"`    // 容器：MKV、MP4等
	SizeRange     *SizeRange        `json:"size_range"`   // 大小范围
	Promotional   bool              `json:"promotional"`  // 是否包含推广
	FreeLeech     bool              `json:"free_leech"`  // 是否免费
	HDR           bool              `json:"hdr"`          // 是否HDR
	Dubbed        bool              `json:"dubbed"`       // 是否配音
	Subtitled     bool              `json:"subtitled"`    // 是否字幕
	Season        int               `json:"season"`       // 季数
	Episode       int               `json:"episode"`      // 集数
	Year          int               `json:"year"`         // 年份
	MinSeeders    int               `json:"min_seeders"`  // 最少做种数
	MinLeechers   int               `json:"min_leechers"` // 最少下载数
	ExcludeWords []string          `json:"exclude_words"` // 排除关键词
	IncludeWords []string          `json:"include_words"` // 包含关键词
	SortBy        string            `json:"sort_by"`       // 排序方式：time、size、seeders、leechers
	SortOrder     string            `json:"sort_order"`    // 排序顺序：asc、desc
	Page          int               `json:"page"`          // 页码
	Limit         int               `json:"limit"`         // 每页数量
	Custom        map[string]string `json:"custom"`        // 自定义过滤条件
}

// SizeRange 大小范围
type SizeRange struct {
	Min int64 `json:"min"` // 最小大小（字节）
	Max int64 `json:"max"` // 最大大小（字节）
}

// TorrentItem 种子项目
type TorrentItem struct {
	ID           string            `json:"id"`             // 种子ID
	Title        string            `json:"title"`          // 标题
	Category     string            `json:"category"`       // 分类
	SubCategory  string            `json:"sub_category"`   // 子分类
	Size         int64             `json:"size"`           // 大小（字节）
	Seeders      int               `json:"seeders"`        // 做种数
	Leechers     int               `json:"leechers"`       // 下载数
	Downloads    int               `json:"downloads"`      // 下载次数
	UploadedAt   time.Time         `json:"uploaded_at"`    // 上传时间
	TTL          time.Duration     `json:"ttl"`           // 剩余时间
	Promotional  bool              `json:"promotional"`    // 是否推广
	FreeLeech    bool              `json:"free_leech"`    // 是否免费
	HDR          bool              `json:"hdr"`           // 是否HDR
	Dubbed       bool              `json:"dubbed"`        // 是否配音
	Subtitled    bool              `json:"subtitled"`     // 是否字幕
	UploadFactor float64           `json:"upload_factor"` // 上传倍率
	DownloadFactor float64          `json:"download_factor"` // 下载倍率
	Comments     int               `json:"comments"`      // 评论数
	IMDBID       string            `json:"imdb_id"`       // IMDB ID
	TMDBID       string            `json:"tmdb_id"`       // TMDB ID
	Details      *TorrentDetail    `json:"details"`       // 详细信息
	Meta         map[string]string `json:"meta"`          // 元数据
}

// TorrentDetail 种子详情
type TorrentDetail struct {
	ID              string            `json:"id"`                // 种子ID
	Title           string            `json:"title"`             // 标题
	Description     string            `json:"description"`       // 描述
	Category        string            `json:"category"`          // 分类
	SubCategory     string            `json:"sub_category"`     // 子分类
	Size            int64             `json:"size"`              // 大小（字节）
	Seeders         int               `json:"seeders"`           // 做种数
	Leechers        int               `json:"leechers"`          // 下载数
	Downloads       int               `json:"downloads"`         // 下载次数
	UploadedAt      time.Time         `json:"uploaded_at"`       // 上传时间
	TTL             time.Duration     `json:"ttl"`              // 剩余时间
	Promotional     bool              `json:"promotional"`       // 是否推广
	FreeLeech       bool              `json:"free_leech"`       // 是否免费
	HDR             bool              `json:"hdr"`              // 是否HDR
	Dubbed          bool              `json:"dubbed"`           // 是否配音
	Subtitled       bool              `json:"subtitled"`        // 是否字幕
	UploadFactor    float64           `json:"upload_factor"`     // 上传倍率
	DownloadFactor  float64           `json:"download_factor"`   // 下载倍率
	Comments        int               `json:"comments"`         // 评论数
	IMDBID          string            `json:"imdb_id"`          // IMDB ID
	TMDBID          string            `json:"tmdb_id"`          // TMDB ID
	TVRageID        string            `json:"tvrage_id"`        // TVRage ID
	Files           []*TorrentFile    `json:"files"`            // 文件列表
	Pictures        []string          `json:"pictures"`         // 图片列表
	CommentsList    []*Comment        `json:"comments_list"`    // 评论列表
	TorrentURL      string            `json:"torrent_url"`      // 种子文件URL
	MagnetURL       string            `json:"magnet_url"`       // 磁力链接
	ReleaseGroup    string            `json:"release_group"`    // 发布组
	MediaInfo       *MediaInfo        `json:"media_info"`       // 媒体信息
	Meta            map[string]interface{} `json:"meta"`        // 元数据
}

// TorrentFile 种子文件
type TorrentFile struct {
	Path     string `json:"path"`     // 文件路径
	Size     int64  `json:"size"`     // 文件大小
	Modified int64  `json:"modified"` // 修改时间
}

// Comment 评论
type Comment struct {
	ID        string    `json:"id"`         // 评论ID
	UserID    string    `json:"user_id"`    // 用户ID
	Username  string    `json:"username"`   // 用户名
	Content   string    `json:"content"`    // 评论内容
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Type        string   `json:"type"`         // 媒体类型：movie、tv、anime
	Title       string   `json:"title"`        // 标题
	OriginalTitle string  `json:"original_title"` // 原标题
	Year        int      `json:"year"`         // 年份
	Season      int      `json:"season"`       // 季数
	Episode     int      `json:"episode"`      // 集数
	IMDBID      string   `json:"imdb_id"`     // IMDB ID
	TMDBID      string   `json:"tmdb_id"`     // TMDB ID
	TVDBID      string   `json:"tvdb_id"`     // TVDB ID
	Overview    string   `json:"overview"`     // 简介
	Genres      []string `json:"genres"`       // 类型
	Poster      string   `json:"poster"`       // 海报
	Backdrop    string   `json:"backdrop"`     // 背景图
	Rating      float64  `json:"rating"`       // 评分
	Runtime     int      `json:"runtime"`      // 时长（分钟）
	ReleaseDate string   `json:"release_date"` // 发布日期
	Status      string   `json:"status"`       // 状态
	Network     string   `json:"network"`     // 电视网
	Language    string   `json:"language"`     // 语言
	Subtitles   []string `json:"subtitles"`    // 字幕
}

// UserInfo 用户信息
type UserInfo struct {
	ID            string    `json:"id"`             // 用户ID
	Username      string    `json:"username"`       // 用户名
	Email         string    `json:"email"`          // 邮箱
	Avatar        string    `json:"avatar"`         // 头像
	Class         string    `json:"class"`          // 用户等级
	Uploaded      int64     `json:"uploaded"`       // 上传量
	Downloaded    int64     `json:"downloaded"`     // 下载量
	Ratio         float64   `json:"ratio"`          // 分享率
	SeedingPoints float64   `json:"seeding_points"` // 做种积分
	Invites       int       `json:"invites"`        // 邀请数量
	JoinedAt      time.Time `json:"joined_at"`      // 加入时间
	LastActive    time.Time `json:"last_active"`    // 最后活跃时间
	UploadSpeed   int64     `json:"upload_speed"`   // 上传速度
	DownloadSpeed int64     `json:"download_speed"` // 下载速度
	BonusPoints   float64   `json:"bonus_points"`   // 魔力值
	Passkey       string    `json:"passkey"`        // Passkey
	Meta          map[string]interface{} `json:"meta"` // 元数据
}

// BaseSpider 基础Spider
type BaseSpider struct {
	name       string
	siteURL    string
	client     *http.Client
	cookieJar  *helper.CookieJar
	userAgent  string
	timeout    time.Duration
	retryCount int
}

// NewBaseSpider 创建基础Spider
func NewBaseSpider(name, siteURL string) *BaseSpider {
	return &BaseSpider{
		name:     name,
		siteURL:  siteURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		timeout:     30 * time.Second,
		retryCount:  3,
	}
}

// Name 获取Spider名称
func (b *BaseSpider) Name() string {
	return b.name
}

// SiteURL 获取站点URL
func (b *BaseSpider) SiteURL() string {
	return b.siteURL
}

// IsSupported 检查是否支持该站点
func (b *BaseSpider) IsSupported(url string) bool {
	return strings.Contains(url, b.siteURL)
}

// SetTimeout 设置超时时间
func (b *BaseSpider) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
	b.client.Timeout = timeout
}

// SetUserAgent 设置User-Agent
func (b *BaseSpider) SetUserAgent(userAgent string) {
	b.userAgent = userAgent
}

// SetCookieJar 设置Cookie管理器
func (b *BaseSpider) SetCookieJar(cookieJar *helper.CookieJar) {
	b.cookieJar = cookieJar
}

// GetHTTPClient 获取HTTP客户端
func (b *BaseSpider) GetHTTPClient() *http.Client {
	return b.client
}

// LogRequest 记录请求日志
func (b *BaseSpider) LogRequest(method, url string, statusCode int, duration time.Duration) {
	logger.Debug("Spider请求",
		zap.String("spider", b.name),
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status", statusCode),
		zap.Duration("duration", duration))
}

// LogError 记录错误日志
func (b *BaseSpider) LogError(operation string, err error, context map[string]interface{}) {
	fields := []zap.Field{
		zap.String("spider", b.name),
		zap.String("operation", operation),
		zap.Error(err),
	}
	
	for k, v := range context {
		fields = append(fields, zap.Any(k, v))
	}
	
	logger.Error("Spider错误", fields...)
}

// Retry 带重试的操作
func (b *BaseSpider) Retry(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for i := 0; i < b.retryCount; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i) * time.Second):
			}
		}
		
		if err := operation(); err != nil {
			lastErr = err
			logger.Warn("Spider操作失败，准备重试",
				zap.String("spider", b.name),
				zap.Int("attempt", i+1),
				zap.Error(err))
			continue
		}
		
		return nil
	}
	
	return fmt.Errorf("操作失败，重试%d次后仍然失败: %w", b.retryCount, lastErr)
}

// ValidateSearchFilters 验证搜索过滤器
func (b *BaseSpider) ValidateSearchFilters(filters *SearchFilters) error {
	if filters == nil {
		return nil
	}
	
	// 验证大小范围
	if filters.SizeRange != nil {
		if filters.SizeRange.Min < 0 || filters.SizeRange.Max < 0 {
			return fmt.Errorf("大小范围不能为负数")
		}
		if filters.SizeRange.Min > filters.SizeRange.Max {
			return fmt.Errorf("最小大小不能大于最大大小")
		}
	}
	
	// 验证年份
	if filters.Year < 1900 || filters.Year > time.Now().Year()+5 {
		return fmt.Errorf("年份无效")
	}
	
	// 验证季集
	if filters.Season < 0 || filters.Episode < 0 {
		return fmt.Errorf("季集数不能为负数")
	}
	
	// 验证分页
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}
	
	return nil
}

// NormalizeKeyword 标准化关键词
func (b *BaseSpider) NormalizeKeyword(keyword string) string {
	// 移除多余的空格和特殊字符
	keyword = strings.TrimSpace(keyword)
	keyword = strings.ReplaceAll(keyword, "  ", " ")
	
	// 转换为小写（部分站点可能需要）
	// keyword = strings.ToLower(keyword)
	
	return keyword
}

// BuildURL 构建完整URL
func (b *BaseSpider) BuildURL(path string, params map[string]string) string {
	baseURL := b.siteURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	fullURL := baseURL + path
	
	if len(params) > 0 {
		values := make([]string, 0, len(params))
		for k, v := range params {
			values = append(values, fmt.Sprintf("%s=%s", k, v))
		}
		fullURL += "?" + strings.Join(values, "&")
	}
	
	return fullURL
}

// ParseSize 解析文件大小
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	
	// 移除空格
	sizeStr = strings.ReplaceAll(sizeStr, " ", "")
	
	var multiplier float64 = 1
	var numberStr string
	
	if strings.Contains(sizeStr, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		numberStr = strings.ReplaceAll(sizeStr, "TB", "")
	} else if strings.Contains(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numberStr = strings.ReplaceAll(sizeStr, "GB", "")
	} else if strings.Contains(sizeStr, "MB") {
		multiplier = 1024 * 1024
		numberStr = strings.ReplaceAll(sizeStr, "MB", "")
	} else if strings.Contains(sizeStr, "KB") {
		multiplier = 1024
		numberStr = strings.ReplaceAll(sizeStr, "KB", "")
	} else if strings.Contains(sizeStr, "B") {
		multiplier = 1
		numberStr = strings.ReplaceAll(sizeStr, "B", "")
	} else {
		numberStr = sizeStr
	}
	
	// 解析数字
	var size float64
	if _, err := fmt.Sscanf(numberStr, "%f", &size); err != nil {
		return 0, fmt.Errorf("解析大小失败: %w", err)
	}
	
	return int64(size * multiplier), nil
}

// FormatSize 格式化文件大小
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CalculateExpiration 计算过期时间
func CalculateExpiration(uploadedAt time.Time, ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 0
	}
	
	elapsed := time.Since(uploadedAt)
	remaining := ttl - elapsed
	
	if remaining < 0 {
		return 0
	}
	
	return remaining
}