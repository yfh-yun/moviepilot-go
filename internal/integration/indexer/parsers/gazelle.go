// Package parsers Gazelle站点解析器
package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/indexer"
	"github.com/yfh-yun/moviepilot-go/internal/utils"
)

// GazelleParser Gazelle站点解析器
type GazelleParser struct {
	schema string
}

// NewGazelleParser 创建Gazelle解析器
func NewGazelleParser() *GazelleParser {
	return &GazelleParser{
		schema: "gazelle",
	}
}

// GetSchema 获取解析器模式
func (p *GazelleParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *GazelleParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " :: Gazelle", "", -1)

	// 解析用户信息
	userInfo, err := p.ParseUser(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &indexer.SiteInfo{
		Name:     siteName,
		Schema:   p.schema,
		Language: "en",
		Enabled:  true,
		Status:   "active",
		Settings: map[string]string{
			"user_id":    userInfo.ID,
			"username":   userInfo.Username,
			"user_class": userInfo.Class,
		},
		LastCheck: time.Now(),
	}, nil
}

// ParseTorrentList 解析种子列表
func (p *GazelleParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// 解析种子表格
	tableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*torrent_table[^"]*"[^>]*>(.*?)</table>`)
	tables := tableRegex.FindStringSubmatch(html)
	
	if len(tables) < 2 {
		// 尝试其他可能的表格
		tableRegex = regexp.MustCompile(`<table[^>]*>(.*?)</table>`)
		tableMatches := tableRegex.FindAllStringSubmatch(html, -1)
		
		for _, table := range tableMatches {
			if len(table) < 2 {
				continue
			}
			if strings.Contains(table[1], "torrent") || strings.Contains(table[1], "seeders") {
				torrents = append(torrents, p.parseGazelleTorrentRows(table[1])...)
				break
			}
		}
	} else {
		torrents = p.parseGazelleTorrentRows(tables[1])
	}

	return torrents, nil
}

// parseGazelleTorrentRows 解析Gazelle种子行
func (p *GazelleParser) parseGazelleTorrentRows(tableHTML string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子行
	rowRegex := regexp.MustCompile(`<tr[^>]*class="[^"]*torrent[^"]*"[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		torrent := p.parseGazelleTorrentRow(row[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseGazelleTorrentRow 解析单行Gazelle种子信息
func (p *GazelleParser) parseGazelleTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="torrents\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatch) < 3 {
		return nil
	}

	torrentID := titleMatch[1]
	title := strings.TrimSpace(titleMatch[2])
	title = utils.CleanHTMLTags(title)

	// 解析大小
	sizeRegex := regexp.MustCompile(`<td[^>]*class="[^"]*size[^"]*"[^>]*>(.*?)</td>`)
	sizeMatch := sizeRegex.FindStringSubmatch(rowHTML)
	if len(sizeMatch) < 2 {
		return nil
	}
	
	size := utils.ParseFileSize(strings.TrimSpace(sizeMatch[1]))

	// 解析做种数和下载数
	seedersRegex := regexp.MustCompile(`<td[^>]*class="[^"]*seeders[^"]*"[^>]*>(\d+)</td>`)
	seedersMatch := seedersRegex.FindStringSubmatch(rowHTML)
	seeders := 0
	if len(seedersMatch) >= 2 {
		seeders, _ = strconv.Atoi(seedersMatch[1])
	}

	leechersRegex := regexp.MustCompile(`<td[^>]*class="[^"]*leechers[^"]*"[^>]*>(\d+)</td>`)
	leechersMatch := leechersRegex.FindStringSubmatch(rowHTML)
	leechers := 0
	if len(leechersMatch) >= 2 {
		leechers, _ = strconv.Atoi(leechersMatch[1])
	}

	// 解析完成数
	completedRegex := regexp.MustCompile(`<td[^>]*class="[^"]*snatched[^"]*"[^>]*>(\d+)</td>`)
	completedMatch := completedRegex.FindStringSubmatch(rowHTML)
	completed := 0
	if len(completedMatch) >= 2 {
		completed, _ = strconv.Atoi(completedMatch[1])
	}

	// 解析上传时间
	timeRegex := regexp.MustCompile(`<span[^>]*title="([^"]*)"[^>]*>.*?</span>`)
	timeMatch := timeRegex.FindStringSubmatch(rowHTML)
	var uploadDate time.Time
	if len(timeMatch) >= 2 {
		uploadDate = utils.ParseTime(timeMatch[1])
	} else {
		uploadDate = time.Now()
	}

	// 解析上传者
	uploaderRegex := regexp.MustCompile(`<a[^>]*href="user\.php\?id=\d+"[^>]*>(.*?)</a>`)
	uploaderMatch := uploaderRegex.FindStringSubmatch(rowHTML)
	uploader := ""
	if len(uploaderMatch) >= 2 {
		uploader = strings.TrimSpace(uploaderMatch[1])
	}

	// 检查免费状态
	freeTorrent := strings.Contains(rowHTML, "freeleech") || strings.Contains(rowHTML, "Freeleech")

	// 检查双倍上传
	doubleUpload := strings.Contains(rowHTML, "double") || strings.Contains(rowHTML, "Double")

	return &indexer.TorrentInfo{
		ID:           torrentID,
		Title:        title,
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    completed,
		UploadDate:   uploadDate,
		DownloadURL:  fmt.Sprintf("torrents.php?action=download&id=%s", torrentID),
		DetailURL:    fmt.Sprintf("torrents.php?id=%s", torrentID),
		Uploader:     uploader,
		FreeTorrent: freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// ParseTorrentDetail 解析种子详情
func (p *GazelleParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// 解析标题
	titleRegex := regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) < 2 {
		return nil, fmt.Errorf("failed to parse torrent title")
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
	imdbRegex := regexp.MustCompile(`imdb\.com/title/tt(\d+)`)
	imdbMatch := imdbRegex.FindStringSubmatch(html)
	imdbID := ""
	if len(imdbMatch) >= 2 {
		imdbID = imdbMatch[1]
	}

	// 解析种子哈希
	hashRegex := regexp.MustCompile(`<[^>]*>([a-fA-F0-9]{40})</[^>]*>`)
	hashMatch := hashRegex.FindStringSubmatch(html)
	infoHash := ""
	if len(hashMatch) >= 2 {
		infoHash = hashMatch[1]
	}

	return &indexer.TorrentDetail{
		ID:          "",
		Title:       title,
		Description: description,
		Files:       fileList,
		InfoHash:    infoHash,
		IMDBID:      imdbID,
		ParsedAt:    time.Now(),
	}, nil
}

// parseFileList 解析文件列表
func (p *GazelleParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// Gazelle站点文件列表通常在特定的表格中
	fileTableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*filelist[^"]*"[^>]*>(.*?)</table>`)
	fileTableMatch := fileTableRegex.FindStringSubmatch(html)
	
	if len(fileTableMatch) >= 2 {
		fileRowsRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
		fileRows := fileRowsRegex.FindAllStringSubmatch(fileTableMatch[1], -1)

		for _, row := range fileRows {
			if len(row) < 2 {
				continue
			}

			file := p.parseFileRow(row[1])
			if file != nil {
				files = append(files, file)
			}
		}
	}

	return files
}

// parseFileRow 解析文件行
func (p *GazelleParser) parseFileRow(rowHTML string) *indexer.TorrentFile {
	// 解析文件名
	nameRegex := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
	nameMatches := nameRegex.FindAllStringSubmatch(rowHTML, -1)
	if len(nameMatches) < 1 {
		return nil
	}

	fileName := strings.TrimSpace(nameMatches[0][1])
	fileName = utils.CleanHTMLTags(fileName)

	// 解析文件大小
	var size int64 = 0
	if len(nameMatches) >= 2 {
		sizeText := strings.TrimSpace(nameMatches[1][1])
		sizeText = utils.CleanHTMLTags(sizeText)
		size = utils.ParseFileSize(sizeText)
	}

	return &indexer.TorrentFile{
		Name: fileName,
		Size: size,
		Path: fileName,
	}
}

// ParseUser 解析用户信息
func (p *GazelleParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 解析用户ID和用户名
	userRegex := regexp.MustCompile(`<a[^>]*href="user\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	userMatch := userRegex.FindStringSubmatch(html)
	if len(userMatch) < 3 {
		return nil, fmt.Errorf("failed to parse user info")
	}

	userID := userMatch[1]
	username := strings.TrimSpace(userMatch[2])

	// 解析上传量
	uploadRegex := regexp.MustCompile(`<span[^>]*id="header-uploaded-value"[^>]*data-value="(\d+)"`)
	uploadMatch := uploadRegex.FindStringSubmatch(html)
	upload := int64(0)
	if len(uploadMatch) >= 2 {
		upload, _ = strconv.ParseInt(uploadMatch[1], 10, 64)
	}

	// 解析下载量
	downloadRegex := regexp.MustCompile(`<span[^>]*id="header-downloaded-value"[^>]*data-value="(\d+)"`)
	downloadMatch := downloadRegex.FindStringSubmatch(html)
	download := int64(0)
	if len(downloadMatch) >= 2 {
		download, _ = strconv.ParseInt(downloadMatch[1], 10, 64)
	}

	// 计算分享率
	ratio := 0.0
	if download > 0 {
		ratio = float64(upload) / float64(download)
	}

	// 解析积分
	bonusRegex := regexp.MustCompile(`<a[^>]*href="bonus\.php"[^>]*data-tooltip="([^"]*)"`)
	bonusMatch := bonusRegex.FindStringSubmatch(html)
	bonus := 0.0
	if len(bonusMatch) >= 2 {
		bonusText := bonusMatch[1]
		bonusRegex2 := regexp.MustCompile(`([\d,\.]+)`)
		bonusMatch2 := bonusRegex2.FindStringSubmatch(bonusText)
		if len(bonusMatch2) >= 2 {
			bonus = utils.ParseFloat(strings.ReplaceAll(bonusMatch2[1], ",", ""))
		}
	}

	// 解析用户等级
	classRegex := regexp.MustCompile(`<span[^>]*class="[^"]*user-class[^"]*"[^>]*>(.*?)</span>`)
	classMatch := classRegex.FindStringSubmatch(html)
	userClass := "Unknown"
	if len(classMatch) >= 2 {
		userClass = strings.TrimSpace(classMatch[1])
		userClass = utils.CleanHTMLTags(userClass)
	}

	return &indexer.UserInfo{
		ID:       userID,
		Username: username,
		Class:    userClass,
		Upload:   upload,
		Download: download,
		Ratio:    ratio,
		Bonus:    bonus,
		ParsedAt: time.Now(),
	}, nil
}