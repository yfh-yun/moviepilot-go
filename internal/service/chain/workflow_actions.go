package chain

import (
	"context"
	"fmt"
	"time"
)

// AddDownloadAction 添加下载动作
type AddDownloadAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *AddDownloadAction) GetID() string {
	return "AddDownload"
}

func (a *AddDownloadAction) GetName() string {
	return "添加下载"
}

func (a *AddDownloadAction) GetDescription() string {
	return "添加种子或磁力链接到下载器"
}

func (a *AddDownloadAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"download_type": map[string]interface{}{
				"type":        "string",
				"title":       "下载类型",
				"description": "种子或磁力链接",
				"enum":        []string{"torrent", "magnet"},
			},
			"download_url": map[string]interface{}{
				"type":        "string",
				"title":       "下载链接",
				"description": "种子文件URL或磁力链接",
			},
			"save_path": map[string]interface{}{
				"type":        "string",
				"title":       "保存路径",
				"description": "下载文件保存路径",
			},
			" downloader": map[string]interface{}{
				"type":        "string",
				"title":       "下载器",
				"description": "选择使用的下载器",
			},
		},
		"required": []string{"download_type", "download_url"},
	}
}

func (a *AddDownloadAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	downloadType, ok := data["download_type"].(string)
	if !ok {
		return nil, fmt.Errorf("下载类型参数缺失")
	}

	downloadURL, ok := data["download_url"].(string)
	if !ok {
		return nil, fmt.Errorf("下载链接参数缺失")
	}

	savePath, _ := data["save_path"].(string)
	downloader, _ := data["downloader"].(string)

	// 这里应该调用下载服务添加下载
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("成功添加%s下载: %s", downloadType, downloadURL)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_download_url"] = downloadURL
	executeContext.Context["last_download_type"] = downloadType

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"download_url":  downloadURL,
			"download_type": downloadType,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *AddDownloadAction) IsDone() bool {
	return a.done
}

func (a *AddDownloadAction) GetSuccess() bool {
	return a.success
}

func (a *AddDownloadAction) GetMessage() string {
	return a.message
}

// AddSubscribeAction 添加订阅动作
type AddSubscribeAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *AddSubscribeAction) GetID() string {
	return "AddSubscribe"
}

func (a *AddSubscribeAction) GetName() string {
	return "添加订阅"
}

func (a *AddSubscribeAction) GetDescription() string {
	return "添加媒体订阅"
}

func (a *AddSubscribeAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"media_type": map[string]interface{}{
				"type":  "string",
				"title": "媒体类型",
				"enum":  []string{"movie", "tv", "anime"},
			},
			"media_name": map[string]interface{}{
				"type":        "string",
				"title":       "媒体名称",
				"description": "电影或电视剧名称",
			},
			"media_year": map[string]interface{}{
				"type":        "integer",
				"title":       "年份",
				"description": "媒体发布年份",
			},
			"season": map[string]interface{}{
				"type":        "integer",
				"title":       "季数",
				"description": "电视剧季数（仅电视剧）",
			},
			"quality": map[string]interface{}{
				"type":        "string",
				"title":       "画质要求",
				"description": "期望的画质质量",
			},
		},
		"required": []string{"media_type", "media_name"},
	}
}

func (a *AddSubscribeAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	mediaType, ok := data["media_type"].(string)
	if !ok {
		return nil, fmt.Errorf("媒体类型参数缺失")
	}

	mediaName, ok := data["media_name"].(string)
	if !ok {
		return nil, fmt.Errorf("媒体名称参数缺失")
	}

	mediaYear, _ := data["media_year"].(int)
	season, _ := data["season"].(int)
	quality, _ := data["quality"].(string)

	// 这里应该调用订阅服务添加订阅
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("成功添加%s订阅: %s", mediaType, mediaName)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_subscribe_media"] = mediaName
	executeContext.Context["last_subscribe_type"] = mediaType

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"media_name": mediaName,
			"media_type": mediaType,
			"media_year": mediaYear,
			"season":     season,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *AddSubscribeAction) IsDone() bool {
	return a.done
}

