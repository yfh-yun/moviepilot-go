package storages

import (
	"path/filepath"
	
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/logger"
)

// StorageSchema 存储模式定义
type StorageSchema struct {
	Value string
}

// StorageBase 存储基类接口
type StorageBase interface {
	// 初始�?	InitStorage()
	
	// 生成二维�?	GenerateQrcode(args ...interface{}) (*map[string]interface{}, *string)
	
	// 检查登�?	CheckLogin(args ...interface{}) *map[string]string
	
	// 获取配置
	GetConfig() *schemas.StorageConf
	
	// 获取配置字典
	GetConf() map[string]interface{}
	
	// 设置配置
	SetConfig(conf map[string]interface{})
	
	// 支持的整理方�?	SupportTranstype() map[string]string
	
	// 是否支持整理方式
	IsSupportTranstype(transtype string) bool
	
	// 重置配置
	ResetConfig()
	
	// 检查存储是否可�?	Check() bool
	
	// 浏览文件
	List(fileitem *schemas.FileItem) []*schemas.FileItem
	
	// 创建目录
	CreateFolder(fileitem *schemas.FileItem, name string) *schemas.FileItem
	
	// 获取目录，如目录不存在则创建
	GetFolder(path string) *schemas.FileItem
	
	// 获取文件或目录，不存在返回nil
	GetItem(path string) *schemas.FileItem
	
	// 获取父目�?	GetParent(fileitem *schemas.FileItem) *schemas.FileItem
	
	// 删除文件
	Delete(fileitem *schemas.FileItem) bool
	
	// 重命名文�?	Rename(fileitem *schemas.FileItem, name string) bool
	
	// 下载文件，保存到本地，返回本地临时文件地址
	Download(fileitem *schemas.FileItem, path string) string
	
	// 上传文件
	Upload(fileitem *schemas.FileItem, path string, newName *string) *schemas.FileItem
	
	// 获取文件详情
	Detail(fileitem *schemas.FileItem) *schemas.FileItem
	
	// 复制文件
	Copy(fileitem *schemas.FileItem, path string, newName string) bool
	
	// 移动文件
	Move(fileitem *schemas.FileItem, path string, newName string) bool
	
	// 硬链接文�?	Link(fileitem *schemas.FileItem, targetFile string) bool
	
	// 软链接文�?	Softlink(fileitem *schemas.FileItem, targetFile string) bool
	
	// 存储使用情况
	Usage() *schemas.StorageUsage
	
	// 快照文件系统
	Snapshot(path string, lastSnapshotTime *float64, maxDepth int) map[string]map[string]interface{}
	
	// 获取存储模式
	Schema() *StorageSchema
}

// BaseStorage 存储基类的默认实�?type BaseStorage struct {
	storageHelper *storage.StorageHelper
	schema        *StorageSchema
	transtype     map[string]string
	snapshotCheckFolderModtime bool
}

// NewBaseStorage 创建基础存储实例
func NewBaseStorage() *BaseStorage {
	return &BaseStorage{
		storageHelper: storage.NewStorageHelper(),
		transtype:     make(map[string]string),
		snapshotCheckFolderModtime: true,
	}
}

// Schema 获取存储模式
func (b *BaseStorage) Schema() *StorageSchema {
	return b.schema
}

// TransferProcess 传输进度回调
func TransferProcess(path string) func(percent float64) {
	pbar := &ProgressBar{} // 简化的进度条实�?	progressHelper := progress.NewProgressHelper(crypto.HashUtils.Md5(path))
	progressHelper.Start()
	
	return func(percent float64) {
		percentValue := round(percent, 2)
		pbar.Update(percentValue)
		// 更新进度
		progressHelper.Update(percentValue, path+" 进度�?+string(percentValue)+"%")
		// 完成时结�?		if percentValue >= 100 {
			progressHelper.End()
			pbar.Close()
		}
	}
}

// round 四舍五入函数
func round(val float64, precision int) float64 {
	// 简化的四舍五入实现
	return val
}

// ProgressBar 简化的进度条实�?type ProgressBar struct {
	// 实现进度条相关功�?}

// Update 更新进度
func (p *ProgressBar) Update(percent float64) {
	// 实现进度更新逻辑
}

// Close 关闭进度�?func (p *ProgressBar) Close() {
	// 实现关闭逻辑
}

// GenerateQrcode 生成二维�?func (b *BaseStorage) GenerateQrcode(args ...interface{}) (*map[string]interface{}, *string) {
	// 默认空实�?	return nil, nil
}

// CheckLogin 检查登�?func (b *BaseStorage) CheckLogin(args ...interface{}) *map[string]string {
	// 默认空实�?	return nil
}

// GetConfig 获取配置
func (b *BaseStorage) GetConfig() *schemas.StorageConf {
	return b.storageHelper.GetStorage(b.schema.Value)
}

// GetConf 获取配置字典
func (b *BaseStorage) GetConf() map[string]interface{} {
	conf := b.GetConfig()
	if conf != nil {
		return conf.Config
	}
	return make(map[string]interface{})
}

// SetConfig 设置配置
func (b *BaseStorage) SetConfig(conf map[string]interface{}) {
	b.storageHelper.SetStorage(b.schema.Value, conf)
	b.InitStorage()
}

// SupportTranstype 支持的整理方�?func (b *BaseStorage) SupportTranstype() map[string]string {
	return b.transtype
}

