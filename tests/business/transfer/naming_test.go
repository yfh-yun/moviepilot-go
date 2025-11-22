package business

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/internal/business/transfer"
	"moviepilot-go/internal/models"
)

func TestNewMovieNamingStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 测试默认模板
	strategy, err := transfer.NewMovieNamingStrategy("", logger)
	assert.NoError(t, err)
	assert.NotNil(t, strategy)

	// 测试自定义模板
	customTemplate := "${title}/${title}.${year}${ext}"
	strategy, err = transfer.NewMovieNamingStrategy(customTemplate, logger)
	assert.NoError(t, err)
	assert.NotNil(t, strategy)

	// 测试无效模板
	invalidTemplate := ""
	strategy, err = transfer.NewMovieNamingStrategy(invalidTemplate, logger)
	assert.NoError(t, err) // 应该使用默认模板
	assert.NotNil(t, strategy)
}

func TestMovieNamingStrategy_GeneratePath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy, err := transfer.NewMovieNamingStrategy("", logger)
	assert.NoError(t, err)

	year := "2023"
	media := &models.Media{
		Title: "Test Movie",
		Year:  &year,
		Type:  "movie",
	}

	metadata := transfer.FileMetadata{
		Resolution: "1080p",
		Source:     "BluRay",
		Codec:      "x264",
	}

	path, err := strategy.GeneratePath(media, "/source/test.mkv", metadata)
	assert.NoError(t, err)
	assert.Contains(t, path, "Test Movie")
	assert.Contains(t, path, "2023")
	assert.Contains(t, path, ".mkv")
}

func TestTVNamingStrategy_GeneratePath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy, err := transfer.NewTVNamingStrategy("", logger)
	assert.NoError(t, err)

	year := "2023"
	season := 1
	episode := 5
	media := &models.Media{
		Title:   "Test TV Show",
		Year:    &year,
		Type:    "tv",
		Season:  &season,
		Episode: &episode,
	}

	metadata := transfer.FileMetadata{
		Resolution: "1080p",
		Source:     "WEB-DL",
	}

	path, err := strategy.GeneratePath(media, "/source/test.mkv", metadata)
	assert.NoError(t, err)
	assert.Contains(t, path, "Test TV Show")
	assert.Contains(t, path, "Season")
	assert.Contains(t, path, "S01")
	assert.Contains(t, path, "E05")
}

func TestAnimeNamingStrategy_GeneratePath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy, err := transfer.NewAnimeNamingStrategy("", logger)
	assert.NoError(t, err)

	season := 1
	episode := 10
	media := &models.Media{
		Title:   "Test Anime",
		Type:    "anime",
		Season:  &season,
		Episode: &episode,
	}

	metadata := transfer.FileMetadata{
		Group: "SubsPlease",
	}

	path, err := strategy.GeneratePath(media, "/source/test.mkv", metadata)
	assert.NoError(t, err)
	assert.Contains(t, path, "Test Anime")
	assert.Contains(t, path, "SubsPlease")
}

func TestNamingManager(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := transfer.NamingConfig{
		MovieTemplate: "${title} (${year})/${title}${ext}",
		TVTemplate:    "${title}/S${season_num}E${episode_num}${ext}",
		AnimeTemplate: "[${group}] ${title} - ${episode_num}${ext}",
	}

	manager, err := transfer.NewNamingManager(config, logger)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// 测试电影
	year := "2023"
	movieMedia := &models.Media{
		Title: "Test Movie",
		Year:  &year,
		Type:  "movie",
	}

	path, err := manager.GeneratePath(movieMedia, "/source/test.mkv", transfer.FileMetadata{})
	assert.NoError(t, err)
	assert.Contains(t, path, "Test Movie")
	assert.Contains(t, path, "2023")

	// 测试电视剧
	season := 1
	episode := 5
	tvMedia := &models.Media{
		Title:   "Test TV",
		Type:    "tv",
		Season:  &season,
		Episode: &episode,
	}

	path, err = manager.GeneratePath(tvMedia, "/source/test.mkv", transfer.FileMetadata{})
	assert.NoError(t, err)
	assert.Contains(t, path, "Test TV")
	assert.Contains(t, path, "S1E5")

	// 测试未知类型(应该回退到电影策略)
	unknownMedia := &models.Media{
		Title: "Unknown",
		Year:  &year,
		Type:  "unknown",
	}

	path, err = manager.GeneratePath(unknownMedia, "/source/test.mkv", transfer.FileMetadata{})
	assert.NoError(t, err)
	assert.Contains(t, path, "Unknown")
}

func TestParseFileMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		want     transfer.FileMetadata
	}{
		{
			name:     "1080p BluRay x264",
			filePath: "/path/Movie.2023.1080p.BluRay.x264.mkv",
			want: transfer.FileMetadata{
				Resolution: "1080p",
				Source:     "BluRay",
				Codec:      "x264",
			},
		},
		{
			name:     "2160p WEB-DL x265",
			filePath: "/path/Movie.2023.2160p.WEB-DL.x265.mkv",
			want: transfer.FileMetadata{
				Resolution: "2160p",
				Source:     "WEB-DL",
				Codec:      "x265",
			},
		},
		{
			name:     "720p WEBRip",
			filePath: "/path/Movie.2023.720p.WEBRip.mkv",
			want: transfer.FileMetadata{
				Resolution: "720p",
				Source:     "WEBRip",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := transfer.ParseFileMetadata(tc.filePath)
			assert.Equal(t, tc.want.Resolution, metadata.Resolution)
			assert.Equal(t, tc.want.Source, metadata.Source)
			if tc.want.Codec != "" {
				assert.Equal(t, tc.want.Codec, metadata.Codec)
			}
		})
	}
}

// 基准测试
func BenchmarkMovieNamingStrategy_GeneratePath(b *testing.B) {
	logger := zaptest.NewLogger(b)
	strategy, _ := transfer.NewMovieNamingStrategy("", logger)

	year := "2023"
	media := &models.Media{
		Title: "Test Movie",
		Year:  &year,
		Type:  "movie",
	}

	metadata := transfer.FileMetadata{
		Resolution: "1080p",
		Source:     "BluRay",
		Codec:      "x264",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = strategy.GeneratePath(media, "/source/test.mkv", metadata)
	}
}

func BenchmarkParseFileMetadata(b *testing.B) {
	filePath := "/path/Movie.2023.1080p.BluRay.x264.mkv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transfer.ParseFileMetadata(filePath)
	}
}
