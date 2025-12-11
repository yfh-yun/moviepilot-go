package benchmarks

import (
	"context"
	"strconv"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/repositories/repositories"
)

// setupTestDB 设置测试数据库
func setupTestDB(b *testing.B) *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=moviepilot_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("无法连接到测试数据库: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&database.Subscribe{}); err != nil {
		b.Fatalf("自动迁移失败: %v", err)
	}

	return db
}

// seedSubscribes 填充测试数据
func seedSubscribes(db *gorm.DB, count int) error {
	for i := 0; i < count; i++ {
		subscribe := &database.Subscribe{
			Name:     "测试订阅" + string(rune(i)),
			Type:     "tv",
			State:    "R",
			Username: "test_user",
		}
		if err := db.Create(subscribe).Error; err != nil {
			return err
		}
	}
	return nil
}

// BenchmarkSubscribeRepository_List 测试列表查询性能
func BenchmarkSubscribeRepository_List(b *testing.B) {
	db := setupTestDB(b)
	repo := repositories.NewSubscribeRepository(db)
	ctx := context.Background()

	// 填充1000条测试数据
	if err := seedSubscribes(db, 1000); err != nil {
		b.Fatalf("填充测试数据失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(ctx, interfaces.ListSubscribeParams{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSubscribeRepository_GetByID 测试单条查询性能
func BenchmarkSubscribeRepository_GetByID(b *testing.B) {
	db := setupTestDB(b)
	repo := repositories.NewSubscribeRepository(db)
	ctx := context.Background()

	// 创建测试数据
	subscribe := &database.Subscribe{
		Name:     "测试订阅",
		Type:     "tv",
		State:    "R",
		Username: "test_user",
	}
	if err := db.Create(subscribe).Error; err != nil {
		b.Fatalf("创建测试数据失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idStr := strconv.FormatUint(uint64(subscribe.ID), 10)
		_, err := repo.GetByID(ctx, idStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSubscribeRepository_Create 测试创建性能
func BenchmarkSubscribeRepository_Create(b *testing.B) {
	db := setupTestDB(b)
	repo := repositories.NewSubscribeRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subscribe := &database.Subscribe{
			Name:     "测试订阅" + string(rune(i)),
			Type:     "tv",
			State:    "R",
			Username: "test_user",
		}
		if err := repo.Create(ctx, subscribe); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSubscribeRepository_Update 测试更新性能
func BenchmarkSubscribeRepository_Update(b *testing.B) {
	db := setupTestDB(b)
	repo := repositories.NewSubscribeRepository(db)
	ctx := context.Background()

	// 创建测试数据
	subscribe := &database.Subscribe{
		Name:     "测试订阅",
		Type:     "tv",
		State:    "R",
		Username: "test_user",
	}
	if err := db.Create(subscribe).Error; err != nil {
		b.Fatalf("创建测试数据失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subscribe.State = "P"
		subscribe.LastUpdate = &time.Time{}
		if err := repo.Update(ctx, subscribe); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSubscribeRepository_ListByState 测试按状态查询性能
func BenchmarkSubscribeRepository_ListByState(b *testing.B) {
	db := setupTestDB(b)
	repo := repositories.NewSubscribeRepository(db)
	ctx := context.Background()

	// 填充测试数据
	if err := seedSubscribes(db, 1000); err != nil {
		b.Fatalf("填充测试数据失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ListByState(ctx, "R")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSubscribeRepository_BatchCreate 测试批量创建性能
func BenchmarkSubscribeRepository_BatchCreate(b *testing.B) {
	db := setupTestDB(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subscribes := make([]*database.Subscribe, 100)
		for j := 0; j < 100; j++ {
			subscribes[j] = &database.Subscribe{
				Name:     "批量订阅" + string(rune(i*100+j)),
				Type:     "tv",
				State:    "R",
				Username: "test_user",
			}
		}
		if err := db.CreateInBatches(subscribes, 100).Error; err != nil {
			b.Fatal(err)
		}
	}
}
