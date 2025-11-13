package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// HDDolbySiteUserInfo HDDolby站点用户信息解析�?type HDDolbySiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewHDDolbySiteUserInfo 创建HDDolby站点用户信息解析器实�?func NewHDDolbySiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *HDDolbySiteUserInfo {

	parser := &HDDolbySiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置请求模式为apikey
	parser.requestMode = "apikey"

	return parser
}

// HDDolbyUserResponse 用户信息API响应结构
type HDDolbyUserResponse struct {
	Status int `json:"status"`
	Data   []struct {
		ID              string `json:"id"`
		Added           string `json:"added"`
		LastAccess      string `json:"last_access"`
		Class           string `json:"class"`
		Uploaded        string `json:"uploaded"`
		Downloaded      string `json:"downloaded"`
		Seedbonus       string `json:"seedbonus"`
		Sebonus         string `json:"sebonus"`
		UnreadMessages  string `json:"unread_messages"`
		Username        string `json:"username"`
	} `json:"data"`
}

// HDDolbyPeersResponse 做种信息API响应结构
type HDDolbyPeersResponse struct {
	Status int `json:"status"`
	Data   []struct {
		Size    int64 `json:"size"`
		Seeders int   `json:"seeders"`
	} `json:"data"`
}

// HDDolby用户级别字典
var HDDolbySysRoleList = map[string]string{
	"0":  "Peasant",
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
	"12": "Helper",
	"13": "Seeder",
	"14": "Transferrer",
	"15": "Uploader",
	"16": "Torrent Manager",
	"17": "Forum Moderator",
	"18": "Coder",
	"19": "Moderator",
	"20": "Administrator",
	"21": "Sysop",
	"22": "Staff Leader",
}

// parseSitePage 获取站点页面地址
func (h *HDDolbySiteUserInfo) parseSitePage(htmlText string) {
	// 更换api地址
	stringUtils := utils.NewStringUtils()
	domain := stringUtils.GetURLDomain(h.baseURL)
	h.baseURL = fmt.Sprintf("https://api.%s", domain)
	h.userTrafficPage = ""
	h.userDetailPage = ""
	h.userBasicPage = "api/v1/user/data"
	h.userBasicParams = make(map[string]string)
	h.userBasicHeaders = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json, text/plain, */*",
	}
	h.sysMailUnreadPage = ""
	h.userMailUnreadPage = ""
	h.mailUnreadParams = make(map[string]string)
	h.torrentSeedingPage = "api/v1/user/peers"
	h.torrentSeedingParams = make(map[string]string)
	h.torrentSeedingHeaders = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json, text/plain, */*",
	}
	h.additionHeaders = map[string]string{
		"x-api-key": h.apikey,
	}
}

// parseLoggedIn 判断是否登录成功
func (h *HDDolbySiteUserInfo) parseLoggedIn(htmlText string) bool {
	// 暂时跳过检测，待后续优�?	return true
}

// parseUserBaseInfo 解析用户基本信息
func (h *HDDolbySiteUserInfo) parseUserBaseInfo(htmlText string) {
	// 解析用户基本信息，这里把_parse_user_traffic_info和_parse_user_detail_info合并到这�?	if htmlText == "" {
		return
	}

	var detail HDDolbyUserResponse
	if err := json.Unmarshal([]byte(htmlText), &detail); err != nil {
		h.logger.Error("解析用户信息JSON失败", zap.Error(err))
		return
	}

	if detail.Status != 0 {
		return
	}

	userInfos := detail.Data
	if len(userInfos) == 0 {
		return
	}

	userInfo := userInfos[0]
	h.userid = userInfo.ID
	h.username = userInfo.Username
	if level, exists := HDDolbySysRoleList[userInfo.Class]; exists {
		h.userLevel = level
	} else {
		// 默认为User级别
		h.userLevel = "User"
	}
	h.joinAt = userInfo.Added
	h.upload, _ = strconv.Atoi(userInfo.Uploaded)
	h.download, _ = strconv.Atoi(userInfo.Downloaded)
	if h.download > 0 {
		h.ratio = math.Round(float64(h.upload)/float64(h.download)*100) / 100
	} else {
		h.ratio = 0
	}
	h.bonus, _ = strconv.ParseFloat(userInfo.Seedbonus, 64)
	h.messageUnread, _ = strconv.Atoi(userInfo.UnreadMessages)
}

// parseUserTrafficInfo 解析用户流量信息
func (h *HDDolbySiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 空实�?}

// parseUserDetailInfo 解析用户详细信息
func (h *HDDolbySiteUserInfo) parseUserDetailInfo(htmlText string) {
	// 空实�?}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (h *HDDolbySiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	// 解析用户做种信息
	if htmlText == "" {
		return ""
	}

	var seedingInfo HDDolbyPeersResponse
	if err := json.Unmarshal([]byte(htmlText), &seedingInfo); err != nil {
		h.logger.Error("解析做种信息JSON失败", zap.Error(err))
		return ""
	}

	if seedingInfo.Status != 0 {
		return ""
	}

	torrents := seedingInfo.Data
	pageSeedingSize := 0
	pageSeedingInfo := make([]interface{}, 0)

	for _, info := range torrents {
		size := info.Size
		seeder := info.Seeders
		if seeder == 0 {
			seeder = 1
		}
		pageSeedingSize += int(size)
		pageSeedingInfo = append(pageSeedingInfo, []interface{}{seeder, size})
	}

	h.seeding += len(torrents)
	h.seedingSize += pageSeedingSize
	h.seedingInfo = append(h.seedingInfo, pageSeedingInfo...)

	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (h *HDDolbySiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	// 解析未读消息链接，这里直接读出详�?	return ""
}

// parseMessageContent 解析消息内容
func (h *HDDolbySiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	// 解析消息内容
	return "", "", ""
}
