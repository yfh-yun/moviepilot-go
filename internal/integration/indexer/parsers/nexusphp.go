// Package parsers NexusPHP站点解析器
package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/integration/indexer"
)

// NexusPHPParser NexusPHP站点解析器
type NexusPHPParser struct {
	schema string
}

// NewNexusPHPParser 创建NexusPHP解析器
func NewNexusPHPParser() *NexusPHPParser {
	return &NexusPHPParser{
		schema: "nexusphp",
	}
}

// GetSchema 获取解析器模式
func (p *NexusPHPParser) GetSchema() string {
	return p.schema
}

// ParseSite 解析站点信息
func (p *NexusPHPParser) ParseSite(html string) (*indexer.SiteInfo, error) {
	// 解析站点名称
	nameRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	nameMatch := nameRegex.FindStringSubmatch(html)
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("failed to parse site name")
	}

	siteName := strings.TrimSpace(nameMatch[1])
	siteName = strings.Replace(siteName, " - Powered by NexusPHP", "", -1)

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
func (p *NexusPHPParser) ParseTorrentList(html string) ([]*indexer.TorrentInfo, error) {
	var torrents []*indexer.TorrentInfo

	// 解析种子表格
	tableRegex := regexp.MustCompile(`<table[^>]*>(.*?)</table>`)
	tables := tableRegex.FindAllStringSubmatch(html, -1)

	for _, table := range tables {
		if len(table) < 2 {
			continue
		}

		// 解析种子行
		rowRegex := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
		rows := rowRegex.FindAllStringSubmatch(table[1], -1)

		for i, row := range rows {
			if len(row) < 2 || i == 0 {
				// 跳过表头
				continue
			}

			torrent, err := p.parseTorrentRow(row[1])
			if err != nil {
				continue
			}

			torrents = append(torrents, torrent)
		}
	}

	return torrents, nil
}

// ParseTorrentDetail 解析种子详情
func (p *NexusPHPParser) ParseTorrentDetail(html string) (*indexer.TorrentDetail, error) {
	torrentInfo, err := p.parseTorrentInfo(html)
	if err != nil {
		return nil, err
	}

	// 解析文件列表
	files, err := p.parseFileList(html)
	if err != nil {
		return nil, err
	}

	// 解析评论
	comments, err := p.parseComments(html)
	if err != nil {
		return nil, err
	}

	// 解析连接信息
	peers, err := p.parsePeers(html)
	if err != nil {
		return nil, err
	}

	return &indexer.TorrentDetail{
		TorrentInfo: torrentInfo,
		Files:       files,
		Comments:    comments,
		Peers:       peers,
	}, nil
}

