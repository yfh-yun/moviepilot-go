package tmdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// DiscoverMovies 发现电影
func (s *TMDBService) DiscoverMovies(ctx context.Context, params *DiscoverParams) (*MovieSearchResponse, error) {
	endpoint := s.buildDiscoverEndpoint("movie", params)
	var result MovieSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 30*time.Minute)
	if err != nil {
		logger.Error("Failed to discover movies", zap.Error(err))
		return nil, err
	}
	return &result, nil
}

// DiscoverTVShows 发现电视剧
func (s *TMDBService) DiscoverTVShows(ctx context.Context, params *DiscoverParams) (*TVSearchResponse, error) {
	endpoint := s.buildDiscoverEndpoint("tv", params)
	var result TVSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 30*time.Minute)
	if err != nil {
		logger.Error("Failed to discover TV shows", zap.Error(err))
		return nil, err
	}
	return &result, nil
}

// buildDiscoverEndpoint 构建发现API端点
func (s *TMDBService) buildDiscoverEndpoint(mediaType string, params *DiscoverParams) string {
	if params == nil {
		params = &DiscoverParams{}
	}

	baseURL := fmt.Sprintf("/discover/%s", mediaType)
	queryParams := []string{}

	// 添加查询参数
	if params.Language != "" {
		queryParams = append(queryParams, fmt.Sprintf("language=%s", params.Language))
	}
	if params.Region != "" {
		queryParams = append(queryParams, fmt.Sprintf("region=%s", params.Region))
	}
	if params.SortBy != "" {
		queryParams = append(queryParams, fmt.Sprintf("sort_by=%s", params.SortBy))
	}
	if params.CertificationCountry != "" {
		queryParams = append(queryParams, fmt.Sprintf("certification_country=%s", params.CertificationCountry))
	}
	if params.Certification != "" {
		queryParams = append(queryParams, fmt.Sprintf("certification=%s", params.Certification))
	}
	if params.CertificationLTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("certification.lte=%s", params.CertificationLTE))
	}
	if params.CertificationGTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("certification.gte=%s", params.CertificationGTE))
	}
	queryParams = append(queryParams, fmt.Sprintf("include_adult=%t", params.IncludeAdult))
	queryParams = append(queryParams, fmt.Sprintf("include_video=%t", params.IncludeVideo))
	if params.PrimaryReleaseYear > 0 {
		queryParams = append(queryParams, fmt.Sprintf("primary_release_year=%d", params.PrimaryReleaseYear))
	}
	if params.PrimaryReleaseDateGTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("primary_release_date.gte=%s", params.PrimaryReleaseDateGTE))
	}
	if params.PrimaryReleaseDateLTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("primary_release_date.lte=%s", params.PrimaryReleaseDateLTE))
	}
	if params.ReleaseDateGTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("release_date.gte=%s", params.ReleaseDateGTE))
	}
	if params.ReleaseDateLTE != "" {
		queryParams = append(queryParams, fmt.Sprintf("release_date.lte=%s", params.ReleaseDateLTE))
	}
	if params.WithReleaseType != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_release_type=%s", params.WithReleaseType))
	}
	if params.Year > 0 {
		queryParams = append(queryParams, fmt.Sprintf("year=%d", params.Year))
	}
	if params.WithGenres != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_genres=%s", params.WithGenres))
	}
	if params.WithCast != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_cast=%s", params.WithCast))
	}
	if params.WithCrew != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_crew=%s", params.WithCrew))
	}
	if params.WithPeople != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_people=%s", params.WithPeople))
	}
	if params.WithCompanies != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_companies=%s", params.WithCompanies))
	}
	if params.WithNetworks != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_networks=%s", params.WithNetworks))
	}
	if params.WatchRegion != "" {
		queryParams = append(queryParams, fmt.Sprintf("watch_region=%s", params.WatchRegion))
	}
	if params.WithWatchProviders != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_watch_providers=%s", params.WithWatchProviders))
	}
	if len(params.WatchMonetizationTypes) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("watch_monetization_types=%s", strings.Join(params.WatchMonetizationTypes, ",")))
	}
	if params.WithoutWatchProviders != "" {
		queryParams = append(queryParams, fmt.Sprintf("without_watch_providers=%s", params.WithoutWatchProviders))
	}
	if params.WithoutGenres != "" {
		queryParams = append(queryParams, fmt.Sprintf("without_genres=%s", params.WithoutGenres))
	}
	if params.WithoutKeywords != "" {
		queryParams = append(queryParams, fmt.Sprintf("without_keywords=%s", params.WithoutKeywords))
	}
	if params.WithKeywords != "" {
		queryParams = append(queryParams, fmt.Sprintf("with_keywords=%s", params.WithKeywords))
	}
	if params.WithRuntimeGTE > 0 {
		queryParams = append(queryParams, fmt.Sprintf("with_runtime.gte=%d", params.WithRuntimeGTE))
	}
	if params.WithRuntimeLTE > 0 {
		queryParams = append(queryParams, fmt.Sprintf("with_runtime.lte=%d", params.WithRuntimeLTE))
	}
	if params.Page > 0 {
		queryParams = append(queryParams, fmt.Sprintf("page=%d", params.Page))
	}

	// 构建完整URL
	if len(queryParams) > 0 {
		return fmt.Sprintf("%s?%s", baseURL, strings.Join(queryParams, "&"))
	}
	return baseURL
}

