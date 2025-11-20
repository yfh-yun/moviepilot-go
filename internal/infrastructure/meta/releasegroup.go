package meta

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ReleaseGroupsMatcher 发布组匹配器
type ReleaseGroupsMatcher struct {
	sync.RWMutex
	releaseGroups []ReleaseGroup
	patterns      []*regexp.Regexp
	initialized   bool
	lastUpdated   time.Time
}

// ReleaseGroup 发布组信息
type ReleaseGroup struct {
	Name      string   `json:"name"`      // 发布组名称
	Aliases   []string `json:"aliases"`   // 别名列表
	Tags      []string `json:"tags"`      // 标签
	Priority  int      `json:"priority"`  // 优先级，数字越大优先级越高
	IsActive  bool     `json:"is_active"` // 是否激活
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// NewReleaseGroupsMatcher 创建发布组匹配器实例
func NewReleaseGroupsMatcher() *ReleaseGroupsMatcher {
	matcher := &ReleaseGroupsMatcher{
		releaseGroups: make([]ReleaseGroup, 0),
		patterns:      make([]*regexp.Regexp, 0),
		initialized:   false,
		lastUpdated:   time.Now(),
	}
	return matcher
}

// Initialize 初始化发布组匹配器
func (m *ReleaseGroupsMatcher) Initialize(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()
	
	if m.initialized {
		return nil
	}
	
	// 加载默认发布组列表
	m.loadDefaultReleaseGroups()
	
	// 编译正则表达式模式
	m.compilePatterns()
	
	m.initialized = true
	m.lastUpdated = time.Now()
	
	return nil
}

// Match 匹配发布组
func (m *ReleaseGroupsMatcher) Match(ctx context.Context, text string) string {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return ""
		}
	}
	
	// 检查是否需要更新（如果上次更新超过24小时）
	if time.Since(m.lastUpdated) > 24*time.Hour {
		go m.updateReleaseGroups(ctx)
	}
	
	m.RLock()
	defer m.RUnlock()
	
	// 空文本检查
	if text == "" {
		return ""
	}
	
	// 转换为小写以进行大小写不敏感匹配
	textLower := strings.ToLower(text)
	
	// 按优先级排序的发布组列表
	sortedGroups := m.getSortedReleaseGroups()
	
	// 尝试精确匹配
	for _, group := range sortedGroups {
		if group.IsActive {
			// 尝试匹配名称
			if strings.Contains(textLower, strings.ToLower(group.Name)) {
				return group.Name
			}
			
			// 尝试匹配别名
			for _, alias := range group.Aliases {
				if strings.Contains(textLower, strings.ToLower(alias)) {
					return group.Name
				}
			}
		}
	}
	
	// 使用正则表达式进行匹配
	for i, pattern := range m.patterns {
		if i < len(sortedGroups) && sortedGroups[i].IsActive {
			matches := pattern.FindStringSubmatch(text)
			if len(matches) > 0 {
				return sortedGroups[i].Name
			}
		}
	}
	
	// 尝试匹配方括号内的内容作为发布组
	simpleGroup := m.extractSimpleGroup(text)
	if simpleGroup != "" {
		return simpleGroup
	}
	
	return ""
}

// MatchAll 匹配所有可能的发布组
func (m *ReleaseGroupsMatcher) MatchAll(ctx context.Context, text string) []string {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return []string{}
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	// 空文本检查
	if text == "" {
		return []string{}
	}
	
	// 转换为小写以进行大小写不敏感匹配
	textLower := strings.ToLower(text)
	
	// 存储匹配结果
	matchedGroups := make(map[string]bool)
	
	// 按优先级排序的发布组列表
	sortedGroups := m.getSortedReleaseGroups()
	
	// 尝试匹配每个发布组
	for _, group := range sortedGroups {
		if group.IsActive {
			matched := false
			
			// 尝试匹配名称
			if strings.Contains(textLower, strings.ToLower(group.Name)) {
				matchedGroups[group.Name] = true
				matched = true
			}
			
			// 尝试匹配别名
			for _, alias := range group.Aliases {
				if strings.Contains(textLower, strings.ToLower(alias)) {
					matchedGroups[group.Name] = true
					matched = true
					break
				}
			}
			
			// 如果已经匹配到，跳过正则匹配以提高性能
			if matched {
				continue
			}
		}
	}
	
	// 使用正则表达式进行匹配
	for i, pattern := range m.patterns {
		if i < len(sortedGroups) && sortedGroups[i].IsActive {
			matches := pattern.FindStringSubmatch(text)
			if len(matches) > 0 {
				matchedGroups[sortedGroups[i].Name] = true
			}
		}
	}
	
	// 转换为切片
	result := make([]string, 0, len(matchedGroups))
	for group := range matchedGroups {
		result = append(result, group)
	}
	
	// 按优先级排序结果
	sort.Slice(result, func(i, j int) bool {
		return m.getGroupPriority(result[i]) > m.getGroupPriority(result[j])
	})
	
	return result
}

