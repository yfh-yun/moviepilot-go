package servarr

// RadarrMovie Radarr电影信息
type RadarrMovie struct {
	ID               *int   `json:"id,omitempty"`
	Title            string `json:"title,omitempty"`
	Year             string `json:"year,omitempty"`
	IsAvailable      bool   `json:"isAvailable,omitempty"`
	Monitored        bool   `json:"monitored,omitempty"`
	TmdbID           *int   `json:"tmdbId,omitempty"`
	ImdbID           string `json:"imdbId,omitempty"`
	TitleSlug        string `json:"titleSlug,omitempty"`
	FolderName       string `json:"folderName,omitempty"`
	Path             string `json:"path,omitempty"`
	ProfileID        *int   `json:"profileId,omitempty"`
	QualityProfileID *int   `json:"qualityProfileId,omitempty"`
	Added            string `json:"added,omitempty"`
	HasFile          bool   `json:"hasFile,omitempty"`
}

// SonarrSeries Sonarr电视剧信息
type SonarrSeries struct {
	ID                *int           `json:"id,omitempty"`
	Title             string         `json:"title,omitempty"`
	SortTitle         string         `json:"sortTitle,omitempty"`
	SeasonCount       *int           `json:"seasonCount,omitempty"`
	Status            string         `json:"status,omitempty"`
	Overview          string         `json:"overview,omitempty"`
	Network           string         `json:"network,omitempty"`
	AirTime           string         `json:"airTime,omitempty"`
	Images            []any          `json:"images,omitempty"`
	RemotePoster      string         `json:"remotePoster,omitempty"`
	Seasons           []any          `json:"seasons,omitempty"`
	Year              string         `json:"year,omitempty"`
	Path              string         `json:"path,omitempty"`
	ProfileID         *int           `json:"profileId,omitempty"`
	LanguageProfileID *int           `json:"languageProfileId,omitempty"`
	SeasonFolder      bool           `json:"seasonFolder,omitempty"`
	Monitored         bool           `json:"monitored,omitempty"`
	UseSceneNumbering bool           `json:"useSceneNumbering,omitempty"`
	Runtime           *int           `json:"runtime,omitempty"`
	TmdbID            *int           `json:"tmdbId,omitempty"`
	ImdbID            string         `json:"imdbId,omitempty"`
	TvdbID            *int           `json:"tvdbId,omitempty"`
	TvRageID          *int           `json:"tvRageId,omitempty"`
	TvMazeID          *int           `json:"tvMazeId,omitempty"`
	FirstAired        string         `json:"firstAired,omitempty"`
	SeriesType        string         `json:"seriesType,omitempty"`
	CleanTitle        string         `json:"cleanTitle,omitempty"`
	TitleSlug         string         `json:"titleSlug,omitempty"`
	Certification     string         `json:"certification,omitempty"`
	Genres            []any          `json:"genres,omitempty"`
	Tags              []any          `json:"tags,omitempty"`
	Added             string         `json:"added,omitempty"`
	Ratings           map[string]any `json:"ratings,omitempty"`
	QualityProfileID  *int           `json:"qualityProfileId,omitempty"`
	Statistics        map[string]any `json:"statistics,omitempty"`
	IsAvailable       *bool          `json:"isAvailable,omitempty"`
	HasFile           *bool          `json:"hasFile,omitempty"`
}
