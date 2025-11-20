// Package meta 元数据处理模块
package meta

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/repositories"
	
)

// WordsMatcher 词汇匹配器
type WordsMatcher struct {
	systemConfigRepo repositories.SystemConfigRepository
}

// NewWordsMatcher 创建词汇匹配器
func NewWordsMatcher(systemConfigRepo repositories.SystemConfigRepository) *WordsMatcher {
	return &WordsMatcher{
		systemConfigRepo: systemConfigRepo,
	}
}

// PrepareResult 预处理结果
type PrepareResult struct {
	Title        string   `json:"title"`         // 处理后的标题
	AppliedWords []string `json:"applied_words"` // 应用的词汇
}

// WordType 词汇类型
type WordType int

const (
	WordTypeBlocked       WordType = iota // 屏蔽词
	WordTypeReplaced                      // 替换词
	WordTypeEpisodeOffset                 // 集数偏移
	WordTypeCombined                      // 组合操作
)

// CustomWord 自定义词汇
type CustomWord struct {
	Type      WordType `json:"type"`
	Original  string   `json:"original"`
	Target    string   `json:"target"`
	FrontWord string   `json:"front_word"`
	BackWord  string   `json:"back_word"`
	Offset    string   `json:"offset"`
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
}

// Prepare 预处理标题，支持三种格式
// 1：屏蔽词
// 2：被替换词 => 替换词
// 3：前定位词 <> 后定位词 >> 偏移量（EP）
func (w *WordsMatcher) Prepare(ctx context.Context, title string, customWords []string) (*PrepareResult, error) {
	if title == "" {
		return &PrepareResult{Title: title, AppliedWords: []string{}}, nil
	}

	appliedWords := []string{}
	words, err := w.getWords(ctx, customWords)
	if err != nil {
		return &PrepareResult{Title: title, AppliedWords: appliedWords},
			fmt.Errorf("获取自定义词汇失败: %w", err)
	}

	processedTitle := title

	for _, wordConfig := range words {
		if wordConfig == "" || strings.HasPrefix(wordConfig, "#") {
			continue
		}

		word, err := w.parseWord(wordConfig)
		if err != nil {
			logger.Warn("解析自定义词汇失败",
				zap.String("词汇", wordConfig),
				zap.Error(err),
				zap.String("标题", title))
			continue
		}

		var newTitle string
		var success bool
		var message string

		switch word.Type {
		case WordTypeCombined:
			// 组合操作：先替换再偏移
			newTitle, message, success = w.executeCombined(processedTitle, word)
		case WordTypeReplaced:
			// 替换词
			newTitle, message, success = w.executeReplacement(processedTitle, word)
		case WordTypeEpisodeOffset:
			// 集数偏移
			newTitle, message, success = w.executeEpisodeOffset(processedTitle, word)
		case WordTypeBlocked:
			// 屏蔽词
			newTitle, message, success = w.executeReplacement(processedTitle, CustomWord{
				Type:     WordTypeReplaced,
				Original: word.Original,
				Target:   "",
			})
		}

		if success {
			processedTitle = newTitle
			appliedWords = append(appliedWords, wordConfig)
			logger.Debug("应用自定义词汇成功",
				zap.String("词汇", wordConfig),
				zap.String("原标题", title),
				zap.String("新标题", processedTitle))
		} else if message != "" {
			logger.Warn("应用自定义词汇失败",
				zap.String("词汇", wordConfig),
				zap.String("原因", message),
				zap.String("标题", title))
		}
	}

	return &PrepareResult{
		Title:        processedTitle,
		AppliedWords: appliedWords,
	}, nil
}

