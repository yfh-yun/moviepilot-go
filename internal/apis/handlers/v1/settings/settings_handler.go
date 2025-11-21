// Package settings 系统设置API处理器
package settings

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// SettingsHandler 系统设置处理器
// 提供系统配置管理、设置获取和更新等功能
type SettingsHandler struct {
	settingsService service.SettingsService
	logger          *zap.Logger
}

// NewSettingsHandler 创建系统设置处理器
func NewSettingsHandler(settingsService service.SettingsService, logger *zap.Logger) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		logger:          logger,
	}
}

// SettingItem 设置项结构体
type SettingItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	DataType    string      `json:"data_type"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	IsAdvanced  bool        `json:"is_advanced"`
}

// UpdateSettingRequest 更新设置请求结构体
type UpdateSettingRequest struct {
	Value interface{} `json:"value" binding:"required"`
}

// UpdateSettingsRequest 批量更新设置请求结构体
type UpdateSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// GetSettings 获取系统设置
// @Summary 获取系统设置
// @Description 获取系统的所有设置项
// @Tags settings
// @Accept json
// @Produce json
// @Param category query string false "设置分类"
// @Success 200 {object} response.SuccessResponse{data=[]SettingItem}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	category := c.Query("category")

	ctx := c.Request.Context()

	settings, err := h.settingsService.GetSettings(ctx, category)
	if err != nil {
		h.logger.Error("获取系统设置失败", zap.Error(err))
		response.InternalServerError(c, "获取系统设置失败")
		return
	}

	// 转换为响应格式
	var responseSettings []SettingItem
	for _, setting := range settings {
		responseSettings = append(responseSettings, SettingItem{
			Key:         setting.Key,
			Value:       setting.Value,
			DataType:    setting.DataType,
			Description: setting.Description,
			Category:    setting.Category,
			IsAdvanced:  setting.IsAdvanced,
		})
	}

	response.Success(c, responseSettings)
}

// GetSetting 获取指定设置项
// @Summary 获取指定设置项
// @Description 根据设置项键名获取对应的设置值
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "设置项键名"
// @Success 200 {object} response.SuccessResponse{data=SettingItem}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/{key} [get]
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	if key == "" {
		response.BadRequest(c, "设置项键名不能为空")
		return
	}

	ctx := c.Request.Context()

	setting, err := h.settingsService.GetSetting(ctx, key)
	if err != nil {
		if err == service.ErrSettingNotFound {
			response.NotFound(c, "设置项不存在")
			return
		}
		h.logger.Error("获取设置项失败", zap.Error(err), zap.String("key", key))
		response.InternalServerError(c, "获取设置项失败")
		return
	}

	responseSetting := SettingItem{
		Key:         setting.Key,
		Value:       setting.Value,
		DataType:    setting.DataType,
		Description: setting.Description,
		Category:    setting.Category,
		IsAdvanced:  setting.IsAdvanced,
	}

	response.Success(c, responseSetting)
}

// UpdateSetting 更新设置项
// @Summary 更新设置项
// @Description 更新指定设置项的值
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "设置项键名"
// @Param request body UpdateSettingRequest true "设置值"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/{key} [put]
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	if key == "" {
		response.BadRequest(c, "设置项键名不能为空")
		return
	}

	var req UpdateSettingRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新设置项请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("更新设置项请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	err := h.settingsService.UpdateSetting(ctx, key, req.Value)
	if err != nil {
		if err == service.ErrSettingNotFound {
			response.NotFound(c, "设置项不存在")
			return
		}
		h.logger.Error("更新设置项失败", zap.Error(err), zap.String("key", key))
		response.InternalServerError(c, "更新设置项失败")
		return
	}

	logger.Info("设置项更新成功", zap.String("key", key))
	response.Success(c, gin.H{
		"message": "设置项更新成功",
		"key":     key,
	})
}

// UpdateSettings 批量更新设置
// @Summary 批量更新设置
// @Description 批量更新多个设置项
// @Tags settings
// @Accept json
// @Produce json
// @Param request body UpdateSettingsRequest true "批量设置值"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/batch [put]
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量更新设置请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("批量更新设置请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	if len(req.Settings) == 0 {
		response.BadRequest(c, "设置项不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.settingsService.UpdateSettings(ctx, req.Settings)
	if err != nil {
		h.logger.Error("批量更新设置失败", zap.Error(err))
		response.InternalServerError(c, "批量更新设置失败")
		return
	}

	logger.Info("批量更新设置成功", zap.Int("count", len(req.Settings)))
	response.Success(c, gin.H{
		"message": "批量更新设置成功",
		"count":   len(req.Settings),
	})
}

// ResetSettings 重置设置
// @Summary 重置设置
// @Description 将所有设置项重置为默认值
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/reset [post]
func (h *SettingsHandler) ResetSettings(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.settingsService.ResetSettings(ctx)
	if err != nil {
		h.logger.Error("重置设置失败", zap.Error(err))
		response.InternalServerError(c, "重置设置失败")
		return
	}

	logger.Info("设置重置成功")
	response.Success(c, gin.H{
		"message": "设置重置成功",
	})
}

// ExportSettings 导出设置
// @Summary 导出设置
// @Description 导出系统设置到配置文件
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/export [get]
func (h *SettingsHandler) ExportSettings(c *gin.Context) {
	ctx := c.Request.Context()

	exportedSettings, err := h.settingsService.ExportSettings(ctx)
	if err != nil {
		h.logger.Error("导出设置失败", zap.Error(err))
		response.InternalServerError(c, "导出设置失败")
		return
	}

	response.Success(c, exportedSettings)
}

// ImportSettings 导入设置
// @Summary 导入设置
// @Description 从配置文件导入系统设置
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body map[string]interface{} true "设置数据"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/import [post]
func (h *SettingsHandler) ImportSettings(c *gin.Context) {
	var settings map[string]interface{}

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&settings); err != nil {
		h.logger.Warn("导入设置请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if len(settings) == 0 {
		response.BadRequest(c, "设置数据不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.settingsService.ImportSettings(ctx, settings)
	if err != nil {
		h.logger.Error("导入设置失败", zap.Error(err))
		response.InternalServerError(c, "导入设置失败")
		return
	}

	logger.Info("设置导入成功", zap.Int("count", len(settings)))
	response.Success(c, gin.H{
		"message": "设置导入成功",
		"count":   len(settings),
	})
}

// GetSettingsCategories 获取设置分类
// @Summary 获取设置分类
// @Description 获取所有设置分类列表
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]string}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/categories [get]
func (h *SettingsHandler) GetSettingsCategories(c *gin.Context) {
	ctx := c.Request.Context()

	categories, err := h.settingsService.GetSettingsCategories(ctx)
	if err != nil {
		h.logger.Error("获取设置分类失败", zap.Error(err))
		response.InternalServerError(c, "获取设置分类失败")
		return
	}

	response.Success(c, categories)
}

// GetSettingValidationRules 获取设置验证规则
// @Summary 获取设置验证规则
// @Description 获取设置的验证规则信息
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "设置项键名"
// @Success 200 {object} response.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/settings/{key}/validation [get]
func (h *SettingsHandler) GetSettingValidationRules(c *gin.Context) {
	key := c.Param("key")

	if key == "" {
		response.BadRequest(c, "设置项键名不能为空")
		return
	}

	ctx := c.Request.Context()

	rules, err := h.settingsService.GetSettingValidationRules(ctx, key)
	if err != nil {
		if err == service.ErrSettingNotFound {
			response.NotFound(c, "设置项不存在")
			return
		}
		h.logger.Error("获取设置验证规则失败", zap.Error(err), zap.String("key", key))
		response.InternalServerError(c, "获取设置验证规则失败")
		return
	}

	response.Success(c, rules)
}
