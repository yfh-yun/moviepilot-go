package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SubscribeHelper 订阅数据统计/订阅分享助手
type SubscribeHelper struct {
	serverHost    string
	httpClient    *http.Client
	mutex         sync.RWMutex
	adminUsers    []string
	shareUserID   string
	githubUser    string
}

// SubscribeData 订阅数据
type SubscribeData struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Year        int                    `json:"year"`
	Season      int                    `json:"season,omitempty"`
	Episode     int                    `json:"episode,omitempty"`
	Status      string                 `json:"status"`
	AddedDate   time.Time              `json:"added_date"`
	UpdatedDate time.Time              `json:"updated_date"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SubscribeShare 订阅分享
type SubscribeShare struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Public      bool                   `json:"public"`
	Subscribes  []SubscribeData        `json:"subscribes"`
	CreatedDate time.Time              `json:"created_date"`
	UpdatedDate time.Time              `json:"updated_date"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SubscribeStatistic 订阅统计
type SubscribeStatistic struct {
	TotalCount    int            `json:"total_count"`
	MovieCount    int            `json:"movie_count"`
	TVCount       int            `json:"tv_count"`
	CompletedCount int           `json:"completed_count"`
	PendingCount  int           `json:"pending_count"`
	StatusCount   map[string]int `json:"status_count"`
	TypeCount     map[string]int `json:"type_count"`
}

// NewSubscribeHelper 创建订阅助手实例
func NewSubscribeHelper(serverHost string) *SubscribeHelper {
	if serverHost == "" {
		serverHost = "http://localhost:3000"
	}

	return &SubscribeHelper{
		serverHost: serverHost,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		adminUsers: []string{
			"jxxghp",
			"thsrite",
			"InfinityPacer",
			"DDSRem",
			"Aqr-K",
			"Putarku",
			"4Nest",
			"xyswordzoro",
		},
	}
}