// ParseUser 解析用户信息
func (p *NexusPHPParser) ParseUser(html string) (*indexer.UserInfo, error) {
	userInfo := &indexer.UserInfo{}

	// 解析用户名
	usernameRegex := regexp.MustCompile(`(?:欢迎|Welcome),\s*<[^>]*>([^<]+)</`)
	usernameMatch := usernameRegex.FindStringSubmatch(html)
	if len(usernameMatch) >= 2 {
		userInfo.Username = strings.TrimSpace(usernameMatch[1])
	}

	// 解析用户ID
	idRegex := regexp.MustCompile(`user\.php\?id=(\d+)`)
	idMatch := idRegex.FindStringSubmatch(html)
	if len(idMatch) >= 2 {
		userInfo.ID = idMatch[1]
	}

	// 解析用户等级
	classRegex := regexp.MustCompile(`(?:等级|Class)[^:]*:\s*([^<\s]+)`)
	classMatch := classRegex.FindStringSubmatch(html)
	if len(classMatch) >= 2 {
		userInfo.Class = strings.TrimSpace(classMatch[1])
	}

	// 解析上传下载量
	uploadRegex := regexp.MustCompile(`(?:上传|Uploaded)[^:]*:\s*([\d.]+\s*[KMGT]?B)`)
	uploadMatch := uploadRegex.FindStringSubmatch(html)
	if len(uploadMatch) >= 2 {
		userInfo.Uploaded = p.parseSize(uploadMatch[1])
	}

	downloadRegex := regexp.MustCompile(`(?:下载|Downloaded)[^:]*:\s*([\d.]+\s*[KMGT]?B)`)
	downloadMatch := downloadRegex.FindStringSubmatch(html)
	if len(downloadMatch) >= 2 {
		userInfo.Downloaded = p.parseSize(downloadMatch[1])
	}

	// 计算分享率
	if userInfo.Downloaded > 0 {
		userInfo.Ratio = float64(userInfo.Uploaded) / float64(userInfo.Downloaded)
	}

	// 解析做种数
	seedingRegex := regexp.MustCompile(`(?:做种|Seeding)[^:]*:\s*(\d+)`)
	seedingMatch := seedingRegex.FindStringSubmatch(html)
	if len(seedingMatch) >= 2 {
		userInfo.Seeding, _ = strconv.Atoi(seedingMatch[1])
	}

	// 解析下载数
	leechingRegex := regexp.MustCompile(`(?:下载中|Leeching)[^:]*:\s*(\d+)`)
	leechingMatch := leechingRegex.FindStringSubmatch(html)
	if len(leechingMatch) >= 2 {
		userInfo.Leeching, _ = strconv.Atoi(leechingMatch[1])
	}

	// 解析魔力值
	bonusRegex := regexp.MustCompile(`(?:魔力值|Bonus)[^:]*:\s*([\d.,]+)`)
	bonusMatch := bonusRegex.FindStringSubmatch(html)
	if len(bonusMatch) >= 2 {
		bonusStr := strings.ReplaceAll(bonusMatch[1], ",", "")
		userInfo.BonusPoints, _ = strconv.ParseFloat(bonusStr, 64)
	}

	// 解析邀请数
	invitesRegex := regexp.MustCompile(`(?:邀请|Invites)[^:]*:\s*(\d+)`)
	invitesMatch := invitesRegex.FindStringSubmatch(html)
	if len(invitesMatch) >= 2 {
		userInfo.Invites, _ = strconv.Atoi(invitesMatch[1])
	}

	return userInfo, nil
}