// parseWord 解析词汇配置
func (w *WordsMatcher) parseWord(wordConfig string) (*CustomWord, error) {
	word := &CustomWord{}

	// 检查组合操作：被替换词 => 替换词 && 前定位词 <> 后定位词 >> 偏移量
	if strings.Contains(wordConfig, " => ") &&
		strings.Contains(wordConfig, " && ") &&
		strings.Contains(wordConfig, " >> ") &&
		strings.Contains(wordConfig, " <> ") {

		// 解析替换词
		replacedPart := strings.Split(strings.Split(wordConfig, " && ")[0], " => ")
		if len(replacedPart) < 2 {
			return nil, fmt.Errorf("无效的替换词格式")
		}
		word.Type = WordTypeCombined
		word.Target = strings.TrimSpace(replacedPart[0])
		word.Original = strings.TrimSpace(replacedPart[1])

		// 解析集偏移
		offsetPart := strings.Split(strings.Split(wordConfig, " && ")[1], " >> ")
		if len(offsetPart) < 2 {
			return nil, fmt.Errorf("无效的偏移量格式")
		}

		frontBackPart := strings.Split(offsetPart[0], " <> ")
		if len(frontBackPart) < 2 {
			return nil, fmt.Errorf("无效的前后定位词格式")
		}
		word.FrontWord = strings.TrimSpace(frontBackPart[0])
		word.BackWord = strings.TrimSpace(frontBackPart[1])
		word.Offset = strings.TrimSpace(offsetPart[1])

		return word, nil
	}

	// 检查替换词：被替换词 => 替换词
	if strings.Contains(wordConfig, " => ") {
		parts := strings.Split(wordConfig, " => ")
		if len(parts) < 2 {
			return nil, fmt.Errorf("无效的替换词格式")
		}
		word.Type = WordTypeReplaced
		word.Original = strings.TrimSpace(parts[0])
		word.Target = strings.TrimSpace(parts[1])
		return word, nil
	}

	// 检查集数偏移：前定位词 <> 后定位词 >> 偏移量
	if strings.Contains(wordConfig, " >> ") && strings.Contains(wordConfig, " <> ") {
		frontBackPart := strings.Split(strings.Split(wordConfig, " >> ")[0], " <> ")
		if len(frontBackPart) < 2 {
			return nil, fmt.Errorf("无效的前后定位词格式")
		}
		word.Type = WordTypeEpisodeOffset
		word.FrontWord = strings.TrimSpace(frontBackPart[0])
		word.BackWord = strings.TrimSpace(frontBackPart[1])
		word.Offset = strings.TrimSpace(strings.Split(wordConfig, " >> ")[1])
		return word, nil
	}

	// 默认为屏蔽词
	word.Type = WordTypeBlocked
	word.Original = strings.TrimSpace(wordConfig)
	return word, nil
}

// executeCombined 执行组合操作
func (w *WordsMatcher) executeCombined(title string, word *CustomWord) (string, string, bool) {
	// 先执行替换
	replacedTitle, message, success := w.executeReplacement(title, &CustomWord{
		Type:     WordTypeReplaced,
		Original: word.Original,
		Target:   word.Target,
	})
	if !success {
		return title, "替换词处理失败: " + message, false
	}

	// 再执行集偏移
	finalTitle, offsetMessage, offsetSuccess := w.executeEpisodeOffset(replacedTitle, &CustomWord{
		Type:      WordTypeEpisodeOffset,
		FrontWord: word.FrontWord,
		BackWord:  word.BackWord,
		Offset:    word.Offset,
	})

	if !offsetSuccess {
		return title, "集数偏移处理失败: " + offsetMessage, false
	}

	return finalTitle, "", true
}

// executeReplacement 执行替换操作
func (w *WordsMatcher) executeReplacement(title string, word *CustomWord) (string, string, bool) {
	if word.Original == "" {
		return title, "原始词汇为空", false
	}

	pattern := regexp.QuoteMeta(word.Original)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return title, "正则表达式编译失败: " + err.Error(), false
	}

	if !re.MatchString(title) {
		return title, "标题中未找到匹配词汇", false
	}

	newTitle := re.ReplaceAllString(title, word.Target)
	return newTitle, "", true
}

