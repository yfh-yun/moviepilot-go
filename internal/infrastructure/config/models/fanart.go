package models

// FanartConfig Fanart配置
type FanartConfig struct {
	Enable  bool   `mapstructure:"FANART_ENABLE" default:"true"`
	Lang    string `mapstructure:"FANART_LANG" default:"zh,en"`
	APIKey  string `mapstructure:"FANART_API_KEY" default:"d2d31f9ecabea050fc7d68aa3146015f"`
	Timeout int    `mapstructure:"FANART_TIMEOUT" default:"30"`
}