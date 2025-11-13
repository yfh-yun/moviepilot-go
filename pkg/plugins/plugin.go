package plugins

import (
	"moviepilot-go/pkg/models"
)

// Plugin 插件接口定义
type Plugin interface {
	// InitPlugin 生效配置信息
	InitPlugin(config map[string]interface{}) error
	
	// GetName 获取插件名称
	GetName() string
	
	// GetDesc 获取插件描述
	GetDesc() string
	
	// GetOrder 获取插件顺序
	GetOrder() int
	
	// GetState 获取插件运行状�?	GetState() bool
	
	// GetCommand 注册插件远程命令
	GetCommand() []map[string]interface{}
	
	// GetRenderMode 获取插件渲染模式
	GetRenderMode() (string, *string)
	
	// GetAPI 注册插件API
	GetAPI() []map[string]interface{}
	
	// GetForm 拼装插件配置页面
	GetForm() ([]map[string]interface{}, map[string]interface{})
	
	// GetPage 拼装插件详情页面
	GetPage() []map[string]interface{}
	
	// GetService 注册插件公共服务
	GetService() []map[string]interface{}
	
	// GetDashboard 获取插件仪表盘页�?	GetDashboard(key string, kwargs map[string]interface{}) (*map[string]interface{}, *map[string]interface{}, *[]map[string]interface{})
	
	// GetDashboardMeta 获取插件仪表盘元信息
	GetDashboardMeta() []map[string]string
	
	// GetModule 获取插件模块声明
	GetModule() map[string]interface{}
	
	// GetActions 获取插件工作流动�?	GetActions() []map[string]interface{}
	
	// StopService 停止插件
	StopService()
	
	// UpdateConfig 更新配置信息
	UpdateConfig(config map[string]interface{}, pluginID *string) bool
	
	// GetConfig 获取配置信息
	GetConfig(pluginID *string) interface{}
	
	// GetDataPath 获取插件数据保存目录
	GetDataPath(pluginID *string) string
	
	// SaveData 保存插件数据
	SaveData(key string, value interface{}, pluginID *string)
	
	// GetData 获取插件数据
	GetData(key *string, pluginID *string) interface{}
	
	// DelData 删除插件数据
	DelData(key string, pluginID *string) interface{}
	
	// PostMessage 发送消�?	PostMessage(channel *models.MessageChannel, mtype *models.NotificationType,
		title, text, image, link, userid, username *string, kwargs map[string]interface{})
	
	// Close 关闭插件
	Close()
}
