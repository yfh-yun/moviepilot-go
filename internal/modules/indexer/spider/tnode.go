package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	
	"moviepilot-go/internal/utils"
)

// TNodeSpider TNode API实现
type TNodeSpider struct {
	indexerID int
	domain    string
	searchURL string
	name      string
	proxy     string
	cookie    string
	ua        string
	size      int
	timeout   int
	
	// URL模板
	baseURL     string
	downloadURL string
	pageURL     string
	
	// 缓存token
	token string
}

// TNodeSearchParams 搜索参数结构
type TNodeSearchParams struct {
	Page         int           `json:"page"`
	Size         int           `json:"size"`
	Type         string        `json:"type"`
	Keyword      string        `json:"keyword"`
	Sorter       string        `json:"sorter"`
	Order        string        `json:"order"`
	Tags         []interface{} `json:"tags"`
	Category     []int         `json:"category"`
	Medium       []interface{} `json:"medium"`
	VideoCoding  []interface{} `json:"videoCoding"`
	AudioCoding  []interface{} `json:"audioCoding"`
	Resolution   []interface{} `json:"resolution"`
	Group        []interface{} `json:"group"`
}

// TNodeResult API返回结果结构
type TNodeResult struct {
	Data struct {
		Torrents []TNodeTorrent `json:"torrents"`
	} `json:"data"`
}

// TNodeTorrent 种子信息结构
type TNodeTorrent struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Size        int64  `json:"size"`
	UploadTime  int64  `json:"upload_time"`
	Seeding     int    `json:"seeding"`
	Leeching    int    `json:"leeching"`
	Complete    int    `json:"complete"`
	DownloadRate float64 `json:"downloadRate"`
	UploadRate  float64 `json:"uploadRate"`
	Imdb        string  `json:"imdb"`
}

// NewTNodeSpider 创建TNodeSpider实例
func NewTNodeSpider(indexer map[string]interface{}) *TNodeSpider {
	spider := &TNodeSpider{
		size:        100,
		timeout:     15,
		baseURL:     "%sapi/torrent/advancedSearch",
		downloadURL: "%sapi/torrent/download/%d",
		pageURL:     "%storrent/info/%d",
	}
	
	if indexer != nil {
		if id, ok := indexer["id"].(int); ok {
			spider.indexerID = id
		}
		
		if domain, ok := indexer["domain"].(string); ok {
			spider.domain = domain
			spider.searchURL = fmt.Sprintf(spider.baseURL, domain)
			spider.downloadURL = fmt.Sprintf(spider.downloadURL, domain, "%d")
			spider.pageURL = fmt.Sprintf(spider.pageURL, domain, "%d")
		}
		
		if name, ok := indexer["name"].(string); ok {
			spider.name = name
		}
		
		if proxy, ok := indexer["proxy"].(bool); ok && proxy {
			// TODO: 设置代理
			// spider.proxy = settings.PROXY
		}
		
		if cookie, ok := indexer["cookie"].(string); ok {
			spider.cookie = cookie
		}
		
		if ua, ok := indexer["ua"].(string); ok {
			spider.ua = ua
		}
		
		if timeout, ok := indexer["timeout"].(int); ok {
			spider.timeout = timeout
		} else if timeout, ok := indexer["timeout"].(float64); ok {
			spider.timeout = int(timeout)
		}
	}
	
	return spider
}

// getToken 获取CSRF token
func (t *TNodeSpider) getToken() string {
	// 检查缓存的token
	if t.token != "" {
		return t.token
	}
	
	if t.domain == "" {
		return ""
	}
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(t.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("GET", t.domain, nil)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 获取token失败，无法创建请�? %v", t.name, err))
		fmt.Printf("%s 获取token失败，无法创建请�? %v\n", t.name, err)
		return ""
	}
	
	// 设置请求�?	if t.ua != "" {
		req.Header.Set("User-Agent", t.ua)
	}
	if t.cookie != "" {
		req.Header.Set("Cookie", t.cookie)
	}
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 获取token失败，无法连�?%s: %v", t.name, t.domain, err))
		fmt.Printf("%s 获取token失败，无法连�?%s: %v\n", t.name, t.domain, err)
		return ""
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 读取响应内容
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		content := buf.String()
		
		// 使用正则表达式提取CSRF token
		re := regexp.MustCompile(`<meta name="x-csrf-token" content="(.+?)">`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			t.token = matches[1]
			return t.token
		}
	}
	
	return ""
}

