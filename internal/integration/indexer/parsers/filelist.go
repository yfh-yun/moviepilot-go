// Package parsers FileList站点解析器
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

// FileListParser FileList站点解析器
type FileListParser struct {
	schema string
}

// NewFileListParser 创建FileList解析器
func NewFileListParser() *FileListParser {
	return &FileListParser{
		schema: "filelist",
	}
}

// GetSchema 获取解析器模式
func (p *FileListParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *FileListParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " :: FileList", "", -1)

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
func (p *FileListParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// FileList使用表格布局显示种子
	tableRegex := regexp.MustCompile(`<table[^>]*class="[^"]*torrents[^"]*"[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	for _, table := range tables {
		if len(table) < 2 {
			continue
		}

		torrents = append(torrents, p.parseFileListTorrentRows(table[1])...)
	}

	// 如果没有找到标准表格，尝试其他可能的结构
	if len(torrents) == 0 {
		torrents = append(torrents, p.parseAlternativeTorrentRows(html)...)
	}

	return torrents, nil
}

// parseFileListTorrentRows 解析FileList种子行
func (p *FileListParser) parseFileListTorrentRows(tableHTML string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 解析种子行
	rowRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	for i, row := range rows {
		if len(row) < 2 {
			continue
		}

		// 跳过表头
		if i == 0 && (strings.Contains(row[1], "Name") || strings.Contains(row[1], "Type")) {
			continue
		}

		torrent := p.parseFileListTorrentRow(row[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseAlternativeTorrentRows 解析替代格式的种子行
func (p *FileListParser) parseAlternativeTorrentRows(html string) []*indexer.TorrentInfo {
	var torrents []*indexer.TorrentInfo

	// 查找包含种子信息的div
	divRegex := regexp.MustCompile(`<div[^>]*class="[^"]*torrent[^"]*"[^>]*>(.*?)</div>`)
	divs := divRegex.FindAllStringSubmatch(html, -1)

	for _, div := range divs {
		if len(div) < 2 {
			continue
		}

		torrent := p.parseFileListTorrentRow(div[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseFileListTorrentRow 解析单行FileList种子信息
func (p *FileListParser) parseFileListTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="details\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatch) < 3 {
		// 尝试其他可能的链接格式
		titleRegex = regexp.MustCompile(`<a[^>]*href="([^"]*id=(\d+)[^"]*)"[^>]*>(.*?)</a>`)
		titleMatch = titleRegex.FindStringSubmatch(rowHTML)
		if len(titleMatch) < 4 {
			return nil
		}
	}

	torrentID := titleMatch[len(titleMatch)-2]
	title := strings.TrimSpace(titleMatch[len(titleMatch)-1])
	title = utils.CleanHTMLTags(title)

	// 解析大小
	size := p.extractSize(rowHTML)

	// 解析做种数和下载数
	seeders := p.extractSeeders(rowHTML)
	leechers := p.extractLeechers(rowHTML)
	completed := p.extractCompleted(rowHTML)

	// 解析上传时间
	uploadDate := p.extractUploadTime(rowHTML)

	// 解析上传者
	uploader := p.extractUploader(rowHTML)

	// 解析类型
	category := p.extractCategory(rowHTML)

	// 检查免费状态
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
		DetailURL:   fmt.Sprintf("details.php?id=%s", torrentID),
		Category:    category,
		Uploader:    uploader,
		FreeTorrent: freeTorrent,
	}
}

// extractSize 提取文件大小
func (p *FileListParser) extractSize(html string) int64 {
	sizePatterns := []string{
		`([0-9.]+\s*[KMGT]?B)`,
		`<td[^>]*>([0-9.]+\s*[KMGT]?B)</td>`,
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
func (p *FileListParser) extractSeeders(html string) int {
	seedersPatterns := []string{
		`<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>`, // 假设第一列是种子数
		`Seeders?[：:\s]*(\d+)`,
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
func (p *FileListParser) extractLeechers(html string) int {
	leechersPatterns := []string{
		`<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>\s*<td[^>]*>\d+</td>`, // 假设第二列是下载数
		`Leechers?[：:\s]*(\d+)`,
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
func (p *FileListParser) extractCompleted(html string) int {
	completedPatterns := []string{
		`<td[^>]*>\d+</td>\s*<td[^>]*>\d+</td>\s*<td[^>]*>(\d+)</td>`, // 假设第三列是完成数
		`Completed?[：:\s]*(\d+)`,
		`class="[^"]*completed[^"]*"[^>]*>(\d+)`,
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
func (p *FileListParser) extractUploadTime(html string) time.Time {
	timePatterns := []string{
		`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`,
		`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2})`,
		`(\d{2}-\d{2}-\d{4}\s+\d{2}:\d{2})`,
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
func (p *FileListParser) extractUploader(html string) string {
	uploaderPatterns := []string{
		`<a[^>]*href="userdetails\.php\?id=\d+"[^>]*>(.*?)</a>`,
		`Uploader[：:\s]*</td>\s*<td[^>]*>(.*?)</td>`,
		`<a[^>]*href="user\.php\?id=\d+"[^>]*>(.*?)</a>`,
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

// extractCategory 提取分类
func (p *FileListParser) extractCategory(html string) string {
	categoryPatterns := []string{
		`<td[^>]*>(.*?)</td>\s*<td[^>]*>.*?</td>`, // 假设第一列是分类
		`Category[：:\s]*</td>\s*<td[^>]*>(.*?)</td>`,
		`class="[^"]*category[^"]*"[^>]*>(.*?)</`,
	}

	for _, pattern := range categoryPatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			category := match[1]
			return utils.CleanHTMLTags(strings.TrimSpace(category))
		}
	}

	return ""
}

// checkFreeTorrent 检查是否免费种子
func (p *FileListParser) checkFreeTorrent(html string) bool {
	freeIndicators := []string{
		`class="[^"]*free[^"]*"`,
		`freeleech`,
		`Free`,
		`100% Free`,
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
func (p *FileListParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// 解析标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) < 2 {
		return nil, fmt.Errorf("failed to parse torrent title")
	}
	title := strings.TrimSpace(titleMatch[1])

	// 解析描述
	descRegex := regexp.MustCompile(`<div[^>]*class="[^"]*nfo[^"]*"[^>]*>(.*?)</div>`)
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
	torrentIDRegex := regexp.MustCompile(`details\.php\?id=(\d+)`)
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
func (p *FileListParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// FileList文件列表可能在NFO区域或表格中
	// 首先尝试查找NFO区域
	nfoRegex := regexp.MustCompile(`<div[^>]*class="[^"]*nfo[^"]*"[^>]*>(.*?)</div>`)
	nfoMatch := nfoRegex.FindStringSubmatch(html)
	
	if len(nfoMatch) >= 2 {
		nfoContent := utils.CleanHTMLTags(nfoMatch[1])
		files = append(files, p.parseTextFileList(nfoContent)...)
	}

	// 尝试表格形式的文件列表
	if len(files) == 0 {
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
	}

	return files
}

// parseTextFileList 从文本中解析文件列表
func (p *FileListParser) parseTextFileList(text string) []*indexer.TorrentFile {
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
func (p *FileListParser) parseFileRow(rowHTML string) *indexer.TorrentFile {
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
func (p *FileListParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 解析用户ID和用户名
	userRegex := regexp.MustCompile(`<a[^>]*href="userdetails\.php\?id=(\d+)"[^>]*>(.*?)</a>`)
	userMatch := userRegex.FindStringSubmatch(html)
	if len(userMatch) < 3 {
		return nil, fmt.Errorf("failed to parse user info")
	}

	userID := userMatch[1]
	username := strings.TrimSpace(userMatch[2])

	// 解析流量信息 - 从用户详情页面
	uploadRegex := regexp.MustCompile(`Uploaded[</td>]+\s*<td[^>]*>(.*?)</td>`)
	uploadMatch := uploadRegex.FindStringSubmatch(html)
	upload := int64(0)
	if len(uploadMatch) >= 2 {
		uploadText := strings.TrimSpace(uploadMatch[1])
		upload = utils.ParseFileSize(uploadText)
	}

	downloadRegex := regexp.MustCompile(`Downloaded[</td>]+\s*<td[^>]*>(.*?)</td>`)
	downloadMatch := downloadRegex.FindStringSubmatch(html)
	download := int64(0)
	if len(downloadMatch) >= 2 {
		downloadText := strings.TrimSpace(downloadMatch[1])
		download = utils.ParseFileSize(downloadText)
	}

	// 计算分享率
	ratio := 0.0
	if download > 0 {
		ratio = float64(upload) / float64(download)
	}

	// 解析用户等级
	classRegex := regexp.MustCompile(`Class[</td>]+\s*<td[^>]*>(.*?)</td>`)
	classMatch := classRegex.FindStringSubmatch(html)
	userClass := "User"
	if len(classMatch) >= 2 {
		userClass = strings.TrimSpace(classMatch[1])
		userClass = utils.CleanHTMLTags(userClass)
	}

	// 解析加入时间
	joinTime := time.Now()
	joinTimeRegex := regexp.MustCompile(`Joined[</td>]+\s*<td[^>]*>(.*?)</td>`)
	joinTimeMatch := joinTimeRegex.FindStringSubmatch(html)
	if len(joinTimeMatch) >= 2 {
		joinTime = utils.ParseTime(strings.TrimSpace(joinTimeMatch[1]))
	}

	return &indexer.UserInfo{
		ID:       userID,
		Username: username,
		Class:    userClass,
		JoinDate: &joinTime,
		Upload:   upload,
		Download: download,
		Ratio:    ratio,
		Bonus:    0.0, // FileList可能没有积分系统
		ParsedAt: time.Now(),
	}, nil
}