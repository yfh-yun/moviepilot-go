package models

// TVDBConfig TVDB配置
type TVDBConfig struct {
	APIKey string `mapstructure:"TVDB_V4_API_KEY" default:"ed2aa66b-7899-4677-92a7-67bc9ce3d93a"`
	APIPin string `mapstructure:"TVDB_V4_API_PIN"`
}