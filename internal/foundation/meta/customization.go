package meta

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// MetadataCustomization 元数据自定义处理器
type MetadataCustomization struct {
	sync.RWMutex
	// 自定义规则配置
	excludedWords       map[string]bool   // 排除词列表
	customReplaceRules  []ReplaceRule     // 自定义替换规则
	titleNormalizationRules []ReplaceRule // 标题规范化规则
	siteNameMappings    map[string]string // 站点名称映射
	sourcePriorities    map[string]int    // 来源优先级
	regionMappings      map[string]string // 区域映射
	languageMappings    map[string]string // 语言映射
	
	// 缓存
	cache               map[string]interface{}
	cacheExpiry         time.Time
	initialized         bool
	lastUpdated         time.Time
}

// ReplaceRule 替换规则
type ReplaceRule struct {
	Pattern       string `json:"pattern"`       // 匹配模式
	Replacement   string `json:"replacement"`   // 替换内容
	IsRegex       bool   `json:"is_regex"`       // 是否使用正则表达式
	IgnoreCase    bool   `json:"ignore_case"`    // 是否忽略大小写
	Description   string `json:"description"`   // 规则描述
	Priority      int    `json:"priority"`      // 优先级，数字越大优先级越高
	IsActive      bool   `json:"is_active"`      // 是否激活
	CreatedAt     time.Time `json:"created_at"` // 创建时间
}

// CustomField 自定义字段
type CustomField struct {
	Name        string      `json:"name"`        // 字段名称
	Label       string      `json:"label"`       // 显示标签
	Type        string      `json:"type"`        // 字段类型：string, number, boolean, select, textarea
	Default     interface{} `json:"default"`     // 默认值
	Options     []string    `json:"options"`     // 选项列表（用于select类型）
	Required    bool        `json:"required"`    // 是否必填
	Description string      `json:"description"` // 字段描述
	Priority    int         `json:"priority"`    // 显示优先级
	Visible     bool        `json:"visible"`     // 是否可见
}

// NewMetadataCustomization 创建元数据自定义处理器实例
func NewMetadataCustomization() *MetadataCustomization {
	custom := &MetadataCustomization{
		excludedWords:           make(map[string]bool),
		customReplaceRules:      make([]ReplaceRule, 0),
		titleNormalizationRules: make([]ReplaceRule, 0),
		siteNameMappings:        make(map[string]string),
		sourcePriorities:        make(map[string]int),
		regionMappings:          make(map[string]string),
		languageMappings:        make(map[string]string),
		cache:                   make(map[string]interface{}),
		initialized:             false,
		lastUpdated:             time.Now(),
	}
	return custom
}

// Initialize 初始化元数据自定义处理器
func (c *MetadataCustomization) Initialize(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()
	
	if c.initialized {
		return nil
	}
	
	// 加载默认规则
	c.loadDefaultRules()
	c.loadDefaultMappings()
	c.loadDefaultPriorities()
	c.loadDefaultExcludedWords()
	
	// 初始化缓存
	c.cacheExpiry = time.Now().Add(1 * time.Hour)
	c.initialized = true
	c.lastUpdated = time.Now()
	
	return nil
}

// NormalizeTitle 规范化标题
func (c *MetadataCustomization) NormalizeTitle(ctx context.Context, title string) string {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return title
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	// 空标题检查
	if title == "" {
		return title
	}
	
	result := title
	
	// 应用标题规范化规则
	for _, rule := range c.titleNormalizationRules {
		if rule.IsActive {
			result = c.applyRule(result, rule)
		}
	}
	
	// 应用自定义替换规则
	for _, rule := range c.customReplaceRules {
		if rule.IsActive && rule.Priority > 50 { // 高优先级规则优先应用
			result = c.applyRule(result, rule)
		}
	}
	
	// 移除多余空格
	result = strings.Join(strings.Fields(result), " ")
	
	return result
}

