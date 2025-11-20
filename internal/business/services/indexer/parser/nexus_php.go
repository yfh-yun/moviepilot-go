// Package parser 索引器Parser包
package parser

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NexusPHPParser NexusPHP站点解析器
type NexusPHPParser struct {
	*BaseParser
}

// NewNexusPHPParser 创建NexusPHP解析器
func NewNexusPHPParser() *NexusPHPParser {
	parser := &NexusPHPParser{
		BaseParser: NewBaseParser("NexusPHP", "nexusphp"),
	}
	
	// 设置站点信息
	parser.siteInfo.Features = []string{
		"免费种子", "推广种子", "上传倍率", "下载倍率",
		"做种积分", "魔力值", "邀请系统", "用户等级",
		"收藏夹", "短评", "感谢", "积分商城",
	}
	
	parser.siteInfo.Categories = []string{
		"电影", "电视剧", "动漫", "音乐", "游戏", "软件", "纪录片", "体育",
		"综艺", "电子书", "其他", "合集",
	}
	
	parser.siteInfo.Resolutions = []string{
		"4K", "2160p", "1080p", "720p", "480p", "360p", "240p",
	}
	
	parser.siteInfo.VideoCodecs = []string{
		"x264", "x265", "HEVC", "AVC", "VC-1", "MPEG2", "DivX", "XviD",
	}
	
	parser.siteInfo.AudioCodecs = []string{
		"AAC", "AC3", "DTS", "FLAC", "MP3", "OGG", "TrueHD", "LPCM",
	}
	
	parser.siteInfo.Containers = []string{
		"MKV", "MP4", "AVI", "MOV", "WMV", "FLV", "M2TS", "ISO",
	}
	
	parser.siteInfo.Sources = []string{
		"BluRay", "BDRip", "BRRip", "DVDRip", "HDDVD", "HDTV", "PDTV", "DSR",
		"WEB-DL", "WEBRip", "CAM", "TS", "TC", "WP", "Workprint",
	}
	
	return parser
}

// ParseTorrentPage 解析种子列表页
func (n *NexusPHPParser) ParseTorrentPage(ctx context.Context, htmlContent string, page int) ([]*TorrentInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	var torrents []*TorrentInfo

	// NexusPHP标准种子列表表格
	doc.Find("table.torrents > tbody > tr").Each(func(i int, s *goquery.Selection) {
		// 跳过表头和空行
		if s.Find("td").Length() < 10 {
			return
		}

		torrent, err := n.parseTorrentRow(s)
		if err != nil {
			// 记录错误但继续处理其他行
			fmt.Printf("解析种子行失败: %v\n", err)
			return
		}

		torrents = append(torrents, torrent)
	})

	return torrents, nil
}

// parseTorrentRow 解析种子行
func (n *NexusPHPParser) parseTorrentRow(row *goquery.Selection) (*TorrentInfo, error) {
	torrent := &TorrentInfo{
		Tags:  []string{},
		Meta:   make(map[string]string),
	}

	cells := row.Find("td")
	if cells.Length() < 10 {
		return nil, fmt.Errorf("种子行列数不足")
	}

	// 解析分类和图标
	if err := n.parseCategory(cells.Eq(0), torrent); err != nil {
		return nil, fmt.Errorf("解析分类失败: %w", err)
	}

	// 解析标题
	if err := n.parseTitle(cells.Eq(1), torrent); err != nil {
		return nil, fmt.Errorf("解析标题失败: %w", err)
	}

	// 解析下载链接
	if err := n.parseDownloadLink(cells.Eq(2), torrent); err != nil {
		return nil, fmt.Errorf("解析下载链接失败: %w", err)
	}

	// 解析评论数
	if err := n.parseComments(cells.Eq(3), torrent); err != nil {
		return nil, fmt.Errorf("解析评论数失败: %w", err)
	}

	// 解析时间
	if err := n.parseTime(cells.Eq(4), torrent); err != nil {
		return nil, fmt.Errorf("解析时间失败: %w", err)
	}

	// 解析大小
	if err := n.parseSize(cells.Eq(5), torrent); err != nil {
		return nil, fmt.Errorf("解析大小失败: %w", err)
	}

	// 解析种子数
	if err := n.parseSeeders(cells.Eq(6), torrent); err != nil {
		return nil, fmt.Errorf("解析种子数失败: %w", err)
	}

	// 解析下载数
	if err := n.parseLeechers(cells.Eq(7), torrent); err != nil {
		return nil, fmt.Errorf("解析下载数失败: %w", err)
	}

	// 解析下载次数
	if err := n.parseDownloads(cells.Eq(8), torrent); err != nil {
		return nil, fmt.Errorf("解析下载次数失败: %w", err)
	}

	// 解析上传者
	if err := n.parseUploader(cells.Eq(9), torrent); err != nil {
		return nil, fmt.Errorf("解析上传者失败: %w", err)
	}

	// 解析特殊标识
	n.parseSpecialTags(row, torrent)

	return torrent, nil
}

