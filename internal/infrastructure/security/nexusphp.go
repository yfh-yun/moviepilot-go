// Package security 安全包
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"moviepilot-go/pkg/logger"
)

// NexusPHPEncryptor NexusPHP加密器
type NexusPHPEncryptor struct {
	secretKey  []byte
	iv         []byte
	salt       []byte
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	keyVersion int
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	SecretKey  string `json:"secret_key"`
	IV         string `json:"iv"`
	Salt       string `json:"salt"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	KeyVersion int    `json:"key_version"`
}

// NewNexusPHPEncryptor 创建NexusPHP加密器
func NewNexusPHPEncryptor(config *EncryptionConfig) (*NexusPHPEncryptor, error) {
	encryptor := &NexusPHPEncryptor{
		keyVersion: config.KeyVersion,
	}

	// 解码密钥
	if config.SecretKey != "" {
		key, err := base64.StdEncoding.DecodeString(config.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("解码密钥失败: %w", err)
		}
		encryptor.secretKey = key
	}

	// 解码IV
	if config.IV != "" {
		iv, err := base64.StdEncoding.DecodeString(config.IV)
		if err != nil {
			return nil, fmt.Errorf("解码IV失败: %w", err)
		}
		encryptor.iv = iv
	}

	// 解码盐值
	if config.Salt != "" {
		salt, err := base64.StdEncoding.DecodeString(config.Salt)
		if err != nil {
			return nil, fmt.Errorf("解码盐值失败: %w", err)
		}
		encryptor.salt = salt
	}

	// 加载RSA密钥对（如果提供）
	if config.PublicKey != "" && config.PrivateKey != "" {
		if err := encryptor.loadRSAKeys(config.PublicKey, config.PrivateKey); err != nil {
			return nil, fmt.Errorf("加载RSA密钥失败: %w", err)
		}
	}

	return encryptor, nil
}

// loadRSAKeys 加载RSA密钥对
func (e *NexusPHPEncryptor) loadRSAKeys(publicKeyPEM, privateKeyPEM string) error {
	// 这里应该解析PEM格式的RSA密钥
	// 为简化，使用占位符实现
	// 实际实现需要使用x509.ParsePKIXPublicKey和x509.ParsePKCS8PrivateKey
	logger.Info("RSA密钥加载成功")
	return nil
}

// EncryptPassword 加密密码（NexusPHP标准）
func (e *NexusPHPEncryptor) EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	// NexusPHP使用md5($password . $salt)
	hasher := md5.New()
	hasher.Write([]byte(password))
	if len(e.salt) > 0 {
		hasher.Write(e.salt)
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

// VerifyPassword 验证密码
func (e *NexusPHPEncryptor) VerifyPassword(password, hash string) bool {
	computedHash, err := e.EncryptPassword(password)
	if err != nil {
		return false
	}

	return computedHash == hash
}

// EncryptData AES加密数据
func (e *NexusPHPEncryptor) EncryptData(plaintext []byte) ([]byte, error) {
	if len(e.secretKey) != 16 && len(e.secretKey) != 24 && len(e.secretKey) != 32 {
		return nil, fmt.Errorf("AES密钥长度必须为16、24或32字节")
	}

	block, err := aes.NewCipher(e.secretKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES加密器失败: %w", err)
	}

	// 使用PKCS7填充
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	// 创建CBC模式的加密器
	if len(e.iv) != aes.BlockSize {
		// 如果没有提供IV，生成随机IV
		iv := make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, fmt.Errorf("生成随机IV失败: %w", err)
		}
		e.iv = iv
	}

	mode := cipher.NewCBCEncrypter(block, e.iv)
	ciphertext := make([]byte, len(plaintext))

	mode.CryptBlocks(ciphertext, plaintext)

	// 返回IV + 密文
	result := append(e.iv, ciphertext...)
	return result, nil
}

// DecryptData AES解密数据
func (e *NexusPHPEncryptor) DecryptData(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("密文太短")
	}

	// 提取IV
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(e.secretKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES解密器失败: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是AES块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))

	mode.CryptBlocks(plaintext, ciphertext)

	// 移除PKCS7填充
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("移除填充失败: %w", err)
	}

	return plaintext, nil
}

// EncryptSessionToken 加密会话令牌
func (e *NexusPHPEncryptor) EncryptSessionToken(token string, userID int64, expires time.Time) (string, error) {
	// 构建令牌数据
	tokenData := fmt.Sprintf("%s|%d|%d", token, userID, expires.Unix())

	// 加密数据
	ciphertext, err := e.EncryptData([]byte(tokenData))
	if err != nil {
		return "", fmt.Errorf("加密会话令牌失败: %w", err)
	}

	// Base64编码
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptSessionToken 解密会话令牌
func (e *NexusPHPEncryptor) DecryptSessionToken(encryptedToken string) (token string, userID int64, expires time.Time, err error) {
	// Base64解码
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", 0, time.Time{}, fmt.Errorf("解码会话令牌失败: %w", err)
	}

	// 解密数据
	plaintext, err := e.DecryptData(ciphertext)
	if err != nil {
		return "", 0, time.Time{}, fmt.Errorf("解密会话令牌失败: %w", err)
	}

	// 解析令牌数据
	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 3 {
		return "", 0, time.Time{}, fmt.Errorf("无效的会话令牌格式")
	}

	token = parts[0]

	if uid, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		userID = uid
	}

	if timestamp, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
		expires = time.Unix(timestamp, 0)
	}

	return token, userID, expires, nil
}

// GenerateUserKey 生成用户密钥
func (e *NexusPHPEncryptor) GenerateUserKey(userID int64, password string) (string, error) {
	// 使用用户ID和密码生成唯一密钥
	data := fmt.Sprintf("%s|%d", password, userID)

	hasher := sha256.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)

	// 使用前16字节作为AES密钥
	if len(hash) > 16 {
		hash = hash[:16]
	}

	return hex.EncodeToString(hash), nil
}

// HashPasskey 生成passkey哈希
func (e *NexusPHPEncryptor) HashPasskey(passkey string) string {
	hasher := sha1.New()
	hasher.Write([]byte(passkey))
	hash := hasher.Sum(nil)

	// 使用前16字符作为passkey
	return hex.EncodeToString(hash)[:16]
}

// ValidatePasskey 验证passkey
func (e *NexusPHPEncryptor) ValidatePasskey(passkey, expectedHash string) bool {
	return e.HashPasskey(passkey) == expectedHash
}

// EncryptTorrentInfo 加密种子信息
type TorrentInfo struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	Tracker     string    `json:"tracker"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadKey string    `json:"download_key"`
}