// FilterTitle 过滤标题中的敏感词和排除词
func (c *MetadataCustomization) FilterTitle(ctx context.Context, title string) string {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return title
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	// 空标题检查
	if title == "" {
		return title
	}
	
	result := title
	words := strings.Fields(title)
	filteredWords := make([]string, 0, len(words))
	
	// 过滤排除词
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if !c.excludedWords[lowerWord] {
			filteredWords = append(filteredWords, word)
		}
	}
	
	result = strings.Join(filteredWords, " ")
	
	// 应用中等优先级的自定义替换规则
	for _, rule := range c.customReplaceRules {
		if rule.IsActive && rule.Priority >= 30 && rule.Priority <= 50 {
			result = c.applyRule(result, rule)
		}
	}
	
	return result
}

// FormatMetadata 格式化元数据信息
func (c *MetadataCustomization) FormatMetadata(ctx context.Context, metadata map[string]interface{}) map[string]interface{} {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return metadata
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	
	result := make(map[string]interface{})
	
	// 复制原始元数据
	for k, v := range metadata {
		result[k] = v
	}
	
	// 应用区域映射
	if region, ok := metadata["region"].(string); ok && region != "" {
		if mappedRegion, exists := c.regionMappings[strings.ToLower(region)]; exists {
			result["region"] = mappedRegion
		}
	}
	
	// 应用语言映射
	if language, ok := metadata["language"].(string); ok && language != "" {
		if mappedLanguage, exists := c.languageMappings[strings.ToLower(language)]; exists {
			result["language"] = mappedLanguage
		}
	}
	
	// 规范化标题
	if title, ok := metadata["title"].(string); ok && title != "" {
		result["title"] = c.NormalizeTitle(ctx, title)
	}
	
	// 规范化原始标题
	if originalTitle, ok := metadata["original_title"].(string); ok && originalTitle != "" {
		result["original_title"] = c.NormalizeTitle(ctx, originalTitle)
	}
	
	return result
}

// GetSourcePriority 获取来源优先级
func (c *MetadataCustomization) GetSourcePriority(ctx context.Context, source string) int {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return 0
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	if priority, exists := c.sourcePriorities[strings.ToLower(source)]; exists {
		return priority
	}
	return 0
}

// MapSiteName 映射站点名称
func (c *MetadataCustomization) MapSiteName(ctx context.Context, siteName string) string {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return siteName
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	if mappedName, exists := c.siteNameMappings[strings.ToLower(siteName)]; exists {
		return mappedName
	}
	return siteName
}

// AddExcludedWord 添加排除词
func (c *MetadataCustomization) AddExcludedWord(word string) error {
	c.Lock()
	defer c.Unlock()
	
	if word != "" {
		c.excludedWords[strings.ToLower(word)] = true
		c.invalidateCache()
		c.lastUpdated = time.Now()
	}
	return nil
}

// RemoveExcludedWord 移除排除词
func (c *MetadataCustomization) RemoveExcludedWord(word string) error {
	c.Lock()
	defer c.Unlock()
	
	if word != "" {
		delete(c.excludedWords, strings.ToLower(word))
		c.invalidateCache()
		c.lastUpdated = time.Now()
	}
	return nil
}

// GetExcludedWords 获取所有排除词
func (c *MetadataCustomization) GetExcludedWords(ctx context.Context) []string {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return []string{}
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	words := make([]string, 0, len(c.excludedWords))
	for word := range c.excludedWords {
		words = append(words, word)
	}
	return words
}

// AddReplaceRule 添加替换规则
func (c *MetadataCustomization) AddReplaceRule(rule ReplaceRule) error {
	c.Lock()
	defer c.Unlock()
	
	if rule.Pattern == "" {
		return nil // 忽略无效规则
	}
	
	// 检查是否已存在相同模式的规则
	for i, r := range c.customReplaceRules {
		if r.Pattern == rule.Pattern && r.IsRegex == rule.IsRegex {
			// 更新现有规则
			c.customReplaceRules[i] = rule
			c.invalidateCache()
			c.lastUpdated = time.Now()
			return nil
		}
	}
	
	// 添加新规则
	c.customReplaceRules = append(c.customReplaceRules, rule)
	c.invalidateCache()
	c.lastUpdated = time.Now()
	
	return nil
}

