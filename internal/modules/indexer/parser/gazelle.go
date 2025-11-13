package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// GazelleSiteUserInfo Gazelle站点用户信息解析�?type GazelleSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewGazelleSiteUserInfo 创建Gazelle站点用户信息解析器实�?func NewGazelleSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *GazelleSiteUserInfo {
	
	parser := &GazelleSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}
	
	return parser
}

// parseUserBaseInfo 解析用户基础信息
func (g *GazelleSiteUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = g.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		g.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	// 获取用户ID、做种页面和用户�?	tmps := htmlquery.Find(doc, `//a[contains(@href, "user.php?id=")]`)
	if len(tmps) > 0 {
		href := htmlquery.SelectAttr(tmps[0], "href")
		userIDMatch := regexp.MustCompile(`user\.php\?id=(\d+)`).FindStringSubmatch(href)
		if len(userIDMatch) > 1 && strings.TrimSpace(userIDMatch[0]) != "" {
			g.userid = userIDMatch[1]
			g.torrentSeedingPage = fmt.Sprintf("torrents.php?type=seeding&userid=%s", g.userid)
			g.userDetailPage = fmt.Sprintf("user.php?id=%s", g.userid)
			g.username = strings.TrimSpace(htmlquery.InnerText(tmps[0]))
		}
	}

	// 获取上传�?	tmps = htmlquery.Find(doc, `//*[@id="header-uploaded-value"]/@data-value`)
	if len(tmps) > 0 {
		stringUtils := utils.NewStringUtils()
		g.upload = stringUtils.NumFilesize(htmlquery.SelectAttr(tmps[0], "data-value"))
	} else {
		tmps = htmlquery.Find(doc, `//li[@id="stats_seeding"]/span/text()`)
		if len(tmps) > 0 {
			stringUtils := utils.NewStringUtils()
			g.upload = stringUtils.NumFilesize(htmlquery.InnerText(tmps[0]))
		}
	}

	// 获取下载�?	tmps = htmlquery.Find(doc, `//*[@id="header-downloaded-value"]/@data-value`)
	if len(tmps) > 0 {
		stringUtils := utils.NewStringUtils()
		g.download = stringUtils.NumFilesize(htmlquery.SelectAttr(tmps[0], "data-value"))
	} else {
		tmps = htmlquery.Find(doc, `//li[@id="stats_leeching"]/span/text()`)
		if len(tmps) > 0 {
			stringUtils := utils.NewStringUtils()
			g.download = stringUtils.NumFilesize(htmlquery.InnerText(tmps[0]))
		}
	}

	// 计算分享�?	if g.download <= 0 {
		g.ratio = 0.0
	} else {
		g.ratio = float64(g.upload) / float64(g.download)
		// 保留3位小�?		g.ratio = float64(int(g.ratio*1000)) / 1000
	}

	// 获取魔力�?	tmps = htmlquery.Find(doc, `//a[contains(@href, "bonus.php")]/@data-tooltip`)
	if len(tmps) > 0 {
		tooltip := htmlquery.SelectAttr(tmps[0], "data-tooltip")
		bonusMatch := regexp.MustCompile(`([\d,.]+)`).FindStringSubmatch(tooltip)
		if len(bonusMatch) > 1 && strings.TrimSpace(bonusMatch[1]) != "" {
			stringUtils := utils.NewStringUtils()
			g.bonus = stringUtils.StrFloat(bonusMatch[1])
		}
	} else {
		tmps = htmlquery.Find(doc, `//a[contains(@href, "bonus.php")]`)
		if len(tmps) > 0 {
			bonusText := htmlquery.InnerText(tmps[0])
			bonusMatch := regexp.MustCompile(`([\d,.]+)`).FindStringSubmatch(bonusText)
			if len(bonusMatch) > 1 && strings.TrimSpace(bonusMatch[1]) != "" {
				stringUtils := utils.NewStringUtils()
				g.bonus = stringUtils.StrFloat(bonusMatch[1])
			}
		}
	}
}

