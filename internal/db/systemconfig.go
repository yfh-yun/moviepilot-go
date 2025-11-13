package db

import (
	"sync"

	"moviepilot-go/internal/logger"
	"moviepilot-go/pkg/models"
)

// SystemConfigOper 系统配置管理
type SystemConfigOper struct {
	configMap map[string]interface{}
	mutex     sync.RWMutex
}

var (
	systemConfigInstance *SystemConfigOper
	systemConfigOnce     sync.Once
)

// NewSystemConfigOper 创建系统配置管理器单例实�?func NewSystemConfigOper() *SystemConfigOper {
	systemConfigOnce.Do(func() {
		systemConfigInstance = &SystemConfigOper{
			configMap: make(map[string]interface{}),
		}
		// 加载配置到内�?		systemConfigInstance.loadConfig()
	})
	return systemConfigInstance
}

// loadConfig 加载配置到内�?func (s *SystemConfigOper) loadConfig() {
	// TODO: 从数据库加载配置�?	// 这里应该从数据库中加载SystemConfig表的数据
	// 由于数据库部分尚未实现，暂时留空
	logger.GetLoggerManager().Info("加载系统配置到内�?)
	
	// 初始化一些默认配置，特别是消息模�?	s.initDefaultConfig()
}

// Set 设置系统设置
func (s *SystemConfigOper) Set(key interface{}, value interface{}) *bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keyStr := s.getKeyString(key)

	// 旧�?	oldValue, exists := s.configMap[keyStr]

	// 更新内存
	s.configMap[keyStr] = value

	// 检查值是否发生变�?	if exists && oldValue == value {
		// 无需更新
		return nil
	}

	// TODO: 写入数据�?	// 这里应该实现数据库写入逻辑
	logger.GetLoggerManager().Debugf("设置系统配置�? %s = %v", keyStr, value)

	if exists {
		return boolPtr(true) // 更新成功
	}
	return boolPtr(true) // 创建成功
}

// Get 获取系统设置
func (s *SystemConfigOper) Get(key interface{}) interface{} {
	s.mutex.RLock()
	defer s.mutex.RLock()

	if key == nil {
		return s.All()
	}

	keyStr := s.getKeyString(key)
	
	// 返回配置值的深拷贝以避免引用共享
	if value, exists := s.configMap[keyStr]; exists {
		return s.deepCopy(value)
	}
	
	return nil
}

// All 获取所有系统设�?func (s *SystemConfigOper) All() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回所有配置的深拷贝以避免引用共享
	result := make(map[string]interface{})
	for k, v := range s.configMap {
		result[k] = s.deepCopy(v)
	}
	return result
}

// Delete 删除系统设置
func (s *SystemConfigOper) Delete(key interface{}) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keyStr := s.getKeyString(key)

	// 更新内存
	delete(s.configMap, keyStr)

	// TODO: 从数据库删除
	// 这里应该实现数据库删除逻辑
	logger.GetLoggerManager().Debugf("删除系统配置�? %s", keyStr)

	return true
}

// getKeyString 将key转换为字符串
func (s *SystemConfigOper) getKeyString(key interface{}) string {
	switch k := key.(type) {
	case models.SystemConfigKey:
		return string(k)
	case string:
		return k
	default:
		return ""
	}
}

// initDefaultConfig 初始化默认配�?func (s *SystemConfigOper) initDefaultConfig() {
	// 初始化默认的消息模板
	defaultTemplates := map[string]interface{}{
		"downloadAdded": "{\n    'title': '{{ title }}',\n    'text': '{% if size %}大小：{{ size }}{% endif %}'\n            '{% if pubdate %}\\n发布时间：{{ pubdate }}{% endif %}'\n            '{% if freedate %}\\n免费时间：{{ freedate }}{% endif %}'\n            '{% if seeders %}\\n做种数：{{ seeders }}{% endif %}'\n            '{% if volume_factor %}\\n促销：{{ volume_factor }}{% endif %}'\n            '{% if hit_and_run %}\\nHit&Run：{{ hit_and_run }}{% endif %}'\n            '{% if labels %}\\n标签：{{ labels }}{% endif %}'\n            '{% if description %}\\n描述：{{ description }}{% endif %}'\n}",
		"subscribeAdded": "{'title': '{{ title_year }}{% if season_fmt %} {{ season_fmt }}{% endif %} 已添加订�?}",
		"subscribeComplete": "{\n    'title': '{{ title_year }}'\n            '{% if season_fmt %} {{ season_fmt }}{% endif %} 已完成{{ msgstr }}',\n    'text': '{% if vote_average %}评分：{{ vote_average }}{% endif %}'\n            '{% if username %}，来自用户：{{ username }}{% endif %}'\n            '{% if actors %}\\n演员：{{ actors }}{% endif %}'\n            '{% if overview %}\\n简介：{{ overview }}{% endif %}'\n}",
	}
	
	s.configMap[string(models.SystemConfigKeyNotificationTemplates)] = defaultTemplates
}

// deepCopy 简单的深拷贝实�?func (s *SystemConfigOper) deepCopy(value interface{}) interface{} {
	// TODO: 实现真正的深拷贝逻辑
	// 当前实现只是返回原值，需要根据实际数据结构完�?	return value
}

// boolPtr 返回布尔值指�?func boolPtr(b bool) *bool {
	return &b
}
