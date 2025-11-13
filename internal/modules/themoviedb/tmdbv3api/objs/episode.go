package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Episode 集对�?type Episode struct {
	tmdb *tmdbv3api.TMDb
}

// NewEpisode 创建Episode实例
func NewEpisode(tmdb *tmdbv3api.TMDb) *Episode {
	return &Episode{
		tmdb: tmdb,
	}
}

// Details 获取TV剧集详情
/*
Get the TV episode details by id.
*/
func (e *Episode) Details(tvID, seasonNum, episodeNum int, appendToResponse string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d", tvID, seasonNum, episodeNum)
	params := fmt.Sprintf("append_to_response=%s", appendToResponse)
	
	result, err := e.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AccountStates 获取剧集的账户状�?/*
Get your rating for a episode.
*/
func (e *Episode) AccountStates(tvID, seasonNum, episodeNum int) (map[string]interface{}, error) {
	sessionID, err := e.tmdb.SessionID()
	if err != nil {
		return nil, err
	}
	
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/account_states", tvID, seasonNum, episodeNum)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	result, err := e.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Changes 获取TV剧集的更改记�?/*
Get the changes for a TV episode. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (e *Episode) Changes(episodeID int, startDate, endDate string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/episode/%d/changes", episodeID)
	params := fmt.Sprintf("page=%d", page)
	if startDate != "" {
		params += fmt.Sprintf("&start_date=%s", startDate)
	}
	if endDate != "" {
		params += fmt.Sprintf("&end_date=%s", endDate)
	}
	key := "changes"
	
	result, err := e.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Credits 获取TV剧集的演职人员信�?/*
Get the credits for TV season.
*/
func (e *Episode) Credits(tvID, seasonNum, episodeNum int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/credits", tvID, seasonNum, episodeNum)
	
	result, err := e.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ExternalIDs 获取TV剧集的外部ID
/*
Get the external ids for a TV episode.
*/
func (e *Episode) ExternalIDs(tvID, seasonNum, episodeNum int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/external_ids", tvID, seasonNum, episodeNum)
	
	result, err := e.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取TV剧集的图�?/*
Get the images that belong to a TV episode.
*/
func (e *Episode) Images(tvID, seasonNum, episodeNum int, includeImageLanguage string) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/images", tvID, seasonNum, episodeNum)
	params := ""
	if includeImageLanguage != "" {
		params = fmt.Sprintf("include_image_language=%s", includeImageLanguage)
	}
	key := "stills"
	
	result, err := e.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Translations 获取剧集的翻译数�?/*
Get the translation data for an episode.
*/
func (e *Episode) Translations(tvID, seasonNum, episodeNum int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/translations", tvID, seasonNum, episodeNum)
	key := "translations"
	
	result, err := e.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// RateTVEpisode 为TV剧集评分
/*
Rate a TV episode.
*/
func (e *Episode) RateTVEpisode(tvID, seasonNum, episodeNum int, rating float64) error {
	sessionID, err := e.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/rating", tvID, seasonNum, episodeNum)
	params := fmt.Sprintf("session_id=%s", sessionID)
	jsonData := map[string]interface{}{
		"value": rating,
	}
	
	_, err = e.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	return err
}

// DeleteRating 删除TV剧集的评�?/*
Remove your rating for a TV episode.
*/
func (e *Episode) DeleteRating(tvID, seasonNum, episodeNum int) error {
	sessionID, err := e.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/rating", tvID, seasonNum, episodeNum)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	_, err = e.tmdb.RequestObj(action, params, "DELETE", nil, nil, nil)
	return err
}

// Videos 获取TV剧集的视�?/*
Get the videos that have been added to a TV episode.
*/
func (e *Episode) Videos(tvID, seasonNum, episodeNum int, includeVideoLanguage string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/tv/%d/season/%d/episode/%d/videos", tvID, seasonNum, episodeNum)
	params := ""
	if includeVideoLanguage != "" {
		params = fmt.Sprintf("include_video_language=%s", includeVideoLanguage)
	}
	
	result, err := e.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
