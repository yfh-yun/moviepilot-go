package modules

// MediaServerBase 媒体服务器基�?type MediaServerBase struct {
	*ServiceBase
}

// NewMediaServerBase 创建一个新的媒体服务器基类实例
func NewMediaServerBase() *MediaServerBase {
	return &MediaServerBase{
		ServiceBase: NewServiceBase(),
	}
}

// GetConfigs 获取已启用的媒体服务器的配置字典
func (m *MediaServerBase) GetConfigs() map[string]interface{} {
	// 注意：这里需要调用ServiceConfigHelper.get_mediaserver_configs()
	// 由于这是Python代码，我们需要在Go中实现相应的功能
	// configs := ServiceConfigHelper.GetMediaServerConfigs()
	
	if m.ServiceName == "" {
		return make(map[string]interface{})
	}
	
	// 过滤出启用的配置
	enabledConfigs := make(map[string]interface{})
	// for _, conf := range configs {
	//     if conf.Type == m.ServiceName && conf.Enabled {
	//         enabledConfigs[conf.Name] = conf
	//     }
	// }
	
	return enabledConfigs
}
