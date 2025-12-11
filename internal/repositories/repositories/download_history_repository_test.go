package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

func setupDownloadTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&database.DownloadHistory{}, &database.DownloadFile{})
	require.NoError(t, err)

	return db
}

func TestDownloadHistoryRepository_Create(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	year := "2023"
	tmdbID := 12345
	history := &database.DownloadHistory{
		Path:         "/downloads/test.mkv",
		Type:         "movie",
		Title:        "Test Movie",
		Year:         &year,
		TMDBID:       &tmdbID,
		DownloadHash: "hash123",
		Downloader:   "qbittorrent",
		Date:         time.Now().Format("2006-01-02 15:04:05"),
	}

	err := repo.Create(ctx, history)
	assert.NoError(t, err)
	assert.NotZero(t, history.ID)
}

func TestDownloadHistoryRepository_GetByHash(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	hash := "test_hash_456"
	history := &database.DownloadHistory{
		Path:         "/downloads/test.mkv",
		Type:         "movie",
		Title:        "Test Movie",
		Year:         &year,
		DownloadHash: hash,
		Date:         time.Now().Format("2006-01-02 15:04:05"),
	}
	err := repo.Create(ctx, history)
	require.NoError(t, err)

	// 查询
	found, err := repo.GetByHash(ctx, hash)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, history.Title, found.Title)
	assert.Equal(t, hash, found.DownloadHash)
}

func TestDownloadHistoryRepository_AddFiles(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	hash := "test_hash_789"
	files := []*database.DownloadFile{
		{
			DownloadHash: hash,
			FullPath:     "/downloads/test/file1.mkv",
			SavePath:     "/downloads/test",
			FileName:     "file1.mkv",
			FileSize:     1024000,
			FileExt:      ".mkv",
			State:        0,
		},
		{
			DownloadHash: hash,
			FullPath:     "/downloads/test/file2.mkv",
			SavePath:     "/downloads/test",
			FileName:     "file2.mkv",
			FileSize:     2048000,
			FileExt:      ".mkv",
			State:        0,
		},
	}

	err := repo.AddFiles(ctx, files)
	assert.NoError(t, err)

	// 验证
	foundFiles, err := repo.GetFilesByHash(ctx, hash, nil)
	assert.NoError(t, err)
	assert.Len(t, foundFiles, 2)
}

func TestDownloadHistoryRepository_GetFilesByHash(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	hash := "test_hash_abc"
	files := []*database.DownloadFile{
		{
			DownloadHash: hash,
			FullPath:     "/downloads/test/file1.mkv",
			SavePath:     "/downloads/test",
			FileName:     "file1.mkv",
			FileSize:     1024000,
			State:        0,
		},
		{
			DownloadHash: hash,
			FullPath:     "/downloads/test/file2.mkv",
			SavePath:     "/downloads/test",
			FileName:     "file2.mkv",
			FileSize:     2048000,
			State:        1, // 已删除
		},
	}
	err := repo.AddFiles(ctx, files)
	require.NoError(t, err)

	// 查询所有文件
	allFiles, err := repo.GetFilesByHash(ctx, hash, nil)
	assert.NoError(t, err)
	assert.Len(t, allFiles, 2)

	// 只查询正常文件
	state := 0
	normalFiles, err := repo.GetFilesByHash(ctx, hash, &state)
	assert.NoError(t, err)
	assert.Len(t, normalFiles, 1)
}

func TestDownloadHistoryRepository_UpdateFileState(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	file := &database.DownloadFile{
		DownloadHash: "hash_def",
		FullPath:     "/downloads/test/file.mkv",
		SavePath:     "/downloads/test",
		FileName:     "file.mkv",
		FileSize:     1024000,
		State:        0,
	}
	err := repo.AddFiles(ctx, []*database.DownloadFile{file})
	require.NoError(t, err)

	// 更新状态
	err = repo.UpdateFileState(ctx, file.FullPath, 1)
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetFileByFullPath(ctx, file.FullPath)
	assert.NoError(t, err)
	assert.Equal(t, 1, found.State)
}

func TestDownloadHistoryRepository_ListByPage(t *testing.T) {
	db := setupDownloadTestDB(t)
	repo := NewDownloadHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	for i := 0; i < 15; i++ {
		history := &database.DownloadHistory{
			Path:         "/downloads/test.mkv",
			Type:         "movie",
			Title:        "Test Movie",
			Year:         &year,
			DownloadHash: "hash",
			Date:         time.Now().Format("2006-01-02 15:04:05"),
		}
		err := repo.Create(ctx, history)
		require.NoError(t, err)
	}

	// 分页查询
	params := interfaces.ListDownloadHistoryParams{
		Page:     1,
		PageSize: 10,
	}
	results, total, err := repo.ListByPage(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, results, 10)
	assert.Equal(t, int64(15), total)
}
