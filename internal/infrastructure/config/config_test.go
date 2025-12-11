package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLoad 测试配置加载
func TestConfigLoad(t *testing.T) {
	// 测试配置加载功能
	loader := NewLoader()
	cfg, err := loader.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.App)
	assert.NotNil(t, cfg.Database)
	// 只验证配置是否被正确加载，不依赖于具体的默认值
	assert.NotEmpty(t, cfg.App.ProjectName)
	assert.NotEmpty(t, cfg.App.Host)
	assert.NotZero(t, cfg.App.Port)
}

// TestConfigUpdate 测试配置更新
func TestConfigUpdate(t *testing.T) {
	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// 测试更新端口配置
	newPort := 8080
	err = cfg.UpdateSetting("PORT", newPort)
	assert.NoError(t, err)
	assert.Equal(t, newPort, cfg.App.Port)

	// 测试更新数据库类型
	newDBType := "postgres"
	err = cfg.UpdateSetting("DB_TYPE", newDBType)
	assert.NoError(t, err)
	assert.Equal(t, newDBType, cfg.Database.Type)

	// 测试更新布尔值配置
	newDebug := true
	err = cfg.UpdateSetting("DEBUG", newDebug)
	assert.NoError(t, err)
	assert.Equal(t, newDebug, cfg.App.Debug)
}

// TestSystemConf 测试系统配置动态计算
func TestSystemConf(t *testing.T) {
	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// 测试普通模式
	cfg.Performance.BigMemoryMode = false
	sysConf := cfg.GetSystemConf()
	assert.Equal(t, 100, sysConf.Torrents)
	assert.Equal(t, 50, sysConf.Refresh)
	assert.Equal(t, 256, sysConf.TMDB)
	assert.Equal(t, 50, sysConf.Scheduler)
	assert.Equal(t, 50, sysConf.ThreadPool)

	// 测试大内存模式
	cfg.Performance.BigMemoryMode = true
	sysConf = cfg.GetSystemConf()
	assert.Equal(t, 200, sysConf.Torrents)
	assert.Equal(t, 100, sysConf.Refresh)
	assert.Equal(t, 1024, sysConf.TMDB)
	assert.Equal(t, 100, sysConf.Scheduler)
	assert.Equal(t, 100, sysConf.ThreadPool)
}

// TestConfigValidation 测试配置校验
func TestConfigValidation(t *testing.T) {
	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// 测试有效配置
	err = ValidateConfig(cfg)
	assert.NoError(t, err)

	// 测试无效端口
	cfg.App.Port = 70000
	err = ValidateConfig(cfg)
	assert.Error(t, err)

	// 测试无效数据库类型
	cfg.App.Port = 3001 // 恢复有效端口
	cfg.Database.Type = "invalid_db_type"
	err = ValidateConfig(cfg)
	assert.Error(t, err)

	// 测试无效超级用户名
	cfg.Database.Type = "sqlite"  // 恢复有效数据库类型
	cfg.Security.SuperUser = "ab" // 太短
	err = ValidateConfig(cfg)
	assert.Error(t, err)
}

// TestConfigLoadFromEnv 测试从环境变量加载配置
func TestConfigLoadFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("PROJECT_NAME", "TestApp")
	os.Setenv("PORT", "9000")
	os.Setenv("DEBUG", "true")
	os.Setenv("DB_TYPE", "postgres")

	defer func() {
		// 清理环境变量
		os.Unsetenv("PROJECT_NAME")
		os.Unsetenv("PORT")
		os.Unsetenv("DEBUG")
		os.Unsetenv("DB_TYPE")
	}()

	// 加载配置
	loader := NewLoader()
	cfg, err := loader.Load()
	assert.NoError(t, err)

	// 验证从环境变量加载的配置
	assert.Equal(t, "TestApp", cfg.App.ProjectName)
	assert.Equal(t, 9000, cfg.App.Port)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "postgres", cfg.Database.Type)
}

// TestGetProxy 测试代理配置获取
func TestGetProxy(t *testing.T) {
	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// 测试无代理配置
	proxy := cfg.GetProxy()
	assert.Nil(t, proxy)

	// 测试有代理配置
	proxyHost := "http://127.0.0.1:8888"
	cfg.Network.ProxyHost = proxyHost
	proxy = cfg.GetProxy()
	assert.NotNil(t, proxy)
	assert.Equal(t, proxyHost, proxy["http"])
	assert.Equal(t, proxyHost, proxy["https"])
}

// TestIsValidDBType 测试数据库类型验证
func TestIsValidDBType(t *testing.T) {
	// 测试有效数据库类型
	assert.True(t, IsValidDBType("sqlite"))
	assert.True(t, IsValidDBType("postgres"))
	assert.True(t, IsValidDBType("postgresql"))

	// 测试无效数据库类型
	assert.False(t, IsValidDBType("invalid"))
	assert.False(t, IsValidDBType("mysql"))
}

// TestIsValidCacheBackendType 测试缓存后端类型验证
func TestIsValidCacheBackendType(t *testing.T) {
	// 测试有效缓存后端类型
	assert.True(t, IsValidCacheBackendType("cachetools"))
	assert.True(t, IsValidCacheBackendType("redis"))
	assert.True(t, IsValidCacheBackendType("memory"))

	// 测试无效缓存后端类型
	assert.False(t, IsValidCacheBackendType("invalid"))
	assert.False(t, IsValidCacheBackendType("memcached"))
}
