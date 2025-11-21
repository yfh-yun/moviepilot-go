package actions

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// PluginHandler 插件调用HTTP处理器接口
type PluginHandler interface {
	// InvokePlugin 调用插件方法
	InvokePlugin(c *gin.Context)
	// GetPluginInfo 获取插件信息
	GetPluginInfo(c *gin.Context)
	// ListPlugins 列出所有可用插件
	ListPlugins(c *gin.Context)
	// CheckPluginStatus 检查插件状态
	CheckPluginStatus(c *gin.Context)
	// GetPluginHistory 获取插件调用历史
	GetPluginHistory(c *gin.Context)
	// BatchInvokePlugins 批量调用插件
	BatchInvokePlugins(c *gin.Context)
}

// pluginHandler 插件调用HTTP处理器实现
type pluginHandler struct {
	pluginInvoker PluginInvoker
	logger        logger.Logger
}

// NewPluginHandler 创建插件调用HTTP处理器实例
func NewPluginHandler(pluginInvoker PluginInvoker, logger logger.Logger) PluginHandler {
	return &pluginHandler{
		pluginInvoker: pluginInvoker,
		logger:        logger,
	}
}

// @Summary 调用插件方法
// @Description 调用指定插件的方法并返回结果
// @Tags plugins
// @Accept json
// @Produce json
// @Param request body PluginInvokeRequest true "插件调用请求"
// @Success 200 {object} PluginInvokeResponse "调用成功"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 401 {object} response.ErrorResponse "未授权"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins/invoke [post]
func (h *pluginHandler) InvokePlugin(c *gin.Context) {
	h.logger.Debug("接收插件调用请求")

	var req PluginInvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("请求参数解析失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 设置默认调用者
	if req.Caller == "" {
		req.Caller = "system"
	}

	// 设置默认超时
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}

	// 调用插件
	result, err := h.pluginInvoker.InvokePlugin(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("插件调用失败", 
			"plugin_id", req.PluginID,
			"method", req.Method,
			"error", err.Error(),
		)
		response.Error(c, http.StatusInternalServerError, "插件调用失败", err.Error())
		return
	}

	h.logger.Info("插件调用成功", 
		"plugin_id", req.PluginID,
		"method", req.Method,
	)
	response.Success(c, result)
}

// @Summary 获取插件信息
// @Description 获取指定插件的详细信息
// @Tags plugins
// @Accept json
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} PluginInfo "插件信息"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 404 {object} response.ErrorResponse "插件不存在"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins/{plugin_id} [get]
func (h *pluginHandler) GetPluginInfo(c *gin.Context) {
	pluginID := c.Param("plugin_id")
	h.logger.Debug("获取插件信息", "plugin_id", pluginID)

	info, err := h.pluginInvoker.GetPluginInfo(c.Request.Context(), pluginID)
	if err != nil {
		h.logger.Error("获取插件信息失败", "plugin_id", pluginID, "error", err.Error())
		response.Error(c, http.StatusNotFound, "获取插件信息失败", err.Error())
		return
	}

	response.Success(c, info)
}

// @Summary 列出所有可用插件
// @Description 列出系统中所有可用的插件，支持按类型过滤
// @Tags plugins
// @Accept json
// @Produce json
// @Param type query string false "插件类型"
// @Success 200 {array} PluginInfo "插件列表"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins [get]
func (h *pluginHandler) ListPlugins(c *gin.Context) {
	pluginType := c.Query("type")
	h.logger.Debug("列出插件", "type", pluginType)

	plugins, err := h.pluginInvoker.ListPlugins(c.Request.Context(), pluginType)
	if err != nil {
		h.logger.Error("列出插件失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, "列出插件失败", err.Error())
		return
	}

	response.Success(c, plugins)
}

// @Summary 检查插件状态
// @Description 检查指定插件的运行状态和健康情况
// @Tags plugins
// @Accept json
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} PluginStatusInfo "插件状态信息"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 404 {object} response.ErrorResponse "插件不存在"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins/{plugin_id}/status [get]
func (h *pluginHandler) CheckPluginStatus(c *gin.Context) {
	pluginID := c.Param("plugin_id")
	h.logger.Debug("检查插件状态", "plugin_id", pluginID)

	status, err := h.pluginInvoker.CheckPluginStatus(c.Request.Context(), pluginID)
	if err != nil {
		h.logger.Error("检查插件状态失败", "plugin_id", pluginID, "error", err.Error())
		response.Error(c, http.StatusInternalServerError, "检查插件状态失败", err.Error())
		return
	}

	response.Success(c, status)
}

// @Summary 获取插件调用历史
// @Description 获取插件的调用历史记录，支持分页和过滤
// @Tags plugins
// @Accept json
// @Produce json
// @Param request body PluginHistoryParams true "查询参数"
// @Success 200 {object} PluginHistoryResponse "历史记录"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins/history [post]
func (h *pluginHandler) GetPluginHistory(c *gin.Context) {
	var params PluginHistoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("请求参数解析失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}

	history, err := h.pluginInvoker.GetPluginHistory(c.Request.Context(), &params)
	if err != nil {
		h.logger.Error("获取插件历史失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, "获取插件历史失败", err.Error())
		return
	}

	response.Success(c, history)
}

// @Summary 批量调用插件
// @Description 批量调用多个插件方法，并行执行
// @Tags plugins
// @Accept json
// @Produce json
// @Param requests body []PluginInvokeRequest true "批量调用请求"
// @Success 200 {array} PluginInvokeResponse "调用结果列表"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/plugins/batch-invoke [post]
func (h *pluginHandler) BatchInvokePlugins(c *gin.Context) {
	var requests []PluginInvokeRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		h.logger.Error("请求参数解析失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 验证请求数量
	if len(requests) == 0 || len(requests) > 10 {
		h.logger.Error("批量调用请求数量无效", "count", len(requests))
		response.Error(c, http.StatusBadRequest, "批量调用请求数量必须在1-10之间")
		return
	}

	// 并行调用多个插件
	results := make([]*PluginInvokeResponse, 0, len(requests))
	for _, req := range requests {
		// 设置默认值
		reqCopy := req // 复制请求避免闭包问题
		if reqCopy.Caller == "" {
			reqCopy.Caller = "system"
		}
		if reqCopy.Timeout == 0 {
			reqCopy.Timeout = 30 * time.Second
		}

		// 调用插件
		result, err := h.pluginInvoker.InvokePlugin(c.Request.Context(), &reqCopy)
		if err != nil {
			h.logger.Error("插件批量调用失败", 
				"plugin_id", reqCopy.PluginID,
				"method", reqCopy.Method,
				"error", err.Error(),
			)
			// 即使单个失败也继续处理其他请求
			results = append(results, &PluginInvokeResponse{
				PluginID: reqCopy.PluginID,
				Method:   reqCopy.Method,
				Error:    err.Error(),
			})
		} else {
			results = append(results, result)
		}
	}

	response.Success(c, results)
}
