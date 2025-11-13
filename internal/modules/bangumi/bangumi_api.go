package bangumi

import (
	"fmt"
	"time"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/utils"
)

// BangumiApi Bangumi API客户�?type BangumiApi struct {
	urls    map[string]string
	baseURL string
	cache   *BangumiCache
}

// NewBangumiApi 创建BangumiApi实例
func NewBangumiApi() *BangumiApi {
	return &BangumiApi{
		urls: map[string]string{
			"discover":       "v0/subjects",
			"search":         "search/subjects/%s?type=2",
			"calendar":       "calendar",
			"detail":         "v0/subjects/%s",
			"credits":        "v0/subjects/%s/persons",
			"subjects":       "v0/subjects/%s/subjects",
			"characters":     "v0/subjects/%s/characters",
			"person_detail":  "v0/persons/%s",
			"person_credits": "v0/persons/%s/subjects",
		},
		baseURL: "https://api.bgm.tv/",
		cache:   NewBangumiCache(),
	}
}

// invoke 调用API
func (b *BangumiApi) invoke(url string, key string, kwargs map[string]interface{}) interface{} {
	// 构建缓存�?	cacheKey := url
	if kwargs != nil {
		for k, v := range kwargs {
			cacheKey += fmt.Sprintf("&%s=%v", k, v)
		}
	}
	
	// 尝试从缓存获�?	if cachedData, exists := b.cache.Get(cacheKey); exists {
		return cachedData
	}
	
	reqURL := b.baseURL + url
	params := make(map[string]string)
	
	if kwargs != nil {
		for k, v := range kwargs {
			if str, ok := v.(string); ok {
				params[k] = str
			}
		}
	}
	
	headers := map[string]string{
		"User-Agent": config.Config.NORMAL_USER_AGENT,
	}
	
	resp := utils.RequestUtils.GetRes(reqURL, params, headers, 0)
	if resp == nil {
		return nil
	}
	
	result := resp.JSON()
	if result == nil {
		return nil
	}
	
	var finalResult interface{}
	if key != "" {
		if val, exists := result[key]; exists {
			finalResult = val
		} else {
			finalResult = nil
		}
	} else {
		finalResult = result
	}
	
	// 缓存结果
	b.cache.Set(cacheKey, finalResult)
	
	return finalResult
}

// asyncInvoke 异步调用API
func (b *BangumiApi) asyncInvoke(url string, key string, kwargs map[string]interface{}) interface{} {
	// 在Go中，我们可以通过goroutine实现异步调用
	// 但为了简化，这里直接调用同步版本
	// 实际项目中可以使用goroutine和channel实现真正的异�?	return b.invoke(url, key, kwargs)
}

// Search 搜索媒体信息
func (b *BangumiApi) Search(name string) []interface{} {
	url := fmt.Sprintf("search/subject/%s", name)
	result := b.invoke(url, "", nil)
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if list, exists := resultMap["list"]; exists {
				if listSlice, ok := list.([]interface{}); ok {
					return listSlice
				}
			}
		}
	}
	
	return []interface{}{}
}

// AsyncSearch 搜索媒体信息（异步版本）
func (b *BangumiApi) AsyncSearch(name string) []interface{} {
	url := fmt.Sprintf("search/subject/%s", name)
	result := b.asyncInvoke(url, "", nil)
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if list, exists := resultMap["list"]; exists {
				if listSlice, ok := list.([]interface{}); ok {
					return listSlice
				}
			}
		}
	}
	
	return []interface{}{}
}

// Calendar 获取每日放送，返回items
func (b *BangumiApi) Calendar() []interface{} {
	retList := []interface{}{}
	ts := time.Now().Format("20060102")
	result := b.invoke(b.urls["calendar"], "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if items, exists := itemMap["items"]; exists {
						if itemsList, ok := items.([]interface{}); ok {
							retList = append(retList, itemsList...)
						}
					}
				}
			}
		}
	}
	
	return retList
}

// AsyncCalendar 获取每日放送，返回items（异步版本）
func (b *BangumiApi) AsyncCalendar() []interface{} {
	retList := []interface{}{}
	ts := time.Now().Format("20060102")
	result := b.asyncInvoke(b.urls["calendar"], "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if items, exists := itemMap["items"]; exists {
						if itemsList, ok := items.([]interface{}); ok {
							retList = append(retList, itemsList...)
						}
					}
				}
			}
		}
	}
	
	return retList
}

