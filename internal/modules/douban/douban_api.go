package douban

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// DoubanApi 豆瓣API客户�?type DoubanApi struct {
	urls           map[string]string
	userAgents     []string
	apiSecretKey   string
	apiKey         string
	apiKey2        string
	baseURL        string
	apiURL         string
	clearAsyncCache bool
}

// NewDoubanApi 创建DoubanApi实例
func NewDoubanApi() *DoubanApi {
	return &DoubanApi{
		urls: map[string]string{
			// 搜索�?			"search":       "/search/weixin",
			"search_agg":   "/search",
			"search_subject": "/search/subjects",
			"imdbid":       "/movie/imdb/%s",

			// 电影探索
			"movie_recommend": "/movie/recommend",
			// 电视剧探�?			"tv_recommend": "/tv/recommend",
			// 搜索
			"movie_tag":    "/movie/tag",
			"tv_tag":       "/tv/tag",
			"movie_search": "/search/movie",
			"tv_search":    "/search/movie",
			"book_search":  "/search/book",
			"group_search": "/search/group",

			// 各类主题合集
			"movie_showing":           "/subject_collection/movie_showing/items",
			"movie_hot_gaia":          "/subject_collection/movie_hot_gaia/items",
			"movie_soon":              "/subject_collection/movie_soon/items",
			"movie_top250":            "/subject_collection/movie_top250/items",
			"movie_scifi":             "/subject_collection/movie_scifi/items",
			"movie_comedy":            "/subject_collection/movie_comedy/items",
			"movie_action":            "/subject_collection/movie_action/items",
			"movie_love":              "/subject_collection/movie_love/items",

			"tv_hot":                  "/subject_collection/tv_hot/items",
			"tv_domestic":             "/subject_collection/tv_domestic/items",
			"tv_american":             "/subject_collection/tv_american/items",
			"tv_japanese":             "/subject_collection/tv_japanese/items",
			"tv_korean":               "/subject_collection/tv_korean/items",
			"tv_animation":            "/subject_collection/tv_animation/items",
			"tv_variety_show":         "/subject_collection/tv_variety_show/items",
			"tv_chinese_best_weekly":  "/subject_collection/tv_chinese_best_weekly/items",
			"tv_global_best_weekly":   "/subject_collection/tv_global_best_weekly/items",

			"show_hot":                "/subject_collection/show_hot/items",
			"show_domestic":           "/subject_collection/show_domestic/items",
			"show_foreign":            "/subject_collection/show_foreign/items",

			"book_bestseller":         "/subject_collection/book_bestseller/items",
			"book_top250":             "/subject_collection/book_top250/items",
			"book_fiction_hot_weekly": "/subject_collection/book_fiction_hot_weekly/items",
			"book_nonfiction_hot_weekly": "/subject_collection/book_nonfiction_hot_weekly/items",

			"music_single":            "/subject_collection/music_single/items",

			"movie_rank_list":         "/movie/rank_list",
			"movie_year_ranks":        "/movie/year_ranks",
			"book_rank_list":          "/book/rank_list",
			"tv_rank_list":            "/tv/rank_list",

			"movie_detail":            "/movie/",
			"movie_rating":            "/movie/%s/rating",
			"movie_photos":            "/movie/%s/photos",
			"movie_trailers":          "/movie/%s/trailers",
			"movie_interests":         "/movie/%s/interests",
			"movie_reviews":           "/movie/%s/reviews",
			"movie_recommendations":   "/movie/%s/recommendations",
			"movie_celebrities":       "/movie/%s/celebrities",

			"tv_detail":               "/tv/",
			"tv_rating":               "/tv/%s/rating",
			"tv_photos":               "/tv/%s/photos",
			"tv_trailers":             "/tv/%s/trailers",
			"tv_interests":            "/tv/%s/interests",
			"tv_reviews":              "/tv/%s/reviews",
			"tv_recommendations":      "/tv/%s/recommendations",
			"tv_celebrities":          "/tv/%s/celebrities",

			"book_detail":             "/book/",
			"book_rating":             "/book/%s/rating",
			"book_interests":          "/book/%s/interests",
			"book_reviews":            "/book/%s/reviews",
			"book_recommendations":    "/book/%s/recommendations",

			"music_detail":            "/music/",
			"music_rating":            "/music/%s/rating",
			"music_interests":         "/music/%s/interests",
			"music_reviews":           "/music/%s/reviews",
			"music_recommendations":   "/music/%s/recommendations",

			"doulist":                 "/doulist/",
			"doulist_items":           "/doulist/%s/items",

			"person_detail":           "/elessar/subject/",
			"person_work":             "/elessar/work_collections/%s/works",
		},
		userAgents: []string{
			"api-client/1 com.douban.frodo/7.22.0.beta9(231) Android/23 product/Mate 40 vendor/HUAWEI model/Mate 40 brand/HUAWEI  rom/android  network/wifi  platform/AndroidPad",
			"api-client/1 com.douban.frodo/7.18.0(230) Android/22 product/MI 9 vendor/Xiaomi model/MI 9 brand/Android  rom/miui6  network/wifi  platform/mobile nd/1",
			"api-client/1 com.douban.frodo/7.1.0(205) Android/29 product/perseus vendor/Xiaomi model/Mi MIX 3  rom/miui6  network/wifi  platform/mobile nd/1",
			"api-client/1 com.douban.frodo/7.3.0(207) Android/22 product/MI 9 vendor/Xiaomi model/MI 9 brand/Android  rom/miui6  network/wifi platform/mobile nd/1",
		},
		apiSecretKey: "bf7dddc7c9cfe6f7",
		apiKey:       "0dad551ec0f84ed02907ff5c42e8ec70",
		apiKey2:      "0ab215a8b1977939201640fa14c66bab",
		baseURL:      "https://frodo.douban.com/api/v2",
		apiURL:       "https://api.douban.com/v2",
	}
}

