package scheduler

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	
	"moviepilot-go/internal/logger"
)

// JobInfo 定时任务信息
type JobInfo struct {
	Name         string
	Func         func()
	Running      bool
	RunInProcess bool
	ProviderName string
	Kwargs       map[string]interface{}
}

// Scheduler 定时任务管理�?type Scheduler struct {
	// 定时服务
	scheduler *cron.Cron
	
	// 退出事�?	event chan struct{}
	
	// �?	mutex sync.RWMutex
	
	// 各服务的运行状�?	jobs map[string]*JobInfo
	
	// 日志记录�?	logger *zap.Logger
	
	// 用户认证失败次数
	authCount int
	
	// 用户认证失败消息发�?	authMessage bool
	
	// 当前上下�?	ctx context.Context
}

// schedulerInstance 定时任务管理器单�?var schedulerInstance *Scheduler
var schedulerOnce sync.Once

// GetScheduler 获取定时任务管理器单�?func GetScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		// 获取日志记录�?		logManager := logger.GetLoggerManager()
		zapLogger := logManager.GetLogger("scheduler")
		
		schedulerInstance = &Scheduler{
			scheduler: cron.New(cron.WithSeconds()),
			event:     make(chan struct{}),
			jobs:      make(map[string]*JobInfo),
			logger:    zapLogger,
			ctx:       context.Background(),
		}
	})
	return schedulerInstance
}

