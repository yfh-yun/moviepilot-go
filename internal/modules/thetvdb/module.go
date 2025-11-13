package thetvdb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/config"
	"moviepilot-go/pkg/models"
	"moviepilot-go/internal/logger"
)

// TheTvDbModule TVDB媒体信息匹配模块
type TheTvDbModule struct {
	timeout     int
	tvdb        *TVDB
	authLock    sync.Mutex
}

// NewTheTvDbModule 创建TheTvDbModule实例
func NewTheTvDbModule() *TheTvDbModule {
	return &TheTvDbModule{
		timeout: 15,
	}
}

// InitModule 初始化模�?func (t *TheTvDbModule) InitModule() error {
	return nil
}

// initializeTvdbSession 创建或刷新TVDB登录会话
func (t *TheTvDbModule) initializeTvdbSession(isRetry bool) error {
	action := "创建"
	if isRetry {
		action = "刷新"
	}
	
	logger.Info(fmt.Sprintf("开�?sTVDB登录会话...", action))
	
	if config.Config.TVDB_V4_API_KEY == "" {
		t.tvdb = nil
		return fmt.Errorf("TVDB API Key 未配置，无法初始化会�?)
	}
	
	proxy := make(map[string]string)
	if config.Config.PROXY != "" {
		proxy["http"] = config.Config.PROXY
		proxy["https"] = config.Config.PROXY
	}
	
	tvdb, err := NewTVDB(config.Config.TVDB_V4_API_KEY, config.Config.TVDB_V4_API_PIN, proxy, t.timeout)
	if err != nil {
		t.tvdb = nil
		return fmt.Errorf("TVDB登录会话%s失败: %v", action, err)
	}
	
	t.tvdb = tvdb
	logger.Info(fmt.Sprintf("TVDB登录会话%s成功", action))
	return nil
}

// ensureTvdbSession 确保TVDB会话存在
func (t *TheTvDbModule) ensureTvdbSession(isRetry bool) error {
	// 第一次检�?无锁)，提高性能，避免不必要锁竞�?	if t.tvdb == nil || isRetry {
		t.authLock.Lock()
		defer t.authLock.Unlock()
		// 第二次检�?有锁)，防止多个线程都通过第一次检查后重复初始�?		if t.tvdb == nil || isRetry {
			return t.initializeTvdbSession(isRetry)
		}
	}
	return nil
}

