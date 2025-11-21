package utils

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
)

// TorrentHelper 种子文件辅助工具
type TorrentHelper struct {
	client *torrent.Client
}

// NewTorrentHelper 创建种子辅助工具实例
func NewTorrentHelper() (*TorrentHelper, error) {
	client, err := torrent.NewClient(&torrent.Config{
		DefaultStorage: storage.NewFile("/tmp/torrents"),
		NoUpload:       false,
		Seed:           true,
	})
	if err != nil {
		return nil, fmt.Errorf("创建种子客户端失败: %w", err)
	}

	return &TorrentHelper{
		client: client,
	}, nil
}

// NewTorrentHelperWithConfig 使用配置创建种子辅助工具
func NewTorrentHelperWithConfig(config *torrent.Config) (*TorrentHelper, error) {
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建种子客户端失败: %w", err)
	}

	return &TorrentHelper{
		client: client,
	}, nil
}

// GetClient 获取种子客户端
func (t *TorrentHelper) GetClient() *torrent.Client {
	return t.client
}

// Close 关闭客户端
func (t *TorrentHelper) Close() error {
	if t.client != nil {
		t.client.Close()
	}
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

// ParseTorrentFile 解析种子文件
func (t *TorrentHelper) ParseTorrentFile(filePath string) (*TorrentInfo, error) {
	metaInfo, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("加载种子文件失败: %w", err)
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return nil, fmt.Errorf("解析种子信息失败: %w", err)
	}

	torrentInfo := &TorrentInfo{
		InfoHash:     metaInfo.HashInfoBytes().String(),
		PieceLength:  int(info.PieceLength),
		PieceCount:   info.NumPieces(),
		Trackers:     metaInfo.AnnounceList,
		CreationDate: time.Unix(metaInfo.CreationDate, 0),
		CreatedBy:    metaInfo.CreatedBy,
		Comment:      metaInfo.Comment,
		Encoding:     string(info.Encoding),
		IsPrivate:    metaInfo.Private,
		Metadata:     make(map[string]string),
	}

	// 解析文件信息
	if info.UpDir {
		// 单文件种子
		torrentInfo.Name = info.Name
		torrentInfo.Size = info.Length
		torrentInfo.Files = []TorrentFile{
			{
				Path:   info.Name,
				Name:   info.Name,
				Size:   info.Length,
				Offset: 0,
			},
		}
	} else {
		// 多文件种子
		torrentInfo.Name = info.Name
		var totalSize int64
		var offset int64

		for _, file := range info.Files {
			filePath := strings.Join(file.Path, string(filepath.Separator))
			fileSize := int64(file.Length)
			
			torrentFile := TorrentFile{
				Path:   filePath,
				Name:   filepath.Base(filePath),
				Size:   fileSize,
				Offset: offset,
			}

			// 计算文件哈希（如果可用）
			if len(file.Md5sum) > 0 {
				torrentFile.MD5Sum = hex.EncodeToString(file.Md5sum)
			}
			if len(file.Ed2k) > 0 {
				torrentFile.ED2KHash = hex.EncodeToString(file.Ed2k)
			}
			if len(file.Sha1) > 0 {
				torrentFile.SHA1Hash = hex.EncodeToString(file.Sha1)
			}

			torrentInfo.Files = append(torrentInfo.Files, torrentFile)
			totalSize += fileSize
			offset += fileSize
		}

		torrentInfo.Size = totalSize
	}

	return torrentInfo, nil
}

// ParseTorrentFromURL 从URL解析种子文件
func (t *TorrentHelper) ParseTorrentFromURL(torrentURL string) (*TorrentInfo, error) {
	resp, err := http.Get(torrentURL)
	if err != nil {
		return nil, fmt.Errorf("下载种子文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载种子文件失败，状态码: %d", resp.StatusCode)
	}

	// 保存到临时文件
	tempFile := filepath.Join(os.TempDir(), "temp_"+strconv.FormatInt(time.Now().Unix(), 10)+".torrent")
	defer os.Remove(tempFile)

	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("保存种子文件失败: %w", err)
	}

	return t.ParseTorrentFile(tempFile)
}