// executeEpisodeOffset 执行集数偏移
func (w *WordsMatcher) executeEpisodeOffset(title string, word *CustomWord) (string, string, bool) {
	if word.FrontWord == "" || word.BackWord == "" || word.Offset == "" {
		return title, "集偏移参数不完整", false
	}

	// 检查前后定位词是否存在
	frontPattern := regexp.QuoteMeta(word.FrontWord)
	backPattern := regexp.QuoteMeta(word.BackWord)

	frontRe, err := regexp.Compile(frontPattern)
	if err != nil {
		return title, "前定位词正则编译失败: " + err.Error(), false
	}

	backRe, err := regexp.Compile(backPattern)
	if err != nil {
		return title, "后定位词正则编译失败: " + err.Error(), false
	}

	if !frontRe.MatchString(title) {
		return title, "标题中未找到前定位词", false
	}

	if word.BackWord != "" && !backRe.MatchString(title) {
		return title, "标题中未找到后定位词", false
	}

	// 构建集数提取正则
	episodePattern := fmt.Sprintf(`(?<=%s.*?)[0-9一二三四五六七八九十]+(?=.*?%s)`,
		regexp.QuoteMeta(word.FrontWord), regexp.QuoteMeta(word.BackWord))

	episodeRe, err := regexp.Compile(episodePattern)
	if err != nil {
		return title, "集数提取正则编译失败: " + err.Error(), false
	}

	episodeMatches := episodeRe.FindAllString(title, -1)
	if len(episodeMatches) == 0 {
		return title, "未找到匹配的集数", false
	}

	// 处理每个集数偏移
	var episodeNumbers []episodeNumber
	var offsetOrder bool // false=向后偏移，true=向前偏移

	for _, match := range episodeMatches {
		num, err := w.parseEpisodeNumber(match)
		if err != nil {
			return title, fmt.Sprintf("解析集数失败: %v", err), false
		}

		// 计算偏移后的集数
		offsetNum, err := w.calculateOffset(num, word.Offset)
		if err != nil {
			return title, fmt.Sprintf("计算偏移失败: %v", err), false
		}

		// 确定偏移方向
		if num.Value > offsetNum.Value {
			offsetOrder = true // 向前偏移
		}

		// 格式化偏移后的集数
		offsetStr, err := w.formatEpisodeNumber(offsetNum, match)
		if err != nil {
			return title, fmt.Sprintf("格式化集数失败: %v", err), false
		}

		episodeNumbers = append(episodeNumbers, episodeNumber{
			Original:      match,
			Offset:        offsetStr,
			OriginalValue: num.Value,
			OffsetValue:   offsetNum.Value,
		})
	}

	// 应用偏移
	newTitle := title

	// 根据偏移方向确定处理顺序
	if offsetOrder {
		// 向前偏移，按集数升序处理
		w.sortEpisodeNumbers(episodeNumbers, true)
	} else {
		// 向后偏移，按集数降序处理
		w.sortEpisodeNumbers(episodeNumbers, false)
	}

	// 应用替换
	for _, ep := range episodeNumbers {
		replacePattern := fmt.Sprintf(`(?<=%s.*?)%s(?=.*?%s)`,
			regexp.QuoteMeta(word.FrontWord),
			regexp.QuoteMeta(ep.Original),
			regexp.QuoteMeta(word.BackWord))

		replaceRe, err := regexp.Compile(replacePattern)
		if err != nil {
			continue
		}

		newTitle = replaceRe.ReplaceAllString(newTitle, ep.Offset)
	}

	return newTitle, "", true
}

// episodeNumber 集数信息
type episodeNumber struct {
	Original      string `json:"original"`
	Offset        string `json:"offset"`
	OriginalValue int    `json:"original_value"`
	OffsetValue   int    `json:"offset_value"`
}

