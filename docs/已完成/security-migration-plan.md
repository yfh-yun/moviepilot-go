# 安全系统迁移计划

> Python: `app/core/security.py`  
> Go: `internal/infrastructure/security/` + `pkg/utils/crypto/`

---

## 1. Python 安全系统概览

### 1.1 核心功能

- **加密解密**：密码哈希、AES 加密、NexusPHP 加密
- **认证授权**：JWT 生成/验证/刷新、API Key 校验
- **安全配置**：允许域名、图片后缀白名单

### 1.2 典型使用场景

- 用户登录 → JWT 生成
- API 请求 → JWT 验证
- 密码存储 → bcrypt 哈希
- 敏感数据存储 → AES 加密

---

## 2. Go 目标设计

### 2.1 分层设计

| 层级 | 位置 | 职责 |
|------|------|------|
| 基础设施层 | `internal/infrastructure/security/` | 安全策略、认证授权实现 |
| 工具层 | `pkg/utils/crypto/` | 基础加密解密功能 |

### 2.2 核心模块

**JWT 模块**：

```go
// internal/infrastructure/security/jwt.go
func GenerateToken(userID string) (string, error)
func ValidateToken(tokenString string) (*Claims, error)
func RefreshToken(tokenString string) (string, error)
```

**密码模块**：

```go
// internal/infrastructure/security/password.go
func HashPassword(password string) (string, error)
func CheckPassword(password, hash string) bool
```

**API Key 模块**：

```go
// internal/infrastructure/security/api_key.go
func ValidateAPIKey(apiKey string) bool
func GenerateAPIKey() string
```

**加密工具**：

```go
// pkg/utils/crypto/hash.go
func SHA256(data string) string

// pkg/utils/crypto/aes.go
func EncryptAES(data, key string) (string, error)
func DecryptAES(encryptedData, key string) (string, error)

// pkg/utils/crypto/nexusphp.go
func NexusPHPEncrypt(data string) string
func NexusPHPDecrypt(data string) string
```

### 2.3 迁移步骤

1. **实现基础加密功能**：在 `pkg/utils/crypto/` 中实现哈希、AES、NexusPHP 加密
2. **实现认证授权**：在 `internal/infrastructure/security/` 中实现 JWT、密码、API Key 功能
3. **替换 Python 安全调用**：将 Python 中的安全相关调用替换为 Go 实现
4. **集成到业务流程**：在 API 层和业务层中使用新的安全系统

---

## 3. 集成与使用

### 3.1 JWT 使用示例

```go
// 生成 Token
token, err := security.GenerateToken(userID)
if err != nil {
    return err
}

// 验证 Token
claims, err := security.ValidateToken(token)
if err != nil {
    return err
}
```

### 3.2 密码使用示例

```go
// 哈希密码
hashedPassword, err := security.HashPassword(password)
if err != nil {
    return err
}

// 验证密码
if !security.CheckPassword(password, hashedPassword) {
    return errors.New("invalid password")
}
```

---

## 4. 检查清单

- [x] 实现 SHA256 哈希
- [x] 实现 AES 加密解密
- [x] 实现 NexusPHP 加密解密
- [x] 实现 JWT 生成/验证/刷新
- [x] 实现密码哈希与验证
- [x] 实现 API Key 生成与验证
- [x] 集成到 API 中间件
- [x] 集成到业务服务

---

## 5. 后续优化

1. **支持更多加密算法**：如 RSA、ECDSA 等
2. **实现 OAuth2.0 支持**：第三方登录集成
3. **添加安全审计**：记录安全相关操作日志
4. **实现权限细粒度控制**：基于角色的访问控制
5. **支持多因素认证**：提高认证安全性
