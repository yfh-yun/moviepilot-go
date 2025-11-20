// Package file 文件管理API处理器
package file

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/file"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// Handler 文件管理API处理器
type Handler struct {
	fileService file.FileService
	logger      *zap.Logger
}

// NewHandler 创建新的文件管理API处理器
func NewHandler(
	fileService file.FileService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		fileService: fileService,
		logger:      logger,
	}
}

// RegisterRoutes 注册文件管理相关路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	fileGroup := router.Group("/file")
	{
		// 文件操作
		fileGroup.GET("/exists", h.FileExists)
		fileGroup.GET("/list", h.ListDirectory)
		fileGroup.GET("/info", h.GetFileInfo)
		fileGroup.GET("/hash", h.GetFileHash)
		fileGroup.GET("/read", h.ReadFile)
		fileGroup.POST("/write", h.WriteFile)
		fileGroup.POST("/create-dir", h.CreateDirectory)
		fileGroup.PUT("/move", h.MoveFile)
		fileGroup.POST("/copy", h.CopyFile)
		fileGroup.DELETE("/delete", h.DeleteFile)
		fileGroup.PUT("/permissions", h.SetFilePermissions)

		// 存储管理
		fileGroup.GET("/storage/info", h.GetStorageInfo)
		fileGroup.GET("/storage/health", h.GetStorageHealth)
		fileGroup.POST("/storage/cleanup", h.CleanupStorage)

		// 备份操作
		fileGroup.POST("/backup", h.CreateBackup)
		fileGroup.GET("/backup/list", h.ListBackups)
		fileGroup.POST("/backup/:id/restore", h.RestoreBackup)
		fileGroup.DELETE("/backup/:id", h.DeleteBackup)

		// 文件搜索
		fileGroup.POST("/search", h.SearchFiles)

		// 文件处理
		fileGroup.POST("/process", h.ProcessFile)
	}
}

// FileExists 检查文件是否存在
// @Summary 检查文件是否存在
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response{data=bool}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/exists [get]
func (h *Handler) FileExists(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	exists, err := h.fileService.FileExists(c.Request.Context(), path)
	if err != nil {
		h.logger.Error("Failed to check file existence", zap.String("path", path), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to check file existence")
		return
	}

	response.Success(c, exists)
}

// ListDirectory 列出目录内容
// @Summary 列出目录内容
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "目录路径"
// @Param recursive query bool false "是否递归列出" default(false)
// @Success 200 {object} response.Response{data=models.DirectoryListing}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/list [get]
func (h *Handler) ListDirectory(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	recursive := false
	if recursiveStr := c.Query("recursive"); recursiveStr != "" {
		if parsed, err := strconv.ParseBool(recursiveStr); err == nil {
			recursive = parsed
		}
	}

	listing, err := h.fileService.ListDirectory(c.Request.Context(), path, recursive)
	if err != nil {
		h.logger.Error("Failed to list directory", zap.String("path", path), zap.Error(err))
		if strings.Contains(err.Error(), "not a directory") {
			response.Error(c, http.StatusBadRequest, "path is not a directory")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to list directory")
		}
		return
	}

	response.Success(c, listing)
}

// ReadFile 读取文件内容
// @Summary 读取文件内容
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response{data=[]byte}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/read [get]
func (h *Handler) ReadFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	data, err := h.fileService.ReadFile(c.Request.Context(), path)
	if err != nil {
		h.logger.Error("Failed to read file", zap.String("path", path), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "file does not exist")
		} else if strings.Contains(err.Error(), "cannot read directory") {
			response.Error(c, http.StatusBadRequest, "cannot read directory as file")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to read file")
		}
		return
	}

	// 设置适当的Content-Type
	contentType := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".log":
		contentType = "text/plain"
	case ".json":
		contentType = "application/json"
	case ".xml":
		contentType = "application/xml"
	case ".html":
		contentType = "text/html"
	}

	c.Data(http.StatusOK, contentType, data)
}

// WriteFile 写入文件内容
// @Summary 写入文件内容
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.WriteFileRequest true "写入文件请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/write [post]
func (h *Handler) WriteFile(c *gin.Context) {
	var req models.WriteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	if err := h.fileService.WriteFile(c.Request.Context(), req.Path, req.Data); err != nil {
		h.logger.Error("Failed to write file", zap.String("path", req.Path), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to write file")
		return
	}

	response.Success(c, nil)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/delete [delete]
func (h *Handler) DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), path); err != nil {
		h.logger.Error("Failed to delete file", zap.String("path", path), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "file does not exist")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to delete file")
		}
		return
	}

	response.Success(c, nil)
}

