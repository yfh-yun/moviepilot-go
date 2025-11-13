package themoviedb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/pkg/config"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

const (
	BaseURL = "https://api.themoviedb.org/3/"
)

// TMDb API结构�?type TMDb struct {
	APIKey   string
	Language string
	Proxy    map[string]string
	Timeout  int
}

// Search 搜索结构�?type Search struct {
	tmdb *TMDb
}

// Movie 电影结构�?type Movie struct {
	tmdb *TMDb
}

// TV 电视剧结构体
type TV struct {
	tmdb *TMDb
}

// Season 季结构体
type Season struct {
	tmdb *TMDb
}

// Episode 集结构体
type Episode struct {
	tmdb *TMDb
}

// Discover 发现结构�?type Discover struct {
	tmdb *TMDb
}

// Trending 趋势结构�?type Trending struct {
	tmdb *TMDb
}

// Person 人物结构�?type Person struct {
	tmdb *TMDb
}

// Collection 合集结构�?type Collection struct {
	tmdb *TMDb
}

// TmdbApi TMDB API主结构体
type TmdbApi struct {
	tmdb      *TMDb
	search    *Search
	movie     *Movie
	tv        *TV
	seasonObj *Season
	episodeObj *Episode
	discover  *Discover
	trending  *Trending
	person    *Person
	collection *Collection
}

// NewTMDb 创建TMDb实例
func NewTMDb(language string) *TMDb {
	return &TMDb{
		APIKey:   config.Config.TMDB_API_KEY,
		Language: language,
		Timeout:  15,
	}
}

// NewSearch 创建Search实例
func NewSearch(language string) *Search {
	return &Search{
		tmdb: NewTMDb(language),
	}
}

// NewMovie 创建Movie实例
func NewMovie(language string) *Movie {
	return &Movie{
		tmdb: NewTMDb(language),
	}
}

// NewTV 创建TV实例
func NewTV(language string) *TV {
	return &TV{
		tmdb: NewTMDb(language),
	}
}

// NewSeason 创建Season实例
func NewSeason(language string) *Season {
	return &Season{
		tmdb: NewTMDb(language),
	}
}

// NewEpisode 创建Episode实例
func NewEpisode(language string) *Episode {
	return &Episode{
		tmdb: NewTMDb(language),
	}
}

// NewDiscover 创建Discover实例
func NewDiscover(language string) *Discover {
	return &Discover{
		tmdb: NewTMDb(language),
	}
}

// NewTrending 创建Trending实例
func NewTrending(language string) *Trending {
	return &Trending{
		tmdb: NewTMDb(language),
	}
}

// NewPerson 创建Person实例
func NewPerson(language string) *Person {
	return &Person{
		tmdb: NewTMDb(language),
	}
}

// NewCollection 创建Collection实例
func NewCollection(language string) *Collection {
	return &Collection{
		tmdb: NewTMDb(language),
	}
}

// NewTmdbApi 创建TmdbApi实例
func NewTmdbApi(language string) *TmdbApi {
	tmdb := NewTMDb(language)
	return &TmdbApi{
		tmdb:      tmdb,
		search:    &Search{tmdb: tmdb},
		movie:     &Movie{tmdb: tmdb},
		tv:        &TV{tmdb: tmdb},
		seasonObj: &Season{tmdb: tmdb},
		episodeObj: &Episode{tmdb: tmdb},
		discover:  &Discover{tmdb: tmdb},
		trending:  &Trending{tmdb: tmdb},
		person:    &Person{tmdb: tmdb},
		collection: &Collection{tmdb: tmdb},
	}
}

// makeRequest 发起API请求
func (t *TMDb) makeRequest(endpoint string, params map[string]string) (map[string]interface{}, error) {
	// 构建URL
	apiURL := BaseURL + endpoint
	
	// 添加基础参数
	if params == nil {
		params = make(map[string]string)
	}
	params["api_key"] = t.APIKey
	if t.Language != "" {
		params["language"] = t.Language
	}
	
	// 构建查询字符�?	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	
	fullURL := apiURL + "?" + query.Encode()
	
	// 设置请求�?	headers := make(map[string]string)
	
	// 发起请求
	response, err := utils.RequestUtils.GetRes(fullURL, headers, t.Proxy, t.Timeout)
	if err != nil {
		return nil, fmt.Errorf("请求TMDB API失败: %v", err)
	}
	
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API返回错误状态码: %d", response.StatusCode)
	}
	
	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, fmt.Errorf("解析TMDB响应失败: %v", err)
	}
	
	return result, nil
}

// SearchMulti 搜索多个媒体类型
func (s *Search) SearchMulti(term string) ([]map[string]interface{}, error) {
	params := map[string]string{
		"query": term,
	}
	
	result, err := s.tmdb.makeRequest("search/multi", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var multiResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				multiResults = append(multiResults, itemMap)
			}
		}
		return multiResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// SearchMovies 搜索电影
