package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
)

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	if log != nil {
		log.Info("starting database auto migration...")
	}

	// 定义所有需要迁移的模型
	modelsToMigrate := []any{
		// 用户相关
		&database.User{},

		// 媒体相关
		&database.Media{},

		// 订阅相关
		&database.Subscribe{},

		// 下载相关
		&database.Download{},
		&database.DownloadHistory{},

		// 转移相关
		&database.TransferHistory{},

		// 站点相关
		&database.Site{},
		&database.SiteStatistic{},
		&database.SiteUserData{},

		// 插件相关
		&database.PluginData{},

		// 系统相关
		&database.SystemConfig{},

		// 搜索相关
		&database.SearchHistory{},
	}

	// 执行迁移
	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		if log != nil {
			log.Error("database auto migration failed", zap.Error(err))
		}
		return fmt.Errorf("auto migration failed: %w", err)
	}

	if log != nil {
		log.Info("database auto migration completed successfully",
			zap.Int("models_count", len(modelsToMigrate)))
	}

	return nil
}

// CreateIndexes 创建索引
func CreateIndexes(db *gorm.DB, log *zap.Logger) error {
	if log != nil {
		log.Info("creating database indexes...")
	}

	// 订阅表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribe_name ON subscribes(name)").Error; err != nil {
		if log != nil {
			log.Error("failed to create subscribe name index", zap.Error(err))
		}
		return fmt.Errorf("failed to create subscribe name index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribe_state ON subscribes(state)").Error; err != nil {
		if log != nil {
			log.Error("failed to create subscribe state index", zap.Error(err))
		}
		return fmt.Errorf("failed to create subscribe state index: %w", err)
	}

	// 站点表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_site_domain ON sites(domain)").Error; err != nil {
		if log != nil {
			log.Error("failed to create site domain index", zap.Error(err))
		}
		return fmt.Errorf("failed to create site domain index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_site_is_active ON sites(is_active)").Error; err != nil {
		if log != nil {
			log.Error("failed to create site is_active index", zap.Error(err))
		}
		return fmt.Errorf("failed to create site is_active index: %w", err)
	}

	// 下载历史索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_download_history_date ON download_histories(date)").Error; err != nil {
		if log != nil {
			log.Error("failed to create download history date index", zap.Error(err))
		}
		return fmt.Errorf("failed to create download history date index: %w", err)
	}

	// 转移历史索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_transfer_history_date ON transfer_histories(date)").Error; err != nil {
		if log != nil {
			log.Error("failed to create transfer history date index", zap.Error(err))
		}
		return fmt.Errorf("failed to create transfer history date index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_transfer_history_status ON transfer_histories(status)").Error; err != nil {
		if log != nil {
			log.Error("failed to create transfer history status index", zap.Error(err))
		}
		return fmt.Errorf("failed to create transfer history status index: %w", err)
	}

	if log != nil {
		log.Info("database indexes created successfully")
	}

	return nil
}

// InitializeData 初始化基础数据
func InitializeData(db *gorm.DB, log *zap.Logger) error {
	if log != nil {
		log.Info("initializing base data...")
	}

	// 检查是否已经初始化
	var count int64
	if err := db.Model(&database.SystemConfig{}).Count(&count).Error; err != nil {
		if log != nil {
			log.Error("failed to check system config", zap.Error(err))
		}
		return fmt.Errorf("failed to check system config: %w", err)
	}

	if count > 0 {
		if log != nil {
			log.Info("base data already initialized, skipping")
		}
		return nil
	}

	// 创建默认系统配置
	defaultConfigs := []database.SystemConfig{
		{
			Key:   "system.version",
			Value: "2.8.1",
		},
		{
			Key:   "system.initialized",
			Value: "true",
		},
	}

	for _, config := range defaultConfigs {
		if err := db.Create(&config).Error; err != nil {
			if log != nil {
				log.Error("failed to create default config", zap.Error(err))
			}
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	if log != nil {
		log.Info("base data initialized successfully")
	}

	return nil
}
