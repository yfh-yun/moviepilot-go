package download

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	downloadbiz "moviepilot-go/internal/business/services/download"
	"moviepilot-go/pkg/logger"
)

// EnhancedHandler 下载增强功能 API 处理器
type EnhancedHandler struct {
	queueService     downloadbiz.QueueService
	limiterService   downloadbiz.LimiterService
	analyticsService downloadbiz.AnalyticsService
	logger           *zap.Logger
}

// NewEnhancedHandler 创建增强处理器
func NewEnhancedHandler(
	queueService downloadbiz.QueueService,
	limiterService downloadbiz.LimiterService,
	analyticsService downloadbiz.AnalyticsService,
) *EnhancedHandler {
	return &EnhancedHandler{
		queueService:     queueService,
		limiterService:   limiterService,
		analyticsService: analyticsService,
		logger:           logger.GetLogger(),
	}
}

// GetQueue 获取下载队列
// @Summary 获取下载队列
// @Description 获取下载队列中的所有任务
// @Tags download-enhanced
// @Produce json
// @Param status query string false "任务状态"
// @Success 200 {array} downloadbiz.DownloadTask
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/queue [get]
func (h *EnhancedHandler) GetQueue(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	status := c.Query("status")

	h.logger.Debug("GetQueue called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("status", status),
	)

	tasks, err := h.queueService.List(c.Request.Context(), status)
	if err != nil {
		h.logger.Error("获取队列失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetQueueStats 获取队列统计
// @Summary 获取队列统计
// @Description 获取下载队列的统计信息
// @Tags download-enhanced
// @Produce json
// @Success 200 {object} downloadbiz.QueueStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/queue/stats [get]
func (h *EnhancedHandler) GetQueueStats(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	stats, err := h.queueService.GetQueueStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取队列统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AddToQueue 添加到队列
// @Summary 添加到队列
// @Description 添加下载任务到队列
// @Tags download-enhanced
// @Accept json
// @Produce json
// @Param task body downloadbiz.DownloadTask true "下载任务"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/queue [post]
func (h *EnhancedHandler) AddToQueue(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var task downloadbiz.DownloadTask

	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Warn("AddToQueue invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.queueService.Add(c.Request.Context(), &task); err != nil {
		h.logger.Error("添加到队列失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("添加到队列成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Any("task_id", task.ID),
	)
	c.JSON(http.StatusOK, gin.H{"message": "添加成功", "task_id": task.ID})
}

// RemoveFromQueue 从队列移除
// @Summary 从队列移除
// @Description 从队列中移除指定任务
// @Tags download-enhanced
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/queue/{id} [delete]
func (h *EnhancedHandler) RemoveFromQueue(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	taskID := c.Param("id")

	if err := h.queueService.Remove(c.Request.Context(), taskID); err != nil {
		h.logger.Error("从队列移除失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("task_id", taskID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("从队列移除成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("task_id", taskID),
	)
	c.JSON(http.StatusOK, gin.H{"message": "移除成功"})
}

// SetGlobalLimit 设置全局限速
// @Summary 设置全局限速
// @Description 设置全局下载和上传速度限制
// @Tags download-enhanced
// @Accept json
// @Produce json
// @Param limit body SpeedLimitRequest true "速度限制"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/limit/global [post]
func (h *EnhancedHandler) SetGlobalLimit(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req SpeedLimitRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("SetGlobalLimit invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.limiterService.SetGlobalLimit(c.Request.Context(), req.DownloadLimit, req.UploadLimit); err != nil {
		h.logger.Error("设置全局限速失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("设置全局限速成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int64("download_limit", req.DownloadLimit),
		zap.Int64("upload_limit", req.UploadLimit),
	)
	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

// GetGlobalLimit 获取全局限速
// @Summary 获取全局限速
// @Description 获取全局速度限制设置
// @Tags download-enhanced
// @Produce json
// @Success 200 {object} downloadbiz.SpeedLimit
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/limit/global [get]
func (h *EnhancedHandler) GetGlobalLimit(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	limit, err := h.limiterService.GetGlobalLimit(c.Request.Context())
	if err != nil {
		h.logger.Error("获取全局限速失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, limit)
}

// SetSchedule 设置定时限速
// @Summary 设置定时限速
// @Description 设置定时速度限制计划
// @Tags download-enhanced
// @Accept json
// @Produce json
// @Param schedule body downloadbiz.LimitSchedule true "限速计划"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/limit/schedule [post]
func (h *EnhancedHandler) SetSchedule(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var schedule downloadbiz.LimitSchedule

	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Warn("SetSchedule invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.limiterService.SetSchedule(c.Request.Context(), &schedule); err != nil {
		h.logger.Error("设置定时限速失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

// GetSchedules 获取定时限速列表
// @Summary 获取定时限速列表
// @Description 获取所有定时限速计划
// @Tags download-enhanced
// @Produce json
// @Success 200 {array} downloadbiz.LimitSchedule
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/limit/schedule [get]
func (h *EnhancedHandler) GetSchedules(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	schedules, err := h.limiterService.GetSchedules(c.Request.Context())
	if err != nil {
		h.logger.Error("获取定时限速列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetOverallStats 获取总体统计
// @Summary 获取总体统计
// @Description 获取下载的总体统计信息
// @Tags download-enhanced
// @Produce json
// @Success 200 {object} downloadbiz.OverallStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/stats/overall [get]
func (h *EnhancedHandler) GetOverallStats(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	stats, err := h.analyticsService.GetOverallStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取总体统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDailyStats 获取每日统计
// @Summary 获取每日统计
// @Description 获取指定天数的每日统计数据
// @Tags download-enhanced
// @Produce json
// @Param days query int false "天数" default(30)
// @Success 200 {array} downloadbiz.DailyStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/stats/daily [get]
func (h *EnhancedHandler) GetDailyStats(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	stats, err := h.analyticsService.GetDailyStats(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("获取每日统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("days", days),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCategoryStats 获取分类统计
// @Summary 获取分类统计
// @Description 获取按分类的统计信息
// @Tags download-enhanced
// @Produce json
// @Success 200 {array} downloadbiz.CategoryStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/stats/category [get]
func (h *EnhancedHandler) GetCategoryStats(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	stats, err := h.analyticsService.GetCategoryStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取分类统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSpeedStats 获取速度统计
// @Summary 获取速度统计
// @Description 获取下载和上传速度统计
// @Tags download-enhanced
// @Produce json
// @Param hours query int false "小时数" default(24)
// @Success 200 {object} downloadbiz.SpeedStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/stats/speed [get]
func (h *EnhancedHandler) GetSpeedStats(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	hoursStr := c.DefaultQuery("hours", "24")
	hours, _ := strconv.Atoi(hoursStr)
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168 // 最多7天
	}

	stats, err := h.analyticsService.GetSpeedStats(c.Request.Context(), hours)
	if err != nil {
		h.logger.Error("获取速度统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("hours", hours),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SpeedLimitRequest 速度限制请求
type SpeedLimitRequest struct {
	DownloadLimit int64 `json:"download_limit" binding:"required"`
	UploadLimit   int64 `json:"upload_limit" binding:"required"`
}
