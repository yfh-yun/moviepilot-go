package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// SubscribeHelper 订阅数据统计/订阅分享�?type SubscribeHelper struct {
	// URL endpoints
	subReg             string
	subDone            string
	subReport          string
	subStatistic       string
	subShare           string
	subShares          string
	subShareStatistic  string
	subFork            string
	sharesCacheRegion  string
	adminUsers         []string

	// User info
	githubUser   *string
	shareUserID  *string
	shareUserMu  sync.RWMutex

	// Singleton instance
}

var (
	subscribeHelperInstance *SubscribeHelper
	subscribeHelperOnce     sync.Once
)

// NewSubscribeHelper 创建SubscribeHelper单例实例
func NewSubscribeHelper() *SubscribeHelper {
	subscribeHelperOnce.Do(func() {
		cfg := config.GetConfig()
		subscribeHelperInstance = &SubscribeHelper{
			subReg:            fmt.Sprintf("%s/subscribe/add", cfg.MP_SERVER_HOST),
			subDone:           fmt.Sprintf("%s/subscribe/done", cfg.MP_SERVER_HOST),
			subReport:         fmt.Sprintf("%s/subscribe/report", cfg.MP_SERVER_HOST),
			subStatistic:      fmt.Sprintf("%s/subscribe/statistic", cfg.MP_SERVER_HOST),
			subShare:          fmt.Sprintf("%s/subscribe/share", cfg.MP_SERVER_HOST),
			subShares:         fmt.Sprintf("%s/subscribe/shares", cfg.MP_SERVER_HOST),
			subShareStatistic: fmt.Sprintf("%s/subscribe/share/statistics", cfg.MP_SERVER_HOST),
			subFork:           fmt.Sprintf("%s/subscribe/fork/%%s", cfg.MP_SERVER_HOST),
			sharesCacheRegion: "subscribe_share",
			adminUsers: []string{
				"jxxghp",
				"thsrite",
				"InfinityPacer",
				"DDSRem",
				"Aqr-K",
				"Putarku",
				"4Nest",
				"xyswordzoro",
				"wikrin",
			},
		}

		// 初始化时检查是否需要上报订阅统�?		if cfg.SUBSCRIBE_STATISTIC_SHARE {
			// TODO: 实现系统配置操作和订阅操�?			// if !systemconfig.Get(SystemConfigKey.SubscribeReport) {
			// 	if subscribeHelperInstance.subReport() {
			// 		systemconfig.Set(SystemConfigKey.SubscribeReport, "1")
			// 	}
			// }
		}

		subscribeHelperInstance.getUserUUID()
		subscribeHelperInstance.getGithubUser()
	})

	return subscribeHelperInstance
}

// checkSubscribeShareEnabled 检查订阅分享功能是否开�?func (s *SubscribeHelper) checkSubscribeShareEnabled() (bool, string) {
	cfg := config.GetConfig()
	if !cfg.SUBSCRIBE_STATISTIC_SHARE {
		return false, "当前没有开启订阅数据共享功�?
	}
	return true, ""
}

// validateSubscribe 验证订阅是否存在
func (s *SubscribeHelper) validateSubscribe(subscribe *models.Subscribe) (bool, string) {
	if subscribe == nil {
		return false, "订阅不存�?
	}
	return true, ""
}

// prepareSubscribeData 准备订阅分享数据
func (s *SubscribeHelper) prepareSubscribeData(subscribe *models.Subscribe) map[string]interface{} {
	// 在Go中，我们不能直接调用to_dict()方法，需要手动构建map
	subscribeMap := map[string]interface{}{
		"id":              subscribe.ID,
		"name":            subscribe.Name,
		"year":            subscribe.Year,
		"type":            subscribe.Type,
		"keyword":         subscribe.Keyword,
		"tmdbid":          subscribe.TMDBID,
		"doubanid":        subscribe.DoubanID,
		"bangumiid":       subscribe.BangumiID,
		"mediaid":         subscribe.MediaID,
		"season":          subscribe.Season,
		"poster":          subscribe.Poster,
		"backdrop":        subscribe.Backdrop,
		"vote":            subscribe.Vote,
		"description":     subscribe.Description,
		"filter":          subscribe.Filter,
		"include":         subscribe.Include,
		"exclude":         subscribe.Exclude,
		"quality":         subscribe.Quality,
		"resolution":      subscribe.Resolution,
		"effect":          subscribe.Effect,
		"total_episode":   subscribe.TotalEpisode,
		"start_episode":   subscribe.StartEpisode,
		"lack_episode":    subscribe.LackEpisode,
		"note":            subscribe.Note,
		"state":           subscribe.State,
		"last_update":     subscribe.LastUpdate,
		"username":        subscribe.Username,
		"sites":           subscribe.Sites,
		"downloader":      subscribe.Downloader,
		"best_version":    subscribe.BestVersion,
		"current_priority": subscribe.CurrentPriority,
		"save_path":       subscribe.SavePath,
		"search_imdbid":   subscribe.SearchIMDbID,
		"date":            subscribe.Date,
		"custom_words":    subscribe.CustomWords,
		"media_category":  subscribe.MediaCategory,
		"filter_groups":   subscribe.FilterGroups,
		"episode_group":   subscribe.EpisodeGroup,
	}

	// 移除id字段
	delete(subscribeMap, "id")
	
	return subscribeMap
}

