package torrents

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// TorrentsService 种子服务
// 原TorrentsChain，负责站点首页或RSS种子处理
type TorrentsService struct {
	*base.ServiceBase
	logger     *zap.Logger
	httpClient *http.Client
}

// NewTorrentsService 创建TorrentsService实例
func NewTorrentsService() *TorrentsService {
	return &TorrentsService{
		ServiceBase: base.NewServiceBase(),
		logger:      logger.GetLogger(),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Initialize 初始化服务
func (s *TorrentsService) Initialize() error {
	s.logger.Info("初始化TorrentsService")
	return nil
}

// Name 获取服务名称
func (s *TorrentsService) Name() string {
	return "TorrentsService"
}

// Close 关闭服务
func (s *TorrentsService) Close() error {
	s.logger.Info("关闭TorrentsService")
	return nil
}

// ParseTorrent 解析种子
func (s *TorrentsService) ParseTorrent(ctx context.Context, torrentURL string) (*dto.TorrentInfo, error) {
	s.logger.Info("解析种子", zap.String("url", torrentURL))

	// 下载种子文件
	torrentData, err := s.downloadTorrentFile(ctx, torrentURL)
	if err != nil {
		s.logger.Error("下载种子文件失败", zap.Error(err))
		return nil, fmt.Errorf("下载种子文件失败: %w", err)
	}

	// 解析种子文件
	mi, err := metainfo.Load(bytes.NewReader(torrentData))
	if err != nil {
		s.logger.Error("解析种子文件失败", zap.Error(err))
		return nil, fmt.Errorf("解析种子文件失败: %w", err)
	}

	// 获取种子信息
	info, err := mi.UnmarshalInfo()
	if err != nil {
		s.logger.Error("解析种子信息失败", zap.Error(err))
		return nil, fmt.Errorf("解析种子信息失败: %w", err)
	}

	// 计算种子哈希
	hash := mi.HashInfoBytes()

	// 提取种子信息
	torrentInfo := &dto.TorrentInfo{
		Title:       info.Name,
		Description: mi.Comment,
		Size:        float64(info.TotalLength()),
	}

	// 记录额外信息（用于日志）
	extraInfo := map[string]any{
		"hash":        hash.HexString(),
		"file_count":  len(info.Files),
		"created_by":  mi.CreatedBy,
		"create_date": time.Unix(mi.CreationDate, 0),
	}

	s.logger.Info("种子解析成功",
		zap.String("title", torrentInfo.Title),
		zap.Any("extra_info", extraInfo),
	)

	return torrentInfo, nil
}

// GetRSSTorrents 获取RSS种子
func (s *TorrentsService) GetRSSTorrents(ctx context.Context, siteID int) ([]*dto.Context, error) {
	s.logger.Info("获取RSS种子", zap.Int("site_id", siteID))

	// TODO: 实现获取RSS种子逻辑
	// 1. 获取站点配置
	// 2. 发送RSS请求
	// 3. 解析RSS响应
	// 4. 构建种子上下文

	// 这里简化处理，返回空切片
	return []*dto.Context{}, nil
}

// FilterTorrents 过滤种子
func (s *TorrentsService) FilterTorrents(ctx context.Context, torrents []*dto.Context, filterRule string) ([]*dto.Context, error) {
	s.logger.Info("过滤种子",
		zap.Int("torrent_count", len(torrents)),
		zap.String("filter_rule", filterRule),
	)

	// 1. 解析过滤规则
	// TODO: 实现规则解析逻辑
	// 这里简化处理，假设规则已经解析完成
	// 实际实现中需要解析filterRule，生成过滤条件

	// 2. 对每个种子应用过滤规则
	// TODO: 实现过滤逻辑
	// 这里简化处理，假设所有种子都符合条件
	// 实际实现中需要根据过滤条件判断种子是否符合要求
	// 例如，根据大小、标题、种子数、下载数等条件过滤
	filteredTorrents := append([]*dto.Context{}, torrents...)

	s.logger.Info("种子过滤完成",
		zap.Int("original_count", len(torrents)),
		zap.Int("filtered_count", len(filteredTorrents)),
	)

	// 3. 返回符合条件的种子
	return filteredTorrents, nil
}

// downloadTorrentFile 下载种子文件
func (s *TorrentsService) downloadTorrentFile(ctx context.Context, url string) ([]byte, error) {
	// 发送HTTP请求
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
