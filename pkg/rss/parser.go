package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
)

// Parser RSS 解析器
type Parser struct {
	client *http.Client
	cache  cache.Cache
	logger *zap.Logger
}

// Config 解析器配置
type Config struct {
	Timeout time.Duration
	Cache   cache.Cache
	Logger  *zap.Logger
}

// NewParser 创建 RSS 解析器
func NewParser(config Config) *Parser {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Parser{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		cache:  config.Cache,
		logger: config.Logger,
	}
}

// RSSFeed RSS 订阅源
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

// Channel RSS 频道
type Channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	PubDate     string    `xml:"pubDate"`
	Items       []RSSItem `xml:"item"`
}

// RSSItem RSS 项目
type RSSItem struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
	Enclosure   Enclosure `xml:"enclosure"`
	Category    string    `xml:"category"`
}

// Enclosure 附件信息
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// ParseURL 解析 RSS URL
func (p *Parser) ParseURL(ctx context.Context, url string) (*RSSFeed, error) {
	// 检查缓存
	if p.cache != nil {
		cacheKey := fmt.Sprintf("rss:%s", url)
		var feed RSSFeed
		err := p.cache.GetJSON(ctx, cacheKey, &feed)
		if err == nil {
			if p.logger != nil {
				p.logger.Debug("RSS cache hit", zap.String("url", url))
			}
			return &feed, nil
		}
	}

	// 创建带 context 的请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 发起请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS fetch failed with status: %d", resp.StatusCode)
	}

	// 读取响应
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS response: %w", err)
	}

	// 解析 XML
	feed, err := p.ParseXML(data)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if p.cache != nil {
		cacheKey := fmt.Sprintf("rss:%s", url)
		if err := p.cache.SetJSON(ctx, cacheKey, feed, 10*time.Minute); err != nil && p.logger != nil {
			p.logger.Warn("failed to cache RSS", zap.Error(err))
		}
	}

	if p.logger != nil {
		p.logger.Info("RSS parsed successfully",
			zap.String("url", url),
			zap.Int("items", len(feed.Channel.Items)))
	}

	return feed, nil
}

// ParseXML 解析 RSS XML 数据
func (p *Parser) ParseXML(data []byte) (*RSSFeed, error) {
	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	return &feed, nil
}

// ParsePubDate 解析发布日期
func ParsePubDate(pubDate string) (time.Time, error) {
	// 尝试多种日期格式
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, pubDate); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", pubDate)
}
