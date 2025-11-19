package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FormatHelper 格式化辅助工具
type FormatHelper struct{}

// NewFormatHelper 创建格式化辅助工具实例
func NewFormatHelper() *FormatHelper {
	return &FormatHelper{}
}

// FormatBytes 格式化字节数
func (f *FormatHelper) FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration 格式化时间间隔
func (f *FormatHelper) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), float64(int(d.Seconds())%60))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh %.0fm", d.Hours(), float64(int(d.Minutes())%60))
	} else {
		days := int(d.Hours()) / 24
		hours := float64(int(d.Hours())%24)
		return fmt.Sprintf("%dd %.0fh", days, hours)
	}
}

// FormatNumber 格式化数字
func (f *FormatHelper) FormatNumber(num float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, num)
}

// FormatFileSize 格式化文件大小
func (f *FormatHelper) FormatFileSize(size int64) string {
	return f.FormatBytes(size)
}

// FormatPercentage 格式化百分比
func (f *FormatHelper) FormatPercentage(value, total float64, decimals int) string {
	if total == 0 {
		return "0%"
	}
	percentage := (value / total) * 100
	format := fmt.Sprintf("%%.%df%%%%", decimals)
	return fmt.Sprintf(format, percentage)
}

// FormatRate 格式化速率
func (f *FormatHelper) FormatRate(bytes int64, duration time.Duration) string {
	if duration == 0 {
		return "0 B/s"
	}
	rate := float64(bytes) / duration.Seconds()
	return f.FormatBytes(int64(rate)) + "/s"
}

// FormatTimestamp 格式化时间戳
func (f *FormatHelper) FormatTimestamp(timestamp int64, layout string) string {
	if timestamp == 0 {
		return ""
	}
	t := time.Unix(timestamp, 0)
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Format(layout)
}

// FormatTimeAgo 格式化为"多久以前"
func (f *FormatHelper) FormatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%.0f分钟前", diff.Minutes())
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%.0f小时前", diff.Hours())
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours()) / 24
		return fmt.Sprintf("%d天前", days)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours()) / (24 * 30)
		return fmt.Sprintf("%d个月前", months)
	} else {
		years := int(diff.Hours()) / (24 * 365)
		return fmt.Sprintf("%d年前", years)
	}
}

// TruncateString 截断字符串
func (f *FormatHelper) TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// PadLeft 左填充
func (f *FormatHelper) PadLeft(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(pad, length-len(s))
	return padding + s
}

// PadRight 右填充
func (f *FormatHelper) PadRight(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(pad, length-len(s))
	return s + padding
}

// CenterString 居中字符串
func (f *FormatHelper) CenterString(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	totalPad := length - len(s)
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	return strings.Repeat(pad, leftPad) + s + strings.Repeat(pad, rightPad)
}

// Capitalize 首字母大写
func (f *FormatHelper) Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// TitleCase 标题格式（每个单词首字母大写）
func (f *FormatHelper) TitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// SnakeCase 蛇形命名
func (f *FormatHelper) SnakeCase(s string) string {
	// 将驼峰命名转换为蛇形命名
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// CamelCase 驼峰命名
func (f *FormatHelper) CamelCase(s string) string {
	words := strings.Split(strings.ToLower(s), "_")
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, "")
}

// KebabCase 短横线命名
func (f *FormatHelper) KebabCase(s string) string {
	return strings.ReplaceAll(f.SnakeCase(s), "_", "-")
}

// RemoveWhitespace 移除空白字符
func (f *FormatHelper) RemoveWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, "")
}

// RemoveSpecialChars 移除特殊字符
func (f *FormatHelper) RemoveSpecialChars(s string) string {
	re := regexp.MustCompile(`[^\w\s]`)
	return re.ReplaceAllString(s, "")
}

// EscapeHTML 转义HTML
func (f *FormatHelper) EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// UnescapeHTML 反转义HTML
func (f *FormatHelper) UnescapeHTML(s string) string {
	return html.UnescapeString(s)
}

