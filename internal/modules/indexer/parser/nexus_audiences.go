package parser

import (
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// NexusAudiencesSiteUserInfo NexusAudiences站点用户信息解析�?type NexusAudiencesSiteUserInfo struct {
	*indexer.NexusPhpSiteUserInfo
}

// NewNexusAudiencesSiteUserInfo 创建NexusAudiences站点用户信息解析器实�?func NewNexusAudiencesSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *NexusAudiencesSiteUserInfo {

	parser := &NexusAudiencesSiteUserInfo{
		NexusPhpSiteUserInfo: indexer.NewNexusPhpSiteUserInfo(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSeedingPages 解析做种页面信息，覆盖基类方�?func (n *NexusAudiencesSiteUserInfo) parseSeedingPages() {
	if n.torrentSeedingPage == "" {
		return
	}
	
	// 设置请求�?	n.torrentSeedingHeaders = map[string]string{
		"Referer": n.joinURL(n.baseURL, n.userDetailPage),
	}
	
	// 获取页面内容
	htmlText := n.getPageContent(
		n.joinURL(n.baseURL, n.torrentSeedingPage),
		n.torrentSeedingParams,
		n.torrentSeedingHeaders,
	)
	
	if htmlText == "" {
		return
	}
	
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 查找Total�?	totalRows := htmlquery.Find(doc, `//table[@class="table table-bordered"]//tr[td[1][normalize-space()="Total"]]`)
	if len(totalRows) == 0 {
		return
	}

	// 提取做种数和做种大小
	seedingCountNodes := htmlquery.Find(totalRows[0], `./td[2]/text()`)
	seedingSizeNodes := htmlquery.Find(totalRows[0], `./td[3]/text()`)

	if len(seedingCountNodes) > 0 {
		seedingCountText := strings.TrimSpace(htmlquery.InnerText(seedingCountNodes[0]))
		n.seeding = stringUtils.StrInt(seedingCountText)
	}

	if len(seedingSizeNodes) > 0 {
		seedingSizeText := strings.TrimSpace(htmlquery.InnerText(seedingSizeNodes[0]))
		n.seedingSize = stringUtils.NumFilesize(seedingSizeText)
	}
}
