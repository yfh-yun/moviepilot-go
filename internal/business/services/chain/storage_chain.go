package chain

import (
	"context"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/cache"
)

// StorageChain 存储管理处理链
type StorageChain struct {
	cache          *cache.Cache
	logger         *logger.Logger
	storageService *service.StorageService
}

// NewStorageChain 创建存储管理处理链实例
func NewStorageChain(cache *cache.Cache, logger *logger.Logger, storageService *service.StorageService) *StorageChain {
	return &StorageChain{
		cache:          cache,
		logger:         logger,
		storageService: storageService,
	}
}

// GetStorageInfo 获取存储信息
func (c *StorageChain) GetStorageInfo(ctx context.Context, path string) (*model.StorageInfo, error) {
	c.logger.Info("获取存储信息", "path", path)

	info, err := c.storageService.GetStorageInfo(ctx, path)
	if err != nil {
		c.logger.Error("获取存储信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取存储信息成功", "path", path)
	return info, nil
}

// ListFiles 列出文件
func (c *StorageChain) ListFiles(ctx context.Context, path string, page, pageSize int) ([]*model.FileItem, int64, error) {
	c.logger.Info("列出文件", "path", path, "page", page, "pageSize", pageSize)

	files, total, err := c.storageService.ListFiles(ctx, path, page, pageSize)
	if err != nil {
		c.logger.Error("列出文件失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("列出文件成功", "count", len(files))
	return files, total, nil
}

// GetFileInfo 获取文件信息
func (c *StorageChain) GetFileInfo(ctx context.Context, path string) (*model.FileItem, error) {
	c.logger.Info("获取文件信息", "path", path)

	fileInfo, err := c.storageService.GetFileInfo(ctx, path)
	if err != nil {
		c.logger.Error("获取文件信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取文件信息成功", "path", path)
	return fileInfo, nil
}

// CreateDirectory 创建目录
func (c *StorageChain) CreateDirectory(ctx context.Context, path string) error {
	c.logger.Info("创建目录", "path", path)

	err := c.storageService.CreateDirectory(ctx, path)
	if err != nil {
		c.logger.Error("创建目录失败", "error", err)
		return err
	}

	c.logger.Info("创建目录成功", "path", path)
	return nil
}

// DeleteFile 删除文件
func (c *StorageChain) DeleteFile(ctx context.Context, path string) error {
	c.logger.Info("删除文件", "path", path)

	err := c.storageService.DeleteFile(ctx, path)
	if err != nil {
		c.logger.Error("删除文件失败", "error", err)
		return err
	}

	c.logger.Info("删除文件成功", "path", path)
	return nil
}

// RenameFile 重命名文件
func (c *StorageChain) RenameFile(ctx context.Context, oldPath, newPath string) error {
	c.logger.Info("重命名文件", "oldPath", oldPath, "newPath", newPath)

	err := c.storageService.RenameFile(ctx, oldPath, newPath)
	if err != nil {
		c.logger.Error("重命名文件失败", "error", err)
		return err
	}

	c.logger.Info("重命名文件成功", "oldPath", oldPath, "newPath", newPath)
	return nil
}

// CopyFile 复制文件
func (c *StorageChain) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	c.logger.Info("复制文件", "srcPath", srcPath, "dstPath", dstPath)

	err := c.storageService.CopyFile(ctx, srcPath, dstPath)
	if err != nil {
		c.logger.Error("复制文件失败", "error", err)
		return err
	}

	c.logger.Info("复制文件成功", "srcPath", srcPath, "dstPath", dstPath)
	return nil
}

// MoveFile 移动文件
func (c *StorageChain) MoveFile(ctx context.Context, srcPath, dstPath string) error {
	c.logger.Info("移动文件", "srcPath", srcPath, "dstPath", dstPath)

	err := c.storageService.MoveFile(ctx, srcPath, dstPath)
	if err != nil {
		c.logger.Error("移动文件失败", "error", err)
		return err
	}

	c.logger.Info("移动文件成功", "srcPath", srcPath, "dstPath", dstPath)
	return nil
}

// UploadFile 上传文件
func (c *StorageChain) UploadFile(ctx context.Context, dstPath string, data []byte) error {
	c.logger.Info("上传文件", "dstPath", dstPath)

	err := c.storageService.UploadFile(ctx, dstPath, data)
	if err != nil {
		c.logger.Error("上传文件失败", "error", err)
		return err
	}

	c.logger.Info("上传文件成功", "dstPath", dstPath)
	return nil
}

// DownloadFile 下载文件
func (c *StorageChain) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	c.logger.Info("下载文件", "path", path)

	data, err := c.storageService.DownloadFile(ctx, path)
	if err != nil {
		c.logger.Error("下载文件失败", "error", err)
		return nil, err
	}

	c.logger.Info("下载文件成功", "path", path)
	return data, nil
}

// GetDirectorySize 获取目录大小
func (c *StorageChain) GetDirectorySize(ctx context.Context, path string) (int64, error) {
	c.logger.Info("获取目录大小", "path", path)

	size, err := c.storageService.GetDirectorySize(ctx, path)
	if err != nil {
		c.logger.Error("获取目录大小失败", "error", err)
		return 0, err
	}

	c.logger.Info("获取目录大小成功", "path", path, "size", size)
	return size, nil
}

// CleanTempFiles 清理临时文件
func (c *StorageChain) CleanTempFiles(ctx context.Context) error {
	c.logger.Info("清理临时文件")

	err := c.storageService.CleanTempFiles(ctx)
	if err != nil {
		c.logger.Error("清理临时文件失败", "error", err)
		return err
	}

	c.logger.Info("清理临时文件成功")
	return nil
}

// GetStorageStats 获取存储统计信息
func (c *StorageChain) GetStorageStats(ctx context.Context) (*model.StorageStats, error) {
	c.logger.Info("获取存储统计信息")

	stats, err := c.storageService.GetStorageStats(ctx)
	if err != nil {
		c.logger.Error("获取存储统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// MountStorage 挂载存储
func (c *StorageChain) MountStorage(ctx context.Context, mountPoint, storageType string, config map[string]interface{}) error {
	c.logger.Info("挂载存储", "mountPoint", mountPoint, "storageType", storageType)

	err := c.storageService.MountStorage(ctx, mountPoint, storageType, config)
	if err != nil {
		c.logger.Error("挂载存储失败", "error", err)
		return err
	}

	c.logger.Info("挂载存储成功", "mountPoint", mountPoint)
	return nil
}

// UnmountStorage 卸载存储
func (c *StorageChain) UnmountStorage(ctx context.Context, mountPoint string) error {
	c.logger.Info("卸载存储", "mountPoint", mountPoint)

	err := c.storageService.UnmountStorage(ctx, mountPoint)
	if err != nil {
		c.logger.Error("卸载存储失败", "error", err)
		return err
	}

	c.logger.Info("卸载存储成功", "mountPoint", mountPoint)
	return nil
}

// GetMountedStorages 获取已挂载的存储列表
func (c *StorageChain) GetMountedStorages(ctx context.Context) ([]*model.MountedStorage, error) {
	c.logger.Info("获取已挂载的存储列表")

	storages, err := c.storageService.GetMountedStorages(ctx)
	if err != nil {
		c.logger.Error("获取已挂载的存储列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取已挂载的存储列表成功", "count", len(storages))
	return storages, nil
}

// CheckStorageHealth 检查存储健康状态
func (c *StorageChain) CheckStorageHealth(ctx context.Context, mountPoint string) (*model.StorageHealth, error) {
	c.logger.Info("检查存储健康状态", "mountPoint", mountPoint)

	health, err := c.storageService.CheckStorageHealth(ctx, mountPoint)
	if err != nil {
		c.logger.Error("检查存储健康状态失败", "error", err)
		return nil, err
	}

	c.logger.Info("检查存储健康状态成功", "mountPoint", mountPoint)
	return health, nil
}
