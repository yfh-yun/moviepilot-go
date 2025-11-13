package douban

import (
	"fmt"
	"regexp"
	"strings"
	
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/common"
	"moviepilot-go/internal/utils/http"
)

// DoubanModule 豆瓣模块
type DoubanModule struct {
	doubanAPI *DoubanApi
	scraper   *DoubanScraper
	cache     *DoubanCache
}

// NewDoubanModule 创建DoubanModule实例
func NewDoubanModule() *DoubanModule {
	return &DoubanModule{}
}

// InitModule 初始化模�?func (dm *DoubanModule) InitModule() error {
	dm.doubanAPI = NewDoubanApi()
	dm.scraper = NewDoubanScraper()
	dm.cache = NewDoubanCache()
	return nil
}

// Stop 停止模块
func (dm *DoubanModule) Stop() {
	dm.doubanAPI.Close()
}

// Test 测试模块连接�?func (dm *DoubanModule) Test() (bool, string) {
	resp := http.RequestUtils.GetRes("https://movie.douban.com/", nil, nil, 0)
	if resp == nil {
		return false, "豆瓣网络连接失败"
	}
	return true, ""
}

// InitSetting 初始化设�?func (dm *DoubanModule) InitSetting() (string, interface{}) {
	return "", nil
}

// GetName 获取模块名称
func (dm *DoubanModule) GetName() string {
	return "豆瓣"
}

// GetType 获取模块类型
func (dm *DoubanModule) GetType() types.ModuleType {
	return types.ModuleTypeMediaRecognize
}

// GetSubtype 获取模块子类�?func (dm *DoubanModule) GetSubtype() types.MediaRecognizeType {
	return types.MediaRecognizeTypeDouban
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (dm *DoubanModule) GetPriority() int {
	return 2
}

