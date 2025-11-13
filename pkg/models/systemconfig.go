package models

// SystemConfig 系统配置
type SystemConfig struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 配置键
	Key string `json:"key" gorm:"not null;uniqueIndex"`
	
	// 配置值
	Value interface{} `json:"value,omitempty" gorm:"serializer:json"`
}

// TableName 设置表名
func (SystemConfig) TableName() string {
	return "systemconfig"
}