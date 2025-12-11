package qbittorrent

import (
	"time"

	"moviepilot-go/internal/integration/downloader"
)

// qbTorrentInfo qBittorrent 种子信息
type qbTorrentInfo struct {
	Hash              string  `json:"hash"`
	Name              string  `json:"name"`
	State             string  `json:"state"`
	Progress          float64 `json:"progress"`
	Size              int64   `json:"size"`
	Downloaded        int64   `json:"downloaded"`
	Uploaded          int64   `json:"uploaded"`
	DlSpeed           int64   `json:"dlspeed"`
	UpSpeed           int64   `json:"upspeed"`
	ETA               int64   `json:"eta"`
	Ratio             float64 `json:"ratio"`
	Category          string  `json:"category"`
	Tags              string  `json:"tags"`
	SavePath          string  `json:"save_path"`
	AddedOn           int64   `json:"added_on"`
	CompletionOn      int64   `json:"completion_on"`
	NumSeeds          int     `json:"num_seeds"`
	NumLeechs         int     `json:"num_leechs"`
	TotalSize         int64   `json:"total_size"`
	AmountLeft        int64   `json:"amount_left"`
	TimeActive        int64   `json:"time_active"`
	SeedsConnected    int     `json:"num_complete"`
	LeechersConnected int     `json:"num_incomplete"`
}

// toTorrent 转换为通用种子信息
func (qt *qbTorrentInfo) toTorrent() *downloader.Torrent {
	torrent := &downloader.Torrent{
		Hash:          qt.Hash,
		Name:          qt.Name,
		State:         qt.mapState(),
		Progress:      qt.Progress,
		Size:          qt.Size,
		Downloaded:    qt.Downloaded,
		Uploaded:      qt.Uploaded,
		DownloadSpeed: qt.DlSpeed,
		UploadSpeed:   qt.UpSpeed,
		ETA:           qt.ETA,
		Ratio:         qt.Ratio,
		Category:      qt.Category,
		SavePath:      qt.SavePath,
		AddedOn:       time.Unix(qt.AddedOn, 0),
	}

	// 解析标签
	if qt.Tags != "" {
		torrent.Tags = splitTags(qt.Tags)
	}

	// 完成时间
	if qt.CompletionOn > 0 {
		completionTime := time.Unix(qt.CompletionOn, 0)
		torrent.CompletionOn = &completionTime
	}

	return torrent
}

// mapState 映射状态
func (qt *qbTorrentInfo) mapState() downloader.TorrentState {
	switch qt.State {
	case "error":
		return downloader.StateError
	case "missingFiles":
		return downloader.StateMissingFiles
	case "uploading":
		return downloader.StateUploading
	case "pausedUP":
		return downloader.StatePausedUP
	case "queuedUP":
		return downloader.StateQueuedUP
	case "stalledUP":
		return downloader.StateStalledUP
	case "checkingUP":
		return downloader.StateCheckingUP
	case "forcedUP":
		return downloader.StateForcedUP
	case "allocating":
		return downloader.StateAllocating
	case "downloading":
		return downloader.StateDownloading
	case "metaDL":
		return downloader.StateMetaDL
	case "pausedDL":
		return downloader.StatePausedDL
	case "queuedDL":
		return downloader.StateQueuedDL
	case "stalledDL":
		return downloader.StateStalledDL
	case "checkingDL":
		return downloader.StateCheckingDL
	case "forcedDL":
		return downloader.StateForcedDL
	case "checkingResumeData":
		return downloader.StateCheckingResumeData
	case "moving":
		return downloader.StateMoving
	default:
		return downloader.StateUnknown
	}
}

// qbTorrentFile qBittorrent 种子文件
type qbTorrentFile struct {
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
	Priority int     `json:"priority"`
	IsSeed   bool    `json:"is_seed"`
}

// toTorrentFile 转换为通用文件信息
func (qf *qbTorrentFile) toTorrentFile() *downloader.TorrentFile {
	return &downloader.TorrentFile{
		Name:     qf.Name,
		Size:     qf.Size,
		Progress: qf.Progress,
		Priority: qf.Priority,
		IsSeed:   qf.IsSeed,
	}
}

// qbTracker qBittorrent tracker 信息
type qbTracker struct {
	URL        string `json:"url"`
	Status     int    `json:"status"`
	NumSeeds   int    `json:"num_seeds"`
	NumLeeches int    `json:"num_leeches"`
	NumPeers   int    `json:"num_peers"`
	Msg        string `json:"msg"`
}

// splitTags 分割标签
func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}

	var result []string
	for _, tag := range splitString(tags, ",") {
		if trimmed := trimSpace(tag); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trimSpace 去除空格
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