// recognizeMediaCore 识别媒体信息的核心逻辑
func (dm *DoubanModule) recognizeMediaCore(meta *models.MetaBase, mtype types.MediaType, 
	doubanid *string, cache bool, doubanInfoFunc func(string, types.MediaType) map[string]interface{},
	matchDoubaninfoFunc func(string, types.MediaType, string, *int) map[string]interface{}) *models.MediaInfo {
	
	if doubanid == nil && meta == nil {
		return nil
	}
	
	if meta != nil && doubanid == nil && config.Config.RECOGNIZE_SOURCE != "douban" {
		return nil
	}
	
	var cacheInfo *CacheData
	if meta == nil {
		// 未提供元数据时，直接查询豆瓣信息，不使用缓存
		cacheInfo = &CacheData{}
	} else if meta.Name == "" {
		logger.Error("识别媒体信息时未提供元数据名�?)
		return nil
	} else {
		// 读取缓存
		if mtype != "" {
			meta.Type = mtype
		}
		if doubanid != nil {
			meta.DoubanID = *doubanid
		}
		cacheInfo = dm.cache.Get(meta)
	}
	
	// 识别豆瓣信息
	var info map[string]interface{}
	if cacheInfo.ID == "" || !cache {
		// 缓存没有或者强制不使用缓存
		if doubanid != nil {
			// 直接查询详情
			info = doubanInfoFunc(*doubanid, mtype)
		} else if meta != nil {
			info = make(map[string]interface{})
			// 简体名�?			zhName := meta.CnName // TODO: 实现繁简转换
			// 使用中英文名分别识别，去重去空，但要保持顺序
			names := []string{}
			if meta.CnName != "" {
				names = append(names, meta.CnName)
			}
			if zhName != "" && zhName != meta.CnName {
				names = append(names, zhName)
			}
			if meta.EnName != "" {
				names = append(names, meta.EnName)
			}
			
			// 去重
			uniqueNames := []string{}
			seen := make(map[string]bool)
			for _, name := range names {
				if name != "" && !seen[name] {
					seen[name] = true
					uniqueNames = append(uniqueNames, name)
				}
			}
			
			for _, name := range uniqueNames {
				if meta.BeginSeason != "" {
					logger.Info(fmt.Sprintf("正在识别 %s �?s�?...", name, meta.BeginSeason))
				} else {
					logger.Info(fmt.Sprintf("正在识别 %s ...", name))
				}
				
				// 匹配豆瓣信息
				var beginSeasonInt *int
				if meta.BeginSeason != "" {
					if season, err := common.StringToInt(meta.BeginSeason); err == nil {
						beginSeasonInt = &season
					}
				}
				
				matchInfo := matchDoubaninfoFunc(name, mtype, meta.Year, beginSeasonInt)
				if len(matchInfo) > 0 {
					// 匹配到豆瓣信�?					info = doubanInfoFunc(matchInfo["id"].(string), mtype)
					if len(info) > 0 {
						break
					}
				}
			}
		} else {
			logger.Error("识别媒体信息时未提供元数据或豆瓣ID")
			return nil
		}
		
		// 保存到缓�?		if meta != nil && cache {
			// 这里需要将map[string]interface{}转换为合适的格式
			dm.cache.Update(meta, info)
		}
	} else {
		// 使用缓存信息
		if cacheInfo.Title != "" {
			logger.Info(fmt.Sprintf("%s 使用豆瓣识别缓存�?s", meta.Name, cacheInfo.Title))
			info = doubanInfoFunc(cacheInfo.ID, cacheInfo.Type)
		} else {
			logger.Info(fmt.Sprintf("%s 使用豆瓣识别缓存：无法识�?, meta.Name))
			info = nil
		}
	}
	
	if len(info) > 0 {
		// 赋值TMDB信息并返�?		mediainfo := models.NewMediaInfoFromDouban(info)
		if meta != nil {
			logger.Info(fmt.Sprintf("%s 豆瓣识别结果�?s %s %s", 
				meta.Name, mediainfo.Type, mediainfo.TitleYear, mediainfo.DoubanID))
		} else {
			logger.Info(fmt.Sprintf("%s 豆瓣识别结果�?s %s", 
				*doubanid, mediainfo.Type, mediainfo.TitleYear))
		}
		return mediainfo
	} else {
		if meta != nil {
			logger.Info(fmt.Sprintf("%s 未匹配到豆瓣媒体信息", meta.Name))
		} else {
			logger.Info(fmt.Sprintf("%s 未匹配到豆瓣媒体信息", *doubanid))
		}
	}
	
	return nil
}

// RecognizeMedia 识别媒体信息
func (dm *DoubanModule) RecognizeMedia(meta *models.MetaBase, mtype types.MediaType, 
	doubanid *string, cache bool, kwargs map[string]interface{}) *models.MediaInfo {
	
	return dm.recognizeMediaCore(
		meta,
		mtype,
		doubanid,
		cache,
		dm.doubanInfo,
		dm.matchDoubaninfo,
	)
}

// doubanInfo 获取豆瓣信息
func (dm *DoubanModule) doubanInfo(doubanid string, mtype types.MediaType) map[string]interface{} {
	
	doubanTv := func() map[string]interface{} {
		info := dm.doubanAPI.TvDetail(doubanid)
		if len(info) > 0 {
			if msg, ok := info["msg"].(string); ok && strings.Contains(msg, "subject_ip_rate_limit") {
				msg := fmt.Sprintf("触发豆瓣IP速率限制，错误信息：%v ...", info)
				logger.Warn(msg)
				// TODO: 抛出APIRateLimitException异常
			}
			celebrities := dm.doubanAPI.TvCelebrities(doubanid)
			if len(celebrities) > 0 {
				info["directors"] = celebrities["directors"]
				info["actors"] = celebrities["actors"]
			}
		}
		return info
	}
	
	doubanMovie := func() map[string]interface{} {
		info := dm.doubanAPI.MovieDetail(doubanid)
		if len(info) > 0 {
			if msg, ok := info["msg"].(string); ok && strings.Contains(msg, "subject_ip_rate_limit") {
				msg := fmt.Sprintf("触发豆瓣IP速率限制，错误信息：%v ...", info)
				logger.Warn(msg)
				// TODO: 抛出APIRateLimitException异常
			}
			celebrities := dm.doubanAPI.MovieCelebrities(doubanid)
			if len(celebrities) > 0 {
				info["directors"] = celebrities["directors"]
				info["actors"] = celebrities["actors"]
			}
		}
		return info
	}
	
	if doubanid == "" {
		return nil
	}
	
	logger.Info(fmt.Sprintf("开始获取豆瓣信息：%s ...", doubanid))
	
	if mtype == types.MediaTypeTV {
		return doubanTv()
	} else if mtype == types.MediaTypeMovie {
		return doubanMovie()
	} else {
		movieResult := doubanMovie()
		if len(movieResult) > 0 {
			return movieResult
		}
		return doubanTv()
	}
}

// processImdbidResult 处理IMDBID查询结果
func (dm *DoubanModule) processImdbidResult(result map[string]interface{}, imdbid string) map[string]interface{} {
	if len(result) > 0 {
		doubanid, _ := result["id"].(string)
		if doubanid != "" && !regexp.MustCompile(`^\d+$`).MatchString(doubanid) {
			if matches := regexp.MustCompile(`\d+`).FindStringSubmatch(doubanid); len(matches) > 0 {
				result["id"] = matches[0]
			}
		}
		logger.Info(fmt.Sprintf("%s 查询到豆瓣信息：%s", imdbid, result["title"]))
		return result
	}
	return nil
}

// processSearchResults 处理搜索结果并进行匹�?func (dm *DoubanModule) processSearchResults(result map[string]interface{}, name string, 
	mtype types.MediaType, year string, season *int) map[string]interface{} {
	
	if len(result) == 0 {
		logger.Warn(fmt.Sprintf("未找�?%s 的豆瓣信�?, name))
		return make(map[string]interface{})
	}
	
	// 触发rate limit检�?	if values, ok := result["values"].(map[string]interface{}); ok {
		for _, v := range values {
			if str, ok := v.(string); ok && strings.Contains(str, "search_access_rate_limit") {
				msg := fmt.Sprintf("触发豆瓣API速率限制，错误信息：%v ...", result)
				logger.Warn(msg)
				// TODO: 抛出APIRateLimitException异常
				break
			}
		}
	}
	
	if items, ok := result["items"].([]interface{}); !ok || len(items) == 0 {
		logger.Warn(fmt.Sprintf("未找�?%s 的豆瓣信�?, name))
		return make(map[string]interface{})
	}
	
	items := result["items"].([]interface{})
	for _, itemObj := range items {
		if itemMap, ok := itemObj.(map[string]interface{}); ok {
			typeName, _ := itemMap["type_name"].(string)
			if typeName != string(types.MediaTypeTV) && typeName != string(types.MediaTypeMovie) {
				continue
			}
			if mtype != "" && mtype != types.MediaTypeUnknown && string(mtype) != typeName {
				continue
			}
			if mtype == types.MediaTypeTV && season == nil {
				s := 1
				season = &s
			}
			
			if target, ok := itemMap["target"].(map[string]interface{}); ok {
				title, _ := target["title"].(string)
				if title == "" {
					continue
				}
				
				// 创建MetaInfo来解析标�?				meta := models.NewMetaInfo(title)
				if typeName == string(types.MediaTypeTV) {
					meta.Type = types.MediaTypeTV
					if meta.BeginSeason == "" {
						meta.BeginSeason = "1"
					}
				}
				
				// 检查匹配条�?				match := meta.Name == name
				if season != nil {
					match = match && meta.BeginSeason == fmt.Sprintf("%d", *season)
				} else {
					match = match && meta.BeginSeason == ""
				}
				if year != "" {
					if itemYear, ok := target["year"].(string); ok {
						match = match && itemYear == year
					}
				}
				
				if match {
					logger.Info(fmt.Sprintf("%s 匹配到豆瓣信息：%s %s", name, target["id"], target["title"]))
					return target
				}
			}
		}
	}
	
	return make(map[string]interface{})
}

// matchDoubaninfo 搜索和匹配豆瓣信�?func (dm *DoubanModule) matchDoubaninfo(name string, imdbid string, mtype types.MediaType, 
	year string, season *int, raiseException bool) map[string]interface{} {
	
	if imdbid != "" {
		// 优先使用IMDBID查询
		logger.Info(fmt.Sprintf("开始使用IMDBID %s 查询豆瓣信息 ...", imdbid))
		result := dm.doubanAPI.Imdbid(imdbid, "")
		processedResult := dm.processImdbidResult(result, imdbid)
		if len(processedResult) > 0 {
			return processedResult
		}
	}
	
	// 搜索
	searchTerm := name
	if year != "" {
		searchTerm = name + " " + year
	}
	logger.Info(fmt.Sprintf("开始使用名�?%s 匹配豆瓣信息 ...", name))
	result := dm.doubanAPI.Search(strings.TrimSpace(searchTerm), 0, 20, "")
	return dm.processSearchResults(result, name, mtype, year, season)
}

// DoubanDiscover 发现豆瓣电影、剧�?func (dm *DoubanModule) DoubanDiscover(mtype types.MediaType, sort, tags string, 
	page, count int) []*models.MediaInfo {
	
	logger.Info(fmt.Sprintf("开始发现豆�?%s ...", mtype))
	var infos map[string]interface{}
	
	if mtype == types.MediaTypeMovie {
		infos = dm.doubanAPI.MovieRecommend(tags, sort, (page-1)*count, count, "")
	} else {
		infos = dm.doubanAPI.TvRecommend(tags, sort, (page-1)*count, count, "")
	}
	
	if infos != nil {
		if items, ok := infos["items"].([]interface{}); ok {
			medias := []*models.MediaInfo{}
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					media := models.NewMediaInfoFromDouban(itemMap)
					// 过滤无效海报
					if media.PosterPath != "" && 
						!strings.Contains(media.PosterPath, "movie_large.jpg") &&
						!strings.Contains(media.PosterPath, "tv_normal.png") &&
						!strings.Contains(media.PosterPath, "tv_normal.jpg") &&
						!strings.Contains(media.PosterPath, "tv_large.jpg") {
						medias = append(medias, media)
					}
				}
			}
			return medias
		}
	}
	
	return []*models.MediaInfo{}
}

