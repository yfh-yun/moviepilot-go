// Package download 下载管理服务
package download

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/integration/qbittorrent"
	"github.com/yfh-yun/moviepilot-go/internal/integration/transmission"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"strings"
	"time"
)

// Service 下载服务
type Service struct {
	downloadRepo      interfaces.DownloadRepository
	downloadFilesRepo interfaces.DownloadFilesRepository
	qbClients         map[string]*qbittorrent.Client
	trClients         map[string]*transmission.Client
}

// NewService 创建下载服务
func NewService(
	downloadRepo interfaces.DownloadRepository,
	downloadFilesRepo interfaces.DownloadFilesRepository,
) *Service {
	return &Service{
		downloadRepo:      downloadRepo,
		downloadFilesRepo: downloadFilesRepo,
		qbClients:         make(map[string]*qbittorrent.Client),
		trClients:         make(map[string]*transmission.Client),
	}
}

// DownloaderConfig 下载器配置
type DownloaderConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // qbittorrent, transmission
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

// TorrentStatus 种子状态
type TorrentStatus struct {
	Hash          string     `json:"hash"`
	Name          string     `json:"name"`
	State         string     `json:"state"`
	Progress      float64    `json:"progress"`
	DownloadSpeed int64      `json:"download_speed"`
	UploadSpeed   int64      `json:"upload_speed"`
	Downloaded    int64      `json:"downloaded"`
	Uploaded      int64      `json:"uploaded"`
	Size          int64      `json:"size"`
	Downloader    string     `json:"downloader"`
	Category      string     `json:"category,omitempty"`
	Tags          string     `json:"tags,omitempty"`
	SavePath      string     `json:"save_path,omitempty"`
	MediaInfo     *MediaInfo `json:"media_info,omitempty"`
}

// MediaInfo 媒体信息(用于前端显示)
type MediaInfo struct {
	TMDBID  *int   `json:"tmdb_id,omitempty"`
	Type    string `json:"type,omitempty"`
	Title   string `json:"title,omitempty"`
	Season  string `json:"season,omitempty"`
	Episode string `json:"episode,omitempty"`
	Image   string `json:"image,omitempty"`
}

