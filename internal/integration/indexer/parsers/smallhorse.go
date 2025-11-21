// Package parsers SmallHorse站点解析器
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

// SmallHorseParser SmallHorse站点解析器
type SmallHorseParser struct {
	schema string
}

// NewSmallHorseParser 创建SmallHorse解析器
func NewSmallHorseParser() *SmallHorseParser {
	return &SmallHorseParser{
		schema: "smallhorse",
	}
}

// GetSchema 获取解析器模式
func (p *SmallHorseParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *SmallHorseParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - Small Horse", "", -1)

	// 解析用户信息
	userInfo, err := p.ParseUser(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &indexer.SiteInfo{
		Name:     siteName,
		Schema:   p.schema,
		Language: "zh-cn",
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
func (p *SmallHorseParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// SmallHorse使用表格布局显示种子
	torrents = append(torrents, p.parseTorrentTable(html)...)
	torrents = append(torrents, p.parseTorrentCards(html)...)

	return torrents, nil
}

// parseTorrentTable 解析种子表格
func (p *SmallHorseParser) parseTorrentTable(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子表格
	tableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*torrents[^"]*"[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	for _, table := range tables {
		if len(table) < 2 {
			continue
		}

		torrents = append(torrents, p.parseSmallHorseTorrentRows(table[1])...)
	}

	return torrents
}

// parseTorrentCards 解析种子卡片
func (p *SmallHorseParser) parseTorrentCards(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 查找种子卡片
	cardRegex := regexp.MustCompile(`<div[^>]*class="[^"]*torrent[^"]*"[^>]*>(.*?)</div>`)
	cards := cardRegex.FindAllStringSubmatch(html, -1)

	for _, card := range cards {
		if len(card) < 2 {
			continue
		}

		torrent := p.parseSmallHorseTorrentCard(card[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseSmallHorseTorrentRows 解析SmallHorse种子行
func (p *SmallHorseParser) parseSmallHorseTorrentRows(tableHTML string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子行
	rowRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	for i, row := range rows {
		if len(row) < 2 {
			continue
		}

		// 跳过表头
		if i == 0 && (strings.Contains(row[1], "标题") || strings.Contains(row[1], "名称")) {
			continue
		}

		torrent := p.parseSmallHorseTorrentRow(row[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseSmallHorseTorrentCard 解析SmallHorse种子卡片
func (p *SmallHorseParser) parseSmallHorseTorrentCard(cardHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="torrents\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(cardHTML)
	if len(titleMatch) < 3 {
		return nil
	}

	torrentID := titleMatch[1]
	title := strings.TrimSpace(titleMatch[2])
	title = utils.CleanHTMLTags(title)

	// 解析各种数据
	size := p.extractSize(cardHTML)
	seeders := p.extractSeeders(cardHTML)
	leechers := p.extractLeechers(cardHTML)
	completed := p.extractCompleted(cardHTML)
	uploadDate := p.extractUploadTime(cardHTML)
	uploader := p.extractUploader(cardHTML)
	freeTorrent := p.checkFreeTorrent(cardHTML)

	return &indexer.TorrentInfo{
		ID:          torrentID,
		Title:       title,
		Size:        size,
		Seeders:     seeders,
		Leechers:    leechers,
		Completed:   completed,
		UploadDate:  uploadDate,
		DownloadURL: fmt.Sprintf("download.php?id=%s", torrentID),
		DetailURL:   fmt.Sprintf("torrents.php?id=%s", torrentID),
		Uploader:    uploader,
		FreeTorrent: freeTorrent,
	}
}

// parseSmallHorseTorrentRow 解析单行SmallHorse种子信息
func (p *SmallHorseParser) parseSmallHorseTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="torrents\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
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

	return &indexer.TorrentInfo{
		ID:          torrentID,
		Title:       title,
		Size:        size,
		Seeders:     seeders,
		Leechers:    leechers,
		Completed:   completed,
		UploadDate:  uploadDate,
		DownloadURL: fmt.Sprintf("download.php?id=%s", torrentID),
		DetailURL:   fmt.Sprintf("torrents.php?id=%s", torrentID),
		Uploader:    uploader,
		FreeTorrent: freeTorrent,
	}
}

// extractSize 提取文件大小
func (p *SmallHorseParser) extractSize(html string) int64 {
	sizePatterns := []string{
		`([0-9.]+\s*[KMGT]?B)`,
		`<td[^>]*>([0-9.]+\s*[KMGT]?B)</td>`,
		`大小[：:\s]*([0-9.]+\s*[KMGT]?B)`,
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
func (p *SmallHorseParser) extractSeeders(html string) int {
	seedersPatterns := []string{
		`<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>`,
		`做种[：:\s]*(\d+)`,
		`class="[^"]*seeders[^"]*"[^>]*>(\d+)`,
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
func (p *SmallHorseParser) extractLeechers(html string) int {
	leechersPatterns := []string{
		`<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>`,
		`下载[：:\s]*(\d+)`,
		`class="[^"]*leechers[^"]*"[^>]*>(\d+)`,
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
func (p *SmallHorseParser) extractCompleted(html string) int {
	completedPatterns := []string{
		`<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>`,
		`完成[：:\s]*(\d+)`,
		`class="[^"]*completed[^"]*"[^>]*>(\d+)`,
		`次数[：:\s]*(\d+)`,
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
func (p *SmallHorseParser) extractUploadTime(html string) time.Time {
	timePatterns := []string{
		`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`,
		`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2})`,
		`(\d{2}-\d{2}-\d{4}\s+\d{2}:\d{2})`,
		`(\d{4}-\d{2}-\d{2})`,
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
func (p *SmallHorseParser) extractUploader(html string) string {
	uploaderPatterns := []string{
		`<a[^>]*href="user\.php\?id=\d+"[^>]*>(.*?)</a>`,
		`上传者[：:\s]*</td>\s*<td[^>]*>(.*?)</td>`,
		`<a[^>]*href="userdetails\.php\?id=\d+"[^>]*>(.*?)</a>`,
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
func (p *SmallHorseParser) checkFreeTorrent(html string) bool {
	freeIndicators := []string{
		`class="[^"]*free[^"]*"`,
		`freeleech`,
		`Free`,
		`免费`,
		`class="[^"]*green[^"]*"`,
	}

	for _, indicator := range freeIndicators {
		regex := regexp.MustCompile(indicator)
		if regex.MatchString(html) {
			return true
		}
	}

	return false
}

// ParseTorrentDetail 解析种子详情
func (p *SmallHorseParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
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
	torrentIDRegex := regexp.MustCompile(`torrents\.php\?id=(\d+)`)
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
		DownloadURL: fmt.Sprintf("download.php?id=%s", torrentID),
		ParsedAt:    time.Now(),
	}, nil
}

// parseFileList 解析文件列表
func (p *SmallHorseParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// 尝试表格形式的文件列表
	tableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*filelist[^"]*"[^>]*>(.*?)</table>`)
	tableMatch := tableRegex.FindStringSubmatch(html)
	
	if len(tableMatch) >= 2 {
		fileRowsRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
		fileRows := fileRowsRegex.FindAllStringSubmatch(tableMatch[1], -1)

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

	// 尝试从文本中解析文件列表
	if len(files) == 0 {
		// 查找可能包含文件列表的区域
		fileAreaRegex := regexp.MustCompile(`<div[^>]*class="[^"]*files[^"]*"[^>]*>(.*?)</div>`)
		fileAreaMatch := fileAreaRegex.FindStringSubmatch(html)
		
		if len(fileAreaMatch) >= 2 {
			fileContent := utils.CleanHTMLTags(fileAreaMatch[1])
			files = append(files, p.parseTextFileList(fileContent)...)
		}
	}

	return files
}

// parseTextFileList 从文本中解析文件列表
func (p *SmallHorseParser) parseTextFileList(text string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 匹配文件名和大小
		fileRegex := regexp.MustCompile(`([^\s]+\.(?:avi|mkv|mp4|mov|wmv|flv|m4v|mpg|mpeg|iso|img|mdf|nrg|srt|ass|sub))\s+([0-9.]+\s*[KMGT]?B)`)
		match := fileRegex.FindStringSubmatch(line)
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

// parseFileRow 解析文件行
func (p *SmallHorseParser) parseFileRow(rowHTML string) *indexer.TorrentFile {
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
func (p *SmallHorseParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 解析用户ID和用户名
	userRegex := regexp.MustCompile(`<a[^>]*href="user\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	userMatch := userRegex.FindStringSubmatch(html)
	if len(userMatch) < 3 {
		return nil, fmt.Errorf("failed to parse user info")
	}

	userID := userMatch[1]
	username := strings.TrimSpace(userMatch[2])

	// SmallHorse的流量信息通常在统计列表中
	statsRegex := regexp.MustCompile(`<ul[^>]*class="[^"]*stats[^"]*"[^>]*>(.*?)</ul>`)
	statsMatch := statsRegex.FindStringSubmatch(html)
	
	upload := int64(0)
	download := int64(0)
	joinTime := time.Now()
	
	if len(statsMatch) >= 2 {
		statsHTML := statsMatch[1]
		
		// 解析上传量
		uploadRegex := regexp.MustCompile(`上传.*?([0-9.]+\s*[KMGT]?B)`)
		uploadMatch := uploadRegex.FindStringSubmatch(statsHTML)
		if len(uploadMatch) >= 2 {
			upload = utils.ParseFileSize(uploadMatch[1])
		}
		
		// 解析下载量
		downloadRegex := regexp.MustCompile(`下载.*?([0-9.]+\s*[KMGT]?B)`)
		downloadMatch := downloadRegex.FindStringSubmatch(statsHTML)
		if len(downloadMatch) >= 2 {
			download = utils.ParseFileSize(downloadMatch[1])
		}
		
		// 解析加入时间
		joinRegex := regexp.MustCompile(`加入.*?(\d{4}-\d{2}-\d{2})`)
		joinMatch := joinRegex.FindStringSubmatch(statsHTML)
		if len(joinMatch) >= 2 {
			joinTime = utils.ParseTime(joinMatch[1])
		}
	}

	// 计算分享率
	ratio := 0.0
	if download > 0 {
		ratio = float64(upload) / float64(download)
	}

	// 解析用户等级
	classRegex := regexp.MustCompile(`class="[^"]*userclass[^"]*"[^>]*>(.*?)</`)
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
		JoinDate: &joinTime,
		Upload:   upload,
		Download: download,
		Ratio:    ratio,
		Bonus:    0.0,
		ParsedAt: time.Now(),
	}, nil
}