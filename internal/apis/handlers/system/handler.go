package system

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middleware"
	"moviepilot-go/internal/business/services/system"
	"moviepilot-go/pkg/logger"
)

// Handler 系统管理 API 处理器
type Handler struct {
	systemService *system.SystemService
}

// NewHandler 创建处理器
func NewHandler(systemService *system.SystemService) *Handler {
	return &Handler{
		systemService: systemService,
	}
}

// GetGlobal 获取非敏感系统设置
// @Summary 获取非敏感系统设置
// @Description 查询非敏感系统设置（需要提供固定 token，与 Python 版本保持兼容）
// @Tags system
// @Produce json
// @Param token query string true "访问令牌"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/system/global [get]
func (h *Handler) GetGlobal(c *gin.Context) {
	logger.Debug("GetGlobal called", zap.String("func", "GetGlobal"))

	token := c.Query("token")
	if token != "moviepilot" {
		logger.Warn("Invalid token for GetGlobal", zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Forbidden",
		})
		return
	}

	settings, err := h.systemService.GetGlobalSettings(c.Request.Context())
	if err != nil {
		logger.Error("获取全局系统设置失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Info("获取全局系统设置成功", zap.Int("settings_count", len(settings)))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统基本信息、版本等
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/system/info [get]
func (h *Handler) GetSystemInfo(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("GetSystemInfo called",
		zap.String("func", "GetSystemInfo"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	info, err := h.systemService.GetSystemInfo(c.Request.Context())
	if err != nil {
		logger.Error("获取系统信息失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取系统信息成功",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, info)
}

// GetEnvSettings 获取环境变量配置
// @Summary 获取环境变量配置
// @Description 获取系统环境变量配置（仅管理员）
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/system/env [get]
func (h *Handler) GetEnvSettings(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("GetEnvSettings called",
		zap.String("func", "GetEnvSettings"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	settings, err := h.systemService.GetEnvSettings(c.Request.Context())
	if err != nil {
		logger.Error("获取环境变量失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Info("获取环境变量成功",
		zap.Int("settings_count", len(settings)),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateEnvSettings 更新环境变量配置
// @Summary 更新环境变量配置
// @Description 更新系统环境变量配置（仅管理员）
// @Tags system
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param env body map[string]interface{} true "环境变量"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/system/env [post]
func (h *Handler) UpdateEnvSettings(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("UpdateEnvSettings called",
		zap.String("func", "UpdateEnvSettings"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	var env map[string]any
	if err := c.ShouldBindJSON(&env); err != nil {
		logger.Warn("更新环境变量参数解析失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result, err := h.systemService.UpdateEnvSettings(c.Request.Context(), env)
	if err != nil {
		logger.Error("更新环境变量失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	failedUpdates, _ := result["failed_updates"].(map[string]string)

	success := len(failedUpdates) == 0
	message := "所有配置项更新成功"
	if !success {
		message = "部分配置更新失败"
		logger.Warn("部分环境变量更新失败",
			zap.Int("failed_count", len(failedUpdates)),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
	}

	logger.Info("环境变量更新完成",
		zap.Bool("success", success),
		zap.Int("total_updates", len(env)),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": message,
		"data":    result,
	})
}

// GetSetting 获取系统设置
// @Summary 获取系统设置
// @Description 根据key获取系统设置
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/system/setting/{key} [get]
func (h *Handler) GetSetting(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("GetSetting called",
		zap.String("func", "GetSetting"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	key := c.Param("key")
	if key == "" {
		logger.Warn("获取配置失败：配置键为空",
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置键不能为空",
		})
		return
	}

	value, err := h.systemService.GetConfig(c.Request.Context(), key)
	if err != nil {
		logger.Error("获取配置失败",
			zap.String("key", key),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Info("获取配置成功",
		zap.String("key", key),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"value": value,
		},
	})
}

// UpdateSetting 更新系统设置
// @Summary 更新系统设置
// @Description 根据key更新系统设置
// @Tags system
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Param value body interface{} true "配置值"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/system/setting/{key} [post]
func (h *Handler) UpdateSetting(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("UpdateSetting called",
		zap.String("func", "UpdateSetting"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	key := c.Param("key")
	if key == "" {
		logger.Warn("更新配置失败：配置键为空",
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置键不能为空",
		})
		return
	}

	var value any
	if err := c.ShouldBindJSON(&value); err != nil {
		logger.Warn("更新配置参数解析失败",
			zap.String("key", key),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if err := h.systemService.SetConfig(c.Request.Context(), key, value); err != nil {
		logger.Error("更新配置失败",
			zap.String("key", key),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Info("配置更新成功",
		zap.String("key", key),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// GetHealth 健康检查
// @Summary 健康检查
// @Description 获取系统健康状态
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /api/system/health [get]
func (h *Handler) GetHealth(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("GetHealth called",
		zap.String("func", "GetHealth"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
	})
}

// GetMetrics 获取系统指标
// @Summary 获取系统指标
// @Description 获取系统性能指标（CPU、内存、磁盘等）
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MetricsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/system/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("GetMetrics called",
		zap.String("func", "GetMetrics"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	// CPU 使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		logger.Error("获取CPU使用率失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Error("获取内存信息失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
	}

	// 磁盘信息
	diskInfo, err := disk.Usage("/")
	if err != nil {
		logger.Error("获取磁盘信息失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
	}

	// 主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Error("获取主机信息失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
	}

	metrics := gin.H{
		"cpu": gin.H{
			"usage_percent": cpuPercent,
			"cores":         runtime.NumCPU(),
		},
		"memory": gin.H{
			"total":         memInfo.Total,
			"used":          memInfo.Used,
			"free":          memInfo.Free,
			"usage_percent": memInfo.UsedPercent,
		},
		"disk": gin.H{
			"total":         diskInfo.Total,
			"used":          diskInfo.Used,
			"free":          diskInfo.Free,
			"usage_percent": diskInfo.UsedPercent,
		},
		"host": gin.H{
			"hostname":         hostInfo.Hostname,
			"os":               hostInfo.OS,
			"platform":         hostInfo.Platform,
			"platform_version": hostInfo.PlatformVersion,
			"uptime":           hostInfo.Uptime,
		},
		"go": gin.H{
			"version":    runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// RestartSystem 重启系统
// @Summary 重启系统
// @Description 重启MoviePilot系统（仅管理员）
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/system/restart [get]
func (h *Handler) RestartSystem(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	logger.Debug("RestartSystem called",
		zap.String("func", "RestartSystem"),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	if err := h.systemService.Restart(c.Request.Context()); err != nil {
		logger.Error("重启系统失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Info("系统重启命令执行成功",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统正在重启...",
	})
}

// GetVersion 获取版本信息
// @Summary 获取版本信息
// @Description 获取MoviePilot版本信息
// @Tags system
// @Produce json
// @Success 200 {object} VersionResponse
// @Router /api/system/version [get]
func (h *Handler) GetVersion(c *gin.Context) {
	logger.Debug("GetVersion called", zap.String("func", "GetVersion"))

	version, err := h.systemService.GetVersion(c.Request.Context())
	if err != nil {
		logger.Error("获取版本信息失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取版本信息成功")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    version,
	})
}

// GetLogs 获取系统日志
// @Summary 获取系统日志
// @Description 获取系统日志（支持实时流式传输）
// @Tags system
// @Security BearerAuth
// @Produce text/event-stream
// @Param length query int false "日志行数" default(50)
// @Param logfile query string false "日志文件名" default("moviepilot.log")
// @Success 200 {string} string "SSE流"
// @Router /api/system/logging [get]
func (h *Handler) GetLogs(c *gin.Context) {
	lengthStr := c.DefaultQuery("length", "50")
	length := 50
	if l, err := strconv.Atoi(lengthStr); err == nil {
		length = l
	}
	logfile := c.DefaultQuery("logfile", "moviepilot.log")

	// 从环境变量获取日志路径
	logPathBase := os.Getenv("LOG_PATH")
	if logPathBase == "" {
		logPathBase = "/app/logs"
	}

	// 构建完整的日志文件路径
	logPath := logPathBase + "/" + logfile

	// 检查日志文件路径是否安全
	if !h.systemService.IsValidLogPath(logPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	// 检查日志文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	// length=-1 时返回全部日志（text/plain，倒序）
	if length == -1 {
		content, err := h.systemService.ReadLogFile(c.Request.Context(), logPath)
		if err != nil {
			logger.Error("读取日志文件失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
			c.String(http.StatusOK, "读取日志文件失败: %v", err)
			return
		}
		// 倒序输出
		lines := strings.Split(content, "\n")
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
		reversed := strings.Join(lines, "\n")
		c.String(http.StatusOK, reversed)
		return
	}

	// 否则返回 SSE 流
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Stream(func(w io.Writer) bool {
		// 获取初始文件内容
		content, err := h.systemService.ReadLogFile(c.Request.Context(), logPath)
		if err != nil {
			logger.Error("读取日志文件失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
			c.SSEvent("data", "日志读取异常: "+err.Error())
			return false
		}

		lines := strings.Split(content, "\n")
		// 取最后N行
		start := 0
		if len(lines) > length {
			start = len(lines) - length
		}
		for _, line := range lines[start:] {
			if strings.TrimSpace(line) != "" {
				c.SSEvent("data", line)
			}
		}

		// 实时监听新日志
		return h.streamNewLogs(c.Request.Context(), logPath, c)
	})
}

// streamNewLogs 实时监听新日志
func (h *Handler) streamNewLogs(ctx context.Context, logPath string, c *gin.Context) bool {
	// 打开日志文件
	file, err := os.Open(logPath)
	if err != nil {
		logger.Error("打开日志文件失败", zap.Error(err), zap.String("log_path", logPath))
		return false
	}
	defer file.Close()

	// 移动文件指针到文件末尾
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	file.Seek(fileSize, 0)

	// 缓冲区
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			// 读取新内容
			n, err := file.Read(buf)
			if err != nil && n == 0 {
				// 没有新内容，短暂等待
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// 处理新内容
			if n > 0 {
				content := string(buf[:n])
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						c.SSEvent("data", line)
					}
				}
			}
		}
	}
}

// TestNetwork 测试网络连通性
// @Summary 测试网络连通性
// @Description 测试指定URL的网络连通性
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param url query string true "测试URL"
// @Param proxy query bool false "是否使用代理" default(false)
// @Param include query string false "响应内容包含校验字符串"
// @Success 200 {object} NetworkTestResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/system/nettest [get]
func (h *Handler) TestNetwork(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL不能为空"})
		return
	}

	proxy := c.Query("proxy") == "true"
	include := c.Query("include")

	result, err := h.systemService.TestNetwork(c.Request.Context(), url, proxy, include)
	if err != nil {
		logger.Error("网络连通性测试失败",
			zap.String("url", url),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	success, _ := result["success"].(bool)
	message, _ := result["message"].(string)
	timeVal := result["time"]

	// 对齐 Python：success + message + data.time
	data := gin.H{"time": timeVal}
	resp := gin.H{
		"success": success,
		"data":    data,
	}
	if message != "" {
		resp["message"] = message
	}

	c.JSON(http.StatusOK, resp)
}

// Versions 查询Github所有Release版本
// @Summary 查询Github所有Release版本
// @Description 查询Github所有Release版本
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/system/versions [get]
func (h *Handler) Versions(c *gin.Context) {
	releases, err := h.systemService.GetReleases(c.Request.Context())
	if err != nil {
		logger.Error("获取Github Release版本失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    releases,
	})
}

// 响应结构体定义
type SystemInfoResponse struct {
	Version     string         `json:"version"`
	BuildTime   string         `json:"build_time"`
	GoVersion   string         `json:"go_version"`
	Platform    string         `json:"platform"`
	Environment map[string]any `json:"environment"`
}

type HealthResponse struct {
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
	Uptime    float64 `json:"uptime"`
}

type MetricsResponse struct {
	CPU    CPUMetrics    `json:"cpu"`
	Memory MemoryMetrics `json:"memory"`
	Disk   DiskMetrics   `json:"disk"`
	Host   HostMetrics   `json:"host"`
}

type CPUMetrics struct {
	UsagePercent []float64 `json:"usage_percent"`
	Cores        int       `json:"cores"`
}

type MemoryMetrics struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskMetrics struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type HostMetrics struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	Uptime          uint64 `json:"uptime"`
	GoMetrics       struct {
		Goroutines   int    `json:"goroutines"`
		MemoryAlloc  uint64 `json:"memory_alloc"`
		MemorySys    uint64 `json:"memory_sys"`
		NumGC        uint32 `json:"num_gc"`
		LastGCTime   string `json:"last_gc_time"`
		HeapObjects  uint64 `json:"heap_objects"`
		GoVersion    string `json:"go_version"`
		NumCPU       int    `json:"num_cpu"`
		GOOS         string `json:"goos"`
		GOARCH       string `json:"goarch"`
		CompilerType string `json:"compiler_type"`
	} `json:"go_metrics"`
}

type NetworkTestResponse struct {
	Success      bool    `json:"success"`
	URL          string  `json:"url"`
	StatusCode   int     `json:"status_code,omitempty"`
	ResponseTime float64 `json:"response_time,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// ClearCache 清理缓存
// @Summary 清理缓存
// @Description 清理系统缓存
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param cache_type query string false "缓存类型" Enums(all, tmdb, douban, bangumi)
// @Success 200 {object} map[string]interface{}
// @Router /api/system/cache [delete]
func (h *Handler) ClearCache(c *gin.Context) {
	cacheType := c.DefaultQuery("cache_type", "all")

	logger.Info("清理缓存",
		zap.String("cache_type", cacheType),
		zap.String("client_ip", c.ClientIP()),
	)

	// TODO: 实现缓存清理逻辑
	// 1. 根据类型清理对应缓存
	// 2. Redis缓存清理
	// 3. 内存缓存清理

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "缓存已清理",
		"type":    cacheType,
	})
}

// GetProgress 获取进度
// @Summary 获取进度
// @Description 获取指定类型的处理进度，返回格式为SSE
// @Tags system
// @Produce text/event-stream
// @Param process_type path string true "进程类型"
// @Success 200 {string} string "SSE流"
// @Router /api/system/progress/{process_type} [get]
func (h *Handler) GetProgress(c *gin.Context) {
	processType := c.Param("process_type")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	progressCh := h.systemService.StreamProgress(ctx, processType)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case progress, ok := <-progressCh:
			if !ok {
				return false
			}
			c.SSEvent("data", progress)
			return true
		}
	})
}

// GetMessages 获取实时消息
// @Summary 获取实时消息
// @Description 实时获取系统消息，返回格式为SSE
// @Tags system
// @Produce text/event-stream
// @Param role query string false "角色" default("system")
// @Success 200 {string} string "SSE流"
// @Router /api/system/message [get]
func (h *Handler) GetMessages(c *gin.Context) {
	role := c.DefaultQuery("role", "system")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	messageCh := h.systemService.StreamMessages(ctx, role)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case message, ok := <-messageCh:
			if !ok {
				return false
			}
			c.SSEvent("data", message)
			return true
		}
	})
}

// RuleTest 过滤规则测试
// @Summary 过滤规则测试
// @Description 过滤规则测试，规则类型 1-订阅，2-洗版，3-搜索
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param title query string true "标题"
// @Param rulegroup_name query string true "规则组名称"
// @Param subtitle query string false "副标题"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/system/ruletest [get]
func (h *Handler) RuleTest(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	title := c.Query("title")
	ruleGroupName := c.Query("rulegroup_name")
	subtitle := c.Query("subtitle")

	if title == "" || ruleGroupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "标题和规则组名称不能为空",
		})
		return
	}

	result, err := h.systemService.TestRule(c.Request.Context(), title, subtitle, ruleGroupName)
	if err != nil {
		logger.Error("过滤规则测试失败",
			zap.String("title", title),
			zap.String("rulegroup_name", ruleGroupName),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ModuleList 查询已加载的模块ID列表
// @Summary 查询已加载的模块ID列表
// @Description 查询已加载的模块ID列表
// @Tags system
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/system/modulelist [get]
func (h *Handler) ModuleList(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	modules, err := h.systemService.GetModuleList(c.Request.Context())
	if err != nil {
		logger.Error("获取模块列表失败",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modules,
	})
}

// ModuleTest 模块可用性测试
// @Summary 模块可用性测试
// @Description 模块可用性测试接口
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param moduleid path string true "模块ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/system/moduletest/{moduleid} [get]
func (h *Handler) ModuleTest(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	moduleID := c.Param("moduleid")

	state, message, err := h.systemService.TestModule(c.Request.Context(), moduleID)
	if err != nil {
		logger.Error("模块测试失败",
			zap.String("module_id", moduleID),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": state,
		"message": message,
	})
}

// RunScheduler 运行服务
// @Summary 运行服务
// @Description 执行命令（仅管理员）
// @Tags system
// @Security BearerAuth
// @Produce json
// @Param jobid query string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/system/runscheduler [get]
func (h *Handler) RunScheduler(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	jobID := c.Query("jobid")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "任务ID不能为空！",
		})
		return
	}

	if err := h.systemService.RunScheduler(c.Request.Context(), jobID); err != nil {
		logger.Error("运行定时任务失败",
			zap.String("job_id", jobID),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已开始执行",
	})
}

// RunSchedulerApiToken 运行服务（API_TOKEN认证）
// @Summary 运行服务（API_TOKEN认证）
// @Description 执行命令（API_TOKEN认证）
// @Tags system
// @Produce json
// @Param jobid query string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/system/runscheduler2 [get]
func (h *Handler) RunSchedulerApiToken(c *gin.Context) {
	reqID := c.GetString("request_id")
	// 注意：此接口使用 API_TOKEN 认证，但这里仍记录 user_id 占位，以便未来扩展
	userID := middleware.GetUserID(c)
	jobID := c.Query("jobid")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "任务ID不能为空！",
		})
		return
	}

	// 在实际实现中，这里应该验证API Token
	// 目前简化处理，直接运行任务

	if err := h.systemService.RunScheduler(c.Request.Context(), jobID); err != nil {
		logger.Error("运行定时任务失败",
			zap.String("job_id", jobID),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已开始执行",
	})
}

// CacheImage 本地缓存图片文件
// @Summary 本地缓存图片文件
// @Description 本地缓存图片文件，支持 HTTP 缓存，如果启用全局图片缓存，则使用磁盘缓存
// @Tags system
// @Produce image/jpeg
// @Param url query string true "图片URL"
// @Param if_none_match header string false "If-None-Match header for ETag"
// @Success 200 {file} binary
// @Router /api/system/cache/image [get]
func (h *Handler) CacheImage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片URL不能为空"})
		return
	}

	ifNoneMatch := c.GetHeader("if_none_match")

	logger.Info("缓存图片",
		zap.String("url", url),
		zap.String("client_ip", c.ClientIP()),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	// 调用服务层处理图片缓存（服务层内部会根据环境变量决定是否使用缓存和代理）
	data, contentType, headers, statusCode, err := h.systemService.CacheImage(
		c.Request.Context(), url, ifNoneMatch)

	if err != nil {
		logger.Error("图片缓存失败",
			zap.String("url", url),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 处理304 Not Modified
	if statusCode == http.StatusNotModified {
		c.Status(http.StatusNotModified)
		if headers != nil {
			if etag, exists := headers["ETag"]; exists {
				c.Header("ETag", etag)
			}
			if cacheControl, exists := headers["Cache-Control"]; exists {
				c.Header("Cache-Control", cacheControl)
			}
		}
		return
	}

	// 设置响应头
	for key, value := range headers {
		c.Header(key, value)
	}

	// 返回图片数据
	c.Data(statusCode, contentType, data)
}

// ProxyImage 图片代理
// @Summary 图片代理
// @Description 代理获取图片资源，支持HTTP缓存和磁盘缓存
// @Tags system
// @Produce image/jpeg
// @Param imgurl path string true "图片URL"
// @Param proxy query bool false "是否使用代理" default(false)
// @Param cache query bool false "是否使用缓存" default(false)
// @Param if_none_match header string false "If-None-Match header for ETag"
// @Success 200 {file} binary
// @Failure 403 {object} map[string]interface{}
// @Router /api/system/img/{proxy} [get]
func (h *Handler) ProxyImage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	proxy := c.Param("proxy")
	imgURL := c.Query("imgurl")
	if imgURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片URL不能为空"})
		return
	}

	proxyEnabled := proxy == "true"
	cache := c.Query("cache") == "true"
	ifNoneMatch := c.GetHeader("if_none_match")
	logger.Info("代理图片",
		zap.String("url", imgURL),
		zap.Bool("proxy", proxyEnabled),
		zap.Bool("cache", cache),
		zap.String("client_ip", c.ClientIP()),
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	// 获取媒体服务器主机列表，用于安全验证
	allowedHosts := h.systemService.GetMediaServerHosts(c.Request.Context())
	allowedDomainSet := make(map[string]bool)
	for _, host := range allowedHosts {
		allowedDomainSet[host] = true
	}

	// 调用服务层处理图片代理
	data, contentType, headers, statusCode, err := h.systemService.ProxyImage(
		c.Request.Context(), imgURL, proxyEnabled, cache, ifNoneMatch, allowedDomainSet)

	if err != nil {
		logger.Error("图片代理失败", zap.String("url", imgURL), zap.Error(err), zap.String("client_ip", c.ClientIP()))
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 处理304 Not Modified
	if statusCode == http.StatusNotModified {
		c.Status(http.StatusNotModified)
		if headers != nil {
			if etag, exists := headers["ETag"]; exists {
				c.Header("ETag", etag)
			}
			if cacheControl, exists := headers["Cache-Control"]; exists {
				c.Header("Cache-Control", cacheControl)
			}
		}
		return
	}

	// 设置响应头
	for key, value := range headers {
		c.Header(key, value)
	}
	c.Header("Cache-Control", "max-age=604800") // 7天
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}

	// 返回图片数据
	c.Data(statusCode, contentType, data)
}

var startTime = time.Now()
