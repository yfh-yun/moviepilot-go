package scheduler

import (
	"net/http"

	"github.com/yfh-yun/moviepilot-go/internal/scheduler"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 调度器处理器
type Handler struct {
	logger           *zap.Logger
	schedulerService *scheduler.SchedulerService
}

// NewHandler 创建新的调度器处理器
func NewHandler(logger *zap.Logger, schedulerService *scheduler.SchedulerService) *Handler {
	return &Handler{
		logger:           logger,
		schedulerService: schedulerService,
	}
}

// ListJobs 获取任务列表
// @Summary 获取任务列表
// @Description 获取所有已配置的任务列表
// @Tags 调度器
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]scheduler.JobConfig}
// @Router /api/v1/scheduler/jobs [get]
func (h *Handler) ListJobs(c *gin.Context) {
	jobs := h.schedulerService.ListJobStatus()

	h.logger.Info("获取任务列表", zap.Int("count", len(jobs)))
	response.Success(c, jobs)
}

// GetJobStatus 获取任务状态
// @Summary 获取任务状态
// @Description 获取指定任务的状态信息
// @Tags 调度器
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} response.Response{data=scheduler.JobResult}
// @Failure 404 {object} response.Response
// @Router /api/v1/scheduler/jobs/{id} [get]
func (h *Handler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")

	if err := validator.Validate.Var(jobID, "required,uuid"); err != nil {
		response.Error(c, http.StatusBadRequest, "任务ID格式无效")
		return
	}

	// 这里需要实现获取任务状态的方法
	// status, err := h.schedulerService.GetJobStatus(jobID)
	// if err != nil {
	// 	h.logger.Error("获取任务状态失败", zap.String("job_id", jobID), zap.Error(err))
	// 	response.Error(c, http.StatusNotFound, "任务不存在")
	// 	return
	// }

	h.logger.Info("获取任务状态", zap.String("job_id", jobID))
	response.Success(c, gin.H{
		"job_id": jobID,
		"status": "running",
	})
}

// EnableJob 启用任务
// @Summary 启用任务
// @Description 启用指定的任务
// @Tags 调度器
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/scheduler/jobs/{id}/enable [post]
func (h *Handler) EnableJob(c *gin.Context) {
	jobID := c.Param("id")

	if err := validator.Validate.Var(jobID, "required,uuid"); err != nil {
		response.Error(c, http.StatusBadRequest, "任务ID格式无效")
		return
	}

	if err := h.schedulerService.EnableJob(jobID); err != nil {
		h.logger.Error("启用任务失败", zap.String("job_id", jobID), zap.Error(err))
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	h.logger.Info("启用任务", zap.String("job_id", jobID))
	response.Success(c, nil)
}

// DisableJob 禁用任务
// @Summary 禁用任务
// @Description 禁用指定的任务
// @Tags 调度器
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/scheduler/jobs/{id}/disable [post]
func (h *Handler) DisableJob(c *gin.Context) {
	jobID := c.Param("id")

	if err := validator.Validate.Var(jobID, "required,uuid"); err != nil {
		response.Error(c, http.StatusBadRequest, "任务ID格式无效")
		return
	}

	if err := h.schedulerService.DisableJob(jobID); err != nil {
		h.logger.Error("禁用任务失败", zap.String("job_id", jobID), zap.Error(err))
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	h.logger.Info("禁用任务", zap.String("job_id", jobID))
	response.Success(c, nil)
}

// ListWorkflowInstances 获取工作流实例列表
// @Summary 获取工作流实例列表
// @Description 获取所有工作流实例的列表
// @Tags 调度器
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]scheduler.WorkflowInstance}
// @Router /api/v1/scheduler/workflows/instances [get]
func (h *Handler) ListWorkflowInstances(c *gin.Context) {
	instances := h.schedulerService.ListWorkflowInstances()

	h.logger.Info("获取工作流实例列表", zap.Int("count", len(instances)))
	response.Success(c, instances)
}

// GetWorkflowInstance 获取工作流实例详情
// @Summary 获取工作流实例详情
// @Description 获取指定工作流实例的详细信息
// @Tags 调度器
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} response.Response{data=scheduler.WorkflowInstance}
// @Failure 404 {object} response.Response
// @Router /api/v1/scheduler/workflows/instances/{id} [get]
func (h *Handler) GetWorkflowInstance(c *gin.Context) {
	instanceID := c.Param("id")

	if err := validator.Validate.Var(instanceID, "required,uuid"); err != nil {
		response.Error(c, http.StatusBadRequest, "实例ID格式无效")
		return
	}

	instance, err := h.schedulerService.GetWorkflowInstance(instanceID)
	if err != nil {
		h.logger.Error("获取工作流实例失败", zap.String("instance_id", instanceID), zap.Error(err))
		response.Error(c, http.StatusNotFound, "工作流实例不存在")
		return
	}

	h.logger.Info("获取工作流实例", zap.String("instance_id", instanceID))
	response.Success(c, instance)
}

// ExecuteWorkflow 执行工作流
// @Summary 执行工作流
// @Description 手动执行指定的工作流
// @Tags 调度器
// @Accept json
// @Produce json
// @Param id path string true "工作流ID"
// @Param data body ExecuteWorkflowRequest true "执行参数"
// @Success 200 {object} response.Response{data=scheduler.WorkflowInstance}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/scheduler/workflows/{id}/execute [post]
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	if err := validator.Validate.Var(workflowID, "required,uuid"); err != nil {
		response.Error(c, http.StatusBadRequest, "工作流ID格式无效")
		return
	}

	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("解析请求参数失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数格式无效")
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	instance, err := h.schedulerService.ExecuteWorkflow(c.Request.Context(), workflowID, req.TriggerData)
	if err != nil {
		h.logger.Error("执行工作流失败", zap.String("workflow_id", workflowID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "执行工作流失败")
		return
	}

	h.logger.Info("执行工作流", zap.String("workflow_id", workflowID), zap.String("instance_id", instance.ID))
	response.Success(c, instance)
}

// ExecuteWorkflowRequest 执行工作流请求
type ExecuteWorkflowRequest struct {
	TriggerData map[string]interface{} `json:"trigger_data" validate:"required"`
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	schedulerGroup := rg.Group("/scheduler")
	{
		schedulerGroup.GET("/jobs", h.ListJobs)
		schedulerGroup.GET("/jobs/:id", h.GetJobStatus)
		schedulerGroup.POST("/jobs/:id/enable", h.EnableJob)
		schedulerGroup.POST("/jobs/:id/disable", h.DisableJob)

		schedulerGroup.GET("/workflows/instances", h.ListWorkflowInstances)
		schedulerGroup.GET("/workflows/instances/:id", h.GetWorkflowInstance)
		schedulerGroup.POST("/workflows/:id/execute", h.ExecuteWorkflow)
	}
}
