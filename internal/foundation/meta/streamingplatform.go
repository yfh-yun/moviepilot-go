package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// StreamingPlatform 流媒体平台信息
type StreamingPlatform struct {
	ID          string            `json:"id"`          // 平台唯一标识
	Name        string            `json:"name"`        // 平台名称
	DisplayName string            `json:"display_name"` // 显示名称
	Icon        string            `json:"icon"`        // 平台图标URL
	Domain      string            `json:"domain"`      // 主域名
	Regions     []string          `json:"regions"`     // 支持的区域
	URLPatterns []string          `json:"url_patterns"` // URL匹配模式
	Priority    int               `json:"priority"`    // 平台优先级
	IsActive    bool              `json:"is_active"`   // 是否激活
	Features    map[string]bool   `json:"features"`    // 支持的功能
	Properties  map[string]string `json:"properties"`  // 其他属性
	LastUpdated time.Time         `json:"last_updated"` // 最后更新时间
}

// StreamingLink 流媒体链接信息
type StreamingLink struct {
	PlatformID  string                 `json:"platform_id"`  // 平台ID
	PlatformName string                `json:"platform_name"` // 平台名称
	URL         string                 `json:"url"`          // 流媒体URL
	Title       string                 `json:"title"`        // 标题
	Quality     string                 `json:"quality"`      // 视频质量
	Language    string                 `json:"language"`     // 语言
	Subtitle    string                 `json:"subtitle"`     // 字幕信息
	Region      string                 `json:"region"`       // 区域限制
	Metadata    map[string]interface{} `json:"metadata"`     // 额外元数据
	ExpiresAt   time.Time              `json:"expires_at"`   // 过期时间
	CreatedAt   time.Time              `json:"created_at"`   // 创建时间
	IsDirectPlay bool                  `json:"is_direct_play"` // 是否支持直接播放
}

// StreamingPlatformManager 流媒体平台管理器
type StreamingPlatformManager struct {
	sync.RWMutex
	platforms       map[string]*StreamingPlatform // 平台ID映射
	nameToID        map[string]string             // 平台名称映射
	domainToID      map[string]string             // 域名映射
	regexCache      map[string]*regexp.Regexp     // 正则表达式缓存
	defaultPriority int                           // 默认优先级
	initialized     bool                          // 是否已初始化
	lastUpdated     time.Time                     // 最后更新时间
}

// NewStreamingPlatformManager 创建流媒体平台管理器实例
func NewStreamingPlatformManager() *StreamingPlatformManager {
	manager := &StreamingPlatformManager{
		platforms:       make(map[string]*StreamingPlatform),
		nameToID:        make(map[string]string),
		domainToID:      make(map[string]string),
		regexCache:      make(map[string]*regexp.Regexp),
		defaultPriority: 50,
		initialized:     false,
		lastUpdated:     time.Now(),
	}
	return manager
}

// Initialize 初始化流媒体平台管理器
func (m *StreamingPlatformManager) Initialize(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()
	
	if m.initialized {
		return nil
	}
	
	// 加载默认流媒体平台
	m.loadDefaultPlatforms()
	m.initializeRegexCache()
	
	m.initialized = true
	m.lastUpdated = time.Now()
	
	return nil
}

// GetPlatform 获取指定ID的流媒体平台信息
func (m *StreamingPlatformManager) GetPlatform(ctx context.Context, platformID string) (*StreamingPlatform, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	platform, exists := m.platforms[strings.ToLower(platformID)]
	if !exists {
		return nil, fmt.Errorf("平台ID %s 不存在", platformID)
	}
	
	return platform, nil
}

// GetPlatformByName 根据名称获取流媒体平台信息
func (m *StreamingPlatformManager) GetPlatformByName(ctx context.Context, name string) (*StreamingPlatform, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	platformID, exists := m.nameToID[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("平台名称 %s 不存在", name)
	}
	
	platform, exists := m.platforms[platformID]
	if !exists {
		return nil, fmt.Errorf("平台ID %s 不存在", platformID)
	}
	
	return platform, nil
}

// GetAllPlatforms 获取所有流媒体平台列表
func (m *StreamingPlatformManager) GetAllPlatforms(ctx context.Context) ([]*StreamingPlatform, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	platforms := make([]*StreamingPlatform, 0, len(m.platforms))
	for _, platform := range m.platforms {
		if platform.IsActive {
			platforms = append(platforms, platform)
		}
	}
	
	// 按优先级排序
	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].Priority > platforms[j].Priority
	})
	
	return platforms, nil
}