func (s *Search) SearchMovies(term string, year string) ([]map[string]interface{}, error) {
	params := map[string]string{
		"query": term,
	}
	
	if year != "" {
		params["year"] = year
	}
	
	result, err := s.tmdb.makeRequest("search/movie", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var movieResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				movieResults = append(movieResults, itemMap)
			}
		}
		return movieResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// SearchTVShows 搜索电视�?func (s *Search) SearchTVShows(term string, releaseYear string) ([]map[string]interface{}, error) {
	params := map[string]string{
		"query": term,
	}
	
	if releaseYear != "" {
		params["first_air_date_year"] = releaseYear
	}
	
	result, err := s.tmdb.makeRequest("search/tv", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var tvResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				tvResults = append(tvResults, itemMap)
			}
		}
		return tvResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// SearchPeople 搜索人物
func (s *Search) SearchPeople(term string) ([]map[string]interface{}, error) {
	params := map[string]string{
		"query": term,
	}
	
	result, err := s.tmdb.makeRequest("search/person", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var personResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				personResults = append(personResults, itemMap)
			}
		}
		return personResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// SearchCollections 搜索合集
func (s *Search) SearchCollections(term string) ([]map[string]interface{}, error) {
	params := map[string]string{
		"query": term,
	}
	
	result, err := s.tmdb.makeRequest("search/collection", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var collectionResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				collectionResults = append(collectionResults, itemMap)
			}
		}
		return collectionResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// GetMovieDetails 获取电影详情
func (m *Movie) GetMovieDetails(movieID int, appendToResponse string) (map[string]interface{}, error) {
	endpoint := "movie/" + strconv.Itoa(movieID)
	params := make(map[string]string)
	
	if appendToResponse != "" {
		params["append_to_response"] = appendToResponse
	}
	
	return m.tmdb.makeRequest(endpoint, params)
}

// GetTVDetails 获取电视剧详�?func (t *TV) GetTVDetails(tvID int, appendToResponse string) (map[string]interface{}, error) {
	endpoint := "tv/" + strconv.Itoa(tvID)
	params := make(map[string]string)
	
	if appendToResponse != "" {
		params["append_to_response"] = appendToResponse
	}
	
	return t.tmdb.makeRequest(endpoint, params)
}

// GetSeasonDetails 获取季详�?func (s *Season) GetSeasonDetails(tvID int, seasonNum int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("tv/%d/season/%d", tvID, seasonNum)
	return s.tmdb.makeRequest(endpoint, nil)
}

// GetEpisodeDetails 获取集详�?func (e *Episode) GetEpisodeDetails(tvID int, seasonNum int, episodeNum int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("tv/%d/season/%d/episode/%d", tvID, seasonNum, episodeNum)
	return e.tmdb.makeRequest(endpoint, nil)
}

// GetCollectionDetails 获取合集详情
func (c *Collection) GetCollectionDetails(collectionID int) (map[string]interface{}, error) {
	endpoint := "collection/" + strconv.Itoa(collectionID)
	return c.tmdb.makeRequest(endpoint, nil)
}

// GetPersonDetails 获取人物详情
func (p *Person) GetPersonDetails(personID int) (map[string]interface{}, error) {
	endpoint := "person/" + strconv.Itoa(personID)
	return p.tmdb.makeRequest(endpoint, nil)
}

// GetMovieCredits 获取电影演职�?func (p *Person) GetMovieCredits(personID int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("person/%d/movie_credits", personID)
	return p.tmdb.makeRequest(endpoint, nil)
}

// GetTVCredits 获取电视剧演职员
func (p *Person) GetTVCredits(personID int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("person/%d/tv_credits", personID)
	return p.tmdb.makeRequest(endpoint, nil)
}

