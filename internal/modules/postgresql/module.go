package postgresql

import (
	"moviepilot-go/internal/modules"
	"moviepilot-go/pkg/models"
)

// PostgreSQLModule PostgreSQL模块
type PostgreSQLModule struct {
	modules.ModuleBase
}

// NewPostgreSQLModule 创建PostgreSQL模块实例
func NewPostgreSQLModule() *PostgreSQLModule {
	return &PostgreSQLModule{}
}

// InitModule 初始化模�?func (pm *PostgreSQLModule) InitModule() error {
	// 初始化模�?	return nil
}

// GetName 获取模块名称
func (pm *PostgreSQLModule) GetName() string {
	return "PostgreSQL"
}

// GetType 获取模块类型
func (pm *PostgreSQLModule) GetType() models.ModuleType {
	return models.ModuleTypeOther
}

// GetSubtype 获取模块子类�?func (pm *PostgreSQLModule) GetSubtype() string {
	return "PostgreSQL"
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (pm *PostgreSQLModule) GetPriority() int {
	return 0
}

// InitSetting 初始化设�?func (pm *PostgreSQLModule) InitSetting() (string, interface{}) {
	return "", nil
}

// Stop 停止模块
func (pm *PostgreSQLModule) Stop() {
	// PostgreSQL模块停止操作
}

// Test 测试模块连接�?func (pm *PostgreSQLModule) Test() (*bool, string) {
	// 创建PostgreSQL实例
	pg := NewPostgreSQL()
	
	// 测试连接�?	success, message := pg.Test()
	if message == "" {
		// 返回nil表示不适用当前配置
		return nil, ""
	}
	
	return &success, message
}
