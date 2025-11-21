// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// ScanFileAction 扫描文件动作
// 对应Python版本app/actions/scan_file.py的ScanFileAction
type ScanFileAction struct {
	fileRepo     interfaces.FileRepository
	storageChain *StorageChain
	fileItems    []*types.File
	hasError     bool
	logger       *zap.Logger
}

// ScanFileParams 扫描文件参数
type ScanFileParams struct {
	Storage    string   `json:"storage" description:"存储类型"`
	Directory  string   `json:"directory" description:"扫描目录"`
	Recursive  bool     `json:"recursive" description:"是否递归扫描"`
	Extensions []string `json:"extensions" description:"文件扩展名列表"`
}

// FileItem 文件项
type FileItem struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	IsDirectory  bool      `json:"is_directory"`
	ParentPath   string    `json:"parent_path"`
	Storage      string    `json:"storage"`
}

// NewScanFileAction 创建扫描文件动作实例
func NewScanFileAction(
	fileRepo interfaces.FileRepository,
	storageChain *StorageChain,
) *ScanFileAction {
	return &ScanFileAction{
		fileRepo:     fileRepo,
		storageChain: storageChain,
		fileItems:    make([]*types.File, 0),
		hasError:     false,
		logger:       logger.Logger,
	}
}

// Execute 执行扫描文件动作
// 实现Python版本ScanFileAction.execute()方法的完整功能
func (sfa *ScanFileAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *types.ActionContext,
) (*types.ActionContext, error) {
	startTime := time.Now()

	// 解析参数
	scanParams, err := sfa.parseParams(params)
	if err != nil {
		sfa.logger.Error("解析扫描文件参数失败", zap.Error(err))
		sfa.hasError = true
		return actionCtx, err
	}

	// 验证参数
	if scanParams.Storage == "" || scanParams.Directory == "" {
		sfa.logger.Error("存储或目录参数为空")
		sfa.hasError = true
		return actionCtx, fmt.Errorf("存储和目录参数不能为空")
	}

	sfa.logger.Info("开始执行扫描文件动作",
		zap.Int64("workflow_id", workflowID),
		zap.String("storage", scanParams.Storage),
		zap.String("directory", scanParams.Directory),
		zap.Bool("recursive", scanParams.Recursive),
	)

	// 获取存储文件项
	fileItem, err := sfa.storageChain.GetFileItem(scanParams.Storage, scanParams.Directory)
	if err != nil {
		sfa.logger.Error("获取存储文件项失败",
			zap.String("storage", scanParams.Storage),
			zap.String("directory", scanParams.Directory),
			zap.Error(err),
		)
		sfa.hasError = true
		return actionCtx, err
	}

	if fileItem == nil {
		sfa.logger.Error("目录不存在",
			zap.String("storage", scanParams.Storage),
			zap.String("directory", scanParams.Directory),
		)
		sfa.hasError = true
		return actionCtx, fmt.Errorf("目录不存在: 【%s】%s", scanParams.Storage, scanParams.Directory)
	}

	// 列出文件
	files, err := sfa.storageChain.ListFiles(fileItem, scanParams.Recursive)
	if err != nil {
		sfa.logger.Error("列出文件失败", zap.Error(err))
		sfa.hasError = true
		return actionCtx, err
	}

	// 处理扫描到的文件
	mediaExtensions := sfa.getMediaExtensions()
	for _, file := range files {
		// 检查工作流是否已停止
		if sfa.isWorkflowStopped(ctx, workflowID) {
			sfa.logger.Info("工作流已停止，终止文件扫描", zap.Int64("workflow_id", workflowID))
			break
		}

		// 检查文件扩展名
		if !sfa.isValidMediaFile(file, scanParams.Extensions, mediaExtensions) {
			continue
		}

		// 添加到文件列表
		sfa.fileItems = append(sfa.fileItems, sfa.convertToFileItem(file))
	}

	// 更新动作上下文
	if len(sfa.fileItems) > 0 {
		actionCtx.Files = append(actionCtx.Files, sfa.fileItems...)
		sfa.logger.Info("文件扫描完成",
			zap.Int("scanned_count", len(sfa.fileItems)),
			zap.Duration("duration", time.Since(startTime)),
		)
	} else {
		sfa.logger.Info("未扫描到媒体文件", zap.Duration("duration", time.Since(startTime)))
	}

	return actionCtx, nil
}

// parseParams 解析动作参数
func (sfa *ScanFileAction) parseParams(params map[string]interface{}) (*ScanFileParams, error) {
	scanParams := &ScanFileParams{
		Storage:   "local", // 默认本地存储
		Recursive: true,    // 默认递归扫描
	}

	// 解析存储类型
	if storage, ok := params["storage"].(string); ok {
		scanParams.Storage = storage
	}

	// 解析目录
	if directory, ok := params["directory"].(string); ok {
		scanParams.Directory = directory
	}

	// 解析是否递归
	if recursive, ok := params["recursive"].(bool); ok {
		scanParams.Recursive = recursive
	}

	// 解析扩展名列表
	if extensions, ok := params["extensions"].([]interface{}); ok {
		for _, ext := range extensions {
			if str, ok := ext.(string); ok {
				scanParams.Extensions = append(scanParams.Extensions, strings.ToLower(str))
			}
		}
	} else if extensionsStr, ok := params["extensions"].(string); ok {
		// 支持逗号分隔的字符串
		scanParams.Extensions = strings.Split(extensionsStr, ",")
		for i, ext := range scanParams.Extensions {
			scanParams.Extensions[i] = strings.TrimSpace(strings.ToLower(ext))
		}
	}

	return scanParams, nil
}

