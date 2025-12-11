package system

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	nethttp "net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/business/services/search"
	"moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/internal/infrastructure/modules"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
	httpclient "moviepilot-go/pkg/utils/http"
	"moviepilot-go/pkg/utils/security"
	urlutil "moviepilot-go/pkg/utils/url"
	"moviepilot-go/pkg/version"

	"go.uber.org/zap"
)

// SystemService 系统服务
// 原SystemChain，负责系统管理
type SystemService struct {
	*base.ServiceBase
	repo          interfaces.SystemConfigRepository
	config        map[string]any
	eventManager  *events.Manager
	env           map[string]string
	searchService search.Service
	moduleManager *modules.Manager
}

// NewSystemService 创建SystemService实例
func NewSystemService(repo interfaces.SystemConfigRepository, eventManager *events.Manager) *SystemService {
	return &SystemService{
		ServiceBase:   base.NewServiceBase(),
		repo:          repo,
		config:        make(map[string]any),
		env:           make(map[string]string),
		eventManager:  eventManager,
		searchService: search.NewService(nil),
		moduleManager: modules.NewManager(logger.GetLogger()),
	}
}

// loadSystemConfig 从仓储加载系统配置到内存
func (s *SystemService) loadSystemConfig() error {
	logger.Debug("Loading system config", zap.String("func", "loadSystemConfig"))

	// 仓储未注入时直接跳过，保持向后兼容
	if s.repo == nil {
		logger.Warn("SystemConfigRepository is nil, skip loading system config")
		return nil
	}

	ctx := context.Background()
	configs, err := s.repo.List(ctx)
	if err != nil {
		logger.Error("Failed to load system configs", zap.Error(err))
		return err
	}

	for _, cfg := range configs {
		var value any = cfg.Value
		switch cfg.Type {
		case "bool":
			if v, err := strconv.ParseBool(cfg.Value); err == nil {
				value = v
			}
		case "int":
			if v, err := strconv.Atoi(cfg.Value); err == nil {
				value = v
			}
		case "json":
			var v any
			if err := json.Unmarshal([]byte(cfg.Value), &v); err == nil {
				value = v
			}
		}

		// 写入内存配置与环境映射
		s.config[cfg.Key] = value
		s.env[cfg.Key] = cfg.Value
	}

	logger.Info("System config loaded successfully", zap.Int("config_count", len(configs)))
	return nil
}

// Initialize 初始化服务
func (s *SystemService) Initialize() error {
	logger.Info("Initializing SystemService", zap.String("func", "Initialize"))

	// 从配置提供者加载系统配置（占位实现）
	if err := s.loadSystemConfig(); err != nil {
		return fmt.Errorf("failed to load system config: %w", err)
	}

	logger.Info("SystemService initialized successfully")
	return nil
}

// Name 获取服务名称
func (s *SystemService) Name() string {
	return "SystemService"
}

// Close 关闭服务
func (s *SystemService) Close() error {
	return nil
}

// GetSystemInfo 获取系统信息
func (s *SystemService) GetSystemInfo(ctx context.Context) (map[string]any, error) {
	logger.Debug("SystemService.GetSystemInfo called")

	info := map[string]any{
		"version":     version.AppVersion,
		"build_time":  time.Now().Format(time.RFC3339),
		"go_version":  runtime.Version(),
		"platform":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"environment": s.env,
	}

	logger.Info("System info retrieved successfully",
		zap.String("version", fmt.Sprintf("%v", info["version"])),
		zap.String("platform", fmt.Sprintf("%v", info["platform"])))

	return info, nil
}

// GetConfig 获取配置
func (s *SystemService) GetConfig(ctx context.Context, key string) (any, error) {
	logger.Debug("SystemService.GetConfig called", zap.String("key", key))

	// 优先从持久化仓储中获取
	if s.repo != nil {
		config, err := s.repo.Get(ctx, key)
		if err != nil {
			logger.Error("Failed to get system config", zap.String("key", key), zap.Error(err))
			return nil, err
		}
		if config == nil {
			logger.Warn("Config item not found", zap.String("key", key))
			return nil, fmt.Errorf("配置项 %s 不存在", key)
		}
		logger.Info("Config retrieved successfully", zap.String("key", key))
		return config.Value, nil
	}

	// 回退到内存配置（占位实现，便于后续平滑迁移）
	if value, exists := s.config[key]; exists {
		logger.Info("Config retrieved from memory", zap.String("key", key))
		return value, nil
	}

	logger.Warn("Config item not found in memory", zap.String("key", key))
	return nil, fmt.Errorf("配置项 %s 不存在", key)
}

