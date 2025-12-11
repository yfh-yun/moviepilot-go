package media

import (
	"fmt"
	"time"
)

// TorrentInfo 种子上下文（站点 + 资源 + 促销信息）
type TorrentInfo struct {
	// 站点相关
	SiteID         int    `json:"site_id"`
	SiteName       string `json:"site_name"`
	SiteCookie     string `json:"site_cookie"`
	SiteUserAgent  string `json:"site_ua"`
	SiteProxy      bool   `json:"site_proxy"`
	SiteOrder      int    `json:"site_order"`
	SiteDownloader string `json:"site_downloader"`

	// 资源内容
	Title       string  `json:"title"`
	Description string  `json:"description"`
	IMDBID      string  `json:"imdbid"`
	Enclosure   string  `json:"enclosure"`
	PageURL     string  `json:"page_url"`
	Size        float64 `json:"size"`

	// 实时状态
	Seeders int       `json:"seeders"`
	Peers   int       `json:"peers"`
	Grabs   int       `json:"grabs"`
	PubDate time.Time `json:"pubdate"` // 解析后的发布时间

	// 促销/规则
	FreeUntil            *time.Time `json:"freedate"`
	UploadVolumeFactor   float64    `json:"uploadvolumefactor"`
	DownloadVolumeFactor float64    `json:"downloadvolumefactor"`
	HitAndRun            bool       `json:"hit_and_run"`
	Labels               []string   `json:"labels"`
	Priority             int        `json:"pri_order"`
	Category             string     `json:"category"`
}

// VolumeFactor 根据上传/下载因子计算促销类型
func (t *TorrentInfo) VolumeFactor() string {
	if t.UploadVolumeFactor == 0 {
		return ""
	}

	var factorStr string

	if t.UploadVolumeFactor > 1 {
		factorStr = fmt.Sprintf("%.0fX ", t.UploadVolumeFactor)
	}

	if t.DownloadVolumeFactor < 1 {
		factorStr += fmt.Sprintf("%.0f%%", t.DownloadVolumeFactor*100)
	}

	return factorStr
}

// FreeDateDiff 计算距离免费结束还剩多久
func (t *TorrentInfo) FreeDateDiff(now time.Time) string {
	if t.FreeUntil == nil {
		return ""
	}

	if now.After(*t.FreeUntil) {
		return "已结束"
	}

	diff := t.FreeUntil.Sub(now)
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}

	return fmt.Sprintf("%d分钟", minutes)
}

// PubMinutes 发布时间距当前的分钟数
func (t *TorrentInfo) PubMinutes(now time.Time) int64 {
	return int64(now.Sub(t.PubDate).Minutes())
}
