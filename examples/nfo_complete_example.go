package main

import (
	"fmt"
	"os"
	"path/filepath"

	"moviepilot-go/internal/helper"
)

func main() {
	// 创建更复杂的示例NFO文件内容
	nfoContent := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
    <title>复仇者联�?/title>
    <originaltitle>The Avengers</originaltitle>
    <year>2012</year>
    <rating>8.0</rating>
    <plot>地球面临的敌人太强大，只有集合所有超级英雄的力量才能拯救世界�?/plot>
    <genre>动作</genre>
    <genre>科幻</genre>
    <genre>冒险</genre>
    <director>Joss Whedon</director>
    <writer>Zak Penn</writer>
    <writer>Joss Whedon</writer>
    <actor>
        <name>Robert Downey Jr.</name>
        <role>Tony Stark / Iron Man</role>
    </actor>
    <actor>
        <name>Chris Evans</name>
        <role>Steve Rogers / Captain America</role>
    </actor>
    <actor>
        <name>Scarlett Johansson</name>
        <role>Natasha Romanoff / Black Widow</role>
    </actor>
    <thumb aspect="poster" preview="poster_thumb.jpg">poster.jpg</thumb>
    <thumb aspect="banner" preview="banner_thumb.jpg">banner.jpg</thumb>
    <fanart>
        <thumb preview="fanart_thumb.jpg">fanart.jpg</thumb>
    </fanart>
    <runtime>143</runtime>
    <country>USA</country>
    <studio>Marvel Studios</studio>
</movie>`

	// 创建临时NFO文件
	tempDir := os.TempDir()
	nfoFilePath := filepath.Join(tempDir, "avengers.nfo")
	
	err := os.WriteFile(nfoFilePath, []byte(nfoContent), 0644)
	if err != nil {
		fmt.Printf("创建示例NFO文件失败: %v\n", err)
		return
	}
	
	defer os.Remove(nfoFilePath) // 清理临时文件

	// 创建NFO读取器实�?	fmt.Println("=== 创建NFO读取�?===")
	nfoReader, err := helper.NewNfoReader(nfoFilePath)
	if err != nil {
		fmt.Printf("创建NFO读取器失�? %v\n", err)
		return
	}
	fmt.Println("NFO读取器创建成�?)

	// 示例1: 获取基本信息
	fmt.Println("\n=== 获取基本信息 ===")
	if title := nfoReader.GetElementValue("title"); title != nil {
		fmt.Printf("电影标题: %s\n", *title)
	}

	if originalTitle := nfoReader.GetElementValue("originaltitle"); originalTitle != nil {
		fmt.Printf("原始标题: %s\n", *originalTitle)
	}

	if year := nfoReader.GetElementValue("year"); year != nil {
		fmt.Printf("年份: %s\n", *year)
	}

	if rating := nfoReader.GetElementValue("rating"); rating != nil {
		fmt.Printf("评分: %s\n", *rating)
	}

	if plot := nfoReader.GetElementValue("plot"); plot != nil {
		fmt.Printf("剧情简�? %s\n", *plot)
	}

	if runtime := nfoReader.GetElementValue("runtime"); runtime != nil {
		fmt.Printf("时长: %s 分钟\n", *runtime)
	}

	// 示例2: 获取多个相同标签的元�?	fmt.Println("\n=== 获取多个元素 ===")
	genres := nfoReader.GetElements("genre")
	fmt.Printf("类型 (%d):\n", len(genres))
	for i, genre := range genres {
		fmt.Printf("  %d. %s\n", i+1, genre.Content)
	}

	writers := nfoReader.GetElements("writer")
	fmt.Printf("编剧 (%d):\n", len(writers))
	for i, writer := range writers {
		fmt.Printf("  %d. %s\n", i+1, writer.Content)
	}

	// 示例3: 获取嵌套元素
	fmt.Println("\n=== 获取嵌套元素 ===")
	actors := nfoReader.GetElements("actor")
	fmt.Printf("演员 (%d):\n", len(actors))
	for i, actor := range actors {
		name := getElementContent(actor, "name")
		role := getElementContent(actor, "role")
		if name != nil && role != nil {
			fmt.Printf("  %d. %s (%s)\n", i+1, *name, *role)
		} else if name != nil {
			fmt.Printf("  %d. %s\n", i+1, *name)
		}
	}

	// 示例4: 获取带属性的元素
	fmt.Println("\n=== 获取带属性的元素 ===")
	thumbs := nfoReader.GetElements("thumb")
	fmt.Printf("缩略�?(%d):\n", len(thumbs))
	for i, thumb := range thumbs {
		aspect := getAttributeValue(thumb, "aspect")
		preview := getAttributeValue(thumb, "preview")
		if aspect != nil && preview != nil {
			fmt.Printf("  %d. %s (aspect: %s, preview: %s)\n", i+1, thumb.Content, *aspect, *preview)
		} else if aspect != nil {
			fmt.Printf("  %d. %s (aspect: %s)\n", i+1, thumb.Content, *aspect)
		} else {
			fmt.Printf("  %d. %s\n", i+1, thumb.Content)
		}
	}

	// 示例5: 获取深层嵌套元素
	fmt.Println("\n=== 获取深层嵌套元素 ===")
	fanartThumbs := nfoReader.GetElements("fanart/thumb")
	fmt.Printf("Fanart缩略�?(%d):\n", len(fanartThumbs))
	for i, thumb := range fanartThumbs {
		preview := getAttributeValue(thumb, "preview")
		if preview != nil {
			fmt.Printf("  %d. %s (preview: %s)\n", i+1, thumb.Content, *preview)
		} else {
			fmt.Printf("  %d. %s\n", i+1, thumb.Content)
		}
	}

	// 示例6: 检查元素是否存�?	fmt.Println("\n=== 检查元素是否存�?===")
	checkElements := []string{"title", "year", "rating", "outline", "tagline", "mpaa"}
	for _, element := range checkElements {
		exists := nfoReader.HasElement(element)
		status := "不存�?
		if exists {
			status = "存在"
		}
		fmt.Printf("%s: %s\n", element, status)
	}

	// 示例7: 获取属性�?	fmt.Println("\n=== 获取特定属性�?===")
	if posterAspect := nfoReader.GetAttrValue("thumb", "aspect"); posterAspect != nil {
		fmt.Printf("第一个thumb元素的aspect属�? %s\n", *posterAspect)
	}

	// 示例8: 获取根节�?	fmt.Println("\n=== 获取根节点信�?===")
	root := nfoReader.GetRoot()
	if root != nil {
		fmt.Printf("根节点名�? %s\n", root.XMLName.Local)
		fmt.Printf("子元素数�? %d\n", len(root.Children))
	}
}

// getElementContent 获取元素的内�?func getElementContent(node *helper.xmlNode, elementName string) *string {
	for _, child := range node.Children {
		if child.XMLName.Local == elementName && child.Content != "" {
			return &child.Content
		}
	}
	return nil
}

// getAttributeValue 获取属性�?func getAttributeValue(node *helper.xmlNode, attrName string) *string {
	for _, attr := range node.Attrs {
		if attr.Name.Local == attrName {
			return &attr.Value
		}
	}
	return nil
}
