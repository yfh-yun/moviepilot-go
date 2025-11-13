package spider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	"moviepilot-go/pkg/models"
)

// MTorrentSpider mTorrent API实现
type MTorrentSpider struct {
	indexerID  int
	domain     string
	url        string
	name       string
	proxy      string
	cookie     string
	ua         string
	size       int
	searchURL  string
	downloadURL string
	pageURL    string
	timeout    int
	
	// 电影分类
	movieCategory []string
	tvCategory    []string
	
	// API KEY
	apikey string
	// JWT Token
	token string
	
	// 标签
	labels map[string]string
}

// MTorrentSearchParams 搜索参数结构
type MTorrentSearchParams struct {
	Keyword    string   `json:"keyword"`
	Categories []string `json:"categories"`
	PageNumber int      `json:"pageNumber"`
	PageSize   int      `json:"pageSize"`
	Visible    int      `json:"visible"`
}

// MTorrentResult API返回结果结构
type MTorrentResult struct {
	Data struct {
		Data []MTorrentTorrent `json:"data"`
	} `json:"data"`
}

// MTorrentTorrent 种子信息结构
type MTorrentTorrent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SmallDescr  string `json:"smallDescr"`
	Size        int64  `json:"size"`
	CreatedDate int64  `json:"createdDate"`
	Category    string `json:"category"`
	Imdb        string `json:"imdb"`
	Status      struct {
		Seeders          int    `json:"seeders"`
		Leechers         int    `json:"leechers"`
		TimesCompleted   int    `json:"timesCompleted"`
		Discount         string `json:"discount"`
		DiscountEndTime  int64  `json:"discountEndTime"`
	} `json:"status"`
	Labels    string   `json:"labels"`
	LabelsNew []string `json:"labelsNew"`
}

