package unit

import (
	"context"
	"testing"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

func TestMemoryCache(t *testing.T) {
	c := cache.NewMemoryCache()
	ctx := context.Background()

	// Test Set and Get
	err := c.Set(ctx, "test_key", "test_value", time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	value, err := c.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test Exists
	exists, err := c.Exists(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to check if key exists: %v", err)
	}

	if !exists {
		t.Error("Key should exist")
	}

	// Test Delete
	err = c.Delete(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to delete cache: %v", err)
	}

	exists, err = c.Exists(ctx, "test_key")
	if err != nil {
		t.Fatalf("Failed to check if key exists after delete: %v", err)
	}

	if exists {
		t.Error("Key should not exist after delete")
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	c := cache.NewMemoryCache()
	ctx := context.Background()

	// Set with very short expiration
	err := c.Set(ctx, "expire_key", "expire_value", time.Millisecond*10)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Wait for expiration
	time.Sleep(time.Millisecond * 20)

	// Check if key still exists
	exists, err := c.Exists(ctx, "expire_key")
	if err != nil {
		t.Fatalf("Failed to check if key exists: %v", err)
	}

	if exists {
		t.Error("Key should have expired")
	}
}