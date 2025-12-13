package models

// StorageConfig 存储配置
type StorageConfig struct {
	// 对rclone进行快照对比时，是否检查文件夹的修改时间
	RcloneSnapshotCheckFolderModtime bool `mapstructure:"RCLONE_SNAPSHOT_CHECK_FOLDER_MODTIME" default:"true"`
	
	// 对OpenList进行快照对比时，是否检查文件夹的修改时间
	OpenListSnapshotCheckFolderModtime bool `mapstructure:"OPENLIST_SNAPSHOT_CHECK_FOLDER_MODTIME" default:"true"`
}