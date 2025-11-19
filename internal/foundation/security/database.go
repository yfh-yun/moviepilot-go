// Package security 安全包
package security

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/database"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// DatabaseSecurity 数据库安全管理器
type DatabaseSecurity struct {
	dbManager     *database.Manager
	encryptor     *NexusPHPEncryptor
	sessionSecret []byte
	config        *DatabaseSecurityConfig
}

// DatabaseSecurityConfig 数据库安全配置
type DatabaseSecurityConfig struct {
	SessionTimeout         int      `json:"session_timeout"`          // 会话超时时间（分钟）
	MaxLoginAttempts       int      `json:"max_login_attempts"`       // 最大登录尝试次数
	PasswordMinLength      int      `json:"password_min_length"`      // 密码最小长度
	PasswordRequireUpper   bool     `json:"password_require_upper"`   // 密码需要大写字母
	PasswordRequireLower   bool     `json:"password_require_lower"`   // 密码需要小写字母
	PasswordRequireNumber  bool     `json:"password_require_number"`  // 密码需要数字
	PasswordRequireSpecial bool     `json:"password_require_special"` // 密码需要特殊字符
	EnableSQLInjection     bool     `json:"enable_sql_injection"`     // 启用SQL注入防护
	EnableXSSFilter        bool     `json:"enable_xss_filter"`        // 启用XSS过滤
	SensitiveFields        []string `json:"sensitive_fields"`         // 敏感字段列表
	EncryptionFields       []string `json:"encryption_fields"`        // 加密字段列表
}

// NewDatabaseSecurity 创建数据库安全管理器
func NewDatabaseSecurity(dbManager *database.Manager, encryptor *NexusPHPEncryptor, config *DatabaseSecurityConfig) *DatabaseSecurity {
	if config == nil {
		config = getDefaultDatabaseSecurityConfig()
	}

	ds := &DatabaseSecurity{
		dbManager: dbManager,
		encryptor: encryptor,
		config:    config,
	}

	// 生成会话密钥
	ds.generateSessionSecret()

	return ds
}

// getDefaultDatabaseSecurityConfig 获取默认数据库安全配置
func getDefaultDatabaseSecurityConfig() *DatabaseSecurityConfig {
	return &DatabaseSecurityConfig{
		SessionTimeout:         30, // 30分钟
		MaxLoginAttempts:       5,  // 5次尝试
		PasswordMinLength:      8,  // 8位最小长度
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: false,
		EnableSQLInjection:     true,
		EnableXSSFilter:        true,
		SensitiveFields: []string{
			"password", "passkey", "email", "phone", "address",
			"secret_key", "private_key", "token", "cookie",
		},
		EncryptionFields: []string{
			"password", "passkey", "secret_key", "private_key",
		},
	}
}

// generateSessionSecret 生成会话密钥
func (ds *DatabaseSecurity) generateSessionSecret() {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		logger.Error("生成会话密钥失败", zap.Error(err))
		// 使用固定密钥作为后备
		secret = []byte("default-session-secret-key-32")
	}
	ds.sessionSecret = secret
}

// ValidatePassword 验证密码强度
func (ds *DatabaseSecurity) ValidatePassword(password string) *PasswordValidation {
	validation := &PasswordValidation{
		IsValid: true,
		Errors:  []string{},
		Score:   0,
	}

	// 检查长度
	if len(password) < ds.config.PasswordMinLength {
		validation.IsValid = false
		validation.Errors = append(validation.Errors,
			fmt.Sprintf("密码长度必须至少%d位", ds.config.PasswordMinLength))
	} else {
		validation.Score += 20
	}

	// 检查大写字母
	if ds.config.PasswordRequireUpper {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "密码必须包含大写字母")
		} else {
			validation.Score += 15
		}
	}

	// 检查小写字母
	if ds.config.PasswordRequireLower {
		hasLower := false
		for _, char := range password {
			if char >= 'a' && char <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "密码必须包含小写字母")
		} else {
			validation.Score += 15
		}
	}

	// 检查数字
	if ds.config.PasswordRequireNumber {
		hasNumber := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "密码必须包含数字")
		} else {
			validation.Score += 15
		}
	}

	// 检查特殊字符
	if ds.config.PasswordRequireSpecial {
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		hasSpecial := false
		for _, char := range password {
			if strings.ContainsRune(specialChars, char) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "密码必须包含特殊字符")
		} else {
			validation.Score += 15
		}
	}

	// 计算复杂度分数
	if len(password) >= 12 {
		validation.Score += 10
	}
	if len(password) >= 16 {
		validation.Score += 10
	}

	return validation
}

