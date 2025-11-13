package helper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// RssHelper RSS帮助类，解析RSS报文、获取RSS地址�?type RssHelper struct {
	// RSS解析限制配置
	MaxRssSize  int64 // 最大RSS文件大小
	MaxRssItems int   // 最大解析条目数

	// 各站点RSS链接获取配置
	rssLinkConf map[string]SiteConfig
}

// SiteConfig 站点配置
type SiteConfig struct {
	XPath   string            `json:"xpath"`
	URL     string            `json:"url"`
	Params  map[string]string `json:"params"`
	Render  bool              `json:"render,omitempty"`
}

// RssItem RSS条目
type RssItem struct {
	Title       string      `json:"title"`
	Enclosure   string      `json:"enclosure"`
	Size        int64       `json:"size"`
	Description string      `json:"description"`
	Link        string      `json:"link"`
	PubDate     time.Time   `json:"pubdate"`
	Nickname    string      `json:"nickname,omitempty"`
}

// NewRssHelper 创建RssHelper实例
func NewRssHelper() *RssHelper {
	helper := &RssHelper{
		MaxRssSize:  50 * 1024 * 1024, // 50MB最大RSS文件大小
		MaxRssItems: 1000,             // 最大解析条目数
		rssLinkConf: map[string]SiteConfig{
			"default": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked":   "0",
					"itemsmalldescr":   "1",
					"showrows":         "50",
					"search_mode":      "1",
				},
			},
			"hares.top": {
				XPath: "//*[@id='layui-layer100001']/div[2]/div/p[4]/a/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"et8.org": {
				XPath: "//*[@id='outer']/table/tbody/tr/td/table/tbody/tr/td/a[2]/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"pttime.org": {
				XPath: "//*[@id='outer']/table/tbody/tr/td/table/tbody/tr/td/text()[5]",
				URL:   "getrss.php",
				Params: map[string]string{
					"showrows":       "10",
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
				},
			},
			"ourbits.club": {
				XPath: "//a[@class='gen_rsslink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"totheglory.im": {
				XPath: "//textarea/text()",
				URL:   "rsstools.php?c51=51&c52=52&c53=53&c54=54&c108=108&c109=109&c62=62&c63=63&c67=67&c69=69&c70=70&c73=73&c76=76&c75=75&c74=74&c87=87&c88=88&c99=99&c90=90&c58=58&c103=103&c101=101&c60=60",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"monikadesign.uk": {
				XPath: "//a/@href",
				URL:   "rss",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"zhuque.in": {
				XPath:  "//a/@href",
				URL:    "user/rss",
				Render: true,
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
				},
			},
			"hdchina.org": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"rsscart":        "0",
				},
			},
			"audiences.me": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"torrent_type":   "1",
					"exp":            "180",
				},
			},
			"shadowflow.org": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"paid":           "0",
					"search_mode":    "0",
					"showrows":       "30",
				},
			},
			"hddolby.com": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"exp":            "180",
				},
			},
			"hdhome.org": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"exp":            "180",
				},
			},
			"pthome.net": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"exp":            "180",
				},
			},
			"ptsbao.club": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "1",
					"size":           "0",
				},
			},
			"leaves.red": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "0",
					"paid":           "2",
				},
			},
			"hdtime.org": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"showrows":       "50",
					"search_mode":    "0",
				},
			},
			"m-team.io": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"showrows":       "50",
					"inclbookmarked": "0",
					"itemsmalldescr": "1",
					"https":          "1",
				},
			},
			"u2.dmhy.org": {
				XPath: "//a[@class='faqlink']/@href",
				URL:   "getrss.php",
				Params: map[string]string{
					"inclbookmarked":    "0",
					"itemsmalldescr":    "1",
					"showrows":          "50",
					"search_mode":       "1",
					"inclautochecked":   "1",
					"trackerssl":        "1",
				},
			},
		},
	}

	return helper
}

