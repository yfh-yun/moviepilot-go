package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RSSHelper RSS辅助工具
type RSSHelper struct {
	client *http.Client
}

// NewRSSHelper 创建RSS辅助工具实例
func NewRSSHelper() *RSSHelper {
	return &RSSHelper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RSSItem RSS条目
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
	Comments    string `xml:"comments"`
	Enclosure   struct {
		URL    string `xml:"url,attr"`
		Type   string `xml:"type,attr"`
		Length string `xml:"length,attr"`
	} `xml:"enclosure"`
}

// RSSChannel RSS频道
type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	Copyright     string    `xml:"copyright"`
	ManagingEditor string    `xml:"managingEditor"`
	WebMaster     string    `xml:"webMaster"`
	PubDate       string    `xml:"pubDate"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Category      string    `xml:"category"`
	Generator     string    `xml:"generator"`
	Docs          string    `xml:"docs"`
	Cloud         struct {
		Domain             string `xml:"domain,attr"`
		Port               string `xml:"port,attr"`
		Path               string `xml:"path,attr"`
		RegisterProcedure  string `xml:"registerProcedure,attr"`
		Protocol           string `xml:"protocol,attr"`
	} `xml:"cloud"`
	TTL        string    `xml:"ttl"`
	Image      struct {
		URL         string `xml:"url"`
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Width       string `xml:"width"`
		Height      string `xml:"height"`
		Description string `xml:"description"`
	} `xml:"image"`
	Rating     string    `xml:"rating"`
	TextInput  struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Name        string `xml:"name"`
		Link        string `xml:"link"`
	} `xml:"textInput"`
	SkipHours struct {
		Hour []string `xml:"hour"`
	} `xml:"skipHours"`
	SkipDays struct {
		Day []string `xml:"day"`
	} `xml:"skipDays"`
	Items []RSSItem `xml:"item"`
}

// RSS RSS根元素
type RSS struct {
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

// AtomEntry Atom条目
type AtomEntry struct {
	Title   string   `xml:"title"`
	Link    []struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	ID        string   `xml:"id"`
	Published string   `xml:"published"`
	Updated   string   `xml:"updated"`
	Author    struct {
		Name  string `xml:"name"`
		Email string `xml:"email"`
		URI   string `xml:"uri"`
	} `xml:"author"`
	Summary string   `xml:"summary"`
	Content string   `xml:"content"`
	Category []struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
}

// AtomFeed Atom订阅源
type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Xmlns   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    []struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	Subtitle string   `xml:"subtitle"`
	ID       string   `xml:"id"`
	Updated  string   `xml:"updated"`
	Author   struct {
		Name  string `xml:"name"`
		Email string `xml:"email"`
		URI   string `xml:"uri"`
	} `xml:"author"`
	Generator struct {
		Version string `xml:"version,attr"`
		URI     string `xml:"uri,attr"`
		Text    string `xml:",chardata"`
	} `xml:"generator"`
	Icon  string      `xml:"icon"`
	Logo  string      `xml:"logo"`
	Rights string     `xml:"rights"`
	Entry []AtomEntry `xml:"entry"`
}

// FeedInfo 订阅源信息
type FeedInfo struct {
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Link        string      `json:"link"`
	Description string      `json:"description"`
	Language    string      `json:"language"`
	Updated     time.Time   `json:"updated"`
	Items       []FeedItem  `json:"items"`
}

// FeedItem 订阅条目
type FeedItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Published   time.Time `json:"published"`
	Updated     time.Time `json:"updated"`
	Author      string    `json:"author"`
	GUID        string    `json:"guid"`
	Categories  []string  `json:"categories"`
	Enclosure   *Enclosure `json:"enclosure,omitempty"`
}

// Enclosure 附件信息
type Enclosure struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Length int64  `json:"length"`
}

// FetchRSS 获取RSS订阅源
func (r *RSSHelper) FetchRSS(url string) (*FeedInfo, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取RSS失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取RSS内容失败: %w", err)
	}

	// 尝试解析RSS
	if strings.Contains(string(body), "<rss") {
		return r.parseRSS(body)
	}

	// 尝试解析Atom
	if strings.Contains(string(body), "<feed") {
		return r.parseAtom(body)
	}

	return nil, fmt.Errorf("不支持的订阅源格式")
}

// parseRSS 解析RSS格式
func (r *RSSHelper) parseRSS(data []byte) (*FeedInfo, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("解析RSS失败: %w", err)
	}

	feed := &FeedInfo{
		Type:        "rss",
		Title:       rss.Channel.Title,
		Link:        rss.Channel.Link,
		Description: rss.Channel.Description,
		Language:    rss.Channel.Language,
	}

	// 解析更新时间
	if rss.Channel.LastBuildDate != "" {
		if updated, err := r.parseDate(rss.Channel.LastBuildDate); err == nil {
			feed.Updated = updated
		}
	} else if rss.Channel.PubDate != "" {
		if updated, err := r.parseDate(rss.Channel.PubDate); err == nil {
			feed.Updated = updated
		}
	}

	// 解析条目
	for _, item := range rss.Channel.Items {
		feedItem := FeedItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: r.cleanHTML(item.Description),
			Content:     r.cleanHTML(item.Content),
			Author:      item.Author,
			GUID:        item.GUID,
		}

		// 解析发布时间
		if item.PubDate != "" {
			if published, err := r.parseDate(item.PubDate); err == nil {
				feedItem.Published = published
			}
		}

		// 解析分类
		if item.Category != "" {
			feedItem.Categories = append(feedItem.Categories, item.Category)
		}

		// 解析附件
		if item.Enclosure.URL != "" {
			var length int64
			if item.Enclosure.Length != "" {
				if parsed, err := parseLength(item.Enclosure.Length); err == nil {
					length = parsed
				}
			}
			feedItem.Enclosure = &Enclosure{
				URL:    item.Enclosure.URL,
				Type:   item.Enclosure.Type,
				Length: length,
			}
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed, nil
}

// parseAtom 解析Atom格式
func (r *RSSHelper) parseAtom(data []byte) (*FeedInfo, error) {
	var atom AtomFeed
	if err := xml.Unmarshal(data, &atom); err != nil {
		return nil, fmt.Errorf("解析Atom失败: %w", err)
	}

	feed := &FeedInfo{
		Type:        "atom",
		Title:       atom.Title,
		Description: atom.Subtitle,
		Language:    "en", // Atom通常不包含语言信息
	}

	// 获取主要链接
	for _, link := range atom.Link {
		if link.Rel == "alternate" || link.Rel == "" {
			feed.Link = link.Href
			break
		}
	}

	// 解析更新时间
	if atom.Updated != "" {
		if updated, err := r.parseDate(atom.Updated); err == nil {
			feed.Updated = updated
		}
	}

	// 解析条目
	for _, entry := range atom.Entry {
		feedItem := FeedItem{
			Title:       entry.Title,
			Description: r.cleanHTML(entry.Summary),
			Content:     r.cleanHTML(entry.Content),
			Author:      entry.Author.Name,
			GUID:        entry.ID,
		}

		// 获取链接
		for _, link := range entry.Link {
			if link.Rel == "alternate" || link.Rel == "" {
				feedItem.Link = link.Href
				break
			}
		}

		// 解析发布时间
		if entry.Published != "" {
			if published, err := r.parseDate(entry.Published); err == nil {
				feedItem.Published = published
			}
		}

		// 解析更新时间
		if entry.Updated != "" {
			if updated, err := r.parseDate(entry.Updated); err == nil {
				feedItem.Updated = updated
			}
		}

		// 解析分类
		for _, category := range entry.Category {
			feedItem.Categories = append(feedItem.Categories, category.Term)
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed, nil
}

// parseDate 解析日期
func (r *RSSHelper) parseDate(dateStr string) (time.Time, error) {
	// 常见的RSS/Atom日期格式
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// cleanHTML 清理HTML标签
func (r *RSSHelper) cleanHTML(html string) string {
	// 简单的HTML标签清理
	replacer := strings.NewReplacer(
		"<p>", "",
		"</p>", "\n",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"<div>", "\n",
		"</div>", "\n",
	)

	// 移除所有HTML标签
	result := replacer.Replace(html)
	
	// 移除剩余的HTML标签
	inTag := false
	var cleaned strings.Builder
	
	for _, r := range result {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			cleaned.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(cleaned.String())
}

// ValidateRSS 验证RSS订阅源
func (r *RSSHelper) ValidateRSS(url string) error {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return err
	}

	if feed.Title == "" {
		return fmt.Errorf("订阅源标题为空")
	}

	if len(feed.Items) == 0 {
		return fmt.Errorf("订阅源没有条目")
	}

	return nil
}

// GetFeedInfo 获取订阅源基本信息
func (r *RSSHelper) GetFeedInfo(url string) (*FeedInfo, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	// 只返回基本信息，不包含条目
	info := &FeedInfo{
		Type:        feed.Type,
		Title:       feed.Title,
		Link:        feed.Link,
		Description: feed.Description,
		Language:    feed.Language,
		Updated:     feed.Updated,
	}

	return info, nil
}

// GetRecentItems 获取最近的条目
func (r *RSSHelper) GetRecentItems(url string, limit int) ([]FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(feed.Items) {
		limit = len(feed.Items)
	}

	return feed.Items[:limit], nil
}

// SearchItems 搜索条目
func (r *RSSHelper) SearchItems(url string, keyword string) ([]FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	var results []FeedItem
	keyword = strings.ToLower(keyword)

	for _, item := range feed.Items {
		if strings.Contains(strings.ToLower(item.Title), keyword) ||
			strings.Contains(strings.ToLower(item.Description), keyword) ||
			strings.Contains(strings.ToLower(item.Content), keyword) {
			results = append(results, item)
		}
	}

	return results, nil
}

// FilterByDate 按日期过滤条目
func (r *RSSHelper) FilterByDate(url string, startDate, endDate time.Time) ([]FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	var results []FeedItem

	for _, item := range feed.Items {
		itemDate := item.Published
		if itemDate.IsZero() {
			itemDate = item.Updated
		}

		if !itemDate.IsZero() && itemDate.After(startDate) && itemDate.Before(endDate) {
			results = append(results, item)
		}
	}

	return results, nil
}

// FilterByCategory 按分类过滤条目
func (r *RSSHelper) FilterByCategory(url string, category string) ([]FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	var results []FeedItem
	category = strings.ToLower(category)

	for _, item := range feed.Items {
		for _, itemCategory := range item.Categories {
			if strings.Contains(strings.ToLower(itemCategory), category) {
				results = append(results, item)
				break
			}
		}
	}

	return results, nil
}

// GetItemByGUID 根据GUID获取条目
func (r *RSSHelper) GetItemByGUID(url string, guid string) (*FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	for _, item := range feed.Items {
		if item.GUID == guid {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("未找到GUID为 %s 的条目", guid)
}

// GetItemByLink 根据链接获取条目
func (r *RSSHelper) GetItemByLink(url string, link string) (*FeedItem, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	for _, item := range feed.Items {
		if item.Link == link {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("未找到链接为 %s 的条目", link)
}

// GetFeedStats 获取订阅源统计信息
func (r *RSSHelper) GetFeedStats(url string) (map[string]interface{}, error) {
	feed, err := r.FetchRSS(url)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"type":           feed.Type,
		"title":          feed.Title,
		"total_items":    len(feed.Items),
		"has_enclosure":  false,
		"categories":     make(map[string]int),
		"oldest_item":    nil,
		"newest_item":    nil,
	}

	var oldestTime, newestTime time.Time
	categoryCount := make(map[string]int)

	for _, item := range feed.Items {
		// 统计附件
		if item.Enclosure != nil {
			stats["has_enclosure"] = true
		}

		// 统计分类
		for _, category := range item.Categories {
			categoryCount[category]++
		}

		// 统计时间范围
		itemTime := item.Published
		if itemTime.IsZero() {
			itemTime = item.Updated
		}

		if !itemTime.IsZero() {
			if oldestTime.IsZero() || itemTime.Before(oldestTime) {
				oldestTime = itemTime
			}
			if newestTime.IsZero() || itemTime.After(newestTime) {
				newestTime = itemTime
			}
		}
	}

	stats["categories"] = categoryCount
	if !oldestTime.IsZero() {
		stats["oldest_item"] = oldestTime
	}
	if !newestTime.IsZero() {
		stats["newest_item"] = newestTime
	}

	return stats, nil
}

// ExportToJSON 导出为JSON格式
func (r *RSSHelper) ExportToJSON(feed *FeedInfo) ([]byte, error) {
	return json.Marshal(feed)
}

// ExportToXML 导出为XML格式
func (r *RSSHelper) ExportToXML(feed *FeedInfo) ([]byte, error) {
	if feed.Type == "rss" {
		rss := RSS{
			Version: "2.0",
			Channel: RSSChannel{
				Title:       feed.Title,
				Link:        feed.Link,
				Description: feed.Description,
				Language:    feed.Language,
				LastBuildDate: feed.Updated.Format(time.RFC1123Z),
			},
		}

		for _, item := range feed.Items {
			rssItem := RSSItem{
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				GUID:        item.GUID,
				Author:      item.Author,
			}

			if !item.Published.IsZero() {
				rssItem.PubDate = item.Published.Format(time.RFC1123Z)
			}

			if item.Enclosure != nil {
				rssItem.Enclosure.URL = item.Enclosure.URL
				rssItem.Enclosure.Type = item.Enclosure.Type
				rssItem.Enclosure.Length = fmt.Sprintf("%d", item.Enclosure.Length)
			}

			rss.Channel.Items = append(rss.Channel.Items, rssItem)
		}

		return xml.MarshalIndent(rss, "", "  ")
	}

	return nil, fmt.Errorf("不支持的导出格式: %s", feed.Type)
}

// parseLength 解析长度字符串
func parseLength(length string) (int64, error) {
	var result int64
	for _, r := range length {
		if r >= '0' && r <= '9' {
			result = result*10 + int64(r-'0')
		}
	}
	return result, nil
}