package chain

import (
	"context"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// SystemChain 系统管理处理链
type SystemChain struct {
	cache         *cache.Cache
	logger        *logger.Logger
	systemService *service.SystemService
}

// NewSystemChain 创建系统管理处理链实例
func NewSystemChain(cache *cache.Cache, logger *logger.Logger, systemService *service.SystemService) *SystemChain {
	return &SystemChain{
		cache:         cache,
		logger:        logger,
		systemService: systemService,
	}
}

// GetSystemInfo 获取系统信息
func (c *SystemChain) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	c.logger.Info("获取系统信息")

	info, err := c.systemService.GetSystemInfo(ctx)
	if err != nil {
		c.logger.Error("获取系统信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取系统信息成功")
	return info, nil
}

// GetSystemStats 获取系统统计信息
func (c *SystemChain) GetSystemStats(ctx context.Context) (*model.SystemStats, error) {
	c.logger.Info("获取系统统计信息")

	stats, err := c.systemService.GetSystemStats(ctx)
	if err != nil {
		c.logger.Error("获取系统统计信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取系统统计信息成功")
	return stats, nil
}

// GetServiceStatus 获取服务状态
func (c *SystemChain) GetServiceStatus(ctx context.Context, serviceName string) (*model.ServiceStatus, error) {
	c.logger.Info("获取服务状态", "serviceName", serviceName)

	status, err := c.systemService.GetServiceStatus(ctx, serviceName)
	if err != nil {
		c.logger.Error("获取服务状态失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取服务状态成功", "serviceName", serviceName)
	return status, nil
}

// RestartService 重启服务
func (c *SystemChain) RestartService(ctx context.Context, serviceName string) error {
	c.logger.Info("重启服务", "serviceName", serviceName)

	err := c.systemService.RestartService(ctx, serviceName)
	if err != nil {
		c.logger.Error("重启服务失败", "error", err)
		return err
	}

	c.logger.Info("重启服务成功", "serviceName", serviceName)
	return nil
}

// StopService 停止服务
func (c *SystemChain) StopService(ctx context.Context, serviceName string) error {
	c.logger.Info("停止服务", "serviceName", serviceName)

	err := c.systemService.StopService(ctx, serviceName)
	if err != nil {
		c.logger.Error("停止服务失败", "error", err)
		return err
	}

	c.logger.Info("停止服务成功", "serviceName", serviceName)
	return nil
}

// StartService 启动服务
func (c *SystemChain) StartService(ctx context.Context, serviceName string) error {
	c.logger.Info("启动服务", "serviceName", serviceName)

	err := c.systemService.StartService(ctx, serviceName)
	if err != nil {
		c.logger.Error("启动服务失败", "error", err)
		return err
	}

	c.logger.Info("启动服务成功", "serviceName", serviceName)
	return nil
}

// GetProcessInfo 获取进程信息
func (c *SystemChain) GetProcessInfo(ctx context.Context) (*model.ProcessInfo, error) {
	c.logger.Info("获取进程信息")

	info, err := c.systemService.GetProcessInfo(ctx)
	if err != nil {
		c.logger.Error("获取进程信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取进程信息成功")
	return info, nil
}

// GetDiskUsage 获取磁盘使用情况
func (c *SystemChain) GetDiskUsage(ctx context.Context) ([]*model.DiskUsage, error) {
	c.logger.Info("获取磁盘使用情况")

	usage, err := c.systemService.GetDiskUsage(ctx)
	if err != nil {
		c.logger.Error("获取磁盘使用情况失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取磁盘使用情况成功")
	return usage, nil
}

// GetNetworkInfo 获取网络信息
func (c *SystemChain) GetNetworkInfo(ctx context.Context) (*model.NetworkInfo, error) {
	c.logger.Info("获取网络信息")

	info, err := c.systemService.GetNetworkInfo(ctx)
	if err != nil {
		c.logger.Error("获取网络信息失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取网络信息成功")
	return info, nil
}

// GetSystemLogs 获取系统日志
func (c *SystemChain) GetSystemLogs(ctx context.Context, lines int) ([]string, error) {
	c.logger.Info("获取系统日志", "lines", lines)

	logs, err := c.systemService.GetSystemLogs(ctx, lines)
	if err != nil {
		c.logger.Error("获取系统日志失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取系统日志成功", "count", len(logs))
	return logs, nil
}

// ClearSystemCache 清理系统缓存
func (c *SystemChain) ClearSystemCache(ctx context.Context) error {
	c.logger.Info("清理系统缓存")

	err := c.systemService.ClearSystemCache(ctx)
	if err != nil {
		c.logger.Error("清理系统缓存失败", "error", err)
		return err
	}

	c.logger.Info("清理系统缓存成功")
	return nil
}

// BackupDatabase 备份数据库
func (c *SystemChain) BackupDatabase(ctx context.Context, backupPath string) error {
	c.logger.Info("备份数据库", "backupPath", backupPath)

	err := c.systemService.BackupDatabase(ctx, backupPath)
	if err != nil {
		c.logger.Error("备份数据库失败", "error", err)
		return err
	}

	c.logger.Info("备份数据库成功", "backupPath", backupPath)
	return nil
}

// RestoreDatabase 恢复数据库
func (c *SystemChain) RestoreDatabase(ctx context.Context, backupPath string) error {
	c.logger.Info("恢复数据库", "backupPath", backupPath)

	err := c.systemService.RestoreDatabase(ctx, backupPath)
	if err != nil {
		c.logger.Error("恢复数据库失败", "error", err)
		return err
	}

	c.logger.Info("恢复数据库成功", "backupPath", backupPath)
	return nil
}

// ShutdownSystem 关闭系统
func (c *SystemChain) ShutdownSystem(ctx context.Context) error {
	c.logger.Info("关闭系统")

	err := c.systemService.ShutdownSystem(ctx)
	if err != nil {
		c.logger.Error("关闭系统失败", "error", err)
		return err
	}

	c.logger.Info("关闭系统成功")
	return nil
}

// RestartSystem 重启系统
func (c *SystemChain) RestartSystem(ctx context.Context) error {
	c.logger.Info("重启系统")

	err := c.systemService.RestartSystem(ctx)
	if err != nil {
		c.logger.Error("重启系统失败", "error", err)
		return err
	}

	c.logger.Info("重启系统成功")
	return nil
}

// GetSystemConfig 获取系统配置
func (c *SystemChain) GetSystemConfig(ctx context.Context) (*model.SystemConfig, error) {
	c.logger.Info("获取系统配置")

	config, err := c.systemService.GetSystemConfig(ctx)
	if err != nil {
		c.logger.Error("获取系统配置失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取系统配置成功")
	return config, nil
}

// UpdateSystemConfig 更新系统配置
func (c *SystemChain) UpdateSystemConfig(ctx context.Context, config model.SystemConfig) error {
	c.logger.Info("更新系统配置")

	err := c.systemService.UpdateSystemConfig(ctx, config)
	if err != nil {
		c.logger.Error("更新系统配置失败", "error", err)
		return err
	}

	c.logger.Info("更新系统配置成功")
	return nil
}

// CheckSystemHealth 检查系统健康状态
func (c *SystemChain) CheckSystemHealth(ctx context.Context) (*model.SystemHealth, error) {
	c.logger.Info("检查系统健康状态")

	health, err := c.systemService.CheckSystemHealth(ctx)
	if err != nil {
		c.logger.Error("检查系统健康状态失败", "error", err)
		return nil, err
	}

	c.logger.Info("检查系统健康状态成功")
	return health, nil
}