// NewMovieDiscoverParams 创建电影发现参数
func NewMovieDiscoverParams() *DiscoverParams {
	return &DiscoverParams{
		Language:     "en-US",
		SortBy:       "popularity.desc",
		IncludeAdult: false,
		IncludeVideo: false,
		Page:         1,
	}
}

// NewTVDiscoverParams 创建电视剧发现参数
func NewTVDiscoverParams() *DiscoverParams {
	return &DiscoverParams{
		Language:     "en-US",
		SortBy:       "popularity.desc",
		IncludeAdult: false,
		Page:         1,
	}
}

// DiscoverParamsBuilder 发现参数构建器
type DiscoverParamsBuilder struct {
	params *DiscoverParams
}

// NewDiscoverParamsBuilder 创建参数构建器
func NewDiscoverParamsBuilder(mediaType string) *DiscoverParamsBuilder {
	var params *DiscoverParams
	if mediaType == "movie" {
		params = NewMovieDiscoverParams()
	} else {
		params = NewTVDiscoverParams()
	}

	return &DiscoverParamsBuilder{params: params}
}

// Language 设置语言
func (b *DiscoverParamsBuilder) Language(language string) *DiscoverParamsBuilder {
	b.params.Language = language
	return b
}

// Region 设置地区
func (b *DiscoverParamsBuilder) Region(region string) *DiscoverParamsBuilder {
	b.params.Region = region
	return b
}

// SortBy 设置排序
func (b *DiscoverParamsBuilder) SortBy(sortBy string) *DiscoverParamsBuilder {
	b.params.SortBy = sortBy
	return b
}

// Genres 设置类型
func (b *DiscoverParamsBuilder) Genres(genres []int) *DiscoverParamsBuilder {
	genreStrs := make([]string, len(genres))
	for i, genre := range genres {
		genreStrs[i] = strconv.Itoa(genre)
	}
	b.params.WithGenres = strings.Join(genreStrs, ",")
	return b
}

// ReleaseYear 设置发行年份
func (b *DiscoverParamsBuilder) ReleaseYear(year int) *DiscoverParamsBuilder {
	b.params.PrimaryReleaseYear = year
	b.params.Year = year
	return b
}

// ReleaseDateRange 设置发行日期范围
func (b *DiscoverParamsBuilder) ReleaseDateRange(gte, lte string) *DiscoverParamsBuilder {
	b.params.ReleaseDateGTE = gte
	b.params.ReleaseDateLTE = lte
	return b
}

// RuntimeRange 设置运行时间范围
func (b *DiscoverParamsBuilder) RuntimeRange(gte, lte int) *DiscoverParamsBuilder {
	b.params.WithRuntimeGTE = gte
	b.params.WithRuntimeLTE = lte
	return b
}

// Certification 设置分级
func (b *DiscoverParamsBuilder) Certification(certification, country string) *DiscoverParamsBuilder {
	b.params.Certification = certification
	b.params.CertificationCountry = country
	return b
}

// IncludeAdult 设置是否包含成人内容
func (b *DiscoverParamsBuilder) IncludeAdult(include bool) *DiscoverParamsBuilder {
	b.params.IncludeAdult = include
	return b
}

// Page 设置页码
func (b *DiscoverParamsBuilder) Page(page int) *DiscoverParamsBuilder {
	b.params.Page = page
	return b
}

// Cast 设置演员
func (b *DiscoverParamsBuilder) Cast(castIDs []int) *DiscoverParamsBuilder {
	castStrs := make([]string, len(castIDs))
	for i, id := range castIDs {
		castStrs[i] = strconv.Itoa(id)
	}
	b.params.WithCast = strings.Join(castStrs, ",")
	return b
}

