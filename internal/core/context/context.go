// Package context 上下文管理模块
package context

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// ContextKeyUser 用户信息键
	ContextKeyUser ContextKey = "user"
	// ContextKeyRequest 请求信息键
	ContextKeyRequest ContextKey = "request"
	// ContextKeyTorrent 种子信息键
	ContextKeyTorrent ContextKey = "torrent"
	// ContextKeyMedia 媒体信息键
	ContextKeyMedia ContextKey = "media"
	// ContextKeyCache 缓存信息键
	ContextKeyCache ContextKey = "cache"
	// ContextKeySession 会话信息键
	ContextKeySession ContextKey = "session"
)

// RequestContext 请求上下文
type RequestContext struct {
	ID        string            `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Query     map[string]string `json:"query"`
	Body      []byte            `json:"body,omitempty"`
	ClientIP  string            `json:"client_ip"`
	UserAgent string            `json:"user_agent"`
	StartTime time.Time         `json:"start_time"`
}

// NewRequestContext 创建请求上下文
func NewRequestContext(c *gin.Context) *RequestContext {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	query := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	return &RequestContext{
		ID:        generateRequestID(),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Headers:   headers,
		Query:     query,
		ClientIP:  c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		StartTime: time.Now(),
	}
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	// 站点信息
	SiteID       int    `json:"site_id"`
	SiteName     string `json:"site_name"`
	SiteCookie   string `json:"site_cookie,omitempty"`
	SiteUA       string `json:"site_ua,omitempty"`
	SiteProxy    bool   `json:"site_proxy"`
	SiteOrder    int    `json:"site_order"`
	SiteDownloader string `json:"site_downloader,omitempty"`

	// 种子基本信息
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	IMDBID      string `json:"imdb_id,omitempty"`
	Enclosure   string `json:"enclosure"`
	PageURL     string `json:"page_url,omitempty"`

	// 种子属性
	Size     float64 `json:"size"`
	Seeders  int     `json:"seeders"`
	Peers    int     `json:"peers"`
	Grabs    int     `json:"grabs"`
	PubDate  string  `json:"pub_date,omitempty"`
	DateElapsed string `json:"date_elapsed,omitempty"`

	// 媒体信息
	MediaType   string `json:"media_type,omitempty"`
	TMDBID      int    `json:"tmdb_id,omitempty"`
	DoubanID    string `json:"douban_id,omitempty"`
	Season      int    `json:"season,omitempty"`
	Episode     int    `json:"episode,omitempty"`
	Year        int    `json:"year,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
	VideoCodec  string `json:"video_codec,omitempty"`
	AudioCodec  string `json:"audio_codec,omitempty"`
	Source      string `json:"source,omitempty"`
	ReleaseGroup string `json:"release_group,omitempty"`

	// 下载信息
	DownloadPath string `json:"download_path,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Priority     int    `json:"priority"`
	AddedAt      time.Time `json:"added_at"`
}

// NewTorrentInfo 创建种子信息
func NewTorrentInfo() *TorrentInfo {
	return &TorrentInfo{
		Tags:     make([]string, 0),
		Priority: 0,
		AddedAt:  time.Now(),
	}
}

// MediaInfo 媒体信息
type MediaInfo struct {
	// 基本信息
	TMDBID      int    `json:"tmdb_id,omitempty"`
	DoubanID    string `json:"douban_id,omitempty"`
	IMDBID      string `json:"imdb_id,omitempty"`
	Title       string `json:"title"`
	OriginalTitle string `json:"original_title,omitempty"`
	Overview    string `json:"overview,omitempty"`
	Tagline     string `json:"tagline,omitempty"`

	// 媒体类型
	MediaType string `json:"media_type"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`

	// 时间信息
	ReleaseDate string `json:"release_date,omitempty"`
	AirDate     string `json:"air_date,omitempty"`
	Runtime     int    `json:"runtime,omitempty"`

	// 分类信息
	Genres     []string `json:"genres,omitempty"`
	Countries  []string `json:"countries,omitempty"`
	Languages  []string `json:"languages,omitempty"`

	// 评分信息
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
	Popularity  float64 `json:"popularity"`

	// 制作信息
	Budget      int64  `json:"budget,omitempty"`
	Revenue     int64  `json:"revenue,omitempty"`
	Status      string `json:"status,omitempty"`

	// 视频质量
	Quality     string `json:"quality,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
	VideoCodec  string `json:"video_codec,omitempty"`
	AudioCodec  string `json:"audio_codec,omitempty"`
	Source      string `json:"source,omitempty"`
	ReleaseGroup string `json:"release_group,omitempty"`

	// 路径信息
	FilePath    string `json:"file_path,omitempty"`
	DownloadPath string `json:"download_path,omitempty"`
	LibraryPath string `json:"library_path,omitempty"`

	// 元数据
	PosterPath  string `json:"poster_path,omitempty"`
	BackdropPath string `json:"backdrop_path,omitempty"`
	LogoPath    string `json:"logo_path,omitempty"`

	// 创建和更新时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMediaInfo 创建媒体信息
func NewMediaInfo(mediaType string) *MediaInfo {
	return &MediaInfo{
		MediaType: mediaType,
		Genres:    make([]string, 0),
		Countries: make([]string, 0),
		Languages: make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsMovie 是否为电影
func (m *MediaInfo) IsMovie() bool {
	return m.MediaType == "movie"
}

// IsTVSeries 是否为电视剧
func (m *MediaInfo) IsTVSeries() bool {
	return m.MediaType == "tv"
}

// IsAnime 是否为动漫
func (m *MediaInfo) IsAnime() bool {
	return m.MediaType == "anime"
}

// CacheInfo 缓存信息
type CacheInfo struct {
	Key        string        `json:"key"`
	Type       string        `json:"type"`
	Size       int64         `json:"size"`
	TTL        time.Duration `json:"ttl"`
	HitCount   int64         `json:"hit_count"`
	MissCount  int64         `json:"miss_count"`
	CreatedAt  time.Time     `json:"created_at"`
	AccessedAt time.Time     `json:"accessed_at"`
}

// NewCacheInfo 创建缓存信息
func NewCacheInfo(key, cacheType string, ttl time.Duration) *CacheInfo {
	return &CacheInfo{
		Key:        key,
		Type:       cacheType,
		TTL:        ttl,
		HitCount:   0,
		MissCount:  0,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
	}
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID          string            `json:"id"`
	UserID      uint              `json:"user_id"`
	Username    string            `json:"username"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	Data        map[string]interface{} `json:"data,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	LastAccess  time.Time         `json:"last_access"`
}

// NewSessionInfo 创建会话信息
func NewSessionInfo(userID uint, username, ipAddress, userAgent string, ttl time.Duration) *SessionInfo {
	return &SessionInfo{
		ID:        generateSessionID(),
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Data:      make(map[string]interface{}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		LastAccess: time.Now(),
	}
}

// IsExpired 检查会话是否过期
func (s *SessionInfo) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UpdateLastAccess 更新最后访问时间
func (s *SessionInfo) UpdateLastAccess() {
	s.LastAccess = time.Now()
}

// Set 设置会话数据
func (s *SessionInfo) Set(key string, value interface{}) {
	s.Data[key] = value
}

// Get 获取会话数据
func (s *SessionInfo) Get(key string) (interface{}, bool) {
	value, exists := s.Data[key]
	return value, exists
}

// ContextManager 上下文管理器
type ContextManager struct {
	sessions map[string]*SessionInfo
	mutex    sync.RWMutex
}

// NewContextManager 创建上下文管理器
func NewContextManager() *ContextManager {
	return &ContextManager{
		sessions: make(map[string]*SessionInfo),
	}
}

// SetRequestContext 设置请求上下文
func (cm *ContextManager) SetRequestContext(c *gin.Context) {
	reqCtx := NewRequestContext(c)
	c.Set(string(ContextKeyRequest), reqCtx)
}

// GetRequestContext 获取请求上下文
func (cm *ContextManager) GetRequestContext(c *gin.Context) (*RequestContext, bool) {
	value, exists := c.Get(string(ContextKeyRequest))
	if !exists {
		return nil, false
	}

	reqCtx, ok := value.(*RequestContext)
	return reqCtx, ok
}

// SetTorrentContext 设置种子上下文
func (cm *ContextManager) SetTorrentContext(c *gin.Context, torrent *TorrentInfo) {
	c.Set(string(ContextKeyTorrent), torrent)
}

// GetTorrentContext 获取种子上下文
func (cm *ContextManager) GetTorrentContext(c *gin.Context) (*TorrentInfo, bool) {
	value, exists := c.Get(string(ContextKeyTorrent))
	if !exists {
		return nil, false
	}

	torrent, ok := value.(*TorrentInfo)
	return torrent, ok
}

// SetMediaContext 设置媒体上下文
func (cm *ContextManager) SetMediaContext(c *gin.Context, media *MediaInfo) {
	c.Set(string(ContextKeyMedia), media)
}

// GetMediaContext 获取媒体上下文
func (cm *ContextManager) GetMediaContext(c *gin.Context) (*MediaInfo, bool) {
	value, exists := c.Get(string(ContextKeyMedia))
	if !exists {
		return nil, false
	}

	media, ok := value.(*MediaInfo)
	return media, ok
}

// CreateSession 创建会话
func (cm *ContextManager) CreateSession(userID uint, username, ipAddress, userAgent string, ttl time.Duration) *SessionInfo {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	session := NewSessionInfo(userID, username, ipAddress, userAgent, ttl)
	cm.sessions[session.ID] = session
	return session
}

// GetSession 获取会话
func (cm *ContextManager) GetSession(sessionID string) (*SessionInfo, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists || session.IsExpired() {
		return nil, false
	}

	session.UpdateLastAccess()
	return session, true
}

// DeleteSession 删除会话
func (cm *ContextManager) DeleteSession(sessionID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.sessions, sessionID)
}

// CleanupExpiredSessions 清理过期会话
func (cm *ContextManager) CleanupExpiredSessions() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for id, session := range cm.sessions {
		if session.IsExpired() {
			delete(cm.sessions, id)
		}
	}
}

// GetSessionCount 获取会话数量
func (cm *ContextManager) GetSessionCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	count := 0
	for _, session := range cm.sessions {
		if !session.IsExpired() {
			count++
		}
	}
	return count
}

// 辅助函数
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}