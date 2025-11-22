package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	// 暂时注释掉torrent相关依赖，等待Go 1.24兼容性修复
	// "github.com/anacrolix/torrent"
	// "github.com/anacrolix/torrent/metainfo"
	// "github.com/anacrolix/torrent/bencode"
)

// MetaInfo 元信息结构（临时定义，避免循环依赖）
type MetaInfo struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	Type        string `json:"type"` // "movie", "tv"
	Season      int    `json:"season"`
	Episode     int    `json:"episode"`
	SeasonList  []int  `json:"season_list"`
	EpisodeList []int  `json:"episode_list"`
}

// TorrentHelper 种子文件辅助工具（简化版本，等待Go 1.24兼容性修复）
type TorrentHelper struct {
	// client *torrent.Client // 暂时注释
}

// NewTorrentHelper 创建种子辅助工具实例
func NewTorrentHelper() (*TorrentHelper, error) {
	// 暂时返回简化版本，等待torrent库的Go 1.24兼容性修复
	logger.Sugar.Warn("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
	return &TorrentHelper{}, nil
}

// NewTorrentHelperWithConfig 使用配置创建种子辅助工具
func NewTorrentHelperWithConfig(config interface{}) (*TorrentHelper, error) {
	// 暂时返回简化版本
	logger.Sugar.Warn("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
	return &TorrentHelper{}, nil
}

// GetClient 获取种子客户端
func (t *TorrentHelper) GetClient() interface{} {
	logger.Sugar.Warn("Torrent客户端暂时不可用")
	return nil
}

// Close 关闭客户端
func (t *TorrentHelper) Close() error {
	logger.Sugar.Info("关闭Torrent辅助工具")
	return nil
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Name         string            `json:"name"`
	InfoHash     string            `json:"info_hash"`
	Size         int64             `json:"size"`
	PieceLength  int               `json:"piece_length"`
	PieceCount   int               `json:"piece_count"`
	Files        []TorrentFile     `json:"files"`
	Trackers     []string          `json:"trackers"`
	CreationDate time.Time         `json:"creation_date"`
	CreatedBy    string            `json:"created_by"`
	Comment      string            `json:"comment"`
	Encoding     string            `json:"encoding"`
	IsPrivate    bool              `json:"is_private"`
	Metadata     map[string]string `json:"metadata"`
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Offset   int64  `json:"offset"`
	MD5Sum   string `json:"md5sum"`
	ED2KHash string `json:"ed2k_hash"`
	SHA1Hash string `json:"sha1_hash"`
}

// TorrentStats 种子统计信息
type TorrentStats struct {
	TotalSize      int64   `json:"total_size"`
	DownloadedSize int64   `json:"downloaded_size"`
	UploadedSize   int64   `json:"uploaded_size"`
	Progress       float64 `json:"progress"`
	DownloadSpeed  int64   `json:"download_speed"`
	UploadSpeed    int64   `json:"upload_speed"`
	Seeds          int     `json:"seeds"`
	Peers          int     `json:"peers"`
	State          string  `json:"state"`
	ETA            int64   `json:"eta"`
}

// 以下方法暂时返回错误或空值，等待torrent库兼容性修复

func (t *TorrentHelper) ParseTorrentFile(filePath string) (*TorrentInfo, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ParseTorrentFromURL(torrentURL string) (*TorrentInfo, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ParseMagnetURI(magnetURI string) (*TorrentInfo, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) CreateTorrentFile(inputPath, outputPath, trackerURL, comment string, isPrivate bool) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

// 保留不依赖torrent库的辅助方法

// ExtractSeasonEpisodes 从种子信息中提取季集信息
func (t *TorrentHelper) ExtractSeasonEpisodes(torrentName string) (*MetaInfo, error) {
	meta := &MetaInfo{}

	// 定义匹配模式
	patterns := []string{
		// S01E01-E02, S1E1-E2格式
		`^(.*?)[. _\-]+S(\d{1,2})E(\d{1,2})[-:]E(\d{1,2})([. _\-].*)?$`,
		// S01E01, S1E1格式
		`^(.*?)[. _\-]+S(\d{1,2})E(\d{1,2})([. _\-].*)?$`,
		// S01, S1格式（整季）
		`^(.*?)[. _\-]+S(\d{1,2})([. _\-].*)?$`,
		// E01-E02, E1-E2格式
		`^(.*?)[. _\-]+E(\d{1,2})[-:]E(\d{1,2})([. _\-].*)?$`,
		// E01, E1格式（仅集数）
		`^(.*?)[. _\-]+E(\d{1,2})([. _\-].*)?$`,
		// 01x02格式
		`^(.*?)[. _\-]+(\d{1,2})x(\d{1,2})([. _\-].*)?$`,
		// 年份格式
		`^(.*?)[. _\-]+(\d{4})([. _\-].*)?$`,
	}

	// 提取标题并解析季集
	title := regexp.MustCompile(`[._\-]+`).ReplaceAllString(torrentName, " ")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)
	meta.Title = title

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(torrentName)
		if len(matches) > 0 {
			// 提取季信息
			if len(matches) > 2 && matches[2] != "" {
				if season, err := strconv.Atoi(matches[2]); err == nil {
					meta.Season = season
				}
			}

			// 提取集信息
			if len(matches) > 3 && matches[3] != "" {
				if episode, err := strconv.Atoi(matches[3]); err == nil {
					meta.Episode = episode
					meta.EpisodeList = []int{episode}
				}
			}

			// 处理连续集数 E01-E02
			if len(matches) > 4 && matches[4] != "" {
				if endEpisode, err := strconv.Atoi(matches[4]); err == nil && meta.Episode > 0 {
					startEpisode := meta.Episode
					var episodes []int
					for i := startEpisode; i <= endEpisode; i++ {
						episodes = append(episodes, i)
					}
					meta.EpisodeList = episodes
				}
			}

			// 季信息记录
			if meta.Season > 0 {
				meta.SeasonList = []int{meta.Season}
			}

			break
		}
	}

	return meta, nil
}

// ValidateSeasonEpisodes 验证季集匹配配置
func (t *TorrentHelper) ValidateSeasonEpisodes(seasonEpisodes map[int][]int) error {
	if len(seasonEpisodes) == 0 {
		return fmt.Errorf("季集配置不能为空")
	}

	for season, episodes := range seasonEpisodes {
		if season <= 0 {
			return fmt.Errorf("季号必须大于0: %d", season)
		}

		if len(episodes) > 0 {
			for _, episode := range episodes {
				if episode <= 0 {
					return fmt.Errorf("集号必须大于0: %d", episode)
				}
			}
		}
	}

	return nil
}

// FilterTorrentsBySeasonEpisodes 根据季集过滤种子列表
func (t *TorrentHelper) FilterTorrentsBySeasonEpisodes(torrents []*TorrentInfo, seasonEpisodes map[int][]int) []*TorrentInfo {
	var filteredTorrents []*TorrentInfo

	for _, torrent := range torrents {
		// 从种子标题提取元数据
		meta, err := t.ExtractSeasonEpisodes(torrent.Name)
		if err != nil {
			logger.Sugar.Warn("提取种子季集信息失败",
				"torrentName", torrent.Name,
				"error", err)
			continue
		}

		// 验证季集匹配（简化版本）
		if t.MatchSeasonEpisodes(torrent, meta, seasonEpisodes) {
			filteredTorrents = append(filteredTorrents, torrent)
		}
	}

	return filteredTorrents
}

// MatchSeasonEpisodes 判断种子是否匹配季集数（简化版本）
func (t *TorrentHelper) MatchSeasonEpisodes(torrentInfo *TorrentInfo, meta *MetaInfo, seasonEpisodes map[int][]int) bool {
	// 匹配季
	seasons := make(map[int]bool)
	for season := range seasonEpisodes {
		seasons[season] = true
	}

	// 种子季
	torrentSeasons := meta.SeasonList
	if len(torrentSeasons) == 0 {
		// 按第一季处理
		torrentSeasons = []int{1}
	}

	// 种子集
	torrentEpisodes := meta.EpisodeList

	// 检查种子季是否在需要的季中
	torrentSeasonSet := make(map[int]bool)
	for _, season := range torrentSeasons {
		torrentSeasonSet[season] = true
	}

	// 验证季匹配
	for _, torrentSeason := range torrentSeasons {
		if !seasons[torrentSeason] {
			// 种子季不在过滤季中
			logger.Sugar.Debug("种子季不在需要的季中",
				"siteName", torrentInfo.Name,
				"title", torrentInfo.Name,
				"torrentSeasons", torrentSeasons,
				"neededSeasons", t.getMapKeys(seasons))
			return false
		}
	}

	// 如果没有集信息，按整季匹配处理
	if len(torrentEpisodes) == 0 {
		return true
	}

	// 单季处理
	if len(torrentSeasons) == 1 {
		season := torrentSeasons[0]
		needEpisodes := seasonEpisodes[season]

		if len(needEpisodes) > 0 {
			// 检查是否有交集
			hasIntersection := false
			for _, torrentEpisode := range torrentEpisodes {
				for _, needEpisode := range needEpisodes {
					if torrentEpisode == needEpisode {
						hasIntersection = true
						break
					}
				}
				if hasIntersection {
					break
				}
			}

			if !hasIntersection {
				// 单季集没有交集的不要
				logger.Sugar.Debug("种子集没有需要的集",
					"siteName", torrentInfo.Name,
					"title", torrentInfo.Name,
					"torrentEpisodes", torrentEpisodes,
					"needEpisodes", needEpisodes)
				return false
			}
		}
	}

	return true
}

// getMapKeys 获取map的键列表
func (t *TorrentHelper) getMapKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMapKeysFromSeasonEpisodes 从季集map获取键列表
func (t *TorrentHelper) getMapKeysFromSeasonEpisodes(m map[int][]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetSeasonEpisodesSummary 获取季集匹配摘要
func (t *TorrentHelper) GetSeasonEpisodesSummary(torrents []*TorrentInfo, seasonEpisodes map[int][]int) *SeasonEpisodesSummary {
	summary := &SeasonEpisodesSummary{
		TotalTorrents:    len(torrents),
		MatchedTorrents:  0,
		SeasonsCount:     len(seasonEpisodes),
		SeasonsSummary:   make(map[int]*SeasonSummary),
		RequiredSeasons:  t.getMapKeysFromSeasonEpisodes(seasonEpisodes),
		RequiredEpisodes: seasonEpisodes,
	}

	for _, torrent := range torrents {
		// 从种子标题提取元数据
		meta, err := t.ExtractSeasonEpisodes(torrent.Name)
		if err != nil {
			continue
		}

		// 检查是否匹配
		if t.MatchSeasonEpisodes(torrent, meta, seasonEpisodes) {
			summary.MatchedTorrents++

			// 统计季信息
			for _, season := range meta.SeasonList {
				if _, exists := summary.SeasonsSummary[season]; !exists {
					summary.SeasonsSummary[season] = &SeasonSummary{
						Season:           season,
						RequiredEpisodes: seasonEpisodes[season],
						MatchedEpisodes:  make([]int, 0),
						TorrentCount:     0,
					}
				}

				seasonSummary := summary.SeasonsSummary[season]
				seasonSummary.TorrentCount++

				// 添加匹配的集
				for _, episode := range meta.EpisodeList {
					episodeExists := false
					for _, existingEpisode := range seasonSummary.MatchedEpisodes {
						if existingEpisode == episode {
							episodeExists = true
							break
						}
					}
					if !episodeExists {
						seasonSummary.MatchedEpisodes = append(seasonSummary.MatchedEpisodes, episode)
					}
				}
			}
		}
	}

	return summary
}

// SeasonEpisodesSummary 季集匹配摘要
type SeasonEpisodesSummary struct {
	TotalTorrents    int                    `json:"total_torrents"`
	MatchedTorrents  int                    `json:"matched_torrents"`
	SeasonsCount     int                    `json:"seasons_count"`
	SeasonsSummary   map[int]*SeasonSummary `json:"seasons_summary"`
	RequiredSeasons  []int                  `json:"required_seasons"`
	RequiredEpisodes map[int][]int          `json:"required_episodes"`
}

// SeasonSummary 季摘要
type SeasonSummary struct {
	Season           int   `json:"season"`
	RequiredEpisodes []int `json:"required_episodes"`
	MatchedEpisodes  []int `json:"matched_episodes"`
	TorrentCount     int   `json:"torrent_count"`
	IsComplete       bool  `json:"is_complete"`
}

// 其他所有torrent相关方法暂时返回错误
func (t *TorrentHelper) AddTorrent(torrentPath string) (interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) AddMagnet(magnetURI string) (interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTorrentStats(torrent interface{}) *TorrentStats {
	return &TorrentStats{
		State: "unavailable",
	}
}

func (t *TorrentHelper) StartTorrent(torrent interface{}) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) StopTorrent(torrent interface{}) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) PauseTorrent(torrent interface{}) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ResumeTorrent(torrent interface{}) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) RemoveTorrent(torrent interface{}, deleteFiles bool) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) VerifyTorrent(torrent interface{}) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTrackers(torrent interface{}) []string {
	return []string{}
}

