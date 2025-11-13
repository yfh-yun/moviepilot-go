package service

import (
	"moviepilot-go/internal/service"
	"moviepilot-go/pkg/models"
)

// MediaServerHelper 媒体服务器帮助类
type MediaServerHelper struct {
	*service.ServiceBaseHelper
}

// NewMediaServerHelper 创建媒体服务器帮助类实例
func NewMediaServerHelper() *MediaServerHelper {
	helper := &MediaServerHelper{
		ServiceBaseHelper: service.NewServiceBaseHelper(),
	}

	// 初始化基类配�?	helper.ConfigKey = "MediaServers"
	helper.ConfType = &models.MediaServerConf{}
	helper.ModuleType = "mediaserver"

	return helper
}

// IsMediaServer 通用的媒体服务器类型判断方法
func (mh *MediaServerHelper) IsMediaServer(serviceType string, serviceInfo *models.ServiceInfo, name string) bool {
	/*
		通用的媒体服务器类型判断方法
		:param serviceType: 媒体服务器的类型名称（如 'plex', 'emby', 'jellyfin'�?		:param serviceInfo: 要判断的服务信息
		:param name: 服务的名�?		:return: 如果服务类型或实例为指定类型，返�?true；否则返�?false
	*/

	// 如果未提�?service 则通过 name 获取服务
	if serviceInfo == nil {
		serviceInfo = mh.GetService(name)
	}

	// 判断服务类型是否为指定类�?	return serviceInfo != nil && serviceInfo.Type == serviceType
}
