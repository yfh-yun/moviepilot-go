package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// UpdateSetting 更新单个配置项
func (c *Config) UpdateSetting(key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log := logger.GetLogger()

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
		log.Error("failed to write config to env file", zap.String("key", key), zap.Error(err))
		return err
	}

	log.Info("config updated successfully", zap.String("key", key), zap.Any("value", convertedValue))
	return nil
}

// UpdateSettings 批量更新配置
func (c *Config) UpdateSettings(settings map[string]any) map[string]error {
	results := make(map[string]error)
	for key, value := range settings {
		results[key] = c.UpdateSetting(key, value)
	}
	return results
}

// isValidKey 检查配置项是否有效
func (c *Config) isValidKey(key string) bool {
	// 将配置结构体转换为map，然后检查key是否存在
	// 这里使用反射实现，遍历所有配置结构体的字段
	configMap := c.toMap()
	_, exists := configMap[key]
	return exists
}

// convertValue 转换配置值类型
func (c *Config) convertValue(key string, value any) (any, error) {
	// 获取配置项的期望类型
	expectedType, err := c.getConfigValueType(key)
	if err != nil {
		return nil, err
	}

	// 转换类型
	valueStr := fmt.Sprintf("%v", value)
	switch expectedType {
	case "string":
		return valueStr, nil
	case "int":
		var intValue int
		_, err := fmt.Sscanf(valueStr, "%d", &intValue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s to int: %w", valueStr, err)
		}
		return intValue, nil
	case "bool":
		var boolValue bool
		_, err := fmt.Sscanf(valueStr, "%t", &boolValue)
		if err != nil {
			// 支持"true"/"false"字符串
			valueStr = strings.ToLower(valueStr)
			if valueStr == "true" {
				return true, nil
			} else if valueStr == "false" {
				return false, nil
			}
			return nil, fmt.Errorf("failed to convert %s to bool", valueStr)
		}
		return boolValue, nil
	case "float64":
		var floatValue float64
		_, err := fmt.Sscanf(valueStr, "%f", &floatValue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s to float64: %w", valueStr, err)
		}
		return floatValue, nil
	default:
		return value, nil
	}
}

