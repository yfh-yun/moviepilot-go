// Package core 核心系统模块
package core

import (
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/context"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/events"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/meta"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/security"
	"time"
)

// Subpackages 导出子包
import (
	_ "github.com/yfh-yun/moviepilot-go/internal/infrastructure/context"
	_ "github.com/yfh-yun/moviepilot-go/internal/infrastructure/events"
	_ "github.com/yfh-yun/moviepilot-go/internal/infrastructure/meta"
	_ "github.com/yfh-yun/moviepilot-go/internal/infrastructure/security"
)

// CoreSystem 核心系统
type CoreSystem struct {
	// 安全组件
	JWTManager     *security.JWTManager
	PasswordManager *security.PasswordManager
	APIKeyManager  *security.APIKeyManager
	AESEncryptor   *security.AESEncryptor
	RSAEncryptor   *security.RSAEncryptor

	// 上下文组件
	ContextManager *context.ContextManager

	// 事件组件
	EventManager   *event.EventManager

	// 元数据组件
	MetaParser     *meta.MetaParser

	// 配置
	config *CoreConfig
}

// CoreConfig 核心系统配置
type CoreConfig struct {
	// JWT配置
	JWTSecret     string        `json:"jwt_secret"`
	TokenDuration time.Duration `json:"token_duration"`

	// 加密配置
	AESKey        string `json:"aes_key"`
	RSABits       int    `json:"rsa_bits"`

	// 事件系统配置
	EventQueueSize    int           `json:"event_queue_size"`
	EventWorkerCount  int           `json:"event_worker_count"`
	EventBatchSize    int           `json:"event_batch_size"`
	EventBatchTimeout time.Duration `json:"event_batch_timeout"`

	// 会话配置
	SessionTimeout time.Duration `json:"session_timeout"`
}

// DefaultCoreConfig 默认核心配置
func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		JWTSecret:        "moviepilot-default-secret-key",
		TokenDuration:    24 * time.Hour,
		AESKey:          "moviepilot-default-aes-key",
		RSABits:         2048,
		EventQueueSize:  1000,
		EventWorkerCount: 4,
		EventBatchSize:   10,
		EventBatchTimeout: 100 * time.Millisecond,
		SessionTimeout:  24 * time.Hour,
	}
}

// NewCoreSystem 创建核心系统
func NewCoreSystem(config *CoreConfig) (*CoreSystem, error) {
	if config == nil {
		config = DefaultCoreConfig()
	}

	core := &CoreSystem{
		config: config,
	}

	// 初始化安全组件
	core.initSecurityComponents()

	// 初始化上下文管理器
	core.ContextManager = context.NewContextManager()

	// 初始化事件管理器
	core.initEventManager()

	// 初始化元数据解析器
	core.MetaParser = meta.NewMetaParser()

	return core, nil
}

// initSecurityComponents 初始化安全组件
func (cs *CoreSystem) initSecurityComponents() {
	// JWT管理器
	cs.JWTManager = security.NewJWTManager(cs.config.JWTSecret, cs.config.TokenDuration)

	// 密码管理器
	cs.PasswordManager = security.NewPasswordManager()

	// API密钥管理器
	cs.APIKeyManager = security.NewAPIKeyManager()

	// AES加密器
	cs.AESEncryptor = security.NewAESEncryptor(cs.config.AESKey)

	// RSA加密器
	rsaKeyPair, err := security.GenerateRSAKeyPair(cs.config.RSABits)
	if err != nil {
		// 如果生成失败，使用默认的空加密器
		cs.RSAEncryptor = nil
	} else {
		cs.RSAEncryptor = security.NewRSAEncryptorFromKeys(rsaKeyPair.PublicKey, rsaKeyPair.PrivateKey)
	}
}

// initEventManager 初始化事件管理器
func (cs *CoreSystem) initEventManager() {
	eventConfig := &event.EventManagerConfig{
		QueueSize:    cs.config.EventQueueSize,
		WorkerCount:  cs.config.EventWorkerCount,
		BatchSize:    cs.config.EventBatchSize,
		BatchTimeout: cs.config.EventBatchTimeout,
		RetryPolicy:  event.DefaultRetryPolicy(),
	}

	cs.EventManager = event.NewEventManager(eventConfig)
}

// Start 启动核心系统
func (cs *CoreSystem) Start() error {
	// 启动事件管理器
	if err := cs.EventManager.Start(); err != nil {
		return err
	}

	return nil
}

// Stop 停止核心系统
func (cs *CoreSystem) Stop() error {
	// 停止事件管理器
	if err := cs.EventManager.Stop(); err != nil {
		return err
	}

	// 清理过期会话
	cs.ContextManager.CleanupExpiredSessions()

	return nil
}

// GetConfig 获取配置
func (cs *CoreSystem) GetConfig() *CoreConfig {
	return cs.config
}

// UpdateConfig 更新配置
func (cs *CoreSystem) UpdateConfig(config *CoreConfig) {
	cs.config = config
	// 重新初始化组件
	cs.initSecurityComponents()
}

// GetStatistics 获取系统统计信息
func (cs *CoreSystem) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// 事件系统统计
	stats["event"] = cs.EventManager.GetStatistics()

	// 会话统计
	stats["sessions"] = map[string]interface{}{
		"count": cs.ContextManager.GetSessionCount(),
	}

	return stats
}

// HealthCheck 健康检查
func (cs *CoreSystem) HealthCheck() map[string]bool {
	health := make(map[string]bool)

	// 检查事件管理器
	health["event_manager"] = cs.EventManager.IsRunning()

	// 检查安全组件
	health["jwt_manager"] = cs.JWTManager != nil
	health["password_manager"] = cs.PasswordManager != nil
	health["api_key_manager"] = cs.APIKeyManager != nil
	health["aes_encryptor"] = cs.AESEncryptor != nil

	// 检查其他组件
	health["context_manager"] = cs.ContextManager != nil
	health["meta_parser"] = cs.MetaParser != nil

	return health
}