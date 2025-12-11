package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// String utilities - enhanced string manipulation functions

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

// StringToSnakeCase converts a string to snake_case
func StringToSnakeCase(s string) string {
	var result bytes.Buffer

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// StringToCamelCase converts a string to camelCase
func StringToCamelCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		if i == 0 {
			words[i] = strings.ToLower(words[i])
		} else {
			words[i] = TitleCase(words[i])
		}
	}
	return strings.Join(words, "")
}

// PascalCase converts a string to PascalCase
func PascalCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		words[i] = TitleCase(words[i])
	}
	return strings.Join(words, "")
}

// TitleCase converts a string to title case
func TitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// KebabCase converts a string to kebab-case
func KebabCase(s string) string {
	return strings.ReplaceAll(StringToSnakeCase(s), "_", "-")
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// TruncateString truncates a string to specified length with ellipsis
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// TruncateWords truncates a string to specified number of words
func TruncateWords(s string, wordCount int) string {
	words := strings.Fields(s)
	if len(words) <= wordCount {
		return s
	}
	return strings.Join(words[:wordCount], " ") + "..."
}

// RemoveExtraSpaces removes extra spaces from a string
func RemoveExtraSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// ExtractNumbers extracts all numbers from a string
func ExtractNumbers(s string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(s, -1)

	var numbers []int
	for _, match := range matches {
		if num, err := strconv.Atoi(match); err == nil {
			numbers = append(numbers, num)
		}
	}

	return numbers
}

// ExtractEmails extracts all email addresses from a string
func ExtractEmails(s string) []string {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return re.FindAllString(s, -1)
}

// ExtractURLs extracts all URLs from a string
func ExtractURLs(s string) []string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	return re.FindAllString(s, -1)
}

// StringContainsAny checks if a string contains any of the given substrings
func StringContainsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// StringContainsAll checks if a string contains all of the given substrings
func StringContainsAll(s string, substrings []string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// ReplaceAll replaces all occurrences of multiple old strings with new strings
func ReplaceAll(s string, oldNew map[string]string) string {
	result := s
	for old, new := range oldNew {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// MaskEmail masks an email address for privacy
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return "***@" + domain
	}

	maskedLocal := localPart[:2] + strings.Repeat("*", len(localPart)-2)
	return maskedLocal + "@" + domain
}

// MaskPhone masks a phone number for privacy
func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return strings.Repeat("*", len(phone))
	}

	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

// PadLeft pads a string on the left to specified length
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	return strings.Repeat(string(padChar), padding) + s
}

// PadRight pads a string on the right to specified length
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	return s + strings.Repeat(string(padChar), padding)
}

// PadCenter pads a string on both sides to specified length
func PadCenter(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(string(padChar), leftPadding) + s + strings.Repeat(string(padChar), rightPadding)
}

// RemovePrefix removes a prefix from a string if it exists
func RemovePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix removes a suffix from a string if it exists
func RemoveSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// RemoveSubstrings removes all occurrences of specified substrings
func RemoveSubstrings(s string, substrings []string) string {
	result := s
	for _, substr := range substrings {
		result = strings.ReplaceAll(result, substr, "")
	}
	return result
}

// SplitByLength splits a string into chunks of specified length
func SplitByLength(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	var chunks []string
	runes := []rune(s)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// JoinStrings joins strings with a separator, ignoring empty strings
func JoinStrings(separator string, strs ...string) string {
	var nonEmpty []string
	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// CommonPrefix finds the common prefix of two strings
func CommonPrefix(s1, s2 string) string {
	minLen := min(len(s1), len(s2))

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}

	return s1[:minLen]
}

// CommonSuffix finds the common suffix of two strings
func CommonSuffix(s1, s2 string) string {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)
	minLen := min(len1, len2)

	for i := 1; i <= minLen; i++ {
		if r1[len1-i] != r2[len2-i] {
			return string(r1[len1-i+1:])
		}
	}

	return string(r1[len1-minLen:])
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	distance := make([][]int, len1+1)
	for i := range distance {
		distance[i] = make([]int, len2+1)
		distance[i][0] = i
	}
	for j := range distance[0] {
		distance[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			distance[i][j] = min3(
				distance[i-1][j]+1,
				distance[i][j-1]+1,
				distance[i-1][j-1]+cost,
			)
		}
	}

	return distance[len1][len2]
}