// Init 初始化定时服�?func (s *Scheduler) Init() {
	// 停止定时服务
	s.Stop()
	
	// 调试模式不启动定时服�?(简化实�?
	// if settings.DEV {
	//     return
	// }
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 各服务的运行状�?	s.jobs = map[string]*JobInfo{
		"cookiecloud": {
			Name:    "同步CookieCloud站点",
			Func:    func() { s.Start("cookiecloud") },
			Running: false,
		},
		"mediaserver_sync": {
			Name:    "同步媒体服务�?,
			Func:    func() { s.Start("mediaserver_sync") },
			Running: false,
		},
		"subscribe_tmdb": {
			Name:    "订阅元数据更�?,
			Func:    func() { s.Start("subscribe_tmdb") },
			Running: false,
		},
		"subscribe_search": {
			Name:    "订阅搜索补全",
			Func:    func() { s.Start("subscribe_search") },
			Running: false,
			Kwargs: map[string]interface{}{
				"state": "R",
			},
		},
		"new_subscribe_search": {
			Name:    "新增订阅搜索",
			Func:    func() { s.Start("new_subscribe_search") },
			Running: false,
			Kwargs: map[string]interface{}{
				"state": "N",
			},
		},
		"subscribe_refresh": {
			Name:    "订阅刷新",
			Func:    func() { s.Start("subscribe_refresh") },
			Running: false,
		},
		"subscribe_follow": {
			Name:    "关注的订阅分�?,
			Func:    func() { s.Start("subscribe_follow") },
			Running: false,
		},
		"transfer": {
			Name:    "下载文件整理",
			Func:    func() { s.Start("transfer") },
			Running: false,
		},
		"clear_cache": {
			Name:    "缓存清理",
			Func:    s.ClearCache,
			Running: false,
		},
		"user_auth": {
			Name:    "用户认证检�?,
			Func:    s.UserAuth,
			Running: false,
		},
		"scheduler_job": {
			Name:    "公共定时服务",
			Func:    func() { s.Start("scheduler_job") },
			Running: false,
		},
		"random_wallpager": {
			Name:    "壁纸缓存",
			Func:    func() { s.Start("random_wallpager") },
			Running: false,
		},
		"sitedata_refresh": {
			Name:    "站点数据刷新",
			Func:    func() { s.Start("sitedata_refresh") },
			Running: false,
		},
		"recommend_refresh": {
			Name:    "推荐缓存",
			Func:    func() { s.Start("recommend_refresh") },
			Running: false,
		},
		"plugin_market_refresh": {
			Name:    "插件市场缓存",
			Func:    func() { s.Start("plugin_market_refresh") },
			Running: false,
			Kwargs: map[string]interface{}{
				"force": true,
			},
		},
		"subscribe_calendar_cache": {
			Name:    "订阅日历缓存",
			Func:    func() { s.Start("subscribe_calendar_cache") },
			Running: false,
		},
		"full_gc": {
			Name:    "主动内存回收",
			Func:    s.FullGC,
			Running: false,
		},
	}
	
	// 启动定时服务进程
	s.scheduler.Start()
	
	// 添加定时任务 (应该从配置中读取时间间隔)
	// CookieCloud定时同步 (默认�?0分钟，应该从配置读取)
	s.scheduler.AddFunc("@every 30m", func() { s.Start("cookiecloud") })
	
	// 媒体服务器同�?(默认�?小时，应该从配置读取)
	s.scheduler.AddFunc("@every 6h", func() { s.Start("mediaserver_sync") })
	
	// 新增订阅时搜索（5分钟检查一次）
	s.scheduler.AddFunc("@every 5m", func() { s.Start("new_subscribe_search") })
	
	// 检查更新订阅TMDB数据（每�?小时�?	s.scheduler.AddFunc("@every 6h", func() { s.Start("subscribe_tmdb") })
	
	// 订阅状态每�?4小时搜索一�?	s.scheduler.AddFunc("@every 24h", func() { s.Start("subscribe_search") })
	
	// RSS订阅刷新 (默认�?0分钟，应该从配置读取)
	s.scheduler.AddFunc("@every 30m", func() { s.Start("subscribe_refresh") })
	
	// 关注订阅分享（每1小时�?	s.scheduler.AddFunc("@every 1h", func() { s.Start("subscribe_follow") })
	
	// 下载器文件转移（�?分钟�?	s.scheduler.AddFunc("@every 5m", func() { s.Start("transfer") })
	
	// 后台刷新TMDB壁纸 (�?0分钟)
	s.scheduler.AddFunc("@every 30m", func() { s.Start("random_wallpager") })
	
	// 公共定时服务 (�?0分钟)
	s.scheduler.AddFunc("@every 10m", func() { s.Start("scheduler_job") })
	
	// 缓存清理服务，每�?4小时
	s.scheduler.AddFunc("@every 24h", func() { s.Start("clear_cache") })
	
	// 定时检查用户认证，每隔10分钟
	s.scheduler.AddFunc("@every 10m", func() { s.Start("user_auth") })
	
	// 站点数据刷新 (默认�?小时，应该从配置读取)
	s.scheduler.AddFunc("@every 6h", func() { s.Start("sitedata_refresh") })
	
	// 推荐缓存 (�?4小时)
	s.scheduler.AddFunc("@every 24h", func() { s.Start("recommend_refresh") })
	
	// 插件市场缓存 (�?0分钟)
	s.scheduler.AddFunc("@every 30m", func() { s.Start("plugin_market_refresh") })
	
	// 订阅日历缓存 (�?小时)
	s.scheduler.AddFunc("@every 6h", func() { s.Start("subscribe_calendar_cache") })
	
	// 主动内存回收 (�?小时，应该从配置读取)
	s.scheduler.AddFunc("@every 1h", func() { s.Start("full_gc") })
	
	// 初始化工作流服务
	s.InitWorkflowJobs()
	
	// 初始化插件服�?	s.InitPluginJobs()
	
	s.logger.Info("定时任务初始化完�?)
}

// prepareJob 准备定时任务
func (s *Scheduler) prepareJob(jobID string) *JobInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	job, exists := s.jobs[jobID]
	if !exists {
		return nil
	}
	
	if job.Running {
		s.logger.Warn(fmt.Sprintf("定时任务 %s - %s 正在运行 ...", jobID, job.Name))
		return nil
	}
	
	s.jobs[jobID].Running = true
	return job
}

