package config

import (
	"fmt"
	"regexp"
)

// Validator 配置校验器
type Validator struct {
	config *Config
}

// NewValidator 创建配置校验器实例
func NewValidator(config *Config) *Validator {
	return &Validator{
		config: config,
	}
}

// Validate 校验所有配置项
func (v *Validator) Validate() error {
	// 校验应用配置
	if err := v.validateAppConfig(); err != nil {
		return err
	}

	// 校验数据库配置
	if err := v.validateDatabaseConfig(); err != nil {
		return err
	}

	// 校验安全配置
	if err := v.validateSecurityConfig(); err != nil {
		return err
	}

	// 校验网络配置
	if err := v.validateNetworkConfig(); err != nil {
		return err
	}

	// 校验TMDB配置
	if err := v.validateTMDBConfig(); err != nil {
		return err
	}

	return nil
}

// validateAppConfig 校验应用配置
func (v *Validator) validateAppConfig() error {
	app := v.config.App

	// 校验端口范围
	if app.Port < 1 || app.Port > 65535 {
		return fmt.Errorf("invalid app port: %d, must be between 1 and 65535", app.Port)
	}

	if app.NginxPort < 0 || app.NginxPort > 65535 {
		return fmt.Errorf("invalid nginx port: %d, must be between 0 and 65535", app.NginxPort)
	}

	return nil
}

// validateDatabaseConfig 校验数据库配置
func (v *Validator) validateDatabaseConfig() error {
	db := v.config.Database

	// 校验数据库类型
	validDBTypes := map[string]bool{
		"sqlite":     true,
		"postgres":   true,
		"postgresql": true,
	}
	if !validDBTypes[db.Type] {
		return fmt.Errorf("invalid database type: %s, must be one of sqlite, postgres, postgresql", db.Type)
	}

	// 校验PostgreSQL配置（如果使用PostgreSQL）
	if db.Type == "postgres" || db.Type == "postgresql" {
		if db.PostgreSQLHost == "" {
			return fmt.Errorf("postgresql host is required when using postgresql database")
		}
		if db.PostgreSQLPort < 1 || db.PostgreSQLPort > 65535 {
			return fmt.Errorf("invalid postgresql port: %d, must be between 1 and 65535", db.PostgreSQLPort)
		}
		if db.PostgreSQLDatabase == "" {
			return fmt.Errorf("postgresql database name is required")
		}
		if db.PostgreSQLUsername == "" {
			return fmt.Errorf("postgresql username is required")
		}
	}

	// 校验连接池配置
	if db.PoolRecycle < 0 {
		return fmt.Errorf("invalid pool recycle time: %d, must be non-negative", db.PoolRecycle)
	}
	if db.PoolTimeout < 0 {
		return fmt.Errorf("invalid pool timeout: %d, must be non-negative", db.PoolTimeout)
	}

	return nil
}

// validateSecurityConfig 校验安全配置
func (v *Validator) validateSecurityConfig() error {
	sec := v.config.Security

	// 校验访问令牌过期时间
	if sec.AccessTokenExpireMinutes <= 0 {
		return fmt.Errorf("invalid access token expire minutes: %d, must be positive", sec.AccessTokenExpireMinutes)
	}

	if sec.ResourceAccessTokenExpireSeconds <= 0 {
		return fmt.Errorf("invalid resource access token expire seconds: %d, must be positive", sec.ResourceAccessTokenExpireSeconds)
	}

	// 校验超级用户名格式
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`).MatchString(sec.SuperUser) {
		return fmt.Errorf("invalid superuser name: %s, must be 3-20 characters, containing only letters, numbers, underscores and hyphens", sec.SuperUser)
	}

	return nil
}

// validateNetworkConfig 校验网络配置
func (v *Validator) validateNetworkConfig() error {
	net := v.config.Network

	// 校验代理端口（如果配置了代理）
	if net.ProxyHost != "" && net.ProxyPort > 0 {
		if net.ProxyPort < 1 || net.ProxyPort > 65535 {
			return fmt.Errorf("invalid proxy port: %d, must be between 1 and 65535", net.ProxyPort)
		}
	}

	return nil
}

// validateTMDBConfig 校验TMDB配置
func (v *Validator) validateTMDBConfig() error {
	tmdb := v.config.TMDB

	// 校验语言格式
	if !regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`).MatchString(tmdb.Language) {
		return fmt.Errorf("invalid tmdb language: %s, must be in format like zh-CN", tmdb.Language)
	}

	// 校验区域格式
	if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(tmdb.Region) {
		return fmt.Errorf("invalid tmdb region: %s, must be 2 uppercase letters", tmdb.Region)
	}

	return nil
}

// ValidateConfig 全局校验函数，方便外部调用
func ValidateConfig(config *Config) error {
	validator := NewValidator(config)
	return validator.Validate()
}

// IsValidDBType 检查数据库类型是否有效
func IsValidDBType(dbType string) bool {
	validTypes := map[string]bool{
		"sqlite":     true,
		"postgres":   true,
		"postgresql": true,
	}
	return validTypes[dbType]
}

// IsValidCacheBackendType 检查缓存后端类型是否有效
func IsValidCacheBackendType(backendType string) bool {
	validTypes := map[string]bool{
		"cachetools": true,
		"redis":      true,
		"memory":     true,
	}
	return validTypes[backendType]
}
