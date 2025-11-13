package helper

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/moby/moby/client"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// SystemHelper 系统工具类，提供系统相关的操作和判断
type SystemHelper struct {
	systemFlagFile string
}

var (
	systemHelperInstance *SystemHelper
	systemHelperOnce     sync.Once
)

// NewSystemHelper 创建SystemHelper单例实例
func NewSystemHelper() *SystemHelper {
	systemHelperOnce.Do(func() {
		systemHelperInstance = &SystemHelper{
			systemFlagFile: "/var/log/nginx/__moviepilot__",
		}
	})
	return systemHelperInstance
}

// HandleConfigChanged 处理配置变更事件，更新日志设�?func (s *SystemHelper) HandleConfigChanged(eventData map[string]interface{}) {
	/*
	 * 处理配置变更事件，更新日志设�?	 * :param eventData: 事件数据
	 */
	if eventData == nil {
		return
	}

	// 检查事件数据中的key
	if keyVal, exists := eventData["key"]; exists {
		if keyStr, ok := keyVal.(string); ok {
			validKeys := []string{"DEBUG", "LOG_LEVEL", "LOG_MAX_FILE_SIZE", "LOG_BACKUP_COUNT", "LOG_FILE_FORMAT", "LOG_CONSOLE_FORMAT"}
			isValidKey := false
			for _, validKey := range validKeys {
				if keyStr == validKey {
					isValidKey = true
					break
				}
			}

			if !isValidKey {
				return
			}
		}
	}

	logger.GetLoggerManager().Info("配置变更，更新日志设�?..")
	// TODO: 实现日志更新逻辑
	// logger.update_loggers()
}

// CanRestart 判断是否可以内部重启
func (s *SystemHelper) CanRestart() bool {
	/*
	 * 判断是否可以内部重启
	 */
	_, err := os.Stat("/var/run/docker.sock")
	dockerSockExists := !os.IsNotExist(err)

	cfg := config.GetConfig()
	return dockerSockExists || cfg.DOCKER_CLIENT_API != "tcp://127.0.0.1:38379"
}

// getContainerID 获取当前容器ID
func (s *SystemHelper) getContainerID() *string {
	/*
	 * 获取当前容器ID
	 */
	containerID := ""
	
	defer func() {
		if r := recover(); r != nil {
			logger.GetLoggerManager().Debugf("获取容器ID失败: %v", r)
		}
	}()

	// 尝试读取 /proc/self/mountinfo 文件
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		logger.GetLoggerManager().Debugf("获取容器ID失败: %v", err)
		return nil
	}

	content := string(data)
	indexResolvConf := strings.Index(content, "resolv.conf")
	if indexResolvConf != -1 {
		indexSecondSlash := strings.LastIndex(content[:indexResolvConf], "/")
		indexFirstSlash := strings.LastIndex(content[:indexSecondSlash], "/") + 1
		containerID = content[indexFirstSlash:indexSecondSlash]
		if len(containerID) < 20 {
			indexResolvConf = strings.Index(content, "/sys/fs/cgroup/devices")
			if indexResolvConf != -1 {
				indexSecondSlash = strings.LastIndex(content[:indexResolvConf], " ")
				indexFirstSlash = strings.LastIndex(content[:indexSecondSlash], "/") + 1
				containerID = content[indexFirstSlash:indexSecondSlash]
			}
		}
	}

	if containerID != "" {
		return &containerID
	}
	return nil
}

// checkRestartPolicy 检查当前容器是否配置了自动重启策略
func (s *SystemHelper) checkRestartPolicy() bool {
	/*
	 * 检查当前容器是否配置了自动重启策略
	 */
	defer func() {
		if r := recover(); r != nil {
			logger.GetLoggerManager().Warningf("检查重启策略失�? %v", r)
		}
	}()

	// 获取当前容器ID
	containerID := s.getContainerID()
	if containerID == nil {
		return false
	}

	// 创建 Docker 客户�?	cfg := config.GetConfig()
	cli, err := client.NewClientWithOpts(client.WithHost(cfg.DOCKER_CLIENT_API))
	if err != nil {
		logger.GetLoggerManager().Warningf("检查重启策略失�? %v", err)
		return false
	}
	defer cli.Close()

	// 获取容器信息
	container, err := cli.ContainerInspect(context.Background(), *containerID)
	if err != nil {
		logger.GetLoggerManager().Warningf("检查重启策略失�? %v", err)
		return false
	}

	// 获取重启策略
	restartPolicy := container.HostConfig.RestartPolicy
	policyName := restartPolicy.Name

	// 检查是否有有效的重启策�?	autoRestartPolicies := []string{"always", "unless-stopped", "on-failure"}
	hasRestartPolicy := false
	for _, policy := range autoRestartPolicies {
		if policyName == policy {
			hasRestartPolicy = true
			break
		}
	}

	logger.GetLoggerManager().Infof("容器重启策略: %s, 支持自动重启: %v", policyName, hasRestartPolicy)
	return hasRestartPolicy
}

