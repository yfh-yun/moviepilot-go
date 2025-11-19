package service

import (
	"context"
	"time"
)

// HistoryService 历史记录服务接口
type HistoryService interface {
	// 下载历史
	GetDownloadHistory(params DownloadHistoryParams) (*PaginatedResponse[DownloadHistoryItem], error)
	CreateDownloadHistory(history *DownloadHistoryItem) error
	UpdateDownloadHistory(id string, history *DownloadHistoryItem) error
	DeleteDownloadHistory(id string) error
	
	// 转移历史
	GetTransferHistory(params TransferHistoryParams) (*PaginatedResponse[TransferHistoryItem], error)
	CreateTransferHistory(history *TransferHistoryItem) error
	UpdateTransferHistory(id string, history *TransferHistoryItem) error
	DeleteTransferHistory(id string) error
	
	// 订阅历史
	GetSubscribeHistory(params SubscribeHistoryParams) (*PaginatedResponse[SubscribeHistoryItem], error)
	CreateSubscribeHistory(history *SubscribeHistoryItem) error
	UpdateSubscribeHistory(id string, history *SubscribeHistoryItem) error
	DeleteSubscribeHistory(id string) error
	
	// 系统历史
	GetSystemHistory(params SystemHistoryParams) (*PaginatedResponse[SystemHistoryItem], error)
	CreateSystemHistory(history *SystemHistoryItem) error
	UpdateSystemHistory(id string, history *SystemHistoryItem) error
	DeleteSystemHistory(id string) error
	
	// 清理历史
	CleanupHistory(ctx context.Context, beforeDate time.Time) error
}

// 下载历史参数
type DownloadHistoryParams struct {
	Page      int    `json:"page"`
	Count     int    `json:"count"`
	Status    string `json:"status,omitempty"`    // success, failed, pending
	MediaType string `json:"media_type,omitempty"` // movie, tv, episode
	MediaID   string `json:"media_id,omitempty"`
	Source    string `json:"source,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// 转移历史参数
type TransferHistoryParams struct {
	Page        int    `json:"page"`
	Count       int    `json:"count"`
	Status      string `json:"status,omitempty"`      // success, failed, pending
	MediaType   string `json:"media_type,omitempty"`  // movie, tv, episode
	MediaID     string `json:"media_id,omitempty"`
	SourcePath  string `json:"source_path,omitempty"`
	DestPath    string `json:"dest_path,omitempty"`
	TransferMode string `json:"transfer_mode,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
}

// 订阅历史参数
type SubscribeHistoryParams struct {
	Page      int    `json:"page"`
	Count     int    `json:"count"`
	Status    string `json:"status,omitempty"`    // success, failed, pending
	MediaType string `json:"media_type,omitempty"` // movie, tv, episode
	MediaID   string `json:"media_id,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// 系统历史参数
type SystemHistoryParams struct {
	Page      int    `json:"page"`
	Count     int    `json:"count"`
	Type      string `json:"type,omitempty"`      // download, transfer, subscribe, plugin
	Level     string `json:"level,omitempty"`     // info, warning, error
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// 下载历史项目
type DownloadHistoryItem struct {
	ID          string    `json:"id"`
	MediaID     string    `json:"media_id"`
	MediaTitle  string    `json:"media_title"`
	MediaType   string    `json:"media_type"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`      // success, failed, pending
	Size        int64     `json:"size"`
	Downloaded  int64     `json:"downloaded"`
	Speed       int64     `json:"speed"`       // bytes per second
	Progress    int       `json:"progress"`    // percentage
	Error       string    `json:"error,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// 转移历史项目
type TransferHistoryItem struct {
	ID           string    `json:"id"`
	MediaID      string    `json:"media_id"`
	MediaTitle   string    `json:"media_title"`
	MediaType    string    `json:"media_type"`
	SourcePath   string    `json:"source_path"`
	DestPath     string    `json:"dest_path"`
	Status       string    `json:"status"`       // success, failed, pending
	Size         int64     `json:"size"`
	Transferred  int64     `json:"transferred"`
	Speed        int64     `json:"speed"`        // bytes per second
	Progress     int       `json:"progress"`     // percentage
	TransferMode string    `json:"transfer_mode"` // move, copy, sync
	Error        string    `json:"error,omitempty"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// 订阅历史项目
type SubscribeHistoryItem struct {
	ID         string    `json:"id"`
	MediaID    string    `json:"media_id"`
	MediaTitle string    `json:"media_title"`
	MediaType  string    `json:"media_type"`
	Season     int       `json:"season,omitempty"`
	Episode    int       `json:"episode,omitempty"`
	Status     string    `json:"status"`      // success, failed, pending
	Error      string    `json:"error,omitempty"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// 系统历史项目
type SystemHistoryItem struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`        // download, transfer, subscribe, plugin
	Level      string    `json:"level"`       // info, warning, error
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	CreateTime time.Time `json:"create_time"`
}

// PaginatedResponse 分页响应
type PaginatedResponse[T any] struct {
	Items      []T      `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Count      int       `json:"count"`
	TotalPages int       `json:"total_pages"`
	HasNext    bool      `json:"has_next"`
	HasPrev    bool      `json:"has_prev"`
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse[T any](items []T, total int64, page, count int) *PaginatedResponse[T] {
	totalPages := int((total + int64(count) - 1) / int64(count))
	if totalPages == 0 {
		totalPages = 1
	}
	
	return &PaginatedResponse[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Count:      count,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}