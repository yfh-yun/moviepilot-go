package business

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"moviepilot-go/internal/business/media"
)

func TestParseFileName(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected *media.FileMetadata
	}{
		{
			name:     "标准电影格式",
			filename: "The.Matrix.1999.1080p.BluRay.x264.mkv",
			expected: &media.FileMetadata{
				Title:      "The Matrix",
				Year:       1999,
				Resolution: "1080p",
				Source:     "BluRay",
				Codec:      "x264",
				Extension:  ".mkv",
				Type:       "movie",
				IsAnime:    false,
			},
		},
		{
			name:     "电视剧格式",
			filename: "Game.of.Thrones.S08E06.FINAL.1080p.WEB-DL.x265.mkv",
			expected: &media.FileMetadata{
				Title:      "Game of Thrones FINAL",
				Season:     8,
				Episode:    6,
				Resolution: "1080p",
				Source:     "WEB-DL",
				Codec:      "x265",
				Extension:  ".mkv",
				Type:       "tv",
				IsAnime:    false,
			},
		},
		{
			name:     "动漫格式",
			filename: "[SubsPlease] Demon Slayer - 01 [1080p].mkv",
			expected: &media.FileMetadata{
				Title:      "Demon Slayer",
				Episode:    1,
				Resolution: "1080p",
				Group:      "SubsPlease",
				Extension:  ".mkv",
				Type:       "anime",
				IsAnime:    true,
			},
		},
		{
			name:     "中文电影",
			filename: "肖申克的救赎.The.Shawshank.Redemption.1994.1080p.BluRay.AAC.mp4",
			expected: &media.FileMetadata{
				Title:      "肖申克的救赎 The Shawshank Redemption",
				Year:       1994,
				Resolution: "1080p",
				Source:     "BluRay",
				Audio:      "AAC",
				Extension:  ".mp4",
				Type:       "movie",
				IsAnime:    false,
			},
		},
		{
			name:     "4K电影",
			filename: "Inception.2010.2160p.UHD.BluRay.x265.mkv",
			expected: &media.FileMetadata{
				Title:      "Inception",
				Year:       2010,
				Resolution: "2160p",
				Source:     "BluRay",
				Codec:      "x265",
				Extension:  ".mkv",
				Type:       "movie",
				IsAnime:    false,
			},
		},
		{
			name:     "季集格式2",
			filename: "Breaking.Bad.S01E01.1080p.WEB-DL.DD5.1.H.264.mkv",
			expected: &media.FileMetadata{
				Title:      "Breaking Bad",
				Season:     1,
				Episode:    1,
				Resolution: "1080p",
				Source:     "WEB-DL",
				Codec:      "x264",
				Extension:  ".mkv",
				Type:       "tv",
				IsAnime:    false,
			},
		},
		{
			name:     "简单格式",
			filename: "Movie.2023.1080p.mkv",
			expected: &media.FileMetadata{
				Title:      "Movie",
				Year:       2023,
				Resolution: "1080p",
				Extension:  ".mkv",
				Type:       "movie",
				IsAnime:    false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := media.ParseFileName(tc.filename)
			
			// 验证基本字段
			assert.Equal(t, tc.expected.Title, result.Title, "标题不匹配")
			assert.Equal(t, tc.expected.Year, result.Year, "年份不匹配")
			assert.Equal(t, tc.expected.Season, result.Season, "季号不匹配")
			assert.Equal(t, tc.expected.Episode, result.Episode, "集号不匹配")
			assert.Equal(t, tc.expected.Resolution, result.Resolution, "分辨率不匹配")
			assert.Equal(t, tc.expected.Source, result.Source, "来源不匹配")
			assert.Equal(t, tc.expected.Codec, result.Codec, "编码不匹配")
			assert.Equal(t, tc.expected.Audio, result.Audio, "音频不匹配")
			assert.Equal(t, tc.expected.Extension, result.Extension, "扩展名不匹配")
			assert.Equal(t, tc.expected.Type, result.Type, "媒体类型不匹配")
			assert.Equal(t, tc.expected.IsAnime, result.IsAnime, "动漫标识不匹配")
			
			// 对于发布组，只在非空时验证
			if tc.expected.Group != "" {
				assert.Equal(t, tc.expected.Group, result.Group, "发布组不匹配")
			}
		})
	}
}

func TestSanitizeForSearch(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "The.Matrix.1999.1080p.BluRay.x264",
			expected: "The Matrix 1999",
		},
		{
			input:    "[SubsPlease] Demon Slayer - 01 [1080p]",
			expected: "Demon Slayer 01",
		},
		{
			input:    "Game.of.Thrones.S08E06.FINAL.1080p.WEB-DL",
			expected: "Game of Thrones S08E06 FINAL",
		},
		{
			input:    "肖申克的救赎.The.Shawshank.Redemption.1994",
			expected: "肖申克的救赎 The Shawshank Redemption 1994",
		},
		{
			input:    "Movie.Title.2023.1080p.BluRay.x264-AVC",
			expected: "Movie Title 2023 AVC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := media.SanitizeForSearch(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractYearFromTitle(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{
			input:    "The Matrix 1999",
			expected: 1999,
		},
		{
			input:    "Movie Title (2023)",
			expected: 2023,
		},
		{
			input:    "Something from 1985",
			expected: 1985,
		},
		{
			input:    "No year here",
			expected: 0,
		},
		{
			input:    "Invalid 1800 year",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := media.ExtractYearFromTitle(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"movie.mp4", true},
		{"movie.mkv", true},
		{"movie.avi", true},
		{"movie.mov", true},
		{"movie.wmv", true},
		{"movie.flv", true},
		{"movie.webm", true},
		{"movie.m4v", true},
		{"subtitle.srt", false},
		{"subtitle.ass", false},
		{"document.txt", false},
		{"image.jpg", false},
		{"", false},
		{"movie", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := media.IsVideoFile(tc.filename)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSubtitleFile(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"subtitle.srt", true},
		{"subtitle.ass", true},
		{"subtitle.ssa", true},
		{"subtitle.sub", true},
		{"subtitle.vtt", true},
		{"movie.mp4", false},
		{"document.txt", false},
		{"image.jpg", false},
		{"", false},
		{"subtitle", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := media.IsSubtitleFile(tc.filename)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDetermineMediaType(t *testing.T) {
	testCases := []struct {
		name     string
		metadata *media.FileMetadata
		expected string
	}{
		{
			name: "电影",
			metadata: &media.FileMetadata{
				Season:  0,
				Episode: 0,
				IsAnime: false,
			},
			expected: "movie",
		},
		{
			name: "电视剧",
			metadata: &media.FileMetadata{
				Season:  1,
				Episode: 1,
				IsAnime: false,
			},
			expected: "tv",
		},
		{
			name: "动漫",
			metadata: &media.FileMetadata{
				Season:  0,
				Episode: 1,
				IsAnime: true,
			},
			expected: "anime",
		},
		{
			name: "动漫电视剧",
			metadata: &media.FileMetadata{
				Season:  1,
				Episode: 1,
				IsAnime: true,
			},
			expected: "anime",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := media.DetermineMediaType(tc.metadata)
			assert.Equal(t, tc.expected, result)
		})
	}
}