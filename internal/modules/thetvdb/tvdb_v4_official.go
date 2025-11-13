package thetvdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/pkg/config"
	"moviepilot-go/internal/utils"
)

const (
	BaseURL = "https://api4.thetvdb.com/v4/"
)

// Auth 认证结构�?type Auth struct {
	Token string
}

// Request 请求结构�?type Request struct {
	AuthToken string
	Links     map[string]interface{}
	Proxy     map[string]string
	Timeout   int
}

// Url URL构建结构�?type Url struct {
	BaseURL string
}

// TVDB 主类
type TVDB struct {
	URL     *Url
	Auth    *Auth
	Request *Request
}

// NewAuth 创建认证实例
func NewAuth(apikey, pin string, proxy map[string]string, timeout int) (*Auth, error) {
	loginURL := BaseURL + "login"
	
	loginInfo := map[string]string{
		"apikey": apikey,
	}
	
	if pin != "" {
		loginInfo["pin"] = pin
	}
	
	loginInfoBytes, err := json.Marshal(loginInfo)
	if err != nil {
		return nil, fmt.Errorf("TVDB认证失败: %v", err)
	}
	
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	
	response, err := utils.RequestUtils.PostRes(loginURL, loginInfoBytes, headers, proxy, timeout)
	if err != nil {
		return nil, fmt.Errorf("TVDB认证失败: %v", err)
	}
	
	if response == nil {
		return nil, fmt.Errorf("TVDB认证失败: 网络连接失败，未收到响应")
	}
	
	if response.StatusCode != http.StatusOK {
		var errorData map[string]interface{}
		if err := json.Unmarshal(response.Body, &errorData); err != nil {
			return nil, fmt.Errorf("TVDB认证失败: Code: %d, 响应解析失败�?v", response.StatusCode, err)
		}
		errorMsg := fmt.Sprintf("Code: %d, %v", response.StatusCode, errorData["message"])
		return nil, fmt.Errorf("TVDB认证失败: %s", errorMsg)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, fmt.Errorf("TVDB认证失败: 响应解析失败�?v", err)
	}
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("TVDB认证失败: 无法获取认证数据")
	}
	
	token, ok := data["token"].(string)
	if !ok {
		return nil, fmt.Errorf("TVDB认证失败: 无法获取认证令牌")
	}
	
	return &Auth{Token: token}, nil
}

// GetToken 获取认证令牌
func (a *Auth) GetToken() string {
	return a.Token
}

// NewRequest 创建请求实例
func NewRequest(authToken string, proxy map[string]string, timeout int) *Request {
	return &Request{
		AuthToken: authToken,
		Proxy:     proxy,
		Timeout:   timeout,
	}
}

// MakeRequest 发起请求
func (r *Request) MakeRequest(urlStr string, ifModifiedSince *time.Time) (interface{}, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + r.AuthToken,
	}
	
	if ifModifiedSince != nil {
		headers["If-Modified-Since"] = ifModifiedSince.String()
	}
	
	response, err := utils.RequestUtils.GetRes(urlStr, headers, r.Proxy, r.Timeout)
	if err != nil {
		return nil, fmt.Errorf("获取 %s 失败: %v", urlStr, err)
	}
	
	if response == nil {
		return nil, fmt.Errorf("获取 %s 失败: 网络连接失败", urlStr)
	}
	
	if response.StatusCode == http.StatusNotModified {
		return map[string]interface{}{
			"code":    http.StatusNotModified,
			"message": "Not-Modified",
		}, nil
	}
	
	if response.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(response.Body, &result); err != nil {
			return nil, fmt.Errorf("获取 %s 失败: 响应解析失败�?v", urlStr, err)
		}
		
		status, ok := result["status"].(string)
		if ok && status == "failure" {
			msg, _ := result["message"].(string)
			return nil, fmt.Errorf("获取 %s 失败: %s", urlStr, msg)
		}
		
		data, exists := result["data"]
		if exists {
			r.Links, _ = result["links"].(map[string]interface{})
			return data, nil
		}
		
		msg, _ := result["message"].(string)
		return nil, fmt.Errorf("获取 %s 失败: %s", urlStr, msg)
	}
	
	// 处理其他HTTP错误状态码
	var errorData map[string]interface{}
	if err := json.Unmarshal(response.Body, &errorData); err != nil {
		return nil, fmt.Errorf("获取 %s 失败: HTTP %d %v", urlStr, response.StatusCode, err)
	}
	
	msg, ok := errorData["message"].(string)
	if !ok {
		msg = fmt.Sprintf("HTTP %d", response.StatusCode)
	}
	
	return nil, fmt.Errorf("获取 %s 失败: %s", urlStr, msg)
}

