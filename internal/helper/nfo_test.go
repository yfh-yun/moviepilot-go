package helper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNfoReader(t *testing.T) {
	// 创建测试NFO文件内容
	nfoContent := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
    <title>测试电影</title>
    <originaltitle>Test Movie</originaltitle>
    <year>2023</year>
    <rating>7.5</rating>
    <plot>这是一个测试电影的剧情简介�?/plot>
    <genre>动作</genre>
    <genre>科幻</genre>
    <director>测试导演</director>
    <actor>
        <name>测试演员1</name>
        <role>主角</role>
    </actor>
    <actor>
        <name>测试演员2</name>
        <role>配角</role>
    </actor>
    <thumb aspect="poster">poster.jpg</thumb>
    <thumb aspect="banner">banner.jpg</thumb>
</movie>`

	// 创建临时测试目录
	tempDir := t.TempDir()
	nfoFilePath := filepath.Join(tempDir, "test.nfo")
	
	// 写入测试文件
	err := os.WriteFile(nfoFilePath, []byte(nfoContent), 0644)
	if err != nil {
		t.Fatalf("创建测试NFO文件失败: %v", err)
	}

	// 测试创建NfoReader实例
	t.Run("创建NfoReader实例", func(t *testing.T) {
		reader, err := NewNfoReader(nfoFilePath)
		if err != nil {
			t.Errorf("创建NfoReader实例失败: %v", err)
		}
		if reader == nil {
			t.Error("NfoReader实例为nil")
		}
	})

	// 测试获取单个元素�?	t.Run("获取单个元素�?, func(t *testing.T) {
		reader, _ := NewNfoReader(nfoFilePath)
		
		// 测试存在的元�?		title := reader.GetElementValue("title")
		if title == nil {
			t.Error("无法获取title元素")
		} else if *title != "测试电影" {
			t.Errorf("title元素值不正确，期�? 测试电影, 实际: %s", *title)
		}
		
		year := reader.GetElementValue("year")
		if year == nil {
			t.Error("无法获取year元素")
		} else if *year != "2023" {
			t.Errorf("year元素值不正确，期�? 2023, 实际: %s", *year)
		}
		
		// 测试不存在的元素
		unknown := reader.GetElementValue("unknown")
		if unknown != nil {
			t.Error("获取不存在的元素应该返回nil")
		}
	})

	// 测试获取多个元素
	t.Run("获取多个元素", func(t *testing.T) {
		reader, _ := NewNfoReader(nfoFilePath)
		
		genres := reader.GetElements("genre")
		if len(genres) != 2 {
			t.Errorf("genre元素数量不正确，期望: 2, 实际: %d", len(genres))
		}
		
		actors := reader.GetElements("actor")
		if len(actors) != 2 {
			t.Errorf("actor元素数量不正确，期望: 2, 实际: %d", len(actors))
		}
	})

	// 测试文件不存在的情况
	t.Run("文件不存�?, func(t *testing.T) {
		_, err := NewNfoReader("nonexistent.nfo")
		if err == nil {
			t.Error("文件不存在时应该返回错误")
		}
	})
}

func TestNfoReaderEdgeCases(t *testing.T) {
	// 测试空内容的元素
	nfoContent := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
    <title>测试电影</title>
    <empty></empty>
    <whitespace>   </whitespace>
    <nested>
        <child>子元素内�?/child>
    </nested>
</movie>`

	tempDir := t.TempDir()
	nfoFilePath := filepath.Join(tempDir, "edge_cases.nfo")
	
	err := os.WriteFile(nfoFilePath, []byte(nfoContent), 0644)
	if err != nil {
		t.Fatalf("创建测试NFO文件失败: %v", err)
	}

	reader, _ := NewNfoReader(nfoFilePath)
	
	// 测试空元�?	t.Run("空元�?, func(t *testing.T) {
		empty := reader.GetElementValue("empty")
		if empty != nil {
			t.Error("空元素应该返回nil")
		}
	})

	// 测试只有空白字符的元�?	t.Run("空白字符元素", func(t *testing.T) {
		whitespace := reader.GetElementValue("whitespace")
		if whitespace == nil {
			t.Error("空白字符元素不应该返回nil")
		} else if *whitespace != "   " {
			t.Errorf("空白字符元素值不正确，期�? '   ', 实际: '%s'", *whitespace)
		}
	})
}
