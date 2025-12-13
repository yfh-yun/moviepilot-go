package media

import (
	"moviepilot-go/pkg/cache"
)

// MediaType 媒体类型枚举
type MediaType string

const (
	MediaTypeMovie   MediaType = "movie"
	MediaTypeTV      MediaType = "tv"
	MediaTypeAnime   MediaType = "anime"
	MediaTypeUnknown MediaType = "unknown"
)

// ResourceType 资源类型枚举
type ResourceType string

const (
	ResourceTypeMovie ResourceType = "movie"
	ResourceTypeTV    ResourceType = "tv"
	ResourceTypeAnime ResourceType = "anime"
	ResourceTypeMusic ResourceType = "music"
	ResourceTypeOther ResourceType = "other"
)

// ResourceEffect 资源效果枚举
type ResourceEffect string

const (
	ResourceEffectBluray  ResourceEffect = "bluray"
	ResourceEffectHDTV    ResourceEffect = "hdtv"
	ResourceEffectWEB     ResourceEffect = "web"
	ResourceEffectDVD     ResourceEffect = "dvd"
	ResourceEffectSD      ResourceEffect = "sd"
	ResourceEffectUnknown ResourceEffect = "unknown"
)

// ResourcePix 资源分辨率枚举
type ResourcePix string

const (
	ResourcePix4K      ResourcePix = "4k"
	ResourcePix1080P   ResourcePix = "1080p"
	ResourcePix720P    ResourcePix = "720p"
	ResourcePix480P    ResourcePix = "480p"
	ResourcePixUnknown ResourcePix = "unknown"
)

// MetaParserDeps 元信息解析器依赖
type MetaParserDeps struct {
	WordsMatcher       *WordsMatcher
	ReleaseMatcher     *ReleaseGroupsMatcher
	CustomizationMatch *CustomizationMatcher
	StreamingPlatforms *StreamingPlatforms
	Cache              cache.Backend
}
