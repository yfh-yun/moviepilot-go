package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	
	"moviepilot-go/pkg/models"
)

// FormatParser 格式化解析器
type FormatParser struct {
	format    string
	startEp   interface{}
	endEp     interface{}
	part      *string
	offset    string
	key       string
	splitChars string
}

// NewFormatParser 创建一个新的格式化解析器实�?/*
 * :params eformat: 格式化字符串
 * :params details: 格式化详�? * :params part: 分集
 * :params offset: 偏移�?-10/EP*2
 * :params key: EP关键�? */
func NewFormatParser(eformat string, details *string, part *string, offset *string, key *string) *FormatParser {
	parser := &FormatParser{
		format:     eformat,
		startEp:    nil,
		endEp:      nil,
		splitChars: `\.|\s+|\(|\)|\[|\]|-|\+|【|】|/|～|;|&|\||#|_|「|」|~`,
	}
	
	// 设置key
	if key != nil {
		parser.key = *key
	} else {
		parser.key = "ep"
	}
	
	// 设置offset
	if offset == nil || *offset == "" {
		parser.offset = "EP"
	} else if strings.Contains(*offset, "EP") {
		parser.offset = *offset
	} else {
		if strings.HasPrefix(*offset, "-") || strings.HasPrefix(*offset, "+") {
			parser.offset = fmt.Sprintf("EP%s", *offset)
		} else {
			parser.offset = fmt.Sprintf("EP+%s", *offset)
		}
	}
	
	// 设置part
	if part != nil {
		parser.part = part
	}
	
	// 处理details
	if details != nil && *details != "" {
		// 检查是否匹�?\d{1,4}-\d{1,4} 格式
		re := regexp.MustCompile(`\d{1,4}-\d{1,4}`)
		if re.MatchString(*details) {
			parser.startEp = *details
			parser.endEp = *details
		} else {
			tmp := strings.Split(*details, ",")
			if len(tmp) > 1 {
				start, err1 := strconv.Atoi(tmp[0])
				end, err2 := strconv.Atoi(tmp[1])
				if err1 == nil && err2 == nil {
					parser.startEp = start
					if start > end {
						parser.endEp = start
					} else {
						parser.endEp = end
					}
				}
			} else {
				ep, err := strconv.Atoi(tmp[0])
				if err == nil {
					parser.startEp = ep
					parser.endEp = ep
				}
			}
		}
	}
	
	return parser
}

// Format 获取格式化字符串
func (fp *FormatParser) Format() string {
	return fp.format
}

// StartEp 获取开始集�?func (fp *FormatParser) StartEp() interface{} {
	return fp.startEp
}

// EndEp 获取结束集数
func (fp *FormatParser) EndEp() interface{} {
	return fp.endEp
}

// Part 获取分集信息
func (fp *FormatParser) Part() *string {
	return fp.part
}

// Offset 获取偏移�?func (fp *FormatParser) Offset() string {
	return fp.offset
}

// Match 检查文件是否匹配格�?func (fp *FormatParser) Match(file string) bool {
	if fp.format == "" {
		return true
	}
	
	s, e := fp.handleSingle(file)
	if s == nil {
		return false
	}
	
	if fp.startEp == nil {
		return true
	}
	
	// 检查是否在指定范围�?	if startEpInt, ok := fp.startEp.(int); ok {
		if endEpInt, ok := fp.endEp.(int); ok {
			sInt := *s
			return startEpInt <= sInt && sInt <= endEpInt
		}
	}
	
	return false
}

// SplitEpisode 拆分集数，返回开始集数，结束集数，Part信息
func (fp *FormatParser) SplitEpisode(fileName string, fileMeta *models.MetaInfo) (*int, *int, *string) {
	// 指定的具体集数，直接返回
	if fp.startEp != nil {
		if fp.startEp == fp.endEp {
			// `details` 格式�?`X-X` 或�?`X`
			if startEpStr, ok := fp.startEp.(string); ok {
				// `details` 格式�?`X-X`
				parts := strings.Split(startEpStr, "-")
				if len(parts) == 2 {
					s := parts[0]
					e := parts[1]
					startEpExpr := strings.Replace(fp.offset, "EP", s, -1)
					endEpExpr := strings.Replace(fp.offset, "EP", e, -1)
					
					startEpVal := fp.evalExpression(startEpExpr)
					endEpVal := fp.evalExpression(endEpExpr)
					
					sInt, _ := strconv.Atoi(s)
					eInt, _ := strconv.Atoi(e)
					if sInt == eInt {
						return &startEpVal, nil, fp.part
					}
					return &startEpVal, &endEpVal, fp.part
				}
			} else if startEpInt, ok := fp.startEp.(int); ok {
				// `details` 格式�?`X`
				startEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(startEpInt), -1)
				startEpVal := fp.evalExpression(startEpExpr)
				return &startEpVal, nil, fp.part
			}
		} else if fp.format == "" {
			// `details` 格式�?`X,X`
			if startEpInt, ok := fp.startEp.(int); ok {
				if endEpInt, ok := fp.endEp.(int); ok {
					startEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(startEpInt), -1)
					endEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(endEpInt), -1)
					
					startEpVal := fp.evalExpression(startEpExpr)
					endEpVal := fp.evalExpression(endEpExpr)
					return &startEpVal, &endEpVal, fp.part
				}
			}
		}
	}
	
	if fp.format == "" {
		// 未填入`集数定位` 且没有`指定集数` 仅处理`集数偏移`
		var startEp, endEp *int
		
		if fileMeta.BeginEpisode != 0 {
			startEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(fileMeta.BeginEpisode), -1)
			startEpVal := fp.evalExpression(startEpExpr)
			startEp = &startEpVal
		}
		
		if fileMeta.EndEpisode != 0 {
			endEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(fileMeta.EndEpisode), -1)
			endEpVal := fp.evalExpression(endEpExpr)
			endEp = &endEpVal
		}
		
		return startEp, endEp, fp.part
	} else {
		// 有`集数定位`
		s, e := fp.handleSingle(fileName)
		var startEp, endEp *int
		
		if s != nil {
			startEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(*s), -1)
			startEpVal := fp.evalExpression(startEpExpr)
			startEp = &startEpVal
		}
		
		if e != nil {
			endEpExpr := strings.Replace(fp.offset, "EP", strconv.Itoa(*e), -1)
			endEpVal := fp.evalExpression(endEpExpr)
			endEp = &endEpVal
		}
		
		return startEp, endEp, fp.part
	}
}

