// Package spiders Gazelle站点爬虫
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

	"go.uber.org/zap"
)

// GazelleSpider Gazelle站点爬虫（如Red, OPS等）
type GazelleSpider struct {
	*BaseSpider
	authKey     string
	userAgent   string
	logger      *zap.Logger
}

// NewGazelleSpider 创建Gazelle爬虫
func NewGazelleSpider(name, domain string) *GazelleSpider {
	spider := &GazelleSpider{
		BaseSpider: NewBaseSpider(name, domain),
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		logger:      logger.Logger,
	}

	// 设置Gazelle特有的Headers
	spider.SetHeaders(map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Connection":      "keep-alive",
	})

	return spider
}

// Login Gazelle登录实现
func (g *GazelleSpider) Login(ctx context.Context, username, password string) error {
	// Gazelle通常使用登录表单
	loginData := map[string]string{
		"username": username,
		"password": password,
		"keeplogged": "1", // 保持登录
		"login":     "Log In",
	}

	// 提交登录请求
	resp, err := g.Post(ctx, "/login.php", loginData)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查登录结果
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// 提取cookies和authkey
	for _, cookie := range resp.Cookies() {
		g.cookies[cookie.Name] = cookie.Value
	}
	g.updateHttpClient()

	// 获取用户信息页面以提取authkey
	userHTML, err := g.GetPageContent(ctx, "/index.php", nil)
	if err == nil {
		g.authKey = g.extractAuthKey(userHTML)
	}

	// 验证是否登录成功
	if !g.IsLoggedIn(ctx) {
		return fmt.Errorf("login failed - still not authenticated")
	}

	g.logger.Info("Gazelle login successful",
		zap.String("site", g.name),
		zap.String("username", username))

	return nil
}

// extractAuthKey 提取认证密钥
func (g *GazelleSpider) extractAuthKey(html string) string {
	authkeyRegex := regexp.MustCompile(`authkey["']?\s*[:=]\s*["']([a-zA-Z0-9]+)["']`)
	matches := authkeyRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	
	// 尝试从JavaScript中提取
	jsRegex := regexp.MustCompile(`var\s+authkey\s*=\s*["']([^"']+)["']`)
	matches = jsRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	
	return ""
}

// Search Gazelle搜索实现
func (g *GazelleSpider) Search(ctx context.Context, keyword, mediaType string) ([]*indexer.TorrentInfo, error) {
	g.RateLimit(ctx)

	// 构建搜索参数
	searchData := map[string]string{
		"searchstr": keyword,
		"action":    "basic",
		"order_by":  "time",
		"order_way": "desc",
	}

	// 根据媒体类型设置搜索类型
	switch mediaType {
	case "movie":
		searchData["filter_cat[1]"] = "1" // Movies
	case "music":
		searchData["filter_cat[2]"] = "1" // Music
	case "game":
		searchData["filter_cat[3]"] = "1" // Games
	case "app":
		searchData["filter_cat[4]"] = "1" // Applications
	case "ebook":
		searchData["filter_cat[5]"] = "1" // E-Books
	case "comic":
		searchData["filter_cat[6]"] = "1" // Comics
	case "tutorial":
		searchData["filter_cat[7]"] = "1" // Tutorials
	}

	// 搜索种子
	html, err := g.PostPageContent(ctx, "/torrents.php", searchData)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// 解析搜索结果
	return g.ParseTorrentList(html)
}

