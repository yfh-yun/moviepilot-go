package indexer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
	"go.uber.org/zap"
)

// SiteSchema 站点框架枚举
type SiteSchema string

const (
	DiscuzX         SiteSchema = "DiscuzX"
	Gazelle         SiteSchema = "Gazelle"
	Ipt             SiteSchema = "IPTorrents"
	NexusPhp        SiteSchema = "NexusPhp"
	NexusProject    SiteSchema = "NexusProject"
	NexusRabbit     SiteSchema = "NexusRabbit"
	NexusHhanclub   SiteSchema = "NexusHhanclub"
	NexusAudiences  SiteSchema = "NexusAudiences"
	SmallHorse      SiteSchema = "Small Horse"
	Unit3d          SiteSchema = "Unit3d"
	TorrentLeech    SiteSchema = "TorrentLeech"
	FileList        SiteSchema = "FileList"
	TNode           SiteSchema = "TNode"
	MTorrent        SiteSchema = "MTorrent"
	Yema            SiteSchema = "Yema"
	HDDolby         SiteSchema = "HDDolby"
)

// SiteParserBaseImpl 站点解析器基础实现
type SiteParserBaseImpl struct {
	// 站点模版
	schema SiteSchema

	// 请求模式 cookie/apikey
	requestMode string

	// 站点信息
	apikey              string
	token               string
	siteName            string
	siteURL             string
	siteDomain          string
	baseURL             string
	siteCookie          string
	session             *http.Client
	ua                  string
	emulate             bool
	proxy               bool
	indexHTML           string

	// 用户信息
	username     string
	userid       string
	userLevel    string
	joinAt       string
	bonus        float64

	// 流量信息
	upload   int
	download int
	ratio    float64

	// 做种信息
	seeding       int
	leeching      int
	seedingSize   int
	leechingSize  int
	uploaded      int
	completed     int
	incomplete    int
	uploadedSize  int
	completedSize int
	incompleteSize int

	// 做种人数, 种子大小
	seedingInfo []interface{}

	// 未读消息
	messageUnread          int
	messageUnreadContents  []interface{}
	messageReadForce       bool

	// 全局附加请求�?	additionHeaders map[string]string

	// 用户基础信息页面
	userBasicPage     string
	userBasicParams   map[string]string
	userBasicHeaders  map[string]string

	// 用户详情信息页面
	userDetailPage    string
	userDetailParams  map[string]string
	userDetailHeaders map[string]string

	// 用户流量信息页面
	userTrafficPage    string
	userTrafficParams  map[string]string
	userTrafficHeaders map[string]string

	// 用户未读消息页面
	userMailUnreadPage string
	sysMailUnreadPage  string

	// 未读消息数参�?	mailUnreadParams   map[string]string
	mailUnreadHeaders  map[string]string
	mailContentParams  map[string]string
	mailContentHeaders map[string]string

	// 用户做种信息页面
	torrentSeedingPage    string
	torrentSeedingParams  map[string]string
	torrentSeedingHeaders map[string]string

	// 错误信息
	errMsg string

	// 日志记录�?	logger *logger.Logger
}

// NewSiteParserBaseImpl 创建站点解析器基础实现
func NewSiteParserBaseImpl(siteName string, url string, siteCookie string, apikey string, token string,
	ua string, emulate bool, proxy bool) *SiteParserBaseImpl {

	log, _ := logger.NewLogger()

	parser := &SiteParserBaseImpl{
		requestMode: "cookie",
		apikey:      apikey,
		token:       token,
		siteName:    siteName,
		siteURL:     url,
		siteCookie:  siteCookie,
		ua:          ua,
		emulate:     emulate,
		proxy:       proxy,
		logger:      log,

		// 初始化默认�?		userDetailPage:      "userdetails.php?id=",
		userTrafficPage:     "index.php",
		userMailUnreadPage:  "messages.php?action=viewmailbox&box=1&unread=yes",
		sysMailUnreadPage:   "messages.php?action=viewmailbox&box=-2&unread=yes",
		torrentSeedingPage:  "getusertorrentlistajax.php?userid=",

		seedingInfo:             make([]interface{}, 0),
		messageUnreadContents:   make([]interface{}, 0),
		additionHeaders:         make(map[string]string),
		userBasicParams:         make(map[string]string),
		userBasicHeaders:        make(map[string]string),
		userDetailParams:        make(map[string]string),
		userDetailHeaders:       make(map[string]string),
		userTrafficParams:       make(map[string]string),
		userTrafficHeaders:      make(map[string]string),
		mailUnreadParams:        make(map[string]string),
		mailUnreadHeaders:       make(map[string]string),
		mailContentParams:       make(map[string]string),
		mailContentHeaders:      make(map[string]string),
		torrentSeedingParams:    make(map[string]string),
		torrentSeedingHeaders:   make(map[string]string),
	}

	// 解析域名
	if parsedURL, err := url.Parse(url); err == nil {
		parser.siteDomain = parsedURL.Host
		parser.baseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	}

	return parser
}