// ParseMagnetURI 解析磁力链接
func (t *TorrentHelper) ParseMagnetURI(magnetURI string) (*TorrentInfo, error) {
	parsed, err := url.Parse(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("解析磁力链接失败: %w", err)
	}

	if parsed.Scheme != "magnet" {
		return nil, fmt.Errorf("不是有效的磁力链接")
	}

	query := parsed.Query()
	
	torrentInfo := &TorrentInfo{
		Trackers: query["tr"],
		Metadata: make(map[string]string),
	}

	// 解析info hash
	if xt := query.Get("xt"); xt != "" {
		if strings.HasPrefix(xt, "urn:btih:") {
			torrentInfo.InfoHash = strings.TrimPrefix(xt, "urn:btih:")
		}
	}

	// 解析名称
	if dn := query.Get("dn"); dn != "" {
		torrentInfo.Name = dn
	}

	// 解析其他参数
	for key, values := range query {
		if len(values) > 0 {
			torrentInfo.Metadata[key] = values[0]
		}
	}

	return torrentInfo, nil
}

// CreateTorrentFile 创建种子文件
func (t *TorrentHelper) CreateTorrentFile(inputPath, outputPath, trackerURL, comment string, isPrivate bool) error {
	var info metainfo.Info

	// 检查输入路径
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	if fileInfo.IsDir() {
		// 创建目录种子
		err = info.BuildFromFilePath(inputPath)
		if err != nil {
			return fmt.Errorf("构建目录种子信息失败: %w", err)
		}
	} else {
		// 创建单文件种子
		info.Name = filepath.Base(inputPath)
		info.Length = fileInfo.Size()
		
		file, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("打开文件失败: %w", err)
		}
		defer file.Close()

		pieces := make([]byte, 0)
		hash := sha1.New()
		buffer := make([]byte, 16*1024) // 16KB pieces

		for {
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取文件失败: %w", err)
			}
			if n == 0 {
				break
			}

			hash.Reset()
			hash.Write(buffer[:n])
			pieces = append(pieces, hash.Sum(nil)...)
		}

		info.Pieces = pieces
		info.PieceLength = 16 * 1024
	}

	// 创建元信息
	metaInfo := metainfo.MetaInfo{
		InfoBytes: info.Marshal(),
		Announce:  trackerURL,
		Comment:   comment,
		CreatedBy: "MoviePilot Go",
		Private:   isPrivate,
	}

	// 保存种子文件
	return metaInfo.WriteFile(outputPath)
}

// AddTorrent 添加种子
func (t *TorrentHelper) AddTorrent(torrentPath string) (*torrent.Torrent, error) {
	torrent, err := t.client.AddTorrentFromFile(torrentPath)
	if err != nil {
		return nil, fmt.Errorf("添加种子失败: %w", err)
	}

	return torrent, nil
}

// AddMagnet 添加磁力链接
func (t *TorrentHelper) AddMagnet(magnetURI string) (*torrent.Torrent, error) {
	torrent, err := t.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("添加磁力链接失败: %w", err)
	}

	return torrent, nil
}

// GetTorrentStats 获取种子统计信息
func (t *TorrentHelper) GetTorrentStats(torrent *torrent.Torrent) *TorrentStats {
	stats := torrent.Stats()
	
	torrentStats := &TorrentStats{
		TotalSize:      torrent.Length(),
		DownloadedSize: stats.BytesReadData.Int64(),
		UploadedSize:   stats.BytesWrittenData.Int64(),
		Progress:       torrent.BytesCompleted() / float64(torrent.Length()),
		DownloadSpeed:  int64(torrent.Stats().TotalDataBytesReadThisSession),
		UploadSpeed:    int64(torrent.Stats().TotalDataBytesWrittenThisSession),
		Seeds:          int(torrent.Stats().ConnectedSeeders),
		Peers:          int(torrent.Stats().ActivePeers),
		State:          t.getTorrentState(torrent),
	}

	// 计算ETA
	if torrentStats.DownloadSpeed > 0 {
		remaining := torrentStats.TotalSize - torrentStats.DownloadedSize
		torrentStats.ETA = remaining / torrentStats.DownloadSpeed
	}

	return torrentStats
}