// SetConfig 设置配置
func (s *SystemService) SetConfig(ctx context.Context, key string, value any) error {
	logger.Debug("SystemService.SetConfig called", zap.String("key", key), zap.Any("value", value))

	// 先写入持久化仓储
	if s.repo != nil {
		// 目前先将值序列化为字符串保存，后续可根据 Type 字段做更精细的类型处理
		strVal := fmt.Sprint(value)
		if err := s.repo.Set(ctx, key, strVal); err != nil {
			logger.Error("Failed to update system config", zap.String("key", key), zap.Error(err))
			return err
		}
	}

	// 再更新内存缓存，作为快速访问和兼容之前实现
	s.config[key] = value
	logger.Info("Config updated successfully", zap.String("key", key))

	// 发布配置变更事件
	if s.eventManager != nil {
		configData := map[string]any{
			"key":         key,
			"value":       value,
			"change_type": "updated",
		}

		// 使用 SendEvent 代替 Publish
		err := s.eventManager.SendEvent("ConfigChanged", configData)
		if err != nil {
			logger.Warn("Failed to publish ConfigChanged event",
				zap.String("key", key),
				zap.Error(err))
		} else {
			logger.Debug("ConfigChanged event published successfully",
				zap.String("key", key))
		}
	}

	return nil
}

// GetProcessInfo 获取进程信息
func (s *SystemService) GetProcessInfo(ctx context.Context) (*dto.ProcessInfo, error) {
	// TODO: 实现获取进程信息逻辑
	return nil, nil
}

// GetSchedules 获取定时任务列表
func (s *SystemService) GetSchedules(ctx context.Context) ([]*dto.ScheduleInfo, error) {
	// TODO: 实现获取定时任务逻辑
	return nil, nil
}

// RunSchedule 运行定时任务
func (s *SystemService) RunSchedule(ctx context.Context, scheduleID string) error {
	// TODO: 实现运行定时任务逻辑
	return nil
}

// CanRestart 判断当前环境是否支持内部重启
// 对应 Python SystemHelper.can_restart
func (s *SystemService) CanRestart(ctx context.Context) bool {
	logger.Debug("Checking if system can restart", zap.String("func", "CanRestart"))

	// 简化实现：返回默认值
	canRestart := false

	logger.Info("System restart capability checked",
		zap.Bool("can_restart", canRestart))

	return canRestart
}

// Restart 重启系统
// 对应 Python SystemHelper.restart：在容器环境中优雅退出，由外部重启；
// 非容器或不支持时返回错误信息
func (s *SystemService) Restart(ctx context.Context) error {
	logger.Info("System restart requested", zap.String("func", "Restart"))

	// 简化实现：直接返回错误
	logger.Warn("System restart is not supported in current environment")
	return fmt.Errorf("当前环境不支持内部重启")
}

// Shutdown 关闭系统
func (s *SystemService) Shutdown(ctx context.Context) error {
	logger.Info("System shutdown requested")
	os.Exit(0)
	return nil
}

// SetSystemModified 设置系统已修改标志
// 对应 Python SystemHelper.set_system_modified：在 Docker 环境下创建标志文件
func (s *SystemService) SetSystemModified(ctx context.Context) error {
	logger.Debug("Setting system modified flag", zap.String("func", "SetSystemModified"))

	// 简化实现：直接返回成功
	logger.Info("System modified flag functionality not supported")
	return nil
}

// IsSystemReset 检查系统是否已被重置
// 对应 Python SystemHelper.is_system_reset：Docker 环境且标志文件不存在视为已重置
func (s *SystemService) IsSystemReset(ctx context.Context) bool {
	logger.Debug("Checking if system is reset", zap.String("func", "IsSystemReset"))

	// 简化实现：返回默认值
	logger.Info("System reset check functionality not supported")
	return false
}

