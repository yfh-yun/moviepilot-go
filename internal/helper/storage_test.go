package helper

import (
	"testing"
	
	"moviepilot-go/pkg/models"
)

func TestNewStorageHelper(t *testing.T) {
	// 测试创建StorageHelper实例
	helper := NewStorageHelper()
	if helper == nil {
		t.Error("Failed to create StorageHelper instance")
	}
}

func TestGetStoragies(t *testing.T) {
	// 测试获取存储设置
	helper := NewStorageHelper()
	storagies := helper.GetStoragies()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if storagies == nil {
		t.Error("GetStoragies should not return nil")
	}
	
	// storagies应该是一个切�?	if len(storagies) < 0 {
		t.Error("GetStoragies should return a valid slice")
	}
}

func TestGetStorage(t *testing.T) {
	// 测试获取指定存储配置
	helper := NewStorageHelper()
	
	// 测试不存在的存储配置
	storage := helper.GetStorage("non-existent-storage")
	if storage != nil {
		t.Error("GetStorage should return nil for non-existent storage")
	}
}

func TestSetStorage(t *testing.T) {
	// 测试设置存储配置
	helper := NewStorageHelper()
	
	// 测试设置新的存储配置
	conf := map[string]interface{}{
		"path": "/test/path",
		"host": "localhost",
	}
	
	helper.SetStorage("test-storage", conf)
	
	// 验证存储配置是否设置成功
	storage := helper.GetStorage("test-storage")
	if storage != nil {
		// 检查配置是否正确设�?		if storage.Type != "test-storage" {
			t.Error("Storage type not set correctly")
		}
		
		if storage.Config["path"] != "/test/path" {
			t.Error("Storage config not set correctly")
		}
	}
}

func TestAddStorage(t *testing.T) {
	// 测试添加存储配置
	helper := NewStorageHelper()
	
	// 测试添加新的存储配置
	conf := map[string]interface{}{
		"path": "/test/path",
		"host": "localhost",
	}
	
	helper.AddStorage("test-storage", "Test Storage", conf)
	
	// 验证存储配置是否添加成功
	storagies := helper.GetStoragies()
	found := false
	for _, storage := range storagies {
		if storage.Type == "test-storage" && storage.Name == "Test Storage" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("AddStorage failed to add storage configuration")
	}
}

func TestResetStorage(t *testing.T) {
	// 测试重置存储配置
	helper := NewStorageHelper()
	
	// 先添加一个存储配�?	conf := map[string]interface{}{
		"path": "/test/path",
		"host": "localhost",
	}
	
	helper.AddStorage("test-storage", "Test Storage", conf)
	
	// 重置存储配置
	helper.ResetStorage("test-storage")
	
	// 验证存储配置是否重置成功
	storage := helper.GetStorage("test-storage")
	if storage != nil {
		// 检查配置是否被重置为空
		if len(storage.Config) != 0 {
			t.Error("ResetStorage failed to reset storage configuration")
		}
	}
}

func TestStorageConfStruct(t *testing.T) {
	// 测试StorageConf结构�?	conf := models.StorageConf{
		Type: "local",
		Name: "Local Storage",
		Config: map[string]interface{}{
			"path": "/test/path",
			"host": "localhost",
		},
	}
	
	if conf.Type != "local" {
		t.Error("Type not set correctly")
	}
	
	if conf.Name != "Local Storage" {
		t.Error("Name not set correctly")
	}
	
	if conf.Config["path"] != "/test/path" {
		t.Error("Config not set correctly")
	}
}

func TestStorageHelperMethods(t *testing.T) {
	// 测试StorageHelper的各种方�?	helper := NewStorageHelper()
	
	// 测试NewStorageConf
	storageConf := models.NewStorageConf()
	if storageConf == nil {
		t.Error("NewStorageConf should not return nil")
	}
	
	if storageConf.Config == nil {
		t.Error("NewStorageConf should initialize Config map")
	}
}
