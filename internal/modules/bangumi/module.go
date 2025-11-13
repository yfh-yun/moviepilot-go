package bangumi

import (
	"fmt"
	"strings"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/http"
)

// BangumiModule Bangumi模块
type BangumiModule struct {
	bangumiAPI *BangumiApi
}

// NewBangumiModule 创建BangumiModule实例
func NewBangumiModule() *BangumiModule {
	return &BangumiModule{}
}

// InitModule 初始化模�?func (b *BangumiModule) InitModule() error {
	b.bangumiAPI = NewBangumiApi()
	return nil
}

// Stop 停止模块
func (b *BangumiModule) Stop() {
	// 停止模块逻辑，Bangumi API不需要特殊处�?	logger.Info("停止Bangumi模块")
}

// Test 测试模块连接�?func (b *BangumiModule) Test() (bool, string) {
	resp := http.RequestUtils.GetRes("https://api.bgm.tv/", nil, nil, 0)
	if resp != nil && resp.StatusCode == 200 {
		return true, ""
	} else if resp != nil {
		return false, fmt.Sprintf("无法连接Bangumi，错误码�?d", resp.StatusCode)
	}
	return false, "Bangumi网络连接失败"
}

// InitSetting 初始化设�?func (b *BangumiModule) InitSetting() (string, interface{}) {
	return "", nil
}

// GetName 获取模块名称
func (b *BangumiModule) GetName() string {
	return "Bangumi"
}

// GetType 获取模块类型
func (b *BangumiModule) GetType() types.ModuleType {
	return types.ModuleTypeMediaRecognize
}

// GetSubtype 获取模块子类�?func (b *BangumiModule) GetSubtype() types.MediaRecognizeType {
	return types.MediaRecognizeTypeBangumi
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (b *BangumiModule) GetPriority() int {
	return 3
}

// RecognizeMedia 识别媒体信息
func (b *BangumiModule) RecognizeMedia(bangumiid *int, kwargs map[string]interface{}) *models.MediaInfo {
	if bangumiid == nil || *bangumiid <= 0 {
		return nil
	}

	// 直接查询详情
	info := b.BangumiInfo(*bangumiid)
	if info != nil {
		// 赋值Bangumi信息并返�?		mediainfo := models.NewMediaInfoFromBangumi(info)
		logger.Info(fmt.Sprintf("%d Bangumi识别结果�?s %s", 
			*bangumiid, mediainfo.Type, mediainfo.TitleYear))
		return mediainfo
	} else {
		logger.Info(fmt.Sprintf("%d 未匹配到Bangumi媒体信息", *bangumiid))
	}

	return nil
}

// AsyncRecognizeMedia 识别媒体信息（异步版本）
func (b *BangumiModule) AsyncRecognizeMedia(bangumiid *int, kwargs map[string]interface{}) *models.MediaInfo {
	if bangumiid == nil || *bangumiid <= 0 {
		return nil
	}

	// 直接查询详情
	info := b.AsyncBangumiInfo(*bangumiid)
	if info != nil {
		// 赋值Bangumi信息并返�?		mediainfo := models.NewMediaInfoFromBangumi(info)
		logger.Info(fmt.Sprintf("%d Bangumi识别结果�?s %s", 
			*bangumiid, mediainfo.Type, mediainfo.TitleYear))
		return mediainfo
	} else {
		logger.Info(fmt.Sprintf("%d 未匹配到Bangumi媒体信息", *bangumiid))
	}

	return nil
}