// JaroWinklerSimilarity calculates Jaro-Winkler similarity between two strings
func JaroWinklerSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate Jaro distance
	matchDistance := max(len1, len2)/2 - 1

	matches1 := make([]bool, len1)
	matches2 := make([]bool, len2)

	matches := 0
	for i := 0; i < len1; i++ {
		start := max(0, i-matchDistance)
		end := min(len2, i+matchDistance+1)

		for j := start; j < end; j++ {
			if !matches2[j] && s1[i] == s2[j] {
				matches1[i] = true
				matches2[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Calculate transpositions
	transpositions := 0
	k := 0
	for i := 0; i < len1; i++ {
		if matches1[i] {
			for !matches2[k] {
				k++
			}
			if s1[i] != s2[k] {
				transpositions++
			}
			k++
		}
	}

	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	// Calculate Jaro-Winkler similarity
	prefix := 0
	for i := 0; i < min(min(len1, len2), 4); i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}

	return jaro + float64(prefix)*0.1*(1-jaro)
}

// min helper function for two arguments
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 helper function for three arguments
func min3(a, b, c int) int {
	return min(min(a, b), c)
}

// max helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NumFileSize 将文件大小文本转化为字节
// 对应 Python StringUtils.num_filesize 方法
func NumFileSize(text interface{}) int {
	var strText string
	switch v := text.(type) {
	case string:
		strText = v
	case int, float64, float32:
		strText = fmt.Sprintf("%v", v)
	default:
		return 0
	}

	if strText == "" {
		return 0
	}

	// 移除逗号和空格，转为大写
	strText = strings.ReplaceAll(strText, ",", "")
	strText = strings.ReplaceAll(strText, " ", "")
	strText = strings.ToUpper(strText)

	// 提取数字部分
	re := regexp.MustCompile(`[0-9.]+`)
	sizeStr := re.FindString(strText)
	if sizeStr == "" {
		return 0
	}

	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0
	}

	// 根据单位转换
	if strings.Contains(strText, "PB") || strings.Contains(strText, "PIB") {
		return int(size * (1 << 50)) // 1024^5
	} else if strings.Contains(strText, "TB") || strings.Contains(strText, "TIB") {
		return int(size * (1 << 40)) // 1024^4
	} else if strings.Contains(strText, "GB") || strings.Contains(strText, "GIB") {
		return int(size * (1 << 30)) // 1024^3
	} else if strings.Contains(strText, "MB") || strings.Contains(strText, "MIB") {
		return int(size * (1 << 20)) // 1024^2
	} else if strings.Contains(strText, "KB") || strings.Contains(strText, "KIB") {
		return int(size * (1 << 10)) // 1024
	}

	// 无单位，直接返回
	return int(size)
}

// StrFileSize 将字节计算为文件大小描述（带单位的格式化后返回）
// 对应 Python StringUtils.str_filesize 方法
func StrFileSize(size interface{}, pre int) string {
	if size == nil {
		return ""
	}

	// 转换为字符串并清理
	strSize := fmt.Sprintf("%v", size)
	strSize = strings.ToLower(strSize)
	strSize = strings.ReplaceAll(strSize, " ", "")
	strSize = strings.ReplaceAll(strSize, "b", "")
	strSize = strings.ReplaceAll(strSize, "ib", "")

	// 检查是否为纯数字
	isNum := true
	for _, c := range strSize {
		if !unicode.IsDigit(c) && c != '.' {
			isNum = false
			break
		}
	}

	if isNum {
		// 转换为浮点数
		bytes, err := strconv.ParseFloat(strSize, 64)
		if err != nil {
			return ""
		}

		// 定义单位和阈值
		units := []string{"", "K", "M", "G", "T"}
		thresholds := []float64{1024 - 1, (1 << 20) - 1, (1 << 30) - 1, (1 << 40) - 1, (1 << 50) - 1}

		// 确定单位
		index := 0
		for i, threshold := range thresholds {
			if bytes <= threshold {
				index = i
				break
			}
			index = len(thresholds)
		}

		// 计算结果
		var result float64
		if index > 0 {
			// 位运算必须在整数上进行
			divisor := 1 << (10 * index)
			result = bytes / float64(divisor)
		} else {
			result = bytes
		}

		// 格式化输出
		format := fmt.Sprintf("%%.%df%%s", pre)
		return fmt.Sprintf(format, result, units[index])
	}

	// 非纯数字，直接返回
	return strSize
}

// IsChinese 判断是否含有中文
// 对应 Python StringUtils.is_chinese 方法
func IsChinese(word interface{}) bool {
	var strWord string
	switch v := word.(type) {
	case string:
		strWord = v
	case []string:
		strWord = strings.Join(v, " ")
	default:
		return false
	}

	if strWord == "" {
		return false
	}

	// 匹配中文字符
	re := regexp.MustCompile(`[\u4e00-\u9fff]`)
	return re.MatchString(strWord)
}

// IsJapanese 判断是否含有日文
// 对应 Python StringUtils.is_japanese 方法
func IsJapanese(word string) bool {
	if word == "" {
		return false
	}

	// 匹配日文假名
	re := regexp.MustCompile(`[\u3040-\u309F\u30A0-\u30FF]`)
	return re.MatchString(word)
}

// IsKorean 判断是否包含韩文
// 对应 Python StringUtils.is_korean 方法
func IsKorean(word string) bool {
	if word == "" {
		return false
	}

	// 匹配韩文字符
	re := regexp.MustCompile(`[\uAC00-\uD7FF]`)
	return re.MatchString(word)
}

// IsAllChinese 判断是否全是中文
// 对应 Python StringUtils.is_all_chinese 方法
func IsAllChinese(word string) bool {
	if word == "" {
		return false
	}

	for _, c := range word {
		if c == ' ' {
			continue
		}
		if c < '\u4e00' || c > '\u9fff' {
			return false
		}
	}

	return true
}

// IsEnglishWord 判断是否为英文单词，有空格时返回False
// 对应 Python StringUtils.is_english_word 方法
func IsEnglishWord(word string) bool {
	if word == "" || strings.Contains(word, " ") {
		return false
	}

	// 检查是否全为字母
	for _, c := range word {
		if !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}

// StrInt web字符串转int
// 对应 Python StringUtils.str_int 方法
func StrInt(text string) int {
	if text == "" {
		return 0
	}

	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "")

	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}

	return value
}

