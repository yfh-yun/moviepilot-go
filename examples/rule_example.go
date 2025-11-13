package main

import (
	"fmt"
	
	"moviepilot-go/internal/helper"
	"moviepilot-go/pkg/models"
)

func main() {
	fmt.Println("Rule Helper Example")
	
	// 创建规则帮助类实�?	ruleHelper := helper.NewRuleHelper()
	
	if ruleHelper == nil {
		fmt.Println("Failed to create RuleHelper")
		return
	}
	
	fmt.Println("RuleHelper created successfully")
	
	// 获取所有规则组
	fmt.Println("\n=== 获取规则�?===")
	ruleGroups := ruleHelper.GetRuleGroups()
	fmt.Printf("找到 %d 个规则组\n", len(ruleGroups))
	
	// 显示规则组信�?	for i, group := range ruleGroups {
		fmt.Printf("  %d. 名称: %s, 媒体类型: %s, 分类: %s\n", 
			i+1, group.Name, group.MediaType, group.Category)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取特定规则�?	fmt.Println("\n=== 获取特定规则�?===")
	group := ruleHelper.GetRuleGroup("non-existent-group")
	if group == nil {
		fmt.Println("未找到指定规则组")
	} else {
		fmt.Printf("找到规则�? %s\n", group.Name)
	}
	
	// 根据媒体信息获取规则�?	fmt.Println("\n=== 根据媒体信息获取规则�?===")
	// 创建一个媒体信息示�?	media := &models.MediaInfo{
		Title: "Test Movie",
		Type:  models.Movie,
		Genres: []string{"Action", "Adventure"},
		Year:  2023,
	}
	
	// 获取匹配的规则组
	mediaGroups := ruleHelper.GetRuleGroupByMedia(media, nil)
	fmt.Printf("根据媒体信息找到 %d 个匹配的规则组\n", len(mediaGroups))
	
	// 指定特定规则组名称获�?	specificGroups := ruleHelper.GetRuleGroupByMedia(media, []string{"test-group"})
	fmt.Printf("根据指定名称找到 %d 个规则组\n", len(specificGroups))
	
	// 获取自定义规�?	fmt.Println("\n=== 获取自定义规�?===")
	customRules := ruleHelper.GetCustomRules()
	fmt.Printf("找到 %d 个自定义规则\n", len(customRules))
	
	// 显示自定义规则信�?	for i, rule := range customRules {
		fmt.Printf("  %d. 名称: %s, ID: %s\n", i+1, rule.Name, rule.ID)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取特定自定义规�?	fmt.Println("\n=== 获取特定自定义规�?===")
	customRule := ruleHelper.GetCustomRule("non-existent-rule")
	if customRule == nil {
		fmt.Println("未找到指定自定义规则")
	} else {
		fmt.Printf("找到自定义规�? %s\n", customRule.Name)
	}
	
	fmt.Println("\nExample completed")
}