// setConfigValue 设置配置值到内存中
func (c *Config) setConfigValue(key string, value any) error {
	// 使用反射设置配置值
	// 这里需要根据key找到对应的配置结构体和字段
	// 简化实现，仅支持顶层配置项

	// 转换为大写，因为配置项都是大写的
	key = strings.ToUpper(key)

	// 配置项映射，将短键名映射到对应的配置结构体和字段
	keyMap := map[string]struct {
		structName string
		fieldName  string
	}{
		// 应用配置
		"PROJECT_NAME":  {"APP", "ProjectName"},
		"APP_DOMAIN":    {"APP", "AppDomain"},
		"API_V1_STR":    {"APP", "APIV1Str"},
		"FRONTEND_PATH": {"APP", "FrontendPath"},
		"TZ":            {"APP", "TZ"},
		"HOST":          {"APP", "Host"},
		"PORT":          {"APP", "Port"},
		"NGINX_PORT":    {"APP", "NginxPort"},
		"CONFIG_DIR":    {"APP", "ConfigDir"},
		"DEBUG":         {"APP", "Debug"},
		"DEV":           {"APP", "Dev"},
		"ADVANCED_MODE": {"APP", "AdvancedMode"},

		// 数据库配置
		"DB_TYPE":                    {"DATABASE", "Type"},
		"DB_ECHO":                    {"DATABASE", "Echo"},
		"DB_TIMEOUT":                 {"DATABASE", "Timeout"},
		"DB_WAL_ENABLE":              {"DATABASE", "WALEnable"},
		"DB_POOL_TYPE":               {"DATABASE", "PoolType"},
		"DB_POOL_PRE_PING":           {"DATABASE", "PoolPrePing"},
		"DB_POOL_RECYCLE":            {"DATABASE", "PoolRecycle"},
		"DB_POOL_TIMEOUT":            {"DATABASE", "PoolTimeout"},
		"DB_SQLITE_POOL_SIZE":        {"DATABASE", "SQLitePoolSize"},
		"DB_SQLITE_MAX_OVERFLOW":     {"DATABASE", "SQLiteMaxOverflow"},
		"DB_POSTGRESQL_HOST":         {"DATABASE", "PostgreSQLHost"},
		"DB_POSTGRESQL_PORT":         {"DATABASE", "PostgreSQLPort"},
		"DB_POSTGRESQL_DATABASE":     {"DATABASE", "PostgreSQLDatabase"},
		"DB_POSTGRESQL_USERNAME":     {"DATABASE", "PostgreSQLUsername"},
		"DB_POSTGRESQL_PASSWORD":     {"DATABASE", "PostgreSQLPassword"},
		"DB_POSTGRESQL_POOL_SIZE":    {"DATABASE", "PostgreSQLPoolSize"},
		"DB_POSTGRESQL_MAX_OVERFLOW": {"DATABASE", "PostgreSQLMaxOverflow"},
	}

	// 检查是否在映射中
	if mapping, exists := keyMap[key]; exists {
		// 找到对应的配置结构体和字段
		var target any
		switch mapping.structName {
		case "APP":
			target = c.App
		case "DATABASE":
			target = c.Database
		case "CACHE":
			target = c.Cache
		case "SECURITY":
			target = c.Security
		case "MEDIA":
			target = c.Media
		case "TMDB":
			target = c.TMDB
		case "SITE":
			target = c.Site
		case "DOWNLOAD":
			target = c.Download
		case "COOKIECLOUD":
			target = c.CookieCloud
		case "TRANSFER":
			target = c.Transfer
		case "PLUGIN":
			target = c.Plugin
		case "PERFORMANCE":
			target = c.Performance
		case "SCHEDULER":
			target = c.Scheduler
		case "SUBSCRIBE":
			target = c.Subscribe
		case "NETWORK":
			target = c.Network
		default:
			return fmt.Errorf("unknown config struct: %s", mapping.structName)
		}

		// 使用反射设置字段值
		return c.setStructField(target, mapping.fieldName, value)
	}

	// 尝试旧的前缀匹配方式
	configFields := []struct {
		name string
		ptr  any
	}{
		{"APP", c.App},
		{"DATABASE", c.Database},
		{"CACHE", c.Cache},
		{"SECURITY", c.Security},
		{"MEDIA", c.Media},
		{"TMDB", c.TMDB},
		{"SITE", c.Site},
		{"DOWNLOAD", c.Download},
		{"COOKIECLOUD", c.CookieCloud},
		{"TRANSFER", c.Transfer},
		{"PLUGIN", c.Plugin},
		{"PERFORMANCE", c.Performance},
		{"SCHEDULER", c.Scheduler},
		{"SUBSCRIBE", c.Subscribe},
		{"NETWORK", c.Network},
	}

	for _, field := range configFields {
		if strings.HasPrefix(key, field.name+"_") {
			// 提取字段名，去掉前缀
			fieldName := strings.TrimPrefix(key, field.name+"_")
			// 使用反射设置字段值
			return c.setStructField(field.ptr, fieldName, value)
		}
	}

	return fmt.Errorf("failed to find config field for key: %s", key)
}

