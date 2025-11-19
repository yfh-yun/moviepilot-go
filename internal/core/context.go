package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// ContextManager 上下文管理器
type ContextManager struct {
	logger        *logger.Logger
	torrentInfos  map[string]*TorrentInfo
	mediaInfos    map[string]*MediaInfo
	contexts      map[string]*Context
	mutex         sync.RWMutex
}

// Context 上下文结构
type Context struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	TorrentInfo   *TorrentInfo      `json:"torrent_info,omitempty"`
	MediaInfo     *MediaInfo        `json:"media_info,omitempty"`
	Status        string            `json:"status"`
	Progress      float64           `json:"progress"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	CreateTime    time.Time         `json:"create_time"`
	UpdateTime    time.Time         `json:"update_time"`
	ExtraData     map[string]interface{} `json:"extra_data,omitempty"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Hash          string            `json:"hash"`
	Name          string            `json:"name"`
	Size          int64             `json:"size"`
	Files         []*TorrentFile    `json:"files"`
	Tracker       string            `json:"tracker,omitempty"`
	Category      string            `json:"category,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Uploader      string            `json:"uploader,omitempty"`
	Description   string            `json:"description,omitempty"`
	DownloadCount int               `json:"download_count,omitempty"`
	Seeders       int               `json:"seeders,omitempty"`
	Leechers      int               `json:"leechers,omitempty"`
	Ratio         float64           `json:"ratio,omitempty"`
	CreateTime    time.Time         `json:"create_time"`
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	Path          string            `json:"path"`
	Name          string            `json:"name"`
	Size          int64             `json:"size"`
	Index         int               `json:"index"`
	Selected      bool              `json:"selected"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title         string            `json:"title"`
	Year          int               `json:"year,omitempty"`
	Season        int               `json:"season,omitempty"`
	Episode       int               `json:"episode,omitempty"`
	Resolution    string            `json:"resolution,omitempty"`
	VideoCodec    string            `json:"video_codec,omitempty"`
	AudioCodec    string            `json:"audio_codec,omitempty"`
	Group         string            `json:"group,omitempty"`
	Language      string            `json:"language,omitempty"`
	IMDBID        string            `json:"imdb_id,omitempty"`
	TMDBID        string            `json:"tmdb_id,omitempty"`
	TVDBID        string            `json:"tvdb_id,omitempty"`
	MediaType     string            `json:"media_type"` // movie, series, anime
	Duration      int               `json:"duration,omitempty"`
	Rating        float64           `json:"rating,omitempty"`
	Overview      string            `json:"overview,omitempty"`
	PosterURL     string            `json:"poster_url,omitempty"`
	BackdropURL   string            `json:"backdrop_url,omitempty"`
	Genres        []string          `json:"genres,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	OriginalTitle string            `json:"original_title,omitempty"`
}

// NewContextManager 创建上下文管理器
func NewContextManager(log *logger.Logger) *ContextManager {
	return &ContextManager{
		logger:       log,
		torrentInfos: make(map[string]*TorrentInfo),
		mediaInfos:   make(map[string]*MediaInfo),
		contexts:     make(map[string]*Context),
	}
}

// CreateContext 创建新的上下文
func (cm *ContextManager) CreateContext(ctxID, name, ctxType string) *Context {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ctx := &Context{
		ID:         ctxID,
		Name:       name,
		Type:       ctxType,
		Status:     "pending",
		Progress:   0,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		ExtraData:  make(map[string]interface{}),
	}

	cm.contexts[ctxID] = ctx
	cm.logger.Info("Context created", "id", ctxID, "name", name, "type", ctxType)
	return ctx
}

// GetContext 获取上下文
func (cm *ContextManager) GetContext(ctxID string) *Context {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	ctx, exists := cm.contexts[ctxID]
	if !exists {
		cm.logger.Warn("Context not found", "id", ctxID)
		return nil
	}
	return ctx
}

// UpdateContext 更新上下文
func (cm *ContextManager) UpdateContext(ctxID string, updates map[string]interface{}) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ctx, exists := cm.contexts[ctxID]
	if !exists {
		cm.logger.Warn("Context not found for update", "id", ctxID)
		return false
	}

	// 更新字段
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				ctx.Status = status
			}
		case "progress":
			if progress, ok := value.(float64); ok {
				ctx.Progress = progress
			}
		case "error_message":
			if msg, ok := value.(string); ok {
				ctx.ErrorMessage = msg
			}
		case "extra_data":
			if data, ok := value.(map[string]interface{}); ok {
				for k, v := range data {
					ctx.ExtraData[k] = v
				}
			}
		}
	}

	ctx.UpdateTime = time.Now()
	cm.logger.Debug("Context updated", "id", ctxID, "status", ctx.Status, "progress", ctx.Progress)
	return true
}

// DeleteContext 删除上下文
func (cm *ContextManager) DeleteContext(ctxID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.contexts[ctxID]; !exists {
		cm.logger.Warn("Context not found for deletion", "id", ctxID)
		return false
	}

	delete(cm.contexts, ctxID)
	cm.logger.Info("Context deleted", "id", ctxID)
	return true
}

// ListContexts 列出所有上下文
func (cm *ContextManager) ListContexts(ctxType string) []*Context {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var result []*Context
	for _, ctx := range cm.contexts {
		if ctxType == "" || ctx.Type == ctxType {
			result = append(result, ctx)
		}
	}
	return result
}

// AddTorrentInfo 添加种子信息
func (cm *ContextManager) AddTorrentInfo(torrent *TorrentInfo) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.torrentInfos[torrent.Hash] = torrent
	cm.logger.Info("Torrent info added", "hash", torrent.Hash, "name", torrent.Name)
}

// GetTorrentInfo 获取种子信息
func (cm *ContextManager) GetTorrentInfo(hash string) *TorrentInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	info, exists := cm.torrentInfos[hash]
	if !exists {
		cm.logger.Warn("Torrent info not found", "hash", hash)
		return nil
	}
	return info
}

// DeleteTorrentInfo 删除种子信息
func (cm *ContextManager) DeleteTorrentInfo(hash string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.torrentInfos[hash]; !exists {
		cm.logger.Warn("Torrent info not found for deletion", "hash", hash)
		return false
	}

	delete(cm.torrentInfos, hash)
	cm.logger.Info("Torrent info deleted", "hash", hash)
	return true
}

// AddMediaInfo 添加媒体信息
func (cm *ContextManager) AddMediaInfo(media *MediaInfo) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	key := cm.generateMediaKey(media)
	cm.mediaInfos[key] = media
	cm.logger.Info("Media info added", "title", media.Title, "year", media.Year)
}

// GetMediaInfo 获取媒体信息
func (cm *ContextManager) GetMediaInfo(title string, year int) *MediaInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	key := fmt.Sprintf("%s_%d", title, year)
	info, exists := cm.mediaInfos[key]
	if !exists {
		cm.logger.Warn("Media info not found", "title", title, "year", year)
		return nil
	}
	return info
}

// FindMediaInfoByID 根据ID查找媒体信息
func (cm *ContextManager) FindMediaInfoByID(mediaID string, idType string) *MediaInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, media := range cm.mediaInfos {
		switch idType {
		case "imdb":
			if media.IMDBID == mediaID {
				return media
			}
		case "tmdb":
			if media.TMDBID == mediaID {
				return media
			}
		case "tvdb":
			if media.TVDBID == mediaID {
				return media
			}
		}
	}

	cm.logger.Warn("Media info not found by ID", "id", mediaID, "type", idType)
	return nil
}

// LinkContextToTorrent 将上下文与种子信息关联
func (cm *ContextManager) LinkContextToTorrent(ctxID string, torrentHash string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ctx, ctxExists := cm.contexts[ctxID]
	torrent, torrentExists := cm.torrentInfos[torrentHash]

	if !ctxExists || !torrentExists {
		cm.logger.Warn("Failed to link context to torrent", "ctx_id", ctxID, "torrent_hash", torrentHash)
		return false
	}

	ctx.TorrentInfo = torrent
	ctx.UpdateTime = time.Now()
	cm.logger.Info("Context linked to torrent", "ctx_id", ctxID, "torrent_hash", torrentHash)
	return true
}

// LinkContextToMedia 将上下文与媒体信息关联
func (cm *ContextManager) LinkContextToMedia(ctxID string, media *MediaInfo) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ctx, exists := cm.contexts[ctxID]
	if !exists {
		cm.logger.Warn("Failed to link context to media", "ctx_id", ctxID)
		return false
	}

	ctx.MediaInfo = media
	ctx.UpdateTime = time.Now()
	cm.logger.Info("Context linked to media", "ctx_id", ctxID, "media_title", media.Title)
	return true
}

// IsMediaFile 检查是否为媒体文件
func (cm *ContextManager) IsMediaFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	mediaExts := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".ts", ".webm", ".iso"}
	for _, ext := range mediaExts {
		if ext == ext {
			return true
		}
	}
	return false
}

// IsSubtitleFile 检查是否为字幕文件
func (cm *ContextManager) IsSubtitleFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	subExts := []string{".srt", ".ass", ".ssa", ".idx", ".sub", ".vtt"}
	for _, ext := range subExts {
		if ext == ext {
			return true
		}
	}
	return false
}

// CleanupOldContexts 清理过期的上下文
func (cm *ContextManager) CleanupOldContexts(maxAge time.Duration) int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	deletedCount := 0

	for id, ctx := range cm.contexts {
		if now.Sub(ctx.UpdateTime) > maxAge {
			delete(cm.contexts, id)
			deletedCount++
			cm.logger.Info("Old context cleaned up", "id", id)
		}
	}

	cm.logger.Info("Cleanup old contexts completed", "deleted_count", deletedCount)
	return deletedCount
}

// generateMediaKey 生成媒体信息的唯一键
func (cm *ContextManager) generateMediaKey(media *MediaInfo) string {
	if media.Year > 0 {
		return fmt.Sprintf("%s_%d", media.Title, media.Year)
	}
	return media.Title
}

// GetStatistics 获取上下文管理器统计信息
func (cm *ContextManager) GetStatistics() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	statusCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, ctx := range cm.contexts {
		statusCount[ctx.Status]++
		typeCount[ctx.Type]++
	}

	return map[string]interface{}{
		"total_contexts":  len(cm.contexts),
		"total_torrents":  len(cm.torrentInfos),
		"total_medias":    len(cm.mediaInfos),
		"status_count":    statusCount,
		"type_count":      typeCount,
	}
}
