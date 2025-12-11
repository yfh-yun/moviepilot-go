package entity

import (
	"time"
)

// SubscribeShare 订阅分享表
type SubscribeShare struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SubscribeID  int       `gorm:"index;not null" json:"subscribe_id"`
	ShareTitle   string    `gorm:"type:varchar(200);not null" json:"share_title"`
	ShareComment string    `gorm:"type:text" json:"share_comment"`
	ShareUser    string    `gorm:"type:varchar(100);not null;index" json:"share_user"`
	ShareUID     string    `gorm:"type:varchar(100);not null;index" json:"share_uid"`
	Name         string    `gorm:"type:varchar(200)" json:"name"`
	Year         string    `gorm:"type:varchar(10)" json:"year"`
	Type         string    `gorm:"type:varchar(20)" json:"type"`
	TmdbID       *int      `gorm:"index" json:"tmdb_id"`
	Season       *int      `json:"season"`
	Poster       string    `gorm:"type:varchar(500)" json:"poster"`
	ForkCount    int       `gorm:"default:0" json:"fork_count"` // 复用次数
	LikeCount    int       `gorm:"default:0" json:"like_count"` // 点赞次数
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SubscribeShare) TableName() string {
	return "subscribe_shares"
}

// UserFollow 用户关注表
type UserFollow struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"type:varchar(100);not null;index" json:"user_id"`
	FollowUID string    `gorm:"type:varchar(100);not null;index" json:"follow_uid"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (UserFollow) TableName() string {
	return "user_follows"
}