// SearchMedias 搜索媒体信息
func (b *BangumiModule) SearchMedias(meta *models.MetaBase) []*models.MediaInfo {
	if config.Config.SEARCH_SOURCE != "" && !strings.Contains(config.Config.SEARCH_SOURCE, "bangumi") {
		return nil
	}
	
	if meta.Name == "" {
		return []*models.MediaInfo{}
	}
	
	infos := b.bangumiAPI.Search(meta.Name)
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				name, _ := infoMap["name"].(string)
				nameCn, _ := infoMap["name_cn"].(string)
				
				// 检查名称是否匹�?				if strings.Contains(strings.ToLower(name), strings.ToLower(meta.Name)) ||
					strings.Contains(strings.ToLower(nameCn), strings.ToLower(meta.Name)) {
					mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
					mediaList = append(mediaList, mediaInfo)
				}
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// AsyncSearchMedias 搜索媒体信息（异步版本）
func (b *BangumiModule) AsyncSearchMedias(meta *models.MetaBase) []*models.MediaInfo {
	if config.Config.SEARCH_SOURCE != "" && !strings.Contains(config.Config.SEARCH_SOURCE, "bangumi") {
		return nil
	}
	
	if meta.Name == "" {
		return []*models.MediaInfo{}
	}
	
	infos := b.bangumiAPI.AsyncSearch(meta.Name)
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				name, _ := infoMap["name"].(string)
				nameCn, _ := infoMap["name_cn"].(string)
				
				// 检查名称是否匹�?				if strings.Contains(strings.ToLower(name), strings.ToLower(meta.Name)) ||
					strings.Contains(strings.ToLower(nameCn), strings.ToLower(meta.Name)) {
					mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
					mediaList = append(mediaList, mediaInfo)
				}
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// BangumiInfo 获取Bangumi信息
func (b *BangumiModule) BangumiInfo(bangumiid int) map[string]interface{} {
	if bangumiid <= 0 {
		return nil
	}
	
	logger.Info(fmt.Sprintf("开始获取Bangumi信息�?d ...", bangumiid))
	return b.bangumiAPI.Detail(bangumiid)
}

// AsyncBangumiInfo 获取Bangumi信息（异步版本）
func (b *BangumiModule) AsyncBangumiInfo(bangumiid int) map[string]interface{} {
	if bangumiid <= 0 {
		return nil
	}
	
	logger.Info(fmt.Sprintf("开始获取Bangumi信息�?d ...", bangumiid))
	return b.bangumiAPI.AsyncDetail(bangumiid)
}

// BangumiCalendar 获取Bangumi每日放�?func (b *BangumiModule) BangumiCalendar() []*models.MediaInfo {
	infos := b.bangumiAPI.Calendar()
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// AsyncBangumiCalendar 获取Bangumi每日放送（异步版本�?func (b *BangumiModule) AsyncBangumiCalendar() []*models.MediaInfo {
	infos := b.bangumiAPI.AsyncCalendar()
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// BangumiCredits 根据BangumiID查询电影演职员表
func (b *BangumiModule) BangumiCredits(bangumiid int) []*schemas.MediaPerson {
	persons := b.bangumiAPI.Credits(bangumiid)
	if len(persons) > 0 {
		personList := []*schemas.MediaPerson{}
		for _, person := range persons {
			mediaPerson := &schemas.MediaPerson{
				Source: "bangumi",
			}
			
			// 从person map中提取字�?			if id, ok := person["id"]; ok {
				mediaPerson.ID = fmt.Sprintf("%v", id)
			}
			
			if name, ok := person["name"]; ok {
				mediaPerson.Name = fmt.Sprintf("%v", name)
			}
			
			if images, ok := person["images"]; ok {
				mediaPerson.Images = images
			}
			
			if career, ok := person["career"]; ok {
				mediaPerson.Career = career
			}
			
			personList = append(personList, mediaPerson)
		}
		return personList
	}
	
	return []*schemas.MediaPerson{}
}

// AsyncBangumiCredits 根据BangumiID查询电影演职员表（异步版本）
func (b *BangumiModule) AsyncBangumiCredits(bangumiid int) []*schemas.MediaPerson {
	persons := b.bangumiAPI.AsyncCredits(bangumiid)
	if len(persons) > 0 {
		personList := []*schemas.MediaPerson{}
		for _, person := range persons {
			mediaPerson := &schemas.MediaPerson{
				Source: "bangumi",
			}
			
			// 从person map中提取字�?			if id, ok := person["id"]; ok {
				mediaPerson.ID = fmt.Sprintf("%v", id)
			}
			
			if name, ok := person["name"]; ok {
				mediaPerson.Name = fmt.Sprintf("%v", name)
			}
			
			if images, ok := person["images"]; ok {
				mediaPerson.Images = images
			}
			
			if career, ok := person["career"]; ok {
				mediaPerson.Career = career
			}
			
			personList = append(personList, mediaPerson)
		}
		return personList
	}
	
	return []*schemas.MediaPerson{}
}

// BangumiRecommend 根据BangumiID查询推荐电影
func (b *BangumiModule) BangumiRecommend(bangumiid int) []*models.MediaInfo {
	subjects := b.bangumiAPI.Subjects(bangumiid)
	if subjects != nil {
		if subjectsList, ok := subjects.([]interface{}); ok {
			mediaList := []*models.MediaInfo{}
			for _, subject := range subjectsList {
				if subjectMap, ok := subject.(map[string]interface{}); ok {
					mediaInfo := models.NewMediaInfoFromBangumi(subjectMap)
					mediaList = append(mediaList, mediaInfo)
				}
			}
			return mediaList
		}
	}
	
	return []*models.MediaInfo{}
}