// AddReleaseGroup 添加发布组
func (m *ReleaseGroupsMatcher) AddReleaseGroup(group ReleaseGroup) error {
	m.Lock()
	defer m.Unlock()
	
	// 检查是否已存在
	for i, g := range m.releaseGroups {
		if strings.EqualFold(g.Name, group.Name) {
			// 更新现有发布组
			m.releaseGroups[i] = group
			m.compilePatterns()
			m.lastUpdated = time.Now()
			return nil
		}
	}
	
	// 添加新发布组
	m.releaseGroups = append(m.releaseGroups, group)
	m.compilePatterns()
	m.lastUpdated = time.Now()
	
	return nil
}

// RemoveReleaseGroup 移除发布组
func (m *ReleaseGroupsMatcher) RemoveReleaseGroup(name string) error {
	m.Lock()
	defer m.Unlock()
	
	// 查找并移除发布组
	for i, group := range m.releaseGroups {
		if strings.EqualFold(group.Name, name) {
			m.releaseGroups = append(m.releaseGroups[:i], m.releaseGroups[i+1:]...)
			m.compilePatterns()
			m.lastUpdated = time.Now()
			return nil
		}
	}
	
	return nil
}

// GetReleaseGroups 获取所有发布组
func (m *ReleaseGroupsMatcher) GetReleaseGroups(ctx context.Context) []ReleaseGroup {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return []ReleaseGroup{}
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	// 返回副本以避免并发修改
	copyGroups := make([]ReleaseGroup, len(m.releaseGroups))
	copy(copyGroups, m.releaseGroups)
	return copyGroups
}

// EnableReleaseGroup 启用发布组
func (m *ReleaseGroupsMatcher) EnableReleaseGroup(name string) error {
	m.Lock()
	defer m.Unlock()
	
	return m.setReleaseGroupActive(name, true)
}

// DisableReleaseGroup 禁用发布组
func (m *ReleaseGroupsMatcher) DisableReleaseGroup(name string) error {
	m.Lock()
	defer m.Unlock()
	
	return m.setReleaseGroupActive(name, false)
}

// UpdateReleaseGroupPriority 更新发布组优先级
func (m *ReleaseGroupsMatcher) UpdateReleaseGroupPriority(name string, priority int) error {
	m.Lock()
	defer m.Unlock()
	
	for i, group := range m.releaseGroups {
		if strings.EqualFold(group.Name, name) {
			m.releaseGroups[i].Priority = priority
			m.lastUpdated = time.Now()
			return nil
		}
	}
	
	return nil
}

// ClearCache 清除缓存
func (m *ReleaseGroupsMatcher) ClearCache() {
	m.Lock()
	defer m.Unlock()
	
	m.initialized = false
	m.lastUpdated = time.Time{}
}

