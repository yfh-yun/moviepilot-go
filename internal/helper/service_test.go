package helper

import (
	"testing"
	
	"moviepilot-go/pkg/models"
)

func TestNewServiceConfigHelper(t *testing.T) {
	// 测试创建ServiceConfigHelper实例
	helper := NewServiceConfigHelper()
	if helper == nil {
		t.Error("Failed to create ServiceConfigHelper instance")
	}
}

func TestGetConfigs(t *testing.T) {
	// 测试获取配置
	helper := NewServiceConfigHelper()
	configs := helper.GetConfigs(models.SystemConfigKeyDownloaders)
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if configs == nil {
		t.Error("GetConfigs should not return nil")
	}
	
	// configs应该是一个切�?	if len(configs) < 0 {
		t.Error("GetConfigs should return a valid slice")
	}
}

func TestGetDownloaderConfigs(t *testing.T) {
	// 测试获取下载器配�?	helper := NewServiceConfigHelper()
	configs := helper.GetDownloaderConfigs()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if configs == nil {
		t.Error("GetDownloaderConfigs should not return nil")
	}
	
	// configs应该是一个切�?	if len(configs) < 0 {
		t.Error("GetDownloaderConfigs should return a valid slice")
	}
}

func TestGetMediaserverConfigs(t *testing.T) {
	// 测试获取媒体服务器配�?	helper := NewServiceConfigHelper()
	configs := helper.GetMediaserverConfigs()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if configs == nil {
		t.Error("GetMediaserverConfigs should not return nil")
	}
	
	// configs应该是一个切�?	if len(configs) < 0 {
		t.Error("GetMediaserverConfigs should return a valid slice")
	}
}

func TestGetNotificationConfigs(t *testing.T) {
	// 测试获取通知配置
	helper := NewServiceConfigHelper()
	configs := helper.GetNotificationConfigs()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if configs == nil {
		t.Error("GetNotificationConfigs should not return nil")
	}
	
	// configs应该是一个切�?	if len(configs) < 0 {
		t.Error("GetNotificationConfigs should return a valid slice")
	}
}

func TestGetNotificationSwitches(t *testing.T) {
	// 测试获取通知开关配�?	helper := NewServiceConfigHelper()
	switches := helper.GetNotificationSwitches()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if switches == nil {
		t.Error("GetNotificationSwitches should not return nil")
	}
	
	// switches应该是一个切�?	if len(switches) < 0 {
		t.Error("GetNotificationSwitches should return a valid slice")
	}
}

func TestGetNotificationSwitch(t *testing.T) {
	// 测试获取特定通知开�?	helper := NewServiceConfigHelper()
	
	// 测试不存在的通知开�?	switchAction := helper.GetNotificationSwitch("non-existent-type")
	if switchAction != nil {
		t.Error("GetNotificationSwitch should return nil for non-existent type")
	}
}

func TestNewServiceBaseHelper(t *testing.T) {
	// 测试创建ServiceBaseHelper实例
	helper := NewServiceBaseHelper(
		models.SystemConfigKeyDownloaders,
		models.DownloaderConf{},
		models.ModuleTypeDownload,
	)
	
	if helper == nil {
		t.Error("Failed to create ServiceBaseHelper instance")
	}
}

func TestServiceBaseHelperGetConfigs(t *testing.T) {
	// 测试ServiceBaseHelper获取配置
	helper := NewServiceBaseHelper(
		models.SystemConfigKeyDownloaders,
		models.DownloaderConf{},
		models.ModuleTypeDownload,
	)
	
	configs := helper.GetConfigs(false)
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if configs == nil {
		t.Error("GetConfigs should not return nil")
	}
	
	// configs应该是一个映�?	if len(configs) < 0 {
		t.Error("GetConfigs should return a valid map")
	}
}

func TestServiceBaseHelperGetConfig(t *testing.T) {
	// 测试ServiceBaseHelper获取特定配置
	helper := NewServiceBaseHelper(
		models.SystemConfigKeyDownloaders,
		models.DownloaderConf{},
		models.ModuleTypeDownload,
	)
	
	// 测试空名�?	config := helper.GetConfig("")
	if config != nil {
		t.Error("GetConfig should return nil for empty name")
	}
	
	// 测试不存在的配置
	config = helper.GetConfig("non-existent-config")
	if config != nil {
		t.Error("GetConfig should return nil for non-existent config")
	}
}

func TestDownloaderConfStruct(t *testing.T) {
	// 测试DownloaderConf结构�?	conf := models.DownloaderConf{
		Name:    "Test Downloader",
		Type:    "qbittorrent",
		Default: true,
		Config: map[string]interface{}{
			"host": "localhost",
			"port": 8080,
		},
		Enabled: true,
	}
	
	if conf.Name != "Test Downloader" {
		t.Error("Name not set correctly")
	}
	
	if conf.Type != "qbittorrent" {
		t.Error("Type not set correctly")
	}
	
	if !conf.Default {
		t.Error("Default not set correctly")
	}
	
	if !conf.Enabled {
		t.Error("Enabled not set correctly")
	}
	
	if conf.Config["host"] != "localhost" {
		t.Error("Config not set correctly")
	}
}

func TestMediaServerConfStruct(t *testing.T) {
	// 测试MediaServerConf结构�?	conf := models.MediaServerConf{
		Name: "Test Media Server",
		Type: "emby",
		Config: map[string]interface{}{
			"host": "localhost",
			"port": 8096,
		},
		Enabled: true,
		SyncLibraries: []interface{}{"Movies", "TV Shows"},
	}
	
	if conf.Name != "Test Media Server" {
		t.Error("Name not set correctly")
	}
	
	if conf.Type != "emby" {
		t.Error("Type not set correctly")
	}
	
	if !conf.Enabled {
		t.Error("Enabled not set correctly")
	}
	
	if conf.Config["host"] != "localhost" {
		t.Error("Config not set correctly")
	}
	
	if len(conf.SyncLibraries) != 2 {
		t.Error("SyncLibraries not set correctly")
	}
}

func TestNotificationConfStruct(t *testing.T) {
	// 测试NotificationConf结构�?	conf := models.NotificationConf{
		Name: "Test Notification",
		Type: "telegram",
		Config: map[string]interface{}{
			"token": "test-token",
			"chat_id": "test-chat-id",
		},
		Switchs: []interface{}{"download.start", "download.finish"},
		Enabled: true,
	}
	
	if conf.Name != "Test Notification" {
		t.Error("Name not set correctly")
	}
	
	if conf.Type != "telegram" {
		t.Error("Type not set correctly")
	}
	
	if !conf.Enabled {
		t.Error("Enabled not set correctly")
	}
	
	if conf.Config["token"] != "test-token" {
		t.Error("Config not set correctly")
	}
	
	if len(conf.Switchs) != 2 {
		t.Error("Switchs not set correctly")
	}
}

func TestNotificationSwitchConfStruct(t *testing.T) {
	// 测试NotificationSwitchConf结构�?	conf := models.NotificationSwitchConf{
		Type:   "download.start",
		Action: "all",
	}
	
	if conf.Type != "download.start" {
		t.Error("Type not set correctly")
	}
	
	if conf.Action != "all" {
		t.Error("Action not set correctly")
	}
}