// getParams 获取搜索参数
func (t *TNodeSpider) getParams(keyword string, page int) *TNodeSearchParams {
	// 确定搜索类型
	searchType := "title"
	if strings.HasPrefix(keyword, "tt") {
		searchType = "imdbid"
	}
	
	return &TNodeSearchParams{
		Page:    page + 1,
		Size:    t.size,
		Type:    searchType,
		Keyword: keyword,
		Sorter:  "id",
		Order:   "desc",
		Tags:    []interface{}{},
		Category: []int{501, 502, 503, 504},
		Medium:       []interface{}{},
		VideoCoding:  []interface{}{},
		AudioCoding:  []interface{}{},
		Resolution:   []interface{}{},
		Group:        []interface{}{},
	}
}

// parseResult 解析搜索结果
func (t *TNodeSpider) parseResult(results []TNodeTorrent) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	
	if len(results) == 0 {
		return torrents
	}
	
	for _, result := range results {
		torrent := map[string]interface{}{
			"title":                result.Title,
			"description":          result.Subtitle,
			"enclosure":            fmt.Sprintf(t.downloadURL, result.ID),
			"pubdate":              t.formatTimestamp(result.UploadTime),
			"size":                 result.Size,
			"seeders":              result.Seeding,
			"peers":                result.Leeching,
			"grabs":                result.Complete,
			"downloadvolumefactor": result.DownloadRate,
			"uploadvolumefactor":   result.UploadRate,
			"page_url":             fmt.Sprintf(t.pageURL, result.ID),
			"imdbid":               result.Imdb,
		}
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// Search 搜索
func (t *TNodeSpider) Search(keyword string, page int) (bool, []map[string]interface{}) {
	// 获取token
	token := t.getToken()
	if token == "" {
		// logger.Warn(fmt.Sprintf("%s 未获取到token，无法搜�?, t.name))
		fmt.Printf("%s 未获取到token，无法搜索\n", t.name)
		return true, []map[string]interface{}{}
	}
	
	// 获取请求参数
	params := t.getParams(keyword, page)
	
	// 将参数转换为JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法序列化参数: %v", t.name, err))
		fmt.Printf("%s 搜索失败，无法序列化参数: %v\n", t.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(t.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", t.searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", t.name, err))
		fmt.Printf("%s 搜索失败，无法创建请�? %v\n", t.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("X-CSRF-TOKEN", token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if t.ua != "" {
		req.Header.Set("User-Agent", t.ua)
	}
	if t.cookie != "" {
		req.Header.Set("Cookie", t.cookie)
	}
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", t.name, t.domain, err))
		fmt.Printf("%s 搜索失败，无法连�?%s: %v\n", t.name, t.domain, err)
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result TNodeResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", t.name, err))
			fmt.Printf("%s 搜索失败，无法解析响�? %v\n", t.name, err)
			return true, []map[string]interface{}{}
		}
		
		return false, t.parseResult(result.Data.Torrents)
	} else {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", t.name, res.StatusCode))
		fmt.Printf("%s 搜索失败，错误码�?d\n", t.name, res.StatusCode)
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索
func (t *TNodeSpider) AsyncSearch(keyword string, page int) (bool, []map[string]interface{}) {
	// 异步搜索实现与同步搜索相�?	// 在Go中，可以通过goroutines实现真正的异步处�?	return t.Search(keyword, page)
}

// formatTimestamp 格式化时间戳
func (t *TNodeSpider) formatTimestamp(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	
	// 转换为时间字符串
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04:05")
}
