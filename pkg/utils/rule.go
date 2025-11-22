package utils

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
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
	logger.Debug("Creating new RuleHelper instance", zap.String("func", "NewRuleHelper"))
	return &RuleHelper{
		ruleGroups: make(map[string]*FilterRuleGroup),
	}
}

// GetRuleGroups 获取所有规则组
func (rh *RuleHelper) GetRuleGroups() []*FilterRuleGroup {
	logger.Debug("Getting all rule groups", zap.String("func", "GetRuleGroups"))
	
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	groups := make([]*FilterRuleGroup, 0, len(rh.ruleGroups))
	for _, group := range rh.ruleGroups {
		groups = append(groups, group)
	}

	logger.Debug("Retrieved rule groups", zap.Int("count", len(groups)), zap.String("func", "GetRuleGroups"))
	return groups
}

// GetRuleGroup 获取指定规则组
func (rh *RuleHelper) GetRuleGroup(groupName string) (*FilterRuleGroup, error) {
	logger.Debug("Getting rule group", zap.String("group_name", groupName), zap.String("func", "GetRuleGroup"))
	
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Invalid group name", zap.String("error", err.Error()), zap.String("func", "GetRuleGroup"))
		return nil, err
	}

	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Rule group not found", zap.String("group_name", groupName), zap.String("error", err.Error()), zap.String("func", "GetRuleGroup"))
		return nil, err
	}

	logger.Debug("Successfully retrieved rule group", zap.String("group_name", groupName), zap.String("func", "GetRuleGroup"))
	return group, nil
}

