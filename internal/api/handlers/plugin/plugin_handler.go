// Package plugin 插件管理API处理器
package plugin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/internal/api/validator"
	
	pluginservice "github.com/yfh-yun/moviepilot-go/internal/service/plugin"
)

// PluginHandler 插件管理处理器
// 提供插件的安装、卸载、启用、禁用、配置和管理功能
type PluginHandler struct {
	pluginService *pluginservice.PluginService
	logger        *zap.Logger
}

// NewPluginHandler 创建插件管理处理器
func NewPluginHandler(pluginService *pluginservice.PluginService, logger *zap.Logger) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
		logger:        logger,
	}
}

// PluginInfoResponse 插件信息响应结构体
type PluginInfoResponse struct {
	ID           string                 `json:"id"`
	Key          string                 `json:"key"`
	Name         string                 `json:"name"`
	State        bool                   `json:"state"`
	Version      string                 `json:"version"`
	Icon         string                 `json:"icon"`
	Author       string                 `json:"author"`
	Description  string                 `json:"description"`
	DescriptionV2 string                 `json:"description_v2"`
	Note         string                 `json:"note"`
	HasPage      bool                   `json:"has_page"`
	PageURL      string                 `json:"page_url"`
	Config       map[string]interface{} `json:"config,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// ListPluginsRequest 插件列表请求结构体
type ListPluginsRequest struct {
	EnabledOnly bool `form:"enabled_only"`
}

// InstallPluginRequest 安装插件请求结构体
type InstallPluginRequest struct {
	PluginID string `json:"plugin_id" binding:"required"`
}

// ConfigurePluginRequest 配置插件请求结构体
type ConfigurePluginRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// ExecutePluginRequest 执行插件请求结构体
type ExecutePluginRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// ListPlugins 获取插件列表
// @Summary 获取插件列表
// @Description 获取插件列表，可选择只获取启用的插件
// @Tags plugin
// @Accept json
// @Produce json
// @Param enabled_only query bool false "是否只获取启用的插件"
// @Success 200 {object} response.APIResponse{data=[]PluginInfoResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins [get]
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	var req ListPluginsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("获取插件列表请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层获取插件列表
	plugins, err := h.pluginService.ListPlugins(ctx, req.EnabledOnly)
	if err != nil {
		h.logger.Error("获取插件列表失败", zap.Error(err))
		response.InternalServerError(c, "获取插件列表失败")
		return
	}

	// 转换为响应格式
	var responsePlugins []PluginInfoResponse
	for _, plugin := range plugins {
		responsePlugins = append(responsePlugins, PluginInfoResponse{
			ID:           strconv.FormatUint(uint64(plugin.ID), 10),
			Key:          plugin.Key,
			Name:         plugin.Name,
			State:        plugin.State,
			Version:      plugin.Version,
			Icon:         plugin.Icon,
			Author:       plugin.Author,
			Description:  plugin.Description,
			DescriptionV2: plugin.DescriptionV2,
			Note:         plugin.Note,
			HasPage:      plugin.HasPage,
			PageURL:      plugin.PageURL,
			CreatedAt:    plugin.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    plugin.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, responsePlugins)
}

// GetPlugin 获取插件详情
// @Summary 获取插件详情
// @Description 根据插件ID获取插件详细信息
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} response.APIResponse{data=PluginInfoResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id} [get]
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层获取插件详情
	plugin, err := h.pluginService.GetPlugin(ctx, pluginID)
	if err != nil {
		if err == plugin.ErrPluginNotFound {
			response.NotFound(c, "插件不存在")
			return
		}
		h.logger.Error("获取插件详情失败", zap.String("plugin_id", pluginID), zap.Error(err))
		response.InternalServerError(c, "获取插件详情失败")
		return
	}

	// 获取插件配置
	config, err := h.pluginService.GetPluginConfig(ctx, pluginID)
	if err != nil {
		h.logger.Warn("获取插件配置失败", zap.String("plugin_id", pluginID), zap.Error(err))
		config = make(map[string]interface{})
	}

	// 转换为响应格式
	responsePlugin := PluginInfoResponse{
		ID:           strconv.FormatUint(uint64(plugin.ID), 10),
		Key:          plugin.Key,
		Name:         plugin.Name,
		State:        plugin.State,
		Version:      plugin.Version,
		Icon:         plugin.Icon,
		Author:       plugin.Author,
		Description:  plugin.Description,
		DescriptionV2: plugin.DescriptionV2,
		Note:         plugin.Note,
		HasPage:      plugin.HasPage,
		PageURL:      plugin.PageURL,
		Config:       config,
		CreatedAt:    plugin.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    plugin.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	response.Success(c, responsePlugin)
}

// InstallPlugin 安装插件
// @Summary 安装插件
// @Description 安装指定的插件
// @Tags plugin
// @Accept json
// @Produce json
// @Param request body InstallPluginRequest true "安装插件请求"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/install [post]
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var req InstallPluginRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("安装插件请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.NewValidator(h.logger).Validate(req); err != nil {
		h.logger.Warn("安装插件请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.NewValidator(h.logger).TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	// 调用服务层安装插件
	err := h.pluginService.InstallPlugin(ctx, req.PluginID)
	if err != nil {
		h.logger.Error("安装插件失败", zap.String("plugin_id", req.PluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginAlreadyInstalled:
			response.BadRequest(c, "插件已安装")
		default:
			response.InternalServerError(c, "安装插件失败")
		}
		return
	}

	response.SuccessWithMessage(c, "插件安装成功", nil)
}

// UninstallPlugin 卸载插件
// @Summary 卸载插件
// @Description 卸载指定的插件
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id} [delete]
func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层卸载插件
	err := h.pluginService.UninstallPlugin(ctx, pluginID)
	if err != nil {
		h.logger.Error("卸载插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginNotFound:
			response.NotFound(c, "插件不存在")
		case plugin.ErrPluginNotInstalled:
			response.BadRequest(c, "插件未安装")
		default:
			response.InternalServerError(c, "卸载插件失败")
		}
		return
	}

	response.SuccessWithMessage(c, "插件卸载成功", nil)
}

// EnablePlugin 启用插件
// @Summary 启用插件
// @Description 启用指定的插件
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id}/enable [post]
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层启用插件
	err := h.pluginService.EnablePlugin(ctx, pluginID)
	if err != nil {
		h.logger.Error("启用插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginNotFound:
			response.NotFound(c, "插件不存在")
		case plugin.ErrPluginAlreadyEnabled:
			response.BadRequest(c, "插件已启用")
		default:
			response.InternalServerError(c, "启用插件失败")
		}
		return
	}

	response.SuccessWithMessage(c, "插件启用成功", nil)
}

// DisablePlugin 禁用插件
// @Summary 禁用插件
// @Description 禁用指定的插件
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id}/disable [post]
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层禁用插件
	err := h.pluginService.DisablePlugin(ctx, pluginID)
	if err != nil {
		h.logger.Error("禁用插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginNotFound:
			response.NotFound(c, "插件不存在")
		case plugin.ErrPluginAlreadyDisabled:
			response.BadRequest(c, "插件已禁用")
		default:
			response.InternalServerError(c, "禁用插件失败")
		}
		return
	}

	response.SuccessWithMessage(c, "插件禁用成功", nil)
}

// ConfigurePlugin 配置插件
// @Summary 配置插件
// @Description 更新插件配置
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param request body ConfigurePluginRequest true "配置插件请求"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id}/config [put]
func (h *PluginHandler) ConfigurePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	var req ConfigurePluginRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("配置插件请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.NewValidator(h.logger).Validate(req); err != nil {
		h.logger.Warn("配置插件请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.NewValidator(h.logger).TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	// 调用服务层更新插件配置
	err := h.pluginService.UpdatePluginConfig(ctx, pluginID, req.Config)
	if err != nil {
		h.logger.Error("更新插件配置失败", zap.String("plugin_id", pluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginNotFound:
			response.NotFound(c, "插件不存在")
		default:
			response.InternalServerError(c, "更新插件配置失败")
		}
		return
	}

	response.SuccessWithMessage(c, "插件配置更新成功", nil)
}

// ExecutePlugin 执行插件
// @Summary 执行插件
// @Description 执行插件的功能
// @Tags plugin
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Param request body ExecutePluginRequest true "执行插件请求"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/plugins/{id}/execute [post]
func (h *PluginHandler) ExecutePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		response.BadRequest(c, "插件ID不能为空")
		return
	}

	var req ExecutePluginRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("执行插件请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.NewValidator(h.logger).Validate(req); err != nil {
		h.logger.Warn("执行插件请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.NewValidator(h.logger).TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	// 调用服务层执行插件
	result, err := h.pluginService.ExecutePlugin(ctx, pluginID, req.Data)
	if err != nil {
		h.logger.Error("执行插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		
		// 根据错误类型返回不同的响应
		switch err {
		case plugin.ErrPluginNotFound:
			response.NotFound(c, "插件不存在")
		default:
			response.InternalServerError(c, "执行插件失败")
		}
		return
	}

	response.Success(c, result)
}