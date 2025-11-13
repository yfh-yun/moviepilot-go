package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
	"go.uber.org/zap"
)

// YemaSpider YemaPT API实现
type YemaSpider struct {
	indexerid   interface{}
	domain      string
	name        string
	proxy       string
	cookie      string
	ua          string
	size        int
	searchURL   string
	downloadURL string
	pageURL     string
	timeout     int

	// 分类
	movieCategory []int
	tvCategory    []int

	// 标签 https://wiki.yemapt.org/developer/constants
	labels map[string]string

	// 日志记录�?	logger *logger.Logger
}

// YemaParams 搜索参数结构
type YemaParams struct {
	PageParam struct {
		Current  int `json:"current"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
	} `json:"pageParam"`
	Sorter  map[string]interface{} `json:"sorter"`
	Keyword string                 `json:"keyword,omitempty"`
}

// YemaResponse API响应结构
type YemaResponse struct {
	Data []YemaTorrent `json:"data"`
}

// YemaTorrent 种子信息结构
type YemaTorrent struct {
	ID                      string      `json:"id"`
	ShowName                string      `json:"showName"`
	ShortDesc               string      `json:"shortDesc"`
	FileSize                int64       `json:"fileSize"`
	SeedNum                 int         `json:"seedNum"`
	LeechNum                int         `json:"leechNum"`
	CompletedNum            int         `json:"completedNum"`
	CategoryId              int         `json:"categoryId"`
	TagList                 []string    `json:"tagList"`
	ListingTime             string      `json:"listingTime"`
	DownloadPromotion       string      `json:"downloadPromotion"`
	UploadPromotion         string      `json:"uploadPromotion"`
	DownloadPromotionEndTime string     `json:"downloadPromotionEndTime"`
}

// NewYemaSpider 创建YemaSpider实例
func NewYemaSpider(indexer map[string]interface{}) *YemaSpider {
	log, _ := logger.NewLogger()

	ys := &YemaSpider{
		size:        40,
		timeout:     15,
		searchURL:   "%sapi/torrent/fetchOpenTorrentList",
		downloadURL: "%sapi/torrent/download?id=%s",
		pageURL:     "%s#/torrent/detail/%s/",
		movieCategory: []int{4},
		tvCategory:    []int{5, 13, 14, 17, 15, 6, 16},
		labels: map[string]string{
			"1":  "禁转",
			"2":  "首发",
			"3":  "官方",
			"4":  "自制",
			"5":  "国语",
			"6":  "中字",
			"7":  "粤语",
			"8":  "英字",
			"9":  "HDR10",
			"10": "杜比视界",
			"11": "分集",
			"12": "完结",
		},
		logger: log,
	}

	if indexer != nil {
		if indexerid, ok := indexer["id"]; ok {
			ys.indexerid = indexerid
		}
		if domain, ok := indexer["domain"].(string); ok {
			ys.domain = domain
			ys.searchURL = fmt.Sprintf(ys.searchURL, domain)
			ys.downloadURL = fmt.Sprintf(ys.downloadURL, domain, "%s")
			ys.pageURL = fmt.Sprintf(ys.pageURL, domain, "%s")
		}
		if name, ok := indexer["name"].(string); ok {
			ys.name = name
		}
		if proxy, ok := indexer["proxy"].(bool); ok && proxy {
			// TODO: 设置代理
			// ys.proxy = settings.PROXY
		}
		if cookie, ok := indexer["cookie"].(string); ok {
			ys.cookie = cookie
		}
		if ua, ok := indexer["ua"].(string); ok {
			ys.ua = ua
		}
		if timeout, ok := indexer["timeout"].(int); ok {
			ys.timeout = timeout
		} else if timeout, ok := indexer["timeout"].(float64); ok {
			ys.timeout = int(timeout)
		}
	}

	return ys
}

// getParams 获取搜索参数
func (ys *YemaSpider) getParams(keyword string, page int) *YemaParams {
	params := &YemaParams{
		Sorter: make(map[string]interface{}),
	}
	params.PageParam.Current = page + 1
	params.PageParam.PageSize = ys.size
	params.PageParam.Total = ys.size

	if keyword != "" {
		params.Keyword = keyword
	}

	return params
}

// parseResult 解析搜索结果
func (ys *YemaSpider) parseResult(results []YemaTorrent) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	stringUtils := utils.NewStringUtils()

	if len(results) == 0 {
		return torrents
	}

	for _, result := range results {
		// 确定分类
		category := models.Unknown
		for _, cat := range ys.tvCategory {
			if result.CategoryId == cat {
				category = models.TV
				break
			}
		}
		if category == models.Unknown {
			for _, cat := range ys.movieCategory {
				if result.CategoryId == cat {
					category = models.Movie
					break
				}
			}
		}

		// 处理标签
		torrentLabels := make([]string, 0)
		for _, labelId := range result.TagList {
			if label, ok := ys.labels[labelId]; ok {
				torrentLabels = append(torrentLabels, label)
			}
		}

		torrent := map[string]interface{}{
			"title":                result.ShowName,
			"description":          result.ShortDesc,
			"enclosure":            ys.getDownloadURL(result.ID),
			"pubdate":              stringUtils.UnifyDateTimeStr(result.ListingTime),
			"size":                 result.FileSize,
			"seeders":              result.SeedNum,
			"peers":                result.LeechNum,
			"grabs":                result.CompletedNum,
			"downloadvolumefactor": ys.getDownloadVolumeFactor(result.DownloadPromotion),
			"uploadvolumefactor":   ys.getUploadVolumeFactor(result.UploadPromotion),
			"freedate":             stringUtils.UnifyDateTimeStr(result.DownloadPromotionEndTime),
			"page_url":             fmt.Sprintf(ys.pageURL, result.ID),
			"labels":               torrentLabels,
			"category":             category,
		}
		torrents = append(torrents, torrent)
	}

	return torrents
}

// Search 搜索种子
func (ys *YemaSpider) Search(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 构造请求参�?	params := ys.getParams(keyword, page)
	jsonData, err := json.Marshal(params)
	if err != nil {
		ys.logger.Warn(fmt.Sprintf("%s 搜索失败，无法构造请求参�? %v", ys.name, err))
		return true, []map[string]interface{}{}
	}

	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(ys.timeout) * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("POST", ys.searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ys.logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", ys.name, err))
		return true, []map[string]interface{}{}
	}

	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	if ys.ua != "" {
		req.Header.Set("User-Agent", ys.ua)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	if ys.cookie != "" {
		req.Header.Set("Cookie", ys.cookie)
	}
	if ys.domain != "" {
		req.Header.Set("Referer", ys.domain)
	}

	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		ys.logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", ys.name, ys.domain, err),
			zap.String("domain", ys.domain),
			zap.String("keyword", keyword))
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()

	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var response YemaResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			ys.logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", ys.name, err))
			return true, []map[string]interface{}{}
		}

		return false, ys.parseResult(response.Data)
	} else {
		ys.logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", ys.name, res.StatusCode),
			zap.String("domain", ys.domain),
			zap.String("keyword", keyword),
			zap.Int("status_code", res.StatusCode))
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索种子
func (ys *YemaSpider) AsyncSearch(keyword string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 异步搜索实现与同步搜索相�?	// 在Go中，可以通过goroutines实现真正的异步处�?	return ys.Search(keyword, mtype, page)
}

// getDownloadVolumeFactor 获取下载系数
func (ys *YemaSpider) getDownloadVolumeFactor(discount string) float64 {
	discountDict := map[string]float64{
		"free": 0,
		"half": 0.5,
		"none": 1,
	}

	if discount != "" {
		if factor, ok := discountDict[discount]; ok {
			return factor
		}
	}
	return 1
}

// getUploadVolumeFactor 获取上传系数
func (ys *YemaSpider) getUploadVolumeFactor(discount string) float64 {
	discountDict := map[string]float64{
		"none":          1,
		"one_half":      1.5,
		"double_upload": 2,
	}

	if discount != "" {
		if factor, ok := discountDict[discount]; ok {
			return factor
		}
	}
	return 1
}

// getDownloadURL 获取下载链接
func (ys *YemaSpider) getDownloadURL(torrentID string) string {
	return fmt.Sprintf(ys.downloadURL, torrentID)
}
