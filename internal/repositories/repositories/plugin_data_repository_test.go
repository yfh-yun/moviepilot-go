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

func setupPluginDataTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&database.PluginData{})
	require.NoError(t, err)

	return db
}

func TestPluginDataRepository_Set(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置插件数据
	err := repo.Set(ctx, "plugin1", "key1", "value1")
	assert.NoError(t, err)

	// 验证
	data, err := repo.Get(ctx, "plugin1", "key1")
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "plugin1", data.PluginKey)
	assert.Equal(t, "key1", data.DataKey)
	assert.Equal(t, "value1", data.DataValue)
}

func TestPluginDataRepository_Get(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置数据
	err := repo.Set(ctx, "plugin1", "test_key", "test_value")
	require.NoError(t, err)

	// 获取数据
	data, err := repo.Get(ctx, "plugin1", "test_key")
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "test_value", data.DataValue)

	// 获取不存在的数据
	data, err = repo.Get(ctx, "plugin1", "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestPluginDataRepository_Delete(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置数据
	err := repo.Set(ctx, "plugin1", "delete_key", "delete_value")
	require.NoError(t, err)

	// 删除数据
	err = repo.Delete(ctx, "plugin1", "delete_key")
	assert.NoError(t, err)

	// 验证已删除
	data, err := repo.Get(ctx, "plugin1", "delete_key")
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestPluginDataRepository_ListByPlugin(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置多个数据
	pluginID := "test_plugin"
	for i := 1; i <= 3; i++ {
		err := repo.Set(ctx, pluginID, "key"+string(rune(i+'0')), "value"+string(rune(i+'0')))
		require.NoError(t, err)
	}

	// 获取插件的所有数据
	dataList, err := repo.ListByPlugin(ctx, pluginID)
	assert.NoError(t, err)
	assert.Len(t, dataList, 3)
}

func TestPluginDataRepository_DeleteByPlugin(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置多个数据
	pluginID := "delete_plugin"
	for i := 1; i <= 3; i++ {
		err := repo.Set(ctx, pluginID, "key"+string(rune(i+'0')), "value"+string(rune(i+'0')))
		require.NoError(t, err)
	}

	// 删除插件的所有数据
	err := repo.DeleteByPlugin(ctx, pluginID)
	assert.NoError(t, err)

	// 验证已删除
	dataList, err := repo.ListByPlugin(ctx, pluginID)
	assert.NoError(t, err)
	assert.Len(t, dataList, 0)
}

func TestPluginDataRepository_BatchSet(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 批量设置
	pluginID := "batch_plugin"
	data := map[string]string{
		"batch_key1": "batch_value1",
		"batch_key2": "batch_value2",
		"batch_key3": "batch_value3",
	}
	err := repo.BatchSet(ctx, pluginID, data)
	assert.NoError(t, err)

	// 验证
	for k, v := range data {
		result, err := repo.Get(ctx, pluginID, k)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, v, result.DataValue)
	}
}

func TestPluginDataRepository_BatchGet(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置数据
	pluginID := "get_plugin"
	data := map[string]string{
		"get_key1": "get_value1",
		"get_key2": "get_value2",
	}
	err := repo.BatchSet(ctx, pluginID, data)
	require.NoError(t, err)

	// 批量获取
	keys := []string{"get_key1", "get_key2"}
	result, err := repo.BatchGet(ctx, pluginID, keys)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "get_value1", result["get_key1"])
	assert.Equal(t, "get_value2", result["get_key2"])
}

func TestPluginDataRepository_Exists(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置数据
	err := repo.Set(ctx, "plugin1", "exists_key", "exists_value")
	require.NoError(t, err)

	// 检查存在
	exists, err := repo.Exists(ctx, "plugin1", "exists_key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在
	exists, err = repo.Exists(ctx, "plugin1", "nonexistent_key")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPluginDataRepository_ListAllPlugins(t *testing.T) {
	db := setupPluginDataTestDB(t)
	repo := NewPluginDataRepository(db)
	ctx := context.Background()

	// 设置多个插件的数据
	plugins := []string{"plugin1", "plugin2", "plugin3"}
	for _, pluginID := range plugins {
		err := repo.Set(ctx, pluginID, "key", "value")
		require.NoError(t, err)
	}

	// 获取所有插件ID
	result, err := repo.ListAllPlugins(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "plugin1")
	assert.Contains(t, result, "plugin2")
	assert.Contains(t, result, "plugin3")
}