// getTorrentState 获取种子状态
func (t *TorrentHelper) getTorrentState(torrent *torrent.Torrent) string {
	switch {
	case torrent.Seeding():
		return "seeding"
	case torrent.Downloading():
		return "downloading"
	case torrent.Checking():
		return "checking"
	case torrent.Stopped():
		return "stopped"
	default:
		return "unknown"
	}
}

// StartTorrent 开始下载种子
func (t *TorrentHelper) StartTorrent(torrent *torrent.Torrent) error {
	torrent.DownloadAll()
	return nil
}

// StopTorrent 停止下载种子
func (t *TorrentHelper) StopTorrent(torrent *torrent.Torrent) error {
	torrent.Drop()
	return nil
}

// PauseTorrent 暂停下载种子
func (t *TorrentHelper) PauseTorrent(torrent *torrent.Torrent) error {
	// torrent库没有直接的暂停功能，可以通过停止来实现
	return t.StopTorrent(torrent)
}

// ResumeTorrent 恢复下载种子
func (t *TorrentHelper) ResumeTorrent(torrent *torrent.Torrent) error {
	return t.StartTorrent(torrent)
}

// RemoveTorrent 移除种子
func (t *TorrentHelper) RemoveTorrent(torrent *torrent.Torrent, deleteFiles bool) error {
	torrent.Drop()
	
	if deleteFiles {
		// 删除文件
		for _, file := range torrent.Files() {
			if file.Path() != "" {
				os.Remove(file.Path())
			}
		}
	}
	
	return nil
}

// VerifyTorrent 验证种子文件
func (t *TorrentHelper) VerifyTorrent(torrent *torrent.Torrent) error {
	// 开始验证
	torrent.VerifyData()
	return nil
}

// GetTrackers 获取种子跟踪器
func (t *TorrentHelper) GetTrackers(torrent *torrent.Torrent) []string {
	return torrent.Trackers()
}

// AddTrackers 添加跟踪器
func (t *TorrentHelper) AddTrackers(torrent *torrent.Torrent, trackers []string) error {
	return torrent.AddTrackers(trackers)
}

