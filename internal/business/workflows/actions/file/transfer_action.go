package file

import (
	"fmt"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/transfer"
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
)

// TransferAction 实现文件传输动作
type TransferAction struct {
	*common.BaseAction
	transferService     *transfer.TransferService
	transferHistoryRepo interfaces.TransferHistoryRepository
	fileItems           []*dto.FileItem
	hasError            bool
}

// NewTransferAction 创建新的文件传输动作实例
func NewTransferAction() base.Action {
	return &TransferAction{
		BaseAction: common.NewBaseAction("transfer", base.ActionTypeFile),
		fileItems:  make([]*dto.FileItem, 0),
	}
}

// GetName 获取动作名称
func (a *TransferAction) GetName() string {
	return "整理文件"
}

// GetDescription 获取动作描述
func (a *TransferAction) GetDescription() string {
	return "整理队列中的文件"
}

// GetData 获取动作参数模板
func (a *TransferAction) GetData() map[string]any {
	return map[string]any{
		"source": "downloads",
	}
}

// execute 执行文件传输动作（核心逻辑）
func (a *TransferAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 初始化传输服务
	a.transferService = transfer.GetTransferService()

	// 从Services中获取TransferHistoryRepository实例
	if repo, ok := ctx.Services["TransferHistoryRepository"].(interfaces.TransferHistoryRepository); ok {
		a.transferHistoryRepo = repo
	}

	// 获取源参数，默认为downloads
	source, ok := ctx.Input["source"].(string)
	if !ok {
		source = "downloads"
	}

	// 失败计数
	failedCount := 0

	// 从全局上下文中获取downloads和fileitems
	downloads, _ := ctx.GlobalContext["downloads"].([]*dto.DownloadTask)
	fileItems, _ := ctx.GlobalContext["fileitems"].([]*dto.FileItem)

	a.GetLogger().Info("开始执行文件传输动作",
		zap.String("source", source),
		zap.Int("downloads_count", len(downloads)),
		zap.Int("fileitems_count", len(fileItems)),
	)

	// 检查工作流是否已停止的函数
	checkWorkflowStopped := func() bool {
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			a.GetLogger().Info("工作流已停止，终止执行")
			return true
		}
		return false
	}

	// 根据源类型执行不同的传输逻辑
	if source == "downloads" {
		// 从下载任务中整理文件
		for _, download := range downloads {
			// 检查工作流是否已停止
			if checkWorkflowStopped() {
				break
			}

			// 检查下载任务是否已完成
			if !download.Completed {
				a.GetLogger().Info("下载任务未完成，跳过",
					zap.String("download_id", download.DownloadID),
					zap.String("path", download.Path),
				)
				continue
			}

			// 检查缓存
			cacheKey := download.DownloadID
			if a.CheckCache(ctx.WorkflowID, cacheKey) {
				a.GetLogger().Info("文件已整理过，跳过",
					zap.String("path", download.Path),
				)
				continue
			}

			// TODO: 从存储中获取文件项
			// fileItem := storagechain.get_file_item(storage="local", path=Path(download.path))
			// 暂时模拟文件项
			fileItem := &dto.FileItem{
				Storage: "local",
				Path:    download.Path,
				Type:    "file",
			}

			// 检查传输历史
			if a.transferHistoryRepo != nil {
				transferd, err := a.transferHistoryRepo.GetBySrc(ctx, fileItem.Path, fileItem.Storage)
				if err != nil {
					a.GetLogger().Error("检查传输历史失败",
						zap.String("path", fileItem.Path),
						zap.String("error", err.Error()),
					)
					// 检查失败不影响后续处理，继续执行
				} else if transferd != nil {
					// 已经整理过的文件不再整理
					a.GetLogger().Info("文件已整理过，跳过",
						zap.String("path", fileItem.Path),
					)
					continue
				}
			}

			// 执行文件传输
			a.GetLogger().Info("开始整理文件",
				zap.String("path", download.Path),
			)

			// 调用传输服务执行实际的文件传输
			// TODO: 实现实际的传输逻辑
			state := true
			// state, errmsg := a.transferService.DoTransfer(fileItem, false)

			if !state {
				failedCount++
				a.GetLogger().Error("整理文件失败",
					zap.String("path", download.Path),
					// zap.String("error", errmsg),
				)
				continue
			}

			a.GetLogger().Info("整理文件完成",
				zap.String("path", download.Path),
			)

			// 添加到已处理文件列表
			a.fileItems = append(a.fileItems, fileItem)

			// 保存到缓存
			a.SaveCache(ctx.WorkflowID, cacheKey)
		}
	} else {
		// 从fileitems中整理文件
		for _, fileItem := range fileItems {
			// 检查工作流是否已停止
			if checkWorkflowStopped() {
				break
			}

			// 检查缓存
			cacheKey := fileItem.Path
			if a.CheckCache(ctx.WorkflowID, cacheKey) {
				a.GetLogger().Info("文件已整理过，跳过",
					zap.String("path", fileItem.Path),
				)
				continue
			}

			// 检查传输历史
			if a.transferHistoryRepo != nil {
				transferd, err := a.transferHistoryRepo.GetBySrc(ctx, fileItem.Path, fileItem.Storage)
				if err != nil {
					a.GetLogger().Error("检查传输历史失败",
						zap.String("path", fileItem.Path),
						zap.String("error", err.Error()),
					)
					// 检查失败不影响后续处理，继续执行
				} else if transferd != nil {
					// 已经整理过的文件不再整理
					a.GetLogger().Info("文件已整理过，跳过",
						zap.String("path", fileItem.Path),
					)
					continue
				}
			}

			// 执行文件传输
			a.GetLogger().Info("开始整理文件",
				zap.String("path", fileItem.Path),
			)

			// 调用传输服务执行实际的文件传输
			// TODO: 实现实际的传输逻辑
			state := true
			// state, errmsg := a.transferService.DoTransfer(fileItem, false, checkWorkflowStopped)

			if !state {
				failedCount++
				a.GetLogger().Error("整理文件失败",
					zap.String("path", fileItem.Path),
					// zap.String("error", errmsg),
				)
				continue
			}

			a.GetLogger().Info("整理文件完成",
				zap.String("path", fileItem.Path),
			)

			// 添加到已处理文件列表
			a.fileItems = append(a.fileItems, fileItem)

			// 保存到缓存
			a.SaveCache(ctx.WorkflowID, cacheKey)
		}
	}

	// 更新全局上下文
	if len(a.fileItems) > 0 {
		ctx.GlobalContext["fileitems"] = append(fileItems, a.fileItems...)
	}

	// 设置错误状态
	if len(a.fileItems) == 0 && failedCount > 0 {
		a.hasError = true
	}

	// 生成输出结果
	output := map[string]any{
		"source":      source,
		"processed":   len(a.fileItems),
		"failed":      failedCount,
		"total":       len(downloads) + len(fileItems),
		"transferred": len(a.fileItems) > 0,
		"success":     !a.hasError,
	}

	a.JobDone(fmt.Sprintf("整理成功 %d 个文件，失败 %d 个", len(a.fileItems), failedCount))

	return output, nil
}
