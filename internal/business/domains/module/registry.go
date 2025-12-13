package module

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	domains_events "moviepilot-go/internal/business/domains/events"
	"moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// Registry 模块注册表
type Registry struct {
	logger     *zap.Logger
	modules    map[string]Module                 // 所有模块（已构造）
	running    map[string]Module                 // 已启用模块
	cfg        any                               // 配置对象
	configRepo interfaces.SystemConfigRepository // 配置仓库
	ctx        context.Context                   // 上下文
	eventBus   events.Bus                        // 事件总线
	cache      cache.Backend                     // 缓存后端
	cacheTTL   int64                             // 缓存过期时间（秒）
}

// NewRegistry 创建新的模块注册表
func NewRegistry(logger *zap.Logger, cfg any, configRepo interfaces.SystemConfigRepository, ctx context.Context, eventBus events.Bus, cache cache.Backend) *Registry {
	return &Registry{
		logger:     logger,
		modules:    make(map[string]Module),
		running:    make(map[string]Module),
		cfg:        cfg,
		configRepo: configRepo,
		ctx:        ctx,
		eventBus:   eventBus,
		cache:      cache,
		cacheTTL:   24 * 3600, // 默认缓存24小时
	}
}

// Register 在应用启动时由各模块调用
func (r *Registry) Register(m Module) {
	id := m.ID()
	r.modules[id] = m
	r.logger.Debug("模块已注册", zap.String("id", id), zap.String("type", string(m.Type())), zap.String("subtype", m.SubType()))
}

// LoadModules 根据配置加载模块
func (r *Registry) LoadModules() {
	// 清理旧的运行模块
	r.running = make(map[string]Module)

	// 清理模块相关缓存
	if r.cache != nil {
		if err := r.cache.Clear("module"); err != nil {
			r.logger.Error("清理模块缓存失败", zap.Error(err))
		}
	}

	for id, m := range r.modules {
		if !r.checkSetting(m) {
			r.logger.Debug("模块未启用（配置未匹配）", zap.String("id", id))
			continue
		}

		err := m.Init(r.cfg)
		if err != nil {
			r.logger.Error("初始化模块失败", zap.String("id", id), zap.Error(err))
			continue
		}

		r.running[id] = m
		r.logger.Info("模块已加载", zap.String("id", id), zap.String("type", string(m.Type())), zap.String("subtype", m.SubType()))
	}
}

// Stop 停止所有运行中的模块
func (r *Registry) Stop() {
	r.logger.Info("正在停止所有模块...")
	for id, m := range r.running {
		if err := m.Stop(); err != nil {
			r.logger.Error("停止模块失败", zap.String("id", id), zap.Error(err))
		} else {
			r.logger.Debug("模块已停止", zap.String("id", id))
		}
	}
	r.logger.Info("所有模块已停止")
}

// Reload 重载模块
func (r *Registry) Reload() {
	r.Stop()
	r.LoadModules()
	// 发送模块重载事件
	if r.eventBus != nil {
		eventType := domains_events.EventType("module.reload")
		if err := r.eventBus.PublishBroadcast(r.ctx, eventType, nil, 10); err != nil {
			r.logger.Error("发送模块重载事件失败", zap.Error(err))
		} else {
			r.logger.Debug("模块重载事件已发送")
		}
	}
	r.logger.Info("模块已重载")
}

// GetRunningModule 根据ID获取运行中的模块
func (r *Registry) GetRunningModule(id string) (Module, bool) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("running_module:%s", id)

	// 检查缓存
	if r.cache != nil {
		var cachedModule Module
		hit, err := r.cache.Get("module", cacheKey, &cachedModule)
		if err == nil && hit {
			return cachedModule, true
		}
	}

	// 缓存未命中，直接从内存获取
	m, exists := r.running[id]

	// 更新缓存
	if exists && r.cache != nil {
		r.cache.Set("module", cacheKey, m, r.cacheTTL)
	}

	return m, exists
}

