// Package spider 索引器Spider包
package spider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// NexusPHP NexusPHP站点Spider
type NexusPHP struct {
	*BaseSpider
	loginURL     string
	searchURL    string
	userURL      string
	torrentURL   string
	loggedIn     bool
	lastLogin    time.Time
	sessionID   string
	passkey      string
	uid          string
}

// NewNexusPHP 创建NexusPHP Spider实例
func NewNexusPHP(siteURL string) *NexusPHP {
	base := NewBaseSpider("NexusPHP", siteURL)
	
	return &NexusPHP{
		BaseSpider: base,
		loginURL:   base.BuildURL("login.php", nil),
		searchURL:  base.BuildURL("torrents.php", nil),
		userURL:    base.BuildURL("userdetails.php", nil),
		torrentURL: base.BuildURL("download.php", nil),
	}
}

// Search 搜索种子
func (n *NexusPHP) Search(ctx context.Context, keyword string, filters *SearchFilters) ([]*TorrentItem, error) {
	if err := n.validateLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录验证失败: %w", err)
	}

	if err := n.ValidateSearchFilters(filters); err != nil {
		return nil, fmt.Errorf("搜索过滤器验证失败: %w", err)
	}

	// 构建搜索参数
	params := n.buildSearchParams(keyword, filters)
	searchURL := n.BuildURL("torrents.php", params)

	logger.Info("开始搜索种子",
		n.loggerFields("search", map[string]interface{}{
			"keyword": keyword,
			"filters": filters,
			"url":     searchURL,
		})...)

	// 执行搜索
	var result []*TorrentItem
	err := n.Retry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", n.userAgent)
		req.Header.Set("Referer", n.siteURL)

		resp, err := n.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
		}

		// 解析搜索结果
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("解析HTML失败: %w", err)
		}

		items, err := n.parseSearchResults(doc, filters)
		if err != nil {
			return fmt.Errorf("解析搜索结果失败: %w", err)
		}

		result = items
		return nil
	})

	if err != nil {
		n.LogError("搜索失败", err, map[string]interface{}{
			"keyword": keyword,
			"filters": filters,
		})
		return nil, err
	}

	logger.Info("搜索完成",
		n.loggerFields("search", map[string]interface{}{
			"keyword": keyword,
			"count":   len(result),
		})...)

	return result, nil
}

// buildSearchParams 构建搜索参数
func (n *NexusPHP) buildSearchParams(keyword string, filters *SearchFilters) map[string]string {
	params := make(map[string]string)

	if keyword != "" {
		params["search"] = n.NormalizeKeyword(keyword)
		params["search_type"] = "all" // 标题+描述
	}

	// 分类过滤
	if filters != nil && filters.Category != "" {
		params["cat"] = n.getCategoryID(filters.Category)
	}

	// 分辨率过滤
	if filters != nil && filters.Resolution != "" {
		params["resolution"] = n.getResolutionID(filters.Resolution)
	}

	// 排序
	sortBy := "time"
	sortOrder := "desc"
	if filters != nil {
		if filters.SortBy != "" {
			sortBy = filters.SortBy
		}
		if filters.SortOrder != "" {
			sortOrder = filters.SortOrder
		}
	}
	params["order"] = sortBy
	params["by"] = sortOrder

	// 分页
	if filters != nil {
		if filters.Page > 1 {
			params["page"] = strconv.Itoa(filters.Page)
		}
		if filters.Limit > 0 {
			params["limit"] = strconv.Itoa(filters.Limit)
		}
	}

	return params
}

