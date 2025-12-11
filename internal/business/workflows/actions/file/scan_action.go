package file

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// ScanAction 实现文件扫描动作
type ScanAction struct {
	*common.BaseAction

	fileItems []map[string]any // 扫描到的文件列表
	hasError  bool             // 是否有错误
}

// NewScanAction 创建新的文件扫描动作实例
func NewScanAction() base.Action {
	return &ScanAction{
		BaseAction: common.NewBaseAction("scan", base.ActionTypeFile),
		fileItems:  []map[string]any{},
		hasError:   false,
	}
}

// GetName 获取动作名称
func (a *ScanAction) GetName() string {
	return "扫描目录"
}

// GetDescription 获取动作描述
func (a *ScanAction) GetDescription() string {
	return "扫描目录文件到队列"
}

// GetData 获取动作参数模板
func (a *ScanAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的ScanFileParams
	return map[string]any{
		"storage": map[string]any{
			"type":        "string",
			"description": "存储",
			"default":     "local",
		},
		"directory": map[string]any{
			"type":        "string",
			"description": "目录",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *ScanAction) Success() bool {
	// 动作是否成功，对应Python中的success属性
	return !a.hasError
}

// execute 执行文件扫描动作（核心逻辑）
func (a *ScanAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	storage, _ := ctx.Input["storage"].(string)
	directory, _ := ctx.Input["directory"].(string)

	// 检查storage和directory是否为空
	if storage == "" || directory == "" {
		return map[string]any{"success": false, "message": "存储和目录不能为空"}, nil
	}

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// storageService, _ := ctx.Services["StorageService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("扫描目录失败: %v", r))
			a.hasError = true
		}
	}()

	// 初始化文件列表和错误状态
	a.fileItems = []map[string]any{}
	a.hasError = false

	// 获取文件项
	// TODO: 实现获取文件项的逻辑
	// fileitem = storagechain.get_file_item(params.storage, Path(params.directory))
	ctx.Logger.Info(fmt.Sprintf("获取目录: 【%s】%s", storage, directory))

	// 列出文件
	// TODO: 实现列出文件的逻辑
	// files = storagechain.list_files(fileitem, recursion=True)
	ctx.Logger.Info(fmt.Sprintf("扫描目录: 【%s】%s", storage, directory))

	// 遍历处理每个文件
	// TODO: 实现遍历处理文件的逻辑
	// for file in files:
	//     if global_vars.is_workflow_stopped(workflow_id):
	//         break
	//     if not file.extension or f".{file.extension.lower()}" not in settings.RMT_MEDIAEXT:
	//         continue
	//     // 添加文件到队列，而不是目录
	//     self._fileitems.append(file)

	// 更新上下文
	if len(a.fileItems) > 0 {
		// 获取现有fileitems
		fileitems, ok := ctx.GlobalContext["fileitems"].([]map[string]any)
		if !ok {
			fileitems = []map[string]any{}
		}
		// 合并文件列表
		fileitems = append(fileitems, a.fileItems...)
		// 更新上下文
		ctx.GlobalContext["fileitems"] = fileitems
	}

	// 标记任务完成
	message := fmt.Sprintf("扫描到 %d 个文件", len(a.fileItems))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":       !a.hasError,
		"storage":       storage,
		"directory":     directory,
		"files_scanned": len(a.fileItems),
		"message":       message,
	}

	return output, nil
}
