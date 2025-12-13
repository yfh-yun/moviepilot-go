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

// VolumeFactor 根据上传/下载因子计算促销类型，与Python版本get_free_string功能一致
func (t *TorrentInfo) VolumeFactor() string {
	freeStrs := map[string]string{
		"1.0 1.0": "普通",
		"1.0 0.0": "免费",
		"2.0 1.0": "2X",
		"4.0 1.0": "4X",
		"2.0 0.0": "2X免费",
		"4.0 0.0": "4X免费",
		"1.0 0.5": "50%",
		"2.0 0.5": "2X 50%",
		"1.0 0.7": "70%",
		"1.0 0.3": "30%",
	}

	key := fmt.Sprintf("%.1f %.1f", t.UploadVolumeFactor, t.DownloadVolumeFactor)
	return freeStrs[key]
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

// ToDict 将TorrentInfo转换为字典，与Python版本to_dict功能一致
func (t *TorrentInfo) ToDict() map[string]interface{} {
	result := make(map[string]interface{})

	// 站点相关字段
	result["site"] = t.SiteID
	result["site_name"] = t.SiteName
	result["site_cookie"] = t.SiteCookie
	result["site_ua"] = t.SiteUserAgent
	result["site_proxy"] = t.SiteProxy
	result["site_order"] = t.SiteOrder
	result["site_downloader"] = t.SiteDownloader

	// 资源内容字段
	result["title"] = t.Title
	result["description"] = t.Description
	result["imdbid"] = t.IMDBID
	result["enclosure"] = t.Enclosure
	result["page_url"] = t.PageURL
	result["size"] = t.Size

	// 实时状态字段
	result["seeders"] = t.Seeders
	result["peers"] = t.Peers
	result["grabs"] = t.Grabs
	result["pubdate"] = t.PubDate.Format("2006-01-02 15:04:05")

	// 促销/规则字段
	if t.FreeUntil != nil {
		result["freedate"] = t.FreeUntil.Format("2006-01-02 15:04:05")
	}
	result["uploadvolumefactor"] = t.UploadVolumeFactor
	result["downloadvolumefactor"] = t.DownloadVolumeFactor
	result["hit_and_run"] = t.HitAndRun
	result["labels"] = t.Labels
	result["pri_order"] = t.Priority
	result["category"] = t.Category

	// 计算属性
	result["volume_factor"] = t.VolumeFactor()
	result["freedate_diff"] = t.FreeDateDiff(time.Now())

	return result
}