// GetEnvSettings 获取环境变量配置
func (s *SystemService) GetEnvSettings(ctx context.Context) (map[string]any, error) {
	// 返回非敏感的环境变量，并追加基础版本信息
	result := make(map[string]any)

	// 完整的敏感字段黑名单（与Python版本保持一致）
	sensitive := map[string]struct{}{
		"SECRET_KEY":           {},
		"RESOURCE_SECRET_KEY":  {},
		"API_TOKEN":            {},
		"TMDB_API_KEY":         {},
		"TVDB_API_KEY":         {},
		"FANART_API_KEY":       {},
		"COOKIECLOUD_KEY":      {},
		"COOKIECLOUD_PASSWORD": {},
		"GITHUB_TOKEN":         {},
		"REPO_GITHUB_TOKEN":    {},
		"U115_APP_ID":          {},
		"ALIPAN_APP_ID":        {},
		"TVDB_V4_API_KEY":      {},
		"TVDB_V4_API_PIN":      {},
	}

	// 首先从环境变量中获取所有配置
	for _, env := range os.Environ() {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			key, value := parts[0], parts[1]
			if _, skip := sensitive[key]; skip {
				continue
			}
			result[key] = value
		}
	}

	// 然后从系统配置表中获取额外配置
	if s.repo != nil {
		configs, err := s.repo.List(ctx)
		if err == nil {
			for _, config := range configs {
				if _, skip := sensitive[config.Key]; skip {
					continue
				}
				// 尝试根据Type字段解析配置值
				switch config.Type {
				case "bool":
					if val, err := strconv.ParseBool(config.Value); err == nil {
						result[config.Key] = val
					} else {
						result[config.Key] = config.Value
					}
				case "int":
					if val, err := strconv.Atoi(config.Value); err == nil {
						result[config.Key] = val
					} else {
						result[config.Key] = config.Value
					}
				case "json":
					var val any
					if err := json.Unmarshal([]byte(config.Value), &val); err == nil {
						result[config.Key] = val
					} else {
						result[config.Key] = config.Value
					}
				default:
					result[config.Key] = config.Value
				}
			}
		} else {
			logger.Error("Failed to get system configs", zap.Error(err))
		}
	}

	// 追加版本信息（占位实现，后续可接入真实版本源）
	if _, exists := result["VERSION"]; !exists {
		if ver := os.Getenv("APP_VERSION"); ver != "" {
			result["VERSION"] = ver
		} else {
			result["VERSION"] = version.AppVersion
		}
	}
	if _, exists := result["AUTH_VERSION"]; !exists {
		if ver := os.Getenv("AUTH_VERSION"); ver != "" {
			result["AUTH_VERSION"] = ver
		} else {
			result["AUTH_VERSION"] = "1.0.0" // 默认版本
		}
	}
	if _, exists := result["INDEXER_VERSION"]; !exists {
		if ver := os.Getenv("INDEXER_VERSION"); ver != "" {
			result["INDEXER_VERSION"] = ver
		} else {
			result["INDEXER_VERSION"] = "1.0.0" // 默认版本
		}
	}
	if _, exists := result["FRONTEND_VERSION"]; !exists {
		if ver := os.Getenv("FRONTEND_VERSION"); ver != "" {
			result["FRONTEND_VERSION"] = ver
		} else {
			result["FRONTEND_VERSION"] = version.FrontendVersion
		}
	}

	return result, nil
}

