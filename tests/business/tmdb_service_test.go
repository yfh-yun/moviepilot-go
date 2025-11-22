package business

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/business/media/tmdb"
	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/cache/memory"
)

// MockService 模拟回退服务
type MockService struct {
	mock.Mock
}

func (m *MockService) Identify(files []media.FileItem, opts media.IdentifyOptions) ([]models.Media, error) {
	args := m.Called(files, opts)
	return args.Get(0).([]models.Media), args.Error(1)
}

func TestTMDBService_DownloadPoster(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建测试媒体
	testMedia := models.Media{
		Title:  "Test Movie",
		Type:   "movie",
		Poster: "https://image.tmdb.org/t/p/w500/test_poster.jpg",
		TMDBID: intPtr(12345),
	}

	// 测试空API密钥情况（service为nil）
	service := media.NewTMDBService("", logger, &MockService{}, memCache)
	assert.Nil(t, service)

	// 创建有API密钥的服务（即使无法连接，也会创建service实例）
	service = media.NewTMDBService("fake_api_key", logger, &MockService{}, memCache)
	assert.NotNil(t, service)

	// 测试没有TMDB ID的情况
	_, err := service.DownloadPoster(context.Background(), models.Media{
		Title:  "Test Movie",
		Type:   "movie",
		Poster: "",
		TMDBID: nil,
	}, "/tmp", tmdb.DownloadOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no poster path or TMDB ID available")

	// 测试有TMDB ID但没有API密钥的情况
	_, err = service.DownloadPoster(context.Background(), testMedia, "/tmp", tmdb.DownloadOptions{})
	assert.Error(t, err)
}

func TestTMDBService_DownloadBackdrop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建TMDB服务（使用假的API密钥，所以client会为nil）
	service := media.NewTMDBService("", logger, &MockService{}, memCache)
	assert.Nil(t, service)

	// 创建有API密钥的服务
	service = media.NewTMDBService("fake_api_key", logger, &MockService{}, memCache)
	assert.NotNil(t, service)

	// 测试没有TMDB ID的情况
	_, err := service.DownloadBackdrop(context.Background(), models.Media{
		Title:    "Test TV Show",
		Type:     "tv",
		Backdrop: "",
		TMDBID:   nil,
	}, "/tmp", tmdb.DownloadOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backdrop path or TMDB ID available")
}

func TestTMDBService_DownloadAllImages(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建测试媒体
	mediaInfo := models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		Poster:   "https://image.tmdb.org/t/p/w500/test_poster.jpg",
		Backdrop: "https://image.tmdb.org/t/p/w1280/test_backdrop.jpg",
		TMDBID:   intPtr(12345),
	}

	// 创建临时目录
	tempDir := t.TempDir()

	// 创建TMDB服务（使用假的API密钥，所以client会为nil）
	service := media.NewTMDBService("", logger, &MockService{}, memCache)
	assert.Nil(t, service)

	// 创建有API密钥的服务
	service = media.NewTMDBService("fake_api_key", logger, &MockService{}, memCache)
	assert.NotNil(t, service)

	// 测试下载所有图片
	_, err := service.DownloadAllImages(context.Background(), mediaInfo, tempDir, tmdb.DownloadOptions{
		Size: "w500",
	})

	// 由于没有真实的TMDB客户端，这些调用会失败
	assert.Error(t, err)

	// 如果有真实的TMDB客户端，应该会下载两个文件
	// 在集成测试中验证
}

func TestTMDBService_Identify(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建模拟回退服务
	fallback := &MockService{}
	fallback.On("Identify", mock.Anything, mock.Anything).Return([]models.Media{
		{
			Title: "Fallback Movie",
			Type:  "movie",
		},
	}, nil)

	// 测试空API密钥情况（会返回nil服务）
	service := media.NewTMDBService("", logger, fallback, memCache)
	assert.Nil(t, service)

	// 测试非空API密钥但无法连接的情况（应该回退）
	service = media.NewTMDBService("fake_key", logger, fallback, memCache)

	files := []media.FileItem{
		{Path: "/test/Movie.2023.1080p.mkv"},
	}

	// 执行识别
	results, err := service.Identify(files, media.IdentifyOptions{})

	// 应该使用回退服务
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Fallback Movie", results[0].Title)

	// 验证回退服务被调用
	fallback.AssertExpectations(t)
}

func TestTMDBService_IdentifyWithCache(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建模拟回退服务
	fallback := &MockService{}
	fallback.On("Identify", mock.Anything, mock.Anything).Return([]models.Media{
		{
			Title: "Cached Movie",
			Type:  "movie",
		},
	}, nil)

	// 创建TMDB服务
	service := media.NewTMDBService("fake_key", logger, fallback, memCache)

	files := []media.FileItem{
		{Path: "/test/CachedMovie.2023.1080p.mkv"},
	}

	// 先设置缓存值
	ctx := context.Background()
	testMedia := models.Media{
		Title:  "Cached Movie",
		Type:   "movie",
		TMDBID: intPtr(12345),
	}
	memCache.SetJSON(ctx, "identify:cachedmovie|movie", testMedia, time.Hour)

	// 执行识别（应该命中缓存）
	results, err := service.Identify(files, media.IdentifyOptions{})

	// 应该从缓存获取
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Cached Movie", results[0].Title)
	assert.Equal(t, intPtr(12345), results[0].TMDBID)

	// 回退服务不应该被调用
	fallback.AssertNotCalled(t, "Identify")
}

func TestTMDBService_IdentifyMultipleFiles(t *testing.T) {
	logger := zaptest.NewLogger(t)
	memCache := memory.NewMemoryCache()
	defer memCache.(*memory.MemoryCache).Stop()

	// 创建模拟回退服务
	fallback := &MockService{}
	fallback.On("Identify", mock.Anything, mock.Anything).Return([]models.Media{
		{
			Title: "Fallback Movie 1",
			Type:  "movie",
		},
		{
			Title: "Fallback TV 1",
			Type:  "tv",
		},
	}, nil)

	// 创建TMDB服务
	service := media.NewTMDBService("fake_key", logger, fallback, memCache)

	files := []media.FileItem{
		{Path: "/test/Movie.2023.1080p.mkv"},
		{Path: "/test/TV.Show.S01E01.1080p.mkv"},
	}

	// 执行识别
	results, err := service.Identify(files, media.IdentifyOptions{})

	// 应该使用回退服务
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// 验证结果
	assert.Equal(t, "Fallback Movie 1", results[0].Title)
	assert.Equal(t, "Fallback TV 1", results[1].Title)
	assert.Equal(t, "movie", results[0].Type)
	assert.Equal(t, "tv", results[1].Type)

	// 验证回退服务被调用
	fallback.AssertExpectations(t)
}

// intPtr 辅助函数已在 nfo_test.go 中定义
