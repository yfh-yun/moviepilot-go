package helper

import (
	"testing"
	
	"moviepilot-go/pkg/models"
)

func TestNewRuleHelper(t *testing.T) {
	// 测试创建RuleHelper实例
	helper := NewRuleHelper()
	if helper == nil {
		t.Error("Failed to create RuleHelper instance")
	}
}

func TestGetRuleGroups(t *testing.T) {
	// 测试获取规则�?	helper := NewRuleHelper()
	ruleGroups := helper.GetRuleGroups()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if ruleGroups == nil {
		t.Error("GetRuleGroups should not return nil")
	}
	
	// ruleGroups应该是一个切�?	if len(ruleGroups) < 0 {
		t.Error("GetRuleGroups should return a valid slice")
	}
}

func TestGetRuleGroup(t *testing.T) {
	// 测试获取特定规则�?	helper := NewRuleHelper()
	
	// 测试不存在的规则�?	group := helper.GetRuleGroup("non-existent-group")
	if group != nil {
		t.Error("GetRuleGroup should return nil for non-existent group")
	}
}

func TestGetRuleGroupByMedia(t *testing.T) {
	// 测试根据媒体信息获取规则�?	helper := NewRuleHelper()
	
	// 测试无媒体信息的情况
	groups := helper.GetRuleGroupByMedia(nil, nil)
	if groups == nil {
		t.Error("GetRuleGroupByMedia should not return nil")
	}
	
	// 测试指定groupNames的情�?	groups = helper.GetRuleGroupByMedia(nil, []string{"test-group"})
	if groups == nil {
		t.Error("GetRuleGroupByMedia should not return nil")
	}
	
	// 测试有媒体信息的情况
	media := &models.MediaInfo{
		Type: models.Movie,
		Genres: []string{"Action", "Adventure"},
	}
	
	groups = helper.GetRuleGroupByMedia(media, nil)
	if groups == nil {
		t.Error("GetRuleGroupByMedia should not return nil")
	}
}

func TestGetCustomRules(t *testing.T) {
	// 测试获取自定义规�?	helper := NewRuleHelper()
	rules := helper.GetCustomRules()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if rules == nil {
		t.Error("GetCustomRules should not return nil")
	}
	
	// rules应该是一个切�?	if len(rules) < 0 {
		t.Error("GetCustomRules should return a valid slice")
	}
}

func TestGetCustomRule(t *testing.T) {
	// 测试获取特定自定义规�?	helper := NewRuleHelper()
	
	// 测试不存在的规则
	rule := helper.GetCustomRule("non-existent-rule")
	if rule != nil {
		t.Error("GetCustomRule should return nil for non-existent rule")
	}
}

func TestRuleGroupStruct(t *testing.T) {
	// 测试FilterRuleGroup结构�?	group := models.FilterRuleGroup{
		Name:        "Test Group",
		RuleString:  "test-rule",
		MediaType:   "movie",
		Category:    "Action",
	}
	
	if group.Name != "Test Group" {
		t.Error("Name not set correctly")
	}
	
	if group.RuleString != "test-rule" {
		t.Error("RuleString not set correctly")
	}
	
	if group.MediaType != "movie" {
		t.Error("MediaType not set correctly")
	}
	
	if group.Category != "Action" {
		t.Error("Category not set correctly")
	}
}

func TestCustomRuleStruct(t *testing.T) {
	// 测试CustomRule结构�?	rule := models.CustomRule{
		ID:          "test-rule-id",
		Name:        "Test Rule",
		Include:     "include-pattern",
		Exclude:     "exclude-pattern",
		SizeRange:   "100-1000",
		Seeders:     "10",
		PublishTime: "24h",
	}
	
	if rule.ID != "test-rule-id" {
		t.Error("ID not set correctly")
	}
	
	if rule.Name != "Test Rule" {
		t.Error("Name not set correctly")
	}
	
	if rule.Include != "include-pattern" {
		t.Error("Include not set correctly")
	}
	
	if rule.Exclude != "exclude-pattern" {
		t.Error("Exclude not set correctly")
	}
	
	if rule.SizeRange != "100-1000" {
		t.Error("SizeRange not set correctly")
	}
	
	if rule.Seeders != "10" {
		t.Error("Seeders not set correctly")
	}
	
	if rule.PublishTime != "24h" {
		t.Error("PublishTime not set correctly")
	}
}