// parseTorrentRow 解析种子行
func (p *NexusPHPParser) parseTorrentRow(rowHTML string) (*indexer.TorrentInfo, error) {
	torrent := &indexer.TorrentInfo{}

	// 解析单元格
	cellRegex := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
	cells := cellRegex.FindAllStringSubmatch(rowHTML, -1)

	if len(cells) < 10 {
		return nil, fmt.Errorf("invalid torrent row")
	}

	// 解析标题和链接
	titleCell := cells[1][1]
	titleRegex := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>([^<]+)</a>`)
	titleMatch := titleRegex.FindStringSubmatch(titleCell)
	if len(titleMatch) >= 3 {
		torrent.Title = strings.TrimSpace(titleMatch[2])
		torrent.DetailURL = titleMatch[1]
	}

	// 解析下载链接
	downloadRegex := regexp.MustCompile(`<a[^>]*href="([^"]*download[^"]*)"[^>]*>`)
	downloadMatch := downloadRegex.FindStringSubmatch(rowHTML)
	if len(downloadMatch) >= 2 {
		torrent.DownloadURL = downloadMatch[1]
	}

	// 解析大小
	sizeCell := cells[4][1]
	sizeRegex := regexp.MustCompile(`([\d.]+\s*[KMGT]?B)`)
	sizeMatch := sizeRegex.FindStringSubmatch(sizeCell)
	if len(sizeMatch) >= 2 {
		torrent.Size = p.parseSize(sizeMatch[1])
	}

	// 解析种子数
	seedersCell := cells[6][1]
	seedersRegex := regexp.MustCompile(`(\d+)`)
	seedersMatch := seedersRegex.FindStringSubmatch(seedersCell)
	if len(seedersMatch) >= 2 {
		torrent.Seeders, _ = strconv.Atoi(seedersMatch[1])
	}

	// 解析下载数
	leechersCell := cells[7][1]
	leechersMatch := seedersRegex.FindStringSubmatch(leechersCell)
	if len(leechersMatch) >= 2 {
		torrent.Leechers, _ = strconv.Atoi(leechersMatch[1])
	}

	// 解析完成数
	completedCell := cells[8][1]
	completedMatch := seedersRegex.FindStringSubmatch(completedCell)
	if len(completedMatch) >= 2 {
		torrent.Completed, _ = strconv.Atoi(completedMatch[1])
	}

	// 解析上传时间
	timeCell := cells[5][1]
	timeRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	timeMatch := timeRegex.FindStringSubmatch(timeCell)
	if len(timeMatch) >= 2 {
		if uploadTime, err := time.Parse("2006-01-02 15:04:05", timeMatch[1]); err == nil {
			torrent.UploadDate = uploadTime
		}
	}

	// 解析免费标记
	if strings.Contains(rowHTML, "free") || strings.Contains(rowHTML, "freeleech") {
		torrent.FreeTorrent = true
	}

	// 解析2X上传
	if strings.Contains(rowHTML, "2x") || strings.Contains(rowHTML, "double") {
		torrent.DoubleUpload = true
	}

	// 解析HDR
	if strings.Contains(strings.ToLower(rowHTML), "hdr") {
		torrent.HDR = true
	}

	// 解析4K
	if strings.Contains(strings.ToUpper(rowHTML), "4K") || strings.Contains(rowHTML, "2160p") {
		torrent.UHD = true
	}

	// 解析分类和标签
	p.parseCategoryAndTags(torrent, rowHTML)

	return torrent, nil
}

// parseTorrentInfo 解析种子信息
func (p *NexusPHPParser) parseTorrentInfo(html string) (*indexer.TorrentInfo, error) {
	torrent := &indexer.TorrentInfo{}

	// 解析标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	if len(titleMatch) >= 2 {
		torrent.Title = strings.TrimSpace(titleMatch[1])
	}

	// 解析大小
	sizeRegex := regexp.MustCompile(`(?:大小|Size)[^:]*:\s*([\d.]+\s*[KMGT]?B)`)
	sizeMatch := sizeRegex.FindStringSubmatch(html)
	if len(sizeMatch) >= 2 {
		torrent.Size = p.parseSize(sizeMatch[1])
	}

	// 解析种子数、下载数、完成数
	seedersRegex := regexp.MustCompile(`(?:种子|Seeders)[^:]*:\s*(\d+)`)
	seedersMatch := seedersRegex.FindStringSubmatch(html)
	if len(seedersMatch) >= 2 {
		torrent.Seeders, _ = strconv.Atoi(seedersMatch[1])
	}

	leechersRegex := regexp.MustCompile(`(?:下载|Leechers)[^:]*:\s*(\d+)`)
	leechersMatch := leechersRegex.FindStringSubmatch(html)
	if len(leechersMatch) >= 2 {
		torrent.Leechers, _ = strconv.Atoi(leechersMatch[1])
	}

	completedRegex := regexp.MustCompile(`(?:完成|Completed)[^:]*:\s*(\d+)`)
	completedMatch := completedRegex.FindStringSubmatch(html)
	if len(completedMatch) >= 2 {
		torrent.Completed, _ = strconv.Atoi(completedMatch[1])
	}

	// 解析IMDB ID
	imdbRegex := regexp.MustCompile(`imdb\.com/title/tt(\d+)`)
	imdbMatch := imdbRegex.FindStringSubmatch(html)
	if len(imdbMatch) >= 2 {
		torrent.IMDBID = "tt" + imdbMatch[1]
	}

	return torrent, nil
}

// parseFileList 解析文件列表
func (p *NexusPHPParser) parseFileList(html string) ([]*indexer.TorrentFile, error) {
	var files []*indexer.TorrentFile

	// 解析文件表格
	fileRegex := regexp.MustCompile(`<tr[^>]*>.*?<td[^>]*>([^<]+)</td>.*?<td[^>]*>([^<]+)</td>.*?</tr>`)
	fileMatches := fileRegex.FindAllStringSubmatch(html, -1)

	for _, match := range fileMatches {
		if len(match) >= 3 {
			file := &indexer.TorrentFile{
				Path: strings.TrimSpace(match[1]),
				Size: p.parseSize(match[2]),
			}
			files = append(files, file)
		}
	}

	return files, nil
}

// parseComments 解析评论
func (p *NexusPHPParser) parseComments(html string) ([]string, error) {
	var comments []string

	// 解析评论内容
	commentRegex := regexp.MustCompile(`<div[^>]*class="comment"[^>]*>(.*?)</div>`)
	commentMatches := commentRegex.FindAllStringSubmatch(html, -1)

	for _, match := range commentMatches {
		if len(match) >= 2 {
			// 移除HTML标签
			content := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(match[1], "")
			content = strings.TrimSpace(content)
			if content != "" {
				comments = append(comments, content)
			}
		}
	}

	return comments, nil
}

// parsePeers 解析连接信息
func (p *NexusPHPParser) parsePeers(html string) ([]*indexer.PeerInfo, error) {
	var peers []*indexer.PeerInfo

	// 解析连接列表
	peerRegex := regexp.MustCompile(`<tr[^>]*>.*?<td[^>]*>([^<]+)</td>.*?<td[^>]*>([^<]+)</td>.*?<td[^>]*>([^<]+)</td>.*?</tr>`)
	peerMatches := peerRegex.FindAllStringSubmatch(html, -1)

	for _, match := range peerMatches {
		if len(match) >= 4 {
			peer := &indexer.PeerInfo{
				IP:     strings.TrimSpace(match[1]),
				Client: strings.TrimSpace(match[2]),
			}

			// 解析进度
			progressRegex := regexp.MustCompile(`([\d.]+)%`)
			progressMatch := progressRegex.FindStringSubmatch(match[3])
			if len(progressMatch) >= 2 {
				if progress, err := strconv.ParseFloat(progressMatch[1], 64); err == nil {
					peer.Progress = progress / 100
				}
			}

			peer.UpdatedAt = time.Now()
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

// parseCategoryAndTags 解析分类和标签
func (p *NexusPHPParser) parseCategoryAndTags(torrent *indexer.TorrentInfo, html string) {
	// 解析分类
	categoryRegex := regexp.MustCompile(`<span[^>]*class="category[^"]*"[^>]*>([^<]+)</span>`)
	categoryMatch := categoryRegex.FindStringSubmatch(html)
	if len(categoryMatch) >= 2 {
		torrent.Category = strings.TrimSpace(categoryMatch[1])
	}

	// 解析标签
	tagRegex := regexp.MustCompile(`<span[^>]*class="tag[^"]*"[^>]*>([^<]+)</span>`)
	tagMatches := tagRegex.FindAllStringSubmatch(html, -1)

	for _, match := range tagMatches {
		if len(match) >= 2 {
			tag := strings.TrimSpace(match[1])
			if tag != "" {
				torrent.Tags = append(torrent.Tags, tag)
			}
		}
	}

	// 解析媒体信息
	p.parseMediaInfo(torrent, html)
}

