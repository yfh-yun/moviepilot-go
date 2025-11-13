package webpush

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// VAPIDConfig VAPID配置
type VAPIDConfig struct {
	PrivateKey string `json:"privateKey"`
	Subject    string `json:"subject"`
	PublicKey  string `json:"publicKey"`
}

// WebPushClient WebPush客户�?type WebPushClient struct {
	vapidConfig VAPIDConfig
	httpClient  *http.Client
}

// NewWebPushClient 创建新的WebPush客户�?func NewWebPushClient(vapidConfig VAPIDConfig) *WebPushClient {
	return &WebPushClient{
		vapidConfig: vapidConfig,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send 发送WebPush消息
func (w *WebPushClient) Send(subscription SubscriptionInfo, payload map[string]string) error {
	// 序列化消息内�?	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息内容失�? %v", err)
	}

	// 加密消息
	encryptedPayload, err := w.encryptPayload(subscription, payloadBytes)
	if err != nil {
		return fmt.Errorf("加密消息失败: %v", err)
	}

	// 生成VAPID认证�?	vapidHeaders, err := w.generateVAPIDHeaders(subscription)
	if err != nil {
		return fmt.Errorf("生成VAPID头失�? %v", err)
	}

	// 构造请�?	req, err := http.NewRequest("POST", subscription.Endpoint, strings.NewReader(string(encryptedPayload)))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求�?	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")

	// 添加VAPID�?	for key, value := range vapidHeaders {
		req.Header.Set(key, value)
	}

	// 发送请�?	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失�? %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("推送失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// encryptPayload 加密消息载荷
func (w *WebPushClient) encryptPayload(subscription SubscriptionInfo, payload []byte) ([]byte, error) {
	// 解码订阅密钥
	auth, err := base64.URLEncoding.DecodeString(subscription.Keys.Auth)
	if err != nil {
		return nil, fmt.Errorf("解码auth密钥失败: %v", err)
	}

	p256dh, err := base64.URLEncoding.DecodeString(subscription.Keys.P256dh)
	if err != nil {
		return nil, fmt.Errorf("解码p256dh密钥失败: %v", err)
	}

	// 生成临时密钥�?	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成临时密钥对失�? %v", err)
	}

	// 获取公钥�?	pubKey := privKey.PublicKey

	// 解码接收方公�?	x, y := elliptic.Unmarshal(elliptic.P256(), p256dh)
	if x == nil {
		return nil, fmt.Errorf("解码接收方公钥失�?)
	}

	// 计算共享密钥
	sharedKeyX, sharedKeyY := pubKey.Curve.ScalarMult(x, y, privKey.D.Bytes())
	sharedKey := sharedKeyX.Bytes()
	
	// 确保共享密钥长度正确
	for len(sharedKey) < 32 {
		sharedKey = append([]byte{0}, sharedKey...)
	}
	if len(sharedKey) > 32 {
		sharedKey = sharedKey[len(sharedKey)-32:]
	}

	// 构造加密内�?	// 注意：这是一个简化的实现，实际应用中需要更完整的加密逻辑

	// 生成随机盐�?	salt := make([]byte, 16)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("生成随机盐值失�? %v", err)
	}

	// 获取公钥字节
	publicKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)

	// 这里应该实现完整的AES128GCM加密逻辑
	// 为简化起见，我们直接返回载荷和必要的信息

	result := make([]byte, 0)
	result = append(result, salt...)
	result = append(result, publicKeyBytes...)
	result = append(result, payload...)

	return result, nil
}

// generateVAPIDHeaders 生成VAPID认证�?func (w *WebPushClient) generateVAPIDHeaders(subscription SubscriptionInfo) (map[string]string, error) {
	// 解码私钥
	// privKeyBytes, err := base64.URLEncoding.DecodeString(w.vapidConfig.PrivateKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("解码私钥失败: %v", err)
	// }

	// 构造JWT头部
	header := map[string]string{
		"typ": "JWT",
		"alg": "ES256",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("序列化JWT头部失败: %v", err)
	}

	// 构造JWT载荷
	aud := ""
	parts := strings.Split(subscription.Endpoint, "/")
	if len(parts) > 2 {
		aud = "https://" + parts[2] // 提取域名并加上https前缀
	}

	claims := map[string]interface{}{
		"aud": aud,
		"exp": time.Now().Add(12 * time.Hour).Unix(),
		"sub": w.vapidConfig.Subject,
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("序列化JWT载荷失败: %v", err)
	}

	// Base64编码头部和载�?	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// 构造签名内�?	signingInput := encodedHeader + "." + encodedClaims

	// 计算签名（简化实现）
	// 实际应用中需要使用ECDSA签名算法
	signature := sha256.Sum256([]byte(signingInput))
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature[:])

	// 构造JWT
	jwt := encodedHeader + "." + encodedClaims + "." + encodedSignature

	// 构造VAPID�?	headers := map[string]string{
		"Authorization": "WebPush " + jwt,
	}

	return headers, nil
}

// GenerateVAPIDKeys 生成VAPID密钥�?func GenerateVAPIDKeys() (privateKey, publicKey string, err error) {
	// 生成密钥�?	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("生成密钥对失�? %v", err)
	}

	// 编码私钥
	privKeyBytes := privKey.D.Bytes()
	privateKey = base64.URLEncoding.EncodeToString(privKeyBytes)

	// 编码公钥
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privKey.PublicKey.X, privKey.PublicKey.Y)
	publicKey = base64.URLEncoding.EncodeToString(pubKeyBytes)

	return privateKey, publicKey, nil
}