// PasswordValidation 密码验证结果
type PasswordValidation struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
	Score   int      `json:"score"`
	Level   string   `json:"level"`
}

// SanitizeInput 清理输入数据
func (ds *DatabaseSecurity) SanitizeInput(input string) string {
	if input == "" {
		return ""
	}

	// 移除前后空白
	input = strings.TrimSpace(input)

	// XSS过滤
	if ds.config.EnableXSSFilter {
		input = ds.filterXSS(input)
	}

	// SQL注入防护
	if ds.config.EnableSQLInjection {
		input = ds.filterSQLInjection(input)
	}

	return input
}

// filterXSS 过滤XSS攻击
func (ds *DatabaseSecurity) filterXSS(input string) string {
	// 移除危险的HTML标签和JavaScript代码
	xssPatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)<script.*?>.*?</script>`), ""},
		{regexp.MustCompile(`(?i)<iframe.*?>.*?</iframe>`), ""},
		{regexp.MustCompile(`(?i)<object.*?>.*?</object>`), ""},
		{regexp.MustCompile(`(?i)<embed.*?>.*?</embed>`), ""},
		{regexp.MustCompile(`(?i)javascript:`), ""},
		{regexp.MustCompile(`(?i)vbscript:`), ""},
		{regexp.MustCompile(`(?i)onload\s*=`), ""},
		{regexp.MustCompile(`(?i)onerror\s*=`), ""},
		{regexp.MustCompile(`(?i)onclick\s*=`), ""},
	}

	for _, xss := range xssPatterns {
		input = xss.pattern.ReplaceAllString(input, xss.replacement)
	}

	return input
}

// filterSQLInjection 过滤SQL注入
func (ds *DatabaseSecurity) filterSQLInjection(input string) string {
	// 移除常见的SQL注入模式
	sqlPatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)\s+`), ""},
		{regexp.MustCompile(`(?i)[\'\"]\s*(or|and)\s+[\'\"]`), ""},
		{regexp.MustCompile(`(?i)--\s*$`), ""},
		{regexp.MustCompile(`(?i)/\*.*?\*/`), ""},
		{regexp.MustCompile(`(?i)[\'"]\s*(\bor\b|\band\b)\s*[\'"].*[\'"]\s*=\s*[\'"]`), ""},
		{regexp.MustCompile(`(?i)waitfor\s+delay`), ""},
		{regexp.MustCompile(`(?i)sleep\s*\(`), ""},
	}

	for _, sql := range sqlPatterns {
		input = sql.pattern.ReplaceAllString(input, sql.replacement)
	}

	return input
}

// EncryptSensitiveData 加密敏感数据
func (ds *DatabaseSecurity) EncryptSensitiveData(data map[string]interface{}) (map[string]interface{}, error) {
	if ds.encryptor == nil {
		return data, nil
	}

	encrypted := make(map[string]interface{})
	for key, value := range data {
		if ds.isSensitiveField(key) {
			if strValue, ok := value.(string); ok {
				encryptedValue, err := ds.encryptor.EncryptData([]byte(strValue))
				if err != nil {
					return nil, fmt.Errorf("加密字段 %s 失败: %w", key, err)
				}
				encrypted[key] = encryptedValue
			} else {
				encrypted[key] = value
			}
		} else {
			encrypted[key] = value
		}
	}

	return encrypted, nil
}

// DecryptSensitiveData 解密敏感数据
func (ds *DatabaseSecurity) DecryptSensitiveData(data map[string]interface{}) (map[string]interface{}, error) {
	if ds.encryptor == nil {
		return data, nil
	}

	decrypted := make(map[string]interface{})
	for key, value := range data {
		if ds.isEncryptionField(key) {
			if encryptedValue, ok := value.([]byte); ok {
				decryptedValue, err := ds.encryptor.DecryptData(encryptedValue)
				if err != nil {
					return nil, fmt.Errorf("解密字段 %s 失败: %w", key, err)
				}
				decrypted[key] = string(decryptedValue)
			} else {
				decrypted[key] = value
			}
		} else {
			decrypted[key] = value
		}
	}

	return decrypted, nil
}