// MoveFile 移动或重命名文件
// @Summary 移动或重命名文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.MoveFileRequest true "移动文件请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/move [put]
func (h *Handler) MoveFile(c *gin.Context) {
	var req models.MoveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	if err := h.fileService.MoveFile(c.Request.Context(), req.OldPath, req.NewPath); err != nil {
		h.logger.Error("Failed to move file", zap.String("oldPath", req.OldPath), zap.String("newPath", req.NewPath), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "source file does not exist")
		} else if strings.Contains(err.Error(), "already exists") {
			response.Error(c, http.StatusBadRequest, "destination path already exists")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to move file")
		}
		return
	}

	response.Success(c, nil)
}

// CopyFile 复制文件
// @Summary 复制文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.CopyFileRequest true "复制文件请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/copy [post]
func (h *Handler) CopyFile(c *gin.Context) {
	var req models.CopyFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	if err := h.fileService.CopyFile(c.Request.Context(), req.SrcPath, req.DstPath); err != nil {
		h.logger.Error("Failed to copy file", zap.String("srcPath", req.SrcPath), zap.String("dstPath", req.DstPath), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "source file does not exist")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to copy file")
		}
		return
	}

	response.Success(c, nil)
}

// CreateDirectory 创建目录
// @Summary 创建目录
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.CreateDirectoryRequest true "创建目录请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/create-dir [post]
func (h *Handler) CreateDirectory(c *gin.Context) {
	var req models.CreateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	if err := h.fileService.CreateDirectory(c.Request.Context(), req.Path); err != nil {
		h.logger.Error("Failed to create directory", zap.String("path", req.Path), zap.Error(err))
		if strings.Contains(err.Error(), "already exists") {
			response.Error(c, http.StatusBadRequest, "directory already exists")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to create directory")
		}
		return
	}

	response.Success(c, nil)
}

// GetFileInfo 获取文件信息
// @Summary 获取文件信息
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response{data=models.FileInfo}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/info [get]
func (h *Handler) GetFileInfo(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	info, err := h.fileService.GetFileInfo(c.Request.Context(), path)
	if err != nil {
		h.logger.Error("Failed to get file info", zap.String("path", path), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "file does not exist")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to get file info")
		}
		return
	}

	response.Success(c, info)
}

// GetFileHash 获取文件哈希值
// @Summary 获取文件哈希值
// @Tags file
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/hash [get]
func (h *Handler) GetFileHash(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}

	hash, err := h.fileService.GetFileHash(c.Request.Context(), path)
	if err != nil {
		h.logger.Error("Failed to get file hash", zap.String("path", path), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "file does not exist")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to get file hash")
		}
		return
	}

	response.Success(c, hash)
}

// SetFilePermissions 设置文件权限
// @Summary 设置文件权限
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.SetPermissionsRequest true "设置权限请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/permissions [put]
func (h *Handler) SetFilePermissions(c *gin.Context) {
	var req models.SetPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	if err := h.fileService.SetFilePermissions(c.Request.Context(), req.Path, req.Permissions); err != nil {
		h.logger.Error("Failed to set file permissions", zap.String("path", req.Path), zap.String("permissions", req.Permissions), zap.Error(err))
		if strings.Contains(err.Error(), "does not exist") {
			response.Error(c, http.StatusNotFound, "file does not exist")
		} else {
			response.Error(c, http.StatusInternalServerError, "failed to set file permissions")
		}
		return
	}

	response.Success(c, nil)
}

// GetStorageInfo 获取存储信息
// @Summary 获取存储信息
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=models.StorageInfo}
// @Failure 500 {object} response.Response
// @Router /api/v1/file/storage/info [get]
func (h *Handler) GetStorageInfo(c *gin.Context) {
	info, err := h.fileService.GetStorageInfo(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get storage info", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get storage info")
		return
	}

	response.Success(c, info)
}

// GetStorageHealth 检查存储健康状态
// @Summary 检查存储健康状态
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=models.StorageHealth}
// @Failure 500 {object} response.Response
// @Router /api/v1/file/storage/health [get]
func (h *Handler) GetStorageHealth(c *gin.Context) {
	health, err := h.fileService.GetStorageHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to check storage health", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to check storage health")
		return
	}

	response.Success(c, health)
}