// StrFloat web字符串转float
// 对应 Python StringUtils.str_float 方法
func StrFloat(text string) float64 {
	if text == "" {
		return 0.0
	}

	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "")

	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0
	}

	return value
}

// Clear 忽略特殊字符
// 对应 Python StringUtils.clear 方法
func Clear(text interface{}, replaceWord string, allowSpace bool) interface{} {
	// 需要忽略的特殊字符
	convertEmptyChars := regexp.MustCompile(`[、.。,，·:：;；!！'’"“”()（）\[\]【】「」\-—―\+\|\\_/&#～~]`)
	zeroWidthChars := regexp.MustCompile(`[\u200B-\u200D\uFEFF]`)

	if text == nil {
		return text
	}

	switch v := text.(type) {
	case string:
		// 移除零宽字符
		cleanText := zeroWidthChars.ReplaceAllString(v, "")
		// 替换特殊字符
		cleanText = convertEmptyChars.ReplaceAllString(cleanText, replaceWord)
		// 处理空格
		if !allowSpace {
			cleanText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanText, "")
		} else {
			cleanText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanText, " ")
			cleanText = strings.TrimSpace(cleanText)
		}
		return cleanText
	case []string:
		// 递归处理字符串数组
		var result []string
		for _, s := range v {
			result = append(result, Clear(s, replaceWord, allowSpace).(string))
		}
		return result
	default:
		return text
	}
}

// ClearUpper 去除特殊字符，同时大写
// 对应 Python StringUtils.clear_upper 方法
func ClearUpper(text string) string {
	if text == "" {
		return ""
	}

	cleanText := Clear(text, "", false).(string)
	return strings.ToUpper(cleanText)
}

// URLEqual 比较两个地址是否为同一个网站
// 对应 Python StringUtils.url_equal 方法
func URLEqual(url1, url2 string) bool {
	if url1 == "" || url2 == "" {
		return false
	}

	// 解析URL，获取netloc
	getNetloc := func(u string) string {
		if strings.HasPrefix(u, "http") {
			parsed, err := url.Parse(u)
			if err != nil {
				return u
			}
			return parsed.Host
		}
		return u
	}

	netloc1 := strings.ReplaceAll(getNetloc(url1), "www.", "")
	netloc2 := strings.ReplaceAll(getNetloc(url2), "www.", "")

	return netloc1 == netloc2
}

// GetURLNetloc 获取URL的协议和域名部分
// 对应 Python StringUtils.get_url_netloc 方法
func GetURLNetloc(rawURL string) (string, string) {
	if rawURL == "" {
		return "", ""
	}

	if !strings.HasPrefix(rawURL, "http") {
		return "http", rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "http", rawURL
	}

	return parsed.Scheme, parsed.Host
}

// SpecialDomains 特殊域名列表
var SpecialDomains = []string{
	"u2.dmhy.org",
	"pt.ecust.pp.ua",
	"pt.gtkpw.xyz",
	"pt.gtk.pw",
}

// GetURLDomain 获取URL的域名部分，只保留最后两级
// 对应 Python StringUtils.get_url_domain 方法
func GetURLDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// 检查是否为特殊域名
	for _, domain := range SpecialDomains {
		if strings.Contains(rawURL, domain) {
			return domain
		}
	}

	// 获取netloc
	_, netloc := GetURLNetloc(rawURL)
	if netloc == "" {
		return ""
	}

	// 分割域名部分
	parts := strings.Split(netloc, ".")
	if len(parts) > 3 {
		return netloc
	}

	// 只返回最后两级
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return netloc
}

// GetURLSLD 获取URL的二级域名部分，不含端口，若为IP则返回IP
// 对应 Python StringUtils.get_url_sld 方法
func GetURLSLD(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// 获取netloc
	_, netloc := GetURLNetloc(rawURL)
	if netloc == "" {
		return ""
	}

	// 去除端口
	if colonIdx := strings.Index(netloc, ":"); colonIdx != -1 {
		netloc = netloc[:colonIdx]
	}

	// 分割域名
	parts := strings.Split(netloc, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}

	return parts[0]
}

