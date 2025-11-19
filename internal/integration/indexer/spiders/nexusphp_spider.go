// Package spiders NexusPHP站点爬虫
package spiders

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/indexer"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"

	"go.uber.org/zap"
)

// NexusPHPSpider NexusPHP站点爬虫
type NexusPHPSpider struct {
	*BaseSpider
	categoryMap map[string]string
	logger      *zap.Logger
}

// NewNexusPHPSpider 创建NexusPHP爬虫
func NewNexusPHPSpider(name, domain string) *NexusPHPSpider {
	spider := &NexusPHPSpider{
		BaseSpider: NewBaseSpider(name, domain),
		categoryMap: map[string]string{
			"movie":    "401", // 电影
			"tv":       "403", // 剧集
			"documentary": "404", // 纪录片
			"anime":    "405", // 动画
			"music":    "406", // 音乐
			"game":     "407", // 游戏
			"software": "408", // 软件
			"other":    "409", // 其他
		},
		logger: logger.Logger,
	}

	// 设置NexusPHP特有的Headers
	spider.SetHeaders(map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Connection":      "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	})

	return spider
}

// Login NexusPHP登录实现
func (n *NexusPHPSpider) Login(ctx context.Context, username, password string) error {
	// 先获取登录页面以获取必要的token和cookies
	loginPage, err := n.GetPageContent(ctx, "/login.php", nil)
	if err != nil {
		return fmt.Errorf("get login page failed: %w", err)
	}

	// 提取CSRF token
	token := n.ExtractFormToken(loginPage)
	if token == "" {
		n.logger.Warn("No CSRF token found in login page")
	}

	// 提取验证码（如果有）
	captchaURL := n.ExtractCaptcha(loginPage)
	if captchaURL != "" {
		captchaURL = n.GetBaseURL() + captchaURL
		captchaText, err := n.SolveCaptcha(ctx, captchaURL)
		if err != nil {
			n.logger.Warn("Captcha solving failed", zap.Error(err))
		} else {
			// 如果有验证码，添加到登录数据
			return n.loginWithCaptcha(ctx, username, password, captchaText, token)
		}
	}

	// 构建登录数据
	loginData := map[string]string{
		"username": username,
		"password": password,
		"track":    "yes", // 保持登录状态
		"submit":   "登录",
	}

	if token != "" {
		loginData["token"] = token
	}

	// 提交登录请求
	resp, err := n.Post(ctx, "/takelogin.php", loginData)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查登录结果
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// 提取cookies
	for _, cookie := range resp.Cookies() {
		n.cookies[cookie.Name] = cookie.Value
	}
	n.updateHttpClient()

	// 验证是否登录成功
	if !n.IsLoggedIn(ctx) {
		return fmt.Errorf("login failed - still not authenticated")
	}

	n.logger.Info("NexusPHP login successful",
		zap.String("site", n.name),
		zap.String("username", username))

	return nil
}

// loginWithCaptcha 带验证码的登录
func (n *NexusPHPSpider) loginWithCaptcha(ctx context.Context, username, password, captcha, token string) error {
	loginData := map[string]string{
		"username": username,
		"password": password,
		"captcha_string": captcha,
		"track":    "yes",
		"submit":   "登录",
	}

	if token != "" {
		loginData["token"] = token
	}

	resp, err := n.Post(ctx, "/takelogin.php", loginData)
	if err != nil {
		return fmt.Errorf("login with captcha failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login with captcha failed with status: %d", resp.StatusCode)
	}

	// 提取cookies
	for _, cookie := range resp.Cookies() {
		n.cookies[cookie.Name] = cookie.Value
	}
	n.updateHttpClient()

	return nil
}

// Search NexusPHP搜索实现
func (n *NexusPHPSpider) Search(ctx context.Context, keyword, mediaType string) ([]*indexer.TorrentInfo, error) {
	// 添加速率限制
	n.RateLimit(ctx)

	// 获取分类ID
	catID := n.getCategoryID(mediaType)
	
	// 构建搜索参数
	searchData := map[string]string{
		"search":    keyword,
		"cat":       catID,
		"search_in": "title",
		"dead":      "0", // 不包含死种
		"incldead":  "0",
		"spstate":    "0", // 包含正常种子
		"sort":      "added", // 按添加时间排序
		"type":      "desc",  // 降序
	}

	// 搜索种子
	html, err := n.PostPageContent(ctx, "/torrents.php", searchData)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// 解析搜索结果
	return n.ParseTorrentList(html)
}

