package redis

import (
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// RedisModule Redis模块
type RedisModule struct {
	modules.ModuleBase
	redisHelper *RedisHelper
}

// NewRedisModule 创建Redis模块实例
func NewRedisModule() *RedisModule {
	return &RedisModule{
		redisHelper: NewRedisHelper(),
	}
}

// InitModule 初始化模�?func (rm *RedisModule) InitModule() error {
	// 初始化模�?	return nil
}

// GetName 获取模块名称
func (rm *RedisModule) GetName() string {
	return "Redis缓存"
}

// GetType 获取模块类型
func (rm *RedisModule) GetType() models.ModuleType {
	return models.ModuleTypeOther
}

// GetSubtype 获取模块子类�?func (rm *RedisModule) GetSubtype() string {
	return "Redis"
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (rm *RedisModule) GetPriority() int {
	return 0
}

// InitSetting 初始化设�?func (rm *RedisModule) InitSetting() (string, interface{}) {
	return "", nil
}

// Stop 停止模块
func (rm *RedisModule) Stop() {
	if rm.redisHelper != nil {
		rm.redisHelper.Close()
	}
}

// Test 测试模块连接�?func (rm *RedisModule) Test() (bool, string) {
	// 检查缓存后端类型是否为Redis
	if utils.Config.CACHE_BACKEND_TYPE != "redis" {
		return true, "" // 如果不是Redis后端，返回成功但不测�?	}
	
	if rm.redisHelper.Test() {
		return true, ""
	}
	return false, "Redis连接失败，请检查配�?
}
