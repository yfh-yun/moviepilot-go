package subtitle

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"golang.org/x/net/html"
)

// SubHDProvider SubHD提供者
type SubHDProvider struct {
	config   map[string]interface{}
	client   *http.Client
	baseURL  string
}

// SubHDConfig SubHD配置
type SubHDConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserAgent string `json:"user_agent"`
}

// SubHDSubtitle SubHD字幕
type SubHDSubtitle struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Language string    `json:"language"`
	Rate     float64   `json:"rate"`
	Downloads int      `json:"downloads"`
	UploadDate time.Time `json:"upload_date"`
	FileURL  string    `json:"file_url"`
	FileSize int64     `json:"file_size"`
	Format   string    `json:"format"`
}

// NewSubHDProvider 创建SubHD提供者
func NewSubHDProvider(config map[string]interface{}) *SubHDProvider {
	return &SubHDProvider{
		config:   config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://subhd.tv",
	}
}

// Name 返回提供者名称
func (p *SubHDProvider) Name() string {
	return "subhd"
}

// Search 搜索字幕
func (p *SubHDProvider) Search(ctx context.Context, req *SearchRequest) ([]*Subtitle, error) {
	// 构建搜索URL
	searchURL := p.buildSearchURL(req)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search subtitles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SubHD returned status %d", resp.StatusCode)
	}

	// 解析HTML响应
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML response: %w", err)
	}

	// 提取字幕信息
	subtitles := p.extractSubtitles(doc, req)
	
	return subtitles, nil
}

// buildSearchURL 构建搜索URL
func (p *SubHDProvider) buildSearchURL(req *SearchRequest) string {
	baseURL := p.baseURL + "/search"
	
	params := url.Values{}
	
	if req.Title != "" {
		params.Set("keyword", req.Title)
	}
	
	if req.Language != "" {
		params.Set("language", p.getLanguageCode(req.Language))
	}
	
	if req.Season > 0 && req.Episode > 0 {
		params.Set("season", fmt.Sprintf("%d", req.Season))
		params.Set("episode", fmt.Sprintf("%d", req.Episode))
	}
	
	if len(params) > 0 {
		baseURL += "?" + params.Encode()
	}
	
	return baseURL
}

// extractSubtitles 提取字幕信息
func (p *SubHDProvider) extractSubtitles(doc *html.Node, req *SearchRequest) []*Subtitle {
	var subtitles []*Subtitle
	
	// 查找字幕列表容器
	var findSubtitleList func(*html.Node)
	findSubtitleList = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "sub-list") {
					p.parseSubtitleList(n, &subtitles, req)
					return
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSubtitleList(c)
		}
	}
	
	findSubtitleList(doc)
	return subtitles
}

// parseSubtitleList 解析字幕列表
func (p *SubHDProvider) parseSubtitleList(node *html.Node, subtitles *[]*Subtitle, req *SearchRequest) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "div" {
			for _, attr := range child.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "sub-item") {
					subtitle := p.parseSubtitleItem(child, req)
					if subtitle != nil {
						*subtitles = append(*subtitles, subtitle)
					}
				}
			}
		}
	}
}

// parseSubtitleItem 解析字幕项
func (p *SubHDProvider) parseSubtitleItem(node *html.Node, req *SearchRequest) *Subtitle {
	subtitle := &Subtitle{
		Provider: p.Name(),
		Extra:    make(map[string]string),
	}
	
	// 解析字幕信息
	var parseItem func(*html.Node)
	parseItem = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				// 提取链接和标题
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						subtitle.FileURL = p.baseURL + attr.Val
						subtitle.ID = p.extractIDFromURL(attr.Val)
					}
					if attr.Key == "title" {
						subtitle.Title = attr.Val
					}
				}
				
				// 提取文本内容
				if text := p.extractText(n); text != "" {
					if strings.Contains(text, "S") && strings.Contains(text, "E") {
						// 解析季集信息
						if season, episode := p.parseSeasonEpisode(text); season > 0 && episode > 0 {
							subtitle.Season = season
							subtitle.Episode = episode
						}
					}
				}
				
			case "span":
				// 提取语言、评分、下载量等信息
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						switch {
						case strings.Contains(attr.Val, "language"):
							subtitle.Language = p.extractText(n)
						case strings.Contains(attr.Val, "rating"):
							if rate := p.extractRating(p.extractText(n)); rate > 0 {
								subtitle.Rate = rate
							}
						case strings.Contains(attr.Val, "downloads"):
							if downloads := p.extractDownloads(p.extractText(n)); downloads > 0 {
								subtitle.Downloads = downloads
							}
						case strings.Contains(attr.Val, "date"):
							if date := p.extractDate(p.extractText(n)); !date.IsZero() {
								subtitle.UploadDate = date
							}
						}
					}
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseItem(c)
		}
	}
	
	parseItem(node)
	
	// 验证必要字段
	if subtitle.ID == "" || subtitle.Title == "" {
		return nil
	}
	
	// 设置默认值
	if subtitle.Language == "" {
		subtitle.Language = "zh"
	}
	
	return subtitle
}

