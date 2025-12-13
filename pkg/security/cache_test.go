package security

import (
	"testing"
	"time"

	"moviepilot-go/pkg/cache"
)

// TestJWTManagerWithCache 测试带缓存的JWT管理器
func TestJWTManagerWithCache(t *testing.T) {
	// 创建内存缓存
	cacheConfig := cache.Config{
		Type:        cache.BackendMemory,
		MaxSize:     100,
		DefaultTTL:  5 * time.Minute,
	}
	cacheBackend, err := cache.NewBackend(cacheConfig)
	if err != nil {
		t.Fatalf("创建缓存失败: %v", err)
	}
	defer cacheBackend.Close()

	// 创建带缓存的JWT管理器
	jwtManager := NewJWTManagerWithCache(
		"test_secret_key",
		1 * time.Hour,
		7 * 24 * time.Hour,
		cacheBackend,
		5 * time.Minute,
	)

	// 生成访问令牌
	userID := uint(1)
	username := "test_user"
	roles := []string{"admin", "user"}
	accessToken, err := jwtManager.GenerateAccessToken(userID, username, roles)
	if err != nil {
		t.Fatalf("生成访问令牌失败: %v", err)
	}

	// 第一次验证（应该命中缓存未命中，执行实际验证）
	claims1, err := jwtManager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("第一次验证令牌失败: %v", err)
	}
	if claims1.UserID != userID || claims1.Username != username {
		t.Fatalf("第一次验证令牌结果不正确: %v", claims1)
	}

	// 第二次验证（应该命中缓存，直接返回缓存结果）
	claims2, err := jwtManager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("第二次验证令牌失败: %v", err)
	}
	if claims2.UserID != userID || claims2.Username != username {
		t.Fatalf("第二次验证令牌结果不正确: %v", claims2)
	}

	// 验证两次验证的结果是否一致
	if claims1.UserID != claims2.UserID || claims1.Username != claims2.Username {
		t.Fatalf("两次验证结果不一致: %v vs %v", claims1, claims2)
	}

	t.Log("JWT管理器缓存测试通过")
}

// TestPasswordManagerWithCache 测试带缓存的密码管理器
func TestPasswordManagerWithCache(t *testing.T) {
	// 创建内存缓存
	cacheConfig := cache.Config{
		Type:        cache.BackendMemory,
		MaxSize:     100,
		DefaultTTL:  5 * time.Minute,
	}
	cacheBackend, err := cache.NewBackend(cacheConfig)
	if err != nil {
		t.Fatalf("创建缓存失败: %v", err)
	}
	defer cacheBackend.Close()

	// 创建带缓存的密码管理器
	passwordManager := NewPasswordManagerWithCache(
		DefaultPasswordConfig,
		cacheBackend,
		5 * time.Minute,
	)

	// 测试密码
	password := "Test12345!"

	// 第一次验证密码强度（应该命中缓存未命中，执行实际验证）
	err = passwordManager.ValidatePasswordStrength(password)
	if err != nil {
		t.Fatalf("第一次验证密码强度失败: %v", err)
	}

	// 第二次验证密码强度（应该命中缓存，直接返回缓存结果）
	err = passwordManager.ValidatePasswordStrength(password)
	if err != nil {
		t.Fatalf("第二次验证密码强度失败: %v", err)
	}

	// 测试弱密码
	weakPassword := "weak"

	// 第一次验证弱密码强度（应该命中缓存未命中，执行实际验证）
	err1 := passwordManager.ValidatePasswordStrength(weakPassword)
	if err1 == nil {
		t.Fatalf("第一次验证弱密码强度应该失败，但实际成功了")
	}

	// 第二次验证弱密码强度（应该命中缓存，直接返回缓存结果）
	err2 := passwordManager.ValidatePasswordStrength(weakPassword)
	if err2 == nil {
		t.Fatalf("第二次验证弱密码强度应该失败，但实际成功了")
	}

	// 验证两次验证的结果是否一致
	if err1.Error() != err2.Error() {
		t.Fatalf("两次验证弱密码结果不一致: %v vs %v", err1, err2)
	}

	t.Log("密码管理器缓存测试通过")
}