// parseSearchResults 解析搜索结果
func (n *NexusPHP) parseSearchResults(doc *goquery.Document, filters *SearchFilters) ([]*TorrentItem, error) {
	var items []*TorrentItem

	// NexusPHP标准的种子列表表格
	doc.Find("table.torrents > tbody > tr").Each(func(i int, s *goquery.Selection) {
		// 跳过表头和空行
		if s.Find("td").Length() == 0 {
			return
		}

		item, err := n.parseTorrentRow(s)
		if err != nil {
			logger.Warn("解析种子行失败", 
				n.loggerFields("parse", map[string]interface{}{
					"row": i + 1,
					"error": err.Error(),
				})...)
			return
		}

		// 应用过滤器
		if filters != nil && !n.matchFilters(item, filters) {
			return
		}

		items = append(items, item)
	})

	return items, nil
}

// parseTorrentRow 解析种子行
func (n *NexusPHP) parseTorrentRow(row *goquery.Selection) (*TorrentItem, error) {
	cells := row.Find("td")
	if cells.Length() < 10 {
		return nil, fmt.Errorf("种子行列数不足")
	}

	item := &TorrentItem{}

	// 解析分类
	categoryCell := cells.Eq(0).Find("img")
	if categoryCell.Length() > 0 {
		categoryCell.Each(func(i int, s *goquery.Selection) {
			if alt, exists := s.Attr("alt"); exists {
				item.Category = alt
			}
			if title, exists := s.Attr("title"); exists {
				item.SubCategory = title
			}
		})
	}

	// 解析标题和链接
	titleCell := cells.Eq(1).Find("a")
	if titleCell.Length() > 0 {
		titleLink := titleCell.First()
		item.Title = strings.TrimSpace(titleLink.Text())
		
		if href, exists := titleLink.Attr("href"); exists {
			if parsedURL, err := url.Parse(href); err == nil {
				if values, err := url.ParseQuery(parsedURL.RawQuery); err == nil {
					if id := values.Get("id"); id != "" {
						item.ID = id
					}
				}
			}
		}
	}

	// 解析下载链接
	downloadLink := cells.Eq(2).Find("a")
	if downloadLink.Length() > 0 {
		if href, exists := downloadLink.Attr("href"); exists {
			item.Meta = make(map[string]string)
			item.Meta["download_url"] = n.BuildURL(strings.TrimPrefix(href, "/"), nil)
		}
	}

	// 解析大小
	sizeCell := cells.Eq(5)
	if sizeText := strings.TrimSpace(sizeCell.Text()); sizeText != "" {
		if size, err := ParseSize(sizeText); err == nil {
			item.Size = size
		}
	}

	// 解析种子数和下载数
	seedersCell := cells.Eq(6)
	if seedersText := strings.TrimSpace(seedersCell.Text()); seedersText != "" {
		if seeders, err := strconv.Atoi(seedersText); err == nil {
			item.Seeders = seeders
		}
	}

	leechersCell := cells.Eq(7)
	if leechersText := strings.TrimSpace(leechersCell.Text()); leechersText != "" {
		if leechers, err := strconv.Atoi(leechersText); err == nil {
			item.Leechers = leechers
		}
	}

	// 解析上传时间
	timeCell := cells.Eq(8)
	if timeText := strings.TrimSpace(timeCell.Text()); timeText != "" {
		if uploadedAt, err := n.parseTime(timeText); err == nil {
			item.UploadedAt = uploadedAt
		}
	}

	// 解析上传者
	uploaderCell := cells.Eq(9)
	if uploaderText := strings.TrimSpace(uploaderCell.Text()); uploaderText != "" {
		item.Meta = make(map[string]string)
		item.Meta["uploader"] = uploaderText
	}

	// 解析特殊标识
	n.parseSpecialTags(row, item)

	return item, nil
}

