package indexer

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// NexusPhpSiteUserInfo NexusPhp站点用户信息解析�?type NexusPhpSiteUserInfo struct {
	*SiteParserBaseImpl
}

// NewNexusPhpSiteUserInfo 创建NexusPhp站点用户信息解析器实�?func NewNexusPhpSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *NexusPhpSiteUserInfo {

	parser := &NexusPhpSiteUserInfo{
		SiteParserBaseImpl: NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.schema = NexusPhp

	return parser
}

// parseSeedingPages 解析做种页面 - 可被子类重写
func (n *NexusPhpSiteUserInfo) parseSeedingPages() {
	if n.torrentSeedingPage != "" {
		// 第一�?		nextPage := n.parseUserTorrentSeedingInfo(
			n.getPageContent(n.joinURL(n.baseURL, n.torrentSeedingPage), n.torrentSeedingParams, n.torrentSeedingHeaders),
			false)

		// 其他页处�?		for nextPage != "" && nextPage != "false" {
			fullURL := n.joinURL(n.baseURL, n.torrentSeedingPage)
			nextFullURL := n.joinURL(fullURL, nextPage)
			nextPage = n.parseUserTorrentSeedingInfo(
				n.getPageContent(nextFullURL, n.torrentSeedingParams, n.torrentSeedingHeaders),
				true)
		}
	}
}

// ParseUserTrafficInfo 解析用户的上传，下载，分享率等信�?- 可被子类调用
func (n *NexusPhpSiteUserInfo) ParseUserTrafficInfo(htmlText string) {
	n.parseUserTrafficInfo(htmlText)
}

// ParseUserDetailInfo 解析用户的详细信�?- 可被子类调用
func (n *NexusPhpSiteUserInfo) ParseUserDetailInfo(htmlText string) {
	n.parseUserDetailInfo(htmlText)
}

// GetUserLevel 获取用户等级 - 可被子类调用
func (n *NexusPhpSiteUserInfo) GetUserLevel(doc *htmlquery.Node) {
	// 基础实现留空，子类可覆盖
}

// parseSitePage 解析站点页面
func (n *NexusPhpSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = n.prepareHTMLText(htmlText)

	// 查找用户详情页面链接
	userDetailMatch := regexp.MustCompile(`userdetails.php\?id=(\d+)`).FindStringSubmatch(htmlText)
	if len(userDetailMatch) > 1 && strings.TrimSpace(userDetailMatch[0]) != "" {
		n.userDetailPage = strings.TrimSpace(userDetailMatch[0])
		n.userid = userDetailMatch[1]
		n.torrentSeedingPage = "getusertorrentlistajax.php?userid=" + n.userid + "&type=seeding"
	} else {
		userDetailMatch = regexp.MustCompile(`(userdetails)`).FindStringSubmatch(htmlText)
		if len(userDetailMatch) > 0 && strings.TrimSpace(userDetailMatch[0]) != "" {
			n.userDetailPage = strings.TrimSpace(userDetailMatch[0])
			n.userid = ""
			n.torrentSeedingPage = ""
		}
	}
}

// parseMessageUnread 解析未读短消息数�?func (n *NexusPhpSiteUserInfo) parseMessageUnread(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 查找消息链接
	messageLabels1 := htmlquery.Find(doc, `//a[@href="messages.php"]/..`)
	messageLabels2 := htmlquery.Find(doc, `//a[contains(@href, "messages.php")]/..`)
	messageLabels := append(messageLabels1, messageLabels2...)

	if len(messageLabels) > 0 {
		messageText := htmlquery.InnerText(messageLabels[0])

		n.logger.Debug("消息原始信息", zap.String("site", n.siteName), zap.String("message", messageText))
		messageUnreadRegex := regexp.MustCompile(`[^Date](信息箱\s*|\((?![^)]*:)|你有\xa0)(\d+)`)
		messageUnreadMatch := messageUnreadRegex.FindAllStringSubmatch(messageText, -1)

		if len(messageUnreadMatch) > 0 && len(messageUnreadMatch[len(messageUnreadMatch)-1]) == 3 {
			n.messageUnread = stringUtils.StrInt(messageUnreadMatch[len(messageUnreadMatch)-1][2])
		} else if stringUtils.IsDigit(messageText) {
			n.messageUnread = stringUtils.StrInt(messageText)
		}
	}
}

