package tmdb

// MultiSearchResponse 多媒体搜索响应
type MultiSearchResponse struct {
	Page         int           `json:"page"`
	Results      []MultiResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// MultiResult 多媒体搜索结果
type MultiResult struct {
	ID               int     `json:"id"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	MediaType        string  `json:"media_type"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`

	// TV特有字段
	FirstAirDate  string   `json:"first_air_date"`
	Name          string   `json:"name"`
	OriginCountry []string `json:"origin_country"`
	OriginalName  string   `json:"original_name"`
}

// MovieSearchResponse 电影搜索响应
type MovieSearchResponse struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// MovieResult 电影搜索结果
type MovieResult struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

// TVSearchResponse 电视剧搜索响应
type TVSearchResponse struct {
	Page         int        `json:"page"`
	Results      []TVResult `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

// TVResult 电视剧搜索结果
type TVResult struct {
	BackdropPath     string   `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIDs         []int    `json:"genre_ids"`
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

// MovieDetails 电影详情
type MovieDetails struct {
	ID                  int                 `json:"id"`
	Adult               bool                `json:"adult"`
	BackdropPath        string              `json:"backdrop_path"`
	BelongsToCollection *MovieCollection    `json:"belongs_to_collection"`
	Budget              int                 `json:"budget"`
	Genres              []Genre             `json:"genres"`
	Homepage            string              `json:"homepage"`
	IMDBID              string              `json:"imdb_id"`
	OriginalLanguage    string              `json:"original_language"`
	OriginalTitle       string              `json:"original_title"`
	Overview            string              `json:"overview"`
	Popularity          float64             `json:"popularity"`
	PosterPath          string              `json:"poster_path"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	ReleaseDate         string              `json:"release_date"`
	Revenue             int                 `json:"revenue"`
	Runtime             int                 `json:"runtime"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Title               string              `json:"title"`
	Video               bool                `json:"video"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`

	// 额外信息
	Images          MovieImages          `json:"images"`
	Credits         Credits              `json:"credits"`
	Keywords        Keywords             `json:"keywords"`
	ReleaseDates    ReleaseDates         `json:"release_dates"`
	Videos          Videos               `json:"videos"`
	Translations    Translations         `json:"translations"`
	Recommendations MovieRecommendations `json:"recommendations"`
	Similar         MovieSimilar         `json:"similar"`
	Reviews         MovieReviews         `json:"reviews"`
	Lists           MovieLists           `json:"lists"`
	Changes         Changes              `json:"changes"`
}

// TVDetails 电视剧详情
type TVDetails struct {
	ID                  int                 `json:"id"`
	BackdropPath        string              `json:"backdrop_path"`
	CreatedBy           []CreatedBy         `json:"created_by"`
	EpisodeRunTime      []int               `json:"episode_run_time"`
	FirstAirDate        string              `json:"first_air_date"`
	Genres              []Genre             `json:"genres"`
	Homepage            string              `json:"homepage"`
	InProduction        bool                `json:"in_production"`
	Languages           []string            `json:"languages"`
	LastAirDate         string              `json:"last_air_date"`
	LastEpisodeToAir    *TVEpisode          `json:"last_episode_to_air"`
	Name                string              `json:"name"`
	Networks            []Network           `json:"networks"`
	NextEpisodeToAir    *TVEpisode          `json:"next_episode_to_air"`
	NumberOfEpisodes    int                 `json:"number_of_episodes"`
	NumberOfSeasons     int                 `json:"number_of_seasons"`
	OriginCountry       []string            `json:"origin_country"`
	OriginalLanguage    string              `json:"original_language"`
	OriginalName        string              `json:"original_name"`
	Overview            string              `json:"overview"`
	Popularity          float64             `json:"popularity"`
	PosterPath          string              `json:"poster_path"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	Seasons             []TVSeason          `json:"seasons"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Type                string              `json:"type"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`

	// 额外信息
	Images          TVImages          `json:"images"`
	Credits         Credits           `json:"credits"`
	Keywords        Keywords          `json:"keywords"`
	ContentRatings  ContentRatings    `json:"content_ratings"`
	ExternalIDs     ExternalIDs       `json:"external_ids"`
	Videos          Videos            `json:"videos"`
	Translations    Translations      `json:"translations"`
	Recommendations TVRecommendations `json:"recommendations"`
	Similar         TVSimilar         `json:"similar"`
	Reviews         TVReviews         `json:"reviews"`
	WatchProviders  WatchProviders    `json:"watch/providers"`
	Changes         Changes           `json:"changes"`
}

// SeasonDetails 季详情
type SeasonDetails struct {
	ID           string      `json:"_id"`
	AirDate      string      `json:"air_date"`
	Episodes     []TVEpisode `json:"episodes"`
	Name         string      `json:"name"`
	Overview     string      `json:"overview"`
	ID2          int         `json:"id"`
	PosterPath   string      `json:"poster_path"`
	SeasonNumber int         `json:"season_number"`
	VoteAverage  float64     `json:"vote_average"`
	VoteCount    int         `json:"vote_count"`

	// 额外信息
	Images      SeasonImages `json:"images"`
	Videos      Videos       `json:"videos"`
	ExternalIDs ExternalIDs  `json:"external_ids"`
	Changes     Changes      `json:"changes"`
	Credits     Credits      `json:"credits"`
}

// EpisodeDetails 集详情
type EpisodeDetails struct {
	AirDate        string      `json:"air_date"`
	Crew           []Crew      `json:"crew"`
	EpisodeNumber  int         `json:"episode_number"`
	GuestStars     []GuestStar `json:"guest_stars"`
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Overview       string      `json:"overview"`
	ProductionCode string      `json:"production_code"`
	Runtime        int         `json:"runtime"`
	SeasonNumber   int         `json:"season_number"`
	StillPath      string      `json:"still_path"`
	VoteAverage    float64     `json:"vote_average"`
	VoteCount      int         `json:"vote_count"`

	// 额外信息
	Images      EpisodeImages `json:"images"`
	Videos      Videos        `json:"videos"`
	ExternalIDs ExternalIDs   `json:"external_ids"`
	Changes     Changes       `json:"changes"`
	Credits     Credits       `json:"credits"`
}

// Credits 演职员信息
type Credits struct {
	ID   int    `json:"id"`
	Cast []Cast `json:"cast"`
	Crew []Crew `json:"crew"`
}

// Cast 演员信息
type Cast struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	ID                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	CastID             int     `json:"cast_id"`
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Order              int     `json:"order"`
}

// Crew 制作人员信息
type Crew struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	ID                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	CreditID           string  `json:"credit_id"`
	Department         string  `json:"department"`
	Job                string  `json:"job"`
}

// TrendingResponse 趋势响应
type TrendingResponse struct {
	Page         int            `json:"page"`
	Results      []TrendingItem `json:"results"`
	TotalPages   int            `json:"total_pages"`
	TotalResults int            `json:"total_results"`
}

// TrendingItem 趋势项
type TrendingItem struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	ID               int     `json:"id"`
	Title            string  `json:"title,omitempty"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title,omitempty"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	MediaType        string  `json:"media_type"`
	GenreIDs         []int   `json:"genre_ids,omitempty"`
	Popularity       float64 `json:"popularity"`
	ReleaseDate      string  `json:"release_date,omitempty"`
	Video            bool    `json:"video,omitempty"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count,omitempty"`

	// TV特有字段
	Name          string   `json:"name,omitempty"`
	OriginalName  string   `json:"original_name,omitempty"`
	FirstAirDate  string   `json:"first_air_date,omitempty"`
	OriginCountry []string `json:"origin_country,omitempty"`
}

// 辅助结构体
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductionCompany struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type ProductionCountry struct {
	ISO31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

type SpokenLanguage struct {
	EnglishName string `json:"english_name"`
	ISO6391     string `json:"iso_639_1"`
	Name        string `json:"name"`
}

type MovieCollection struct {
	ID           int    `json:"id"`
	BackdropPath string `json:"backdrop_path"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
}

type CreatedBy struct {
	ID          int    `json:"id"`
	CreditID    string `json:"credit_id"`
	Name        string `json:"name"`
	Gender      int    `json:"gender"`
	ProfilePath string `json:"profile_path"`
}

type Network struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type TVSeason struct {
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

type TVEpisode struct {
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ProductionCode string  `json:"production_code"`
	SeasonNumber   int     `json:"season_number"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
}

type GuestStar struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	CreditID           string  `json:"credit_id"`
	Order              int     `json:"order"`
	Character          string  `json:"character"`
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	KnownForDepartment string  `json:"known_for_department"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}

// 图片相关结构体
type MovieImages struct {
	Backdrops []Image `json:"backdrops"`
	Logos     []Image `json:"logos"`
	Posters   []Image `json:"posters"`
}

type TVImages struct {
	Backdrops []Image `json:"backdrops"`
	Logos     []Image `json:"logos"`
	Posters   []Image `json:"posters"`
}

type SeasonImages struct {
	Posters []Image `json:"posters"`
}

type EpisodeImages struct {
	Stills []Image `json:"stills"`
}

type Image struct {
	AspectRatio float64 `json:"aspect_ratio"`
	FilePath    string  `json:"file_path"`
	Height      int     `json:"height"`
	ISO6391     string  `json:"iso_639_1"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
	Width       int     `json:"width"`
}

// 其他响应结构体（简化版）
type Keywords struct {
	Keywords []Keyword `json:"keywords"`
}

type Keyword struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ReleaseDates struct {
	Results []ReleaseDateResult `json:"results"`
}

type ReleaseDateResult struct {
	ISO31661     string        `json:"iso_3166_1"`
	ReleaseDates []ReleaseDate `json:"release_dates"`
}

type ReleaseDate struct {
	Certification string `json:"certification"`
	ISO6391       string `json:"iso_639_1"`
	Note          string `json:"note"`
	ReleaseDate   string `json:"release_date"`
	Type          int    `json:"type"`
}

type Videos struct {
	ID      int     `json:"id"`
	Results []Video `json:"results"`
}

type Video struct {
	ID        string `json:"id"`
	ISO6391   string `json:"iso_639_1"`
	ISO31661  string `json:"iso_3166_1"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Official  bool   `json:"official"`
	Published string `json:"published_at"`
	Site      string `json:"site"`
	Size      int    `json:"size"`
	Type      string `json:"type"`
}

type Translations struct {
	Translations []Translation `json:"translations"`
}

type Translation struct {
	ISO6391     string          `json:"iso_639_1"`
	ISO31661    string          `json:"iso_3166_1"`
	Name        string          `json:"name"`
	EnglishName string          `json:"english_name"`
	Data        TranslationData `json:"data"`
}

type TranslationData struct {
	Title    string `json:"title,omitempty"`
	Overview string `json:"overview,omitempty"`
	Tagline  string `json:"tagline,omitempty"`
	Homepage string `json:"homepage,omitempty"`
}

type MovieRecommendations struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

type MovieSimilar struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

type MovieReviews struct {
	Page         int      `json:"page"`
	Results      []Review `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type MovieLists struct {
	Page         int         `json:"page"`
	Results      []MovieList `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type MovieList struct {
	Description   string `json:"description"`
	FavoriteCount int    `json:"favorite_count"`
	ID            int    `json:"id"`
	ItemCount     int    `json:"item_count"`
	ISO6391       string `json:"iso_639_1"`
	ListType      string `json:"list_type"`
	Name          string `json:"name"`
	PosterPath    string `json:"poster_path"`
}

type TVRecommendations struct {
	Page         int        `json:"page"`
	Results      []TVResult `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

type TVSimilar struct {
	Page         int        `json:"page"`
	Results      []TVResult `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

type TVReviews struct {
	Page         int      `json:"page"`
	Results      []Review `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type Review struct {
	ID         string `json:"id"`
	Author     string `json:"author"`
	Content    string `json:"content"`
	ISO6391    string `json:"iso_639_1"`
	MediaID    int    `json:"media_id"`
	MediaTitle string `json:"media_title"`
	MediaType  string `json:"media_type"`
	URL        string `json:"url"`
}

type ContentRatings struct {
	Results []ContentRating `json:"results"`
}

type ContentRating struct {
	ISO31661 string `json:"iso_3166_1"`
	Rating   string `json:"rating"`
}

type ExternalIDs struct {
	IMDBID      string `json:"imdb_id"`
	FreebaseMID string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TVDBID      int    `json:"tvdb_id"`
	TVRageID    int    `json:"tvrage_id"`
	FacebookID  string `json:"facebook_id"`
	InstagramID string `json:"instagram_id"`
	TwitterID   string `json:"twitter_id"`
	WikipediaID string `json:"wikipedia_id"`
}

type WatchProviders struct {
	ID      int                       `json:"id"`
	Results map[string]ProviderResult `json:"results"`
}

type ProviderResult struct {
	Link          string     `json:"link"`
	Flatrate      []Provider `json:"flatrate"`
	FlatrateAgain []Provider `json:"flatrate_again"`
	Buy           []Provider `json:"buy"`
	Rent          []Provider `json:"rent"`
	Free          []Provider `json:"free"`
	Ads           []Provider `json:"ads"`
}

type Provider struct {
	DisplayPriority int    `json:"display_priority"`
	LogoPath        string `json:"logo_path"`
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
}

type Changes struct {
	Changes []ChangeItem `json:"changes"`
}

type ChangeItem struct {
	Key   string   `json:"key"`
	Items []Change `json:"items"`
}

type Change struct {
	ID      string `json:"id"`
	Action  string `json:"action"`
	Time    string `json:"time"`
	ISO6391 string `json:"iso_639_1"`
	Value   any    `json:"value"`
}

// 图片相关类型定义

// Images 图片集合响应
type Images struct {
	ID        int         `json:"id"`
	Backdrops []ImageData `json:"backdrops"`
	Logos     []ImageData `json:"logos"`
	Posters   []ImageData `json:"posters"`
	Profiles  []ImageData `json:"profiles"`
	Stills    []ImageData `json:"stills"`
}

// ImageData 单个图片数据
type ImageData struct {
	AspectRatio float64 `json:"aspect_ratio"`
	FilePath    string  `json:"file_path"`
	Height      int     `json:"height"`
	ISO6391     string  `json:"iso_639_1"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
	Width       int     `json:"width"`
}

// PersonImages 人物图片响应
type PersonImages struct {
	ID       int         `json:"id"`
	Profiles []ImageData `json:"profiles"`
}

// Configuration TMDB配置信息
type Configuration struct {
	Images     ImageConfiguration `json:"images"`
	ChangeKeys []string           `json:"change_keys"`
}

// ImageConfiguration 图片配置
type ImageConfiguration struct {
	BaseURL       string   `json:"base_url"`
	SecureBaseURL string   `json:"secure_base_url"`
	BackdropSizes []string `json:"backdrop_sizes"`
	LogoSizes     []string `json:"logo_sizes"`
	PosterSizes   []string `json:"poster_sizes"`
	ProfileSizes  []string `json:"profile_sizes"`
	StillSizes    []string `json:"still_sizes"`
}

// ImageConfig 图片下载配置
type ImageConfig struct {
	Size           string `json:"size"`
	BaseURL        string `json:"base_url"`
	Quality        int    `json:"quality"`
	UserAgent      string `json:"user_agent"`
	MaxConcurrency int    `json:"max_concurrency"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	ImageConfig    ImageConfig `json:"image_config"`
	OutputDir      string      `json:"output_dir"`
	MaxConcurrency int         `json:"max_concurrency"`
	Timeout        int         `json:"timeout"`
	UserAgent      string      `json:"user_agent"`
}

// DiscoverParams 发现API参数
type DiscoverParams struct {
	Language               string   `json:"language"`
	Region                 string   `json:"region"`
	SortBy                 string   `json:"sort_by"`
	CertificationCountry   string   `json:"certification_country"`
	Certification          string   `json:"certification"`
	CertificationLTE       string   `json:"certification.lte"`
	CertificationGTE       string   `json:"certification.gte"`
	IncludeAdult           bool     `json:"include_adult"`
	IncludeVideo           bool     `json:"include_video"`
	PrimaryReleaseYear     int      `json:"primary_release_year"`
	PrimaryReleaseDateGTE  string   `json:"primary_release_date.gte"`
	PrimaryReleaseDateLTE  string   `json:"primary_release_date.lte"`
	ReleaseDateGTE         string   `json:"release_date.gte"`
	ReleaseDateLTE         string   `json:"release_date.lte"`
	WithReleaseType        string   `json:"with_release_type"`
	Year                   int      `json:"year"`
	WithGenres             string   `json:"with_genres"`
	WithCast               string   `json:"with_cast"`
	WithCrew               string   `json:"with_crew"`
	WithPeople             string   `json:"with_people"`
	WithCompanies          string   `json:"with_companies"`
	WithNetworks           string   `json:"with_networks"`
	WatchRegion            string   `json:"watch_region"`
	WithWatchProviders     string   `json:"with_watch_providers"`
	WatchMonetizationTypes []string `json:"watch_monetization_types"`
	WithoutWatchProviders  string   `json:"without_watch_providers"`
	WithoutGenres          string   `json:"without_genres"`
	WithoutKeywords        string   `json:"without_keywords"`
	WithKeywords           string   `json:"with_keywords"`
	WithRuntimeGTE         int      `json:"with_runtime.gte"`
	WithRuntimeLTE         int      `json:"with_runtime.lte"`
	Page                   int      `json:"page"`
}

// 人物相关类型定义

// PersonDetails 人物详情
type PersonDetails struct {
	Birthday           string   `json:"birthday"`
	KnownForDepartment string   `json:"known_for_department"`
	Deathday           string   `json:"deathday"`
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	AlsoKnownAs        []string `json:"also_known_as"`
	Gender             int      `json:"gender"`
	Biography          string   `json:"biography"`
	Popularity         float64  `json:"popularity"`
	PlaceOfBirth       string   `json:"place_of_birth"`
	ProfilePath        string   `json:"profile_path"`
	Adult              bool     `json:"adult"`
	IMDBID             string   `json:"imdb_id"`
	Homepage           string   `json:"homepage"`
}

// PersonSearchResponse 人物搜索响应
type PersonSearchResponse struct {
	Page         int            `json:"page"`
	Results      []PersonResult `json:"results"`
	TotalPages   int            `json:"total_pages"`
	TotalResults int            `json:"total_results"`
}

// PersonResult 人物搜索结果
type PersonResult struct {
	Adult              bool    `json:"adult"`
	ID                 int     `json:"id"`
	KnownFor           []Movie `json:"known_for"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}

// PersonMovieCredits 人物电影作品
type PersonMovieCredits struct {
	ID   int          `json:"id"`
	Cast []PersonCast `json:"cast"`
	Crew []PersonCrew `json:"crew"`
}

// PersonTVCredits 人物电视剧作品
type PersonTVCredits struct {
	ID   int          `json:"id"`
	Cast []PersonCast `json:"cast"`
	Crew []PersonCrew `json:"crew"`
}

// PersonCombinedCredits 人物综合作品
type PersonCombinedCredits struct {
	ID   int                `json:"id"`
	Cast []PersonCastCredit `json:"cast"`
	Crew []PersonCrewCredit `json:"crew"`
}

// PersonCast 人物演员作品
type PersonCast struct {
	ID               int     `json:"id"`
	Character        string  `json:"character"`
	CreditID         string  `json:"credit_id"`
	Order            int     `json:"order"`
	Adult            bool    `json:"adult"`
	GenreIDs         []int   `json:"genre_ids"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

// PersonCrew 人物工作人员作品
type PersonCrew struct {
	ID               int     `json:"id"`
	Department       string  `json:"department"`
	CreditID         string  `json:"credit_id"`
	Job              string  `json:"job"`
	Adult            bool    `json:"adult"`
	GenreIDs         []int   `json:"genre_ids"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

// PersonCastCredit 人物综合演员作品（包含电影和电视剧）
type PersonCastCredit struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Character        string  `json:"character"`
	CreditID         string  `json:"credit_id"`
	Order            int     `json:"order"`
	MediaType        string  `json:"media_type"`
	// TV特有字段
	FirstAirDate  string   `json:"first_air_date"`
	Name          string   `json:"name"`
	OriginCountry []string `json:"origin_country"`
	OriginalName  string   `json:"original_name"`
}

// PersonCrewCredit 人物综合工作人员作品（包含电影和电视剧）
type PersonCrewCredit struct {
	ID               int     `json:"id"`
	Department       string  `json:"department"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	CreditID         string  `json:"credit_id"`
	Job              string  `json:"job"`
	MediaType        string  `json:"media_type"`
	// TV特有字段
	FirstAirDate  string   `json:"first_air_date"`
	Name          string   `json:"name"`
	OriginCountry []string `json:"origin_country"`
	OriginalName  string   `json:"original_name"`
}

// PersonExternalIDs 人物ID
type PersonExternalIDs struct {
	ID          int    `json:"id"`
	FreebaseMid string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	IMDBID      string `json:"imdb_id"`
	TVRageID    int    `json:"tvrage_id"`
	TVDBID      int    `json:"tvdb_id"`
}

// Movie 电影信息（用于人物作品）
type Movie struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}
