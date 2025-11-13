package modules

// DownloaderBase 下载器基�?type DownloaderBase struct {
	*ServiceBase
	// 默认配置名称
	DefaultConfigName *string
}

// NewDownloaderBase 创建一个新的下载器基类实例
func NewDownloaderBase() *DownloaderBase {
	return &DownloaderBase{
		ServiceBase:       NewServiceBase(),
		DefaultConfigName: nil,
	}
}

// GetDefaultConfigName 获取默认服务配置的名�?// 优先从所有下载器中查找配置了默认的下载器，如果没有配置，则获取第一个下载器名称
func (d *DownloaderBase) GetDefaultConfigName() *string {
	// 优先查找默认配置
	if d.DefaultConfigName != nil {
		return d.DefaultConfigName
	}
	
	// 注意：这里需要调用ServiceConfigHelper.get_downloader_configs()
	// 由于这是Python代码，我们需要在Go中实现相应的功能
	// configs := ServiceConfigHelper.GetDownloaderConfigs()
	
	// 查找默认配置
	// for _, conf := range configs {
	//     if conf.Default {
	//         d.DefaultConfigName = &conf.Name
	//         return d.DefaultConfigName
	//     }
	// }
	
	// 如果没有默认配置，返回第一个配置的名称
	// if len(configs) > 0 {
	//     d.DefaultConfigName = &configs[0].Name
	//     return d.DefaultConfigName
	// }
	
	return nil
}

// GetConfigs 获取已启用的下载器的配置字典
func (d *DownloaderBase) GetConfigs() map[string]interface{} {
	// 注意：这里需要调用ServiceConfigHelper.get_downloader_configs()
	// 由于这是Python代码，我们需要在Go中实现相应的功能
	// configs := ServiceConfigHelper.GetDownloaderConfigs()
	
	if d.ServiceName == "" {
		return make(map[string]interface{})
	}
	
	// 过滤出启用的配置
	enabledConfigs := make(map[string]interface{})
	// for _, conf := range configs {
	//     if conf.Type == d.ServiceName && conf.Enabled {
	//         enabledConfigs[conf.Name] = conf
	//     }
	// }
	
	return enabledConfigs
}