// getMediaExtensions 获取支持的媒体文件扩展名
func (sfa *ScanFileAction) getMediaExtensions() []string {
	// 对应Python版本中的settings.RMT_MEDIAEXT
	return []string{
		// 视频文件
		".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".3gp",
		".mpg", ".mpeg", ".ts", ".mts", ".m2ts", ".vob", ".f4v", ".asf", ".rm", ".rmvb",

		// 音频文件
		".mp3", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".wav", ".ape",

		// 字幕文件
		".srt", ".ass", ".ssa", ".sub", ".vtt",
	}
}

// isValidMediaFile 检查是否为有效的媒体文件
func (sfa *ScanFileAction) isValidMediaFile(file *FileItem, customExtensions, mediaExtensions []string) bool {
	// 检查是否有扩展名
	if file.Extension == "" {
		return false
	}

	ext := strings.ToLower(file.Extension)

	// 如果指定了自定义扩展名，优先使用
	if len(customExtensions) > 0 {
		for _, customExt := range customExtensions {
			if strings.HasPrefix(customExt, ".") {
				if ext == customExt {
					return true
				}
			} else {
				if "."+ext == customExt {
					return true
				}
			}
		}
		return false
	}

	// 检查是否为媒体文件扩展名
	for _, mediaExt := range mediaExtensions {
		if ext == mediaExt || "."+ext == mediaExt {
			return true
		}
	}

	return false
}

// convertToFileItem 转换为文件项
func (sfa *ScanFileAction) convertToFileItem(file *FileItem) *types.File {
	return &types.File{
		Path:       file.Path,
		Name:       file.Name,
		Extension:  file.Extension,
		Size:       file.Size,
		ParentPath: file.ParentPath,
		Status:     "local",
		UserID:     0, // 默认用户
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// isWorkflowStopped 检查工作流是否已停止
func (sfa *ScanFileAction) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// GetSuccess 获取执行结果
func (sfa *ScanFileAction) GetSuccess() bool {
	return !sfa.hasError
}

// GetFileItems 获取扫描到的文件列表
func (sfa *ScanFileAction) GetFileItems() []*types.File {
	return sfa.fileItems
}

// GetName 获取动作名称
func (sfa *ScanFileAction) GetName() string {
	return "扫描目录"
}

// GetDescription 获取动作描述
func (sfa *ScanFileAction) GetDescription() string {
	return "扫描目录文件到队列"
}

// GetData 获取动作参数定义
func (sfa *ScanFileAction) GetData() map[string]interface{} {
	return map[string]interface{}{
		"storage": map[string]interface{}{
			"type":        "string",
			"description": "存储类型",
			"default":     "local",
		},
		"directory": map[string]interface{}{
			"type":        "string",
			"description": "扫描目录",
			"default":     "",
		},
		"recursive": map[string]interface{}{
			"type":        "boolean",
			"description": "是否递归扫描",
			"default":     true,
		},
		"extensions": map[string]interface{}{
			"type":        "array",
			"description": "文件扩展名列表（为空则使用默认媒体扩展名）",
			"default":     []string{},
		},
	}
}

// ScanDirectory 直接扫描目录方法（提供给其他服务调用）
func (sfa *ScanFileAction) ScanDirectory(ctx context.Context, storage, directory string, recursive bool) ([]*types.File, error) {
	params := map[string]interface{}{
		"storage":   storage,
		"directory": directory,
		"recursive": recursive,
	}

	actionCtx := &types.ActionContext{
		Files: make([]*types.File, 0),
	}

	resultCtx, err := sfa.Execute(ctx, 0, params, actionCtx)
	if err != nil {
		return nil, err
	}

	return resultCtx.Files, nil
}

// ValidateDirectory 验证目录是否可访问
func (sfa *ScanFileAction) ValidateDirectory(ctx context.Context, storage, directory string) error {
	// 获取存储文件项
	fileItem, err := sfa.storageChain.GetFileItem(storage, directory)
	if err != nil {
		return err
	}

	if fileItem == nil {
		return fmt.Errorf("目录不存在")
	}

	// 检查目录是否可读
	fileInfo, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("无法访问目录: %w", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("路径不是目录")
	}

	// 检查读权限
	file, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("目录无读权限: %w", err)
	}
	defer file.Close()

	return nil
}

// GetDirectoryInfo 获取目录信息
func (sfa *ScanFileAction) GetDirectoryInfo(ctx context.Context, storage, directory string) (*DirectoryInfo, error) {
	fileItem, err := sfa.storageChain.GetFileItem(storage, directory)
	if err != nil {
		return nil, err
	}

	if fileItem == nil {
		return nil, fmt.Errorf("目录不存在")
	}

	// 获取目录信息
	fileInfo, err := os.Stat(directory)
	if err != nil {
		return nil, err
	}

	// 计算目录大小和文件数量
	var totalSize int64
	var fileCount int

	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &DirectoryInfo{
		Path:         directory,
		Storage:      storage,
		Size:         totalSize,
		FileCount:    fileCount,
		ModifiedTime: fileInfo.ModTime(),
		IsAccessible: err == nil,
	}, nil
}

// DirectoryInfo 目录信息
type DirectoryInfo struct {
	Path         string    `json:"path"`
	Storage      string    `json:"storage"`
	Size         int64     `json:"size"`
	FileCount    int       `json:"file_count"`
	ModifiedTime time.Time `json:"modified_time"`
	IsAccessible bool      `json:"is_accessible"`
}
