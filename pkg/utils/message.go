package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MessageHelper 消息辅助工具
type MessageHelper struct{}

// NewMessageHelper 创建消息辅助工具实例
func NewMessageHelper() *MessageHelper {
	return &MessageHelper{}
}

// MessageType 消息类型
type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypeMarkdown
	MessageTypeHTML
	MessageTypeImage
	MessageTypeVideo
	MessageTypeAudio
	MessageTypeFile
	MessageTypeLink
	MessageTypeCode
	MessageTypeQuote
	MessageTypeTable
)

// Message 消息结构
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Content   string                 `json:"content"`
	Title     string                 `json:"title"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewMessage 创建新消息
func (m *MessageHelper) NewMessage(msgType MessageType, content, title, author string) *Message {
	return &Message{
		ID:        m.GenerateMessageID(content),
		Type:      msgType,
		Content:   content,
		Title:     title,
		Timestamp: time.Now(),
		Author:    author,
		Metadata:  make(map[string]interface{}),
	}
}

// GenerateMessageID 生成消息ID
func (m *MessageHelper) GenerateMessageID(content string) string {
	hash := md5.Sum([]byte(content + time.Now().String()))
	return hex.EncodeToString(hash[:])
}

// ValidateMessage 验证消息
func (m *MessageHelper) ValidateMessage(msg *Message) error {
	if msg.ID == "" {
		return fmt.Errorf("消息ID不能为空")
	}
	if msg.Content == "" {
		return fmt.Errorf("消息内容不能为空")
	}
	if msg.Timestamp.IsZero() {
		return fmt.Errorf("消息时间戳不能为空")
	}
	return nil
}

// SanitizeMessage 清理消息内容
func (m *MessageHelper) SanitizeMessage(content string) string {
	// 移除潜在的恶意脚本
	scriptPattern := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = scriptPattern.ReplaceAllString(content, "")

	// 移除危险的HTML标签
	dangerousTags := []string{"iframe", "object", "embed", "form", "input", "button"}
	for _, tag := range dangerousTags {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>.*?</%s>`, tag, tag))
		content = pattern.ReplaceAllString(content, "")
	}

	// 清理多余的空白字符
	spacePattern := regexp.MustCompile(`\s+`)
	content = spacePattern.ReplaceAllString(content, " ")

	return strings.TrimSpace(content)
}

// ExtractLinks 从消息中提取链接
func (m *MessageHelper) ExtractLinks(content string) []string {
	linkPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\^[\]` + "`" + `]+`)
	return linkPattern.FindAllString(content, -1)
}

// ExtractMentions 从消息中提取提及
func (m *MessageHelper) ExtractMentions(content string) []string {
	mentionPattern := regexp.MustCompile(`@(\w+)`)
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	var mentions []string
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}
	return mentions
}

// ExtractHashtags 从消息中提取标签
func (m *MessageHelper) ExtractHashtags(content string) []string {
	hashtagPattern := regexp.MustCompile(`#(\w+)`)
	matches := hashtagPattern.FindAllStringSubmatch(content, -1)
	var hashtags []string
	for _, match := range matches {
		if len(match) > 1 {
			hashtags = append(hashtags, match[1])
		}
	}
	return hashtags
}

// ExtractEmails 从消息中提取邮箱
func (m *MessageHelper) ExtractEmails(content string) []string {
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return emailPattern.FindAllString(content, -1)
}

// ExtractPhoneNumbers 从消息中提取电话号码
func (m *MessageHelper) ExtractPhoneNumbers(content string) []string {
	phonePattern := regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b|\b\d{10}\b|\+?\d{1,3}[-.\s]?\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`)
	return phonePattern.FindAllString(content, -1)
}

// DetectLanguage 检测消息语言（简单实现）
func (m *MessageHelper) DetectLanguage(content string) string {
	// 简单的语言检测逻辑
	chinesePattern := regexp.MustCompile(`[\u4e00-\u9fff]`)
	if chinesePattern.MatchString(content) {
		return "zh"
	}

	englishPattern := regexp.MustCompile(`[a-zA-Z]`)
	if englishPattern.MatchString(content) {
		return "en"
	}

	return "unknown"
}

// CountWords 统计单词数量
func (m *MessageHelper) CountWords(content string) int {
	// 移除标点符号
	punctuationPattern := regexp.MustCompile(`[^\w\s]`)
	cleanContent := punctuationPattern.ReplaceAllString(content, " ")

	// 分割单词
	words := strings.Fields(strings.TrimSpace(cleanContent))
	return len(words)
}

// CountCharacters 统计字符数量
func (m *MessageHelper) CountCharacters(content string) int {
	return len([]rune(content))
}

// EstimateReadingTime 估算阅读时间（分钟）
func (m *MessageHelper) EstimateReadingTime(content string) int {
	wordCount := m.CountWords(content)
	// 假设每分钟阅读200个单词
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}
	return readingTime
}

// IsEmpty 检查消息是否为空
func (m *MessageHelper) IsEmpty(content string) bool {
	return strings.TrimSpace(content) == ""
}

// IsTooLong 检查消息是否过长
func (m *MessageHelper) IsTooLong(content string, maxLength int) bool {
	return len([]rune(content)) > maxLength
}

// TruncateMessage 截断消息
func (m *MessageHelper) TruncateMessage(content string, maxLength int) string {
	runes := []rune(content)
	if len(runes) <= maxLength {
		return content
	}
	if maxLength <= 3 {
		return string(runes[:maxLength])
	}
	return string(runes[:maxLength-3]) + "..."
}

// HighlightKeywords 高亮关键词
func (m *MessageHelper) HighlightKeywords(content string, keywords []string, wrapper string) string {
	result := content
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(keyword))
		result = pattern.ReplaceAllString(result, wrapper+"$0"+wrapper)
	}
	return result
}

// RemoveMarkdown 移除Markdown格式
func (m *MessageHelper) RemoveMarkdown(content string) string {
	// 移除标题
	titlePattern := regexp.MustCompile(`^#{1,6}\s+`)
	content = titlePattern.ReplaceAllString(content, "")

	// 移除粗体和斜体
	boldPattern := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	content = boldPattern.ReplaceAllString(content, "$1")

	italicPattern := regexp.MustCompile(`\*([^*]+)\*`)
	content = italicPattern.ReplaceAllString(content, "$1")

	// 移除链接
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	content = linkPattern.ReplaceAllString(content, "$1")

	// 移除代码块
	codeBlockPattern := regexp.MustCompile("```[^`]*```")
	content = codeBlockPattern.ReplaceAllString(content, "")

	// 移除行内代码
	inlineCodePattern := regexp.MustCompile("`([^`]+)`")
	content = inlineCodePattern.ReplaceAllString(content, "$1")

	return content
}

