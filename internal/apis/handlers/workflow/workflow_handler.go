package workflow

import (
	"github.com/yfh-yun/moviepilot-go/internal/business/services/workflow"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 工作流API处理器
type Handler struct {
	workflowService *workflow.WorkflowService
}

// NewHandler 创建工作流处理器
func NewHandler(workflowService *workflow.WorkflowService) *Handler {
	return &Handler{
		workflowService: workflowService,
	}
}

// CreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建新的工作流配置
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.CreateWorkflowRequest true "创建工作流请求"
// @Success 200 {object} response.Response{data=workflow.WorkflowResponse} "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow [post]
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req workflow.CreateWorkflowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数解析错误", err)
		return
	}

	// 验证参数
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数验证失败", err)
		return
	}

	// 调用服务层
	result, err := h.workflowService.CreateWorkflow(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建工作流失败", err)
		return
	}

	response.Success(c, result)
}

// GetWorkflow 获取工作流详情
// @Summary 获取工作流详情
// @Description 根据ID获取工作流详细信息
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response{data=workflow.WorkflowResponse} "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow/{id} [get]
func (h *Handler) GetWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的工作流ID", err)
		return
	}

	// 调用服务层
	result, err := h.workflowService.GetWorkflow(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "workflow not found" {
			response.Error(c, http.StatusNotFound, "工作流不存在", nil)
		} else {
			response.Error(c, http.StatusInternalServerError, "获取工作流失败", err)
		}
		return
	}

	response.Success(c, result)
}

// UpdateWorkflow 更新工作流
// @Summary 更新工作流
// @Description 更新现有工作流配置
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param request body workflow.UpdateWorkflowRequest true "更新工作流请求"
// @Success 200 {object} response.Response{data=workflow.WorkflowResponse} "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow/{id} [put]
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的工作流ID", err)
		return
	}

	var req workflow.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数解析错误", err)
		return
	}

	// 设置ID
	req.ID = uint(id)

	// 验证参数
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数验证失败", err)
		return
	}

	// 调用服务层
	result, err := h.workflowService.UpdateWorkflow(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "workflow not found" {
			response.Error(c, http.StatusNotFound, "工作流不存在", nil)
		} else {
			response.Error(c, http.StatusInternalServerError, "更新工作流失败", err)
		}
		return
	}

	response.Success(c, result)
}

// DeleteWorkflow 删除工作流
// @Summary 删除工作流
// @Description 删除指定ID的工作流
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow/{id} [delete]
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的工作流ID", err)
		return
	}

	// 调用服务层
	err = h.workflowService.DeleteWorkflow(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除工作流失败", err)
		return
	}

	response.Success(c, nil)
}

// ListWorkflows 获取工作流列表
// @Summary 获取工作流列表
// @Description 分页获取工作流列表，支持关键词搜索
// @Tags workflow
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页大小，默认为20"
// @Param keyword query string false "搜索关键词"
// @Param type query string false "工作流类型"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} response.Response{data=workflow.WorkflowListResponse} "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow [get]
func (h *Handler) ListWorkflows(c *gin.Context) {
	var req workflow.ListWorkflowsRequest

	// 解析查询参数
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Keyword = c.Query("keyword")
	req.Type = c.Query("type")

	if enabledStr := c.Query("enabled"); enabledStr != "" {
		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "无效的启用状态", err)
			return
		}
		req.Enabled = &enabled
	}

	// 验证参数
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数验证失败", err)
		return
	}

	// 调用服务层
	result, err := h.workflowService.ListWorkflows(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取工作流列表失败", err)
		return
	}

	response.Success(c, result)
}

// ExecuteWorkflow 执行工作流
// @Summary 执行工作流
// @Description 手动执行指定ID的工作流
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param request body map[string]interface{} false "触发数据"
// @Success 200 {object} response.Response{data=workflow.WorkflowExecutionResult} "成功响应"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/workflow/{id}/execute [post]
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的工作流ID", err)
		return
	}

	var triggerData map[string]interface{}
	if err := c.ShouldBindJSON(&triggerData); err != nil && err.Error() != "EOF" {
		response.Error(c, http.StatusBadRequest, "参数解析错误", err)
		return
	}

	// 调用服务层
	result, err := h.workflowService.ExecuteWorkflow(c.Request.Context(), uint(id), triggerData)
	if err != nil {
		if err.Error() == "workflow not found" {
			response.Error(c, http.StatusNotFound, "工作流不存在", nil)
		} else {
			response.Error(c, http.StatusInternalServerError, "执行工作流失败", err)
		}
		return
	}

	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	workflowGroup := router.Group("/workflow")
	{
		workflowGroup.GET("", h.ListWorkflows)
		workflowGroup.POST("", h.CreateWorkflow)
		workflowGroup.GET("/:id", h.GetWorkflow)
		workflowGroup.PUT("/:id", h.UpdateWorkflow)
		workflowGroup.DELETE("/:id", h.DeleteWorkflow)
		workflowGroup.POST("/:id/execute", h.ExecuteWorkflow)
	}
}
