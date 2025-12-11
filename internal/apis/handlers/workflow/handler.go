package workflowapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// WorkflowService 定义 handler 所需的最小接口，方便测试替换。
type WorkflowService interface {
	StartLocalFileWorkflow(ctx context.Context, req StartLocalFileWorkflowRequest) (*StartLocalFileWorkflowResponse, error)
	ListWorkflows(ctx context.Context) ([]dto.Workflow, error)
	CreateWorkflow(ctx context.Context, workflow dto.Workflow) error
	GetPluginActions(ctx context.Context, pluginID string) ([]map[string]any, error)
	ListActions(ctx context.Context) ([]map[string]any, error)
	GetEventTypes(ctx context.Context) ([]map[string]any, error)
	ShareWorkflow(ctx context.Context, share dto.WorkflowShare) error
	DeleteShare(ctx context.Context, shareID int) error
	ForkWorkflow(ctx context.Context, workflow dto.WorkflowShare) error
	GetShares(ctx context.Context, name string, page, count int) ([]dto.WorkflowShare, error)
	RunWorkflow(ctx context.Context, workflowID int, fromBegin bool) error
	StartWorkflow(ctx context.Context, workflowID int) error
	PauseWorkflow(ctx context.Context, workflowID int) error
	ResetWorkflow(ctx context.Context, workflowID int) error
	GetWorkflow(ctx context.Context, workflowID int) (dto.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflowID int, workflow dto.Workflow) error
	DeleteWorkflow(ctx context.Context, workflowID int) error
}

// Handler 负责处理 workflow 相关 API。
type Handler struct {
	service WorkflowService
	logger  *zap.Logger
}

// NewHandler 创建 Handler。
func NewHandler(service WorkflowService, log *zap.Logger) *Handler {
	if log == nil {
		log = logger.GetLogger()
	}
	return &Handler{service: service, logger: log}
}

// StartLocalFileWorkflow 触发“本地文件 → 刮削 → 转移”链路。
func (h *Handler) StartLocalFileWorkflow(c *gin.Context) {
	var req StartLocalFileWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("workflow request validation failed",
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("workflow request accepted",
		zap.String("path", c.FullPath()),
		zap.String("root_path", req.RootPath),
		zap.Bool("include_fetch", req.IncludeFetch),
		zap.Bool("wait_for_completion", req.WaitForCompletion),
	)

	result, err := h.service.StartLocalFileWorkflow(ctx, req)
	if err != nil {
		h.logger.Error("workflow start failed",
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "workflow start failed",
		})
		return
	}

	statusCode := http.StatusAccepted
	message := "workflow started"
	if req.WaitForCompletion {
		statusCode = http.StatusOK
		if result != nil && result.Message != "" {
			message = result.Message
		} else {
			message = "workflow completed"
		}
	}

	c.JSON(statusCode, gin.H{
		"success":   true,
		"code":      200,
		"message":   message,
		"data":      result,
		"timestamp": time.Now().Unix(),
	})
}

// ListWorkflows 获取工作流列表
// @Summary 所有工作流
// @Description 获取工作流列表
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.Workflow
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow [get]
func (h *Handler) ListWorkflows(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API ListWorkflows started", zap.String("request_id", reqID))

	ctx := c.Request.Context()

	workflows, err := h.service.ListWorkflows(ctx)
	if err != nil {
		h.logger.Error("Workflow API ListWorkflows failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取工作流列表失败",
		})
		return
	}

	h.logger.Info("Workflow API ListWorkflows succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_count", len(workflows)),
	)

	c.JSON(http.StatusOK, workflows)
}

// CreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建工作流
// @Tags workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param workflow body dto.Workflow true "工作流信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow [post]
func (h *Handler) CreateWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API CreateWorkflow started", zap.String("request_id", reqID))

	var workflow dto.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		h.logger.Warn("Workflow API CreateWorkflow invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.CreateWorkflow(ctx, workflow); err != nil {
		h.logger.Error("Workflow API CreateWorkflow failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "创建工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API CreateWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.String("workflow_name", workflow.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "创建工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// GetPluginActions 查询插件动作
// @Summary 查询插件动作
// @Description 获取所有动作
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param plugin_id query string false "插件ID"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/plugin/actions [get]
func (h *Handler) GetPluginActions(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API GetPluginActions started", zap.String("request_id", reqID))

	pluginID := c.Query("plugin_id")

	ctx := c.Request.Context()

	actions, err := h.service.GetPluginActions(ctx, pluginID)
	if err != nil {
		h.logger.Error("Workflow API GetPluginActions failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取插件动作失败",
		})
		return
	}

	h.logger.Info("Workflow API GetPluginActions succeeded",
		zap.String("request_id", reqID),
		zap.String("plugin_id", pluginID),
		zap.Int("action_count", len(actions)),
	)

	c.JSON(http.StatusOK, actions)
}

// ListActions 所有动作
// @Summary 所有动作
// @Description 获取所有动作
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/actions [get]
func (h *Handler) ListActions(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API ListActions started", zap.String("request_id", reqID))

	ctx := c.Request.Context()

	actions, err := h.service.ListActions(ctx)
	if err != nil {
		h.logger.Error("Workflow API ListActions failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取动作列表失败",
		})
		return
	}

	h.logger.Info("Workflow API ListActions succeeded",
		zap.String("request_id", reqID),
		zap.Int("action_count", len(actions)),
	)

	c.JSON(http.StatusOK, actions)
}

// GetEventTypes 获取所有事件类型
// @Summary 获取所有事件类型
// @Description 获取所有事件类型
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/event_types [get]
func (h *Handler) GetEventTypes(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API GetEventTypes started", zap.String("request_id", reqID))

	ctx := c.Request.Context()

	eventTypes, err := h.service.GetEventTypes(ctx)
	if err != nil {
		h.logger.Error("Workflow API GetEventTypes failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取事件类型失败",
		})
		return
	}

	h.logger.Info("Workflow API GetEventTypes succeeded",
		zap.String("request_id", reqID),
		zap.Int("event_type_count", len(eventTypes)),
	)

	c.JSON(http.StatusOK, eventTypes)
}