// parseUserBaseInfo 解析用户基本信息
func (n *NexusPhpSiteUserInfo) parseUserBaseInfo(htmlText string) {
	// 合并解析，减少额外请求调�?	n.parseUserTrafficInfo(htmlText)
	n.userTrafficPage = ""

	n.parseMessageUnread(htmlText)

	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 查找用户�?	ret1 := htmlquery.Find(doc, `//a[contains(@href, "userdetails") and contains(@href, "`+n.userid+`")]//b//text()`)
	if len(ret1) > 0 {
		n.username = htmlquery.InnerText(ret1[0])
		return
	}

	ret2 := htmlquery.Find(doc, `//a[contains(@href, "userdetails") and contains(@href, "`+n.userid+`")]//text()`)
	if len(ret2) > 0 {
		n.username = htmlquery.InnerText(ret2[0])
		return
	}

	ret3 := htmlquery.Find(doc, `//a[contains(@href, "userdetails")]//strong//text()`)
	if len(ret3) > 0 {
		n.username = htmlquery.InnerText(ret3[0])
		return
	}
}

// parseUserTrafficInfo 解析用户流量信息
func (n *NexusPhpSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	htmlText = n.prepareHTMLText(htmlText)
	stringUtils := utils.NewStringUtils()

	// 上传�?	uploadMatch := regexp.MustCompile(`[^总]上[传傳]�?[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(htmlText)
	if len(uploadMatch) > 1 {
		n.upload = stringUtils.NumFilesize(strings.TrimSpace(uploadMatch[1]))
	} else {
		n.upload = 0
	}

	// 下载�?	downloadMatch := regexp.MustCompile(`[^总子影力]下[载載]�?[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(htmlText)
	if len(downloadMatch) > 1 {
		n.download = stringUtils.NumFilesize(strings.TrimSpace(downloadMatch[1]))
	} else {
		n.download = 0
	}

	// 分享�?	ratioMatch := regexp.MustCompile(`分享率[:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+)`).FindStringSubmatch(htmlText)
	// 计算分享�?	calcRatio := 0.0
	if n.download > 0 {
		calcRatio = float64(n.upload) / float64(n.download)
		// 保留3位小�?		calcRatio = float64(int(calcRatio*1000)) / 1000
	}
	// 优先使用页面上的分享�?	if len(ratioMatch) > 1 && strings.TrimSpace(ratioMatch[1]) != "" {
		n.ratio = stringUtils.StrFloat(ratioMatch[1])
	} else {
		n.ratio = calcRatio
	}

	// 下载�?	leechingMatch := regexp.MustCompile(`(Torrents leeching|下载�?[\u4E00-\u9FA5\D\s]+(\d+)[\s\S]+<`).FindStringSubmatch(htmlText)
	if len(leechingMatch) > 2 && strings.TrimSpace(leechingMatch[2]) != "" {
		n.leeching = stringUtils.StrInt(leechingMatch[2])
	} else {
		n.leeching = 0
	}

	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	// 解析魔力�?	hasUcoin, bonus := n.parseUcoin(doc)
	if hasUcoin {
		n.bonus = bonus
		return
	}

	tmps := htmlquery.Find(doc, `//a[contains(@href,"mybonus")]/text()`)
	if len(tmps) > 0 {
		bonusText := strings.TrimSpace(htmlquery.InnerText(tmps[0]))
		bonusMatch := regexp.MustCompile(`([\d,.]+)`).FindStringSubmatch(bonusText)
		if len(bonusMatch) > 1 && strings.TrimSpace(bonusMatch[1]) != "" {
			n.bonus = stringUtils.StrFloat(bonusMatch[1])
			return
		}
	}

	bonusMatch := regexp.MustCompile(`mybonus.[\[\]:�?>/a-zA-Z_\-="'\s#;.(使用魔力值豆]+\s*([\d,.]+)[<()&\s]`).FindStringSubmatch(htmlText)
	if len(bonusMatch) > 1 && strings.TrimSpace(bonusMatch[1]) != "" {
		n.bonus = stringUtils.StrFloat(bonusMatch[1])
		return
	}

	bonusMatch2 := regexp.MustCompile(`[魔力值|\]][\[\]:�?>/a-zA-Z_\-="'\s#;]+\s*([\d,.]+|\"[\d,.]+\")[<>()&\s]`).FindStringSubmatch(htmlText)
	if len(bonusMatch2) > 1 && strings.TrimSpace(bonusMatch2[1]) != "" {
		n.bonus = stringUtils.StrFloat(strings.Trim(bonusMatch2[1], "\""))
	}
}

