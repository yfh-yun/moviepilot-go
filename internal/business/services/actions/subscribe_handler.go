// Package actions 提供订阅管理的HTTP处理器实现
package actions

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
)

// SubscribeHandler 订阅管理处理器
type SubscribeHandler struct {
	logger           logger.Logger
	subscribeManager SubscribeManager
	subscribeRepo    interfaces.SubscribeRepository
	validator        *SubscribeValidator
}

// NewSubscribeHandler 创建订阅处理器实例
func NewSubscribeHandler(
	subscribeManager SubscribeManager,
	subscribeRepo interfaces.SubscribeRepository,
) *SubscribeHandler {
	return &SubscribeHandler{
		logger:           logger.NewLogger("subscribe_handler"),
		subscribeManager: subscribeManager,
		subscribeRepo:    subscribeRepo,
		validator:        NewSubscribeValidator(),
	}
}

// CreateSubscribe 创建订阅
// @Summary 创建新订阅
// @Description 创建一个新的订阅任务
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param subscribe body AddSubscribeParams true "订阅信息"
// @Success 201 {object} SubscribeResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes [post]
func (h *SubscribeHandler) CreateSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	var params AddSubscribeParams

	// 解析请求体
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("请求参数解析失败", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMS",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 添加订阅
	result, err := h.subscribeManager.AddSubscribe(ctx, &params)
	if err != nil {
		h.logger.Error("创建订阅失败", "error", err.Error(), "name", params.Name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_SUBSCRIBE_FAILED",
			Message: "创建订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetSubscribe 获取订阅详情
// @Summary 获取订阅详情
// @Description 根据ID获取订阅的详细信息
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} SubscribeInfo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id} [get]
func (h *SubscribeHandler) GetSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 获取订阅信息
	subscribe, err := h.subscribeManager.GetSubscribe(ctx, subscribeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("获取订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_SUBSCRIBE_FAILED",
			Message: "获取订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

// UpdateSubscribe 更新订阅
// @Summary 更新订阅信息
// @Description 更新指定订阅的配置信息
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Param subscribe body UpdateSubscribeParams true "更新的订阅信息"
// @Success 200 {object} SubscribeResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id} [put]
func (h *SubscribeHandler) UpdateSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")
	var params UpdateSubscribeParams

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("请求参数解析失败", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMS",
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 更新订阅
	result, err := h.subscribeManager.UpdateSubscribe(ctx, subscribeID, &params)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("更新订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_SUBSCRIBE_FAILED",
			Message: "更新订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteSubscribe 删除订阅
// @Summary 删除订阅
// @Description 根据ID删除指定的订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id} [delete]
func (h *SubscribeHandler) DeleteSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 删除订阅
	err := h.subscribeManager.DeleteSubscribe(ctx, subscribeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("删除订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_SUBSCRIBE_FAILED",
			Message: "删除订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "订阅删除成功",
	})
}

// PauseSubscribe 暂停订阅
// @Summary 暂停订阅
// @Description 暂停指定的订阅任务
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id}/pause [post]
func (h *SubscribeHandler) PauseSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 暂停订阅
	err := h.subscribeManager.PauseSubscribe(ctx, subscribeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("暂停订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "PAUSE_SUBSCRIBE_FAILED",
			Message: "暂停订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "订阅暂停成功",
	})
}

// ResumeSubscribe 恢复订阅
// @Summary 恢复订阅
// @Description 恢复指定的订阅任务
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id}/resume [post]
func (h *SubscribeHandler) ResumeSubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 恢复订阅
	err := h.subscribeManager.ResumeSubscribe(ctx, subscribeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("恢复订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RESUME_SUBSCRIBE_FAILED",
			Message: "恢复订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "订阅恢复成功",
	})
}

// ListSubscribes 获取订阅列表
// @Summary 获取订阅列表
// @Description 分页获取订阅列表，支持过滤条件
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param type query string false "订阅类型过滤"
// @Param status query string false "订阅状态过滤"
// @Param tags query string false "标签过滤，逗号分隔"
// @Param keywords query string false "关键词搜索"
// @Param limit query int false "每页数量，默认20，最大100"
// @Param offset query int false "偏移量，默认0"
// @Success 200 {object} PaginatedResponse{data=[]SubscribeInfo}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes [get]
func (h *SubscribeHandler) ListSubscribes(c *gin.Context) {
	ctx := c.Request.Context()

	// 构建过滤条件
	filter := &SubscribeFilter{}

	// 解析查询参数
	if typeStr := c.Query("type"); typeStr != "" {
		filter.Types = []SubscribeType{SubscribeType(typeStr)}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Statuses = []SubscribeStatus{SubscribeStatus(statusStr)}
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		filter.Tags = strings.Split(tagsStr, ",")
	}

	if keywordsStr := c.Query("keywords"); keywordsStr != "" {
		filter.Keywords = strings.Split(keywordsStr, ",")
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// 获取订阅列表
	subscribes, total, err := h.subscribeManager.ListSubscribes(ctx, filter)
	if err != nil {
		h.logger.Error("获取订阅列表失败", "error", err.Error(), "filter", filter)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LIST_SUBSCRIBES_FAILED",
			Message: "获取订阅列表失败: " + err.Error(),
		})
		return
	}

	// 返回分页结果
	c.JSON(http.StatusOK, PaginatedResponse{
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
		Data:   subscribes,
	})
}