// Detail 获取番剧详情
func (b *BangumiApi) Detail(bid int) map[string]interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["detail"], bid)
	result := b.invoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			return resultMap
		}
	}
	
	return nil
}

// AsyncDetail 获取番剧详情（异步版本）
func (b *BangumiApi) AsyncDetail(bid int) map[string]interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["detail"], bid)
	result := b.asyncInvoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			return resultMap
		}
	}
	
	return nil
}

// Credits 获取番剧人物
func (b *BangumiApi) Credits(bid int) []map[string]interface{} {
	retList := []map[string]interface{}{}
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["characters"], bid)
	result := b.invoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					characterID := itemMap["id"]
					actors := itemMap["actors"]
					
					if characterID != nil && actors != nil {
						if actorsList, ok := actors.([]interface{}); ok && len(actorsList) > 0 {
							if actorInfo, ok := actorsList[0].(map[string]interface{}); ok {
								// 更新角色信息
								career := []interface{}{itemMap["name"]}
								actorInfo["career"] = career
								retList = append(retList, actorInfo)
							}
						}
					}
				}
			}
		}
	}
	
	return retList
}

// AsyncCredits 获取番剧人物（异步版本）
func (b *BangumiApi) AsyncCredits(bid int) []map[string]interface{} {
	retList := []map[string]interface{}{}
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["characters"], bid)
	result := b.asyncInvoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					characterID := itemMap["id"]
					actors := itemMap["actors"]
					
					if characterID != nil && actors != nil {
						if actorsList, ok := actors.([]interface{}); ok && len(actorsList) > 0 {
							if actorInfo, ok := actorsList[0].(map[string]interface{}); ok {
								// 更新角色信息
								career := []interface{}{itemMap["name"]}
								actorInfo["career"] = career
								retList = append(retList, actorInfo)
							}
						}
					}
				}
			}
		}
	}
	
	return retList
}

// Subjects 获取关联条目信息
func (b *BangumiApi) Subjects(bid int) interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["subjects"], bid)
	return b.invoke(url, "", map[string]interface{}{"_ts": ts})
}

// AsyncSubjects 获取关联条目信息（异步版本）
func (b *BangumiApi) AsyncSubjects(bid int) interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["subjects"], bid)
	return b.asyncInvoke(url, "", map[string]interface{}{"_ts": ts})
}

// PersonDetail 获取人物详细信息
func (b *BangumiApi) PersonDetail(personID int) map[string]interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["person_detail"], personID)
	result := b.invoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			return resultMap
		}
	}
	
	return nil
}

// AsyncPersonDetail 获取人物详细信息（异步版本）
func (b *BangumiApi) AsyncPersonDetail(personID int) map[string]interface{} {
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["person_detail"], personID)
	result := b.asyncInvoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			return resultMap
		}
	}
	
	return nil
}

// PersonCredits 获取人物参演作品
func (b *BangumiApi) PersonCredits(personID int) []interface{} {
	retList := []interface{}{}
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["person_credits"], personID)
	result := b.invoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				retList = append(retList, item)
			}
		}
	}
	
	return retList
}

// AsyncPersonCredits 获取人物参演作品（异步版本）
func (b *BangumiApi) AsyncPersonCredits(personID int) []interface{} {
	retList := []interface{}{}
	ts := time.Now().Format("20060102")
	url := fmt.Sprintf(b.urls["person_credits"], personID)
	result := b.asyncInvoke(url, "", map[string]interface{}{"_ts": ts})
	
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			for _, item := range resultList {
				retList = append(retList, item)
			}
		}
	}
	
	return retList
}

// Discover 发现
func (b *BangumiApi) Discover(kwargs map[string]interface{}) []interface{} {
	params := map[string]interface{}{
		"_ts": time.Now().Format("20060102"),
	}
	
	if kwargs != nil {
		for k, v := range kwargs {
			params[k] = v
		}
	}
	
	result := b.invoke(b.urls["discover"], "data", params)
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			return resultList
		}
	}
	
	return []interface{}{}
}

// AsyncDiscover 发现（异步版本）
func (b *BangumiApi) AsyncDiscover(kwargs map[string]interface{}) []interface{} {
	params := map[string]interface{}{
		"_ts": time.Now().Format("20060102"),
	}
	
	if kwargs != nil {
		for k, v := range kwargs {
			params[k] = v
		}
	}
	
	result := b.asyncInvoke(b.urls["discover"], "data", params)
	if result != nil {
		if resultList, ok := result.([]interface{}); ok {
			return resultList
		}
	}
	
	return []interface{}{}
}
