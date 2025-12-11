package resource

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// DownloadsAction 实现下载资源获取动作
type DownloadsAction struct {
	*common.BaseAction
}

// NewDownloadsAction 创建新的下载资源获取动作实例
func NewDownloadsAction() base.Action {
	return &DownloadsAction{
		BaseAction: common.NewBaseAction("downloads", base.ActionTypeResource),
	}
}

// GetDescription 获取动作描述
func (a *DownloadsAction) GetDescription() string {
	return "获取下载队列中的任务状态"
}

// GetData 获取动作参数模板
func (a *DownloadsAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchDownloadsParams
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *DownloadsAction) Success() bool {
	// 动作是否成功取决于是否完成，对应Python中的success属性
	return a.Done()
}

// execute 执行下载资源获取动作（核心逻辑）
func (a *DownloadsAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取downloads
	downloads, ok := ctx.GlobalContext["downloads"].([]map[string]any)
	if !ok || len(downloads) == 0 {
		return map[string]any{"success": true, "downloads_checked": 0}, nil
	}

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// downloadService, _ := ctx.Services["DownloadService"].(interface{})

	// 遍历处理每个download
	allComplete := true
	for i, download := range downloads {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		downloadID, _ := download["download_id"].(string)
		ctx.Logger.Info(fmt.Sprintf("获取下载任务 %s 状态 ...", downloadID))

		// 获取下载任务状态
		// TODO: 实现获取下载任务状态的逻辑
		// torrents := ActionChain().list_torrents(hashs=[download.download_id])
		ctx.Logger.Debug(fmt.Sprintf("获取下载任务 %s 的状态", downloadID))

		// 更新下载任务状态
		// TODO: 实现更新下载任务状态的逻辑
		// 这里模拟下载任务状态，实际应该调用downloadService获取
		completed := false
		// 随机模拟完成状态
		// if rand.Float32() > 0.5 {
		//     completed = true
		//     ctx.Logger.Info(fmt.Sprintf("下载任务 %s 已完成", downloadID))
		// } else {
		//     ctx.Logger.Info(fmt.Sprintf("下载任务 %s 未完成", downloadID))
		// }

		// 更新downloads中的下载任务状态
		downloads[i]["completed"] = completed
		if !completed {
			allComplete = false
		}
	}

	// 如果所有下载任务都完成，标记任务完成
	if allComplete {
		message := "所有下载任务已完成"
		ctx.Logger.Info(message)
		a.JobDone(message)
	}

	// 更新上下文中的downloads
	ctx.GlobalContext["downloads"] = downloads

	// 输出结果
	output := map[string]any{
		"success":           true,
		"downloads_checked": len(downloads),
		"all_complete":      allComplete,
	}

	return output, nil
}