// parseUcoin 解析ucoin, 统一转换为铜�?func (n *NexusPhpSiteUserInfo) parseUcoin(doc *htmlquery.Node) (bool, float64) {
	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElementNode(doc) {
		return false, 0.0
	}

	var gold, silver, copper float64
	var hasGold, hasSilver, hasCopper bool

	golds := htmlquery.Find(doc, `//span[@class = "ucoin-symbol ucoin-gold"]//text()`)
	if len(golds) > 0 {
		gold = stringUtils.StrFloat(htmlquery.InnerText(golds[len(golds)-1]))
		hasGold = true
	}

	silvers := htmlquery.Find(doc, `//span[@class = "ucoin-symbol ucoin-silver"]//text()`)
	if len(silvers) > 0 {
		silver = stringUtils.StrFloat(htmlquery.InnerText(silvers[len(silvers)-1]))
		hasSilver = true
	}

	coppers := htmlquery.Find(doc, `//span[@class = "ucoin-symbol ucoin-copper"]//text()`)
	if len(coppers) > 0 {
		copper = stringUtils.StrFloat(htmlquery.InnerText(coppers[len(coppers)-1]))
		hasCopper = true
	}

	if hasGold || hasSilver || hasCopper {
		if !hasGold {
			gold = 0
		}
		if !hasSilver {
			silver = 0
		}
		if !hasCopper {
			copper = 0
		}
		return true, gold*100*100 + silver*100 + copper
	}

	return false, 0.0
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (n *NexusPhpSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	// 替换转义字符
	htmlText = strings.ReplaceAll(htmlText, `\/`, "/")
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	// 首页存在扩展链接，使用扩展链�?	seedingURLText := htmlquery.Find(doc, `//a[contains(@href,"torrents.php") and contains(@href,"seeding")]/@href`)
	if !multiPage && len(seedingURLText) > 0 {
		seedingPage := strings.TrimSpace(htmlquery.SelectAttr(seedingURLText[0], "href"))
		if seedingPage != "" {
			n.torrentSeedingPage = seedingPage
			return n.torrentSeedingPage
		}
	}

	sizeCol := 3
	seedersCol := 4

	// 搜索size�?	sizeColXPath := `//tr[position()=1]/td[(img[@class="size"] and img[@alt="size"]) or (text() = "大小") or (a/img[@class="size" and @alt="size"])]`
	if len(htmlquery.Find(doc, sizeColXPath)) > 0 {
		precedingNodes := htmlquery.Find(doc, sizeColXPath+`/preceding-sibling::td`)
		sizeCol = len(precedingNodes) + 1
	}

	// 搜索seeders�?	seedersColXPath := `//tr[position()=1]/td[(img[@class="seeders"] and img[@alt="seeders"]) or (text() = "在做�?) or (a/img[@class="seeders" and @alt="seeders"])]`
	if len(htmlquery.Find(doc, seedersColXPath)) > 0 {
		precedingNodes := htmlquery.Find(doc, seedersColXPath+`/preceding-sibling::td`)
		seedersCol = len(precedingNodes) + 1
	}

	pageSeeding := 0
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	// 如果 table class="torrents"，则增加table[@class="torrents"]
	tableClass := ""
	if len(htmlquery.Find(doc, `//table[@class="torrents"]`)) > 0 {
		tableClass = `//table[@class="torrents"]`
	}

	seedingSizes := htmlquery.Find(doc, tableClass+`//tr[position()>1]/td[`+string(rune(sizeCol))+`]`)
	seedingSeeders := htmlquery.Find(doc, tableClass+`//tr[position()>1]/td[`+string(rune(seedersCol))+`]/b/a/text()`)
	
	if len(seedingSeeders) == 0 {
		seedingSeeders = htmlquery.Find(doc, tableClass+`//tr[position()>1]/td[`+string(rune(seedersCol))+`]//text()`)
	}

	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		pageSeeding = len(seedingSizes)

		for i := 0; i < len(seedingSizes); i++ {
			size := stringUtils.NumFilesize(strings.TrimSpace(htmlquery.InnerText(seedingSizes[i])))
			seeders := 0
			if i < len(seedingSeeders) {
				seeders = stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i])))
			}

			pageSeedingSize += size
			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	n.seeding += pageSeeding
	n.seedingSize += pageSeedingSize
	n.seedingInfo = append(n.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	nextPage := ""
	nextPageText := htmlquery.Find(doc, `//a[contains(.//text(), "下一�?) or contains(.//text(), "下一�?) or contains(.//text(), ">")]/@href`)

	// 防止识别到详情页
	for len(nextPageText) > 0 {
		nextPage = strings.TrimSpace(htmlquery.SelectAttr(nextPageText[len(nextPageText)-1], "href"))
		nextPageText = nextPageText[:len(nextPageText)-1]
		if !strings.HasPrefix(nextPage, "details.php") {
			break
		}
		nextPage = ""
	}

	// fix up page url
	if nextPage != "" {
		if !strings.Contains(nextPage, n.userid) {
			nextPage += "&userid=" + n.userid + "&type=seeding"
		}
	}

	return nextPage
}