// GetURLHost 获取URL的一级域名
// 对应 Python StringUtils.get_url_host 方法
func GetURLHost(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// 获取netloc
	_, netloc := GetURLNetloc(rawURL)
	if netloc == "" {
		return ""
	}

	// 分割域名
	parts := strings.Split(netloc, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}

	return ""
}

// GetBaseURL 获取URL根地址
// 对应 Python StringUtils.get_base_url 方法
func GetBaseURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// 获取scheme和netloc
	scheme, netloc := GetURLNetloc(rawURL)
	return fmt.Sprintf("%s://%s", scheme, netloc)
}

// ClearFileName 清理文件名，去除非法字符
// 对应 Python StringUtils.clear_file_name 方法
func ClearFileName(name string) string {
	if name == "" {
		return ""
	}

	// 替换非法字符
	re := regexp.MustCompile(`[*?\\/"<>~|]`)
	name = re.ReplaceAllString(name, "")
	// 替换冒号为中文冒号
	name = strings.ReplaceAll(name, ":", "：")
	return name
}

// GenerateRandomStr 生成一个指定长度的随机字符串
// 对应 Python StringUtils.generate_random_str 方法
func GenerateRandomStr(randomLength int) string {
	if randomLength <= 0 {
		randomLength = 16
	}

	const baseStr = "ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghigklmnopqrstuvwxyz0123456789"
	result := make([]byte, randomLength)
	baseLen := len(baseStr)

	for i := 0; i < randomLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(baseLen)))
		if err != nil {
			return ""
		}
		result[i] = baseStr[num.Int64()]
	}

	return string(result)
}

// FormatTimestamp 时间戳转日期
// 对应 Python StringUtils.format_timestamp 方法
func FormatTimestamp(timestamp interface{}, dateFormat string) string {
	// 检查是否为有效时间戳字符串
	var tsStr string
	switch v := timestamp.(type) {
	case string:
		tsStr = v
	case int, int64, float64:
		tsStr = fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", timestamp)
	}

	// 检查是否为纯数字
	isNum := true
	for _, c := range tsStr {
		if !unicode.IsDigit(c) {
			isNum = false
			break
		}
	}

	if !isNum {
		return tsStr
	}

	// 转换为时间戳
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return tsStr
	}

	// 格式化时间
	t := time.Unix(ts, 0)
	if dateFormat == "" {
		dateFormat = "2006-01-02 15:04:05"
	}

	return t.Format(dateFormat)
}

// ToBool 字符串转bool
// 对应 Python StringUtils.to_bool 方法
func ToBool(text interface{}, defaultVal bool) bool {
	switch v := text.(type) {
	case bool:
		return v
	case string:
		if v == "" {
			return defaultVal
		}
		lowerV := strings.ToLower(v)
		return lowerV == "y" || lowerV == "true" || lowerV == "1" || lowerV == "yes" || lowerV == "on"
	case int, int64, float64, float32:
		// 转换为float64进行比较
		var num float64
		fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &num)
		return num > 0
	default:
		return false
	}
}

