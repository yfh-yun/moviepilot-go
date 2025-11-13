package models

// TmdbSeason TMDB季信�?type TmdbSeason struct {
	AirDate      string  `json:"air_date,omitempty"`
	EpisodeCount int     `json:"episode_count,omitempty"`
	Name         string  `json:"name,omitempty"`
	Overview     string  `json:"overview,omitempty"`
	PosterPath   string  `json:"poster_path,omitempty"`
	SeasonNumber int     `json:"season_number,omitempty"`
	VoteAverage  float64 `json:"vote_average,omitempty"`
}

// TmdbEpisode TMDB集信�?type TmdbEpisode struct {
	AirDate       string        `json:"air_date,omitempty"`
	EpisodeNumber int           `json:"episode_number,omitempty"`
	EpisodeType   string        `json:"episode_type,omitempty"`
	Name          string        `json:"name,omitempty"`
	Overview      string        `json:"overview,omitempty"`
	Runtime       int           `json:"runtime,omitempty"`
	SeasonNumber  int           `json:"season_number,omitempty"`
	StillPath     string        `json:"still_path,omitempty"`
	VoteAverage   float64       `json:"vote_average,omitempty"`
	Crew          []interface{} `json:"crew,omitempty"`
	GuestStars    []interface{} `json:"guest_stars,omitempty"`
}

// NewTmdbSeason 创建一个新�?TmdbSeason 实例
func NewTmdbSeason() *TmdbSeason {
	return &TmdbSeason{}
}

// NewTmdbEpisode 创建一个新�?TmdbEpisode 实例
func NewTmdbEpisode() *TmdbEpisode {
	return &TmdbEpisode{
		Crew:       make([]interface{}, 0),
		GuestStars: make([]interface{}, 0),
	}
}
