# config.py 配置系统迁移计划

> Python: `app/core/config.py` (876 行)  
> Go: `internal/infrastructure/config/` + `pkg/config/`

---

## 📋 目录

1. [Python 配置系统分析](#1-python-配置系统分析)
2. [Go 设计方案](#2-go-设计方案)
3. [配置项分类与迁移](#3-配置项分类与迁移)
4. [实现计划](#4-实现计划)
5. [测试策略](#5-测试策略)

---

## 1. Python 配置系统分析

### 1.1 核心组件

| 组件 | 职责 | 代码行数 |
|------|------|---------|
| `SystemConfModel` | 系统资源配置（缓存大小、线程池等） | 23-44 |
| `ConfigModel` | 所有配置项定义（400+ 配置项） | 47-408 |
| `Settings` | 配置加载与管理（继承 Pydantic BaseSettings） | 410-791 |
| `GlobalVar` | 全局运行时状态（停止事件、订阅等） | 797-876 |

### 1.2 配置来源

```python
class Settings(BaseSettings, ConfigModel, LogConfigModel):
    class Config:
        case_sensitive = True
        env_file = SystemUtils.get_env_path()  # 从 .env 文件加载
        env_file_encoding = "utf-8"
```

**加载优先级**：
1. 环境变量（最高优先级）
2. `.env` 文件（`app.env`）
3. 代码中的默认值

### 1.3 核心特性

#### 1.3.1 自动类型转换与校验

```python
@validator('*', pre=True, always=True)
def generic_type_validator(cls, value: Any, field):
    # 自动将字符串转换为 int/bool/list 等类型
    # 校验失败时使用默认值并记录日志
    # 自动修正并写回 .env 文件
```

#### 1.3.2 运行时配置更新

```python
def update_setting(self, key: str, value: Any) -> Tuple[Optional[bool], str]:
    # 1. 类型转换与校验
    # 2. 写入 .env 文件
    # 3. 更新内存中的配置
```

#### 1.3.3 动态属性（Property）

```python
@property
def CONFIG_PATH(self):
    if self.CONFIG_DIR:
        return Path(self.CONFIG_DIR)
    elif SystemUtils.is_docker():
        return Path("/config")
    elif SystemUtils.is_frozen():
        return Path(sys.executable).parent / "config"
    return self.ROOT_PATH / "config"
```

### 1.4 配置分类统计

| 分类 | 配置项数量 | 示例 |
|------|-----------|------|
| 基础应用配置 | 10 | PROJECT_NAME, HOST, PORT, DEBUG |
| 安全认证配置 | 8 | SECRET_KEY, API_TOKEN, SUPERUSER |
| 数据库配置 | 17 | DB_TYPE, DB_POSTGRESQL_HOST, DB_POOL_SIZE |
| 缓存配置 | 7 | CACHE_BACKEND_TYPE, CACHE_REDIS_MAXMEMORY |
| 网络代理配置 | 5 | PROXY_HOST, DOH_ENABLE, DOH_DOMAINS |
| 媒体元数据配置 | 4 | SEARCH_SOURCE, RECOGNIZE_SOURCE |
| TMDB 配置 | 6 | TMDB_API_KEY, TMDB_IMAGE_DOMAIN |
| 站点配置 | 7 | SITEDATA_REFRESH_INTERVAL, OCR_HOST |
| 下载配置 | 5 | TORRENT_TAG, DOWNLOAD_SUBTITLE |
| CookieCloud 配置 | 6 | COOKIECLOUD_HOST, COOKIECLOUD_INTERVAL |
| 整理配置 | 6 | MOVIE_RENAME_FORMAT, TV_RENAME_FORMAT |
| 插件配置 | 4 | PLUGIN_MARKET, PLUGIN_AUTO_RELOAD |
| 性能配置 | 5 | BIG_MEMORY_MODE, MEMORY_GC_INTERVAL |
| 其他配置 | 30+ | 订阅、媒体服务器、工作流等 |

**总计：120+ 配置项**

---

## 2. Go 设计方案

### 2.1 目录结构

```
internal/infrastructure/config/
├── config.go           # 配置主结构体
├── loader.go           # 配置加载器（env + yaml + 默认值）
├── validator.go        # 配置校验器
├── updater.go          # 运行时配置更新
├── models/             # 配置模型（分类）
│   ├── app.go          # 应用配置
│   ├── database.go     # 数据库配置
│   ├── cache.go        # 缓存配置
│   ├── security.go     # 安全配置
│   ├── media.go        # 媒体配置
│   ├── plugin.go       # 插件配置
│   └── performance.go  # 性能配置
└── defaults/           # 默认配置
    └── defaults.go
```

### 2.2 核心结构设计

**config.go**：

```go
package config

import (
    "sync"
    "time"
)

// Config 全局配置结构体
type Config struct {
    mu sync.RWMutex  // 读写锁，支持运行时更新

    // 各分类配置（组合模式）
    App         *AppConfig
    Database    *DatabaseConfig
    Cache       *CacheConfig
    Security    *SecurityConfig
    Media       *MediaConfig
    TMDB        *TMDBConfig
    Site        *SiteConfig
    Download    *DownloadConfig
    CookieCloud *CookieCloudConfig
    Transfer    *TransferConfig
    Plugin      *PluginConfig
    Performance *PerformanceConfig
    Scheduler   *SchedulerConfig
    Subscribe   *SubscribeConfig
    Network     *NetworkConfig
    
    // 动态计算属性（通过方法实现）
    paths *PathConfig  // 内部字段
}

// AppConfig 应用基础配置
type AppConfig struct {
    ProjectName   string `mapstructure:"PROJECT_NAME" default:"MoviePilot"`
    AppDomain     string `mapstructure:"APP_DOMAIN"`
    APIV1Str      string `mapstructure:"API_V1_STR" default:"/api/v1"`
    FrontendPath  string `mapstructure:"FRONTEND_PATH" default:"/public"`
    TZ            string `mapstructure:"TZ" default:"Asia/Shanghai"`
    Host          string `mapstructure:"HOST" default:"0.0.0.0"`
    Port          int    `mapstructure:"PORT" default:"3001"`
    NginxPort     int    `mapstructure:"NGINX_PORT" default:"3000"`
    ConfigDir     string `mapstructure:"CONFIG_DIR"`
    Debug         bool   `mapstructure:"DEBUG" default:"false"`
    Dev           bool   `mapstructure:"DEV" default:"false"`
    AdvancedMode  bool   `mapstructure:"ADVANCED_MODE" default:"true"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Type              string `mapstructure:"DB_TYPE" default:"sqlite"`
    Echo              bool   `mapstructure:"DB_ECHO" default:"false"`
    Timeout           int    `mapstructure:"DB_TIMEOUT" default:"60"`
    WALEnable         bool   `mapstructure:"DB_WAL_ENABLE" default:"true"`
    PoolType          string `mapstructure:"DB_POOL_TYPE" default:"QueuePool"`
    PoolPrePing       bool   `mapstructure:"DB_POOL_PRE_PING" default:"true"`
    PoolRecycle       int    `mapstructure:"DB_POOL_RECYCLE" default:"300"`
    PoolTimeout       int    `mapstructure:"DB_POOL_TIMEOUT" default:"30"`
    
    // SQLite 配置
    SQLitePoolSize    int    `mapstructure:"DB_SQLITE_POOL_SIZE" default:"10"`
    SQLiteMaxOverflow int    `mapstructure:"DB_SQLITE_MAX_OVERFLOW" default:"50"`
    
    // PostgreSQL 配置
    PostgreSQLHost        string `mapstructure:"DB_POSTGRESQL_HOST" default:"localhost"`
    PostgreSQLPort        int    `mapstructure:"DB_POSTGRESQL_PORT" default:"5432"`
    PostgreSQLDatabase    string `mapstructure:"DB_POSTGRESQL_DATABASE" default:"moviepilot"`
    PostgreSQLUsername    string `mapstructure:"DB_POSTGRESQL_USERNAME" default:"moviepilot"`
    PostgreSQLPassword    string `mapstructure:"DB_POSTGRESQL_PASSWORD" default:"moviepilot"`
    PostgreSQLPoolSize    int    `mapstructure:"DB_POSTGRESQL_POOL_SIZE" default:"10"`
    PostgreSQLMaxOverflow int    `mapstructure:"DB_POSTGRESQL_MAX_OVERFLOW" default:"50"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
    BackendType       string `mapstructure:"CACHE_BACKEND_TYPE" default:"cachetools"`
    BackendURL        string `mapstructure:"CACHE_BACKEND_URL" default:"redis://localhost:6379"`
    RedisMaxMemory    string `mapstructure:"CACHE_REDIS_MAXMEMORY"`
    GlobalImageCache  bool   `mapstructure:"GLOBAL_IMAGE_CACHE" default:"false"`
    ImageCacheDays    int    `mapstructure:"GLOBAL_IMAGE_CACHE_DAYS" default:"7"`
    TempFileDays      int    `mapstructure:"TEMP_FILE_DAYS" default:"3"`
    MetaCacheExpire   int    `mapstructure:"META_CACHE_EXPIRE" default:"0"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
    SecretKey                      string   `mapstructure:"SECRET_KEY"`
    ResourceSecretKey              string   `mapstructure:"RESOURCE_SECRET_KEY"`
    AllowedHosts                   []string `mapstructure:"ALLOWED_HOSTS" default:"[\"*\"]"`
    AccessTokenExpireMinutes       int      `mapstructure:"ACCESS_TOKEN_EXPIRE_MINUTES" default:"11520"`
    ResourceAccessTokenExpireSeconds int    `mapstructure:"RESOURCE_ACCESS_TOKEN_EXPIRE_SECONDS" default:"1800"`
    SuperUser                      string   `mapstructure:"SUPERUSER" default:"admin"`
    SuperUserPassword              string   `mapstructure:"SUPERUSER_PASSWORD"`
    AuxiliaryAuthEnable            bool     `mapstructure:"AUXILIARY_AUTH_ENABLE" default:"false"`
    APIToken                       string   `mapstructure:"API_TOKEN"`
    AuthSite                       string   `mapstructure:"AUTH_SITE"`
    ImageDomains                   []string `mapstructure:"SECURITY_IMAGE_DOMAINS"`
    ImageSuffixes                  []string `mapstructure:"SECURITY_IMAGE_SUFFIXES"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
    BigMemoryMode                    bool    `mapstructure:"BIG_MEMORY_MODE" default:"false"`
    EncodingDetectionPerformanceMode bool    `mapstructure:"ENCODING_DETECTION_PERFORMANCE_MODE" default:"true"`
    EncodingDetectionMinConfidence   float64 `mapstructure:"ENCODING_DETECTION_MIN_CONFIDENCE" default:"0.8"`
    MemoryGCInterval                 int     `mapstructure:"MEMORY_GC_INTERVAL" default:"30"`
}

// SystemConfModel 系统资源配置（根据性能模式动态计算）
type SystemConfModel struct {
    Torrents   int
    Refresh    int
    TMDB       int
    Douban     int
    Bangumi    int
    Fanart     int
    Meta       int
    Scheduler  int
    ThreadPool int
}
```

### 2.3 配置加载器设计

**loader.go**：

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/joho/godotenv"
    "github.com/spf13/viper"
    "go.uber.org/zap"

    "moviepilot-go/pkg/logger"
    "moviepilot-go/pkg/utils"
)

type Loader struct {
    logger *zap.Logger
}

func NewLoader() *Loader {
    return &Loader{
        logger: logger.GetLogger(),
    }
}

// Load 加载配置（优先级：环境变量 > .env 文件 > 默认值）
func (l *Loader) Load() (*Config, error) {
    // 1. 加载 .env 文件
    envPath := l.getEnvPath()
    if err := godotenv.Load(envPath); err != nil {
        l.logger.Warn("failed to load .env file", zap.String("path", envPath), zap.Error(err))
    }

    // 2. 初始化 viper
    v := viper.New()
    v.AutomaticEnv()  // 自动读取环境变量
    
    // 3. 设置默认值
    l.setDefaults(v)

    // 4. 解析配置
    cfg := &Config{}
    if err := l.unmarshal(v, cfg); err != nil {
        return nil, err
    }

    // 5. 校验配置
    if err := l.validate(cfg); err != nil {
        return nil, err
    }

    // 6. 初始化动态属性
    cfg.initPaths()

    l.logger.Info("configuration loaded successfully")
    return cfg, nil
}

func (l *Loader) getEnvPath() string {
    // 优先级：环境变量 > Docker > 二进制 > 开发环境
    if envPath := os.Getenv("ENV_FILE_PATH"); envPath != "" {
        return envPath
    }
    
    if utils.IsDocker() {
        return "/config/app.env"
    }
    
    if utils.IsFrozen() {
        execPath, _ := os.Executable()
        return filepath.Join(filepath.Dir(execPath), "config", "app.env")
    }
    
    return "./config/app.env"
}

func (l *Loader) setDefaults(v *viper.Viper) {
    // 应用配置
    v.SetDefault("PROJECT_NAME", "MoviePilot")
    v.SetDefault("API_V1_STR", "/api/v1")
    v.SetDefault("FRONTEND_PATH", "/public")
    v.SetDefault("TZ", "Asia/Shanghai")
    v.SetDefault("HOST", "0.0.0.0")
    v.SetDefault("PORT", 3001)
    v.SetDefault("NGINX_PORT", 3000)
    v.SetDefault("DEBUG", false)
    v.SetDefault("DEV", false)
    v.SetDefault("ADVANCED_MODE", true)
    
    // 数据库配置
    v.SetDefault("DB_TYPE", "sqlite")
    v.SetDefault("DB_ECHO", false)
    v.SetDefault("DB_TIMEOUT", 60)
    // ... 其他默认值
}

func (l *Loader) unmarshal(v *viper.Viper, cfg *Config) error {
    cfg.App = &AppConfig{}
    cfg.Database = &DatabaseConfig{}
    cfg.Cache = &CacheConfig{}
    cfg.Security = &SecurityConfig{}
    // ... 其他配置

    if err := v.Unmarshal(&cfg.App); err != nil {
        return fmt.Errorf("failed to unmarshal app config: %w", err)
    }
    
    if err := v.Unmarshal(&cfg.Database); err != nil {
        return fmt.Errorf("failed to unmarshal database config: %w", err)
    }
    
    // ... 其他配置解析
    
    return nil
}
```

### 2.4 运行时配置更新

**updater.go**：

```go
package config

import (
    "fmt"
    "os"

    "github.com/joho/godotenv"
)

// UpdateSetting 更新单个配置项
func (c *Config) UpdateSetting(key string, value interface{}) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // 1. 校验配置项是否存在
    if !c.isValidKey(key) {
        return fmt.Errorf("invalid config key: %s", key)
    }

    // 2. 类型转换与校验
    convertedValue, err := c.convertValue(key, value)
    if err != nil {
        return err
    }

    // 3. 更新内存中的配置
    if err := c.setConfigValue(key, convertedValue); err != nil {
        return err
    }

    // 4. 写入 .env 文件
    if err := c.writeToEnvFile(key, convertedValue); err != nil {
        return err
    }

    return nil
}

// UpdateSettings 批量更新配置
func (c *Config) UpdateSettings(settings map[string]interface{}) map[string]error {
    results := make(map[string]error)
    for key, value := range settings {
        results[key] = c.UpdateSetting(key, value)
    }
    return results
}

func (c *Config) writeToEnvFile(key string, value interface{}) error {
    envPath := c.getEnvPath()
    
    // 读取现有 .env 文件
    envMap, err := godotenv.Read(envPath)
    if err != nil {
        envMap = make(map[string]string)
    }

    // 更新配置项
    envMap[key] = fmt.Sprintf("%v", value)

    // 写回文件
    return godotenv.Write(envMap, envPath)
}
```

### 2.5 动态属性实现

**config.go**（续）：

```go
// PathConfig 路径配置（动态计算）
type PathConfig struct {
    Root       string
    Config     string
    Temp       string
    Cache      string
    PluginData string
    Log        string
    Cookie     string
}

func (c *Config) initPaths() {
    c.paths = &PathConfig{}
    
    // Root Path
    c.paths.Root = c.getRootPath()
    
    // Config Path
    if c.App.ConfigDir != "" {
        c.paths.Config = c.App.ConfigDir
    } else if utils.IsDocker() {
        c.paths.Config = "/config"
    } else if utils.IsFrozen() {
        execPath, _ := os.Executable()
        c.paths.Config = filepath.Join(filepath.Dir(execPath), "config")
    } else {
        c.paths.Config = filepath.Join(c.paths.Root, "config")
    }
    
    // 其他路径
    c.paths.Temp = filepath.Join(c.paths.Config, "temp")
    c.paths.Cache = filepath.Join(c.paths.Config, "cache")
    c.paths.PluginData = filepath.Join(c.paths.Config, "plugins")
    c.paths.Log = filepath.Join(c.paths.Config, "logs")
    c.paths.Cookie = filepath.Join(c.paths.Config, "cookies")
}

// GetConfigPath 获取配置目录路径
func (c *Config) GetConfigPath() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.paths.Config
}

// GetTempPath 获取临时文件路径
func (c *Config) GetTempPath() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.paths.Temp
}

// GetSystemConf 根据性能模式返回系统配置
func (c *Config) GetSystemConf() *SystemConfModel {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.Performance.BigMemoryMode {
        metaExpire := c.Cache.MetaCacheExpire
        if metaExpire == 0 {
            metaExpire = 72
        }
        return &SystemConfModel{
            Torrents:   200,
            Refresh:    100,
            TMDB:       1024,
            Douban:     512,
            Bangumi:    512,
            Fanart:     512,
            Meta:       metaExpire * 3600,
            Scheduler:  100,
            ThreadPool: 100,
        }
    }
    
    metaExpire := c.Cache.MetaCacheExpire
    if metaExpire == 0 {
        metaExpire = 24
    }
    return &SystemConfModel{
        Torrents:   100,
        Refresh:    50,
        TMDB:       256,
        Douban:     256,
        Bangumi:    256,
        Fanart:     128,
        Meta:       metaExpire * 3600,
        Scheduler:  50,
        ThreadPool: 50,
    }
}

// GetProxy 获取代理配置
func (c *Config) GetProxy() map[string]string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.Network.ProxyHost != "" {
        return map[string]string{
            "http":  c.Network.ProxyHost,
            "https": c.Network.ProxyHost,
        }
    }
    return nil
}
```

---

## 3. 配置项分类与迁移

### 3.1 Phase 1: 核心配置（Week 1）

**优先级：🔴 高**

| Python 配置项 | Go 配置项 | 所属结构 | 默认值 |
|--------------|----------|---------|--------|
| `PROJECT_NAME` | `App.ProjectName` | AppConfig | "MoviePilot" |
| `HOST` | `App.Host` | AppConfig | "0.0.0.0" |
| `PORT` | `App.Port` | AppConfig | 3001 |
| `DEBUG` | `App.Debug` | AppConfig | false |
| `CONFIG_DIR` | `App.ConfigDir` | AppConfig | "" |
| `DB_TYPE` | `Database.Type` | DatabaseConfig | "sqlite" |
| `DB_POSTGRESQL_*` | `Database.PostgreSQL*` | DatabaseConfig | - |
| `SECRET_KEY` | `Security.SecretKey` | SecurityConfig | 自动生成 |
| `API_TOKEN` | `Security.APIToken` | SecurityConfig | 自动生成 |

**实现任务**：
- [x] 创建基础配置结构体
- [ ] 实现配置加载器（viper + godotenv）
- [ ] 实现默认值设置
- [ ] 实现环境变量读取
- [ ] 单元测试

### 3.2 Phase 2: 缓存与性能配置（Week 2）

**优先级：🟠 中高**

| Python 配置项 | Go 配置项 | 所属结构 |
|--------------|----------|---------|
| `CACHE_BACKEND_TYPE` | `Cache.BackendType` | CacheConfig |
| `CACHE_BACKEND_URL` | `Cache.BackendURL` | CacheConfig |
| `BIG_MEMORY_MODE` | `Performance.BigMemoryMode` | PerformanceConfig |
| `MEMORY_GC_INTERVAL` | `Performance.MemoryGCInterval` | PerformanceConfig |

**实现任务**：
- [ ] 缓存配置结构体
- [ ] 性能配置结构体
- [ ] 动态系统配置计算（`GetSystemConf()`）

### 3.3 Phase 3: 媒体与站点配置（Week 3）

**优先级：🟡 中**

| Python 配置项 | Go 配置项 | 所属结构 |
|--------------|----------|---------|
| `TMDB_API_KEY` | `TMDB.APIKey` | TMDBConfig |
| `TMDB_IMAGE_DOMAIN` | `TMDB.ImageDomain` | TMDBConfig |
| `SEARCH_SOURCE` | `Media.SearchSource` | MediaConfig |
| `RECOGNIZE_SOURCE` | `Media.RecognizeSource` | MediaConfig |
| `SITEDATA_REFRESH_INTERVAL` | `Site.DataRefreshInterval` | SiteConfig |

### 3.4 Phase 4: 插件与扩展配置（Week 4）

**优先级：🟢 低**

| Python 配置项 | Go 配置项 | 所属结构 |
|--------------|----------|---------|
| `PLUGIN_MARKET` | `Plugin.Market` | PluginConfig |
| `PLUGIN_AUTO_RELOAD` | `Plugin.AutoReload` | PluginConfig |
| `COOKIECLOUD_*` | `CookieCloud.*` | CookieCloudConfig |

---

## 4. 实现计划

### 4.1 Week 1: 基础架构

**目标**：完成配置加载器和核心配置

- [ ] **Day 1-2**：设计配置结构体
  - 创建 `internal/infrastructure/config/` 目录
  - 定义 `Config` 主结构体
  - 定义 `AppConfig`、`DatabaseConfig`、`SecurityConfig`

- [ ] **Day 3-4**：实现配置加载器
  - 集成 `viper` + `godotenv`
  - 实现环境变量读取
  - 实现默认值设置
  - 实现 `.env` 文件加载

- [ ] **Day 5**：配置校验
  - 实现配置项校验逻辑
  - 自动生成 `SECRET_KEY` 和 `API_TOKEN`
  - 校验必填项

- [ ] **Day 6-7**：测试与文档
  - 单元测试（覆盖率 > 80%）
  - 集成测试
  - 编写使用文档

### 4.2 Week 2: 缓存与性能

**目标**：完成缓存和性能配置

- [ ] **Day 1-2**：缓存配置
  - 定义 `CacheConfig`
  - 实现缓存后端选择逻辑

- [ ] **Day 3-4**：性能配置
  - 定义 `PerformanceConfig`
  - 实现 `GetSystemConf()` 动态计算

- [ ] **Day 5-7**：运行时配置更新
  - 实现 `UpdateSetting()` 方法
  - 实现 `.env` 文件写入
  - 实现配置热更新（可选）

### 4.3 Week 3: 媒体与站点

**目标**：完成媒体、TMDB、站点配置

- [ ] **Day 1-3**：媒体配置
  - 定义 `MediaConfig`、`TMDBConfig`、`SiteConfig`
  - 实现配置加载

- [ ] **Day 4-7**：其他配置
  - 定义 `DownloadConfig`、`TransferConfig`、`SubscribeConfig`
  - 实现配置加载

### 4.4 Week 4: 插件与扩展

**目标**：完成插件、CookieCloud、网络配置

- [ ] **Day 1-3**：插件配置
  - 定义 `PluginConfig`
  - 实现插件市场地址解析

- [ ] **Day 4-5**：CookieCloud 配置
  - 定义 `CookieCloudConfig`

- [ ] **Day 6-7**：网络配置
  - 定义 `NetworkConfig`
  - 实现代理配置解析（`GetProxy()`）

---

## 5. 测试策略

### 5.1 单元测试

**config_test.go**：

```go
func TestConfigLoad(t *testing.T) {
    // 测试默认配置加载
    cfg, err := NewLoader().Load()
    assert.NoError(t, err)
    assert.Equal(t, "MoviePilot", cfg.App.ProjectName)
    assert.Equal(t, 3001, cfg.App.Port)
}

