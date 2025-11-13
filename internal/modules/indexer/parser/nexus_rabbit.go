package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/modules/indexer"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// NexusRabbitSiteUserInfo NexusRabbit站点用户信息解析�?type NexusRabbitSiteUserInfo struct {
	*indexer.SiteParserBaseImpl
}

// NewNexusRabbitSiteUserInfo 创建NexusRabbit站点用户信息解析器实�?func NewNexusRabbitSiteUserInfo(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *NexusRabbitSiteUserInfo {

	parser := &NexusRabbitSiteUserInfo{
		SiteParserBaseImpl: indexer.NewSiteParserBaseImpl(siteName, url, siteCookie, apikey, token, ua, emulate, proxy),
	}

	// 设置站点模式
	parser.SiteParserBaseImpl.GetSchema().(indexer.SiteSchema)

	return parser
}

// parseSitePage 解析站点页面
func (n *NexusRabbitSiteUserInfo) parseSitePage(htmlText string) {
	htmlText = n.prepareHTMLText(htmlText)

	userDetail := regexp.MustCompile(`user.php\?id=(\d+)`).FindStringSubmatch(htmlText)

	if len(userDetail) < 2 || strings.TrimSpace(userDetail[0]) == "" {
		return
	}

	n.userid = userDetail[1]
	n.userDetailPage = fmt.Sprintf("user.php?id=%s", n.userid)

	n.userTrafficPage = ""

	n.torrentSeedingPage = "api/general"
	n.torrentSeedingParams = map[string]string{
		"page":  "1",
		"limit": "5000000",
		"action": "userTorrentsList",
		"data": fmt.Sprintf(`{"type": "seeding", "id": %s}`, n.userid),
	}
	n.torrentSeedingHeaders = map[string]string{
		"Content-Type":   "application/json",
		"Accept":         "application/json, text/plain, */*",
		"X-Requested-With": "XMLHttpRequest", // 必须要加上这一条，不然返回的是空数�?	}

	n.userMailUnreadPage = ""
	n.sysMailUnreadPage = "api/general"
	n.mailUnreadParams = map[string]string{
		"page":  "1",
		"limit": "5000000",
		"action": "getMessageIn",
	}
	n.mailUnreadHeaders = map[string]string{
		"Content-Type":   "application/json",
		"Accept":         "application/json, text/plain, */*",
		"X-Requested-With": "XMLHttpRequest",
	}
}

// parseUserTorrentSeedingInfo 解析用户做种信息
func (n *NexusRabbitSiteUserInfo) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	var torrents []map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &torrents)
	if err != nil {
		n.logger.Error("解析做种信息失败", zap.Error(err))
		return ""
	}

	seedingSize := 0
	seedingInfo := make([]interface{}, 0)

	stringUtils := utils.NewStringUtils()
	for _, torrent := range torrents {
		seeders := 0
		if s, ok := torrent["seeders"]; ok {
			if si, ok := s.(float64); ok {
				seeders = int(si)
			}
		}

		size := 0
		if s, ok := torrent["size"]; ok {
			if ss, ok := s.(string); ok {
				size = stringUtils.NumFilesize(ss)
			}
		}

		seedingSize += size
		seedingInfo = append(seedingInfo, []interface{}{seeders, size})
	}

	n.seeding = len(torrents)
	n.seedingSize = seedingSize
	n.seedingInfo = seedingInfo

	return ""
}

// parseMessageUnreadLinks 解析未读消息链接
func (n *NexusRabbitSiteUserInfo) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	unreadIDs := make([]interface{}, 0)
	
	var messages []map[string]interface{}
	err := json.Unmarshal([]byte(htmlText), &messages)
	if err != nil {
		n.logger.Error("解析未读消息失败", zap.Error(err))
		return ""
	}

	stringUtils := utils.NewStringUtils()
	for _, msg := range messages {
		var msgID, msgUnread interface{}
		var hasID, hasUnread bool
		
		if msgID, hasID = msg["id"]; !hasID {
			continue
		}
		
		if msgUnread, hasUnread = msg["unread"]; !hasUnread {
			continue
		}
		
		if unread, ok := msgUnread.(string); ok && unread == "no" {
			continue
		}
		
		unreadIDs = append(unreadIDs, msgID)
		
		head, date, content := "", "", ""
		if h, ok := msg["subject"]; ok {
			if hs, ok := h.(string); ok {
				head = hs
			}
		}
		
		if d, ok := msg["added"]; ok {
			if ds, ok := d.(string); ok {
				date = ds
			}
		}
		
		if c, ok := msg["msg"]; ok {
			if cs, ok := c.(string); ok {
				content = cs
			}
		}
		
		if head != "" && date != "" && content != "" {
			n.messageUnreadContents = append(n.messageUnreadContents, []interface{}{head, date, content})
		}
	}
	
	n.messageUnread = len(unreadIDs)
	if len(unreadIDs) > 0 {
		// 标记消息为已�?		n.getPageContent(
			n.joinURL(n.baseURL, "api/general?loading=true"),
			map[string]string{
				"action": "readMessage",
				"data": fmt.Sprintf(`{"ids": %v}`, unreadIDs),
			},
			map[string]string{
				"Content-Type":   "application/json",
				"Accept":         "application/json, text/plain, */*",
				"X-Requested-With": "XMLHttpRequest",
			},
		)
	}
	
	return ""
}

