package models

import (
	"time"

	"gorm.io/gorm"
)

// Site 站点模型
type Site struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	Name     string `gorm:"size:100;not null" json:"name"`
	URL      string `gorm:"size:500;not null" json:"url"`
	Type     string `gorm:"size:20;not null" json:"type"` // pt, public, rss
	Priority int    `gorm:"default:5" json:"priority"`
	Enabled  bool   `gorm:"default:true;not null" json:"enabled"`

	// 认证信息
	Cookie    string `gorm:"type:text" json:"cookie,omitempty"`
	UserAgent string `gorm:"size:500" json:"user_agent,omitempty"`
	Proxy     string `gorm:"size:200" json:"proxy,omitempty"`

	// 签到配置
	CheckinEnabled bool   `gorm:"default:false;not null" json:"checkin_enabled"`
	CheckinCron    string `gorm:"size:50;default:'0 8 * * *'" json:"checkin_cron"`
	CheckinURL     string `gorm:"size:500" json:"checkin_url,omitempty"`

	// 流量统计
	Upload   int64   `gorm:"default:0" json:"upload"`
	Download int64   `gorm:"default:0" json:"download"`
	Ratio    float64 `gorm:"type:decimal(10,2);default:0" json:"ratio"`

	// 状态信息
	Status       string     `gorm:"size:20;default:active;not null" json:"status"` // active, error, disabled
	LastCheckin  *time.Time `json:"last_checkin,omitempty"`
	LastSync     *time.Time `json:"last_sync,omitempty"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Cookies     []SiteCookie `gorm:"foreignKey:SiteID" json:"cookies,omitempty"`
	CheckinLogs []CheckinLog `gorm:"foreignKey:SiteID" json:"checkin_logs,omitempty"`
	Stats       []SiteStats  `gorm:"foreignKey:SiteID" json:"stats,omitempty"`
	SyncLogs    []SyncLog    `gorm:"foreignKey:SiteID" json:"sync_logs,omitempty"`
}

// TableName 指定表名
func (Site) TableName() string {
	return "sites"
}

// IsActive 检查站点是否活跃
func (s *Site) IsActive() bool {
	return s.Status == "active" && s.Enabled
}

// IsError 检查站点是否有错误
func (s *Site) IsError() bool {
	return s.Status == "error"
}

// IsPT 检查是否是 PT 站点
func (s *Site) IsPT() bool {
	return s.Type == "pt"
}

// IsPublic 检查是否是公开 Tracker
func (s *Site) IsPublic() bool {
	return s.Type == "public"
}

// IsRSS 检查是否是 RSS 订阅
func (s *Site) IsRSS() bool {
	return s.Type == "rss"
}
