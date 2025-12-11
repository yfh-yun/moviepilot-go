package file

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// ScrapeAction 实现文件刮削动作
type ScrapeAction struct {
	*common.BaseAction

	scrapedFiles []map[string]any // 成功刮削的文件列表
	hasError     bool             // 是否有错误
	failedCount  int              // 失败次数
}

// NewScrapeAction 创建新的文件刮削动作实例
func NewScrapeAction() base.Action {
	return &ScrapeAction{
		BaseAction:   common.NewBaseAction("scrape", base.ActionTypeFile),
		scrapedFiles: []map[string]any{},
		hasError:     false,
		failedCount:  0,
	}
}

// GetName 获取动作名称
func (a *ScrapeAction) GetName() string {
	return "刮削文件"
}

// GetDescription 获取动作描述
func (a *ScrapeAction) GetDescription() string {
	return "刮削媒体信息和图片"
}

// GetData 获取动作参数模板
func (a *ScrapeAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的ScrapeFileParams
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *ScrapeAction) Success() bool {
	// 动作是否成功，对应Python中的success属性
	return !a.hasError
}

// execute 执行文件刮削动作（核心逻辑）
func (a *ScrapeAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取fileitems
	fileitems, ok := ctx.GlobalContext["fileitems"].([]map[string]any)
	if !ok || len(fileitems) == 0 {
		return map[string]any{"success": true, "scraped_files": 0, "failed_files": 0}, nil
	}

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// mediaService, _ := ctx.Services["MediaService"].(interface{})
	// storageService, _ := ctx.Services["StorageService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("刮削文件失败: %v", r))
			a.hasError = true
		}
	}()

	// 初始化文件列表、错误状态和失败次数
	a.scrapedFiles = []map[string]any{}
	a.hasError = false
	a.failedCount = 0

	// 遍历处理每个fileitem
	for _, fileitem := range fileitems {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 获取文件路径
		filePath, ok := fileitem["path"].(string)
		if !ok || filePath == "" {
			ctx.Logger.Error("文件路径无效")
			a.failedCount++
			continue
		}

		// 检查文件是否已刮削过
		hasScraped := false
		for _, scrapedFile := range a.scrapedFiles {
			if scrapedFile["path"] == filePath {
				hasScraped = true
				break
			}
		}
		if hasScraped {
			continue
		}

		// 检查文件是否存在
		// TODO: 实现检查文件是否存在的逻辑
		// if not StorageChain().exists(fileitem):
		//     continue
		ctx.Logger.Info(fmt.Sprintf("检查文件是否存在: %s", filePath))

		// 检查缓存
		cacheKey := fmt.Sprintf("%s", filePath)
		if a.CheckCache(ctx.WorkflowID, cacheKey) {
			ctx.Logger.Info(fmt.Sprintf("%s 已刮削过，跳过", filePath))
			continue
		}

		// 识别媒体信息
		// TODO: 实现识别媒体信息的逻辑
		// meta = MetaInfoPath(Path(fileitem.path))
		// mediachain = MediaChain()
		// mediainfo = mediachain.recognize_media(meta)
		ctx.Logger.Info(fmt.Sprintf("识别媒体信息: %s", filePath))

		// 刮削元数据
		// TODO: 实现刮削元数据的逻辑
		// mediachain.scrape_metadata(fileitem=fileitem, meta=meta, mediainfo=mediainfo)
		ctx.Logger.Info(fmt.Sprintf("刮削元数据: %s", filePath))

		// 添加到刮削成功列表
		a.scrapedFiles = append(a.scrapedFiles, fileitem)

		// 保存缓存
		a.SaveCache(ctx.WorkflowID, cacheKey)
	}

	// 更新错误状态
	if len(a.scrapedFiles) == 0 && a.failedCount > 0 {
		a.hasError = true
	}

	// 标记任务完成
	message := fmt.Sprintf("成功刮削 %d 个文件，失败 %d 个", len(a.scrapedFiles), a.failedCount)
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":      !a.hasError,
		"scraped_files": len(a.scrapedFiles),
		"failed_files":  a.failedCount,
		"message":      message,
	}

	return output, nil
}
