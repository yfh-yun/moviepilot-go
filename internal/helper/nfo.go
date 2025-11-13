package helper

import (
	"encoding/xml"
	"os"
)

// NfoReader NFO文件读取�?type NfoReader struct {
	// xml文件路径
	xmlFilePath string
	// XML根节�?	root *xmlNode
}

// xmlNode XML节点结构
type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []xmlNode  `xml:",any"`
}

// NewNfoReader 创建NFO读取器实�?func NewNfoReader(xmlFilePath string) (*NfoReader, error) {
	/*
	 * 创建NFO读取器实�?	 * :param xmlFilePath: XML文件路径
	 * :return: NFO读取器实例和错误信息
	 */
	reader := &NfoReader{
		xmlFilePath: xmlFilePath,
	}

	// 解析XML文件
	err := reader.parse()
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// parse 解析XML文件
func (n *NfoReader) parse() error {
	/*
	 * 解析XML文件
	 * :return: 错误信息
	 */
	// 检查文件是否存�?	if _, err := os.Stat(n.xmlFilePath); os.IsNotExist(err) {
		return err
	}

	// 读取文件内容
	data, err := os.ReadFile(n.xmlFilePath)
	if err != nil {
		return err
	}

	// 解析XML
	root := &xmlNode{}
	err = xml.Unmarshal(data, root)
	if err != nil {
		return err
	}

	n.root = root
	return nil
}

// GetElementValue 获取元素�?(对应Python版本的get_element_value)
func (n *NfoReader) GetElementValue(elementPath string) *string {
	/*
	 * 获取元素�?	 * :param elementPath: 元素路径
	 * :return: 元素�?	 */
	if n.root == nil {
		return nil
	}

	// 查找元素
	element := n.findElement(n.root, elementPath)
	if element != nil {
		// 如果元素有内容则返回内容，否则返回nil
		if element.Content != "" {
			return &element.Content
		}
		// 如果元素没有内容但有子元素，返回第一个子元素的内�?		if len(element.Children) > 0 && element.Children[0].Content != "" {
			return &element.Children[0].Content
		}
	}
	
	return nil
}

// GetElements 获取元素列表 (对应Python版本的get_elements)
func (n *NfoReader) GetElements(elementPath string) []*xmlNode {
	/*
	 * 获取元素列表
	 * :param elementPath: 元素路径
	 * :return: 元素列表
	 */
	if n.root == nil {
		return make([]*xmlNode, 0)
	}

	// 查找所有匹配的元素
	elements := n.findElements(n.root, elementPath)
	return elements
}

// findElement 查找单个元素
func (n *NfoReader) findElement(node *xmlNode, path string) *xmlNode {
	/*
	 * 查找单个元素
	 * :param node: 起始节点
	 * :param path: 元素路径
	 * :return: 元素节点
	 */
	// 简化实现：仅支持单层路径查找，符合Python原始版本功能
	for i := range node.Children {
		child := &node.Children[i]
		if child.XMLName.Local == path {
			return child
		}
	}
	return nil
}

// findElements 查找多个元素
func (n *NfoReader) findElements(node *xmlNode, path string) []*xmlNode {
	/*
	 * 查找多个元素
	 * :param node: 起始节点
	 * :param path: 元素路径
	 * :return: 元素节点列表
	 */
	elements := make([]*xmlNode, 0)
	
	// 简化实现：仅支持单层路径查找，符合Python原始版本功能
	for i := range node.Children {
		child := &node.Children[i]
		if child.XMLName.Local == path {
			elements = append(elements, child)
		}
	}
	
	return elements
}

// GetAttrValue 获取属性�?func (n *NfoReader) GetAttrValue(elementPath string, attrName string) *string {
	/*
	 * 获取属性�?	 * :param elementPath: 元素路径
	 * :param attrName: 属性名�?	 * :return: 属性�?	 */
	if n.root == nil {
		return nil
	}

	// 查找元素
	element := n.findElement(n.root, elementPath)
	if element != nil {
		// 查找属�?		for _, attr := range element.Attrs {
			if attr.Name.Local == attrName {
				return &attr.Value
			}
		}
	}
	
	return nil
}