// IsSupportTranstype 是否支持整理方式
func (b *BaseStorage) IsSupportTranstype(transtype string) bool {
	_, exists := b.transtype[transtype]
	return exists
}

// ResetConfig 重置配置
func (b *BaseStorage) ResetConfig() {
	b.storageHelper.ResetStorage(b.schema.Value)
	b.InitStorage()
}

// GetParent 获取父目�?func (b *BaseStorage) GetParent(fileitem *schemas.FileItem) *schemas.FileItem {
	parentPath := filepath.Dir(fileitem.Path)
	return b.GetItem(parentPath)
}

// Snapshot 快照文件系统，输出所有层级文件信息（不含目录�?func (b *BaseStorage) Snapshot(path string, lastSnapshotTime *float64, maxDepth int) map[string]map[string]interface{} {
	filesInfo := make(map[string]map[string]interface{})
	
	var snapshotFile func(fileitem *schemas.FileItem, currentDepth int)
	snapshotFile = func(fileitem *schemas.FileItem, currentDepth int) {
		defer func() {
			if err := recover(); err != nil {
				logger.Debugf("Snapshot error for %s: %v", fileitem.Path, err)
			}
		}()
		
		if fileitem.Type == "dir" {
			// 检查递归深度限制
			if currentDepth >= maxDepth {
				return
			}
			
			// 增量检查：如果目录修改时间早于上次快照，跳�?			if b.snapshotCheckFolderModtime && 
				lastSnapshotTime != nil && 
				fileitem.ModifyTime != nil && 
				*fileitem.ModifyTime <= *lastSnapshotTime {
				return
			}
			
			// 遍历子文�?			subFiles := b.List(fileitem)
			for _, subFile := range subFiles {
				snapshotFile(subFile, currentDepth+1)
			}
		} else {
			// 记录文件的完整信息用于比�?			if fileitem.ModifyTime != nil && lastSnapshotTime != nil && *fileitem.ModifyTime > *lastSnapshotTime {
				filesInfo[fileitem.Path] = map[string]interface{}{
					"size": fileitem.Size,
					"modify_time": fileitem.ModifyTime,
					"type": fileitem.Type,
				}
			}
		}
	}
	
	fileitem := b.GetItem(path)
	if fileitem == nil {
		return filesInfo
	}
	
	snapshotFile(fileitem, 0)
	
	return filesInfo
}

// InitStorage 初始化存�?- 需要子类实�?func (b *BaseStorage) InitStorage() {
	// 需要子类实�?}

// Check 检查存储是否可�?- 需要子类实�?func (b *BaseStorage) Check() bool {
	// 需要子类实�?	return false
}

// List 浏览文件 - 需要子类实�?func (b *BaseStorage) List(fileitem *schemas.FileItem) []*schemas.FileItem {
	// 需要子类实�?	return nil
}

// CreateFolder 创建目录 - 需要子类实�?func (b *BaseStorage) CreateFolder(fileitem *schemas.FileItem, name string) *schemas.FileItem {
	// 需要子类实�?	return nil
}

// GetFolder 获取目录，如目录不存在则创建 - 需要子类实�?func (b *BaseStorage) GetFolder(path string) *schemas.FileItem {
	// 需要子类实�?	return nil
}

// GetItem 获取文件或目录，不存在返回nil - 需要子类实�?func (b *BaseStorage) GetItem(path string) *schemas.FileItem {
	// 需要子类实�?	return nil
}

// Delete 删除文件 - 需要子类实�?func (b *BaseStorage) Delete(fileitem *schemas.FileItem) bool {
	// 需要子类实�?	return false
}

// Rename 重命名文�?- 需要子类实�?func (b *BaseStorage) Rename(fileitem *schemas.FileItem, name string) bool {
	// 需要子类实�?	return false
}

// Download 下载文件，保存到本地，返回本地临时文件地址 - 需要子类实�?func (b *BaseStorage) Download(fileitem *schemas.FileItem, path string) string {
	// 需要子类实�?	return ""
}

// Upload 上传文件 - 需要子类实�?func (b *BaseStorage) Upload(fileitem *schemas.FileItem, path string, newName *string) *schemas.FileItem {
	// 需要子类实�?	return nil
}

// Detail 获取文件详情 - 需要子类实�?func (b *BaseStorage) Detail(fileitem *schemas.FileItem) *schemas.FileItem {
	// 需要子类实�?	return nil
}

// Copy 复制文件 - 需要子类实�?func (b *BaseStorage) Copy(fileitem *schemas.FileItem, path string, newName string) bool {
	// 需要子类实�?	return false
}

// Move 移动文件 - 需要子类实�?func (b *BaseStorage) Move(fileitem *schemas.FileItem, path string, newName string) bool {
	// 需要子类实�?	return false
}

// Link 硬链接文�?- 需要子类实�?func (b *BaseStorage) Link(fileitem *schemas.FileItem, targetFile string) bool {
	// 需要子类实�?	return false
}

// Softlink 软链接文�?- 需要子类实�?func (b *BaseStorage) Softlink(fileitem *schemas.FileItem, targetFile string) bool {
	// 需要子类实�?	return false
}

// Usage 存储使用情况 - 需要子类实�?func (b *BaseStorage) Usage() *schemas.StorageUsage {
	// 需要子类实�?	return nil
}