// ShareWorkflow 分享工作流
// @Summary 分享工作流
// @Description 分享工作流
// @Tags workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param workflow body dto.WorkflowShare true "工作流分享信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/share [post]
func (h *Handler) ShareWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API ShareWorkflow started", zap.String("request_id", reqID))

	var share dto.WorkflowShare
	if err := c.ShouldBindJSON(&share); err != nil {
		h.logger.Warn("Workflow API ShareWorkflow invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.ShareWorkflow(ctx, share); err != nil {
		h.logger.Error("Workflow API ShareWorkflow failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "分享工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API ShareWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", share.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "分享工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// DeleteShare 删除分享
// @Summary 删除分享
// @Description 删除分享
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param share_id path int true "分享ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/share/{share_id} [delete]
func (h *Handler) DeleteShare(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API DeleteShare started", zap.String("request_id", reqID))

	shareIDStr := c.Param("share_id")
	shareID, err := strconv.Atoi(shareIDStr)
	if err != nil {
		h.logger.Warn("Workflow API DeleteShare invalid share_id",
			zap.String("request_id", reqID),
			zap.String("share_id", shareIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的分享ID",
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.DeleteShare(ctx, shareID); err != nil {
		h.logger.Error("Workflow API DeleteShare failed",
			zap.String("request_id", reqID),
			zap.Int("share_id", shareID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "删除分享失败",
		})
		return
	}

	h.logger.Info("Workflow API DeleteShare succeeded",
		zap.String("request_id", reqID),
		zap.Int("share_id", shareID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "删除分享成功",
		"timestamp": time.Now().Unix(),
	})
}

// ForkWorkflow 复用工作流
// @Summary 复用工作流
// @Description 复用工作流
// @Tags workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param workflow body dto.WorkflowShare true "工作流复用信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/fork [post]
func (h *Handler) ForkWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API ForkWorkflow started", zap.String("request_id", reqID))

	var workflow dto.WorkflowShare
	if err := c.ShouldBindJSON(&workflow); err != nil {
		h.logger.Warn("Workflow API ForkWorkflow invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.ForkWorkflow(ctx, workflow); err != nil {
		h.logger.Error("Workflow API ForkWorkflow failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "复用工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API ForkWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.String("workflow_name", workflow.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "复用工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// GetShares 查询分享的工作流
// @Summary 查询分享的工作流
// @Description 查询分享的工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param name query string false "工作流名称"
// @Param page query int false "页码"
// @Param count query int false "每页数量"
// @Success 200 {array} dto.WorkflowShare
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/shares [get]
func (h *Handler) GetShares(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API GetShares started", zap.String("request_id", reqID))

	name := c.Query("name")
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	countStr := c.DefaultQuery("count", "30")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 30
	}

	ctx := c.Request.Context()

	shares, err := h.service.GetShares(ctx, name, page, count)
	if err != nil {
		h.logger.Error("Workflow API GetShares failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取分享工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API GetShares succeeded",
		zap.String("request_id", reqID),
		zap.String("name", name),
		zap.Int("page", page),
		zap.Int("count", count),
		zap.Int("share_count", len(shares)),
	)

	c.JSON(http.StatusOK, shares)
}

// RunWorkflow 执行工作流
// @Summary 执行工作流
// @Description 执行工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Param from_begin query bool false "是否从头开始"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id}/run [post]
func (h *Handler) RunWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API RunWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API RunWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	fromBeginStr := c.DefaultQuery("from_begin", "true")
	fromBegin := fromBeginStr == "true"

	ctx := c.Request.Context()

	if err := h.service.RunWorkflow(ctx, workflowID, fromBegin); err != nil {
		h.logger.Error("Workflow API RunWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Bool("from_begin", fromBegin),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "执行工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API RunWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
		zap.Bool("from_begin", fromBegin),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "执行工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// StartWorkflow 启用工作流
// @Summary 启用工作流
// @Description 启用工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id}/start [post]
func (h *Handler) StartWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API StartWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API StartWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.StartWorkflow(ctx, workflowID); err != nil {
		h.logger.Error("Workflow API StartWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "启用工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API StartWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "启用工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// PauseWorkflow 停用工作流
// @Summary 停用工作流
// @Description 停用工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id}/pause [post]
func (h *Handler) PauseWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API PauseWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API PauseWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.PauseWorkflow(ctx, workflowID); err != nil {
		h.logger.Error("Workflow API PauseWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "停用工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API PauseWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "停用工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// ResetWorkflow 重置工作流
// @Summary 重置工作流
// @Description 重置工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id}/reset [post]
func (h *Handler) ResetWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API ResetWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API ResetWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.ResetWorkflow(ctx, workflowID); err != nil {
		h.logger.Error("Workflow API ResetWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "重置工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API ResetWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "重置工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// GetWorkflow 工作流详情
// @Summary 工作流详情
// @Description 获取工作流详情
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} dto.Workflow
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id} [get]
func (h *Handler) GetWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API GetWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API GetWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	ctx := c.Request.Context()

	workflow, err := h.service.GetWorkflow(ctx, workflowID)
	if err != nil {
		h.logger.Error("Workflow API GetWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "获取工作流详情失败",
		})
		return
	}

	h.logger.Info("Workflow API GetWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
		zap.String("workflow_name", workflow.Name),
	)

	c.JSON(http.StatusOK, workflow)
}

// UpdateWorkflow 更新工作流
// @Summary 更新工作流
// @Description 更新工作流
// @Tags workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Param workflow body dto.Workflow true "工作流信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id} [put]
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API UpdateWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API UpdateWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	var workflow dto.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		h.logger.Warn("Workflow API UpdateWorkflow invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.UpdateWorkflow(ctx, workflowID, workflow); err != nil {
		h.logger.Error("Workflow API UpdateWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "更新工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API UpdateWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
		zap.String("workflow_name", workflow.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "更新工作流成功",
		"timestamp": time.Now().Unix(),
	})
}

// DeleteWorkflow 删除工作流
// @Summary 删除工作流
// @Description 删除工作流
// @Tags workflow
// @Security BearerAuth
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/{workflow_id} [delete]
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Workflow API DeleteWorkflow started", zap.String("request_id", reqID))

	workflowIDStr := c.Param("workflow_id")
	workflowID, err := strconv.Atoi(workflowIDStr)
	if err != nil {
		h.logger.Warn("Workflow API DeleteWorkflow invalid workflow_id",
			zap.String("request_id", reqID),
			zap.String("workflow_id", workflowIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    400,
			"message": "无效的工作流ID",
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.service.DeleteWorkflow(ctx, workflowID); err != nil {
		h.logger.Error("Workflow API DeleteWorkflow failed",
			zap.String("request_id", reqID),
			zap.Int("workflow_id", workflowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    500,
			"message": "删除工作流失败",
		})
		return
	}

	h.logger.Info("Workflow API DeleteWorkflow succeeded",
		zap.String("request_id", reqID),
		zap.Int("workflow_id", workflowID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"code":      200,
		"message":   "删除工作流成功",
		"timestamp": time.Now().Unix(),
	})
}