// Parse 解析RSS订阅URL，获取RSS中的种子信息
// url: RSS地址
// proxy: 是否使用代理
// timeout: 请求超时(�?
// headers: 自定义请求头
// 返回: 种子信息列表，如为nil代表Rss过期，如果为error则为错误
func (r *RssHelper) Parse(urlStr string, proxy bool, timeout int, headers map[string]string) ([]RssItem, *bool, error) {
	// 开始处�?	retArray := make([]RssItem, 0)
	
	if urlStr == "" {
		result := false
		return nil, &result, nil
	}

	// 创建HTTP客户�?	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	// 设置代理
	if proxy {
		// 注意：在实际实现中需要根据配置设置代�?		// proxyURL, _ := url.Parse(config.GetConfig().PROXY)
		// client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	// 创建请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.Errorf("获取RSS失败：请求创建失败，URL: %s, 错误: %v", urlStr, err)
		result := false
		return nil, &result, err
	}

	// 设置请求�?	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 发送请�?	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("获取RSS失败：请求发送失败，URL: %s, 错误: %v", urlStr, err)
		result := false
		return nil, &result, err
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		logger.Errorf("RSS请求失败，状态码: %d, URL: %s", resp.StatusCode, urlStr)
		result := false
		return nil, &result, nil
	}

	// 读取响应内容
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("读取RSS响应失败，URL: %s, 错误: %v", urlStr, err)
		result := false
		return nil, &result, err
	}

	// 检查响应大小，避免处理过大的RSS文件
	if int64(len(rawData)) > r.MaxRssSize {
		logger.Warningf("RSS文件过大: %.1fMB，跳过解�?, float64(len(rawData))/1024/1024)
		result := false
		return nil, &result, nil
	}

	// 尝试检测编码并解码为字符串
	retXml := string(rawData)
	// 在Go中，我们通常假设XML是UTF-8编码，除非另有说�?
	// 验证RSS内容是否有效
	if retXml == "" || strings.TrimSpace(retXml) == "" {
		logger.Error("RSS内容为空")
		result := false
		return nil, &result, nil
	}

	// 检查是否包含基本的RSS/XML结构
	retXmlStripped := strings.TrimSpace(retXml)
	if !strings.HasPrefix(retXmlStripped, "<") {
		logger.Error("RSS内容不是有效的XML格式")
		result := false
		return nil, &result, nil
	}

	// RSS过期检�?	rssExpiredMsg := []string{
		"RSS 链接已过�? 您需要获得一个新�?",
		"RSS Link has expired, You need to get a new one!",
		"RSS Link has expired, You need to get new!",
	}
	
	for _, msg := range rssExpiredMsg {
		if strings.Contains(retXml, msg) {
			// RSS已过期，返回nil表示过期
			return nil, nil, nil
		}
	}

	// 解析XML
	doc, err := xmlquery.Parse(strings.NewReader(retXml))
	if err != nil {
		logger.Debugf("XML解析失败�?v，尝试HTML解析", err)
		// 如果XML解析失败，尝试作为HTML解析
		htmlDoc, err := htmlquery.Parse(strings.NewReader(retXml))
		if err != nil {
			logger.Errorf("HTML解析也失败：%v", err)
			result := false
			return nil, &result, err
		}
		
		// 查找RSS根节�?		rssNodes := htmlquery.Find(htmlDoc, "//rss | //feed")
		if len(rssNodes) > 0 {
			doc = rssNodes[0]
		} else {
			logger.Error("无法解析RSS内容")
			result := false
			return nil, &result, nil
		}
	}

	// 查找所有item或entry节点
	items := xmlquery.Find(doc, ".//item | .//entry")

	// 限制处理的条目数�?	itemsCount := len(items)
	if itemsCount > r.MaxRssItems {
		logger.Warningf("RSS条目过多: %d，仅处理�?d�?, itemsCount, r.MaxRssItems)
		itemsCount = r.MaxRssItems
	}

	// 解析每个条目
	for i := 0; i < itemsCount; i++ {
		item := items[i]
		
		// 标题
		titleNodes := xmlquery.Find(item, ".//title")
		title := ""
		if len(titleNodes) > 0 && titleNodes[0].InnerText() != "" {
			title = titleNodes[0].InnerText()
		}
		if title == "" {
			continue
		}

		// 描述
		descNodes := xmlquery.Find(item, ".//description | .//summary")
		description := ""
		if len(descNodes) > 0 && descNodes[0].InnerText() != "" {
			description = descNodes[0].InnerText()
		}

		// 种子页面链接
		linkNodes := xmlquery.Find(item, ".//link")
		link := ""
		if len(linkNodes) > 0 {
			if linkNodes[0].InnerText() != "" {
				link = linkNodes[0].InnerText()
			} else {
				link = htmlquery.SelectAttr(linkNodes[0], "href")
			}
		}

		// 种子链接
		enclosureNodes := xmlquery.Find(item, ".//enclosure")
		enclosure := ""
		if len(enclosureNodes) > 0 {
			enclosure = htmlquery.SelectAttr(enclosureNodes[0], "url")
		}
		if enclosure == "" && link == "" {
			continue
		}
		// 部分RSS只有link没有enclosure
		if enclosure == "" && link != "" {
			enclosure = link
		}

		// 大小
		var size int64 = 0
		if len(enclosureNodes) > 0 {
			sizeAttr := htmlquery.SelectAttr(enclosureNodes[0], "length")
			if sizeAttr != "" {
				if s, err := strconv.ParseInt(sizeAttr, 10, 64); err == nil {
					size = s
				}
			}
		}

		// 发布日期
		pubdateNodes := xmlquery.Find(item, ".//pubDate | .//published | .//updated")
		var pubdate time.Time
		if len(pubdateNodes) > 0 && pubdateNodes[0].InnerText() != "" {
			// 这里应该调用时间解析工具函数
			pubdate = utils.GetTime(pubdateNodes[0].InnerText())
			// 转为本地时区（Go中默认已经是本地时区�?		}

		// 获取豆瓣昵称
		nicknameNodes := xmlquery.Find(item, ".//*[local-name()='creator']")
		nickname := ""
		if len(nicknameNodes) > 0 && nicknameNodes[0].InnerText() != "" {
			nickname = nicknameNodes[0].InnerText()
		}

		// 构造返回对�?		tmpDict := RssItem{
			Title:       title,
			Enclosure:   enclosure,
			Size:        size,
			Description: description,
			Link:        link,
			PubDate:     pubdate,
		}
		
		// 如果豆瓣昵称不为空，返回数据增加豆瓣昵称
		if nickname != "" {
			tmpDict.Nickname = nickname
		}
		
		retArray = append(retArray, tmpDict)
	}

	return retArray, nil, nil
}

