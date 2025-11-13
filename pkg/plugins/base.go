package plugins

import (
	"path/filepath"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/db"
	"moviepilot-go/internal/core"
	"moviepilot-go/internal/helper"
	"moviepilot-go/pkg/models"
)

// PluginBase 插件模块基类，通过继承该类实现插件功能
type PluginBase struct {
	// 插件名称
	PluginName string
	// 插件描述
	PluginDesc string
	// 插件顺序
	PluginOrder int
	// 是否为插件分�?	IsClone bool
	
	// 插件数据操作
	PluginData *db.PluginDataOper
	// 处理�?	Chain *PluginChain
	// 系统配置
	SystemConfig *core.SystemConfigOper
	// 系统消息
	SystemMessage *helper.MessageHelper
	// 事件管理�?	EventManager *core.EventManager
}

// NewPluginBase 创建一个新的插件基类实�?func NewPluginBase() *PluginBase {
	return &PluginBase{
		PluginName:    "",
		PluginDesc:    "",
		PluginOrder:   9999,
		IsClone:       false,
		PluginData:    db.NewPluginDataOper(),
		Chain:         NewPluginChain(),
		SystemConfig:  core.NewSystemConfigOper(),
		SystemMessage: helper.NewMessageHelper(),
		EventManager:  core.NewEventManager(),
	}
}

// InitPlugin 生效配置信息
// 子类需要实现此方法
func (p *PluginBase) InitPlugin(config map[string]interface{}) {
	// 子类需要实现此方法
}

// GetName 获取插件名称
func (p *PluginBase) GetName() string {
	return p.PluginName
}

// GetState 获取插件运行状�?// 子类需要实现此方法
func (p *PluginBase) GetState() bool {
	// 子类需要实现此方法
	return false
}

// GetCommand 注册插件远程命令
func (p *PluginBase) GetCommand() []map[string]interface{} {
	// 子类可以重写此方�?	return nil
}

// GetRenderMode 获取插件渲染模式
// 返回: 1、渲染模式，支持：vue/vuetify，默认vuetify�?、vue模式下编译后文件的相对路径，默认为`dist/asserts`，vuetify模式下为None
func (p *PluginBase) GetRenderMode() (string, *string) {
	return "vuetify", nil
}

// GetAPI 注册插件API
// 子类需要实现此方法
func (p *PluginBase) GetAPI() []map[string]interface{} {
	// 子类需要实现此方法
	return nil
}

// GetForm 拼装插件配置页面
// 子类需要实现此方法
func (p *PluginBase) GetForm() ([]map[string]interface{}, map[string]interface{}) {
	// 子类需要实现此方法
	return nil, nil
}

// GetPage 拼装插件详情页面
// 子类需要实现此方法
func (p *PluginBase) GetPage() []map[string]interface{} {
	// 子类需要实现此方法
	return nil
}

// GetService 注册插件公共服务
func (p *PluginBase) GetService() []map[string]interface{} {
	// 子类可以重写此方�?	return nil
}

// GetDashboard 获取插件仪表盘页�?func (p *PluginBase) GetDashboard(key string, kwargs map[string]interface{}) (*map[string]interface{}, *map[string]interface{}, *[]map[string]interface{}) {
	// 子类可以重写此方�?	return nil, nil, nil
}

// GetDashboardMeta 获取插件仪表盘元信息
func (p *PluginBase) GetDashboardMeta() []map[string]string {
	// 子类可以重写此方�?	return nil
}

// GetModule 获取插件模块声明，用于胁持系统模块实�?func (p *PluginBase) GetModule() map[string]interface{} {
	// 子类可以重写此方�?	return nil
}

// GetActions 获取插件工作流动�?func (p *PluginBase) GetActions() []map[string]interface{} {
	// 子类可以重写此方�?	return nil
}

// StopService 停止插件
// 子类需要实现此方法
func (p *PluginBase) StopService() {
	// 子类需要实现此方法
}

// UpdateConfig 更新配置信息
func (p *PluginBase) UpdateConfig(config map[string]interface{}, pluginID *string) bool {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	key := "plugin." + id
	return p.SystemConfig.Set(key, config)
}

// GetConfig 获取配置信息
func (p *PluginBase) GetConfig(pluginID *string) interface{} {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	key := "plugin." + id
	return p.SystemConfig.Get(key)
}

// GetDataPath 获取插件数据保存目录
func (p *PluginBase) GetDataPath(pluginID *string) string {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	dataPath := filepath.Join(config.Settings.PluginDataPath, id)
	// 确保目录存在
	// os.MkdirAll(dataPath, 0755)
	return dataPath
}

// SaveData 保存插件数据
func (p *PluginBase) SaveData(key string, value interface{}, pluginID *string) {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	p.PluginData.Save(id, key, value)
}

// GetData 获取插件数据
func (p *PluginBase) GetData(key *string, pluginID *string) interface{} {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	k := ""
	if key != nil {
		k = *key
	}
	
	return p.PluginData.GetData(id, k)
}

// DelData 删除插件数据
func (p *PluginBase) DelData(key string, pluginID *string) interface{} {
	id := p.getClassName()
	if pluginID != nil {
		id = *pluginID
	}
	
	return p.PluginData.DelData(id, key)
}

// PostMessage 发送消�?func (p *PluginBase) PostMessage(channel *models.MessageChannel, mtype *models.NotificationType, 
	title, text, image, link, userid, username *string, kwargs map[string]interface{}) {
	
	// 如果没有提供链接，则使用默认链接
	l := ""
	if link == nil {
		l = config.Settings.MPDomain("#/plugins?tab=installed&id=" + p.getClassName())
	} else {
		l = *link
	}
	
	notification := &models.Notification{
		Channel:  channel,
		MType:    mtype,
		Title:    title,
		Text:     text,
		Image:    image,
		Link:     &l,
		UserID:   userid,
		Username: username,
	}
	
	// 合并kwargs中的额外参数
	if kwargs != nil {
		// 处理额外参数
	}
	
	p.Chain.PostMessage(notification)
}

// Close 关闭插件
func (p *PluginBase) Close() {
	// 清理资源
}

// getClassName 获取类名（在Go中模拟）
func (p *PluginBase) getClassName() string {
	// 在Go中，我们无法直接获取类型名，需要在具体插件中重写此方法
	return "PluginBase"
}
