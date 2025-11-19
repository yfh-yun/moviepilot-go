// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/service/actions/types"

	"go.uber.org/zap"
)

// TransferFileAction 整理文件动作
// 对应Python版本app/actions/transfer_file.py的TransferFileAction
type TransferFileAction struct {
	fileRepo         interfaces.FileRepository
	downloadRepo     interfaces.DownloadRepository
	transferRepo     interfaces.TransferHistoryRepository
	storageChain     *StorageChain
	transferChain    *TransferChain
	cache            *WorkflowCache
	transferredFiles []*types.File
	hasError         bool
	logger           *zap.Logger
}

// TransferFileParams 整理文件参数
type TransferFileParams struct {
	Source string `json:"source" description:"来源"` // downloads, fileitems
}

// TransferHistory 传输历史记录
type TransferHistory struct {
	ID         int64     `json:"id"`
	SrcPath    string    `json:"src_path"`
	DstPath    string    `json:"dst_path"`
	Storage    string    `json:"storage"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
}

// NewTransferFileAction 创建整理文件动作实例
func NewTransferFileAction(
	fileRepo interfaces.FileRepository,
	downloadRepo interfaces.DownloadRepository,
	transferRepo interfaces.TransferHistoryRepository,
	storageChain *StorageChain,
	transferChain *TransferChain,
	cache *WorkflowCache,
) *TransferFileAction {
	return &TransferFileAction{
		fileRepo:         fileRepo,
		downloadRepo:     downloadRepo,
		transferRepo:     transferRepo,
		storageChain:     storageChain,
		transferChain:    transferChain,
		cache:            cache,
		transferredFiles: make([]*types.File, 0),
		hasError:         false,
		logger:           logger.Logger,
	}
}

// Execute 执行整理文件动作
// 实现Python版本TransferFileAction.execute()方法的完整功能
func (tfa *TransferFileAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *types.ActionContext,
) (*types.ActionContext, error) {
	startTime := time.Now()
	failedCount := 0

	// 解析参数
	transferParams, err := tfa.parseParams(params)
	if err != nil {
		tfa.logger.Error("解析整理文件参数失败", zap.Error(err))
		return actionCtx, err
	}

	tfa.logger.Info("开始执行整理文件动作",
		zap.Int64("workflow_id", workflowID),
		zap.String("source", transferParams.Source),
		zap.Int("download_count", len(actionCtx.Downloads)),
		zap.Int("file_count", len(actionCtx.Files)),
	)

	// 检查继续函数
	checkContinue := func() bool {
		return tfa.checkContinue(ctx, workflowID)
	}

	if transferParams.Source == "downloads" {
		// 从下载任务中整理文件
		failedCount = tfa.transferFromDownloads(ctx, workflowID, actionCtx, checkContinue)
	} else {
		// 从文件项中整理文件
		failedCount = tfa.transferFromFileItems(ctx, workflowID, actionCtx, checkContinue)
	}

	// 更新动作上下文
	if len(tfa.transferredFiles) > 0 {
		actionCtx.Files = append(actionCtx.Files, tfa.transferredFiles...)
		tfa.logger.Info("文件整理完成",
			zap.Int("success_count", len(tfa.transferredFiles)),
			zap.Int("failed_count", failedCount),
			zap.Duration("duration", time.Since(startTime)),
		)
	} else if failedCount > 0 {
		tfa.hasError = true
		tfa.logger.Error("所有文件整理都失败了", zap.Int("failed_count", failedCount))
	} else {
		tfa.logger.Info("没有文件需要整理", zap.Duration("duration", time.Since(startTime)))
	}

	return actionCtx, nil
}

// parseParams 解析动作参数
func (tfa *TransferFileAction) parseParams(params map[string]interface{}) (*TransferFileParams, error) {
	transferParams := &TransferFileParams{
		Source: "downloads", // 默认来源
	}

	if source, ok := params["source"].(string); ok {
		transferParams.Source = source
	}

	return transferParams, nil
}

// transferFromDownloads 从下载任务中整理文件
func (tfa *TransferFileAction) transferFromDownloads(
	ctx context.Context,
	workflowID int64,
	actionCtx *types.ActionContext,
	checkContinue func() bool,
) int {
	failedCount := 0

	for _, download := range actionCtx.Downloads {
		// 检查工作流是否已停止
		if !checkContinue() {
			tfa.logger.Info("工作流已停止，终止文件整理", zap.Int64("workflow_id", workflowID))
			break
		}

		// 检查下载是否完成
		if !tfa.isDownloadCompleted(download) {
			tfa.logger.Info("下载任务未完成，跳过整理", zap.String("download_id", download.ID))
			continue
		}

		// 检查缓存
		cacheKey := download.ID
		if tfa.checkCache(ctx, workflowID, cacheKey) {
			tfa.logger.Info("下载任务已整理过，跳过", zap.String("download_id", download.ID))
			continue
		}

		// 执行文件整理
		success, fileItem := tfa.transferSingleFile(ctx, download.DownloadID, download, checkContinue)
		if success && fileItem != nil {
			tfa.transferredFiles = append(tfa.transferredFiles, fileItem)
			// 保存缓存
			if err := tfa.saveCache(ctx, workflowID, cacheKey); err != nil {
				tfa.logger.Warn("保存缓存失败", zap.Error(err))
			}
			tfa.logger.Info("文件整理成功", zap.String("path", fileItem.Path))
		} else {
			failedCount++
			tfa.logger.Error("文件整理失败", zap.String("download_id", download.ID))
		}
	}

	return failedCount
}

// transferFromFileItems 从文件项中整理文件
func (tfa *TransferFileAction) transferFromFileItems(
	ctx context.Context,
	workflowID int64,
	actionCtx *types.ActionContext,
	checkContinue func() bool,
) int {
	failedCount := 0

	// 创建文件副本以避免修改原始切片
	filesToProcess := make([]*types.File, len(actionCtx.Files))
	copy(filesToProcess, actionCtx.Files)

	for _, file := range filesToProcess {
		// 检查工作流是否已停止
		if !checkContinue() {
			tfa.logger.Info("工作流已停止，终止文件整理", zap.Int64("workflow_id", workflowID))
			break
		}

		// 检查缓存
		cacheKey := file.Path
		if tfa.checkCache(ctx, workflowID, cacheKey) {
			tfa.logger.Info("文件已整理过，跳过", zap.String("path", file.Path))
			continue
		}

		// 检查是否已整理过
		if tfa.isFileAlreadyTransferred(file) {
			continue
		}

		// 执行文件整理
		success, transferredFile := tfa.transferFileItem(ctx, file, checkContinue)
		if success && transferredFile != nil {
			tfa.transferredFiles = append(tfa.transferredFiles, transferredFile)
			// 从原始文件列表中移除已整理的文件
			tfa.removeFileFromList(actionCtx.Files, file)
			// 记录已整理的文件
			if err := tfa.saveCache(ctx, workflowID, cacheKey); err != nil {
				tfa.logger.Warn("保存缓存失败", zap.Error(err))
			}
			tfa.logger.Info("文件整理成功", zap.String("path", file.Path))
		} else {
			failedCount++
			tfa.logger.Error("文件整理失败", zap.String("path", file.Path))
		}
	}

	return failedCount
}

// transferSingleFile 整理单个下载文件
func (tfa *TransferFileAction) transferSingleFile(
	ctx context.Context,
	downloadID string,
	download *types.Download,
	checkContinue func() bool,
) (bool, *types.File) {
	// 获取文件项
	fileItem, err := tfa.storageChain.GetFileItem("local", download.URL)
	if err != nil {
		tfa.logger.Error("获取文件项失败", zap.String("path", download.URL), zap.Error(err))
		return false, nil
	}

	if fileItem == nil {
		tfa.logger.Error("文件不存在", zap.String("path", download.URL))
		return false, nil
	}

	// 检查是否已传输过
	if tfa.isAlreadyTransferred(fileItem.Path, fileItem.Storage) {
		tfa.logger.Info("文件已传输过，跳过", zap.String("path", fileItem.Path))
		return false, nil
	}

	// 执行传输
	tfa.logger.Info("开始整理文件", zap.String("path", fileItem.Path))
	state, errMsg := tfa.transferChain.DoTransfer(fileItem, false, checkContinue)
	if !state {
		tfa.logger.Error("整理文件失败", zap.String("path", fileItem.Path), zap.String("error", errMsg))
		return false, nil
	}

	// 转换为文件对象
	file := tfa.convertToFileItem(fileItem)
	file.TransferPath = fileItem.DstPath
	file.Status = "transferred"
	file.TransferredAt = &[]time.Time{time.Now()}[0]
	file.UpdatedAt = time.Now()

	return true, file
}

// transferFileItem 整理文件项
func (tfa *TransferFileAction) transferFileItem(
	ctx context.Context,
	file *types.File,
	checkContinue func() bool,
) (bool, *types.File) {
	// 检查是否已传输过
	if tfa.isAlreadyTransferred(file.Path, "local") {
		return false, nil
	}

	// 转换为文件项
	fileItem := &FileItem{
		Path:       file.Path,
		Name:       file.Name,
		Extension:  file.Extension,
		Size:       file.Size,
		ParentPath: file.ParentPath,
		Storage:    "local",
	}

	// 执行传输
	tfa.logger.Info("开始整理文件", zap.String("path", file.Path))
	state, errMsg := tfa.transferChain.DoTransfer(fileItem, false, checkContinue)
	if !state {
		tfa.logger.Error("整理文件失败", zap.String("path", file.Path), zap.String("error", errMsg))
		return false, nil
	}

	// 创建传输后的文件对象
	transferredFile := *file
	transferredFile.TransferPath = fileItem.DstPath
	transferredFile.Status = "transferred"
	transferredFile.TransferredAt = &[]time.Time{time.Now()}[0]
	transferredFile.UpdatedAt = time.Now()

	return true, &transferredFile
}

// isDownloadCompleted 检查下载是否完成
func (tfa *TransferFileAction) isDownloadCompleted(download *types.Download) bool {
	return download.Status == "completed"
}

// isAlreadyTransferred 检查文件是否已传输过
func (tfa *TransferFileAction) isAlreadyTransferred(path, storage string) bool {
	// 查询传输历史
	history, err := tfa.transferRepo.GetBySrc(context.Background(), path, storage)
	if err != nil {
		tfa.logger.Warn("查询传输历史失败", zap.Error(err))
		return false
	}

	return history != nil
}

// isFileAlreadyTransferred 检查文件是否已整理过
func (tfa *TransferFileAction) isFileAlreadyTransferred(file *types.File) bool {
	return tfa.isAlreadyTransferred(file.Path, "local")
}

// convertToFileItem 转换为文件对象
func (tfa *TransferFileAction) convertToFileItem(item *FileItem) *types.File {
	return &types.File{
		Path:       item.Path,
		Name:       item.Name,
		Extension:  item.Extension,
		Size:       item.Size,
		ParentPath: item.ParentPath,
		Status:     "local",
		UserID:     0, // 默认用户
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// removeFileFromList 从文件列表中移除文件
func (tfa *TransferFileAction) removeFileFromList(files []*types.File, target *types.File) {
	for i, file := range files {
		if file.Path == target.Path {
			// 删除元素
			files = append(files[:i], files[i+1:]...)
			break
		}
	}
}

// checkContinue 检查是否继续整理文件
func (tfa *TransferFileAction) checkContinue(ctx context.Context, workflowID int64) bool {
	// 检查工作流是否已停止
	return !tfa.isWorkflowStopped(ctx, workflowID)
}

// checkCache 检查缓存
func (tfa *TransferFileAction) checkCache(ctx context.Context, workflowID int64, key string) bool {
	if tfa.cache == nil {
		return false
	}

	cacheKey := fmt.Sprintf("transfer_cache_%d", workflowID)
	exists, err := tfa.cache.Exists(ctx, cacheKey, key)
	if err != nil {
		tfa.logger.Warn("检查缓存失败", zap.Error(err))
		return false
	}

	return exists
}

// saveCache 保存缓存
func (tfa *TransferFileAction) saveCache(ctx context.Context, workflowID int64, key string) error {
	if tfa.cache == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("transfer_cache_%d", workflowID)
	return tfa.cache.Set(ctx, cacheKey, key, 24*time.Hour)
}

// isWorkflowStopped 检查工作流是否已停止
func (tfa *TransferFileAction) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// GetSuccess 获取执行结果
func (tfa *TransferFileAction) GetSuccess() bool {
	return !tfa.hasError
}

// GetTransferredFiles 获取已传输的文件列表
func (tfa *TransferFileAction) GetTransferredFiles() []*types.File {
	return tfa.transferredFiles
}

// GetName 获取动作名称
func (tfa *TransferFileAction) GetName() string {
	return "整理文件"
}

// GetDescription 获取动作描述
func (tfa *TransferFileAction) GetDescription() string {
	return "整理队列中的文件"
}

// GetData 获取动作参数定义
func (tfa *TransferFileAction) GetData() map[string]interface{} {
	return map[string]interface{}{
		"source": map[string]interface{}{
			"type":        "string",
			"description": "来源",
			"default":     "downloads",
			"options":     []string{"downloads", "fileitems"},
		},
	}
}

// TransferFile 转移文件方法（提供给其他服务调用）
func (tfa *TransferFileAction) TransferFile(
	ctx context.Context,
	filePath string,
	sourceStorage string,
) (*types.File, error) {
	// 创建文件对象
	file := &types.File{
		Path:      filePath,
		Name:      filepath.Base(filePath),
		Extension: filepath.Ext(filePath),
		Status:    "local",
		UserID:    0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 执行转移
	success, transferredFile := tfa.transferFileItem(ctx, file, func() bool { return true })
	if !success {
		return nil, fmt.Errorf("文件转移失败")
	}

	return transferredFile, nil
}

// BatchTransferFiles 批量转移文件
func (tfa *TransferFileAction) BatchTransferFiles(
	ctx context.Context,
	filePaths []string,
	sourceStorage string,
) ([]*types.File, error) {
	var transferredFiles []*types.File

	for _, filePath := range filePaths {
		file, err := tfa.TransferFile(ctx, filePath, sourceStorage)
		if err != nil {
			tfa.logger.Error("批量转移文件失败", zap.String("path", filePath), zap.Error(err))
			continue
		}
		transferredFiles = append(transferredFiles, file)
	}

	return transferredFiles, nil
}

// GetTransferStatus 获取转移状态
func (tfa *TransferFileAction) GetTransferStatus(ctx context.Context, filePath string) (*TransferStatus, error) {
	// 查询传输历史
	history, err := tfa.transferRepo.GetBySrc(ctx, filePath, "local")
	if err != nil {
		return nil, err
	}

	if history == nil {
		return &TransferStatus{
			FilePath: filePath,
			Status:   "not_transferred",
		}, nil
	}

	return &TransferStatus{
		FilePath:     filePath,
		Status:       history.Status,
		Destination:  history.DstPath,
		TransferTime: history.CreateTime,
	}, nil
}

// TransferStatus 转移状态
type TransferStatus struct {
	FilePath     string    `json:"file_path"`
	Status       string    `json:"status"`
	Destination  string    `json:"destination"`
	TransferTime time.Time `json:"transfer_time"`
	ErrorMessage string    `json:"error_message,omitempty"`
}