// SearchMedias 搜索媒体信息
func (dm *DoubanModule) SearchMedias(meta *models.MetaBase) []*models.MediaInfo {
	if config.Config.SEARCH_SOURCE != "" && !strings.Contains(config.Config.SEARCH_SOURCE, "douban") {
		return nil
	}
	
	if meta.Name == "" {
		return []*models.MediaInfo{}
	}
	
	result := dm.doubanAPI.Search(meta.Name, 0, 20, "")
	if result == nil {
		return []*models.MediaInfo{}
	}
	
	if items, ok := result["items"].([]interface{}); !ok || len(items) == 0 {
		return []*models.MediaInfo{}
	}
	
	// 返回数据
	retMedias := []*models.MediaInfo{}
	items := result["items"].([]interface{})
	
	for _, itemObj := range items {
		if itemMap, ok := itemObj.(map[string]interface{}); ok {
			typeName, _ := itemMap["type_name"].(string)
			if meta.Type != types.MediaTypeUnknown && string(meta.Type) != typeName {
				continue
			}
			if typeName != string(types.MediaTypeTV) && typeName != string(types.MediaTypeMovie) {
				continue
			}
			
			if target, ok := itemMap["target"].(map[string]interface{}); ok {
				title, _ := target["title"].(string)
				if !strings.Contains(title, meta.Name) {
					continue
				}
				retMedias = append(retMedias, models.NewMediaInfoFromDouban(target))
			}
		}
	}
	
	// 将搜索词中的季写入标题中
	if len(retMedias) > 0 && meta.BeginSeason != "" {
		// TODO: 小写数据转大�?		seasonStr := meta.BeginSeason // 简化处�?		for _, media := range retMedias {
			if media.Type == types.MediaTypeTV {
				media.Title = fmt.Sprintf("%s �?s�?, media.Title, seasonStr)
				media.Season = meta.BeginSeason
			}
		}
	}
	
	return retMedias
}

