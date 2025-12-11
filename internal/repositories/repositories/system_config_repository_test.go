package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
)

func setupSystemConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&database.SystemConfig{})
	require.NoError(t, err)

	return db
}

func TestSystemConfigRepository_Set(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置配置
	err := repo.Set(ctx, "test_key", "test_value")
	assert.NoError(t, err)

	// 验证
	config, err := repo.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test_key", config.Key)
	assert.Equal(t, "test_value", config.Value)
}

func TestSystemConfigRepository_Get(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置配置
	err := repo.Set(ctx, "key1", "value1")
	require.NoError(t, err)

	// 获取配置
	config, err := repo.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "value1", config.Value)

	// 获取不存在的配置
	config, err = repo.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestSystemConfigRepository_Delete(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置配置
	err := repo.Set(ctx, "key_to_delete", "value")
	require.NoError(t, err)

	// 删除配置
	err = repo.Delete(ctx, "key_to_delete")
	assert.NoError(t, err)

	// 验证已删除
	config, err := repo.Get(ctx, "key_to_delete")
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestSystemConfigRepository_List(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置多个配置
	configs := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	for k, v := range configs {
		err := repo.Set(ctx, k, v)
		require.NoError(t, err)
	}

	// 获取所有配置
	list, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestSystemConfigRepository_BatchSet(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 批量设置
	configs := map[string]string{
		"batch_key1": "batch_value1",
		"batch_key2": "batch_value2",
		"batch_key3": "batch_value3",
	}
	err := repo.BatchSet(ctx, configs)
	assert.NoError(t, err)

	// 验证
	for k, v := range configs {
		config, err := repo.Get(ctx, k)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, v, config.Value)
	}
}

func TestSystemConfigRepository_BatchGet(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置配置
	configs := map[string]string{
		"get_key1": "get_value1",
		"get_key2": "get_value2",
	}
	err := repo.BatchSet(ctx, configs)
	require.NoError(t, err)

	// 批量获取
	keys := []string{"get_key1", "get_key2"}
	result, err := repo.BatchGet(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "get_value1", result["get_key1"])
	assert.Equal(t, "get_value2", result["get_key2"])
}

func TestSystemConfigRepository_Exists(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	repo := NewSystemConfigRepository(db)
	ctx := context.Background()

	// 设置配置
	err := repo.Set(ctx, "exists_key", "exists_value")
	require.NoError(t, err)

	// 检查存在
	exists, err := repo.Exists(ctx, "exists_key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在
	exists, err = repo.Exists(ctx, "nonexistent_key")
	assert.NoError(t, err)
	assert.False(t, exists)
}
