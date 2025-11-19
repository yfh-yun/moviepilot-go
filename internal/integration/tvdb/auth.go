package tvdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthManager 认证管理器
type AuthManager struct {
	client      *Client
	mu          sync.RWMutex
	lastRefresh time.Time
	tokenTTL    time.Duration
}

// NewAuthManager 创建认证管理器
func NewAuthManager(client *Client) *AuthManager {
	return &AuthManager{
		client:   client,
		tokenTTL: 24 * time.Hour, // TVDB令牌TTL通常为24小时
	}
}

// RefreshToken 刷新令牌
func (am *AuthManager) RefreshToken() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 检查是否需要刷新
	if time.Since(am.lastRefresh) < am.tokenTTL-time.Hour {
		return nil // 令牌还有效
	}

	// 执行认证
	payload := map[string]string{
		"apikey": am.client.apiKey,
		"pin":    am.client.pin,
	}

	resp, err := am.client.httpClient.Post(context.Background(), "login", nil, payload)
	if err != nil {
		return fmt.Errorf("刷新令牌请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("刷新令牌失败，状态码: %d", resp.StatusCode)
	}

	var authResp struct {
		Status string `json:"status"`
		Data   struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("解析刷新响应失败: %w", err)
	}

	if authResp.Status != "success" {
		return fmt.Errorf("刷新令牌失败: %s", authResp.Status)
	}

	// 更新令牌和刷新时间
	am.client.accessToken = authResp.Data.Token
	am.lastRefresh = time.Now()

	return nil
}

// IsTokenValid 检查令牌是否有效
func (am *AuthManager) IsTokenValid() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.client.accessToken == "" {
		return false
	}

	return time.Since(am.lastRefresh) < am.tokenTTL
}

// HandleAuthError 处理认证错误
func (am *AuthManager) HandleAuthError(err error) error {
	if IsUnauthorizedError(err) {
		// 尝试刷新令牌
		if refreshErr := am.RefreshToken(); refreshErr != nil {
			return fmt.Errorf("认证失败且刷新令牌失败: %w", refreshErr)
		}
		return nil // 令牌已刷新，可以重试
	}
	return err
}

// ForceRefresh 强制刷新令牌
func (am *AuthManager) ForceRefresh() error {
	am.lastRefresh = time.Time{} // 设置为零时间，强制刷新
	return am.RefreshToken()
}

// GetTokenInfo 获取令牌信息
func (am *AuthManager) GetTokenInfo() (token string, refreshed time.Time, ttl time.Duration) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.client.accessToken, am.lastRefresh, am.tokenTTL - time.Since(am.lastRefresh)
}

// Reset 重置认证状态
func (am *AuthManager) Reset() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.client.accessToken = ""
	am.lastRefresh = time.Time{}
}

// IsUnauthorizedError 检查是否为认证错误
func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误消息中是否包含认证相关的关键词
	errorMsg := err.Error()
	unauthorizedKeywords := []string{
		"401",
		"unauthorized",
		"认证失败",
		"token expired",
		"invalid token",
	}

	for _, keyword := range unauthorizedKeywords {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(keyword)) {
			return true
		}
	}

	// 使用TVDB错误类型检查
	return IsAuthError(err)
}
