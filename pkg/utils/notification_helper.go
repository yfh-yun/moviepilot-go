package utils

import (
	"sync"

	"go.uber.org/zap"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeMovieAdd           NotificationType = "movie_add"
	NotificationTypeMovieDownload      NotificationType = "movie_download"
	NotificationTypeMovieTransfer      NotificationType = "movie_transfer"
	NotificationTypeTVAdd              NotificationType = "tv_add"
	NotificationTypeTVDownload         NotificationType = "tv_download"
	NotificationTypeTVTransfer         NotificationType = "tv_transfer"
	NotificationTypeSubscriptionAdd    NotificationType = "subscription_add"
	NotificationTypeSubscriptionUpdate NotificationType = "subscription_update"
	NotificationTypeSystemError        NotificationType = "system_error"
	NotificationTypeSystemStart        NotificationType = "system_start"
	NotificationTypeSystemStop         NotificationType = "system_stop"
	NotificationTypeTorrentAdd         NotificationType = "torrent_add"
	NotificationTypeTorrentFinish      NotificationType = "torrent_finish"
	NotificationTypeTorrentError       NotificationType = "torrent_error"
)

// NotificationChannel 通知渠道类型
type NotificationChannel string

const (
	NotificationChannelWeChat       NotificationChannel = "wechat"
	NotificationChannelTelegram     NotificationChannel = "telegram"
	NotificationChannelSlack        NotificationChannel = "slack"
	NotificationChannelSynologyChat NotificationChannel = "synologychat"
	NotificationChannelWebPush      NotificationChannel = "webpush"
	NotificationChannelVoiceChat    NotificationChannel = "voicechat"
)

// NotificationConf 通知配置
type NotificationConf struct {
	Name    string              `json:"name"`
	Type    NotificationChannel `json:"type"`
	Enabled bool                `json:"enabled"`
	Config  map[string]any      `json:"config"`
}

// NotificationSwitchConf 通知开关配置
type NotificationSwitchConf struct {
	Type   NotificationType `json:"type"`
	Action string           `json:"action"` // on, off, custom
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	Type    NotificationType `json:"type"`
	Title   string           `json:"title"`
	Content string           `json:"content"`
	Image   string           `json:"image,omitempty"`
	Data    map[string]any   `json:"data,omitempty"`
}

// NotificationService 通知服务接口
type NotificationService interface {
	// Name 返回服务名称
	Name() string
	// Type 返回服务类型
	Type() NotificationChannel
	// Send 发送通知
	Send(message NotificationMessage) error
	// IsEnabled 检查服务是否启用
	IsEnabled() bool
	// Configure 配置服务
	Configure(config NotificationConf)
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name     string              `json:"name"`
	Instance NotificationService `json:"instance"`
	Type     NotificationChannel `json:"type"`
	Config   NotificationConf    `json:"config"`
}

// NotificationHelper 通知帮助类
type NotificationHelper struct {
	logger       *zap.Logger
	services     map[string]*ServiceInfo
	serviceMutex sync.RWMutex
	configs      map[string]NotificationConf
	configMutex  sync.RWMutex
	switches     map[NotificationType]NotificationSwitchConf
	switchMutex  sync.RWMutex
}

// NewNotificationHelper 创建通知帮助类实例
func NewNotificationHelper(logger *zap.Logger) *NotificationHelper {
	return &NotificationHelper{
		logger:   logger,
		services: make(map[string]*ServiceInfo),
		configs:  make(map[string]NotificationConf),
		switches: make(map[NotificationType]NotificationSwitchConf),
	}
}

// AddService 添加通知服务
func (h *NotificationHelper) AddService(service *ServiceInfo) {
	h.serviceMutex.Lock()
	defer h.serviceMutex.Unlock()
	h.services[service.Name] = service
	h.logger.Info("添加通知服务", zap.String("name", service.Name), zap.String("type", string(service.Type)))
}

// RemoveService 移除通知服务
func (h *NotificationHelper) RemoveService(name string) {
	h.serviceMutex.Lock()
	defer h.serviceMutex.Unlock()
	if _, exists := h.services[name]; exists {
		delete(h.services, name)
		h.logger.Info("移除通知服务", zap.String("name", name))
	}
}

// GetService 获取指定名称的通知服务
func (h *NotificationHelper) GetService(name string) *ServiceInfo {
	h.serviceMutex.RLock()
	defer h.serviceMutex.RUnlock()
	return h.services[name]
}

// GetServices 获取所有通知服务
func (h *NotificationHelper) GetServices() map[string]*ServiceInfo {
	h.serviceMutex.RLock()
	defer h.serviceMutex.RUnlock()
	// 返回副本，避免并发修改问题
	services := make(map[string]*ServiceInfo)
	for k, v := range h.services {
		services[k] = v
	}
	return services
}

// GetServicesByType 根据类型获取通知服务
func (h *NotificationHelper) GetServicesByType(serviceType NotificationChannel) map[string]*ServiceInfo {
	h.serviceMutex.RLock()
	defer h.serviceMutex.RUnlock()
	services := make(map[string]*ServiceInfo)
	for name, service := range h.services {
		if service.Type == serviceType {
			services[name] = service
		}
	}
	return services
}

