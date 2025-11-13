package parser

import (
	"regexp"
	"strings"

	"moviepilot-go/internal/modules/indexer"
)

// NexusProjectSiteUserInfo NexusProject站点用户信息解析�?type NexusProjectSiteUserInfo struct {
	*indexer.NexusPhpSiteUserInfo
}

// NewNexusProjectSiteUserInfo 创建NexusProject站点用户信息解析器实�?func NewNexusProjectSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *NexusProjectSiteUserInfo {

	parser := &NexusProjectSiteUserInfo{
		NexusPhpSiteUserInfo: indexer.NewNexusPhpSiteUserInfo(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面
func (n *NexusProjectSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = n.prepareHTMLText(htmlText)

	// 查找用户详情页面链接
	userDetailMatch := regexp.MustCompile(`userdetails.php\?id=(\d+)`).FindStringSubmatch(htmlText)
	if len(userDetailMatch) > 1 && strings.TrimSpace(userDetailMatch[0]) != "" {
		n.userDetailPage = strings.TrimSpace(userDetailMatch[0])
		n.userid = userDetailMatch[1]
	}

	// 设置做种页面
	n.torrentSeedingPage = "viewusertorrents.php?id=" + n.userid + "&show=seeding"
}
