package objs

import (
	"fmt"
	"net/url"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Season 季对�?type Season struct {
	tmdb *tmdbv3api.TMDb
}

// NewSeason 创建Season实例
func NewSeason(tmdb *tmdbv3api.TMDb) *Season {
	return &Season{
		tmdb: tmdb,
	}
}

// Details 获取电视季详�?/*
Get the TV season details by id.
:param tv_id: 电视节目ID
:param season_num: 季数
:param append_to_response: 附加响应参数
:return: 电视季详�?*/
func (s *Season) Details(tvID, seasonNum int, appendToResponse string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d", tvID, seasonNum)
	params := ""
	if appendToResponse != "" {
		params = "append_to_response=" + appendToResponse
	}

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AccountStates 获取用户对季中各集的评分
/*
Get all of the user ratings for the season's episodes.
:param tv_id: 电视节目ID
:param season_num: 季数
:return: 用户评分状�?*/
func (s *Season) AccountStates(tvID, seasonNum int) ([]interface{}, error) {
	sessionID, err := s.tmdb.SessionID()
	if err != nil {
		return nil, err
	}

	action := fmt.Sprintf("/tv/%d/season/%d/account_states", tvID, seasonNum)
	params := "session_id=" + sessionID
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AggregateCredits 获取电视季的总演职员�?/*
Get the aggregate credits for TV season.
This call differs from the main credits call in that it does not only return the season credits,
but rather is a view of all the cast & crew for all of the episodes belonging to a season.
:param tv_id: 电视节目ID
:param season_num: 季数
:return: 总演职员�?*/
func (s *Season) AggregateCredits(tvID, seasonNum int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/aggregate_credits", tvID, seasonNum)

	result, err := s.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Changes 获取电视季变更记�?/*
Get the changes for a TV season. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
:param season_id: 季ID
:param start_date: 开始日�?:param end_date: 结束日期
:param page: 页码
:return: 变更记录
*/
func (s *Season) Changes(seasonID int, startDate *string, endDate *string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/season/%d/changes", seasonID)
	params := "page=" + strconv.Itoa(page)
	if startDate != nil {
		params += "&start_date=" + *startDate
	}
	if endDate != nil {
		params += "&end_date=" + *endDate
	}
	key := "changes"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Credits 获取电视季演职员�?/*
Get the credits for TV season.
:param tv_id: 电视节目ID
:param season_num: 季数
:return: 演职员表
*/
func (s *Season) Credits(tvID, seasonNum int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/credits", tvID, seasonNum)

	result, err := s.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ExternalIDs 获取电视季外部ID
/*
Get the external ids for a TV season.
:param tv_id: 电视节目ID
:param season_num: 季数
:return: 外部ID
*/
func (s *Season) ExternalIDs(tvID, seasonNum int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/external_ids", tvID, seasonNum)

	result, err := s.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取电视季图�?/*
Get the images that belong to a TV season.
:param tv_id: 电视节目ID
:param season_num: 季数
:param include_image_language: 包含的图片语言
:return: 图片列表
*/
func (s *Season) Images(tvID, seasonNum int, includeImageLanguage *string) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/images", tvID, seasonNum)
	params := ""
	if includeImageLanguage != nil {
		params = "include_image_language=" + url.QueryEscape(*includeImageLanguage)
	}
	key := "posters"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Translations 获取电视季翻译信�?/*
Get a list of the translations that exist for a TV show.
:param tv_id: 电视节目ID
:param season_num: 季数
:return: 翻译信息列表
*/
func (s *Season) Translations(tvID, seasonNum int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/translations", tvID, seasonNum)
	key := "translations"

	result, err := s.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Videos 获取电视季视频信�?/*
Get the videos that have been added to a TV show.
:param tv_id: 电视节目ID
:param season_num: 季数
:param include_video_language: 包含的视频语言
:param page: 页码
:return: 视频信息列表
*/
func (s *Season) Videos(tvID, seasonNum int, includeVideoLanguage *string, page int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/videos", tvID, seasonNum)
	params := "page=" + strconv.Itoa(page)
	if includeVideoLanguage != nil {
		params += "&include_video_language=" + *includeVideoLanguage
	}

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
