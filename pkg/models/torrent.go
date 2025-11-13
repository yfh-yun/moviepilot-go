package models

import (
	"time"
)

// TorrentStatus 种子状态枚�?type TorrentStatus string

const (
	Transfer      TorrentStatus = "transfer"
	Downloading   TorrentStatus = "downloading"
	Completed     TorrentStatus = "completed"
	Stopped       TorrentStatus = "stopped"
	Errored       TorrentStatus = "errored"
)

// TorrentInfo 种子信息结构�?type TorrentInfo struct {
	ID            int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Hash          string        `json:"hash" gorm:"index;unique"`
	Name          string        `json:"name"`
	Size          int64         `json:"size"`
	Status        TorrentStatus `json:"status"`
	Progress      float64       `json:"progress"`
	DownloadSpeed int64         `json:"download_speed"`
	UploadSpeed   int64         `json:"upload_speed"`
	Downloaded    int64         `json:"downloaded"`
	Uploaded      int64         `json:"uploaded"`
	Ratio         float64       `json:"ratio"`
	ETA           int64         `json:"eta"` // 剩余时间(�?
	AddedTime     time.Time     `json:"added_time"`
	CompletedTime time.Time     `json:"completed_time"`
	SavePath      string        `json:"save_path"`
	Category      string        `json:"category"`
	Tags          []string      `json:"tags" gorm:"serializer:json"`
	Tracker       string        `json:"tracker"`
	MediaInfoID   *int64        `json:"media_info_id"` // 关联的媒体信息ID
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// TableName 设置表名
func (t *TorrentInfo) TableName() string {
	return "torrents"
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TorrentID int64     `json:"torrent_id" gorm:"index"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Progress  float64   `json:"progress"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (tf *TorrentFile) TableName() string {
	return "torrent_files"
}
