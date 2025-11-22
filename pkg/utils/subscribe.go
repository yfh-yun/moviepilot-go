package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
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
	logger.Debug("SubscribeHelper.NewSubscribeHelper called", zap.String("func", "NewSubscribeHelper"), zap.String("server_host", serverHost))
	
	if serverHost == "" {
		serverHost = "http://localhost:3000"
		logger.Info("Using default server host", zap.String("default_host", serverHost))
	}

	helper := &SubscribeHelper{
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

	logger.Info("SubscribeHelper created successfully", 
		zap.String("server_host", helper.serverHost),
		zap.Int("admin_users_count", len(helper.adminUsers)),
		zap.Duration("timeout", helper.httpClient.Timeout))

	return helper
}

// AddSubscribe 添加订阅
func (sh *SubscribeHelper) AddSubscribe(subscribe *SubscribeData) error {
	logger.Debug("SubscribeHelper.AddSubscribe called", zap.String("func", "AddSubscribe"), zap.Bool("has_subscribe", subscribe != nil), zap.String("subscribe_name", func() string { if subscribe != nil { return subscribe.Name }; return "" }()))

	if subscribe == nil {
		logger.Error("Subscribe cannot be nil", zap.String("func", "AddSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe cannot be nil", "")
	}

	if subscribe.Name == "" {
		logger.Error("Subscribe name cannot be empty", zap.String("func", "AddSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe name cannot be empty", "")
	}

	logger.Info("Adding subscribe", zap.String("func", "AddSubscribe"), zap.String("name", subscribe.Name), zap.String("type", subscribe.Type), zap.Int("year", subscribe.Year))

	url := fmt.Sprintf("%s/subscribe/add", sh.serverHost)
	err := sh.sendRequest("POST", url, subscribe)
	if err != nil {
		logger.Error("Failed to add subscribe", zap.String("func", "AddSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_name", subscribe.Name))
		return err
	}

	logger.Info("Subscribe added successfully", zap.String("func", "AddSubscribe"), zap.String("subscribe_name", subscribe.Name))
	return nil
}

// UpdateSubscribe 更新订阅
func (sh *SubscribeHelper) UpdateSubscribe(subscribe *SubscribeData) error {
	logger.Debug("SubscribeHelper.UpdateSubscribe called", zap.String("func", "UpdateSubscribe"), zap.Bool("has_subscribe", subscribe != nil), zap.String("subscribe_id", func() string { if subscribe != nil { return subscribe.ID }; return "" }()))

	if subscribe == nil {
		logger.Error("Subscribe cannot be nil", zap.String("func", "UpdateSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe cannot be nil", "")
	}

	if subscribe.ID == "" {
		logger.Error("Subscribe ID cannot be empty", zap.String("func", "UpdateSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe ID cannot be empty", "")
	}

	logger.Info("Updating subscribe", zap.String("func", "UpdateSubscribe"), zap.String("id", subscribe.ID), zap.String("name", subscribe.Name), zap.String("status", subscribe.Status))

	url := fmt.Sprintf("%s/subscribe/update", sh.serverHost)
	err := sh.sendRequest("PUT", url, subscribe)
	if err != nil {
		logger.Error("Failed to update subscribe", zap.String("func", "UpdateSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribe.ID))
		return err
	}

	logger.Info("Subscribe updated successfully", zap.String("func", "UpdateSubscribe"), zap.String("subscribe_id", subscribe.ID))
	return nil
}

// DeleteSubscribe 删除订阅
func (sh *SubscribeHelper) DeleteSubscribe(subscribeID string) error {
	logger.Debug("SubscribeHelper.DeleteSubscribe called", zap.String("func", "DeleteSubscribe"), zap.String("subscribe_id", subscribeID))

	if subscribeID == "" {
		logger.Error("Subscribe ID cannot be empty", zap.String("func", "DeleteSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe ID cannot be empty", "")
	}

	logger.Info("Deleting subscribe", zap.String("func", "DeleteSubscribe"), zap.String("subscribe_id", subscribeID))

	url := fmt.Sprintf("%s/subscribe/delete/%s", sh.serverHost, subscribeID)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logger.Error("Failed to create delete request", zap.String("func", "DeleteSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribeID))
		return errors.WrapError(err, "failed to create request")
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send delete request", zap.String("func", "DeleteSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribeID))
		return errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status", zap.String("func", "DeleteSubscribe"), zap.Int("status_code", resp.StatusCode), zap.String("subscribe_id", subscribeID))
		return errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	logger.Info("Subscribe deleted successfully", zap.String("func", "DeleteSubscribe"), zap.String("subscribe_id", subscribeID))
	return nil
}

// GetSubscribe 获取订阅
func (sh *SubscribeHelper) GetSubscribe(subscribeID string) (*SubscribeData, error) {
	logger.Debug("SubscribeHelper.GetSubscribe called", zap.String("func", "GetSubscribe"), zap.String("subscribe_id", subscribeID))

	if subscribeID == "" {
		logger.Error("Subscribe ID cannot be empty", zap.String("func", "GetSubscribe"))
		return nil, errors.NewAppError(http.StatusBadRequest, "subscribe ID cannot be empty", "")
	}

	logger.Info("Getting subscribe", zap.String("func", "GetSubscribe"), zap.String("subscribe_id", subscribeID))

	url := fmt.Sprintf("%s/subscribe/get/%s", sh.serverHost, subscribeID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create get request", zap.String("func", "GetSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribeID))
		return nil, errors.WrapError(err, "failed to create request")
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send get request", zap.String("func", "GetSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribeID))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status", zap.String("func", "GetSubscribe"), zap.Int("status_code", resp.StatusCode), zap.String("subscribe_id", subscribeID))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var subscribe SubscribeData
	if err := json.NewDecoder(resp.Body).Decode(&subscribe); err != nil {
		logger.Error("Failed to decode response", zap.String("func", "GetSubscribe"), zap.String("error", err.Error()), zap.String("subscribe_id", subscribeID))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("Subscribe retrieved successfully", zap.String("func", "GetSubscribe"), zap.String("subscribe_id", subscribeID), zap.String("subscribe_name", subscribe.Name))
	return &subscribe, nil
}

// GetAllSubscribes 获取所有订阅
func (sh *SubscribeHelper) GetAllSubscribes() ([]SubscribeData, error) {
	logger.Debug("SubscribeHelper.GetAllSubscribes called", zap.String("func", "GetAllSubscribes"))

	url := fmt.Sprintf("%s/subscribe/list", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create list request", zap.String("func", "GetAllSubscribes"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to create request")
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send list request", zap.String("func", "GetAllSubscribes"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status", zap.String("func", "GetAllSubscribes"), zap.Int("status_code", resp.StatusCode))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var subscribes []SubscribeData
	if err := json.NewDecoder(resp.Body).Decode(&subscribes); err != nil {
		logger.Error("Failed to decode response", zap.String("func", "GetAllSubscribes"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("All subscribes retrieved successfully", zap.String("func", "GetAllSubscribes"), zap.Int("count", len(subscribes)))
	return subscribes, nil
}

// GetSubscribeStatistic 获取订阅统计
func (sh *SubscribeHelper) GetSubscribeStatistic() (*SubscribeStatistic, error) {
	logger.Debug("SubscribeHelper.GetSubscribeStatistic called", zap.String("func", "GetSubscribeStatistic"))

	url := fmt.Sprintf("%s/subscribe/statistic", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create statistic request", zap.String("func", "GetSubscribeStatistic"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to create request")
	}

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send statistic request", zap.String("func", "GetSubscribeStatistic"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status", zap.String("func", "GetSubscribeStatistic"), zap.Int("status_code", resp.StatusCode))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var statistic SubscribeStatistic
	if err := json.NewDecoder(resp.Body).Decode(&statistic); err != nil {
		logger.Error("Failed to decode response", zap.String("func", "GetSubscribeStatistic"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("Subscribe statistic retrieved successfully", 
		zap.String("func", "GetSubscribeStatistic"),
		zap.Int("total_count", statistic.TotalCount),
		zap.Int("movie_count", statistic.MovieCount),
		zap.Int("tv_count", statistic.TVCount))

	return &statistic, nil
}

// ShareSubscribe 分享订阅
func (sh *SubscribeHelper) ShareSubscribe(subscribeIDs []string, name, description string, public bool) (*SubscribeShare, error) {
	logger.Debug("SubscribeHelper.ShareSubscribe called", zap.String("func", "ShareSubscribe"), zap.Strings("subscribe_ids", subscribeIDs), zap.String("name", name), zap.Bool("public", public))

	if len(subscribeIDs) == 0 {
		logger.Error("Subscribe IDs cannot be empty", zap.String("func", "ShareSubscribe"))
		return nil, errors.NewAppError(http.StatusBadRequest, "subscribe IDs cannot be empty", "")
	}

	if name == "" {
		logger.Error("Share name cannot be empty", zap.String("func", "ShareSubscribe"))
		return nil, errors.NewAppError(http.StatusBadRequest, "share name cannot be empty", "")
	}

	logger.Info("Creating share", zap.String("func", "ShareSubscribe"), zap.String("name", name), zap.Int("subscribes_count", len(subscribeIDs)), zap.Bool("public", public))

	shareData := map[string]interface{}{
		"subscribe_ids": subscribeIDs,
		"name":          name,
		"description":   description,
		"public":        public,
	}

	// 将数据编码为JSON
	jsonData, err := json.Marshal(shareData)
	if err != nil {
		logger.Error("Failed to encode share data", zap.String("func", "ShareSubscribe"), zap.String("error", err.Error()), zap.String("name", name))
		return nil, errors.WrapError(err, "failed to encode share data")
	}

	url := fmt.Sprintf("%s/subscribe/share", sh.serverHost)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create share request", zap.String("func", "ShareSubscribe"), zap.String("error", err.Error()), zap.String("name", name))
		return nil, errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send share request", zap.String("func", "ShareSubscribe"), zap.String("error", err.Error()), zap.String("name", name))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status for share", zap.String("func", "ShareSubscribe"), zap.Int("status_code", resp.StatusCode), zap.String("name", name))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		logger.Error("Failed to decode share response", zap.String("func", "ShareSubscribe"), zap.String("error", err.Error()), zap.String("name", name))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("Share created successfully", zap.String("func", "ShareSubscribe"), zap.String("share_id", share.ID), zap.String("share_name", share.Name))
	return &share, nil
}

// GetShare 获取分享
func (sh *SubscribeHelper) GetShare(shareID string) (*SubscribeShare, error) {
	logger.Debug("SubscribeHelper.GetShare called", zap.String("func", "GetShare"), zap.String("share_id", shareID))

	if shareID == "" {
		logger.Error("Share ID cannot be empty", zap.String("func", "GetShare"))
		return nil, errors.NewAppError(http.StatusBadRequest, "share ID cannot be empty", "")
	}

	logger.Info("Getting share", zap.String("func", "GetShare"), zap.String("share_id", shareID))

	url := fmt.Sprintf("%s/subscribe/share/%s", sh.serverHost, shareID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create get share request", zap.String("func", "GetShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send get share request", zap.String("func", "GetShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status for get share", zap.String("func", "GetShare"), zap.Int("status_code", resp.StatusCode), zap.String("share_id", shareID))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		logger.Error("Failed to decode get share response", zap.String("func", "GetShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("Share retrieved successfully", zap.String("func", "GetShare"), zap.String("share_id", shareID), zap.String("share_name", share.Name))
	return &share, nil
}

// GetAllShares 获取所有分享
func (sh *SubscribeHelper) GetAllShares() ([]SubscribeShare, error) {
	logger.Debug("SubscribeHelper.GetAllShares called", zap.String("func", "GetAllShares"))

	url := fmt.Sprintf("%s/subscribe/shares", sh.serverHost)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create get all shares request", zap.String("func", "GetAllShares"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send get all shares request", zap.String("func", "GetAllShares"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status for get all shares", zap.String("func", "GetAllShares"), zap.Int("status_code", resp.StatusCode))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var shares []SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&shares); err != nil {
		logger.Error("Failed to decode get all shares response", zap.String("func", "GetAllShares"), zap.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("All shares retrieved successfully", zap.String("func", "GetAllShares"), zap.Int("count", len(shares)))
	return shares, nil
}

// ForkShare 复刻分享
func (sh *SubscribeHelper) ForkShare(shareID string) (*SubscribeShare, error) {
	logger.Debug("SubscribeHelper.ForkShare called", zap.String("func", "ForkShare"), zap.String("share_id", shareID))

	if shareID == "" {
		logger.Error("Share ID cannot be empty", zap.String("func", "ForkShare"))
		return nil, errors.NewAppError(http.StatusBadRequest, "share ID cannot be empty", "")
	}

	logger.Info("Forking share", zap.String("func", "ForkShare"), zap.String("share_id", shareID))

	url := fmt.Sprintf("%s/subscribe/fork/%s", sh.serverHost, shareID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Error("Failed to create fork share request", zap.String("func", "ForkShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send fork share request", zap.String("func", "ForkShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status for fork share", zap.String("func", "ForkShare"), zap.Int("status_code", resp.StatusCode), zap.String("share_id", shareID))
		return nil, errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var share SubscribeShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		logger.Error("Failed to decode fork share response", zap.String("func", "ForkShare"), zap.String("error", err.Error()), zap.String("share_id", shareID))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	logger.Info("Share forked successfully", zap.String("func", "ForkShare"), zap.String("share_id", shareID), zap.String("forked_share_id", share.ID))
	return &share, nil
}

// ReportInstall 报告安装
func (sh *SubscribeHelper) ReportInstall() error {
	logger.Debug("SubscribeHelper.ReportInstall called", zap.String("func", "ReportInstall"))

	url := fmt.Sprintf("%s/subscribe/report", sh.serverHost)
	
	reportData := map[string]interface{}{
		"action":    "install",
		"timestamp": time.Now().Unix(),
		"user_agent": "MoviePilot-SubscribeHelper/1.0",
	}

	logger.Info("Reporting install", zap.String("func", "ReportInstall"), zap.Int64("timestamp", reportData["timestamp"].(int64)))
	
	err := sh.sendRequest("POST", url, reportData)
	if err != nil {
		logger.Error("Failed to report install", zap.String("func", "ReportInstall"), zap.String("error", err.Error()))
		return err
	}

	logger.Info("Install reported successfully", zap.String("func", "ReportInstall"))
	return nil
}

// ReportStatistic 报告统计
func (sh *SubscribeHelper) ReportStatistic(statistic *SubscribeStatistic) error {
	logger.Debug("SubscribeHelper.ReportStatistic called", zap.String("func", "ReportStatistic"), zap.Bool("has_statistic", statistic != nil))

	if statistic == nil {
		logger.Error("Statistic cannot be nil", zap.String("func", "ReportStatistic"))
		return errors.NewAppError(http.StatusBadRequest, "statistic cannot be nil", "")
	}

	logger.Info("Reporting statistic", zap.String("func", "ReportStatistic"), zap.Int("total_count", statistic.TotalCount))

	url := fmt.Sprintf("%s/subscribe/statistic", sh.serverHost)
	
	err := sh.sendRequest("POST", url, statistic)
	if err != nil {
		logger.Error("Failed to report statistic", zap.String("func", "ReportStatistic"), zap.String("error", err.Error()))
		return err
	}

	logger.Info("Statistic reported successfully", zap.String("func", "ReportStatistic"))
	return nil
}

// sendRequest 发送HTTP请求
func (sh *SubscribeHelper) sendRequest(method, url string, data interface{}) error {
	logger.Debug("SubscribeHelper.sendRequest called", zap.String("func", "sendRequest"), zap.String("method", method), zap.String("url", url), zap.Bool("has_data", data != nil))

	var err error
	var req *http.Request

	if data != nil {
		// 将数据编码为JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			logger.Error("Failed to encode request data", zap.String("func", "sendRequest"), zap.String("error", err.Error()), zap.String("method", method), zap.String("url", url))
			return errors.WrapError(err, "failed to encode request data")
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		logger.Error("Failed to create request", zap.String("func", "sendRequest"), zap.String("error", err.Error()), zap.String("method", method), zap.String("url", url))
		return errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-SubscribeHelper/1.0")

	logger.Debug("Sending HTTP request", zap.String("func", "sendRequest"), zap.String("method", method), zap.String("url", url), zap.String("user_agent", req.Header.Get("User-Agent")))

	resp, err := sh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send request", zap.String("func", "sendRequest"), zap.String("error", err.Error()), zap.String("method", method), zap.String("url", url))
		return errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	logger.Debug("Received HTTP response", zap.String("func", "sendRequest"), zap.String("method", method), zap.String("url", url), zap.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error status", zap.String("func", "sendRequest"), zap.Int("status_code", resp.StatusCode), zap.String("method", method), zap.String("url", url))
		return errors.NewAppError(resp.StatusCode, "server returned error status", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	logger.Info("HTTP request completed successfully", zap.String("func", "sendRequest"), zap.String("method", method), zap.String("url", url), zap.Int("status_code", resp.StatusCode))
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
	logger.Debug("SubscribeHelper.ValidateSubscribe called", zap.String("func", "ValidateSubscribe"), zap.Bool("has_subscribe", subscribe != nil), zap.String("subscribe_name", func() string { if subscribe != nil { return subscribe.Name }; return "" }()))

	if subscribe == nil {
		logger.Error("Subscribe cannot be nil", zap.String("func", "ValidateSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe cannot be nil", "")
	}

	if subscribe.Name == "" {
		logger.Error("Subscribe name cannot be empty", zap.String("func", "ValidateSubscribe"))
		return errors.NewAppError(http.StatusBadRequest, "subscribe name cannot be empty", "")
	}

	if subscribe.Type == "" {
		logger.Error("Subscribe type cannot be empty", zap.String("func", "ValidateSubscribe"), zap.String("subscribe_name", subscribe.Name))
		return errors.NewAppError(http.StatusBadRequest, "subscribe type cannot be empty", "")
	}

	// 验证类型
	validTypes := []string{"movie", "tv", "documentary", "animation"}
	if !containsString(validTypes, subscribe.Type) {
		logger.Error("Invalid subscribe type", zap.String("func", "ValidateSubscribe"), zap.String("type", subscribe.Type), zap.Strings("valid_types", validTypes), zap.String("subscribe_name", subscribe.Name))
		return errors.NewAppError(http.StatusBadRequest, "invalid subscribe type", fmt.Sprintf("type: %s", subscribe.Type))
	}

	// 验证状态
	validStatuses := []string{"pending", "downloading", "completed", "failed", "paused"}
	if subscribe.Status != "" && !containsString(validStatuses, subscribe.Status) {
		logger.Error("Invalid subscribe status", zap.String("func", "ValidateSubscribe"), zap.String("status", subscribe.Status), zap.Strings("valid_statuses", validStatuses), zap.String("subscribe_name", subscribe.Name))
		return errors.NewAppError(http.StatusBadRequest, "invalid subscribe status", fmt.Sprintf("status: %s", subscribe.Status))
	}

	logger.Debug("Subscribe validation passed", zap.String("func", "ValidateSubscribe"), zap.String("subscribe_name", subscribe.Name), zap.String("type", subscribe.Type), zap.String("status", subscribe.Status))
	return nil
}

// ValidateShare 验证分享数据
func (sh *SubscribeHelper) ValidateShare(share *SubscribeShare) error {
	logger.Debug("SubscribeHelper.ValidateShare called", zap.String("func", "ValidateShare"), zap.Bool("has_share", share != nil), zap.String("share_name", func() string { if share != nil { return share.Name }; return "" }()))

	if share == nil {
		logger.Error("Share cannot be nil", zap.String("func", "ValidateShare"))
		return errors.NewAppError(http.StatusBadRequest, "share cannot be nil", "")
	}

	if share.Name == "" {
		logger.Error("Share name cannot be empty", zap.String("func", "ValidateShare"))
		return errors.NewAppError(http.StatusBadRequest, "share name cannot be empty", "")
	}

	if share.Author == "" {
		logger.Error("Share author cannot be empty", zap.String("func", "ValidateShare"), zap.String("share_name", share.Name))
		return errors.NewAppError(http.StatusBadRequest, "share author cannot be empty", "")
	}

	// 验证订阅数据
	for i, subscribe := range share.Subscribes {
		if err := sh.ValidateSubscribe(&subscribe); err != nil {
			logger.Error("Invalid subscribe in share", zap.String("func", "ValidateShare"), zap.String("error", err.Error()), zap.String("share_name", share.Name), zap.Int("subscribe_index", i))
			return errors.WrapError(err, fmt.Sprintf("invalid subscribe in share at index %d", i))
		}
	}

	logger.Debug("Share validation passed", zap.String("func", "ValidateShare"), zap.String("share_name", share.Name), zap.String("author", share.Author), zap.Int("subscribes_count", len(share.Subscribes)))
	return nil
}

// ExportSubscribes 导出订阅数据
func (sh *SubscribeHelper) ExportSubscribes() ([]SubscribeData, error) {
	return sh.GetAllSubscribes()
}

// ImportSubscribes 导入订阅数据
func (sh *SubscribeHelper) ImportSubscribes(subscribes []SubscribeData) error {
	logger.Debug("SubscribeHelper.ImportSubscribes called", zap.String("func", "ImportSubscribes"), zap.Int("subscribes_count", len(subscribes)))

	if subscribes == nil {
		logger.Error("Subscribes cannot be nil", zap.String("func", "ImportSubscribes"))
		return errors.NewAppError(http.StatusBadRequest, "subscribes cannot be nil", "")
	}

	successCount := 0
	for i, subscribe := range subscribes {
		if err := sh.ValidateSubscribe(&subscribe); err != nil {
			logger.Error("Invalid subscribe during import", zap.String("func", "ImportSubscribes"), zap.String("error", err.Error()), zap.String("subscribe_name", subscribe.Name), zap.Int("index", i))
			return errors.WrapError(err, fmt.Sprintf("invalid subscribe at index %d: %s", i, subscribe.Name))
		}

		if err := sh.AddSubscribe(&subscribe); err != nil {
			logger.Error("Failed to add subscribe during import", zap.String("func", "ImportSubscribes"), zap.String("error", err.Error()), zap.String("subscribe_name", subscribe.Name), zap.Int("index", i))
			return errors.WrapError(err, fmt.Sprintf("failed to add subscribe %s", subscribe.Name))
		}
		successCount++
	}

	logger.Info("Subscribes imported successfully", zap.String("func", "ImportSubscribes"), zap.Int("total_count", len(subscribes)), zap.Int("success_count", successCount))
	return nil
}

// GetStats 获取统计信息
func (sh *SubscribeHelper) GetStats() map[string]interface{} {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	stats := map[string]interface{}{
		"server_host":   sh.serverHost,
		"admin_users":   len(sh.adminUsers),
		"share_user_id": sh.shareUserID,
		"github_user":   sh.githubUser,
		"timeout":       sh.httpClient.Timeout.String(),
	}

	logger.Debug("SubscribeHelper stats retrieved", zap.String("func", "GetStats"), zap.Any("stats", stats))
	return stats
}