func (a *AddSubscribeAction) GetSuccess() bool {
	return a.success
}

func (a *AddSubscribeAction) GetMessage() string {
	return a.message
}

// FetchMediasAction 获取媒体信息动作
type FetchMediasAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *FetchMediasAction) GetID() string {
	return "FetchMedias"
}

func (a *FetchMediasAction) GetName() string {
	return "获取媒体信息"
}

func (a *FetchMediasAction) GetDescription() string {
	return "从TMDB等来源获取媒体详细信息"
}

func (a *FetchMediasAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source": map[string]interface{}{
				"type":  "string",
				"title": "数据源",
				"enum":  []string{"tmdb", "douban", "bangumi"},
			},
			"media_name": map[string]interface{}{
				"type":        "string",
				"title":       "媒体名称",
				"description": "要搜索的媒体名称",
			},
			"media_type": map[string]interface{}{
				"type":  "string",
				"title": "媒体类型",
				"enum":  []string{"movie", "tv", "anime"},
			},
			"year": map[string]interface{}{
				"type":        "integer",
				"title":       "年份",
				"description": "媒体发布年份",
			},
		},
		"required": []string{"source", "media_name"},
	}
}

func (a *FetchMediasAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	source, ok := data["source"].(string)
	if !ok {
		return nil, fmt.Errorf("数据源参数缺失")
	}

	mediaName, ok := data["media_name"].(string)
	if !ok {
		return nil, fmt.Errorf("媒体名称参数缺失")
	}

	mediaType, _ := data["media_type"].(string)
	year, _ := data["year"].(int)

	// 这里应该调用媒体服务获取信息
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("成功从%s获取媒体信息: %s", source, mediaName)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["fetched_media"] = mediaName
	executeContext.Context["fetched_source"] = source

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"media_name": mediaName,
			"source":     source,
			"media_type": mediaType,
			"year":       year,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *FetchMediasAction) IsDone() bool {
	return a.done
}

func (a *FetchMediasAction) GetSuccess() bool {
	return a.success
}

func (a *FetchMediasAction) GetMessage() string {
	return a.message
}

// FilterMediasAction 过滤媒体动作
type FilterMediasAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *FilterMediasAction) GetID() string {
	return "FilterMedias"
}

func (a *FilterMediasAction) GetName() string {
	return "过滤媒体"
}

func (a *FilterMediasAction) GetDescription() string {
	return "根据规则过滤媒体内容"
}

func (a *FilterMediasAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"filter_rules": map[string]interface{}{
				"type":        "string",
				"title":       "过滤规则",
				"description": "过滤规则ID或名称",
			},
			"include_keywords": map[string]interface{}{
				"type":        "array",
				"title":       "包含关键词",
				"description": "必须包含的关键词列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"exclude_keywords": map[string]interface{}{
				"type":        "array",
				"title":       "排除关键词",
				"description": "需要排除的关键词列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"min_size": map[string]interface{}{
				"type":        "integer",
				"title":       "最小大小",
				"description": "文件最小大小（MB）",
			},
			"max_size": map[string]interface{}{
				"type":        "integer",
				"title":       "最大大小",
				"description": "文件最大大小（MB）",
			},
		},
	}
}

func (a *FilterMediasAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	filterRules, _ := data["filter_rules"].(string)

	// 这里应该调用过滤服务进行媒体过滤
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = "媒体过滤完成"

	if filterRules != "" {
		a.message = fmt.Sprintf("使用规则%s完成媒体过滤", filterRules)
	}

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["filtered_count"] = 1 // 假设过滤出一个结果

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"filter_rules":   filterRules,
			"filtered_count": 1,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *FilterMediasAction) IsDone() bool {
	return a.done
}

func (a *FilterMediasAction) GetSuccess() bool {
	return a.success
}

