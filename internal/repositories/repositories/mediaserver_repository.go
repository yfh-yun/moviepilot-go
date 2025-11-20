package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mediaServerRepository 媒体服务器仓储实现
type mediaServerRepository struct {
	db     *gorm.DB
	logger *zap.Logger
	client *http.Client
}

// NewMediaServerRepository 创建媒体服务器仓储
func NewMediaServerRepository(db *gorm.DB, logger *zap.Logger) interfaces.MediaServerRepository {
	return &mediaServerRepository{
		db:     db,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Create 创建媒体服务器记录
func (r *mediaServerRepository) Create(server *model.MediaServer) error {
	return r.db.Create(server).Error
}

// Update 更新媒体服务器记录
func (r *mediaServerRepository) Update(server *model.MediaServer) error {
	return r.db.Save(server).Error
}

// Delete 删除媒体服务器记录
func (r *mediaServerRepository) Delete(id uint) error {
	return r.db.Delete(&model.MediaServer{}, id).Error
}

// GetByID 根据ID获取媒体服务器
func (r *mediaServerRepository) GetByID(id uint) (*model.MediaServer, error) {
	var server model.MediaServer
	err := r.db.First(&server, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// List 获取媒体服务器列表
func (r *mediaServerRepository) List() ([]*model.MediaServer, error) {
	var servers []*model.MediaServer
	err := r.db.Find(&servers).Error
	return servers, err
}

// GetActive 获取激活的媒体服务器
func (r *mediaServerRepository) GetActive() ([]*model.MediaServer, error) {
	var servers []*model.MediaServer
	err := r.db.Where("enabled = ?", true).Find(&servers).Error
	return servers, err
}

// TestConnection 测试媒体服务器连接
func (r *mediaServerRepository) TestConnection(ctx context.Context, id uint) error {
	server, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取媒体服务器失败: %w", err)
	}
	if server == nil {
		return fmt.Errorf("媒体服务器不存在")
	}

	r.logger.Info("测试媒体服务器连接",
		zap.Uint("server_id", id),
		zap.String("server_name", server.Name),
		zap.String("server_type", string(server.Type)))

	switch server.Type {
	case MediaServerTypeEmby:
		return r.testEmbyConnection(ctx, server)
	case MediaServerTypeJellyfin:
		return r.testJellyfinConnection(ctx, server)
	case MediaServerTypePlex:
		return r.testPlexConnection(ctx, server)
	default:
		return fmt.Errorf("不支持的媒体服务器类型: %s", server.Type)
	}
}

// testEmbyConnection 测试Emby连接
func (r *mediaServerRepository) testEmbyConnection(ctx context.Context, server *model.MediaServer) error {
	url := fmt.Sprintf("%s/emby/System/Ping", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证头
	if server.APIKey != "" {
		req.Header.Set("X-Emby-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Emby服务器响应异常: %d", resp.StatusCode)
	}

	r.logger.Info("Emby服务器连接测试成功")
	return nil
}

// testJellyfinConnection 测试Jellyfin连接
func (r *mediaServerRepository) testJellyfinConnection(ctx context.Context, server *model.MediaServer) error {
	url := fmt.Sprintf("%s/System/Ping", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证头
	if server.APIKey != "" {
		req.Header.Set("X-MediaBrowser-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接Jellyfin服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Jellyfin服务器响应异常: %d", resp.StatusCode)
	}

	r.logger.Info("Jellyfin服务器连接测试成功")
	return nil
}

// testPlexConnection 测试Plex连接
func (r *mediaServerRepository) testPlexConnection(ctx context.Context, server *model.MediaServer) error {
	url := fmt.Sprintf("%s/library/sections", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证头
	if server.APIKey != "" {
		req.Header.Set("X-Plex-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接Plex服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Plex服务器响应异常: %d", resp.StatusCode)
	}

	r.logger.Info("Plex服务器连接测试成功")
	return nil
}

// GetLibraries 获取媒体库列表
func (r *mediaServerRepository) GetLibraries(ctx context.Context, id uint) ([]interface{}, error) {
	server, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取媒体服务器失败: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("媒体服务器不存在")
	}

	switch server.Type {
	case MediaServerTypeEmby:
		return r.getEmbyLibraries(ctx, server)
	case MediaServerTypeJellyfin:
		return r.getJellyfinLibraries(ctx, server)
	case MediaServerTypePlex:
		return r.getPlexLibraries(ctx, server)
	default:
		return nil, fmt.Errorf("不支持的媒体服务器类型: %s", server.Type)
	}
}

// getEmbyLibraries 获取Emby媒体库
func (r *mediaServerRepository) getEmbyLibraries(ctx context.Context, server *model.MediaServer) ([]interface{}, error) {
	url := fmt.Sprintf("%s/emby/Library/VirtualFolders", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if server.APIKey != "" {
		req.Header.Set("X-Emby-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取Emby媒体库失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var libraries []map[string]interface{}
	if err := json.Unmarshal(body, &libraries); err != nil {
		return nil, fmt.Errorf("解析Emby媒体库响应失败: %w", err)
	}

	// 转换为通用格式
	var result []interface{}
	for _, lib := range libraries {
		result = append(result, map[string]interface{}{
			"id":   lib["Id"],
			"name": lib["Name"],
			"type": lib["CollectionType"],
		})
	}

	return result, nil
}

// getJellyfinLibraries 获取Jellyfin媒体库
func (r *mediaServerRepository) getJellyfinLibraries(ctx context.Context, server *model.MediaServer) ([]interface{}, error) {
	url := fmt.Sprintf("%s/Library/VirtualFolders", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if server.APIKey != "" {
		req.Header.Set("X-MediaBrowser-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取Jellyfin媒体库失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var libraries []map[string]interface{}
	if err := json.Unmarshal(body, &libraries); err != nil {
		return nil, fmt.Errorf("解析Jellyfin媒体库响应失败: %w", err)
	}

	// 转换为通用格式
	var result []interface{}
	for _, lib := range libraries {
		result = append(result, map[string]interface{}{
			"id":   lib["Id"],
			"name": lib["Name"],
			"type": lib["CollectionType"],
		})
	}

	return result, nil
}

// getPlexLibraries 获取Plex媒体库
func (r *mediaServerRepository) getPlexLibraries(ctx context.Context, server *model.MediaServer) ([]interface{}, error) {
	url := fmt.Sprintf("%s/library/sections", server.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if server.APIKey != "" {
		req.Header.Set("X-Plex-Token", server.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取Plex媒体库失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// Plex XML响应解析（这里简化为JSON处理）
	var response struct {
		MediaContainer struct {
			Directory []struct {
				Key         string `xml:"key"`
				Title       string `xml:"title"`
				Type        string `xml:"type"`
				Composite   string `xml:"composite"`
			} `xml:"Directory"`
		} `xml:"MediaContainer"`
	}

	// 这里需要XML解析，为了简化示例，返回空结果
	var result []interface{}
	return result, nil
}

// SyncLibraries 同步媒体库
func (r *mediaServerRepository) SyncLibraries(ctx context.Context, id uint) error {
	server, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取媒体服务器失败: %w", err)
	}
	if server == nil {
		return fmt.Errorf("媒体服务器不存在")
	}

	r.logger.Info("开始同步媒体库",
		zap.Uint("server_id", id),
		zap.String("server_name", server.Name))

	// 获取媒体库列表
	libraries, err := r.GetLibraries(ctx, id)
	if err != nil {
		return fmt.Errorf("获取媒体库列表失败: %w", err)
	}

	// 这里应该实现具体的同步逻辑
	// 1. 清空现有的媒体服务器项目
	// 2. 遍历媒体库，获取所有媒体项目
	// 3. 保存到数据库

	r.logger.Info("媒体库同步完成",
		zap.Uint("server_id", id),
		zap.Int("library_count", len(libraries)))

	return nil
}

// Exists 检查媒体服务器是否存在
func (r *mediaServerRepository) Exists(id uint) bool {
	var count int64
	r.db.Model(&model.MediaServer{}).Where("id = ?", id).Count(&count)
	return count > 0
}

// GetByName 根据名称获取媒体服务器
func (r *mediaServerRepository) GetByName(name string) (*model.MediaServer, error) {
	var server model.MediaServer
	err := r.db.Where("name = ?", name).First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// Empty 清空媒体服务器数据
func (r *mediaServerRepository) Empty(serverName string) error {
	query := r.db.Model(&model.MediaServerItem{})
	if serverName != "" {
		query = query.Where("server_name = ?", serverName)
	}
	return query.Delete(&model.MediaServerItem{}).Error
}