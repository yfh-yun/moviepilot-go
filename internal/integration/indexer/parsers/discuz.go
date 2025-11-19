// Package parsers Discuz站点解析器
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

// DiscuzParser Discuz站点解析器
type DiscuzParser struct {
	schema string
}

// NewDiscuzParser 创建Discuz解析器
func NewDiscuzParser() *DiscuzParser {
	return &DiscuzParser{
		schema: "discuz",
	}
}

// GetSchema 获取解析器模式
func (p *DiscuzParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *DiscuzParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - Powered by Discuz!", "", -1)
	siteName = strings.Replace(siteName, " - Discuz! Board", "", -1)

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
func (p *DiscuzParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// Discuz站点的种子列表可能在不同的表格结构中
	tableRegex := regexp.MustCompile(`<table[^>]*summary="[^"]*torrent[^"]*"[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	for _, table := range tables {
		if len(table) < 2 {
			continue
		}

		torrents = append(torrents, p.parseDiscuzTorrentRows(table[1])...)
	}

	// 如果没有找到标准表格，尝试其他可能的结构
	if len(torrents) == 0 {
		tableRegex = regexp.MustCompile(`<table[^>]*>(.*?)</table>`)
		tableMatches := tableRegex.FindAllStringSubmatch(html, -1)

		for _, table := range tableMatches {
			if len(table) < 2 {
				continue
			}
			// 检查是否包含种子相关的内容
			if strings.Contains(table[1], "种子") || strings.Contains(table[1], "torrent") ||
				strings.Contains(table[1], "大小") || strings.Contains(table[1], "完成") {
				torrents = append(torrents, p.parseDiscuzTorrentRows(table[1])...)
			}
		}
	}

	return torrents, nil
}

// parseDiscuzTorrentRows 解析Discuz种子行
func (p *DiscuzParser) parseDiscuzTorrentRows(tableHTML string) []*indexer.TorrentInfo {
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

		torrent := p.parseDiscuzTorrentRow(row[1])
		if torrent != nil {
			torrents = append(torrents, torrent)
		}
	}

	return torrents
}

// parseDiscuzTorrentRow 解析单行Discuz种子信息
func (p *DiscuzParser) parseDiscuzTorrentRow(rowHTML string) *indexer.TorrentInfo {
	// 解析标题和链接
	titleRegex := regexp.MustCompile(`<a[^>]*href="([^"]*torrent[^"]*id=(\d+)[^"]*)"[^>]*>(.*?)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(rowHTML)
	if len(titleMatch) < 4 {
		// 尝试另一种格式
		titleRegex = regexp.MustCompile(`<a[^>]*href="viewthread\.php\?tid=(\d+)"[^>]*>(.*?)</a>`)
		titleMatch = titleRegex.FindStringSubmatch(rowHTML)
		if len(titleMatch) < 3 {
			return nil
		}
	}

	torrentID := titleMatch[len(titleMatch)-2]
	title := strings.TrimSpace(titleMatch[len(titleMatch)-1])
	title = utils.CleanHTMLTags(title)

	// 解析大小
	sizeRegex := regexp.MustCompile(`([0-9.]+\s*[KMGT]?B)`)
	sizeMatch := sizeRegex.FindStringSubmatch(rowHTML)
	size := int64(0)
	if len(sizeMatch) >= 2 {
		size = utils.ParseFileSize(sizeMatch[1])
	}

	// 解析做种数和下载数
	// Discuz格式可能不同，需要适应各种可能的表达方式
	seeders := p.extractNumber(rowHTML, []string{"做种数", "种子数", "seeder"})
	leechers := p.extractNumber(rowHTML, []string{"下载数", "下载中", "leecher"})
	completed := p.extractNumber(rowHTML, []string{"完成数", "完成量", "complete"})

	// 解析上传时间
	uploadDate := p.parseUploadTime(rowHTML)

	// 解析上传者
	uploader := p.extractUploader(rowHTML)

	return &indexer.TorrentInfo{
		ID:         torrentID,
		Title:      title,
		Size:       size,
		Seeders:    seeders,
		Leechers:   leechers,
		Completed:  completed,
		UploadDate: uploadDate,
		DetailURL:  fmt.Sprintf("viewthread.php?tid=%s", torrentID),
		Uploader:   uploader,
	}
}