// NewUrl 创建URL构建实例
func NewUrl() *Url {
	return &Url{
		BaseURL: BaseURL,
	}
}

// Construct 构建API URL
func (u *Url) Construct(urlSect string, urlID *int, urlSubsect, urlLang string, params map[string]interface{}) string {
	urlStr := u.BaseURL + urlSect
	
	if urlID != nil {
		urlStr += "/" + strconv.Itoa(*urlID)
	}
	
	if urlSubsect != "" {
		urlStr += "/" + urlSubsect
	}
	
	if urlLang != "" {
		urlStr += "/" + urlLang
	}
	
	if len(params) > 0 {
		var queryParams []string
		for key, value := range params {
			if value != nil {
				queryParams = append(queryParams, key+"="+url.QueryEscape(fmt.Sprintf("%v", value)))
			}
		}
		
		if len(queryParams) > 0 {
			urlStr += "?" + strings.Join(queryParams, "&")
		}
	}
	
	return urlStr
}

// NewTVDB 创建TVDB实例
func NewTVDB(apikey, pin string, proxy map[string]string, timeout int) (*TVDB, error) {
	tvdbURL := NewUrl()
	
	auth, err := NewAuth(apikey, pin, proxy, timeout)
	if err != nil {
		return nil, err
	}
	
	authToken := auth.GetToken()
	request := NewRequest(authToken, proxy, timeout)
	
	return &TVDB{
		URL:     tvdbURL,
		Auth:    auth,
		Request: request,
	}, nil
}

// GetReqLinks 获取请求链接信息
func (t *TVDB) GetReqLinks() map[string]interface{} {
	return t.Request.Links
}

// GetSeries 获取剧集信息
func (t *TVDB) GetSeries(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("series", &id, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetSeriesBySlug 通过 slug (别名) 返回单个剧集信息的字�?func (t *TVDB) GetSeriesBySlug(slug string, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("series/slug", nil, slug, "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetSeriesExtended 获取剧集扩展信息
func (t *TVDB) GetSeriesExtended(id int, meta *string, short bool, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	if short {
		params["short"] = true
	}
	
	urlStr := t.URL.Construct("series", &id, "extended", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetMovie 获取电影信息
func (t *TVDB) GetMovie(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("movies", &id, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetMovieExtended 获取电影扩展信息
func (t *TVDB) GetMovieExtended(id int, meta *string, short bool, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	if short {
		params["short"] = true
	}
	
	urlStr := t.URL.Construct("movies", &id, "extended", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetMovieBySlug 通过 slug (别名) 返回单个电影信息的字�?func (t *TVDB) GetMovieBySlug(slug string, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("movies/slug", nil, slug, "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetSeason 获取季信�?func (t *TVDB) GetSeason(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("seasons", &id, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetSeasonExtended 获取季扩展信�?func (t *TVDB) GetSeasonExtended(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("seasons", &id, "extended", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetEpisode 获取集信�?func (t *TVDB) GetEpisode(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("episodes", &id, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetEpisodeExtended 获取集扩展信�?func (t *TVDB) GetEpisodeExtended(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("episodes", &id, "extended", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetPerson 获取人物信息
func (t *TVDB) GetPerson(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("people", &id, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// GetPersonExtended 获取人物扩展信息
func (t *TVDB) GetPersonExtended(id int, meta *string, ifModifiedSince *time.Time) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if meta != nil {
		params["meta"] = *meta
	}
	
	urlStr := t.URL.Construct("people", &id, "extended", "", params)
	data, err := t.Request.MakeRequest(urlStr, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.(map[string]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// Search 搜索
func (t *TVDB) Search(query string, params map[string]interface{}) ([]interface{}, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	params["query"] = query
	
	urlStr := t.URL.Construct("search", nil, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, nil)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.([]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}

// SearchByRemoteID 通过外部 ID 精确匹配搜索，并返回结果列表
func (t *TVDB) SearchByRemoteID(remoteid string) ([]interface{}, error) {
	params := make(map[string]interface{})
	params["remoteid"] = remoteid
	
	urlStr := t.URL.Construct("search/remoteid", nil, "", "", params)
	data, err := t.Request.MakeRequest(urlStr, nil)
	if err != nil {
		return nil, err
	}
	
	if result, ok := data.([]interface{}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("返回的数据格式不正确")
}