// setStructField 使用反射设置结构体字段值
func (c *Config) setStructField(ptr any, fieldName string, value any) error {
	v := reflect.ValueOf(ptr).Elem()
	f := v.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, fieldName)
	})

	if !f.IsValid() {
		return fmt.Errorf("field not found: %s", fieldName)
	}

	if !f.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// 转换值类型
	valueV := reflect.ValueOf(value)
	if valueV.Type() == f.Type() {
		f.Set(valueV)
		return nil
	}

	// 尝试类型转换
	switch f.Kind() {
	case reflect.String:
		f.SetString(fmt.Sprintf("%v", value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, ok := value.(int); ok {
			f.SetInt(int64(intVal))
		} else {
			return fmt.Errorf("cannot convert %v to int for field %s", value, fieldName)
		}
	case reflect.Bool:
		if boolVal, ok := value.(bool); ok {
			f.SetBool(boolVal)
		} else {
			return fmt.Errorf("cannot convert %v to bool for field %s", value, fieldName)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, ok := value.(float64); ok {
			f.SetFloat(floatVal)
		} else {
			return fmt.Errorf("cannot convert %v to float for field %s", value, fieldName)
		}
	default:
		return fmt.Errorf("unsupported field type: %s for field %s", f.Type(), fieldName)
	}

	return nil
}

// writeToEnvFile 写入配置到 .env 文件
func (c *Config) writeToEnvFile(key string, value any) error {
	envPath := c.getEnvPath()

	// 读取现有 .env 文件
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		// 如果文件不存在，创建一个新的
		envMap = make(map[string]string)
	}

	// 更新配置项
	envMap[key] = fmt.Sprintf("%v", value)

	// 写回文件
	return godotenv.Write(envMap, envPath)
}

// getEnvPath 获取 .env 文件路径
func (c *Config) getEnvPath() string {
	// 优先级：环境变量 > Docker > 二进制 > 开发环境
	if envPath := os.Getenv("ENV_FILE_PATH"); envPath != "" {
		return envPath
	}

	if isDocker() {
		return "/config/app.env"
	}

	if isFrozen() {
		execPath, _ := os.Executable()
		return fmt.Sprintf("%s/config/app.env", execPath)
	}

	return "./config/app.env"
}

// getConfigValueType 获取配置项的类型
func (c *Config) getConfigValueType(key string) (string, error) {
	// 简化实现，仅支持顶层配置项
	configMap := c.toMap()
	if val, exists := configMap[key]; exists {
		return reflect.TypeOf(val).Kind().String(), nil
	}
	return "", fmt.Errorf("config key not found: %s", key)
}

// toMap 将配置转换为map，用于检查配置项是否存在
func (c *Config) toMap() map[string]any {
	// 简化实现，仅返回配置项名称，不返回实际值
	// 实际使用时，应该遍历所有配置结构体的字段，生成完整的配置项列表

	// 这里返回一个静态的配置项列表，实际项目中应该通过反射动态生成
	configMap := make(map[string]any)

	// 应用配置
	configMap["PROJECT_NAME"] = ""
	configMap["APP_DOMAIN"] = ""
	configMap["API_V1_STR"] = ""
	configMap["FRONTEND_PATH"] = ""
	configMap["TZ"] = ""
	configMap["HOST"] = ""
	configMap["PORT"] = 0
	configMap["NGINX_PORT"] = 0
	configMap["CONFIG_DIR"] = ""
	configMap["DEBUG"] = false
	configMap["DEV"] = false
	configMap["ADVANCED_MODE"] = false

	// 数据库配置
	configMap["DB_TYPE"] = ""
	configMap["DB_ECHO"] = false
	configMap["DB_TIMEOUT"] = 0
	configMap["DB_WAL_ENABLE"] = false
	configMap["DB_POOL_TYPE"] = ""
	configMap["DB_POOL_PRE_PING"] = false
	configMap["DB_POOL_RECYCLE"] = 0
	configMap["DB_POOL_TIMEOUT"] = 0
	configMap["DB_SQLITE_POOL_SIZE"] = 0
	configMap["DB_SQLITE_MAX_OVERFLOW"] = 0
	configMap["DB_POSTGRESQL_HOST"] = ""
	configMap["DB_POSTGRESQL_PORT"] = 0
	configMap["DB_POSTGRESQL_DATABASE"] = ""
	configMap["DB_POSTGRESQL_USERNAME"] = ""
	configMap["DB_POSTGRESQL_PASSWORD"] = ""
	configMap["DB_POSTGRESQL_POOL_SIZE"] = 0
	configMap["DB_POSTGRESQL_MAX_OVERFLOW"] = 0

	// 缓存配置
	configMap["CACHE_BACKEND_TYPE"] = ""
	configMap["CACHE_BACKEND_URL"] = ""
	configMap["CACHE_REDIS_MAXMEMORY"] = ""
	configMap["GLOBAL_IMAGE_CACHE"] = false
	configMap["GLOBAL_IMAGE_CACHE_DAYS"] = 0
	configMap["TEMP_FILE_DAYS"] = 0
	configMap["META_CACHE_EXPIRE"] = 0

	// 安全配置
	configMap["SECRET_KEY"] = ""
	configMap["RESOURCE_SECRET_KEY"] = ""
	configMap["ALLOWED_HOSTS"] = []string{}
	configMap["ACCESS_TOKEN_EXPIRE_MINUTES"] = 0
	configMap["RESOURCE_ACCESS_TOKEN_EXPIRE_SECONDS"] = 0
	configMap["SUPERUSER"] = ""
	configMap["SUPERUSER_PASSWORD"] = ""
	configMap["AUXILIARY_AUTH_ENABLE"] = false
	configMap["API_TOKEN"] = ""
	configMap["AUTH_SITE"] = ""
	configMap["SECURITY_IMAGE_DOMAINS"] = []string{}
	configMap["SECURITY_IMAGE_SUFFIXES"] = []string{}

	// 媒体配置
	configMap["SEARCH_SOURCE"] = ""
	configMap["RECOGNIZE_SOURCE"] = ""
	configMap["MOVIE_RENAME_FORMAT"] = ""
	configMap["TV_RENAME_FORMAT"] = ""

	// TMDB配置
	configMap["TMDB_API_KEY"] = ""
	configMap["TMDB_IMAGE_DOMAIN"] = ""
	configMap["TMDB_LANGUAGE"] = ""
	configMap["TMDB_REGION"] = ""
	configMap["TMDB_PROXY_ENABLE"] = false
	configMap["TMDB_PROXY_URL"] = ""

	// 站点配置
	configMap["SITEDATA_REFRESH_INTERVAL"] = 0
	configMap["OCR_HOST"] = ""
	configMap["OCR_API_KEY"] = ""
	configMap["SITE_MAX_CONCURRENT_TASKS"] = 0
	configMap["SITE_REQUEST_TIMEOUT"] = 0
	configMap["SITE_RETRY_TIMES"] = 0
	configMap["SITE_RETRY_INTERVAL"] = 0

	// 下载配置
	configMap["TORRENT_TAG"] = ""
	configMap["DOWNLOAD_SUBTITLE"] = false
	configMap["MAX_CONCURRENT_DOWNLOADS"] = 0
	configMap["DOWNLOAD_PATH"] = ""
	configMap["TEMP_PATH"] = ""

	// 整理配置
	configMap["TRANSFER_ENABLE"] = false
	configMap["MOVIE_PATH"] = ""
	configMap["TV_PATH"] = ""
	configMap["AUTO_TRANSFER"] = false
	configMap["DELETE_SOURCE"] = false

	// 插件配置
	configMap["PLUGIN_MARKET"] = ""
	configMap["PLUGIN_AUTO_RELOAD"] = false
	configMap["PLUGIN_MAX_WORKERS"] = 0
	configMap["PLUGIN_TIMEOUT"] = 0

	// CookieCloud配置
	configMap["COOKIECLOUD_HOST"] = ""
	configMap["COOKIECLOUD_KEY"] = ""
	configMap["COOKIECLOUD_PASSWORD"] = ""
	configMap["COOKIECLOUD_INTERVAL"] = 0
	configMap["COOKIECLOUD_ENABLE"] = false

	// 性能配置
	configMap["BIG_MEMORY_MODE"] = false
	configMap["ENCODING_DETECTION_PERFORMANCE_MODE"] = false
	configMap["ENCODING_DETECTION_MIN_CONFIDENCE"] = 0.0
	configMap["MEMORY_GC_INTERVAL"] = 0

	// 网络配置
	configMap["PROXY_HOST"] = ""
	configMap["PROXY_PORT"] = 0
	configMap["PROXY_USERNAME"] = ""
	configMap["PROXY_PASSWORD"] = ""
	configMap["DOH_ENABLE"] = false
	configMap["DOH_DOMAINS"] = []string{}

	// 调度器配置
	configMap["SCHEDULER_MAX_CONCURRENT_TASKS"] = 0
	configMap["SCHEDULER_TASK_TIMEOUT"] = 0

	// 订阅配置
	configMap["SUBSCRIBE_CHECK_INTERVAL"] = 0
	configMap["SUBSCRIBE_MAX_CONCURRENT_CHECKS"] = 0
	configMap["SUBSCRIBE_NOTIFICATION_ENABLED"] = false

	return configMap
}