// GetGlobalSettings 获取非敏感系统设置视图
// 对应 Python get_global_setting：基于全局配置排除敏感字段，并追加用户相关占位信息
func (s *SystemService) GetGlobalSettings(ctx context.Context) (map[string]any, error) {
	// 先复用当前的环境配置视图
	result, err := s.GetEnvSettings(ctx)
	if err != nil {
		return nil, err
	}

	// 进一步排除更敏感的字段（参照 Python 端 exclude 列表做子集）
	extraSensitive := []string{
		"TMDB_API_KEY",
		"TVDB_API_KEY",
		"FANART_API_KEY",
		"COOKIECLOUD_KEY",
		"COOKIECLOUD_PASSWORD",
		"GITHUB_TOKEN",
		"REPO_GITHUB_TOKEN",
		"U115_APP_ID",
		"ALIPAN_APP_ID",
		"TVDB_V4_API_KEY",
		"TVDB_V4_API_PIN",
	}
	for _, k := range extraSensitive {
		delete(result, k)
	}

	// 占位的用户相关信息（后续可接入真实用户/订阅逻辑）
	if _, ok := result["USER_UNIQUE_ID"]; !ok {
		result["USER_UNIQUE_ID"] = s.getUserUUID(ctx)
	}
	if _, ok := result["SUBSCRIBE_SHARE_MANAGE"]; !ok {
		result["SUBSCRIBE_SHARE_MANAGE"] = s.isSubscribeAdmin(ctx)
	}
	if _, ok := result["WORKFLOW_SHARE_MANAGE"]; !ok {
		result["WORKFLOW_SHARE_MANAGE"] = s.isSubscribeAdmin(ctx)
	}

	return result, nil
}

// getUserUUID 获取用户唯一ID（占位实现）
func (s *SystemService) getUserUUID(ctx context.Context) string {
	// 占位实现，后续可接入真实用户逻辑
	if uuid := os.Getenv("USER_UUID"); uuid != "" {
		return uuid
	}
	return "default-user-uuid"
}

// isSubscribeAdmin 检查是否是订阅管理员（占位实现）
func (s *SystemService) isSubscribeAdmin(ctx context.Context) bool {
	// 占位实现，后续可接入真实订阅逻辑
	admin := os.Getenv("SUBSCRIBE_SHARE_ADMIN")
	return admin == "true"
}

// UpdateEnvSettings 更新环境变量配置
func (s *SystemService) UpdateEnvSettings(ctx context.Context, env map[string]any) (map[string]any, error) {
	successUpdates := make(map[string]any)
	failedUpdates := make(map[string]string)

	for k, v := range env {
		strVal, ok := v.(string)
		if !ok {
			failedUpdates[k] = "仅支持字符串类型的环境变量值"
			continue
		}
		if err := os.Setenv(k, strVal); err != nil {
			failedUpdates[k] = err.Error()
			continue
		}
		s.env[k] = strVal
		successUpdates[k] = strVal
	}

	logger.Info("Environment variables updated",
		zap.Int("success_count", len(successUpdates)),
		zap.Int("failed_count", len(failedUpdates)))

	// 发布配置变更事件（批量更新）
	if s.eventManager != nil && len(successUpdates) > 0 {
		configData := map[string]any{
			"updates":     successUpdates,
			"change_type": "batch_updated",
		}

		// 使用 SendEvent 代替 Publish
		err := s.eventManager.SendEvent("ConfigChanged", configData)
		if err != nil {
			logger.Warn("Failed to publish ConfigChanged event for batch update",
				zap.Int("update_count", len(successUpdates)),
				zap.Error(err))
		} else {
			logger.Debug("ConfigChanged event published successfully for batch update",
				zap.Int("update_count", len(successUpdates)))
		}
	}

	result := map[string]any{
		"success_updates": successUpdates,
		"failed_updates":  failedUpdates,
	}
	return result, nil
}

// GetVersion 获取版本信息
func (s *SystemService) GetVersion(ctx context.Context) (map[string]any, error) {
	version := map[string]any{
		"current":  "1.0.0",
		"latest":   "1.0.0",
		"releases": []string{"1.0.0"},
	}
	return version, nil
}