// parseUserDetailInfo 解析用户详细信息，加入时间，等级
func (n *NexusPhpSiteUserInfo) parseUserDetailInfo(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	n.GetUserLevel(doc)
	n.fixupTrafficInfo(doc)

	// 加入日期
	joinAtText := htmlquery.Find(doc, `//tr/td[text()="加入日期" or text()="注册日期" or *[text()="加入日期"]]/following-sibling::td[1]//text() | //div/b[text()="加入日期"]/../text()`)
	if len(joinAtText) > 0 {
		joinAtStr := strings.TrimSpace(htmlquery.InnerText(joinAtText[0]))
		if idx := strings.Index(joinAtStr, " ("); idx != -1 {
			joinAtStr = joinAtStr[:idx]
		}
		n.joinAt = stringUtils.UnifyDateTimeStr(joinAtStr)
	}

	// 做种体积 & 做种�?	// seeding 页面获取不到的话，此处再获取一�?	seedingSizes := htmlquery.Find(doc, `//tr/td[text()="当前上传"]/following-sibling::td[1]//table[tr[1][td[4 and text()="尺寸"]]]//tr[position()>1]/td[4]`)
	seedingSeeders := htmlquery.Find(doc, `//tr/td[text()="当前上传"]/following-sibling::td[1]//table[tr[1][td[5 and text()="做种�?]]]//tr[position()>1]/td[5]//text()`)
	
	tmpSeeding := len(seedingSizes)
	tmpSeedingSize := 0
	tmpSeedingInfo := make([]interface{}, 0)
	
	for i := 0; i < len(seedingSizes); i++ {
		size := stringUtils.NumFilesize(strings.TrimSpace(htmlquery.InnerText(seedingSizes[i])))
		seeders := 0
		if i < len(seedingSeeders) {
			seeders = stringUtils.StrInt(strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i])))
		}

		tmpSeedingSize += size
		tmpSeedingInfo = append(tmpSeedingInfo, []interface{}{seeders, size})
	}

	if n.seedingSize == 0 {
		n.seedingSize = tmpSeedingSize
	}
	if n.seeding == 0 {
		n.seeding = tmpSeeding
	}
	if len(n.seedingInfo) == 0 {
		n.seedingInfo = tmpSeedingInfo
	}

	seedingSizesText := htmlquery.Find(doc, `//tr/td[text()="做种统计"]/following-sibling::td[1]//text()`)
	if len(seedingSizesText) > 0 {
		seedingText := htmlquery.InnerText(seedingSizesText[0])
		seedingMatch := regexp.MustCompile(`总做种数:\s+(\d+)`).FindStringSubmatch(seedingText)
		seedingSizeMatch := regexp.MustCompile(`总做种体�?\s+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(seedingText)
		
		tmpSeeding = 0
		tmpSeedingSize = 0
		
		if len(seedingMatch) > 1 {
			tmpSeeding = stringUtils.StrInt(seedingMatch[1])
		}
		if len(seedingSizeMatch) > 1 {
			tmpSeedingSize = stringUtils.NumFilesize(strings.TrimSpace(seedingSizeMatch[1]))
		}
	}

	if n.seedingSize == 0 {
		n.seedingSize = tmpSeedingSize
	}
	if n.seeding == 0 {
		n.seeding = tmpSeeding
	}

	n.fixupTorrentSeedingPage(doc)
}

// fixupTorrentSeedingPage 修正种子页面链接
func (n *NexusPhpSiteUserInfo) fixupTorrentSeedingPage(doc *htmlquery.Node) {
	// 单独的种子页�?	seedingURLText := htmlquery.Find(doc, `//a[contains(@href,"getusertorrentlist.php") and contains(@href,"seeding")]/@href`)
	if len(seedingURLText) > 0 {
		n.torrentSeedingPage = strings.TrimSpace(htmlquery.SelectAttr(seedingURLText[0], "href"))
		return
	}

	// 从JS调用种获取用户ID
	seedingURLText = htmlquery.Find(doc, `//a[contains(@href, "javascript: getusertorrentlistajax") and contains(@href,"seeding")]/@href`)
	csrfText := htmlquery.Find(doc, `//meta[@name="x-csrf"]/@content`)
	
	if n.torrentSeedingPage == "" && len(seedingURLText) > 0 {
		userJsMatch := regexp.MustCompile(`javascript: getusertorrentlistajax\(\s*'(\d+)`).FindStringSubmatch(htmlquery.SelectAttr(seedingURLText[0], "href"))
		if len(userJsMatch) > 1 && strings.TrimSpace(userJsMatch[1]) != "" {
			n.userid = userJsMatch[1]
			n.torrentSeedingPage = "getusertorrentlistajax.php?userid=" + n.userid + "&type=seeding"
			return
		}
	} else if len(seedingURLText) > 0 && len(csrfText) > 0 {
		csrf := strings.TrimSpace(htmlquery.SelectAttr(csrfText[0], "content"))
		if csrf != "" {
			n.torrentSeedingPage = "ajax_getusertorrentlist.php"
			n.torrentSeedingParams = map[string]string{
				"userid": n.userid,
				"type":   "seeding",
				"csrf":   csrf,
			}
		}
	}

	// 分类做种模式
	// 临时屏蔽
	// seedingUrlText = htmlquery.Find(doc, `//tr/td[text()="当前做种"]/following-sibling::td[1]/table//td/a[contains(@href,"seeding")]/@href`)
	// if len(seedingUrlText) > 0 {
	//    n.torrentSeedingPage = htmlquery.SelectAttr(seedingUrlText[0], "href")
	// }
}

// getUserLevel 获取用户等级
func (n *NexusPhpSiteUserInfo) getUserLevel(doc *htmlquery.Node) {
	// 等级 获取同一行等级数据，图片格式等级，取title信息，否则取文本信息
	userLevelsText := htmlquery.Find(doc, `//tr/td[text()="等級" or text()="等级" or *[text()="等级"]]/following-sibling::td[1]/img[1]/@title`)
	if len(userLevelsText) > 0 {
		n.userLevel = strings.TrimSpace(htmlquery.SelectAttr(userLevelsText[0], "title"))
		return
	}

	userLevelsText = htmlquery.Find(doc, `//tr/td[text()="等級" or text()="等级"]/following-sibling::td[1 and not(img)] | //tr/td[text()="等級" or text()="等级"]/following-sibling::td[1 and img[not(@title)]]`)
	if len(userLevelsText) > 0 {
		n.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelsText[0]))
		return
	}

	userLevelsText = htmlquery.Find(doc, `//tr/td[text()="等級" or text()="等级"]/following-sibling::td[1]`)
	if len(userLevelsText) > 0 {
		n.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelsText[0]))
		return
	}

	userLevelsText = htmlquery.Find(doc, `//a[contains(@href, "userdetails")]/text()`)
	if n.userLevel == "" && len(userLevelsText) > 0 {
		for _, userLevelText := range userLevelsText {
			userLevelMatch := regexp.MustCompile(`\[(.*)\]`).FindStringSubmatch(htmlquery.InnerText(userLevelText))
			if len(userLevelMatch) > 1 && strings.TrimSpace(userLevelMatch[1]) != "" {
				n.userLevel = strings.TrimSpace(userLevelMatch[1])
				break
			}
		}
	}
}