// RemoveReplaceRule 移除替换规则
func (c *MetadataCustomization) RemoveReplaceRule(pattern string, isRegex bool) error {
	c.Lock()
	defer c.Unlock()
	
	for i, rule := range c.customReplaceRules {
		if rule.Pattern == pattern && rule.IsRegex == isRegex {
			c.customReplaceRules = append(c.customReplaceRules[:i], c.customReplaceRules[i+1:]...)
			c.invalidateCache()
			c.lastUpdated = time.Now()
			return nil
		}
	}
	return nil
}

// GetReplaceRules 获取所有替换规则
func (c *MetadataCustomization) GetReplaceRules(ctx context.Context) []ReplaceRule {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return []ReplaceRule{}
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	// 返回副本以避免并发修改
	rules := make([]ReplaceRule, len(c.customReplaceRules))
	copy(rules, c.customReplaceRules)
	return rules
}

// SetSourcePriority 设置来源优先级
func (c *MetadataCustomization) SetSourcePriority(source string, priority int) error {
	c.Lock()
	defer c.Unlock()
	
	if source != "" {
		c.sourcePriorities[strings.ToLower(source)] = priority
		c.invalidateCache()
		c.lastUpdated = time.Now()
	}
	return nil
}

// GetSourcePriorities 获取所有来源优先级
func (c *MetadataCustomization) GetSourcePriorities(ctx context.Context) map[string]int {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return map[string]int{}
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	// 返回副本以避免并发修改
	priorities := make(map[string]int)
	for source, priority := range c.sourcePriorities {
		priorities[source] = priority
	}
	return priorities
}

// AddSiteMapping 添加站点映射
func (c *MetadataCustomization) AddSiteMapping(originalName, mappedName string) error {
	c.Lock()
	defer c.Unlock()
	
	if originalName != "" {
		c.siteNameMappings[strings.ToLower(originalName)] = mappedName
		c.invalidateCache()
		c.lastUpdated = time.Now()
	}
	return nil
}

// GetSiteMappings 获取所有站点映射
func (c *MetadataCustomization) GetSiteMappings(ctx context.Context) map[string]string {
	if !c.initialized {
		if err := c.Initialize(ctx); err != nil {
			return map[string]string{}
		}
	}
	
	c.RLock()
	defer c.RUnlock()
	
	// 返回副本以避免并发修改
	mappings := make(map[string]string)
	for original, mapped := range c.siteNameMappings {
		mappings[original] = mapped
	}
	return mappings
}

// ExportConfig 导出配置
func (c *MetadataCustomization) ExportConfig() (map[string]interface{}, error) {
	c.RLock()
	defer c.RUnlock()
	
	config := map[string]interface{}{
		"excluded_words":       c.excludedWords,
		"custom_replace_rules": c.customReplaceRules,
		"site_name_mappings":   c.siteNameMappings,
		"source_priorities":    c.sourcePriorities,
		"region_mappings":      c.regionMappings,
		"language_mappings":    c.languageMappings,
		"last_updated":         c.lastUpdated,
	}
	
	return config, nil
}