// RemoveTrackers 移除跟踪器
func (t *TorrentHelper) RemoveTrackers(torrent *torrent.Torrent, trackers []string) error {
	// torrent库没有直接的移除跟踪器功能
	// 需要重新添加所有跟踪器（不包括要移除的）
	currentTrackers := t.GetTrackers(torrent)
	var newTrackers []string
	
	for _, tracker := range currentTrackers {
		shouldRemove := false
		for _, removeTracker := range trackers {
			if tracker == removeTracker {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			newTrackers = append(newTrackers, tracker)
		}
	}
	
	return t.AddTrackers(torrent, newTrackers)
}

// GetPeers 获取连接的对等节点
func (t *TorrentHelper) GetPeers(torrent *torrent.Torrent) []torrent.Peer {
	return torrent.PeerConns()
}

// GetFiles 获取种子文件列表
func (t *TorrentHelper) GetFiles(torrent *torrent.Torrent) []torrent.File {
	return torrent.Files()
}

// DownloadFile 下载指定文件
func (t *TorrentHelper) DownloadFile(torrent *torrent.Torrent, filePath string) error {
	for _, file := range torrent.Files() {
		if file.Path() == filePath {
			file.Download()
			return nil
		}
	}
	return fmt.Errorf("未找到文件: %s", filePath)
}

// SetPriority 设置文件优先级
func (t *TorrentHelper) SetPriority(torrent *torrent.Torrent, filePath string, priority int) error {
	for _, file := range torrent.Files() {
		if file.Path() == filePath {
			file.SetPriority(priority)
			return nil
		}
	}
	return fmt.Errorf("未找到文件: %s", filePath)
}

// GetDownloadPath 获取下载路径
func (t *TorrentHelper) GetDownloadPath(torrent *torrent.Torrent) string {
	return torrent.Storage().String()
}

// SetDownloadPath 设置下载路径
func (t *TorrentHelper) SetDownloadPath(torrent *torrent.Torrent, path string) error {
	// torrent库不支持动态更改下载路径
	// 需要在创建客户端时指定
	return fmt.Errorf("不支持动态更改下载路径")
}

// GetMagnetURI 获取磁力链接
func (t *TorrentHelper) GetMagnetURI(torrent *torrent.Torrent) string {
	return torrent.Magnet().String()
}

// ExportTorrent 导出种子文件
func (t *TorrentHelper) ExportTorrent(torrent *torrent.Torrent, outputPath string) error {
	metaInfo := torrent.Metainfo()
	return metaInfo.WriteFile(outputPath)
}

// ImportTorrent 导入种子文件
func (t *TorrentHelper) ImportTorrent(torrentPath string) (*torrent.Torrent, error) {
	return t.AddTorrent(torrentPath)
}

// SearchTorrents 搜索种子（需要外部搜索API）
func (t *TorrentHelper) SearchTorrents(keyword string, category string) ([]map[string]interface{}, error) {
	// 这里需要实现具体的搜索逻辑
	// 可以调用外部的搜索API，如The Pirate Bay、1337x等
	return nil, fmt.Errorf("搜索功能需要实现具体的搜索API")
}

// GetTorrentList 获取种子列表
func (t *TorrentHelper) GetTorrentList() []*torrent.Torrent {
	var torrents []*torrent.Torrent
	for _, torrent := range t.client.Torrents() {
		torrents = append(torrents, torrent)
	}
	return torrents
}

// GetTorrentByHash 根据哈希值获取种子
func (t *TorrentHelper) GetTorrentByHash(infoHash string) (*torrent.Torrent, error) {
	for _, torrent := range t.client.Torrents() {
		if torrent.InfoHash().String() == infoHash {
			return torrent, nil
		}
	}
	return nil, fmt.Errorf("未找到哈希为 %s 的种子", infoHash)
}

// GetTorrentByName 根据名称获取种子
func (t *TorrentHelper) GetTorrentByName(name string) (*torrent.Torrent, error) {
	for _, torrent := range t.client.Torrents() {
		if torrent.Name() == name {
			return torrent, nil
		}
	}
	return nil, fmt.Errorf("未找到名称为 %s 的种子", name)
}

// CalculateInfoHash 计算信息哈希
func (t *TorrentHelper) CalculateInfoHash(data []byte) (string, error) {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

// ValidateTorrentFile 验证种子文件
func (t *TorrentHelper) ValidateTorrentFile(filePath string) error {
	_, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("种子文件无效: %w", err)
	}
	return nil
}

// ValidateMagnetURI 验证磁力链接
func (t *TorrentHelper) ValidateMagnetURI(magnetURI string) error {
	_, err := t.ParseMagnetURI(magnetURI)
	if err != nil {
		return fmt.Errorf("磁力链接无效: %w", err)
	}
	return nil
}

// GetTorrentProgress 获取下载进度
func (t *TorrentHelper) GetTorrentProgress(torrent *torrent.Torrent) float64 {
	if torrent.Length() == 0 {
		return 0
	}
	return float64(torrent.BytesCompleted()) / float64(torrent.Length())
}

// IsTorrentComplete 检查种子是否下载完成
func (t *TorrentHelper) IsTorrentComplete(torrent *torrent.Torrent) bool {
	return torrent.BytesCompleted() == torrent.Length()
}

// GetRemainingTime 获取剩余时间
func (t *TorrentHelper) GetRemainingTime(torrent *torrent.Torrent) time.Duration {
	stats := t.GetTorrentStats(torrent)
	if stats.DownloadSpeed <= 0 {
		return 0
	}
	
	remaining := torrent.Length() - torrent.BytesCompleted()
	return time.Duration(remaining/stats.DownloadSpeed) * time.Second
}

// GetDownloadSpeed 获取下载速度
func (t *TorrentHelper) GetDownloadSpeed(torrent *torrent.Torrent) int64 {
	return int64(torrent.Stats().TotalDataBytesReadThisSession)
}

// GetUploadSpeed 获取上传速度
func (t *TorrentHelper) GetUploadSpeed(torrent *torrent.Torrent) int64 {
	return int64(torrent.Stats().TotalDataBytesWrittenThisSession)
}

// GetSeedRatio 获取分享率
func (t *TorrentHelper) GetSeedRatio(torrent *torrent.Torrent) float64 {
	stats := t.GetTorrentStats(torrent)
	if stats.DownloadedSize == 0 {
		return 0
	}
	return float64(stats.UploadedSize) / float64(stats.DownloadedSize)
}

// SetSeedRatio 设置分享率限制
func (t *TorrentHelper) SetSeedRatio(torrent *torrent.Torrent, ratio float64) error {
	// torrent库没有直接的分享率限制功能
	// 需要手动监控和停止
	return fmt.Errorf("分享率限制需要手动实现")
}

// GetActivePeers 获取活跃对等节点数量
func (t *TorrentHelper) GetActivePeers(torrent *torrent.Torrent) int {
	return int(torrent.Stats().ActivePeers)
}

// GetConnectedSeeds 获取连接的种子数量
func (t *TorrentHelper) GetConnectedSeeds(torrent *torrent.Torrent) int {
	return int(torrent.Stats().ConnectedSeeders)
}

// GetTotalPeers 获取总对等节点数量
func (t *TorrentHelper) GetTotalPeers(torrent *torrent.Torrent) int {
	return len(torrent.PeerConns())
}

// GetAvailableSeeds 获取可用种子数量
func (t *TorrentHelper) GetAvailableSeeds(torrent *torrent.Torrent) int {
	return int(torrent.Stats().KnownSeeders)
}

// GetAvailability 获取可用性
func (t *TorrentHelper) GetAvailability(torrent *torrent.Torrent) float64 {
	peers := t.GetTotalPeers(torrent)
	seeds := t.GetConnectedSeeds(torrent)
	
	if peers == 0 {
		return float64(seeds)
	}
	
	return float64(seeds) / float64(peers)
}

// GetPieceStatus 获取分片状态
func (t *TorrentHelper) GetPieceStatus(torrent *torrent.Torrent) ([]bool, error) {
	pieces := torrent.Pieces()
	status := make([]bool, len(pieces))
	
	for i, piece := range pieces {
		status[i] = piece.Complete()
	}
	
	return status, nil
}

// GetCompletedPieces 获取已完成分片数量
func (t *TorrentHelper) GetCompletedPieces(torrent *torrent.Torrent) int {
	pieces := torrent.Pieces()
	completed := 0
	
	for _, piece := range pieces {
		if piece.Complete() {
			completed++
		}
	}
	
	return completed
}

// GetTotalPieces 获取总分片数量
func (t *TorrentHelper) GetTotalPieces(torrent *torrent.Torrent) int {
	return len(torrent.Pieces())
}

// GetPieceSize 获取分片大小
func (t *TorrentHelper) GetPieceSize(torrent *torrent.Torrent) int {
	if len(torrent.Pieces()) > 0 {
		return len(torrent.Pieces()[0].Data())
	}
	return 0
}

// GetTorrentInfo 获取种子详细信息
func (t *TorrentHelper) GetTorrentInfo(torrent *torrent.Torrent) (*TorrentInfo, error) {
	metaInfo := torrent.Metainfo()
	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return nil, fmt.Errorf("解析种子信息失败: %w", err)
	}

	torrentInfo := &TorrentInfo{
		Name:         info.Name,
		InfoHash:     torrent.InfoHash().String(),
		Size:         torrent.Length(),
		PieceLength:  int(info.PieceLength),
		PieceCount:   info.NumPieces(),
		Trackers:     metaInfo.AnnounceList,
		CreationDate: time.Unix(metaInfo.CreationDate, 0),
		CreatedBy:    metaInfo.CreatedBy,
		Comment:      metaInfo.Comment,
		Encoding:     string(info.Encoding),
		IsPrivate:    metaInfo.Private,
		Metadata:     make(map[string]string),
	}

	// 解析文件信息
	files := t.GetFiles(torrent)
	for _, file := range files {
		torrentFile := TorrentFile{
			Path:   file.Path(),
			Name:   file.DisplayPath(),
			Size:   file.Length(),
			Offset: file.Offset(),
		}
		torrentInfo.Files = append(torrentInfo.Files, torrentFile)
	}

	return torrentInfo, nil
}