// GetReleases 获取Github上MoviePilot的所有Release版本列表
// 对应 Python /system/versions latest_version 接口
func (s *SystemService) GetReleases(ctx context.Context) (any, error) {
	apiURL := "https://api.github.com/repos/jxxghp/MoviePilot/releases"
	client := &nethttp.Client{Timeout: 10 * time.Second}
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	// 简单设置一个UA，避免部分环境下被Github拒绝
	if ua := os.Getenv("GITHUB_USER_AGENT"); ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("github releases 请求失败，状态码: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var releases any
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// TestNetwork 测试网络连通性
// 尽量对齐 Python /system/nettest 的行为（占位实现，部分字段从环境变量读取）
func (s *SystemService) TestNetwork(ctx context.Context, rawURL string, useProxy bool, include string) (map[string]any, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	// 使用URL工具包替换占位符
	testURL := urlutil.ReplaceProxyPlaceholders(rawURL)

	// 创建HTTP客户端配置
	clientConfig := httpclient.ClientConfig{
		Timeout: 10 * time.Second,
	}

	// Github 连通性测试时添加自定义头
	if regexp.MustCompile(`(?i)github`).MatchString(testURL) {
		if ua := os.Getenv("GITHUB_HEADERS_USER_AGENT"); ua != "" {
			if clientConfig.Headers == nil {
				clientConfig.Headers = make(map[string]string)
			}
			clientConfig.Headers["User-Agent"] = ua
			clientConfig.UserAgent = ua
		}
	}

	// 设置代理
	if useProxy {
		if proxyURL := os.Getenv("HTTP_PROXY"); proxyURL != "" {
			clientConfig.Proxy = proxyURL
		}
	}

	client := httpclient.NewClient(clientConfig)

	// 使用HTTP工具包测试网络
	result, err := client.TestNetwork(ctx, testURL, include)
	if err != nil {
		return map[string]any{
			"success": false,
			"message": err.Error(),
			"time":    result.ResponseTime,
		}, nil
	}

	// 构建返回结果，对齐Python格式
	if result.Success {
		return map[string]any{
			"success": true,
			"time":    result.ResponseTime,
		}, nil
	}

	// 构建错误信息
	message := result.Error
	if message == "" {
		message = fmt.Sprintf("网络测试失败，状态码：%d", result.StatusCode)
	}

	return map[string]any{
		"success": false,
		"message": message,
		"time":    result.ResponseTime,
	}, nil
}

// ReadLogFile 读取日志文件全部内容（用于 length=-1 的 text/plain 模式）
func (s *SystemService) ReadLogFile(ctx context.Context, logPath string) (string, error) {
	if !s.IsValidLogPath(logPath) {
		return "", fmt.Errorf("invalid log path")
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsValidLogPath 公开的日志路径安全校验方法
func (s *SystemService) IsValidLogPath(logPath string) bool {
	// 获取日志路径基础目录
	logPathBase := os.Getenv("LOG_PATH")
	if logPathBase == "" {
		logPathBase = "/app/logs"
	}

	return security.IsSafeLogPath(logPath, logPathBase)
}

// StreamProgress 简单的进度流（占位实现，后续可对接真实 ProgressHelper）
func (s *SystemService) StreamProgress(ctx context.Context, processType string) <-chan map[string]any {
	progressCh := make(chan map[string]any)
	go func() {
		defer close(progressCh)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		percent := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if percent < 100 {
					percent += 5
					if percent > 100 {
						percent = 100
					}
				}
				progress := map[string]any{
					"type":    processType,
					"percent": percent,
					"status":  statusFromPercent(percent),
					"message": fmt.Sprintf("%s 进度 %d%%", processType, percent),
				}
				select {
				case <-ctx.Done():
					return
				case progressCh <- progress:
				}
				if percent == 100 {
					return
				}
			}
		}
	}()
	return progressCh
}

// StreamMessages 实时获取系统消息
func (s *SystemService) StreamMessages(ctx context.Context, role string) <-chan string {
	messageCh := make(chan string)
	go func() {
		defer close(messageCh)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				message := fmt.Sprintf("[%s] 系统消息 - %s", role, time.Now().Format("2006-01-02 15:04:05"))
				select {
				case <-ctx.Done():
					return
				case messageCh <- message:
				}
			}
		}
	}()
	return messageCh
}

// TestRule 测试过滤规则
func (s *SystemService) TestRule(ctx context.Context, title, subtitle, ruleGroupName string) (map[string]any, error) {
	logger.Debug("Testing rule", zap.String("rule_group", ruleGroupName), zap.String("title", title))

	// 简化实现：返回默认值
	result := map[string]any{
		"priority":      100,
		"allowed":       false,
		"matched_rules": []string{},
		"rule_group":    ruleGroupName,
	}

	logger.Info("Rule test completed",
		zap.String("rule_group", ruleGroupName),
		zap.String("title", title),
		zap.Bool("allowed", false),
		zap.Int("priority", 100))

	return result, nil
}