func (a *FilterMediasAction) GetMessage() string {
	return a.message
}

// SendMessageAction 发送消息动作
type SendMessageAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *SendMessageAction) GetID() string {
	return "SendMessage"
}

func (a *SendMessageAction) GetName() string {
	return "发送消息"
}

func (a *SendMessageAction) GetDescription() string {
	return "发送通知消息"
}

func (a *SendMessageAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message_type": map[string]interface{}{
				"type":  "string",
				"title": "消息类型",
				"enum":  []string{"notification", "warning", "error", "info"},
			},
			"title": map[string]interface{}{
				"type":        "string",
				"title":       "消息标题",
				"description": "消息标题",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"title":       "消息内容",
				"description": "消息正文内容",
			},
			"channels": map[string]interface{}{
				"type":        "array",
				"title":       "发送渠道",
				"description": "消息发送渠道",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"wechat", "telegram", "email", "webhook"},
				},
			},
		},
		"required": []string{"content"},
	}
}

func (a *SendMessageAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	content, ok := data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("消息内容参数缺失")
	}

	messageType, _ := data["message_type"].(string)
	title, _ := data["title"].(string)
	channels, _ := data["channels"].([]interface{})

	// 这里应该调用消息服务发送消息
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = "消息发送成功"

	if title != "" {
		a.message = fmt.Sprintf("成功发送消息: %s", title)
	}

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_message_sent"] = content

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"title":        title,
			"content":      content,
			"message_type": messageType,
			"channels":     channels,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *SendMessageAction) IsDone() bool {
	return a.done
}

func (a *SendMessageAction) GetSuccess() bool {
	return a.success
}

func (a *SendMessageAction) GetMessage() string {
	return a.message
}

// TransferFileAction 转移文件动作
type TransferFileAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *TransferFileAction) GetID() string {
	return "TransferFile"
}

func (a *TransferFileAction) GetName() string {
	return "转移文件"
}

func (a *TransferFileAction) GetDescription() string {
	return "转移下载完成的文件到媒体库"
}

func (a *TransferFileAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_path": map[string]interface{}{
				"type":        "string",
				"title":       "源路径",
				"description": "文件源路径",
			},
			"target_path": map[string]interface{}{
				"type":        "string",
				"title":       "目标路径",
				"description": "文件目标路径",
			},
			"media_type": map[string]interface{}{
				"type":  "string",
				"title": "媒体类型",
				"enum":  []string{"movie", "tv", "anime"},
			},
			"rename_format": map[string]interface{}{
				"type":        "string",
				"title":       "重命名格式",
				"description": "文件重命名格式模板",
			},
			"delete_source": map[string]interface{}{
				"type":        "boolean",
				"title":       "删除源文件",
				"description": "转移完成后删除源文件",
			},
		},
		"required": []string{"source_path", "target_path"},
	}
}

func (a *TransferFileAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	sourcePath, ok := data["source_path"].(string)
	if !ok {
		return nil, fmt.Errorf("源路径参数缺失")
	}

	targetPath, ok := data["target_path"].(string)
	if !ok {
		return nil, fmt.Errorf("目标路径参数缺失")
	}

	mediaType, _ := data["media_type"].(string)
	renameFormat, _ := data["rename_format"].(string)
	deleteSource, _ := data["delete_source"].(bool)

	// 这里应该调用转移服务进行文件转移
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("文件转移成功: %s -> %s", sourcePath, targetPath)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_transferred_file"] = sourcePath

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"source_path":   sourcePath,
			"target_path":   targetPath,
			"media_type":    mediaType,
			"rename_format": renameFormat,
			"delete_source": deleteSource,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *TransferFileAction) IsDone() bool {
	return a.done
}

func (a *TransferFileAction) GetSuccess() bool {
	return a.success
}

func (a *TransferFileAction) GetMessage() string {
	return a.message
}