// IsNotification 检查是否为指定类型的通知服务
func (h *NotificationHelper) IsNotification(serviceType NotificationChannel, service *ServiceInfo, name string) bool {
	if service == nil {
		service = h.GetService(name)
	}
	return service != nil && service.Type == serviceType
}

// SetConfigs 设置通知配置
func (h *NotificationHelper) SetConfigs(configs []NotificationConf) {
	h.configMutex.Lock()
	defer h.configMutex.Unlock()
	// 清空现有配置
	h.configs = make(map[string]NotificationConf)
	// 添加新配置
	for _, config := range configs {
		h.configs[config.Name] = config
	}
	h.logger.Info("设置通知配置", zap.Int("count", len(configs)))
}

// GetConfigs 获取通知配置
func (h *NotificationHelper) GetConfigs() map[string]NotificationConf {
	h.configMutex.RLock()
	defer h.configMutex.RUnlock()
	// 返回副本，避免并发修改问题
	configs := make(map[string]NotificationConf)
	for k, v := range h.configs {
		configs[k] = v
	}
	return configs
}

// GetConfig 获取指定名称的通知配置
func (h *NotificationHelper) GetConfig(name string) *NotificationConf {
	h.configMutex.RLock()
	defer h.configMutex.RUnlock()
	if config, exists := h.configs[name]; exists {
		return &config
	}
	return nil
}

// SetSwitches 设置通知开关配置
func (h *NotificationHelper) SetSwitches(switches []NotificationSwitchConf) {
	h.switchMutex.Lock()
	defer h.switchMutex.Unlock()
	// 清空现有开关
	h.switches = make(map[NotificationType]NotificationSwitchConf)
	// 添加新开关
	for _, switchConf := range switches {
		h.switches[switchConf.Type] = switchConf
	}
	h.logger.Info("设置通知开关", zap.Int("count", len(switches)))
}

// GetSwitches 获取通知开关配置
func (h *NotificationHelper) GetSwitches() map[NotificationType]NotificationSwitchConf {
	h.switchMutex.RLock()
	defer h.switchMutex.RUnlock()
	// 返回副本，避免并发修改问题
	switches := make(map[NotificationType]NotificationSwitchConf)
	for k, v := range h.switches {
		switches[k] = v
	}
	return switches
}

// GetSwitch 获取指定类型的通知开关
func (h *NotificationHelper) GetSwitch(notificationType NotificationType) *NotificationSwitchConf {
	h.switchMutex.RLock()
	defer h.switchMutex.RUnlock()
	if switchConf, exists := h.switches[notificationType]; exists {
		return &switchConf
	}
	return nil
}

// CanSendNotification 检查是否可以发送指定类型的通知
func (h *NotificationHelper) CanSendNotification(notificationType NotificationType) bool {
	switchConf := h.GetSwitch(notificationType)
	if switchConf == nil {
		return false
	}
	return switchConf.Action == "on"
}

// SendNotification 发送通知
func (h *NotificationHelper) SendNotification(message NotificationMessage) error {
	// 检查是否可以发送该类型的通知
	if !h.CanSendNotification(message.Type) {
		h.logger.Debug("通知类型已关闭，跳过发送", zap.String("type", string(message.Type)))
		return nil
	}

	// 获取所有启用的通知服务
	services := h.GetServices()
	if len(services) == 0 {
		h.logger.Debug("没有启用的通知服务，跳过发送")
		return nil
	}

	// 并发发送通知
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for name, service := range services {
		if !service.Instance.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(name string, service *ServiceInfo) {
			defer wg.Done()
			h.logger.Debug("通过通知服务发送通知", zap.String("type", string(service.Type)), zap.String("title", message.Title))
			if err := service.Instance.Send(message); err != nil {
				h.logger.Error("通过通知服务发送通知失败", zap.String("type", string(service.Type)), zap.Error(err))
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			} else {
				h.logger.Info("通过通知服务发送通知成功", zap.String("type", string(service.Type)), zap.String("title", message.Title))
			}
		}(name, service)
	}

	wg.Wait()

	if len(errors) > 0 {
		return errors[0] // 返回第一个错误
	}

	return nil
}

// RegisterService 注册通知服务
func (h *NotificationHelper) RegisterService(service NotificationService) {
	h.serviceMutex.Lock()
	defer h.serviceMutex.Unlock()

	// 获取服务配置
	config := h.GetConfig(service.Name())
	if config != nil {
		// 配置服务
		service.Configure(*config)
	}

	// 创建服务信息
	serviceInfo := &ServiceInfo{
		Name:     service.Name(),
		Instance: service,
		Type:     service.Type(),
		Config:   *config,
	}

	// 添加到服务列表
	h.services[service.Name()] = serviceInfo
	h.logger.Info("注册通知服务", zap.String("name", service.Name()), zap.String("type", string(service.Type())))
}

// UnregisterService 注销通知服务
func (h *NotificationHelper) UnregisterService(name string) {
	h.serviceMutex.Lock()
	defer h.serviceMutex.Unlock()
	if _, exists := h.services[name]; exists {
		delete(h.services, name)
		h.logger.Info("注销通知服务", zap.String("name", name))
	}
}