// GetActivePlatforms 获取所有激活的流媒体平台
func (m *StreamingPlatformManager) GetActivePlatforms(ctx context.Context) ([]*StreamingPlatform, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	platforms := make([]*StreamingPlatform, 0, len(m.platforms))
	for _, platform := range m.platforms {
		if platform.IsActive {
			platforms = append(platforms, platform)
		}
	}
	
	return platforms, nil
}

// AddPlatform 添加新的流媒体平台
func (m *StreamingPlatformManager) AddPlatform(ctx context.Context, platform *StreamingPlatform) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if platform == nil {
		return fmt.Errorf("平台信息不能为空")
	}
	
	if platform.ID == "" {
		return fmt.Errorf("平台ID不能为空")
	}
	
	if platform.Name == "" {
		return fmt.Errorf("平台名称不能为空")
	}
	
	m.Lock()
	defer m.Unlock()
	
	// 标准化ID和名称
	platformID := strings.ToLower(platform.ID)
	platformName := strings.ToLower(platform.Name)
	
	// 如果没有设置显示名称，使用名称作为显示名称
	if platform.DisplayName == "" {
		platform.DisplayName = platform.Name
	}
	
	// 如果没有设置优先级，使用默认优先级
	if platform.Priority == 0 {
		platform.Priority = m.defaultPriority
	}
	
	// 初始化功能映射
	if platform.Features == nil {
		platform.Features = make(map[string]bool)
	}
	
	// 初始化属性映射
	if platform.Properties == nil {
		platform.Properties = make(map[string]string)
	}
	
	// 更新时间戳
	platform.LastUpdated = time.Now()
	
	// 保存平台信息
	m.platforms[platformID] = platform
	m.nameToID[platformName] = platformID
	
	// 更新域名映射
	if platform.Domain != "" {
		m.domainToID[strings.ToLower(platform.Domain)] = platformID
	}
	
	// 更新正则表达式缓存
	for _, pattern := range platform.URLPatterns {
		if pattern != "" {
			if _, err := m.compileRegex(pattern); err != nil {
				// 记录错误但不中断流程
				fmt.Printf("编译URL模式失败: %s, 错误: %v\n", pattern, err)
			}
		}
	}
	
	m.lastUpdated = time.Now()
	
	return nil
}

// UpdatePlatform 更新流媒体平台信息
func (m *StreamingPlatformManager) UpdatePlatform(ctx context.Context, platform *StreamingPlatform) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if platform == nil {
		return fmt.Errorf("平台信息不能为空")
	}
	
	if platform.ID == "" {
		return fmt.Errorf("平台ID不能为空")
	}
	
	m.Lock()
	defer m.Unlock()
	
	platformID := strings.ToLower(platform.ID)
	oldPlatform, exists := m.platforms[platformID]
	
	if !exists {
		return fmt.Errorf("平台ID %s 不存在", platformID)
	}
	
	// 记录旧的名称和域名，用于更新映射
	oldName := strings.ToLower(oldPlatform.Name)
	oldDomain := strings.ToLower(oldPlatform.Domain)
	
	// 更新平台信息
	platform.LastUpdated = time.Now()
	m.platforms[platformID] = platform
	
	// 更新名称映射
	if oldName != "" {
		delete(m.nameToID, oldName)
	}
	m.nameToID[strings.ToLower(platform.Name)] = platformID
	
	// 更新域名映射
	if oldDomain != "" {
		delete(m.domainToID, oldDomain)
	}
	if platform.Domain != "" {
		m.domainToID[strings.ToLower(platform.Domain)] = platformID
	}
	
	// 更新正则表达式缓存
	// 先清理旧的URL模式对应的缓存
	for _, pattern := range oldPlatform.URLPatterns {
		delete(m.regexCache, pattern)
	}
	// 添加新的URL模式对应的缓存
	for _, pattern := range platform.URLPatterns {
		if pattern != "" {
			if _, err := m.compileRegex(pattern); err != nil {
				fmt.Printf("编译URL模式失败: %s, 错误: %v\n", pattern, err)
			}
		}
	}
	
	m.lastUpdated = time.Now()
	
	return nil
}

