package enums

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie      MediaType = "电影"
	MediaTypeTV         MediaType = "电视剧"
	MediaTypeCollection MediaType = "系列"
	MediaTypeUnknown    MediaType = "未知"
)

// MediaImageType 媒体图片类型
type MediaImageType string

const (
	MediaImageTypePoster   MediaImageType = "poster_path"
	MediaImageTypeBackdrop MediaImageType = "backdrop_path"
)

// MediaRecognizeType 识别器类型
type MediaRecognizeType string

const (
	// 豆瓣
	MediaRecognizeTypeDouban MediaRecognizeType = "豆瓣"
	// TMDB
	MediaRecognizeTypeTMDB MediaRecognizeType = "TheMovieDb"
	// TVDB
	MediaRecognizeTypeTVDB MediaRecognizeType = "TheTvDb"
	// bangumi
	MediaRecognizeTypeBangumi MediaRecognizeType = "Bangumi"
)
