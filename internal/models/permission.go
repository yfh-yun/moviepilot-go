package models

import (
	"time"
)

// Permission 权限模型
type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"` // 格式：action:resource
	Resource    string    `gorm:"size:50;not null" json:"resource"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	// 关联
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// String 返回权限的字符串表示
func (p *Permission) String() string {
	return p.Name
}