// RemovePlatform 移除流媒体平台
func (m *StreamingPlatformManager) RemovePlatform(ctx context.Context, platformID string) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if platformID == "" {
		return fmt.Errorf("平台ID不能为空")
	}
	
	m.Lock()
	defer m.Unlock()
	
	platformID = strings.ToLower(platformID)
	platform, exists := m.platforms[platformID]
	
	if !exists {
		return fmt.Errorf("平台ID %s 不存在", platformID)
	}
	
	// 移除映射
	delete(m.nameToID, strings.ToLower(platform.Name))
	if platform.Domain != "" {
		delete(m.domainToID, strings.ToLower(platform.Domain))
	}
	
	// 清理正则表达式缓存
	for _, pattern := range platform.URLPatterns {
		delete(m.regexCache, pattern)
	}
	
	// 移除平台
	delete(m.platforms, platformID)
	
	m.lastUpdated = time.Now()
	
	return nil
}

// TogglePlatformStatus 切换平台激活状态
func (m *StreamingPlatformManager) TogglePlatformStatus(ctx context.Context, platformID string) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if platformID == "" {
		return fmt.Errorf("平台ID不能为空")
	}
	
	m.Lock()
	defer m.Unlock()
	
	platformID = strings.ToLower(platformID)
	platform, exists := m.platforms[platformID]
	
	if !exists {
		return fmt.Errorf("平台ID %s 不存在", platformID)
	}
	
	// 切换状态
	platform.IsActive = !platform.IsActive
	platform.LastUpdated = time.Now()
	
	m.lastUpdated = time.Now()
	
	return nil
}

// DetectPlatformFromURL 从URL检测流媒体平台
func (m *StreamingPlatformManager) DetectPlatformFromURL(ctx context.Context, streamingURL string) (*StreamingPlatform, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if streamingURL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	
	m.RLock()
	defer m.RUnlock()
	
	// 尝试解析URL获取域名
	parsedURL, err := url.Parse(streamingURL)
	var domain string
	if err == nil && parsedURL.Host != "" {
		domain = strings.ToLower(parsedURL.Host)
		// 移除端口号
		if idx := strings.LastIndex(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
		
		// 尝试直接通过域名查找
		if platformID, exists := m.domainToID[domain]; exists {
			if platform, ok := m.platforms[platformID]; ok && platform.IsActive {
				return platform, nil
			}
		}
		
		// 尝试通过子域名查找
		parts := strings.Split(domain, ".")
		for i := 0; i < len(parts)-1; i++ {
			subDomain := strings.Join(parts[i:], ".")
			if platformID, exists := m.domainToID[subDomain]; exists {
				if platform, ok := m.platforms[platformID]; ok && platform.IsActive {
					return platform, nil
				}
			}
		}
	}
	
	// 使用URL模式匹配
	for _, platform := range m.platforms {
		if !platform.IsActive {
			continue
		}
		
		for _, pattern := range platform.URLPatterns {
			if pattern == "" {
				continue
			}
			
			re, exists := m.regexCache[pattern]
			if !exists {
				// 如果缓存中没有，尝试编译
				var compileErr error
				re, compileErr = m.compileRegex(pattern)
				if compileErr != nil {
					continue
				}
			}
			
			if re.MatchString(streamingURL) {
				return platform, nil
			}
		}
	}
	
	return nil, fmt.Errorf("无法识别流媒体平台")
}

// ParseStreamingLink 解析流媒体链接信息
func (m *StreamingPlatformManager) ParseStreamingLink(ctx context.Context, streamingURL string) (*StreamingLink, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	if streamingURL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	
	// 检测平台
	platform, err := m.DetectPlatformFromURL(ctx, streamingURL)
	if err != nil {
		// 如果无法检测平台，创建一个未知平台的链接
		return &StreamingLink{
			PlatformID:   "unknown",
			PlatformName: "未知平台",
			URL:          streamingURL,
			CreatedAt:    time.Now(),
		}, nil
	}
	
	// 创建链接信息
	link := &StreamingLink{
		PlatformID:   platform.ID,
		PlatformName: platform.DisplayName,
		URL:          streamingURL,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	// 尝试从URL中提取其他信息
	m.extractInfoFromURL(streamingURL, link)
	
	// 设置平台特性
	link.IsDirectPlay = platform.Features["direct_play"]
	
	return link, nil
}

// GetStreamingLinksForMedia 获取媒体的所有流媒体链接
func (m *StreamingPlatformManager) GetStreamingLinksForMedia(ctx context.Context, mediaID string, mediaType string) ([]*StreamingLink, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	// 预留接口，用于未来从数据库或外部服务获取流媒体链接
	// 当前版本返回空列表
	return []*StreamingLink{}, nil
}

// ExportPlatforms 导出所有平台配置
func (m *StreamingPlatformManager) ExportPlatforms() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	
	platforms := make([]*StreamingPlatform, 0, len(m.platforms))
	for _, platform := range m.platforms {
		platforms = append(platforms, platform)
	}
	
	data, err := json.MarshalIndent(platforms, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化平台配置失败: %w", err)
	}
	
	return data, nil
}

// ImportPlatforms 导入平台配置
func (m *StreamingPlatformManager) ImportPlatforms(ctx context.Context, data []byte) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	var platforms []*StreamingPlatform
	if err := json.Unmarshal(data, &platforms); err != nil {
		return fmt.Errorf("反序列化平台配置失败: %w", err)
	}
	
	m.Lock()
	defer m.Unlock()
	
	// 清空现有数据
	m.platforms = make(map[string]*StreamingPlatform)
	m.nameToID = make(map[string]string)
	m.domainToID = make(map[string]string)
	m.regexCache = make(map[string]*regexp.Regexp)
	
	// 导入新平台
	for _, platform := range platforms {
		if platform.ID == "" || platform.Name == "" {
			continue
		}
		
		platformID := strings.ToLower(platform.ID)
		platformName := strings.ToLower(platform.Name)
		
		if platform.DisplayName == "" {
			platform.DisplayName = platform.Name
		}
		
		if platform.Priority == 0 {
			platform.Priority = m.defaultPriority
		}
		
		if platform.Features == nil {
			platform.Features = make(map[string]bool)
		}
		
		if platform.Properties == nil {
			platform.Properties = make(map[string]string)
		}
		
		m.platforms[platformID] = platform
		m.nameToID[platformName] = platformID
		
		if platform.Domain != "" {
			m.domainToID[strings.ToLower(platform.Domain)] = platformID
		}
		
		// 编译URL模式
		for _, pattern := range platform.URLPatterns {
			if pattern != "" {
				if _, err := m.compileRegex(pattern); err != nil {
					fmt.Printf("编译URL模式失败: %s, 错误: %v\n", pattern, err)
				}
			}
		}
	}
	
	m.lastUpdated = time.Now()
	
	return nil
}