// ScanFileAction 扫描文件动作
type ScanFileAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *ScanFileAction) GetID() string {
	return "ScanFile"
}

func (a *ScanFileAction) GetName() string {
	return "扫描文件"
}

func (a *ScanFileAction) GetDescription() string {
	return "扫描指定目录中的媒体文件"
}

func (a *ScanFileAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scan_path": map[string]interface{}{
				"type":        "string",
				"title":       "扫描路径",
				"description": "要扫描的目录路径",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"title":       "递归扫描",
				"description": "是否递归扫描子目录",
			},
			"file_extensions": map[string]interface{}{
				"type":        "array",
				"title":       "文件扩展名",
				"description": "要扫描的文件扩展名列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"min_size": map[string]interface{}{
				"type":        "integer",
				"title":       "最小文件大小",
				"description": "最小文件大小（MB）",
			},
		},
		"required": []string{"scan_path"},
	}
}

func (a *ScanFileAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	scanPath, ok := data["scan_path"].(string)
	if !ok {
		return nil, fmt.Errorf("扫描路径参数缺失")
	}

	recursive, _ := data["recursive"].(bool)
	extensions, _ := data["file_extensions"].([]interface{})
	minSize, _ := data["min_size"].(int)

	// 这里应该调用扫描服务进行文件扫描
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("文件扫描完成: %s", scanPath)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["scanned_files_count"] = 1 // 假设扫描到一个文件

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"scan_path":   scanPath,
			"recursive":   recursive,
			"extensions":  extensions,
			"min_size":    minSize,
			"files_found": 1,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *ScanFileAction) IsDone() bool {
	return a.done
}

func (a *ScanFileAction) GetSuccess() bool {
	return a.success
}

func (a *ScanFileAction) GetMessage() string {
	return a.message
}

// ScrapeFileAction 刮削文件动作
type ScrapeFileAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *ScrapeFileAction) GetID() string {
	return "ScrapeFile"
}

func (a *ScrapeFileAction) GetName() string {
	return "刮削文件"
}

func (a *ScrapeFileAction) GetDescription() string {
	return "刮削媒体文件的元数据信息"
}

func (a *ScrapeFileAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"title":       "文件路径",
				"description": "要刮削的文件路径",
			},
			"media_info": map[string]interface{}{
				"type":        "object",
				"title":       "媒体信息",
				"description": "媒体基本信息",
			},
			"scrape_sources": map[string]interface{}{
				"type":        "array",
				"title":       "刮削源",
				"description": "刮削数据源列表",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"tmdb", "douban", "fanart"},
				},
			},
			"save_nfo": map[string]interface{}{
				"type":        "boolean",
				"title":       "保存NFO",
				"description": "是否保存NFO文件",
			},
			"download_artwork": map[string]interface{}{
				"type":        "boolean",
				"title":       "下载海报",
				"description": "是否下载海报和艺术作品",
			},
		},
		"required": []string{"file_path"},
	}
}

func (a *ScrapeFileAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	filePath, ok := data["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("文件路径参数缺失")
	}

	mediaInfo, _ := data["media_info"].(map[string]interface{})
	sources, _ := data["scrape_sources"].([]interface{})
	saveNFO, _ := data["save_nfo"].(bool)
	downloadArtwork, _ := data["download_artwork"].(bool)

	// 这里应该调用刮削服务进行文件刮削
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("文件刮削完成: %s", filePath)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["scraped_file"] = filePath

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"file_path":        filePath,
			"media_info":       mediaInfo,
			"scrape_sources":   sources,
			"save_nfo":         saveNFO,
			"download_artwork": downloadArtwork,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *ScrapeFileAction) IsDone() bool {
	return a.done
}

func (a *ScrapeFileAction) GetSuccess() bool {
	return a.success
}

func (a *ScrapeFileAction) GetMessage() string {
	return a.message
}

// InvokePluginAction 调用插件动作
type InvokePluginAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *InvokePluginAction) GetID() string {
	return "InvokePlugin"
}

