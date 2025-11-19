package actions

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RSSItem RSS项结构体
type RSSItem struct {
	Title       string    `xml:"title" json:"title"`
	Link        string    `xml:"link" json:"link"`
	Description string    `xml:"description" json:"description"`
	PubDate     time.Time `xml:"pubDate" json:"pubDate"`
	Guid        string    `xml:"guid" json:"guid"`
	Category    string    `xml:"category" json:"category"`
}

// RSSFeed RSS源结构体
type RSSFeed struct {
	Title string    `xml:"channel>title" json:"title"`
	Link  string    `xml:"channel>link" json:"link"`
	Items []RSSItem `xml:"channel>item" json:"items"`
}

// ParseRule 解析规则
type ParseRule struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Type        string `json:"type"` // "regex", "string"
}

// RSSParser RSS解析器接口
type RSSParser interface {
	Parse(content string) (*RSSFeed, error)
	Supports(format string) bool
}

// StandardRSSParser 标准RSS解析器
type StandardRSSParser struct {
	logger *zap.Logger
}

// NewStandardRSSParser 创建标准RSS解析器
func NewStandardRSSParser(logger *zap.Logger) *StandardRSSParser {
	return &StandardRSSParser{
		logger: logger,
	}
}

// Parse 解析RSS内容
func (p *StandardRSSParser) Parse(content string) (*RSSFeed, error) {
	var feed RSSFeed

	// 尝试解析标准RSS格式
	if err := xml.Unmarshal([]byte(content), &feed); err != nil {
		p.logger.Error("RSS解析失败", zap.Error(err))
		return nil, fmt.Errorf("RSS解析失败: %w", err)
	}

	// 处理日期格式
	for i := range feed.Items {
		if parsedTime, err := p.parseDate(feed.Items[i].PubDate); err == nil {
			feed.Items[i].PubDate = parsedTime
		}
	}

	return &feed, nil
}

// Supports 检查是否支持该格式
func (p *StandardRSSParser) Supports(format string) bool {
	return strings.Contains(strings.ToLower(format), "rss") ||
		strings.Contains(strings.ToLower(format), "xml")
}

// parseDate 解析日期格式
func (p *StandardRSSParser) parseDate(dateStr string) (time.Time, error) {
	// 尝试多种日期格式
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// CustomRSSParser 自定义RSS解析器
type CustomRSSParser struct {
	siteRules map[string]ParseRule
	logger    *zap.Logger
}

// NewCustomRSSParser 创建自定义RSS解析器
func NewCustomRSSParser(siteRules map[string]ParseRule, logger *zap.Logger) *CustomRSSParser {
	return &CustomRSSParser{
		siteRules: siteRules,
		logger:    logger,
	}
}

// Parse 使用自定义规则解析RSS
func (p *CustomRSSParser) Parse(content string) (*RSSFeed, error) {
	// 首先尝试标准解析
	standardParser := NewStandardRSSParser(p.logger)
	feed, err := standardParser.Parse(content)
	if err != nil {
		return nil, err
	}

	// 应用自定义规则
	for i := range feed.Items {
		p.applyCustomRules(&feed.Items[i])
	}

	return feed, nil
}

// Supports 检查是否支持该站点
func (p *CustomRSSParser) Supports(site string) bool {
	_, exists := p.siteRules[site]
	return exists
}

// applyCustomRules 应用自定义规则
func (p *CustomRSSParser) applyCustomRules(item *RSSItem) {
	// 根据站点URL匹配规则
	for site, rule := range p.siteRules {
		if strings.Contains(item.Link, site) {
			switch rule.Type {
			case "regex":
				if matched, err := regexp.MatchString(rule.Pattern, item.Title); err == nil && matched {
					item.Title = p.applyReplacement(item.Title, rule)
				}
			case "string":
				if strings.Contains(item.Title, rule.Pattern) {
					item.Title = p.applyReplacement(item.Title, rule)
				}
			}
			break
		}
	}
}

// applyReplacement 应用替换规则
func (p *CustomRSSParser) applyReplacement(text string, rule ParseRule) string {
	if rule.Type == "regex" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			p.logger.Error("正则表达式编译失败", zap.Error(err))
			return text
		}
		return regex.ReplaceAllString(text, rule.Replacement)
	}

	// 字符串替换
	return strings.ReplaceAll(text, rule.Pattern, rule.Replacement)
}

// RSSParserManager RSS解析器管理器
type RSSParserManager struct {
	parsers []RSSParser
	logger  *zap.Logger
}

// NewRSSParserManager 创建RSS解析器管理器
func NewRSSParserManager(logger *zap.Logger) *RSSParserManager {
	return &RSSParserManager{
		parsers: []RSSParser{
			NewStandardRSSParser(logger),
		},
		logger: logger,
	}
}

// RegisterParser 注册解析器
func (m *RSSParserManager) RegisterParser(parser RSSParser) {
	m.parsers = append(m.parsers, parser)
}

// ParseRSS 解析RSS内容
func (m *RSSParserManager) ParseRSS(ctx context.Context, content, format string) (*RSSFeed, error) {
	// 查找支持的解析器
	for _, parser := range m.parsers {
		if parser.Supports(format) {
			feed, err := parser.Parse(content)
			if err != nil {
				m.logger.Warn("解析器处理失败", zap.String("format", format), zap.Error(err))
				continue
			}
			return feed, nil
		}
	}

	return nil, fmt.Errorf("没有找到支持格式 '%s' 的解析器", format)
}

// FilterRSSItems 过滤RSS项
func (m *RSSParserManager) FilterRSSItems(ctx context.Context, items []RSSItem, filters []Filter) []RSSItem {
	var filtered []RSSItem

	for _, item := range items {
		if m.matchesFilters(ctx, item, filters) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// matchesFilters 检查是否匹配过滤器
func (m *RSSParserManager) matchesFilters(ctx context.Context, item RSSItem, filters []Filter) bool {
	for _, filter := range filters {
		if !filter.Match(item) {
			return false
		}
	}

	return true
}

// Filter 过滤器接口
type Filter interface {
	Match(item RSSItem) bool
}

// TitleFilter 标题过滤器
type TitleFilter struct {
	Pattern string
	Type    string // "contains", "regex"
}

// Match 检查标题是否匹配
func (f *TitleFilter) Match(item RSSItem) bool {
	switch f.Type {
	case "contains":
		return strings.Contains(strings.ToLower(item.Title), strings.ToLower(f.Pattern))
	case "regex":
		matched, err := regexp.MatchString(f.Pattern, item.Title)
		return err == nil && matched
	default:
		return strings.Contains(strings.ToLower(item.Title), strings.ToLower(f.Pattern))
	}
}

// CategoryFilter 分类过滤器
type CategoryFilter struct {
	Pattern string
}

// Match 检查分类是否匹配
func (f *CategoryFilter) Match(item RSSItem) bool {
	return strings.Contains(strings.ToLower(item.Category), strings.ToLower(f.Pattern))
}

// DateFilter 日期过滤器
type DateFilter struct {
	From time.Time
	To   time.Time
}

// Match 检查日期是否在范围内
func (f *DateFilter) Match(item RSSItem) bool {
	return !item.PubDate.Before(f.From) && !item.PubDate.After(f.To)
}