// MetadataNfo 获取NFO文件内容文本
func (dm *DoubanModule) MetadataNfo(mediainfo *models.MediaInfo, season *int, kwargs map[string]interface{}) string {
	if config.Config.SCRAP_SOURCE != "douban" {
		return ""
	}
	return dm.scraper.GetMetadataNfo(mediainfo, season)
}

// MetadataImg 获取图片名称和url
func (dm *DoubanModule) MetadataImg(mediainfo *models.MediaInfo, season, episode *int) map[string]string {
	if config.Config.SCRAP_SOURCE != "douban" {
		return nil
	}
	return dm.scraper.GetMetadataImg(mediainfo, season, episode)
}

// validateDoubanObtainImagesParams 验证豆瓣 obtain_images 参数
func (dm *DoubanModule) validateDoubanObtainImagesParams(mediainfo *models.MediaInfo) *models.MediaInfo {
	if config.Config.RECOGNIZE_SOURCE != "douban" {
		return nil
	}
	if mediainfo.DoubanID == "" {
		return nil
	}
	if mediainfo.BackdropPath != "" {
		// 没有图片缺失
		return mediainfo
	}
	return nil
}

// processDoubanImages 处理豆瓣图片数据
func (dm *DoubanModule) processDoubanImages(mediainfo *models.MediaInfo, info map[string]interface{}) *models.MediaInfo {
	if len(info) == 0 {
		return mediainfo
	}
	
	if photos, ok := info["photos"].([]interface{}); ok && len(photos) > 0 {
		if photo, ok := photos[0].(map[string]interface{}); ok {
			if image, ok := photo["image"].(map[string]interface{}); ok {
				if large, ok := image["large"].(map[string]interface{}); ok {
					if url, ok := large["url"].(string); ok && url != "" {
						mediainfo.BackdropPath = url
					}
				}
			}
		}
	}
	
	return mediainfo
}

// ObtainImages 补充抓取媒体信息图片
func (dm *DoubanModule) ObtainImages(mediainfo *models.MediaInfo) *models.MediaInfo {
	// 验证参数
	result := dm.validateDoubanObtainImagesParams(mediainfo)
	if result != nil {
		return result
	}
	
	// 调用图片接口
	var info map[string]interface{}
	if mediainfo.Type == types.MediaTypeMovie {
		info = dm.doubanAPI.MoviePhotos(mediainfo.DoubanID, 0, 20, "")
	} else {
		info = dm.doubanAPI.TvPhotos(mediainfo.DoubanID, 0, 20, "")
	}
	
	// 处理图片数据
	return dm.processDoubanImages(mediainfo, info)
}

// ClearCache 清除缓存
func (dm *DoubanModule) ClearCache() {
	logger.Info("开始清除豆瓣缓�?...")
	dm.doubanAPI.ClearCache()
	dm.cache.Clear()
	logger.Info("豆瓣缓存清除完成")
}