// isSensitiveField 检查是否为敏感字段
func (ds *DatabaseSecurity) isSensitiveField(fieldName string) bool {
	fieldName = strings.ToLower(fieldName)
	for _, sensitive := range ds.config.SensitiveFields {
		if strings.Contains(fieldName, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// isEncryptionField 检查是否为加密字段
func (ds *DatabaseSecurity) isEncryptionField(fieldName string) bool {
	fieldName = strings.ToLower(fieldName)
	for _, encryption := range ds.config.EncryptionFields {
		if strings.Contains(fieldName, strings.ToLower(encryption)) {
			return true
		}
	}
	return false
}

// CreateSession 创建会话
func (ds *DatabaseSecurity) CreateSession(ctx context.Context, userID int64, userAgent string, ip string) (string, error) {
	sessionID := ds.generateSessionID()

	// 加密会话令牌
	token, err := ds.encryptor.EncryptSessionToken(sessionID, userID, time.Now().Add(time.Duration(ds.config.SessionTimeout)*time.Minute))
	if err != nil {
		return "", fmt.Errorf("创建会话令牌失败: %w", err)
	}

	// 存储会话到数据库
	if err := ds.storeSession(ctx, sessionID, userID, userAgent, ip); err != nil {
		return "", fmt.Errorf("存储会话失败: %w", err)
	}

	logger.Info("创建用户会话",
		zap.Int64("用户ID", userID),
		zap.String("会话ID", sessionID),
		zap.String("IP", ip))

	return token, nil
}

// ValidateSession 验证会话
func (ds *DatabaseSecurity) ValidateSession(ctx context.Context, token string) (*SessionInfo, error) {
	// 解密会话令牌
	sessionID, userID, expires, err := ds.encryptor.DecryptSessionToken(token)
	if err != nil {
		return nil, fmt.Errorf("解密会话令牌失败: %w", err)
	}

	// 检查过期时间
	if time.Now().After(expires) {
		return nil, fmt.Errorf("会话已过期")
	}

	// 从数据库验证会话
	sessionInfo, err := ds.getSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话信息失败: %w", err)
	}

	if sessionInfo == nil || sessionInfo.UserID != userID {
		return nil, fmt.Errorf("无效的会话")
	}

	// 检查会话是否被撤销
	if sessionInfo.Revoked {
		return nil, fmt.Errorf("会话已被撤销")
	}

	// 更新最后访问时间
	if err := ds.updateSessionAccess(ctx, sessionID); err != nil {
		logger.Warn("更新会话访问时间失败", zap.Error(err))
	}

	return sessionInfo, nil
}

// RevokeSession 撤销会话
func (ds *DatabaseSecurity) RevokeSession(ctx context.Context, token string) error {
	sessionID, _, _, err := ds.encryptor.DecryptSessionToken(token)
	if err != nil {
		return fmt.Errorf("解密会话令牌失败: %w", err)
	}

	if err := ds.revokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("撤销会话失败: %w", err)
	}

	logger.Info("撤销会话", zap.String("会话ID", sessionID))
	return nil
}

// RevokeAllSessions 撤销用户所有会话
func (ds *DatabaseSecurity) RevokeAllSessions(ctx context.Context, userID int64) error {
	if err := ds.revokeUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("撤销用户会话失败: %w", err)
	}

	logger.Info("撤销用户所有会话", zap.Int64("用户ID", userID))
	return nil
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID  string    `json:"session_id"`
	UserID     int64     `json:"user_id"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"`
}

