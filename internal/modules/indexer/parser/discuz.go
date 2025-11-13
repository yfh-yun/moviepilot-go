package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// DiscuzUserInfo Discuz站点用户信息解析�?type DiscuzUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewDiscuzUserInfo 创建Discuz站点用户信息解析器实�?func NewDiscuzUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *DiscuzUserInfo {
	
	parser := &DiscuzUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}
	
	// 设置站点模式
	// parser.SiteParserBaseImpl.GetSchema().(*indexer.SiteSchema)
	
	return parser
}

// parseUserBaseInfo 解析用户基础信息
func (d *DiscuzUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = d.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		d.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	// 查找用户信息链接
	userInfoNodes := htmlquery.Find(doc, `//a[contains(@href, "&uid=")]`)
	if len(userInfoNodes) > 0 {
		href := htmlquery.SelectAttr(userInfoNodes[0], "href")
		userIDMatch := regexp.MustCompile(`&uid=(\d+)`).FindStringSubmatch(href)
		if len(userIDMatch) > 1 && strings.TrimSpace(userIDMatch[1]) != "" {
			d.userid = userIDMatch[1]
			d.torrentSeedingPage = "forum.php?&mod=torrents&cat_5up=on"
			d.userDetailPage = href
			d.username = strings.TrimSpace(htmlquery.InnerText(userInfoNodes[0]))
		}
	}
}

// parseSitePage 解析站点相关信息页面
func (d *DiscuzUserInfo) parseSitePage(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息，加入时间，等级
func (d *DiscuzUserInfo) parseUserDetailInfo(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		d.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()

	// 用户等级
	userLevelsText := htmlquery.Find(doc, `//a[contains(@href, "usergroup")]/text()`)
	if len(userLevelsText) > 0 {
		d.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelsText[len(userLevelsText)-1]))
	}

	// 加入日期
	joinAtText := htmlquery.Find(doc, `//li[em[text()="注册时间"]]/text()`)
	if len(joinAtText) > 0 {
		d.joinAt = stringUtils.UnifyDateTimeStr(strings.TrimSpace(htmlquery.InnerText(joinAtText[0])))
	}

	// 分享�?	ratioText := htmlquery.Find(doc, `//li[contains(.//text(), "分享�?)]//text()`)
	if len(ratioText) > 0 {
		ratioMatch := regexp.MustCompile(`\(([\d,.]+)\)`).FindStringSubmatch(htmlquery.InnerText(ratioText[0]))
		if len(ratioMatch) > 1 && strings.TrimSpace(ratioMatch[1]) != "" {
			d.bonus = stringUtils.StrFloat(ratioMatch[1])
		}
	}

	// 积分
	bounsText := htmlquery.Find(doc, `//li[em[text()="积分"]]/text()`)
	if len(bounsText) > 0 {
		d.bonus = stringUtils.StrFloat(strings.TrimSpace(htmlquery.InnerText(bounsText[0])))
	}

	// 上传
	uploadText := htmlquery.Find(doc, `//li[em[contains(text(),"上传�?)]]/text()`)
	if len(uploadText) > 0 {
		uploadStr := strings.TrimSpace(htmlquery.InnerText(uploadText[0]))
		parts := strings.Split(uploadStr, "/")
		if len(parts) > 1 {
			d.upload = stringUtils.NumFilesize(parts[len(parts)-1])
		}
	}

	// 下载
	downloadText := htmlquery.Find(doc, `//li[em[contains(text(),"下载�?)]]/text()`)
	if len(downloadText) > 0 {
		downloadStr := strings.TrimSpace(htmlquery.InnerText(downloadText[0]))
		parts := strings.Split(downloadStr, "/")
		if len(parts) > 1 {
			d.download = stringUtils.NumFilesize(parts[len(parts)-1])
		}
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (d *DiscuzUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		d.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	// 检查HTML元素是否有效
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 3
	seedersCol := 4
	
	// 搜索size�?	sizeColNodes := htmlquery.Find(doc, `//tr[position()=1]/td[.//img[@class="size"] and .//img[@alt="size"]]`)
	if len(sizeColNodes) > 0 {
		precedingNodes := htmlquery.Find(doc, `//tr[position()=1]/td[.//img[@class="size"] and .//img[@alt="size"]]/preceding-sibling::td`)
		sizeCol = len(precedingNodes) + 1
	}
	
	// 搜索seeders�?	seedersColNodes := htmlquery.Find(doc, `//tr[position()=1]/td[.//img[@class="seeders"] and .//img[@alt="seeders"]]`)
	if len(seedersColNodes) > 0 {
		precedingNodes := htmlquery.Find(doc, `//tr[position()=1]/td[.//img[@class="seeders"] and .//img[@alt="seeders"]]/preceding-sibling::td`)
		seedersCol = len(precedingNodes) + 1
	}

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	// 获取做种大小和做种者信�?	sizeXPath := fmt.Sprintf(`//tr[position()>1]/td[%d]`, sizeCol)
	seedingSizes := htmlquery.Find(doc, sizeXPath)
	
	seedersXPath := fmt.Sprintf(`//tr[position()>1]/td[%d]//text()`, seedersCol)
	seedingSeeders := htmlquery.Find(doc, seedersXPath)

	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			sizeText := strings.TrimSpace(htmlquery.InnerText(seedingSizes[i]))
			size := stringUtils.NumFilesize(sizeText)

			var seeders int = 0
			if i < len(seedingSeeders) {
				seedersText := strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i]))
				seeders = stringUtils.StrInt(seedersText)
			}

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	d.seeding += pageSeeding
	d.seedingSize += pageSeedingSize
	d.seedingInfo = append(d.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	var nextPage string
	nextPageText := htmlquery.Find(doc, `//a[contains(.//text(), "下一�?) or contains(.//text(), "下一�?)]/@href`)
	if len(nextPageText) > 0 {
		nextPage = strings.TrimSpace(htmlquery.SelectAttr(nextPageText[len(nextPageText)-1], "href"))
	}

	return nextPage
}

// parseUserTrafficInfo 解析用户流量信息
func (d *DiscuzUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseMessageUnreadLinks 解析未读消息链接
func (d *DiscuzUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (d *DiscuzUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