// parseCategory 解析分类
func (n *NexusPHPParser) parseCategory(cell *goquery.Selection, torrent *TorrentInfo) error {
	img := cell.Find("img:first-child")
	if img.Length() == 0 {
		return nil
	}

	// 获取图片alt属性作为分类
	if alt := n.GetAttr(img, "alt"); alt != "" {
		torrent.Category = alt
	}

	// 获取图片title属性作为子分类
	if title := n.GetAttr(img, "title"); title != "" {
		torrent.SubCategory = title
	}

	// 通过图片src判断具体分类
	if src := n.GetAttr(img, "src"); src != "" {
		torrent.Meta["category_icon"] = src
		
		// 根据图标映射分类
		if strings.Contains(src, "cat_401") {
			torrent.Category = "电影"
		} else if strings.Contains(src, "cat_402") {
			torrent.Category = "电视剧"
		} else if strings.Contains(src, "cat_403") {
			torrent.Category = "动漫"
		}
	}

	return nil
}

// parseTitle 解析标题
func (n *NexusPHPParser) parseTitle(cell *goquery.Selection, torrent *TorrentInfo) error {
	link := cell.Find("a[href*='details.php']:first-child")
	if link.Length() == 0 {
		return nil
	}

	// 获取标题文本
	torrent.Title = n.CleanText(n.GetText(link))

	// 获取详情页URL
	if href := n.GetAttr(link, "href"); href != "" {
		torrent.URL = href
		torrent.Meta["detail_url"] = href

		// 从URL中提取种子ID
		if parsedURL, err := url.Parse(href); err == nil {
			if values, err := url.ParseQuery(parsedURL.RawQuery); err == nil {
				if id := values.Get("id"); id != "" {
					torrent.ID = id
				}
			}
		}
	}

	// 解析媒体信息
	n.parseMediaInfoFromTitle(torrent)

	// 解析IMDB/TMDB ID
	n.parseMediaIDs(cell, torrent)

	return nil
}

// parseDownloadLink 解析下载链接
func (n *NexusPHPParser) parseDownloadLink(cell *goquery.Selection, torrent *TorrentInfo) error {
	link := cell.Find("a[href*='download.php']:first-child")
	if link.Length() == 0 {
		return nil
	}

	if href := n.GetAttr(link, "href"); href != "" {
		torrent.DownloadURL = href
		torrent.Meta["download_url"] = href
	}

	return nil
}

// parseComments 解析评论数
func (n *NexusPHPParser) parseComments(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.Comments = n.ExtractNumber(text)
	return nil
}

// parseTime 解析时间
func (n *NexusPHPParser) parseTime(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.UploadedAt = n.ParseTime(text)
	
	// 解析剩余时间TTL
	if span := cell.Find("span[title]"); span.Length() > 0 {
		if title := n.GetAttr(span, "title"); title != "" {
			torrent.TTL = title
		}
	}

	return nil
}

// parseSize 解析大小
func (n *NexusPHPParser) parseSize(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.Size = n.ExtractSize(text)
	return nil
}

// parseSeeders 解析种子数
func (n *NexusPHPParser) parseSeeders(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.Seeders = n.ExtractNumber(text)
	return nil
}

// parseLeechers 解析下载数
func (n *NexusPHPParser) parseLeechers(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.Leechers = n.ExtractNumber(text)
	return nil
}

// parseDownloads 解析下载次数
func (n *NexusPHPParser) parseDownloads(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	torrent.Downloads = n.ExtractNumber(text)
	return nil
}

// parseUploader 解析上传者
func (n *NexusPHPParser) parseUploader(cell *goquery.Selection, torrent *TorrentInfo) error {
	text := n.GetText(cell)
	if text != "" {
		torrent.Meta["uploader"] = text
	}

	// 检查是否为匿名上传者
	if cell.Find("i").Length() > 0 || strings.Contains(text, "匿名") {
		torrent.Meta["anonymous"] = "true"
	}

	return nil
}

