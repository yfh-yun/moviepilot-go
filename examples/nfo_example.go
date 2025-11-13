package main

import (
	"fmt"
	"os"
	"path/filepath"

	"moviepilot-go/internal/helper"
)

func main() {
	// 创建示例NFO文件内容
	nfoContent := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
    <title>示例电影</title>
    <originaltitle>Example Movie</originaltitle>
    <year>2023</year>
    <rating>8.5</rating>
    <plot>这是一个示例电影的剧情简介�?/plot>
    <genre>动作</genre>
    <genre>科幻</genre>
    <director>张三</director>
    <actor>
        <name>李四</name>
        <role>主角</role>
    </actor>
    <actor>
        <name>王五</name>
        <role>配角</role>
    </actor>
    <thumb aspect="poster">poster.jpg</thumb>
    <thumb aspect="banner">banner.jpg</thumb>
</movie>`

	// 创建临时NFO文件
	tempDir := os.TempDir()
	nfoFilePath := filepath.Join(tempDir, "example.nfo")
	
	err := os.WriteFile(nfoFilePath, []byte(nfoContent), 0644)
	if err != nil {
		fmt.Printf("创建示例NFO文件失败: %v\n", err)
		return
	}
	
	defer os.Remove(nfoFilePath) // 清理临时文件

	// 创建NFO读取器实�?	nfoReader, err := helper.NewNfoReader(nfoFilePath)
	if err != nil {
		fmt.Printf("创建NFO读取器失�? %v\n", err)
		return
	}

	// 示例1: 获取单个元素�?	fmt.Println("=== 获取单个元素�?===")
	if title := nfoReader.GetElementValue("title"); title != nil {
		fmt.Printf("电影标题: %s\n", *title)
	} else {
		fmt.Println("未找到电影标�?)
	}

	if year := nfoReader.GetElementValue("year"); year != nil {
		fmt.Printf("年份: %s\n", *year)
	} else {
		fmt.Println("未找到年�?)
	}

	if rating := nfoReader.GetElementValue("rating"); rating != nil {
		fmt.Printf("评分: %s\n", *rating)
	} else {
		fmt.Println("未找到评�?)
	}

	// 示例2: 获取多个元素
	fmt.Println("\n=== 获取多个元素 ===")
	genres := nfoReader.GetElements("genre")
	fmt.Printf("类型数量: %d\n", len(genres))
	for i, genre := range genres {
		if genre.Content != "" {
			fmt.Printf("类型%d: %s\n", i+1, genre.Content)
		}
	}

	actors := nfoReader.GetElements("actor")
	fmt.Printf("演员数量: %d\n", len(actors))
	for i, actor := range actors {
		name := nfoReader.findActorName(actor)
		if name != nil {
			fmt.Printf("演员%d: %s\n", i+1, *name)
		}
	}

	// 示例3: 获取属性�?	fmt.Println("\n=== 获取属性�?===")
	thumbs := nfoReader.GetElements("thumb")
	for i, thumb := range thumbs {
		if aspect := nfoReader.findThumbAspect(thumb); aspect != nil {
			fmt.Printf("缩略�?d (aspect=%s): %s\n", i+1, *aspect, thumb.Content)
		} else {
			fmt.Printf("缩略�?d: %s\n", i+1, thumb.Content)
		}
	}

	// 示例4: 获取不存在的元素
	fmt.Println("\n=== 获取不存在的元素 ===")
	if unknown := nfoReader.GetElementValue("unknown"); unknown != nil {
		fmt.Printf("未知元素�? %s\n", *unknown)
	} else {
		fmt.Println("未找到未知元�?)
	}
}

// findActorName 查找演员名称
func (n *helper.NfoReader) findActorName(actor *helper.xmlNode) *string {
	for _, child := range actor.Children {
		if child.XMLName.Local == "name" && child.Content != "" {
			return &child.Content
		}
	}
	return nil
}

// findThumbAspect 查找缩略图aspect属�?func (n *helper.NfoReader) findThumbAspect(thumb *helper.xmlNode) *string {
	for _, attr := range thumb.Attrs {
		if attr.Name.Local == "aspect" {
			return &attr.Value
		}
	}
	return nil
}