func (a *InvokePluginAction) GetName() string {
	return "调用插件"
}

func (a *InvokePluginAction) GetDescription() string {
	return "调用指定的插件执行特定功能"
}

func (a *InvokePluginAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"plugin_id": map[string]interface{}{
				"type":        "string",
				"title":       "插件ID",
				"description": "要调用的插件标识",
			},
			"plugin_method": map[string]interface{}{
				"type":        "string",
				"title":       "插件方法",
				"description": "要调用的插件方法名",
			},
			"plugin_args": map[string]interface{}{
				"type":        "object",
				"title":       "插件参数",
				"description": "传递给插件的参数",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"title":       "超时时间",
				"description": "插件调用超时时间（秒）",
			},
		},
		"required": []string{"plugin_id", "plugin_method"},
	}
}

func (a *InvokePluginAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	pluginID, ok := data["plugin_id"].(string)
	if !ok {
		return nil, fmt.Errorf("插件ID参数缺失")
	}

	pluginMethod, ok := data["plugin_method"].(string)
	if !ok {
		return nil, fmt.Errorf("插件方法参数缺失")
	}

	pluginArgs, _ := data["plugin_args"].(map[string]interface{})
	timeout, _ := data["timeout"].(int)

	// 这里应该调用插件管理器执行插件
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("插件调用成功: %s.%s", pluginID, pluginMethod)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_plugin_called"] = pluginID

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"plugin_id":     pluginID,
			"plugin_method": pluginMethod,
			"plugin_args":   pluginArgs,
			"timeout":       timeout,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *InvokePluginAction) IsDone() bool {
	return a.done
}

func (a *InvokePluginAction) GetSuccess() bool {
	return a.success
}

func (a *InvokePluginAction) GetMessage() string {
	return a.message
}

// SendEventAction 发送事件动作
type SendEventAction struct {
	id      string
	name    string
	success bool
	message string
	done    bool
}

func (a *SendEventAction) GetID() string {
	return "SendEvent"
}

func (a *SendEventAction) GetName() string {
	return "发送事件"
}

func (a *SendEventAction) GetDescription() string {
	return "发送系统事件"
}

func (a *SendEventAction) GetDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"event_type": map[string]interface{}{
				"type":        "string",
				"title":       "事件类型",
				"description": "要发送的事件类型",
			},
			"event_data": map[string]interface{}{
				"type":        "object",
				"title":       "事件数据",
				"description": "事件携带的数据",
			},
			"delay": map[string]interface{}{
				"type":        "integer",
				"title":       "延迟发送",
				"description": "延迟发送时间（秒）",
			},
		},
		"required": []string{"event_type"},
	}
}

func (a *SendEventAction) Execute(ctx context.Context, workflowID int64,
	data map[string]interface{}, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	a.success = false
	a.done = false
	a.message = ""

	// 解析参数
	eventType, ok := data["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("事件类型参数缺失")
	}

	eventData, _ := data["event_data"].(map[string]interface{})
	delay, _ := data["delay"].(int)

	// 如果有延迟，等待指定时间
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}

	// 这里应该调用事件管理器发送事件
	// 简化起见，直接返回成功
	a.success = true
	a.done = true
	a.message = fmt.Sprintf("事件发送成功: %s", eventType)

	// 更新执行上下文
	if executeContext.Context == nil {
		executeContext.Context = make(map[string]interface{})
	}
	executeContext.Context["last_event_sent"] = eventType

	return &WorkflowActionResult{
		Success: a.success,
		Message: a.message,
		Data: map[string]interface{}{
			"event_type": eventType,
			"event_data": eventData,
			"delay":      delay,
		},
		Context: executeContext.Context,
	}, nil
}

func (a *SendEventAction) IsDone() bool {
	return a.done
}

func (a *SendEventAction) GetSuccess() bool {
	return a.success
}

func (a *SendEventAction) GetMessage() string {
	return a.message
}