// finishJob 完成定时任务
func (s *Scheduler) finishJob(jobID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if job, exists := s.jobs[jobID]; exists {
		job.Running = false
	}
}

// Start 启动定时服务
func (s *Scheduler) Start(jobID string) {
	// 获取定时任务
	job := s.prepareJob(jobID)
	if job == nil {
		return
	}
	
	// 开始运�?	defer s.finishJob(jobID)
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error(fmt.Sprintf("定时任务 %s 执行失败�?v", job.Name, r))
				// 应该发送错误通知，类似Python版本中的实现
				// MessageHelper().put(title=f"{job.get('name')} 执行失败",
				//                    message=str(e),
				//                    role="system")
			}
		}()
		
		// 执行任务函数
		if job.Func != nil {
			job.Func()
		}
	}()
}

// ClearCache 清理缓存
func (s *Scheduler) ClearCache() {
	s.logger.Info("正在清理缓存...")
	// 实际的缓存清理逻辑应该在这里实�?	s.logger.Info("缓存清理完成")
}

// FullGC 主动内存回收
func (s *Scheduler) FullGC() {
	s.logger.Info("正在进行主动内存回收...")
	
	// 执行GC
	runtime.GC()
	
	s.logger.Info("主动内存回收完成")
}

// UserAuth 用户认证检�?func (s *Scheduler) UserAuth() {
	// 最大重试次�?	maxTry := 30
	if s.authCount > maxTry {
		if !s.authMessage {
			s.logger.Error("用户认证失败次数过多，将不再尝试认证�?)
			s.authMessage = true
		}
		return
	}
	
	s.logger.Info("用户未认证，正在尝试认证...")
	
	// 简化实现，实际应该检查用户认证状�?	// status, msg := SitesHelper().check_user()
	status := false // 简化实�?	msg := "认证失败" // 简化实�?	
	if status {
		s.authCount = 0
		s.logger.Info(fmt.Sprintf("%s 用户认证成功", msg))
		// 认证通过后重新初始化插件
		// PluginManager().init_config()
		// s.init_plugin_jobs()
	} else {
		s.authCount++
		s.logger.Error(fmt.Sprintf("用户认证失败�?s，共失败 %d �?, msg, s.authCount))
		if s.authCount >= maxTry {
			s.logger.Error("用户认证失败次数过多，将不再尝试认证�?)
		}
	}
}

// Stop 关闭定时服务
func (s *Scheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.scheduler != nil {
		s.logger.Info("正在停止定时任务...")
		// 发送退出事�?		close(s.event)
		// 停止定时任务
		s.scheduler.Stop()
		s.logger.Info("定时任务停止完成")
	}
}

// InitPluginJobs 初始化插件定时服�?func (s *Scheduler) InitPluginJobs() {
	// 实际应该初始化插件定时服�?	s.logger.Info("正在初始化插件定时服�?..")
	// 这里应该调用插件管理器获取运行中的插件ID并更新它们的定时任务
	// 示例:
	// for _, pid := range PluginManager().GetRunningPluginIDs() {
	//     s.UpdatePluginJob(pid)
	// }
}

// InitWorkflowJobs 初始化工作流定时服务
func (s *Scheduler) InitWorkflowJobs() {
	// 实际应该初始化工作流定时服务
	s.logger.Info("正在初始化工作流定时服务...")
	// 这里应该调用工作流链获取定时工作流并更新它们的定时任�?	// 示例:
	// for _, workflow := range WorkflowChain().GetTimerWorkflows() {
	//     s.UpdateWorkflowJob(workflow)
	// }
}

// UpdateWorkflowJob 更新工作流定时服�?func (s *Scheduler) UpdateWorkflowJob(workflowID string, name string, timer string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	jobID := fmt.Sprintf("workflow-%s", workflowID)
	
	// 移除已存在的任务
	s.removeJob(jobID)
	
	// 添加新的工作流任�?	s.jobs[jobID] = &JobInfo{
		Name:         name,
		Func:         func() { s.Start(jobID) },
		Running:      false,
		ProviderName: "工作�?,
	}
	
	// 添加到定时器
	// 解析cron表达�?	schedulerID, err := s.scheduler.AddFunc(timer, func() { s.Start(jobID) })
	if err != nil {
		s.logger.Error(fmt.Sprintf("添加工作流定时任务失�? %v", err))
		return
	}
	
	s.logger.Info(fmt.Sprintf("注册工作流服务：%s - %s (ID: %d)", name, timer, schedulerID))
}