// parseEpisodeNumber 解析集数
func (w *WordsMatcher) parseEpisodeNumber(numStr string) (*episodeNumber, error) {
	// 尝试直接解析为数字
	if num, err := strconv.Atoi(numStr); err == nil {
		return &episodeNumber{Value: num}, nil
	}

	// 解析中文数字
	num, err := w.chineseNumberToInt(numStr)
	if err != nil {
		return nil, err
	}

	return &episodeNumber{Value: num}, nil
}

// calculateOffset 计算偏移量
func (w *WordsMatcher) calculateOffset(num *episodeNumber, offsetStr string) (*episodeNumber, error) {
	// 替换EP为实际集数
	offsetExpression := strings.ReplaceAll(offsetStr, "EP", strconv.Itoa(num.Value))

	// 计算表达式
	result, err := w.evaluateExpression(offsetExpression)
	if err != nil {
		return nil, err
	}

	return &episodeNumber{Value: result}, nil
}

// formatEpisodeNumber 格式化集数
func (w *WordsMatcher) formatEpisodeNumber(num *episodeNumber, original string) (string, error) {
	// 如果原始是中文数字，转换为中文数字
	if !w.isDigitString(original) {
		return w.intToChineseNumber(num.Value), nil
	}

	// 如果原始是阿拉伯数字，保持前导零
	prefixZeros := w.extractPrefixZeros(original)
	if prefixZeros > 0 {
		return fmt.Sprintf("%s%d", strings.Repeat("0", prefixZeros), num.Value), nil
	}

	return strconv.Itoa(num.Value), nil
}

// sortEpisodeNumbers 排序集数
func (w *WordsMatcher) sortEpisodeNumbers(numbers []episodeNumber, ascending bool) {
	if ascending {
		// 升序
		for i := 0; i < len(numbers)-1; i++ {
			for j := i + 1; j < len(numbers); j++ {
				if numbers[i].OffsetValue > numbers[j].OffsetValue {
					numbers[i], numbers[j] = numbers[j], numbers[i]
				}
			}
		}
	} else {
		// 降序
		for i := 0; i < len(numbers)-1; i++ {
			for j := i + 1; j < len(numbers); j++ {
				if numbers[i].OffsetValue < numbers[j].OffsetValue {
					numbers[i], numbers[j] = numbers[j], numbers[i]
				}
			}
		}
	}
}

// chineseNumberToInt 中文数字转整数
func (w *WordsMatcher) chineseNumberToInt(chineseStr string) (int, error) {
	digitMap := map[rune]int{
		'零': 0, '一': 1, '二': 2, '三': 3, '四': 4,
		'五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
		'十': 10, '百': 100, '千': 1000, '万': 10000,
	}

	result := 0
	temp := 0

	for _, char := range chineseStr {
		if digit, exists := digitMap[char]; exists {
			if digit >= 10 {
				// 十、百、千、万等位数单位
				if temp == 0 {
					temp = 1
				}
				temp *= digit
				if digit >= 10000 {
					result += temp
					temp = 0
				}
			} else {
				// 数字
				temp += digit
			}
		}
	}

	result += temp
	return result, nil
}

// intToChineseNumber 整数转中文数字
func (w *WordsMatcher) intToChineseNumber(num int) string {
	if num == 0 {
		return "零"
	}

	digitMap := map[int]string{
		0: "零", 1: "一", 2: "二", 3: "三", 4: "四",
		5: "五", 6: "六", 7: "七", 8: "八", 9: "九",
	}

	unitMap := map[int]string{
		1: "十", 2: "百", 3: "千", 4: "万",
	}

	if num <= 10 {
		return digitMap[num]
	}

	var result string
	digits := []int{}
	temp := num

	for temp > 0 {
		digits = append([]int{temp % 10}, digits...)
		temp /= 10
	}

	for i, digit := range digits {
		if digit == 0 {
			continue
		}

		pos := len(digits) - i
		if pos > 1 {
			unit, exists := unitMap[pos]
			if exists {
				result += digitMap[digit] + unit
			}
		} else {
			result += digitMap[digit]
		}
	}

	return result
}