// GetAllElements 获取所有元�?func (n *NfoReader) GetAllElements() *xmlNode {
	/*
	 * 获取所有元�?	 * :return: 根节�?	 */
	return n.root
}

// GetRoot 获取根节�?func (n *NfoReader) GetRoot() *xmlNode {
	/*
	 * 获取根节�?	 * :return: 根节�?	 */
	return n.root
}

// HasElement 检查元素是否存�?func (n *NfoReader) HasElement(elementPath string) bool {
	/*
	 * 检查元素是否存�?	 * :param elementPath: 元素路径
	 * :return: 是否存在
	 */
	return n.findElement(n.root, elementPath) != nil
}

// GetElementCount 获取元素数量
func (n *NfoReader) GetElementCount(elementPath string) int {
	/*
	 * 获取元素数量
	 * :param elementPath: 元素路径
	 * :return: 元素数量
	 */
	elements := n.GetElements(elementPath)
	return len(elements)
}

// GetElementContent 获取元素内容
func (n *NfoReader) GetElementContent(node *xmlNode) *string {
	/*
	 * 获取元素内容
	 * :param node: XML节点
	 * :return: 元素内容
	 */
	if node != nil && node.Content != "" {
		return &node.Content
	}
	return nil
}

// GetAttributeValue 获取节点属性�?func (n *NfoReader) GetAttributeValue(node *xmlNode, attrName string) *string {
	/*
	 * 获取节点属性�?	 * :param node: XML节点
	 * :param attrName: 属性名�?	 * :return: 属性�?	 */
	if node != nil {
		for _, attr := range node.Attrs {
			if attr.Name.Local == attrName {
				return &attr.Value
			}
		}
	}
	return nil
}

// GetChildElements 获取子元�?func (n *NfoReader) GetChildElements(node *xmlNode, childName string) []*xmlNode {
	/*
	 * 获取子元�?	 * :param node: XML节点
	 * :param childName: 子元素名�?	 * :return: 子元素列�?	 */
	children := make([]*xmlNode, 0)
	if node != nil {
		for i := range node.Children {
			child := &node.Children[i]
			if child.XMLName.Local == childName {
				children = append(children, child)
			}
		}
	}
	return children
}

// GetFirstChild 获取第一个子元素
func (n *NfoReader) GetFirstChild(node *xmlNode, childName string) *xmlNode {
	/*
	 * 获取第一个子元素
	 * :param node: XML节点
	 * :param childName: 子元素名�?	 * :return: 第一个子元素
	 */
	if node != nil {
		for i := range node.Children {
			child := &node.Children[i]
			if child.XMLName.Local == childName {
				return child
			}
		}
	}
	return nil
}

// GetElementNames 获取所有子元素名称
func (n *NfoReader) GetElementNames(node *xmlNode) []string {
	/*
	 * 获取所有子元素名称
	 * :param node: XML节点
	 * :return: 子元素名称列�?	 */
	names := make([]string, 0)
	if node != nil {
		nameMap := make(map[string]bool)
		for _, child := range node.Children {
			if !nameMap[child.XMLName.Local] {
				nameMap[child.XMLName.Local] = true
				names = append(names, child.XMLName.Local)
			}
		}
	}
	return names
}

// GetElementText 获取元素的文本内容（包括所有子元素的文本）
func (n *NfoReader) GetElementText(elementPath string) *string {
	/*
	 * 获取元素的文本内容（包括所有子元素的文本）
	 * :param elementPath: 元素路径
	 * :return: 元素文本内容
	 */
	if n.root == nil {
		return nil
	}

	element := n.findElement(n.root, elementPath)
	if element != nil {
		text := n.collectText(element)
		if text != "" {
			return &text
		}
	}
	
	return nil
}

// collectText 收集节点的所有文本内�?func (n *NfoReader) collectText(node *xmlNode) string {
	/*
	 * 收集节点的所有文本内�?	 * :param node: XML节点
	 * :return: 文本内容
	 */
	text := node.Content
	
	for _, child := range node.Children {
		text += n.collectText(&child)
	}
	
	return text
}
