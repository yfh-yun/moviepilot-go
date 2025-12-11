package transmission

import (
	"encoding/json"
	"fmt"
	"time"

	"moviepilot-go/internal/integration/downloader"
)

// rpcRequest RPC 请求
type rpcRequest struct {
	Method    string `json:"method"`
	Arguments any    `json:"arguments,omitempty"`
	Tag       int    `json:"tag,omitempty"`
}

// rpcResponse RPC 响应
type rpcResponse struct {
	Arguments json.RawMessage `json:"arguments"`
	Result    string          `json:"result"`
	Tag       int             `json:"tag,omitempty"`
}

// trTorrent Transmission 种子信息
type trTorrent struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Status             int      `json:"status"`
	PercentDone        float64  `json:"percentDone"`
	TotalSize          int64    `json:"totalSize"`
	DownloadedEver     int64    `json:"downloadedEver"`
	UploadedEver       int64    `json:"uploadedEver"`
	RateDownload       int64    `json:"rateDownload"`
	RateUpload         int64    `json:"rateUpload"`
	ETA                int64    `json:"eta"`
	UploadRatio        float64  `json:"uploadRatio"`
	Labels             []string `json:"labels"`
	DownloadDir        string   `json:"downloadDir"`
	AddedDate          int64    `json:"addedDate"`
	DoneDate           int64    `json:"doneDate"`
	PeersConnected     int      `json:"peersConnected"`
	PeersGettingFromUs int      `json:"peersGettingFromUs"`
	PeersSendingToUs   int      `json:"peersSendingToUs"`
	Error              int      `json:"error"`
	ErrorString        string   `json:"errorString"`

	// 详细信息字段
	Files        []trFile        `json:"files"`
	FileStats    []trFileStat    `json:"fileStats"`
	Trackers     []trTracker     `json:"trackers"`
	TrackerStats []trTrackerStat `json:"trackerStats"`
	PieceCount   int             `json:"pieceCount"`
	PieceSize    int64           `json:"pieceSize"`
	Comment      string          `json:"comment"`
	Creator      string          `json:"creator"`
	DateCreated  int64           `json:"dateCreated"`
	HashString   string          `json:"hashString"`
}

// trFile Transmission 文件信息
type trFile struct {
	Name           string `json:"name"`
	Length         int64  `json:"length"`
	BytesCompleted int64  `json:"bytesCompleted"`
}

// trFileStat Transmission 文件统计
type trFileStat struct {
	BytesCompleted int64 `json:"bytesCompleted"`
	Wanted         bool  `json:"wanted"`
	Priority       int   `json:"priority"`
}

// trTracker Transmission tracker 信息
type trTracker struct {
	ID       int    `json:"id"`
	Announce string `json:"announce"`
	Scrape   string `json:"scrape"`
	Tier     int    `json:"tier"`
}

// trTrackerStat Transmission tracker 统计
type trTrackerStat struct {
	ID                    int    `json:"id"`
	Announce              string `json:"announce"`
	AnnounceState         int    `json:"announceState"`
	DownloadCount         int    `json:"downloadCount"`
	HasAnnounced          bool   `json:"hasAnnounced"`
	HasScraped            bool   `json:"hasScraped"`
	Host                  string `json:"host"`
	IsBackup              bool   `json:"isBackup"`
	LastAnnouncePeerCount int    `json:"lastAnnouncePeerCount"`
	LastAnnounceResult    string `json:"lastAnnounceResult"`
	LastAnnounceStartTime int64  `json:"lastAnnounceStartTime"`
	LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded"`
	LastAnnounceTime      int64  `json:"lastAnnounceTime"`
	LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut"`
	LastScrapeResult      string `json:"lastScrapeResult"`
	LastScrapeStartTime   int64  `json:"lastScrapeStartTime"`
	LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded"`
	LastScrapeTime        int64  `json:"lastScrapeTime"`
	LastScrapeTimedOut    bool   `json:"lastScrapeTimedOut"`
	LeecherCount          int    `json:"leecherCount"`
	NextAnnounceTime      int64  `json:"nextAnnounceTime"`
	NextScrapeTime        int64  `json:"nextScrapeTime"`
	Scrape                string `json:"scrape"`
	ScrapeState           int    `json:"scrapeState"`
	SeederCount           int    `json:"seederCount"`
	Tier                  int    `json:"tier"`
}

// toTorrent 转换为通用种子信息
func (tr *trTorrent) toTorrent() *downloader.Torrent {
	torrent := &downloader.Torrent{
		Hash:          fmt.Sprintf("%d", tr.ID), // Transmission 使用 ID
		Name:          tr.Name,
		State:         tr.mapState(),
		Progress:      tr.PercentDone,
		Size:          tr.TotalSize,
		Downloaded:    tr.DownloadedEver,
		Uploaded:      tr.UploadedEver,
		DownloadSpeed: tr.RateDownload,
		UploadSpeed:   tr.RateUpload,
		ETA:           tr.ETA,
		Ratio:         tr.UploadRatio,
		Tags:          tr.Labels,
		SavePath:      tr.DownloadDir,
		AddedOn:       time.Unix(tr.AddedDate, 0),
	}

	// 分类（从标签中提取第一个作为分类）
	if len(tr.Labels) > 0 {
		torrent.Category = tr.Labels[0]
	}

	// 完成时间
	if tr.DoneDate > 0 {
		doneTime := time.Unix(tr.DoneDate, 0)
		torrent.CompletionOn = &doneTime
	}

	return torrent
}

// mapState 映射状态
// Transmission 状态码:
// 0: 已停止
// 1: 检查等待
// 2: 检查中
// 3: 下载等待
// 4: 下载中
// 5: 做种等待
// 6: 做种中
func (tr *trTorrent) mapState() downloader.TorrentState {
	// 检查错误
	if tr.Error != 0 {
		return downloader.StateError
	}

	switch tr.Status {
	case 0: // 已停止
		if tr.PercentDone >= 1.0 {
			return downloader.StatePausedUP
		}
		return downloader.StatePausedDL

	case 1: // 检查等待
		return downloader.StateCheckingResumeData

	case 2: // 检查中
		if tr.PercentDone >= 1.0 {
			return downloader.StateCheckingUP
		}
		return downloader.StateCheckingDL

	case 3: // 下载等待
		return downloader.StateQueuedDL

	case 4: // 下载中
		// 检查是否停滞
		if tr.RateDownload == 0 && tr.PeersSendingToUs == 0 {
			return downloader.StateStalledDL
		}
		return downloader.StateDownloading

	case 5: // 做种等待
		return downloader.StateQueuedUP

	case 6: // 做种中
		// 检查是否停滞
		if tr.RateUpload == 0 && tr.PeersGettingFromUs == 0 {
			return downloader.StateStalledUP
		}
		return downloader.StateUploading

	default:
		return downloader.StateUnknown
	}
}