// parseSitePage 解析站点页面
func (g *GazelleSiteUserInfo) parseSitePage(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (g *GazelleSiteUserInfo) parseUserDetailInfo(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		g.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	// 检查HTML元素是否有效
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 用户等级
	userLevelsText := htmlquery.Find(doc, `//*[@id="class-value"]/@data-value`)
	if len(userLevelsText) > 0 {
		g.userLevel = strings.TrimSpace(htmlquery.SelectAttr(userLevelsText[0], "data-value"))
	} else {
		userLevelsText = htmlquery.Find(doc, `//li[contains(text(), "用户等级")]/text()`)
		if len(userLevelsText) > 0 {
			text := htmlquery.InnerText(userLevelsText[0])
			parts := strings.Split(text, ":")
			if len(parts) > 1 {
				g.userLevel = strings.TrimSpace(parts[1])
			}
		}
	}

	// 加入日期
	joinAtText := htmlquery.Find(doc, `//*[@id="join-date-value"]/@data-value`)
	if len(joinAtText) > 0 {
		joinDate := htmlquery.SelectAttr(joinAtText[0], "data-value")
		g.joinAt = stringUtils.UnifyDateTimeStr(strings.TrimSpace(joinDate))
	} else {
		joinAtText = htmlquery.Find(doc, `//div[contains(@class, "box_userinfo_stats")]//li[contains(text(), "加入时间")]/span/text()`)
		if len(joinAtText) > 0 {
			g.joinAt = stringUtils.UnifyDateTimeStr(strings.TrimSpace(htmlquery.InnerText(joinAtText[0])))
		}
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (g *GazelleSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		g.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	// 检查HTML元素是否有效
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 3
	// 搜索size�?	tds := htmlquery.Find(doc, `//table[contains(@id, "torrent")]//tr[1]/td`)
	if len(tds) > 0 {
		sizeCol = len(tds) - 3
	}
	// 搜索seeders�?	seedersCol := sizeCol + 2

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)
	
	// 做种大小
	sizeXPath := fmt.Sprintf(`//table[contains(@id, "torrent")]//tr[position()>1]/td[%d]`, sizeCol)
	seedingSizes := htmlquery.Find(doc, sizeXPath)
	
	// 做种�?	seedersXPath := fmt.Sprintf(`//table[contains(@id, "torrent")]//tr[position()>1]/td[%d]/text()`, seedersCol)
	seedingSeeders := htmlquery.Find(doc, seedersXPath)
	
	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			sizeText := strings.TrimSpace(htmlquery.InnerText(seedingSizes[i]))
			size := stringUtils.NumFilesize(sizeText)

			seeders := 0
			if i < len(seedingSeeders) {
				seedersText := htmlquery.InnerText(seedingSeeders[i])
				seedersInt, err := strconv.Atoi(seedersText)
				if err == nil {
					seeders = seedersInt
				}
			}

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	if multiPage {
		g.seeding += pageSeeding
		g.seedingSize += pageSeedingSize
		g.seedingInfo = append(g.seedingInfo, pageSeedingInfo...)
	} else {
		if g.seeding == 0 {
			g.seeding = pageSeeding
		}
		if g.seedingSize == 0 {
			g.seedingSize = pageSeedingSize
		}
		if len(g.seedingInfo) == 0 {
			g.seedingInfo = pageSeedingInfo
		}
	}

	// 是否存在下页数据
	var nextPage string
	nextPageText := htmlquery.Find(doc, `//a[contains(.//text(), "Next") or contains(.//text(), "下一�?)]/@href`)
	if len(nextPageText) > 0 {
		nextPage = strings.TrimSpace(htmlquery.SelectAttr(nextPageText[len(nextPageText)-1], "href"))
	}

	return nextPage
}

// parseUserTrafficInfo 解析用户流量信息
func (g *GazelleSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseMessageUnreadLinks 解析未读消息链接
func (g *GazelleSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (g *GazelleSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
