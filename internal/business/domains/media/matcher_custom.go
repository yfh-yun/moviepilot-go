package media

import (
	"strings"
)

// CustomizationConfigProvider 自定义占位符配置提供者接口
type CustomizationConfigProvider interface {
	GetCustomization() ([]string, error)
}

// CustomizationMatcher 自定义占位符匹配器
type CustomizationMatcher struct {
	customProvider CustomizationConfigProvider
}

// NewCustomizationMatcher 创建新的CustomizationMatcher实例
func NewCustomizationMatcher(provider CustomizationConfigProvider) *CustomizationMatcher {
	return &CustomizationMatcher{
		customProvider: provider,
	}
}

// Match 匹配自定义占位符
func (cm *CustomizationMatcher) Match(title string) string {
	if title == "" {
		return ""
	}

	// 获取自定义占位符
	var customizations []string
	if cm.customProvider != nil {
		customizations, _ = cm.customProvider.GetCustomization()
	}

	if len(customizations) == 0 {
		return ""
	}

	var matches []string
	processedTitle := title

	// 匹配所有自定义占位符
	for _, customization := range customizations {
		customization = strings.TrimSpace(customization)
		if customization == "" {
			continue
		}

		if strings.Contains(processedTitle, customization) {
			matches = append(matches, customization)
			// 从标题中移除匹配到的占位符
			processedTitle = strings.ReplaceAll(processedTitle, customization, "")
		}
	}

	// 去重并返回
	uniqueMatches := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			uniqueMatches = append(uniqueMatches, match)
		}
	}

	return strings.Join(uniqueMatches, "@")
}