// ResetToDefaults 重置为默认配置
func (m *StreamingPlatformManager) ResetToDefaults() error {
	m.Lock()
	defer m.Unlock()
	
	m.platforms = make(map[string]*StreamingPlatform)
	m.nameToID = make(map[string]string)
	m.domainToID = make(map[string]string)
	m.regexCache = make(map[string]*regexp.Regexp)
	
	m.loadDefaultPlatforms()
	m.initializeRegexCache()
	
	m.lastUpdated = time.Now()
	
	return nil
}

// GetStatistics 获取统计信息
func (m *StreamingPlatformManager) GetStatistics() map[string]interface{} {
	m.RLock()
	defer m.RUnlock()
	
	totalPlatforms := len(m.platforms)
	activePlatforms := 0
	
	for _, platform := range m.platforms {
		if platform.IsActive {
			activePlatforms++
		}
	}
	
	return map[string]interface{}{
		"total_platforms":  totalPlatforms,
		"active_platforms": activePlatforms,
		"inactive_platforms": totalPlatforms - activePlatforms,
		"regex_cache_size":  len(m.regexCache),
		"last_updated":      m.lastUpdated,
		"initialized":       m.initialized,
	}
}

// 私有方法

// loadDefaultPlatforms 加载默认流媒体平台
func (m *StreamingPlatformManager) loadDefaultPlatforms() {
	// 定义默认流媒体平台
	defaultPlatforms := []*StreamingPlatform{
		{
			ID:          "bilibili",
			Name:        "bilibili",
			DisplayName: "哔哩哔哩",
			Icon:        "https://i0.hdslb.com/bfs/archive/a4a0972d242813e43a2c4c16a1f856184a2e42cf.png",
			Domain:      "bilibili.com",
			Regions:     []string{"cn"},
			URLPatterns: []string{
				`https?://www\.bilibili\.com/video/av\d+`,
				`https?://www\.bilibili\.com/video/BV[0-9A-Za-z]+`,
				`https?://b23\.tv/[0-9A-Za-z]+`,
			},
			Priority: 90,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": true,
				"hd":          true,
				"subtitle":    true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "iqiyi",
			Name:        "iqiyi",
			DisplayName: "爱奇艺",
			Icon:        "https://www.iqiyi.com/favicon.ico",
			Domain:      "iqiyi.com",
			Regions:     []string{"cn"},
			URLPatterns: []string{
				`https?://www\.iqiyi\.com/v_\w+\.html`,
				`https?://www\.iqiyi\.com/a_\w+\.html`,
			},
			Priority: 85,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"vip":         true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "youku",
			Name:        "youku",
			DisplayName: "优酷",
			Icon:        "https://static.youku.com/v1.0.0723/img/favicon.ico",
			Domain:      "youku.com",
			Regions:     []string{"cn"},
			URLPatterns: []string{
				`https?://v\.youku\.com/v_show/id_\w+\.html`,
				`https?://www\.youku\.com/show_page/id_\w+\.html`,
			},
			Priority: 80,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"vip":         true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "tencent",
			Name:        "tencent",
			DisplayName: "腾讯视频",
			Icon:        "https://v.qq.com/favicon.ico",
			Domain:      "qq.com",
			Regions:     []string{"cn"},
			URLPatterns: []string{
				`https?://v\.qq\.com/x/cover/\w+\.html`,
				`https?://v\.qq\.com/x/page/\w+\.html`,
				`https?://www\.v.qq\.com/x/cover/\w+\.html`,
			},
			Priority: 85,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"vip":         true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "netflix",
			Name:        "netflix",
			DisplayName: "Netflix",
			Icon:        "https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2023.ico",
			Domain:      "netflix.com",
			Regions:     []string{"us", "uk", "ca", "au", "jp", "kr", "de", "fr", "es", "it"},
			URLPatterns: []string{
				`https?://www\.netflix\.com/watch/\d+`,
				`https?://www\.netflix\.com/title/\d+`,
			},
			Priority: 95,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"4k":          true,
				"hdr":         true,
				"dolby_vision": true,
				"dolby_atmos":  true,
				"subtitle":     true,
				"vip":          true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "youtube",
			Name:        "youtube",
			DisplayName: "YouTube",
			Icon:        "https://www.youtube.com/s/desktop/05a91f42/img/favicon_32x32.png",
			Domain:      "youtube.com",
			Regions:     []string{"global"},
			URLPatterns: []string{
				`https?://www\.youtube\.com/watch\?v=([^&]+)`,
				`https?://youtu\.be/([^?]+)`,
				`https?://www\.youtube\.com/embed/([^?]+)`,
			},
			Priority: 100,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": true,
				"hd":          true,
				"4k":          true,
				"8k":          true,
				"hdr":         true,
				"subtitle":    true,
				"live":        true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "disney",
			Name:        "disney",
			DisplayName: "Disney+",
			Icon:        "https://disneyplus.com/favicon.ico",
			Domain:      "disneyplus.com",
			Regions:     []string{"us", "uk", "ca", "au", "jp", "kr", "de", "fr", "es", "it"},
			URLPatterns: []string{
				`https?://www\.disneyplus\.com/title/\w+`,
				`https?://www\.disneyplus\.com/series/\w+`,
			},
			Priority: 90,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"4k":          true,
				"hdr":         true,
				"dolby_vision": true,
				"dolby_atmos":  true,
				"subtitle":     true,
				"vip":          true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:          "amazon",
			Name:        "amazon",
			DisplayName: "Amazon Prime Video",
			Icon:        "https://www.primevideo.com/favicon.ico",
			Domain:      "primevideo.com",
			Regions:     []string{"us", "uk", "ca", "au", "jp", "kr", "de", "fr", "es", "it"},
			URLPatterns: []string{
				`https?://www\.primevideo\.com/detail/\w+`,
				`https?://www\.amazon\.com/gp/video/detail/\w+`,
			},
			Priority: 85,
			IsActive: true,
			Features: map[string]bool{
				"direct_play": false,
				"hd":          true,
				"4k":          true,
				"hdr":         true,
				"dolby_vision": true,
				"dolby_atmos":  true,
				"subtitle":     true,
				"vip":          true,
			},
			Properties: map[string]string{
				"type": "video",
			},
			LastUpdated: time.Now(),
		},
	}
	
	// 添加到管理器
	for _, platform := range defaultPlatforms {
		platformID := strings.ToLower(platform.ID)
		platformName := strings.ToLower(platform.Name)
		
		m.platforms[platformID] = platform
		m.nameToID[platformName] = platformID
		
		if platform.Domain != "" {
			m.domainToID[strings.ToLower(platform.Domain)] = platformID
		}
	}
}

