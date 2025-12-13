package meta

import (
	"regexp"
	"strings"
	"sync"

	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// SystemConfigOper 系统配置操作接口
type SystemConfigOper interface {
	Get(key string) (interface{}, error)
}

// CustomizationMatcher 识别自定义占位符
type CustomizationMatcher struct {
	systemConfig SystemConfigOper
	cache        cache.CacheBackend
	mutex        sync.RWMutex

	customization     string
	customizationList []string
	customSeparator   string
}

var (
	once     sync.Once
	instance *CustomizationMatcher
)

// NewCustomizationMatcher 创建CustomizationMatcher实例
func NewCustomizationMatcher(systemConfig SystemConfigOper, cache cache.CacheBackend) *CustomizationMatcher {
	once.Do(func() {
		instance = &CustomizationMatcher{
			systemConfig: systemConfig,
			cache:        cache,
		}
	})
	return instance
}

// GetInstance 获取CustomizationMatcher单例实例
func GetInstance() *CustomizationMatcher {
	if instance == nil {
		logger.Warn("CustomizationMatcher instance not initialized, creating with default values", zap.String("component", "CustomizationMatcher"))
		// 创建默认缓存实例
		defaultCache := cache.Cache("ttl", 100, 300) // 100个条目，5分钟TTL
		instance = &CustomizationMatcher{
			cache: defaultCache,
		}
	}
	return instance
}

// Match 匹配自定义占位符
func (cm *CustomizationMatcher) Match(title string) string {
	if title == "" {
		return ""
	}

	// 尝试从缓存获取结果
	cacheKey := cm.generateCacheKey(title)
	if cachedResult, found, _ := cm.cache.Get(cacheKey, ""); found {
		if result, ok := cachedResult.(string); ok {
			logger.Debug("Cache hit for customization match", zap.String("title", title), zap.String("result", result))
			return result
		}
	}

	// 构建自定义正则表达式
	customizationRegex, err := cm.buildCustomizationRegex()
	if err != nil || customizationRegex == nil {
		cm.cache.Set(cacheKey, "", 0, "")
		return ""
	}

	// 处理重复多次的情况，保留先后顺序（按添加自定义占位符的顺序）
	uniqueCustomization := make(map[string]int)

	// 查找所有匹配项，包括捕获组
	matches := customizationRegex.FindAllStringSubmatch(title, -1)
	for _, match := range matches {
		// 遍历所有捕获组（第一个是完整匹配，后面是分组匹配）
		for i := 1; i < len(match); i++ {
			item := match[i]
			if item != "" && uniqueCustomization[item] == 0 {
				// 记录匹配项
				uniqueCustomization[item] = i - 1
			}
		}
	}

	// 按原始自定义列表顺序排序
	var sortedMatches []string
	for _, item := range cm.customizationList {
		if _, exists := uniqueCustomization[item]; exists {
			sortedMatches = append(sortedMatches, item)
		}
	}

	separator := cm.customSeparator
	if separator == "" {
		separator = "@"
	}

	result := strings.Join(sortedMatches, separator)

	// 缓存结果
	cm.cache.Set(cacheKey, result, 300, "") // 5分钟TTL
	logger.Debug("Cache miss for customization match", zap.String("title", title), zap.String("result", result))

	return result
}

// 构建自定义占位符正则表达式
func (cm *CustomizationMatcher) buildCustomizationRegex() (*regexp.Regexp, error) {
	cm.mutex.RLock()
	customization := cm.customization
	cm.mutex.RUnlock()

	if customization != "" {
		return regexp.Compile(customization)
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 双重检查锁定
	if cm.customization != "" {
		return regexp.Compile(cm.customization)
	}

	// 从配置获取自定义占位符
	customizationConfig, err := cm.systemConfig.Get("Customization")
	if err != nil || customizationConfig == nil {
		logger.Error("Failed to get customization config", zap.Error(err))
		return nil, err
	}

	var customizationList []string

	switch v := customizationConfig.(type) {
	case string:
		// 处理字符串格式，支持换行和|分隔
		customStr := strings.ReplaceAll(v, "\n", ";")
		customStr = strings.ReplaceAll(customStr, "|", ";")
		customStr = strings.Trim(customStr, ";")
		if customStr != "" {
			customizationList = strings.Split(customStr, ";")
		}
	case []string:
		customizationList = v
	}

	if len(customizationList) == 0 {
		return nil, nil
	}

	// 过滤空项
	var filteredList []string
	for _, item := range customizationList {
		item = strings.TrimSpace(item)
		if item != "" {
			filteredList = append(filteredList, item)
		}
	}

	if len(filteredList) == 0 {
		return nil, nil
	}

	// 保存原始列表用于后续处理
	cm.customizationList = filteredList
	// 构建正则表达式，为每个自定义项创建一个捕获组
	var regexParts []string
	for _, item := range filteredList {
		regexParts = append(regexParts, "("+item+")")
	}
	customizationPattern := strings.Join(regexParts, "|")
	cm.customization = customizationPattern

	return regexp.Compile(customizationPattern)
}

// 设置自定义分隔符
func (cm *CustomizationMatcher) SetCustomSeparator(separator string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.customSeparator = separator
}

// 生成缓存键
func (cm *CustomizationMatcher) generateCacheKey(title string) string {
	return "customization:match:" + title
}

// ClearCache 清空缓存
func (cm *CustomizationMatcher) ClearCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.cache.Clear("")
	cm.customization = ""      // 同时清空编译的正则表达式，下次会重新生成
	cm.customizationList = nil // 清空自定义列表
	logger.Info("Customization matcher cache cleared", zap.String("component", "CustomizationMatcher"))
}
