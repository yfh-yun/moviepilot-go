package utils

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// RSSItem RSS项目
type RSSItem struct {
	Title       string
	Link        string
	Description string
	PubDate     time.Time
	GUID        string
	Enclosure   *Enclosure
	TorrentURL  string
	MagnetURL   string
}

// Enclosure 附件
type Enclosure struct {
	URL    string
	Length int64
	Type   string
}

// RSSFeed RSS订阅
type RSSFeed struct {
	Title       string
	Link        string
	Description string
	Items       []*RSSItem
}

// Parser RSS解析器
type Parser interface {
	// ParseURL 解析指定URL的RSS订阅
	ParseURL(ctx context.Context, url string) (*RSSFeed, error)

	// ParseContent 解析RSS内容
	ParseContent(ctx context.Context, content string) (*RSSFeed, error)
}

// rssParser RSS解析器实现
type rssParser struct {
	logger     *zap.Logger
	httpClient *http.Client
	parser     *gofeed.Parser
}

// NewParser 创建RSS解析器
func NewParser() Parser {
	return &rssParser{
		logger: logger.GetLogger(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		parser: gofeed.NewParser(),
	}
}

// ParseURL 解析指定URL的RSS订阅
func (p *rssParser) ParseURL(ctx context.Context, url string) (*RSSFeed, error) {
	p.logger.Info("解析RSS订阅", zap.String("url", url))

	// 发送HTTP请求获取RSS内容
	resp, err := p.httpClient.Get(url)
	if err != nil {
		p.logger.Error("获取RSS内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取RSS内容失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		p.logger.Error("获取RSS内容失败", zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("获取RSS内容失败，状态码: %d", resp.StatusCode)
	}

	// 解析RSS内容
	feed, err := p.parser.Parse(resp.Body)
	if err != nil {
		p.logger.Error("解析RSS内容失败", zap.Error(err))
		return nil, fmt.Errorf("解析RSS内容失败: %w", err)
	}

	// 转换为自定义RSSFeed结构
	rssFeed := p.convertToRSSFeed(feed)

	p.logger.Info("RSS订阅解析成功",
		zap.String("title", rssFeed.Title),
		zap.Int("item_count", len(rssFeed.Items)),
	)

	return rssFeed, nil
}

// ParseContent 解析RSS内容
func (p *rssParser) ParseContent(ctx context.Context, content string) (*RSSFeed, error) {
	p.logger.Info("解析RSS内容")

	// 解析RSS内容
	feed, err := p.parser.ParseString(content)
	if err != nil {
		p.logger.Error("解析RSS内容失败", zap.Error(err))
		return nil, fmt.Errorf("解析RSS内容失败: %w", err)
	}

	// 转换为自定义RSSFeed结构
	rssFeed := p.convertToRSSFeed(feed)

	p.logger.Info("RSS内容解析成功",
		zap.String("title", rssFeed.Title),
		zap.Int("item_count", len(rssFeed.Items)),
	)

	return rssFeed, nil
}

// convertToRSSFeed 将gofeed.Feed转换为自定义RSSFeed结构
func (p *rssParser) convertToRSSFeed(feed *gofeed.Feed) *RSSFeed {
	rssFeed := &RSSFeed{
		Title:       feed.Title,
		Link:        feed.Link,
		Description: feed.Description,
		Items:       make([]*RSSItem, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		rssItem := &RSSItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			PubDate:     item.PublishedParsed.UTC(),
			GUID:        item.GUID,
		}

		// 处理附件
		if len(item.Enclosures) > 0 {
			enclosure := item.Enclosures[0]
			// 将string类型的Length转换为int64
			length, _ := strconv.ParseInt(enclosure.Length, 10, 64)
			rssItem.Enclosure = &Enclosure{
				URL:    enclosure.URL,
				Length: length,
				Type:   enclosure.Type,
			}

			// 提取种子URL
			if enclosure.Type == "application/x-bittorrent" ||
				enclosure.Type == "application/octet-stream" ||
				len(enclosure.URL) > 8 && enclosure.URL[len(enclosure.URL)-8:] == ".torrent" {
				rssItem.TorrentURL = enclosure.URL
			}
		}

		// 提取磁力链接
		if item.Link != "" && len(item.Link) > 7 && item.Link[:7] == "magnet:" {
			rssItem.MagnetURL = item.Link
		}

		rssFeed.Items = append(rssFeed.Items, rssItem)
	}

	return rssFeed
}