// Parse 解析站点信息
func (sp *SiteParserBaseImpl) Parse() error {
	defer sp.close()

	// Cookie模式时，获取站点首页html
	if sp.requestMode == "apikey" {
		if sp.apikey == "" && sp.token == "" {
			sp.logger.Warn(fmt.Sprintf("%s 未设置cookie �?apikey/token，跳过后续操�?, sp.siteName))
			return nil
		}
		sp.indexHTML = ""
	} else {
		// 检查是否已经登�?		sp.indexHTML = sp.getPageContent(sp.siteURL, nil, nil)
		if !sp.parseLoggedIn(sp.indexHTML) {
			return nil
		}
	}

	// 解析站点页面
	sp.parseSitePage(sp.indexHTML)

	// 解析用户基础信息
	if sp.userBasicPage != "" {
		content := sp.getPageContent(sp.joinURL(sp.baseURL, sp.userBasicPage), sp.userBasicParams, sp.userBasicHeaders)
		sp.parseUserBaseInfo(content)
	} else {
		sp.parseUserBaseInfo(sp.indexHTML)
	}

	// 解析用户详细信息
	if sp.userDetailPage != "" {
		content := sp.getPageContent(sp.joinURL(sp.baseURL, sp.userDetailPage), sp.userDetailParams, sp.userDetailHeaders)
		sp.parseUserDetailInfo(content)
	}

	// 解析用户未读消息
	// TODO: 检查settings.SITE_MESSAGE配置
	sp.parseUnreadMsgs()

	// 解析用户上传、下载、分享率等信�?	if sp.userTrafficPage != "" {
		content := sp.getPageContent(sp.joinURL(sp.baseURL, sp.userTrafficPage), sp.userTrafficParams, sp.userTrafficHeaders)
		sp.parseUserTrafficInfo(content)
	}

	// 解析用户做种信息
	sp.parseSeedingPages()

	return nil
}

// parseUnreadMsgs 解析所有未读消息标题和内容
func (sp *SiteParserBaseImpl) parseUnreadMsgs() {
	unreadMsgLinks := make([]string, 0)

	if sp.messageUnread > 0 || sp.messageReadForce {
		links := []string{sp.userMailUnreadPage, sp.sysMailUnreadPage}
		for _, link := range links {
			if link == "" {
				continue
			}

			msgLinks := make([]string, 0)
			nextPage := sp.parseMessageUnreadLinks(sp.getPageContent(sp.joinURL(sp.baseURL, link), sp.mailUnreadParams, sp.mailUnreadHeaders), msgLinks)
			for nextPage != "" {
				nextPage = sp.parseMessageUnreadLinks(sp.getPageContent(sp.joinURL(sp.baseURL, nextPage), sp.mailUnreadParams, sp.mailUnreadHeaders), msgLinks)
			}
			unreadMsgLinks = append(unreadMsgLinks, msgLinks...)
		}
	}

	// 重新更新未读消息数（99999表示有消息但数量未知�?	if len(unreadMsgLinks) > 0 && sp.messageUnread == 0 {
		sp.messageUnread = len(unreadMsgLinks)
	}

	// 解析未读消息内容
	for _, msgLink := range unreadMsgLinks {
		sp.logger.Debug(fmt.Sprintf("%s 信息链接 %s", sp.siteName, msgLink))
		head, date, content := sp.parseMessageContent(sp.getPageContent(sp.joinURL(sp.baseURL, msgLink), sp.mailContentParams, sp.mailContentHeaders))
		sp.logger.Debug(fmt.Sprintf("%s 标题 %s 时间 %s 内容 %s", sp.siteName, head, date, content))
		sp.messageUnreadContents = append(sp.messageUnreadContents, []interface{}{head, date, content})
	}
}

