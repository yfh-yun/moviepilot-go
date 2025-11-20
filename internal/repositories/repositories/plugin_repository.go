package repositories

import (
	"gorm.io/gorm"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// PluginRepository 插件仓储实现
type PluginRepository struct {
	db *gorm.DB
}

// NewPluginRepository 创建插件仓储
func NewPluginRepository(db *gorm.DB) interfaces.PluginRepository {
	return &model.PluginRepository{db: db}
}

// Create 创建插件记录
func (r *PluginRepository) Create(plugin *model.Plugin) error {
	return r.db.Create(plugin).Error
}

// Update 更新插件记录
func (r *PluginRepository) Update(plugin *model.Plugin) error {
	return r.db.Save(plugin).Error
}

// Delete 删除插件记录
func (r *PluginRepository) Delete(pluginID string) error {
	return r.db.Delete(&model.Plugin{}, "id = ?", pluginID).Error
}

// GetByID 根据ID获取插件
func (r *PluginRepository) GetByID(pluginID string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.Where("id = ?", pluginID).First(&plugin).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &plugin, err
}

// GetByName 根据名称获取插件
func (r *PluginRepository) GetByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.Where("name = ?", name).First(&plugin).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &plugin, err
}

// GetAll 获取所有插件
func (r *PluginRepository) GetAll() ([]*model.Plugin, error) {
	var plugins []*model.Plugin
	err := r.db.Order("name ASC").Find(&plugins).Error
	return plugins, err
}

// GetEnabled 获取已启用的插件
func (r *PluginRepository) GetEnabled() ([]*model.Plugin, error) {
	var plugins []*model.Plugin
	err := r.db.Where("enabled = ?", true).Order("name ASC").Find(&plugins).Error
	return plugins, err
}

// GetInstalled 获取已安装的插件
func (r *PluginRepository) GetInstalled() ([]*model.Plugin, error) {
	var plugins []*model.Plugin
	err := r.db.Where("installed = ?", true).Order("name ASC").Find(&plugins).Error
	return plugins, err
}

// GetByType 根据类型获取插件
func (r *PluginRepository) GetByType(pluginType string) ([]*model.Plugin, error) {
	var plugins []*model.Plugin
	err := r.db.Where("type = ?", pluginType).Order("name ASC").Find(&plugins).Error
	return plugins, err
}

// ListByPage 分页获取插件列表
func (r *PluginRepository) ListByPage(page, count int) ([]*model.Plugin, int64, error) {
	var plugins []*model.Plugin
	var total int64

	// 统计总数
	err := r.db.Model(&model.Plugin{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * count
	if offset < 0 {
		offset = 0
	}

	// 获取分页数据
	err = r.db.Offset(offset).Limit(count).Order("name ASC").Find(&plugins).Error
	return plugins, total, err
}

// Search 搜索插件
func (r *PluginRepository) Search(keyword string, page, count int) ([]*model.Plugin, int64, error) {
	var plugins []*model.Plugin
	var total int64

	// 构建搜索条件
	db := r.db.Model(&model.Plugin{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?",
			"%"+keyword+"%",
			"%"+keyword+"%")
	}

	// 统计总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * count
	if offset < 0 {
		offset = 0
	}

	// 获取分页数据
	err = db.Offset(offset).Limit(count).Order("name ASC").Find(&plugins).Error
	return plugins, total, err
}

// Count 统计插件数量
func (r *PluginRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Plugin{}).Count(&count).Error
	return count, err
}

// CountEnabled 统计已启用的插件数量
func (r *PluginRepository) CountEnabled() (int64, error) {
	var count int64
	err := r.db.Model(&model.Plugin{}).Where("enabled = ?", true).Count(&count).Error
	return count, err
}

// CountInstalled 统计已安装的插件数量
func (r *PluginRepository) CountInstalled() (int64, error) {
	var count int64
	err := r.db.Model(&model.Plugin{}).Where("installed = ?", true).Count(&count).Error
	return count, err
}

// UpdateConfig 更新插件配置
func (r *PluginRepository) UpdateConfig(pluginID, config string) error {
	return r.db.Model(&model.Plugin{}).
		Where("id = ?", pluginID).
		Update("config", config).Error
}

// UpdateStatus 更新插件状态
func (r *PluginRepository) UpdateStatus(pluginID string, enabled bool) error {
	return r.db.Model(&model.Plugin{}).
		Where("id = ?", pluginID).
		Update("enabled", enabled).Error
}

// UpdateInstallStatus 更新插件安装状态
func (r *PluginRepository) UpdateInstallStatus(pluginID string, installed bool) error {
	return r.db.Model(&model.Plugin{}).
		Where("id = ?", pluginID).
		Update("installed", installed).Error
}
