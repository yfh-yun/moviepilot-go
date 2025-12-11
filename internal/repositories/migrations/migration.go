package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/logger"
)

// Migration 数据库迁移
type Migration struct {
	db *gorm.DB
}

// NewMigration 创建迁移实例
func NewMigration(db *gorm.DB) *Migration {
	return &Migration{db: db}
}

// Run 运行数据库迁移
func (m *Migration) Run() error {
	zapLogger := logger.GetLogger()
	zapLogger.Info("Starting database migrations")

	// 自动迁移所有模型
	err := m.db.AutoMigrate(
		// 用户相关
		&database.User{},
		&database.UserConfig{},

		// 媒体相关
		&database.Media{},

		// 订阅相关
		&database.Subscribe{},
		&database.SubscribeHistory{},

		// 工作流相关
		&database.Workflow{},

		// 消息相关
		&database.Message{},

		// 站点相关
		&database.SiteIcon{},
		&database.SiteStatistic{},
		&database.SiteUserData{},

		// 文件相关
		&database.File{},
		&database.TransferHistory{},

		// 下载相关
		&database.DownloadHistory{},
		&database.DownloadFiles{},

		// 其他模型
		&database.Plugin{},
		&database.SystemConfig{},
	)

	if err != nil {
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	zapLogger.Info("Database migrations completed successfully")
	return nil
}

// CreateIndexes 创建索引
func (m *Migration) CreateIndexes() error {
	zapLogger := logger.GetLogger()
	zapLogger.Info("Creating database indexes")

	// 这里可以添加自定义索引创建逻辑
	// 例如：
	// if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
	//     return err
	// }

	zapLogger.Info("Database indexes created successfully")
	return nil
}

// SeedData 初始化种子数据
func (m *Migration) SeedData() error {
	zapLogger := logger.GetLogger()
	zapLogger.Info("Seeding initial data")

	// 这里可以添加种子数据初始化逻辑
	// 例如创建默认管理员用户等

	zapLogger.Info("Initial data seeded successfully")
	return nil
}
