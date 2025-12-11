package models

import (
	"time"
)

// Role 角色模型
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	DisplayName string    `gorm:"size:100" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	IsSystem    bool      `gorm:"default:false;not null" json:"is_system"` // 系统角色不可删除
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// IsAdmin 检查是否是管理员角色
func (r *Role) IsAdmin() bool {
	return r.Name == "admin"
}

// IsUser 检查是否是普通用户角色
func (r *Role) IsUser() bool {
	return r.Name == "user"
}

// IsGuest 检查是否是访客角色
func (r *Role) IsGuest() bool {
	return r.Name == "guest"
}
