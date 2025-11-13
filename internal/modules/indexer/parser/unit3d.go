package parser

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// Unit3dSiteUserInfo Unit3d站点用户信息解析�?type Unit3dSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewUnit3dSiteUserInfo 创建Unit3d站点用户信息解析器实�?func NewUnit3dSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *Unit3dSiteUserInfo {

	parser := &Unit3dSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseUserBaseInfo 解析用户基础信息
func (u *Unit3dSiteUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = u.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		u.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	// 查找用户设置链接
	tmps := htmlquery.Find(doc, `//a[contains(@href, "/users/") and contains(@href, "settings")]/@href`)
	if len(tmps) > 0 {
		href := htmlquery.SelectAttr(tmps[0], "href")
		userNameMatch := regexp.MustCompile(`/users/(.+)/settings`).FindStringSubmatch(href)
		if len(userNameMatch) > 1 && strings.TrimSpace(userNameMatch[0]) != "" {
			u.username = userNameMatch[1]
			u.torrentSeedingPage = "/users/" + u.username + "/active?perPage=100&client=&seeding=include"
			u.userDetailPage = "/users/" + u.username
		}
	}

	// 查找魔力�?	tmps = htmlquery.Find(doc, `//a[contains(@href, "bonus/earnings")]`)
	if len(tmps) > 0 {
		bonusText := htmlquery.InnerText(tmps[0])
		bonusMatch := regexp.MustCompile(`([\d,.]+)`).FindStringSubmatch(bonusText)
		if len(bonusMatch) > 1 && strings.TrimSpace(bonusMatch[1]) != "" {
			stringUtils := utils.NewStringUtils()
			u.bonus = stringUtils.StrFloat(bonusMatch[1])
		}
	}
}

// parseSitePage 解析站点页面
func (u *Unit3dSiteUserInfo) parseSitePage(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (u *Unit3dSiteUserInfo) parseUserDetailInfo(htmlText string) {
	/*
	 解析用户额外信息，加入时间，等级
	*/
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		u.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 用户等级
	userLevelsText := htmlquery.Find(doc, `//div[contains(@class, "content")]//span[contains(@class, "badge-user")]/text()`)
	if len(userLevelsText) > 0 {
		u.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelsText[0]))
	}

	// 加入日期
	joinAtText := htmlquery.Find(doc, `//div[contains(@class, "content")]//h4[contains(text(), "注册日期") or contains(text(), "註冊日期") or contains(text(), "Registration date")]/text()`)
	if len(joinAtText) > 0 {
		joinAtStr := htmlquery.InnerText(joinAtText[0])
		joinAtStr = strings.Replace(joinAtStr, "注册日期", "", -1)
		joinAtStr = strings.Replace(joinAtStr, "註冊日期", "", -1)
		joinAtStr = strings.Replace(joinAtStr, "Registration date", "", -1)
		u.joinAt = stringUtils.UnifyDateTimeStr(joinAtStr)
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (u *Unit3dSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		u.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 9
	seedersCol := 2
	
	// 搜索size�?	sizeColNodes := htmlquery.Find(doc, `//thead//th[contains(@class,"size")]`)
	if len(sizeColNodes) > 0 {
		precedingNodes := htmlquery.Find(doc, `//thead//th[contains(@class,"size")][1]/preceding-sibling::th`)
		sizeCol = len(precedingNodes) + 1
	}
	
	// 搜索seeders�?	seedersColNodes := htmlquery.Find(doc, `//thead//th[contains(@class,"seeders")]`)
	if len(seedersColNodes) > 0 {
		precedingNodes := htmlquery.Find(doc, `//thead//th[contains(@class,"seeders")]/preceding-sibling::th`)
		seedersCol = len(precedingNodes) + 1
	}

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	seedingSizes := htmlquery.Find(doc, `//tr[position()]/td[`+string(rune(sizeCol))+`]`)
	seedingSeeders := htmlquery.Find(doc, `//tr[position()]/td[`+string(rune(seedersCol))+`]`)

	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			size := stringUtils.NumFilesize(strings.TrimSpace(htmlquery.InnerText(seedingSizes[i])))
			seeders := stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i])))

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	u.seeding += pageSeeding
	u.seedingSize += pageSeedingSize
	u.seedingInfo = append(u.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	nextPage := ""
	nextPages := htmlquery.Find(doc, `//ul[@class="pagination"]/li[contains(@class,"active")]/following-sibling::li`)
	if len(nextPages) > 1 {
		pageNum := strings.TrimSpace(htmlquery.InnerText(nextPages[0]))
		if stringUtils.IsDigit(pageNum) {
			nextPage = u.torrentSeedingPage + "&page=" + pageNum
		}
	}

	return nextPage
}

// parseUserTrafficInfo 解析用户流量信息
func (u *Unit3dSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	htmlText = u.prepareHTMLText(htmlText)
	stringUtils := utils.NewStringUtils()

	// 上传�?	uploadMatch := regexp.MustCompile(`[^总]上[传傳]�?[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(htmlText)
	if len(uploadMatch) > 1 {
		u.upload = stringUtils.NumFilesize(strings.TrimSpace(uploadMatch[1]))
	} else {
		u.upload = 0
	}

	// 下载�?	downloadMatch := regexp.MustCompile(`[^总子影力]下[载載]�?[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(htmlText)
	if len(downloadMatch) > 1 {
		u.download = stringUtils.NumFilesize(strings.TrimSpace(downloadMatch[1]))
	} else {
		u.download = 0
	}

	// 分享�?	ratioMatch := regexp.MustCompile(`分享率[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+)`).FindStringSubmatch(htmlText)
	if len(ratioMatch) > 1 && strings.TrimSpace(ratioMatch[1]) != "" {
		u.ratio = stringUtils.StrFloat(ratioMatch[1])
	} else {
		u.ratio = 0.0
	}
}

// parseMessageUnreadLinks 解析未读消息链接
func (u *Unit3dSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (u *Unit3dSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
