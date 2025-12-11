package downloader

import (
	"context"
	"time"
)

// Client 下载器客户端接口
type Client interface {
	// AddTorrent 添加种子
	AddTorrent(ctx context.Context, req *AddTorrentRequest) (*Torrent, error)

	// ListTorrents 列出所有种子
	ListTorrents(ctx context.Context, filter *TorrentFilter) ([]*Torrent, error)

	// GetTorrentInfo 获取种子详情
	GetTorrentInfo(ctx context.Context, hash string) (*TorrentInfo, error)

	// PauseTorrent 暂停种子
	PauseTorrent(ctx context.Context, hash string) error

	// ResumeTorrent 恢复种子
	ResumeTorrent(ctx context.Context, hash string) error

	// RemoveTorrent 删除种子
	RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error

	// SetTorrentCategory 设置种子分类
	SetTorrentCategory(ctx context.Context, hash string, category string) error

	// SetTorrentTags 设置种子标签
	SetTorrentTags(ctx context.Context, hash string, tags []string) error

	// GetVersion 获取下载器版本
	GetVersion(ctx context.Context) (string, error)

	// TestConnection 测试连接
	TestConnection(ctx context.Context) error
}

// AddTorrentRequest 添加种子请求
type AddTorrentRequest struct {
	// URL 种子URL或磁力链接
	URL string

	// TorrentData 种子文件数据（二进制）
	TorrentData []byte

	// SavePath 保存路径
	SavePath string

	// Category 分类
	Category string

	// Tags 标签
	Tags []string

	// Paused 是否暂停
	Paused bool

	// SkipChecking 跳过校验
	SkipChecking bool

	// SequentialDownload 顺序下载
	SequentialDownload bool

	// FirstLastPiecePrio 优先下载首尾块
	FirstLastPiecePrio bool
}

// TorrentFilter 种子过滤器
type TorrentFilter struct {
	// Category 分类过滤
	Category string

	// Tag 标签过滤
	Tag string

	// State 状态过滤
	State TorrentState

	// Hashes 指定hash列表
	Hashes []string
}

// Torrent 种子基本信息
type Torrent struct {
	// Hash 种子哈希
	Hash string

	// Name 种子名称
	Name string

	// State 状态
	State TorrentState

	// Progress 进度 (0-1)
	Progress float64

	// Size 总大小（字节）
	Size int64

	// Downloaded 已下载（字节）
	Downloaded int64

	// Uploaded 已上传（字节）
	Uploaded int64

	// DownloadSpeed 下载速度（字节/秒）
	DownloadSpeed int64

	// UploadSpeed 上传速度（字节/秒）
	UploadSpeed int64

	// ETA 预计剩余时间（秒）
	ETA int64

	// Ratio 分享率
	Ratio float64

	// Category 分类
	Category string

	// Tags 标签
	Tags []string

	// SavePath 保存路径
	SavePath string

	// AddedOn 添加时间
	AddedOn time.Time

	// CompletionOn 完成时间
	CompletionOn *time.Time
}

// TorrentInfo 种子详细信息
type TorrentInfo struct {
	*Torrent

	// Seeders 做种者数量
	Seeders int

	// Leechers 下载者数量
	Leechers int

	// TotalSize 总大小
	TotalSize int64

	// PieceSize 分块大小
	PieceSize int64

	// NumPieces 分块数量
	NumPieces int

	// Comment 备注
	Comment string

	// CreationDate 创建日期
	CreationDate time.Time

	// Creator 创建者
	Creator string

	// Files 文件列表
	Files []*TorrentFile
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	// Name 文件名
	Name string

	// Size 文件大小
	Size int64

	// Progress 下载进度
	Progress float64

	// Priority 优先级
	Priority int

	// IsSeed 是否完成
	IsSeed bool
}

// TorrentState 种子状态
type TorrentState string

const (
	// StateError 错误
	StateError TorrentState = "error"

	// StateMissingFiles 文件丢失
	StateMissingFiles TorrentState = "missingFiles"

	// StateUploading 上传中
	StateUploading TorrentState = "uploading"

	// StatePausedUP 暂停上传
	StatePausedUP TorrentState = "pausedUP"

	// StateQueuedUP 排队上传
	StateQueuedUP TorrentState = "queuedUP"

	// StateStalledUP 停滞上传
	StateStalledUP TorrentState = "stalledUP"

	// StateCheckingUP 检查上传
	StateCheckingUP TorrentState = "checkingUP"

	// StateForcedUP 强制上传
	StateForcedUP TorrentState = "forcedUP"

	// StateAllocating 分配空间
	StateAllocating TorrentState = "allocating"

	// StateDownloading 下载中
	StateDownloading TorrentState = "downloading"

	// StateMetaDL 下载元数据
	StateMetaDL TorrentState = "metaDL"

	// StatePausedDL 暂停下载
	StatePausedDL TorrentState = "pausedDL"

	// StateQueuedDL 排队下载
	StateQueuedDL TorrentState = "queuedDL"

	// StateStalledDL 停滞下载
	StateStalledDL TorrentState = "stalledDL"

	// StateCheckingDL 检查下载
	StateCheckingDL TorrentState = "checkingDL"

	// StateForcedDL 强制下载
	StateForcedDL TorrentState = "forcedDL"

	// StateCheckingResumeData 检查恢复数据
	StateCheckingResumeData TorrentState = "checkingResumeData"

	// StateMoving 移动中
	StateMoving TorrentState = "moving"

	// StateUnknown 未知状态
	StateUnknown TorrentState = "unknown"
)

// IsDownloading 是否正在下载
func (s TorrentState) IsDownloading() bool {
	return s == StateDownloading || s == StateMetaDL || s == StateForcedDL
}

// IsCompleted 是否已完成
func (s TorrentState) IsCompleted() bool {
	return s == StateUploading || s == StateStalledUP || s == StateCheckingUP || s == StateForcedUP
}

// IsPaused 是否已暂停
func (s TorrentState) IsPaused() bool {
	return s == StatePausedDL || s == StatePausedUP
}

// IsError 是否错误
func (s TorrentState) IsError() bool {
	return s == StateError || s == StateMissingFiles
}

// Factory 下载器工厂
type Factory struct {
	clients map[string]Client
}

// NewFactory 创建下载器工厂
func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]Client),
	}
}

// Register 注册下载器客户端
func (f *Factory) Register(name string, client Client) {
	f.clients[name] = client
}

// GetClient 获取下载器客户端
func (f *Factory) GetClient(name string) (Client, bool) {
	client, ok := f.clients[name]
	return client, ok
}

// ListClients 列出所有下载器
func (f *Factory) ListClients() []string {
	names := make([]string, 0, len(f.clients))
	for name := range f.clients {
		names = append(names, name)
	}
	return names
}
