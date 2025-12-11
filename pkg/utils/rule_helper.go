package utils

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

// FilterRuleGroup 规则组
type FilterRuleGroup struct {
	Name      string       `json:"name"`
	MediaType MediaType    `json:"media_type"`
	Category  string       `json:"category"`
	Rules     []FilterRule `json:"rules"`
}

// FilterRule 过滤规则
type FilterRule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

// CustomRule 自定义规则
type CustomRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Rules       []string `json:"rules"`
	Enabled     bool     `json:"enabled"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title    string    `json:"title"`
	Type     MediaType `json:"type"`
	Category string    `json:"category"`
}

// ConfigStore 配置存储接口
type ConfigStore interface {
	// Get 获取配置
	Get(key string, defaultValue any) any
	// Set 设置配置
	Set(key string, value any)
}

// MemoryConfigStore 内存配置存储实现
type MemoryConfigStore struct {
	data  map[string]any
	mutex sync.RWMutex
}

// NewMemoryConfigStore 创建内存配置存储
func NewMemoryConfigStore() *MemoryConfigStore {
	return &MemoryConfigStore{
		data: make(map[string]any),
	}
}

// Get 获取配置
func (s *MemoryConfigStore) Get(key string, defaultValue any) any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if val, ok := s.data[key]; ok {
		return val
	}
	return defaultValue
}

// Set 设置配置
func (s *MemoryConfigStore) Set(key string, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
}

// RuleHelper 规则帮助类
type RuleHelper struct {
	logger      *zap.Logger
	configStore ConfigStore
	mutex       sync.RWMutex
}

// RuleSystemConfigKey 规则系统配置键
type RuleSystemConfigKey string

const (
	RuleSystemConfigKeyUserFilterRuleGroups RuleSystemConfigKey = "UserFilterRuleGroups"
	RuleSystemConfigKeyCustomFilterRules    RuleSystemConfigKey = "CustomFilterRules"
)

// NewRuleHelper 创建规则帮助类实例
func NewRuleHelper(configStore ...ConfigStore) *RuleHelper {
	var store ConfigStore
	if len(configStore) > 0 {
		store = configStore[0]
	} else {
		store = NewMemoryConfigStore()
	}
	return &RuleHelper{
		logger:      logger.GetLogger(),
		configStore: store,
	}
}

// GetRuleGroups 获取所有规则组
func (h *RuleHelper) GetRuleGroups() []*FilterRuleGroup {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 从配置存储中获取规则组
	ruleGroups := h.configStore.Get(string(RuleSystemConfigKeyUserFilterRuleGroups), []*FilterRuleGroup{}).([]*FilterRuleGroup)
	if ruleGroups == nil {
		return []*FilterRuleGroup{}
	}

	return ruleGroups
}

// GetRuleGroup 获取指定名称的规则组
func (h *RuleHelper) GetRuleGroup(groupName string) (*FilterRuleGroup, error) {
	ruleGroups := h.GetRuleGroups()
	for _, group := range ruleGroups {
		if group.Name == groupName {
			return group, nil
		}
	}
	return nil, fmt.Errorf("规则组不存在: %s", groupName)
}

// GetRuleGroupByMedia 根据媒体信息获取规则组
func (h *RuleHelper) GetRuleGroupByMedia(media *MediaInfo, groupNames []string) []*FilterRuleGroup {
	retGroups := make([]*FilterRuleGroup, 0)
	ruleGroups := h.GetRuleGroups()

	// 如果指定了规则组名称，则过滤规则组
	if len(groupNames) > 0 {
		filteredGroups := make([]*FilterRuleGroup, 0)
		for _, group := range ruleGroups {
			for _, name := range groupNames {
				if group.Name == name {
					filteredGroups = append(filteredGroups, group)
					break
				}
			}
		}
		ruleGroups = filteredGroups
	}

	// 根据媒体信息过滤规则组
	for _, group := range ruleGroups {
		if group.MediaType == "" {
			// 没有指定媒体类型，匹配所有媒体
			retGroups = append(retGroups, group)
		} else if media != nil {
			if group.Category == "" && group.MediaType == media.Type {
				// 匹配媒体类型
				retGroups = append(retGroups, group)
			} else if group.Category == media.Category {
				// 匹配分类
				retGroups = append(retGroups, group)
			}
		}
	}

	return retGroups
}

// GetCustomRules 获取所有自定义规则
func (h *RuleHelper) GetCustomRules() []*CustomRule {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 从配置存储中获取自定义规则
	customRules := h.configStore.Get(string(RuleSystemConfigKeyCustomFilterRules), []*CustomRule{}).([]*CustomRule)
	if customRules == nil {
		return []*CustomRule{}
	}

	return customRules
}

// GetCustomRule 获取指定ID的自定义规则
func (h *RuleHelper) GetCustomRule(ruleID string) *CustomRule {
	customRules := h.GetCustomRules()
	for _, rule := range customRules {
		if rule.ID == ruleID {
			return rule
		}
	}
	return nil
}

// AddRuleGroup 添加规则组
func (h *RuleHelper) AddRuleGroup(group *FilterRuleGroup) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	// 检查是否已存在
	for _, g := range ruleGroups {
		if g.Name == group.Name {
			return fmt.Errorf("规则组已存在: %s", group.Name)
		}
	}
	ruleGroups = append(ruleGroups, group)
	h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
	h.logger.Info("添加规则组成功", zap.String("name", group.Name))
	return nil
}

// UpdateRuleGroup 更新规则组（支持部分更新）
func (h *RuleHelper) UpdateRuleGroup(groupName string, updates map[string]any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	for i, g := range ruleGroups {
		if g.Name == groupName {
			// 应用更新
			if mediaType, ok := updates["media_type"].(string); ok {
				g.MediaType = MediaType(mediaType)
			}
			if category, ok := updates["category"].(string); ok {
				g.Category = category
			}
			if rules, ok := updates["rules"].([]FilterRule); ok {
				g.Rules = rules
			}
			ruleGroups[i] = g
			h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
			h.logger.Info("更新规则组成功", zap.String("name", groupName))
			return nil
		}
	}

	return fmt.Errorf("规则组不存在: %s", groupName)
}

// RemoveRuleGroup 删除规则组（别名）
func (h *RuleHelper) RemoveRuleGroup(groupName string) error {
	return h.DeleteRuleGroup(groupName)
}

// DeleteRuleGroup 删除规则组
func (h *RuleHelper) DeleteRuleGroup(groupName string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	for i, g := range ruleGroups {
		if g.Name == groupName {
			ruleGroups = append(ruleGroups[:i], ruleGroups[i+1:]...)
			h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
			h.logger.Info("删除规则组成功", zap.String("name", groupName))
			return nil
		}
	}

	return fmt.Errorf("规则组不存在: %s", groupName)
}

// AddCustomRule 添加自定义规则
func (h *RuleHelper) AddCustomRule(rule *CustomRule) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	customRules := h.GetCustomRules()
	customRules = append(customRules, rule)
	h.configStore.Set(string(RuleSystemConfigKeyCustomFilterRules), customRules)
	h.logger.Info("添加自定义规则成功", zap.String("id", rule.ID), zap.String("name", rule.Name))
}

// UpdateCustomRule 更新自定义规则
func (h *RuleHelper) UpdateCustomRule(rule *CustomRule) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	customRules := h.GetCustomRules()
	for i, r := range customRules {
		if r.ID == rule.ID {
			customRules[i] = rule
			h.configStore.Set(string(RuleSystemConfigKeyCustomFilterRules), customRules)
			h.logger.Info("更新自定义规则成功", zap.String("id", rule.ID), zap.String("name", rule.Name))
			return nil
		}
	}

	return fmt.Errorf("自定义规则不存在: %s", rule.ID)
}

// DeleteCustomRule 删除自定义规则
func (h *RuleHelper) DeleteCustomRule(ruleID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	customRules := h.GetCustomRules()
	for i, r := range customRules {
		if r.ID == ruleID {
			customRules = append(customRules[:i], customRules[i+1:]...)
			h.configStore.Set(string(RuleSystemConfigKeyCustomFilterRules), customRules)
			h.logger.Info("删除自定义规则成功", zap.String("id", ruleID))
			return nil
		}
	}

	return fmt.Errorf("自定义规则不存在: %s", ruleID)
}

// AddRule 添加规则到规则组
func (h *RuleHelper) AddRule(groupName string, rule *FilterRule) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	for i, g := range ruleGroups {
		if g.Name == groupName {
			// 检查规则是否已存在
			for _, r := range g.Rules {
				if r.Name == rule.Name {
					return fmt.Errorf("规则已存在: %s", rule.Name)
				}
			}
			g.Rules = append(g.Rules, *rule)
			ruleGroups[i] = g
			h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
			h.logger.Info("添加规则成功", zap.String("group", groupName), zap.String("rule", rule.Name))
			return nil
		}
	}
	return fmt.Errorf("规则组不存在: %s", groupName)
}

// UpdateRule 更新规则
func (h *RuleHelper) UpdateRule(groupName string, ruleName string, updates map[string]any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	for i, g := range ruleGroups {
		if g.Name == groupName {
			for j, r := range g.Rules {
				if r.Name == ruleName {
					// 应用更新
					if ruleType, ok := updates["type"].(string); ok {
						r.Type = ruleType
					}
					if operator, ok := updates["operator"].(string); ok {
						r.Operator = operator
					}
					if value, ok := updates["value"].(string); ok {
						r.Value = value
					}
					if action, ok := updates["action"].(string); ok {
						r.Action = action
					}
					if priority, ok := updates["priority"].(int); ok {
						r.Priority = priority
					}
					if enabled, ok := updates["enabled"].(bool); ok {
						r.Enabled = enabled
					}
					g.Rules[j] = r
					ruleGroups[i] = g
					h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
					h.logger.Info("更新规则成功", zap.String("group", groupName), zap.String("rule", ruleName))
					return nil
				}
			}
			return fmt.Errorf("规则不存在: %s", ruleName)
		}
	}
	return fmt.Errorf("规则组不存在: %s", groupName)
}

// RemoveRule 删除规则
func (h *RuleHelper) RemoveRule(groupName string, ruleName string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := h.GetRuleGroups()
	for i, g := range ruleGroups {
		if g.Name == groupName {
			for j, r := range g.Rules {
				if r.Name == ruleName {
					g.Rules = append(g.Rules[:j], g.Rules[j+1:]...)
					ruleGroups[i] = g
					h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
					h.logger.Info("删除规则成功", zap.String("group", groupName), zap.String("rule", ruleName))
					return nil
				}
			}
			return fmt.Errorf("规则不存在: %s", ruleName)
		}
	}
	return fmt.Errorf("规则组不存在: %s", groupName)
}

// GetRule 获取规则
func (h *RuleHelper) GetRule(groupName string, ruleName string) (*FilterRule, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	ruleGroups := h.GetRuleGroups()
	for _, g := range ruleGroups {
		if g.Name == groupName {
			for _, r := range g.Rules {
				if r.Name == ruleName {
					rule := r // 复制
					return &rule, nil
				}
			}
			return nil, fmt.Errorf("规则不存在: %s", ruleName)
		}
	}
	return nil, fmt.Errorf("规则组不存在: %s", groupName)
}

// ApplyRules 应用规则到媒体
func (h *RuleHelper) ApplyRules(media *MediaInfo, groupNames []string) (bool, []string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	matchedRules := make([]string, 0)
	allowed := true

	// 获取适用的规则组
	ruleGroups := h.GetRuleGroupByMedia(media, groupNames)

	// 应用规则
	for _, group := range ruleGroups {
		for _, rule := range group.Rules {
			if !rule.Enabled {
				continue
			}
			// TODO: 实现规则匹配逻辑
			// 这里简化处理，实际需要根据rule.Type, rule.Operator, rule.Value进行匹配
			matchedRules = append(matchedRules, rule.Name)
			if rule.Action == "block" {
				allowed = false
			}
		}
	}

	return allowed, matchedRules
}

// GetGroupCount 获取规则组数量
func (h *RuleHelper) GetGroupCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.GetRuleGroups())
}

// GetRuleCount 获取规则总数
func (h *RuleHelper) GetRuleCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	count := 0
	ruleGroups := h.GetRuleGroups()
	for _, g := range ruleGroups {
		count += len(g.Rules)
	}
	return count
}

// ImportRules 导入规则
func (h *RuleHelper) ImportRules(rules map[string]*FilterRuleGroup) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ruleGroups := make([]*FilterRuleGroup, 0, len(rules))
	for _, group := range rules {
		ruleGroups = append(ruleGroups, group)
	}
	h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), ruleGroups)
	h.logger.Info("导入规则成功", zap.Int("count", len(rules)))
	return nil
}

// ExportRules 导出规则
func (h *RuleHelper) ExportRules() map[string]*FilterRuleGroup {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	ruleGroups := h.GetRuleGroups()
	result := make(map[string]*FilterRuleGroup, len(ruleGroups))
	for _, g := range ruleGroups {
		result[g.Name] = g
	}
	return result
}

// ClearRules 清空所有规则
func (h *RuleHelper) ClearRules() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.configStore.Set(string(RuleSystemConfigKeyUserFilterRuleGroups), []*FilterRuleGroup{})
	h.logger.Info("清空所有规则成功")
}