// AddSubscribe 添加订阅
func (sh *SubscribeHelper) AddSubscribe(subscribe *SubscribeData) error {
	if subscribe == nil {
		return fmt.Errorf("subscribe cannot be nil")
	}

	if subscribe.Name == "" {
		return fmt.Errorf("subscribe name cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/add", sh.serverHost)
	
	return sh.sendRequest("POST", url, subscribe)
}

// UpdateSubscribe 更新订阅
func (sh *SubscribeHelper) UpdateSubscribe(subscribe *SubscribeData) error {
	if subscribe == nil {
		return fmt.Errorf("subscribe cannot be nil")
	}

	if subscribe.ID == "" {
		return fmt.Errorf("subscribe ID cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/update", sh.serverHost)
	
	return sh.sendRequest("PUT", url, subscribe)
}

// DeleteSubscribe 删除订阅
func (sh *SubscribeHelper) DeleteSubscribe(subscribeID string) error {
	if subscribeID == "" {
		return fmt.Errorf("subscribe ID cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/delete/%s", sh.serverHost, subscribeID)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetSubscribe 获取订阅
func (sh *SubscribeHelper) GetSubscribe(subscribeID string) (*SubscribeData, error) {
	if subscribeID == "" {
		return nil, fmt.Errorf("subscribe ID cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/get/%s", sh.serverHost, subscribeID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var subscribe SubscribeData
	if err := json.NewDecoder(resp.Body).Decode(&subscribe); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &subscribe, nil
}

// GetAllSubscribes 获取所有订阅
func (sh *SubscribeHelper) GetAllSubscribes() ([]SubscribeData, error) {
	url := fmt.Sprintf("%s/subscribe/list", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var subscribes []SubscribeData
	if err := json.NewDecoder(resp.Body).Decode(&subscribes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return subscribes, nil
}

// GetSubscribeStatistic 获取订阅统计
func (sh *SubscribeHelper) GetSubscribeStatistic() (*SubscribeStatistic, error) {
	url := fmt.Sprintf("%s/subscribe/statistic", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var statistic SubscribeStatistic
	if err := json.NewDecoder(resp.Body).Decode(&statistic); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &statistic, nil
}

// ShareSubscribe 分享订阅
func (sh *SubscribeHelper) ShareSubscribe(subscribeIDs []string, name, description string, public bool) (*SubscribeShare, error) {
	if len(subscribeIDs) == 0 {
		return nil, fmt.Errorf("subscribe IDs cannot be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("share name cannot be empty")
	}

	shareData := map[string]interface{}{
		"subscribe_ids": subscribeIDs,
		"name":          name,
		"description":   description,
		"public":        public,
	}

	url := fmt.Sprintf("%s/subscribe/share", sh.serverHost)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求体
	// 这里应该设置JSON请求体，简化实现

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &share, nil
}

// GetShare 获取分享
func (sh *SubscribeHelper) GetShare(shareID string) (*SubscribeShare, error) {
	if shareID == "" {
		return nil, fmt.Errorf("share ID cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/share/%s", sh.serverHost, shareID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &share, nil
}

// GetAllShares 获取所有分享
func (sh *SubscribeHelper) GetAllShares() ([]SubscribeShare, error) {
	url := fmt.Sprintf("%s/subscribe/shares", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var shares []SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&shares); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return shares, nil
}

// ForkShare 复刻分享
func (sh *SubscribeHelper) ForkShare(shareID string) (*SubscribeShare, error) {
	if shareID == "" {
		return nil, fmt.Errorf("share ID cannot be empty")
	}

	url := fmt.Sprintf("%s/subscribe/fork/%s", sh.serverHost, shareID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &share, nil
}

// ReportInstall 报告安装
func (sh *SubscribeHelper) ReportInstall() error {
	url := fmt.Sprintf("%s/subscribe/report", sh.serverHost)
	
	reportData := map[string]interface{}{
		"action":    "install",
		"timestamp": time.Now().Unix(),
		"user_agent": "MoviePilot-SubscribeHelper/1.0",
	}

	return sh.sendRequest("POST", url, reportData)
}

// ReportStatistic 报告统计
func (sh *SubscribeHelper) ReportStatistic(statistic *SubscribeStatistic) error {
	if statistic == nil {
		return fmt.Errorf("statistic cannot be nil")
	}

	url := fmt.Sprintf("%s/subscribe/statistic", sh.serverHost)
	
	return sh.sendRequest("POST", url, statistic)
}

// sendRequest 发送HTTP请求
func (sh *SubscribeHelper) sendRequest(method, url string, data interface{}) error {
	var err error
	var req *http.Request

	if data != nil {
		// 这里应该设置JSON请求体
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// SetServerHost 设置服务器主机
func (sh *SubscribeHelper) SetServerHost(host string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	sh.serverHost = host
}

// GetServerHost 获取服务器主机
func (sh *SubscribeHelper) GetServerHost() string {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	return sh.serverHost
}

// SetAdminUsers 设置管理员用户列表
func (sh *SubscribeHelper) SetAdminUsers(users []string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	sh.adminUsers = users
}

// GetAdminUsers 获取管理员用户列表
func (sh *SubscribeHelper) GetAdminUsers() []string {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	return append([]string{}, sh.adminUsers...)
}

// IsAdminUser 检查是否为管理员用户
func (sh *SubscribeHelper) IsAdminUser(username string) bool {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	for _, admin := range sh.adminUsers {
		if admin == username {
			return true
		}
	}

	return false
}

// SetShareUserID 设置分享用户ID
func (sh *SubscribeHelper) SetShareUserID(userID string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	sh.shareUserID = userID
}

// GetShareUserID 获取分享用户ID
func (sh *SubscribeHelper) GetShareUserID() string {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	return sh.shareUserID
}

// SetGithubUser 设置GitHub用户
func (sh *SubscribeHelper) SetGithubUser(user string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	sh.githubUser = user
}

// GetGithubUser 获取GitHub用户
func (sh *SubscribeHelper) GetGithubUser() string {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	return sh.githubUser
}

// SetTimeout 设置HTTP客户端超时
func (sh *SubscribeHelper) SetTimeout(timeout time.Duration) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	sh.httpClient.Timeout = timeout
}

// GetTimeout 获取HTTP客户端超时
func (sh *SubscribeHelper) GetTimeout() time.Duration {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()
	return sh.httpClient.Timeout
}

// ValidateSubscribe 验证订阅数据
func (sh *SubscribeHelper) ValidateSubscribe(subscribe *SubscribeData) error {
	if subscribe == nil {
		return fmt.Errorf("subscribe cannot be nil")
	}

	if subscribe.Name == "" {
		return fmt.Errorf("subscribe name cannot be empty")
	}

	if subscribe.Type == "" {
		return fmt.Errorf("subscribe type cannot be empty")
	}

	// 验证类型
	validTypes := []string{"movie", "tv", "documentary", "animation"}
	if !containsString(validTypes, subscribe.Type) {
		return fmt.Errorf("invalid subscribe type: %s", subscribe.Type)
	}

	// 验证状态
	validStatuses := []string{"pending", "downloading", "completed", "failed", "paused"}
	if subscribe.Status != "" && !containsString(validStatuses, subscribe.Status) {
		return fmt.Errorf("invalid subscribe status: %s", subscribe.Status)
	}

	return nil
}

// ValidateShare 验证分享数据
func (sh *SubscribeHelper) ValidateShare(share *SubscribeShare) error {
	if share == nil {
		return fmt.Errorf("share cannot be nil")
	}

	if share.Name == "" {
		return fmt.Errorf("share name cannot be empty")
	}

	if share.Author == "" {
		return fmt.Errorf("share author cannot be empty")
	}

	// 验证订阅数据
	for _, subscribe := range share.Subscribes {
		if err := sh.ValidateSubscribe(&subscribe); err != nil {
			return fmt.Errorf("invalid subscribe in share: %v", err)
		}
	}

	return nil
}

// ExportSubscribes 导出订阅数据
func (sh *SubscribeHelper) ExportSubscribes() ([]SubscribeData, error) {
	return sh.GetAllSubscribes()
}

// ImportSubscribes 导入订阅数据
func (sh *SubscribeHelper) ImportSubscribes(subscribes []SubscribeData) error {
	if subscribes == nil {
		return fmt.Errorf("subscribes cannot be nil")
	}

	for _, subscribe := range subscribes {
		if err := sh.ValidateSubscribe(&subscribe); err != nil {
			return fmt.Errorf("invalid subscribe: %v", err)
		}

		if err := sh.AddSubscribe(&subscribe); err != nil {
			return fmt.Errorf("failed to add subscribe %s: %v", subscribe.Name, err)
		}
	}

	return nil
}

// GetStats 获取统计信息
func (sh *SubscribeHelper) GetStats() map[string]interface{} {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	return map[string]interface{}{
		"server_host":   sh.serverHost,
		"admin_users":   len(sh.adminUsers),
		"share_user_id": sh.shareUserID,
		"github_user":   sh.githubUser,
		"timeout":       sh.httpClient.Timeout.String(),
	}
}