// ToMarkdown 转换为Markdown格式
func (m *MessageHelper) ToMarkdown(msg *Message) string {
	var builder strings.Builder

	if msg.Title != "" {
		builder.WriteString("# " + msg.Title + "\n\n")
	}

	switch msg.Type {
	case MessageTypeCode:
		builder.WriteString("```\n" + msg.Content + "\n```\n")
	case MessageTypeQuote:
		lines := strings.Split(msg.Content, "\n")
		for _, line := range lines {
			builder.WriteString("> " + line + "\n")
		}
	case MessageTypeLink:
		builder.WriteString("[" + msg.Title + "](" + msg.Content + ")\n")
	default:
		builder.WriteString(msg.Content + "\n")
	}

	if msg.Author != "" {
		builder.WriteString("\n---\n*" + msg.Author + " - " + msg.Timestamp.Format("2006-01-02 15:04:05") + "*\n")
	}

	return builder.String()
}

// ToHTML 转换为HTML格式
func (m *MessageHelper) ToHTML(msg *Message) string {
	var builder strings.Builder

	builder.WriteString("<div class=\"message\">\n")

	if msg.Title != "" {
		builder.WriteString("<h1>" + msg.Title + "</h1>\n")
	}

	switch msg.Type {
	case MessageTypeCode:
		builder.WriteString("<pre><code>" + msg.Content + "</code></pre>\n")
	case MessageTypeQuote:
		builder.WriteString("<blockquote>" + msg.Content + "</blockquote>\n")
	case MessageTypeLink:
		builder.WriteString("<a href=\"" + msg.Content + "\">" + msg.Title + "</a>\n")
	default:
		builder.WriteString("<p>" + strings.ReplaceAll(msg.Content, "\n", "<br>") + "</p>\n")
	}

	if msg.Author != "" {
		builder.WriteString("<hr>\n")
		builder.WriteString("<small><em>" + msg.Author + " - " + msg.Timestamp.Format("2006-01-02 15:04:05") + "</em></small>\n")
	}

	builder.WriteString("</div>")
	return builder.String()
}

// ToPlainText 转换为纯文本格式
func (m *MessageHelper) ToPlainText(msg *Message) string {
	var builder strings.Builder

	if msg.Title != "" {
		builder.WriteString(msg.Title + "\n\n")
	}

	builder.WriteString(msg.Content + "\n")

	if msg.Author != "" {
		builder.WriteString("\n---\n" + msg.Author + " - " + msg.Timestamp.Format("2006-01-02 15:04:05") + "\n")
	}

	return builder.String()
}

// AddMetadata 添加元数据
func (m *MessageHelper) AddMetadata(msg *Message, key string, value interface{}) {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}
	msg.Metadata[key] = value
}

// GetMetadata 获取元数据
func (m *MessageHelper) GetMetadata(msg *Message, key string) (interface{}, bool) {
	if msg.Metadata == nil {
		return nil, false
	}
	value, exists := msg.Metadata[key]
	return value, exists
}

// RemoveMetadata 移除元数据
func (m *MessageHelper) RemoveMetadata(msg *Message, key string) {
	if msg.Metadata != nil {
		delete(msg.Metadata, key)
	}
}

// CloneMessage 克隆消息
func (m *MessageHelper) CloneMessage(msg *Message) *Message {
	clone := &Message{
		ID:        msg.ID,
		Type:      msg.Type,
		Content:   msg.Content,
		Title:     msg.Title,
		Timestamp: msg.Timestamp,
		Author:    msg.Author,
		Metadata:  make(map[string]interface{}),
	}

	// 深拷贝元数据
	for k, v := range msg.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// MergeMessages 合并消息
func (m *MessageHelper) MergeMessages(messages []*Message) *Message {
	if len(messages) == 0 {
		return nil
	}

	if len(messages) == 1 {
		return messages[0]
	}

	var builder strings.Builder
	var authors []string

	for _, msg := range messages {
		if msg.Title != "" {
			builder.WriteString(msg.Title + "\n")
		}
		builder.WriteString(msg.Content + "\n\n")
		if msg.Author != "" {
			authors = append(authors, msg.Author)
		}
	}

	mergedAuthor := strings.Join(authors, ", ")
	mergedContent := strings.TrimSpace(builder.String())

	return m.NewMessage(MessageTypeText, mergedContent, "", mergedAuthor)
}

// FilterMessages 过滤消息
func (m *MessageHelper) FilterMessages(messages []*Message, filter func(*Message) bool) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if filter(msg) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// SortMessages 排序消息
func (m *MessageHelper) SortMessages(messages []*Message, compare func(*Message, *Message) bool) {
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if compare(messages[j], messages[i]) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
}