// buildSharePayload 构建分享请求载荷
func (s *SubscribeHelper) buildSharePayload(shareTitle, shareComment, shareUser string, subscribeDict map[string]interface{}) map[string]interface{} {
	s.shareUserMu.RLock()
	shareUserID := s.shareUserID
	s.shareUserMu.RUnlock()

	payload := map[string]interface{}{
		"share_title":   shareTitle,
		"share_comment": shareComment,
		"share_user":    shareUser,
	}

	if shareUserID != nil {
		payload["share_uid"] = *shareUserID
	}

	// 合并订阅数据到载荷中
	for k, v := range subscribeDict {
		payload[k] = v
	}

	return payload
}

// handleResponse 处理HTTP响应
func (s *SubscribeHelper) handleResponse(res *http.Response, clearCache bool) (bool, string) {
	if res == nil {
		return false, "连接MoviePilot服务器失�?
	}

	// 检查响应状�?	if res.StatusCode == 200 {
		// 清除缓存
		if clearCache {
			// TODO: 实现缓存清除逻辑
			// 在Go版本中，可能需要使用其他缓存机�?		}
		return true, ""
	} else {
		// 尝试解析错误消息
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return false, "无法读取响应内容"
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return false, "无法解析响应内容"
		}

		if message, ok := result["message"]; ok {
			if msgStr, ok := message.(string); ok {
				return false, msgStr
			}
		}

		return false, "未知错误"
	}
}

// handleListResponse 处理返回List的HTTP响应
func (s *SubscribeHelper) handleListResponse(res *http.Response) []map[string]interface{} {
	if res != nil && res.StatusCode == 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return []map[string]interface{}{}
		}

		var result []map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return []map[string]interface{}{}
		}

		return result
	}
	return []map[string]interface{}{}
}

// getStatistic 获取订阅统计数据
func (s *SubscribeHelper) GetStatistic(stype string, page, count int, genreID *int, minRating, maxRating *float64, sortType *string) []map[string]interface{} {
	/*
	 * 获取订阅统计数据
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return []map[string]interface{}{}
	}

	// 构建查询参数
	params := url.Values{}
	params.Set("stype", stype)
	params.Set("page", strconv.Itoa(page))
	params.Set("count", strconv.Itoa(count))

	// 添加可选参�?	if genreID != nil {
		params.Set("genre_id", strconv.Itoa(*genreID))
	}
	if minRating != nil {
		params.Set("min_rating", strconv.FormatFloat(*minRating, 'f', -1, 64))
	}
	if maxRating != nil {
		params.Set("max_rating", strconv.FormatFloat(*maxRating, 'f', -1, 64))
	}
	if sortType != nil {
		params.Set("sort_type", *sortType)
	}

	// 构建完整URL
	urlStr := fmt.Sprintf("%s?%s", s.subStatistic, params.Encode())

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}
	defer res.Body.Close()

	return s.handleListResponse(res)
}

// subReg 新增订阅统计
func (s *SubscribeHelper) SubReg(sub map[string]interface{}) bool {
	/*
	 * 新增订阅统计
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return false
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	// 序列化数�?	jsonData, err := json.Marshal(sub)
	if err != nil {
		logger.GetLoggerManager().Errorf("序列化订阅数据失�? %v", err)
		return false
	}

	req, err := http.NewRequest("POST", s.subReg, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return false
	}
	
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true
	}
	return false
}

// subDone 完成订阅统计
func (s *SubscribeHelper) SubDone(sub map[string]interface{}) bool {
	/*
	 * 完成订阅统计
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return false
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	// 序列化数�?	jsonData, err := json.Marshal(sub)
	if err != nil {
		logger.GetLoggerManager().Errorf("序列化订阅数据失�? %v", err)
		return false
	}

	req, err := http.NewRequest("POST", s.subDone, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return false
	}
	
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true
	}
	return false
}

// subRegAsync 异步新增订阅统计
func (s *SubscribeHelper) SubRegAsync(sub map[string]interface{}) bool {
	/*
	 * 异步新增订阅统计
	 */
	go s.SubReg(sub)
	return true
}

