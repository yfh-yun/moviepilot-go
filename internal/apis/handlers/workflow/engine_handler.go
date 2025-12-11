package workflowapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	workflowbiz "moviepilot-go/internal/business/services/workflow"
	"moviepilot-go/pkg/logger"
)

// EngineHandler 工作流引擎 API 处理器
type EngineHandler struct {
	engine workflowbiz.Engine
	logger *zap.Logger
}

// NewEngineHandler 创建引擎处理器
func NewEngineHandler(engine workflowbiz.Engine) *EngineHandler {
	return &EngineHandler{
		engine: engine,
		logger: logger.GetLogger(),
	}
}

// Execute 执行工作流
// @Summary 执行工作流
// @Description 同步执行工作流
// @Tags workflow
// @Accept json
// @Produce json
// @Param workflow body workflowbiz.Workflow true "工作流定义"
// @Success 200 {object} workflowbiz.ExecutionResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/execute [post]
func (h *EngineHandler) Execute(c *gin.Context) {
	var workflow workflowbiz.Workflow

	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.engine.Execute(c.Request.Context(), &workflow)
	if err != nil {
		h.logger.Error("执行工作流失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExecuteAsync 异步执行工作流
// @Summary 异步执行工作流
// @Description 异步执行工作流，立即返回执行ID
// @Tags workflow
// @Accept json
// @Produce json
// @Param workflow body workflowbiz.Workflow true "工作流定义"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/execute/async [post]
func (h *EngineHandler) ExecuteAsync(c *gin.Context) {
	var workflow workflowbiz.Workflow

	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	executionID, err := h.engine.ExecuteAsync(c.Request.Context(), &workflow)
	if err != nil {
		h.logger.Error("异步执行工作流失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": executionID,
		"message":      "工作流已开始执行",
	})
}

// GetExecution 获取执行状态
// @Summary 获取执行状态
// @Description 获取工作流执行的当前状态
// @Tags workflow
// @Produce json
// @Param execution_id path string true "执行ID"
// @Success 200 {object} workflowbiz.Execution
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/execution/{execution_id} [get]
func (h *EngineHandler) GetExecution(c *gin.Context) {
	executionID := c.Param("execution_id")

	execution, err := h.engine.GetExecution(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("获取执行状态失败", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// CancelExecution 取消执行
// @Summary 取消执行
// @Description 取消正在执行的工作流
// @Tags workflow
// @Produce json
// @Param execution_id path string true "执行ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/workflow/execution/{execution_id}/cancel [post]
func (h *EngineHandler) CancelExecution(c *gin.Context) {
	executionID := c.Param("execution_id")

	if err := h.engine.CancelExecution(c.Request.Context(), executionID); err != nil {
		h.logger.Error("取消执行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "执行已取消"})
}
