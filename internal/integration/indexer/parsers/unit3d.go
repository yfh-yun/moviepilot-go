// Package parsers Unit3D站点解析器
package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/integration/indexer"
	"moviepilot-go/pkg/utils"
)

// Unit3DParser Unit3D站点解析器
type Unit3DParser struct {
	schema string
}

// NewUnit3DParser 创建Unit3D解析器
func NewUnit3DParser() *Unit3DParser {
	return &Unit3DParser{
		schema: "unit3d",
	}
}

// GetSchema 获取解析器模式
func (p *Unit3DParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *Unit3DParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - UNIT3D", "", -1)

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
func (p *Unit3DParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// Unit3D使用卡片或表格布局显示种子
	torrents = append(torrents, p.parseTorrentCards(html)...)
	torrents = append(torrents, p.parseTorrentTable(html)...)

	return torrents, nil
}

// parseTorrentCards 解析种子卡片布局
func (p *Unit3DParser) parseTorrentCards(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// Unit3D卡片布局通常是<div class="torrent">或类似的
	cardRegex := regexp.MustCompile(`<div[^>]*class="[^"]*torrent[^"]*"[^>]*>(.*?)</div>\s*(?=<div|</div>|$)`)
	cards := cardRegex.FindAllStringSubmatch(html, -1)

	for _, card := range cards {
		if len(card) < 2 {
			continue
		}

		torrent := p.parseUnit3DTorrentCard(card[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseTorrentTable 解析种子表格布局
func (p *Unit3DParser) parseTorrentTable(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子表格
	tableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*torrents[^"]*"[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	for _, table := range tables {
		if len(table) < 2 {
			continue
		}

		torrents = append(torrents, p.parseUnit3DTorrentRows(table[1])...)
	}

	return torrents
}

// parseUnit3DTorrentRows 解析Unit3D种子行
func (p *Unit3DParser) parseUnit3DTorrentRows(tableHTML string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子行
	rowRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	for i, row := range rows {
		if len(row) < 2 {
			continue
		}

		// 跳过表头
		if i == 0 && (strings.Contains(row[1], "Name") || strings.Contains(row[1], "Title")) {
			continue
		}

		torrent := p.parseUnit3DTorrentRow(row[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseUnit3DTorrentCard 解析Unit3D种子卡片
func (p *Unit3DParser) parseUnit3DTorrentCard(cardHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="/torrents/(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(cardHTML)
	if len(titleMatch) < 3 {
		return nil
	}

	torrentID := titleMatch[1]
	title := strings.TrimSpace(titleMatch[2])
	title = utils.CleanHTMLTags(title)

	// 解析大小
	size := p.extractSize(cardHTML)

	// 解析做种数和下载数
	seeders := p.extractSeeders(cardHTML)
	leechers := p.extractLeechers(cardHTML)
	completed := p.extractCompleted(cardHTML)

	// 解析上传时间
	uploadDate := p.extractUploadTime(cardHTML)

	// 解析上传者
	uploader := p.extractUploader(cardHTML)

	// 检查免费状态
	freeTorrent := p.checkFreeTorrent(cardHTML)

	// 检查双倍上传
	doubleUpload := p.checkDoubleUpload(cardHTML)

	return &indexer.TorrentInfo{
		ID:           torrentID,
		Title:        title,
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    completed,
		UploadDate:   uploadDate,
		DownloadURL:  fmt.Sprintf("/torrents/download/%s", torrentID),
		DetailURL:    fmt.Sprintf("/torrents/%s", torrentID),
		Uploader:     uploader,
		FreeTorrent:  freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// parseUnit3DTorrentRow 解析单行Unit3D种子信息
func (p *Unit3DParser) parseUnit3DTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="/torrents/(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatch) < 3 {
		return nil
	}

	torrentID := titleMatch[1]
	title := strings.TrimSpace(titleMatch[2])
	title = utils.CleanHTMLTags(title)

	// 解析各种数据
	size := p.extractSize(rowHTML)
	seeders := p.extractSeeders(rowHTML)
	leechers := p.extractLeechers(rowHTML)
	completed := p.extractCompleted(rowHTML)
	uploadDate := p.extractUploadTime(rowHTML)
	uploader := p.extractUploader(rowHTML)
	freeTorrent := p.checkFreeTorrent(rowHTML)
	doubleUpload := p.checkDoubleUpload(rowHTML)

	return &indexer.TorrentInfo{
		ID:           torrentID,
		Title:        title,
		Size:         size,
		Seeders:      seeders,
		Leechers:     leechers,
		Completed:    completed,
		UploadDate:   uploadDate,
		DownloadURL:  fmt.Sprintf("/torrents/download/%s", torrentID),
		DetailURL:    fmt.Sprintf("/torrents/%s", torrentID),
		Uploader:     uploader,
		FreeTorrent:  freeTorrent,
		DoubleUpload: doubleUpload,
	}
}

// extractSize 提取文件大小
func (p *Unit3DParser) extractSize(html string) int64 {
	sizePatterns := []string{
		`([0-9.]+\s*[KMGT]?B)`,
		`<span[^>]*class="[^"]*size[^"]*"[^>]*>([0-9.]+\s*[KMGT]?B)</span>`,
		`Size[：:\s]*([0-9.]+\s*[KMGT]?B)`,
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
func (p *Unit3DParser) extractSeeders(html string) int {
	seedersPatterns := []string{
		`<span[^>]*class="[^"]*seeders[^"]*"[^>]*>(\d+)</span>`,
		`<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>`, // 假设第一列是种子数
		`Seeders?[：:\s]*(\d+)`,
	}

	for _, pattern := range seedersPatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			num, _ := strconv.Atoi(match[1])
			return num
		}
	}

	return 0
}

// extractLeechers 提取下载数
func (p *Unit3DParser) extractLeechers(html string) int {
	leechersPatterns := []string{
		`<span[^>]*class="[^"]*leechers[^"]*"[^>]*>(\d+)</span>`,
		`<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>`, // 假设第二列是下载数
		`Leechers?[：:\s]*(\d+)`,
	}

	for _, pattern := range leechersPatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			num, _ := strconv.Atoi(match[1])
			return num
		}
	}

	return 0
}

// extractCompleted 提取完成数
func (p *Unit3DParser) extractCompleted(html string) int {
	completedPatterns := []string{
		`<span[^>]*class="[^"]*completed[^"]*"[^>]*>(\d+)</span>`,
		`<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>`, // 假设第三列是完成数
		`Completed?[：:\s]*(\d+)`,
		`Times[：:\s]*(\d+)`,
	}

	for _, pattern := range completedPatterns {
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
func (p *Unit3DParser) extractUploadTime(html string) time.Time {
	timePatterns := []string{
		`<time[^>]*datetime="([^"]*)"[^>]*>`,
		`<span[^>]*title="([^"]*)"[^>]*class="[^"]*time[^"]*"`,
		`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`,
		`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2})`,
	}

	for _, pattern := range timePatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			return utils.ParseTime(match[1])
		}
	}

	return time.Now()
}

// extractUploader 提取上传者
func (p *Unit3DParser) extractUploader(html string) string {
	uploaderPatterns := []string{
		`<a[^>]*href="/users/([^"/]+)"[^>]*>(.*?)</a>`,
		`<a[^>]*href="/users/\d+"[^>]*>(.*?)</a>`,
		`Uploader[：:\s]*</td>\s*<td[^>]*>(.*?)</td>`,
	}

	for _, pattern := range uploaderPatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			uploader := match[len(match)-1]
			return utils.CleanHTMLTags(strings.TrimSpace(uploader))
		}
	}

	return ""
}

// checkFreeTorrent 检查是否免费种子
func (p *Unit3DParser) checkFreeTorrent(html string) bool {
	freeIndicators := []string{
		`class="[^"]*free[^"]*"`,
		`freeleech`,
		`Free`,
		`100% Free`,
		`class="[^"]*green[^"]*"`, // 通常绿色表示免费
	}

	for _, indicator := range freeIndicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}

	return false
}

// checkDoubleUpload 检查是否双倍上传
func (p *Unit3DParser) checkDoubleUpload(html string) bool {
	doubleIndicators := []string{
		`class="[^"]*double[^"]*"`,
		`double upload`,
		`2x upload`,
		`class="[^"]*orange[^"]*"`, // 通常橙色表示双倍
	}

	for _, indicator := range doubleIndicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}

	return false
}

// ParseTorrentDetail 解析种子详情
func (p *Unit3DParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// 解析标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
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

	// 解析种子ID
	torrentIDRegex := regexp.MustCompile(`/torrents/(\d+)`)
	torrentIDMatch := torrentIDRegex.FindStringSubmatch(html)
	torrentID := ""
	if len(torrentIDMatch) >= 2 {
		torrentID = torrentIDMatch[1]
	}

	return &indexer.TorrentDetail{
		ID:          torrentID,
		Title:       title,
		Description: description,
		Files:       fileList,
		IMDBID:      imdbID,
		DownloadURL: fmt.Sprintf("/torrents/download/%s", torrentID),
		ParsedAt:    time.Now(),
	}, nil
}

// parseFileList 解析文件列表
func (p *Unit3DParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// Unit3D文件列表可能在表格中
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

	// 尝试其他可能的文件列表格式
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

// parseFileRow 解析文件行
func (p *Unit3DParser) parseFileRow(rowHTML string) *indexer.TorrentFile {
	// 解析文件名
	nameRegex := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
	nameMatches := nameRegex.FindAllStringSubmatch(rowHTML, -1)
	if len(nameMatches) < 2 {
		return nil
	}

	fileName := strings.TrimSpace(nameMatches[1][1])
	fileName = utils.CleanHTMLTags(fileName)

	// 解析文件大小
	var size int64 = 0
	if len(nameMatches) >= 3 {
		sizeText := strings.TrimSpace(nameMatches[2][1])
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
func (p *Unit3DParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 解析用户名
	usernameRegex := regexp.MustCompile(`<a[^>]*href="/users/([^"/]+)"[^>]*>(.*?)</a>`)
	usernameMatch := usernameRegex.FindStringSubmatch(html)
	if len(usernameMatch) < 3 {
		// 尝试其他格式
		usernameRegex = regexp.MustCompile(`<a[^>]*href="/users/\d+"[^>]*>(.*?)</a>`)
		usernameMatch = usernameRegex.FindStringSubmatch(html)
		if len(usernameMatch) < 2 {
			return nil, fmt.Errorf("failed to parse username")
		}
	}

	username := strings.TrimSpace(usernameMatch[len(usernameMatch)-1])
	userID := usernameMatch[1]

	// 解析积分
	bonusRegex := regexp.MustCompile(`<a[^>]*href="/bonus/earnings"[^>]*>(.*?)</a>`)
	bonusMatch := bonusRegex.FindStringSubmatch(html)
	bonus := 0.0
	if len(bonusMatch) >= 2 {
		bonusText := strings.TrimSpace(bonusMatch[1])
		bonusText = utils.CleanHTMLTags(bonusText)
		bonusText = regexp.MustCompile(`([0-9,.]+)`).FindString(bonusText)
		bonus = utils.ParseFloat(bonusText)
	}

	// 解析流量信息 - Unit3D通常在用户详情页面显示
	upload := int64(0)
	download := int64(0)
	ratio := 0.0

	// 尝试从页面提取流量信息
	uploadRegex := regexp.MustCompile(`Upload[：:\s]*([0-9.]+\s*[KMGT]?B)`)
	uploadMatch := uploadRegex.FindStringSubmatch(html)
	if len(uploadMatch) >= 2 {
		upload = utils.ParseFileSize(uploadMatch[1])
	}

	downloadRegex := regexp.MustCompile(`Download[：:\s]*([0-9.]+\s*[KMGT]?B)`)
	downloadMatch := downloadRegex.FindStringSubmatch(html)
	if len(downloadMatch) >= 2 {
		download = utils.ParseFileSize(downloadMatch[1])
	}

	if download > 0 {
		ratio = float64(upload) / float64(download)
	}

	// 解析用户等级
	classRegex := regexp.MustCompile(`<span[^>]*class="[^"]*user-class[^"]*"[^>]*>(.*?)</span>`)
	classMatch := classRegex.FindStringSubmatch(html)
	userClass := "User"
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