// ConvertToMagnet 转换为磁力链接
func (t *TorrentHelper) ConvertToMagnet(torrentPath string) (string, error) {
	torrent, err := t.AddTorrent(torrentPath)
	if err != nil {
		return "", fmt.Errorf("添加种子失败: %w", err)
	}
	defer t.RemoveTorrent(torrent, false)

	return t.GetMagnetURI(torrent), nil
}

// ConvertToTorrent 转换为种子文件
func (t *TorrentHelper) ConvertToTorrent(magnetURI, outputPath string) error {
	torrent, err := t.AddMagnet(magnetURI)
	if err != nil {
		return fmt.Errorf("添加磁力链接失败: %w", err)
	}

	// 等待元数据下载完成
	<-torrent.GotInfo()

	return t.ExportTorrent(torrent, outputPath)
}

// BatchOperation 批量操作
func (t *TorrentHelper) BatchOperation(torrents []*torrent.Torrent, operation string) error {
	for _, torrent := range torrents {
		switch operation {
		case "start":
			if err := t.StartTorrent(torrent); err != nil {
				return fmt.Errorf("启动种子失败: %w", err)
			}
		case "stop":
			if err := t.StopTorrent(torrent); err != nil {
				return fmt.Errorf("停止种子失败: %w", err)
			}
		case "pause":
			if err := t.PauseTorrent(torrent); err != nil {
				return fmt.Errorf("暂停种子失败: %w", err)
			}
		case "resume":
			if err := t.ResumeTorrent(torrent); err != nil {
				return fmt.Errorf("恢复种子失败: %w", err)
			}
		case "verify":
			if err := t.VerifyTorrent(torrent); err != nil {
				return fmt.Errorf("验证种子失败: %w", err)
			}
		default:
			return fmt.Errorf("不支持的操作: %s", operation)
		}
	}
	return nil
}