// GetRssLink 获取站点rss地址
// url: 站点地址
// cookie: 站点cookie
// ua: 站点ua
// proxy: 是否使用代理
// timeout: 请求超时时间
// 返回: rss地址、错误信�?func (r *RssHelper) GetRssLink(urlStr, cookie, ua string, proxy bool, timeout int) (string, string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("获取RSS链接时发生未预期错误: %v", err)
		}
	}()

	// 获取站点域名
	domain := utils.GetUrlDomain(urlStr)
	
	// 获取配置
	var siteConf SiteConfig
	if conf, exists := r.rssLinkConf[domain]; exists {
		siteConf = conf
	} else {
		siteConf = r.rssLinkConf["default"]
	}
	
	// RSS地址
	rssURL, err := url.JoinPath(urlStr, siteConf.URL)
	if err != nil {
		return "", fmt.Sprintf("构建RSS链接失败�?v", err)
	}
	
	// RSS请求参数
	rssParams := siteConf.Params
	
	// 请求RSS页面
	var htmlText string
	if siteConf.Render {
		// 注意：PlaywrightHelper在Go中需要单独实现或使用其他方案
		// htmlText = PlaywrightHelper().GetPageSource(...)
		return "", "渲染模式暂未实现"
	} else {
		// 构造POST请求的表单数�?		formData := url.Values{}
		for key, value := range rssParams {
			formData.Set(key, value)
		}
		
		// 创建HTTP客户�?		clientTimeout := 30
		if timeout > 0 {
			clientTimeout = timeout
		}
		
		client := &http.Client{
			Timeout: time.Duration(clientTimeout) * time.Second,
		}
		
		// 创建POST请求
		req, err := http.NewRequest("POST", rssURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return "", fmt.Sprintf("创建请求失败�?v", err)
		}
		
		// 设置请求�?		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if ua != "" {
			req.Header.Set("User-Agent", ua)
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		
		// 设置代理
		if proxy {
			// 注意：在实际实现中需要根据配置设置代�?			// proxyURL, _ := url.Parse(config.GetConfig().PROXY)
			// client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
		
		// 发送请�?		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Sprintf("获取RSS链接失败：无法连�?%s，错�? %v", urlStr, err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			return "", fmt.Sprintf("获取 %s RSS链接失败，错误码�?d，错误原因：%s", urlStr, resp.StatusCode, resp.Status)
		}
		
		// 读取响应内容
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Sprintf("读取响应内容失败�?v", err)
		}
		
		htmlText = string(body)
	}
	
	// 解析HTML
	if htmlText != "" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
		if err != nil {
			return "", fmt.Sprintf("解析HTML失败�?v", err)
		}
		
		// 使用XPath查找RSS链接
		// 注意：goquery不直接支持XPath，需要使用其他库或者转换为CSS选择�?		// 这里简化处理，使用CSS选择�?		
		// 将XPath转换为CSS选择器（这是一个简化的处理方式�?		// 实际项目中可能需要更复杂的XPath到CSS的转�?		cssSelector := convertXPathToCSS(siteConf.XPath)
		
		// 查找元素
		var rssLink string
		doc.Find(cssSelector).Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				rssLink = href
			} else {
				rssLink = s.Text()
			}
		})
		
		if rssLink != "" {
			return rssLink, ""
		}
	}
	
	return "", fmt.Sprintf("获取RSS链接失败�?s", urlStr)
}

// convertXPathToCSS 简单的XPath到CSS选择器转换（实际项目中可能需要更复杂的实现）
func convertXPathToCSS(xpath string) string {
	// 这是一个非常简化的转换，仅适用于部分简单情�?	// 实际项目中应该使用专门的XPath到CSS转换�?	
	// 移除@href等属性选择器部分，只保留元素选择器部�?	if strings.Contains(xpath, "/@") {
		parts := strings.Split(xpath, "/@")
		xpath = parts[0]
	}
	
	// 简单替�?	xpath = strings.ReplaceAll(xpath, "//", "")
	xpath = strings.ReplaceAll(xpath, "/", " ")
	xpath = strings.ReplaceAll(xpath, "[@class=", ".")
	xpath = strings.ReplaceAll(xpath, "]", "")
	xpath = strings.ReplaceAll(xpath, "'", "")
	xpath = strings.ReplaceAll(xpath, "\"", "")
	
	return xpath
}
