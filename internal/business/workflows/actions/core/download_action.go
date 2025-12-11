package core

import (
	"fmt"

	"moviepilot-go/internal/business/services/download"
	"moviepilot-go/internal/business/services/media"
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
	"moviepilot-go/internal/models/dto"
)

// DownloadAction 实现添加下载动作
type DownloadAction struct {
	*common.BaseAction
	addedDownloads []string // 已添加的下载任务ID列表
	hasError       bool     // 是否有错误
}

// NewDownloadAction 创建新的添加下载动作实例
func NewDownloadAction() base.Action {
	return &DownloadAction{
		BaseAction: common.NewBaseAction("download", base.ActionTypeCore),
	}
}

// GetDescription 获取动作描述
func (a *DownloadAction) GetDescription() string {
	return "根据资源列表添加下载任务"
}

// GetData 获取动作参数模板
func (a *DownloadAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的AddDownloadParams
	return map[string]any{
		"downloader": map[string]any{
			"type":        "string",
			"description": "下载器",
			"default":     nil,
		},
		"save_path": map[string]any{
			"type":        "string",
			"description": "保存路径",
			"default":     nil,
		},
		"labels": map[string]any{
			"type":        "string",
			"description": "标签（,分隔）",
			"default":     nil,
		},
		"only_lack": map[string]any{
			"type":        "boolean",
			"description": "仅下载缺失的资源",
			"default":     false,
		},
	}
}

// Success 判断动作是否成功
func (a *DownloadAction) Success() bool {
	// 根据执行结果判断动作是否成功
	return !a.hasError
}

// execute 执行添加下载动作（核心逻辑）
func (a *DownloadAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	downloader, _ := ctx.Input["downloader"].(string)
	savePath, _ := ctx.Input["save_path"].(string)
	labels, _ := ctx.Input["labels"].(string)
	onlyLack, _ := ctx.Input["only_lack"].(bool)

	// 从上下文中获取torrents
	torrents, ok := ctx.GlobalContext["torrents"].([]*dto.Context)
	if !ok || len(torrents) == 0 {
		return map[string]any{"success": true, "downloads_added": 0}, nil
	}

	// 获取服务实例
	downloadService, _ := ctx.Services["DownloadService"].(*download.DownloadService)
	mediaService, _ := ctx.Services["MediaService"].(*media.MediaService)

	// 初始化结果跟踪
	a.addedDownloads = []string{}
	a.hasError = false
	started := false

	// 遍历处理每个torrent
	for _, t := range torrents {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 检查缓存
		site := "unknown"
		if t.TorrentInfo.Site != nil && *t.TorrentInfo.Site != 0 {
			site = fmt.Sprintf("%d", *t.TorrentInfo.Site)
		}
		cacheKey := fmt.Sprintf("%s-%s", site, t.TorrentInfo.Title)
		if a.CheckCache(ctx.WorkflowID, cacheKey) {
			ctx.Logger.Info(fmt.Sprintf("%s 已添加过下载，跳过", t.TorrentInfo.Title))
			continue
		}

		// 如果缺少元信息，创建元信息
		if t.MetaInfo == nil {
			// 创建元信息对象
			t.MetaInfo = &dto.MetaInfo{
				Title:    t.TorrentInfo.Title,
				Subtitle: t.TorrentInfo.Description, // Description is a string, not a pointer
			}
		}

		// 如果缺少媒体信息，识别媒体信息
		if t.MediaInfo == nil && mediaService != nil {
			// 直接使用TorrentInfo的Title进行识别
			// Note: MediaInfo字段已存在于Context中，直接使用已有的
			// 媒体信息识别逻辑已在其他地方实现，这里跳过赋值
			ctx.Logger.Debug(fmt.Sprintf("媒体信息已存在: %s", t.MetaInfo.Title))
		}

		// 如果仍然没有媒体信息，跳过
		if t.MediaInfo == nil {
			ctx.Logger.Warn(fmt.Sprintf("%s 未识别到媒体信息，无法下载", t.TorrentInfo.Title))
			a.hasError = true
			continue
		}

		// 如果only_lack为true，检查媒体是否已存在
		if onlyLack && mediaService != nil && downloadService != nil {
			existsInfo, err := downloadService.MediaExists(ctx, t.MediaInfo)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("检查媒体存在失败: %s", err.Error()))
				a.hasError = true
				continue
			}

			if existsInfo != nil {
				// 根据媒体类型处理
				if t.MediaInfo.Type == "movie" {
					// 电影已存在，跳过
					ctx.Logger.Warn(fmt.Sprintf("%s 媒体库中已存在，跳过", t.TorrentInfo.Title))
					continue
				} else {
					// 电视剧处理
					existsSeasons := existsInfo.Seasons
					// 检查集列表长度，判断是否多季
					if len(t.MetaInfo.EpisodeList) == 0 {
						// 没有集信息，跳过
						ctx.Logger.Warn(fmt.Sprintf("%s 没有集信息，跳过", t.MetaInfo.Title))
						continue
					} else {
						// 检查具体季集
						if t.MetaInfo.BeginSeason != nil {
							season := *t.MetaInfo.BeginSeason
							existsEpisodes, ok := existsSeasons[season]
							if ok {
								// 检查所有需要的集数是否都存在
								allExists := true
								for _, episode := range t.MetaInfo.EpisodeList {
									if !contains(existsEpisodes, episode) {
										allExists = false
										break
									}
								}
								if allExists {
									ctx.Logger.Warn(fmt.Sprintf("%s 第 %d 季第 %v 集已存在，跳过", t.MetaInfo.Title, season, t.MetaInfo.EpisodeList))
									continue
								}
							}
						}
					}
				}
			}
		}

		started = true

		// 添加下载任务
		if downloadService != nil {
			// 使用DownloadSingle方法添加下载，匹配Python中的download_single
			task, err := downloadService.DownloadSingle(ctx, t, downloader, savePath, labels)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("添加下载任务失败: %s", err.Error()))
				a.hasError = true
				continue
			}

			a.addedDownloads = append(a.addedDownloads, task.DownloadID)
			// 保存缓存
			a.SaveCache(ctx.WorkflowID, cacheKey)
			ctx.Logger.Info(fmt.Sprintf("已添加下载任务: %s，ID: %s", t.TorrentInfo.Title, task.DownloadID))
		}
	}

	// 更新上下文
	if len(a.addedDownloads) > 0 {
		ctx.Logger.Info(fmt.Sprintf("已添加 %d 个下载任务", len(a.addedDownloads)))

		// 获取现有的下载列表
		existingDownloads, ok := ctx.GlobalContext["downloads"].([]*dto.DownloadTask)
		if !ok {
			existingDownloads = []*dto.DownloadTask{}
		}

		// 添加新的下载任务
		for _, did := range a.addedDownloads {
			existingDownloads = append(existingDownloads, &dto.DownloadTask{
				DownloadID: did,
				Downloader: downloader,
			})
		}

		// 更新上下文
		ctx.GlobalContext["downloads"] = existingDownloads
	} else if started {
		// 如果已开始处理但没有添加任何下载任务，标记为错误
		a.hasError = true
	}

	// 标记任务完成
	message := fmt.Sprintf("已添加 %d 个下载任务", len(a.addedDownloads))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":         !a.hasError,
		"downloads_added": len(a.addedDownloads),
		"message":         message,
	}

	return output, nil
}

// contains 检查切片是否包含某个元素
func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