// extractNumber 从HTML中提取数字
func (p *DiscuzParser) extractNumber(html string, keywords []string) int {
	for _, keyword := range keywords {
		// 查找关键词后面的数字
		regex := regexp.MustCompile(fmt.Sprintf(`%s[：:\s]*(\d+)`, keyword))
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			num, _ := strconv.Atoi(match[1])
			return num
		}

		// 查找表格单元格中的数字
		cellRegex := regexp.MustCompile(fmt.Sprintf(`<td[^>]*>[^<]*%s[^<]*</td>\s*<td[^>]*>(\d+)</td>`, keyword))
		cellMatch := cellRegex.FindStringSubmatch(html)
		if len(cellMatch) >= 2 {
			num, _ := strconv.Atoi(cellMatch[1])
			return num
		}
	}
	return 0
}

// parseUploadTime 解析上传时间
func (p *DiscuzParser) parseUploadTime(html string) time.Time {
	// 尝试解析各种时间格式
	timePatterns := []string{
		`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`,
		`(\d{4}-\d{2}-\d{2})`,
		`(\d{2}:\d{2}:\d{2}\s+\d{4}-\d{2}-\d{2})`,
		`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2})`,
	}

	for _, pattern := range timePatterns {
		regex := regexp.MustCompile(pattern)
		match := regex.FindStringSubmatch(html)
		if len(match) >= 2 {
			return utils.ParseTime(match[1])
		}
	}

	// 尝试解析相对时间
	relativeTimeRegex := regexp.MustCompile(`(\d+)\s*(秒|分钟|小时|天|周|月|年)前`)
	relativeMatch := relativeTimeRegex.FindStringSubmatch(html)
	if len(relativeMatch) >= 3 {
		num, _ := strconv.Atoi(relativeMatch[1])
		unit := relativeMatch[2]
		
		duration := time.Duration(num) * time.Second
		switch unit {
		case "分钟":
			duration *= time.Minute
		case "小时":
			duration *= time.Hour
		case "天":
			duration *= 24 * time.Hour
		case "周":
			duration *= 7 * 24 * time.Hour
		case "月":
			duration *= 30 * 24 * time.Hour
		case "年":
			duration *= 365 * 24 * time.Hour
		}
		
		return time.Now().Add(-duration)
	}

	return time.Now()
}

// extractUploader 提取上传者信息
func (p *DiscuzParser) extractUploader(html string) string {
	// 解析上传者链接
	uploaderRegex := regexp.MustCompile(`<a[^>]*href="[^"]*uid=(\d+)"[^>]*>(.*?)</a>`)
	uploaderMatch := uploaderRegex.FindStringSubmatch(html)
	if len(uploaderMatch) >= 3 {
		return strings.TrimSpace(uploaderMatch[2])
	}

	// 尝试其他可能的格式
	uploaderRegex = regexp.MustCompile(`上传者[：:\s]*</td>\s*<td[^>]*>(.*?)</td>`)
	uploaderMatch = uploaderRegex.FindStringSubmatch(html)
	if len(uploaderMatch) >= 2 {
		return utils.CleanHTMLTags(strings.TrimSpace(uploaderMatch[1]))
	}

	return ""
}

// ParseTorrentDetail 解析种子详情
func (p *DiscuzParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	// 解析标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) < 2 {
		return nil, fmt.Errorf("failed to parse torrent title")
	}
	title := strings.TrimSpace(titleMatch[1])

	// 解析描述
	descRegex := regexp.MustCompile(`<div[^>]*class="[^"]*postmessage[^"]*"[^>]*>(.*?)</div>`)
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

	// 解析下载链接
	downloadRegex := regexp.MustCompile(`<a[^>]*href="([^"]*attachment[^"]*id=(\d+)[^"]*)"[^>]*>(.*?)</a>`)
	downloadMatch := downloadRegex.FindStringSubmatch(html)
	torrentID := ""
	downloadURL := ""
	if len(downloadMatch) >= 3 {
		torrentID = downloadMatch[2]
		downloadURL = downloadMatch[1]
	}

	return &indexer.TorrentDetail{
		ID:          torrentID,
		Title:       title,
		Description: description,
		Files:       fileList,
		IMDBID:      imdbID,
		DownloadURL: downloadURL,
		ParsedAt:    time.Now(),
	}, nil
}

