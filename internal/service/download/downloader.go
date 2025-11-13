package service

import (
	"moviepilot-go/internal/service"
	"moviepilot-go/pkg/models"
)

// DownloaderHelper 下载器帮助类
type DownloaderHelper struct {
	*service.ServiceBaseHelper
}

// NewDownloaderHelper 创建下载器帮助类实例
func NewDownloaderHelper() *DownloaderHelper {
	helper := &DownloaderHelper{
		ServiceBaseHelper: service.NewServiceBaseHelper(),
	}
	
	// 初始化基类配�?	helper.ConfigKey = "downloaders"
	helper.ConfType = &models.DownloaderConf{}
	helper.ModuleType = "downloader"
	
	return helper
}

// IsDownloader 通用的下载器类型判断方法
func (dh *DownloaderHelper) IsDownloader(serviceType string, serviceInfo *models.ServiceInfo, name string) bool {
	/*
		通用的下载器类型判断方法
		:param serviceType: 下载器的类型名称（如 'qbittorrent', 'transmission'�?		:param serviceInfo: 要判断的服务信息
		:param name: 服务的名�?		:return: 如果服务类型或实例为指定类型，返�?true；否则返�?false
	*/
	
	// 如果未提�?service 则通过 name 获取服务
	if serviceInfo == nil {
		serviceInfo = dh.GetService(name)
	}
	
	// 判断服务类型是否为指定类�?	return serviceInfo != nil && serviceInfo.Type == serviceType
}
