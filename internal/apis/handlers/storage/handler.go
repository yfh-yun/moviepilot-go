package storage

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middleware"
	"moviepilot-go/internal/business/services/storage"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// Handler 存储 API 处理器
type Handler struct {
	logger         *zap.Logger
	storageService *storage.StorageService
}

// NewHandler 创建存储 API 处理器
func NewHandler(storageService *storage.StorageService) *Handler {
	return &Handler{
		logger:         logger.GetLogger(),
		storageService: storageService,
	}
}

// QRCode 生成二维码
// @Summary 生成二维码内容
// @Description 生成二维码
// @Tags storage
// @Produce json
// @Param name path string true "存储名称"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/qrcode/{name} [get]
func (h *Handler) QRCode(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	h.logger.Debug("QRCode called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	qrcodeData, errmsg, err := h.storageService.GenerateQRCode(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("生成二维码失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if qrcodeData != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    qrcodeData,
			"message": errmsg,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errmsg,
		})
	}
}

// Check 二维码登录确认
// @Summary 二维码登录确认
// @Description 二维码登录确认
// @Tags storage
// @Produce json
// @Param name path string true "存储名称"
// @Param ck query string false "ck参数"
// @Param t query string false "t参数"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/check/{name} [get]
func (h *Handler) Check(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	ck := c.Query("ck")
	t := c.Query("t")
	h.logger.Debug("Check called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
		zap.String("ck", ck),
		zap.String("t", t),
	)

	data, errmsg, err := h.storageService.CheckLogin(c.Request.Context(), name, ck, t)
	if err != nil {
		h.logger.Error("二维码登录确认失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if data != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errmsg,
		})
	}
}

// SaveConfig 保存存储配置
// @Summary 保存存储配置
// @Description 保存存储配置
// @Tags storage
// @Accept json
// @Produce json
// @Param name path string true "存储名称"
// @Param conf body map[string]interface{} true "存储配置"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/save/{name} [post]
func (h *Handler) SaveConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	h.logger.Debug("SaveConfig called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	var conf map[string]any
	if err := c.ShouldBindJSON(&conf); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.storageService.SaveConfig(c.Request.Context(), name, conf)
	if err != nil {
		h.logger.Error("保存存储配置失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ResetConfig 重置存储配置
// @Summary 重置存储配置
// @Description 重置存储配置
// @Tags storage
// @Produce json
// @Param name path string true "存储名称"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/reset/{name} [get]
func (h *Handler) ResetConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	h.logger.Debug("ResetConfig called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	err := h.storageService.ResetConfig(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("重置存储配置失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ListFiles 所有目录和文件
// @Summary 所有目录和文件
// @Description 查询当前目录下所有目录和文件
// @Tags storage
// @Accept json
// @Produce json
// @Param fileitem body dto.FileItem true "文件项"
// @Param sort query string false "排序方式，name:按名称排序，time:按修改时间排序"
// @Success 200 {object} []dto.FileItem
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/list [post]
func (h *Handler) ListFiles(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	sort := c.DefaultQuery("sort", "updated_at")
	h.logger.Debug("ListFiles called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("sort", sort),
	)

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileList, err := h.storageService.ListFilesByItem(c.Request.Context(), &fileitem)
	if err != nil {
		h.logger.Error("获取文件列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// TODO: 实现排序逻辑
	c.JSON(http.StatusOK, fileList)
}

// Mkdir 创建目录
// @Summary 创建目录
// @Description 创建目录
// @Tags storage
// @Accept json
// @Produce json
// @Param fileitem body dto.FileItem true "文件项"
// @Param name query string true "目录名称"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/mkdir [post]
func (h *Handler) Mkdir(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Query("name")
	h.logger.Debug("Mkdir called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
		})
		return
	}

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.storageService.CreateFolder(c.Request.Context(), &fileitem, name)
	if result {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
	}
}

// Delete 删除文件或目录
// @Summary 删除文件或目录
// @Description 删除文件或目录
// @Tags storage
// @Accept json
// @Produce json
// @Param fileitem body dto.FileItem true "文件项"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/delete [post]
func (h *Handler) Delete(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	h.logger.Debug("Delete called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.storageService.DeleteFileByItem(c.Request.Context(), &fileitem)
	if result {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
	}
}

// Download 下载文件
// @Summary 下载文件
// @Description 下载文件或目录
// @Tags storage
// @Accept json
// @Produce octet-stream
// @Param fileitem body dto.FileItem true "文件项"
// @Success 200 {file} file
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/download [post]
func (h *Handler) Download(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	h.logger.Debug("Download called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpFile, err := h.storageService.DownloadFile(c.Request.Context(), &fileitem)
	if err != nil {
		h.logger.Error("下载文件失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if tmpFile != "" {
		c.File(tmpFile)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
	}
}

// Image 预览图片
// @Summary 预览图片
// @Description 预览图片
// @Tags storage
// @Accept json
// @Produce image/jpeg
// @Param fileitem body dto.FileItem true "文件项"
// @Success 200 {file} file
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/image [post]
func (h *Handler) Image(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	h.logger.Debug("Image called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpFile, err := h.storageService.DownloadFile(c.Request.Context(), &fileitem)
	if err != nil {
		h.logger.Error("下载图片失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if tmpFile != "" {
		c.File(tmpFile)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "图片读取出错",
		})
	}
}

// Rename 重命名文件或目录
// @Summary 重命名文件或目录
// @Description 重命名文件或目录
// @Tags storage
// @Accept json
// @Produce json
// @Param fileitem body dto.FileItem true "文件项"
// @Param new_name query string true "新名称"
// @Param recursive query bool false "是否递归修改"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/rename [post]
func (h *Handler) Rename(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	newName := c.Query("new_name")
	recursiveStr := c.DefaultQuery("recursive", "false")
	recursive, _ := strconv.ParseBool(recursiveStr)
	h.logger.Debug("Rename called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("new_name", newName),
		zap.Bool("recursive", recursive),
	)

	if newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "新名称为空",
		})
		return
	}

	var fileitem dto.FileItem
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现递归重命名逻辑

	// 重命名自己
	result := h.storageService.RenameFile(c.Request.Context(), &fileitem, newName)
	if result {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
	}
}

// Usage 存储空间信息
// @Summary 存储空间信息
// @Description 查询存储空间
// @Tags storage
// @Produce json
// @Param name path string true "存储名称"
// @Success 200 {object} dto.StorageUsage
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/usage/{name} [get]
func (h *Handler) Usage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	h.logger.Debug("Usage called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	ret, err := h.storageService.GetStorageUsage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("获取存储空间信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if ret != nil {
		c.JSON(http.StatusOK, ret)
	} else {
		c.JSON(http.StatusOK, &dto.StorageUsage{})
	}
}

// TransType 支持的整理方式获取
// @Summary 支持的整理方式获取
// @Description 查询支持的整理方式
// @Tags storage
// @Produce json
// @Param name path string true "存储名称"
// @Success 200 {object} dto.StorageTransType
// @Failure 500 {object} map[string]interface{}
// @Router /api/storage/transtype/{name} [get]
func (h *Handler) TransType(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	name := c.Param("name")
	h.logger.Debug("TransType called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("name", name),
	)

	ret, err := h.storageService.SupportTransType(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("获取支持的整理方式失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if ret != nil {
		c.JSON(http.StatusOK, &dto.StorageTransType{TransType: ret})
	} else {
		c.JSON(http.StatusOK, &dto.StorageTransType{})
	}
}
