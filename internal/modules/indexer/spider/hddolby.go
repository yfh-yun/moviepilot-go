package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	
	"moviepilot-go/pkg/models"
)

// HddolbySpider HDDolby API实现
type HddolbySpider struct {
	indexerID  int
	domain     string
	domainHost string
	name       string
	proxy      string
	cookie     string
	ua         string
	apikey     string
	size       int
	pageURL    string
	timeout    int
	searchURL  string
	
	// 分类
	movieCategory []int
	tvCategory    []int
	
	// 标签
	labels map[string]string
}

// HddolbyResult API返回结果结构
type HddolbyResult struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Data []HddolbyTorrent `json:"data"`
}

// HddolbyTorrent 种子信息结构
type HddolbyTorrent struct {
	ID                  int    `json:"id"`
	PromotionTimeType   int    `json:"promotion_time_type"`
	PromotionUntil      string `json:"promotion_until"`
	Category            int    `json:"category"`
	Name                string `json:"name"`
	SmallDescr          string `json:"small_descr"`
	TimesCompleted      int    `json:"times_completed"`
	Size                int64  `json:"size"`
	Added               string `json:"added"`
	Leechers            int    `json:"leechers"`
	Seeders             int    `json:"seeders"`
	Downhash            string `json:"downhash"`
	Tags                string `json:"tags"`
}

// HddolbySearchParams 搜索参数结构
type HddolbySearchParams struct {
	Keyword     string `json:"keyword"`
	PageNumber  int    `json:"page_number"`
	PageSize    int    `json:"page_size"`
	Categories  []int  `json:"categories"`
	Visible     int    `json:"visible"`
}