// parseSpecialTags 解析特殊标签
func (n *NexusPHPParser) parseSpecialTags(row *goquery.Selection, torrent *TorrentInfo) {
	// 检查免费种子
	if row.Find("img[alt*='Free'], img[alt*='免费']").Length() > 0 {
		torrent.FreeLeech = true
		torrent.DownloadFactor = 0.0
		torrent.Tags = append(torrent.Tags, "免费")
	}

	// 检查推广种子
	if row.Find("img[alt*='2X'], img[alt*='50%'], img[alt*='上传']").Length() > 0 {
		torrent.Promotional = true
		torrent.UploadFactor = 2.0
		torrent.Tags = append(torrent.Tags, "推广")
	}

	// 检查HDR
	if strings.Contains(strings.ToUpper(torrent.Title), "HDR") ||
		row.Find("span:contains('HDR'), span:contains('hdr')").Length() > 0 {
		torrent.HDR = true
		torrent.Tags = append(torrent.Tags, "HDR")
	}

	// 检查配音
	if strings.Contains(strings.ToUpper(torrent.Title), "DUB") ||
		row.Find("span:contains('配音'), span:contains('国语')").Length() > 0 {
		torrent.Dubbed = true
		torrent.Tags = append(torrent.Tags, "配音")
	}

	// 检查字幕
	if strings.Contains(strings.ToUpper(torrent.Title), "SUB") ||
		row.Find("span:contains('字幕'), span:contains('中字')").Length() > 0 {
		torrent.Subtitled = true
		torrent.Tags = append(torrent.Tags, "字幕")
	}

	// 解析分辨率
	if res := n.parseResolution(torrent.Title); res != "" {
		torrent.Tags = append(torrent.Tags, res)
		torrent.Meta["resolution"] = res
	}

	// 解析视频编码
	if codec := n.parseVideoCodec(torrent.Title); codec != "" {
		torrent.Tags = append(torrent.Tags, codec)
		torrent.Meta["video_codec"] = codec
	}

	// 解析音频编码
	if codec := n.parseAudioCodec(torrent.Title); codec != "" {
		torrent.Tags = append(torrent.Tags, codec)
		torrent.Meta["audio_codec"] = codec
	}
}

// parseMediaInfoFromTitle 从标题解析媒体信息
func (n *NexusPHPParser) parseMediaInfoFromTitle(torrent *TorrentInfo) {
	title := strings.ToUpper(torrent.Title)

	// 判断媒体类型
	if strings.Contains(title, "S0") && strings.Contains(title, "E0") {
		torrent.MediaType = "episode"
	} else if strings.Contains(title, "COMPLETE") || strings.Contains(title, "全季") {
		torrent.MediaType = "tv"
	} else {
		torrent.MediaType = "movie"
	}

	// 解析年份
	yearRe := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if match := yearRe.FindString(title); match != "" {
		if year, err := strconv.Atoi(match); err == nil {
			torrent.MediaInfo = &MediaInfo{Year: year}
		}
	}

	// 解析季集信息
	if torrent.MediaType == "episode" {
		seasonRe := regexp.MustCompile(`S(\d{1,2})`)
		episodeRe := regexp.MustCompile(`E(\d{1,3})`)

		if match := seasonRe.FindStringSubmatch(title); len(match) > 1 {
			if season, err := strconv.Atoi(match[1]); err == nil && torrent.MediaInfo != nil {
				torrent.MediaInfo.Season = season
			}
		}

		if match := episodeRe.FindStringSubmatch(title); len(match) > 1 {
			if episode, err := strconv.Atoi(match[1]); err == nil && torrent.MediaInfo != nil {
				torrent.MediaInfo.Episode = episode
			}
		}
	}
}

// parseMediaIDs 解析媒体ID
func (n *NexusPHPParser) parseMediaIDs(cell *goquery.Selection, torrent *TorrentInfo) {
	// 查找IMDB链接
	if imdbLink := cell.Find("a[href*='imdb.com']"); imdbLink.Length() > 0 {
		if href := n.GetAttr(imdbLink, "href"); href != "" {
			if imdbID := n.extractIMDBID(href); imdbID != "" {
				torrent.IMDBID = imdbID
				if torrent.MediaInfo == nil {
					torrent.MediaInfo = &MediaInfo{}
				}
				torrent.MediaInfo.IMDBID = imdbID
			}
		}
	}

	// 查找豆瓣链接
	if doubanLink := cell.Find("a[href*='douban.com']"); doubanLink.Length() > 0 {
		if href := n.GetAttr(doubanLink, "href"); href != "" {
			torrent.Meta["douban_url"] = href
		}
	}
}

