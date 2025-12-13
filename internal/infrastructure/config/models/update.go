package models

// UpdateConfig 系统升级配置
type UpdateConfig struct {
	// 重启自动升级
	AutoUpdate string `mapstructure:"MOVIEPILOT_AUTO_UPDATE" default:"release"`
	
	// 自动检查和更新站点资源包
	AutoUpdateResource bool `mapstructure:"AUTO_UPDATE_RESOURCE" default:"true"`
}