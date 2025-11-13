package utils

import (
	"strings"

	"github.com/antchfx/htmlquery"
)

// SiteUtils 站点工具�?type SiteUtils struct{}

// NewSiteUtils 创建新的站点工具类实�?func NewSiteUtils() *SiteUtils {
	return &SiteUtils{}
}

// IsLoggedIn 判断站点是否已经登陆
// htmlText: HTML文本内容
// 返回�? 已登录返回true，未登录返回false
func (s *SiteUtils) IsLoggedIn(htmlText string) bool {
	// 如果HTML文本为空，返回false
	if strings.TrimSpace(htmlText) == "" {
		return false
	}

	// 解析HTML文档
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		return false
	}

	// 检查是否为有效的HTML元素
	if doc == nil {
		return false
	}

	// 存在明显的密码输入框，说明未登录
	passwordInputs, err := htmlquery.QueryAll(doc, "//input[@type='password']")
	if err == nil && len(passwordInputs) > 0 {
		return false
	}

	// 是否存在登出和用户面板等链接
	xpaths := []string{
		`//a[contains(@href, "logout") or contains(@data-url, "logout") or contains(@href, "mybonus") or contains(@onclick, "logout") or contains(@href, "usercp") or contains(@lay-on, "logout")]`,
		`//form[contains(@action, "logout")]`,
		`//div[@class="user-info-side"]`,
		`//a[@id="myitem"]`,
	}

	// 检查是否存在任何登出或用户相关的元�?	for _, xpath := range xpaths {
		nodes, err := htmlquery.QueryAll(doc, xpath)
		if err == nil && len(nodes) > 0 {
			return true
		}
	}

	return false
}

// IsCheckin 判断站点是否已经签到
// htmlText: HTML文本内容
// 返回�? 已签到返回true，未签到返回false
func (s *SiteUtils) IsCheckin(htmlText string) bool {
	// 如果HTML文本为空，返回false
	if strings.TrimSpace(htmlText) == "" {
		return false
	}

	// 解析HTML文档
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		return false
	}

	// 检查是否为有效的HTML元素
	if doc == nil {
		return false
	}

	// 站点签到支持的识别XPATH
	xpaths := []string{
		`//a[@id="signed"]`,
		`//a[contains(@href, "attendance")]`,
		`//a[contains(text(), "签到")]`,
		`//a/b[contains(text(), "�?�?)]`,
		`//span[@id="sign_in"]/a`,
		`//a[contains(@href, "addbonus")]`,
		`//input[@class="dt_button"][contains(@value, "打卡")]`,
		`//a[contains(@href, "sign_in")]`,
		`//a[contains(@onclick, "do_signin")]`,
		`//a[@id="do-attendance"]`,
		`//shark-icon-button[@href="attendance.php"]`,
	}

	// 检查是否存在任何签到相关的元素
	// 如果存在这些元素，说明还未签到，返回false
	for _, xpath := range xpaths {
		nodes, err := htmlquery.QueryAll(doc, xpath)
		if err == nil && len(nodes) > 0 {
			return false
		}
	}

	// 如果没有找到未签到的元素，说明已经签�?	return true
}
