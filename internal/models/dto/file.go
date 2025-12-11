package dto

import "path/filepath"

// FileItem 文件项
type FileItem struct {
	// 存储类型
	Storage string `json:"storage,omitempty"`
	// 类型 dir/file
	Type string `json:"type,omitempty"`
	// 文件路径
	Path string `json:"path,omitempty"`
	// 文件名
	Name string `json:"name,omitempty"`
	// 文件名
	Basename string `json:"basename,omitempty"`
	// 文件后缀
	Extension string `json:"extension,omitempty"`
	// 文件大小
	Size *int64 `json:"size,omitempty"`
	// 修改时间
	ModifyTime *float64 `json:"modify_time,omitempty"`
	// 子节点
	Children []any `json:"children,omitempty"`
	// ID
	FileID string `json:"fileid,omitempty"`
	// 父ID
	ParentFileID string `json:"parent_fileid,omitempty"`
	// 缩略图
	Thumbnail string `json:"thumbnail,omitempty"`
	// 115 pickcode
	Pickcode string `json:"pickcode,omitempty"`
	// drive_id
	DriveID string `json:"drive_id,omitempty"`
	// url
	URL string `json:"url,omitempty"`
}

// Dir 返回文件所在目录
func (f *FileItem) Dir() string {
	return filepath.Dir(f.Path)
}

// Base 返回文件名
func (f *FileItem) Base() string {
	return filepath.Base(f.Path)
}

// IsDirectory 判断是否为目录
func (f *FileItem) IsDirectory() bool {
	return f.Type == "dir"
}

// IsFile 判断是否为文件
func (f *FileItem) IsFile() bool {
	return f.Type == "file"
}

// StorageUsage 存储使用情况
type StorageUsage struct {
	// 总空间
	Total float64 `json:"total,omitempty"`
	// 剩余空间
	Available float64 `json:"available,omitempty"`
}

// StorageTransType 存储传输类型
type StorageTransType struct {
	// 传输类型
	TransType map[string]any `json:"transtype,omitempty"`
}
