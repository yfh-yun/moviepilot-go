package business

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/models"
)

func TestWriteMovieNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	mediaInfo := &models.Media{
		Title:       "Test Movie",
		Description: "This is a test movie",
		Year:        stringPtr("2023"),
		TMDBID:      intPtr(12345),
		IMDBID:      stringPtr("tt1234567"),
		Vote:        float64Ptr(8.5),
		Runtime:     intPtr(120),
		Genres:      `["Action", "Drama"]`,
		Poster:      "https://example.com/poster.jpg",
	}

	nfoPath := filepath.Join(tempDir, "movie.nfo")

	err := media.WriteMovieNFO(mediaInfo, nfoPath, logger)
	require.NoError(t, err)

	// 验证文件存在
	assert.FileExists(t, nfoPath)

	// 读取并验证内容
	data, err := os.ReadFile(nfoPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "<movie>")
	assert.Contains(t, content, "<title>Test Movie</title>")
	assert.Contains(t, content, "<plot>This is a test movie</plot>")
	assert.Contains(t, content, "<year>2023</year>")
	assert.Contains(t, content, "<tmdbid>12345</tmdbid>")
	assert.Contains(t, content, "<imdbid>tt1234567</imdbid>")
	assert.Contains(t, content, "<rating>8.5</rating>")
	assert.Contains(t, content, "<runtime>120</runtime>")
	assert.Contains(t, content, "<genre>Action</genre>")
	assert.Contains(t, content, "<genre>Drama</genre>")
	assert.Contains(t, content, "<thumb>https://example.com/poster.jpg</thumb>")
	assert.Contains(t, content, "</movie>")
}

func TestWriteTVShowNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	mediaInfo := &models.Media{
		Title:       "Test TV Show",
		Description: "This is a test TV show",
		Year:        stringPtr("2022"),
		TMDBID:      intPtr(54321),
		IMDBID:      stringPtr("tt7654321"),
		Vote:        float64Ptr(9.0),
		Genres:      `["Comedy", "Drama"]`,
		Poster:      "https://example.com/tvposter.jpg",
	}

	nfoPath := filepath.Join(tempDir, "tvshow.nfo")

	err := media.WriteTVShowNFO(mediaInfo, nfoPath, logger)
	require.NoError(t, err)

	// 验证文件存在
	assert.FileExists(t, nfoPath)

	// 读取并验证内容
	data, err := os.ReadFile(nfoPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "<tvshow>")
	assert.Contains(t, content, "<title>Test TV Show</title>")
	assert.Contains(t, content, "<plot>This is a test TV show</plot>")
	assert.Contains(t, content, "<year>2022</year>")
	assert.Contains(t, content, "<tmdbid>54321</tmdbid>")
	assert.Contains(t, content, "<imdbid>tt7654321</imdbid>")
	assert.Contains(t, content, "<rating>9</rating>")
	assert.Contains(t, content, "<genre>Comedy</genre>")
	assert.Contains(t, content, "<genre>Drama</genre>")
	assert.Contains(t, content, "<thumb>https://example.com/tvposter.jpg</thumb>")
	assert.Contains(t, content, "</tvshow>")
}

func TestWriteEpisodeNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	mediaInfo := &models.Media{
		Title:       "Test TV Show",
		Description: "This is a test episode",
		Year:        stringPtr("2022"),
		TMDBID:      intPtr(54321),
		IMDBID:      stringPtr("tt7654321"),
		Vote:        float64Ptr(8.8),
		Runtime:     intPtr(45),
		Poster:      "https://example.com/eposter.jpg",
	}

	nfoPath := filepath.Join(tempDir, "episode.nfo")

	err := media.WriteEpisodeNFO(mediaInfo, 1, 1, "Pilot Episode", nfoPath, logger)
	require.NoError(t, err)

	// 验证文件存在
	assert.FileExists(t, nfoPath)

	// 读取并验证内容
	data, err := os.ReadFile(nfoPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "<episodedetails>")
	assert.Contains(t, content, "<title>Pilot Episode</title>")
	assert.Contains(t, content, "<showtitle>Test TV Show</showtitle>")
	assert.Contains(t, content, "<season>1</season>")
	assert.Contains(t, content, "<episode>1</episode>")
	assert.Contains(t, content, "<plot>This is a test episode</plot>")
	assert.Contains(t, content, "<tmdbid>54321</tmdbid>")
	assert.Contains(t, content, "<imdbid>tt7654321</imdbid>")
	assert.Contains(t, content, "<rating>8.8</rating>")
	assert.Contains(t, content, "<runtime>45</runtime>")
	assert.Contains(t, content, "<thumb>https://example.com/eposter.jpg</thumb>")
	assert.Contains(t, content, "</episodedetails>")
}

func TestReadMovieNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	// 创建测试NFO文件
	nfoContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>Test Movie</title>
    <plot>This is a test movie</plot>
    <year>2023</year>
    <rating>8.5</rating>
    <runtime>120</runtime>
    <tmdbid>12345</tmdbid>
    <imdbid>tt1234567</imdbid>
    <genre>Action</genre>
    <genre>Drama</genre>
    <thumb>https://example.com/poster.jpg</thumb>