// GetGlobalStats 获取全局统计信息
func (t *TorrentHelper) GetGlobalStats() map[string]interface{} {
	stats := t.client.Stats()
	
	return map[string]interface{}{
		"total_torrents":    len(t.client.Torrents()),
		"active_torrents":  stats.ActiveTorrents,
		"total_peers":      stats.TotalPeers,
		"connected_peers":   stats.ConnectedPeers,
		"half_open_peers":   stats.HalfOpenPeers,
		"total_downloaded": stats.BytesReadData.Int64(),
		"total_uploaded":   stats.BytesWrittenData.Int64(),
		"download_speed":   stats.TotalDataBytesReadThisSession,
		"upload_speed":     stats.TotalDataBytesWrittenThisSession,
	}
}

// SetGlobalLimits 设置全局限制
func (t *TorrentHelper) SetGlobalLimits(maxDownloads, maxUploads int, downloadSpeed, uploadSpeed int64) error {
	// torrent库的配置需要在创建客户端时设置
	// 这里只是示例，实际实现可能需要重新创建客户端
	return fmt.Errorf("全局限制需要在创建客户端时设置")
}

// GetTorrentLimits 获取种子限制
func (t *TorrentHelper) GetTorrentLimits(torrent *torrent.Torrent) map[string]interface{} {
	// torrent库没有直接的限制查询功能
	return map[string]interface{}{
		"max_connections": 0,
		"max_uploads":     0,
		"download_speed": 0,
		"upload_speed":   0,
	}
}

