package plugin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/plugin"
	"moviepilot-go/pkg/logger"
)

// Handler 插件处理器
type Handler struct {
	logger        *zap.Logger
	pluginService plugin.Service
}

// NewHandler 创建插件处理器
func NewHandler(pluginService plugin.Service) *Handler {
	return &Handler{
		logger:        logger.GetLogger(),
		pluginService: pluginService,
	}
}

// AllPlugins 获取所有插件
// @Summary 获取所有插件清单
// @Description 查询所有插件清单，包括本地插件和在线插件，插件状态：installed, market, all
// @Tags plugin
// @Produce json
// @Param state query string false "插件状态：installed, market, all" default(all)
// @Param force query bool false "是否强制刷新" default(false)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/ [get]
func (h *Handler) AllPlugins(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	state := c.DefaultQuery("state", "all")
	force := c.DefaultQuery("force", "false") == "true"

	h.logger.Info("获取所有插件请求",
		zap.String("state", state),
		zap.Bool("force", force),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用插件服务获取插件列表
	plugins, err := h.pluginService.ListPlugins(c.Request.Context())
	if err != nil {
		h.logger.Error("获取插件列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取插件列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// InstalledPlugins 获取已安装插件
// @Summary 获取已安装插件
// @Description 查询用户已安装插件清单
// @Tags plugin
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/installed [get]
func (h *Handler) InstalledPlugins(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("获取已安装插件请求",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取已安装插件逻辑
	c.JSON(http.StatusOK, []string{})
}

// PluginStatistic 获取插件统计信息
// @Summary 插件安装统计
// @Description 插件安装统计
// @Tags plugin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/statistic [get]
func (h *Handler) PluginStatistic(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("获取插件统计信息请求",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件统计信息逻辑
	c.JSON(http.StatusOK, map[string]any{})
}

// ReloadPlugin 重新加载插件
// @Summary 重新加载插件
// @Description 重新加载插件
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/reload/{plugin_id} [get]
func (h *Handler) ReloadPlugin(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("重新加载插件请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用插件服务重新加载插件
	err := h.pluginService.ReloadPlugin(c.Request.Context(), pluginID)
	if err != nil {
		h.logger.Error("重新加载插件失败",
			zap.Error(err),
			zap.String("plugin_id", pluginID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "重新加载插件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件重新加载成功",
	})
}

// InstallPlugin 安装插件
// @Summary 安装插件
// @Description 安装插件
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Param repo_url query string false "仓库地址"
// @Param force query bool false "是否强制安装" default(false)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/install/{plugin_id} [get]
func (h *Handler) InstallPlugin(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")
	repoURL := c.DefaultQuery("repo_url", "")
	force := c.DefaultQuery("force", "false") == "true"

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("安装插件请求",
		zap.String("plugin_id", pluginID),
		zap.String("repo_url", repoURL),
		zap.Bool("force", force),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现安装插件逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件安装成功",
	})
}

// PluginRemotes 获取插件联邦组件列表
// @Summary 获取插件联邦组件列表
// @Description 获取插件联邦组件列表
// @Tags plugin
// @Produce json
// @Param token query string true "认证令牌"
// @Success 200 {array} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/remotes [get]
func (h *Handler) PluginRemotes(c *gin.Context) {
	reqID := c.GetString("request_id")
	token := c.Query("token")

	if token != "moviepilot" {
		h.logger.Error("无效的认证令牌",
			zap.String("token", token),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "禁止访问",
		})
		return
	}

	h.logger.Info("获取插件联邦组件列表请求",
		zap.String("request_id", reqID),
	)

	// TODO: 实现获取插件联邦组件列表逻辑
	c.JSON(http.StatusOK, []map[string]any{})
}

// PluginForm 获取插件表单页面
// @Summary 获取插件表单页面
// @Description 根据插件ID获取插件配置表单或Vue组件URL
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/form/{plugin_id} [get]
func (h *Handler) PluginForm(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("获取插件表单页面请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件表单页面逻辑
	c.JSON(http.StatusOK, map[string]any{
		"render_mode": "form",
		"conf":        map[string]any{},
		"model":       map[string]any{},
	})
}

// PluginPage 获取插件数据页面
// @Summary 获取插件数据页面
// @Description 根据插件ID获取插件数据页面
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/page/{plugin_id} [get]
func (h *Handler) PluginPage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("获取插件数据页面请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件数据页面逻辑
	c.JSON(http.StatusOK, map[string]any{
		"render_mode": "page",
		"page":        []any{},
	})
}

// PluginDashboardMeta 获取所有插件仪表板元信息
// @Summary 获取所有插件仪表板元信息
// @Description 获取所有插件仪表板元信息
// @Tags plugin
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/dashboard/meta [get]
func (h *Handler) PluginDashboardMeta(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("获取所有插件仪表板元信息请求",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取所有插件仪表板元信息逻辑
	c.JSON(http.StatusOK, []map[string]any{})
}

// PluginDashboardByKey 获取插件仪表板配置
// @Summary 获取插件仪表板配置
// @Description 根据插件ID和key获取插件仪表板配置
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Param key path string true "仪表板key"
// @Param user_agent header string false "用户代理"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/dashboard/{plugin_id}/{key} [get]
func (h *Handler) PluginDashboardByKey(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")
	key := c.Param("key")
	userAgent := c.GetHeader("User-Agent")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("获取插件仪表板配置请求",
		zap.String("plugin_id", pluginID),
		zap.String("key", key),
		zap.String("user_agent", userAgent),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件仪表板配置逻辑
	c.JSON(http.StatusOK, map[string]any{})
}

// PluginDashboard 获取插件仪表板配置
// @Summary 获取插件仪表板配置
// @Description 根据插件ID获取插件仪表板配置
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Param user_agent header string false "用户代理"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/dashboard/{plugin_id} [get]
func (h *Handler) PluginDashboard(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")
	userAgent := c.GetHeader("User-Agent")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("获取插件仪表板配置请求",
		zap.String("plugin_id", pluginID),
		zap.String("user_agent", userAgent),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件仪表板配置逻辑
	c.JSON(http.StatusOK, map[string]any{})
}

// ResetPlugin 重置插件配置及数据
// @Summary 重置插件配置及数据
// @Description 根据插件ID重置插件配置及数据
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/reset/{plugin_id} [get]
func (h *Handler) ResetPlugin(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("重置插件配置及数据请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现重置插件配置及数据逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件配置及数据重置成功",
	})
}

// PluginStaticFile 获取插件静态文件
// @Summary 获取插件静态文件
// @Description 获取插件静态文件
// @Tags plugin
// @Produce octet-stream
// @Param plugin_id path string true "插件ID"
// @Param filepath path string true "文件路径"
// @Success 200 {file} file
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/file/{plugin_id}/{filepath:path} [get]
func (h *Handler) PluginStaticFile(c *gin.Context) {
	reqID := c.GetString("request_id")
	pluginID := c.Param("plugin_id")
	filepath := c.Param("filepath")

	// 基础安全检查
	if strings.Contains(filepath, "..") || strings.Contains(pluginID, "..") {
		h.logger.Warn("检测到路径遍历尝试",
			zap.String("plugin_id", pluginID),
			zap.String("filepath", filepath),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "禁止访问",
		})
		return
	}

	h.logger.Info("获取插件静态文件请求",
		zap.String("plugin_id", pluginID),
		zap.String("filepath", filepath),
		zap.String("request_id", reqID),
	)

	// TODO: 实现获取插件静态文件逻辑
	c.JSON(http.StatusNotFound, gin.H{
		"error": "文件不存在",
	})
}

// GetPluginFolders 获取插件文件夹配置
// @Summary 获取插件文件夹配置
// @Description 获取插件文件夹分组配置
// @Tags plugin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/folders [get]
func (h *Handler) GetPluginFolders(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("获取插件文件夹配置请求",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取插件文件夹配置逻辑
	c.JSON(http.StatusOK, map[string]any{})
}

// SavePluginFolders 保存插件文件夹配置
// @Summary 保存插件文件夹配置
// @Description 保存插件文件夹分组配置
// @Tags plugin
// @Accept json
// @Produce json
// @Param folders body map[string]interface{} true "文件夹配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/folders [post]
func (h *Handler) SavePluginFolders(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	var folders map[string]any
	if err := c.ShouldBindJSON(&folders); err != nil {
		h.logger.Error("无效的文件夹配置",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.logger.Info("保存插件文件夹配置请求",
		zap.Any("folders", folders),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现保存插件文件夹配置逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件文件夹配置保存成功",
	})
}

// CreatePluginFolder 创建插件文件夹
// @Summary 创建插件文件夹
// @Description 创建新的插件文件夹
// @Tags plugin
// @Produce json
// @Param folder_name path string true "文件夹名称"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/folders/{folder_name} [post]
func (h *Handler) CreatePluginFolder(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	folderName := c.Param("folder_name")

	if folderName == "" {
		h.logger.Error("缺少文件夹名称",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件夹名称不能为空",
		})
		return
	}

	h.logger.Info("创建插件文件夹请求",
		zap.String("folder_name", folderName),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现创建插件文件夹逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件文件夹创建成功",
	})
}

// DeletePluginFolder 删除插件文件夹
// @Summary 删除插件文件夹
// @Description 删除插件文件夹
// @Tags plugin
// @Produce json
// @Param folder_name path string true "文件夹名称"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/folders/{folder_name} [delete]
func (h *Handler) DeletePluginFolder(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	folderName := c.Param("folder_name")

	if folderName == "" {
		h.logger.Error("缺少文件夹名称",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件夹名称不能为空",
		})
		return
	}

	h.logger.Info("删除插件文件夹请求",
		zap.String("folder_name", folderName),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现删除插件文件夹逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件文件夹删除成功",
	})
}

// UpdateFolderPlugins 更新文件夹中的插件
// @Summary 更新文件夹中的插件
// @Description 更新指定文件夹中的插件列表
// @Tags plugin
// @Accept json
// @Produce json
// @Param folder_name path string true "文件夹名称"
// @Param plugin_ids body []string true "插件ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/folders/{folder_name}/plugins [put]
func (h *Handler) UpdateFolderPlugins(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	folderName := c.Param("folder_name")

	var pluginIDs []string
	if err := c.ShouldBindJSON(&pluginIDs); err != nil {
		h.logger.Error("无效的插件ID列表",
			zap.Error(err),
			zap.String("folder_name", folderName),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.logger.Info("更新文件夹中的插件请求",
		zap.String("folder_name", folderName),
		zap.Int("plugin_count", len(pluginIDs)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现更新文件夹中的插件逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件夹中的插件已更新",
	})
}

// ClonePlugin 创建插件分身
// @Summary 创建插件分身
// @Description 创建插件分身
// @Tags plugin
// @Accept json
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Param clone_data body map[string]interface{} true "分身数据"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/clone/{plugin_id} [post]
func (h *Handler) ClonePlugin(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	var cloneData map[string]any
	if err := c.ShouldBindJSON(&cloneData); err != nil {
		h.logger.Error("无效的分身数据",
			zap.Error(err),
			zap.String("plugin_id", pluginID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.logger.Info("创建插件分身请求",
		zap.String("plugin_id", pluginID),
		zap.Any("clone_data", cloneData),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现创建插件分身逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件分身创建成功",
	})
}

// GetPluginConfig 获取插件配置
// @Summary 获取插件配置
// @Description 根据插件ID获取插件配置信息
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/{plugin_id} [get]
func (h *Handler) GetPluginConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("获取插件配置请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用插件服务获取插件配置
	config, err := h.pluginService.GetPluginConfig(c.Request.Context(), pluginID)
	if err != nil {
		h.logger.Error("获取插件配置失败",
			zap.Error(err),
			zap.String("plugin_id", pluginID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取插件配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdatePluginConfig 更新插件配置
// @Summary 更新插件配置
// @Description 更新插件配置
// @Tags plugin
// @Accept json
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Param conf body map[string]interface{} true "插件配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/{plugin_id} [put]
func (h *Handler) UpdatePluginConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	var conf map[string]any
	if err := c.ShouldBindJSON(&conf); err != nil {
		h.logger.Error("无效的插件配置",
			zap.Error(err),
			zap.String("plugin_id", pluginID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.logger.Info("更新插件配置请求",
		zap.String("plugin_id", pluginID),
		zap.Any("config", conf),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用插件服务更新插件配置
	err := h.pluginService.ConfigurePlugin(c.Request.Context(), pluginID, conf)
	if err != nil {
		h.logger.Error("更新插件配置失败",
			zap.Error(err),
			zap.String("plugin_id", pluginID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新插件配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件配置更新成功",
	})
}

// UninstallPlugin 卸载插件
// @Summary 卸载插件
// @Description 卸载插件
// @Tags plugin
// @Produce json
// @Param plugin_id path string true "插件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/plugin/{plugin_id} [delete]
func (h *Handler) UninstallPlugin(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	pluginID := c.Param("plugin_id")

	if pluginID == "" {
		h.logger.Error("缺少插件ID",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "插件ID不能为空",
		})
		return
	}

	h.logger.Info("卸载插件请求",
		zap.String("plugin_id", pluginID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现卸载插件逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件卸载成功",
	})
}