// loadDefaultReleaseGroups 加载默认发布组列表
func (m *ReleaseGroupsMatcher) loadDefaultReleaseGroups() {
	// 默认发布组列表
	defaultGroups := []ReleaseGroup{
		// 动漫字幕组
		{Name: "LoliHouse", Aliases: []string{"Loli", "LH"}, Priority: 10, IsActive: true, CreatedAt: time.Now()},
		{Name: "Erai-raws", Aliases: []string{"Erai", "ERAI"}, Priority: 9, IsActive: true, CreatedAt: time.Now()},
		{Name: "DMG", Aliases: []string{}, Priority: 8, IsActive: true, CreatedAt: time.Now()},
		{Name: "SUBPIG", Aliases: []string{"SUBPIG猪猪字幕组", "猪猪字幕组"}, Priority: 8, IsActive: true, CreatedAt: time.Now()},
		{Name: "蜜柑计划", Aliases: []string{"蜜柑", "Mikan"}, Priority: 7, IsActive: true, CreatedAt: time.Now()},
		{Name: "漫游字幕组", Aliases: []string{"漫游", "Roaming"}, Priority: 7, IsActive: true, CreatedAt: time.Now()},
		{Name: "千夏字幕组", Aliases: []string{"千夏", "Chinatsu"}, Priority: 6, IsActive: true, CreatedAt: time.Now()},
		{Name: "诸神字幕组", Aliases: []string{"诸神", "Kamigami"}, Priority: 6, IsActive: true, CreatedAt: time.Now()},
		{Name: "幻樱字幕组", Aliases: []string{"幻樱", "HY"}, Priority: 6, IsActive: true, CreatedAt: time.Now()},
		{Name: "动漫国字幕组", Aliases: []string{"动漫国", "DMG"}, Priority: 5, IsActive: true, CreatedAt: time.Now()},
		{Name: "豌豆字幕组", Aliases: []string{"豌豆"}, Priority: 5, IsActive: true, CreatedAt: time.Now()},
		{Name: "极影字幕组", Aliases: []string{"极影"}, Priority: 5, IsActive: true, CreatedAt: time.Now()},
		{Name: "澄空字幕组", Aliases: []string{"澄空", "CK"}, Priority: 5, IsActive: true, CreatedAt: time.Now()},
		{Name: "异域字幕组", Aliases: []string{"异域"}, Priority: 5, IsActive: true, CreatedAt: time.Now()},
		
		// 电影剧集发布组
		{Name: "CHS", Aliases: []string{"简中"}, Priority: 4, IsActive: true, CreatedAt: time.Now()},
		{Name: "CHT", Aliases: []string{"繁中"}, Priority: 4, IsActive: true, CreatedAt: time.Now()},
		{Name: "双语", Aliases: []string{"CHS&CHT", "简繁"}, Priority: 4, IsActive: true, CreatedAt: time.Now()},
		{Name: "WEBDL", Aliases: []string{}, Priority: 3, IsActive: true, CreatedAt: time.Now()},
		{Name: "HDRip", Aliases: []string{}, Priority: 3, IsActive: true, CreatedAt: time.Now()},
		{Name: "BDRip", Aliases: []string{}, Priority: 3, IsActive: true, CreatedAt: time.Now()},
		{Name: "BluRay", Aliases: []string{}, Priority: 3, IsActive: true, CreatedAt: time.Now()},
		{Name: "1080p", Aliases: []string{}, Priority: 2, IsActive: true, CreatedAt: time.Now()},
		{Name: "720p", Aliases: []string{}, Priority: 2, IsActive: true, CreatedAt: time.Now()},
		{Name: "2160p", Aliases: []string{}, Priority: 2, IsActive: true, CreatedAt: time.Now()},
		{Name: "4K", Aliases: []string{}, Priority: 2, IsActive: true, CreatedAt: time.Now()},
		{Name: "H264", Aliases: []string{}, Priority: 1, IsActive: true, CreatedAt: time.Now()},
		{Name: "H265", Aliases: []string{"HEVC"}, Priority: 1, IsActive: true, CreatedAt: time.Now()},
		{Name: "x264", Aliases: []string{}, Priority: 1, IsActive: true, CreatedAt: time.Now()},
		{Name: "x265", Aliases: []string{}, Priority: 1, IsActive: true, CreatedAt: time.Now()},
	}
	
	m.releaseGroups = defaultGroups
}