// subDoneAsync 异步完成订阅统计
func (s *SubscribeHelper) SubDoneAsync(sub map[string]interface{}) bool {
	/*
	 * 异步完成订阅统计
	 */
	go s.SubDone(sub)
	return true
}

// subReport 上报存量订阅统计
func (s *SubscribeHelper) SubReport() bool {
	/*
	 * 上报存量订阅统计
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return false
	}

	// TODO: 实现获取订阅列表的逻辑
	// subscribes := SubscribeOper().list()
	// if !subscribes {
	// 	return true
	// }

	// 构造请求数�?	// data := map[string]interface{}{
	// 	"subscribes": subscribes,
	// }

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	// req, err := http.NewRequest("POST", s.subReport, bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
	// 	return false
	// }
	
	// req.Header.Set("Content-Type", "application/json")

	// res, err := client.Do(req)
	// if err != nil {
	// 	logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
	// 	return false
	// }
	// defer res.Body.Close()

	return true
}

// SubShare 分享订阅
func (s *SubscribeHelper) SubShare(subscribeID int, shareTitle, shareComment, shareUser string) (bool, string) {
	/*
	 * 分享订阅
	 */
	// 检查功能是否开�?	enabled, message := s.checkSubscribeShareEnabled()
	if !enabled {
		return false, message
	}

	// TODO: 获取订阅信息
	// subscribe := SubscribeOper().get(subscribeID)

	// 验证订阅
	// valid, message := s.validateSubscribe(subscribe)
	// if !valid {
	// 	return false, message
	// }

	// 准备数据
	// subscribeDict := s.prepareSubscribeData(subscribe)
	// payload := s.buildSharePayload(shareTitle, shareComment, shareUser, subscribeDict)

	// 发送分享请�?	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	// jsonData, err := json.Marshal(payload)
	// if err != nil {
	// 	logger.GetLoggerManager().Errorf("序列化分享数据失�? %v", err)
	// 	return false, "数据序列化失�?
	// }

	// req, err := http.NewRequest("POST", s.subShare, bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
	// 	return false, "创建请求失败"
	// }
	
	// req.Header.Set("Content-Type", "application/json")

	// res, err := client.Do(req)
	// if err != nil {
	// 	logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
	// 	return false, "发送请求失�?
	// }
	// defer res.Body.Close()

	// return s.handleResponse(res, true)
	return true, ""
}

// ShareDelete 删除分享
func (s *SubscribeHelper) ShareDelete(shareID int) (bool, string) {
	/*
	 * 删除分享
	 */
	// 检查功能是否开�?	enabled, message := s.checkSubscribeShareEnabled()
	if !enabled {
		return false, message
	}

	s.shareUserMu.RLock()
	shareUserID := s.shareUserID
	s.shareUserMu.RUnlock()

	// 构建查询参数
	params := url.Values{}
	if shareUserID != nil {
		params.Set("share_uid", *shareUserID)
	}

	// 构建完整URL
	urlStr := fmt.Sprintf("%s/%d?%s", s.subShare, shareID, params.Encode())

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	req, err := http.NewRequest("DELETE", urlStr, nil)
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return false, "创建请求失败"
	}

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return false, "发送请求失�?
	}
	defer res.Body.Close()

	return s.handleResponse(res, true)
}

// SubFork 复用分享的订�?func (s *SubscribeHelper) SubFork(shareID int) (bool, string) {
	/*
	 * 复用分享的订�?	 */
	// 检查功能是否开�?	enabled, message := s.checkSubscribeShareEnabled()
	if !enabled {
		return false, message
	}

	// 构建完整URL
	urlStr := fmt.Sprintf(s.subFork, shareID)

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return false, "创建请求失败"
	}
	
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return false, "发送请求失�?
	}
	defer res.Body.Close()

	return s.handleResponse(res, false)
}

