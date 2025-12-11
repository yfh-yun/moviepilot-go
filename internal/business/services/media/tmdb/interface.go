package tmdb

import (
	"context"
)

// Service 定义 TMDB 服务接口
type Service interface {
	// 搜索相关
	MultiSearch(ctx context.Context, query string, page int) (*MultiSearchResponse, error)
	SearchMovies(ctx context.Context, query string, page int) (*MovieSearchResponse, error)
	SearchTVShows(ctx context.Context, query string, page int) (*TVSearchResponse, error)
	SearchPeople(ctx context.Context, query string, page int) (*PersonSearchResponse, error)

	// 详情相关
	GetMovieDetails(ctx context.Context, id int, language string) (*MovieDetails, error)
	GetTVDetails(ctx context.Context, id int, language string) (*TVDetails, error)
	GetSeasonDetails(ctx context.Context, tvID, seasonNumber int, language string) (*SeasonDetails, error)
	GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int, language string) (*EpisodeDetails, error)
	GetPersonDetails(ctx context.Context, id int, language string) (*PersonDetails, error)

	// 演职员相关
	GetMovieCredits(ctx context.Context, id int, language string) (*Credits, error)
	GetTVCredits(ctx context.Context, id int, language string) (*Credits, error)
	GetPersonMovieCredits(ctx context.Context, id int, language string) (*PersonMovieCredits, error)
	GetPersonTVCredits(ctx context.Context, id int, language string) (*PersonTVCredits, error)
	GetPersonCombinedCredits(ctx context.Context, id int, language string) (*PersonCombinedCredits, error)

	// 图片相关
	GetMovieImages(ctx context.Context, id int, language string) (*MovieImages, error)
	GetTVImages(ctx context.Context, id int, language string) (*TVImages, error)
	GetPersonImages(ctx context.Context, id int, language string) (*PersonImages, error)
	DownloadImage(ctx context.Context, filePath string, config *ImageConfig) ([]byte, error)
	DownloadImages(ctx context.Context, filePaths []string, config *ImageConfig) (map[string][]byte, error)

	// 发现相关
	DiscoverMovies(ctx context.Context, params *DiscoverParams) (*MovieSearchResponse, error)
	DiscoverTVShows(ctx context.Context, params *DiscoverParams) (*TVSearchResponse, error)

	// 趋势相关
	GetTrending(ctx context.Context, mediaType, timeWindow string, page int) (*TrendingResponse, error)

	// 配置相关
	GetConfiguration(ctx context.Context) (*Configuration, error)
}