// compilePatterns 编译正则表达式模式
func (m *ReleaseGroupsMatcher) compilePatterns() {
	m.patterns = make([]*regexp.Regexp, 0, len(m.releaseGroups))
	
	for _, group := range m.releaseGroups {
		// 为每个发布组名称和别名创建正则表达式
		patterns := []string{}
		
		// 添加名称模式
		patterns = append(patterns, regexp.QuoteMeta(group.Name))
		
		// 添加别名模式
		for _, alias := range group.Aliases {
			patterns = append(patterns, regexp.QuoteMeta(alias))
		}
		
		// 创建组合正则表达式
		if len(patterns) > 0 {
			// 尝试匹配方括号、圆括号或无括号版本
			pattern := "(?:\\[|\\(|\[|\()?(?:" + strings.Join(patterns, "|") + ")(?:\\]|\\)|\]|\))?"
			re, err := regexp.Compile(`(?i)` + pattern) // (?i) 使匹配不区分大小写
			if err == nil {
				m.patterns = append(m.patterns, re)
			} else {
				// 如果编译失败，使用简单的字符串匹配（在Match方法中处理）
				m.patterns = append(m.patterns, nil)
			}
		} else {
			m.patterns = append(m.patterns, nil)
		}
	}
}

// getSortedReleaseGroups 获取按优先级排序的发布组列表
func (m *ReleaseGroupsMatcher) getSortedReleaseGroups() []ReleaseGroup {
	sorted := make([]ReleaseGroup, len(m.releaseGroups))
	copy(sorted, m.releaseGroups)
	
	// 按优先级降序排序
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})
	
	return sorted
}

// extractSimpleGroup 提取简单的发布组信息
func (m *ReleaseGroupsMatcher) extractSimpleGroup(text string) string {
	// 匹配方括号内的内容
	bracketPattern := regexp.MustCompile(`\[(.*?)\]`)
	matches := bracketPattern.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		if len(match) >= 2 {
			group := strings.TrimSpace(match[1])
			// 简单过滤：长度不能太短，且不能包含常见的质量标记
			if len(group) >= 2 && !m.isQualityTag(group) && !m.isResolutionTag(group) && !m.isCodecTag(group) {
				return group
			}
		}
	}
	
	// 匹配圆括号内的内容
	parenPattern := regexp.MustCompile(`\((.*?)\)`)
	matches = parenPattern.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		if len(match) >= 2 {
			group := strings.TrimSpace(match[1])
			// 简单过滤：长度不能太短，且不能包含常见的质量标记
			if len(group) >= 2 && !m.isQualityTag(group) && !m.isResolutionTag(group) && !m.isCodecTag(group) {
				return group
			}
		}
	}
	
	return ""
}

// isQualityTag 判断是否为质量标记
func (m *ReleaseGroupsMatcher) isQualityTag(text string) bool {
	qualityTags := []string{"WEB", "WEB-DL", "BD", "BluRay", "DVD", "HDTV", "HD", "SD", "UHD", "4K", "HDR", "SDR"}
	textLower := strings.ToLower(text)
	
	for _, tag := range qualityTags {
		if strings.ToLower(tag) == textLower {
			return true
		}
	}
	
	return false
}

// isResolutionTag 判断是否为分辨率标记
func (m *ReleaseGroupsMatcher) isResolutionTag(text string) bool {
	resolutionPattern := regexp.MustCompile(`^\d{3,4}[pP]$`)
	return resolutionPattern.MatchString(text)
}

// isCodecTag 判断是否为编码标记
func (m *ReleaseGroupsMatcher) isCodecTag(text string) bool {
	codecTags := []string{"h264", "h265", "hevc", "x264", "x265", "mpeg4", "divx", "xvid", "vp9", "av1"}
	textLower := strings.ToLower(text)
	
	for _, tag := range codecTags {
		if textLower == tag {
			return true
		}
	}
	
	return false
}

// getGroupPriority 获取发布组优先级
func (m *ReleaseGroupsMatcher) getGroupPriority(name string) int {
	for _, group := range m.releaseGroups {
		if strings.EqualFold(group.Name, name) {
			return group.Priority
		}
	}
	return 0
}

// setReleaseGroupActive 设置发布组激活状态
func (m *ReleaseGroupsMatcher) setReleaseGroupActive(name string, active bool) error {
	for i, group := range m.releaseGroups {
		if strings.EqualFold(group.Name, name) {
			m.releaseGroups[i].IsActive = active
			m.lastUpdated = time.Now()
			return nil
		}
	}
	
	return nil
}