// DiscoverMovies 发现电影
func (d *Discover) DiscoverMovies(paramsTuple []interface{}) ([]map[string]interface{}, error) {
	params := make(map[string]string)
	
	// 转换参数元组为map
	for i := 0; i < len(paramsTuple); i += 2 {
		if i+1 < len(paramsTuple) {
			key := fmt.Sprintf("%v", paramsTuple[i])
			value := fmt.Sprintf("%v", paramsTuple[i+1])
			params[key] = value
		}
	}
	
	result, err := d.tmdb.makeRequest("discover/movie", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var movieResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				movieResults = append(movieResults, itemMap)
			}
		}
		return movieResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// DiscoverTVShows 发现电视�?func (d *Discover) DiscoverTVShows(paramsTuple []interface{}) ([]map[string]interface{}, error) {
	params := make(map[string]string)
	
	// 转换参数元组为map
	for i := 0; i < len(paramsTuple); i += 2 {
		if i+1 < len(paramsTuple) {
			key := fmt.Sprintf("%v", paramsTuple[i])
			value := fmt.Sprintf("%v", paramsTuple[i+1])
			params[key] = value
		}
	}
	
	result, err := d.tmdb.makeRequest("discover/tv", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var tvResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				tvResults = append(tvResults, itemMap)
			}
		}
		return tvResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// TrendingAllWeek 获取周趋�?func (t *Trending) TrendingAllWeek(page int) ([]map[string]interface{}, error) {
	params := make(map[string]string)
	if page > 0 {
		params["page"] = strconv.Itoa(page)
	}
	
	result, err := t.tmdb.makeRequest("trending/all/week", params)
	if err != nil {
		return nil, err
	}
	
	if results, ok := result["results"].([]interface{}); ok {
		var trendingResults []map[string]interface{}
		for _, item := range results {
			if itemMap, ok := item.(map[string]interface{}); ok {
				trendingResults = append(trendingResults, itemMap)
			}
		}
		return trendingResults, nil
	}
	
	return []map[string]interface{}{}, nil
}

// Match 根据名称匹配媒体信息
func (t *TmdbApi) Match(name string, mtype models.MediaType, year string, seasonYear string, seasonNumber int, groupSeasons []map[string]interface{}) map[string]interface{} {
	if mtype != models.MediaTypeTV {
		yearRange := t.generateYearRange(year)
		for _, searchYear := range yearRange {
			info := t.searchMovieByName(name, searchYear)
			if len(info) > 0 {
				return info
			}
		}
	} else {
		// 有当前季和当前季集年份，使用精确匹配
		if seasonYear != "" && seasonNumber > 0 {
			info := t.searchTVBySeason(name, seasonYear, seasonNumber, groupSeasons)
			if len(info) > 0 {
				return info
			}
		}
		
		yearRange := t.generateYearRange(year)
		for _, searchYear := range yearRange {
			info := t.searchTVByName(name, searchYear)
			if len(info) > 0 {
				return info
			}
		}
	}
	
	return make(map[string]interface{})
}

// generateYearRange 生成年份范围
func (t *TmdbApi) generateYearRange(year string) []string {
	yearRange := []string{year}
	if year != "" {
		yearInt, err := strconv.Atoi(year)
		if err == nil {
			yearRange = append(yearRange, strconv.Itoa(yearInt+1))
			yearRange = append(yearRange, strconv.Itoa(yearInt-1))
		}
	}
	return yearRange
}

// searchMovieByName 根据名称搜索电影
func (t *TmdbApi) searchMovieByName(name string, year string) map[string]interface{} {
	var movies []map[string]interface{}
	var err error
	
	if year != "" {
		movies, err = t.search.SearchMovies(name, year)
	} else {
		movies, err = t.search.SearchMovies(name, "")
	}
	
	if err != nil || len(movies) == 0 {
		return make(map[string]interface{})
	}
	
	// 按年份降序排�?	sort.Slice(movies, func(i, j int) bool {
		date1 := ""
		if releaseDate, ok := movies[i]["release_date"].(string); ok {
			date1 = releaseDate
		}
		
		date2 := ""
		if releaseDate, ok := movies[j]["release_date"].(string); ok {
			date2 = releaseDate
		}
		
		return date1 > date2
	})
	
	for _, movie := range movies {
		// 年份
		movieYear := ""
		if releaseDate, ok := movie["release_date"].(string); ok && len(releaseDate) >= 4 {
			movieYear = releaseDate[0:4]
		}
		
		if year != "" && movieYear != year {
			// 年份不匹�?			continue
		}
		
		// 匹配标题、原标题
		if t.compareNames(name, movie["title"]) {
			return movie
		}
		
		if t.compareNames(name, movie["original_title"]) {
			return movie
		}
	}
	
	return make(map[string]interface{})
}

// searchTVByName 根据名称搜索电视�?func (t *TmdbApi) searchTVByName(name string, year string) map[string]interface{} {
	var tvs []map[string]interface{}
	var err error
	
	if year != "" {
		tvs, err = t.search.SearchTVShows(name, year)
	} else {
		tvs, err = t.search.SearchTVShows(name, "")
	}
	
	if err != nil || len(tvs) == 0 {
		return make(map[string]interface{})
	}
	
	// 按年份降序排�?	sort.Slice(tvs, func(i, j int) bool {
		date1 := ""
		if firstAirDate, ok := tvs[i]["first_air_date"].(string); ok {
			date1 = firstAirDate
		}
		
		date2 := ""
		if firstAirDate, ok := tvs[j]["first_air_date"].(string); ok {
			date2 = firstAirDate
		}
		
		return date1 > date2
	})
	
	for _, tv := range tvs {
		tvYear := ""
		if firstAirDate, ok := tv["first_air_date"].(string); ok && len(firstAirDate) >= 4 {
			tvYear = firstAirDate[0:4]
		}
		
		if year != "" && tvYear != year {
			// 年份不匹�?			continue
		}
		
		// 匹配标题、原标题
		if t.compareNames(name, tv["name"]) {
			return tv
		}
		
		if t.compareNames(name, tv["original_name"]) {
			return tv
		}
	}
	
	return make(map[string]interface{})
}