// AddRuleGroup 添加规则组
func (rh *RuleHelper) AddRuleGroup(group *FilterRuleGroup) error {
	logger.Debug("Adding rule group", zap.String("group_name", group.Name), zap.String("func", "AddRuleGroup"))
	
	if group == nil {
		err := errors.NewAppError(400, "rule group cannot be nil", "")
		logger.Error("Failed to add rule group", zap.String("error", err.Error()), zap.String("func", "AddRuleGroup"))
		return err
	}

	if group.Name == "" {
		err := errors.NewAppError(400, "rule group name cannot be empty", "")
		logger.Error("Failed to add rule group", zap.String("error", err.Error()), zap.String("func", "AddRuleGroup"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	// 检查是否已存在
	if _, exists := rh.ruleGroups[group.Name]; exists {
		err := errors.NewAppError(409, fmt.Sprintf("rule group already exists: %s", group.Name), "")
		logger.Warn("Rule group already exists", zap.String("group_name", group.Name), zap.String("error", err.Error()), zap.String("func", "AddRuleGroup"))
		return err
	}

	rh.ruleGroups[group.Name] = group
	logger.Info("Successfully added rule group", zap.String("group_name", group.Name), zap.String("media_type", group.MediaType), zap.String("func", "AddRuleGroup"))
	return nil
}

// UpdateRuleGroup 更新规则组
func (rh *RuleHelper) UpdateRuleGroup(groupName string, updates map[string]interface{}) error {
	logger.Debug("Updating rule group", zap.String("group_name", groupName), zap.Any("updates", updates), zap.String("func", "UpdateRuleGroup"))
	
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to update rule group", zap.String("error", err.Error()), zap.String("func", "UpdateRuleGroup"))
		return err
	}

	if len(updates) == 0 {
		err := errors.NewAppError(400, "updates cannot be empty", "")
		logger.Error("Failed to update rule group", zap.String("error", err.Error()), zap.String("func", "UpdateRuleGroup"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to update rule group", zap.String("group_name", groupName), zap.String("error", err.Error()), zap.String("func", "UpdateRuleGroup"))
		return err
	}

	// 应用更新
	originalName := group.Name
	for key, value := range updates {
		switch key {
		case "name":
			if newName, ok := value.(string); ok && newName != "" {
				delete(rh.ruleGroups, groupName)
				group.Name = newName
				rh.ruleGroups[newName] = group
				groupName = newName
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

	logger.Info("Successfully updated rule group", zap.String("original_name", originalName), zap.String("new_name", groupName), zap.String("func", "UpdateRuleGroup"))
	return nil
}

// RemoveRuleGroup 移除规则组
func (rh *RuleHelper) RemoveRuleGroup(groupName string) error {
	logger.Debug("Removing rule group", zap.String("group_name", groupName), zap.String("func", "RemoveRuleGroup"))
	
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to remove rule group", zap.String("error", err.Error()), zap.String("func", "RemoveRuleGroup"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to remove rule group", zap.String("group_name", groupName), zap.String("error", err.Error()), zap.String("func", "RemoveRuleGroup"))
		return err
	}

	delete(rh.ruleGroups, groupName)
	logger.Info("Successfully removed rule group", zap.String("group_name", groupName), zap.String("media_type", group.MediaType), zap.String("func", "RemoveRuleGroup"))
	return nil
}

// GetRuleGroupsByMedia 根据媒体信息获取规则组
func (rh *RuleHelper) GetRuleGroupsByMedia(media *MediaInfo, groupNames []string) []*FilterRuleGroup {
	logger.Debug("Getting rule groups by media", zap.String("media_type", media.Type), zap.String("media_title", media.Title), zap.Strings("group_names", groupNames), zap.String("func", "GetRuleGroupsByMedia"))
	
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	var retGroups []*FilterRuleGroup
	groups := rh.getFilteredGroups(groupNames)

	for _, group := range groups {
		if rh.matchesMedia(group, media) {
			retGroups = append(retGroups, group)
		}
	}

	logger.Debug("Retrieved rule groups by media", zap.Int("count", len(retGroups)), zap.String("media_title", media.Title), zap.String("func", "GetRuleGroupsByMedia"))
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
	logger.Debug("Applying rules to media", zap.String("media_type", media.Type), zap.String("media_title", media.Title), zap.Strings("group_names", groupNames), zap.String("func", "ApplyRules"))
	
	groups := rh.GetRuleGroupsByMedia(media, groupNames)
	
	if len(groups) == 0 {
		logger.Debug("No matching rule groups found", zap.String("media_title", media.Title), zap.String("func", "ApplyRules"))
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
				ruleName := fmt.Sprintf("%s.%s", group.Name, rule.Name)
				matchedRules = append(matchedRules, ruleName)
				logger.Debug("Rule matched", zap.String("rule_name", ruleName), zap.String("media_title", media.Title), zap.String("func", "ApplyRules"))
				
				// 如果是排除规则，则不允许
				if rule.Type == "exclude" {
					allowed = false
					logger.Debug("Exclude rule matched, media not allowed", zap.String("rule_name", ruleName), zap.String("media_title", media.Title), zap.String("func", "ApplyRules"))
				}
			}
		}
	}

	logger.Info("Rules applied to media", zap.String("media_title", media.Title), zap.Bool("allowed", allowed), zap.Int("matched_rules_count", len(matchedRules)), zap.String("func", "ApplyRules"))
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
		matched := strings.Contains(strings.ToLower(fieldValue), strings.ToLower(valueStr))
		logger.Debug("Contains comparison", zap.String("field_value", fieldValue), zap.String("rule_value", valueStr), zap.Bool("matched", matched), zap.String("func", "compareContains"))
		return matched
	}
	return false
}

// compareStartsWith 比较开始于
func (rh *RuleHelper) compareStartsWith(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		matched := strings.HasPrefix(strings.ToLower(fieldValue), strings.ToLower(valueStr))
		logger.Debug("Starts with comparison", zap.String("field_value", fieldValue), zap.String("rule_value", valueStr), zap.Bool("matched", matched), zap.String("func", "compareStartsWith"))
		return matched
	}
	return false
}

// compareEndsWith 比较结束于
func (rh *RuleHelper) compareEndsWith(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		matched := strings.HasSuffix(strings.ToLower(fieldValue), strings.ToLower(valueStr))
		logger.Debug("Ends with comparison", zap.String("field_value", fieldValue), zap.String("rule_value", valueStr), zap.Bool("matched", matched), zap.String("func", "compareEndsWith"))
		return matched
	}
	return false
}