// ParseTorrentList 解析种子列表
func (n *NexusPHPSpider) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// 使用正则表达式解析种子信息
	// NexusPHP标准模式的种子行
	torrentRegex := regexp.MustCompile(`<tr[^>]*class="torrent[^"]*"[^>]*>(.*?)</tr>`)
	matches := torrentRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		torrent := n.parseTorrentRow(match[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	if len(torrents) == 0 {
		n.logger.Warn("No torrents found in search results")
	}

	return torrents, nil
}

// parseTorrentRow 解析单个种子行
func (n *NexusPHPSpider) parseTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 提取种子ID和详情链接
	idRegex := regexp.MustCompile(`<a[^>]*href=["']details\.php\?id=(\d+)["'][^>]*>`)
	idMatches := idRegex.FindStringSubmatch(rowHTML)
	if len(idMatches) < 2 {
		return nil
	}

	id := idMatches[1]

	// 提取标题
	titleRegex := regexp.MustCompile(`<a[^>]*class="torrentname"[^>]*>(.*?)</a>`)
	titleMatches := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatches) < 2 {
		return nil
	}

	title := n.cleanTitle(titleMatches[1])

	// 提取分类
	categoryRegex := regexp.MustCompile(`<a[^>]*href=["']browse\.php\?cat=\d+["'][^>]*>(.*?)</a>`)
	categoryMatches := categoryRegex.FindStringSubmatch(rowHTML)
	category := ""
	if len(categoryMatches) >= 2 {
		category = n.cleanTitle(categoryMatches[1])
	}

	// 提取大小
	sizeRegex := regexp.MustCompile(`(\d+\.?\d*)\s*(bytes|KB|MB|GB|TB)`)
	sizeMatches := sizeRegex.FindStringSubmatch(rowHTML)
	var size int64
	if len(sizeMatches) >= 3 {
		size = n.parseSize(sizeMatches[1], sizeMatches[2])
	}

	// 提取种子数、下载数、完成数
	seeders := n.extractNumberFromHTML(rowHTML, "seeders")
	leechers := n.extractNumberFromHTML(rowHTML, "leechers")
	completed := n.extractNumberFromHTML(rowHTML, "completed")

	// 提取上传时间
	uploadDate := n.extractUploadDate(rowHTML)

	// 提取发布者
	uploaderRegex := regexp.MustCompile(`<a[^>]*href=["']userdetails\.php\?id=\d+["'][^>]*>(.*?)</a>`)
	uploaderMatches := uploaderRegex.FindStringSubmatch(rowHTML)
	uploader := ""
	if len(uploaderMatches) >= 2 {
		uploader = n.cleanTitle(uploaderMatches[1])
	}

	// 提取下载链接
	downloadRegex := regexp.MustCompile(`<a[^>]*href=["']download\.php\?id=\d+["'][^>]*>`)
	downloadMatches := downloadRegex.FindStringSubmatch(rowHTML)
	downloadURL := ""
	if len(downloadMatches) >= 1 {
		downloadURL = n.GetBaseURL() + downloadMatches[0]
		downloadRegex2 := regexp.MustCompile(`href=["']([^"']+)["']`)
		downloadMatches2 := downloadRegex2.FindStringSubmatch(downloadMatches[0])
		if len(downloadMatches2) >= 2 {
			downloadURL = n.GetBaseURL() + downloadMatches2[1]
		}
	}

	// 检查免费种子、双倍上传等状态
	freeTorrent := strings.Contains(rowHTML, "free") || strings.Contains(rowHTML, "promotion")
	doubleUpload := strings.Contains(rowHTML, "2x") || strings.Contains(rowHTML, "double")

	// 提取IMDB ID
	imdbID := n.extractIMDBID(rowHTML)

	return &indexer.TorrentInfo{
		ID:           id,
		Title:        title,
		Description:  "",
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    completed,
		UploadDate:   uploadDate,
		DownloadURL:  downloadURL,
		DetailURL:    fmt.Sprintf("%s/details.php?id=%s", n.GetBaseURL(), id),
		Category:     category,
		Tags:         n.extractTags(rowHTML),
		Uploader:     uploader,
		IMDBID:       imdbID,
		FreeTorrent:  freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// GetTorrentDetail 获取种子详情
func (n *NexusPHPSpider) GetTorrentDetail(ctx context.Context, id string) (*indexer.TorrentDetail, error) {
	n.RateLimit(ctx)

	path := fmt.Sprintf("/details.php?id=%s", id)
	html, err := n.GetPageContent(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get torrent detail failed: %w", err)
	}

	return n.ParseTorrentDetail(html), nil
}

// ParseTorrentDetail 解析种子详情
func (n *NexusPHPSpider) ParseTorrentDetail(html string) *indexer.TorrentDetail {
	detail := &indexer.TorrentDetail{
		Files:     []*indexer.TorrentFile{},
		Comments:  []*indexer.Comment{},
	}

	// 提取基本信息
	// 这里可以添加更多详情页面的解析逻辑

	return detail
}

// GetUserInfo 获取用户信息
func (n *NexusPHPSpider) GetUserInfo(ctx context.Context) (*indexer.UserInfo, error) {
	n.RateLimit(ctx)

	html, err := n.GetPageContent(ctx, "/user.php", nil)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	return n.ParseUserInfo(html), nil
}

// ParseUserInfo 解析用户信息
func (n *NexusPHPSpider) ParseUserInfo(html string) *indexer.UserInfo {
	userInfo := &indexer.UserInfo{}

	// 提取用户名
	usernameRegex := regexp.MustCompile(`(?:欢迎回来|Hello),\s*([^<\s]+)`)
	matches := usernameRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Username = matches[1]
	}

	// 提取用户等级
	rankRegex := regexp.MustCompile(`等级[：:]\s*([^<\s]+)`)
	matches = rankRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Rank = matches[1]
	}

	// 提取上传量
	uploadRegex := regexp.MustCompile(`上传量[：:]\s*([\d.]+\s*[KMGT]?B)`)
	matches = uploadRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Upload = n.parseSizeFromString(matches[1])
	}

	// 提取下载量
	downloadRegex := regexp.MustCompile(`下载量[：:]\s*([\d.]+\s*[KMGT]?B)`)
	matches = downloadRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Download = n.parseSizeFromString(matches[1])
	}

	// 提取分享率
	ratioRegex := regexp.MustCompile(`分享率[：:]\s*([\d.]+)`)
	matches = ratioRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if ratio, err := strconv.ParseFloat(matches[1], 64); err == nil {
			userInfo.Ratio = ratio
		}
	}

	return userInfo
}