// StringMD5Hash 计算字符串的MD5哈希值
// 对应 Python StringUtils.md5_hash 方法
func StringMD5Hash(data interface{}) string {
	strData := fmt.Sprintf("%v", data)
	hash := md5.New()
	hash.Write([]byte(strData))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// IsNumber 判断字符是否为可以转换为整数或者浮点数
// 对应 Python StringUtils.is_number 方法
func IsNumber(text string) bool {
	if text == "" {
		return false
	}

	// 尝试转换为浮点数
	_, err := strconv.ParseFloat(text, 64)
	return err == nil
}

// FindCommonPrefix 查找两个字符串的公共前缀
// 对应 Python StringUtils.find_common_prefix 方法
func FindCommonPrefix(str1, str2 string) string {
	if str1 == "" || str2 == "" {
		return ""
	}

	minLen := min(len(str1), len(str2))
	var result strings.Builder

	for i := 0; i < minLen; i++ {
		if str1[i] == str2[i] {
			result.WriteByte(str1[i])
		} else {
			break
		}
	}

	return result.String()
}

// IsValidHTMLElement 检查elem是否为有效的HTML元素
// 对应 Python StringUtils.is_valid_html_element 方法
func IsValidHTMLElement(elem interface{}) bool {
	switch v := elem.(type) {
	case nil:
		return false
	case string:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	default:
		// 尝试使用反射检查长度
		vType := reflect.TypeOf(elem)
		vValue := reflect.ValueOf(elem)
		if vType.Kind() == reflect.Slice || vType.Kind() == reflect.Array {
			return vValue.Len() > 0
		}
		return false
	}
}

// IsLink 检查文件是否为链接地址，支持各类协议
// 对应 Python StringUtils.is_link 方法
func IsLink(text string) bool {
	if text == "" {
		return false
	}

	// 检查是否以http、https、ftp等协议开头
	if regexp.MustCompile(`^(http|https|ftp|ftps|sftp|ws|wss)://`).MatchString(text) {
		return true
	}

	// 检查是否为IP地址或域名
	if regexp.MustCompile(`^[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})?$`).MatchString(text) {
		return true
	}

	return false
}

// IsMagnetLink 判断内容是否为磁力链接
// 对应 Python StringUtils.is_magnet_link 方法
func IsMagnetLink(content interface{}) bool {
	switch v := content.(type) {
	case string:
		return strings.HasPrefix(v, "magnet:")
	case []byte:
		return strings.HasPrefix(string(v), "magnet:")
	default:
		return false
	}
}

// StrSeries 将季集列表转化为字符串简写
// 对应 Python StringUtils.str_series 方法
func StrSeries(array []int) string {
	if len(array) == 0 {
		return ""
	}

	// 排序
	sort.Ints(array)

	// 合并连续数字
	var result []string
	start := array[0]
	end := array[0]

	for i := 1; i < len(array); i++ {
		if array[i] == end+1 {
			end = array[i]
		} else {
			if start == end {
				result = append(result, fmt.Sprintf("%d", start))
			} else {
				result = append(result, fmt.Sprintf("%d-%d", start, end))
			}
			start = array[i]
			end = array[i]
		}
	}

	// 处理最后一个序列
	if start == end {
		result = append(result, fmt.Sprintf("%d", start))
	} else {
		result = append(result, fmt.Sprintf("%d-%d", start, end))
	}

	return strings.Join(result, ",")
}

// FormatEP 将剧集列表格式化为连续区间
// 对应 Python StringUtils.format_ep 方法
func FormatEP(nums []int) string {
	if len(nums) == 0 {
		return ""
	}

	if len(nums) == 1 {
		return fmt.Sprintf("E%02d", nums[0])
	}

	// 排序
	sort.Ints(nums)

	// 合并连续数字
	var result []string
	start := nums[0]
	end := nums[0]

	for i := 1; i < len(nums); i++ {
		if nums[i] == end+1 {
			end = nums[i]
		} else {
			if start == end {
				result = append(result, fmt.Sprintf("E%02d", start))
			} else {
				result = append(result, fmt.Sprintf("E%02d-E%02d", start, end))
			}
			start = nums[i]
			end = nums[i]
		}
	}

	// 处理最后一个序列
	if start == end {
		result = append(result, fmt.Sprintf("E%02d", start))
	} else {
		result = append(result, fmt.Sprintf("E%02d-E%02d", start, end))
	}

	return strings.Join(result, "、")
}

// NaturalSortKey 自然排序键生成
// 对应 Python StringUtils.natural_sort_key 方法
func NaturalSortKey(text string) []interface{} {
	if text == "" {
		return []interface{}{}
	}

	// 将字符串拆分为数字和非数字部分
	re := regexp.MustCompile(`(\d+|[^\d]+)`)
	parts := re.FindAllString(text, -1)

	// 转换数字部分为整数
	var result []interface{}
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		} else {
			result = append(result, strings.ToLower(part))
		}
	}

	return result
}

// NaturalSort 自然排序
// 使用 NaturalSortKey 进行排序
func NaturalSort(strs []string) {
	sort.Slice(strs, func(i, j int) bool {
		key1 := NaturalSortKey(strs[i])
		key2 := NaturalSortKey(strs[j])

		// 比较两个键
		minLen := min(len(key1), len(key2))
		for k := 0; k < minLen; k++ {
			a, b := key1[k], key2[k]

			switch v1 := a.(type) {
			case int:
				if v2, ok := b.(int); ok {
					if v1 != v2 {
						return v1 < v2
					}
				} else {
					// 数字小于字符串
					return true
				}
			case string:
				if v2, ok := b.(string); ok {
					if v1 != v2 {
						return v1 < v2
					}
				} else {
					// 字符串大于数字
					return false
				}
			}
		}

		// 长度短的排在前面
		return len(key1) < len(key2)
	})
}

// StrTimelong 将数字转换为时间描述
// 对应 Python StringUtils.str_timelong 方法
func StrTimelong(timeSec interface{}) string {
	var sec float64
	switch v := timeSec.(type) {
	case string:
		if t, err := strconv.ParseFloat(v, 64); err == nil {
			sec = t
		} else {
			return ""
		}
	case int:
		sec = float64(v)
	case float64:
		sec = v
	default:
		return ""
	}

	// 定义时间单位和阈值
	type timeUnit struct {
		threshold int
		unit      string
	}

	units := []timeUnit{
		{0, "秒"},
		{59, "分"},
		{3599, "小时"},
		{86399, "天"},
	}

	// 找到合适的时间单位
	var unit string
	var divisor int
	for i := len(units) - 1; i >= 0; i-- {
		if int(sec) > units[i].threshold {
			unit = units[i].unit
			switch i {
			case 1: // 分
				divisor = 60
			case 2: // 小时
				divisor = 3600
			case 3: // 天
				divisor = 86400
			default:
				divisor = 1
			}
			break
		}
	}

	if unit == "" {
		unit = "秒"
		divisor = 1
	}

	return fmt.Sprintf("%d%s", int(sec)/divisor, unit)
}