// Restart 执行Docker重启操作
func (s *SystemHelper) Restart() (bool, string) {
	/*
	 * 执行Docker重启操作
	 */
	if !utils.NewSystemUtils().IsDocker() {
		return false, "非Docker环境，无法重启！"
	}

	defer func() {
		if r := recover(); r != nil {
			logger.GetLoggerManager().Errorf("重启失败: %v", r)
			// 降级为Docker API重启
			logger.GetLoggerManager().Warning("降级为Docker API重启...")
			s.dockerAPIRestart()
		}
	}()

	// 检查容器是否配置了自动重启策略
	hasRestartPolicy := s.checkRestartPolicy()
	if hasRestartPolicy {
		// 有重启策略，使用优雅退出方�?		logger.GetLoggerManager().Info("检测到容器配置了自动重启策略，使用优雅重启方式...")
		// 启动优雅退出超时监�?		s.startGracefulShutdownMonitor()
		// 发送SIGTERM信号给当前进程，触发优雅停止
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
		return true, ""
	} else {
		// 没有重启策略，使用Docker API强制重启
		logger.GetLoggerManager().Info("容器未配置自动重启策略，使用Docker API重启...")
		return s.dockerAPIRestart()
	}
}

// startGracefulShutdownMonitor 启动优雅退出超时监�?func (s *SystemHelper) startGracefulShutdownMonitor() {
	/*
	 * 启动优雅退出超时监�?	 * 如果30秒内进程没有退出，则使用Docker API强制重启
	 */

	go func() {
		// 等待30�?		time.Sleep(30 * time.Second)
		logger.GetLoggerManager().Warning("优雅退出超�?0秒，使用Docker API强制重启...")
		defer func() {
			if r := recover(); r != nil {
				logger.GetLoggerManager().Errorf("强制重启失败: %v", r)
			}
		}()
		s.dockerAPIRestart()
	}()
}

// dockerAPIRestart 使用Docker API重启容器，并尝试优雅停止
func (s *SystemHelper) dockerAPIRestart() (bool, string) {
	/*
	 * 使用Docker API重启容器，并尝试优雅停止
	 */
	defer func() {
		if r := recover(); r != nil {
			logger.GetLoggerManager().Errorf("重启时发生错�? %v", r)
		}
	}()

	// 创建 Docker 客户�?	cfg := config.GetConfig()
	cli, err := client.NewClientWithOpts(client.WithHost(cfg.DOCKER_CLIENT_API))
	if err != nil {
		return false, fmt.Sprintf("重启时发生错误：%v", err)
	}
	defer cli.Close()

	containerID := s.getContainerID()
	if containerID == nil {
		return false, "获取容器ID失败�?
	}

	// 重启容器
	err = cli.ContainerRestart(context.Background(), *containerID, nil)
	if err != nil {
		return false, fmt.Sprintf("重启时发生错误：%v", err)
	}

	return true, ""
}

// SetSystemModified 设置系统已修改标�?func (s *SystemHelper) SetSystemModified() {
	/*
	 * 设置系统已修改标�?	 */
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("设置系统修改标志失败: %v\n", r)
		}
	}()

	if utils.NewSystemUtils().IsDocker() {
		flagFile := s.systemFlagFile
		// 确保目录存在
		dir := filepath.Dir(flagFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("创建目录失败: %v\n", err)
				return
			}
		}
		
		// 创建或更新文�?		file, err := os.Create(flagFile)
		if err != nil {
			fmt.Printf("创建标志文件失败: %v\n", err)
			return
		}
		defer file.Close()
		
		// 更新文件的修改时�?		now := time.Now()
		err = os.Chtimes(flagFile, now, now)
		if err != nil {
			fmt.Printf("更新标志文件时间失败: %v\n", err)
		}
	}
}

// IsSystemReset 检查系统是否已被重�?func (s *SystemHelper) IsSystemReset() bool {
	/*
	 * 检查系统是否已被重�?	 * :return: 如果系统已重置，返回 True；否则返�?False
	 */
	if utils.NewSystemUtils().IsDocker() {
		_, err := os.Stat(s.systemFlagFile)
		return os.IsNotExist(err)
	}
	return false
}

// SetupSignalHandler 设置信号处理器，用于处理优雅关闭
func (s *SystemHelper) SetupSignalHandler() {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	
	// 注册要捕获的信号
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	
	// 启动goroutine处理信号
	go func() {
		sig := <-sigChan
		logger.GetLoggerManager().Infof("接收到信�? %v", sig)
		
		// 在这里可以执行清理工�?		// 例如：关闭数据库连接、停止定时任务等
		
		// 退出程�?		os.Exit(0)
	}()
}
