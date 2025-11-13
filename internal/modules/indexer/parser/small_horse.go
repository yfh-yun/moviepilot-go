package parser

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// SmallHorseSiteUserInfo SmallHorse站点用户信息解析�?type SmallHorseSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewSmallHorseSiteUserInfo 创建SmallHorse站点用户信息解析器实�?func NewSmallHorseSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *SmallHorseSiteUserInfo {

	parser := &SmallHorseSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面
func (s *SmallHorseSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = s.prepareHTMLText(htmlText)

	userDetail := regexp.MustCompile(`user.php\?id=(\d+)`).FindStringSubmatch(htmlText)
	if len(userDetail) > 1 && strings.TrimSpace(userDetail[0]) != "" {
		s.userDetailPage = strings.TrimSpace(userDetail[0])
		s.userid = userDetail[1]
		s.torrentSeedingPage = "torrents.php?type=seeding&userid=" + s.userid
	}
	s.userTrafficPage = "user.php?id=" + s.userid
}

// parseUserBaseInfo 解析用户基础信息
func (s *SmallHorseSiteUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = s.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		s.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	ret := htmlquery.Find(doc, `//a[contains(@href, "user.php")]//text()`)
	if len(ret) > 0 {
		s.username = htmlquery.InnerText(ret[0])
	}
}

// parseUserTrafficInfo 解析用户流量信息
func (s *SmallHorseSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 上传/下载/分享�?[做种�?魔力值]
	htmlText = s.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		s.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	tmps := htmlquery.Find(doc, `//ul[@class = "stats nobullet"]`)
	if len(tmps) > 0 {
		// 解析加入时间
		if len(tmps) > 1 {
			liNodes := htmlquery.Find(tmps[1], "li")
			if len(liNodes) > 0 {
				spanTextNodes := htmlquery.Find(liNodes[0], "span//text()")
				if len(spanTextNodes) > 0 {
					s.joinAt = stringUtils.UnifyDateTimeStr(htmlquery.InnerText(spanTextNodes[0]))
				}
			}

			// 上传�?			if len(liNodes) > 2 {
				textNodes := htmlquery.Find(liNodes[2], "text()")
				if len(textNodes) > 0 {
					uploadText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
					s.upload = stringUtils.NumFilesize(strings.TrimSpace(uploadText))
				}
			}

			// 下载�?			if len(liNodes) > 3 {
				textNodes := htmlquery.Find(liNodes[3], "text()")
				if len(textNodes) > 0 {
					downloadText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
					s.download = stringUtils.NumFilesize(strings.TrimSpace(downloadText))
				}
			}

			// 分享�?			if len(liNodes) > 4 {
				spanTextNodes := htmlquery.Find(liNodes[4], "span//text()")
				if len(spanTextNodes) > 0 {
					ratioText := strings.Replace(htmlquery.InnerText(spanTextNodes[0]), "�?, "0", -1)
					s.ratio = stringUtils.StrFloat(ratioText)
				} else if len(liNodes) > 5 {
					textNodes := htmlquery.Find(liNodes[5], "text()")
					if len(textNodes) > 0 {
						ratioText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
						s.ratio = stringUtils.StrFloat(strings.TrimSpace(ratioText))
					}
				}
			}

			// 魔力�?			if len(liNodes) > 5 {
				textNodes := htmlquery.Find(liNodes[5], "text()")
				if len(textNodes) > 0 {
					bonusText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
					s.bonus = stringUtils.StrFloat(strings.TrimSpace(bonusText))
				}
			}
		}

		// 用户等级
		if len(tmps) > 3 {
			liNodes := htmlquery.Find(tmps[3], "li")
			if len(liNodes) > 0 {
				textNodes := htmlquery.Find(liNodes[0], "text()")
				if len(textNodes) > 0 {
					levelText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
					s.userLevel = strings.TrimSpace(levelText)
				}
			}
		}

		// 下载�?		if len(tmps) > 4 {
			liNodes := htmlquery.Find(tmps[4], "li")
			if len(liNodes) > 6 {
				textNodes := htmlquery.Find(liNodes[6], "text()")
				if len(textNodes) > 0 {
					leechingText := strings.Split(htmlquery.InnerText(textNodes[0]), ":")[1]
					leechingText = strings.Replace(leechingText, "[", "", -1)
					s.leeching = stringUtils.StrInt(strings.TrimSpace(leechingText))
				}
			}
		}
	}
}

// parseUserDetailInfo 解析用户详细信息
func (s *SmallHorseSiteUserInfo) parseUserDetailInfo(htmlText string) {
	// 空实�?}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (s *SmallHorseSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		s.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 6
	seedersCol := 8

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	seedingSizes := htmlquery.Find(doc, `//table[@id="torrent_table"]//tr[position()>1]/td[`+string(rune(sizeCol))+`]`)
	seedingSeeders := htmlquery.Find(doc, `//table[@id="torrent_table"]//tr[position()>1]/td[`+string(rune(seedersCol))+`]`)

	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			size := stringUtils.NumFilesize(strings.TrimSpace(htmlquery.InnerText(seedingSizes[i])))
			seeders := stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i])))

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	s.seeding += pageSeeding
	s.seedingSize += pageSeedingSize
	s.seedingInfo = append(s.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	nextPage := ""
	nextPages := htmlquery.Find(doc, `//ul[@class="pagination"]/li[contains(@class,"active")]/following-sibling::li`)
	if len(nextPages) > 1 {
		pageNum := strings.TrimSpace(htmlquery.InnerText(nextPages[0]))
		if stringUtils.IsDigit(pageNum) {
			nextPage = s.torrentSeedingPage + "&page=" + pageNum
		}
	}

	return nextPage
}

// parseMessageUnreadLinks 解析未读消息链接
func (s *SmallHorseSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (s *SmallHorseSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
