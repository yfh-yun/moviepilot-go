package impl

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// HistoryServiceImpl 历史记录服务实现
type HistoryServiceImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewHistoryService 创建历史记录服务
func NewHistoryService(db *gorm.DB, logger *zap.Logger) service.HistoryService {
	return &HistoryServiceImpl{
		db:     db,
		logger: logger,
	}
}

// GetDownloadHistory 获取下载历史
func (h *HistoryServiceImpl) GetDownloadHistory(params service.DownloadHistoryParams) (*service.PaginatedResponse[service.DownloadHistoryItem], error) {
	var histories []models.DownloadHistory
	var total int64
	
	query := h.buildDownloadQuery(params)
	
	// 获取总数
	if err := query.Model(&models.DownloadHistory{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询下载历史总数失败: %w", err)
	}
	
	// 获取分页数据
	offset := (params.Page - 1) * params.Count
	if err := query.Offset(offset).Limit(params.Count).Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("查询下载历史失败: %w", err)
	}
	
	// 转换为服务层模型
	items := make([]service.DownloadHistoryItem, len(histories))
	for i, history := range histories {
		items[i] = h.convertToDownloadHistoryItem(history)
	}
	
	return service.NewPaginatedResponse(items, total, params.Page, params.Count), nil
}

// CreateDownloadHistory 创建下载历史
func (h *HistoryServiceImpl) CreateDownloadHistory(history *service.DownloadHistoryItem) error {
	model := h.convertFromDownloadHistoryItem(history)
	
	if err := h.db.Create(model).Error; err != nil {
		return fmt.Errorf("创建下载历史失败: %w", err)
	}
	
	h.logger.Info("下载历史已创建", 
		zap.String("id", history.ID),
		zap.String("media_title", history.MediaTitle))
	
	return nil
}

// UpdateDownloadHistory 更新下载历史
func (h *HistoryServiceImpl) UpdateDownloadHistory(id string, history *service.DownloadHistoryItem) error {
	model := h.convertFromDownloadHistoryItem(history)
	model.ID = id
	
	if err := h.db.Save(model).Error; err != nil {
		return fmt.Errorf("更新下载历史失败: %w", err)
	}
	
	h.logger.Info("下载历史已更新", 
		zap.String("id", id),
		zap.String("status", history.Status))
	
	return nil
}

