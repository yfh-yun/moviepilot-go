package models

// PluginData 插件数据表
type PluginData struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 插件ID
	PluginID string `json:"plugin_id,omitempty" gorm:"index;not null"`
	
	// 数据key
	Key string `json:"key,omitempty" gorm:"index;not null"`
	
	// 数据值
	Value interface{} `json:"value,omitempty" gorm:"serializer:json"`
}

// TableName 设置表名
func (PluginData) TableName() string {
	return "plugindata"
}