// ParseTorrentList 解析种子列表
func (g *GazelleSpider) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// Gazelle使用表格结构
	rowRegex := regexp.MustCompile(`<tr[^>]*class="torrent[^"]*"[^>]*>(.*?)</tr>`)
	matches := rowRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		torrent := g.parseTorrentRow(match[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	if len(torrents) == 0 {
		g.logger.Warn("No torrents found in Gazelle search results")
	}

	return torrents, nil
}

// parseTorrentRow 解析单个种子行
func (g *GazelleSpider) parseTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 提取种子ID
	idRegex := regexp.MustCompile(`torrents\.php\?(?:\w+=\w+&)*id=(\d+)`)
	idMatches := idRegex.FindStringSubmatch(rowHTML)
	if len(idMatches) < 2 {
		return nil
	}

	id := idMatches[1]

	// 提取标题
	titleRegex := regexp.MustCompile(`<a[^>]*href=["']torrents\.php\?[^"']*id=\d+[^"']*["'][^>]*>([^<]+)</a>`)
	titleMatches := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatches) < 2 {
		return nil
	}

	title := g.cleanTitle(titleMatches[1])

	// 提取艺术家（音乐类别）
	artistRegex := regexp.MustCompile(`<a[^>]*href=["']artist\.php\?[^"']*["'][^>]*>([^<]+)</a>`)
	artistMatches := artistRegex.FindStringSubmatch(rowHTML)
	if len(artistMatches) >= 2 {
		title = g.cleanTitle(artistMatches[1]) + " - " + title
	}

	// 提取大小
	sizeRegex := regexp.MustCompile(`(\d+\.?\d*)\s*(bytes|KB|MB|GB|TB)`)
	sizeMatches := sizeRegex.FindStringSubmatch(rowHTML)
	var size int64
	if len(sizeMatches) >= 3 {
		size = g.parseSize(sizeMatches[1], sizeMatches[2])
	}

	// 提取种子数和下载数
	seeders := g.extractSeeders(rowHTML)
	leechers := g.extractLeechers(rowHTML)

	// 提取时间
	uploadDate := g.extractUploadDate(rowHTML)

	// 提取上传者
	uploaderRegex := regexp.MustCompile(`<a[^>]*href=["']user\.php\?id=\d+["'][^>]*>([^<]+)</a>`)
	uploaderMatches := uploaderRegex.FindStringSubmatch(rowHTML)
	uploader := ""
	if len(uploaderMatches) >= 2 {
		uploader = g.cleanTitle(uploaderMatches[1])
	}

	// 构建下载URL（需要authkey）
	downloadURL := ""
	if g.authKey != "" {
		downloadURL = fmt.Sprintf("%s/torrents.php?action=download&id=%s&authkey=%s", 
			g.GetBaseURL(), id, g.authKey)
	}

	return &indexer.TorrentInfo{
		ID:           id,
		Title:        title,
		Description:  "",
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    0, // Gazelle通常不显示完成数
		UploadDate:   uploadDate,
		DownloadURL:  downloadURL,
		DetailURL:    fmt.Sprintf("%s/torrents.php?id=%s", g.GetBaseURL(), id),
		Category:     g.extractCategory(rowHTML),
		Tags:         g.extractTags(rowHTML),
		Uploader:     uploader,
		IMDBID:       g.extractIMDBID(rowHTML),
		FreeTorrent:  g.isFreeTorrent(rowHTML),
		DoubleUpload: g.isDoubleUpload(rowHTML),
	}
}

// GetTorrentDetail 获取种子详情
func (g *GazelleSpider) GetTorrentDetail(ctx context.Context, id string) (*indexer.TorrentDetail, error) {
	g.RateLimit(ctx)

	path := fmt.Sprintf("/torrents.php?id=%s", id)
	html, err := g.GetPageContent(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get torrent detail failed: %w", err)
	}

	return g.ParseTorrentDetail(html), nil
}

// ParseTorrentDetail 解析种子详情
func (g *GazelleSpider) ParseTorrentDetail(html string) *indexer.TorrentDetail {
	detail := &indexer.TorrentDetail{
		Files:     []*indexer.TorrentFile{},
		Comments:  []*indexer.Comment{},
	}

	// 提取文件列表
	files := g.extractFileList(html)
	detail.Files = files

	// 提取评论
	comments := g.extractComments(html)
	detail.Comments = comments

	return detail
}

// GetUserInfo 获取用户信息
func (g *GazelleSpider) GetUserInfo(ctx context.Context) (*indexer.UserInfo, error) {
	g.RateLimit(ctx)

	html, err := g.GetPageContent(ctx, "/user.php", nil)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	return g.ParseUserInfo(html), nil
}

