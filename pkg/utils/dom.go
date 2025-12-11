package utils

import (
	"fmt"
)

// XMLNode 是一个轻量级的 DOM 结构，用于在 Go 中模拟 Python DomUtils 所依赖的 XML DOM 行为
// 它不直接绑定到 encoding/xml，调用方可以自行在解析/构建 XML 时与该结构互转。
type XMLNode struct {
	Name     string            // 标签名
	Attrs    map[string]string // 属性
	Children []*XMLNode        // 子节点
	Text     string            // 文本内容（若有）
}

// XMLDocument 模拟 XML 文档，用于创建节点
type XMLDocument struct {
	Root *XMLNode // 根节点
}

// NewXMLDocument 创建一个新的 XML 文档
func NewXMLDocument() *XMLDocument {
	return &XMLDocument{}
}

// CreateElement 创建一个新的元素节点
func (doc *XMLDocument) CreateElement(name string) *XMLNode {
	return &XMLNode{
		Name:  name,
		Attrs: make(map[string]string),
	}
}

// CreateTextNode 创建一个新的文本节点
func (doc *XMLDocument) CreateTextNode(value string) *XMLNode {
	return &XMLNode{
		Text: value,
	}
}

// TagValue 对应 Python DomUtils.tag_value：
// 在给定节点的直接子节点中查找第一个标签名为 tagName 的节点，
// 如果指定了 attrName，则返回该属性的值；否则返回该节点的 Text；
// 若均不存在，则返回 defaultValue。
// 支持多种类型的 defaultValue（string 或 int）
func TagValue(node *XMLNode, tagName, attrName string, defaultValue interface{}) interface{} {
	if node == nil {
		return defaultValue
	}

	// 只查找直接子节点，与 Python getElementsByTagName 行为一致
	var found *XMLNode
	for _, child := range node.Children {
		if child.Name == tagName {
			found = child
			break
		}
	}

	if found == nil {
		return defaultValue
	}

	if attrName != "" {
		if found.Attrs != nil {
			if v, ok := found.Attrs[attrName]; ok && v != "" {
				return v
			}
		}
	}

	if found.Text != "" {
		return found.Text
	}

	return defaultValue
}

// TagValueString 是 TagValue 的字符串版本，返回字符串类型
func TagValueString(node *XMLNode, tagName, attrName, defaultValue string) string {
	result := TagValue(node, tagName, attrName, defaultValue)
	if str, ok := result.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", result)
}

// TagValueInt 是 TagValue 的整数版本，返回整数类型
func TagValueInt(node *XMLNode, tagName, attrName string, defaultValue int) int {
	result := TagValue(node, tagName, attrName, defaultValue)
	if i, ok := result.(int); ok {
		return i
	}
	return defaultValue
}

// AddNode 对应 Python DomUtils.add_node：在 parent 下创建一个名为 name 的子节点，
// 若 value 非空，则设置其 Text 字段，最后返回新创建的节点。
func AddNode(doc *XMLDocument, parent *XMLNode, name string, value interface{}) *XMLNode {
	if parent == nil {
		return nil
	}

	// 创建新节点
	node := doc.CreateElement(name)
	parent.Children = append(parent.Children, node)

	// 设置文本值（如果有）
	if value != nil {
		text := doc.CreateTextNode(fmt.Sprintf("%v", value))
		// 在Go实现中，我们直接将文本值赋给节点的Text字段
		node.Text = text.Text
	}

	return node
}