// UpdatePluginJob 更新插件定时服务
func (s *Scheduler) UpdatePluginJob(pluginID string, jobID string, name string, timer string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	fullJobID := fmt.Sprintf("%s_%s", pluginID, jobID)
	
	// 移除已存在的任务
	s.removeJob(fullJobID)
	
	// 添加新的插件任务
	s.jobs[fullJobID] = &JobInfo{
		Name:         name,
		Func:         func() { s.Start(fullJobID) },
		Running:      false,
		ProviderName: "插件",
	}
	
	// 添加到定时器
	// 解析cron表达�?	schedulerID, err := s.scheduler.AddFunc(timer, func() { s.Start(fullJobID) })
	if err != nil {
		s.logger.Error(fmt.Sprintf("添加插件定时任务失败: %v", err))
		return
	}
	
	s.logger.Info(fmt.Sprintf("注册插件服务�?s - %s (ID: %d)", name, timer, schedulerID))
}

// removeJob 移除定时服务
func (s *Scheduler) removeJob(jobID string) {
	// 实际应该移除定时任务
	if _, exists := s.jobs[jobID]; exists {
		delete(s.jobs, jobID)
		s.logger.Info(fmt.Sprintf("移除定时服务�?s", jobID))
	}
}

// RemoveWorkflowJob 移除工作流服�?func (s *Scheduler) RemoveWorkflowJob(workflowID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	jobID := fmt.Sprintf("workflow-%s", workflowID)
	
	// 从任务列表中移除
	if _, exists := s.jobs[jobID]; exists {
		delete(s.jobs, jobID)
		s.logger.Info(fmt.Sprintf("移除工作流服务：%s", jobID))
	}
}

// RemovePluginJob 移除插件服务
func (s *Scheduler) RemovePluginJob(pluginID string, jobID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	fullJobID := fmt.Sprintf("%s_%s", pluginID, jobID)
	
	// 从任务列表中移除
	if _, exists := s.jobs[fullJobID]; exists {
		delete(s.jobs, fullJobID)
		s.logger.Info(fmt.Sprintf("移除插件服务�?s", fullJobID))
	}
}

// List 当前所有任�?func (s *Scheduler) List() []*ScheduleInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 返回计时任务
	schedulers := make([]*ScheduleInfo, 0)
	
	// 如果调度器未初始化或未运行，返回空列�?	if s.scheduler == nil {
		return schedulers
	}
	
	// 将正在运行的任务提取出来 (保障一次性任务正常显�?
	for jobID, job := range s.jobs {
		if job.Running && job.Name != "" {
			schedulers = append(schedulers, &ScheduleInfo{
				ID:       jobID,
				Name:     job.Name,
				Provider: job.ProviderName,
				Status:   "正在运行",
			})
		}
	}
	
	// 获取其他待执行任�?	jobs := s.scheduler.Entries()
	for _, job := range jobs {
		// 简化实现，实际应该获取任务的详细信�?		schedulers = append(schedulers, &ScheduleInfo{
			ID:       strconv.Itoa(int(job.ID)),
			Name:     "定时任务",
			Provider: "[系统]",
			Status:   "等待",
			NextRun:  job.Next.String(),
		})
	}
	
	return schedulers
}