// extractIMDBID 从URL中提取IMDB ID
func (n *NexusPHPParser) extractIMDBID(url string) string {
	re := regexp.MustCompile(`tt\d{7,8}`)
	if match := re.FindString(url); match != "" {
		return match
	}
	return ""
}

// parseResolution 解析分辨率
func (n *NexusPHPParser) parseResolution(title string) string {
	title = strings.ToUpper(title)
	
	resolutions := []string{
		"4K", "2160P", "1080P", "720P", "480P", "360P",
	}
	
	for _, res := range resolutions {
		if strings.Contains(title, res) {
			return res
		}
	}
	
	return ""
}

// parseVideoCodec 解析视频编码
func (n *NexusPHPParser) parseVideoCodec(title string) string {
	title = strings.ToUpper(title)
	
	codecs := []string{
		"X265", "HEVC", "X264", "AVC", "VC1", "MPEG2",
	}
	
	for _, codec := range codecs {
		if strings.Contains(title, codec) {
			return codec
		}
	}
	
	return ""
}

// parseAudioCodec 解析音频编码
func (n *NexusPHPParser) parseAudioCodec(title string) string {
	title = strings.ToUpper(title)
	
	codecs := []string{
		"DTS", "AC3", "AAC", "FLAC", "MP3", "TRUEHD", "LPCM",
	}
	
	for _, codec := range codecs {
		if strings.Contains(title, codec) {
			return codec
		}
	}
	
	return ""
}

// ParseTorrentDetail 解析种子详情页
func (n *NexusPHPParser) ParseTorrentDetail(ctx context.Context, htmlContent string) (*TorrentDetail, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	detail := &TorrentDetail{
		Files:          []*TorrentFile{},
		Screenshots:     []string{},
		CommentsList:    []*Comment{},
		Meta:            make(map[string]string),
		Tags:            []string{},
	}

	// 解析基本信息
	if err := n.parseDetailBasicInfo(doc, detail); err != nil {
		return nil, fmt.Errorf("解析基本信息失败: %w", err)
	}

	// 解析文件列表
	if err := n.parseDetailFileList(doc, detail); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}

	// 解析截图
	if err := n.parseDetailScreenshots(doc, detail); err != nil {
		return nil, fmt.Errorf("解析截图失败: %w", err)
	}

	// 解析技术信息
	if err := n.parseDetailTechnicalInfo(doc, detail); err != nil {
		return nil, fmt.Errorf("解析技术信息失败: %w", err)
	}

	// 解析评论
	if err := n.parseDetailComments(doc, detail); err != nil {
		return nil, fmt.Errorf("解析评论失败: %w", err)
	}

	return detail, nil
}

// parseDetailBasicInfo 解析详情页基本信息
func (n *NexusPHPParser) parseDetailBasicInfo(doc *goquery.Document, detail *TorrentDetail) error {
	// 解析标题
	if title := doc.Find("h1").First().Text(); title != "" {
		detail.Title = n.CleanText(title)
	}

	// 解析描述
	if desc := doc.Find("#description").First().Text(); desc != "" {
		detail.Description = n.CleanText(desc)
	}

	// 解析下载链接
	if downloadLink := doc.Find("a[href*='download.php']").First(); downloadLink.Length() > 0 {
		if href := n.GetAttr(downloadLink, "href"); href != "" {
			detail.DownloadURL = href
		}
	}

	// 解析磁力链接
	if magnetLink := doc.Find("a[href^='magnet:']").First(); magnetLink.Length() > 0 {
		if href := n.GetAttr(magnetLink, "href"); href != "" {
			detail.MagnetURL = href
		}
	}

	// 解析基本信息表格
	doc.Find("table:contains('大小') tr, table:contains('大小') > tbody > tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 2 {
			label := n.GetText(cells.Eq(0))
			value := n.GetText(cells.Eq(1))

			switch {
			case strings.Contains(label, "大小"):
				detail.Size = n.ExtractSize(value)
			case strings.Contains(label, "种子数"):
				detail.Seeders = n.ExtractNumber(value)
			case strings.Contains(label, "下载数"):
				detail.Leechers = n.ExtractNumber(value)
			case strings.Contains(label, "完成数"):
				detail.Downloads = n.ExtractNumber(value)
			}
		}
	})

	return nil
}

