package utils

import (
	"fmt"
	"sync"
)

// NotificationHelper 消息通知帮助类
type NotificationHelper struct {
	services map[string]*NotificationService
	mutex    sync.RWMutex
}

// NotificationService 消息通知服务信息
type NotificationService struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
}

// NotificationConfig 消息通知配置
type NotificationConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	Webhook    string                 `json:"webhook,omitempty"`
	Token      string                 `json:"token,omitempty"`
	ChatID     string                 `json:"chat_id,omitempty"`
	APIKey     string                 `json:"api_key,omitempty"`
	Secret     string                 `json:"secret,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Password   string                 `json:"password,omitempty"`
	From       string                 `json:"from,omitempty"`
	To         []string               `json:"to,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image,omitempty"`
	Link    string `json:"link,omitempty"`
	Level   string `json:"level,omitempty"` // info, warning, error, success
}

// NewNotificationHelper 创建消息通知助手实例
func NewNotificationHelper() *NotificationHelper {
	return &NotificationHelper{
		services: make(map[string]*NotificationService),
	}
}

// AddService 添加消息通知服务
func (nh *NotificationHelper) AddService(service *NotificationService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service.Type == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	nh.mutex.Lock()
	defer nh.mutex.Unlock()

	nh.services[service.Name] = service
	return nil
}

// RemoveService 移除消息通知服务
func (nh *NotificationHelper) RemoveService(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	nh.mutex.Lock()
	defer nh.mutex.Unlock()

	if _, exists := nh.services[name]; !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	delete(nh.services, name)
	return nil
}

// GetService 获取消息通知服务
func (nh *NotificationHelper) GetService(name string) (*NotificationService, error) {
	if name == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	service, exists := nh.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	return service, nil
}

// GetAllServices 获取所有消息通知服务
func (nh *NotificationHelper) GetAllServices() []*NotificationService {
	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	services := make([]*NotificationService, 0, len(nh.services))
	for _, service := range nh.services {
		services = append(services, service)
	}

	return services
}

// GetEnabledServices 获取启用的消息通知服务
func (nh *NotificationHelper) GetEnabledServices() []*NotificationService {
	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	var enabledServices []*NotificationService
	for _, service := range nh.services {
		if service.Enabled {
			enabledServices = append(enabledServices, service)
		}
	}

	return enabledServices
}

// IsNotification 判断是否为指定类型的消息通知服务
func (nh *NotificationHelper) IsNotification(serviceType string, service *NotificationService, name string) bool {
	if service == nil && name != "" {
		s, err := nh.GetService(name)
		if err != nil {
			return false
		}
		service = s
	}

	return service != nil && service.Type == serviceType
}

// IsWechat 判断是否为微信通知
func (nh *NotificationHelper) IsWechat(service *NotificationService, name string) bool {
	return nh.IsNotification("wechat", service, name)
}

// IsTelegram 判断是否为Telegram通知
func (nh *NotificationHelper) IsTelegram(service *NotificationService, name string) bool {
	return nh.IsNotification("telegram", service, name)
}

// IsServerChan 判断是否为Server酱通知
func (nh *NotificationHelper) IsServerChan(service *NotificationService, name string) bool {
	return nh.IsNotification("serverchan", service, name)
}

// IsBark 判断是否为Bark通知
func (nh *NotificationHelper) IsBark(service *NotificationService, name string) bool {
	return nh.IsNotification("bark", service, name)
}

// IsPushPlus 判断是否为PushPlus通知
func (nh *NotificationHelper) IsPushPlus(service *NotificationService, name string) bool {
	return nh.IsNotification("pushplus", service, name)
}

// IsEmail 判断是否为邮件通知
func (nh *NotificationHelper) IsEmail(service *NotificationService, name string) bool {
	return nh.IsNotification("email", service, name)
}

// GetServiceByType 根据类型获取消息通知服务
func (nh *NotificationHelper) GetServiceByType(serviceType string) []*NotificationService {
	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	var services []*NotificationService
	for _, service := range nh.services {
		if service.Type == serviceType {
			services = append(services, service)
		}
	}

	return services
}

// GetServiceByPriority 根据优先级获取消息通知服务
func (nh *NotificationHelper) GetServiceByPriority() []*NotificationService {
	services := nh.GetEnabledServices()
	
	// 按优先级排序（优先级数字越小，优先级越高）
	for i := 0; i < len(services)-1; i++ {
		for j := i + 1; j < len(services); j++ {
			if services[i].Priority > services[j].Priority {
				services[i], services[j] = services[j], services[i]
			}
		}
	}

	return services
}

// UpdateService 更新消息通知服务
func (nh *NotificationHelper) UpdateService(name string, updates map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	nh.mutex.Lock()
	defer nh.mutex.Unlock()

	service, exists := nh.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if newName, ok := value.(string); ok {
				delete(nh.services, name)
				service.Name = newName
				nh.services[newName] = service
			}
		case "type":
			if serviceType, ok := value.(string); ok {
				service.Type = serviceType
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				service.Enabled = enabled
			}
		case "priority":
			if priority, ok := value.(int); ok {
				service.Priority = priority
			}
		case "config":
			if config, ok := value.(map[string]interface{}); ok {
				service.Config = config
			}
		}
	}

	return nil
}

// EnableService 启用消息通知服务
func (nh *NotificationHelper) EnableService(name string) error {
	return nh.UpdateService(name, map[string]interface{}{"enabled": true})
}

// DisableService 禁用消息通知服务
func (nh *NotificationHelper) DisableService(name string) error {
	return nh.UpdateService(name, map[string]interface{}{"enabled": false})
}

// ValidateConfig 验证消息通知配置
func (nh *NotificationHelper) ValidateConfig(config *NotificationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("notification type cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("notification name cannot be empty")
	}

	// 根据类型进行特定验证
	switch config.Type {
	case "wechat":
		return nh.validateWechatConfig(config)
	case "telegram":
		return nh.validateTelegramConfig(config)
	case "serverchan":
		return nh.validateServerChanConfig(config)
	case "bark":
		return nh.validateBarkConfig(config)
	case "pushplus":
		return nh.validatePushPlusConfig(config)
	case "email":
		return nh.validateEmailConfig(config)
	default:
		return fmt.Errorf("unsupported notification type: %s", config.Type)
	}
}

// validateWechatConfig 验证微信配置
func (nh *NotificationHelper) validateWechatConfig(config *NotificationConfig) error {
	if config.Webhook == "" {
		return fmt.Errorf("wechat webhook URL is required")
	}
	return nil
}

// validateTelegramConfig 验证Telegram配置
func (nh *NotificationHelper) validateTelegramConfig(config *NotificationConfig) error {
	if config.Token == "" {
		return fmt.Errorf("telegram bot token is required")
	}
	if config.ChatID == "" {
		return fmt.Errorf("telegram chat ID is required")
	}
	return nil
}

// validateServerChanConfig 验证Server酱配置
func (nh *NotificationHelper) validateServerChanConfig(config *NotificationConfig) error {
	if config.Token == "" {
		return fmt.Errorf("serverchan sendkey is required")
	}
	return nil
}

// validateBarkConfig 验证Bark配置
func (nh *NotificationHelper) validateBarkConfig(config *NotificationConfig) error {
	if config.Webhook == "" {
		return fmt.Errorf("bark server URL is required")
	}
	if config.APIKey == "" {
		return fmt.Errorf("bark device key is required")
	}
	return nil
}

// validatePushPlusConfig 验证PushPlus配置
func (nh *NotificationHelper) validatePushPlusConfig(config *NotificationConfig) error {
	if config.Token == "" {
		return fmt.Errorf("pushplus token is required")
	}
	return nil
}

// validateEmailConfig 验证邮件配置
func (nh *NotificationHelper) validateEmailConfig(config *NotificationConfig) error {
	if config.Username == "" {
		return fmt.Errorf("email username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("email password is required")
	}
	if config.From == "" {
		return fmt.Errorf("email from address is required")
	}
	if len(config.To) == 0 {
		return fmt.Errorf("email to addresses are required")
	}
	return nil
}

// GetServiceCount 获取服务数量
func (nh *NotificationHelper) GetServiceCount() int {
	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	return len(nh.services)
}

// GetEnabledServiceCount 获取启用服务数量
func (nh *NotificationHelper) GetEnabledServiceCount() int {
	services := nh.GetEnabledServices()
	return len(services)
}

// ClearServices 清空所有服务
func (nh *NotificationHelper) ClearServices() {
	nh.mutex.Lock()
	defer nh.mutex.Unlock()

	nh.services = make(map[string]*NotificationService)
}

// ImportServices 导入服务配置
func (nh *NotificationHelper) ImportServices(configs []*NotificationConfig) error {
	if configs == nil {
		return fmt.Errorf("configs cannot be nil")
	}

	for _, config := range configs {
		if err := nh.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid config for %s: %v", config.Name, err)
		}

		service := &NotificationService{
			Name:     config.Name,
			Type:     config.Type,
			Enabled:  config.Enabled,
			Priority: config.Priority,
			Config: map[string]interface{}{
				"webhook":  config.Webhook,
				"token":    config.Token,
				"chat_id":  config.ChatID,
				"api_key":  config.APIKey,
				"secret":   config.Secret,
				"username": config.Username,
				"password": config.Password,
				"from":     config.From,
				"to":       config.To,
			},
		}

		// 添加额外配置
		for key, value := range config.Extra {
			service.Config[key] = value
		}

		if err := nh.AddService(service); err != nil {
			return fmt.Errorf("failed to add service %s: %v", config.Name, err)
		}
	}

	return nil
}

// ExportServices 导出服务配置
func (nh *NotificationHelper) ExportServices() []*NotificationConfig {
	services := nh.GetAllServices()
	configs := make([]*NotificationConfig, 0, len(services))

	for _, service := range services {
		config := &NotificationConfig{
			Type:     service.Type,
			Name:     service.Name,
			Enabled:  service.Enabled,
			Priority: service.Priority,
			Extra:    make(map[string]interface{}),
		}

		// 从配置中提取基本信息
		if webhook, ok := service.Config["webhook"].(string); ok {
			config.Webhook = webhook
		}
		if token, ok := service.Config["token"].(string); ok {
			config.Token = token
		}
		if chatID, ok := service.Config["chat_id"].(string); ok {
			config.ChatID = chatID
		}
		if apiKey, ok := service.Config["api_key"].(string); ok {
			config.APIKey = apiKey
		}
		if secret, ok := service.Config["secret"].(string); ok {
			config.Secret = secret
		}
		if username, ok := service.Config["username"].(string); ok {
			config.Username = username
		}
		if password, ok := service.Config["password"].(string); ok {
			config.Password = password
		}
		if from, ok := service.Config["from"].(string); ok {
			config.From = from
		}
		if to, ok := service.Config["to"].([]string); ok {
			config.To = to
		}

		// 添加额外配置
		for key, value := range service.Config {
			if key != "webhook" && key != "token" && key != "chat_id" && 
			   key != "api_key" && key != "secret" && key != "username" && 
			   key != "password" && key != "from" && key != "to" {
				config.Extra[key] = value
			}
		}

		configs = append(configs, config)
	}

	return configs
}

// SendMessage 发送消息
func (nh *NotificationHelper) SendMessage(message *NotificationMessage) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	services := nh.GetEnabledServices()
	if len(services) == 0 {
		return fmt.Errorf("no enabled notification service found")
	}

	var lastError error
	for _, service := range services {
		if err := nh.sendToService(service, message); err != nil {
			lastError = err
			continue
		}
		// 如果发送成功，返回nil（至少有一个服务发送成功）
		return nil
	}

	return fmt.Errorf("failed to send message to all services: %v", lastError)
}

// sendToService 发送消息到指定服务
func (nh *NotificationHelper) sendToService(service *NotificationService, message *NotificationMessage) error {
	// 这里是简化的发送逻辑，实际实现需要根据不同服务类型调用相应的API
	switch service.Type {
	case "wechat":
		return nh.sendWechatMessage(service, message)
	case "telegram":
		return nh.sendTelegramMessage(service, message)
	case "serverchan":
		return nh.sendServerChanMessage(service, message)
	case "bark":
		return nh.sendBarkMessage(service, message)
	case "pushplus":
		return nh.sendPushPlusMessage(service, message)
	case "email":
		return nh.sendEmailMessage(service, message)
	default:
		return fmt.Errorf("unsupported notification type: %s", service.Type)
	}
}

// 各种服务的发送方法（简化实现）
func (nh *NotificationHelper) sendWechatMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要调用微信API
	return fmt.Errorf("wechat notification not implemented")
}

func (nh *NotificationHelper) sendTelegramMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要调用Telegram Bot API
	return fmt.Errorf("telegram notification not implemented")
}

func (nh *NotificationHelper) sendServerChanMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要调用Server酱API
	return fmt.Errorf("serverchan notification not implemented")
}

func (nh *NotificationHelper) sendBarkMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要调用Bark API
	return fmt.Errorf("bark notification not implemented")
}

func (nh *NotificationHelper) sendPushPlusMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要调用PushPlus API
	return fmt.Errorf("pushplus notification not implemented")
}

func (nh *NotificationHelper) sendEmailMessage(service *NotificationService, message *NotificationMessage) error {
	// 实际实现需要发送邮件
	return fmt.Errorf("email notification not implemented")
}