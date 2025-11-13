package command

// CommandRegisterEventData 命令注册事件数据
type CommandRegisterEventData struct {
	Commands map[string]interface{} `json:"commands"`
	Origin   string                 `json:"origin"`
	Service  interface{}            `json:"service"`
	Cancel   bool                   `json:"cancel"`
	Source   string                 `json:"source"`
}

// CommandExecuteEvent 命令执行事件数据
type CommandExecuteEvent struct {
	Cmd     string      `json:"cmd"`
	Channel string      `json:"channel"`
	Source  string      `json:"source"`
	User    interface{} `json:"user"`
}

// ModuleReloadEvent 模块重载事件数据
type ModuleReloadEvent struct {
	// 可以根据需要添加相关字�?}
