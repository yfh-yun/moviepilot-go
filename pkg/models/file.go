package models

import (
	"time"
)

// FileItem 文件�?type FileItem struct {
	// 存储类型
	Storage string `json:"storage,omitempty"`
	
	// 类型 dir/file
	Type string `json:"type,omitempty"`
	
	// 文件路径
	Path string `json:"path,omitempty"`
	
	// 文件�?	Name string `json:"name,omitempty"`
	
	// 文件�?	Basename string `json:"basename,omitempty"`
	
	// 文件后缀
	Extension string `json:"extension,omitempty"`
	
	// 文件大小
	Size int64 `json:"size,omitempty"`
	
	// 修改时间
	ModifyTime time.Time `json:"modify_time,omitempty"`
	
	// 子节�?	Children []FileItem `json:"children,omitempty"`
	
	// ID
	Fileid string `json:"fileid,omitempty"`
	
	// 父ID
	ParentFileid string `json:"parent_fileid,omitempty"`
	
	// 缩略�?	Thumbnail string `json:"thumbnail,omitempty"`
	
	// 115 pickcode
	Pickcode string `json:"pickcode,omitempty"`
	
	// drive_id
	DriveID string `json:"drive_id,omitempty"`
	
	// url
	URL string `json:"url,omitempty"`
}

// StorageUsage 存储使用情况
type StorageUsage struct {
	// 总空�?	Total float64 `json:"total,omitempty"`
	
	// 剩余空间
	Available float64 `json:"available,omitempty"`
}

// StorageTransType 存储传输类型
type StorageTransType struct {
	// 传输类型
	Transtype map[string]interface{} `json:"transtype,omitempty"`
}

// NewFileItem 创建一个新�?FileItem 实例，带有默认�?func NewFileItem() *FileItem {
	return &FileItem{
		Storage:  "local",
		Path:     "/",
		Children: make([]FileItem, 0),
	}
}

// NewStorageUsage 创建一个新�?StorageUsage 实例
func NewStorageUsage() *StorageUsage {
	return &StorageUsage{
		Total:     0.0,
		Available: 0.0,
	}
}

// NewStorageTransType 创建一个新�?StorageTransType 实例
func NewStorageTransType() *StorageTransType {
	return &StorageTransType{
		Transtype: make(map[string]interface{}),
	}
}
