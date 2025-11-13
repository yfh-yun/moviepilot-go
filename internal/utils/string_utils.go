package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// StringUtils 字符串工具类
type StringUtils struct{}

// NewStringUtils 创建字符串工具类实例
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// IsChinese 判断字符串是否包含中文字�?func (su *StringUtils) IsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

// FormatTimestamp 格式化时间戳
func (su *StringUtils) FormatTimestamp(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	
	// 转换为时间字符串
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// FormatTimestampStr 时间戳转日期
func (su *StringUtils) FormatTimestampStr(timestamp string, dateFormat ...string) string {
	// 如果没有提供dateFormat参数，则使用默认格式
	format := "2006-01-02 15:04:05"
	if len(dateFormat) > 0 {
		format = dateFormat[0]
	}
	
	// 如果timestamp不是数字字符串，直接返回
	if timestamp == "" {
		return timestamp
	}
	
	// 检查是否为数字字符�?	for _, r := range timestamp {
		if r < '0' || r > '9' {
			return timestamp
		}
	}
	
	// 转换为整�?	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return timestamp
	}
	
	// 转换为时间并格式�?	t := time.Unix(ts, 0)
	return t.Format(format)
}

// UnifyDateTimeStr 日期时间格式�?统一转成 2020-10-14 07:48:04 这种格式
func (su *StringUtils) UnifyDateTimeStr(dateTimeStr string) string {
	// 传入的参数如果是空字符串直接返回
	if dateTimeStr == "" {
		return dateTimeStr
	}
	
	// 尝试多种时间格式解析
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 MST",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	
	// 如果所有格式都无法解析，返回原始字符串
	return dateTimeStr
}

// IsLink 检查文件是否为链接地址，支持各类协�?func (su *StringUtils) IsLink(text string) bool {
	if text == "" {
		return false
	}
	
	// 检查是否以http、https、ftp等协议开�?	linkRegex := regexp.MustCompile(`^(http|https|ftp|ftps|sftp|ws|wss)://`)
	if linkRegex.MatchString(text) {
		return true
	}
	
	// 检查是否为IP地址或域�?	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})?$`)
	if domainRegex.MatchString(text) {
		return true
	}
	
	return false
}

// IsMagnetLink 判断内容是否为磁力链�?func (su *StringUtils) IsMagnetLink(content string) bool {
	if content == "" {
		return false
	}
	
	if strings.HasPrefix(content, "magnet:") {
		return true
	}
	
	return false
}

// IsValidHTMLElement 检查HTML元素是否有效
func (su *StringUtils) IsValidHTMLElement(elem string) bool {
	return elem != "" && len(elem) > 0
}

// NumFilesize 将文件大小文本转化为字节
func (su *StringUtils) NumFilesize(text interface{}) int {
	// 如果text为空，返�?
	if text == nil {
		return 0
	}
	
	// 转换为字符串
	var textStr string
	switch v := text.(type) {
	case string:
		textStr = v
	case int:
		return v
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		textStr = strconv.Itoa(v)
	}
	
	// 如果为空字符串，返回0
	if textStr == "" {
		return 0
	}
	
	// 如果是纯数字字符串，直接转换返回
	if _, err := strconv.Atoi(textStr); err == nil {
		size, _ := strconv.Atoi(textStr)
		return size
	}
	
	// 去除逗号和空格并转为大写
	textStr = strings.ReplaceAll(textStr, ",", "")
	textStr = strings.ReplaceAll(textStr, " ", "")
	textStr = strings.ToUpper(textStr)
	
	// 提取数字部分（去除单位）
	sizeRegex := regexp.MustCompile(`(?i)[KMGTPI]*B?`)
	sizeStr := sizeRegex.ReplaceAllString(textStr, "")
	
	var size float64
	var err error
	if sizeStr != "" {
		size, err = strconv.ParseFloat(sizeStr, 64)
		if err != nil {
			return 0
		}
	}
	
	// 根据单位计算字节�?	if strings.Contains(textStr, "PB") || strings.Contains(textStr, "PIB") {
		size *= math.Pow(1024, 5)
	} else if strings.Contains(textStr, "TB") || strings.Contains(textStr, "TIB") {
		size *= math.Pow(1024, 4)
	} else if strings.Contains(textStr, "GB") || strings.Contains(textStr, "GIB") {
		size *= math.Pow(1024, 3)
	} else if strings.Contains(textStr, "MB") || strings.Contains(textStr, "MIB") {
		size *= math.Pow(1024, 2)
	} else if strings.Contains(textStr, "KB") || strings.Contains(textStr, "KIB") {
		size *= 1024
	}
	
	return int(math.Round(size))
}

// StrInt 将字符串转换为整�?func (su *StringUtils) StrInt(text string) int {
	if text == "" {
		return 0
	}
	
	// 去除空格和逗号
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "")
	
	// 转换为整�?	if num, err := strconv.Atoi(text); err == nil {
		return num
	}
	
	return 0
}

// StrFloat 将字符串转换为浮点数
func (su *StringUtils) StrFloat(text string) float64 {
	if text == "" {
		return 0.0
	}
	
	// 去除空格和逗号
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "")
	
	// 转换为浮点数
	if num, err := strconv.ParseFloat(text, 64); err == nil {
		return num
	}
	
	return 0.0
}

// GetUrlDomain 获取URL的域�?func (su *StringUtils) GetUrlDomain(domain string) string {
	if domain == "" {
		return domain
	}
	
	// 移除端口�?	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	
	// 移除可能的端口部�?	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	
	return domain
}