// ParseUserInfo 解析用户信息
func (g *GazelleSpider) ParseUserInfo(html string) *indexer.UserInfo {
	userInfo := &indexer.UserInfo{}

	// 提取用户名
	usernameRegex := regexp.MustCompile(`(?:Welcome|Hello)\s+([^<\s]+)`)
	matches := usernameRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Username = matches[1]
	}

	// 提取用户等级
	rankRegex := regexp.MustCompile(`Class[^:]*:\s*([^<\s]+)`)
	matches = rankRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Rank = matches[1]
	}

	// 提取上传量
	uploadRegex := regexp.MustCompile(`Upload[^:]*:\s*([\d.]+\s*[KMGT]?B)`)
	matches = uploadRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Upload = g.parseSizeFromString(matches[1])
	}

	// 提取下载量
	downloadRegex := regexp.MustCompile(`Download[^:]*:\s*([\d.]+\s*[KMGT]?B)`)
	matches = downloadRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		userInfo.Download = g.parseSizeFromString(matches[1])
	}

	// 提取分享率
	ratioRegex := regexp.MustCompile(`Ratio[^:]*:\s*([\d.]+)`)
	matches = ratioRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if ratio, err := strconv.ParseFloat(matches[1], 64); err == nil {
			userInfo.Ratio = ratio
		}
	}

	return userInfo
}