// parseSeedingPages 解析做种页面
func (sp *SiteParserBaseImpl) parseSeedingPages() {
	if sp.torrentSeedingPage != "" {
		// 第一�?		nextPage := sp.parseUserTorrentSeedingInfo(
			sp.getPageContent(sp.joinURL(sp.baseURL, sp.torrentSeedingPage), sp.torrentSeedingParams, sp.torrentSeedingHeaders),
			false)

		// 其他页处�?		for nextPage != "" && nextPage != "false" {
			fullURL := sp.joinURL(sp.baseURL, sp.torrentSeedingPage)
			nextFullURL := sp.joinURL(fullURL, nextPage)
			nextPage = sp.parseUserTorrentSeedingInfo(
				sp.getPageContent(nextFullURL, sp.torrentSeedingParams, sp.torrentSeedingHeaders),
				true)
		}
	}
}

// prepareHTMLText 处理掉HTML中的干扰部分
func (sp *SiteParserBaseImpl) prepareHTMLText(htmlText string) string {
	// 处理掉HTML中的干扰部分
	re1 := regexp.MustCompile(`#\d+`)
	result := re1.ReplaceAllString(htmlText, "")

	re2 := regexp.MustCompile(`\d+px`)
	result = re2.ReplaceAllString(result, "")

	return result
}

// getPageContent 获取页面内容
func (sp *SiteParserBaseImpl) getPageContent(url string, params map[string]string, headers map[string]string) string {
	var reqHeaders map[string]string
	var proxies interface{} // TODO: 实现代理设置

	if sp.ua != "" || len(headers) > 0 || len(sp.additionHeaders) > 0 {
		if sp.requestMode == "apikey" {
			reqHeaders = make(map[string]string)
		} else {
			reqHeaders = map[string]string{
				"User-Agent": sp.ua,
			}
		}

		if len(headers) > 0 {
			for k, v := range headers {
				reqHeaders[k] = v
			}
		} else {
			reqHeaders["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
		}

		if len(sp.additionHeaders) > 0 {
			for k, v := range sp.additionHeaders {
				reqHeaders[k] = v
			}
		}
	}

	var cookie string
	var session *http.Client

	if sp.requestMode == "apikey" {
		// 使用apikey请求，通过请求头传�?		cookie = ""
		session = nil
	} else {
		// 使用cookie请求
		cookie = sp.siteCookie
		session = sp.session
	}

	// 创建HTTP客户�?	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	var res *http.Response
	var err error

	if len(params) > 0 {
		if contentType, ok := reqHeaders["Content-Type"]; ok && contentType == "application/json" {
			// JSON请求
			jsonData, _ := json.Marshal(params)
			req, _ := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
			for k, v := range reqHeaders {
				req.Header.Set(k, v)
			}
			res, err = client.Do(req)
		} else {
			// 表单请求
			formData := url.Values{}
			for k, v := range params {
				formData.Set(k, v)
			}
			req, _ := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
			for k, v := range reqHeaders {
				req.Header.Set(k, v)
			}
			res, err = client.Do(req)
		}
	} else {
		// GET请求
		req, _ := http.NewRequest("GET", url, nil)
		for k, v := range reqHeaders {
			req.Header.Set(k, v)
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		res, err = client.Do(req)
	}

	if err != nil {
		sp.logger.Error(fmt.Sprintf("%s 请求失败: %v", sp.siteName, err))
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode == 200 || res.StatusCode == 500 || res.StatusCode == 403 {
		if accept, ok := reqHeaders["Accept"]; ok && strings.Contains(accept, "application/json") {
			var result map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
				sp.logger.Error(fmt.Sprintf("%s API响应JSON解析失败: %v", sp.siteName, err))
				return ""
			}
			jsonStr, _ := json.Marshal(result)
			return string(jsonStr)
		} else {
			// 读取响应内容
			buf := make([]byte, 1024*1024) // 1MB缓冲�?			n, _ := res.Body.Read(buf)
			content := string(buf[:n])

			// 如果cloudflare 有防护，尝试使用浏览器仿�?			// TODO: 实现under_challenge检�?			/*
			if underChallenge(content) {
				sp.logger.Warn(fmt.Sprintf("%s 检测到Cloudflare，请更新Cookie和UA", sp.siteName))
				return ""
			}
			*/

			// TODO: 实现RequestUtils.get_decoded_html_content
			return content
		}
	}

	return ""
}

// joinURL 连接URL
func (sp *SiteParserBaseImpl) joinURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return base + ref
	}

	return baseURL.ResolveReference(refURL).String()
}

// parseLoggedIn 解析用户是否已经登陆
func (sp *SiteParserBaseImpl) parseLoggedIn(htmlText string) bool {
	siteUtils := utils.NewSiteUtils()
	loggedIn := siteUtils.IsLoggedIn(htmlText)
	if !loggedIn {
		sp.errMsg = "未检测到已登陆，请检查cookies是否过期"
		sp.logger.Warn(fmt.Sprintf("%s 未登录，跳过后续操作", sp.siteName))
	}
	return loggedIn
}

