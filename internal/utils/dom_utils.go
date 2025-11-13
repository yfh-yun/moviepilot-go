// Package utils 提供DOM处理相关的工具函�?package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// DomUtils DOM工具�?type DomUtils struct{}

// NewDomUtils 创建一个新�?DomUtils 实例
func NewDomUtils() *DomUtils {
	return &DomUtils{}
}

// RemoveXMLTags 移除XML标签
func (d *DomUtils) RemoveXMLTags(in string) string {
	// 使用正则表达式移除XML标签
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(in, "")
}

// TrimNbspAndSpaces 去除HTML中的&nbsp;和空�?func (d *DomUtils) TrimNbspAndSpaces(in string) string {
	// 替换&nbsp;为空格，然后去除首尾空格
	result := strings.ReplaceAll(in, "&nbsp;", " ")
	result = strings.TrimSpace(result)
	return result
}

// MetaValue 获取meta标签的�?func (d *DomUtils) MetaValue(metaNodes []interface{}, name string) string {
	// 这是一个简化实现，实际项目中可能需要解析HTML节点
	for _, node := range metaNodes {
		// 在实际实现中，你需要检查节点属性是否匹配name
		// 这里只是一个示例实�?		if nodeStr, ok := node.(string); ok {
			if strings.Contains(nodeStr, name) {
				// 提取content属性�?				re := regexp.MustCompile(`content=["']([^"']*)["']`)
				matches := re.FindStringSubmatch(nodeStr)
				if len(matches) > 1 {
					return matches[1]
				}
			}
		}
	}
	return ""
}

// FilterText 过滤文本节点
func (d *DomUtils) FilterText(xpathObj interface{}) string {
	// 这是一个简化实�?	if text, ok := xpathObj.(string); ok {
		return text
	}
	return ""
}

// HTMLUnescape HTML解码
func (d *DomUtils) HTMLUnescape(text string) string {
	return html.UnescapeString(text)
}

// GetPageHTML 获取页面HTML
func (d *DomUtils) GetPageHTML(htmlText string, url string) string {
	// 这是一个简化实�?	// 在实际项目中，你可能需要处理相对链接等
	return htmlText
}

// GetPageTitle 获取页面标题
func (d *DomUtils) GetPageTitle(htmlText string) string {
	// 使用正则表达式提取标�?	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(htmlText)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetPageImage 获取页面图片
func (d *DomUtils) GetPageImage(htmlText string, url string, imgName string) string {
	// 使用正则表达式提取图片链�?	re := regexp.MustCompile(`<img[^>]*src=["']([^"']*)["'][^>]*>`)
	matches := re.FindAllStringSubmatch(htmlText, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			src := match[1]
			// 如果指定了imgName，检查是否匹�?			if imgName == "" || strings.Contains(src, imgName) {
				return src
			}
		}
	}
	return ""
}

// SpaceSubstitution 空格替换
func (d *DomUtils) SpaceSubstitution(text string) string {
	// 替换多种空格字符为空�?	text = strings.ReplaceAll(text, "\u00A0", " ") // &nbsp;
	text = strings.ReplaceAll(text, "\u2002", " ") // en空格
	text = strings.ReplaceAll(text, "\u2003", " ") // em空格
	text = strings.ReplaceAll(text, "\u2009", " ") // 窄空�?	text = strings.ReplaceAll(text, "\u200A", " ") // 发丝空格
	text = strings.ReplaceAll(text, "\u200B", " ") // 零宽空格
	text = strings.ReplaceAll(text, "\u3000", " ") // 中文空格
	text = strings.ReplaceAll(text, "\t", " ")     // 制表�?	text = strings.ReplaceAll(text, "\n", " ")     // 换行�?	text = strings.ReplaceAll(text, "\r", " ")     // 回车�?	
	// 压缩多个连续空格为单个空�?	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// GetHTMLText 获取HTML文本
func (d *DomUtils) GetHTMLText(htmlText string) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlText, "")
	
	// 解码HTML实体
	text = html.UnescapeString(text)
	
	// 替换空格
	text = d.SpaceSubstitution(text)
	
	return text
}