// UpdateSubscribeItems 手动更新订阅项
// @Summary 手动更新订阅项
// @Description 立即更新指定订阅的订阅项
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} UpdateSubscribeItemsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribes/{id}/update [post]
func (h *SubscribeHandler) UpdateSubscribeItems(c *gin.Context) {
	ctx := c.Request.Context()
	subscribeID := c.Param("id")

	if subscribeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_SUBSCRIBE_ID",
			Message: "订阅ID不能为空",
		})
		return
	}

	// 更新订阅项
	newItems, err := h.subscribeManager.UpdateSubscribeItems(ctx, subscribeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SUBSCRIBE_NOT_FOUND",
				Message: "订阅不存在",
			})
			return
		}

		h.logger.Error("更新订阅项失败", "error", err.Error(), "subscribe_id", subscribeID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_SUBSCRIBE_ITEMS_FAILED",
			Message: "更新订阅项失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateSubscribeItemsResponse{
		Message:  "订阅项更新成功",
		NewItems: len(newItems),
		Items:    newItems,
	})
}

// ListSubscribeItems 获取订阅项列表
// @Summary 获取订阅项列表
// @Description 分页获取订阅项列表，支持过滤条件
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param subscribe_id query string false "订阅ID过滤"
// @Param downloaded query bool false "是否已下载"
// @Param categories query string false "分类过滤，逗号分隔"
// @Param keywords query string false "标题关键词搜索"
// @Param limit query int false "每页数量，默认50，最大200"
// @Param offset query int false "偏移量，默认0"
// @Param order_by query string false "排序字段"
// @Param order_dir query string false "排序方向"
// @Success 200 {object} PaginatedResponse{data=[]SubscribeItem}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribe-items [get]
func (h *SubscribeHandler) ListSubscribeItems(c *gin.Context) {
	ctx := c.Request.Context()

	// 构建过滤条件
	filter := &SubscribeItemFilter{}

	// 解析查询参数
	if subscribeID := c.Query("subscribe_id"); subscribeID != "" {
		filter.SubscribeID = subscribeID
	}

	if downloadedStr := c.Query("downloaded"); downloadedStr != "" {
		if downloaded, err := strconv.ParseBool(downloadedStr); err == nil {
			filter.Downloaded = &downloaded
		}
	}

	if categoriesStr := c.Query("categories"); categoriesStr != "" {
		filter.Categories = strings.Split(categoriesStr, ",")
	}

	if keywordsStr := c.Query("keywords"); keywordsStr != "" {
		filter.Keywords = strings.Split(keywordsStr, ",")
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if orderBy := c.Query("order_by"); orderBy != "" {
		filter.OrderBy = orderBy
	}

	if orderDir := c.Query("order_dir"); orderDir != "" {
		filter.OrderDir = orderDir
	}

	// 获取订阅项列表
	items, total, err := h.subscribeManager.GetSubscribeItems(ctx, filter)
	if err != nil {
		h.logger.Error("获取订阅项列表失败", "error", err.Error(), "filter", filter)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LIST_SUBSCRIBE_ITEMS_FAILED",
			Message: "获取订阅项列表失败: " + err.Error(),
		})
		return
	}

	// 返回分页结果
	c.JSON(http.StatusOK, PaginatedResponse{
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
		Data:   items,
	})
}

// SubscribeValidator 订阅验证器
type SubscribeValidator struct {
	logger logger.Logger
}

// NewSubscribeValidator 创建订阅验证器实例
func NewSubscribeValidator() *SubscribeValidator {
	return &SubscribeValidator{
		logger: logger.NewLogger("subscribe_validator"),
	}
}

// 响应结构定义

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误信息
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"` // 成功消息
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Total  int64       `json:"total"`  // 总数量
	Limit  int         `json:"limit"`  // 每页数量
	Offset int         `json:"offset"` // 偏移量
	Data   interface{} `json:"data"`   // 数据
}

// UpdateSubscribeItemsResponse 更新订阅项响应
type UpdateSubscribeItemsResponse struct {
	Message  string           `json:"message"`   // 操作消息
	NewItems int              `json:"new_items"` // 新增项目数
	Items    []*SubscribeItem `json:"items"`     // 订阅项列表
}
