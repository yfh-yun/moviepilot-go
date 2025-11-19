package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ChineseNumberToInt 将中文数字转换为阿拉伯数字
func ChineseNumberToInt(chineseNum string) (int, error) {
	if chineseNum == "" {
		return 0, fmt.Errorf("empty chinese number")
	}

	// 先尝试直接转换为数字
	if num, err := strconv.Atoi(chineseNum); err == nil {
		return num, nil
	}

	// 中文数字映射
	chineseDigits := map[rune]int{
		'零': 0, '〇': 0,
		'一': 1, '壹': 1,
		'二': 2, '贰': 2,
		'三': 3, '叁': 3,
		'四': 4, '肆': 4,
		'五': 5, '伍': 5,
		'六': 6, '陆': 6,
		'七': 7, '柒': 7,
		'八': 8, '捌': 8,
		'九': 9, '玖': 9,
	}

	// 中文单位映射
	chineseUnits := map[rune]int{
		'十': 10, '拾': 10,
		'百': 100, '佰': 100,
		'千': 1000, '仟': 1000,
		'万': 10000, '萬': 10000,
		'亿': 100000000,
	}

	var result int
	var temp int
	var hasUnit bool

	for _, char := range chineseNum {
		if digit, exists := chineseDigits[char]; exists {
			temp = temp*10 + digit
		} else if unit, exists := chineseUnits[char]; exists {
			if unit == 10 && temp == 0 {
				// 处理"十"开头的情况，如"十五"
				temp = 1
			}
			result += temp * unit
			temp = 0
			hasUnit = true
		} else {
			// 遇到非中文数字字符，尝试处理前面的部分
			break
		}
	}

	// 加上剩余的部分
	result += temp

	// 如果没有单位且temp为0，可能是单个数字
	if !hasUnit && temp == 0 {
		for _, char := range chineseNum {
			if digit, exists := chineseDigits[char]; exists {
				result = result*10 + digit
			}
		}
	}

	return result, nil
}

// IsAllChinese 检查字符串是否全为中文
func IsAllChinese(s string) bool {
	for _, r := range s {
		if r < 0x4e00 || r > 0x9fff {
			return false
		}
	}
	return true
}

// IsAllLetters 检查字符串是否全为字母
func IsAllLetters(s string) bool {
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}
	return true
}

// ContainsChinese 检查字符串是否包含中文
func ContainsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}



// NormalizeChineseSpaces 规范化中文字符串的空格
func NormalizeChineseSpaces(s string) string {
	// 移除多余的空格，保留中文之间的适当空格
	s = strings.ReplaceAll(s, "　", " ") // 全角空格转半角
	s = strings.ReplaceAll(s, "  ", " ")  // 多个空格转单个
	return strings.TrimSpace(s)
}

// IntToChineseNumber 将阿拉伯数字转换为中文数字
func IntToChineseNumber(num int) string {
	if num == 0 {
		return "零"
	}

	chineseDigits := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}

	if num < 10 {
		return chineseDigits[num]
	}

	if num <= 99 {
		tens := num / 10
		units := num % 10
		if tens == 1 {
			if units == 0 {
				return "十"
			}
			return "十" + chineseDigits[units]
		}
		if units == 0 {
			return chineseDigits[tens] + "十"
		}
		return chineseDigits[tens] + "十" + chineseDigits[units]
	}

	// 对于更大的数字，简化处理
	return strconv.Itoa(num)
}