// Package parsers HDDolby站点解析器
package parsers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/integration/indexer"
	"moviepilot-go/pkg/utils"
)

// HDDolbyParser HDDolby站点解析器
type HDDolbyParser struct {
	schema string
}

// HDDolbyUserLevelMap HDDolby用户级别映射
var HDDolbyUserLevelMap = map[string]string{
	"0":  "Peasant",
	"1":  "User",
	"2":  "Power User",
	"3":  "Elite User",
	"4":  "Crazy User",
	"5":  "Insane User",
	"6":  "Veteran User",
	"7":  "Extreme User",
	"8":  "Ultimate User",
	"9":  "Nexus Master",
	"10": "VIP",
	"11": "Retiree",
	"12": "Helper",
	"13": "Seeder",
	"14": "Transferrer",
	"15": "Uploader",
	"16": "Moderator",
	"17": "Administrator",
	"18": "SysOP",
}

// HDDolbyUserData HDDolby用户数据结构
type HDDolbyUserData struct {
	Status int `json:"status"`
	Data   struct {
		ID       int     `json:"id"`
		Username string  `json:"username"`
		Uploaded int64   `json:"uploaded"`
		Downloaded int64  `json:"downloaded"`
		Bonus    float64 `json:"bonus"`
		UserRole int     `json:"userRole"`
		RegTime  int64   `json:"regTime"`
		Seeding  int     `json:"seeding"`
		Leeching int     `json:"leeching"`
		UploadedCount int `json:"uploadedCount"`
		FinishedCount  int `json:"finishedCount"`
	} `json:"data"`
}

// HDDolbyTorrentData HDDolby种子数据结构
type HDDolbyTorrentData struct {
	Status int `json:"status"`
	Data   struct {
		List []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Size        int64  `json:"size"`
			Added       int64  `json:"added"`
			Seeders     int    `json:"seeders"`
			Leechers    int    `json:"leechers"`
			Completed   int    `json:"completed"`
			Owner       struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
			} `json:"owner"`
			Category struct {
				Name string `json:"name"`
			} `json:"category"`
			Type     string `json:"type"`
			FreeType string `json:"freeType"`
			DoubleUp bool   `json:"doubleUp"`
			IMDBID   string `json:"imdbId"`
			Tags     []string `json:"tags"`
		} `json:"list"`
		Total int `json:"total"`
	} `json:"data"`
}

// NewHDDolbyParser 创建HDDolby解析器
func NewHDDolbyParser() *HDDolbyParser {
	return &HDDolbyParser{
		schema: "hddolby",
	}
}

// GetSchema 获取解析器模式
func (p *HDDolbyParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *HDDolbyParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - HD-Dolby", "", -1)

	// HDDolby主要使用API
	return &indexer.SiteInfo{
		Name:     siteName,
		Schema:   p.schema,
		Language: "zh-cn",
		Enabled:  true,
		Status:   "active",
		Settings: map[string]string{
			"request_mode": "apikey",
		},
		LastCheck: time.Now(),
	}, nil
}

// ParseTorrentList 解析种子列表
func (p *HDDolbyParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	// HDDolby主要使用API返回JSON格式
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseTorrentJSON(html)
	}

	// 如果不是JSON，尝试解析HTML
	return p.parseTorrentHTML(html), nil
}