// initializeRegexCache 初始化正则表达式缓存
func (m *StreamingPlatformManager) initializeRegexCache() {
	for _, platform := range m.platforms {
		for _, pattern := range platform.URLPatterns {
			if pattern != "" {
				if _, err := m.compileRegex(pattern); err != nil {
					fmt.Printf("编译URL模式失败: %s, 错误: %v\n", pattern, err)
				}
			}
		}
	}
}

// compileRegex 编译正则表达式并缓存
func (m *StreamingPlatformManager) compileRegex(pattern string) (*regexp.Regexp, error) {
	// 检查缓存
	re, exists := m.regexCache[pattern]
	if exists {
		return re, nil
	}
	
	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	
	// 缓存编译后的正则表达式
	m.regexCache[pattern] = re
	
	return re, nil
}

// extractInfoFromURL 从URL中提取信息
func (m *StreamingPlatformManager) extractInfoFromURL(streamingURL string, link *StreamingLink) {
	// 尝试从URL查询参数中提取信息
	parsedURL, err := url.Parse(streamingURL)
	if err != nil {
		return
	}
	
	// 提取查询参数
	query := parsedURL.Query()
	
	// 提取质量信息
	if quality := query.Get("quality"); quality != "" {
		link.Quality = quality
	}
	
	// 提取语言信息
	if language := query.Get("language"); language != "" {
		link.Language = language
	}
	
	// 提取字幕信息
	if subtitle := query.Get("subtitle"); subtitle != "" {
		link.Subtitle = subtitle
	}
	
	// 提取区域信息
	if region := query.Get("region"); region != "" {
		link.Region = region
	}
	
	// 尝试从URL路径中提取视频质量
	path := parsedURL.Path
	if strings.Contains(path, "4k") || strings.Contains(path, "2160p") {
		link.Quality = "4K"
	} else if strings.Contains(path, "1080p") {
		link.Quality = "1080p"
	} else if strings.Contains(path, "720p") {
		link.Quality = "720p"
	} else if strings.Contains(path, "480p") {
		link.Quality = "480p"
	}
	
	// 尝试从URL路径中提取语言信息
	if strings.Contains(strings.ToLower(path), "zh") || strings.Contains(strings.ToLower(path), "chinese") {
		if strings.Contains(strings.ToLower(path), "cn") || strings.Contains(strings.ToLower(path), "simplified") {
			link.Language = "简体中文"
		} else if strings.Contains(strings.ToLower(path), "hk") || strings.Contains(strings.ToLower(path), "tw") || strings.Contains(strings.ToLower(path), "traditional") {
			link.Language = "繁体中文"
		} else {
			link.Language = "中文"
		}
	} else if strings.Contains(strings.ToLower(path), "en") || strings.Contains(strings.ToLower(path), "english") {
		link.Language = "英语"
	}
	
	// 检查是否是直接播放链接（基于文件扩展名）
	if strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".mkv") || strings.HasSuffix(path, ".avi") || 
	   strings.HasSuffix(path, ".mov") || strings.HasSuffix(path, ".wmv") || strings.HasSuffix(path, ".flv") ||
	   strings.HasSuffix(path, ".webm") || strings.HasSuffix(path, ".m3u8") {
		link.IsDirectPlay = true
	}
}

// GetPlatformFeatures 获取平台支持的功能列表
func (m *StreamingPlatformManager) GetPlatformFeatures(ctx context.Context, platformID string) ([]string, error) {
	platform, err := m.GetPlatform(ctx, platformID)
	if err != nil {
		return nil, err
	}
	
	features := make([]string, 0)
	for feature, supported := range platform.Features {
		if supported {
			features = append(features, feature)
		}
	}
	
	return features, nil
}

// SetDefaultPriority 设置默认优先级
func (m *StreamingPlatformManager) SetDefaultPriority(ctx context.Context, priority int) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.Lock()
	defer m.Unlock()
	
	m.defaultPriority = priority
	m.lastUpdated = time.Now()
	
	return nil
}

// GetDefaultPriority 获取默认优先级
func (m *StreamingPlatformManager) GetDefaultPriority(ctx context.Context) (int, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return 0, fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	return m.defaultPriority, nil
}

// ClearCache 清除缓存
func (m *StreamingPlatformManager) ClearCache() {
	m.Lock()
	defer m.Unlock()
	
	m.regexCache = make(map[string]*regexp.Regexp)
	m.initializeRegexCache()
}