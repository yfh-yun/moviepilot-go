package media

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaService_IsAnime(t *testing.T) {
	// 创建测试依赖
	deps := MetaParserDeps{
		WordsMatcher:       NewSimpleWordsMatcher(),
		ReleaseMatcher:     NewReleaseGroupsMatcher(nil),
		CustomizationMatch: NewCustomizationMatcher(nil),
		StreamingPlatforms: NewStreamingPlatforms(),
	}

	// 创建MetaService实例
	ms := NewMetaService(deps)

	// 测试动漫标题
	animeTitles := []string{
		"【2023】【1080P】动漫名称【字幕组】",
		"动漫名称 - 01【1080P】【字幕组】",
		"[2023][1080P]动漫名称[字幕组]",
	}

	for _, title := range animeTitles {
		assert.True(t, ms.IsAnime(title), "title should be anime: %s", title)
	}

	// 测试非动漫标题
	nonAnimeTitles := []string{
		"Movie Title 2023 1080p BluRay",
		"TV Series S01E01 1080p WEB-DL",
		"Documentary Title 2023 720p",
	}

	for _, title := range nonAnimeTitles {
		assert.False(t, ms.IsAnime(title), "title should not be anime: %s", title)
	}
}

func TestMetaService_FindMetaInfo(t *testing.T) {
	// 创建测试依赖
	deps := MetaParserDeps{
		WordsMatcher:       NewSimpleWordsMatcher(),
		ReleaseMatcher:     NewReleaseGroupsMatcher(nil),
		CustomizationMatch: NewCustomizationMatcher(nil),
		StreamingPlatforms: NewStreamingPlatforms(),
	}

	// 创建MetaService实例
	ms := NewMetaService(deps)

	// 测试内嵌元信息标签
	testCases := []struct {
		title           string
		expectedTitle   string
		expectedTMDBID  int64
		expectedType    MediaType
		expectedSeason  *int
		expectedEpisode *int
	}{
		{
			title:          "Movie Title{[tmdbid=12345;type=movie]}",
			expectedTitle:  "Movie Title",
			expectedTMDBID: 12345,
			expectedType:   MediaTypeMovie,
		},
		{
			title:           "TV Series{[tmdbid=67890;type=tv;s=1;e=1-10]}",
			expectedTitle:   "TV Series",
			expectedTMDBID:  67890,
			expectedType:    MediaTypeTV,
			expectedSeason:  intPtr(1),
			expectedEpisode: intPtr(1),
		},
	}

	for _, tc := range testCases {
		cleanedTitle, result := ms.FindMetaInfo(tc.title)
		assert.Equal(t, tc.expectedTitle, cleanedTitle)
		assert.Equal(t, tc.expectedTMDBID, result.TMDBID)
		assert.Equal(t, tc.expectedType, result.Type)
		if tc.expectedSeason != nil {
			assert.NotNil(t, result.BeginSeason)
			assert.Equal(t, *tc.expectedSeason, *result.BeginSeason)
		}
		if tc.expectedEpisode != nil {
			assert.NotNil(t, result.BeginEpisode)
			assert.Equal(t, *tc.expectedEpisode, *result.BeginEpisode)
		}
	}
}

func TestMetaService_MetaInfo(t *testing.T) {
	// 创建测试依赖
	deps := MetaParserDeps{
		WordsMatcher:       NewSimpleWordsMatcher(),
		ReleaseMatcher:     NewReleaseGroupsMatcher(nil),
		CustomizationMatch: NewCustomizationMatcher(nil),
		StreamingPlatforms: NewStreamingPlatforms(),
	}

	// 创建MetaService实例
	ms := NewMetaService(deps)

	// 测试MetaInfo方法
	testCases := []struct {
		title        string
		subtitle     string
		isFile       bool
		expectedType MediaType
	}{
		{
			title:        "Movie Title 2023",
			subtitle:     "",
			isFile:       false,
			expectedType: MediaTypeUnknown,
		},
		{
			title:        "【2023】动漫名称【字幕组】",
			subtitle:     "",
			isFile:       false,
			expectedType: MediaTypeAnime,
		},
	}

	for _, tc := range testCases {
		meta, err := ms.MetaInfo(context.Background(), tc.title, tc.subtitle, tc.isFile, nil)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedType, meta.Type)
	}
}

// intPtr 返回int指针
func intPtr(i int) *int {
	return &i
}