// Download 下载种子
func (n *NexusPHPSpider) Download(ctx context.Context, id string) ([]byte, error) {
	n.RateLimit(ctx)

	path := fmt.Sprintf("/download.php?id=%s", id)
	return n.BaseSpider.Download(ctx, id)
}

// 辅助方法

// getCategoryID 获取分类ID
func (n *NexusPHPSpider) getCategoryID(mediaType string) string {
	if catID, exists := n.categoryMap[mediaType]; exists {
		return catID
	}
	return "0" // 所有分类
}

// cleanTitle 清理标题
func (n *NexusPHPSpider) cleanTitle(title string) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	title = re.ReplaceAllString(title, "")
	
	// 移除多余的空白字符
	re = regexp.MustCompile(`\s+`)
	title = re.ReplaceAllString(title, " ")
	
	// 去除首尾空格
	return strings.TrimSpace(title)
}

// parseSize 解析大小
func (n *NexusPHPSpider) parseSize(size string, unit string) int64 {
	sizeFloat, _ := strconv.ParseFloat(size, 64)
	
	multiplier := int64(1)
	switch strings.ToUpper(unit) {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	
	return int64(sizeFloat * float64(multiplier))
}

// parseSizeFromString 从字符串解析大小
func (n *NexusPHPSpider) parseSizeFromString(sizeStr string) int64 {
	sizeRegex := regexp.MustCompile(`(\d+\.?\d*)\s*(bytes|KB|MB|GB|TB)`)
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) >= 3 {
		return n.parseSize(matches[1], matches[2])
	}
	return 0
}

// extractNumberFromHTML 从HTML中提取数字
func (n *NexusPHPSpider) extractNumberFromHTML(html, className string) int {
	regex := regexp.MustCompile(fmt.Sprintf(`<[^>]*class="[^"]*%s[^"]*"[^>]*>(\d+)`, className))
	matches := regex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return 0
}

// extractUploadDate 提取上传时间
func (n *NexusPHPSpider) extractUploadDate(html string) time.Time {
	// NexusPHP时间格式通常为 "2023-12-01 12:34:56"
	dateRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	matches := dateRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if t, err := time.Parse("2006-01-02 15:04:05", matches[1]); err == nil {
			return t
		}
	}
	
	// 尝试其他时间格式
	formats := []string{
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, matches[1]); err == nil {
			return t
		}
	}
	
	return time.Now()
}

// extractIMDBID 提取IMDB ID
func (n *NexusPHPSpider) extractIMDBID(html string) string {
	imdbRegex := regexp.MustCompile(`(?:imdb\.com/title/tt|tt)(\d{7,8})`)
	matches := imdbRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractTags 提取标签
func (n *NexusPHPSpider) extractTags(html string) []string {
	var tags []string
	
	// 查找常见的标签标识
	tagIdentifiers := []string{
		"free", "2x", "50%", "30%", "hot", "new", "recommended",
	}
	
	for _, identifier := range tagIdentifiers {
		if strings.Contains(html, identifier) {
			tags = append(tags, identifier)
		}
	}
	
	return tags
}

// IsLoggedIn 检查登录状态
func (n *NexusPHPSpider) IsLoggedIn(ctx context.Context) bool {
	// 检查用户页面是否可访问
	html, err := n.GetPageContent(ctx, "/user.php", nil)
	if err != nil {
		return false
	}
	
	// 检查是否包含登录页面的特征
	if strings.Contains(html, "login") || strings.Contains(html, "登录") {
		return false
	}
	
	// 检查是否包含用户信息的特征
	return strings.Contains(html, "ratio") || strings.Contains(html, "分享率") || 
		   strings.Contains(html, "upload") || strings.Contains(html, "上传量")
}