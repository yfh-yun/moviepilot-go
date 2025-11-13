package helper

import (
	"testing"
)

func TestNewResourceHelper(t *testing.T) {
	// 测试创建ResourceHelper实例
	helper := NewResourceHelper()
	if helper == nil {
		t.Error("Failed to create ResourceHelper instance")
	}
	
	// 检查基本属性是否正确设�?	if helper.repo == "" {
		t.Error("Repo URL should not be empty")
	}
	
	if helper.filesAPI == "" {
		t.Error("Files API URL should not be empty")
	}
	
	if helper.baseDir == "" {
		t.Error("Base directory should not be empty")
	}
}

func TestGetProxies(t *testing.T) {
	// 测试获取代理设置
	helper := NewResourceHelper()
	proxy := helper.getProxies()
	
	// 由于这是测试环境，proxy可能为空，但我们至少要确保函数能正常执行
	t.Logf("Proxy setting: %s", proxy)
}

func TestResourceHelperCheck(t *testing.T) {
	// 测试Check方法
	helper := NewResourceHelper()
	
	// 由于Check方法会尝试连接外部资源，我们只测试其能正常执行而不报错
	// 在实际环境中，可能需要mock网络请求来进行完整测�?	err := helper.Check()
	if err != nil {
		t.Logf("Check method returned error (may be due to network): %v", err)
	} else {
		t.Log("Check method executed successfully")
	}
}

func TestGetSitesVersions(t *testing.T) {
	// 测试获取站点版本的方�?	authVersion := getSitesAuthVersion()
	if authVersion == "" {
		t.Error("Auth version should not be empty")
	}
	
	indexerVersion := getSitesIndexerVersion()
	if indexerVersion == "" {
		t.Error("Indexer version should not be empty")
	}
	
	t.Logf("Sites auth version: %s", authVersion)
	t.Logf("Sites indexer version: %s", indexerVersion)
}

func TestResourceStructs(t *testing.T) {
	// 测试资源结构�?	resourceInfo := ResourceInfo{
		Version: "1.0.0",
		Resources: map[string]Resource{
			"test": {
				Type:     "auth",
				Platform: "linux",
				Target:   "/path/to/target",
				Version:  "1.0.0",
			},
		},
	}
	
	if resourceInfo.Version != "1.0.0" {
		t.Error("Version not set correctly")
	}
	
	if len(resourceInfo.Resources) != 1 {
		t.Error("Resources not set correctly")
	}
	
	resource := resourceInfo.Resources["test"]
	if resource.Type != "auth" {
		t.Error("Resource type not set correctly")
	}
	
	if resource.Platform != "linux" {
		t.Error("Resource platform not set correctly")
	}
	
	if resource.Target != "/path/to/target" {
		t.Error("Resource target not set correctly")
	}
	
	if resource.Version != "1.0.0" {
		t.Error("Resource version not set correctly")
	}
	
	// 测试文件信息结构�?	fileInfo := FileInfo{
		Name:        "test.txt",
		DownloadURL: "http://example.com/test.txt",
	}
	
	if fileInfo.Name != "test.txt" {
		t.Error("File name not set correctly")
	}
	
	if fileInfo.DownloadURL != "http://example.com/test.txt" {
		t.Error("Download URL not set correctly")
	}
}
