package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Movie 电影对象
type Movie struct {
	tmdb *tmdbv3api.TMDb
}

// NewMovie 创建Movie实例
func NewMovie(tmdb *tmdbv3api.TMDb) *Movie {
	return &Movie{
		tmdb: tmdb,
	}
}

// Details 获取电影的主要信�?/*
Get the primary information about a movie.
*/
func (m *Movie) Details(movieID int, appendToResponse string) (map[string]interface{}, error) {
	action := "/movie/" + strconv.Itoa(movieID)
	params := ""
	if appendToResponse != "" {
		params = "append_to_response=" + appendToResponse
	}
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AccountStates 获取会话的账户状�?/*
Grab the following account states for a session:
Movie rating, If it belongs to your watchlist, or If it belongs to your favourite list.
*/
func (m *Movie) AccountStates(movieID int) (map[string]interface{}, error) {
	sessionID, err := m.tmdb.SessionID()
	if err != nil {
		return nil, err
	}
	
	action := fmt.Sprintf("/movie/%d/account_states", movieID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AlternativeTitles 获取电影的所有备选标�?/*
Get all of the alternative titles for a movie.
*/
func (m *Movie) AlternativeTitles(movieID int, country string) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/alternative_titles", movieID)
	params := ""
	if country != "" {
		params = "country=" + country
	}
	key := "titles"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Changes 获取电影的更改记�?/*
Get the changes for a movie. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (m *Movie) Changes(movieID int, startDate, endDate string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/changes", movieID)
	params := fmt.Sprintf("page=%d", page)
	if startDate != "" {
		params += "&start_date=" + startDate
	}
	if endDate != "" {
		params += "&end_date=" + endDate
	}
	key := "changes"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Credits 获取电影的演员和工作人员
/*
Get the cast and crew for a movie.
*/
func (m *Movie) Credits(movieID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/credits", movieID)
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ExternalIDs 获取电影的外部ID
/*
Get the external ids for a movie.
*/
func (m *Movie) ExternalIDs(movieID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/external_ids", movieID)
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取电影的图�?/*
Get the images that belong to a movie.
Querying images with a language parameter will filter the results.
If you want to include a fallback language (especially useful for backdrops)
you can use the include_image_language parameter.
This should be a comma separated value like so: include_image_language=en,null.
*/
func (m *Movie) Images(movieID int, includeImageLanguage string) (map[string]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/images", movieID)
	params := ""
	if includeImageLanguage != "" {
		params = "include_image_language=" + includeImageLanguage
	}
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Keywords 获取与电影关联的关键�?/*
Get the keywords associated to a movie.
*/
func (m *Movie) Keywords(movieID int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/keywords", movieID)
	key := "keywords"
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Lists 获取该电影所属的列表
/*
Get a list of lists that this movie belongs to.
*/
func (m *Movie) Lists(movieID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/lists", movieID)
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Recommendations 获取推荐的电影列�?/*
Get a list of recommended movies for a movie.
*/
func (m *Movie) Recommendations(movieID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/recommendations", movieID)
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ReleaseDates 获取电影的发布日期和分级信息
/*
Get the release date along with the certification for a movie.
*/
func (m *Movie) ReleaseDates(movieID int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/release_dates", movieID)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Reviews 获取电影的用户评�?/*
Get the user reviews for a movie.
*/
func (m *Movie) Reviews(movieID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/reviews", movieID)
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Similar 获取相似电影列表
/*
Get a list of similar movies.
*/
func (m *Movie) Similar(movieID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/similar", movieID)
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Translations 获取电影的翻译列�?/*
Get a list of translations that have been created for a movie.
*/
func (m *Movie) Translations(movieID int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/translations", movieID)
	key := "translations"
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Videos 获取电影的视�?/*
Get the videos that have been added to a movie.
*/
func (m *Movie) Videos(movieID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/videos", movieID)
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// WatchProviders 获取电影的观看提供商信息
/*
You can query this method to get a list of the availabilities per country by provider.
*/
func (m *Movie) WatchProviders(movieID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/movie/%d/watch/providers", movieID)
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// RateMovie 为电影评�?/*
Rate a movie.
*/
func (m *Movie) RateMovie(movieID int, rating float64) error {
	sessionID, err := m.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/movie/%d/rating", movieID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	jsonData := map[string]interface{}{
		"value": rating,
	}
	
	_, err = m.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	return err
}

// DeleteRating 删除电影评分
/*
Remove your rating for a movie.
*/
func (m *Movie) DeleteRating(movieID int) error {
	sessionID, err := m.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/movie/%d/rating", movieID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	_, err = m.tmdb.RequestObj(action, params, "DELETE", nil, nil, nil)
	return err
}

// Latest 获取最新创建的电影
/*
Get the most newly created movie. This is a live response and will continuously change.
*/
func (m *Movie) Latest() (map[string]interface{}, error) {
	action := "/movie/latest"
	
	result, err := m.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// NowPlaying 获取正在影院上映的电影列�?/*
Get a list of movies in theatres.
*/
func (m *Movie) NowPlaying(region string, page int) ([]interface{}, error) {
	action := "/movie/now_playing"
	params := fmt.Sprintf("page=%d", page)
	if region != "" {
		params += "&region=" + region
	}
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Popular 获取当前热门电影列表
/*
Get a list of the current popular movies on TMDb. This list updates daily.
*/
func (m *Movie) Popular(region string, page int) ([]interface{}, error) {
	action := "/movie/popular"
	params := fmt.Sprintf("page=%d", page)
	if region != "" {
		params += "&region=" + region
	}
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TopRated 获取评分最高的电影
/*
Get the top rated movies on TMDb.
*/
func (m *Movie) TopRated(region string, page int) ([]interface{}, error) {
	action := "/movie/top_rated"
	params := fmt.Sprintf("page=%d", page)
	if region != "" {
		params += "&region=" + region
	}
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Upcoming 获取即将上映的电�?/*
Get a list of upcoming movies in theatres.
*/
func (m *Movie) Upcoming(region string, page int) ([]interface{}, error) {
	action := "/movie/upcoming"
	params := fmt.Sprintf("page=%d", page)
	if region != "" {
		params += "&region=" + region
	}
	key := "results"
	
	result, err := m.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
