package actions

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"moviepilot-go/pkg/logger"
)

// RSSFetcher RSS获取器接口
type RSSFetcher interface {
	// FetchRSS 从指定URL获取RSS数据
	FetchRSS(ctx context.Context, params *FetchRSSParams) (*RSSResponse, error)
	// FetchMultipleRSS 批量获取多个RSS源的数据
	FetchMultipleRSS(ctx context.Context, paramsList []*FetchRSSParams) ([]*RSSResponse, error)
	// ValidateRSSFeed 验证RSS源是否有效
	ValidateRSSFeed(ctx context.Context, feedURL string) (bool, error)
	// GetRSSStats 获取RSS统计信息
	GetRSSStats(ctx context.Context) (*RSSStats, error)
	// ClearCache 清除指定RSS源的缓存
	ClearCache(feedURL string) error
}

// RSSFetcherImpl RSS获取器实现
type RSSFetcherImpl struct {
	httpClient *http.Client
	cache      *sync.Map // 简单的内存缓存，实际项目中可能需要使用Redis等
	logger     *zap.Logger
	stats      *RSSStats
	statsMutex sync.Mutex
}

// NewRSSFetcher 创建RSS获取器实例
func NewRSSFetcher() *RSSFetcherImpl {
	return &RSSFetcherImpl{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:  &sync.Map{},
		logger: logger.Logger,
		stats:  &RSSStats{},
	}
}

// FetchRSS 从指定URL获取RSS数据
func (f *RSSFetcherImpl) FetchRSS(ctx context.Context, params *FetchRSSParams) (*RSSResponse, error) {
	startTime := time.Now()
	response := &RSSResponse{
		Success: true,
		Entries: []RSSEntry{},
	}

	// 验证参数
	if err := f.validateFetchParams(params); err != nil {
		response.Success = false
		response.Error = &RSSError{
			Type:      RSSErrorTypeValidation,
			Message:   "参数验证失败",
			Details:   err.Error(),
			Timestamp: time.Now(),
		}
		return response, err
	}

	// 检查缓存
	if params.CacheEnabled {
		if cached, found := f.getFromCache(params.FeedURL); found {
			response.CacheHit = true
			response.Entries = cached.Entries
			response.Feed = cached.Feed
			response.Total = cached.Total
			response.ProcessingTime = time.Since(startTime)
			f.updateStats(true)
			return response, nil
		}
	}

	// 设置默认值
	if params.Timeout == 0 {
		params.Timeout = 30
	}
	if params.Retries == 0 {
		params.Retries = 3
	}
	if params.Delay == 0 {
		params.Delay = 2
	}
	if params.Format == "" {
		params.Format = RSSFormatXML
	}

	// 尝试获取RSS数据
	var data []byte
	var err error

	for attempt := 1; attempt <= params.Retries; attempt++ {
		data, err = f.fetchWithRetry(ctx, params, attempt)
		if err == nil {
			break
		}

		// 如果是最后一次尝试，则不重试
		if attempt == params.Retries {
			response.Success = false
			response.Error = f.createRSSError(err)
			f.updateStats(false)
			return response, err
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return response, ctx.Err()
		case <-time.After(time.Duration(params.Delay) * time.Second):
		}
	}

	// 解析RSS数据
	feed, entries, err := f.parseRSS(data, params.Format)
	if err != nil {
		response.Success = false
		response.Error = &RSSError{
			Type:      RSSErrorTypeParse,
			Message:   "解析RSS数据失败",
			Details:   err.Error(),
			Timestamp: time.Now(),
		}
		f.updateStats(false)
		return response, err
	}

	// 应用过滤器
	filteredEntries := f.applyFilters(entries, params.Filters)
	response.Entries = filteredEntries
	response.Total = len(entries)
	response.Filtered = len(entries) - len(filteredEntries)
	response.Feed = feed

	// 限制数量
	if params.Limit > 0 && len(response.Entries) > params.Limit {
		response.Entries = response.Entries[:params.Limit]
	}

	// 更新源的最后获取时间
	if feed != nil {
		feed.LastFetch = time.Now()
		feed.LastSuccess = time.Now()
		feed.ErrorCount = 0
	}

	// 缓存结果
	if params.CacheEnabled {
		ttl := 30 * time.Minute // 默认30分钟
		if params.CacheTTL > 0 {
			ttl = time.Duration(params.CacheTTL) * time.Minute
		}
		f.cacheResponse(params.FeedURL, response, ttl)
	}

	response.ProcessingTime = time.Since(startTime)
	f.updateStats(true)

	return response, nil
}

