package fanart

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	
	"moviepilot-go/internal/core/cache"
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/utils"
)

// FanartModule Fanart模块
type FanartModule struct {
	// 代理
	proxies map[string]string
	
	// Fanart Api
	movieURL string
	tvURL    string
}

// NewFanartModule 创建FanartModule实例
func NewFanartModule() *FanartModule {
	return &FanartModule{
		proxies:  config.Config.PROXY,
		movieURL: fmt.Sprintf("https://webservice.fanart.tv/v3/movies/%%s?api_key=%s", config.Config.FANART_API_KEY),
		tvURL:    fmt.Sprintf("https://webservice.fanart.tv/v3/tv/%%s?api_key=%s", config.Config.FANART_API_KEY),
	}
}

// InitModule 初始化模�?func (f *FanartModule) InitModule() error {
	return nil
}

// Stop 停止模块
func (f *FanartModule) Stop() {
	// 空实�?}

// Test 测试模块连接�?func (f *FanartModule) Test() (bool, string) {
	resp, err := utils.RequestUtils.GetRes("https://webservice.fanart.tv", nil, nil, 0)
	if err != nil {
		return false, "fanart网络连接失败"
	}
	
	if resp.StatusCode == 200 {
		return true, ""
	}
	
	return false, fmt.Sprintf("无法连接fanart，错误码�?d", resp.StatusCode)
}

// InitSetting 初始化设�?func (f *FanartModule) InitSetting() (string, interface{}) {
	return "FANART_API_KEY", true
}

// GetName 获取模块名称
func (f *FanartModule) GetName() string {
	return "Fanart"
}

// GetType 获取模块类型
func (f *FanartModule) GetType() models.ModuleType {
	return models.ModuleTypeOther
}

// GetSubtype 获取模块子类�?func (f *FanartModule) GetSubtype() models.OtherModulesType {
	return models.OtherModulesTypeFanart
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (f *FanartModule) GetPriority() int {
	return 0
}

// ObtainImages 获取图片
func (f *FanartModule) ObtainImages(mediainfo *models.MediaInfo) *models.MediaInfo {
	if !config.Config.FANART_ENABLE {
		return nil
	}
	
	if mediainfo.TmdbID == 0 && mediainfo.TvdbID == 0 {
		return nil
	}
	
	var result map[string]interface{}
	if mediainfo.Type == models.MediaTypeMovie {
		result = f.requestFanart(mediainfo.Type, mediainfo.TmdbID)
	} else {
		if mediainfo.TvdbID != 0 {
			result = f.requestFanart(mediainfo.Type, mediainfo.TvdbID)
		} else {
			logger.Infof("%s 没有tvdbid，无法获取fanart图片", mediainfo.TitleYear)
			return nil
		}
	}
	
	if result == nil || (result["status"] != nil && result["status"] == "error") {
		logger.Warnf("没有获取�?%s 的fanart图片数据", mediainfo.TitleYear)
		return nil
	}
	
	// 获取所有图�?	for name, images := range result {
		if images == nil {
			continue
		}
		
		// 将interface{}转换为切�?		imagesSlice, ok := images.([]interface{})
		if !ok || len(imagesSlice) == 0 {
			continue
		}
		
		// 图片属性xx_path
		imageName := f.name(name)
		if strings.HasPrefix(imageName, "season") {
			// 季图片，图片格式seasonxx-xxxx/season-specials-xxxx
			for _, imageObj := range imagesSlice {
				imageMap, ok := imageObj.(map[string]interface{})
				if !ok {
					continue
				}
				
				imageSeason := imageMap["season"]
				if imageSeason != nil {
					// 包括poster,thumb,banner
					var seasonImage string
					if imageSeason == "0" {
						seasonImage = fmt.Sprintf("season-specials-%s", imageName[6:])
					} else {
						seasonImage = fmt.Sprintf("season%s-%s", fmt.Sprintf("%02s", imageSeason), imageName[6:])
					}
					// 设置图片，没有图片才设置
					if mediainfo.GetImage(seasonImage) == "" {
						if url, ok := imageMap["url"].(string); ok {
							mediainfo.SetImage(seasonImage, url)
						}
					}
				}
			}
		} else {
			// 其他图片，优先环境变量指定语言，再like最�?			bestImage := f.pickBestImage(imagesSlice)
			// 设置图片，没有图片才设置
			if mediainfo.GetImage(imageName) == "" {
				if url, ok := bestImage["url"].(string); ok {
					mediainfo.SetImage(imageName, url)
				}
			}
		}
	}
	
	return mediainfo
}

// name 转换Fanart图片的名�?func (f *FanartModule) name(fanartName string) string {
	wordsToRemove := `tv|movie|hdmovie|hdtv|show|hd`
	re := regexp.MustCompile(wordsToRemove)
	result := re.ReplaceAllString(strings.ToLower(fanartName), "")
	return result
}