// NewMTorrentSpider 创建MTorrentSpider实例
func NewMTorrentSpider(indexer map[string]interface{}) *MTorrentSpider {
	spider := &MTorrentSpider{
		size:          100,
		timeout:       15,
		movieCategory: []string{"401", "419", "420", "421", "439", "405", "404"},
		tvCategory:    []string{"403", "402", "435", "438", "404", "405"},
		labels: map[string]string{
			"0": "",
			"1": "DIY",
			"2": "国配",
			"3": "DIY 国配",
			"4": "中字",
			"5": "DIY 中字",
			"6": "国配 中字",
			"7": "DIY 国配 中字",
		},
	}
	
	if indexer != nil {
		if id, ok := indexer["id"].(int); ok {
			spider.indexerID = id
		}
		
		if domain, ok := indexer["domain"].(string); ok {
			spider.url = domain
			spider.domain = getURLDomain(domain)
			spider.searchURL = fmt.Sprintf("https://api.%s/api/torrent/search", spider.domain)
			spider.downloadURL = fmt.Sprintf("https://api.%s/api/torrent/genDlToken", spider.domain)
			spider.pageURL = fmt.Sprintf("%sdetail/%%s", domain)
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
		
		if apikey, ok := indexer["apikey"].(string); ok {
			spider.apikey = apikey
		}
		
		if token, ok := indexer["token"].(string); ok {
			spider.token = token
		}
		
		if timeout, ok := indexer["timeout"].(int); ok {
			spider.timeout = timeout
		} else if timeout, ok := indexer["timeout"].(float64); ok {
			spider.timeout = int(timeout)
		}
	}
	
	return spider
}

// getParams 获取请求参数
func (m *MTorrentSpider) getParams(keyword string, mtype models.MediaType, page int) *MTorrentSearchParams {
	var categories []string
	
	if mtype == "" {
		categories = []string{}
	} else if mtype == models.MediaTypeTV {
		categories = m.tvCategory
	} else {
		categories = m.movieCategory
	}
	
	// mtorrent搜索imdb需要输入完整imdb链接
	if strings.HasPrefix(keyword, "tt") {
		keyword = fmt.Sprintf("https://www.imdb.com/title/%s", keyword)
	}
	
	return &MTorrentSearchParams{
		Keyword:    keyword,
		Categories: categories,
		PageNumber: page + 1,
		PageSize:   m.size,
		Visible:    1,
	}
}

// parseResult 解析搜索结果
func (m *MTorrentSpider) parseResult(results []MTorrentTorrent) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	
	if len(results) == 0 {
		return torrents
	}
	
	for _, result := range results {
		// 判断分类
		var category string
		isTVCat := false
		isMovieCat := false
		
		for _, cat := range m.tvCategory {
			if cat == result.Category {
				isTVCat = true
				break
			}
		}
		
		for _, cat := range m.movieCategory {
			if cat == result.Category {
				isMovieCat = true
				break
			}
		}
		
		if isTVCat && !isMovieCat {
			category = string(models.MediaTypeTV)
		} else if isMovieCat {
			category = string(models.MediaTypeMovie)
		} else {
			category = string(models.MediaTypeUnknown)
		}
		
		// 处理标签
		labels := make([]string, 0)
		if len(result.LabelsNew) > 0 {
			// 新版标签本身就是list
			labels = result.LabelsNew
		} else {
			// 旧版标签
			labelsValue := m.labels[result.Labels]
			if labelsValue != "" {
				labels = strings.Split(labelsValue, " ")
			}
		}
		
		// 构造种子信�?		torrent := map[string]interface{}{
			"title":                result.Name,
			"description":          result.SmallDescr,
			"enclosure":            m.getDownloadURL(result.ID),
			"pubdate":              m.formatTimestamp(result.CreatedDate),
			"size":                 result.Size,
			"seeders":              result.Status.Seeders,
			"peers":                result.Status.Leechers,
			"grabs":                result.Status.TimesCompleted,
			"downloadvolumefactor": m.getDownloadVolumeFactor(result.Status.Discount),
			"uploadvolumefactor":   m.getUploadVolumeFactor(result.Status.Discount),
			"page_url":             fmt.Sprintf(m.pageURL, result.ID),
			"imdbid":               m.findIMDbID(result.Imdb),
			"labels":               labels,
			"category":             category,
		}
		
		// 添加免费截止时间
		if result.Status.DiscountEndTime > 0 {
			torrent["freedate"] = m.formatTimestamp(result.Status.DiscountEndTime)
		}
		
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// Search 搜索
func (m *MTorrentSpider) Search(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 检查ApiKey
	if m.apikey == "" {
		return true, []map[string]interface{}{}
	}
	
	// 获取请求参数
	params := m.getParams(keyword, mtype, page)
	
	// 将参数转换为JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法序列化参数: %v", m.name, err))
		fmt.Printf("%s 搜索失败，无法序列化参数: %v\n", m.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(m.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", m.searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", m.name, err))
		fmt.Printf("%s 搜索失败，无法创建请�? %v\n", m.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", m.ua)
	req.Header.Set("x-api-key", m.apikey)
	req.Header.Set("Referer", fmt.Sprintf("%sbrowse", m.domain))
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", m.name, m.domain, err))
		fmt.Printf("%s 搜索失败，无法连�?%s: %v\n", m.name, m.domain, err)
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result MTorrentResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", m.name, err))
			fmt.Printf("%s 搜索失败，无法解析响�? %v\n", m.name, err)
			return true, []map[string]interface{}{}
		}
		
		return false, m.parseResult(result.Data.Data)
	} else {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", m.name, res.StatusCode))
		fmt.Printf("%s 搜索失败，错误码�?d\n", m.name, res.StatusCode)
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索
func (m *MTorrentSpider) AsyncSearch(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 异步搜索实现与同步搜索相�?	// 在Go中，可以通过goroutines实现真正的异步处�?	return m.Search(keyword, mtype, page)
}

// findIMDbID 从imdb链接中提取imdbid
func (m *MTorrentSpider) findIMDbID(imdb string) string {
	if imdb != "" {
		re := regexp.MustCompile(`tt\d+`)
		match := re.FindString(imdb)
		if match != "" {
			return match
		}
	}
	return ""
}

// getDownloadVolumeFactor 获取下载系数
func (m *MTorrentSpider) getDownloadVolumeFactor(discount string) float64 {
	discountDict := map[string]float64{
		"FREE":            0,
		"PERCENT_50":      0.5,
		"PERCENT_70":      0.3,
		"_2X_FREE":        0,
		"_2X_PERCENT_50":  0.5,
	}
	
	if discount != "" {
		if factor, exists := discountDict[discount]; exists {
			return factor
		}
	}
	return 1
}

// getUploadVolumeFactor 获取上传系数
func (m *MTorrentSpider) getUploadVolumeFactor(discount string) float64 {
	uploadVolumeFactorDict := map[string]float64{
		"_2X":             2.0,
		"_2X_FREE":        2.0,
		"_2X_PERCENT_50":  2.0,
	}
	
	if discount != "" {
		if factor, exists := uploadVolumeFactorDict[discount]; exists {
			return factor
		}
	}
	return 1
}

// getDownloadURL 获取下载链接，返回base64编码的json字符串及URL
func (m *MTorrentSpider) getDownloadURL(torrentID string) string {
	// 构造URL
	url := fmt.Sprintf(m.downloadURL, m.domain)
	
	// 构造参�?	params := map[string]interface{}{
		"method": "post",
		"cookie": false,
		"params": map[string]string{
			"id": torrentID,
		},
		"header": map[string]string{
			"User-Agent": m.ua,
			"Accept":     "application/json, text/plain, */*",
			"x-api-key":  m.apikey,
		},
		"proxy":  m.proxy != "",
		"result": "data",
	}
	
	// 转换为JSON并base64编码
	jsonData, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	
	base64Str := base64.StdEncoding.EncodeToString(jsonData)
	return fmt.Sprintf("[%s]%s", base64Str, url)
}

// formatTimestamp 格式化时间戳
func (m *MTorrentSpider) formatTimestamp(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	
	// 转换为时间字符串
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// getURLDomain 获取URL域名
func getURLDomain(urlStr string) string {
	// 使用net/url包解析URL获取域名
	if urlStr == "" {
		return ""
	}
	
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	return u.Host
}