// parseDetailFileList 解析文件列表
func (n *NexusPHPParser) parseDetailFileList(doc *goquery.Document, detail *TorrentDetail) error {
	doc.Find("#filelist tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 2 {
			file := &TorrentFile{}
			file.Path = n.GetText(cells.Eq(0))
			file.Size = n.ExtractSize(n.GetText(cells.Eq(1)))
			
			// 提取文件扩展名
			if lastDot := strings.LastIndex(file.Path, "."); lastDot != -1 {
				file.Extension = strings.ToUpper(file.Path[lastDot+1:])
			}
			
			detail.Files = append(detail.Files, file)
		}
	})

	return nil
}

// parseDetailScreenshots 解析截图
func (n *NexusPHPParser) parseDetailScreenshots(doc *goquery.Document, detail *TorrentDetail) error {
	doc.Find(".postimg img, .screenshot img").Each(func(i int, s *goquery.Selection) {
		if src := n.GetAttr(s, "src"); src != "" {
			detail.Screenshots = append(detail.Screenshots, src)
		}
	})

	return nil
}

// parseDetailTechnicalInfo 解析技术信息
func (n *NexusPHPParser) parseDetailTechnicalInfo(doc *goquery.Document, detail *TorrentDetail) error {
	detail.TechnicalInfo = &TechnicalInfo{}

	// 解析技术信息表格
	doc.Find("table:contains('视频') tr, table:contains('音频') tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 2 {
			label := n.GetText(cells.Eq(0))
			value := n.GetText(cells.Eq(1))

			switch {
			case strings.Contains(label, "容器"):
				detail.TechnicalInfo.Container = value
			case strings.Contains(label, "视频编码"):
				detail.TechnicalInfo.VideoCodec = value
			case strings.Contains(label, "分辨率"):
				detail.TechnicalInfo.Resolution = value
			case strings.Contains(label, "音频编码"):
				detail.TechnicalInfo.AudioCodec = value
			case strings.Contains(label, "字幕"):
				detail.TechnicalInfo.Subtitle = value
			}
		}
	})

	return nil
}

// parseDetailComments 解析评论
func (n *NexusPHPParser) parseDetailComments(doc *goquery.Document, detail *TorrentDetail) error {
	doc.Find(".comment").Each(func(i int, s *goquery.Selection) {
		comment := &Comment{}

		// 评论ID
		if id := n.GetAttr(s, "id"); id != "" {
			comment.ID = id
		}

		// 用户信息
		if userCell := s.Find(".comment_user, .comment_author").First(); userCell.Length() > 0 {
			comment.Username = n.GetText(userCell)
		}

		// 评论内容
		if contentCell := s.Find(".comment_content, .comment_text").First(); contentCell.Length() > 0 {
			comment.Content = n.GetText(contentCell)
		}

		// 时间
		if timeCell := s.Find(".comment_time, .comment_date").First(); timeCell.Length() > 0 {
			comment.CreatedAt = n.GetText(timeCell)
		}

		detail.CommentsList = append(detail.CommentsList, comment)
	})

	return nil
}

// 其他方法的默认实现
func (n *NexusPHPParser) ParseSearchResult(ctx context.Context, htmlContent string, keyword string, page int) ([]*TorrentInfo, error) {
	// NexusPHP的搜索结果页面与种子列表页面格式相同
	return n.ParseTorrentPage(ctx, htmlContent, page)
}

func (n *NexusPHPParser) ParseUserPage(ctx context.Context, htmlContent string, userID string) (*UserInfo, error) {
	userInfo := &UserInfo{
		ID: userID,
		Meta: make(map[string]string),
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	// 解析用户基本信息
	if title := doc.Find("h1").First().Text(); title != "" {
		userInfo.Username = n.CleanText(title)
	}

	// 解析用户等级
	if classCell := doc.Find("table:contains('等级') tr").Find("td").Eq(1); classCell.Length() > 0 {
		userInfo.Class = n.GetText(classCell)
	}

	// 解析上传下载量
	doc.Find("table:contains('上传量') tr, table:contains('下载量') tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 2 {
			label := n.GetText(cells.Eq(0))
			value := n.GetText(cells.Eq(1))

			if strings.Contains(label, "上传量") {
				userInfo.Uploaded = n.ExtractSize(value)
			} else if strings.Contains(label, "下载量") {
				userInfo.Downloaded = n.ExtractSize(value)
			}
		}
	})

	// 计算分享率
	if userInfo.Downloaded > 0 {
		userInfo.Ratio = float64(userInfo.Uploaded) / float64(userInfo.Downloaded)
	}

	return userInfo, nil
}