// GetModuleList 获取已加载的模块ID列表
func (s *SystemService) GetModuleList(ctx context.Context) (map[string]any, error) {
	logger.Debug("Getting module list", zap.String("func", "GetModuleList"))

	// 获取所有模块
	modules := s.moduleManager.GetAll()

	// 转换为API响应格式
	moduleList := make([]map[string]any, 0, len(modules))
	for _, module := range modules {
		moduleInfo := map[string]any{
			"id":       module.GetID(),
			"name":     module.GetName(),
			"priority": module.GetPriority(),
			"status":   "running", // 默认状态
			"version":  "1.0.0",   // 默认版本
		}
		moduleList = append(moduleList, moduleInfo)
	}

	result := map[string]any{
		"modules": moduleList,
	}

	logger.Info("Module list retrieved successfully",
		zap.Int("module_count", len(moduleList)))

	return result, nil
}

// TestModule 测试模块可用性
func (s *SystemService) TestModule(ctx context.Context, moduleID string) (bool, string, error) {
	logger.Debug("Testing module", zap.String("module_id", moduleID))

	if moduleID == "" {
		return false, "模块ID不能为空", nil
	}

	// 获取模块（避免与modules包名冲突，使用mod变量名）
	mod, err := s.moduleManager.Get(moduleID)
	if err != nil {
		logger.Error("Failed to get module",
			zap.String("module_id", moduleID),
			zap.Error(err))
		return false, "模块不存在", fmt.Errorf("模块 %s 不存在！", moduleID)
	}

	// 测试模块可用性 - 简化实现
	// 因为modules.Module接口没有GetStatus()方法，所以直接返回可用
	var testResult bool
	var message string

	// 尝试调用模块的一些方法来测试可用性
	if err := mod.Initialize(); err != nil {
		testResult = false
		message = "模块初始化失败: " + err.Error()
	} else {
		testResult = true
		message = "模块可用"
	}

	logger.Info("Module test completed",
		zap.String("module_id", moduleID),
		zap.Bool("test_result", testResult),
		zap.String("message", message))

	return testResult, message, nil
}

// RunScheduler 运行定时任务
func (s *SystemService) RunScheduler(ctx context.Context, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("任务ID不能为空")
	}

	logger.Info("Running scheduled task", zap.String("job_id", jobID))

	// 占位实现，后续可对接真实的Scheduler
	// 这里应该根据jobID执行对应的任务

	return nil
}

// TestNetworkWithTimeout 测试网络连通性，支持超时设置
func (s *SystemService) TestNetworkWithTimeout(ctx context.Context, rawURL string, useProxy bool, include string, timeout time.Duration) (map[string]any, error) {
	// 使用带超时的context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.TestNetwork(ctxWithTimeout, rawURL, useProxy, include)
}

func statusFromPercent(percent int) string {
	if percent >= 100 {
		return "completed"
	}
	return "processing"
}

