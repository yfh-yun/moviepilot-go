package parser

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// IptSiteUserInfo IPT站点用户信息解析�?type IptSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewIptSiteUserInfo 创建IPT站点用户信息解析器实�?func NewIptSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *IptSiteUserInfo {

	parser := &IptSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)
	
	// 设置特定页面URL
	parser.SiteParserBaseImpl.UserDetailPage = "user.php?u="
	parser.SiteParserBaseImpl.TorrentSeedingPage = "peers?u="

	return parser
}

// parseUserBaseInfo 解析用户基础信息
func (i *IptSiteUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = i.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		i.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	// 查找用户信息链接
	userNodes := htmlquery.Find(doc, `//a[contains(@href, "/u/")]`)
	if len(userNodes) > 0 {
		// 获取用户�?		i.username = strings.TrimSpace(htmlquery.InnerText(userNodes[len(userNodes)-1]))

		// 获取用户ID
		href := htmlquery.SelectAttr(userNodes[0], "href")
		userIDMatch := regexp.MustCompile(`/u/(\d+)`).FindStringSubmatch(href)
		if len(userIDMatch) > 1 && strings.TrimSpace(userIDMatch[1]) != "" {
			i.userid = userIDMatch[1]
			i.userDetailPage = "user.php?u=" + i.userid
			i.torrentSeedingPage = "peers?u=" + i.userid
		}
	}

	// 解析上传、下载等统计信息
	statsNodes := htmlquery.Find(doc, `//div[@class = "stats"]/div/div`)
	if len(statsNodes) > 0 {
		stringUtils := utils.NewStringUtils()
		
		// 上传�?(第二个span)
		uploadNodes := htmlquery.Find(statsNodes[0], `span[position()=2]/text()`)
		if len(uploadNodes) > 0 {
			uploadText := strings.TrimSpace(htmlquery.InnerText(uploadNodes[0]))
			i.upload = stringUtils.NumFilesize(uploadText)
		}

		// 下载�?(第三个span)
		downloadNodes := htmlquery.Find(statsNodes[0], `span[position()=3]/text()`)
		if len(downloadNodes) > 0 {
			downloadText := strings.TrimSpace(htmlquery.InnerText(downloadNodes[0]))
			i.download = stringUtils.NumFilesize(downloadText)
		}

		// 做种数和下载�?(第三个a标签下的两个text节点)
		seedingLeechingNodes := htmlquery.Find(statsNodes[0], `a[position()=3]/text()`)
		if len(seedingLeechingNodes) >= 2 {
			i.seeding = stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingLeechingNodes[0])))
			i.leeching = stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingLeechingNodes[1])))
		}

		// 分享�?(第一个span)
		ratioNodes := htmlquery.Find(statsNodes[0], `span[position()=1]/text()`)
		if len(ratioNodes) > 0 {
			ratioText := strings.TrimSpace(htmlquery.InnerText(ratioNodes[0]))
			i.ratio = stringUtils.StrFloat(strings.ReplaceAll(ratioText, "-", "0"))
		}

		// 积分 (第四个a标签)
		bonusNodes := htmlquery.Find(statsNodes[0], `a[position()=4]/text()`)
		if len(bonusNodes) > 0 {
			bonusText := strings.TrimSpace(htmlquery.InnerText(bonusNodes[0]))
			i.bonus = stringUtils.StrFloat(bonusText)
		}
	}
}

// parseSitePage 解析站点相关信息页面
func (i *IptSiteUserInfo) parseSitePage(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (i *IptSiteUserInfo) parseUserDetailInfo(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		i.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 用户等级
	userLevelNodes := htmlquery.Find(doc, `//tr/th[text()="Class"]/following-sibling::td[1]/text()`)
	if len(userLevelNodes) > 0 {
		i.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelNodes[0]))
	}

	// 加入日期
	joinDateNodes := htmlquery.Find(doc, `//tr/th[text()="Join date"]/following-sibling::td[1]/text()`)
	if len(joinDateNodes) > 0 {
		joinDateText := strings.TrimSpace(htmlquery.InnerText(joinDateNodes[0]))
		// 只取括号前的部分
		if idx := strings.Index(joinDateText, " ("); idx != -1 {
			joinDateText = joinDateText[:idx]
		}
		i.joinAt = stringUtils.UnifyDateTimeStr(joinDateText)
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (i *IptSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		i.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	// 计算做种结束位置
	seedingEndPos := 3
	leechersNodes := htmlquery.Find(doc, `//tr/td[text() = "Leechers"]`)
	if len(leechersNodes) > 0 {
		// 计算"Leechers"行前面有多少�?		precedingNodes := htmlquery.Find(doc, `//tr/td[text() = "Leechers"]/../preceding-sibling::tr`)
		seedingEndPos = len(precedingNodes) + 1 - 3
	}

	pageSeeding := 0
	pageSeedingSize := 0

	// 获取做种大小信息
	seedingTorrents := htmlquery.Find(doc, `//tr/td[text() = "Seeders"]/../following-sibling::tr/td[position()=6]/text()`)
	if len(seedingTorrents) > 0 {
		pageSeeding = seedingEndPos
		for idx, seedingTorrent := range seedingTorrents {
			if idx >= seedingEndPos {
				break
			}
			
			perSize := strings.TrimSpace(htmlquery.InnerText(seedingTorrent))
			
			// 提取括号内的内容
			if strings.Contains(perSize, "(") && strings.Contains(perSize, ")") {
				start := strings.LastIndex(perSize, "(") + 1
				end := strings.Index(perSize, ")")
				if start < end {
					perSize = perSize[start:end]
				}
			}

			pageSeedingSize += stringUtils.NumFilesize(perSize)
		}
	}

	i.seeding = pageSeeding
	i.seedingSize = pageSeedingSize

	return ""
}

// parseUserTrafficInfo 解析用户流量信息
func (i *IptSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseMessageUnreadLinks 解析未读消息链接
func (i *IptSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (i *IptSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
