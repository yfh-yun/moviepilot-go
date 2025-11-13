package parser

import (
	"encoding/json"
	"regexp"
	"strings"

	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// TNodeSiteUserInfo TNode站点用户信息解析�?type TNodeSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewTNodeSiteUserInfo 创建TNode站点用户信息解析器实�?func NewTNodeSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *TNodeSiteUserInfo {

	parser := &TNodeSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面
func (t *TNodeSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = t.prepareHTMLText(htmlText)

	// <meta name="x-csrf-token" content="fd169876a7b4846f3a7a16fcd5cccf8d">
	csrfToken := regexp.MustCompile(`<meta name="x-csrf-token" content="(.+?)">`).FindStringSubmatch(htmlText)
	if len(csrfToken) > 1 {
		t.additionHeaders = map[string]string{"X-CSRF-TOKEN": csrfToken[1]}
		t.userDetailPage = "api/user/getMainInfo"
		t.torrentSeedingPage = "api/user/listTorrentActivity?id=&type=seeding&page=1&size=20000"
	}
}

// parseLoggedIn 判断是否登录成功
func (t *TNodeSiteUserInfo) parseLoggedIn(htmlText string) bool {
	// 判断是否登录成功, 通过判断是否存在用户信息
	// 暂时跳过检测，待后续优�?	return true
}

// parseUserBaseInfo 解析用户基础信息
func (t *TNodeSiteUserInfo) parseUserBaseInfo(htmlText string) {
	t.username = t.userid
}

// parseUserTrafficInfo 解析用户流量信息
func (t *TNodeSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (t *TNodeSiteUserInfo) parseUserDetailInfo(htmlText string) {
	var detail map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &detail)
	if err != nil {
		return
	}
	
	if status, ok := detail["status"]; !ok || status != 200 {
		return
	}

	userInfo := detail["data"].(map[string]interface{})
	if id, ok := userInfo["id"]; ok {
		t.userid = string(rune(int(id.(float64))))
	}
	
	if username, ok := userInfo["username"]; ok {
		t.username = username.(string)
	}
	
	if classInfo, ok := userInfo["class"].(map[string]interface{}); ok {
		if name, ok := classInfo["name"]; ok {
			t.userLevel = name.(string)
		}
	}
	
	if regTime, ok := userInfo["regTime"]; ok {
		t.joinAt = string(rune(int(regTime.(float64))))
		stringUtils := utils.NewStringUtils()
		t.joinAt = stringUtils.UnifyDateTimeStr(t.joinAt)
	}

	if upload, ok := userInfo["upload"]; ok {
		t.upload = int(upload.(float64))
	}
	
	if download, ok := userInfo["download"]; ok {
		t.download = int(download.(float64))
	}
	
	if t.download <= 0 {
		t.ratio = 0
	} else {
		t.ratio = float64(t.upload) / float64(t.download)
		// 保留3位小�?		t.ratio = float64(int(t.ratio*1000)) / 1000
	}
	
	if bonus, ok := userInfo["bonus"]; ok {
		t.bonus = bonus.(float64)
	}

	unreadAdmin := 0
	unreadInbox := 0
	unreadSystem := 0
	
	if unreadAdminVal, ok := userInfo["unreadAdmin"]; ok {
		unreadAdmin = int(unreadAdminVal.(float64))
	}
	
	if unreadInboxVal, ok := userInfo["unreadInbox"]; ok {
		unreadInbox = int(unreadInboxVal.(float64))
	}
	
	if unreadSystemVal, ok := userInfo["unreadSystem"]; ok {
		unreadSystem = int(unreadSystemVal.(float64))
	}
	
	t.messageUnread = unreadAdmin + unreadInbox + unreadSystem
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (t *TNodeSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	var seedingInfo map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &seedingInfo)
	if err != nil {
		return ""
	}
	
	if status, ok := seedingInfo["status"]; !ok || status != 200 {
		return ""
	}

	data := seedingInfo["data"].(map[string]interface{})
	torrents := data["torrents"].([]interface{})

	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)
	
	for _, item := range torrents {
		torrent := item.(map[string]interface{})
		
		size := 0
		if sizeVal, ok := torrent["size"]; ok {
			size = int(sizeVal.(float64))
		}
		
		seeders := 0
		if seedersVal, ok := torrent["seeding"]; ok {
			seeders = int(seedersVal.(float64))
		}

		pageSeedingSize += size
		pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
	}

	t.seeding += len(torrents)
	t.seedingSize += pageSeedingSize
	t.seedingInfo = append(t.seedingInfo, pageSeedingInfo...)

	// 是否存在下页数据
	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (t *TNodeSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	return ""
}

// parseMessageContent 解析消息内容
func (t *TNodeSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	/*
	 系统信息 api/message/listSystem?page=1&size=20
	 收件箱信�?api/message/listInbox?page=1&size=20
	 管理员信�?api/message/listAdmin?page=1&size=20
	*/
	return "", "", ""
}