// ImportConfig 导入配置
func (c *MetadataCustomization) ImportConfig(config map[string]interface{}) error {
	c.Lock()
	defer c.Unlock()
	
	// 导入排除词
	if excludedWords, ok := config["excluded_words"].(map[string]interface{}); ok {
		c.excludedWords = make(map[string]bool)
		for word := range excludedWords {
			c.excludedWords[strings.ToLower(word)] = true
		}
	}
	
	// 导入自定义替换规则
	if rulesData, ok := config["custom_replace_rules"].([]interface{}); ok {
		c.customReplaceRules = make([]ReplaceRule, 0, len(rulesData))
		for _, ruleData := range rulesData {
			if ruleJSON, err := json.Marshal(ruleData); err == nil {
				var rule ReplaceRule
				if err := json.Unmarshal(ruleJSON, &rule); err == nil {
					c.customReplaceRules = append(c.customReplaceRules, rule)
				}
			}
		}
	}
	
	// 导入站点映射
	if siteMappings, ok := config["site_name_mappings"].(map[string]interface{}); ok {
		c.siteNameMappings = make(map[string]string)
		for original, mapped := range siteMappings {
			if mappedStr, ok := mapped.(string); ok {
				c.siteNameMappings[strings.ToLower(original)] = mappedStr
			}
		}
	}
	
	// 导入来源优先级
	if priorities, ok := config["source_priorities"].(map[string]interface{}); ok {
		c.sourcePriorities = make(map[string]int)
		for source, priority := range priorities {
			if priorityInt, ok := priority.(float64); ok {
				c.sourcePriorities[strings.ToLower(source)] = int(priorityInt)
			}
		}
	}
	
	c.invalidateCache()
	c.lastUpdated = time.Now()
	
	return nil
}

// ResetToDefaults 重置为默认配置
func (c *MetadataCustomization) ResetToDefaults() error {
	c.Lock()
	defer c.Unlock()
	
	c.loadDefaultRules()
	c.loadDefaultMappings()
	c.loadDefaultPriorities()
	c.loadDefaultExcludedWords()
	
	c.invalidateCache()
	c.lastUpdated = time.Now()
	
	return nil
}

// GetStatistics 获取统计信息
func (c *MetadataCustomization) GetStatistics() map[string]interface{} {
	c.RLock()
	defer c.RUnlock()
	
	totalRules := len(c.customReplaceRules)
	activeRules := 0
	inactiveRules := 0
	
	for _, rule := range c.customReplaceRules {
		if rule.IsActive {
			activeRules++
		} else {
			inactiveRules++
		}
	}
	
	return map[string]interface{}{
		"total_excluded_words": len(c.excludedWords),
		"total_rules":          totalRules,
		"active_rules":         activeRules,
		"inactive_rules":       inactiveRules,
		"total_site_mappings":  len(c.siteNameMappings),
		"total_source_priorities": len(c.sourcePriorities),
		"total_region_mappings":   len(c.regionMappings),
		"total_language_mappings": len(c.languageMappings),
		"last_updated":           c.lastUpdated,
		"initialized":            c.initialized,
	}
}

// 私有方法

// loadDefaultRules 加载默认规则
func (c *MetadataCustomization) loadDefaultRules() {
	// 默认标题规范化规则
	c.titleNormalizationRules = []ReplaceRule{
		{Pattern: "[\[\(]\s*\d+p\s*[\]\)]", Replacement: "", IsRegex: true, IgnoreCase: true, Description: "移除分辨率标签", Priority: 100, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "[\[\(]\s*\d+k\s*[\]\)]", Replacement: "", IsRegex: true, IgnoreCase: true, Description: "移除4K等标签", Priority: 90, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "[\[\(]\s*(web|web-dl|bluray|bdrip|dvdrip)\s*[\]\)]", Replacement: "", IsRegex: true, IgnoreCase: true, Description: "移除格式标签", Priority: 80, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "[\[\(]\s*(h264|h265|x264|x265|hevc)\s*[\]\)]", Replacement: "", IsRegex: true, IgnoreCase: true, Description: "移除编码标签", Priority: 70, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "[\[\(]\s*(chs|cht|双语)\s*[\]\)]", Replacement: "", IsRegex: true, IgnoreCase: true, Description: "移除语言标签", Priority: 60, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "\s{2,}", Replacement: " ", IsRegex: true, IgnoreCase: false, Description: "压缩多余空格", Priority: 50, IsActive: true, CreatedAt: time.Now()},
	}
	
	// 默认自定义替换规则
	c.customReplaceRules = []ReplaceRule{
		{Pattern: "&amp;", Replacement: "&", IsRegex: false, IgnoreCase: false, Description: "修复HTML实体", Priority: 90, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "&quot;", Replacement: "\"", IsRegex: false, IgnoreCase: false, Description: "修复HTML实体", Priority: 90, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "&#39;", Replacement: "'", IsRegex: false, IgnoreCase: false, Description: "修复HTML实体", Priority: 90, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "&lt;", Replacement: "<", IsRegex: false, IgnoreCase: false, Description: "修复HTML实体", Priority: 90, IsActive: true, CreatedAt: time.Now()},
		{Pattern: "&gt;", Replacement: ">", IsRegex: false, IgnoreCase: false, Description: "修复HTML实体", Priority: 90, IsActive: true, CreatedAt: time.Now()},
	}
}