// searchTVBySeason 根据电视剧的名称和季的年份及序号匹配TMDB
func (t *TmdbApi) searchTVBySeason(name string, seasonYear string, seasonNumber int, groupSeasons []map[string]interface{}) map[string]interface{} {
	tvs, err := t.search.SearchTVShows(name, "")
	if err != nil || len(tvs) == 0 {
		return make(map[string]interface{})
	}
	
	// 按年份降序排�?	sort.Slice(tvs, func(i, j int) bool {
		date1 := ""
		if firstAirDate, ok := tvs[i]["first_air_date"].(string); ok {
			date1 = firstAirDate
		}
		
		date2 := ""
		if firstAirDate, ok := tvs[j]["first_air_date"].(string); ok {
			date2 = firstAirDate
		}
		
		return date1 > date2
	})
	
	for _, tv := range tvs {
		// 年份
		tvYear := ""
		if firstAirDate, ok := tv["first_air_date"].(string); ok && len(firstAirDate) >= 4 {
			tvYear = firstAirDate[0:4]
		}
		
		if (t.compareNames(name, tv["name"]) || t.compareNames(name, tv["original_name"])) && (tvYear == seasonYear) {
			return tv
		}
	}
	
	return make(map[string]interface{})
}

// compareNames 比较文件名是否匹配，忽略大小写和特殊字符
func (t *TmdbApi) compareNames(fileName string, tmdbNames interface{}) bool {
	if fileName == "" || tmdbNames == nil {
		return false
	}
	
	var namesList []string
	switch v := tmdbNames.(type) {
	case string:
		namesList = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				namesList = append(namesList, str)
			}
		}
	case []string:
		namesList = v
	default:
		return false
	}
	
	fileName = strings.ToUpper(strings.TrimSpace(fileName))
	
	for _, tmdbName := range namesList {
		tmdbName = strings.ToUpper(strings.TrimSpace(tmdbName))
		if fileName == tmdbName {
			return true
		}
	}
	
	return false
}

// GetInfo 根据TMDB ID获取媒体信息
func (t *TmdbApi) GetInfo(mtype models.MediaType, tmdbid int) map[string]interface{} {
	if mtype == models.MediaTypeMovie {
		tmdbInfo, err := t.movie.GetMovieDetails(tmdbid, "images,credits,alternative_titles,translations,release_dates,external_ids")
		if err == nil && len(tmdbInfo) > 0 {
			tmdbInfo["media_type"] = "movie"
			return tmdbInfo
		}
	} else if mtype == models.MediaTypeTV {
		tmdbInfo, err := t.tv.GetTVDetails(tmdbid, "images,credits,alternative_titles,translations,content_ratings,external_ids,episode_groups")
		if err == nil && len(tmdbInfo) > 0 {
			tmdbInfo["media_type"] = "tv"
			return tmdbInfo
		}
	} else {
		// 未指定类型时，尝试两种类�?		tvInfo, tvErr := t.tv.GetTVDetails(tmdbid, "images,credits,alternative_titles,translations,content_ratings,external_ids,episode_groups")
		movieInfo, movieErr := t.movie.GetMovieDetails(tmdbid, "images,credits,alternative_titles,translations,release_dates,external_ids")
		
		if tvInfo != nil && len(tvInfo) > 0 && movieInfo != nil && len(movieInfo) > 0 {
			// 无法判断是电影还是电视剧
			return make(map[string]interface{})
		} else if tvInfo != nil && len(tvInfo) > 0 {
			tvInfo["media_type"] = "tv"
			return tvInfo
		} else if movieInfo != nil && len(movieInfo) > 0 {
			movieInfo["media_type"] = "movie"
			return movieInfo
		}
	}
	
	return make(map[string]interface{})
}

// SearchMultiis 同时查询模糊匹配的电影、电视剧TMDB信息
func (t *TmdbApi) SearchMultiis(title string) []map[string]interface{} {
	if title == "" {
		return []map[string]interface{}{}
	}
	
	retInfos := []map[string]interface{}{}
	multis, err := t.search.SearchMulti(title)
	if err != nil {
		return retInfos
	}
	
	for _, multi := range multis {
		if mediaType, ok := multi["media_type"].(string); ok {
			if mediaType == "movie" || mediaType == "tv" {
				retInfos = append(retInfos, multi)
			}
		}
	}
	
	return retInfos
}

// Close 关闭连接
func (t *TmdbApi) Close() {
	// 在Go中不需要特别关闭HTTP连接
}
