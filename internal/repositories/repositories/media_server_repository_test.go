package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

func setupMediaServerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&database.MediaServer{})
	require.NoError(t, err)

	return db
}

func TestMediaServerRepository_Create(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	server := &database.MediaServer{
		Name:     "Emby Server",
		Type:     "emby",
		Host:     "localhost",
		Port:     8096,
		APIKey:   "test_api_key",
		IsActive: true,
	}

	err := repo.Create(ctx, server)
	assert.NoError(t, err)
	assert.NotZero(t, server.ID)
}

func TestMediaServerRepository_GetByID(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "Test Server",
		Type:   "jellyfin",
		Host:   "localhost",
		Port:   8096,
		APIKey: "key123",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 查询
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, server.Name, found.Name)
}

func TestMediaServerRepository_GetByName(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "My Plex",
		Type:   "plex",
		Host:   "localhost",
		Port:   32400,
		APIKey: "plex_key",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 根据名称查询
	found, err := repo.GetByName(ctx, "My Plex")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "plex", found.Type)
}

func TestMediaServerRepository_Update(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "Update Test",
		Type:   "emby",
		Host:   "localhost",
		Port:   8096,
		APIKey: "old_key",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 更新
	server.APIKey = "new_key"
	server.Port = 8097
	err = repo.Update(ctx, server)
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "new_key", found.APIKey)
	assert.Equal(t, 8097, found.Port)
}

func TestMediaServerRepository_Delete(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "Delete Test",
		Type:   "emby",
		Host:   "localhost",
		Port:   8096,
		APIKey: "key",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 删除
	err = repo.Delete(ctx, server.ID)
	assert.NoError(t, err)

	// 验证已删除
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestMediaServerRepository_List(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建多个服务器
	servers := []*database.MediaServer{
		{Name: "Emby1", Type: "emby", Host: "host1", Port: 8096, APIKey: "key1", IsActive: true},
		{Name: "Emby2", Type: "emby", Host: "host2", Port: 8096, APIKey: "key2", IsActive: false},
		{Name: "Plex1", Type: "plex", Host: "host3", Port: 32400, APIKey: "key3", IsActive: true},
	}
	for _, s := range servers {
		err := repo.Create(ctx, s)
		require.NoError(t, err)
	}

	// 查询所有
	params := interfaces.ListMediaServerParams{}
	list, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, list, 3)

	// 按类型查询
	params.Type = "emby"
	list, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, list, 2)

	// 按状态查询
	active := true
	params = interfaces.ListMediaServerParams{IsActive: &active}
	list, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMediaServerRepository_ListActive(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	servers := []*database.MediaServer{
		{Name: "Active1", Type: "emby", Host: "host1", Port: 8096, APIKey: "key1", IsActive: true},
		{Name: "Inactive", Type: "emby", Host: "host2", Port: 8096, APIKey: "key2", IsActive: false},
		{Name: "Active2", Type: "plex", Host: "host3", Port: 32400, APIKey: "key3", IsActive: true},
	}
	for _, s := range servers {
		err := repo.Create(ctx, s)
		require.NoError(t, err)
	}

	// 查询活跃服务器
	list, err := repo.ListActive(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMediaServerRepository_UpdateAPIKey(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "API Test",
		Type:   "emby",
		Host:   "localhost",
		Port:   8096,
		APIKey: "old_api_key",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 更新API Key
	err = repo.UpdateAPIKey(ctx, server.ID, "new_api_key")
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "new_api_key", found.APIKey)
}

func TestMediaServerRepository_SetActive(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:     "Active Test",
		Type:     "emby",
		Host:     "localhost",
		Port:     8096,
		APIKey:   "key",
		IsActive: true,
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 设置为非活跃
	err = repo.SetActive(ctx, server.ID, false)
	assert.NoError(t, err)

	// 验证
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.False(t, found.IsActive)
}

func TestMediaServerRepository_Exists(t *testing.T) {
	db := setupMediaServerTestDB(t)
	repo := NewMediaServerRepository(db)
	ctx := context.Background()

	// 创建服务器
	server := &database.MediaServer{
		Name:   "Exists Test",
		Type:   "emby",
		Host:   "localhost",
		Port:   8096,
		APIKey: "key",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// 检查存在
	exists, err := repo.Exists(ctx, "Exists Test")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在
	exists, err = repo.Exists(ctx, "Nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}