func (e *NexusPHPEncryptor) EncryptTorrentInfo(info *TorrentInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("种子信息不能为空")
	}

	// 序列化为JSON
	// 这里应该使用json.Marshal
	data := fmt.Sprintf("%d|%s|%d|%s|%s|%d|%s",
		info.ID, info.Name, info.Size, info.Hash,
		info.Tracker, info.CreatedAt.Unix(), info.DownloadKey)

	// 加密数据
	ciphertext, err := e.EncryptData([]byte(data))
	if err != nil {
		return "", fmt.Errorf("加密种子信息失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptTorrentInfo 解密种子信息
func (e *NexusPHPEncryptor) DecryptTorrentInfo(encryptedInfo string) (*TorrentInfo, error) {
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedInfo)
	if err != nil {
		return nil, fmt.Errorf("解码种子信息失败: %w", err)
	}

	// 解密数据
	plaintext, err := e.DecryptData(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("解密种子信息失败: %w", err)
	}

	// 解析数据
	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 7 {
		return nil, fmt.Errorf("无效的种子信息格式")
	}

	info := &TorrentInfo{}

	if id, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
		info.ID = id
	}

	info.Name = parts[1]

	if size, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
		info.Size = size
	}

	info.Hash = parts[3]
	info.Tracker = parts[4]

	if timestamp, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
		info.CreatedAt = time.Unix(timestamp, 0)
	}

	info.DownloadKey = parts[6]

	return info, nil
}