// ProxyImage 代理获取图片（与Python的fetch_image函数对齐）
func (s *SystemService) ProxyImage(ctx context.Context, imgURL string, proxy bool, cache bool, ifNoneMatch string, allowedDomains map[string]bool) ([]byte, string, map[string]string, int, error) {
	if imgURL == "" {
		return nil, "", nil, nethttp.StatusBadRequest, fmt.Errorf("图片URL不能为空")
	}

	// 1. 验证URL安全性
	if !s.isSafeURL(imgURL, allowedDomains) {
		logger.Warn("Blocked unsafe image URL", zap.String("url", imgURL))
		return nil, "", nil, http.StatusForbidden, fmt.Errorf("不安全的图片URL")
	}

	// 2. 创建HTTP客户端
	clientConfig := httpclient.ClientConfig{
		Timeout: 30 * time.Second,
	}

	// 3. 设置代理
	if proxy {
		if proxyURL := os.Getenv("HTTP_PROXY"); proxyURL != "" {
			clientConfig.Proxy = proxyURL
		}
	}

	client := httpclient.NewClient(clientConfig)

	// 4. 获取图片
	imgResp, err := client.GetImage(ctx, imgURL, ifNoneMatch)
	if err != nil {
		return nil, "", nil, http.StatusInternalServerError, fmt.Errorf("获取图片失败: %w", err)
	}

	// 5. 如果是304 Not Modified，直接返回
	if imgResp.StatusCode == nethttp.StatusNotModified {
		return nil, imgResp.ContentType, imgResp.Headers, nethttp.StatusNotModified, nil
	}

	// 6. 检查响应状态
	if imgResp.StatusCode != nethttp.StatusOK {
		return nil, "", nil, imgResp.StatusCode, fmt.Errorf("图片获取失败，状态码: %d", imgResp.StatusCode)
	}

	// 7. 验证图片内容有效性
	if !s.isValidImage(imgResp.Data) {
		return nil, "", nil, nethttp.StatusBadRequest, fmt.Errorf("无效的图片格式")
	}

	// 8. 保存到缓存（如果启用）
	if cache {
		cachePath := s.getImageCachePath(imgURL)
		if err := s.saveImageToCache(ctx, cachePath, imgResp.Data); err != nil {
			logger.Warn("Failed to cache image", zap.String("path", cachePath), zap.Error(err))
		}
	}

	// 9. 确保有ETag
	if imgResp.Headers["etag"] == "" {
		etag := fmt.Sprintf("%x", md5.Sum(imgResp.Data))
		if imgResp.Headers == nil {
			imgResp.Headers = make(map[string]string)
		}
		imgResp.Headers["etag"] = etag
		imgResp.Headers["cache-control"] = "public, max-age=604800" // 7天缓存
	}

	return imgResp.Data, imgResp.ContentType, imgResp.Headers, imgResp.StatusCode, nil
}

// CacheImage 缓存图片（与Python的cache_img函数对齐）
func (s *SystemService) CacheImage(ctx context.Context, imgURL string, ifNoneMatch string) ([]byte, string, map[string]string, int, error) {
	if imgURL == "" {
		return nil, "", nil, nethttp.StatusBadRequest, fmt.Errorf("图片URL不能为空")
	}

	// 如果没有启用全局图片缓存，则不使用磁盘缓存
	globalImageCache := os.Getenv("GLOBAL_IMAGE_CACHE")
	useCache := globalImageCache == "true"

	// doubanio.com 不使用代理
	useProxy := !strings.Contains(imgURL, "doubanio.com")

	return s.ProxyImage(ctx, imgURL, useProxy, useCache, ifNoneMatch, nil)
}

// isSafeURL 检查URL是否安全（与Python的SecurityUtils.is_safe_url对齐）
func (s *SystemService) isSafeURL(imgURL string, allowedDomains map[string]bool) bool {
	if imgURL == "" {
		return false
	}

	// 使用安全工具包检查图片URL
	return security.IsSafeImagePath(imgURL)
}

// getImageCachePath 获取图片缓存路径（与Python的SecurityUtils.sanitize_url_path对齐）
func (s *SystemService) getImageCachePath(imgURL string) string {
	// 解析URL
	parsedURL, err := neturl.Parse(imgURL)
	if err != nil {
		// 如果解析失败，使用简单的URL清理
		return s.sanitizeURLPath(imgURL)
	}

	// 构建缓存路径
	cachePath := parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		cachePath += "_" + parsedURL.RawQuery
	}

	// 清理路径中的特殊字符
	cachePath = s.sanitizeURLPath(cachePath)

	// 如果没有文件类型，则添加默认后缀
	ext := filepath.Ext(cachePath)
	if ext == "" {
		cachePath += ".jpg"
	}

	return cachePath
}

// sanitizeURLPath 清理URL路径（与Python的SecurityUtils.sanitize_url_path对齐）
func (s *SystemService) sanitizeURLPath(path string) string {
	// 替换特殊字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return replacer.Replace(path)
}