// updateReleaseGroups 更新发布组列表（可以扩展为从配置或远程源加载）
func (m *ReleaseGroupsMatcher) updateReleaseGroups(ctx context.Context) {
	// 这里可以实现从配置文件、数据库或远程API更新发布组列表的逻辑
	// 目前简单地重新加载默认列表
	m.Lock()
	defer m.Unlock()
	
	// 保留当前激活状态的发布组
	activeGroups := make(map[string]bool)
	for _, group := range m.releaseGroups {
		activeGroups[group.Name] = group.IsActive
	}
	
	// 重新加载默认列表
	m.loadDefaultReleaseGroups()
	
	// 恢复激活状态
	for i, group := range m.releaseGroups {
		if active, exists := activeGroups[group.Name]; exists {
			m.releaseGroups[i].IsActive = active
		}
	}
	
	// 重新编译模式
	m.compilePatterns()
	m.lastUpdated = time.Now()
}

// GetReleaseGroupInfo 获取发布组详细信息
func (m *ReleaseGroupsMatcher) GetReleaseGroupInfo(ctx context.Context, name string) (ReleaseGroup, bool) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return ReleaseGroup{}, false
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	for _, group := range m.releaseGroups {
		if strings.EqualFold(group.Name, name) {
			return group, true
		}
	}
	
	return ReleaseGroup{}, false
}

// SearchReleaseGroups 搜索发布组
func (m *ReleaseGroupsMatcher) SearchReleaseGroups(ctx context.Context, keyword string) []ReleaseGroup {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return []ReleaseGroup{}
		}
	}
	
	m.RLock()
	defer m.RUnlock()
	
	if keyword == "" {
		return m.getSortedReleaseGroups()
	}
	
	keywordLower := strings.ToLower(keyword)
	results := make([]ReleaseGroup, 0)
	
	for _, group := range m.releaseGroups {
		// 搜索名称
		if strings.Contains(strings.ToLower(group.Name), keywordLower) {
			results = append(results, group)
			continue
		}
		
		// 搜索别名
		for _, alias := range group.Aliases {
			if strings.Contains(strings.ToLower(alias), keywordLower) {
				results = append(results, group)
				break
			}
		}
	}
	
	// 按优先级排序结果
	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority > results[j].Priority
	})
	
	return results
}

// ExportReleaseGroups 导出发布组列表（用于备份）
func (m *ReleaseGroupsMatcher) ExportReleaseGroups() []ReleaseGroup {
	m.RLock()
	defer m.RUnlock()
	
	// 返回副本以避免并发修改
	copyGroups := make([]ReleaseGroup, len(m.releaseGroups))
	copy(copyGroups, m.releaseGroups)
	return copyGroups
}

// ImportReleaseGroups 导入发布组列表（用于恢复）
func (m *ReleaseGroupsMatcher) ImportReleaseGroups(groups []ReleaseGroup) error {
	m.Lock()
	defer m.Unlock()
	
	// 验证导入的数据
	for _, group := range groups {
		if group.Name == "" {
			continue // 跳过无效的发布组
		}
	}
	
	// 导入发布组
	m.releaseGroups = groups
	
	// 重新编译模式
	m.compilePatterns()
	m.lastUpdated = time.Now()
	
	return nil
}

// GetStatistics 获取匹配器统计信息
func (m *ReleaseGroupsMatcher) GetStatistics() map[string]interface{} {
	m.RLock()
	defer m.RUnlock()
	
	totalGroups := len(m.releaseGroups)
	activeGroups := 0
	inactiveGroups := 0
	
	for _, group := range m.releaseGroups {
		if group.IsActive {
			activeGroups++
		} else {
			inactiveGroups++
		}
	}
	
	return map[string]interface{}{
		"total_groups":   totalGroups,
		"active_groups":  activeGroups,
		"inactive_groups": inactiveGroups,
		"last_updated":   m.lastUpdated,
		"initialized":    m.initialized,
	}
}

// ResetToDefaults 重置为默认发布组列表
func (m *ReleaseGroupsMatcher) ResetToDefaults() error {
	m.Lock()
	defer m.Unlock()
	
	m.loadDefaultReleaseGroups()
	m.compilePatterns()
	m.lastUpdated = time.Now()
	
	return nil
}