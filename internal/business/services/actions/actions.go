// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"

	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/business/services/actions/types"
)

// AddDownloadAction 添加下载动作
// 实现Python版本AddDownloadAction的完整功能
type AddDownloadAction struct {
	*BaseActionImpl
	downloadManager *DownloadManager
}

// NewAddDownloadAction 创建添加下载动作
func NewAddDownloadAction(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *AddDownloadAction {
	action := &AddDownloadAction{
		BaseActionImpl:  NewBaseAction("add_download", cache),
		downloadManager: NewDownloadManager(downloadRepo, mediaRepo, cache),
	}
	return action
}

// Name 获取动作名称
func (a *AddDownloadAction) Name() string {
	return "添加下载"
}

// Description 获取动作描述
func (a *AddDownloadAction) Description() string {
	return "根据资源列表添加下载任务"
}

// Data 获取动作数据
func (a *AddDownloadAction) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": a.GetActionID(),
		"name":      a.Name(),
		"desc":      a.Description(),
		"params": map[string]interface{}{
			"downloader": "",
			"save_path":  "",
			"labels":     []string{},
			"only_lack":  false,
			"sites":      []int{},
			"quality":    "",
			"resolution": "",
		},
	}
}

// doExecute 执行动作的具体逻辑
func (a *AddDownloadAction) doExecute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionContext *types.ActionContext,
) (*types.ActionContext, error) {
	// 解析参数
	downloadParams := &AddDownloadParams{
		Downloader: getStringParam(params, "downloader", ""),
		SavePath:   getStringParam(params, "save_path", ""),
		Labels:     getStringSliceParam(params, "labels", []string{}),
		OnlyLack:   getBoolParam(params, "only_lack", false),
		Sites:      getIntSliceParam(params, "sites", []int{}),
		Quality:    getStringParam(params, "quality", ""),
		Resolution: getStringParam(params, "resolution", ""),
	}

	// 执行下载添加
	results, err := a.downloadManager.AddDownload(ctx, workflowID, downloadParams, actionContext.Torrents)
	if err != nil {
		a.SetError(fmt.Sprintf("添加下载失败: %v", err))
		return actionContext, err
	}

	// 统计结果
	successCount := 0
	var downloadTasks []*types.Download
	for _, result := range results {
		if result.Success {
			successCount++
			// 创建下载任务对象
			downloadTask := &types.Download{
				ID:         result.DownloadID,
				Downloader: downloadParams.Downloader,
				Status:     "pending",
			}
			downloadTasks = append(downloadTasks, downloadTask)
		}
	}

	// 更新上下文
	actionContext.Downloads = append(actionContext.Downloads, downloadTasks...)

	// 设置完成状态
	if successCount > 0 {
		a.SetDone(fmt.Sprintf("已添加 %d 个下载任务", successCount))
	} else {
		a.SetError("没有成功添加任何下载任务")
	}

	return actionContext, nil
}

// FetchMediasAction 获取媒体数据动作
// 实现Python版本FetchMediasAction的完整功能
type FetchMediasAction struct {
	*BaseActionImpl
	mediaFetcher *MediaFetcher
}

// NewFetchMediasAction 创建获取媒体数据动作
func NewFetchMediasAction(cache *WorkflowCache) *FetchMediasAction {
	action := &FetchMediasAction{
		BaseActionImpl: NewBaseAction("fetch_medias", cache),
		mediaFetcher:   NewMediaFetcher(cache),
	}
	return action
}

// Name 获取动作名称
func (a *FetchMediasAction) Name() string {
	return "获取媒体数据"
}

// Description 获取动作描述
func (a *FetchMediasAction) Description() string {
	return "从各种数据源获取媒体信息"
}

