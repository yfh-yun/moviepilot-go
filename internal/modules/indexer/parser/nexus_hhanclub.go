package parser

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// NexusHhanclubSiteUserInfo NexusHhanclub站点用户信息解析�?type NexusHhanclubSiteUserInfo struct {
	*indexer.NexusPhpSiteUserInfo
}

// NewNexusHhanclubSiteUserInfo 创建NexusHhanclub站点用户信息解析器实�?func NewNexusHhanclubSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *NexusHhanclubSiteUserInfo {

	parser := &NexusHhanclubSiteUserInfo{
		NexusPhpSiteUserInfo: indexer.NewNexusPhpSiteUserInfo(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseUserTrafficInfo 解析用户流量信息
func (n *NexusHhanclubSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 调用父类方法
	n.NexusPhpSiteUserInfo.parseUserTrafficInfo(htmlText)

	htmlText = n.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()

	// 上传、下载、分享率
	uploadNodes := htmlquery.Find(doc, `//*[@id="user-info-panel"]/div[2]/div[2]/div[4]/text()`)
	if len(uploadNodes) > 0 {
		uploadText := htmlquery.InnerText(uploadNodes[0])
		uploadMatch := regexp.MustCompile(`[_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(uploadText)
		if len(uploadMatch) > 1 {
			n.upload = stringUtils.NumFilesize(strings.TrimSpace(uploadMatch[1]))
		}
	}

	downloadNodes := htmlquery.Find(doc, `//*[@id="user-info-panel"]/div[2]/div[2]/div[5]/text()`)
	if len(downloadNodes) > 0 {
		downloadText := htmlquery.InnerText(downloadNodes[0])
		downloadMatch := regexp.MustCompile(`[_<>/a-zA-Z-="'\s#;]+([\d,.\s]+[KMGTPI]*B)`).FindStringSubmatch(downloadText)
		if len(downloadMatch) > 1 {
			n.download = stringUtils.NumFilesize(strings.TrimSpace(downloadMatch[1]))
		}
	}

	ratioNodes := htmlquery.Find(doc, `//*[@id="user-info-panel"]/div[2]/div[1]/div[1]/div/text()`)
	if len(ratioNodes) > 0 {
		ratioText := htmlquery.InnerText(ratioNodes[0])
		ratioMatch := regexp.MustCompile(`分享率][:：_<>/a-zA-Z-="'\s#;]+([\d,.\s]+)`).FindStringSubmatch(ratioText)
		// 计算分享�?		calcRatio := 0.0
		if n.download > 0 {
			calcRatio = float64(n.upload) / float64(n.download)
			if calcRatio > 3 { // 保留3位小�?				calcRatio = float64(int(calcRatio*1000)) / 1000
			}
		}
		if len(ratioMatch) > 1 && strings.TrimSpace(ratioMatch[1]) != "" {
			n.ratio = stringUtils.StrFloat(ratioMatch[1])
		} else {
			n.ratio = calcRatio
		}
	}
}

// parseUserDetailInfo 解析用户详细信息
func (n *NexusHhanclubSiteUserInfo) parseUserDetailInfo(htmlText string) {
	// 调用父类方法
	n.NexusPhpSiteUserInfo.parseUserDetailInfo(htmlText)

	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 加入时间
	joinAtNodes := htmlquery.Find(doc, `//*[@id="mainContent"]/div/div[2]/div[4]/div[3]/span[2]/text()`)
	if len(joinAtNodes) > 0 {
		joinAtText := strings.TrimSpace(htmlquery.InnerText(joinAtNodes[0]))
		if idx := strings.Index(joinAtText, " ("); idx != -1 {
			joinAtText = joinAtText[:idx]
		}
		n.joinAt = stringUtils.UnifyDateTimeStr(joinAtText)
	}
}

// getUserLevel 获取用户等级
func (n *NexusHhanclubSiteUserInfo) getUserLevel(doc *htmlquery.Node) {
	// 调用父类方法
	n.NexusPhpSiteUserInfo.getUserLevel(doc)

	userLevelNodes := htmlquery.Find(doc, `//*[@id="mainContent"]/div/div[2]/div[2]/div[4]/span[2]/img/@title`)
	if len(userLevelNodes) > 0 {
		n.userLevel = htmlquery.SelectAttr(userLevelNodes[0], "title")
	}
}
