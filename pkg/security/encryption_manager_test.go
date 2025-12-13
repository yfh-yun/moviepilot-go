package security

import (
	"crypto/rand"
	"testing"
	"time"

	"moviepilot-go/pkg/cache"
)

// 生成测试用的 Fernet 密钥
func generateTestFernetKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}

// 生成测试用的 AES 密钥
func generateTestAESKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return string(key)
}

// 测试 EncryptMessage 方法的缓存功能
func TestEncryptionManager_EncryptMessage_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := "test message"
	key := generateTestFernetKey()

	// 第一次加密（应该写入缓存）
	result1, err := em.EncryptMessage(message, key)
	if err != nil {
		t.Fatalf("第一次加密失败: %v", err)
	}

	// 第二次加密（应该从缓存读取）
	result2, err := em.EncryptMessage(message, key)
	if err != nil {
		t.Fatalf("第二次加密失败: %v", err)
	}

	// 验证结果一致
	if result1 != result2 {
		t.Fatalf("两次加密结果不一致: %s != %s", result1, result2)
	}
}

// 测试 Decrypt 方法的缓存功能
func TestEncryptionManager_Decrypt_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := "test message"
	key := generateTestFernetKey()

	// 先加密得到密文
	encrypted, err := em.EncryptMessage(message, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 第一次解密（应该写入缓存）
	result1, err := em.Decrypt([]byte(encrypted), key)
	if err != nil {
		t.Fatalf("第一次解密失败: %v", err)
	}

	// 第二次解密（应该从缓存读取）
	result2, err := em.Decrypt([]byte(encrypted), key)
	if err != nil {
		t.Fatalf("第二次解密失败: %v", err)
	}

	// 验证结果一致
	if string(result1) != string(result2) {
		t.Fatalf("两次解密结果不一致: %s != %s", string(result1), string(result2))
	}
}

// 测试 HashSHA256 方法的缓存功能
func TestEncryptionManager_HashSHA256_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := "test message"

	// 第一次哈希计算（应该写入缓存）
	result1 := em.HashSHA256(message)

	// 第二次哈希计算（应该从缓存读取）
	result2 := em.HashSHA256(message)

	// 验证结果一致
	if result1 != result2 {
		t.Fatalf("两次哈希结果不一致: %s != %s", result1, result2)
	}
}

// 测试 AESDecrypt 方法的缓存功能
func TestEncryptionManager_AESDecrypt_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	original := "test message"
	key := generateTestAESKey()

	// 先加密得到密文
	encrypted, err := em.AESEncrypt(original, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 第一次解密（应该写入缓存）
	result1, err := em.AESDecrypt(encrypted, key)
	if err != nil {
		t.Fatalf("第一次解密失败: %v", err)
	}

	// 第二次解密（应该从缓存读取）
	result2, err := em.AESDecrypt(encrypted, key)
	if err != nil {
		t.Fatalf("第二次解密失败: %v", err)
	}

	// 验证结果一致
	if result1 != result2 {
		t.Fatalf("两次解密结果不一致: %s != %s", result1, result2)
	}

	// 验证结果正确
	if result1 != original {
		t.Fatalf("解密结果不正确: %s != %s", result1, original)
	}
}

// 测试 AESEncrypt 方法的缓存功能
func TestEncryptionManager_AESEncrypt_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := "test message"
	key := generateTestAESKey()

	// 第一次加密（应该写入缓存）
	result1, err := em.AESEncrypt(message, key)
	if err != nil {
		t.Fatalf("第一次加密失败: %v", err)
	}

	// 第二次加密（应该从缓存读取）
	result2, err := em.AESEncrypt(message, key)
	if err != nil {
		t.Fatalf("第二次加密失败: %v", err)
	}

	// 验证结果一致
	if result1 != result2 {
		t.Fatalf("两次加密结果不一致: %s != %s", result1, result2)
	}
}

// 测试 NexusPHPEncrypt 方法的缓存功能
func TestEncryptionManager_NexusPHPEncrypt_Cache(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := "test message"
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	// 第一次加密（应该写入缓存）
	result1, err := em.NexusPHPEncrypt(message, key)
	if err != nil {
		t.Fatalf("第一次加密失败: %v", err)
	}

	// 第二次加密（应该从缓存读取）
	result2, err := em.NexusPHPEncrypt(message, key)
	if err != nil {
		t.Fatalf("第二次加密失败: %v", err)
	}

	// 验证结果一致
	if result1 != result2 {
		t.Fatalf("两次加密结果不一致: %s != %s", result1, result2)
	}
}

// 测试不带缓存的加密管理器
func TestEncryptionManager_WithoutCache(t *testing.T) {
	// 创建不带缓存的加密管理器
	em := NewEncryptionManager()

	// 测试数据
	message := "test message"
	key := generateTestFernetKey()

	// 加密（应该不使用缓存）
	result, err := em.EncryptMessage(message, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 验证结果不为空
	if result == "" {
		t.Fatalf("加密结果为空")
	}
}

// 测试空字符串的处理
func TestEncryptionManager_EmptyString(t *testing.T) {
	// 创建内存缓存
	cacheBackend := cache.NewMemoryBackend(cache.Config{
		Type:       cache.BackendMemory,
		MaxSize:    1024,
		DefaultTTL: 5 * time.Minute,
	})
	cacheTTL := 5 * time.Minute

	// 创建带缓存的加密管理器
	em := NewEncryptionManagerWithCache(cacheBackend, cacheTTL)

	// 测试数据
	message := ""
	key := generateTestAESKey()

	// 加密空字符串
	encrypted, err := em.AESEncrypt(message, key)
	if err != nil {
		t.Fatalf("加密空字符串失败: %v", err)
	}

	// 解密空字符串
	decrypted, err := em.AESDecrypt(encrypted, key)
	if err != nil {
		t.Fatalf("解密空字符串失败: %v", err)
	}

	// 验证结果为空字符串
	if decrypted != "" {
		t.Fatalf("解密空字符串结果不为空: %s", decrypted)
	}
}
