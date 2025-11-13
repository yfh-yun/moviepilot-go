package spider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// HaiDanSpider haidan.video API实现
type HaiDanSpider struct {
	indexerID   int
	domain      string
	searchURL   string
	detailURL   string
	name        string
	proxy       string
	cookie      string
	ua          string
	size        int
	timeout     int
	
	// 电影分类
	movieCategory []string
	tvCategory    []string
	
	// 足销状�?1-普通，2-免费�?-2X�?-2X免费�?-50%�?-2X50%�?-30%
	dlState map[string]float64
	upState map[string]float64
}

// HaiDanResult API返回结果结构
type HaiDanResult struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

// NewHaiDanSpider 创建HaiDanSpider实例
func NewHaiDanSpider(indexer map[string]interface{}) *HaiDanSpider {
	spider := &HaiDanSpider{
		size:          100,
		timeout:       15,
		movieCategory: []string{"401", "404", "405"},
		tvCategory:    []string{"402", "403", "404", "405"},
		dlState: map[string]float64{
			"1": 1,
			"2": 0,
			"3": 1,
			"4": 0,
			"5": 0.5,
			"6": 0.5,
			"7": 0.3,
		},
		upState: map[string]float64{
			"1": 1,
			"2": 1,
			"3": 2,
			"4": 2,
			"5": 1,
			"6": 2,
			"7": 1,
		},
	}
	
	if indexer != nil {
		if id, ok := indexer["id"].(int); ok {
			spider.indexerID = id
		}
		
		if domain, ok := indexer["domain"].(string); ok {
			spider.domain = domain
			spider.searchURL = domain + "torrents.php"
			spider.detailURL = domain + "details.php?group_id=%s&torrent_id=%s"
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

// getParams 获取请求参数
func (h *HaiDanSpider) getParams(keyword string, mtype models.MediaType) string {
	params := url.Values{}
	
	// 搜索类型
	if strings.HasPrefix(keyword, "tt") {
		params.Set("search_area", "4") // 4-IMDb
	} else {
		params.Set("search_area", "0") // 0-标题
	}
	
	// 分类
	var categories []string
	if mtype == "" {
		categories = []string{}
	} else if mtype == models.MediaTypeTV {
		categories = h.tvCategory
	} else {
		categories = h.movieCategory
	}
	
	// 设置参数
	params.Set("isapi", "1")
	params.Set("search", keyword)
	params.Set("search_mode", "0") // 0-�?1-�?2-精准
	
	// 添加分类参数
	if len(categories) > 0 {
		params.Set("cat", strings.Join(categories, ","))
	}
	
	return params.Encode()
}

// parseResult 解析结果
func (h *HaiDanSpider) parseResult(result *HaiDanResult) []map[string]interface{} {
	torrents := make([]map[string]interface{}, 0)
	
	if result.Data == nil {
		return torrents
	}
	
	for tid, item := range result.Data {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		// 判断分类
		categoryValue := ""
		if cat, exists := itemMap["category"]; exists {
			categoryValue = fmt.Sprintf("%v", cat)
		}
		
		var category string
		isTVCat := false
		isMovieCat := false
		
		for _, cat := range h.tvCategory {
			if cat == categoryValue {
				isTVCat = true
				break
			}
		}
		
		for _, cat := range h.movieCategory {
			if cat == categoryValue {
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
		
		// 构造种子信�?		torrent := map[string]interface{}{
			"title":        itemMap["name"],
			"description":  itemMap["small_descr"],
			"enclosure":    itemMap["url"],
			"pubdate":      h.formatTimestamp(itemMap["added"]),
			"size":         h.parseInt(itemMap["size"]),
			"seeders":      h.parseInt(itemMap["seeders"]),
			"peers":        h.parseInt(itemMap["leechers"]),
			"grabs":        h.parseInt(itemMap["times_completed"]),
			"downloadvolumefactor": h.getDownloadVolumeFactor(h.getString(itemMap["sp_state"])),
			"uploadvolumefactor":   h.getUploadVolumeFactor(h.getString(itemMap["sp_state"])),
			"page_url":     fmt.Sprintf(h.detailURL, h.getString(itemMap["group_id"]), tid),
			"labels":       []string{},
			"category":     category,
		}
		
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// formatTimestamp 格式化时间戳
func (h *HaiDanSpider) formatTimestamp(timestamp interface{}) string {
	// TODO: 实现时间戳格式化
	// 暂时返回空字符串，需要根据StringUtils.format_timestamp实现
	// 这里简单处理一下时间戳
	if timestamp == nil {
		return ""
	}
	
	// 尝试转换为整�?	ts := h.parseInt(timestamp)
	if ts <= 0 {
		return ""
	}
	
	// 转换为时间字符串
	t := time.Unix(int64(ts), 0)
	return t.Format("2006-01-02 15:04:05")
}

// parseInt 安全地转换为整数
func (h *HaiDanSpider) parseInt(value interface{}) int {
	if value == nil {
		return 0
	}
	
	switch v := value.(type) {
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	}
	
	return 0
}

// getString 安全地转换为字符�?func (h *HaiDanSpider) getString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	return fmt.Sprintf("%v", value)
}

// getDownloadVolumeFactor 获取下载系数
func (h *HaiDanSpider) getDownloadVolumeFactor(discount string) float64 {
	if discount != "" {
		if factor, exists := h.dlState[discount]; exists {
			return factor
		}
	}
	return 1
}

// getUploadVolumeFactor 获取上传系数
func (h *HaiDanSpider) getUploadVolumeFactor(discount string) float64 {
	if discount != "" {
		if factor, exists := h.upState[discount]; exists {
			return factor
		}
	}
	return 1
}

// Search 搜索
func (h *HaiDanSpider) Search(keyword string, mtype models.MediaType) (bool, []map[string]interface{}) {
	// 检查cookie
	if h.cookie == "" {
		return true, []map[string]interface{}{}
	}
	
	// 获取参数
	paramsStr := h.getParams(keyword, mtype)
	
	// 构造完整URL
	fullURL := h.searchURL + "?" + paramsStr
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(h.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", h.name, err))
		fmt.Printf("%s 搜索失败，无法创建请�? %v\n", h.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("User-Agent", h.ua)
	req.Header.Set("Cookie", h.cookie)
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", h.name, h.domain, err))
		fmt.Printf("%s 搜索失败，无法连�?%s: %v\n", h.name, h.domain, err)
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result HaiDanResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", h.name, err))
			fmt.Printf("%s 搜索失败，无法解析响�? %v\n", h.name, err)
			return true, []map[string]interface{}{}
		}
		
		// 检查返回码
		if result.Code != 0 {
			// logger.Warn(fmt.Sprintf("%s 搜索失败�?s", h.name, result.Msg))
			fmt.Printf("%s 搜索失败�?s\n", h.name, result.Msg)
			return true, []map[string]interface{}{}
		}
		
		return false, h.parseResult(&result)
	} else {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", h.name, res.StatusCode))
		fmt.Printf("%s 搜索失败，错误码�?d\n", h.name, res.StatusCode)
		return true, []map[string]interface{}{}
	}
}

// AsyncSearch 异步搜索
func (h *HaiDanSpider) AsyncSearch(keyword string, mtype models.MediaType) (bool, []map[string]interface{}) {
	// 检查cookie
	if h.cookie == "" {
		return true, []map[string]interface{}{}
	}
	
	// 获取参数
	paramsStr := h.getParams(keyword, mtype)
	
	// 构造完整URL
	fullURL := h.searchURL + "?" + paramsStr
	
	// 创建HTTP客户端，设置超时时间
	client := &http.Client{
		Timeout: time.Duration(h.timeout) * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法创建请�? %v", h.name, err))
		fmt.Printf("%s 搜索失败，无法创建请�? %v\n", h.name, err)
		return true, []map[string]interface{}{}
	}
	
	// 设置请求�?	req.Header.Set("User-Agent", h.ua)
	req.Header.Set("Cookie", h.cookie)
	
	// 发送请�?	res, err := client.Do(req)
	if err != nil {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，无法连�?%s: %v", h.name, h.domain, err))
		fmt.Printf("%s 搜索失败，无法连�?%s: %v\n", h.name, h.domain, err)
		return true, []map[string]interface{}{}
	}
	defer res.Body.Close()
	
	// 检查响应状�?	if res.StatusCode == 200 {
		// 解析JSON响应
		var result HaiDanResult
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			// logger.Warn(fmt.Sprintf("%s 搜索失败，无法解析响�? %v", h.name, err))
			fmt.Printf("%s 搜索失败，无法解析响�? %v\n", h.name, err)
			return true, []map[string]interface{}{}
		}
		
		// 检查返回码
		if result.Code != 0 {
			// logger.Warn(fmt.Sprintf("%s 搜索失败�?s", h.name, result.Msg))
			fmt.Printf("%s 搜索失败�?s\n", h.name, result.Msg)
			return true, []map[string]interface{}{}
		}
		
		return false, h.parseResult(&result)
	} else {
		// logger.Warn(fmt.Sprintf("%s 搜索失败，错误码�?d", h.name, res.StatusCode))
		fmt.Printf("%s 搜索失败，错误码�?d\n", h.name, res.StatusCode)
		return true, []map[string]interface{}{}
	}
}