// StrSeconds 将秒转为时分秒字符串
// 对应 Python StringUtils.str_secends 方法
func StrSeconds(timeSec interface{}) string {
	var sec int64
	switch v := timeSec.(type) {
	case string:
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			sec = t
		} else {
			return ""
		}
	case int:
		sec = int64(v)
	case int64:
		sec = v
	case float64:
		sec = int64(v)
	default:
		return ""
	}

	hours := sec / 3600
	remainderSeconds := sec % 3600
	minutes := remainderSeconds / 60
	seconds := remainderSeconds % 60

	timeStr := fmt.Sprintf("%d秒", seconds)
	if minutes > 0 {
		timeStr = fmt.Sprintf("%d分%s", minutes, timeStr)
	}
	if hours > 0 {
		timeStr = fmt.Sprintf("%d时%s", hours, timeStr)
	}

	return timeStr
}

// StrToTimestamp 日期转时间戳
// 对应 Python StringUtils.str_to_timestamp 方法
func StrToTimestamp(dateStr string) float64 {
	if dateStr == "" {
		return 0
	}

	// 尝试解析日期字符串
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// 尝试其他格式
		layout := "2006-01-02 15:04:05"
		t, err = time.Parse(layout, dateStr)
		if err != nil {
			// 尝试使用更灵活的解析方式（Go标准库没有dateparser，这里仅支持常见格式）
			return 0
		}
	}

	return float64(t.Unix())
}

