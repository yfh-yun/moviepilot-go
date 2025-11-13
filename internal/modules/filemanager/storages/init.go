package storages

// Storage 模块初始化文�?
// StorageFactories 存储工厂映射
var StorageFactories = make(map[string]func() StorageBase)

// RegisterStorage 注册存储类型
func RegisterStorage(name string, factory func() StorageBase) {
	StorageFactories[name] = factory
}

// CreateStorage 创建存储实例
func CreateStorage(name string) StorageBase {
	factory, exists := StorageFactories[name]
	if !exists {
		return nil
	}
	return factory()
}

// RegisterAliPan 注册阿里云盘存储
func RegisterAliPan() {
	RegisterStorage("alipan", func() StorageBase {
		return NewAliPan()
	})
}

// RegisterAlist 注册AList存储
func RegisterAlist() {
	RegisterStorage("alist", func() StorageBase {
		return NewAlist()
	})
}

// RegisterLocalStorage 注册本地存储
func RegisterLocalStorage() {
	RegisterStorage("local", func() StorageBase {
		return NewLocalStorage()
	})
}

// RegisterRclone 注册rclone存储
func RegisterRclone() {
	RegisterStorage("rclone", func() StorageBase {
		return NewRclone()
	})
}

// RegisterSMB 注册SMB存储
func RegisterSMB() {
	RegisterStorage("smb", func() StorageBase {
		return NewSMB()
	})
}