// FetchMultipleRSS 批量获取多个RSS源的数据
func (f *RSSFetcherImpl) FetchMultipleRSS(ctx context.Context, paramsList []*FetchRSSParams) ([]*RSSResponse, error) {
	responses := make([]*RSSResponse, len(paramsList))
	wg := &sync.WaitGroup{}
	errCh := make(chan error, len(paramsList))
	mutex := &sync.Mutex{}

	// 并发获取RSS数据
	for i, params := range paramsList {
		wg.Add(1)
		go func(index int, p *FetchRSSParams) {
			defer wg.Done()

			// 创建带超时的上下文
			timeout := 30 * time.Second
			if p.Timeout > 0 {
				timeout = time.Duration(p.Timeout) * time.Second
			}
			subCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			response, err := f.FetchRSS(subCtx, p)
			mutex.Lock()
			responses[index] = response
			mutex.Unlock()

			if err != nil {
				errCh <- fmt.Errorf("fetch rss failed for %s: %w", p.FeedURL, err)
			}
		}(i, params)
	}

	// 等待所有获取完成
	wg.Wait()
	close(errCh)

	// 收集错误信息
	errors := []string{}
	for err := range errCh {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return responses, fmt.Errorf("some rss fetch failed: %s", strings.Join(errors, "; "))
	}

	return responses, nil
}