// parseSpecialTags 解析特殊标签
func (n *NexusPHP) parseSpecialTags(row *goquery.Selection, item *TorrentItem) {
	// 检查免费标识
	if row.Find("img[alt*='Free'], img[alt*='免费']").Length() > 0 {
		item.FreeLeech = true
		item.DownloadFactor = 0
	}

	// 检查推广标识
	if row.Find("img[alt*='2X'], img[alt*='50%'], img[alt*='上传']").Length() > 0 {
		item.Promotional = true
		item.UploadFactor = 2.0
	}

	// 检查HDR标识
	if strings.Contains(strings.ToLower(item.Title), "hdr") || 
	   row.Find("span:contains('HDR'), span:contains('hdr')").Length() > 0 {
		item.HDR = true
	}

	// 检查配音标识
	if strings.Contains(strings.ToUpper(item.Title), "DUB") ||
	   row.Find("span:contains('配音'), span:contains('国语')").Length() > 0 {
		item.Dubbed = true
	}

	// 检查字幕标识
	if strings.Contains(strings.ToUpper(item.Title), "SUB") ||
	   row.Find("span:contains('字幕'), span:contains('中字')").Length() > 0 {
		item.Subtitled = true
	}
}

// GetTorrentDetails 获取种子详情
func (n *NexusPHP) GetTorrentDetails(ctx context.Context, torrentID string) (*TorrentDetail, error) {
	if err := n.validateLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录验证失败: %w", err)
	}

	detailURL := n.BuildURL("details.php", map[string]string{
		"id":  torrentID,
		"hit": "1",
	})

	logger.Info("获取种子详情",
		n.loggerFields("details", map[string]interface{}{
			"torrent_id": torrentID,
			"url":        detailURL,
		})...)

	var detail *TorrentDetail
	err := n.Retry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", detailURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", n.userAgent)
		req.Header.Set("Referer", n.siteURL)

		resp, err := n.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("解析HTML失败: %w", err)
		}

		parsedDetail, err := n.parseDetailPage(doc, torrentID)
		if err != nil {
			return fmt.Errorf("解析详情页失败: %w", err)
		}

		detail = parsedDetail
		return nil
	})

	if err != nil {
		n.LogError("获取种子详情失败", err, map[string]interface{}{
			"torrent_id": torrentID,
		})
		return nil, err
	}

	logger.Info("种子详情获取成功",
		n.loggerFields("details", map[string]interface{}{
			"torrent_id": torrentID,
			"title":      detail.Title,
		})...)

	return detail, nil
}

// parseDetailPage 解析详情页
func (n *NexusPHP) parseDetailPage(doc *goquery.Document, torrentID string) (*TorrentDetail, error) {
	detail := &TorrentDetail{
		ID:     torrentID,
		Meta:   make(map[string]interface{}),
		Files:  make([]*TorrentFile, 0),
		Pictures: make([]string, 0),
	}

	// 解析标题
	titleText := doc.Find("h1").First().Text()
	detail.Title = strings.TrimSpace(titleText)

	// 解析描述
	descText := doc.Find("#description").First().Text()
	detail.Description = strings.TrimSpace(descText)

	// 解析基本信息
	n.parseBasicInfo(doc, detail)

	// 解析文件列表
	n.parseFileList(doc, detail)

	// 解析图片
	n.parsePictures(doc, detail)

	// 解析评论
	n.parseComments(doc, detail)

	return detail, nil
}

// parseBasicInfo 解析基本信息
func (n *NexusPHP) parseBasicInfo(doc *goquery.Document, detail *TorrentDetail) {
	// 解析大小
	doc.Find("td:contains('大小') + td").Each(func(i int, s *goquery.Selection) {
		if sizeText := strings.TrimSpace(s.Text()); sizeText != "" {
			if size, err := ParseSize(sizeText); err == nil {
				detail.Size = size
			}
		}
	})

	// 解析种子数和下载数
	doc.Find("td:contains('种子数') + td").Each(func(i int, s *goquery.Selection) {
		if seedersText := strings.TrimSpace(s.Text()); seedersText != "" {
			if seeders, err := strconv.Atoi(seedersText); err == nil {
				detail.Seeders = seeders
			}
		}
	})

	doc.Find("td:contains('下载数') + td").Each(func(i int, s *goquery.Selection) {
		if leechersText := strings.TrimSpace(s.Text()); leechersText != "" {
			if leechers, err := strconv.Atoi(leechersText); err == nil {
				detail.Leechers = leechers
			}
		}
	})

	// 解析下载链接
	downloadLink := doc.Find("a[href*='download.php']").First()
	if downloadLink.Length() > 0 {
		if href, exists := downloadLink.Attr("href"); exists {
			detail.TorrentURL = n.BuildURL(strings.TrimPrefix(href, "/"), nil)
		}
	}

	// 解析磁力链接
	magnetLink := doc.Find("a[href^='magnet:']").First()
	if magnetLink.Length() > 0 {
		if href, exists := magnetLink.Attr("href"); exists {
			detail.MagnetURL = href
		}
	}
}