func TestConfigUpdate(t *testing.T) {
    cfg, _ := NewLoader().Load()
    
    // 测试配置更新
    err := cfg.UpdateSetting("PORT", 8080)
    assert.NoError(t, err)
    assert.Equal(t, 8080, cfg.App.Port)
}

func TestSystemConf(t *testing.T) {
    cfg, _ := NewLoader().Load()
    
    // 测试普通模式
    cfg.Performance.BigMemoryMode = false
    sysConf := cfg.GetSystemConf()
    assert.Equal(t, 100, sysConf.Torrents)
    
    // 测试大内存模式
    cfg.Performance.BigMemoryMode = true
    sysConf = cfg.GetSystemConf()
    assert.Equal(t, 200, sysConf.Torrents)
}
```

### 5.2 集成测试

**integration_test.go**：

```go
func TestConfigLoadFromEnv(t *testing.T) {
    // 创建临时 .env 文件
    envContent := `
PROJECT_NAME=TestApp
PORT=9000
DEBUG=true
DB_TYPE=postgres
`
    tmpFile := createTempEnvFile(t, envContent)
    defer os.Remove(tmpFile)
    
    // 加载配置
    os.Setenv("ENV_FILE_PATH", tmpFile)
    cfg, err := NewLoader().Load()
    
    assert.NoError(t, err)
    assert.Equal(t, "TestApp", cfg.App.ProjectName)
    assert.Equal(t, 9000, cfg.App.Port)
    assert.True(t, cfg.App.Debug)
    assert.Equal(t, "postgres", cfg.Database.Type)
}
```

---

## 6. 与 Python 的差异

| 特性 | Python (Pydantic) | Go (Viper) |
|------|-------------------|-----------|
| 配置来源 | BaseSettings 自动读取 | 手动集成 viper + godotenv |
| 类型校验 | Pydantic 自动校验 | 需手动实现校验逻辑 |
| 默认值 | Field(default=...) | viper.SetDefault() |
| 运行时更新 | 直接修改属性 | 需加锁 + 写文件 |
| 动态属性 | @property 装饰器 | 方法实现 |

---

## 7. 注意事项

### 7.1 配置安全

- ✅ `SECRET_KEY` 和 `API_TOKEN` 必须自动生成
- ✅ 敏感配置（密码、Token）不记录到日志
- ✅ 配置文件权限设置为 600

### 7.2 配置兼容性

- ✅ 保持与 Python 版本的配置项名称一致
- ✅ 支持从旧版本配置文件迁移
- ✅ 配置项废弃时提供警告

### 7.3 性能优化

- ✅ 使用读写锁支持并发读取
- ✅ 缓存动态计算的配置（如路径）
- ✅ 避免频繁写入 .env 文件

---

## 8. 后续优化

### 8.1 配置热更新

通过文件监控（fsnotify）实现配置文件变更自动重载：

```go
func (c *Config) WatchConfigFile() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(c.getEnvPath())
    
    go func() {
        for event := range watcher.Events {
            if event.Op&fsnotify.Write == fsnotify.Write {
                c.Reload()
            }
        }
    }()
}
```

### 8.2 配置中心集成

支持从配置中心（如 etcd、Consul）读取配置：

```go
type ConfigSource interface {
    Load() (*Config, error)
    Watch(callback func(*Config))
}

type EtcdConfigSource struct {
    client *etcd.Client
}
```

---

## 9. 检查清单

### 开发阶段

- [ ] 所有配置项已定义
- [ ] 默认值已设置
- [ ] 配置校验已实现
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过

### 部署阶段

- [ ] `.env.example` 文件已创建
- [ ] 配置文档已编写
- [ ] 迁移指南已编写
- [ ] Docker 环境变量已配置

---

**相关文档**：
- [core-migration-app-core.md](./core-migration-app-core.md)
- [startup-migration.md](./startup-migration.md)
- [migration-overview.md](./migration-overview.md)

**最后更新**：2025-11-26  
**状态**：设计阶段