// compareRegex 比较正则表达式
func (rh *RuleHelper) compareRegex(fieldValue string, ruleValue interface{}) bool {
	if valueStr, ok := ruleValue.(string); ok {
		// 编译正则表达式
		pattern, err := regexp.Compile(valueStr)
		if err != nil {
			logger.Error("Invalid regex pattern", zap.String("pattern", valueStr), zap.String("error", err.Error()), zap.String("func", "compareRegex"))
			return false
		}
		
		matched := pattern.MatchString(fieldValue)
		logger.Debug("Regex comparison", zap.String("field_value", fieldValue), zap.String("pattern", valueStr), zap.Bool("matched", matched), zap.String("func", "compareRegex"))
		return matched
	}
	return false
}

// AddRule 添加规则到规则组
func (rh *RuleHelper) AddRule(groupName string, rule *FilterRule) error {
	logger.Debug("Adding rule to group", zap.String("group_name", groupName), zap.String("rule_name", rule.Name), zap.String("func", "AddRule"))
	
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to add rule", zap.String("error", err.Error()), zap.String("func", "AddRule"))
		return err
	}

	if rule == nil {
		err := errors.NewAppError(400, "rule cannot be nil", "")
		logger.Error("Failed to add rule", zap.String("error", err.Error()), zap.String("func", "AddRule"))
		return err
	}

	// 验证规则
	if err := rh.ValidateRule(rule); err != nil {
		logger.Error("Invalid rule", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "AddRule"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to add rule", zap.String("group_name", groupName), zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "AddRule"))
		return err
	}

	// 检查规则是否已存在
	for _, existingRule := range group.Rules {
		if existingRule.Name == rule.Name {
			err := errors.NewAppError(409, fmt.Sprintf("rule already exists in group: %s.%s", groupName, rule.Name), "")
			logger.Warn("Rule already exists", zap.String("group_name", groupName), zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "AddRule"))
			return err
		}
	}

	group.Rules = append(group.Rules, *rule)
	logger.Info("Successfully added rule to group", zap.String("group_name", groupName), zap.String("rule_name", rule.Name), zap.String("rule_type", rule.Type), zap.String("func", "AddRule"))
	return nil
}

// RemoveRule 从规则组移除规则
func (rh *RuleHelper) RemoveRule(groupName, ruleName string) error {
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to remove rule", zap.String("error", err.Error()), zap.String("func", "RemoveRule"))
		return err
	}

	if ruleName == "" {
		err := errors.NewAppError(400, "rule name cannot be empty", "")
		logger.Error("Failed to remove rule", zap.String("error", err.Error()), zap.String("func", "RemoveRule"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to remove rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "RemoveRule"))
		return err
	}

	for i, rule := range group.Rules {
		if rule.Name == ruleName {
			group.Rules = append(group.Rules[:i], group.Rules[i+1:]...)
			logger.Info("Successfully removed rule from group", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("func", "RemoveRule"))
			return nil
		}
	}

	err := errors.NewAppError(404, fmt.Sprintf("rule not found: %s", ruleName), "")
	logger.Error("Failed to remove rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "RemoveRule"))
	return err
}

// UpdateRule 更新规则
func (rh *RuleHelper) UpdateRule(groupName, ruleName string, updates map[string]interface{}) error {
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to update rule", zap.String("error", err.Error()), zap.String("func", "UpdateRule"))
		return err
	}

	if ruleName == "" {
		err := errors.NewAppError(400, "rule name cannot be empty", "")
		logger.Error("Failed to update rule", zap.String("error", err.Error()), zap.String("func", "UpdateRule"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to update rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "UpdateRule"))
		return err
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
			logger.Info("Successfully updated rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("func", "UpdateRule"))
			return nil
		}
	}

	err := errors.NewAppError(404, fmt.Sprintf("rule not found: %s", ruleName), "")
	logger.Error("Failed to update rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "UpdateRule"))
	return err
}