// GetRunningModulesByType 根据类型获取运行中的模块
func (r *Registry) GetRunningModulesByType(t ModuleType) []Module {
	// 生成缓存键
	cacheKey := fmt.Sprintf("running_modules_by_type:%s", t)

	// 检查缓存
	if r.cache != nil {
		var cachedModules []Module
		hit, err := r.cache.Get("module", cacheKey, &cachedModules)
		if err == nil && hit {
			return cachedModules
		}
	}

	// 缓存未命中，从内存获取
	var result []Module
	for _, m := range r.running {
		if m.Type() == t {
			result = append(result, m)
		}
	}

	// 更新缓存
	if r.cache != nil {
		r.cache.Set("module", cacheKey, result, r.cacheTTL)
	}

	return result
}

// GetRunningModulesBySubType 根据子类型获取运行中的模块
func (r *Registry) GetRunningModulesBySubType(subType string) []Module {
	// 生成缓存键
	cacheKey := fmt.Sprintf("running_modules_by_subtype:%s", subType)

	// 检查缓存
	if r.cache != nil {
		var cachedModules []Module
		hit, err := r.cache.Get("module", cacheKey, &cachedModules)
		if err == nil && hit {
			return cachedModules
		}
	}

	// 缓存未命中，从内存获取
	var result []Module
	for _, m := range r.running {
		if m.SubType() == subType {
			result = append(result, m)
		}
	}

	// 更新缓存
	if r.cache != nil {
		r.cache.Set("module", cacheKey, result, r.cacheTTL)
	}

	return result
}

// GetAllModules 获取所有注册的模块
func (r *Registry) GetAllModules() map[string]Module {
	return r.modules
}

// GetRunningModules 获取所有运行中的模块
func (r *Registry) GetRunningModules() map[string]Module {
	return r.running
}

// checkSetting 检查模块配置是否匹配
func (r *Registry) checkSetting(m Module) bool {
	// 假设每个模块提供 SettingInfo() (*Setting, bool)
	if cs, ok := m.(interface{ SettingInfo() (*Setting, bool) }); ok {
		setting, exists := cs.SettingInfo()
		if !exists || setting == nil {
			return true
		}
		return r.matchConfig(setting)
	}
	return true
}

// matchConfig 匹配配置，行为与Python check_setting尽量一致
func (r *Registry) matchConfig(s *Setting) bool {
	if s == nil {
		return true
	}

	// 从配置仓库获取配置值
	config, err := r.configRepo.Get(r.ctx, s.Key)
	if err != nil {
		r.logger.Error("获取配置失败", zap.String("key", s.Key), zap.Error(err))
		return false
	}

	if config == nil {
		// 配置不存在，视为未启用
		r.logger.Debug("配置不存在", zap.String("key", s.Key))
		return false
	}

	// 配置值为空，视为未启用
	if config.Value == "" {
		r.logger.Debug("配置值为空", zap.String("key", s.Key))
		return false
	}

	// 如果Setting.Value为true，且配置值为真（非空、非false），则启用
	if s.Value == true {
		// 尝试将配置值转换为布尔值
		boolVal, err := strconv.ParseBool(config.Value)
		if err == nil {
			return boolVal
		}
		// 如果转换失败，非空字符串视为true
		return config.Value != ""
	}

	// 如果Setting.Value是字符串，检查是否在配置值中
	if strVal, ok := s.Value.(string); ok {
		// 支持多种分隔符：空格、逗号、分号
		configValues := []string{config.Value}
		for _, sep := range []string{",", ";", " "} {
			if strings.Contains(config.Value, sep) {
				configValues = strings.Split(config.Value, sep)
				break
			}
		}

		// 检查是否包含指定值
		for _, val := range configValues {
			val = strings.TrimSpace(val)
			if val == strVal {
				return true
			}
		}
	}

	// 其他情况，检查是否相等
	return fmt.Sprintf("%v", s.Value) == config.Value
}

// GetModuleCount 获取运行中的模块数量
func (r *Registry) GetModuleCount() int {
	return len(r.running)
}