// parseMediaInfo 解析媒体信息
func (p *NexusPHPParser) parseMediaInfo(torrent *indexer.TorrentInfo, html string) {
	lowerHTML := strings.ToLower(html)

	// 解析分辨率
	resolutionRegex := regexp.MustCompile(`(2160p|1080p|720p|480p)`)
	resolutionMatch := resolutionRegex.FindStringSubmatch(lowerHTML)
	if len(resolutionMatch) >= 2 {
		torrent.Resolution = strings.ToUpper(resolutionMatch[1])
	}

	// 解析编码
	codecRegex := regexp.MustCompile(`(x264|x265|h\.264|h\.265|hevc|avc)`)
	codecMatch := codecRegex.FindStringSubmatch(lowerHTML)
	if len(codecMatch) >= 2 {
		torrent.Codec = strings.ToUpper(codecMatch[1])
	}

	// 解析容器
	containerRegex := regexp.MustCompile(`\.(mkv|mp4|avi|mov|wmv)`)
	containerMatch := containerRegex.FindStringSubmatch(lowerHTML)
	if len(containerMatch) >= 2 {
		torrent.Container = strings.ToUpper(containerMatch[1])
	}

	// 解析音频
	audioRegex := regexp.MustCompile(`(dts|ac3|aac|flac|mp3)`)
	audioMatches := audioRegex.FindAllStringSubmatch(lowerHTML, -1)
	for _, match := range audioMatches {
		if len(match) >= 1 {
			audio := strings.ToUpper(match[1])
			torrent.Audio = audio
			break // 只取第一个音频格式
		}
	}

	// 解析字幕
	subtitleRegex := regexp.MustCompile(`(中英|中字|英字|简体|繁体|chinese|english)`)
	subtitleMatches := subtitleRegex.FindAllStringSubmatch(lowerHTML, -1)
	for _, match := range subtitleMatches {
		if len(match) >= 1 {
			torrent.Subtitles = append(torrent.Subtitles, match[1])
		}
	}
}

// parseSize 解析大小字符串
func (p *NexusPHPParser) parseSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	sizeStr = strings.ReplaceAll(sizeStr, " ", "")

	if strings.HasSuffix(sizeStr, "B") {
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	} else if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "T")
	}

	if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		return int64(size * float64(multiplier))
	}

	return 0
}