</movie>`

	nfoPath := filepath.Join(tempDir, "movie.nfo")
	err := os.WriteFile(nfoPath, []byte(nfoContent), 0644)
	require.NoError(t, err)

	// 读取NFO文件
	nfoData, err := media.ReadNFO(nfoPath, logger)
	require.NoError(t, err)

	// 验证内容
	assert.Equal(t, "Test Movie", nfoData.Title)
	assert.Equal(t, "This is a test movie", nfoData.Plot)
	assert.Equal(t, 2023, nfoData.Year)
	assert.Equal(t, 8.5, nfoData.Rating)
	assert.Equal(t, 120, nfoData.Runtime)
	assert.Equal(t, 12345, nfoData.TMDBID)
	assert.Equal(t, "tt1234567", nfoData.IMDBID)
	assert.Contains(t, nfoData.Genres, "Action")
	assert.Contains(t, nfoData.Genres, "Drama")
	assert.Equal(t, "https://example.com/poster.jpg", nfoData.Poster)
}

func TestReadTVShowNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	// 创建测试NFO文件
	nfoContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<tvshow>
    <title>Test TV Show</title>
    <plot>This is a test TV show</plot>
    <year>2022</year>
    <rating>9.0</rating>
    <tmdbid>54321</tmdbid>
    <imdbid>tt7654321</imdbid>
    <genre>Comedy</genre>
    <genre>Drama</genre>
    <thumb>https://example.com/tvposter.jpg</thumb>
</tvshow>`

	nfoPath := filepath.Join(tempDir, "tvshow.nfo")
	err := os.WriteFile(nfoPath, []byte(nfoContent), 0644)
	require.NoError(t, err)

	// 读取NFO文件
	nfoData, err := media.ReadTVShowNFO(nfoPath, logger)
	require.NoError(t, err)

	// 验证内容
	assert.Equal(t, "Test TV Show", nfoData.Title)
	assert.Equal(t, "This is a test TV show", nfoData.Plot)
	assert.Equal(t, 2022, nfoData.Year)
	assert.Equal(t, 9.0, nfoData.Rating)
	assert.Equal(t, 54321, nfoData.TMDBID)
	assert.Equal(t, "tt7654321", nfoData.IMDBID)
	assert.Contains(t, nfoData.Genres, "Comedy")
	assert.Contains(t, nfoData.Genres, "Drama")
	assert.Equal(t, "https://example.com/tvposter.jpg", nfoData.Poster)
}

func TestReadEpisodeNFO(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	// 创建测试NFO文件
	nfoContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<episodedetails>
    <title>Pilot Episode</title>
    <showtitle>Test TV Show</showtitle>
    <season>1</season>
    <episode>1</episode>
    <plot>This is a test episode</plot>
    <rating>8.8</rating>
    <runtime>45</runtime>
    <tmdbid>54321</tmdbid>
    <imdbid>tt7654321</imdbid>
    <thumb>https://example.com/eposter.jpg</thumb>
</episodedetails>`

	nfoPath := filepath.Join(tempDir, "episode.nfo")
	err := os.WriteFile(nfoPath, []byte(nfoContent), 0644)
	require.NoError(t, err)

	// 读取NFO文件
	nfoData, err := media.ReadEpisodeNFO(nfoPath, logger)
	require.NoError(t, err)

	// 验证内容
	assert.Equal(t, "Pilot Episode", nfoData.Title)
	assert.Equal(t, "Test TV Show", nfoData.ShowTitle)
	assert.Equal(t, 1, nfoData.Season)
	assert.Equal(t, 1, nfoData.Episode)
	assert.Equal(t, "This is a test episode", nfoData.Plot)
	assert.Equal(t, 8.8, nfoData.Rating)
	assert.Equal(t, 45, nfoData.Runtime)
	assert.Equal(t, 54321, nfoData.TMDBID)
	assert.Equal(t, "tt7654321", nfoData.IMDBID)
	assert.Equal(t, "https://example.com/eposter.jpg", nfoData.Thumb)
}

func TestDetectNFOType(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{
			path:     "/Movies/Test Movie/tvshow.nfo",
			expected: "tvshow",
		},
		{
			path:     "/TV Shows/Test Show/Season 01/episode.nfo",
			expected: "episode",
		},
		{
			path:     "/TV Shows/Test Show/S01E01.nfo",
			expected: "episode",
		},
		{
			path:     "/Movies/Test Movie/movie.nfo",
			expected: "movie",
		},
		{
			path:     "/Test Movie.nfo",
			expected: "movie",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := media.DetectNFOType(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNFOToMedia(t *testing.T) {
	nfoData := &media.NFOData{
		Title:   "Test Movie",
		Plot:    "Test plot",
		Year:    2023,
		Rating:  8.5,
		Runtime: 120,
		TMDBID:  12345,
		IMDBID:  "tt1234567",
		Poster:  "https://example.com/poster.jpg",
		Genres:  []string{"Action", "Drama"},
	}

	mediaResult := media.NFOToMedia(nfoData, "movie")

	assert.Equal(t, "Test Movie", mediaResult.Title)
	assert.Equal(t, "Test plot", mediaResult.Description)
	assert.Equal(t, "movie", mediaResult.Type)
	assert.Equal(t, "2023", *mediaResult.Year)
	assert.Equal(t, 8.5, *mediaResult.Vote)
	assert.Equal(t, 120, *mediaResult.Runtime)
	assert.Equal(t, 12345, *mediaResult.TMDBID)
	assert.Equal(t, "tt1234567", *mediaResult.IMDBID)
	assert.Equal(t, "https://example.com/poster.jpg", mediaResult.Poster)
	assert.Contains(t, mediaResult.Genres, "Action")
	assert.Contains(t, mediaResult.Genres, "Drama")
}

func TestNFOFileCreation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	// 测试目录自动创建
	deepPath := filepath.Join(tempDir, "Movies", "Test Movie", "2023")
	nfoPath := filepath.Join(deepPath, "movie.nfo")

	mediaInfo := &models.Media{
		Title: "Test Movie",
		Year:  stringPtr("2023"),
	}

	err := media.WriteMovieNFO(mediaInfo, nfoPath, logger)
	require.NoError(t, err)

	// 验证目录和文件都被创建
	assert.FileExists(t, nfoPath)
	assert.DirExists(t, deepPath)
}