// CleanupStorage 清理存储空间
// @Summary 清理存储空间
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.CleanupOptions true "清理选项"
// @Success 200 {object} response.Response{data=models.CleanupResult}
// @Failure 500 {object} response.Response
// @Router /api/v1/file/storage/cleanup [post]
func (h *Handler) CleanupStorage(c *gin.Context) {
	var options models.CleanupOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		// 使用默认选项
		options = models.CleanupOptions{
			CleanTempFiles:  true,
			RemoveEmptyDirs: true,
			MaxAgeDays:      30,
		}
	}

	result, err := h.fileService.CleanupStorage(c.Request.Context(), &options)
	if err != nil {
		h.logger.Error("Failed to cleanup storage", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to cleanup storage")
		return
	}

	response.Success(c, result)
}

// CreateBackup 创建备份
// @Summary 创建备份
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.CreateBackupRequest true "创建备份请求"
// @Success 200 {object} response.Response{data=models.BackupInfo}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/backup [post]
func (h *Handler) CreateBackup(c *gin.Context) {
	var req models.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.ValidateError(c, err)
		return
	}

	backupInfo, err := h.fileService.CreateBackup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create backup", zap.String("name", req.Name), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to create backup")
		return
	}

	response.Success(c, backupInfo)
}

// ListBackups 列出备份列表
// @Summary 列出备份列表
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.BackupInfo}
// @Failure 500 {object} response.Response
// @Router /api/v1/file/backup/list [get]
func (h *Handler) ListBackups(c *gin.Context) {
	backups, err := h.fileService.ListBackups(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list backups", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to list backups")
		return
	}

	response.Success(c, backups)
}

// RestoreBackup 恢复备份
// @Summary 恢复备份
// @Tags file
// @Accept json
// @Produce json
// @Param id path string true "备份ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/backup/{id}/restore [post]
func (h *Handler) RestoreBackup(c *gin.Context) {
	backupID := c.Param("id")
	if backupID == "" {
		response.Error(c, http.StatusBadRequest, "backup ID is required")
		return
	}

	if err := h.fileService.RestoreBackup(c.Request.Context(), backupID); err != nil {
		h.logger.Error("Failed to restore backup", zap.String("backupID", backupID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to restore backup")
		return
	}

	response.Success(c, nil)
}

// DeleteBackup 删除备份
// @Summary 删除备份
// @Tags file
// @Accept json
// @Produce json
// @Param id path string true "备份ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/backup/{id} [delete]
func (h *Handler) DeleteBackup(c *gin.Context) {
	backupID := c.Param("id")
	if backupID == "" {
		response.Error(c, http.StatusBadRequest, "backup ID is required")
		return
	}

	if err := h.fileService.DeleteBackup(c.Request.Context(), backupID); err != nil {
		h.logger.Error("Failed to delete backup", zap.String("backupID", backupID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete backup")
		return
	}

	response.Success(c, nil)
}

// SearchFiles 搜索文件
// @Summary 搜索文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.FileSearchQuery true "文件搜索查询"
// @Success 200 {object} response.Response{data=models.FileSearchResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/search [post]
func (h *Handler) SearchFiles(c *gin.Context) {
	var query models.FileSearchQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.fileService.SearchFiles(c.Request.Context(), &query)
	if err != nil {
		h.logger.Error("Failed to search files", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to search files")
		return
	}

	response.Success(c, result)
}

// ProcessFile 处理文件
// @Summary 处理文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body models.FileProcessRequest true "文件处理请求"
// @Success 200 {object} response.Response{data=models.FileProcessResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/file/process [post]
func (h *Handler) ProcessFile(c *gin.Context) {
	var req models.FileProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.fileService.ProcessFile(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process file", zap.String("operation", req.Operation), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to process file")
		return
	}

	response.Success(c, result)
}

// GetHandlerInfo 获取处理器信息（用于监控）
func (h *Handler) GetHandlerInfo() *models.FileHandlerInfo {
	return &models.FileHandlerInfo{
		HandlerName: "FileHandler",
		Version:     "1.0.0",
		RegisteredRoutes: []string{
			"/api/v1/file/exists",
			"/api/v1/file/list",
			"/api/v1/file/read",
			"/api/v1/file/write",
			"/api/v1/file/delete",
			"/api/v1/file/move",
			"/api/v1/file/copy",
			"/api/v1/file/create-dir",
			"/api/v1/file/info",
			"/api/v1/file/hash",
			"/api/v1/file/permissions",
			"/api/v1/file/storage/info",
			"/api/v1/file/storage/health",
			"/api/v1/file/storage/cleanup",
			"/api/v1/file/backup",
			"/api/v1/file/backup/list",
			"/api/v1/file/backup/{id}/restore",
			"/api/v1/file/backup/{id}",
			"/api/v1/file/search",
			"/api/v1/file/process",
		},
	}
}