// AddTorrentRequest 添加种子请求
type AddTorrentRequest struct {
	TorrentURL     string   `json:"torrent_url,omitempty"`
	TorrentFile    []byte   `json:"torrent_file,omitempty"`
	DownloadDir    string   `json:"download_dir,omitempty"`
	Category       string   `json:"category,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	IsPaused       bool     `json:"is_paused,omitempty"`
	DownloaderName string   `json:"downloader_name"`
	SequentialDL   bool     `json:"sequential_dl,omitempty"`    // qBittorrent专用
	FirstLastPiece bool     `json:"first_last_piece,omitempty"` // qBittorrent专用
}

// InitializeDownloaders 初始化下载器
func (s *Service) InitializeDownloaders(configs []DownloaderConfig) error {
	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		switch strings.ToLower(config.Type) {
		case "qbittorrent":
			client, err := qbittorrent.NewClient(config.Host, config.Port, config.Username, config.Password)
			if err != nil {
				logger.Error("Failed to initialize qBittorrent client", "name", config.Name, "error", err)
				continue
			}
			s.qbClients[config.Name] = client
			logger.Info("qBittorrent client initialized", "name", config.Name)

		case "transmission":
			client, err := transmission.NewClient("http", config.Host, config.Port, config.Username, config.Password)
			if err != nil {
				logger.Error("Failed to initialize Transmission client", "name", config.Name, "error", err)
				continue
			}
			s.trClients[config.Name] = client
			logger.Info("Transmission client initialized", "name", config.Name)

		default:
			logger.Warn("Unknown downloader type", "type", config.Type)
		}
	}

	return nil
}

// AddTorrent 添加种子
func (s *Service) AddTorrent(req *AddTorrentRequest) (string, error) {
	if req.TorrentURL == "" && len(req.TorrentFile) == 0 {
		return "", errors.New("torrent_url or torrent_file is required")
	}

	// 查找下载器
	var hash string
	var err error

	// 先尝试qBittorrent
	if qbClient, ok := s.qbClients[req.DownloaderName]; ok {
		tagsStr := strings.Join(req.Tags, ",")
		err = qbClient.AddTorrent(
			req.TorrentURL,
			req.TorrentFile,
			req.DownloadDir,
			req.Category,
			tagsStr,
			req.IsPaused,
			req.SequentialDL,
			req.FirstLastPiece,
		)
		if err != nil {
			return "", fmt.Errorf("failed to add torrent to qBittorrent: %w", err)
		}

		// 等待种子添加完成并获取hash
		time.Sleep(2 * time.Second)
		torrents, err := qbClient.GetTorrents("", req.Category, strings.Join(req.Tags, ","), "", false, 1, 0)
		if err != nil || len(torrents) == 0 {
			return "", errors.New("failed to get torrent hash")
		}
		hash = torrents[0].Hash

	} else if trClient, ok := s.trClients[req.DownloaderName]; ok {
		// Transmission
		torrent, err := trClient.AddTorrent(req.TorrentURL, req.TorrentFile, req.DownloadDir, req.IsPaused, req.Tags)
		if err != nil {
			return "", fmt.Errorf("failed to add torrent to Transmission: %w", err)
		}
		hash = torrent.HashString

	} else {
		return "", fmt.Errorf("downloader not found: %s", req.DownloaderName)
	}

	logger.Info("Torrent added successfully", "hash", hash, "downloader", req.DownloaderName)
	return hash, nil
}

// GetTorrents 获取种子列表
func (s *Service) GetTorrents(downloaderName string, status string) ([]*TorrentStatus, error) {
	var torrents []*TorrentStatus

	// qBittorrent
	if downloaderName == "" || downloaderName == "qbittorrent" {
		for name, client := range s.qbClients {
			qbTorrents, err := client.GetTorrents(status, "", "", "", false, 0, 0)
			if err != nil {
				logger.Error("Failed to get torrents from qBittorrent", "name", name, "error", err)
				continue
			}

			for _, t := range qbTorrents {
				ts := &TorrentStatus{
					Hash:          t.Hash,
					Name:          t.Name,
					State:         t.State,
					Progress:      t.Progress * 100,
					DownloadSpeed: t.DownloadSpeed,
					UploadSpeed:   t.UploadSpeed,
					Downloaded:    t.Downloaded,
					Uploaded:      t.Uploaded,
					Size:          t.Size,
					Downloader:    name,
					Category:      t.Category,
					Tags:          t.Tags,
					SavePath:      t.SavePath,
				}

				// 从数据库获取媒体信息
				if history, err := s.downloadRepo.GetByHash(t.Hash); err == nil && history != nil {
					ts.MediaInfo = &MediaInfo{
						TMDBID:  history.TMDBID,
						Type:    history.Type,
						Title:   history.Title,
						Season:  history.Seasons,
						Episode: history.Episodes,
						Image:   history.Image,
					}
				}

				torrents = append(torrents, ts)
			}
		}
	}

	// Transmission
	if downloaderName == "" || downloaderName == "transmission" {
		for name, client := range s.trClients {
			trTorrents, err := client.GetTorrents(nil, nil)
			if err != nil {
				logger.Error("Failed to get torrents from Transmission", "name", name, "error", err)
				continue
			}

			for _, t := range trTorrents {
				// 过滤状态
				if status != "" && t.GetStatusString() != status {
					continue
				}

				ts := &TorrentStatus{
					Hash:          t.HashString,
					Name:          t.Name,
					State:         t.GetStatusString(),
					Progress:      t.PercentDone * 100,
					DownloadSpeed: t.RateDownload,
					UploadSpeed:   t.RateUpload,
					Downloaded:    t.DownloadedEver,
					Uploaded:      t.UploadedEver,
					Size:          t.TotalSize,
					Downloader:    name,
					Tags:          strings.Join(t.Labels, ","),
					SavePath:      t.DownloadDir,
				}

				// 从数据库获取媒体信息
				if history, err := s.downloadRepo.GetByHash(t.HashString); err == nil && history != nil {
					ts.MediaInfo = &MediaInfo{
						TMDBID:  history.TMDBID,
						Type:    history.Type,
						Title:   history.Title,
						Season:  history.Seasons,
						Episode: history.Episodes,
						Image:   history.Image,
					}
				}

				torrents = append(torrents, ts)
			}
		}
	}

	return torrents, nil
}

// StartTorrent 启动种子
func (s *Service) StartTorrent(hash string, downloaderName string) error {
	// qBittorrent
	if qbClient, ok := s.qbClients[downloaderName]; ok {
		return qbClient.ResumeTorrents([]string{hash})
	}

	// Transmission - 需要先获取种子ID
	if trClient, ok := s.trClients[downloaderName]; ok {
		torrents, err := trClient.GetTorrents(nil, nil)
		if err != nil {
			return err
		}
		for _, t := range torrents {
			if t.HashString == hash {
				return trClient.StartTorrents([]int64{t.ID})
			}
		}
		return errors.New("torrent not found")
	}

	return fmt.Errorf("downloader not found: %s", downloaderName)
}

// StopTorrent 停止种子
func (s *Service) StopTorrent(hash string, downloaderName string) error {
	// qBittorrent
	if qbClient, ok := s.qbClients[downloaderName]; ok {
		return qbClient.PauseTorrents([]string{hash})
	}

	// Transmission
	if trClient, ok := s.trClients[downloaderName]; ok {
		torrents, err := trClient.GetTorrents(nil, nil)
		if err != nil {
			return err
		}
		for _, t := range torrents {
			if t.HashString == hash {
				return trClient.StopTorrents([]int64{t.ID})
			}
		}
		return errors.New("torrent not found")
	}

	return fmt.Errorf("downloader not found: %s", downloaderName)
}

// DeleteTorrent 删除种子
func (s *Service) DeleteTorrent(hash string, downloaderName string, deleteFiles bool) error {
	// qBittorrent
	if qbClient, ok := s.qbClients[downloaderName]; ok {
		return qbClient.DeleteTorrents([]string{hash}, deleteFiles)
	}

	// Transmission
	if trClient, ok := s.trClients[downloaderName]; ok {
		torrents, err := trClient.GetTorrents(nil, nil)
		if err != nil {
			return err
		}
		for _, t := range torrents {
			if t.HashString == hash {
				return trClient.DeleteTorrents([]int64{t.ID}, deleteFiles)
			}
		}
		return errors.New("torrent not found")
	}

	return fmt.Errorf("downloader not found: %s", downloaderName)
}

// CreateDownloadHistory 创建下载历史
func (s *Service) CreateDownloadHistory(history *models.DownloadHistory) error {
	if history.Date == "" {
		history.Date = time.Now().Format("2006-01-02 15:04:05")
	}
	return s.downloadRepo.Create(history)
}

// GetDownloadHistory 获取下载历史
func (s *Service) GetDownloadHistory(page, pageSize int) ([]*models.DownloadHistory, int64, error) {
	return s.downloadRepo.ListByPage(page, pageSize)
}

// GetDownloadHistoryByHash 根据Hash获取下载历史
func (s *Service) GetDownloadHistoryByHash(hash string) (*models.DownloadHistory, error) {
	return s.downloadRepo.GetByHash(hash)
}

// DeleteDownloadHistory 删除下载历史
func (s *Service) DeleteDownloadHistory(id uint) error {
	return s.downloadRepo.Delete(id)
}

// GetDownloadFiles 获取下载文件列表
func (s *Service) GetDownloadFiles(hash string) ([]*models.DownloadFiles, error) {
	return s.downloadFilesRepo.GetByHash(hash, nil)
}

// AddDownloadFiles 添加下载文件
func (s *Service) AddDownloadFiles(files []*models.DownloadFiles) error {
	for _, file := range files {
		if err := s.downloadFilesRepo.Create(file); err != nil {
			return err
		}
	}
	return nil
}