// pickBestImage 选择最佳图�?func (f *FanartModule) pickBestImage(images []interface{}) map[string]interface{} {
	// 转换为map切片以便处理
	imageMaps := make([]map[string]interface{}, 0)
	for _, img := range images {
		if imgMap, ok := img.(map[string]interface{}); ok {
			imageMaps = append(imageMaps, imgMap)
		}
	}
	
	langEnv := config.Config.FANART_LANG
	if langEnv != "" {
		langs := strings.Split(langEnv, ",")
		for i, lang := range langs {
			langs[i] = strings.TrimSpace(lang)
		}
		
		for _, lang := range langs {
			langImages := make([]map[string]interface{}, 0)
			for _, img := range imageMaps {
				if imgLang, ok := img["lang"].(string); ok && imgLang == lang {
					langImages = append(langImages, img)
				}
			}
			
			if len(langImages) > 0 {
				// 按likes排序
				sort.Slice(langImages, func(i, j int) bool {
					likesI := 0
					likesJ := 0
					
					if likes, ok := langImages[i]["likes"].(string); ok {
						fmt.Sscanf(likes, "%d", &likesI)
					}
					
					if likes, ok := langImages[j]["likes"].(string); ok {
						fmt.Sscanf(likes, "%d", &likesJ)
					}
					
					return likesI > likesJ
				})
				return langImages[0]
			}
		}
	}
	
	// 没设置或没找到，按原逻辑 zh、en、like最�?	zhImages := make([]map[string]interface{}, 0)
	enImages := make([]map[string]interface{}, 0)
	
	for _, img := range imageMaps {
		if imgLang, ok := img["lang"].(string); ok {
			if imgLang == "zh" {
				zhImages = append(zhImages, img)
			} else if imgLang == "en" {
				enImages = append(enImages, img)
			}
		}
	}
	
	if len(zhImages) > 0 {
		// 按likes排序
		sort.Slice(zhImages, func(i, j int) bool {
			likesI := 0
			likesJ := 0
			
			if likes, ok := zhImages[i]["likes"].(string); ok {
				fmt.Sscanf(likes, "%d", &likesI)
			}
			
			if likes, ok := zhImages[j]["likes"].(string); ok {
				fmt.Sscanf(likes, "%d", &likesJ)
			}
			
			return likesI > likesJ
		})
		return zhImages[0]
	}
	
	if len(enImages) > 0 {
		// 按likes排序
		sort.Slice(enImages, func(i, j int) bool {
			likesI := 0
			likesJ := 0
			
			if likes, ok := enImages[i]["likes"].(string); ok {
				fmt.Sscanf(likes, "%d", &likesI)
			}
			
			if likes, ok := enImages[j]["likes"].(string); ok {
				fmt.Sscanf(likes, "%d", &likesJ)
			}
			
			return likesI > likesJ
		})
		return enImages[0]
	}
	
	// 按likes排序
	sort.Slice(imageMaps, func(i, j int) bool {
		likesI := 0
		likesJ := 0
		
		if likes, ok := imageMaps[i]["likes"].(string); ok {
			fmt.Sscanf(likes, "%d", &likesI)
		}
		
		if likes, ok := imageMaps[j]["likes"].(string); ok {
			fmt.Sscanf(likes, "%d", &likesJ)
		}
		
		return likesI > likesJ
	})
	
	if len(imageMaps) > 0 {
		return imageMaps[0]
	}
	
	// 返回空map而不是nil避免panic
	return make(map[string]interface{})
}

// requestFanart 请求fanart图片数据
func (f *FanartModule) requestFanart(mediaType models.MediaType, queryid int) map[string]interface{} {
	var imageURL string
	if mediaType == models.MediaTypeMovie {
		imageURL = fmt.Sprintf(f.movieURL, queryid)
	} else {
		imageURL = fmt.Sprintf(f.tvURL, queryid)
	}
	
	// 使用缓存
	cacheKey := fmt.Sprintf("fanart:%s:%d", mediaType, queryid)
	
	// 尝试从缓存获�?	if cachedData, found := cache.GlobalCache.Get(cacheKey); found {
		if result, ok := cachedData.(map[string]interface{}); ok {
			return result
		}
	}
	
	resp, err := utils.RequestUtils.GetRes(imageURL, nil, f.proxies, 10)
	if err != nil {
		logger.Errorf("获取%d的Fanart图片失败�?s", queryid, err.Error())
		return nil
	}
	
	if resp == nil {
		logger.Debugf("未能获取�?%d 的Fanart图片", queryid)
		return make(map[string]interface{})
	}
	
	defer resp.Body.Close()
	
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.Debugf("获取Fanart图片失败，状态码: %d", resp.StatusCode)
		return make(map[string]interface{})
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("解析Fanart返回数据失败�?s", err.Error())
		return make(map[string]interface{})
	}
	
	// 缓存结果
	cache.GlobalCache.Set(cacheKey, result, config.Config.CONF.Meta)
	
	return result
}
