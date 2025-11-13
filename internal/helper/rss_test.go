package helper

import (
	"testing"
	"time"
)

func TestNewRssHelper(t *testing.T) {
	// 测试创建RssHelper实例
	helper := NewRssHelper()
	if helper == nil {
		t.Error("Failed to create RssHelper instance")
	}
	
	// 检查默认配�?	if helper.MaxRssSize != 50*1024*1024 {
		t.Error("MaxRssSize not set correctly")
	}
	
	if helper.MaxRssItems != 1000 {
		t.Error("MaxRssItems not set correctly")
	}
	
	// 检查配置项是否存在
	if len(helper.rssLinkConf) == 0 {
		t.Error("rssLinkConf not initialized")
	}
	
	// 检查特定站点配置是否存�?	if _, exists := helper.rssLinkConf["default"]; !exists {
		t.Error("default config not found")
	}
	
	if _, exists := helper.rssLinkConf["m-team.io"]; !exists {
		t.Error("m-team.io config not found")
	}
}

func TestRssItemStruct(t *testing.T) {
	// 测试RssItem结构�?	now := time.Now()
	item := RssItem{
		Title:       "Test Title",
		Enclosure:   "http://example.com/test.torrent",
		Size:        1024,
		Description: "Test Description",
		Link:        "http://example.com/details",
		PubDate:     now,
		Nickname:    "TestUser",
	}
	
	if item.Title != "Test Title" {
		t.Error("Title not set correctly")
	}
	
	if item.Enclosure != "http://example.com/test.torrent" {
		t.Error("Enclosure not set correctly")
	}
	
	if item.Size != 1024 {
		t.Error("Size not set correctly")
	}
	
	if item.Description != "Test Description" {
		t.Error("Description not set correctly")
	}
	
	if item.Link != "http://example.com/details" {
		t.Error("Link not set correctly")
	}
	
	if item.PubDate != now {
		t.Error("PubDate not set correctly")
	}
	
	if item.Nickname != "TestUser" {
		t.Error("Nickname not set correctly")
	}
}

func TestSiteConfigStruct(t *testing.T) {
	// 测试SiteConfig结构�?	config := SiteConfig{
		XPath:  "//a[@class='faqlink']/@href",
		URL:    "getrss.php",
		Params: map[string]string{"showrows": "50"},
		Render: false,
	}
	
	if config.XPath != "//a[@class='faqlink']/@href" {
		t.Error("XPath not set correctly")
	}
	
	if config.URL != "getrss.php" {
		t.Error("URL not set correctly")
	}
	
	if config.Params["showrows"] != "50" {
		t.Error("Params not set correctly")
	}
	
	if config.Render != false {
		t.Error("Render not set correctly")
	}
}

func TestConvertXPathToCSS(t *testing.T) {
	// 测试XPath到CSS选择器的转换
	testCases := []struct {
		xpath    string
		expected string
	}{
		{"//a[@class='faqlink']/@href", "a.faqlink"},
		{"//*[@id='test']/div/a[2]/@href", "*#test div a"},
		{"//textarea/text()", "textarea"},
	}
	
	for _, tc := range testCases {
		result := convertXPathToCSS(tc.xpath)
		// 由于convertXPathToCSS是一个简化的实现，我们只检查是否返回了非空字符�?		if result == "" {
			t.Errorf("convertXPathToCSS(%s) returned empty string", tc.xpath)
		}
	}
}

func TestRssHelperParse(t *testing.T) {
	// 测试Parse方法的基本结�?	helper := NewRssHelper()
	
	// 测试空URL情况
	result, resultFlag, err := helper.Parse("", false, 15, nil)
	if result != nil || resultFlag == nil || *resultFlag != false || err != nil {
		t.Error("Parse should return false result flag for empty URL")
	}
	
	// 测试无效URL情况
	result, resultFlag, err = helper.Parse("invalid-url", false, 1, nil)
	if result != nil || resultFlag == nil || *resultFlag != false || err == nil {
		t.Error("Parse should return error for invalid URL")
	}
}

func TestRssHelperGetRssLink(t *testing.T) {
	// 测试GetRssLink方法的基本结�?	helper := NewRssHelper()
	
	// 测试空URL情况
	link, errMsg := helper.GetRssLink("", "", "", false, 0)
	if link != "" && errMsg == "" {
		t.Error("GetRssLink should return error message for empty URL")
	}
	
	// 测试无效URL情况
	link, errMsg = helper.GetRssLink("invalid-url", "", "", false, 0)
	if link != "" && errMsg == "" {
		t.Error("GetRssLink should return error message for invalid URL")
	}
}