// GetShares 获取订阅分享数据
func (s *SubscribeHelper) GetShares(name *string, page, count int, genreID *int, minRating, maxRating *float64, sortType *string) []map[string]interface{} {
	/*
	 * 获取订阅分享数据
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return []map[string]interface{}{}
	}

	// 构建查询参数
	params := url.Values{}
	if name != nil {
		params.Set("name", *name)
	}
	params.Set("page", strconv.Itoa(page))
	params.Set("count", strconv.Itoa(count))

	// 添加可选参�?	if genreID != nil {
		params.Set("genre_id", strconv.Itoa(*genreID))
	}
	if minRating != nil {
		params.Set("min_rating", strconv.FormatFloat(*minRating, 'f', -1, 64))
	}
	if maxRating != nil {
		params.Set("max_rating", strconv.FormatFloat(*maxRating, 'f', -1, 64))
	}
	if sortType != nil {
		params.Set("sort_type", *sortType)
	}

	// 构建完整URL
	urlStr := fmt.Sprintf("%s?%s", s.subShares, params.Encode())

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}
	defer res.Body.Close()

	return s.handleListResponse(res)
}

// GetShareStatistics 获取订阅分享统计数据
func (s *SubscribeHelper) GetShareStatistics() []map[string]interface{} {
	/*
	 * 获取订阅分享统计数据
	 */
	enabled, _ := s.checkSubscribeShareEnabled()
	if !enabled {
		return []map[string]interface{}{}
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	// 设置代理
	cfg := config.GetConfig()
	if cfg.PROXY != "" {
		// TODO: 设置代理
	}

	req, err := http.NewRequest("GET", s.subShareStatistic, nil)
	if err != nil {
		logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}

	res, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
		return []map[string]interface{}{}
	}
	defer res.Body.Close()

	return s.handleListResponse(res)
}

// getUserUUID 获取用户uuid
func (s *SubscribeHelper) getUserUUID() string {
	/*
	 * 获取用户uuid
	 */
	s.shareUserMu.Lock()
	defer s.shareUserMu.Unlock()

	if s.shareUserID == nil {
		uuid := utils.GenerateUserUniqueID()
		s.shareUserID = &uuid
		logger.GetLoggerManager().Infof("当前用户UUID: %s", *s.shareUserID)
	}
	return *s.shareUserID
}

// getGithubUser 获取github用户
func (s *SubscribeHelper) getGithubUser() *string {
	/*
	 * 获取github用户
	 */
	s.shareUserMu.Lock()
	defer s.shareUserMu.Unlock()

	cfg := config.GetConfig()
	if s.githubUser == nil && len(cfg.GITHUB_HEADERS) > 0 {
		// 发送HTTP请求
		client := &http.Client{
			Timeout: 15 * time.Second,
		}
		
		// 设置代理
		if cfg.PROXY != "" {
			// TODO: 设置代理
		}

		// 设置请求�?		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			logger.GetLoggerManager().Errorf("创建HTTP请求失败: %v", err)
			return nil
		}

		// 添加请求�?		for key, value := range cfg.GITHUB_HEADERS {
			req.Header.Set(key, value)
		}

		res, err := client.Do(req)
		if err != nil {
			logger.GetLoggerManager().Errorf("发送HTTP请求失败: %v", err)
			return nil
		}
		defer res.Body.Close()

		if res.StatusCode == 200 {
			body, err := io.ReadAll(res.Body)
			if err != nil {
				logger.GetLoggerManager().Errorf("读取响应内容失败: %v", err)
				return nil
			}

			var result map[string]interface{}
			err = json.Unmarshal(body, &result)
			if err != nil {
				logger.GetLoggerManager().Errorf("解析响应内容失败: %v", err)
				return nil
			}

			if login, ok := result["login"]; ok {
				if loginStr, ok := login.(string); ok {
					s.githubUser = &loginStr
					logger.GetLoggerManager().Infof("当前Github用户: %s", *s.githubUser)
				}
			}
		}
	}
	return s.githubUser
}

// IsAdminUser 判断是否是管理员
func (s *SubscribeHelper) IsAdminUser() bool {
	/*
	 * 判断是否是管理员
	 */
	s.shareUserMu.RLock()
	githubUser := s.githubUser
	s.shareUserMu.RUnlock()

	if githubUser == nil {
		return false
	}

	for _, adminUser := range s.adminUsers {
		if *githubUser == adminUser {
			return true
		}
	}
	return false
}
