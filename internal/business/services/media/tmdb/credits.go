package tmdb

import (
	"context"
	"fmt"
	"strings"

	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// GetTopCast 获取主要演员（前N个）
func (s *TMDBService) GetTopCast(ctx context.Context, mediaType string, id int, language string, limit int) ([]Cast, error) {
	var credits *Credits
	var err error

	// 根据媒体类型获取演职员信息
	switch mediaType {
	case "movie":
		credits, err = s.GetMovieCredits(ctx, id, language)
	case "tv":
		credits, err = s.GetTVCredits(ctx, id, language)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}

	if err != nil {
		return nil, err
	}

	if len(credits.Cast) == 0 {
		return []Cast{}, nil
	}

	// 限制返回数量
	if limit <= 0 || limit > len(credits.Cast) {
		limit = len(credits.Cast)
	}

	logger.Debug("Getting top cast", zap.String("media_type", mediaType), zap.Int("id", id), zap.Int("limit", limit), zap.Int("total", len(credits.Cast)))

	return credits.Cast[:limit], nil
}

// GetTopCrew 获取主要工作人员（按部门分组）
func (s *TMDBService) GetTopCrew(ctx context.Context, mediaType string, id int, language string, departments []string) (map[string][]Crew, error) {
	var credits *Credits
	var err error

	// 根据媒体类型获取演职员信息
	switch mediaType {
	case "movie":
		credits, err = s.GetMovieCredits(ctx, id, language)
	case "tv":
		credits, err = s.GetTVCredits(ctx, id, language)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}

	if err != nil {
		return nil, err
	}

	if len(credits.Crew) == 0 {
		return map[string][]Crew{}, nil
	}

	// 如果没有指定部门，返回所有部门
	if len(departments) == 0 {
		departments = []string{"Directing", "Writing", "Production", "Editing", "Sound", "Art", "Camera"}
	}

	result := make(map[string][]Crew)
	deptSet := make(map[string]bool)
	for _, dept := range departments {
		deptSet[strings.ToLower(dept)] = true
	}

	// 按部门分组
	for _, crew := range credits.Crew {
		if deptSet[strings.ToLower(crew.Department)] {
			if _, exists := result[crew.Department]; !exists {
				result[crew.Department] = []Crew{}
			}
			result[crew.Department] = append(result[crew.Department], crew)
		}
	}

	logger.Debug("Getting top crew", zap.String("media_type", mediaType), zap.Int("id", id), zap.Strings("departments", departments), zap.Int("total_crew", len(credits.Crew)))

	return result, nil
}

// GetDirector 获取导演
func (s *TMDBService) GetDirector(ctx context.Context, mediaType string, id int, language string) ([]Crew, error) {
	crewMap, err := s.GetTopCrew(ctx, mediaType, id, language, []string{"Directing"})
	if err != nil {
		return nil, err
	}

	directingCrew, exists := crewMap["Directing"]
	if !exists {
		return []Crew{}, nil
	}

	// 过滤出导演
	var directors []Crew
	for _, crew := range directingCrew {
		if strings.Contains(strings.ToLower(crew.Job), "director") {
			directors = append(directors, crew)
		}
	}

	logger.Debug("Getting directors", zap.String("media_type", mediaType), zap.Int("id", id), zap.Int("count", len(directors)))

	return directors, nil
}

// GetWriters 获取编剧
func (s *TMDBService) GetWriters(ctx context.Context, mediaType string, id int, language string) ([]Crew, error) {
	crewMap, err := s.GetTopCrew(ctx, mediaType, id, language, []string{"Writing"})
	if err != nil {
		return nil, err
	}

	writingCrew, exists := crewMap["Writing"]
	if !exists {
		return []Crew{}, nil
	}

	// 过滤出编剧
	var writers []Crew
	writerJobs := []string{"writer", "screenplay", "story", "novel", "comic book"}
	for _, crew := range writingCrew {
		jobLower := strings.ToLower(crew.Job)
		for _, writerJob := range writerJobs {
			if strings.Contains(jobLower, writerJob) {
				writers = append(writers, crew)
				break
			}
		}
	}

	logger.Debug("Getting writers", zap.String("media_type", mediaType), zap.Int("id", id), zap.Int("count", len(writers)))

	return writers, nil
}

// GetPersonDetailsWithCredits 获取人物详情及作品
func (s *TMDBService) GetPersonDetailsWithCredits(ctx context.Context, id int, language string) (*PersonDetailsWithCredits, error) {
	// 并发获取人物详情和作品
	type result struct {
		details  *PersonDetails
		movies   *PersonMovieCredits
		tv       *PersonTVCredits
		combined *PersonCombinedCredits
		err      error
	}

	ch := make(chan result, 4)

	// 获取人物详情
	go func() {
		details, err := s.GetPersonDetails(ctx, id, language)
		ch <- result{details: details, err: err}
	}()

	// 获取电影作品
	go func() {
		movies, err := s.GetPersonMovieCredits(ctx, id, language)
		ch <- result{movies: movies, err: err}
	}()

	// 获取电视剧作品
	go func() {
		tv, err := s.GetPersonTVCredits(ctx, id, language)
		ch <- result{tv: tv, err: err}
	}()

	// 获取综合作品
	go func() {
		combined, err := s.GetPersonCombinedCredits(ctx, id, language)
		ch <- result{combined: combined, err: err}
	}()

	// 收集结果
	var details *PersonDetails
	var movies *PersonMovieCredits
	var tv *PersonTVCredits
	var combined *PersonCombinedCredits
	var errors []error

	for i := 0; i < 4; i++ {
		res := <-ch
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
		if res.details != nil {
			details = res.details
		}
		if res.movies != nil {
			movies = res.movies
		}
		if res.tv != nil {
			tv = res.tv
		}
		if res.combined != nil {
			combined = res.combined
		}
	}

	if len(errors) > 0 {
		logger.Error("Failed to get person details with credits", zap.Int("id", id), zap.Int("errors_count", len(errors)))
		return nil, fmt.Errorf("failed to get person details: %v", errors[0])
	}

	personResult := &PersonDetailsWithCredits{
		PersonDetails:   details,
		MovieCredits:    movies,
		TVCredits:       tv,
		CombinedCredits: combined,
	}

	logger.Debug("Got person details with credits", zap.Int("id", id), zap.String("name", details.Name))

	return personResult, nil
}

// PersonDetailsWithCredits 包含人物详情和作品的结构体
type PersonDetailsWithCredits struct {
	*PersonDetails
	MovieCredits    *PersonMovieCredits    `json:"movie_credits"`
	TVCredits       *PersonTVCredits       `json:"tv_credits"`
	CombinedCredits *PersonCombinedCredits `json:"combined_credits"`
}
