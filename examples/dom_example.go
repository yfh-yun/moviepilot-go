// Package main 提供DOM工具使用示例
package main

import (
	"fmt"
	
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== DOM工具使用示例 ===")
	
	domUtils := &utils.DomUtils{}
	
	// 测试XML数据
	xmlData := `
		<root>
			<item id="1" name="item1">Item 1 Content</item>
			<item id="2" name="item2">Item 2 Content</item>
			<book isbn="123456">Go语言编程</book>
			<selfclosing />
			<empty></empty>
			<htmlcontent>&lt;div&gt;HTML内容&lt;/div&gt;</htmlcontent>
		</root>
	`
	
	// 示例1: 获取标签内容
	fmt.Println("\n1. 获取标签内容:")
	content := domUtils.GetElementByName(xmlData, "book", "默认�?)
	fmt.Printf("book标签内容: %s\n", content)
	
	// 示例2: 获取标签属�?	fmt.Println("\n2. 获取标签属�?")
	idValue := domUtils.TagValue(xmlData, "item", "id", "默认ID")
	nameValue := domUtils.TagValue(xmlData, "item", "name", "默认名称")
	fmt.Printf("item标签的id属�? %s\n", idValue)
	fmt.Printf("item标签的name属�? %s\n", nameValue)
	
	// 示例3: 获取所有匹配标�?	fmt.Println("\n3. 获取所有匹配标�?")
	items := domUtils.GetElementsByName([]byte(xmlData), "item")
	for i, item := range items {
		fmt.Printf("�?d个item内容: %s\n", i+1, item)
	}
	
	// 示例4: 添加节点
	fmt.Println("\n4. 添加节点:")
	newXML := `<root><item id="1">Item 1</item></root>`
	updatedXML := domUtils.AddNodeToXML(newXML, "root", "item", "Item 2")
	fmt.Printf("更新后的XML: %s\n", updatedXML)
	
	// 示例5: 创建新节�?	fmt.Println("\n5. 创建新节�?")
	newNode := domUtils.AddNode("", "title", "这是标题")
	fmt.Printf("创建的新节点: %s\n", newNode)
	
	newNodeWithNum := domUtils.AddNode("", "number", 123)
	fmt.Printf("创建的数字节�? %s\n", newNodeWithNum)
	
	// 示例6: 处理自闭合标签和空标�?	fmt.Println("\n6. 处理特殊标签:")
	selfClosingContent := domUtils.GetElementByName(xmlData, "selfclosing", "无内�?)
	emptyContent := domUtils.GetElementByName(xmlData, "empty", "无内�?)
	fmt.Printf("自闭合标签内�? '%s'\n", selfClosingContent)
	fmt.Printf("空标签内�? '%s'\n", emptyContent)
	
	// 示例7: 处理HTML实体
	fmt.Println("\n7. 处理HTML实体:")
	htmlContent := domUtils.GetElementByName(xmlData, "htmlcontent", "无内�?)
	fmt.Printf("HTML内容标签: '%s'\n", htmlContent)
	
	// 示例8: 使用默认�?	fmt.Println("\n8. 使用默认�?")
	nonExistent := domUtils.TagValue(xmlData, "nonexistent", "", "默认内容")
	fmt.Printf("不存在标签的内容: '%s'\n", nonExistent)
	
	nonExistentAttr := domUtils.TagValue(xmlData, "item", "nonexistent", "默认属�?)
	fmt.Printf("不存在属性的�? '%s'\n", nonExistentAttr)
}