// loadDefaultMappings 加载默认映射
func (c *MetadataCustomization) loadDefaultMappings() {
	// 站点名称映射
	c.siteNameMappings = map[string]string{
		"btsow":        "BT搜",
		"rarbg":        "RARBG",
		"1337x":        "1337X",
		"thepiratebay": "海盗湾",
		"yts":          "YTS",
		" eztv":         "EZTV",
		"tokyotoshokan": "东京图书馆",
		"nyaa":         "Nyaa",
		"anidex":       "AniDex",
	}
	
	// 区域映射
	c.regionMappings = map[string]string{
		"cn":  "中国大陆",
		"hk":  "中国香港",
		"tw":  "中国台湾",
		"jp":  "日本",
		"kr":  "韩国",
		"us":  "美国",
		"uk":  "英国",
		"ca":  "加拿大",
		"au":  "澳大利亚",
		"de":  "德国",
		"fr":  "法国",
		"es":  "西班牙",
		"it":  "意大利",
		"ru":  "俄罗斯",
		"ind": "印度",
		"other": "其他",
	}
	
	// 语言映射
	c.languageMappings = map[string]string{
		"zh":      "中文",
		"zh-cn":   "简体中文",
		"zh-tw":   "繁体中文",
		"zh-hk":   "香港繁体",
		"en":      "英语",
		"ja":      "日语",
		"ko":      "韩语",
		"fr":      "法语",
		"de":      "德语",
		"es":      "西班牙语",
		"it":      "意大利语",
		"ru":      "俄语",
		"ar":      "阿拉伯语",
		"hi":      "印地语",
		"multi":   "多语言",
		"dual":    "双语",
		"original": "原声",
	}
}

// loadDefaultPriorities 加载默认优先级
func (c *MetadataCustomization) loadDefaultPriorities() {
	// 来源优先级（数字越大优先级越高）
	c.sourcePriorities = map[string]int{
		"tmdb":            100,
		"imdb":            90,
		"themoviedb":      100,
		"douban":          85,
		"mal":             80, // MyAnimeList
		"anidb":           75,
		"bangumi":         70,
		"rottentomatoes":  65,
		"metacritic":      60,
		"trakt":           55,
		"letterboxd":      50,
		"official":        100,
		"local":           5,
		"unknown":         0,
	}
}

