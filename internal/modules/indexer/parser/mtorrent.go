package parser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// MTorrentSiteUserInfo MTorrent站点用户信息解析�?type MTorrentSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// MTeamSysRoleList 用户级别字典
var MTeamSysRoleList = map[string]string{
	"1":  "User",
	"2":  "Power User",
	"3":  "Elite User",
	"4":  "Crazy User",
	"5":  "Insane User",
	"6":  "Veteran User",
	"7":  "Extreme User",
	"8":  "Ultimate User",
	"9":  "Nexus Master",
	"10": "VIP",
	"11": "Retiree",
	"12": "Uploader",
	"13": "Moderator",
	"14": "Administrator",
	"15": "Sysop",
	"16": "Staff",
	"17": "Offer memberStaff",
	"18": "Bet memberStaff",
}

// NewMTorrentSiteUserInfo 创建MTorrent站点用户信息解析器实�?func NewMTorrentSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *MTorrentSiteUserInfo {

	parser := &MTorrentSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	// 设置请求模式为apikey
	parser.SiteParserBaseImpl.RequestMode = "apikey"

	// 更换api地址
	baseURL, _ := url.Parse(url)
	domain := baseURL.Host
	// 提取域名部分，去掉端�?	host := strings.Split(domain, ":")[0]
	// 构造api地址
	apiDomain := "api." + host
	if strings.Contains(domain, ":") {
		// 保留端口部分
		parts := strings.Split(domain, ":")
		if len(parts) > 1 {
			apiDomain += ":" + parts[1]
		}
	}
	parser.SiteParserBaseImpl.BaseURL = fmt.Sprintf("%s://%s", baseURL.Scheme, apiDomain)

	// 设置各种页面URL和参�?	parser.SiteParserBaseImpl.UserTrafficPage = ""
	parser.SiteParserBaseImpl.UserDetailPage = ""
	parser.SiteParserBaseImpl.UserBasicPage = "api/member/profile"
	parser.SiteParserBaseImpl.SysMailUnreadPage = ""
	parser.SiteParserBaseImpl.UserMailUnreadPage = "api/msg/search"

	// 设置参数
	parser.SiteParserBaseImpl.MailUnreadParams = map[string]string{
		"keyword":  "",
		"box":      "-2",
		"type":     "pageNumber",
		"pageSize": "100",
	}

	// 设置做种信息页面和参�?	parser.SiteParserBaseImpl.TorrentSeedingPage = "api/member/getUserTorrentList"
	parser.SiteParserBaseImpl.TorrentSeedingHeaders = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json, text/plain, */*",
	}

	// 设置附加请求�?	parser.SiteParserBaseImpl.AdditionHeaders = map[string]string{
		"x-api-key": apikey,
	}

	return parser
}

// parseSitePage 解析站点页面地址
func (m *MTorrentSiteUserInfo) parseSitePage(htmlText string) {
	// 在NewMTorrentSiteUserInfo中已经设置了所有页面地址和参数，这里留空实现
}

// parseLoggedIn 判断是否登录成功
func (m *MTorrentSiteUserInfo) parseLoggedIn(htmlText string) bool {
	// 暂时跳过检测，待后续优�?	return true
}

// parseUserBaseInfo 解析用户基本信息
func (m *MTorrentSiteUserInfo) parseUserBaseInfo(htmlText string) {
	if htmlText == "" {
		return
	}

	var detail map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &detail)
	if err != nil {
		m.logger.Error("解析JSON失败", zap.Error(err))
		return
	}

	if detail["code"] != "0" {
		return
	}

	userInfo := detail["data"].(map[string]interface{})
	m.userid = fmt.Sprintf("%.0f", userInfo["id"].(float64))
	m.username = userInfo["username"].(string)

	// 获取用户等级
	role := "1"
	if userInfo["role"] != nil {
		role = fmt.Sprintf("%.0f", userInfo["role"].(float64))
	}
	if level, exists := MTeamSysRoleList[role]; exists {
		m.userLevel = level
	}

	// 加入时间
	if userInfo["memberStatus"] != nil {
		memberStatus := userInfo["memberStatus"].(map[string]interface{})
		if memberStatus["createdDate"] != nil {
			m.joinAt = memberStatus["createdDate"].(string)
		}
	}

	// 流量信息
	if userInfo["memberCount"] != nil {
		memberCount := userInfo["memberCount"].(map[string]interface{})
		
		if memberCount["uploaded"] != nil {
			uploaded, _ := strconv.ParseFloat(fmt.Sprintf("%v", memberCount["uploaded"]), 64)
			m.upload = int(uploaded)
		}
		
		if memberCount["downloaded"] != nil {
			downloaded, _ := strconv.ParseFloat(fmt.Sprintf("%v", memberCount["downloaded"]), 64)
			m.download = int(downloaded)
		}
		
		if memberCount["shareRate"] != nil {
			m.ratio, _ = strconv.ParseFloat(fmt.Sprintf("%v", memberCount["shareRate"]), 64)
		}
		
		if memberCount["bonus"] != nil {
			m.bonus, _ = strconv.ParseFloat(fmt.Sprintf("%v", memberCount["bonus"]), 64)
		}
	}

	// 设置强制读取消息
	m.messageReadForce = true

	// 设置做种参数
	m.torrentSeedingParams = map[string]string{
		"pageNumber": "1",
		"pageSize":   "200",
		"type":       "SEEDING",
		"userid":     m.userid,
	}
}

// parseUserTrafficInfo 解析用户流量信息
func (m *MTorrentSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实现，信息已在parseUserBaseInfo中解�?}

// parseUserDetailInfo 解析用户详细信息
func (m *MTorrentSiteUserInfo) parseUserDetailInfo(htmlText string) {
	// 空实现，信息已在parseUserBaseInfo中解�?}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (m *MTorrentSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	if htmlText == "" {
		return ""
	}

	var seedingInfo map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &seedingInfo)
	if err != nil {
		m.logger.Error("解析JSON失败", zap.Error(err))
		return ""
	}

	if seedingInfo["code"] != "0" {
		return ""
	}

	data := seedingInfo["data"].(map[string]interface{})
	torrents := data["data"].([]interface{})

	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	for _, info := range torrents {
		torrentInfo := info.(map[string]interface{})["torrent"].(map[string]interface{})
		
		size := 0
		if torrentInfo["size"] != nil {
			size, _ = strconv.Atoi(fmt.Sprintf("%.0f", torrentInfo["size"].(float64)))
		}
		
		seeders := 0
		if torrentInfo["source"] != nil {
			seeders, _ = strconv.Atoi(fmt.Sprintf("%.0f", torrentInfo["source"].(float64)))
		}
		
		pageSeedingSize += size
		pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeders, size})
	}

	m.seeding += len(torrents)
	m.seedingSize += pageSeedingSize
	m.seedingInfo = append(m.seedingInfo, pageSeedingInfo...)

	// 查询总做种数
	seederCount := 0
	result := m.getPageContent(m.joinURL(m.baseURL, "api/tracker/myPeerStatus"), map[string]string{"uid": m.userid}, nil)
	if result != "" {
		var seederInfo map[string]interface{}
		err := json.Unmarshal([]byte(result), &seederInfo)
		if err == nil && seederInfo["data"] != nil {
			data := seederInfo["data"].(map[string]interface{})
			if data["seeder"] != nil {
				seederCount, _ = strconv.Atoi(fmt.Sprintf("%.0f", data["seeder"].(float64)))
			}
		}
	}

	if seederCount == 0 {
		return ""
	}

	if m.seeding >= seederCount {
		return ""
	}

	// 还有下一�?	pageNumber, _ := strconv.Atoi(m.torrentSeedingParams["pageNumber"])
	m.torrentSeedingParams["pageNumber"] = strconv.Itoa(pageNumber + 1)

	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (m *MTorrentSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	if htmlText == "" {
		return ""
	}

	var messagesInfo map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &messagesInfo)
	if err != nil {
		m.logger.Error("解析JSON失败", zap.Error(err))
		return ""
	}

	if messagesInfo["code"] != "0" {
		return ""
	}

	data := messagesInfo["data"].(map[string]interface{})
	messages := data["data"].([]interface{})

	for _, msg := range messages {
		message := msg.(map[string]interface{})
		
		// 检查是否未�?		if unread, ok := message["unread"].(bool); !ok || !unread {
			continue
		}

		head := ""
		date := ""
		content := ""

		if message["title"] != nil {
			head = message["title"].(string)
		}
		if message["createdDate"] != nil {
			date = message["createdDate"].(string)
		}
		if message["context"] != nil {
			content = message["context"].(string)
		}

		if head != "" && date != "" && content != "" {
			m.messageUnreadContents = append(m.messageUnreadContents, []interface{}{head, date, content})
			
			// 设置已读
			if message["id"] != nil {
				msgID := fmt.Sprintf("%.0f", message["id"].(float64))
				m.getPageContent(m.joinURL(m.baseURL, "api/msg/markRead"), map[string]string{"msgId": msgID}, nil)
			}
		}
	}

	// 是否存在下页数据
	return ""
}

// parseMessageContent 解析消息内容
func (m *MTorrentSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	// 空实现，消息内容已在parseMessageUnreadLinks中解�?	return "", "", ""
}
