package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// FormatParser 格式化解析器
type FormatParser struct {
	format     string
	startEp    any
	endEp      any
	offset     string
	key        string
	part       string
	splitChars string
}

// NewFormatParser 创建格式化解析器
func NewFormatParser(format string, details string, part string, offset string, key string) *FormatParser {
	fp := &FormatParser{
		format:     format,
		key:        key,
		part:       part,
		splitChars: `\.|\s+|\(|\)|\[|]|-|\+|【|】|/|～|;|&|\||#|_|「|」|~`,
	}

	// 处理偏移量
	fp.handleOffset(offset)

	// 处理详情
	fp.handleDetails(details)

	return fp
}

// handleOffset 处理偏移量
func (fp *FormatParser) handleOffset(offset string) {
	if offset == "" {
		fp.offset = "EP"
		return
	}

	if strings.Contains(offset, "EP") {
		fp.offset = offset
		return
	}

	if strings.HasPrefix(offset, "-") || strings.HasPrefix(offset, "+") {
		fp.offset = fmt.Sprintf("EP%s", offset)
	} else {
		fp.offset = fmt.Sprintf("EP+%s", offset)
	}
}

// handleDetails 处理详情
func (fp *FormatParser) handleDetails(details string) {
	if details == "" {
		return
	}

	// 匹配 XXXX-XXXX 格式
	rangeRegex := regexp.MustCompile(`^\d{1,4}-\d{1,4}$`)
	if rangeRegex.MatchString(details) {
		fp.startEp = details
		fp.endEp = details
		return
	}

	// 处理逗号分隔的格式
	parts := strings.Split(details, ",")
	if len(parts) > 1 {
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		if start > end {
			fp.startEp = start
			fp.endEp = start
		} else {
			fp.startEp = start
			fp.endEp = end
		}
	} else {
		// 处理单个数字
		val, _ := strconv.Atoi(parts[0])
		fp.startEp = val
		fp.endEp = val
	}
}

// Match 匹配文件名
func (fp *FormatParser) Match(file string) bool {
	if fp.format == "" {
		return true
	}

	// 只使用第一个返回值，忽略第二个
	start, _ := fp.handleSingle(file)
	if start == 0 {
		return false
	}

	if fp.startEp == nil {
		return true
	}

	// 检查集数是否在范围内
	if start >= fp.getStartEp() && start <= fp.getEndEp() {
		return true
	}

	return false
}

// SplitEpisode 拆分剧集
func (fp *FormatParser) SplitEpisode(fileName string, beginEp, endEp int) (int, int, string) {
	// 指定了具体集数
	if fp.startEp != nil {
		if fp.startEp == fp.endEp {
			// 格式为 X-X 或 X
			if str, ok := fp.startEp.(string); ok {
				// 格式为 X-X
				parts := strings.Split(str, "-")
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])

				startEp := fp.applyOffset(start)
				endEp := fp.applyOffset(end)

				if start == end {
					return startEp, 0, fp.part
				}
				return startEp, endEp, fp.part
			} else {
				// 格式为 X
				start := fp.startEp.(int)
				startEp := fp.applyOffset(start)
				return startEp, 0, fp.part
			}
		} else if fp.format == "" {
			// 格式为 X,X
			start := fp.startEp.(int)
			end := fp.endEp.(int)
			startEp := fp.applyOffset(start)
			endEp := fp.applyOffset(end)
			return startEp, endEp, fp.part
		}
	}

	if fp.format == "" {
		// 未指定格式，仅处理偏移
		startEp := fp.applyOffset(beginEp)
		endEp := fp.applyOffset(endEp)
		return startEp, endEp, fp.part
	} else {
		// 有格式
		start, end := fp.handleSingle(fileName)
		startEp := fp.applyOffset(start)
		endEp := fp.applyOffset(end)
		return startEp, endEp, fp.part
	}
}

// handleSingle 处理单集
func (fp *FormatParser) handleSingle(file string) (int, int) {
	if fp.format == "" {
		return 0, 0
	}

	// 使用正则表达式匹配格式
	re, err := regexp.Compile(fp.format)
	if err != nil {
		logger.GetLogger().Debug("格式正则编译失败", zap.Error(err), zap.String("format", fp.format))
		return 0, 0
	}

	// 约定：第一个捕获组为 episodes 字段
	matches := re.FindStringSubmatch(file)
	if len(matches) < 2 {
		logger.GetLogger().Debug("未匹配到集数字段", zap.String("file", file), zap.String("format", fp.format))
		return 0, 0
	}
	episodes := matches[1]

	// 验证集数格式
	epRegex := regexp.MustCompile(`^(EP)?(\d{1,4})(-(EP)?(\d{1,4}))?$`)
	if !epRegex.MatchString(episodes) {
		logger.GetLogger().Debug("集数格式不匹配", zap.String("episodes", episodes))
		return 0, 0
	}

	// 拆分集数
	splitRegex := regexp.MustCompile(fp.splitChars)
	epParts := splitRegex.Split(episodes, -1)

	// 过滤出包含数字的部分
	var validParts []string
	for _, part := range epParts {
		if numRegex := regexp.MustCompile(`[a-zA-Z]*\d{1,4}`); numRegex.MatchString(part) {
			validParts = append(validParts, part)
		}
	}

	if len(validParts) == 0 {
		return 0, 0
	}

	// 提取数字
	numRegex := regexp.MustCompile(`\d+`)
	startStr := numRegex.FindString(validParts[0])
	start, _ := strconv.Atoi(startStr)

	if len(validParts) == 1 {
		return start, 0
	}

	endStr := numRegex.FindString(validParts[1])
	end, _ := strconv.Atoi(endStr)

	return start, end
}

// applyOffset 应用偏移
func (fp *FormatParser) applyOffset(ep int) int {
	if fp.offset == "EP" {
		return ep
	}

	// 解析偏移格式
	offsetRegex := regexp.MustCompile(`EP([+-]?\d+)`)
	matches := offsetRegex.FindStringSubmatch(fp.offset)
	if len(matches) != 2 {
		return ep
	}

	offset, _ := strconv.Atoi(matches[1])
	return ep + offset
}

// getStartEp 获取开始集数
func (fp *FormatParser) getStartEp() int {
	if fp.startEp == nil {
		return 0
	}

	if str, ok := fp.startEp.(string); ok {
		parts := strings.Split(str, "-")
		start, _ := strconv.Atoi(parts[0])
		return start
	}

	return fp.startEp.(int)
}

// getEndEp 获取结束集数
func (fp *FormatParser) getEndEp() int {
	if fp.endEp == nil {
		return 0
	}

	if str, ok := fp.endEp.(string); ok {
		parts := strings.Split(str, "-")
		end, _ := strconv.Atoi(parts[1])
		return end
	}

	return fp.endEp.(int)
}