// ToJSON 转换为JSON字符串
func (f *FormatHelper) ToJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("转换为JSON失败: %w", err)
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (f *FormatHelper) FromJSON(data string, v interface{}) error {
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		return fmt.Errorf("从JSON解析失败: %w", err)
	}
	return nil
}

// FormatPhoneNumber 格式化电话号码
func (f *FormatHelper) FormatPhoneNumber(phone string) string {
	// 移除所有非数字字符
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(phone, "")

	// 根据长度格式化
	switch len(digits) {
	case 11: // 中国手机号
		return fmt.Sprintf("%s-%s-%s", digits[:3], digits[3:7], digits[7:])
	case 10: // 美国手机号
		return fmt.Sprintf("(%s) %s-%s", digits[:3], digits[3:6], digits[6:])
	default:
		return phone
	}
}

// FormatCreditCard 格式化信用卡号
func (f *FormatHelper) FormatCreditCard(card string) string {
	// 移除所有非数字字符
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(card, "")

	// 每4位添加空格
	var formatted strings.Builder
	for i, digit := range digits {
		if i > 0 && i%4 == 0 {
			formatted.WriteString(" ")
		}
		formatted.WriteRune(digit)
	}
	return formatted.String()
}

// MaskString 遮蔽字符串
func (f *FormatHelper) MaskString(s string, start, end int, mask string) string {
	if len(s) <= start+end {
		return s
	}
	masked := strings.Repeat(mask, len(s)-start-end)
	return s[:start] + masked + s[len(s)-end:]
}

// FormatVersion 格式化版本号
func (f *FormatHelper) FormatVersion(major, minor, patch int) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// ParseVersion 解析版本号
func (f *FormatHelper) ParseVersion(version string) (major, minor, patch int, err error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("无效的版本号格式")
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析主版本号失败: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析次版本号失败: %w", err)
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析修订版本号失败: %w", err)
	}

	return major, minor, patch, nil
}

// CompareVersion 比较版本号
func (f *FormatHelper) CompareVersion(v1, v2 string) int {
	major1, minor1, patch1, err := f.ParseVersion(v1)
	if err != nil {
		return 0
	}

	major2, minor2, patch2, err := f.ParseVersion(v2)
	if err != nil {
		return 0
	}

	if major1 != major2 {
		if major1 > major2 {
			return 1
		}
		return -1
	}

	if minor1 != minor2 {
		if minor1 > minor2 {
			return 1
		}
		return -1
	}

	if patch1 != patch2 {
		if patch1 > patch2 {
			return 1
		}
		return -1
	}

	return 0
}

// FormatMacAddress 格式化MAC地址
func (f *FormatHelper) FormatMacAddress(mac string) string {
	// 移除所有非十六进制字符
	re := regexp.MustCompile(`[^0-9a-fA-F]`)
	cleaned := re.ReplaceAllString(mac, "")

	// 转换为大写并格式化
	cleaned = strings.ToUpper(cleaned)
	if len(cleaned) != 12 {
		return mac
	}

	var formatted strings.Builder
	for i := 0; i < len(cleaned); i += 2 {
		if i > 0 {
			formatted.WriteString(":")
		}
		formatted.WriteString(cleaned[i : i+2])
	}
	return formatted.String()
}

// FormatIPAddress 格式化IP地址
func (f *FormatHelper) FormatIPAddress(ip string) string {
	// 简单的IP地址格式化验证
	re := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if re.MatchString(ip) {
		return ip
	}
	return ip
}

// IndentString 缩进字符串
func (f *FormatHelper) IndentString(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// WrapText 文本换行
func (f *FormatHelper) WrapText(text string, width int) string {
	var buf bytes.Buffer
	var currentLine strings.Builder

	words := strings.Fields(text)
	for i, word := range words {
		if currentLine.Len() > 0 {
			if currentLine.Len()+1+len(word) > width {
				buf.WriteString(currentLine.String())
				buf.WriteString("\n")
				currentLine.Reset()
			} else {
				currentLine.WriteString(" ")
			}
		}
		currentLine.WriteString(word)

		if i == len(words)-1 {
			buf.WriteString(currentLine.String())
		}
	}

	return buf.String()
}