// GetRule 获取规则
func (rh *RuleHelper) GetRule(groupName, ruleName string) (*FilterRule, error) {
	if groupName == "" {
		err := errors.NewAppError(400, "group name cannot be empty", "")
		logger.Error("Failed to get rule", zap.String("error", err.Error()), zap.String("func", "GetRule"))
		return nil, err
	}

	if ruleName == "" {
		err := errors.NewAppError(400, "rule name cannot be empty", "")
		logger.Error("Failed to get rule", zap.String("error", err.Error()), zap.String("func", "GetRule"))
		return nil, err
	}

	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	group, exists := rh.ruleGroups[groupName]
	if !exists {
		err := errors.NewAppError(404, fmt.Sprintf("rule group not found: %s", groupName), "")
		logger.Error("Failed to get rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "GetRule"))
		return nil, err
	}

	for _, rule := range group.Rules {
		if rule.Name == ruleName {
			logger.Debug("Successfully retrieved rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("func", "GetRule"))
			return &rule, nil
		}
	}

	err := errors.NewAppError(404, fmt.Sprintf("rule not found: %s", ruleName), "")
	logger.Error("Failed to get rule", zap.String("group_name", groupName), zap.String("rule_name", ruleName), zap.String("error", err.Error()), zap.String("func", "GetRule"))
	return nil, err
}

// ValidateRule 验证规则
func (rh *RuleHelper) ValidateRule(rule *FilterRule) error {
	logger.Debug("Validating rule", zap.String("rule_name", rule.Name), zap.String("func", "ValidateRule"))
	
	if rule == nil {
		err := errors.NewAppError(400, "rule cannot be nil", "")
		logger.Error("Rule validation failed", zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	if rule.Name == "" {
		err := errors.NewAppError(400, "rule name cannot be empty", "")
		logger.Error("Rule validation failed", zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	if rule.Type == "" {
		err := errors.NewAppError(400, "rule type cannot be empty", "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	if rule.Field == "" {
		err := errors.NewAppError(400, "rule field cannot be empty", "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	if rule.Operator == "" {
		err := errors.NewAppError(400, "rule operator cannot be empty", "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	// 验证规则类型
	validTypes := []string{"include", "exclude", "regex"}
	if !containsString(validTypes, rule.Type) {
		err := errors.NewAppError(400, fmt.Sprintf("invalid rule type: %s", rule.Type), "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	// 验证操作符
	validOperators := []string{"equals", "contains", "starts_with", "ends_with", "regex"}
	if !containsString(validOperators, rule.Operator) {
		err := errors.NewAppError(400, fmt.Sprintf("invalid operator: %s", rule.Operator), "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	// 验证字段
	validFields := []string{"title", "year", "genre", "category", "overview"}
	if !containsString(validFields, rule.Field) {
		err := errors.NewAppError(400, fmt.Sprintf("invalid field: %s", rule.Field), "")
		logger.Error("Rule validation failed", zap.String("rule_name", rule.Name), zap.String("error", err.Error()), zap.String("func", "ValidateRule"))
		return err
	}

	logger.Debug("Rule validation successful", zap.String("rule_name", rule.Name), zap.String("func", "ValidateRule"))
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
	logger.Info("Clearing all rules", zap.String("func", "ClearRules"))
	
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	count := len(rh.ruleGroups)
	rh.ruleGroups = make(map[string]*FilterRuleGroup)
	
	logger.Info("Cleared all rules", zap.Int("previous_count", count), zap.String("func", "ClearRules"))
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
		err := errors.NewAppError(400, "rules cannot be nil", "")
		logger.Error("Failed to import rules", zap.String("error", err.Error()), zap.String("func", "ImportRules"))
		return err
	}

	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	rh.ruleGroups = make(map[string]*FilterRuleGroup)
	for name, group := range rules {
		rh.ruleGroups[name] = group
	}

	logger.Info("Successfully imported rules", zap.Int("group_count", len(rules)), zap.String("func", "ImportRules"))
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