// extractText 提取文本内容
func (p *SubHDProvider) extractText(node *html.Node) string {
	var text strings.Builder
	
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	
	extract(node)
	return strings.TrimSpace(text.String())
}

// extractIDFromURL 从URL提取ID
func (p *SubHDProvider) extractIDFromURL(url string) string {
	re := regexp.MustCompile(`/d/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseSeasonEpisode 解析季集信息
func (p *SubHDProvider) parseSeasonEpisode(text string) (int, int) {
	re := regexp.MustCompile(`S(\d+)E(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 2 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		return season, episode
	}
	return 0, 0
}

// extractRating 提取评分
func (p *SubHDProvider) extractRating(text string) float64 {
	re := regexp.MustCompile(`(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if rating, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return rating
		}
	}
	return 0
}

// extractDownloads 提取下载量
func (p *SubHDProvider) extractDownloads(text string) int {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if downloads, err := strconv.Atoi(matches[1]); err == nil {
			return downloads
		}
	}
	return 0
}

// extractDate 提取日期
func (p *SubHDProvider) extractDate(text string) time.Time {
	// 尝试多种日期格式
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01-02",
		"01/02",
	}
	
	for _, format := range formats {
		if date, err := time.Parse(format, text); err == nil {
			// 如果只有月日，使用当前年份
			if format == "01-02" || format == "01/02" {
				now := time.Now()
				date = date.AddDate(now.Year(), 0, 0)
			}
			return date
		}
	}
	
	return time.Time{}
}

// Download 下载字幕
func (p *SubHDProvider) Download(ctx context.Context, subtitle *Subtitle) ([]byte, error) {
	if subtitle.FileURL == "" {
		return nil, fmt.Errorf("no download URL available")
	}

	// 访问字幕详情页获取下载链接
	httpReq, err := http.NewRequestWithContext(ctx, "GET", subtitle.FileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create detail request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtitle detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subtitle detail page returned status %d", resp.StatusCode)
	}

	// 解析详情页获取下载链接
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse detail page: %w", err)
	}

	downloadURL := p.extractDownloadURL(doc)
	if downloadURL == "" {
		return nil, fmt.Errorf("download URL not found")
	}

	// 下载字幕文件
	return p.downloadFile(ctx, downloadURL)
}

// extractDownloadURL 提取下载URL
func (p *SubHDProvider) extractDownloadURL(doc *html.Node) string {
	var downloadURL string
	
	var findDownloadLink func(*html.Node)
	findDownloadLink = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "download") {
					downloadURL = p.baseURL + attr.Val
					return
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDownloadLink(c)
		}
	}
	
	findDownloadLink(doc)
	return downloadURL
}

// downloadFile 下载文件
func (p *SubHDProvider) downloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file download returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return buf.Bytes(), nil
}

// Test 测试连接
func (p *SubHDProvider) Test(ctx context.Context) error {
	// 简单的搜索测试
	req := &SearchRequest{
		Title:    "test",
		Language: "zh",
	}

	_, err := p.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("SubHD test failed: %w", err)
	}

	logger.Info("SubHD provider test successful")
	return nil
}

// setHeaders 设置请求头
func (p *SubHDProvider) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", p.getUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

// getUserAgent 获取User-Agent
func (p *SubHDProvider) getUserAgent() string {
	if userAgent, ok := p.config["user_agent"].(string); ok && userAgent != "" {
		return userAgent
	}
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
}

// getLanguageCode 获取语言代码
func (p *SubHDProvider) getLanguageCode(language string) string {
	languageMap := map[string]string{
		"zh": "zh-cn",
		"zh-cn": "zh-cn",
		"zh-tw": "zh-tw",
		"en": "en",
		"ja": "ja",
		"ko": "ko",
		"fr": "fr",
		"de": "de",
		"es": "es",
		"it": "it",
		"pt": "pt",
		"ru": "ru",
	}
	
	if code, ok := languageMap[language]; ok {
		return code
	}
	return language
}