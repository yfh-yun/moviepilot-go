package scraper

import "context"

// Scraper 刮削器接口
type Scraper interface {
	ScrapeMovie(ctx context.Context, title string, year int) (*MovieMetadata, error)
	ScrapeTV(ctx context.Context, title string, year int) (*TVMetadata, error)
	ScrapeSeason(ctx context.Context, tvID int, seasonNumber int) (*SeasonMetadata, error)
	ScrapeEpisode(ctx context.Context, tvID int, seasonNumber int, episodeNumber int) (*EpisodeMetadata, error)
	DownloadImage(ctx context.Context, imageURL string) ([]byte, error)
}

// MovieMetadata 电影元数据
type MovieMetadata struct {
	TMDBID              int      `json:"tmdb_id"`
	Title               string   `json:"title"`
	OriginalTitle       string   `json:"original_title"`
	Overview            string   `json:"overview"`
	ReleaseDate         string   `json:"release_date"`
	Runtime             int      `json:"runtime"`
	VoteAverage         float64  `json:"vote_average"`
	VoteCount           int      `json:"vote_count"`
	PosterPath          string   `json:"poster_path"`
	BackdropPath        string   `json:"backdrop_path"`
	Genres              []string `json:"genres"`
	ProductionCompanies []string `json:"production_companies"`
}

// TVMetadata 电视剧元数据
type TVMetadata struct {
	TMDBID           int      `json:"tmdb_id"`
	Name             string   `json:"name"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	FirstAirDate     string   `json:"first_air_date"`
	LastAirDate      string   `json:"last_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	PosterPath       string   `json:"poster_path"`
	BackdropPath     string   `json:"backdrop_path"`
	Genres           []string `json:"genres"`
	Networks         []string `json:"networks"`
	NumberOfSeasons  int      `json:"number_of_seasons"`
	NumberOfEpisodes int      `json:"number_of_episodes"`
}

// SeasonMetadata 季元数据
type SeasonMetadata struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	SeasonNumber int    `json:"season_number"`
	AirDate      string `json:"air_date"`
	PosterPath   string `json:"poster_path"`
	EpisodeCount int    `json:"episode_count"`
}

// EpisodeMetadata 集元数据
type EpisodeMetadata struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	SeasonNumber  int     `json:"season_number"`
	EpisodeNumber int     `json:"episode_number"`
	AirDate       string  `json:"air_date"`
	Runtime       int     `json:"runtime"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	StillPath     string  `json:"still_path"`
}

// NewTMDbScraper 创建TMDb刮削器
func NewTMDbScraper(apiKey string) Scraper {
	// TODO: 实现TMDb刮削器
	return nil
}
