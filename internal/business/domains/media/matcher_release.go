package media

import (
	"regexp"
	"strings"
)

// ReleaseGroupConfigProvider 制作组配置提供者接口
type ReleaseGroupConfigProvider interface {
	GetCustomReleaseGroups() ([]string, error)
}

// ReleaseGroupsMatcher 制作组匹配器
type ReleaseGroupsMatcher struct {
	builtInPatterns []string
	customProvider  ReleaseGroupConfigProvider
}

// builtInReleaseGroups 内置制作组正则
var builtInReleaseGroups = []string{
	`\[([^\]]+)\]`,     // [Group]
	`\(([^\)]+)\)`,     // (Group)
	`-\s*([^-]+)$`,     // - Group
	`\.(\w+)$`,         // .Group
	`\[([^\]]+)\]\s*$`, // [Group] at end
	`\(([^\)]+)\)\s*$`, // (Group) at end
}

// NewReleaseGroupsMatcher 创建新的ReleaseGroupsMatcher实例
func NewReleaseGroupsMatcher(provider ReleaseGroupConfigProvider) *ReleaseGroupsMatcher {
	return &ReleaseGroupsMatcher{
		builtInPatterns: builtInReleaseGroups,
		customProvider:  provider,
	}
}

// Match 匹配制作组
func (rgm *ReleaseGroupsMatcher) Match(title string, groupsPattern string) string {
	if title == "" {
		return ""
	}

	// 合并内置和自定义模式
	patterns := rgm.builtInPatterns
	if rgm.customProvider != nil {
		customPatterns, err := rgm.customProvider.GetCustomReleaseGroups()
		if err == nil {
			patterns = append(patterns, customPatterns...)
		}
	}

	var groups []string
	processedTitle := title

	// 应用所有模式
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(processedTitle, -1)
		for _, match := range matches {
			if len(match) > 1 {
				group := strings.TrimSpace(match[1])
				if group != "" {
					groups = append(groups, group)
					// 从标题中移除匹配到的制作组
					processedTitle = re.ReplaceAllString(processedTitle, "")
				}
			}
		}
	}

	// 去重并返回
	uniqueGroups := make([]string, 0, len(groups))
	seen := make(map[string]bool)
	for _, group := range groups {
		if !seen[group] {
			seen[group] = true
			uniqueGroups = append(uniqueGroups, group)
		}
	}

	return strings.Join(uniqueGroups, "@")
}
