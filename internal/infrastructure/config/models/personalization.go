package models

// PersonalizationConfig 个性化配置
type PersonalizationConfig struct {
	// 登录页面电影海报,tmdb/bing/mediaserver
	Wallpaper string `mapstructure:"WALLPAPER" default:"tmdb"`
	
	// 自定义壁纸api地址
	CustomWallpaperAPIURL string `mapstructure:"CUSTOMIZE_WALLPAPER_API_URL"`
}