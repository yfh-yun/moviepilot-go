package utils

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// 挑战页面标题列表
var challengeTitles = []string{
	// Cloudflare
	"Just a moment...",
	"请稍候…",
	// DDoS-GUARD
	"DDOS-GUARD",
}

// 挑战页面CSS选择器列表
var challengeSelectors = []string{
	// Cloudflare
	"#cf-challenge-running", ".ray_id", ".attack-box", "#cf-please-wait", "#challenge-spinner", "#trk_jschal_js",
	// Custom CloudFlare for EbookParadijs, Film-Paleis, MuziekFabriek and Puur-Hollands
	"td.info #js_info",
	// Fairlane / pararius.com
	"div.vc div.text-box h2",
}

// UnderChallenge 检查页面是否处于挑战状态
func UnderChallenge(htmlText string) bool {
	logger := logger.GetLogger()
	logger.Debug("检查页面是否处于挑战状态")

	if htmlText == "" {
		return false
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		logger.Error("解析HTML失败", zap.Error(err))
		return false
	}

	// 检查页面标题
	pageTitle := doc.Find("title").Text()
	logger.Debug("页面标题", zap.String("title", pageTitle))

	for _, title := range challengeTitles {
		if strings.EqualFold(pageTitle, title) {
			logger.Debug("匹配到挑战页面标题", zap.String("title", title))
			return true
		}
	}

	// 检查CSS选择器
	for _, selector := range challengeSelectors {
		if doc.Find(selector).Length() > 0 {
			logger.Debug("匹配到挑战页面选择器", zap.String("selector", selector))
			return true
		}
	}

	return false
}