// handleTvdbCall 包裹TVDB调用，处理token失效情况并尝试重新初始化
func (t *TheTvDbModule) handleTvdbCall(method string, args ...interface{}) (interface{}, error) {
	err := t.ensureTvdbSession(false)
	if err != nil {
		return nil, err
	}
	
	switch method {
	case "get_series":
		if len(args) < 1 {
			return nil, fmt.Errorf("参数不足")
		}
		
		id, ok := args[0].(int)
		if !ok {
			return nil, fmt.Errorf("参数类型错误")
		}
		
		var meta *string = nil
		if len(args) > 1 && args[1] != nil {
			if m, ok := args[1].(string); ok {
				meta = &m
			}
		}
		
		result, err := t.tvdb.GetSeries(id, meta, nil)
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "Unauthorized") || strings.Contains(fmt.Sprintf("%v", err), "401")) {
			logger.Warning("TVDB Token 可能已失效，正在尝试重新登录...")
			retryErr := t.ensureTvdbSession(true)
			if retryErr != nil {
				logger.Error(fmt.Sprintf("TVDB Token失效后重新登录失�? %v", retryErr))
				return nil, retryErr
			}
			
			result, err = t.tvdb.GetSeries(id, meta, nil)
		}
		
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "NotFoundException") || strings.Contains(fmt.Sprintf("%v", err), "ID not found")) {
			logger.Warning(fmt.Sprintf("TVDB 资源未找�?(调用 %s): %v", method, err))
			return nil, nil
		}
		
		return result, err
		
	case "get_series_extended":
		if len(args) < 1 {
			return nil, fmt.Errorf("参数不足")
		}
		
		id, ok := args[0].(int)
		if !ok {
			return nil, fmt.Errorf("参数类型错误")
		}
		
		var meta *string = nil
		if len(args) > 1 && args[1] != nil {
			if m, ok := args[1].(string); ok {
				meta = &m
			}
		}
		
		short := false
		if len(args) > 2 && args[2] != nil {
			if s, ok := args[2].(bool); ok {
				short = s
			}
		}
		
		result, err := t.tvdb.GetSeriesExtended(id, meta, short, nil)
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "Unauthorized") || strings.Contains(fmt.Sprintf("%v", err), "401")) {
			logger.Warning("TVDB Token 可能已失效，正在尝试重新登录...")
			retryErr := t.ensureTvdbSession(true)
			if retryErr != nil {
				logger.Error(fmt.Sprintf("TVDB Token失效后重新登录失�? %v", retryErr))
				return nil, retryErr
			}
			
			result, err = t.tvdb.GetSeriesExtended(id, meta, short, nil)
		}
		
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "NotFoundException") || strings.Contains(fmt.Sprintf("%v", err), "ID not found")) {
			logger.Warning(fmt.Sprintf("TVDB 资源未找�?(调用 %s): %v", method, err))
			return nil, nil
		}
		
		return result, err
		
	case "search":
		if len(args) < 1 {
			return nil, fmt.Errorf("参数不足")
		}
		
		query, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("参数类型错误")
		}
		
		params := make(map[string]interface{})
		if len(args) > 1 && args[1] != nil {
			if p, ok := args[1].(map[string]interface{}); ok {
				params = p
			}
		}
		
		result, err := t.tvdb.Search(query, params)
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "Unauthorized") || strings.Contains(fmt.Sprintf("%v", err), "401")) {
			logger.Warning("TVDB Token 可能已失效，正在尝试重新登录...")
			retryErr := t.ensureTvdbSession(true)
			if retryErr != nil {
				logger.Error(fmt.Sprintf("TVDB Token失效后重新登录失�? %v", retryErr))
				return nil, retryErr
			}
			
			result, err = t.tvdb.Search(query, params)
		}
		
		if err != nil && (strings.Contains(fmt.Sprintf("%v", err), "NotFoundException") || strings.Contains(fmt.Sprintf("%v", err), "ID not found")) {
			logger.Warning(fmt.Sprintf("TVDB 资源未找�?(调用 %s): %v", method, err))
			return nil, nil
		}
		
		return result, err
		
	default:
		return nil, fmt.Errorf("不支持的方法: %s", method)
	}
}

// GetName 获取模块名称
func (t *TheTvDbModule) GetName() string {
	return "TheTvDb"
}

// GetType 获取模块类型
func (t *TheTvDbModule) GetType() models.ModuleType {
	return models.MediaRecognize
}

// GetSubtype 获取模块子类�?func (t *TheTvDbModule) GetSubtype() models.MediaRecognizeType {
	return models.TVDB
}

// GetPriority 获取模块优先�?func (t *TheTvDbModule) GetPriority() int {
	return 4
}

// Stop 停止模块
func (t *TheTvDbModule) Stop() {
	logger.Info("TheTvDbModule 停止。正在清�?TVDB 会话�?)
	t.authLock.Lock()
	defer t.authLock.Unlock()
	t.tvdb = nil
}

// Test 测试模块连接�?func (t *TheTvDbModule) Test() (bool, string) {
	_, err := t.handleTvdbCall("get_series", 81189)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

// InitSetting 初始化设�?func (t *TheTvDbModule) InitSetting() (string, interface{}) {
	return "", nil
}