func (t *TorrentHelper) AddTrackers(torrent interface{}, trackers []string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) RemoveTrackers(torrent interface{}, trackers []string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetPeers(torrent interface{}) []interface{} {
	return []interface{}{}
}

func (t *TorrentHelper) GetFiles(torrent interface{}) []interface{} {
	return []interface{}{}
}

func (t *TorrentHelper) DownloadFile(torrent interface{}, filePath string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) SetPriority(torrent interface{}, filePath string, priority int) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetDownloadPath(torrent interface{}) string {
	return "/tmp/torrents"
}

func (t *TorrentHelper) SetDownloadPath(torrent interface{}, path string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetMagnetURI(torrent interface{}) string {
	return ""
}

func (t *TorrentHelper) ExportTorrent(torrent interface{}, outputPath string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ImportTorrent(torrentPath string) (interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) SearchTorrents(keyword string, category string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTorrentList() []interface{} {
	return []interface{}{}
}

func (t *TorrentHelper) GetTorrentByHash(infoHash string) (interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTorrentByName(name string) (interface{}, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) CalculateInfoHash(data []byte) (string, error) {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

func (t *TorrentHelper) ValidateTorrentFile(filePath string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ValidateMagnetURI(magnetURI string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTorrentProgress(torrent interface{}) float64 {
	return 0
}

func (t *TorrentHelper) IsTorrentComplete(torrent interface{}) bool {
	return false
}

func (t *TorrentHelper) GetRemainingTime(torrent interface{}) time.Duration {
	return 0
}

func (t *TorrentHelper) GetDownloadSpeed(torrent interface{}) int64 {
	return 0
}

func (t *TorrentHelper) GetUploadSpeed(torrent interface{}) int64 {
	return 0
}

func (t *TorrentHelper) GetSeedRatio(torrent interface{}) float64 {
	return 0
}

func (t *TorrentHelper) SetSeedRatio(torrent interface{}, ratio float64) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetActivePeers(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetConnectedSeeds(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetTotalPeers(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetAvailableSeeds(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetAvailability(torrent interface{}) float64 {
	return 0
}

func (t *TorrentHelper) GetPieceStatus(torrent interface{}) ([]bool, error) {
	return []bool{}, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetCompletedPieces(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetTotalPieces(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetPieceSize(torrent interface{}) int {
	return 0
}

func (t *TorrentHelper) GetTorrentInfo(torrent interface{}) (*TorrentInfo, error) {
	return nil, fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ConvertToMagnet(torrentPath string) (string, error) {
	return "", fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) ConvertToTorrent(magnetURI, outputPath string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) BatchOperation(torrents []interface{}, operation string) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetGlobalStats() map[string]interface{} {
	return map[string]interface{}{
		"status": "unavailable",
	}
}

func (t *TorrentHelper) SetGlobalLimits(maxDownloads, maxUploads int, downloadSpeed, uploadSpeed int64) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}

func (t *TorrentHelper) GetTorrentLimits(torrent interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": "unavailable",
	}
}

func (t *TorrentHelper) SetTorrentLimits(torrent interface{}, maxConnections, maxUploads int, downloadSpeed, uploadSpeed int64) error {
	return fmt.Errorf("Torrent功能暂时不可用，等待Go 1.24兼容性修复")
}