// TagValue 解析XML标签�?// xmlData: XML数据（字符串或[]byte�?// tagName: 标签�?// attrName: 属性名（可选）
// defaultValue: 默认�?// 返回标签值或属性�?func (d *DomUtils) TagValue(xmlData interface{}, tagName, attrName string, defaultValue interface{}) interface{} {
	var xmlStr string
	
	// 处理不同类型的输�?	switch v := xmlData.(type) {
	case string:
		xmlStr = v
	case []byte:
		xmlStr = string(v)
	default:
		return defaultValue
	}
	
	// 如果指定了属性名，则查找属性�?	if attrName != "" {
		// 构造属性匹配的正则表达�?		attrPattern := fmt.Sprintf(`<%s[^>]*%s="([^"]*)"`, tagName, attrName)
		re := regexp.MustCompile(attrPattern)
		matches := re.FindStringSubmatch(xmlStr)
		
		if len(matches) > 1 {
			return matches[1]
		}
		
		// 尝试单引�?		attrPattern = fmt.Sprintf(`<%s[^>]*%s='([^']*)'`, tagName, attrName)
		re = regexp.MustCompile(attrPattern)
		matches = re.FindStringSubmatch(xmlStr)
		
		if len(matches) > 1 {
			return matches[1]
		}
		
		return defaultValue
	}
	
	// 否则查找标签内容
	// 构造标签内容匹配的正则表达�?	tagPattern := fmt.Sprintf(`<%s(?:[^>]*>)?([^<]*)</%s>`, tagName, tagName)
	re := regexp.MustCompile(tagPattern)
	matches := re.FindStringSubmatch(xmlStr)
	
	if len(matches) > 1 {
		// 对HTML实体进行解码
		return html.UnescapeString(strings.TrimSpace(matches[1]))
	}
	
	// 尝试自闭合标�?	selfClosingPattern := fmt.Sprintf(`<%s[^>]*/?>`, tagName)
	re = regexp.MustCompile(selfClosingPattern)
	matches = re.FindStringSubmatch(xmlStr)
	
	if len(matches) > 0 {
		// 自闭合标签没有内容，返回空字符串
		return ""
	}
	
	return defaultValue
}

// AddNode 添加一个XML节点
// parent: 父元�?// name: 节点名称
// value: 节点值（可选）
// 返回完整的XML节点字符�?func (d *DomUtils) AddNode(parent, name string, value interface{}) string {
	var nodeValue string
	
	// 处理不同类型的值输�?	if value != nil {
		nodeValue = fmt.Sprintf("%v", value)
	}
	
	// 转义特殊字符
	escapedValue := html.EscapeString(nodeValue)
	
	// 构造节�?	if escapedValue == "" {
		return fmt.Sprintf("<%s></%s>", name, name)
	}
	
	return fmt.Sprintf("<%s>%s</%s>", name, escapedValue, name)
}

// AddNodeToXML 向XML中添加节点（针对XML字符串）
// xmlData: 原始XML数据
// parentTag: 父标签名
// name: 新节点名�?// value: 新节点�?// 返回更新后的XML字符�?func (d *DomUtils) AddNodeToXML(xmlData, parentTag, name string, value interface{}) string {
	var nodeValue string
	
	// 处理不同类型的值输�?	if value != nil {
		nodeValue = fmt.Sprintf("%v", value)
	}
	
	// 转义特殊字符
	escapedValue := html.EscapeString(nodeValue)
	
	// 构造新节点
	newNode := ""
	if escapedValue == "" {
		newNode = fmt.Sprintf("<%s></%s>", name, name)
	} else {
		newNode = fmt.Sprintf("<%s>%s</%s>", name, escapedValue, name)
	}
	
	// 查找父标签的结束位置
	endTag := fmt.Sprintf("</%s>", parentTag)
	pos := strings.LastIndex(xmlData, endTag)
	
	if pos == -1 {
		// 如果没找到结束标签，直接追加
		return xmlData + newNode
	}
	
	// 在父标签结束前插入新节点
	return xmlData[:pos] + newNode + xmlData[pos:]
}

// GetElementsByName 通过名称获取所有匹配的元素内容
// xmlData: XML数据
// tagName: 标签�?// 返回所有匹配的元素内容数组
func (d *DomUtils) GetElementsByName(xmlData []byte, tagName string) []string {
	xmlStr := string(xmlData)
	
	// 构造标签内容匹配的正则表达式（匹配所有相同标签）
	tagPattern := fmt.Sprintf(`<%s(?:[^>]*>)?([^<]*)</%s>`, tagName, tagName)
	re := regexp.MustCompile(tagPattern)
	matches := re.FindAllStringSubmatch(xmlStr, -1)
	
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			// 对HTML实体进行解码
			results = append(results, html.UnescapeString(strings.TrimSpace(match[1])))
		}
	}
	
	return results
}

// GetElementByName 通过名称获取第一个匹配的元素内容
// xmlData: XML数据 ([]byte或string)
// tagName: 标签�?// defaultValue: 默认�?// 返回第一个匹配的元素内容
func (d *DomUtils) GetElementByName(xmlData interface{}, tagName string, defaultValue string) string {
	var xmlStr string
	
	// 处理不同类型的输�?	switch v := xmlData.(type) {
	case string:
		xmlStr = v
	case []byte:
		xmlStr = string(v)
	default:
		return defaultValue
	}
	
	// 构造标签内容匹配的正则表达�?	tagPattern := fmt.Sprintf(`<%s(?:[^>]*>)?([^<]*)</%s>`, tagName, tagName)
	re := regexp.MustCompile(tagPattern)
	matches := re.FindStringSubmatch(xmlStr)
	
	if len(matches) > 1 {
		// 对HTML实体进行解码
		return html.UnescapeString(strings.TrimSpace(matches[1]))
	}
	
	return defaultValue
}