// DeleteDownloadHistory 删除下载历史
func (h *HistoryServiceImpl) DeleteDownloadHistory(id string) error {
	if err := h.db.Delete(&models.DownloadHistory{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除下载历史失败: %w", err)
	}
	
	h.logger.Info("下载历史已删除", zap.String("id", id))
	return nil
}

// GetTransferHistory 获取转移历史
func (h *HistoryServiceImpl) GetTransferHistory(params service.TransferHistoryParams) (*service.PaginatedResponse[service.TransferHistoryItem], error) {
	var histories []models.TransferHistory
	var total int64
	
	query := h.buildTransferQuery(params)
	
	// 获取总数
	if err := query.Model(&models.TransferHistory{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询转移历史总数失败: %w", err)
	}
	
	// 获取分页数据
	offset := (params.Page - 1) * params.Count
	if err := query.Offset(offset).Limit(params.Count).Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("查询转移历史失败: %w", err)
	}
	
	// 转换为服务层模型
	items := make([]service.TransferHistoryItem, len(histories))
	for i, history := range histories {
		items[i] = h.convertToTransferHistoryItem(history)
	}
	
	return service.NewPaginatedResponse(items, total, params.Page, params.Count), nil
}

// CreateTransferHistory 创建转移历史
func (h *HistoryServiceImpl) CreateTransferHistory(history *service.TransferHistoryItem) error {
	model := h.convertFromTransferHistoryItem(history)
	
	if err := h.db.Create(model).Error; err != nil {
		return fmt.Errorf("创建转移历史失败: %w", err)
	}
	
	h.logger.Info("转移历史已创建", 
		zap.String("id", history.ID),
		zap.String("media_title", history.MediaTitle))
	
	return nil
}

// UpdateTransferHistory 更新转移历史
func (h *HistoryServiceImpl) UpdateTransferHistory(id string, history *service.TransferHistoryItem) error {
	model := h.convertFromTransferHistoryItem(history)
	model.ID = id
	
	if err := h.db.Save(model).Error; err != nil {
		return fmt.Errorf("更新转移历史失败: %w", err)
	}
	
	h.logger.Info("转移历史已更新", 
		zap.String("id", id),
		zap.String("status", history.Status))
	
	return nil
}

// DeleteTransferHistory 删除转移历史
func (h *HistoryServiceImpl) DeleteTransferHistory(id string) error {
	if err := h.db.Delete(&models.TransferHistory{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除转移历史失败: %w", err)
	}
	
	h.logger.Info("转移历史已删除", zap.String("id", id))
	return nil
}

// GetSubscribeHistory 获取订阅历史
func (h *HistoryServiceImpl) GetSubscribeHistory(params service.SubscribeHistoryParams) (*service.PaginatedResponse[service.SubscribeHistoryItem], error) {
	var histories []models.SubscribeHistory
	var total int64
	
	query := h.buildSubscribeQuery(params)
	
	// 获取总数
	if err := query.Model(&models.SubscribeHistory{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询订阅历史总数失败: %w", err)
	}
	
	// 获取分页数据
	offset := (params.Page - 1) * params.Count
	if err := query.Offset(offset).Limit(params.Count).Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("查询订阅历史失败: %w", err)
	}
	
	// 转换为服务层模型
	items := make([]service.SubscribeHistoryItem, len(histories))
	for i, history := range histories {
		items[i] = h.convertToSubscribeHistoryItem(history)
	}
	
	return service.NewPaginatedResponse(items, total, params.Page, params.Count), nil
}

// CreateSubscribeHistory 创建订阅历史
func (h *HistoryServiceImpl) CreateSubscribeHistory(history *service.SubscribeHistoryItem) error {
	model := h.convertFromSubscribeHistoryItem(history)
	
	if err := h.db.Create(model).Error; err != nil {
		return fmt.Errorf("创建订阅历史失败: %w", err)
	}
	
	h.logger.Info("订阅历史已创建", 
		zap.String("id", history.ID),
		zap.String("media_title", history.MediaTitle))
	
	return nil
}

// UpdateSubscribeHistory 更新订阅历史
func (h *HistoryServiceImpl) UpdateSubscribeHistory(id string, history *service.SubscribeHistoryItem) error {
	model := h.convertFromSubscribeHistoryItem(history)
	model.ID = id
	
	if err := h.db.Save(model).Error; err != nil {
		return fmt.Errorf("更新订阅历史失败: %w", err)
	}
	
	h.logger.Info("订阅历史已更新", 
		zap.String("id", id),
		zap.String("status", history.Status))
	
	return nil
}

// DeleteSubscribeHistory 删除订阅历史
func (h *HistoryServiceImpl) DeleteSubscribeHistory(id string) error {
	if err := h.db.Delete(&models.SubscribeHistory{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除订阅历史失败: %w", err)
	}
	
	h.logger.Info("订阅历史已删除", zap.String("id", id))
	return nil
}

// GetSystemHistory 获取系统历史
func (h *HistoryServiceImpl) GetSystemHistory(params service.SystemHistoryParams) (*service.PaginatedResponse[service.SystemHistoryItem], error) {
	var histories []models.SystemHistory
	var total int64
	
	query := h.buildSystemQuery(params)
	
	// 获取总数
	if err := query.Model(&models.SystemHistory{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询系统历史总数失败: %w", err)
	}
	
	// 获取分页数据
	offset := (params.Page - 1) * params.Count
	if err := query.Offset(offset).Limit(params.Count).Order("create_time DESC").Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("查询系统历史失败: %w", err)
	}
	
	// 转换为服务层模型
	items := make([]service.SystemHistoryItem, len(histories))
	for i, history := range histories {
		items[i] = h.convertToSystemHistoryItem(history)
	}
	
	return service.NewPaginatedResponse(items, total, params.Page, params.Count), nil
}

// CreateSystemHistory 创建系统历史
func (h *HistoryServiceImpl) CreateSystemHistory(history *service.SystemHistoryItem) error {
	model := h.convertFromSystemHistoryItem(history)
	
	if err := h.db.Create(model).Error; err != nil {
		return fmt.Errorf("创建系统历史失败: %w", err)
	}
	
	h.logger.Debug("系统历史已创建", 
		zap.String("id", history.ID),
		zap.String("type", history.Type))
	
	return nil
}

// UpdateSystemHistory 更新系统历史
func (h *HistoryServiceImpl) UpdateSystemHistory(id string, history *service.SystemHistoryItem) error {
	model := h.convertFromSystemHistoryItem(history)
	model.ID = id
	
	if err := h.db.Save(model).Error; err != nil {
		return fmt.Errorf("更新系统历史失败: %w", err)
	}
	
	h.logger.Debug("系统历史已更新", 
		zap.String("id", id),
		zap.String("type", history.Type))
	
	return nil
}

// DeleteSystemHistory 删除系统历史
func (h *HistoryServiceImpl) DeleteSystemHistory(id string) error {
	if err := h.db.Delete(&models.SystemHistory{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除系统历史失败: %w", err)
	}
	
	h.logger.Debug("系统历史已删除", zap.String("id", id))
	return nil
}

// CleanupHistory 清理历史记录
func (h *HistoryServiceImpl) CleanupHistory(ctx context.Context, beforeDate time.Time) error {
	h.logger.Info("开始清理历史记录", zap.Time("before_date", beforeDate))
	
	// 清理下载历史
	if err := h.db.WithContext(ctx).Where("create_time < ?", beforeDate).Delete(&models.DownloadHistory{}).Error; err != nil {
		return fmt.Errorf("清理下载历史失败: %w", err)
	}
	
	// 清理转移历史
	if err := h.db.WithContext(ctx).Where("create_time < ?", beforeDate).Delete(&models.TransferHistory{}).Error; err != nil {
		return fmt.Errorf("清理转移历史失败: %w", err)
	}
	
	// 清理订阅历史
	if err := h.db.WithContext(ctx).Where("create_time < ?", beforeDate).Delete(&models.SubscribeHistory{}).Error; err != nil {
		return fmt.Errorf("清理订阅历史失败: %w", err)
	}
	
	// 清理系统历史（保留更长时间）
	systemBeforeDate := beforeDate.AddDate(0, -6, 0) // 保留6个月
	if err := h.db.WithContext(ctx).Where("create_time < ?", systemBeforeDate).Delete(&models.SystemHistory{}).Error; err != nil {
		return fmt.Errorf("清理系统历史失败: %w", err)
	}
	
	h.logger.Info("历史记录清理完成", zap.Time("before_date", beforeDate))
	return nil
}

// 辅助方法

// buildDownloadQuery 构建下载历史查询
func (h *HistoryServiceImpl) buildDownloadQuery(params service.DownloadHistoryParams) *gorm.DB {
	query := h.db.Model(&models.DownloadHistory{})
	
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.MediaType != "" {
		query = query.Where("media_type = ?", params.MediaType)
	}
	if params.MediaID != "" {
		query = query.Where("media_id = ?", params.MediaID)
	}
	if params.Source != "" {
		query = query.Where("source = ?", params.Source)
	}
	if params.StartTime != "" {
		query = query.Where("create_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}
	
	return query
}

// buildTransferQuery 构建转移历史查询
func (h *HistoryServiceImpl) buildTransferQuery(params service.TransferHistoryParams) *gorm.DB {
	query := h.db.Model(&models.TransferHistory{})
	
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.MediaType != "" {
		query = query.Where("media_type = ?", params.MediaType)
	}
	if params.MediaID != "" {
		query = query.Where("media_id = ?", params.MediaID)
	}
	if params.SourcePath != "" {
		query = query.Where("source_path LIKE ?", "%"+params.SourcePath+"%")
	}
	if params.DestPath != "" {
		query = query.Where("dest_path LIKE ?", "%"+params.DestPath+"%")
	}
	if params.TransferMode != "" {
		query = query.Where("transfer_mode = ?", params.TransferMode)
	}
	if params.StartTime != "" {
		query = query.Where("create_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}
	
	return query
}

// buildSubscribeQuery 构建订阅历史查询
func (h *HistoryServiceImpl) buildSubscribeQuery(params service.SubscribeHistoryParams) *gorm.DB {
	query := h.db.Model(&models.SubscribeHistory{})
	
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.MediaType != "" {
		query = query.Where("media_type = ?", params.MediaType)
	}
	if params.MediaID != "" {
		query = query.Where("media_id = ?", params.MediaID)
	}
	if params.Season > 0 {
		query = query.Where("season = ?", params.Season)
	}
	if params.Episode > 0 {
		query = query.Where("episode = ?", params.Episode)
	}
	if params.StartTime != "" {
		query = query.Where("create_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}
	
	return query
}

// buildSystemQuery 构建系统历史查询
func (h *HistoryServiceImpl) buildSystemQuery(params service.SystemHistoryParams) *gorm.DB {
	query := h.db.Model(&models.SystemHistory{})
	
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Level != "" {
		query = query.Where("level = ?", params.Level)
	}
	if params.StartTime != "" {
		query = query.Where("create_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}
	
	return query
}

// 转换方法

func (h *HistoryServiceImpl) convertToDownloadHistoryItem(model models.DownloadHistory) service.DownloadHistoryItem {
	return service.DownloadHistoryItem{
		ID:          model.ID,
		MediaID:     model.MediaID,
		MediaTitle:  model.MediaTitle,
		MediaType:   model.MediaType,
		Source:      model.Source,
		Status:      model.Status,
		Size:        model.Size,
		Downloaded:  model.Downloaded,
		Speed:       model.Speed,
		Progress:    model.Progress,
		Error:       model.Error,
		CreateTime:  model.CreateTime,
		UpdateTime:  model.UpdateTime,
		CompletedAt: model.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertFromDownloadHistoryItem(item *service.DownloadHistoryItem) *models.DownloadHistory {
	return &models.DownloadHistory{
		ID:          item.ID,
		MediaID:     item.MediaID,
		MediaTitle:  item.MediaTitle,
		MediaType:   item.MediaType,
		Source:      item.Source,
		Status:      item.Status,
		Size:        item.Size,
		Downloaded:  item.Downloaded,
		Speed:       item.Speed,
		Progress:    item.Progress,
		Error:       item.Error,
		CreateTime:  item.CreateTime,
		UpdateTime:  item.UpdateTime,
		CompletedAt: item.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertToTransferHistoryItem(model models.TransferHistory) service.TransferHistoryItem {
	return service.TransferHistoryItem{
		ID:           model.ID,
		MediaID:      model.MediaID,
		MediaTitle:   model.MediaTitle,
		MediaType:    model.MediaType,
		SourcePath:   model.SourcePath,
		DestPath:     model.DestPath,
		Status:       model.Status,
		Size:         model.Size,
		Transferred:  model.Transferred,
		Speed:        model.Speed,
		Progress:     model.Progress,
		TransferMode: model.TransferMode,
		Error:        model.Error,
		CreateTime:   model.CreateTime,
		UpdateTime:   model.UpdateTime,
		CompletedAt:  model.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertFromTransferHistoryItem(item *service.TransferHistoryItem) *models.TransferHistory {
	return &models.TransferHistory{
		ID:           item.ID,
		MediaID:      item.MediaID,
		MediaTitle:   item.MediaTitle,
		MediaType:    item.MediaType,
		SourcePath:   item.SourcePath,
		DestPath:     item.DestPath,
		Status:       item.Status,
		Size:         item.Size,
		Transferred:  item.Transferred,
		Speed:        item.Speed,
		Progress:     item.Progress,
		TransferMode: item.TransferMode,
		Error:        item.Error,
		CreateTime:   item.CreateTime,
		UpdateTime:   item.UpdateTime,
		CompletedAt:  item.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertToSubscribeHistoryItem(model models.SubscribeHistory) service.SubscribeHistoryItem {
	return service.SubscribeHistoryItem{
		ID:         model.ID,
		MediaID:    model.MediaID,
		MediaTitle: model.MediaTitle,
		MediaType:  model.MediaType,
		Season:     model.Season,
		Episode:    model.Episode,
		Status:     model.Status,
		Error:      model.Error,
		CreateTime: model.CreateTime,
		UpdateTime: model.UpdateTime,
		CompletedAt: model.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertFromSubscribeHistoryItem(item *service.SubscribeHistoryItem) *models.SubscribeHistory {
	return &models.SubscribeHistory{
		ID:         item.ID,
		MediaID:    item.MediaID,
		MediaTitle: item.MediaTitle,
		MediaType:  item.MediaType,
		Season:     item.Season,
		Episode:    item.Episode,
		Status:     item.Status,
		Error:      item.Error,
		CreateTime: item.CreateTime,
		UpdateTime: item.UpdateTime,
		CompletedAt: item.CompletedAt,
	}
}

func (h *HistoryServiceImpl) convertToSystemHistoryItem(model models.SystemHistory) service.SystemHistoryItem {
	return service.SystemHistoryItem{
		ID:         model.ID,
		Type:       model.Type,
		Level:      model.Level,
		Message:    model.Message,
		Details:    model.Details,
		CreateTime: model.CreateTime,
	}
}

func (h *HistoryServiceImpl) convertFromSystemHistoryItem(item *service.SystemHistoryItem) *models.SystemHistory {
	return &models.SystemHistory{
		ID:         item.ID,
		Type:       item.Type,
		Level:      item.Level,
		Message:    item.Message,
		Details:    item.Details,
		CreateTime: item.CreateTime,
	}
}