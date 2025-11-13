package utils

import (
	"regexp"
	"strings"
)

// SiteUtils 站点工具�?type SiteUtils struct{}

// NewSiteUtils 创建站点工具类实�?func NewSiteUtils() *SiteUtils {
	return &SiteUtils{}
}

// IsLoggedIn 判断站点是否已经登陆
func (su *SiteUtils) IsLoggedIn(htmlText string) bool {
	// 简化实�?- 检查是否存在明显的登出链接或用户相关信�?	logoutPatterns := []string{
		`logout`,
		`登出`,
		`退出`,
		`signout`,
		`usercp`,
		`mybonus`,
	}
	
	for _, pattern := range logoutPatterns {
		if strings.Contains(strings.ToLower(htmlText), pattern) {
			return true
		}
	}
	
	// 检查是否存在密码输入框（如果存在可能未登录�?	if strings.Contains(htmlText, `type="password"`) {
		return false
	}
	
	// 默认实现返回true
	return true
}

// IsCheckin 判断站点是否已经签到
func (su *SiteUtils) IsCheckin(htmlText string) bool {
	// 简化实�?- 检查是否存在签到相关关键词
	checkinPatterns := []string{
		`签到`,
		`打卡`,
		`attendance`,
		`sign_in`,
		`do_signin`,
	}
	
	for _, pattern := range checkinPatterns {
		if strings.Contains(strings.ToLower(htmlText), strings.ToLower(pattern)) {
			// 如果存在签到按钮，说明未签到
			return false
		}
	}
	
	// 默认实现返回true（已签到�?	return true
}

// IsValidHTMLElement 检查HTML元素是否有效
func (su *SiteUtils) IsValidHTMLElement(elem string) bool {
	return elem != "" && len(elem) > 0
}
