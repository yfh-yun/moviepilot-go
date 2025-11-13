package helper

import (
	"moviepilot-go/internal/db"
	"moviepilot-go/pkg/models"
)

// RuleHelper 规划帮助�?type RuleHelper struct{}

// NewRuleHelper 创建RuleHelper实例
func NewRuleHelper() *RuleHelper {
	return &RuleHelper{}
}

// GetRuleGroups 获取用户所有规则组
func (r *RuleHelper) GetRuleGroups() []models.FilterRuleGroup {
	// 从系统配置中获取规则�?	ruleGroups := db.NewSystemConfigOper().Get(models.SystemConfigKeyUserFilterRuleGroups)
	if ruleGroups == nil {
		return []models.FilterRuleGroup{}
	}
	
	// 类型断言获取规则组数�?	groups, ok := ruleGroups.([]interface{})
	if !ok {
		return []models.FilterRuleGroup{}
	}
	
	// 转换为FilterRuleGroup数组
	result := make([]models.FilterRuleGroup, 0, len(groups))
	for _, group := range groups {
		// 将map转换为FilterRuleGroup结构�?		if groupMap, ok := group.(map[string]interface{}); ok {
			filterGroup := models.FilterRuleGroup{}
			
			if name, exists := groupMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					filterGroup.Name = nameStr
				}
			}
			
			if ruleString, exists := groupMap["rule_string"]; exists {
				if ruleStr, ok := ruleString.(string); ok {
					filterGroup.RuleString = ruleStr
				}
			}
			
			if mediaType, exists := groupMap["media_type"]; exists {
				if mediaTypeStr, ok := mediaType.(string); ok {
					filterGroup.MediaType = mediaTypeStr
				}
			}
			
			if category, exists := groupMap["category"]; exists {
				if categoryStr, ok := category.(string); ok {
					filterGroup.Category = categoryStr
				}
			}
			
			result = append(result, filterGroup)
		}
	}
	
	return result
}

// GetRuleGroup 获取规则�?func (r *RuleHelper) GetRuleGroup(groupName string) *models.FilterRuleGroup {
	ruleGroups := r.GetRuleGroups()
	for _, group := range ruleGroups {
		// 注意：这里需要使用局部变量避免闭包问�?		g := group
		if g.Name == groupName {
			return &g
		}
	}
	return nil
}

// GetRuleGroupByMedia 根据媒体信息获取规则�?func (r *RuleHelper) GetRuleGroupByMedia(media *models.MediaInfo, groupNames []string) []models.FilterRuleGroup {
	retGroups := make([]models.FilterRuleGroup, 0)
	ruleGroups := r.GetRuleGroups()
	
	// 如果指定了groupNames，则过滤规则�?	if len(groupNames) > 0 {
		filteredGroups := make([]models.FilterRuleGroup, 0)
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
	
	// 根据媒体信息筛选规则组
	for _, group := range ruleGroups {
		// 如果没有指定媒体类型，则添加该组
		if group.MediaType == "" {
			retGroups = append(retGroups, group)
		} else if media != nil && group.Category == "" && group.MediaType == string(media.Type) {
			// 如果媒体类型匹配但没有指定分类，则添加该�?			retGroups = append(retGroups, group)
		} else if media != nil && group.Category != "" && media.Genres != nil {
			// 检查媒体分类是否匹�?			for _, category := range media.Genres {
				if group.Category == category {
					retGroups = append(retGroups, group)
					break
				}
			}
		}
	}
	
	return retGroups
}

// GetCustomRules 获取用户所有自定义规则
func (r *RuleHelper) GetCustomRules() []models.CustomRule {
	// 从系统配置中获取自定义规�?	rules := db.NewSystemConfigOper().Get(models.SystemConfigKeyCustomFilterRules)
	if rules == nil {
		return []models.CustomRule{}
	}
	
	// 类型断言获取规则数据
	rulesList, ok := rules.([]interface{})
	if !ok {
		return []models.CustomRule{}
	}
	
	// 转换为CustomRule数组
	result := make([]models.CustomRule, 0, len(rulesList))
	for _, rule := range rulesList {
		// 将map转换为CustomRule结构�?		if ruleMap, ok := rule.(map[string]interface{}); ok {
			customRule := models.CustomRule{}
			
			if id, exists := ruleMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					customRule.ID = idStr
				}
			}
			
			if name, exists := ruleMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					customRule.Name = nameStr
				}
			}
			
			if include, exists := ruleMap["include"]; exists {
				if includeStr, ok := include.(string); ok {
					customRule.Include = includeStr
				}
			}
			
			if exclude, exists := ruleMap["exclude"]; exists {
				if excludeStr, ok := exclude.(string); ok {
					customRule.Exclude = excludeStr
				}
			}
			
			if sizeRange, exists := ruleMap["size_range"]; exists {
				if sizeRangeStr, ok := sizeRange.(string); ok {
					customRule.SizeRange = sizeRangeStr
				}
			}
			
			if seeders, exists := ruleMap["seeders"]; exists {
				if seedersStr, ok := seeders.(string); ok {
					customRule.Seeders = seedersStr
				}
			}
			
			if publishTime, exists := ruleMap["publish_time"]; exists {
				if publishTimeStr, ok := publishTime.(string); ok {
					customRule.PublishTime = publishTimeStr
				}
			}
			
			result = append(result, customRule)
		}
	}
	
	return result
}

// GetCustomRule 获取自定义规�?func (r *RuleHelper) GetCustomRule(ruleID string) *models.CustomRule {
	rules := r.GetCustomRules()
	for _, rule := range rules {
		// 注意：这里需要使用局部变量避免闭包问�?		r := rule
		if r.ID == ruleID {
			return &r
		}
	}
	return nil
}
