package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// CloudflareHelper Cloudflare挑战检测助手
type CloudflareHelper struct {
	challengeTitles    []string
	challengeSelectors []string
}

// NewCloudflareHelper 创建Cloudflare助手实例
func NewCloudflareHelper() *CloudflareHelper {
	return &CloudflareHelper{
		challengeTitles: []string{
			"Just a moment...",
			"请稍候…",
			"DDOS-GUARD",
		},
		challengeSelectors: []string{
			"#cf-challenge-running",
			".ray_id",
			".attack-box",
			"#cf-please-wait",
			"#challenge-spinner",
			"#trk_jschal_js",
			"td.info #js_info",
			"div.vc div.text-box h2",
		},
	}
}

// UnderChallenge 检查页面是否处于挑战状态
func (cf *CloudflareHelper) UnderChallenge(htmlText string) bool {
	if htmlText == "" {
		return false
	}

	// 提取页面标题
	title := cf.extractTitle(htmlText)
	if title == "" {
		return false
	}

	// 检查标题匹配
	for _, challengeTitle := range cf.challengeTitles {
		if strings.ToLower(title) == strings.ToLower(challengeTitle) {
			return true
		}
	}

	// 检查选择器匹配
	for _, selector := range cf.challengeSelectors {
		if cf.matchesSelector(htmlText, selector) {
			return true
		}
	}

	return false
}

// extractTitle 从HTML中提取标题
func (cf *CloudflareHelper) extractTitle(htmlText string) string {
	// 简单的正则表达式提取标题
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(htmlText)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// matchesSelector 检查HTML是否匹配选择器
func (cf *CloudflareHelper) matchesSelector(htmlText, selector string) bool {
	// 简化的选择器匹配，只支持ID和类选择器
	if strings.HasPrefix(selector, "#") {
		// ID选择器
		id := strings.TrimPrefix(selector, "#")
		idRegex := regexp.MustCompile(fmt.Sprintf(`(?i)id\s*=\s*["']?%s["']?`, regexp.QuoteMeta(id)))
		return idRegex.MatchString(htmlText)
	} else if strings.HasPrefix(selector, ".") {
		// 类选择器
		class := strings.TrimPrefix(selector, ".")
		classRegex := regexp.MustCompile(fmt.Sprintf(`(?i)class\s*=\s*["'][^"']*%s[^"']*["']`, regexp.QuoteMeta(class)))
		return classRegex.MatchString(htmlText)
	} else if strings.Contains(selector, " ") {
		// 后代选择器（简化处理）
		parts := strings.Split(selector, " ")
		if len(parts) == 2 {
			parent := parts[0]
			child := parts[1]
			parentRegex := cf.selectorToRegex(parent)
			childRegex := cf.selectorToRegex(child)
			return parentRegex.MatchString(htmlText) && childRegex.MatchString(htmlText)
		}
	}
	return false
}

// selectorToRegex 将CSS选择器转换为正则表达式
func (cf *CloudflareHelper) selectorToRegex(selector string) *regexp.Regexp {
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		pattern := fmt.Sprintf(`(?i)id\s*=\s*["']?%s["']?`, regexp.QuoteMeta(id))
		return regexp.MustCompile(pattern)
	} else if strings.HasPrefix(selector, ".") {
		class := strings.TrimPrefix(selector, ".")
		pattern := fmt.Sprintf(`(?i)class\s*=\s*["'][^"']*%s[^"']*["']`, regexp.QuoteMeta(class))
		return regexp.MustCompile(pattern)
	}
	// 标签选择器
	pattern := fmt.Sprintf(`(?i)<%s[^>]*>`, regexp.QuoteMeta(selector))
	return regexp.MustCompile(pattern)
}

// IsCloudflareError 判断是否为Cloudflare相关错误
func (cf *CloudflareHelper) IsCloudflareError(errorMsg string) bool {
	cfErrors := []string{
		"cloudflare",
		"cf-ray",
		"challenge",
		"just a moment",
		"captcha",
		"security check",
	}
	
	lowerError := strings.ToLower(errorMsg)
	for _, cfError := range cfErrors {
		if strings.Contains(lowerError, cfError) {
			return true
		}
	}
	return false
}

// GetRetryDelay 获取重试延迟时间（秒）
func (cf *CloudflareHelper) GetRetryDelay() int {
	return 6 // 默认6秒重试
}

// GetTimeout 获取超时时间（秒）
func (cf *CloudflareHelper) GetTimeout() int {
	return 60 // 默认60秒超时
}