// sign 签名
func (d *DoubanApi) sign(urlStr, ts, method string) string {
	urlPath := urlStr
	if u, err := url.Parse(urlStr); err == nil {
		urlPath = u.Path
	}
	rawSign := strings.ToUpper(method) + "&" + url.QueryEscape(urlPath) + "&" + ts
	h := hmac.New(sha1.New, []byte(d.apiSecretKey))
	h.Write([]byte(rawSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// invokeRecommend 推荐/发现类API
func (d *DoubanApi) invokeRecommend(urlKey string, kwargs map[string]interface{}) map[string]interface{} {
	return d.invoke(urlKey, kwargs)
}

// invokeSearch 搜索类API
func (d *DoubanApi) invokeSearch(urlKey string, kwargs map[string]interface{}) map[string]interface{} {
	return d.invoke(urlKey, kwargs)
}

// prepareGetRequest 准备GET请求的URL和参�?func (d *DoubanApi) prepareGetRequest(urlKey string, kwargs map[string]interface{}) (string, map[string]string) {
	reqURL := d.baseURL + d.urls[urlKey]

	params := make(map[string]string)
	params["apiKey"] = d.apiKey

	// 添加kwargs参数
	for k, v := range kwargs {
		if str, ok := v.(string); ok {
			params[k] = str
		}
	}

	ts := ""
	if t, exists := params["_ts"]; exists {
		ts = t
		delete(params, "_ts")
	} else {
		ts = time.Now().Format("20060102")
	}

	params["os_rom"] = "android"
	params["apiKey"] = d.apiKey
	params["_ts"] = ts
	params["_sig"] = d.sign(reqURL, ts, "GET")

	return reqURL, params
}

// handleResponse 处理HTTP响应
func (d *DoubanApi) handleResponse(resp *utils.HttpResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}
	return resp.JSON()
}

// invoke GET请求
func (d *DoubanApi) invoke(urlKey string, kwargs map[string]interface{}) map[string]interface{} {
	// 检查缓存逻辑应该在这里实现，但为了简化，我们直接调用
	reqURL, params := d.prepareGetRequest(urlKey, kwargs)
	
	ua := d.userAgents[rand.Intn(len(d.userAgents))]
	resp := utils.RequestUtils.GetRes(reqURL, params, map[string]string{"User-Agent": ua}, 0)
	return d.handleResponse(resp)
}

// preparePostRequest 准备POST请求的URL和参�?func (d *DoubanApi) preparePostRequest(urlKey string, kwargs map[string]interface{}) (string, map[string]string) {
	reqURL := d.apiURL + fmt.Sprintf(d.urls[urlKey])
	params := map[string]string{"apikey": d.apiKey2}
	
	// 添加kwargs参数
	for k, v := range kwargs {
		if str, ok := v.(string); ok {
			params[k] = str
		}
	}
	
	if _, exists := params["_ts"]; exists {
		delete(params, "_ts")
	}
	
	return reqURL, params
}