// ValidateRSSFeed 验证RSS源是否有效
func (f *RSSFetcherImpl) ValidateRSSFeed(ctx context.Context, feedURL string) (bool, error) {
	if feedURL == "" {
		return false, errors.New("feed URL is required")
	}

	// 创建简单的验证参数
	params := &FetchRSSParams{
		FeedURL:  feedURL,
		Timeout:  10,
		Retries:  1,
		Limit:    1, // 只获取一个条目用于验证
	}

	// 尝试获取
	response, err := f.FetchRSS(ctx, params)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

// GetRSSStats 获取RSS统计信息
func (f *RSSFetcherImpl) GetRSSStats(ctx context.Context) (*RSSStats, error) {
	f.statsMutex.Lock()
	defer f.statsMutex.Unlock()

	// 返回统计信息的副本
	stats := *f.stats
	return &stats, nil
}

// ClearCache 清除指定RSS源的缓存
func (f *RSSFetcherImpl) ClearCache(feedURL string) error {
	if feedURL == "" {
		return errors.New("feed URL is required")
	}

	f.cache.Delete(feedURL)
	return nil
}

// 私有辅助方法

// validateFetchParams 验证获取参数
func (f *RSSFetcherImpl) validateFetchParams(params *FetchRSSParams) error {
	if params == nil {
		return errors.New("params cannot be nil")
	}

	if params.FeedURL == "" {
		return errors.New("feed URL is required")
	}

	// 验证URL格式
	_, err := url.ParseRequestURI(params.FeedURL)
	if err != nil {
		return fmt.Errorf("invalid feed URL: %w", err)
	}

	// 验证超时时间
	if params.Timeout < 0 {
		return errors.New("timeout cannot be negative")
	}

	// 验证重试次数
	if params.Retries < 0 {
		return errors.New("retries cannot be negative")
	}

	// 验证延迟时间
	if params.Delay < 0 {
		return errors.New("delay cannot be negative")
	}

	return nil
}

// fetchWithRetry 带重试的HTTP请求
func (f *RSSFetcherImpl) fetchWithRetry(ctx context.Context, params *FetchRSSParams, attempt int) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", params.FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	f.setupRequestHeaders(req, params)

	// 执行请求
	resp, err := f.httpClient.Do(req)
	if err != nil {
		f.logger.Warn("HTTP request failed", 
			zap.String("url", params.FeedURL),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		f.logger.Warn("HTTP request returned non-200 status", 
			zap.String("url", params.FeedURL),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", attempt),
		)
		return nil, err
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// setupRequestHeaders 设置请求头
func (f *RSSFetcherImpl) setupRequestHeaders(req *http.Request, params *FetchRSSParams) {
	// 设置默认User-Agent
	if params.UserAgent != "" {
		req.Header.Set("User-Agent", params.UserAgent)
	} else {
		req.Header.Set("User-Agent", "MoviePilot-RSS-Fetcher/1.0")
	}

	// 设置Accept头
	req.Header.Set("Accept", "application/rss+xml, application/xml, application/json, text/xml, text/html")

	// 设置认证
	if params.Username != "" && params.Password != "" {
		req.SetBasicAuth(params.Username, params.Password)
	}

	// 设置Cookies
	if params.Cookies != "" {
		req.Header.Set("Cookie", params.Cookies)
	}

	// 设置自定义头
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}
}

// parseRSS 解析RSS数据
func (f *RSSFetcherImpl) parseRSS(data []byte, format RSSFormat) (*RSSFeed, []RSSEntry, error) {
	switch format {
	case RSSFormatXML:
		return f.parseRSSXML(data)
	case RSSFormatJSON:
		return f.parseRSSJSON(data)
	case RSSFormatCustom:
		// 自定义格式需要特殊处理，这里简化处理
		return f.parseRSSXML(data)
	default:
		// 自动检测格式
		return f.autoDetectAndParse(data)
	}
}

// parseRSSXML 解析XML格式的RSS
func (f *RSSFetcherImpl) parseRSSXML(data []byte) (*RSSFeed, []RSSEntry, error) {
	// 简化的XML解析实现
	// 实际项目中应该使用更完善的RSS解析库
	var rss struct {
		XMLName     xml.Name `xml:"rss"`
		Version     string   `xml:"version,attr"`
		Channel     struct {
			Title       string     `xml:"title"`
			Link        string     `xml:"link"`
			Description string     `xml:"description"`
			Language    string     `xml:"language"`
			PubDate     string     `xml:"pubDate"`
			LastBuild   string     `xml:"lastBuildDate"`
			Generator   string     `xml:"generator"`
			Items       []struct {
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				Description string `xml:"description"`
				PubDate     string `xml:"pubDate"`
				GUID        string `xml:"guid"`
				Enclosure   struct {
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// 创建Feed对象
	feed := &RSSFeed{
		ID:          f.generateFeedID(rss.Channel.Link),
		Title:       rss.Channel.Title,
		Link:        rss.Channel.Link,
		Description: rss.Channel.Description,
		Language:    rss.Channel.Language,
		Generator:   rss.Channel.Generator,
		Format:      RSSFormatXML,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 解析日期
	if rss.Channel.LastBuild != "" {
		if t, err := f.parseTime(rss.Channel.LastBuild); err == nil {
			feed.LastBuild = t
		}
	}

	// 创建条目
	entries := make([]RSSEntry, 0, len(rss.Channel.Items))
	for _, item := range rss.Channel.Items {
		entry := RSSEntry{
			ID:          f.generateEntryID(item.GUID),
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			GUID:        item.GUID,
		}

		// 解析发布日期
		if item.PubDate != "" {
			if t, err := f.parseTime(item.PubDate); err == nil {
				entry.PublishedAt = t
			}
		}

		// 处理附件
		if item.Enclosure.URL != "" {
			length, _ := f.parseFileSize(item.Enclosure.Length)
			entry.Enclosure = &Enclosure{
				URL:    item.Enclosure.URL,
				Length: length,
				Type:   item.Enclosure.Type,
			}

			// 检测下载链接类型
			if strings.HasSuffix(item.Enclosure.URL, ".torrent") {
				entry.TorrentURL = item.Enclosure.URL
			} else if strings.HasPrefix(item.Enclosure.URL, "magnet:") {
				entry.MagnetURL = item.Enclosure.URL
			}
		}

		// 从标题中提取媒体信息
		f.extractMediaInfoFromTitle(&entry)

		entries = append(entries, entry)
	}

	return feed, entries, nil
}

// parseRSSJSON 解析JSON格式的RSS
func (f *RSSFetcherImpl) parseRSSJSON(data []byte) (*RSSFeed, []RSSEntry, error) {
	// 简化的JSON解析实现
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 这里需要根据具体的JSON格式进行解析
	// 由于没有具体的JSON格式规范，这里返回一些模拟数据
	feed := &RSSFeed{
		ID:          "json-feed-" + time.Now().Format("20060102150405"),
		Title:       "JSON Feed",
		Link:        "https://example.com",
		Description: "JSON格式的RSS源",
		Format:      RSSFormatJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	entries := []RSSEntry{
		{
			ID:          "json-entry-1",
			Title:       "JSON条目示例",
			Link:        "https://example.com/entry1",
			PublishedAt: time.Now(),
		},
	}

	return feed, entries, nil
}

// autoDetectAndParse 自动检测并解析RSS格式
func (f *RSSFetcherImpl) autoDetectAndParse(data []byte) (*RSSFeed, []RSSEntry, error) {
	// 简单的格式检测
	dataStr := string(data)
	dataStr = strings.TrimSpace(dataStr)

	if strings.HasPrefix(dataStr, "{" && strings.HasSuffix(dataStr, "}")) {
		// JSON格式
		return f.parseRSSJSON(data)
	} else if strings.HasPrefix(dataStr, "<?xml") || strings.HasPrefix(dataStr, "<rss") || strings.HasPrefix(dataStr, "<feed") {
		// XML格式
		return f.parseRSSXML(data)
	}

	return nil, nil, errors.New("unable to detect RSS format")
}

// applyFilters 应用过滤器
func (f *RSSFetcherImpl) applyFilters(entries []RSSEntry, filters *RSSFilters) []RSSEntry {
	if filters == nil {
		return entries
	}

	filtered := []RSSEntry{}

	for _, entry := range entries {
		if f.matchesFilters(&entry, filters) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// matchesFilters 检查条目是否匹配过滤条件
func (f *RSSFetcherImpl) matchesFilters(entry *RSSEntry, filters *RSSFilters) bool {
	// 检查包含标题关键词
	if len(filters.IncludeTitle) > 0 {
		match := false
		for _, keyword := range filters.IncludeTitle {
			if strings.Contains(strings.ToLower(entry.Title), strings.ToLower(keyword)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// 检查排除标题关键词
	if len(filters.ExcludeTitle) > 0 {
		for _, keyword := range filters.ExcludeTitle {
			if strings.Contains(strings.ToLower(entry.Title), strings.ToLower(keyword)) {
				return false
			}
		}
	}

	// 检查包含关键词
	content := strings.ToLower(entry.Title + " " + entry.Description)
	if len(filters.IncludeKeywords) > 0 {
		match := false
		for _, keyword := range filters.IncludeKeywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// 检查排除关键词
	if len(filters.ExcludeKeywords) > 0 {
		for _, keyword := range filters.ExcludeKeywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				return false
			}
		}
	}

	// 检查文件大小
	if filters.MinSize > 0 && entry.FileSize < filters.MinSize {
		return false
	}
	if filters.MaxSize > 0 && entry.FileSize > filters.MaxSize {
		return false
	}

	// 检查做种数
	if filters.MinSeeders > 0 && entry.Seeders < filters.MinSeeders {
		return false
	}

	// 检查媒体类型
	if len(filters.MediaTypes) > 0 && entry.MediaType != "" {
		match := false
		for _, mt := range filters.MediaTypes {
			if entry.MediaType == mt {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

// extractMediaInfoFromTitle 从标题中提取媒体信息
func (f *RSSFetcherImpl) extractMediaInfoFromTitle(entry *RSSEntry) {
	// 使用正则表达式从标题中提取信息
	title := entry.Title

	// 检测分辨率
	resolutionRegex := regexp.MustCompile(`(\d{3,4}p)`)
	if matches := resolutionRegex.FindStringSubmatch(title); len(matches) > 0 {
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["resolution"] = matches[1]
	}

	// 检测编码
	codecRegex := regexp.MustCompile(`(H\.264|H\.265|x264|x265|HEVC|AVC)`)
	if matches := codecRegex.FindStringSubmatch(title); len(matches) > 0 {
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["codec"] = matches[1]
	}

	// 检测媒体类型
	if strings.Contains(strings.ToLower(title), "series") || 
	   strings.Contains(strings.ToLower(title), "season") ||
	   strings.Contains(strings.ToLower(title), "s\d+") {
		entry.MediaType = RSSTypeSeries
	} else if strings.Contains(strings.ToLower(title), "anime") {
		entry.MediaType = RSSTypeAnimation
	} else if strings.Contains(strings.ToLower(title), "documentary") {
		entry.MediaType = RSSTypeDocumentary
	} else {
		entry.MediaType = RSSTypeMovie
	}
}

// 缓存相关方法

// getFromCache 从缓存获取数据
func (f *RSSFetcherImpl) getFromCache(feedURL string) (*RSSResponse, bool) {
	cached, found := f.cache.Load(feedURL)
	if !found {
		return nil, false
	}

	cacheItem, ok := cached.(*cacheItem)
	if !ok || cacheItem.expired() {
		f.cache.Delete(feedURL)
		return nil, false
	}

	return cacheItem.data, true
}

// cacheResponse 缓存响应数据
func (f *RSSFetcherImpl) cacheResponse(feedURL string, response *RSSResponse, ttl time.Duration) {
	item := &cacheItem{
		data:      response,
		expiresAt: time.Now().Add(ttl),
	}

	f.cache.Store(feedURL, item)
}

// cacheItem 缓存项结构
type cacheItem struct {
	data      *RSSResponse
	expiresAt time.Time
}

// expired 检查缓存是否过期
func (c *cacheItem) expired() bool {
	return time.Now().After(c.expiresAt)
}

// 工具方法

// generateFeedID 生成Feed ID
func (f *RSSFetcherImpl) generateFeedID(link string) string {
	// 使用简单的哈希算法生成ID
	// 实际项目中可以使用更复杂的算法
	return fmt.Sprintf("feed-%d", f.simpleHash(link))
}

// generateEntryID 生成条目ID
func (f *RSSFetcherImpl) generateEntryID(guid string) string {
	if guid != "" {
		return fmt.Sprintf("entry-%d", f.simpleHash(guid))
	}
	return fmt.Sprintf("entry-%d", time.Now().UnixNano())
}

// simpleHash 简单的哈希函数
func (f *RSSFetcherImpl) simpleHash(s string) int {
	hash := 0
	for _, char := range s {
		hash = 31*hash + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// parseTime 解析时间字符串
func (f *RSSFetcherImpl) parseTime(timeStr string) (time.Time, error) {
	// 尝试多种时间格式
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"02 Jan 2006 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// parseFileSize 解析文件大小字符串
func (f *RSSFetcherImpl) parseFileSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	// 尝试直接解析数字
	var size int64
	if _, err := fmt.Sscanf(sizeStr, "%d", &size); err == nil {
		return size, nil
	}

	return 0, errors.New("unable to parse file size")
}

// createRSSError 创建RSS错误对象
func (f *RSSFetcherImpl) createRSSError(err error) *RSSError {
	errType := RSSErrorTypeUnknown
	message := err.Error()

	// 根据错误信息判断错误类型
	if strings.Contains(strings.ToLower(message), "timeout") {
		errType = RSSErrorTypeTimeout
	} else if strings.Contains(strings.ToLower(message), "network") || 
	          strings.Contains(strings.ToLower(message), "connection") ||
	          strings.Contains(strings.ToLower(message), "http") {
		errType = RSSErrorTypeNetwork
	} else if strings.Contains(strings.ToLower(message), "auth") || 
	          strings.Contains(strings.ToLower(message), "permission") {
		errType = RSSErrorTypeAuth
	} else if strings.Contains(strings.ToLower(message), "parse") || 
	          strings.Contains(strings.ToLower(message), "invalid") {
		errType = RSSErrorTypeParse
	}

	return &RSSError{
		Type:      errType,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// updateStats 更新统计信息
func (f *RSSFetcherImpl) updateStats(success bool) {
	f.statsMutex.Lock()
	defer f.statsMutex.Unlock()

	f.stats.LastUpdated = time.Now()

	// 更新成功率统计
	// 这里简化处理，实际项目中应该维护更详细的统计信息
}