// TvdbInfo 获取TVDB信息
func (t *TheTvDbModule) TvdbInfo(tvdbid int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB信息: %d ...", tvdbid))
	result, err := t.handleTvdbCall("get_series_extended", tvdbid)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// SearchTvdb 用标题搜索TVDB剧集
func (t *TheTvDbModule) SearchTvdb(title string) []map[string]interface{} {
	logger.Info(fmt.Sprintf("开始用标题搜索TVDB剧集: %s ...", title))
	result, err := t.handleTvdbCall("search", title)
	if err != nil {
		logger.Error(fmt.Sprintf("用标题搜索TVDB剧集失败 (%s): %v", title, err))
		return []map[string]interface{}{}
	}
	
	if result == nil {
		return []map[string]interface{}{}
	}
	
	if res, ok := result.([]interface{}); ok {
		var seriesList []map[string]interface{}
		for _, item := range res {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, exists := itemMap["type"]; exists && itemType == "series" {
					seriesList = append(seriesList, itemMap)
				}
			}
		}
		return seriesList
	}
	
	logger.Warning(fmt.Sprintf("TVDB 搜索 '%s' 未返回列表：%T", title, result))
	return []map[string]interface{}{}
}

// SearchTvdbByRemoteID 通过外部ID搜索TVDB剧集
func (t *TheTvDbModule) SearchTvdbByRemoteID(remoteid string) []map[string]interface{} {
	logger.Info(fmt.Sprintf("开始用外部ID搜索TVDB剧集: %s ...", remoteid))
	result, err := t.handleTvdbCall("search_by_remote_id", remoteid)
	if err != nil {
		logger.Error(fmt.Sprintf("用外部ID搜索TVDB剧集失败 (%s): %v", remoteid, err))
		return []map[string]interface{}{}
	}
	
	if result == nil {
		return []map[string]interface{}{}
	}
	
	if res, ok := result.([]interface{}); ok {
		var seriesList []map[string]interface{}
		for _, item := range res {
			if itemMap, ok := item.(map[string]interface{}); ok {
				seriesList = append(seriesList, itemMap)
			}
		}
		return seriesList
	}
	
	logger.Warning(fmt.Sprintf("TVDB 搜索 '%s' 未返回列表：%T", remoteid, result))
	return []map[string]interface{}{}
}

// GetSeriesBySlug 通过slug获取剧集信息
func (t *TheTvDbModule) GetSeriesBySlug(slug string) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB剧集信息: %s ...", slug))
	result, err := t.handleTvdbCall("get_series_by_slug", slug)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB剧集信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetMovie 获取电影信息
func (t *TheTvDbModule) GetMovie(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB电影信息: %d ...", id))
	result, err := t.handleTvdbCall("get_movie", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB电影信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetMovieExtended 获取电影扩展信息
func (t *TheTvDbModule) GetMovieExtended(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB电影扩展信息: %d ...", id))
	result, err := t.handleTvdbCall("get_movie_extended", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB电影扩展信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetSeason 获取季信�?func (t *TheTvDbModule) GetSeason(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB季信�? %d ...", id))
	result, err := t.handleTvdbCall("get_season", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB季信息失�? %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetSeasonExtended 获取季扩展信�?func (t *TheTvDbModule) GetSeasonExtended(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB季扩展信�? %d ...", id))
	result, err := t.handleTvdbCall("get_season_extended", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB季扩展信息失�? %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetEpisode 获取集信�?func (t *TheTvDbModule) GetEpisode(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB集信�? %d ...", id))
	result, err := t.handleTvdbCall("get_episode", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB集信息失�? %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetEpisodeExtended 获取集扩展信�?func (t *TheTvDbModule) GetEpisodeExtended(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB集扩展信�? %d ...", id))
	result, err := t.handleTvdbCall("get_episode_extended", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB集扩展信息失�? %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetPerson 获取人物信息
func (t *TheTvDbModule) GetPerson(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB人物信息: %d ...", id))
	result, err := t.handleTvdbCall("get_person", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB人物信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}

// GetPersonExtended 获取人物扩展信息
func (t *TheTvDbModule) GetPersonExtended(id int) map[string]interface{} {
	logger.Info(fmt.Sprintf("开始获取TVDB人物扩展信息: %d ...", id))
	result, err := t.handleTvdbCall("get_person_extended", id)
	if err != nil {
		logger.Error(fmt.Sprintf("获取TVDB人物扩展信息失败: %v", err))
		return nil
	}
	
	if result == nil {
		return nil
	}
	
	if res, ok := result.(map[string]interface{}); ok {
		return res
	}
	
	return nil
}