// fixupTrafficInfo 修正流量信息
func (n *NexusPhpSiteUserInfo) fixupTrafficInfo(doc *htmlquery.Node) {
	// fixup bonus
	if n.bonus == 0 {
		bonusText := htmlquery.Find(doc, `//tr/td[text()="魔力�? or text()="猫粮"]/following-sibling::td[1]/text()`)
		if len(bonusText) > 0 {
			stringUtils := utils.NewStringUtils()
			n.bonus = stringUtils.StrFloat(strings.TrimSpace(htmlquery.InnerText(bonusText[0])))
		}
	}
}

// parseMessageUnreadLinks 获取未阅读消息链�?func (n *NexusPhpSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	messageLinks := htmlquery.Find(doc, `//tr[not(./td/img[@alt="Read"])]/td/a[contains(@href, "viewmessage")]/@href`)
	for _, link := range messageLinks {
		msgLinks = append(msgLinks, htmlquery.SelectAttr(link, "href"))
	}

	// 是否存在下页数据
	nextPage := ""
	nextPageText := htmlquery.Find(doc, `//a[contains(.//text(), "下一�?) or contains(.//text(), "下一�?)]/@href`)
	if len(nextPageText) > 0 {
		nextPage = strings.TrimSpace(htmlquery.SelectAttr(nextPageText[len(nextPageText)-1], "href"))
	}

	return nextPage
}

