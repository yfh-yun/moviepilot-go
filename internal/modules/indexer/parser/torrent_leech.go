package parser

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// TorrentLeechSiteUserInfo TorrentLeech站点用户信息解析�?type TorrentLeechSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewTorrentLeechSiteUserInfo 创建TorrentLeech站点用户信息解析器实�?func NewTorrentLeechSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *TorrentLeechSiteUserInfo {

	parser := &TorrentLeechSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面
func (t *TorrentLeechSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = t.prepareHTMLText(htmlText)

	userDetail := regexp.MustCompile(`/profile/([^/]+)/`).FindStringSubmatch(htmlText)
	if len(userDetail) > 1 && strings.TrimSpace(userDetail[0]) != "" {
		t.userDetailPage = strings.TrimSpace(userDetail[0])
		t.userid = userDetail[1]
	}
	t.userTrafficPage = "profile/" + t.userid + "/view"
	t.torrentSeedingPage = "profile/" + t.userid + "/seeding"
}

// parseUserBaseInfo 解析用户基础信息
func (t *TorrentLeechSiteUserInfo) parseUserBaseInfo(htmlText string) {
	t.username = t.userid
}

// parseUserTrafficInfo 解析用户流量信息
func (t *TorrentLeechSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	/*
	 上传/下载/分享�?[做种�?魔力值]
	*/
	htmlText = t.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		t.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()

	// 上传�?	uploadHTML := htmlquery.Find(doc, `//div[contains(@class,"profile-uploaded")]//span/text()`)
	if len(uploadHTML) > 0 {
		t.upload = stringUtils.NumFilesize(htmlquery.InnerText(uploadHTML[0]))
	}

	// 下载�?	downloadHTML := htmlquery.Find(doc, `//div[contains(@class,"profile-downloaded")]//span/text()`)
	if len(downloadHTML) > 0 {
		t.download = stringUtils.NumFilesize(htmlquery.InnerText(downloadHTML[0]))
	}

	// 分享�?	ratioHTML := htmlquery.Find(doc, `//div[contains(@class,"profile-ratio")]//span/text()`)
	if len(ratioHTML) > 0 {
		ratioText := strings.Replace(htmlquery.InnerText(ratioHTML[0]), "�?, "0", -1)
		t.ratio = stringUtils.StrFloat(ratioText)
	}

	// 用户等级
	userLevelHTML := htmlquery.Find(doc, `//table[contains(@class, "profileViewTable")]//tr/td[text()="Class"]/following-sibling::td/text()`)
	if len(userLevelHTML) > 0 {
		t.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelHTML[0]))
	}

	// 加入时间
	joinAtHTML := htmlquery.Find(doc, `//table[contains(@class, "profileViewTable")]//tr/td[text()="Registration date"]/following-sibling::td/text()`)
	if len(joinAtHTML) > 0 {
		t.joinAt = stringUtils.UnifyDateTimeStr(strings.TrimSpace(htmlquery.InnerText(joinAtHTML[0])))
	}

	// 魔力�?	bonusHTML := htmlquery.Find(doc, `//span[contains(@class, "total-TL-points")]/text()`)
	if len(bonusHTML) > 0 {
		t.bonus = stringUtils.StrFloat(strings.TrimSpace(htmlquery.InnerText(bonusHTML[0])))
	}
}

// parseUserDetailInfo 解析用户详细信息
func (t *TorrentLeechSiteUserInfo) parseUserDetailInfo(htmlText string) {
	// 空实�?}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (t *TorrentLeechSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		t.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 2
	seedersCol := 7

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	seedingSizes := htmlquery.Find(doc, `//tbody/tr/td[`+string(rune(sizeCol))+`]`)
	seedingSeeders := htmlquery.Find(doc, `//tbody/tr/td[`+string(rune(seedersCol))+`]/text()`)

	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			size := stringUtils.NumFilesize(strings.TrimSpace(htmlquery.InnerText(seedingSizes[i])))
			seeders := stringUtils.StrInt(htmlquery.InnerText(seedingSeeders[i]))

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	t.seeding += pageSeeding
	t.seedingSize += pageSeedingSize
	t.seedingInfo = append(t.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (t *TorrentLeechSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (t *TorrentLeechSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
