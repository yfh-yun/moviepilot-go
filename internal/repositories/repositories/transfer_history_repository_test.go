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

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&database.TransferHistory{})
	require.NoError(t, err)

	return db
}

func TestTransferHistoryRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	year := "2023"
	tmdbID := 12345
	history := &database.TransferHistory{
		Type:       "movie",
		Title:      "Test Movie",
		Year:       &year,
		TMDBID:     &tmdbID,
		Src:        "/source/test.mkv",
		Dest:       "/target/test.mkv",
		Source:     "/source/test.mkv",
		SourcePath: "/source/test.mkv",
		Target:     "/target/test.mkv",
		TargetPath: "/target/test.mkv",
		Mode:       "copy",
		Status:     true,
		Date:       time.Now().Format("2006-01-02 15:04:05"),
	}

	err := repo.Create(ctx, history)
	assert.NoError(t, err)
	assert.NotZero(t, history.ID)
}

func TestTransferHistoryRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	history := &database.TransferHistory{
		Type:       "movie",
		Title:      "Test Movie",
		Year:       &year,
		Src:        "/source/test.mkv",
		Dest:       "/target/test.mkv",
		Source:     "/source/test.mkv",
		SourcePath: "/source/test.mkv",
		Target:     "/target/test.mkv",
		TargetPath: "/target/test.mkv",
		Status:     true,
		Date:       time.Now().Format("2006-01-02 15:04:05"),
	}
	err := repo.Create(ctx, history)
	require.NoError(t, err)

	// 查询
	found, err := repo.GetByID(ctx, history.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, history.Title, found.Title)
	assert.Equal(t, history.Type, found.Type)
}

func TestTransferHistoryRepository_GetByTitle(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	// 创建多条测试数据
	year := "2023"
	for i := 0; i < 3; i++ {
		history := &database.TransferHistory{
			Type:       "movie",
			Title:      "Test Movie",
			Year:       &year,
			Src:        "/source/test.mkv",
			Dest:       "/target/test.mkv",
			Source:     "/source/test.mkv",
			SourcePath: "/source/test.mkv",
			Target:     "/target/test.mkv",
			TargetPath: "/target/test.mkv",
			Status:     true,
			Date:       time.Now().Format("2006-01-02 15:04:05"),
		}
		err := repo.Create(ctx, history)
		require.NoError(t, err)
	}

	// 查询
	results, err := repo.GetByTitle(ctx, "Test")
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestTransferHistoryRepository_ListByHash(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	year := "2023"
	hash := "test_hash_123"

	// 创建测试数据
	for i := 0; i < 2; i++ {
		history := &database.TransferHistory{
			Type:         "movie",
			Title:        "Test Movie",
			Year:         &year,
			Src:          "/source/test.mkv",
			Dest:         "/target/test.mkv",
			Source:       "/source/test.mkv",
			SourcePath:   "/source/test.mkv",
			Target:       "/target/test.mkv",
			TargetPath:   "/target/test.mkv",
			DownloadHash: hash,
			Status:       true,
			Date:         time.Now().Format("2006-01-02 15:04:05"),
		}
		err := repo.Create(ctx, history)
		require.NoError(t, err)
	}

	// 查询
	results, err := repo.ListByHash(ctx, hash)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestTransferHistoryRepository_ListByPage(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	for i := 0; i < 15; i++ {
		history := &database.TransferHistory{
			Type:       "movie",
			Title:      "Test Movie",
			Year:       &year,
			Src:        "/source/test.mkv",
			Dest:       "/target/test.mkv",
			Source:     "/source/test.mkv",
			SourcePath: "/source/test.mkv",
			Target:     "/target/test.mkv",
			TargetPath: "/target/test.mkv",
			Status:     true,
			Date:       time.Now().Format("2006-01-02 15:04:05"),
		}
		err := repo.Create(ctx, history)
		require.NoError(t, err)
	}

	// 分页查询
	params := interfaces.ListTransferHistoryParams{
		Page:     1,
		PageSize: 10,
	}
	results, total, err := repo.ListByPage(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, results, 10)
	assert.Equal(t, int64(15), total)

	// 第二页
	params.Page = 2
	results, total, err = repo.ListByPage(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
	assert.Equal(t, int64(15), total)
}

func TestTransferHistoryRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	history := &database.TransferHistory{
		Type:       "movie",
		Title:      "Test Movie",
		Year:       &year,
		Src:        "/source/test.mkv",
		Dest:       "/target/test.mkv",
		Source:     "/source/test.mkv",
		SourcePath: "/source/test.mkv",
		Target:     "/target/test.mkv",
		TargetPath: "/target/test.mkv",
		Status:     true,
		Date:       time.Now().Format("2006-01-02 15:04:05"),
	}
	err := repo.Create(ctx, history)
	require.NoError(t, err)

	// 更新
	history.Status = false
	history.ErrMsg = "Test error"
	err = repo.Update(ctx, history)
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetByID(ctx, history.ID)
	assert.NoError(t, err)
	assert.False(t, found.Status)
	assert.Equal(t, "Test error", found.ErrMsg)
}

func TestTransferHistoryRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTransferHistoryRepository(db)
	ctx := context.Background()

	// 创建测试数据
	year := "2023"
	history := &database.TransferHistory{
		Type:       "movie",
		Title:      "Test Movie",
		Year:       &year,
		Src:        "/source/test.mkv",
		Dest:       "/target/test.mkv",
		Source:     "/source/test.mkv",
		SourcePath: "/source/test.mkv",
		Target:     "/target/test.mkv",
		TargetPath: "/target/test.mkv",
		Status:     true,
		Date:       time.Now().Format("2006-01-02 15:04:05"),
	}
	err := repo.Create(ctx, history)
	require.NoError(t, err)

	// 删除
	err = repo.Delete(ctx, history.ID)
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetByID(ctx, history.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}