// parseMessageContent 解析短消息内�?func (n *NexusPhpSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return "", "", ""
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return "", "", ""
	}

	// 标题
	messageHeadText := ""
	messageHead := htmlquery.Find(doc, `//h1/text() | //div[@class="layui-card-header"]/span[1]/text()`)
	if len(messageHead) > 0 {
		messageHeadText = strings.TrimSpace(htmlquery.InnerText(messageHead[len(messageHead)-1]))
	}

	// 消息时间
	messageDateText := ""
	messageDate := htmlquery.Find(doc, `//h1/following-sibling::table[.//tr/td[@class="colhead"]]//tr[2]/td[2] | //div[@class="layui-card-header"]/span[2]/span[2]`)
	if len(messageDate) > 0 {
		messageDateText = strings.TrimSpace(htmlquery.InnerText(messageDate[0]))
	}

	// 消息内容
	messageContentText := ""
	messageContent := htmlquery.Find(doc, `//h1/following-sibling::table[.//tr/td[@class="colhead"]]//tr[3]/td | //div[contains(@class,"layui-card-body")]`)
	if len(messageContent) > 0 {
		messageContentText = strings.TrimSpace(htmlquery.InnerText(messageContent[0]))
	}

	return messageHeadText, messageDateText, messageContentText
}