// NewHddolbySpider 创建HddolbySpider实例
func NewHddolbySpider(indexer map[string]interface{}) *HddolbySpider {
	spider := &HddolbySpider{
		size:          40,
		timeout:       15,
		movieCategory: []int{401, 405},
		tvCategory:    []int{402, 403, 404, 405},
		labels: map[string]string{
			"gf":   "官方",
			"gy":   "国语",
			"yy":   "粤语",
			"ja":   "日语",
			"ko":   "韩语",
			"zz":   "中文字幕",
			"jz":   "禁转",
			"xz":   "限转",
			"diy":  "DIY",
			"sf":   "首发",
			"yq":   "应求",
			"m0":   "零魔",
			"yc":   "原创",
			"gz":   "官字",
			"db":   "Dolby Vision",
			"hdr10": "HDR10",
			"hdrm": "HDR10+",
			"tx":   "特效",
			"lz":   "连载",
			"wj":   "完结",
			"hdrv": "HDR Vivid",
			"hlg":  "HLG",
			"hq":   "高码�?,
			"hfr":  "高帧�?,
		},
	}
	
	if indexer != nil {
		if id, ok := indexer["id"].(int); ok {
			spider.indexerID = id
		}
		
		if domain, ok := indexer["domain"].(string); ok {
			spider.domain = domain
			spider.domainHost = getURLDomain(domain) // 需要实现这个函�?			spider.searchURL = fmt.Sprintf("https://api.%s/api/v1/torrent/search", spider.domainHost)
			spider.pageURL = fmt.Sprintf("%sdetails.php?id=%%s&hit=1", domain)
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
		
		if timeout, ok := indexer["timeout"].(int); ok {
			spider.timeout = timeout
		} else if timeout, ok := indexer["timeout"].(float64); ok {
			spider.timeout = int(timeout)
		}
	}
	
	return spider
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

// getParams 获取请求参数
func (h *HddolbySpider) getParams(keyword string, mtype models.MediaType, page int) *HddolbySearchParams {
	var categories []int
	
	switch mtype {
	case models.MediaTypeTV:
		categories = h.tvCategory
	case models.MediaTypeMovie:
		categories = h.movieCategory
	default:
		// 合并电影和电视剧分类，并去重
		categoryMap := make(map[int]bool)
		for _, cat := range h.movieCategory {
			categoryMap[cat] = true
		}
		for _, cat := range h.tvCategory {
			categoryMap[cat] = true
		}
		
		categories = make([]int, 0, len(categoryMap))
		for cat := range categoryMap {
			categories = append(categories, cat)
		}
	}
	
	return &HddolbySearchParams{
		Keyword:    keyword,
		PageNumber: page,
		PageSize:   100,
		Categories: categories,
		Visible:    1,
	}
}

// parseResult 解析搜索结果
func (h *HddolbySpider) parseResult(results []HddolbyTorrent) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	
	if len(results) == 0 {
		return torrents
	}
	
	for _, result := range results {
		// 类别
		var category string
		isTVCat := false
		isMovieCat := false
		
		for _, cat := range h.tvCategory {
			if cat == result.Category {
				isTVCat = true
				break
			}
		}
		
		for _, cat := range h.movieCategory {
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
		
		// 标签
		torrentLabelIds := strings.Split(result.Tags, ";")
		torrentLabels := make([]string, 0)
		for _, labelId := range torrentLabelIds {
			if label, exists := h.labels[labelId]; exists {
				torrentLabels = append(torrentLabels, label)
			}
		}
		
		// 种子信息
		torrent := map[string]interface{}{
			"title":                result.Name,
			"description":          result.SmallDescr,
			"enclosure":            h.getDownloadURL(result.ID, result.Downhash),
			"pubdate":              result.Added,
			"size":                 result.Size,
			"seeders":              result.Seeders,
			"peers":                result.Leechers,
			"grabs":                result.TimesCompleted,
			"downloadvolumefactor": h.getDownloadVolumeFactor(result.PromotionTimeType),
			"uploadvolumefactor":   h.getUploadVolumeFactor(result.PromotionTimeType),
			"freedate":             result.PromotionUntil,
			"page_url":             fmt.Sprintf(h.pageURL, strconv.Itoa(result.ID)),
			"labels":               torrentLabels,
			"category":             category,
		}
		
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// Search 搜索
func (h *HddolbySpider) Search(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 准备参数
	params := h.getParams(keyword, mtype, page)
	
	// 将参数转换为JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法序列化参数: %v", h.name, err))
		fmt.Printf("%s 搜索失败，无法序列化参数: %v\n", h.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(h.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", h.searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", h.name, err))
		fmt.Printf("%s 搜索失败，无法创建请�? %v\n", h.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("x-api-key", h.apikey)
	if h.cookie != "" {
		req.Header.Set("Cookie", h.cookie)
	}
	req.Header.Set("Referer", h.domain)
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", h.name, h.domain, err))
		fmt.Printf("%s 搜索失败，无法连�?%s: %v\n", h.name, h.domain, err)
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result HddolbyResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", h.name, err))
			fmt.Printf("%s 搜索失败，无法解析响�? %v\n", h.name, err)
			return true, []map[string]interface{}{}
		}
		
		// 检查是否有错误
		if result.Error != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，错误信息：%s", h.name, result.Error.Message))
			fmt.Printf("%s 搜索失败，错误信息：%s\n", h.name, result.Error.Message)
			return true, []map[string]interface{}{}
		}
		
		return false, h.parseResult(result.Data)
	} else {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", h.name, res.StatusCode))
		fmt.Printf("%s 搜索失败，错误码�?d\n", h.name, res.StatusCode)
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索
func (h *HddolbySpider) AsyncSearch(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 异步搜索实现与同步搜索相�?	// 在Go中，可以通过goroutines实现真正的异步处�?	return h.Search(keyword, mtype, page)
}

// getDownloadVolumeFactor 获取下载系数
func (h *HddolbySpider) getDownloadVolumeFactor(discount int) float64 {
	discountDict := map[int]float64{
		2: 0,
		5: 0.5,
		6: 1,
		7: 0.3,
	}
	
	if discount != 0 {
		if factor, exists := discountDict[discount]; exists {
			return factor
		}
	}
	return 1
}

// getUploadVolumeFactor 获取上传系数
func (h *HddolbySpider) getUploadVolumeFactor(discount int) float64 {
	discountDict := map[int]float64{
		3: 2,
		4: 2,
		6: 2,
	}
	
	if discount != 0 {
		if factor, exists := discountDict[discount]; exists {
			return factor
		}
	}
	return 1
}

// getDownloadURL 获取下载链接
func (h *HddolbySpider) getDownloadURL(torrentID int, downhash string) string {
	return fmt.Sprintf("%sdownload.php?id=%d&downhash=%s", h.domain, torrentID, downhash)
}
