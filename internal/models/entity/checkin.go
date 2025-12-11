package entity

import (
	"time"
)

// SiteCheckinHistory 站点签到历史表
type SiteCheckinHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SiteID    uint      `gorm:"index;not null" json:"site_id"`
	SiteName  string    `gorm:"type:varchar(200)" json:"site_name"`
	Success   bool      `gorm:"not null;default:false" json:"success"`
	Message   string    `gorm:"type:varchar(500)" json:"message"`
	Bonus     float64   `gorm:"type:decimal(10,2);default:0" json:"bonus"` // 获得的积分/魔力值
	Upload    int64     `gorm:"default:0" json:"upload"`                   // 获得的上传量（字节）
	Download  int64     `gorm:"default:0" json:"download"`                 // 获得的下载量（字节）
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SiteCheckinHistory) TableName() string {
	return "site_checkin_history"
}