// Clear 清理资源
func (sp *SiteParserBaseImpl) Clear() {
	sp.indexHTML = ""
	sp.seedingInfo = make([]interface{}, 0)
	sp.messageUnreadContents = make([]interface{}, 0)
}

// close 关闭会话
func (sp *SiteParserBaseImpl) close() {
	// 在Go中不需要显式关闭HTTP客户�?	sp.session = nil
}

// GetUserID 获取用户ID
func (sp *SiteParserBaseImpl) GetUserID() string {
	return sp.userid
}

// GetUsername 获取用户�?func (sp *SiteParserBaseImpl) GetUsername() string {
	return sp.username
}

// GetUserLevel 获取用户等级
func (sp *SiteParserBaseImpl) GetUserLevel() string {
	return sp.userLevel
}

// GetJoinAt 获取加入时间
func (sp *SiteParserBaseImpl) GetJoinAt() string {
	return sp.joinAt
}

// GetUpload 获取上传�?func (sp *SiteParserBaseImpl) GetUpload() int {
	return sp.upload
}

// GetDownload 获取下载�?func (sp *SiteParserBaseImpl) GetDownload() int {
	return sp.download
}

// GetRatio 获取分享�?func (sp *SiteParserBaseImpl) GetRatio() float64 {
	return sp.ratio
}

// GetBonus 获取积分
func (sp *SiteParserBaseImpl) GetBonus() float64 {
	return sp.bonus
}

// GetSeeding 获取做种�?func (sp *SiteParserBaseImpl) GetSeeding() int {
	return sp.seeding
}

// GetLeeching 获取下载�?func (sp *SiteParserBaseImpl) GetLeeching() int {
	return sp.leeching
}

// GetSeedingSize 获取做种体积
func (sp *SiteParserBaseImpl) GetSeedingSize() int {
	return sp.seedingSize
}

// GetLeechingSize 获取下载体积
func (sp *SiteParserBaseImpl) GetLeechingSize() int {
	return sp.leechingSize
}

// GetSeedingInfo 获取做种信息
func (sp *SiteParserBaseImpl) GetSeedingInfo() []interface{} {
	return sp.seedingInfo
}

// GetMessageUnread 获取未读消息�?func (sp *SiteParserBaseImpl) GetMessageUnread() int {
	return sp.messageUnread
}

// GetMessageUnreadContents 获取未读消息内容
func (sp *SiteParserBaseImpl) GetMessageUnreadContents() []interface{} {
	return sp.messageUnreadContents
}

// GetErrMsg 获取错误信息
func (sp *SiteParserBaseImpl) GetErrMsg() string {
	return sp.errMsg
}

// GetSchema 获取站点解析器模�?func (sp *SiteParserBaseImpl) GetSchema() interface{} {
	return sp.schema
}

// SiteSchema 获取站点解析模型
func (sp *SiteParserBaseImpl) SiteSchema() SiteSchema {
	return sp.schema
}

// 以下为需要子类实现的抽象方法占位�?
// parseMessageUnreadLinks 获取未阅读消息链�?func (sp *SiteParserBaseImpl) parseMessageUnreadLinks(htmlText string, msgLinks []string) string {
	// 需要子类实�?	return ""
}

// parseSitePage 解析站点相关信息页面
func (sp *SiteParserBaseImpl) parseSitePage(htmlText string) {
	// 需要子类实�?}

// parseUserBaseInfo 解析用户基础信息
func (sp *SiteParserBaseImpl) parseUserBaseInfo(htmlText string) {
	// 需要子类实�?}

// parseUserTrafficInfo 解析用户的上传，下载，分享率等信�?func (sp *SiteParserBaseImpl) parseUserTrafficInfo(htmlText string) {
	// 需要子类实�?}

// parseUserTorrentSeedingInfo 解析用户的做种相关信�?func (sp *SiteParserBaseImpl) parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string {
	// 需要子类实�?	return ""
}

// parseUserDetailInfo 解析用户的详细信�?func (sp *SiteParserBaseImpl) parseUserDetailInfo(htmlText string) {
	// 需要子类实�?}

// parseMessageContent 解析短消息内�?func (sp *SiteParserBaseImpl) parseMessageContent(htmlText string) (string, string, string) {
	// 需要子类实�?	return "", "", ""
}