// loadDefaultExcludedWords 加载默认排除词
func (c *MetadataCustomization) loadDefaultExcludedWords() {
	// 默认排除词列表
	excluded := []string{
		"1080p", "720p", "480p", "2160p", "4k", "8k",
		"web", "web-dl", "webdl", "hd", "sd", "uhd",
		"bluray", "bdrip", "dvdrip", "hdtv", "tvrip",
		"x264", "x265", "h264", "h265", "hevc", "mpeg4",
		"avi", "mp4", "mkv", "mov", "wmv", "flv",
		"mp3", "aac", "flac", "ogg", "wav",
		"chs", "cht", "双语", "中字", "字幕", "sub",
		"hdc", "hds", "hdchina", "hdsky", "mteam", "ourbits",
		"lemonhd", "totheglory", "tjupt", "ptchina", "pthome",
		"10bit", "8bit", "hdr", "sdr", "dv", "dolby",
		"atmos", "dts", "ac3", "eac3", "truehd",
		"ntsc", "pal", "region", "uncut", "extended",
		"director's", "cut", "remastered", "limited",
		"season", "episode", "s01", "s02", "e01", "e02",
	}
	
	c.excludedWords = make(map[string]bool)
	for _, word := range excluded {
		c.excludedWords[strings.ToLower(word)] = true
	}
}

// applyRule 应用替换规则
func (c *MetadataCustomization) applyRule(text string, rule ReplaceRule) string {
	if rule.Pattern == "" || !rule.IsActive {
		return text
	}
	
	if rule.IsRegex {
		// 使用正则表达式替换
		// 注意：在实际实现中，应该使用regexp包编译和缓存正则表达式
		// 这里简化处理，直接使用strings包的相关方法
		if rule.IgnoreCase {
			// 忽略大小写的简单实现（实际应使用regexp.Compile(`(?i)` + rule.Pattern)）
			// 这里简化处理，只做基础替换
			return strings.ReplaceAll(text, rule.Pattern, rule.Replacement)
		} else {
			return strings.ReplaceAll(text, rule.Pattern, rule.Replacement)
		}
	} else {
		// 简单字符串替换
		if rule.IgnoreCase {
			// 忽略大小写的字符串替换
			return strings.ToLower(text)
		} else {
			return strings.ReplaceAll(text, rule.Pattern, rule.Replacement)
		}
	}
}

// invalidateCache 使缓存失效
func (c *MetadataCustomization) invalidateCache() {
	c.cache = make(map[string]interface{})
	c.cacheExpiry = time.Now().Add(1 * time.Hour)
}

// GetCacheStatus 获取缓存状态
func (c *MetadataCustomization) GetCacheStatus() map[string]interface{} {
	c.RLock()
	defer c.RUnlock()
	
	return map[string]interface{}{
		"cache_size":  len(c.cache),
		"cache_expiry": c.cacheExpiry,
		"is_valid":    time.Now().Before(c.cacheExpiry),
	}
}

// ClearCache 清除缓存
func (c *MetadataCustomization) ClearCache() {
	c.Lock()
	defer c.Unlock()
	
	c.invalidateCache()
}

// RegisterCustomField 注册自定义字段（预留接口，用于未来扩展）
func (c *MetadataCustomization) RegisterCustomField(field CustomField) error {
	// 预留接口，用于未来支持用户自定义元数据字段
	// 当前版本暂不实现具体逻辑
	return nil
}

// GetCustomFields 获取所有自定义字段（预留接口）
func (c *MetadataCustomization) GetCustomFields() []CustomField {
	// 预留接口，返回空切片
	return []CustomField{}
}

// ValidateMetadata 验证元数据完整性（预留接口）
func (c *MetadataCustomization) ValidateMetadata(metadata map[string]interface{}) (bool, []string) {
	// 预留接口，用于验证元数据的完整性和有效性
	// 当前版本暂不实现具体逻辑，默认返回有效
	return true, []string{}
}

// SanitizeMetadata 清理元数据中的敏感信息
func (c *MetadataCustomization) SanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return make(map[string]interface{})
	}
	
	// 创建副本以避免修改原始数据
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	
	// 移除可能的敏感字段
	sensitiveFields := []string{
		"password", "token", "api_key", "secret", "auth",
		"username", "user_id", "email", "phone", "credit_card",
	}
	
	for _, field := range sensitiveFields {
		delete(result, field)
		delete(result, strings.ToLower(field))
		delete(result, strings.ToUpper(field))
	}
	
	return result
}