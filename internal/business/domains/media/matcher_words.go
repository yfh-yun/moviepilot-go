package media

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
)

// IdentifierConfigProvider 识别词配置提供者接口
type IdentifierConfigProvider interface {
	GetCustomIdentifiers() ([]string, error)
}

// WordsMatcher 识别词匹配器
type WordsMatcher struct {
	provider IdentifierConfigProvider
	once     sync.Once
}

var (
	// wordsMatcherInstance 单例实例
	wordsMatcherInstance *WordsMatcher
	// wordsMatcherMutex 单例初始化互斥锁
	wordsMatcherMutex sync.Mutex
)

// NewWordsMatcher 创建新的WordsMatcher实例
func NewWordsMatcher(provider IdentifierConfigProvider) *WordsMatcher {
	// 双重检查锁定实现单例模式
	if wordsMatcherInstance == nil {
		wordsMatcherMutex.Lock()
		defer wordsMatcherMutex.Unlock()
		if wordsMatcherInstance == nil {
			wordsMatcherInstance = &WordsMatcher{
				provider: provider,
			}
		}
	}
	return wordsMatcherInstance
}

// Prepare 预处理标题，支持三种格式
// 1：屏蔽词
// 2：被替换词 => 替换词
// 3：前定位词 <> 后定位词 >> 偏移量（EP）
func (wm *WordsMatcher) Prepare(title string, customWords []string) (string, []string) {
	if title == "" {
		return title, nil
	}

	var appliedWords []string
	processedTitle := title

	// 读取自定义识别词
	words := customWords
	if wm.provider != nil {
		configWords, err := wm.provider.GetCustomIdentifiers()
		if err == nil {
			words = append(words, configWords...)
		}
	}

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}

		// 复杂替换规则：替换词 => 被替换词 && 集偏移前字段 <> 集偏移后字段 >> 集偏移
		if strings.Contains(word, " => ") && strings.Contains(word, " && ") && strings.Contains(word, " >> ") && strings.Contains(word, " <>") {
			// 提取各部分
			// 注意：<> 不是正则特殊字符，不需要转义
			thcRegex := regexp.MustCompile(`(.*?)\s*=>`)
			thcMatches := thcRegex.FindStringSubmatch(word)
			if len(thcMatches) < 2 {
				continue
			}
			thc := strings.TrimSpace(thcMatches[1])

			bthcRegex := regexp.MustCompile(`=>\s*(.*?)\s*&&`)
			bthcMatches := bthcRegex.FindStringSubmatch(word)
			if len(bthcMatches) < 2 {
				continue
			}
			bthc := strings.TrimSpace(bthcMatches[1])

			pyqRegex := regexp.MustCompile(`&&\s*(.*?)\s*<>`)
			pyqMatches := pyqRegex.FindStringSubmatch(word)
			if len(pyqMatches) < 2 {
				continue
			}
			pyq := strings.TrimSpace(pyqMatches[1])

			pyhRegex := regexp.MustCompile(`<>(.*?)\s*>>`)
			pyhMatches := pyhRegex.FindStringSubmatch(word)
			if len(pyhMatches) < 2 {
				continue
			}
			pyh := strings.TrimSpace(pyhMatches[1])

			offsetsRegex := regexp.MustCompile(`>>\s*(.*?)$`)
			offsetsMatches := offsetsRegex.FindStringSubmatch(word)
			if len(offsetsMatches) < 2 {
				continue
			}
			offsets := strings.TrimSpace(offsetsMatches[1])

			// 替换词
			newTitle, _, state := wm.replaceRegex(processedTitle, thc, bthc)
			if state {
				processedTitle = newTitle
				// 替换词成功再进行集偏移
				newTitle, _, offsetState := wm.episodeOffset(processedTitle, pyq, pyh, offsets)
				if offsetState {
					processedTitle = newTitle
				}
				appliedWords = append(appliedWords, word)
			}
			// 简单替换规则：A => B
		} else if strings.Contains(word, " => ") {
			stringsParts := strings.SplitN(word, " => ", 2)
			if len(stringsParts) == 2 {
				oldStr := strings.TrimSpace(stringsParts[0])
				newStr := strings.TrimSpace(stringsParts[1])
				newTitle, _, state := wm.replaceRegex(processedTitle, oldStr, newStr)
				if state {
					processedTitle = newTitle
					appliedWords = append(appliedWords, word)
				}
			}
			// 集偏移规则：前定位词 <> 后定位词 >> 偏移量
		} else if strings.Contains(word, " >> ") && strings.Contains(word, " <>") {
			// 提取各部分
			parts := strings.SplitN(word, " <>", 2)
			if len(parts) < 2 {
				continue
			}
			front := strings.TrimSpace(parts[0])
			backParts := strings.SplitN(parts[1], " >> ", 2)
			if len(backParts) < 2 {
				continue
			}
			back := strings.TrimSpace(backParts[0])
			offset := strings.TrimSpace(backParts[1])

			newTitle, _, state := wm.episodeOffset(processedTitle, front, back, offset)
			if state {
				processedTitle = newTitle
				appliedWords = append(appliedWords, word)
			}
			// 屏蔽词规则
		} else {
			newTitle, _, state := wm.replaceRegex(processedTitle, word, "")
			if state {
				processedTitle = newTitle
				appliedWords = append(appliedWords, word)
			}
		}
	}

	return processedTitle, appliedWords
}

