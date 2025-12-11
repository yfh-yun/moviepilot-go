package enums

// StorageSchema 支持的存储类型
type StorageSchema string

const (
	// 存储类型
	StorageSchemaLocal  StorageSchema = "local"
	StorageSchemaAlipan StorageSchema = "alipan"
	StorageSchemaU115   StorageSchema = "u115"
	StorageSchemaRclone StorageSchema = "rclone"
	StorageSchemaAlist  StorageSchema = "alist"
	StorageSchemaSMB    StorageSchema = "smb"
)
