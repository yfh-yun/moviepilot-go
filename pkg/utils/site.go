package utils

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// IsLoggedIn 对应 Python SiteUtils.is_logged_in：
// - HTML 无效或包含明显密码输入框 => 未登录
// - 存在登出/用户面板等典型元素 => 已登录
func IsLoggedIn(htmlText string) bool {
	if strings.TrimSpace(htmlText) == "" {
		return false
	}
	reader := strings.NewReader(htmlText)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return false
	}

	// 若存在密码输入框，则认为未登录
	if doc.Find("input[type='password']").Length() > 0 {
		return false
	}

	// 参考 Python 中的若干 XPath，用 CSS/属性选择器近似实现
	selectors := []string{
		"a[href*='logout']",
		"a[data-url*='logout']",
		"a[href*='mybonus']",
		"a[onclick*='logout']",
		"a[href*='usercp']",
		"a[lay-on*='logout']",
		"form[action*='logout']",
		"div.user-info-side",
		"a#myitem",
	}

	for _, sel := range selectors {
		if doc.Find(sel).Length() > 0 {
			return true
		}
	}

	return false
}

// IsCheckin 对应 Python SiteUtils.is_checkin：
// - 若能找到“签到入口”相关元素 => 认为尚未签到（返回 false）
// - 否则认为已签到（返回 true）
func IsCheckin(htmlText string) bool {
	if strings.TrimSpace(htmlText) == "" {
		return false
	}
	reader := strings.NewReader(htmlText)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return false
	}

	selectors := []string{
		"a#signed",
		"a[href*='attendance']",
		"a:contains('签到')",
		"a b:contains('签 到')",
		"span#sign_in a",
		"a[href*='addbonus']",
		"input.dt_button[value*='打卡']",
		"a[href*='sign_in']",
		"a[onclick*='do_signin']",
		"a#do-attendance",
		"shark-icon-button[href='attendance.php']",
	}

	for _, sel := range selectors {
		if doc.Find(sel).Length() > 0 {
			return false
		}
	}

	return true
}
