package command

// CommandCategory 命令分类
type CommandCategory string

const (
	CategorySite      CommandCategory = "站点"
	CategorySubscribe CommandCategory = "订阅"
	CategoryManage    CommandCategory = "管理"
)

// CommandType 命令类型
type CommandType string

const (
	TypeScheduler CommandType = "scheduler"
	TypeCommand   CommandType = "command"
)