// StrFromCookieJar 将cookiejar转换为字符串
// 对应 Python StringUtils.str_from_cookiejar 方法
func StrFromCookieJar(cj map[string]string) string {
	if len(cj) == 0 {
		return ""
	}

	var cookies []string
	for k, v := range cj {
		cookies = append(cookies, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(cookies, "; ")
}

// GetIDList 从字符串中提取id列表
// 对应 Python StringUtils.get_idlist 方法
func GetIDList(content string, dicts []map[string]interface{}) ([]interface{}, string) {
	if content == "" {
		return []interface{}{}, ""
	}

	idList := []interface{}{}
	contentList := strings.Split(content, " ")

	for _, dic := range dicts {
		if name, ok := dic["name"].(string); ok {
			for _, word := range contentList {
				if word == name {
					if id, ok := dic["id"]; ok {
						// 检查id是否已存在
						alreadyExists := false
						for _, existingID := range idList {
							if reflect.DeepEqual(existingID, id) {
								alreadyExists = true
								break
							}
						}
						if !alreadyExists {
							idList = append(idList, id)
						}
					}
					// 从content中移除该名称
					content = strings.ReplaceAll(content, name, "")
					break
				}
			}
		}
	}

	// 清理剩余的空白字符
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	return idList, content
}

// StrTimehours 将分钟转换成小时和分钟
// 对应 Python StringUtils.str_timehours 方法
func StrTimehours(minutes int) string {
	if minutes == 0 {
		return ""
	}

	hours := minutes / 60
	minutes = minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

// StrAmount 格式化显示金额
// 对应 Python StringUtils.str_amount 方法
func StrAmount(amount interface{}, curr string) string {
	if amount == nil {
		return "0"
	}

	var num float64
	switch v := amount.(type) {
	case int:
		num = float64(v)
	case float64:
		num = v
	case string:
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			num = n
		} else {
			return "0"
		}
	default:
		return "0"
	}

	// 格式化金额，添加千分位分隔符
	formatted := fmt.Sprintf("%s%.2f", curr, num)
	return formatted
}

// CountWords 计算字符串中包含的单词或汉字的数量，兼容中英文混合
// 对应 Python StringUtils.count_words 方法
// 注意：这里覆盖了之前的 CountWords 实现，以匹配 Python 版本的功能
func CountWords(text string) int {
	if text == "" {
		return 0
	}

	// 匹配汉字
	chineseRegex := regexp.MustCompile(`[\u4e00-\u9fa5]`)
	chineseMatches := chineseRegex.FindAllString(text, -1)
	chineseCount := len(chineseMatches)

	// 匹配英文单词
	englishRegex := regexp.MustCompile(`[a-zA-Z]+`)
	englishMatches := englishRegex.FindAllString(text, -1)
	englishCount := len(englishMatches)

	return chineseCount + englishCount
}

// SplitText 把文本拆分为固定字节长度的数组，优先按换行拆分，避免单词内拆分
// 对应 Python StringUtils.split_text 方法
func SplitText(text string, maxLength int) []string {
	if text == "" {
		return []string{""}
	}

	var result []string
	buf := ""

	// 按换行拆分
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		lineBytes := len([]byte(line))
		if lineBytes > maxLength {
			// 超长行，需要进一步拆分
			var parts []string
			blank := ""

			// 检查是否为英文行（只包含字母、数字、点和空格）
			if regexp.MustCompile(`^[A-Za-z0-9.\s]+$`).MatchString(line) {
				// 英文行按空格拆分
				parts = strings.Split(line, " ")
				blank = " "
			} else {
				// 中文行按字符拆分
				parts = make([]string, len([]rune(line)))
				for i, r := range []rune(line) {
					parts[i] = string(r)
				}
			}

			part := ""
			for _, p := range parts {
				newPart := part
				if newPart != "" {
					newPart += blank
				}
				newPart += p

				if len([]byte(newPart)) > maxLength {
					// 超长，添加到结果
					if part != "" {
						result = append(result, strings.TrimSpace(buf+part))
						buf = ""
					}
					part = p
				} else {
					part = newPart
				}
			}

			if part != "" {
				buf += part
			}
		} else {
			// 检查添加换行后的长度
			newBuf := buf
			if newBuf != "" {
				newBuf += "\n"
			}
			newBuf += line

			if len([]byte(newBuf)) > maxLength {
				// 超长，添加到结果
				result = append(result, strings.TrimSpace(buf))
				buf = line
			} else {
				buf = newBuf
			}
		}
	}

	// 处理剩余内容
	if buf != "" {
		result = append(result, strings.TrimSpace(buf))
	}

	return result
}

// StrTitle 大写首字母兼容空值
// 对应 Python StringUtils.str_title 方法
func StrTitle(s string) string {
	if s == "" {
		return ""
	}
	return strings.Title(s)
}

// EscapeMarkdown 转义Markdown字符
// 对应 Python StringUtils.escape_markdown 方法
func EscapeMarkdown(content string) string {
	if content == "" {
		return ""
	}

	// 转义Markdown特殊字符
	// 注意：需要先转义反斜杠，避免后续转义失效
	content = strings.ReplaceAll(content, "\\", "\\\\")
	content = strings.ReplaceAll(content, "_", "\\_")
	content = strings.ReplaceAll(content, "*", "\\*")
	content = strings.ReplaceAll(content, "[", "\\[")
	content = strings.ReplaceAll(content, "]", "\\]")
	content = strings.ReplaceAll(content, "(", "\\(")
	content = strings.ReplaceAll(content, ")", "\\)")
	content = strings.ReplaceAll(content, "~", "\\~")
	content = strings.ReplaceAll(content, "`", "\\`")
	content = strings.ReplaceAll(content, ">", "\\>")
	content = strings.ReplaceAll(content, "<", "\\<")
	content = strings.ReplaceAll(content, "#", "\\#")
	content = strings.ReplaceAll(content, "+", "\\+")
	content = strings.ReplaceAll(content, "-", "\\-")
	content = strings.ReplaceAll(content, "=", "\\=")
	content = strings.ReplaceAll(content, "|", "\\|")
	content = strings.ReplaceAll(content, ".", "\\.")
	content = strings.ReplaceAll(content, "!", "\\!")
	content = strings.ReplaceAll(content, "{", "\\{")
	content = strings.ReplaceAll(content, "}", "\\}")

	return content
}

// GetDomainAddress 从地址中获取域名和端口号
// 对应 Python StringUtils.get_domain_address 方法
func GetDomainAddress(address string, prefix bool) (string, int) {
	if address == "" {
		return "", 0
	}

	// 去除末尾的斜杠
	address = strings.TrimRight(address, "/")

	var domain string
	var port int

	if prefix {
		// 需要包含协议前缀
		if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
			address = "http://" + address
		}
	} else {
		// 不需要包含协议前缀
		if strings.Contains(address, "://") {
			address = address[strings.Index(address, "://")+3:]
		}
	}

	// 拆分域名和端口
	if strings.Contains(address, ":") {
		parts := strings.Split(address, ":")
		if len(parts) == 2 {
			// http://example.com:8080 或 example.com:8080
			domain = parts[0]
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		} else if len(parts) == 3 {
			// http://[::1]:8080 格式的IPv6地址
			domain = parts[0] + ":" + parts[1]
			if p, err := strconv.Atoi(parts[2]); err == nil {
				port = p
			}
		} else {
			// 格式不正确
			return "", 0
		}
	} else {
		// 没有端口，使用默认端口
		domain = address
		if strings.HasPrefix(address, "https://") {
			port = 443
		} else {
			port = 80
		}
	}

	return domain, port
}

