// Package parsers YEMA站点解析器
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

// YemaParser YEMA站点解析器
type YemaParser struct {
	schema string
}

// YemaUserData YEMA用户数据结构
type YemaUserData struct {
	Success bool `json:"success"`
	Data    struct {
		ID           int     `json:"id"`
		Username     string  `json:"username"`
		Email        string  `json:"email"`
		Uploaded     int64   `json:"uploaded"`
		Downloaded   int64   `json:"downloaded"`
		UploadBonus  float64 `json:"uploadBonus"`
		SeedTime     int64   `json:"seedTime"`
		UserClass    struct {
			Name string `json:"name"`
		} `json:"userClass"`
		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		Status    string `json:"status"`
	} `json:"data"`
}

// YemaTorrentData YEMA种子数据结构
type YemaTorrentData struct {
	Success bool `json:"success"`
	Data    struct {
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

// NewYemaParser 创建YEMA解析器
func NewYemaParser() *YemaParser {
	return &YemaParser{
		schema: "yema",
	}
}

// GetSchema 获取解析器模式
func (p *YemaParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *YemaParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - YEMA", "", -1)

	// YEMA主要使用API，这里返回基本信息
	return &indexer.SiteInfo{
		Name:     siteName,
		Schema:   p.schema,
		Language: "zh-cn",
		Enabled:  true,
		Status:   "active",
		Settings: map[string]string{
			"api_based": "true",
		},
		LastCheck: time.Now(),
	}, nil
}

// ParseTorrentList 解析种子列表
func (p *YemaParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	// 尝试解析JSON格式的种子列表
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseTorrentJSON(html)
	}

	// 如果不是JSON，尝试解析HTML
	return p.parseTorrentHTML(html), nil
}

// parseTorrentJSON 解析JSON格式的种子列表
func (p *YemaParser) parseTorrentJSON(html string) ([]*indexer.TorrentInfo, error) {
	var torrentData YemaTorrentData
	if err := json.Unmarshal([]byte(html), &torrentData); err != nil {
		return nil, fmt.Errorf("failed to parse torrent JSON: %w", err)
	}

	if !torrentData.Success {
		return nil, fmt.Errorf("API returned success: false")
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
			DownloadURL:  fmt.Sprintf("/torrent/download/%d", item.ID),
			DetailURL:    fmt.Sprintf("/torrent/%d", item.ID),
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
func (p *YemaParser) parseTorrentHTML(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子卡片或行
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
func (p *YemaParser) parseTorrentHTMLItem(itemHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="/torrent/(\d+)"[^>]*>(.*?)</a>`)
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
		DownloadURL:  fmt.Sprintf("/torrent/download/%s", torrentID),
		DetailURL:    fmt.Sprintf("/torrent/%s", torrentID),
		Uploader:     uploader,
		IMDBID:       imdbID,
		FreeTorrent:  freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// ParseTorrentDetail 解析种子详情
func (p *YemaParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// 尝试解析JSON格式
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseDetailJSON(html)
	}

	// HTML格式解析
	return p.parseDetailHTML(html), nil
}

// parseDetailJSON 解析JSON格式的种子详情
func (p *YemaParser) parseDetailJSON(html string) (*indexer.TorrentDetail, error) {
	var detail map[string]interface{}
	if err := json.Unmarshal([]byte(html), &detail); err != nil {
		return nil, fmt.Errorf("failed to parse detail JSON: %w", err)
	}

	// 提取基本信息
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
func (p *YemaParser) parseDetailHTML(html string) *indexer.TorrentDetail {
	// 解析标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) < 2 {
		return nil
	}
	title := strings.TrimSpace(titleMatch[1])

	// 解析描述
	descRegex := regexp.MustCompile(`<div[^>]*class="[^"]*description[^"]*"[^>]*>(.*?)</div>`)
	descMatch := descRegex.FindStringSubmatch(html)
	description := ""
	if len(descMatch) >= 2 {
		description = utils.CleanHTMLTags(descMatch[1])
	}

	// 解析文件列表
	fileList := p.parseFileList(html)

	// 解析IMDB ID
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
func (p *YemaParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 尝试解析JSON格式的用户信息
	if strings.HasPrefix(strings.TrimSpace(html), "{") {
		return p.parseUserJSON(html)
	}

	// 如果不是JSON，使用默认值
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
func (p *YemaParser) parseUserJSON(html string) (*indexer.UserInfo, error) {
	var userData YemaUserData
	if err := json.Unmarshal([]byte(html), &userData); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	if !userData.Success {
		return nil, fmt.Errorf("API returned success: false")
	}

	// 计算分享率
	ratio := 0.0
	if userData.Data.Downloaded > 0 {
		ratio = float64(userData.Data.Uploaded) / float64(userData.Data.Downloaded)
	}

	// 解析加入时间
	var joinTime *time.Time
	if userData.Data.CreatedAt > 0 {
		t := time.Unix(userData.Data.CreatedAt, 0)
		joinTime = &t
	}

	return &indexer.UserInfo{
		ID:       strconv.Itoa(userData.Data.ID),
		Username: userData.Data.Username,
		Class:    userData.Data.UserClass.Name,
		JoinDate: joinTime,
		Upload:   userData.Data.Uploaded,
		Download: userData.Data.Downloaded,
		Ratio:    ratio,
		Bonus:    userData.Data.UploadBonus,
		ParsedAt: time.Now(),
	}, nil
}

// extractSize 提取文件大小
func (p *YemaParser) extractSize(html string) int64 {
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

// extractSeeders 提取做种数
func (p *YemaParser) extractSeeders(html string) int {
	patterns := []string{
		`"seeders":\s*(\d+)`,
		`<span[^>]*class="[^"]*seeders[^"]*"[^>]*>(\d+)</span>`,
		`做种[：:\s]*(\d+)`,
	}

	return p.extractNumber(html, patterns)
}

// extractLeechers 提取下载数
func (p *YemaParser) extractLeechers(html string) int {
	patterns := []string{
		`"leechers":\s*(\d+)`,
		`<span[^>]*class="[^"]*leechers[^"]*"[^>]*>(\d+)</span>`,
		`下载[：:\s]*(\d+)`,
	}

	return p.extractNumber(html, patterns)
}

// extractCompleted 提取完成数
func (p *YemaParser) extractCompleted(html string) int {
	patterns := []string{
		`"completed":\s*(\d+)`,
		`<span[^>]*class="[^"]*completed[^"]*"[^>]*>(\d+)</span>`,
		`完成[：:\s]*(\d+)`,
	}

	return p.extractNumber(html, patterns)
}

// extractNumber 从HTML中提取数字（通用方法）
func (p *YemaParser) extractNumber(html string, patterns []string) int {
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

// extractUploadTime 提取上传时间
func (p *YemaParser) extractUploadTime(html string) time.Time {
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

// extractUploader 提取上传者
func (p *YemaParser) extractUploader(html string) string {
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

// extractIMDBID 提取IMDB ID
func (p *YemaParser) extractIMDBID(html string) string {
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

// checkFreeTorrent 检查是否免费种子
func (p *YemaParser) checkFreeTorrent(html string) bool {
	indicators := []string{
		`"freeType":\s*"[^"]*free[^"]*"`,
		`class="[^"]*free[^"]*"`,
		`freeleech`,
		`free`,
	}

	for _, indicator := range indicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}

	return false
}

// checkDoubleUpload 检查是否双倍上传
func (p *YemaParser) checkDoubleUpload(html string) bool {
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

// parseFileList 解析文件列表
func (p *YemaParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// JSON格式文件列表
	if strings.Contains(html, `"files"`) {
		fileListRegex := regexp.MustCompile(`"files":\s*\[(.*?)\]`)
		fileListMatch := fileListRegex.FindStringSubmatch(html)
		if len(fileListMatch) >= 2 {
			var fileItems []interface{}
			if err := json.Unmarshal([]byte("["+fileListMatch[1]+"]"), &fileItems); err == nil {
				for _, fileItem := range fileItems {
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
		}
	}

	// HTML格式文件列表
	if len(files) == 0 {
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
	}

	return files
}