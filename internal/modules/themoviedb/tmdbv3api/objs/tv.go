package objs

import (
	"fmt"
	"net/url"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// TV 电视剧对�?type TV struct {
	tmdb *tmdbv3api.TMDb
}

// NewTV 创建TV实例
func NewTV(tmdb *tmdbv3api.TMDb) *TV {
	return &TV{
		tmdb: tmdb,
	}
}

// Details 获取电视剧详�?/*
Get the primary TV show details by id.
:param tv_id: 电视剧ID
:param append_to_response: 附加响应参数
:return: 电视剧详�?*/
func (t *TV) Details(tvID int, appendToResponse string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d", tvID)
	params := ""
	if appendToResponse != "" {
		params = "append_to_response=" + appendToResponse
	}

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AccountStates 获取账户状�?/*
Grab the following account states for a session:
TV show rating, If it belongs to your watchlist, or If it belongs to your favourite list.
:param tv_id: 电视剧ID
:return: 账户状�?*/
func (t *TV) AccountStates(tvID int) (map[string]interface{}, error) {
	sessionID, err := t.tmdb.SessionID()
	if err != nil {
		return nil, err
	}

	action := fmt.Sprintf("/tv/%d/account_states", tvID)
	params := "session_id=" + sessionID

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AggregateCredits 获取总演职员�?/*
Get the aggregate credits (cast and crew) that have been added to a TV show.
This call differs from the main credits call in that it does not return the newest season but rather,
is a view of all the entire cast & crew for all episodes belonging to a TV show.
:param tv_id: 电视剧ID
:return: 总演职员�?*/
func (t *TV) AggregateCredits(tvID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/aggregate_credits", tvID)

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AlternativeTitles 获取备选标�?/*
Returns all of the alternative titles for a TV show.
:param tv_id: 电视剧ID
:return: 备选标题列�?*/
func (t *TV) AlternativeTitles(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/alternative_titles", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Changes 获取变更记录
/*
Get the changes for a TV show. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
:param tv_id: 电视剧ID
:param start_date: 开始日�?:param end_date: 结束日期
:param page: 页码
:return: 变更记录
*/
func (t *TV) Changes(tvID int, startDate *string, endDate *string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/changes", tvID)
	params := "page=" + strconv.Itoa(page)
	if startDate != nil {
		params += "&start_date=" + *startDate
	}
	if endDate != nil {
		params += "&end_date=" + *endDate
	}
	key := "changes"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ContentRatings 获取内容分级
/*
Get the list of content ratings (certifications) that have been added to a TV show.
:param tv_id: 电视剧ID
:return: 内容分级列表
*/
func (t *TV) ContentRatings(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/content_ratings", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Credits 获取演职员表
/*
Get the credits (cast and crew) that have been added to a TV show.
:param tv_id: 电视剧ID
:return: 演职员表
*/
func (t *TV) Credits(tvID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/credits", tvID)

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// EpisodeGroups 获取剧集�?/*
Get all of the episode groups that have been created for a TV show.
:param tv_id: 电视剧ID
:return: 剧集组列�?*/
func (t *TV) EpisodeGroups(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/episode_groups", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// GroupEpisodes 获取剧集组所有剧�?/*
查询剧集组所有剧�?:param group_id: 组ID
:return: 剧集�?*/
func (t *TV) GroupEpisodes(groupID string) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/episode_group/%s", groupID)
	key := "groups"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ExternalIDs 获取外部ID
/*
Get the external ids for a TV show.
:param tv_id: 电视剧ID
:return: 外部ID
*/
func (t *TV) ExternalIDs(tvID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/external_ids", tvID)

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取图片
/*
Get the images that belong to a TV show.
Querying images with a language parameter will filter the results.
If you want to include a fallback language (especially useful for backdrops)
you can use the include_image_language parameter.
This should be a comma separated value like so: include_image_language=en,null.
:param tv_id: 电视剧ID
:param include_image_language: 包含的图片语言
:return: 图片信息
*/
func (t *TV) Images(tvID int, includeImageLanguage *string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/images", tvID)
	params := ""
	if includeImageLanguage != nil {
		params = "include_image_language=" + *includeImageLanguage
	}

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Keywords 获取关键�?/*
Get the keywords that have been added to a TV show.
:param tv_id: 电视剧ID
:return: 关键词列�?*/
func (t *TV) Keywords(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/keywords", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Recommendations 获取推荐
/*
Get the list of TV show recommendations for this item.
:param tv_id: 电视剧ID
:param page: 页码
:return: 推荐列表
*/
func (t *TV) Recommendations(tvID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/recommendations", tvID)
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Reviews 获取评论
/*
Get the reviews for a TV show.
:param tv_id: 电视剧ID
:param page: 页码
:return: 评论列表
*/
func (t *TV) Reviews(tvID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/reviews", tvID)
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ScreenedTheatrically 获取影院上映信息
/*
Get a list of seasons or episodes that have been screened in a film festival or theatre.
:param tv_id: 电视剧ID
:return: 影院上映信息列表
*/
func (t *TV) ScreenedTheatrically(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/screened_theatrically", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Similar 获取相似电视�?/*
Get the primary TV show details by id.
:param tv_id: 电视剧ID
:param page: 页码
:return: 相似电视剧列�?*/
func (t *TV) Similar(tvID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/similar", tvID)
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Translations 获取翻译信息
/*
Get a list of the translations that exist for a TV show.
:param tv_id: 电视剧ID
:return: 翻译信息列表
*/
func (t *TV) Translations(tvID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/translations", tvID)
	key := "translations"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Videos 获取视频信息
/*
Get the videos that have been added to a TV show.
:param tv_id: 电视剧ID
:param include_video_language: 包含的视频语言
:param page: 页码
:return: 视频信息
*/
func (t *TV) Videos(tvID int, includeVideoLanguage *string, page int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/videos", tvID)
	params := "page=" + strconv.Itoa(page)
	if includeVideoLanguage != nil {
		params += "&include_video_language=" + *includeVideoLanguage
	}

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// WatchProviders 获取观看提供�?/*
You can query this method to get a list of the availabilities per country by provider.
:param tv_id: 电视剧ID
:return: 观看提供商列�?*/
func (t *TV) WatchProviders(tvID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/watch/providers", tvID)
	key := "results"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// RateTVShow 评价电视�?/*
Rate a TV show.
:param tv_id: 电视剧ID
:param rating: 评分�?*/
func (t *TV) RateTVShow(tvID int, rating float64) error {
	sessionID, err := t.tmdb.SessionID()
	if err != nil {
		return err
	}

	action := fmt.Sprintf("/tv/%d/rating", tvID)
	params := "session_id=" + sessionID
	jsonData := map[string]interface{}{"value": rating}

	_, err = t.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	return err
}

// DeleteRating 删除评分
/*
Remove your rating for a TV show.
:param tv_id: 电视剧ID
*/
func (t *TV) DeleteRating(tvID int) error {
	sessionID, err := t.tmdb.SessionID()
	if err != nil {
		return err
	}

	action := fmt.Sprintf("/tv/%d/rating", tvID)
	params := "session_id=" + sessionID

	_, err = t.tmdb.RequestObj(action, params, "DELETE", nil, nil, nil)
	return err
}

// Latest 获取最新电视剧
/*
Get the most newly created TV show. This is a live response and will continuously change.
:return: 最新电视剧
*/
func (t *TV) Latest() (map[string]interface{}, error) {
	action := "/tv/latest"

	result, err := t.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AiringToday 获取今日播出电视�?/*
Get a list of TV shows that are airing today.
This query is purely day based as we do not currently support airing times.
:param page: 页码
:return: 今日播出电视剧列�?*/
func (t *TV) AiringToday(page int) ([]interface{}, error) {
	action := "/tv/airing_today"
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// OnTheAir 获取正在播出的电视剧
/*
Get a list of shows that are currently on the air.
:param page: 页码
:return: 正在播出的电视剧列表
*/
func (t *TV) OnTheAir(page int) ([]interface{}, error) {
	action := "/tv/on_the_air"
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Popular 获取热门电视�?/*
Get a list of the current popular TV shows on TMDb. This list updates daily.
:param page: 页码
:return: 热门电视剧列�?*/
func (t *TV) Popular(page int) ([]interface{}, error) {
	action := "/tv/popular"
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TopRated 获取高评分电视剧
/*
Get a list of the top rated TV shows on TMDb.
:param page: 页码
:return: 高评分电视剧列表
*/
func (t *TV) TopRated(page int) ([]interface{}, error) {
	action := "/tv/top_rated"
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
