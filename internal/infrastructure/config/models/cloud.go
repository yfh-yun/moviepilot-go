package models

// CloudConfig 云盘配置
type CloudConfig struct {
	// 115云盘配置
	U115AppID string `mapstructure:"U115_APP_ID" default:"100196807"`
	
	// 阿里云盘配置
	AlipanAppID string `mapstructure:"ALIPAN_APP_ID" default:"ac1bf04dc9fd4d9aaabb65b4a668d403"`
}