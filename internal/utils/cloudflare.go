package utils

import (
	"os"
	"strconv"
	"strings"
	
	"github.com/PuerkitoBio/goquery"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// Cloudflare工具相关常量
var (
	// 挑战页面标题
	CHALLENGE_TITLES = []string{
		// Cloudflare
		"Just a moment...",
		"请稍候�?,
		// DDoS-GUARD
		"DDOS-GUARD",
	}
	
	// 挑战页面选择�?	CHALLENGE_SELECTORS = []string{
		// Cloudflare
		"#cf-challenge-running", 
		".ray_id", 
		".attack-box", 
		"#cf-please-wait", 
		"#challenge-spinner", 
		"#trk_jschal_js",
		// Custom CloudFlare for EbookParadijs, Film-Paleis, MuziekFabriek and Puur-Hollands
		"td.info #js_info",
		// Fairlane / pararius.com
		"div.vc div.text-box h2",
	}
	
	// 短超时时�?	SHORT_TIMEOUT = 6
	
	// Cloudflare超时时间
	CF_TIMEOUT int
)

func init() {
	// 初始化CF_TIMEOUT，从环境变量获取或使用默认�?0
	CF_TIMEOUT = 60
	if timeoutStr := os.Getenv("NASTOOL_CF_TIMEOUT"); timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil && val > 0 {
			CF_TIMEOUT = val
		}
	}
}

// UnderChallenge 检查页面是否处于挑战状�?// :param htmlText: HTML文本内容
// :return: 如果页面处于挑战状态返回true，否则返回false
func UnderChallenge(htmlText string) bool {
	// 检查htmlText是否为空
	if htmlText == "" {
		return false
	}
	
	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		logger.GetLoggerManager().Debug("解析HTML文档失败", zap.Error(err))
		return false
	}
	
	// 获取页面标题
	pageTitle := doc.Find("title").Text()
	logger.GetLoggerManager().Debug("检查挑战页面标�? " + pageTitle)
	
	// 检查标题是否匹配挑战标�?	for _, title := range CHALLENGE_TITLES {
		if strings.ToLower(pageTitle) == strings.ToLower(title) {
			return true
		}
	}
	
	// 检查选择器是否匹�?	for _, selector := range CHALLENGE_SELECTORS {
		if doc.Find(selector).Length() > 0 {
			return true
		}
	}
	
	return false
}