// Download 下载种子
func (g *GazelleSpider) Download(ctx context.Context, id string) ([]byte, error) {
	g.RateLimit(ctx)

	if g.authKey == "" {
		// 重新获取authkey
		if userHTML, err := g.GetPageContent(ctx, "/index.php", nil); err == nil {
			g.authKey = g.extractAuthKey(userHTML)
		}
	}

	if g.authKey == "" {
		return nil, fmt.Errorf("no authkey available")
	}

	// 构建下载URL
	downloadURL := fmt.Sprintf("/torrents.php?action=download&id=%s&authkey=%s", id, g.authKey)
	resp, err := g.Get(ctx, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("download torrent failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 读取种子内容
	content, err := g.readResponseContent(resp)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// 辅助方法

// cleanTitle 清理标题
func (g *GazelleSpider) cleanTitle(title string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	title = re.ReplaceAllString(title, "")
	
	re = regexp.MustCompile(`\s+`)
	title = re.ReplaceAllString(title, " ")
	
	return strings.TrimSpace(title)
}

// parseSize 解析大小
func (g *GazelleSpider) parseSize(size string, unit string) int64 {
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
func (g *GazelleSpider) parseSizeFromString(sizeStr string) int64 {
	sizeRegex := regexp.MustCompile(`(\d+\.?\d*)\s*(bytes|KB|MB|GB|TB)`)
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) >= 3 {
		return g.parseSize(matches[1], matches[2])
	}
	return 0
}

// extractSeeders 提取种子数
func (g *GazelleSpider) extractSeeders(html string) int {
	regex := regexp.MustCompile(`(\d+)\s*seeders?`)
	matches := regex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return 0
}

// extractLeechers 提取下载数
func (g *GazelleSpider) extractLeechers(html string) int {
	regex := regexp.MustCompile(`(\d+)\s*leechers?`)
	matches := regex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return 0
}

// extractUploadDate 提取上传时间
func (g *GazelleSpider) extractUploadDate(html string) time.Time {
	// Gazelle时间格式通常为相对时间或绝对时间
	dateRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2})`)
	matches := dateRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		if t, err := time.Parse("2006-01-02 15:04", matches[1]); err == nil {
			return t
		}
	}
	return time.Now()
}

// extractCategory 提取分类
func (g *GazelleSpider) extractCategory(html string) string {
	if strings.Contains(strings.ToLower(html), "movie") {
		return "movie"
	}
	if strings.Contains(strings.ToLower(html), "music") {
		return "music"
	}
	if strings.Contains(strings.ToLower(html), "game") {
		return "game"
	}
	if strings.Contains(strings.ToLower(html), "application") {
		return "app"
	}
	if strings.Contains(strings.ToLower(html), "ebook") {
		return "ebook"
	}
	return "other"
}

// extractTags 提取标签
func (g *GazelleSpider) extractTags(html string) []string {
	var tags []string
	
	tagIdentifiers := []string{
		"Freeleech!", "Personal Freeleech", "Neutral Leech", "Staff Pick",
		"Lossless", "FLAC", "MP3", "WEB", "DVD", "BluRay",
	}
	
	for _, identifier := range tagIdentifiers {
		if strings.Contains(html, identifier) {
			tags = append(tags, identifier)
		}
	}
	
	return tags
}

// extractIMDBID 提取IMDB ID
func (g *GazelleSpider) extractIMDBID(html string) string {
	imdbRegex := regexp.MustCompile(`(?:imdb\.com/title/tt|tt)(\d{7,8})`)
	matches := imdbRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// isFreeTorrent 检查是否为免费种子
func (g *GazelleSpider) isFreeTorrent(html string) bool {
	return strings.Contains(strings.ToLower(html), "freeleech") ||
		   strings.Contains(strings.ToLower(html), "neutral leech")
}

// isDoubleUpload 检查是否为双倍上传
func (g *GazelleSpider) isDoubleUpload(html string) bool {
	return strings.Contains(strings.ToLower(html), "double upload")
}

// extractFileList 提取文件列表
func (g *GazelleSpider) extractFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile
	
	// 解析文件列表表格
	fileRegex := regexp.MustCompile(`<tr[^>]*file_row[^>]*>.*?</tr>`)
	matches := fileRegex.FindAllString(html, -1)
	
	for _, match := range matches {
		file := g.parseFileRow(match)
		if file != nil {
			files = append(files, file)
		}
	}
	
	return files
}

// parseFileRow 解析文件行
func (g *GazelleSpider) parseFileRow(rowHTML string) *indexer.TorrentFile {
	// 提取文件名
	nameRegex := regexp.MustCompile(`<td[^>]*>([^<]+)</td>`)
	nameMatches := nameRegex.FindStringSubmatch(rowHTML)
	if len(nameMatches) < 2 {
		return nil
	}
	
	// 提取文件大小
	sizeRegex := regexp.MustCompile(`(\d+\.?\d*)\s*(bytes|KB|MB|GB|TB)`)
	sizeMatches := sizeRegex.FindStringSubmatch(rowHTML)
	var size int64
	if len(sizeMatches) >= 3 {
		size = g.parseSize(sizeMatches[1], sizeMatches[2])
	}
	
	return &indexer.TorrentFile{
		Name: g.cleanTitle(nameMatches[1]),
		Size: size,
		Path: "",
	}
}

// extractComments 提取评论
func (g *GazelleSpider) extractComments(html string) []*indexer.Comment {
	var comments []*indexer.Comment
	
	// 解析评论
	commentRegex := regexp.MustCompile(`<div[^>]*comment[^>]*>.*?</div>`)
	matches := commentRegex.FindAllString(html, -1)
	
	for i, match := range matches {
		comment := g.parseComment(match, i)
		if comment != nil {
			comments = append(comments, comment)
		}
	}
	
	return comments
}

// parseComment 解析评论
func (g *GazelleSpider) parseComment(commentHTML string, index int) *indexer.Comment {
	// 提取评论内容
	contentRegex := regexp.MustCompile(`<div[^>]*comment_body[^>]*>(.*?)</div>`)
	contentMatches := contentRegex.FindStringSubmatch(commentHTML)
	if len(contentMatches) < 2 {
		return nil
	}
	
	// 提取作者
	authorRegex := regexp.MustCompile(`<a[^>]*user\.php\?[^>]*>([^<]+)</a>`)
	authorMatches := authorRegex.FindStringSubmatch(commentHTML)
	author := "Anonymous"
	if len(authorMatches) >= 2 {
		author = g.cleanTitle(authorMatches[1])
	}
	
	return &indexer.Comment{
		ID:       fmt.Sprintf("comment_%d", index),
		Author:   author,
		Content:  g.cleanTitle(contentMatches[1]),
		PostDate: time.Now(), // 实际应用中应该从HTML中提取
	}
}

// readResponseContent 读取响应内容
func (g *GazelleSpider) readResponseContent(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	
	content := make([]byte, 0)
	buf := make([]byte, 4096)
	
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			content = append(content, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	
	return content, nil
}

// IsLoggedIn 检查登录状态
func (g *GazelleSpider) IsLoggedIn(ctx context.Context) bool {
	html, err := g.GetPageContent(ctx, "/index.php", nil)
	if err != nil {
		return false
	}
	
	// 检查是否包含登录页面的特征
	if strings.Contains(strings.ToLower(html), "login") || strings.Contains(strings.ToLower(html), "log in") {
		return false
	}
	
	// 检查是否包含用户信息的特征
	return strings.Contains(strings.ToLower(html), "ratio") || 
		   strings.Contains(strings.ToLower(html), "upload") ||
		   strings.Contains(strings.ToLower(html), "logout")
}