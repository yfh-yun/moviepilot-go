package parser

import (
	"encoding/json"
	"math"

	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// TYemaSiteUserInfo Yema站点用户信息解析�?type TYemaSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewTYemaSiteUserInfo 创建Yema站点用户信息解析器实�?func NewTYemaSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *TYemaSiteUserInfo {

	parser := &TYemaSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面地址
func (y *TYemaSiteUserInfo) parseSitePage(htmlText string) {
	/*
	 获取站点页面地址
	*/
	y.userTrafficPage = ""
	y.userDetailPage = ""
	y.userBasicPage = "api/consumer/fetchSelfDetail"
	y.userBasicParams = make(map[string]string)
	y.sysMailUnreadPage = ""
	y.userMailUnreadPage = ""
	y.mailUnreadParams = make(map[string]string)
	y.torrentSeedingPage = "/api/userTorrent/fetchSeedTorrentInfo"
	y.torrentSeedingParams = map[string]string{
		// 虽然这个参数是无意义的，但这�?API 必须�?POST
		"status": "seeding",
	}
	y.torrentSeedingHeaders = make(map[string]string)
	y.additionHeaders = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json, text/plain, */*",
	}
}

// parseLoggedIn 判断是否登录成功
func (y *TYemaSiteUserInfo) parseLoggedIn(htmlText string) bool {
	/*
	 判断是否登录成功, 通过判断是否存在用户信息
	 暂时跳过检测，待后续优�?	*/
	return true
}

// parseUserBaseInfo 解析用户基本信息
func (y *TYemaSiteUserInfo) parseUserBaseInfo(htmlText string) {
	/*
	 解析用户基本信息，这里把_parse_user_traffic_info和_parse_user_detail_info合并到这�?	*/
	if htmlText == "" {
		return
	}
	
	var detail map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &detail)
	if err != nil {
		return
	}
	
	if detail["success"] == nil || !detail["success"].(bool) {
		return
	}
	
	userInfo := detail["data"].(map[string]interface{})
	y.userid = userInfo["id"].(string)
	y.username = userInfo["name"].(string)
	y.userLevel = userInfo["level"].(string)
	
	stringUtils := utils.NewStringUtils()
	y.joinAt = stringUtils.UnifyDateTimeStr(userInfo["registerTime"].(string))

	y.upload = int(userInfo["uploadSize"].(float64))
	// 使用 promotionDownloadSize 获取真实下载量（考虑促销因素�?	if _, exists := userInfo["promotionDownloadSize"]; exists {
		y.download = int(userInfo["promotionDownloadSize"].(float64))
	} else {
		y.download = int(userInfo["downloadSize"].(float64))
	}
	
	if y.download == 0 {
		y.ratio = 0
	} else {
		y.ratio = math.Round(float64(y.upload)/float64(y.download)*100) / 100
	}
	
	y.bonus = userInfo["bonus"].(float64)
	y.messageUnread = 0
}

// parseUserTrafficInfo 解析用户流量信息
func (y *TYemaSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	/*
	 解析用户流量信息
	*/
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (y *TYemaSiteUserInfo) parseUserDetailInfo(htmlText string) {
	/*
	 解析用户详细信息
	*/
	// 空实�?}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (y *TYemaSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	/*
	 解析用户做种信息
	*/
	if htmlText == "" {
		return ""
	}
	
	var seedingInfo map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &seedingInfo)
	if err != nil {
		return ""
	}
	
	if seedingInfo["success"] == nil || !seedingInfo["success"].(bool) || seedingInfo["data"] == nil {
		return ""
	}

	torrents := seedingInfo["data"].(map[string]interface{})

	y.seeding += int(torrents["num"].(float64))
	y.seedingSize += int(torrents["fileSize"].(float64))

	// 是否存在下页数据
	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (y *TYemaSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	/*
	 解析未读消息链接，这里直接读出详�?	*/
	// 空实�?	return ""
}

// parseMessageContent 解析消息内容
func (y *TYemaSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	/*
	 解析消息内容
	*/
	// 空实�?	return "", "", ""
}