// handleSingle 处理单集，返回单集的开始和结束集数
func (fp *FormatParser) handleSingle(file string) (*int, *int) {
	if fp.format == "" {
		return nil, nil
	}
	
	// 简化的解析逻辑，Go中没有Python的parse库，需要手动实�?	// 这里实现一个基础版本的解析逻辑
	
	// 将format中的{key}替换为正则表达式捕获�?	escapedFormat := regexp.QuoteMeta(fp.format)
	pattern := strings.Replace(escapedFormat, "\\{"+fp.key+"\\}", `([^/]+)`, -1)
	
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(file)
	
	if len(matches) <= 1 {
		return nil, nil
	}
	
	// 第一个捕获组是匹配的内容
	episodes := matches[1]
	
	// 验证episodes格式
	epRe := regexp.MustCompile(`^(EP)?(\d{1,4})(-(EP)?(\d{1,4}))?$`)
	if !epRe.MatchString(episodes) {
		return nil, nil
	}
	
	// 分割集数
	splitRe := regexp.MustCompile(fp.splitChars)
	episodeSplits := splitRe.Split(episodes, -1)
	
	// 过滤出包含数字的部分
	var filteredSplits []string
	numRe := regexp.MustCompile(`[a-zA-Z]*\d{1,4}`)
	for _, split := range episodeSplits {
		if numRe.MatchString(split) {
			filteredSplits = append(filteredSplits, split)
		}
	}
	
	if len(filteredSplits) == 0 {
		return nil, nil
	}
	
	// 移除字母前缀，只保留数字
	numOnlyRe := regexp.MustCompile(`[a-zA-Z]*`)
	firstNumStr := numOnlyRe.ReplaceAllString(filteredSplits[0], "")
	firstNum, err := strconv.Atoi(firstNumStr)
	if err != nil {
		return nil, nil
	}
	
	if len(filteredSplits) == 1 {
		return &firstNum, nil
	}
	
	secondNumStr := numOnlyRe.ReplaceAllString(filteredSplits[1], "")
	secondNum, err := strconv.Atoi(secondNumStr)
	if err != nil {
		return &firstNum, nil
	}
	
	return &firstNum, &secondNum
}

// evalExpression 计算表达式（简化版�?func (fp *FormatParser) evalExpression(expr string) int {
	// 简化版表达式计算，仅支持基本的 + - * / 运算
	// 移除EP前缀并计算表达式
	
	// 替换EP为数�?	re := regexp.MustCompile(`EP`)
	expr = re.ReplaceAllString(expr, "")
	
	// 解析简单的数学表达�?	// 支持 +, -, *, / 运算
	
	// 处理乘法和除�?	mulDivRe := regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s*([*/])\s*(-?\d+(?:\.\d+)?)`)
	for mulDivRe.MatchString(expr) {
		match := mulDivRe.FindStringSubmatch(expr)
		if len(match) == 4 {
			left, _ := strconv.ParseFloat(match[1], 64)
			right, _ := strconv.ParseFloat(match[3], 64)
			var result float64
			if match[2] == "*" {
				result = left * right
			} else {
				result = left / right
			}
			expr = strings.Replace(expr, match[0], strconv.FormatFloat(result, 'f', -1, 64), 1)
		}
	}
	
	// 处理加法和减�?	addSubRe := regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s*([+\-])\s*(-?\d+(?:\.\d+)?)`)
	for addSubRe.MatchString(expr) {
		match := addSubRe.FindStringSubmatch(expr)
		if len(match) == 4 {
			left, _ := strconv.ParseFloat(match[1], 64)
			right, _ := strconv.ParseFloat(match[3], 64)
			var result float64
			if match[2] == "+" {
				result = left + right
			} else {
				result = left - right
			}
			expr = strings.Replace(expr, match[0], strconv.FormatFloat(result, 'f', -1, 64), 1)
		}
	}
	
	result, _ := strconv.ParseFloat(expr, 64)
	return int(math.Round(result))
}
