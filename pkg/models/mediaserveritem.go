package models

// MediaServerItem 媒体服务器媒体条目表
type MediaServerItem struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 服务器类型
	Server string `json:"server,omitempty" gorm:"index"`
	
	// 媒体库ID
	Library string `json:"library,omitempty"`
	
	// ID
	ItemID string `json:"item_id,omitempty" gorm:"index"`
	
	// 类型
	ItemType string `json:"item_type,omitempty"`
	
	// 标题
	Title string `json:"title,omitempty" gorm:"index"`
	
	// 原标题
	OriginalTitle string `json:"original_title,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// TMDBID
	TmdbID int `json:"tmdbid,omitempty" gorm:"index"`
	
	// IMDBID
	ImdbID string `json:"imdbid,omitempty" gorm:"index"`
	
	// TVDBID
	TvdbID string `json:"tvdbid,omitempty" gorm:"index"`
	
	// 路径
	Path string `json:"path,omitempty"`
	
	// 季集
	Seasoninfo map[string]interface{} `json:"seasoninfo,omitempty" gorm:"serializer:json;default:'{}'"`
	
	// 备注
	Note map[string]interface{} `json:"note,omitempty" gorm:"serializer:json"`
	
	// 同步时间
	LstModDate string `json:"lst_mod_date,omitempty"`
}

// TableName 设置表名
func (MediaServerItem) TableName() string {
	return "mediaserveritem"
}