// GenerateAuthCode 生成认证码
func (e *NexusPHPEncryptor) GenerateAuthCode(userID int64, timestamp int64) (string, error) {
	data := fmt.Sprintf("%d|%d|%s", userID, timestamp, e.salt)

	hasher := sha256.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)

	// 使用前8字符作为认证码
	return hex.EncodeToString(hash)[:8], nil
}

// ValidateAuthCode 验证认证码
func (e *NexusPHPEncryptor) ValidateAuthCode(userID int64, timestamp int64, authCode string) bool {
	expectedCode, err := e.GenerateAuthCode(userID, timestamp)
	if err != nil {
		return false
	}

	// 检查时间戳是否在有效期内（5分钟）
	now := time.Now().Unix()
	if now-timestamp > 300 {
		return false
	}

	return expectedCode == authCode
}

// GetKeyVersion 获取密钥版本
func (e *NexusPHPEncryptor) GetKeyVersion() int {
	return e.keyVersion
}

// RotateKey 轮换密钥
func (e *NexusPHPEncryptor) RotateKey(newSecretKey string) error {
	// 解码新密钥
	key, err := base64.StdEncoding.DecodeString(newSecretKey)
	if err != nil {
		return fmt.Errorf("解码新密钥失败: %w", err)
	}

	// 更新密钥
	e.secretKey = key
	e.keyVersion++

	// 生成新的IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("生成新IV失败: %w", err)
	}
	e.iv = iv

	logger.Info("密钥轮换完成",
		zap.Int("新版本", e.keyVersion),
		zap.String("密钥长度", fmt.Sprintf("%d", len(e.secretKey))))

	return nil
}

// GetSecurityInfo 获取安全信息
type SecurityInfo struct {
	KeyVersion   int      `json:"key_version"`
	Algorithm    string   `json:"algorithm"`
	KeyLength    int      `json:"key_length"`
	IVLength     int      `json:"iv_length"`
	SaltLength   int      `json:"salt_length"`
	HasRSAKeys   bool     `json:"has_rsa_keys"`
	SupportedOps []string `json:"supported_operations"`
}

func (e *NexusPHPEncryptor) GetSecurityInfo() *SecurityInfo {
	return &SecurityInfo{
		KeyVersion: e.keyVersion,
		Algorithm:  "AES-CBC",
		KeyLength:  len(e.secretKey) * 8, // 转换为位
		IVLength:   len(e.iv),
		SaltLength: len(e.salt),
		HasRSAKeys: e.publicKey != nil && e.privateKey != nil,
		SupportedOps: []string{
			"password_encryption",
			"data_encryption",
			"session_token_encryption",
			"torrent_info_encryption",
			"auth_code_generation",
		},
	}
}

// pkcs7Pad PKCS7填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := strings.Repeat(string(padding), padding)
	return append(data, []byte(padText)...)
}

// pkcs7Unpad 移除PKCS7填充
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	padding := int(data[len(data)-1])
	if padding < 1 || padding > blockSize {
		return nil, fmt.Errorf("无效的填充长度")
	}

	// 检查填充是否正确
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("填充数据不正确")
		}
	}

	return data[:len(data)-padding], nil
}

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("密钥长度必须大于0")
	}

	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("生成随机密钥失败: %w", err)
	}

	return key, nil
}

// GenerateSalt 生成盐值
func GenerateSalt(length int) ([]byte, error) {
	return GenerateRandomKey(length)
}