// generateSessionID 生成会话ID
func (ds *DatabaseSecurity) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// storeSession 存储会话
func (ds *DatabaseSecurity) storeSession(ctx context.Context, sessionID string, userID int64, userAgent, ip string) error {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_sessions (session_id, user_id, user_agent, ip, created_at, last_access, expires_at, revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(ctx, query,
		sessionID, userID, userAgent, ip,
		time.Now(), time.Now(),
		time.Now().Add(time.Duration(ds.config.SessionTimeout)*time.Minute),
		false,
	)

	return err
}

// getSession 获取会话信息
func (ds *DatabaseSecurity) getSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT session_id, user_id, user_agent, ip, created_at, last_access, expires_at, revoked
		FROM user_sessions 
		WHERE session_id = ? AND revoked = FALSE
	`

	row := db.QueryRow(ctx, query, sessionID)

	var session SessionInfo
	err = row.Scan(&session.SessionID, &session.UserID, &session.UserAgent,
		&session.IP, &session.CreatedAt, &session.LastAccess,
		&session.ExpiresAt, &session.Revoked)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// updateSessionAccess 更新会话访问时间
func (ds *DatabaseSecurity) updateSessionAccess(ctx context.Context, sessionID string) error {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return err
	}

	query := `UPDATE user_sessions SET last_access = ? WHERE session_id = ?`
	_, err = db.Exec(ctx, query, time.Now(), sessionID)
	return err
}

// revokeSession 撤销会话
func (ds *DatabaseSecurity) revokeSession(ctx context.Context, sessionID string) error {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return err
	}

	query := `UPDATE user_sessions SET revoked = TRUE WHERE session_id = ?`
	_, err = db.Exec(ctx, query, sessionID)
	return err
}

// revokeUserSessions 撤销用户所有会话
func (ds *DatabaseSecurity) revokeUserSessions(ctx context.Context, userID int64) error {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return err
	}

	query := `UPDATE user_sessions SET revoked = TRUE WHERE user_id = ?`
	_, err = db.Exec(ctx, query, userID)
	return err
}

// CleanupExpiredSessions 清理过期会话
func (ds *DatabaseSecurity) CleanupExpiredSessions(ctx context.Context) error {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return err
	}

	query := `DELETE FROM user_sessions WHERE expires_at < ? OR (revoked = TRUE AND last_access < ?)`

	deleteTime := time.Now().AddDate(0, 0, -7) // 保留7天的被撤销会话
	result, err := db.Exec(ctx, query, time.Now(), deleteTime)
	if err != nil {
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err == nil {
		logger.Info("清理过期会话", zap.Int64("删除数量", rowsAffected))
	}

	return nil
}

// GetSecurityReport 获取安全报告
type SecurityReport struct {
	ActiveSessions   int                      `json:"active_sessions"`
	ExpiredSessions  int                      `json:"expired_sessions"`
	RevokedSessions  int                      `json:"revoked_sessions"`
	SuspiciousLogins []SuspiciousLoginAttempt `json:"suspicious_logins"`
	LastCleanup      time.Time                `json:"last_cleanup"`
}

// SuspiciousLoginAttempt 可疑登录尝试
type SuspiciousLoginAttempt struct {
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Attempts  int       `json:"attempts"`
	LastTime  time.Time `json:"last_time"`
	Blocked   bool      `json:"blocked"`
}

// GetSecurityReport 获取安全报告
func (ds *DatabaseSecurity) GetSecurityReport(ctx context.Context) (*SecurityReport, error) {
	db, err := ds.dbManager.GetConnection()
	if err != nil {
		return nil, err
	}

	report := &SecurityReport{}

	// 统计活跃会话
	activeQuery := `SELECT COUNT(*) FROM user_sessions WHERE revoked = FALSE AND expires_at > ?`
	row := db.QueryRow(ctx, activeQuery, time.Now())
	if err := row.Scan(&report.ActiveSessions); err != nil {
		logger.Warn("统计活跃会话失败", zap.Error(err))
	}

	// 统计过期会话
	expiredQuery := `SELECT COUNT(*) FROM user_sessions WHERE expires_at <= ?`
	row = db.QueryRow(ctx, expiredQuery, time.Now())
	if err := row.Scan(&report.ExpiredSessions); err != nil {
		logger.Warn("统计过期会话失败", zap.Error(err))
	}

	// 统计被撤销会话
	revokedQuery := `SELECT COUNT(*) FROM user_sessions WHERE revoked = TRUE`
	row = db.QueryRow(ctx, revokedQuery)
	if err := row.Scan(&report.RevokedSessions); err != nil {
		logger.Warn("统计被撤销会话失败", zap.Error(err))
	}

	report.LastCleanup = time.Now()

	return report, nil
}

// GetConfig 获取安全配置
func (ds *DatabaseSecurity) GetConfig() *DatabaseSecurityConfig {
	// 返回配置的副本
	config := *ds.config
	return &config
}

// UpdateConfig 更新安全配置
func (ds *DatabaseSecurity) UpdateConfig(config *DatabaseSecurityConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	ds.config = config
	logger.Info("更新数据库安全配置", zap.Any("配置", config))
	return nil
}