// Data 获取动作数据
func (a *FetchMediasAction) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": a.GetActionID(),
		"name":      a.Name(),
		"desc":      a.Description(),
		"params": map[string]interface{}{
			"source_type": "ranking",
			"sources":     []string{},
			"api_path":    "",
			"limit":       0,
			"genres":      []string{},
			"year":        0,
			"rating":      0.0,
			"countries":   []string{},
			"languages":   []string{},
			"sort_by":     "",
			"order_by":    "",
		},
	}
}

// doExecute 执行动作的具体逻辑
func (a *FetchMediasAction) doExecute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionContext *types.ActionContext,
) (*types.ActionContext, error) {
	// 解析参数
	fetchParams := &FetchMediasParams{
		SourceType: getStringParam(params, "source_type", "ranking"),
		Sources:    getStringSliceParam(params, "sources", []string{}),
		APIPath:    getStringParam(params, "api_path", ""),
		Limit:      getIntParam(params, "limit", 0),
		Genres:     getStringSliceParam(params, "genres", []string{}),
		Year:       getIntParam(params, "year", 0),
		Rating:     getFloat64Param(params, "rating", 0.0),
		Countries:  getStringSliceParam(params, "countries", []string{}),
		Languages:  getStringSliceParam(params, "languages", []string{}),
		SortBy:     getStringParam(params, "sort_by", ""),
		OrderBy:    getStringParam(params, "order_by", ""),
	}

	// 执行媒体获取
	results, err := a.mediaFetcher.FetchMedias(ctx, workflowID, fetchParams)
	if err != nil {
		a.SetError(fmt.Sprintf("获取媒体数据失败: %v", err))
		return actionContext, err
	}

	// 收集所有媒体数据
	var allMedias []*types.MediaInfo
	totalCount := 0
	successCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
			allMedias = append(allMedias, result.Medias...)
			totalCount += len(result.Medias)
		}
	}

	// 更新上下文
	actionContext.Medias = append(actionContext.Medias, allMedias...)

	// 设置完成状态
	if totalCount > 0 {
		a.SetDone(fmt.Sprintf("已获取 %d 个媒体数据，来源 %d 个", totalCount, successCount))
	} else {
		a.SetError("没有获取到任何媒体数据")
	}

	return actionContext, nil
}

// SendMessageAction 发送消息动作
// 实现Python版本SendMessageAction的完整功能
type SendMessageAction struct {
	*BaseActionImpl
	messageSender *MessageSender
}

// NewSendMessageAction 创建发送消息动作
func NewSendMessageAction(
	messageRepo interfaces.MessageRepository,
	userRepo interfaces.UserRepository,
	cache *WorkflowCache,
) *SendMessageAction {
	action := &SendMessageAction{
		BaseActionImpl: NewBaseAction("send_message", cache),
		messageSender:  NewMessageSender(messageRepo, userRepo, cache),
	}
	return action
}

// Name 获取动作名称
func (a *SendMessageAction) Name() string {
	return "发送消息"
}

// Description 获取动作描述
func (a *SendMessageAction) Description() string {
	return "发送任务执行消息"
}

// Data 获取动作数据
func (a *SendMessageAction) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": a.GetActionID(),
		"name":      a.Name(),
		"desc":      a.Description(),
		"params": map[string]interface{}{
			"clients":  []string{},
			"userid":   "",
			"title":    "",
			"content":  "",
			"type":     "",
			"template": "",
			"async":    false,
			"priority": 5,
		},
	}
}