// parseFileList 解析文件列表
func (n *NexusPHP) parseFileList(doc *goquery.Document, detail *TorrentDetail) {
	doc.Find("#filelist tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() < 2 {
			return
		}

		file := &TorrentFile{}
		
		// 文件路径
		file.Path = strings.TrimSpace(cells.Eq(0).Text())
		
		// 文件大小
		sizeText := strings.TrimSpace(cells.Eq(1).Text())
		if size, err := ParseSize(sizeText); err == nil {
			file.Size = size
		}

		detail.Files = append(detail.Files, file)
	})
}

// parsePictures 解析图片
func (n *NexusPHP) parsePictures(doc *goquery.Document, detail *TorrentDetail) {
	doc.Find(".postimg img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			detail.Pictures = append(detail.Pictures, n.BuildURL(strings.TrimPrefix(src, "/"), nil))
		}
	})
}

// parseComments 解析评论
func (n *NexusPHP) parseComments(doc *goquery.Document, detail *TorrentDetail) {
	doc.Find(".comment").Each(func(i int, s *goquery.Selection) {
		comment := &Comment{}
		
		// 评论ID
		if id, exists := s.Attr("id"); exists {
			comment.ID = id
		}
		
		// 用户信息
		userCell := s.Find(".comment_user")
		if userCell.Length() > 0 {
			comment.Username = strings.TrimSpace(userCell.Text())
		}
		
		// 评论内容
		contentCell := s.Find(".comment_content")
		if contentCell.Length() > 0 {
			comment.Content = strings.TrimSpace(contentCell.Text())
		}
		
		// 时间
		timeCell := s.Find(".comment_time")
		if timeCell.Length() > 0 {
			if timeText := strings.TrimSpace(timeCell.Text()); timeText != "" {
				if createdAt, err := n.parseTime(timeText); err == nil {
					comment.CreatedAt = createdAt
				}
			}
		}
		
		detail.CommentsList = append(detail.CommentsList, comment)
	})
}

// DownloadTorrent 下载种子文件
func (n *NexusPHP) DownloadTorrent(ctx context.Context, torrentID string) ([]byte, error) {
	if err := n.validateLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录验证失败: %w", err)
	}

	downloadURL := n.BuildURL("download.php", map[string]string{
		"id": torrentID,
	})

	logger.Info("下载种子文件",
		n.loggerFields("download", map[string]interface{}{
			"torrent_id": torrentID,
			"url":        downloadURL,
		})...)

	var data []byte
	err := n.Retry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", n.userAgent)
		req.Header.Set("Referer", n.siteURL)

		resp, err := n.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
		}

		// 读取响应数据
		buf := make([]byte, resp.ContentLength)
		if _, err := resp.Body.Read(buf); err != nil {
			return fmt.Errorf("读取种子文件失败: %w", err)
		}

		data = buf
		return nil
	})

	if err != nil {
		n.LogError("下载种子文件失败", err, map[string]interface{}{
			"torrent_id": torrentID,
		})
		return nil, err
	}

	logger.Info("种子文件下载成功",
		n.loggerFields("download", map[string]interface{}{
			"torrent_id": torrentID,
			"size":       len(data),
		})...)

	return data, nil
}