// replaceRegex 正则替换
func (wm *WordsMatcher) replaceRegex(title, replaced, replace string) (string, string, bool) {
	if title == "" || replaced == "" {
		return title, "", false
	}

	// 使用regexp2库支持更复杂的正则表达式
	re, err := regexp2.Compile(replaced, regexp2.None)
	if err != nil {
		return title, fmt.Sprintf("正则编译失败: %v", err), false
	}

	matches, err := re.FindStringMatch(title)
	if err != nil || matches == nil {
		return title, "", false
	}

	// 执行替换
	newTitle, err := re.Replace(title, replace, -1, -1)
	if err != nil {
		return title, fmt.Sprintf("正则替换失败: %v", err), false
	}

	return newTitle, "", true
}

// episodeOffset 集数偏移处理
func (wm *WordsMatcher) episodeOffset(title, front, back, offset string) (string, string, bool) {
	if title == "" {
		return title, "", false
	}

	// 构建正则表达式，匹配front后面跟着的数字（支持中文和阿拉伯数字）
	pattern := fmt.Sprintf(`(%s)([0-9一二三四五六七八九十]+)`, regexp.QuoteMeta(front))
	re, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return title, fmt.Sprintf("正则编译失败: %v", err), false
	}

	// 查找所有匹配的集数
	var matches []*regexp2.Match
	match, err := re.FindStringMatch(title)
	for err == nil && match != nil {
		matches = append(matches, match)
		match, err = re.FindNextMatch(match)
	}

	if len(matches) == 0 {
		return title, "", false
	}

	// 执行替换
	newTitle := title
	for _, match := range matches {
		if len(match.Groups()) < 3 {
			continue
		}

		fullMatch := match.String()
		prefix := match.Groups()[1].String()
		epStr := match.Groups()[2].String()

		// 转换为阿拉伯数字（支持中文和阿拉伯数字）
		anNum, err := wm.chineseToArabic(epStr)
		if err != nil {
			continue
		}

		// 计算偏移后的集数
		calcOffset := strings.Replace(offset, "EP", fmt.Sprintf("%d", anNum), -1)
		offsetNum, err := wm.calculateOffset(calcOffset)
		if err != nil {
			continue
		}

		// 根据原格式转换回相应的数字格式
		var offsetStr string
		if isChineseNumber(epStr) {
			// 中文数字转换回中文
			offsetStr = wm.arabicToChinese(offsetNum)
		} else {
			// 阿拉伯数字，保留前导零
			leadingZeroRegex := regexp.MustCompile(`^0+`)
			leadingZero := leadingZeroRegex.FindString(epStr)
			offsetStr = fmt.Sprintf("%s%d", leadingZero, offsetNum)
		}

		// 构建新的匹配字符串
		newMatch := fmt.Sprintf("%s%s", prefix, offsetStr)

		// 替换整个匹配
		newTitle = strings.Replace(newTitle, fullMatch, newMatch, 1)
	}

	return newTitle, "", true
}

// calculateOffset 简单的表达式计算，仅支持基本的加减乘除
func (wm *WordsMatcher) calculateOffset(expr string) (int, error) {
	// 简单实现，仅支持基本的加减乘除
	// 更复杂的表达式可以考虑使用第三方库

	// 移除空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 支持的运算符：+、-、*、/
	operators := []string{"+", "-", "*", "/"}

	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			if len(parts) != 2 {
				continue
			}

			a, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			b, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			switch op {
			case "+":
				return a + b, nil
			case "-":
				return a - b, nil
			case "*":
				return a * b, nil
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("除以零错误")
				}
				return a / b, nil
			}
		}
	}

	// 如果没有运算符，直接返回数字
	return strconv.Atoi(expr)
}

// isChineseNumber 判断是否为中文数字
func isChineseNumber(s string) bool {
	if s == "" {
		return false
	}
	chineseNums := "一二三四五六七八九十"
	for _, r := range s {
		if !strings.ContainsRune(chineseNums, r) {
			return false
		}
	}
	return true
}

// chineseToArabic 将中文数字转换为阿拉伯数字
func (wm *WordsMatcher) chineseToArabic(chineseNum string) (int, error) {
	// 简单的中文数字转换，只处理一到十
	chineseToNum := map[string]int{
		"一": 1,
		"二": 2,
		"三": 3,
		"四": 4,
		"五": 5,
		"六": 6,
		"七": 7,
		"八": 8,
		"九": 9,
		"十": 10,
	}

	// 尝试直接转换
	if num, ok := chineseToNum[chineseNum]; ok {
		return num, nil
	}

	// 如果是阿拉伯数字字符串，直接转换
	return strconv.Atoi(chineseNum)
}

// arabicToChinese 将阿拉伯数字转换为中文数字
func (wm *WordsMatcher) arabicToChinese(num int) string {
	// 简单的阿拉伯数字转换，只处理一到十
	numToChinese := map[int]string{
		1:  "一",
		2:  "二",
		3:  "三",
		4:  "四",
		5:  "五",
		6:  "六",
		7:  "七",
		8:  "八",
		9:  "九",
		10: "十",
	}

	// 尝试直接转换
	if chinese, ok := numToChinese[num]; ok {
		return chinese
	}

	// 超出范围的数字直接返回字符串
	return fmt.Sprintf("%d", num)
}

// SimpleWordsMatcher 简单的WordsMatcher实现，不依赖外部配置
type SimpleWordsMatcher struct{}

// NewSimpleWordsMatcher 创建简单的WordsMatcher实例
func NewSimpleWordsMatcher() *WordsMatcher {
	return &WordsMatcher{
		provider: &SimpleWordsMatcher{},
	}
}

// GetCustomIdentifiers 获取自定义识别词
func (swm *SimpleWordsMatcher) GetCustomIdentifiers() ([]string, error) {
	// 简单实现，返回空列表
	return []string{}, nil
}
