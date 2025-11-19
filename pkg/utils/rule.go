package utils

import (
	"fmt"
	"sync"
)

// RuleHelper 规则帮助类
type RuleHelper struct {
	ruleGroups map[string]*FilterRuleGroup
	mutex      sync.RWMutex
}

// FilterRuleGroup 过滤规则组
type FilterRuleGroup struct {
	Name      string        `json:"name"`
	MediaType string        `json:"media_type,omitempty"`
	Category  string        `json:"category,omitempty"`
	Rules     []FilterRule  `json:"rules"`
	Enabled   bool          `json:"enabled"`
	Priority  int           `json:"priority"`
}

// FilterRule 过滤规则
type FilterRule struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`     // include, exclude, regex
	Pattern  string      `json:"pattern"`
	Field    string      `json:"field"`    // title, year, genre, etc.
	Operator string      `json:"operator"` // equals, contains, starts_with, ends_with
	Value    interface{} `json:"value"`
	Enabled  bool        `json:"enabled"`
}

// CustomRule 自定义规则
type CustomRule struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Type     string `json:"type"`     // movie, tv
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Category string `json:"category"`
	Overview string `json:"overview"`
}

// NewRuleHelper 创建规则助手实例
func NewRuleHelper() *RuleHelper {
	return &RuleHelper{
		ruleGroups: make(map[string]*FilterRuleGroup),
	}
}

// GetRuleGroups 获取所有规则组
func (rh *RuleHelper) GetRuleGroups() []*FilterRuleGroup {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	groups := make([]*FilterRuleGroup, 0, len(rh.ruleGroups))
	for _, group := range rh.ruleGroups {
		groups = append(groups, group)
	}

	return groups
}

// GetRuleGroup 获取指定规则组
func (rh *RuleHelper) GetRuleGroup(groupName string) (*FilterRuleGroup, error) {
	if groupName == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("rule group not found: %s", groupName)
	}

	return group, nil
}

// AddRuleGroup 添加规则组
func (rh *RuleHelper) AddRuleGroup(group *FilterRuleGroup) error {
	if group == nil {
		return fmt.Errorf("rule group cannot be nil")
	}

	if group.Name == "" {
		return fmt.Errorf("rule group name cannot be empty")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	rh.ruleGroups[group.Name] = group
	return nil
}

// UpdateRuleGroup 更新规则组
func (rh *RuleHelper) UpdateRuleGroup(groupName string, updates map[string]interface{}) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return fmt.Errorf("rule group not found: %s", groupName)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if newName, ok := value.(string); ok {
				delete(rh.ruleGroups, groupName)
				group.Name = newName
				rh.ruleGroups[newName] = group
			}
		case "media_type":
			if mediaType, ok := value.(string); ok {
				group.MediaType = mediaType
			}
		case "category":
			if category, ok := value.(string); ok {
				group.Category = category
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				group.Enabled = enabled
			}
		case "priority":
			if priority, ok := value.(int); ok {
				group.Priority = priority
			}
		case "rules":
			if rules, ok := value.([]FilterRule); ok {
				group.Rules = rules
			}
		}
	}

	return nil
}

// RemoveRuleGroup 移除规则组
func (rh *RuleHelper) RemoveRuleGroup(groupName string) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	if _, exists := rh.ruleGroups[groupName]; !exists {
		return fmt.Errorf("rule group not found: %s", groupName)
	}

	delete(rh.ruleGroups, groupName)
	return nil
}

// GetRuleGroupsByMedia 根据媒体信息获取规则组
func (rh *RuleHelper) GetRuleGroupsByMedia(media *MediaInfo, groupNames []string) []*FilterRuleGroup {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	var retGroups []*FilterRuleGroup
	groups := rh.getFilteredGroups(groupNames)

	for _, group := range groups {
		if rh.matchesMedia(group, media) {
			retGroups = append(retGroups, group)
		}
	}

	return retGroups
}

// getFilteredGroups 获取过滤后的规则组
func (rh *RuleHelper) getFilteredGroups(groupNames []string) []*FilterRuleGroup {
	var groups []*FilterRuleGroup

	for _, group := range rh.ruleGroups {
		if !group.Enabled {
			continue
		}

		if len(groupNames) > 0 {
			found := false
			for _, name := range groupNames {
				if group.Name == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		groups = append(groups, group)
	}

	return groups
}

// matchesMedia 检查规则组是否匹配媒体信息
func (rh *RuleHelper) matchesMedia(group *FilterRuleGroup, media *MediaInfo) bool {
	if group.MediaType != "" && media != nil && group.MediaType != media.Type {
		return false
	}

	if group.Category != "" && media != nil && group.Category != media.Category {
		return false
	}

	return true
}

// ApplyRules 应用规则到媒体信息
func (rh *RuleHelper) ApplyRules(media *MediaInfo, groupNames []string) (bool, []string) {
	groups := rh.GetRuleGroupsByMedia(media, groupNames)
	
	if len(groups) == 0 {
		return false, []string{}
	}

	var matchedRules []string
	allowed := true

	for _, group := range groups {
		for _, rule := range group.Rules {
			if !rule.Enabled {
				continue
			}

			if rh.evaluateRule(rule, media) {
				matchedRules = append(matchedRules, fmt.Sprintf("%s.%s", group.Name, rule.Name))
				
				// 如果是排除规则，则不允许
				if rule.Type == "exclude" {
					allowed = false
				}
			}
		}
	}

	return allowed, matchedRules
}

// evaluateRule 评估单个规则
func (rh *RuleHelper) evaluateRule(rule FilterRule, media *MediaInfo) bool {
	if media == nil {
		return false
	}

	fieldValue := rh.getFieldValue(media, rule.Field)
	if fieldValue == "" {
		return false
	}

	switch rule.Operator {
	case "equals":
		return rh.compareEquals(fieldValue, rule.Value)
	case "contains":
		return rh.compareContains(fieldValue, rule.Value)
	case "starts_with":
		return rh.compareStartsWith(fieldValue, rule.Value)
	case "ends_with":
		return rh.compareEndsWith(fieldValue, rule.Value)
	case "regex":
		return rh.compareRegex(fieldValue, rule.Value)
	default:
		return false
	}
}

// getFieldValue 获取字段值
func (rh *RuleHelper) getFieldValue(media *MediaInfo, field string) string {
	switch field {
	case "title":
		return media.Title
	case "year":
		return fmt.Sprintf("%d", media.Year)
	case "genre":
		return media.Genre
	case "category":
		return media.Category
	case "overview":
		return media.Overview
	default:
		return ""
	}
}

// compareEquals 比较相等
func (rh *RuleHelper) compareEquals(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		return fieldValue == valueStr
	}
	return false
}

// compareContains 比较包含
func (rh *RuleHelper) compareContains(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		return contains(fieldValue, valueStr)
	}
	return false
}

// compareStartsWith 比较开始于
func (rh *RuleHelper) compareStartsWith(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		return hasPrefix(fieldValue, valueStr)
	}
	return false
}

// compareEndsWith 比较结束于
func (rh *RuleHelper) compareEndsWith(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		return hasSuffix(fieldValue, valueStr)
	}
	return false
}

// compareRegex 比较正则表达式
func (rh *RuleHelper) compareRegex(fieldValue string, ruleValue interface{}) bool {
	// 简化实现，实际应该使用正则表达式
	if valueStr, ok := ruleValue.(string); ok {
		return contains(fieldValue, valueStr)
	}
	return false
}

// AddRule 添加规则到规则组
func (rh *RuleHelper) AddRule(groupName string, rule *FilterRule) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return fmt.Errorf("rule group not found: %s", groupName)
	}

	group.Rules = append(group.Rules, *rule)
	return nil
}

// RemoveRule 从规则组移除规则
func (rh *RuleHelper) RemoveRule(groupName, ruleName string) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if ruleName == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return fmt.Errorf("rule group not found: %s", groupName)
	}

	for i, rule := range group.Rules {
		if rule.Name == ruleName {
			group.Rules = append(group.Rules[:i], group.Rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleName)
}

// UpdateRule 更新规则
func (rh *RuleHelper) UpdateRule(groupName, ruleName string, updates map[string]interface{}) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if ruleName == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return fmt.Errorf("rule group not found: %s", groupName)
	}

	for i, rule := range group.Rules {
		if rule.Name == ruleName {
			// 应用更新
			for key, value := range updates {
				switch key {
				case "name":
					if name, ok := value.(string); ok {
						group.Rules[i].Name = name
					}
				case "type":
					if ruleType, ok := value.(string); ok {
						group.Rules[i].Type = ruleType
					}
				case "pattern":
					if pattern, ok := value.(string); ok {
						group.Rules[i].Pattern = pattern
					}
				case "field":
					if field, ok := value.(string); ok {
						group.Rules[i].Field = field
					}
				case "operator":
					if operator, ok := value.(string); ok {
						group.Rules[i].Operator = operator
					}
				case "value":
					group.Rules[i].Value = value
				case "enabled":
					if enabled, ok := value.(bool); ok {
						group.Rules[i].Enabled = enabled
					}
				}
			}
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleName)
}

// GetRule 获取规则
func (rh *RuleHelper) GetRule(groupName, ruleName string) (*FilterRule, error) {
	if groupName == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	if ruleName == "" {
		return nil, fmt.Errorf("rule name cannot be empty")
	}

	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("rule group not found: %s", groupName)
	}

	for _, rule := range group.Rules {
		if rule.Name == ruleName {
			return &rule, nil
		}
	}

	return nil, fmt.Errorf("rule not found: %s", ruleName)
}

// ValidateRule 验证规则
func (rh *RuleHelper) ValidateRule(rule *FilterRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	if rule.Type == "" {
		return fmt.Errorf("rule type cannot be empty")
	}

	if rule.Field == "" {
		return fmt.Errorf("rule field cannot be empty")
	}

	if rule.Operator == "" {
		return fmt.Errorf("rule operator cannot be empty")
	}

	// 验证规则类型
	validTypes := []string{"include", "exclude", "regex"}
	if !containsString(validTypes, rule.Type) {
		return fmt.Errorf("invalid rule type: %s", rule.Type)
	}

	// 验证操作符
	validOperators := []string{"equals", "contains", "starts_with", "ends_with", "regex"}
	if !containsString(validOperators, rule.Operator) {
		return fmt.Errorf("invalid operator: %s", rule.Operator)
	}

	// 验证字段
	validFields := []string{"title", "year", "genre", "category", "overview"}
	if !containsString(validFields, rule.Field) {
		return fmt.Errorf("invalid field: %s", rule.Field)
	}

	return nil
}

// GetRuleCount 获取规则数量
func (rh *RuleHelper) GetRuleCount() int {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	count := 0
	for _, group := range rh.ruleGroups {
		count += len(group.Rules)
	}

	return count
}

// GetGroupCount 获取规则组数量
func (rh *RuleHelper) GetGroupCount() int {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	return len(rh.ruleGroups)
}

// ClearRules 清空所有规则
func (rh *RuleHelper) ClearRules() {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	rh.ruleGroups = make(map[string]*FilterRuleGroup)
}

// ExportRules 导出规则
func (rh *RuleHelper) ExportRules() map[string]*FilterRuleGroup {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	// 返回副本
	export := make(map[string]*FilterRuleGroup)
	for name, group := range rh.ruleGroups {
		export[name] = &FilterRuleGroup{
			Name:      group.Name,
			MediaType: group.MediaType,
			Category:  group.Category,
			Rules:     append([]FilterRule{}, group.Rules...),
			Enabled:   group.Enabled,
			Priority:  group.Priority,
		}
	}

	return export
}

// ImportRules 导入规则
func (rh *RuleHelper) ImportRules(rules map[string]*FilterRuleGroup) error {
	if rules == nil {
		return fmt.Errorf("rules cannot be nil")
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	rh.ruleGroups = make(map[string]*FilterRuleGroup)
	for name, group := range rules {
		rh.ruleGroups[name] = group
	}

	return nil
}

// containsString 检查字符串是否在切片中
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}