// parseTorrentJSON 解析JSON格式的种子列表
func (p *HDDolbyParser) parseTorrentJSON(html string) ([]*indexer.TorrentInfo, error) {
	var torrentData HDDolbyTorrentData
	if err := json.Unmarshal([]byte(html), &torrentData); err != nil {
		return nil, fmt.Errorf("failed to parse torrent JSON: %w", err)
	}

	if torrentData.Status != 200 {
		return nil, fmt.Errorf("API returned status: %d", torrentData.Status)
	}

	var torrents []*indexer.TorrentInfo
	for _, item := range torrentData.Data.List {
		torrent := &indexer.TorrentInfo{
			ID:           strconv.Itoa(item.ID),
			Title:        item.Name,
			Size:         item.Size,
			Seeders:      item.Seeders,
			Leechers:     item.Leechers,
			Completed:    item.Completed,
			UploadDate:   time.Unix(item.Added, 0),
			DownloadURL:  fmt.Sprintf("/download.php?id=%d", item.ID),
			DetailURL:    fmt.Sprintf("/details.php?id=%d", item.ID),
			Category:     item.Category.Name,
			Tags:         item.Tags,
			Uploader:     item.Owner.Username,
			IMDBID:       item.IMDBID,
			FreeTorrent:  item.FreeType != "none",
			DoubleUpload: item.DoubleUp,
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// parseTorrentHTML 解析HTML格式的种子列表
func (p *HDDolbyParser) parseTorrentHTML(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子行
	torrentRegex := regexp.MustCompile(`<div[^>]*class="[^"]*torrent[^"]*"[^>]*>(.*?)</div>`)
	torrentMatches := torrentRegex.FindAllStringSubmatch(html, -1)

	for _, match := range torrentMatches {
		if len(match) < 2 {
			continue
		}

		torrent := p.parseTorrentHTMLItem(match[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseTorrentHTMLItem 解析HTML种子项
func (p *HDDolbyParser) parseTorrentHTMLItem(itemHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="/details\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(itemHTML)
	if len(titleMatch) < 3 {
		return nil
	}

	torrentID := titleMatch[1]
	title := strings.TrimSpace(titleMatch[2])
	title = utils.CleanHTMLTags(title)

	// 解析各种数据
	size := p.extractSize(itemHTML)
	seeders := p.extractSeeders(itemHTML)
	leechers := p.extractLeechers(itemHTML)
	completed := p.extractCompleted(itemHTML)
	uploadDate := p.extractUploadTime(itemHTML)
	uploader := p.extractUploader(itemHTML)
	imdbID := p.extractIMDBID(itemHTML)
	freeTorrent := p.checkFreeTorrent(itemHTML)
	doubleUpload := p.checkDoubleUpload(itemHTML)

	return &indexer.TorrentInfo{
		ID:           torrentID,
		Title:        title,
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    completed,
		UploadDate:   uploadDate,
		DownloadURL:  fmt.Sprintf("/download.php?id=%s", torrentID),
		DetailURL:    fmt.Sprintf("/details.php?id=%s", torrentID),
		Uploader:     uploader,
		IMDBID:       imdbID,
		FreeTorrent:  freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// ParseTorrentDetail 解析种子详情
func (p *HDDolbyParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// HDDolby可能返回JSON格式的详情
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseDetailJSON(html)
	}

	return p.parseDetailHTML(html), nil
}

// parseDetailJSON 解析JSON格式的种子详情
func (p *HDDolbyParser) parseDetailJSON(html string) (*indexer.TorrentDetail, error) {
	var detail map[string]interface{}
	if err := json.Unmarshal([]byte(html), &detail); err != nil {
		return nil, fmt.Errorf("failed to parse detail JSON: %w", err)
	}

	title := utils.GetStringFromMap(detail, "name", "")
	description := utils.GetStringFromMap(detail, "description", "")
	imdbID := utils.GetStringFromMap(detail, "imdbId", "")

	// 提取文件列表
	var files []*indexer.TorrentFile
	if fileListData, ok := detail["files"].([]interface{}); ok {
		for _, fileItem := range fileListData {
			if fileMap, ok := fileItem.(map[string]interface{}); ok {
				file := &indexer.TorrentFile{
					Name: utils.GetStringFromMap(fileMap, "name", ""),
					Size: utils.GetInt64FromMap(fileMap, "size", 0),
					Path: utils.GetStringFromMap(fileMap, "path", ""),
				}
				files = append(files, file)
			}
		}
	}

	return &indexer.TorrentDetail{
		Title:       title,
		Description: description,
		Files:       files,
		IMDBID:      imdbID,
		ParsedAt:    time.Now(),
	}, nil
}

// parseDetailHTML 解析HTML格式的种子详情
func (p *HDDolbyParser) parseDetailHTML(html string) *indexer.TorrentDetail {
	titleRegex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) < 2 {
		return nil
	}
	title := strings.TrimSpace(titleMatch[1])

	descRegex := regexp.MustCompile(`<div[^>]*class="[^"]*description[^"]*"[^>]*>(.*?)</div>`)
	descMatch := descRegex.FindStringSubmatch(html)
	description := ""
	if len(descMatch) >= 2 {
		description = utils.CleanHTMLTags(descMatch[1])
	}

	fileList := p.parseFileList(html)
	imdbID := p.extractIMDBID(html)

	return &indexer.TorrentDetail{
		Title:       title,
		Description: description,
		Files:       fileList,
		IMDBID:      imdbID,
		ParsedAt:    time.Now(),
	}
}

// ParseUser 解析用户信息
func (p *HDDolbyParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 尝试解析JSON格式的用户信息
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseUserJSON(html)
	}

	return &indexer.UserInfo{
		Username: "unknown",
		Class:    "User",
		Upload:   0,
		Download: 0,
		Ratio:    0.0,
		Bonus:    0.0,
		ParsedAt: time.Now(),
	}, nil
}

// parseUserJSON 解析JSON格式的用户信息
func (p *HDDolbyParser) parseUserJSON(html string) (*indexer.UserInfo, error) {
	var userData HDDolbyUserData
	if err := json.Unmarshal([]byte(html), &userData); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	if userData.Status != 200 {
		return nil, fmt.Errorf("API returned status: %d", userData.Status)
	}

	// 获取用户等级
	userClass := HDDolbyUserLevelMap[fmt.Sprintf("%d", userData.Data.UserRole)]
	if userClass == "" {
		userClass = "User"
	}

	// 计算分享率
	ratio := 0.0
	if userData.Data.Downloaded > 0 {
		ratio = float64(userData.Data.Uploaded) / float64(userData.Data.Downloaded)
	}

	// 解析加入时间
	var joinTime *time.Time
	if userData.Data.RegTime > 0 {
		t := time.Unix(userData.Data.RegTime, 0)
		joinTime = &t
	}

	return &indexer.UserInfo{
		ID:         strconv.Itoa(userData.Data.ID),
		Username:   userData.Data.Username,
		Class:      userClass,
		JoinDate:   joinTime,
		Upload:     userData.Data.Uploaded,
		Download:   userData.Data.Downloaded,
		Ratio:      ratio,
		Bonus:      userData.Data.Bonus,
		Seeding:    userData.Data.Seeding,
		Leeching:   userData.Data.Leeching,
		Uploaded:   userData.Data.UploadedCount,
		Completed:  userData.Data.FinishedCount,
		ParsedAt:   time.Now(),
	}, nil
}

// 通用提取方法
func (p *HDDolbyParser) extractSize(html string) int64 {
	sizePatterns := []string{
		`"size":\s*(\d+)`,
		`([0-9.]+\s*[KMGT]?B)`,
		`<span[^>]*class="[^"]*size[^"]*"[^>]*>([0-9.]+\s*[KMGT]?B)</span>`,
	}

	for _, pattern := range sizePatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			return utils.ParseFileSize(match[1])
		}
	}
	return 0
}

func (p *HDDolbyParser) extractSeeders(html string) int {
	patterns := []string{
		`"seeders":\s*(\d+)`,
		`<span[^>]*class="[^"]*seeders[^"]*"[^>]*>(\d+)</span>`,
		`做种[：:\s]*(\d+)`,
	}
	return p.extractNumber(html, patterns)
}

func (p *HDDolbyParser) extractLeechers(html string) int {
	patterns := []string{
		`"leechers":\s*(\d+)`,
		`<span[^>]*class="[^"]*leechers[^"]*"[^>]*>(\d+)</span>`,
		`下载[：:\s]*(\d+)`,
	}
	return p.extractNumber(html, patterns)
}

func (p *HDDolbyParser) extractCompleted(html string) int {
	patterns := []string{
		`"completed":\s*(\d+)`,
		`<span[^>]*class="[^"]*completed[^"]*"[^>]*>(\d+)</span>`,
		`完成[：:\s]*(\d+)`,
	}
	return p.extractNumber(html, patterns)
}

func (p *HDDolbyParser) extractNumber(html string, patterns []string) int {
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			num, _ := strconv.Atoi(match[1])
			return num
		}
	}
	return 0
}

func (p *HDDolbyParser) extractUploadTime(html string) time.Time {
	patterns := []string{
		`"added":\s*(\d+)`,
		`<time[^>]*datetime="([^"]*)"[^>]*>`,
		`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			if timestamp, err := strconv.ParseInt(match[1], 10, 64); err == nil && timestamp > 1000000000 {
				return time.Unix(timestamp, 0)
			}
			return utils.ParseTime(match[1])
		}
	}
	return time.Now()
}

func (p *HDDolbyParser) extractUploader(html string) string {
	patterns := []string{
		`"username":\s*"([^"]*)"`,
		`<a[^>]*href="/user/(\d+)"[^>]*>(.*?)</a>`,
		`上传者[：:\s]*([^<\s]*)`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			uploader := match[len(match)-1]
			return utils.CleanHTMLTags(strings.TrimSpace(uploader))
		}
	}
	return ""
}

