package models

// SearchConfig 搜索配置
type SearchConfig struct {
	// 搜索多个名称
	SearchMultipleName bool `mapstructure:"SEARCH_MULTIPLE_NAME" default:"false"`
	
	// 最大搜索名称数量
	MaxSearchNameLimit int `mapstructure:"MAX_SEARCH_NAME_LIMIT" default:"2"`
}