// AsyncBangumiRecommend 根据BangumiID查询推荐电影（异步版本）
func (b *BangumiModule) AsyncBangumiRecommend(bangumiid int) []*models.MediaInfo {
	subjects := b.bangumiAPI.AsyncSubjects(bangumiid)
	if subjects != nil {
		if subjectsList, ok := subjects.([]interface{}); ok {
			mediaList := []*models.MediaInfo{}
			for _, subject := range subjectsList {
				if subjectMap, ok := subject.(map[string]interface{}); ok {
					mediaInfo := models.NewMediaInfoFromBangumi(subjectMap)
					mediaList = append(mediaList, mediaInfo)
				}
			}
			return mediaList
		}
	}
	
	return []*models.MediaInfo{}
}

// BangumiPersonDetail 获取人物详细信息
func (b *BangumiModule) BangumiPersonDetail(personID int) *schemas.MediaPerson {
	personinfo := b.bangumiAPI.PersonDetail(personID)
	if personinfo != nil {
		mediaPerson := &schemas.MediaPerson{
			Source: "bangumi",
		}
		
		if id, exists := personinfo["id"]; exists {
			mediaPerson.ID = fmt.Sprintf("%v", id)
		}
		
		if name, exists := personinfo["name"]; exists {
			mediaPerson.Name = fmt.Sprintf("%v", name)
		}
		
		if images, exists := personinfo["images"]; exists {
			mediaPerson.Images = images
		}
		
		if summary, exists := personinfo["summary"]; exists {
			mediaPerson.Biography = fmt.Sprintf("%v", summary)
		}
		
		if birthDay, exists := personinfo["birth_day"]; exists {
			mediaPerson.Birthday = fmt.Sprintf("%v", birthDay)
		}
		
		if gender, exists := personinfo["gender"]; exists {
			mediaPerson.Gender = fmt.Sprintf("%v", gender)
		}
		
		return mediaPerson
	}
	
	return nil
}

// AsyncBangumiPersonDetail 获取人物详细信息（异步版本）
func (b *BangumiModule) AsyncBangumiPersonDetail(personID int) *schemas.MediaPerson {
	personinfo := b.bangumiAPI.AsyncPersonDetail(personID)
	if personinfo != nil {
		mediaPerson := &schemas.MediaPerson{
			Source: "bangumi",
		}
		
		if id, exists := personinfo["id"]; exists {
			mediaPerson.ID = fmt.Sprintf("%v", id)
		}
		
		if name, exists := personinfo["name"]; exists {
			mediaPerson.Name = fmt.Sprintf("%v", name)
		}
		
		if images, exists := personinfo["images"]; exists {
			mediaPerson.Images = images
		}
		
		if summary, exists := personinfo["summary"]; exists {
			mediaPerson.Biography = fmt.Sprintf("%v", summary)
		}
		
		if birthDay, exists := personinfo["birth_day"]; exists {
			mediaPerson.Birthday = fmt.Sprintf("%v", birthDay)
		}
		
		if gender, exists := personinfo["gender"]; exists {
			mediaPerson.Gender = fmt.Sprintf("%v", gender)
		}
		
		return mediaPerson
	}
	
	return nil
}

// BangumiPersonCredits 根据BangumiID查询人物参演作品
func (b *BangumiModule) BangumiPersonCredits(personID int) []*models.MediaInfo {
	creditsInfo := b.bangumiAPI.PersonCredits(personID)
	if len(creditsInfo) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, credit := range creditsInfo {
			if creditMap, ok := credit.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(creditMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// AsyncBangumiPersonCredits 根据BangumiID查询人物参演作品（异步版本）
func (b *BangumiModule) AsyncBangumiPersonCredits(personID int) []*models.MediaInfo {
	creditsInfo := b.bangumiAPI.AsyncPersonCredits(personID)
	if len(creditsInfo) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, credit := range creditsInfo {
			if creditMap, ok := credit.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(creditMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// BangumiDiscover 发现Bangumi番剧
func (b *BangumiModule) BangumiDiscover(kwargs map[string]interface{}) []*models.MediaInfo {
	infos := b.bangumiAPI.Discover(kwargs)
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}

// AsyncBangumiDiscover 发现Bangumi番剧（异步版本）
func (b *BangumiModule) AsyncBangumiDiscover(kwargs map[string]interface{}) []*models.MediaInfo {
	infos := b.bangumiAPI.AsyncDiscover(kwargs)
	if len(infos) > 0 {
		mediaList := []*models.MediaInfo{}
		for _, info := range infos {
			if infoMap, ok := info.(map[string]interface{}); ok {
				mediaInfo := models.NewMediaInfoFromBangumi(infoMap)
				mediaList = append(mediaList, mediaInfo)
			}
		}
		return mediaList
	}
	
	return []*models.MediaInfo{}
}