// Crew 设置工作人员
func (b *DiscoverParamsBuilder) Crew(crewIDs []int) *DiscoverParamsBuilder {
	crewStrs := make([]string, len(crewIDs))
	for i, id := range crewIDs {
		crewStrs[i] = strconv.Itoa(id)
	}
	b.params.WithCrew = strings.Join(crewStrs, ",")
	return b
}

// Companies 设置制作公司
func (b *DiscoverParamsBuilder) Companies(companyIDs []int) *DiscoverParamsBuilder {
	companyStrs := make([]string, len(companyIDs))
	for i, id := range companyIDs {
		companyStrs[i] = strconv.Itoa(id)
	}
	b.params.WithCompanies = strings.Join(companyStrs, ",")
	return b
}

// Networks 设置电视网络
func (b *DiscoverParamsBuilder) Networks(networkIDs []int) *DiscoverParamsBuilder {
	networkStrs := make([]string, len(networkIDs))
	for i, id := range networkIDs {
		networkStrs[i] = strconv.Itoa(id)
	}
	b.params.WithNetworks = strings.Join(networkStrs, ",")
	return b
}

// Keywords 设置关键词
func (b *DiscoverParamsBuilder) Keywords(keywordIDs []int) *DiscoverParamsBuilder {
	keywordStrs := make([]string, len(keywordIDs))
	for i, id := range keywordIDs {
		keywordStrs[i] = strconv.Itoa(id)
	}
	b.params.WithKeywords = strings.Join(keywordStrs, ",")
	return b
}

// WithoutGenres 排除类型
func (b *DiscoverParamsBuilder) WithoutGenres(genres []int) *DiscoverParamsBuilder {
	genreStrs := make([]string, len(genres))
	for i, genre := range genres {
		genreStrs[i] = strconv.Itoa(genre)
	}
	b.params.WithoutGenres = strings.Join(genreStrs, ",")
	return b
}

// WithoutKeywords 排除关键词
func (b *DiscoverParamsBuilder) WithoutKeywords(keywords []int) *DiscoverParamsBuilder {
	keywordStrs := make([]string, len(keywords))
	for i, id := range keywords {
		keywordStrs[i] = strconv.Itoa(id)
	}
	b.params.WithoutKeywords = strings.Join(keywordStrs, ",")
	return b
}

// WatchProviders 设置观看提供商
func (b *DiscoverParamsBuilder) WatchProviders(providers []int, region string, monetizationTypes []string) *DiscoverParamsBuilder {
	if region != "" {
		b.params.WatchRegion = region
	}

	if len(providers) > 0 {
		providerStrs := make([]string, len(providers))
		for i, id := range providers {
			providerStrs[i] = strconv.Itoa(id)
		}
		b.params.WithWatchProviders = strings.Join(providerStrs, ",")
	}

	if len(monetizationTypes) > 0 {
		b.params.WatchMonetizationTypes = monetizationTypes
	}

	return b
}

// Build 构建参数
func (b *DiscoverParamsBuilder) Build() *DiscoverParams {
	return b.params
}

// Validate 验证参数
func (p *DiscoverParams) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.WithRuntimeGTE > p.WithRuntimeLTE && p.WithRuntimeLTE > 0 {
		return fmt.Errorf("runtime GTE cannot be greater than LTE")
	}

	if p.WithRuntimeGTE < 0 {
		p.WithRuntimeGTE = 0
	}

	if p.WithRuntimeLTE < 0 {
		p.WithRuntimeLTE = 0
	}

	// 验证日期格式
	if p.ReleaseDateGTE != "" {
		if !isValidDate(p.ReleaseDateGTE) {
			return fmt.Errorf("invalid release date GTE format: %s", p.ReleaseDateGTE)
		}
	}

	if p.ReleaseDateLTE != "" {
		if !isValidDate(p.ReleaseDateLTE) {
			return fmt.Errorf("invalid release date LTE format: %s", p.ReleaseDateLTE)
		}
	}

	return nil
}

// isValidDate 验证日期格式 (YYYY-MM-DD)
func isValidDate(date string) bool {
	if len(date) != 10 {
		return false
	}

	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return false
	}

	year, err1 := strconv.Atoi(parts[0])
	month, err2 := strconv.Atoi(parts[1])
	day, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}

	if year < 1900 || year > 2100 {
		return false
	}

	if month < 1 || month > 12 {
		return false
	}

	if day < 1 || day > 31 {
		return false
	}

	return true
}