// parseUserBaseInfo 解析用户基础信息
func (n *NexusRabbitSiteUserInfo) parseUserBaseInfo(htmlText string) {
	// 只有奶糖余额才需要在 base 中获取，其它均可以在详情页拿�?	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	bonusNodes := htmlquery.Find(doc, `//div[contains(text(), "奶糖余额")]/following-sibling::div[1]/text()`)
	if len(bonusNodes) > 0 {
		n.bonus = stringUtils.StrFloat(strings.TrimSpace(htmlquery.InnerText(bonusNodes[0])))
	}
}

// parseUserDetailInfo 解析用户详细信息
func (n *NexusRabbitSiteUserInfo) parseUserDetailInfo(htmlText string) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		n.logger.Error("解析HTML失败", zap.Error(err))
		return
	}

	stringUtils := utils.NewStringUtils()
	if !stringUtils.IsValidHTMLElement(htmlText) {
		return
	}

	// 缩小一下查找范围，所有的信息都在这个 div �?	userInfoNodes := htmlquery.Find(doc, `//div[contains(@class, "layui-hares-user-info-right")]`)
	if len(userInfoNodes) == 0 {
		return
	}
	userInfo := userInfoNodes[0]

	// 用户�?	usernameNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "用户�?)]/a/span/text()`)
	if len(usernameNodes) > 0 {
		n.username = strings.TrimSpace(htmlquery.InnerText(usernameNodes[0]))
	}

	// 等级
	userLevelNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "等级")]/b/text()`)
	if len(userLevelNodes) > 0 {
		n.userLevel = strings.TrimSpace(htmlquery.InnerText(userLevelNodes[0]))
	}

	// 加入日期
	joinDateNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "注册日期")]/text()`)
	if len(joinDateNodes) > 0 {
		joinDate := strings.TrimSpace(htmlquery.InnerText(joinDateNodes[0]))
		parts := strings.Split(joinDate, "\r")
		if len(parts) > 0 {
			joinDate = strings.TrimPrefix(parts[0], "注册日期�?)
			n.joinAt = stringUtils.UnifyDateTimeStr(joinDate)
		}
	}

	// 上传�?	uploadNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "上传�?)]/text()`)
	if len(uploadNodes) > 0 {
		uploadText := strings.TrimSpace(htmlquery.InnerText(uploadNodes[0]))
		uploadText = strings.TrimPrefix(uploadText, "上传量：")
		n.upload = stringUtils.NumFilesize(uploadText)
	}

	// 下载�?	downloadNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "下载�?)]/text()`)
	if len(downloadNodes) > 0 {
		downloadText := strings.TrimSpace(htmlquery.InnerText(downloadNodes[0]))
		downloadText = strings.TrimPrefix(downloadText, "下载量：")
		n.download = stringUtils.NumFilesize(downloadText)
	}

	// 分享�?	ratioNodes := htmlquery.Find(userInfo, `.//span[contains(text(), "分享�?)]/em/text()`)
	if len(ratioNodes) > 0 {
		n.ratio = stringUtils.StrFloat(strings.TrimSpace(htmlquery.InnerText(ratioNodes[0])))
	}
}

// parseMessageContent 解析短消息内�?func (n *NexusRabbitSiteUserInfo) parseMessageContent(htmlText string) (string, string, string) {
	// 解析短消息内容，已经�?parseMessageUnreadLinks 内实现，这里留空防止接口报错
	return "", "", ""
}

// parseUserTrafficInfo 解析用户流量信息
func (n *NexusRabbitSiteUserInfo) parseUserTrafficInfo(htmlText string) {
	// 解析用户的上传，下载，分享率等信息，已经�?parseUserDetailInfo 内实现，这里留空防止接口报错
}
