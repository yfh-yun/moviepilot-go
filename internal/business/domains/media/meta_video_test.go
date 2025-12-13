package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaVideo_ParseBasic(t *testing.T) {
	// 创建MetaVideo实例，测试基本解析功能
	video := NewMetaVideo("Test Movie 2023", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Movie", video.EnName)
	assert.Equal(t, "2023", video.Year)
	assert.Equal(t, MediaTypeUnknown, video.Type)
}

func TestMetaVideo_ParseTVSeries(t *testing.T) {
	// 测试电视剧解析
	video := NewMetaVideo("Test Series S01E01", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Series", video.EnName)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)
	assert.Equal(t, 1, *video.BeginEpisode)
	assert.Equal(t, 1, video.TotalSeason)
	assert.Equal(t, 1, video.TotalEpisode)
	assert.Equal(t, "S01", video.Season())
	assert.Equal(t, "E01", video.Episode())
}

func TestMetaVideo_ParseTVSeriesWithYear(t *testing.T) {
	// 测试带年份的电视剧解析
	video := NewMetaVideo("Test Series 2023 S01E01", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Series", video.EnName)
	assert.Equal(t, "2023", video.Year)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)
	assert.Equal(t, 1, *video.BeginEpisode)
}

func TestMetaVideo_ParseMultipleEpisodes(t *testing.T) {
	// 测试多集解析
	video := NewMetaVideo("Test Series S01E01-E03", "", false, MetaParserDeps{})

	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)
	assert.Equal(t, 1, *video.BeginEpisode)
	assert.Equal(t, 3, *video.EndEpisode)
	assert.Equal(t, 3, video.TotalEpisode)
	assert.Equal(t, "E01-E03", video.Episode())
}

func TestMetaVideo_ParseResolution(t *testing.T) {
	// 测试分辨率解析
	video := NewMetaVideo("Test Movie 1080p", "", false, MetaParserDeps{})
	assert.Equal(t, ResourcePix1080P, video.ResourcePix)

	video = NewMetaVideo("Test Movie 720p", "", false, MetaParserDeps{})
	assert.Equal(t, ResourcePix720P, video.ResourcePix)

	video = NewMetaVideo("Test Movie 4K", "", false, MetaParserDeps{})
	assert.Equal(t, ResourcePix4K, video.ResourcePix)
}

func TestMetaVideo_ParseResourceType(t *testing.T) {
	// 测试资源类型解析
	video := NewMetaVideo("Test Movie BluRay", "", false, MetaParserDeps{})
	assert.Equal(t, ResourceType("BluRay"), video.ResourceType)

	video = NewMetaVideo("Test Movie WEB-DL", "", false, MetaParserDeps{})
	assert.Equal(t, ResourceType("WEB-DL"), video.ResourceType)

	video = NewMetaVideo("Test Movie HDTV", "", false, MetaParserDeps{})
	assert.Equal(t, ResourceType("HDTV"), video.ResourceType)
}

func TestMetaVideo_ParseEncode(t *testing.T) {
	// 测试编码解析
	video := NewMetaVideo("Test Movie x264", "", false, MetaParserDeps{})
	assert.Equal(t, "x264", video.VideoEncode)

	video = NewMetaVideo("Test Movie H265", "", false, MetaParserDeps{})
	assert.Equal(t, "H265", video.VideoEncode)

	video = NewMetaVideo("Test Movie DTS", "", false, MetaParserDeps{})
	assert.Equal(t, "DTS", video.AudioEncode)
}

func TestMetaVideo_ParseChineseTitle(t *testing.T) {
	// 测试中文标题解析
	video := NewMetaVideo("测试电影 2023", "", false, MetaParserDeps{})
	assert.Equal(t, "测试电影", video.CnName)
	assert.Equal(t, "2023", video.Year)

	// 测试中文电视剧解析
	video = NewMetaVideo("测试剧集 S01E01", "", false, MetaParserDeps{})
	assert.Equal(t, "测试剧集", video.CnName)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)
	assert.Equal(t, 1, *video.BeginEpisode)
}

func TestMetaVideo_ParseComplexTitle(t *testing.T) {
	// 测试复杂标题解析
	video := NewMetaVideo("Test Series 2023 S01E01 1080p BluRay x264 DTS", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Series", video.EnName)
	assert.Equal(t, "2023", video.Year)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)
	assert.Equal(t, 1, *video.BeginEpisode)
	assert.Equal(t, ResourcePix1080P, video.ResourcePix)
	assert.Equal(t, ResourceType("BluRay"), video.ResourceType)
	assert.Equal(t, "x264", video.VideoEncode)
	assert.Equal(t, "DTS", video.AudioEncode)
}

func TestMetaVideo_ParseOnlySeason(t *testing.T) {
	// 测试只有季的情况
	video := NewMetaVideo("Test Series S02", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Series", video.EnName)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 2, *video.BeginSeason)
	assert.Nil(t, video.BeginEpisode)
}

func TestMetaVideo_ParseOnlyEpisode(t *testing.T) {
	// 测试只有集的情况
	video := NewMetaVideo("Test Series E05", "", false, MetaParserDeps{})

	assert.Equal(t, "Test Series", video.EnName)
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.NotNil(t, video.BeginSeason, "BeginSeason should not be nil for TV series")
	assert.Equal(t, 1, *video.BeginSeason) // 自动设置为第1季
	assert.Equal(t, 5, *video.BeginEpisode)
}

func TestMetaVideo_ParseSpecialCases(t *testing.T) {
	// 测试纯数字文件名
	video := NewMetaVideo("01", "", true, MetaParserDeps{})
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginEpisode)

	// 测试Season xx格式
	video = NewMetaVideo("Season 01", "", false, MetaParserDeps{})
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 1, *video.BeginSeason)

	// 测试Sxx格式
	video = NewMetaVideo("S03", "", false, MetaParserDeps{})
	assert.Equal(t, MediaTypeTV, video.Type)
	assert.Equal(t, 3, *video.BeginSeason)
}