// getCachedImage 从缓存获取图片
func (s *SystemService) getCachedImage(ctx context.Context, cachePath string, ifNoneMatch string) ([]byte, string, map[string]string, error) {
	// 确定缓存目录
	cacheDir := os.Getenv("CACHE_PATH")
	if cacheDir == "" {
		cacheDir = "/tmp/cache"
	}

	fullPath := filepath.Join(cacheDir, "images", cachePath)

	// 检查缓存文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, "", nil, fmt.Errorf("缓存文件不存在")
	}

	// 读取缓存文件
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", nil, err
	}

	// 生成ETag
	etag := fmt.Sprintf("%x", md5.Sum(content))

	// 检查 If-None-Match
	if ifNoneMatch == etag {
		return nil, "", nil, fmt.Errorf("缓存未修改")
	}

	// 确定内容类型
	ext := filepath.Ext(cachePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// 构建响应头
	headers := map[string]string{
		"ETag":          etag,
		"Cache-Control": "public, max-age=604800", // 7天缓存
	}

	return content, contentType, headers, nil
}

// fetchRemoteImage 从远程获取图片（与Python的AsyncRequestUtils.get_res对齐）
func (s *SystemService) fetchRemoteImage(ctx context.Context, imgURL string, proxy bool) ([]byte, string, map[string]string, error) {
	// 构造请求
	req, err := nethttp.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return nil, "", nil, err
	}

	// 设置User-Agent
	userAgent := os.Getenv("NORMAL_USER_AGENT")
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	// 设置Accept头
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,*/*")

	// 设置Referer头（用于豆瓣）
	if strings.Contains(imgURL, "doubanio.com") {
		req.Header.Set("Referer", "https://movie.douban.com/")
	}

	// 构造HTTP客户端
	client := &nethttp.Client{Timeout: 10 * time.Second}

	// 设置代理
	if proxy {
		httpProxy := os.Getenv("HTTP_PROXY")
		if httpProxy != "" {
			proxyURL, err := neturl.Parse(httpProxy)
			if err == nil {
				client.Transport = &nethttp.Transport{
					Proxy: nethttp.ProxyURL(proxyURL),
				}
			}
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", nil, err
	}

	// 获取内容类型
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 从URL推断内容类型
		ext := filepath.Ext(imgURL)
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "image/jpeg"
		}
	}

	// 构建响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return content, contentType, headers, nil
}

// isValidImage 验证图片有效性（与Python的PIL.Image.verify对齐）
func (s *SystemService) isValidImage(content []byte) bool {
	// 简单实现：检查文件头是否为常见图片格式
	if len(content) < 4 {
		return false
	}

	// JPEG: FF D8 FF
	if content[0] == 0xFF && content[1] == 0xD8 && content[2] == 0xFF {
		return true
	}

	// PNG: 89 50 4E 47
	if content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 {
		return true
	}

	// GIF: 47 49 46 38
	if content[0] == 0x47 && content[1] == 0x49 && content[2] == 0x46 && content[3] == 0x38 {
		return true
	}

	// WebP: 52 49 46 46
	if content[8] == 0x57 && content[9] == 0x45 && content[10] == 0x42 && content[11] == 0x50 {
		return true
	}

	// AVIF: 基于 ftyp box 的检查
	if len(content) > 12 && content[4] == 0x66 && content[5] == 0x74 && content[6] == 0x79 && content[7] == 0x70 {
		return true
	}

	return false
}

// saveImageToCache 保存图片到缓存
func (s *SystemService) saveImageToCache(ctx context.Context, cachePath string, content []byte) error {
	// 确定缓存目录
	cacheDir := os.Getenv("CACHE_PATH")
	if cacheDir == "" {
		cacheDir = "/tmp/cache"
	}

	// 创建缓存目录
	imagesDir := filepath.Join(cacheDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return err
	}

	// 保存文件
	fullPath := filepath.Join(imagesDir, cachePath)
	return os.WriteFile(fullPath, content, 0644)
}

// GetMediaServerHosts 获取媒体服务器主机列表（与Python的MediaServerHelper对齐）
func (s *SystemService) GetMediaServerHosts(ctx context.Context) []string {
	// 占位实现，后续可接入真实媒体服务器逻辑
	// 从环境变量获取媒体服务器配置
	mediaServers := os.Getenv("MEDIA_SERVER_HOSTS")
	if mediaServers != "" {
		return strings.Split(mediaServers, ",")
	}

	// 默认空列表
	return []string{}
}
