package models

// DockerConfig Docker配置
type DockerConfig struct {
	// Docker Client API地址
	ClientAPI string `mapstructure:"DOCKER_CLIENT_API" default:"tcp://127.0.0.1:38379"`
}