// post POST请求
func (d *DoubanApi) post(urlKey string, kwargs map[string]interface{}) map[string]interface{} {
	// 检查缓存逻辑应该在这里实现，但为了简化，我们直接调用
	reqURL, params := d.preparePostRequest(urlKey, kwargs)
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=utf-8",
		"Cookie":       "bid=J9zb1zA5sJc",
		"User-Agent":   config.Config.NORMAL_USER_AGENT,
	}
	
	resp := utils.RequestUtils.PostRes(reqURL, headers, "", 0, params)
	return d.handleResponse(resp)
}

// Imdbid IMDBID搜索
func (d *DoubanApi) Imdbid(imdbid string, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.post("imdbid", map[string]interface{}{
		"_ts": ts,
	})
}

// Search 关键字搜�?func (d *DoubanApi) Search(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("search", map[string]interface{}{
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieSearch 电影搜索
func (d *DoubanApi) MovieSearch(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("movie_search", map[string]interface{}{
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvSearch 电视搜索
func (d *DoubanApi) TvSearch(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("tv_search", map[string]interface{}{
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// BookSearch 书籍搜索
func (d *DoubanApi) BookSearch(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("book_search", map[string]interface{}{
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// GroupSearch 小组搜索
func (d *DoubanApi) GroupSearch(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("group_search", map[string]interface{}{
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// PersonSearch 人物搜索
func (d *DoubanApi) PersonSearch(keyword string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeSearch("search_subject", map[string]interface{}{
		"type":  "person",
		"q":     keyword,
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieShowing 正在热映
func (d *DoubanApi) MovieShowing(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("movie_showing", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieSoon 即将上映
func (d *DoubanApi) MovieSoon(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("movie_soon", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieHotGaia 热门电影
func (d *DoubanApi) MovieHotGaia(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("movie_hot_gaia", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvHot 热门剧集
func (d *DoubanApi) TvHot(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_hot", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvAnimation 动画
func (d *DoubanApi) TvAnimation(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_animation", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvVarietyShow 综艺
func (d *DoubanApi) TvVarietyShow(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_variety_show", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvRankList 电视剧排行榜
func (d *DoubanApi) TvRankList(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_rank_list", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// ShowHot 综艺热门
func (d *DoubanApi) ShowHot(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("show_hot", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieDetail 电影详情
func (d *DoubanApi) MovieDetail(subjectID string) map[string]interface{} {
	urlKey := "movie_detail" + subjectID
	return d.invokeSearch(urlKey, nil)
}

// MovieCelebrities 电影演职�?func (d *DoubanApi) MovieCelebrities(subjectID string) map[string]interface{} {
	urlKey := fmt.Sprintf(d.urls["movie_celebrities"], subjectID)
	return d.invokeSearch(urlKey, nil)
}

// TvDetail 电视剧详�?func (d *DoubanApi) TvDetail(subjectID string) map[string]interface{} {
	urlKey := "tv_detail" + subjectID
	return d.invokeSearch(urlKey, nil)
}

// TvCelebrities 电视剧演职员
func (d *DoubanApi) TvCelebrities(subjectID string) map[string]interface{} {
	urlKey := fmt.Sprintf(d.urls["tv_celebrities"], subjectID)
	return d.invokeSearch(urlKey, nil)
}

// BookDetail 书籍详情
func (d *DoubanApi) BookDetail(subjectID string) map[string]interface{} {
	urlKey := "book_detail" + subjectID
	return d.invokeSearch(urlKey, nil)
}

// MovieTop250 电影TOP250
func (d *DoubanApi) MovieTop250(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("movie_top250", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieRecommend 电影探索
func (d *DoubanApi) MovieRecommend(tags, sort string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	params := map[string]interface{}{
		"sort":  sort,
		"start": start,
		"count": count,
		"_ts":   ts,
	}
	
	if tags != "" {
		params["tags"] = tags
	}
	
	return d.invokeRecommend("movie_recommend", params)
}

// TvRecommend 电视剧探�?func (d *DoubanApi) TvRecommend(tags, sort string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	params := map[string]interface{}{
		"sort":  sort,
		"start": start,
		"count": count,
		"_ts":   ts,
	}
	
	if tags != "" {
		params["tags"] = tags
	}
	
	return d.invokeRecommend("tv_recommend", params)
}

// TvChineseBestWeekly 华语口碑周榜
func (d *DoubanApi) TvChineseBestWeekly(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_chinese_best_weekly", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvGlobalBestWeekly 全球口碑周榜
func (d *DoubanApi) TvGlobalBestWeekly(start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	return d.invokeRecommend("tv_global_best_weekly", map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// DoulistDetail 豆列详情
func (d *DoubanApi) DoulistDetail(subjectID string) map[string]interface{} {
	urlKey := "doulist" + subjectID
	return d.invokeSearch(urlKey, nil)
}

// DoulistItems 豆列列表
func (d *DoubanApi) DoulistItems(subjectID string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["doulist_items"], subjectID)
	return d.invokeSearch(urlKey, map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MovieRecommendations 电影推荐
func (d *DoubanApi) MovieRecommendations(subjectID string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["movie_recommendations"], subjectID)
	return d.invokeRecommend(urlKey, map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvRecommendations 电视剧推�?func (d *DoubanApi) TvRecommendations(subjectID string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["tv_recommendations"], subjectID)
	return d.invokeRecommend(urlKey, map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// MoviePhotos 电影剧照
func (d *DoubanApi) MoviePhotos(subjectID string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["movie_photos"], subjectID)
	return d.invokeSearch(urlKey, map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// TvPhotos 电视剧剧�?func (d *DoubanApi) TvPhotos(subjectID string, start, count int, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["tv_photos"], subjectID)
	return d.invokeSearch(urlKey, map[string]interface{}{
		"start": start,
		"count": count,
		"_ts":   ts,
	})
}

// PersonDetail 用户详情
func (d *DoubanApi) PersonDetail(subjectID int) map[string]interface{} {
	urlKey := "person_detail" + fmt.Sprintf("%d", subjectID)
	return d.invokeSearch(urlKey, nil)
}

// PersonWork 用户作品�?func (d *DoubanApi) PersonWork(subjectID int, start, count int, sortBy, collectionTitle, ts string) map[string]interface{} {
	if ts == "" {
		ts = time.Now().Format("20060102")
	}
	urlKey := fmt.Sprintf(d.urls["person_work"], subjectID)
	return d.invokeSearch(urlKey, map[string]interface{}{
		"sortby":           sortBy,
		"collection_title": collectionTitle,
		"start":            start,
		"count":            count,
		"_ts":              ts,
	})
}

// ClearCache 清空LRU缓存
func (d *DoubanApi) ClearCache() {
	// 清空缓存的逻辑应该在这里实�?	d.clearAsyncCache = true
}

// Close 关闭连接
func (d *DoubanApi) Close() {
	// 关闭连接的逻辑
	logger.Info("关闭豆瓣API连接")
}
