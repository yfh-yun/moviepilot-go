package models

// SecurityConfig 安全配置
type SecurityConfig struct {
	SecretKey                        string   `mapstructure:"SECRET_KEY"`
	ResourceSecretKey                string   `mapstructure:"RESOURCE_SECRET_KEY"`
	AllowedHosts                     []string `mapstructure:"ALLOWED_HOSTS" default:"[\"*\"]"`
	AccessTokenExpireMinutes         int      `mapstructure:"ACCESS_TOKEN_EXPIRE_MINUTES" default:"11520"`
	ResourceAccessTokenExpireSeconds int      `mapstructure:"RESOURCE_ACCESS_TOKEN_EXPIRE_SECONDS" default:"1800"`
	SuperUser                        string   `mapstructure:"SUPERUSER" default:"admin"`
	SuperUserPassword                string   `mapstructure:"SUPERUSER_PASSWORD"`
	AuxiliaryAuthEnable              bool     `mapstructure:"AUXILIARY_AUTH_ENABLE" default:"false"`
	APIToken                         string   `mapstructure:"API_TOKEN"`
	AuthSite                         string   `mapstructure:"AUTH_SITE"`
	ImageDomains                     []string `mapstructure:"SECURITY_IMAGE_DOMAINS"`
	ImageSuffixes                    []string `mapstructure:"SECURITY_IMAGE_SUFFIXES"`
}