// SetTorrentLimits 设置种子限制
func (t *TorrentHelper) SetTorrentLimits(torrent *torrent.Torrent, maxConnections, maxUploads int, downloadSpeed, uploadSpeed int64) error {
	// torrent库没有直接的限制设置功能
	return fmt.Errorf("种子限制需要手动实现")
}

// MatchSeasonEpisodes 判断种子是否匹配季集数
func (t *TorrentHelper) MatchSeasonEpisodes(torrentInfo *TorrentInfo, meta *model.MetaInfo, seasonEpisodes map[int][]int) bool {
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
	allSeasonsMatch := true
	for _, torrentSeason := range torrentSeasons {
		if !seasons[torrentSeason] {
			// 种子季不在过滤季中
			logger.Global.Debug("种子季不在需要的季中",
				zap.String("siteName", torrentInfo.Name),
				zap.String("title", torrentInfo.Name),
				zap.Ints("torrentSeasons", torrentSeasons),
				zap.Ints("neededSeasons", t.getMapKeys(seasons)))
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
				logger.Global.Debug("种子集没有需要的集",
					zap.String("siteName", torrentInfo.Name),
					zap.String("title", torrentInfo.Name),
					zap.Ints("torrentEpisodes", torrentEpisodes),
					zap.Ints("needEpisodes", needEpisodes))
				return false
			}
		}
	}
	
	return true
}

// ExtractSeasonEpisodes 从种子信息中提取季集信息
func (t *TorrentHelper) ExtractSeasonEpisodes(torrentName string) (*model.MetaInfo, error) {
	meta := &model.MetaInfo{}
	
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
			logger.Global.Warn("提取种子季集信息失败", 
				zap.String("torrentName", torrent.Name), 
				zap.Error(err))
			continue
		}
		
		// 验证季集匹配
		if t.MatchSeasonEpisodes(torrent, meta, seasonEpisodes) {
			filteredTorrents = append(filteredTorrents, torrent)
		}
	}
	
	return filteredTorrents
}

// GetSeasonEpisodesSummary 获取季集匹配摘要
func (t *TorrentHelper) GetSeasonEpisodesSummary(torrents []*TorrentInfo, seasonEpisodes map[int][]int) *SeasonEpisodesSummary {
	summary := &SeasonEpisodesSummary{
		TotalTorrents:    len(torrents),
		MatchedTorrents:  0,
		SeasonsCount:     len(seasonEpisodes),
		SeasonsSummary:   make(map[int]*SeasonSummary),
		RequiredSeasons:   t.getMapKeys(seasonEpisodes),
		RequiredEpisodes:  seasonEpisodes,
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
						MatchedEpisodes: make([]int, 0),
						TorrentCount:    0,
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
	TotalTorrents    int                     `json:"total_torrents"`
	MatchedTorrents  int                     `json:"matched_torrents"`
	SeasonsCount     int                     `json:"seasons_count"`
	SeasonsSummary   map[int]*SeasonSummary   `json:"seasons_summary"`
	RequiredSeasons  []int                   `json:"required_seasons"`
	RequiredEpisodes map[int][]int            `json:"required_episodes"`
}

// SeasonSummary 季摘要
type SeasonSummary struct {
	Season           int   `json:"season"`
	RequiredEpisodes []int `json:"required_episodes"`
	MatchedEpisodes  []int `json:"matched_episodes"`
	TorrentCount     int   `json:"torrent_count"`
	IsComplete       bool  `json:"is_complete"`
}

// getMapKeys 获取map的键列表
func (t *TorrentHelper) getMapKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}