// parseFileList 解析文件列表
func (p *DiscuzParser) parseFileList(html string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	// Discuz文件列表可能在帖子内容中
	fileRegex := regexp.MustCompile(`([^\s]+\.(avi|mkv|mp4|mov|wmv|flv|m4v|mpg|mpeg|iso|img|mdf|nrg))\s*\(([^)]+)\)`)
	fileMatches := fileRegex.FindAllStringSubmatch(html, -1)

	for _, match := range fileMatches {
		if len(match) >= 4 {
			fileName := strings.TrimSpace(match[1])
			sizeText := strings.TrimSpace(match[3])
			size := utils.ParseFileSize(sizeText)

			files = append(files, &indexer.TorrentFile{
				Name: fileName,
				Size: size,
				Path: fileName,
			})
		}
	}

	// 尝试其他可能的文件列表格式
	if len(files) == 0 {
		// 查找代码块中的文件列表
		codeRegex := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
		codeMatches := codeRegex.FindAllStringSubmatch(html, -1)

		for _, codeMatch := range codeMatches {
			if len(codeMatch) >= 2 {
				codeContent := utils.CleanHTMLTags(codeMatch[1])
				codeFiles := p.parseTextFileList(codeContent)
				files = append(files, codeFiles...)
			}
		}
	}

	return files
}

// parseTextFileList 从文本中解析文件列表
func (p *DiscuzParser) parseTextFileList(text string) []*indexer.TorrentFile {
	var files []*indexer.TorrentFile

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 匹配文件名和大小
		fileRegex := regexp.MustCompile(`([^\s]+\.(?:avi|mkv|mp4|mov|wmv|flv|m4v|mpg|mpeg|iso|img|mdf|nrg))\s+([0-9.]+\s*[KMGT]?B)`)
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

// ParseUser 解析用户信息
func (p *DiscuzParser) ParseUser(html string) (*indexer.UserInfo, error) {
	// 解析用户ID和用户名
	userRegex := regexp.MustCompile(`<a[^>]*href="[^"]*uid=(\d+)"[^>]*>(.*?)</a>`)
	userMatch := userRegex.FindStringSubmatch(html)
	if len(userMatch) < 3 {
		return nil, fmt.Errorf("failed to parse user info")
	}

	userID := userMatch[1]
	username := strings.TrimSpace(userMatch[2])

	// 解析用户等级
	classRegex := regexp.MustCompile(`<span[^>]*class="[^"]*usergroup[^"]*"[^>]*>(.*?)</span>`)
	classMatch := classRegex.FindStringSubmatch(html)
	userClass := "Unknown"
	if len(classMatch) >= 2 {
		userClass = strings.TrimSpace(classMatch[1])
		userClass = utils.CleanHTMLTags(userClass)
	}

	// 解析加入时间
	joinTime := time.Now()
	joinTimeRegex := regexp.MustCompile(`注册时间[：:\s]*(\d{4}-\d{2}-\d{2})`)
	joinTimeMatch := joinTimeRegex.FindStringSubmatch(html)
	if len(joinTimeMatch) >= 2 {
		joinTime = utils.ParseTime(joinTimeMatch[1])
	}

	// Discuz站点通常不直接显示流量信息，需要从其他页面获取
	// 这里使用默认值
	upload := int64(0)
	download := int64(0)
	ratio := 0.0
	bonus := 0.0

	return &indexer.UserInfo{
		ID:       userID,
		Username: username,
		Class:    userClass,
		JoinDate: &joinTime,
		Upload:   upload,
		Download: download,
		Ratio:    ratio,
		Bonus:    bonus,
		ParsedAt: time.Now(),
	}, nil
}