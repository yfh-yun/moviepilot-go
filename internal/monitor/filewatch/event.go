package filewatch

// Operation 文件操作类型
type Operation int

const (
	// Create 创建文件
	Create Operation = iota
	// Write 写入文件
	Write
	// Remove 删除文件
	Remove
	// Rename 重命名文件
	Rename
)

// String 返回操作类型的字符串表示
func (op Operation) String() string {
	switch op {
	case Create:
		return "CREATE"
	case Write:
		return "WRITE"
	case Remove:
		return "REMOVE"
	case Rename:
		return "RENAME"
	default:
		return "UNKNOWN"
	}
}

// Event 文件系统事件
type Event struct {
	Path string
	Op   Operation
}

// EventHandler 事件处理函数类型
type EventHandler func(event Event)
