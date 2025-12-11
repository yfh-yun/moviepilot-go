package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// FetchDownloadsAction 实现获取下载任务动作
type FetchDownloadsAction struct {
	*common.BaseAction
	downloads []map[string]any // 下载任务列表
}

// NewFetchDownloadsAction 创建新的获取下载任务动作实例
func NewFetchDownloadsAction() base.Action {
	return &FetchDownloadsAction{
		BaseAction: common.NewBaseAction("fetch_downloads", base.ActionTypeCore),
		downloads:  []map[string]any{},
	}
}

// GetName 获取动作名称
func (a *FetchDownloadsAction) GetName() string {
	return "获取下载任务"
}

// GetDescription 获取动作描述
func (a *FetchDownloadsAction) GetDescription() string {
	return "获取下载队列中的任务状态"
}

// GetData 获取动作参数模板
func (a *FetchDownloadsAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchDownloadsParams
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *FetchDownloadsAction) Success() bool {
	// 成功条件与Python一致：动作已完成
	return a.Done()
}

// execute 执行获取下载任务动作（核心逻辑）
func (a *FetchDownloadsAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取downloads
	downloads, ok := ctx.GlobalContext["downloads"].([]map[string]any)
	if !ok {
		// 如果上下文中没有downloads，使用动作自身的downloads
		downloads = a.downloads
	}

	if len(downloads) == 0 {
		// 没有下载任务，直接返回
		return map[string]any{
			"success":  true,
			"downloads_checked": 0,
		}, nil
	}

	// 获取下载服务实例
	downloadService, _ := ctx.Services["DownloadService"].(interface {
		// ListTorrents 获取种子列表（对应Python的list_torrents）
		ListTorrents(ctx any, hashs []string) ([]map[string]any, error)
	})

	allComplete := true

	// 遍历处理每个download
	for i, download := range downloads {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 获取download_id
		downloadID, ok := download["download_id"].(string)
		if !ok {
			continue
		}

		ctx.Logger.Info(fmt.Sprintf("获取下载任务 %s 状态 ...", downloadID))

		// 调用下载服务获取种子状态
		torrents := []map[string]any{}
		var err error
		if downloadService != nil {
			torrents, err = downloadService.ListTorrents(ctx, []string{downloadID})
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("获取下载任务 %s 状态失败: %s", downloadID, err.Error()))
				continue
			}
		}

		// 更新下载任务状态
		completed := false
		if len(torrents) == 0 {
			// 种子不存在，标记为已完成
			completed = true
			ctx.Logger.Info(fmt.Sprintf("下载任务 %s 已完成", downloadID))
		} else {
			// 检查每个种子的进度
			for _, t := range torrents {
				if progress, ok := t["progress"].(float64); ok && progress >= 100 {
					completed = true
					ctx.Logger.Info(fmt.Sprintf("下载任务 %s 已完成", downloadID))
					break
				} else {
					completed = false
					ctx.Logger.Info(fmt.Sprintf("下载任务 %s 未完成", downloadID))
				}
			}
		}

		// 更新download的completed状态
		download["completed"] = completed
		downloads[i] = download

		// 检查是否所有任务都已完成
		if !completed {
			allComplete = false
		}
	}

	// 如果所有任务都已完成，标记动作完成
	if allComplete {
		a.JobDone("所有下载任务已完成")
	}

	// 更新上下文
	ctx.GlobalContext["downloads"] = downloads

	// 输出结果
	output := map[string]any{
		"success":          true,
		"downloads_checked": len(downloads),
		"all_complete":     allComplete,
	}

	return output, nil
}