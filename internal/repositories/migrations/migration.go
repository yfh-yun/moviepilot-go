package migrations

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"gorm.io/gorm"
)

// Migration 数据库迁移
type Migration struct {
	db *gorm.DB
}

// NewMigration 创建迁移实例
func NewMigration(db *gorm.DB) *Migration {
	return &Migration{db: db}
}

// AutoMigrate 自动迁移所有表
func (m *Migration) AutoMigrate() error {
	return m.db.AutoMigrate(
		&model.User{},
		&model.Media{},
		&model.Subscribe{},
		&model.DownloadHistory{},
		&model.DownloadFiles{},
		&model.TransferHistory{},
		&model.MediaServer{},
		&model.Message{},
		&model.PluginData{},
		&model.SystemConfig{},
		&model.UserConfig{},
		&model.Site{},
		&model.SiteUserData{},
		&model.SiteStatistic{},
		&model.SubscribeHistory{},
		&model.SiteIcon{},
		&model.Workflow{},
	)
}

// CreateIndexes 创建索引
func (m *Migration) CreateIndexes() error {
	// 用户表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_users_name ON users(name)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return err
	}

	// 媒体表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_medias_tmdbid ON medias(tmdbid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_medias_doubanid ON medias(doubanid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_medias_type ON medias(type)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_medias_title ON medias(title)").Error; err != nil {
		return err
	}

	// 订阅表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_name ON subscribes(name)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_tmdbid ON subscribes(tmdbid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_doubanid ON subscribes(doubanid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_bangumiid ON subscribes(bangumiid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_mediaid ON subscribes(mediaid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribes_state ON subscribes(state)").Error; err != nil {
		return err
	}

	// 下载历史表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_histories_path ON download_histories(path)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_histories_tmdbid ON download_histories(tmdbid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_histories_download_hash ON download_histories(download_hash)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_histories_type ON download_histories(type)").Error; err != nil {
		return err
	}

	// 下载文件表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_files_download_hash ON download_files(download_hash)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_files_fullpath ON download_files(fullpath)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_download_files_savepath ON download_files(savepath)").Error; err != nil {
		return err
	}

	// 转移历史表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_transfer_histories_tmdbid ON transfer_histories(tmdbid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_transfer_histories_tvdbid ON transfer_histories(tvdbid)").Error; err != nil {
		return err
	}

	// 站点表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_sites_name ON sites(name)").Error; err != nil {
		return err
	}

	// 站点用户数据表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_site_user_datas_site_name ON site_user_datas(site_name)").Error; err != nil {
		return err
	}

	// 站点统计表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_site_statistics_site_name ON site_statistics(site_name)").Error; err != nil {
		return err
	}

	// 订阅历史表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribe_histories_subscribe_id ON subscribe_histories(subscribe_id)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribe_histories_tmdbid ON subscribe_histories(tmdbid)").Error; err != nil {
		return err
	}
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_subscribe_histories_tvdbid ON subscribe_histories(tvdbid)").Error; err != nil {
		return err
	}

	// 站点图标表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_site_icons_site_name ON site_icons(site_name)").Error; err != nil {
		return err
	}

	// 系统配置表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(key)").Error; err != nil {
		return err
	}

	// 用户配置表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_user_configs_userid ON user_configs(userid)").Error; err != nil {
		return err
	}

	// 插件数据表索引
	if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_plugin_datas_plugin_key ON plugin_datas(plugin_key)").Error; err != nil {
		return err
	}

	return nil
}

// SeedData 初始化基础数据
func (m *Migration) SeedData() error {
	// 创建默认管理员用户
	adminUser := &model.User{
		Name:         "admin",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		IsActive:     true,
		IsSuperuser:  true,
		Permissions:  `{"all": true}`,
		Settings:     `{"theme": "light", "language": "zh-CN"}`,
	}

	// 检查是否已存在管理员
	var count int64
	if err := m.db.Model(&model.User{}).Where("name = ?", adminUser.Name).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		if err := m.db.Create(adminUser).Error; err != nil {
			return err
		}
	}

	// 初始化系统配置
	systemConfigs := []model.SystemConfig{
		{Key: "site.title", Value: "MoviePilot", Type: "string", Remark: "站点标题"},
		{Key: "site.subtitle", Value: "自动化媒体库管理工具", Type: "string", Remark: "站点副标题"},
		{Key: "site.logo", Value: "", Type: "string", Remark: "站点Logo"},
		{Key: "site.favicon", Value: "", Type: "string", Remark: "站点图标"},
		{Key: "site.theme", Value: "light", Type: "string", Remark: "站点主题"},
		{Key: "site.language", Value: "zh-CN", Type: "string", Remark: "站点语言"},
		{Key: "site.version", Value: "2.8.1", Type: "string", Remark: "系统版本"},
		{Key: "download.watch_dirs", Value: "[]", Type: "json", Remark: "下载监控目录"},
		{Key: "transfer.download_dirs", Value: "[]", Type: "json", Remark: "下载目录映射"},
		{Key: "transfer.library_dirs", Value: "[]", Type: "json", Remark: "媒体库目录映射"},
		{Key: "media.type", Value: "movie,tv", Type: "string", Remark: "支持的媒体类型"},
		{Key: "media.quality", Value: "1080p,720p", Type: "string", Remark: "优先视频质量"},
		{Key: "search.enable", Value: "true", Type: "bool", Remark: "启用搜索"},
		{Key: "search.default_sites", Value: "[]", Type: "json", Remark: "默认搜索站点"},
		{Key: "notification.enable", Value: "true", Type: "bool", Remark: "启用通知"},
		{Key: "notification.types", Value: "[]", Type: "json", Remark: "通知类型"},
	}

	for _, config := range systemConfigs {
		var count int64
		if err := m.db.Model(&model.SystemConfig{}).Where("key = ?", config.Key).Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := m.db.Create(&config).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// Run 运行完整迁移
func (m *Migration) Run() error {
	// 自动迁移表结构
	if err := m.AutoMigrate(); err != nil {
		return err
	}

	// 创建索引
	if err := m.CreateIndexes(); err != nil {
		return err
	}

	// 初始化基础数据
	if err := m.SeedData(); err != nil {
		return err
	}

	return nil
}