func (p *HDDolbyParser) extractIMDBID(html string) string {
	patterns := []string{
		`"imdbId":\s*"([^"]*)"`,
		`imdb\.com/title/tt(\d+)`,
		`tt(\d{7,8})`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			return match[1]
		}
	}
	return ""
}

func (p *HDDolbyParser) checkFreeTorrent(html string) bool {
	indicators := []string{
		`"freeType":\s*"[^"]*free[^"]*"`,
		`class="[^"]*free[^"]*"`,
		`freeleech`,
		`免费`,
	}

	for _, indicator := range indicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}
	return false
}

func (p *HDDolbyParser) checkDoubleUpload(html string) bool {
	indicators := []string{
		`"doubleUp":\s*true`,
		`class="[^"]*double[^"]*"`,
		`double.*upload`,
	}

	for _, indicator := range indicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}
	return false
}

func (p *HDDolbyParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	fileRegex := regexp.MustCompile(`<div[^>]*class="[^"]*file[^"]*"[^>]*>.*?<span[^>]*>([^<]+)</span>.*?<span[^>]*>([^<]+)</span>.*?</div>`)
	fileMatches := fileRegex.FindAllStringSubmatch(html, -1)

	for _, match := range fileMatches {
		if len(match) >= 3 {
			fileName := strings.TrimSpace(match[1])
			sizeText := strings.TrimSpace(match[2])
			size := utils.ParseFileSize(sizeText)

			files = append(files, &indexer.TorrentFile{
				Name: fileName,
				Size: size,
				Path: fileName,
			})
		}
	}

	return files
}