// validateLogin 验证登录状态
func (n *NexusPHP) validateLogin(ctx context.Context) error {
	// 检查是否已登录且会话仍然有效
	if n.loggedIn && time.Since(n.lastLogin) < 30*time.Minute {
		return nil
	}

	// 执行登录验证
	return n.login(ctx)
}

// login 执行登录
func (n *NexusPHP) login(ctx context.Context) error {
	// 这里需要具体的登录逻辑
	// 包括：获取登录表单、提交用户名密码、处理验证码等
	logger.Info("NexusPHP登录验证", 
		n.loggerFields("login", map[string]interface{}{
			"site": n.siteURL,
		})...)

	// 模拟登录成功
	n.loggedIn = true
	n.lastLogin = time.Now()

	return nil
}

// 辅助方法

// getCategoryID 获取分类ID
func (n *NexusPHP) getCategoryID(category string) string {
	categoryMap := map[string]string{
		"movie":  "401",
		"tv":     "402",
		"anime":  "403",
		"music":   "404",
		"game":   "405",
		"app":    "406",
		"other":  "407",
	}
	
	if id, exists := categoryMap[category]; exists {
		return id
	}
	return ""
}

// getResolutionID 获取分辨率ID
func (n *NexusPHP) getResolutionID(resolution string) string {
	resolutionMap := map[string]string{
		"4K":   "2160p",
		"1080p": "1080p",
		"720p":  "720p",
		"480p":  "480p",
	}
	
	if id, exists := resolutionMap[resolution]; exists {
		return id
	}
	return ""
}

// matchFilters 检查是否匹配过滤器
func (n *NexusPHP) matchFilters(item *TorrentItem, filters *SearchFilters) bool {
	// 大小过滤
	if filters.SizeRange != nil {
		if item.Size < filters.SizeRange.Min || item.Size > filters.SizeRange.Max {
			return false
		}
	}

	// 最少做种数过滤
	if filters.MinSeeders > 0 && item.Seeders < filters.MinSeeders {
		return false
	}

	// 最少下载数过滤
	if filters.MinLeechers > 0 && item.Leechers < filters.MinLeechers {
		return false
	}

	// 免费种子过滤
	if filters.FreeLeech && !item.FreeLeech {
		return false
	}

	// HDR过滤
	if filters.HDR && !item.HDR {
		return false
	}

	// 配音过滤
	if filters.Dubbed && !item.Dubbed {
		return false
	}

	// 字幕过滤
	if filters.Subtitled && !item.Subtitled {
		return false
	}

	// 关键词过滤
	title := strings.ToLower(item.Title)
	
	for _, word := range filters.ExcludeWords {
		if strings.Contains(title, strings.ToLower(word)) {
			return false
		}
	}
	
	for _, word := range filters.IncludeWords {
		if !strings.Contains(title, strings.ToLower(word)) {
			return false
		}
	}

	return true
}

// parseTime 解析时间字符串
func (n *NexusPHP) parseTime(timeStr string) (time.Time, error) {
	// 处理相对时间
	if strings.Contains(timeStr, "前") || strings.Contains(timeStr, "ago") {
		return time.Now(), nil // 简化处理
	}
	
	// 处理标准时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	
	return time.Now(), fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// loggerFields 构建日志字段
func (n *NexusPHP) loggerFields(operation string, fields map[string]interface{}) []zap.Field {
	result := []zap.Field{
		zap.String("spider", n.name),
		zap.String("operation", operation),
		zap.String("site", n.siteURL),
	}
	
	for k, v := range fields {
		result = append(result, zap.Any(k, v))
	}
	
	return result
}

// 其他方法实现...

// GetUserTorrents 获取用户种子列表
func (n *NexusPHP) GetUserTorrents(ctx context.Context, userID string) ([]*TorrentItem, error) {
	// TODO: 实现用户种子列表获取
	return nil, fmt.Errorf("方法未实现")
}

// GetUserInfo 获取用户信息
func (n *NexusPHP) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	// TODO: 实现用户信息获取
	return nil, fmt.Errorf("方法未实现")
}