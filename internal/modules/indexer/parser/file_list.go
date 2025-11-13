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

// FileListSiteUserInfo FileList站点用户信息解析�?type FileListSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewFileListSiteUserInfo 创建FileList站点用户信息解析器实�?func NewFileListSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *FileListSiteUserInfo {
	
	parser := &FileListSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}
	
	return parser
}

// parseSitePage 解析站点页面
func (f *FileListSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = f.prepareHTMLText(htmlText)

	userDetail := regexp.MustCompile(`userdetails\.php\?id=(\d+)`).FindStringSubmatch(htmlText)
	if len(userDetail) > 1 && strings.TrimSpace(userDetail[0]) != "" {
		f.userDetailPage = strings.TrimLeft(strings.TrimSpace(userDetail[0]), "/")
		f.userid = userDetail[1]
	}

	f.torrentSeedingPage = fmt.Sprintf("snatchlist.php?id=%s&action=torrents&type=seeding", f.userid)
}

// parseUserBaseInfo 解析用户基础信息
func (f *FileListSiteUserInfo) parseUserBaseInfo(htmlText string) {
	htmlText = f.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		f.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	xpath := fmt.Sprintf(`//a[contains(@href, "userdetails") and contains(@href, "%s")]//text()`, f.userid)
	ret := htmlquery.Find(doc, xpath)
	if len(ret) > 0 {
		f.username = htmlquery.InnerText(ret[0])
	}
}

// parseUserTrafficInfo 解析用户流量信息
func (f *FileListSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (f *FileListSiteUserInfo) parseUserDetailInfo(htmlText string) {
	htmlText = f.prepareHTMLText(htmlText)
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		f.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()

	// 上传�?	uploadHTML := htmlquery.Find(doc, `//table//tr/td[text()="Uploaded"]/following-sibling::td//text()`)
	if len(uploadHTML) > 0 {
		f.upload = stringUtils.NumFilesize(htmlquery.InnerText(uploadHTML[0]))
	}

	// 下载�?	downloadHTML := htmlquery.Find(doc, `//table//tr/td[text()="Downloaded"]/following-sibling::td//text()`)
	if len(downloadHTML) > 0 {
		f.download = stringUtils.NumFilesize(htmlquery.InnerText(downloadHTML[0]))
	}

	// 分享�?	ratioHTML := htmlquery.Find(doc, `//table//tr/td[text()="Share ratio"]/following-sibling::td//text()`)
	shareRatio := 0.0
	if len(ratioHTML) > 0 {
		shareRatio = stringUtils.StrFloat(htmlquery.InnerText(ratioHTML[0]))
	} else {
		shareRatio = 0
	}
	if f.download <= 0 {
		f.ratio = 0
	} else {
		f.ratio = shareRatio
	}

	// 做种信息
	seedHTML := htmlquery.Find(doc, `//table//tr/td[text()="Seed bonus"]/following-sibling::td//text()`)
	if len(seedHTML) > 0 {
		// 做种数量
		if len(seedHTML) > 1 {
			f.seeding = stringUtils.StrInt(htmlquery.InnerText(seedHTML[1]))
		}
		// 做种大小
		if len(seedHTML) > 3 {
			f.seedingSize = stringUtils.NumFilesize(htmlquery.InnerText(seedHTML[3]))
		}
	}

	// 用户等级
	userLevelHTML := htmlquery.Find(doc, `//table//tr/td[text()="Class"]/following-sibling::td//text()`)
	if len(userLevelHTML) > 0 {
		f.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelHTML[0]))
	}

	// 加入时间
	joinAtHTML := htmlquery.Find(doc, `//table//tr/td[contains(text(), "Join")]/following-sibling::td//text()`)
	if len(joinAtHTML) > 0 {
		joinAt := strings.Split(htmlquery.InnerText(joinAtHTML[0]), "(")[0]
		f.joinAt = stringUtils.UnifyDateTimeStr(strings.TrimSpace(joinAt))
	}

	// 魔力�?	bonusHTML := htmlquery.Find(doc, `//a[contains(@href, "shop.php")]`)
	if len(bonusHTML) > 0 {
		f.bonus = stringUtils.StrFloat(htmlquery.InnerText(bonusHTML[0]))
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (f *FileListSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		f.logger.Error("解析HTML失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	// 检查HTML元素是否有效
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return ""
	}

	sizeCol := 6
	seedersCol := 7

	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)
	
	// 做种大小
	sizeXPath := fmt.Sprintf(`//table/tr[position()>1]/td[%d]`, sizeCol)
	seedingSizes := htmlquery.Find(doc, sizeXPath)
	
	// 做种�?	seedersXPath := fmt.Sprintf(`//table/tr[position()>1]/td[%d]`, seedersCol)
	seedingSeeders := htmlquery.Find(doc, seedersXPath)
	
	if len(seedingSizes) > 0 && len(seedingSeeders) > 0 {
		for i := 0; i < len(seedingSizes); i++ {
			sizeText := strings.TrimSpace(htmlquery.InnerText(seedingSizes[i]))
			size := stringUtils.NumFilesize(sizeText)
			pageSeedingSize += size

			seedersText := ""
			if i < len(seedingSeeders) {
				seedersText = strings.TrimSpace(htmlquery.InnerText(seedingSeeders[i]))
			}
			seeders := stringUtils.StrInt(seedersText)

			pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
		}
	}

	f.seedingInfo = append(f.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	var nextPage string
	// 根据Python版本，这里返回空字符串，表示没有下一�?	return nextPage
}

// parseMessageUnreadLinks 解析未读消息链接
func (f *FileListSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (f *FileListSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	return "", "", ""
}