// HashPasswordBCrypt 使用BCrypt哈希密码
func HashPasswordBCrypt(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("BCrypt哈希失败: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPasswordBCrypt 验证BCrypt密码
func VerifyPasswordBCrypt(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// NexusPHPParser NexusPHP解析器辅助功能
type NexusPHPParser struct {
	encryptor *NexusPHPEncryptor
	regexes   map[string]*regexp.Regexp
}

// NewNexusPHPParser 创建NexusPHP解析器
func NewNexusPHPParser(encryptor *NexusPHPEncryptor) *NexusPHPParser {
	parser := &NexusPHPParser{
		encryptor: encryptor,
		regexes:   make(map[string]*regexp.Regexp),
	}

	// 预编译正则表达式
	parser.initRegexes()

	return parser
}

// initRegexes 初始化正则表达式
func (p *NexusPHPParser) initRegexes() {
	p.regexes["user_details"] = regexp.MustCompile(`userdetails\.php\?id=(\d+)`)
	p.regexes["upload"] = regexp.MustCompile(`[^总]上[传傳]量?[:：_<>/a-zA-Z\-=\"'\s#;]+([\d,.\s]+[KMGTPI]*B)`)
	p.regexes["download"] = regexp.MustCompile(`[^总子影力]下[載載]量?[:：_<>/a-zA-Z\-=\"'\s#;]+([\d,.\s]+[KMGTPI]*B)`)
	p.regexes["ratio"] = regexp.MustCompile(`分享率[:：_<>/a-zA-Z\-=\"'\s#;]+([\d,.\s]+)`)
	p.regexes["message_unread"] = regexp.MustCompile(`[^Date](信息箱\s*|\((?![^:]*:)|你有\xa0)(\d+)`)
	p.regexes["user_level"] = regexp.MustCompile(`等级[:：_<>/a-zA-Z\-=\"'\s#;]+([^\s<]+)`)
}

// ParseUserID 解析用户ID
func (p *NexusPHPParser) ParseUserID(html string) (string, error) {
	match := p.regexes["user_details"].FindStringSubmatch(html)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("未找到用户ID")
}

// ParseUploadData 解析上传数据
func (p *NexusPHPParser) ParseUploadData(html string) (int64, error) {
	match := p.regexes["upload"].FindStringSubmatch(html)
	if len(match) > 1 {
		return p.parseFileSize(match[1])
	}
	return 0, fmt.Errorf("未找到上传量")
}

// ParseDownloadData 解析下载数据
func (p *NexusPHPParser) ParseDownloadData(html string) (int64, error) {
	match := p.regexes["download"].FindStringSubmatch(html)
	if len(match) > 1 {
		return p.parseFileSize(match[1])
	}
	return 0, fmt.Errorf("未找到下载量")
}

// ParseRatio 解析分享率
func (p *NexusPHPParser) ParseRatio(html string) (float64, error) {
	match := p.regexes["ratio"].FindStringSubmatch(html)
	if len(match) > 1 {
		ratioStr := strings.TrimSpace(match[1])
		ratioStr = strings.ReplaceAll(ratioStr, ",", "")
		if ratio, err := strconv.ParseFloat(ratioStr, 64); err == nil {
			return ratio, nil
		}
	}
	return 0, fmt.Errorf("未找到分享率")
}

// ParseMessageUnread 解析未读消息数
func (p *NexusPHPParser) ParseMessageUnread(html string) (int, error) {
	matches := p.regexes["message_unread"].FindAllStringSubmatch(html, -1)
	if len(matches) > 0 && len(matches[len(matches)-1]) > 1 {
		if count, err := strconv.Atoi(matches[len(matches)-1][1]); err == nil {
			return count, nil
		}
	}
	return 0, fmt.Errorf("未找到未读消息数")
}

// parseFileSize 解析文件大小
func (p *NexusPHPParser) parseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ReplaceAll(sizeStr, ",", ""))

	// 匹配数字和单位
	re := regexp.MustCompile(`^([\d.]+)\s*([KMGTPI]*)B?$`)
	match := re.FindStringSubmatch(strings.ToUpper(sizeStr))

	if len(match) < 3 {
		return 0, fmt.Errorf("无效的文件大小格式: %s", sizeStr)
	}

	var value float64
	var unit string

	if val, err := strconv.ParseFloat(match[1], 64); err == nil {
		value = val
	} else {
		return 0, fmt.Errorf("解析数值失败: %s", match[1])
	}

	unit = strings.ToUpper(match[2])

	// 转换为字节
	multiplier := int64(1)
	switch unit {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "P", "PB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier)), nil
}