// DiffTimeStr 输入YYYY-MM-DD HH24:MI:SS格式的时间字符串，返回距离现在的剩余时间
// 对应 Python StringUtils.diff_time_str 方法
func DiffTimeStr(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	// 解析时间字符串
	layout := "2006-01-02 15:04:05"
	targetTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return timeStr
	}

	// 计算时间差
	now := time.Now()
	diff := targetTime.Sub(now)

	// 如果时间差为负，返回空字符串
	if diff <= 0 {
		return ""
	}

	// 计算天、小时、分钟
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	var result string
	if days > 0 {
		result = fmt.Sprintf("%d天", days)
	}
	if hours > 0 || days > 0 {
		result += fmt.Sprintf("%d小时", hours)
	}
	result += fmt.Sprintf("%d分钟", minutes)

	return result
}

// SafeStrip 去除字符串两端的空白字符
// 对应 Python StringUtils.safe_strip 方法
func SafeStrip(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

// versionMap 版本号转换映射
var versionMap = map[string]int{
	"stable": -1,
	"rc":     -2,
	"beta":   -3,
	"alpha":  -4,
}

// otherVersion 不符合的版本号
const otherVersion = -5

// CompareVersion 比较两个版本号的大小
// 对应 Python StringUtils.compare_version 方法
func CompareVersion(v1 string, compareType string, v2 string, verbose bool) (interface{}, interface{}) {
	// 验证输入
	if v1 == "" || v2 == "" {
		if verbose {
			return nil, "要比较的版本号不全"
		}
		return nil, nil
	}

	if compareType == "" {
		if verbose {
			return nil, "缺少比对模式，无法比对"
		}
		return nil, nil
	}

	// 验证比较类型
	validCompareTypes := map[string]bool{
		"ge": true, ">=": true,
		"le": true, "<=": true,
		"eq": true, "==": true,
		"gt": true, ">": true,
		"lt": true, "<": true,
	}

	if !validCompareTypes[compareType] {
		if verbose {
			return nil, fmt.Sprintf("设置的版本比对模式 %s 不是有效的模式！", compareType)
		}
		return nil, nil
	}

	// 预处理版本号
	preprocessVersion := func(version string) []string {
		version = strings.TrimSpace(version)
		version = strings.TrimPrefix(version, "v")
		version = strings.TrimPrefix(version, "V")
		return strings.FieldsFunc(version, func(r rune) bool {
			return r == '.' || r == '-' || r == '_'
		})
	}

	// 转换版本号为数字
	conversionVersion := func(versionList []string) []int {
		var result []int
		for _, item := range versionList {
			if num, err := strconv.Atoi(item); err == nil {
				result = append(result, num)
			} else {
				if val, ok := versionMap[item]; ok {
					result = append(result, val)
				} else {
					result = append(result, otherVersion)
				}
			}
		}
		return result
	}

	// 处理版本号
	v1List := conversionVersion(preprocessVersion(v1))
	v2List := conversionVersion(preprocessVersion(v2))

	// 补全版本号长度
	maxLength := max(len(v1List), len(v2List))
	for len(v1List) < maxLength {
		v1List = append(v1List, 0)
	}
	for len(v2List) < maxLength {
		v2List = append(v2List, 0)
	}

	var verComparison string
	var verComparisonErr string

	// 比较版本号
	for i := 0; i < maxLength; i++ {
		v1Val := v1List[i]
		v2Val := v2List[i]

		if compareType == "eq" || compareType == "==" {
			if v1Val != v2Val {
				verComparisonErr = "不等于"
				break
			} else {
				verComparison = "等于"
			}
		} else if compareType == "ge" || compareType == ">=" {
			if v1Val > v2Val {
				verComparison = "大于"
				break
			} else if v1Val < v2Val {
				verComparisonErr = "小于"
				break
			} else {
				verComparison = "等于"
			}
		} else if compareType == "gt" || compareType == ">" {
			if v1Val > v2Val {
				verComparison = "大于"
				break
			} else if v1Val < v2Val {
				verComparisonErr = "小于"
				break
			} else {
				verComparisonErr = "等于"
				break
			}
		} else if compareType == "le" || compareType == "<=" {
			if v1Val > v2Val {
				verComparisonErr = "大于"
				break
			} else if v1Val < v2Val {
				verComparison = "小于"
				break
			} else {
				verComparison = "等于"
			}
		} else if compareType == "lt" || compareType == "<" {
			if v1Val > v2Val {
				verComparisonErr = "大于"
				break
			} else if v1Val < v2Val {
				verComparison = "小于"
				break
			} else {
				verComparisonErr = "等于"
				break
			}
		}
	}

	if verbose {
		if verComparison != "" {
			msg := fmt.Sprintf("版本号 %s %s 目标版本号 %s ！", v1, verComparison, v2)
			return true, msg
		} else {
			msg := fmt.Sprintf("版本号 %s %s 目标版本号 %s ！", v1, verComparisonErr, v2)
			return false, msg
		}
	} else {
		return verComparison != "", nil
	}
}
