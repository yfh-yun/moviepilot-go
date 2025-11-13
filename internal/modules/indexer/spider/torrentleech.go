package spider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	
	"moviepilot-go/internal/utils"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// TorrentLeech TorrentLeech实现
type TorrentLeech struct {
	indexer map[string]interface{}
	proxy   string
	size    int
	timeout int
	
	// URL模板
	searchURL   string
	browseURL   string
	downloadURL string
	pageURL     string
	
	// 日志记录�?	logger *logger.Logger
}

// TorrentLeechResult API返回结果结构
type TorrentLeechResult struct {
	TorrentList []TorrentLeechTorrent `json:"torrentList"`
}

// TorrentLeechTorrent 种子信息结构
type TorrentLeechTorrent struct {
	Name               string  `json:"name"`
	Fid                int     `json:"fid"`
	Filename           string  `json:"filename"`
	AddedTimestamp     int64   `json:"addedTimestamp"`
	Size               int64   `json:"size"`
	Seeders            int     `json:"seeders"`
	Leechers           int     `json:"leechers"`
	Completed          int     `json:"completed"`
	DownloadMultiplier float64 `json:"download_multiplier"`
	ImdbID             string  `json:"imdbID"`
}

// NewTorrentLeech 创建TorrentLeech实例
func NewTorrentLeech(indexer map[string]interface{}) *TorrentLeech {
	log, _ := logger.NewLogger()
	
	tl := &TorrentLeech{
		indexer:     indexer,
		size:        100,
		timeout:     15,
		searchURL:   "%storrents/browse/list/query/%s",
		browseURL:   "%storrents/browse/list/page/%d",
		downloadURL: "%sdownload/%d/%s",
		pageURL:     "%storrent/%d",
		logger:      log,
	}
	
	if indexer != nil {
		if proxy, ok := indexer["proxy"].(bool); ok && proxy {
			// TODO: 设置代理
			// tl.proxy = settings.PROXY
			if timeout, ok := indexer["timeout"].(int); ok {
				tl.timeout = timeout
			} else if timeout, ok := indexer["timeout"].(float64); ok {
				tl.timeout = int(timeout)
			}
		}
	}
	
	return tl
}

// parseResult 解析搜索结果
func (tl *TorrentLeech) parseResult(results []TorrentLeechTorrent) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	
	if len(results) == 0 {
		return torrents
	}
	
	domain := ""
	if d, ok := tl.indexer["domain"].(string); ok {
		domain = d
	}
	
	stringUtils := utils.NewStringUtils()
	
	for _, result := range results {
		torrent := map[string]interface{}{
			"title":                result.Name,
			"enclosure":            fmt.Sprintf(tl.downloadURL, domain, result.Fid, result.Filename),
			"pubdate":              stringUtils.FormatTimestamp(result.AddedTimestamp),
			"size":                 result.Size,
			"seeders":              result.Seeders,
			"peers":                result.Leechers,
			"grabs":                result.Completed,
			"downloadvolumefactor": result.DownloadMultiplier,
			"uploadvolumefactor":   1,
			"page_url":             fmt.Sprintf(tl.pageURL, domain, result.Fid),
			"imdbid":               result.ImdbID,
		}
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// Search 搜索种子
func (tl *TorrentLeech) Search(keyword string, page int) (bool, []map[string]interface{}) {
	// 检查是否包含中�?	stringUtils := utils.NewStringUtils()
	if stringUtils.IsChinese(keyword) {
		// 不支持中�?		return true, []map[string]interface{}{}
	}
	
	var searchURL string
	domain := ""
	if d, ok := tl.indexer["domain"].(string); ok {
		domain = d
	}
	
	if keyword != "" {
		searchURL = fmt.Sprintf(tl.searchURL, domain, url.QueryEscape(keyword))
	} else {
		searchURL = fmt.Sprintf(tl.browseURL, domain, page+1)
	}
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(tl.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		tl.logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", getName(tl.indexer), err),
			zap.String("domain", domain),
			zap.String("keyword", keyword))
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if ua, ok := tl.indexer["ua"].(string); ok {
		req.Header.Set("User-Agent", ua)
	}
	if cookie, ok := tl.indexer["cookie"].(string); ok {
		req.Header.Set("Cookie", cookie)
	}
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		tl.logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", getName(tl.indexer), domain, err),
			zap.String("domain", domain),
			zap.String("keyword", keyword))
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result TorrentLeechResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			tl.logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", getName(tl.indexer), err),
				zap.String("domain", domain),
				zap.String("keyword", keyword))
			return true, []map[string]interface{}{}
		}
		
		return false, tl.parseResult(result.TorrentList)
	} else {
		tl.logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", getName(tl.indexer), res.StatusCode),
			zap.String("domain", domain),
			zap.String("keyword", keyword),
			zap.Int("status_code", res.StatusCode))
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索种子
func (tl *TorrentLeech) AsyncSearch(keyword string, page int) (bool, []map[string]interface{}) {
	// 异步搜索实现与同步搜索相�?	// 在Go中，可以通过goroutines实现真正的异步处�?	return tl.Search(keyword, page)
}

// getName 获取索引器名�?func getName(indexer map[string]interface{}) string {
	if name, ok := indexer["name"].(string); ok {
		return name
	}
	return ""
}
