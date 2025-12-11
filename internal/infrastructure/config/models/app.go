package models

// AppConfig 应用基础配置
type AppConfig struct {
	ProjectName  string `mapstructure:"PROJECT_NAME" default:"MoviePilot"`
	AppDomain    string `mapstructure:"APP_DOMAIN"`
	APIV1Str     string `mapstructure:"API_V1_STR" default:"/api/v1"`
	FrontendPath string `mapstructure:"FRONTEND_PATH" default:"/public"`
	TZ           string `mapstructure:"TZ" default:"Asia/Shanghai"`
	Host         string `mapstructure:"HOST" default:"0.0.0.0"`
	Port         int    `mapstructure:"PORT" default:"3001"`
	NginxPort    int    `mapstructure:"NGINX_PORT" default:"3000"`
	ConfigDir    string `mapstructure:"CONFIG_DIR"`
	Debug        bool   `mapstructure:"DEBUG" default:"false"`
	Dev          bool   `mapstructure:"DEV" default:"false"`
	AdvancedMode bool   `mapstructure:"ADVANCED_MODE" default:"true"`
}