// doExecute 执行动作的具体逻辑
func (a *SendMessageAction) doExecute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionContext *types.ActionContext,
) (*types.ActionContext, error) {
	// 解析参数
	sendParams := &SendMessageParams{
		Clients:  getStringSliceParam(params, "clients", []string{}),
		UserID:   getStringParam(params, "userid", ""),
		Title:    getStringParam(params, "title", ""),
		Content:  getStringParam(params, "content", ""),
		Type:     getStringParam(params, "type", ""),
		Template: getStringParam(params, "template", ""),
		Async:    getBoolParam(params, "async", false),
		Priority: getIntParam(params, "priority", 5),
	}

	// 如果没有指定标题和内容，使用默认的进度消息
	if sendParams.Title == "" && sendParams.Content == "" {
		sendParams.Title = "工作流进度"
		sendParams.Content = fmt.Sprintf("当前进度：%d%%", actionContext.Progress)
		sendParams.Type = "info"
	}

	// 执行消息发送
	results, err := a.messageSender.SendMessage(ctx, workflowID, sendParams, actionContext.Messages)
	if err != nil {
		a.SetError(fmt.Sprintf("发送消息失败: %v", err))
		return actionContext, err
	}

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	// 设置完成状态
	if successCount > 0 {
		a.SetDone(fmt.Sprintf("已发送 %d 条消息", successCount))
	} else {
		a.SetError("没有成功发送任何消息")
	}

	return actionContext, nil
}

// AddSubscribeAction 添加订阅动作
// 基于现有的SubscribeManager实现
type AddSubscribeAction struct {
	*BaseActionImpl
	subscribeManager *SubscribeManager
}

// NewAddSubscribeAction 创建添加订阅动作
func NewAddSubscribeAction(
	subscribeRepo interfaces.SubscribeRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *AddSubscribeAction {
	action := &AddSubscribeAction{
		BaseActionImpl:   NewBaseAction("add_subscribe", cache),
		subscribeManager: NewSubscribeManager(subscribeRepo, mediaRepo, cache),
	}
	return action
}

// Name 获取动作名称
func (a *AddSubscribeAction) Name() string {
	return "添加订阅"
}

// Description 获取动作描述
func (a *AddSubscribeAction) Description() string {
	return "根据媒体列表添加订阅"
}

// Data 获取动作数据
func (a *AddSubscribeAction) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": a.GetActionID(),
		"name":      a.Name(),
		"desc":      a.Description(),
		"params": map[string]interface{}{
			"username": "",
			"priority": 5,
		},
	}
}

// doExecute 执行动作的具体逻辑
func (a *AddSubscribeAction) doExecute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionContext *types.ActionContext,
) (*types.ActionContext, error) {
	// 解析参数
	username := getStringParam(params, "username", "")
	priority := getIntParam(params, "priority", 5)

	// 执行订阅添加
	results, err := a.subscribeManager.AddSubscribe(ctx, workflowID, username, priority, actionContext.Medias)
	if err != nil {
		a.SetError(fmt.Sprintf("添加订阅失败: %v", err))
		return actionContext, err
	}

	// 统计结果
	successCount := 0
	var subscribes []*types.Subscribe
	for _, result := range results {
		if result.Success {
			successCount++
			if result.Subscribe != nil {
				subscribes = append(subscribes, result.Subscribe)
			}
		}
	}

	// 更新上下文
	actionContext.Subscribes = append(actionContext.Subscribes, subscribes...)

	// 设置完成状态
	if successCount > 0 {
		a.SetDone(fmt.Sprintf("已添加 %d 个订阅", successCount))
	} else {
		a.SetError("没有成功添加任何订阅")
	}

	return actionContext, nil
}

// 参数解析辅助函数

func getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := params[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			var i int
			if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func getFloat64Param(params map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

func getStringSliceParam(params map[string]interface{}, key string, defaultValue []string) []string {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case []string:
			return v
		case []interface{}:
			result := make([]string, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		case string:
			// 如果是逗号分隔的字符串
			if v != "" {
				return []string{v}
			}
		}
	}
	return defaultValue
}

func getIntSliceParam(params map[string]interface{}, key string, defaultValue []int) []int {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case []int:
			return v
		case []interface{}:
			result := make([]int, len(v))
			for i, item := range v {
				switch item := item.(type) {
				case int:
					result[i] = item
				case float64:
					result[i] = int(item)
				}
			}
			return result
		}
	}
	return defaultValue
}
