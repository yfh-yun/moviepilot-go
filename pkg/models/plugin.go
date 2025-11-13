package models

// Plugin 插件信息
type Plugin struct {
	// ID
	ID string `json:"id,omitempty"`
	// 插件名称
	PluginName string `json:"plugin_name,omitempty"`
	// 插件描述
	PluginDesc string `json:"plugin_desc,omitempty"`
	// 插件图标
	PluginIcon string `json:"plugin_icon,omitempty"`
	// 插件版本
	PluginVersion string `json:"plugin_version,omitempty"`
	// 插件标签
	PluginLabel string `json:"plugin_label,omitempty"`
	// 插件作�?	PluginAuthor string `json:"plugin_author,omitempty"`
	// 作者主�?	AuthorURL string `json:"author_url,omitempty"`
	// 插件配置项ID前缀
	PluginConfigPrefix string `json:"plugin_config_prefix,omitempty"`
	// 加载顺序
	PluginOrder int `json:"plugin_order,omitempty"`
	// 可使用的用户级别
	AuthLevel int `json:"auth_level,omitempty"`
	// 是否已安�?	Installed bool `json:"installed,omitempty"`
	// 运行状�?	State bool `json:"state,omitempty"`
	// 是否有详情页�?	HasPage bool `json:"has_page,omitempty"`
	// 是否有新版本
	HasUpdate bool `json:"has_update,omitempty"`
	// 是否本地
	IsLocal bool `json:"is_local,omitempty"`
	// 仓库地址
	RepoURL string `json:"repo_url,omitempty"`
	// 安装次数
	InstallCount int `json:"install_count,omitempty"`
	// 更新记录
	History map[string]interface{} `json:"history,omitempty"`
	// 添加时间，值越小表示越靠后发布
	AddTime int `json:"add_time,omitempty"`
	// 插件公钥
	PluginPublicKey string `json:"plugin_public_key,omitempty"`
}

// PluginDashboard 插件仪表�?type PluginDashboard struct {
	// 继承Plugin的所有字�?	Plugin
	// 名称
	Name string `json:"name,omitempty"`
	// 仪表板key
	Key string `json:"key,omitempty"`
	// 演染模式
	RenderMode string `json:"render_mode,omitempty"`
	// 全局配置
	Attrs map[string]interface{} `json:"attrs,omitempty"`
	// col列数
	Cols map[string]interface{} `json:"cols,omitempty"`
	// 页面元素
	Elements []map[string]interface{} `json:"elements,omitempty"`
}

// PluginMemoryInfo 插件内存信息
type PluginMemoryInfo struct {
	// 插件ID
	PluginID string `json:"plugin_id"`
	// 插件名称
	PluginName string `json:"plugin_name"`
	// 插件版本
	PluginVersion string `json:"plugin_version"`
	// 总内存使用量(字节)
	TotalMemoryBytes int `json:"total_memory_bytes"`
	// 总内存使用量(MB)
	TotalMemoryMB float64 `json:"total_memory_mb"`
	// 对象数量
	ObjectCount int `json:"object_count"`
	// 计算耗时(毫秒)
	CalculationTimeMS float64 `json:"calculation_time_ms"`
	// 统计时间�?	Timestamp float64 `json:"timestamp"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// 大对象详�?	ObjectDetails []map[string]interface{} `json:"object_details,omitempty"`
}

// NewPlugin 创建一个新�?Plugin 实例
func NewPlugin() *Plugin {
	return &Plugin{
		History: make(map[string]interface{}),
	}
}

// NewPluginDashboard 创建一个新�?PluginDashboard 实例
func NewPluginDashboard() *PluginDashboard {
	return &PluginDashboard{
		Plugin:     *NewPlugin(),
		RenderMode: "vuetify",
		Attrs:      make(map[string]interface{}),
		Cols:       make(map[string]interface{}),
		Elements:   make([]map[string]interface{}, 0),
	}
}

// NewPluginMemoryInfo 创建一个新�?PluginMemoryInfo 实例
func NewPluginMemoryInfo() *PluginMemoryInfo {
	return &PluginMemoryInfo{
		ObjectDetails: make([]map[string]interface{}, 0),
	}
}