// evaluateExpression 计算数学表达式
func (w *WordsMatcher) evaluateExpression(expr string) (int, error) {
	// 简单的表达式计算器，支持基本运算
	// 这里使用Go的eval包或自己实现解析
	// 为了简化，这里只支持加减乘除

	// 移除空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 尝试直接解析为数字
	if num, err := strconv.Atoi(expr); err == nil {
		return num, nil
	}

	// 简单的加减法处理
	// 这里可以使用更复杂的表达式解析库
	result := 0
	currentNum := 0
	currentOp := '+'

	expr = "+" + expr // 确保第一个数字能被处理

	for i := 0; i < len(expr); {
		if expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/' {
			if currentOp == '+' {
				result += currentNum
			} else if currentOp == '-' {
				result -= currentNum
			} else if currentOp == '*' {
				result *= currentNum
			} else if currentOp == '/' {
				if currentNum == 0 {
					return 0, fmt.Errorf("除零错误")
				}
				result /= currentNum
			}
			currentOp = rune(expr[i])
			currentNum = 0
			i++
		} else if unicode.IsDigit(rune(expr[i])) {
			start := i
			for i < len(expr) && unicode.IsDigit(rune(expr[i])) {
				i++
			}
			num, _ := strconv.Atoi(expr[start:i])
			currentNum = num
		} else {
			i++
		}
	}

	// 处理最后一个数字
	if currentOp == '+' {
		result += currentNum
	} else if currentOp == '-' {
		result -= currentNum
	} else if currentOp == '*' {
		result *= currentNum
	} else if currentOp == '/' {
		if currentNum == 0 {
			return 0, fmt.Errorf("除零错误")
		}
		result /= currentNum
	}

	return result, nil
}

// isDigitString 检查字符串是否只包含数字
func (w *WordsMatcher) isDigitString(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

// extractPrefixZeros 提取前导零数量
func (w *WordsMatcher) extractPrefixZeros(s string) int {
	count := 0
	for _, char := range s {
		if char == '0' {
			count++
		} else {
			break
		}
	}
	return count
}

// getWords 获取自定义词汇列表
func (w *WordsMatcher) getWords(ctx context.Context, customWords []string) ([]string, error) {
	if customWords != nil {
		return customWords, nil
	}

	// 从系统配置获取
	wordsConfig, err := w.systemConfigRepo.Get(ctx, types.SystemConfigKeyCustomIdentifiers)
	if err != nil {
		return nil, err
	}

	if wordsConfig == nil || wordsConfig.Value == "" {
		return []string{}, nil
	}

	// 假设配置以数组格式存储，这里简化处理
	var words []string
	if strings.Contains(wordsConfig.Value, "[") {
		// JSON数组格式
		// 需要JSON解析，这里简化处理
		words = strings.Split(strings.Trim(wordsConfig.Value, "[]"), ",")
	} else {
		// 换行或逗号分隔
		words = strings.FieldsFunc(wordsConfig.Value, func(r rune) bool {
			return r == '\n' || r == ','
		})
	}

	// 清理空项和引号
	var cleanWords []string
	for _, word := range words {
		word = strings.TrimSpace(strings.Trim(word, "\""))
		if word != "" {
			cleanWords = append(cleanWords, word)
		}
	}

	return cleanWords, nil
}

// ValidateWord 验证词汇配置格式
func (w *WordsMatcher) ValidateWord(wordConfig string) error {
	if wordConfig == "" {
		return nil
	}

	_, err := w.parseWord(wordConfig)
	return err
}

// ValidateWords 批量验证词汇配置
func (w *WordsMatcher) ValidateWords(words []string) error {
	for i, word := range words {
		if err := w.ValidateWord(word); err != nil {
			return fmt.Errorf("词汇 %d 格式无效: %w", i+1, err)
		}
	}
	return nil
}
