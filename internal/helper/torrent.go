package helper

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/core/cache"
	"moviepilot-go/internal/db"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/types"
	"moviepilot-go/internal/utils"
)

// TorrentHelper 种子帮助�?type TorrentHelper struct {
	InvalidTorrents *cache.TTLCache
}

// NewTorrentHelper 创建TorrentHelper实例
func NewTorrentHelper() *TorrentHelper {
	return &TorrentHelper{
		InvalidTorrents: cache.NewTTLCache(128, 24*time.Hour),
	}
}

// DownloadResult 下载结果
type DownloadResult struct {
	CachePath   string
	Content     []byte
	FolderName  string
	FileList    []string
	ErrorMsg    string
	IsMagnet    bool
}

// DownloadTorrent 把种子下载到本地
// 返回: 种子缓存相对路径【用于索引缓存�? 种子内容、种子主目录、种子文件清单、错误信�?func (t *TorrentHelper) DownloadTorrent(urlStr, cookie, ua, referer string, proxy bool) *DownloadResult {
	result := &DownloadResult{}

	if strings.HasPrefix(urlStr, "magnet:") {
		result.Content = []byte(urlStr)
		result.IsMagnet = true
		result.ErrorMsg = "磁力链接"
		return result
	}

	// 构建 torrent 种子文件的缓存路�?	cachePath := fmt.Sprintf("%s.torrent", md5Hash(urlStr))
	result.CachePath = cachePath

	// 缓存处理�?	cacheBackend := cache.NewFileCache()

	// 读取缓存的种子文�?	torrentContent, err := cacheBackend.Get(cachePath, "torrents")
	if err == nil && torrentContent != nil {
		// 缓存已存�?		folderName, fileList, err := t.GetFileInfoFromTorrentContent(torrentContent)
		if err == nil && folderName != "" && len(fileList) > 0 {
			// 成功拿到种子数据
			result.Content = torrentContent
			result.FolderName = folderName
			result.FileList = fileList
			return result
		}
		logger.Errorf("处理缓存的种子文�?%s 时出�? %v，将重新下载", cachePath, err)
	}

	// 下载种子文件
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if proxy && config.GlobalConfig.Proxy != "" {
		proxyURL, err := url.Parse(config.GlobalConfig.Proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		result.ErrorMsg = "无法打开链接"
		return result
	}

	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		result.ErrorMsg = "无法打开链接"
		return result
	}
	defer resp.Body.Close()

	for resp.StatusCode == 301 || resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" && strings.HasPrefix(location, "magnet:") {
			result.Content = []byte(location)
			result.IsMagnet = true
			result.ErrorMsg = "获取到磁力链�?
			return result
		}

		req, err = http.NewRequest("GET", location, nil)
		if err != nil {
			result.ErrorMsg = "无法打开链接"
			return result
		}

		if ua != "" {
			req.Header.Set("User-Agent", ua)
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if referer != "" {
			req.Header.Set("Referer", referer)
		}

		resp, err = client.Do(req)
		if err != nil {
			result.ErrorMsg = "无法打开链接"
			return result
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			result.ErrorMsg = "未下载到种子数据"
			return result
		}

		if len(content) == 0 {
			result.ErrorMsg = "未下载到种子数据"
			return result
		}

		// 解析内容格式
		if strings.HasPrefix(string(content), "magnet:") {
			// 磁力链接
			result.Content = content
			result.IsMagnet = true
			result.ErrorMsg = "获取到磁力链�?
			return result
		}

		if bytes.Contains(content, []byte("下载种子文件")) {
			// 首次下载提示页面
			skipFlag := false
			// 这里简化处理，实际应该解析HTML表单并提�?			if !skipFlag {
				result.ErrorMsg = "种子数据有误，请确认链接是否正确，如为PT站点则需手工在站点下载一次种�?
				return result
			}
		}

		// 种子内容
		if len(content) > 0 {
			// 检查是不是种子文件
			folderName, fileList, err := t.GetFileInfoFromTorrentContent(content)
			if err == nil && len(fileList) > 0 {
				// 保存到缓�?				err = cacheBackend.Set(cachePath, content, "torrents")
				if err != nil {
					logger.Errorf("保存种子到缓存失�? %v", err)
				}
				// 成功拿到种子数据
				result.Content = content
				result.FolderName = folderName
				result.FileList = fileList
				return result
			}
			logger.Errorf("种子文件解析失败: %v", err)
			// 种子数据仍然错误
			result.ErrorMsg = "种子数据有误，请确认链接是否正确"
			return result
		}
		// 返回失败
		result.ErrorMsg = ""
		return result
	} else if resp.StatusCode == 429 {
		result.ErrorMsg = "触发站点流控，请稍后重试"
		return result
	} else {
		// 把错误的种子记下来，避免重复使用
		t.AddInvalid(urlStr)
		result.ErrorMsg = fmt.Sprintf("下载种子出错，状态码�?d", resp.StatusCode)
		return result
	}
}

// GetTorrentInfo 从种子文件路径获取种子信�?func (t *TorrentHelper) GetTorrentInfo(torrentPath string) (string, []string, error) {
	if torrentPath == "" {
		return "", []string{}, nil
	}

	mi, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		logger.Errorf("种子文件解析失败: %v", err)
		return "", []string{}, err
	}

	return t.GetFileInfoFromTorrent(mi)
}

// GetFileInfoFromTorrent 从种子文件中获取文件夹名和文件清�?func (t *TorrentHelper) GetFileInfoFromTorrent(mi *metainfo.MetaInfo) (string, []string, error) {
	if mi == nil {
		return "", []string{}, nil
	}

	info, err := mi.UnmarshalInfo()
	if err != nil {
		return "", []string{}, err
	}

	var folderName string
	var fileList []string

	files := info.UpvertedFiles()
	if len(files) == 1 && filepath.Base(files[0].Path) == info.Name {
		// 单文件种子目录名返回�?		folderName = ""
		// 单文件种�?		fileList = []string{info.Name}
	} else {
		// 目录�?		folderName = info.Name
		// 文件清单，如果一级目录与种子名相同则去掉
		fileList = make([]string, 0)
		for _, file := range files {
			filePath := file.Path
			// 根路�?			if len(filePath) > 0 {
				rootPath := filePath[0]
				if rootPath == folderName && len(filePath) > 1 {
					relPath := filepath.Join(filePath[1:]...)
					fileList = append(fileList, relPath)
				} else {
					fullPath := filepath.Join(filePath...)
					fileList = append(fileList, fullPath)
				}
			}
		}
	}

	logger.Debugf("解析种子�?s => 目录�?s，文件清单：%v", info.Name, folderName, fileList)
	return folderName, fileList, nil
}

// GetFileInfoFromTorrentContent 从种子内容中获取文件夹名和文件清�?func (t *TorrentHelper) GetFileInfoFromTorrentContent(torrentContent []byte) (string, []string, error) {
	if len(torrentContent) == 0 {
		return "", []string{}, nil
	}

	// 检查是否为磁力链接
	if utils.IsMagnetLink(string(torrentContent)) {
		return "", []string{}, nil
	}

	// 解析种子内容
	r := bytes.NewReader(torrentContent)
	mi, err := metainfo.Load(r)
	if err != nil {
		logger.Errorf("种子内容解析失败: %v", err)
		return "", []string{}, err
	}

	// 获取文件清单
	return t.GetFileInfoFromTorrent(mi)
}

// GetUrlFilename 从下载请求中获取种子文件�?func (t *TorrentHelper) GetUrlFilename(resp *http.Response, urlStr string) string {
	if resp == nil {
		return ""
	}

	disposition := resp.Header.Get("content-disposition")
	var fileName string

	if disposition != "" {
		re := regexp.MustCompile(`filename="?([^"]+)"?`)
		matches := re.FindStringSubmatch(disposition)
		if len(matches) > 1 {
			decodedName, err := url.QueryUnescape(matches[1])
			if err == nil {
				parts := strings.Split(strings.TrimSpace(decodedName), ";")
				fileName = parts[0]
				if strings.HasSuffix(fileName, "\"") {
					fileName = fileName[:len(fileName)-1]
				}
			}
		}
	} else if urlStr != "" && strings.HasSuffix(urlStr, ".torrent") {
		parts := strings.Split(urlStr, "/")
		fileName = parts[len(parts)-1]
		decodedName, err := url.QueryUnescape(fileName)
		if err == nil {
			fileName = decodedName
		}
	} else {
		fileName = time.Now().String()
	}

	return fileName
}

// TorrentContext 种子上下�?type TorrentContext struct {
	MetaInfo   *utils.MetaBase
	TorrentInfo *types.TorrentInfo
	MediaInfo  *types.MediaInfo
}

// SortTorrents 对种子进行排序：torrent、site、upload、seeder
func (t *TorrentHelper) SortTorrents(torrentList []*TorrentContext) []*TorrentContext {
	if len(torrentList) == 0 {
		return []*TorrentContext{}
	}

	// 下载规则
	sysConfig := db.NewSystemConfigOper()
	priorityRuleInterface := sysConfig.Get(string(types.SystemConfigKeyTorrentsPriority))
	priorityRule := []string{"torrent", "upload", "seeder"}
	if priorityRuleInterface != nil {
		if rules, ok := priorityRuleInterface.([]string); ok {
			priorityRule = rules
		}
	}

	// 站点上传�?	siteOper := db.NewSiteOper()
	userData := siteOper.GetUserDataLatest()
	siteUploads := make(map[string]int64)
	for _, site := range userData {
		siteUploads[site.Name] = site.Upload
	}

	// 定义排序键生成函�?	getSortStr := func(context *TorrentContext) string {
		meta := context.MetaInfo
		torrent := context.TorrentInfo
		media := context.MediaInfo

		// 标题
		title := fmt.Sprintf("%-200s", media.Title)

		// 站点优先�?		siteOrder := fmt.Sprintf("%03d", 999-torrent.SiteOrder)

		// 站点上传�?		upload := fmt.Sprintf("%030d", siteUploads[torrent.SiteName])

		// 资源优先�?		torrentOrder := fmt.Sprintf("%03d", torrent.PriOrder)

		// 资源做种�?		seeders := fmt.Sprintf("%010d", torrent.Seeders)

		// 季集
		var seasonEpisode string
		if len(meta.EpisodeList) == 0 {
			// 无集数的排最前面
			seasonEpisode = fmt.Sprintf("%03d9999", len(meta.SeasonList))
		} else {
			// 集数越多的排越前�?			seasonEpisode = fmt.Sprintf("%03d%04d", len(meta.SeasonList), len(meta.EpisodeList))
		}

		// 根据下载规则的顺序拼装排序字符串
		sortStr := title
		for _, rule := range priorityRule {
			switch rule {
			case "torrent":
				sortStr += torrentOrder
			case "site":
				sortStr += siteOrder
			case "upload":
				sortStr += upload
			case "seeder":
				sortStr += seeders
			}
		}
		sortStr += seasonEpisode
		return sortStr
	}

	// 排序
	sort.Slice(torrentList, func(i, j int) bool {
		return getSortStr(torrentList[i]) > getSortStr(torrentList[j])
	})

	return torrentList
}

// SortGroupTorrents 对媒体信息进行排序、去�?func (t *TorrentHelper) SortGroupTorrents(torrentList []*TorrentContext) []*TorrentContext {
	if len(torrentList) == 0 {
		return []*TorrentContext{}
	}

	// 排序
	torrentList = t.SortTorrents(torrentList)

	// 控重
	result := make([]*TorrentContext, 0)
	added := make([]string, 0)

	// 排序后重新加入数组，按真实名称控重，即只取每个名称的第一�?	for _, context := range torrentList {
		// 控重的主链是名称、年份、季、集
		meta := context.MetaInfo
		media := context.MediaInfo
		var mediaName string

		if media.Type == types.MediaTypeTV {
			mediaName = fmt.Sprintf("%s%s", media.TitleYear, meta.SeasonEpisode)
		} else {
			mediaName = media.TitleYear
		}

		// 检查是否已经添�?		found := false
		for _, name := range added {
			if name == mediaName {
				found = true
				break
			}
		}

		if !found {
			added = append(added, mediaName)
			result = append(result, context)
		}
	}

	return result
}

// GetTorrentEpisodes 从种子的文件清单中获取所有集�?func (t *TorrentHelper) GetTorrentEpisodes(files []string) []int {
	episodes := make([]int, 0)
	mediaExt := config.GlobalConfig.RMTMediaExt

	for _, file := range files {
		if file == "" {
			continue
		}

		filePath := file
		ext := filepath.Ext(filePath)
		if ext == "" {
			continue
		}

		// 转换为小写进行比�?		extLower := strings.ToLower(ext)
		isMedia := false
		for _, validExt := range mediaExt {
			if extLower == strings.ToLower(validExt) {
				isMedia = true
				break
			}
		}

		if !isMedia {
			continue
		}

		// 只使用文件名识别
		filename := filepath.Base(filePath)
		meta := utils.NewMetaInfo(filename)
		if meta.BeginEpisode <= 0 {
			continue
		}

		// 合并集数列表
		for _, ep := range meta.EpisodeList {
			found := false
			for _, existingEp := range episodes {
				if existingEp == ep {
					found = true
					break
				}
			}
			if !found {
				episodes = append(episodes, ep)
			}
		}
	}

	return episodes
}

// IsInvalid 判断种子是否是无效种�?func (t *TorrentHelper) IsInvalid(urlStr string) bool {
	_, exists := t.InvalidTorrents.Get(urlStr)
	return exists
}

// AddInvalid 添加无效种子
func (t *TorrentHelper) AddInvalid(urlStr string) {
	t.InvalidTorrents.Set(urlStr, true)
}

// MatchTorrent 检查种子是否匹配媒体信�?func (t *TorrentHelper) MatchTorrent(mediainfo *types.MediaInfo, torrentMeta *utils.MetaBase, torrent *types.TorrentInfo) bool {
	// 比对词条指定的tmdbid
	if torrentMeta.TMDBID != 0 || torrentMeta.DoubanID != "" {
		if torrentMeta.TMDBID != 0 && torrentMeta.TMDBID == mediainfo.TMDBID {
			logger.Infof("%s 通过词表指定TMDBID匹配到资源：%s - %s", mediainfo.Title, torrent.SiteName, torrent.Title)
			return true
		}
		if torrentMeta.DoubanID != "" && torrentMeta.DoubanID == mediainfo.DoubanID {
			logger.Infof("%s 通过词表指定豆瓣ID匹配到资源：%s - %s", mediainfo.Title, torrent.SiteName, torrent.Title)
			return true
		}
	}

	// 要匹配的媒体标题、原标题
	mediaTitles := make(map[string]bool)
	if mediainfo.Title != "" {
		mediaTitles[utils.ClearUpper(mediainfo.Title)] = true
	}
	if mediainfo.OriginalTitle != "" {
		mediaTitles[utils.ClearUpper(mediainfo.OriginalTitle)] = true
	}

	// 要匹配的媒体别名、译�?	mediaNames := make(map[string]bool)
	for _, name := range mediainfo.Names {
		if name != "" {
			mediaNames[utils.ClearUpper(name)] = true
		}
	}

	// 识别的种子中英文�?	metaNames := make(map[string]bool)
	if torrentMeta.CnName != "" {
		metaNames[utils.ClearUpper(torrentMeta.CnName)] = true
	}
	if torrentMeta.EnName != "" {
		metaNames[utils.ClearUpper(torrentMeta.EnName)] = true
	}

	// 比对种子识别类型
	if torrentMeta.Type == types.MediaTypeTV && mediainfo.Type != types.MediaTypeTV {
		logger.Debugf("%s - %s 种子标题类型�?%s，不匹配 %s", torrent.SiteName, torrent.Title, torrentMeta.Type, mediainfo.Type)
		return false
	}

	// 比对种子在站点中的类�?	if torrent.Category == string(types.MediaTypeTV) && mediainfo.Type != types.MediaTypeTV {
		logger.Debugf("%s - %s 种子在站点中归类�?%s，不匹配 %s", torrent.SiteName, torrent.Title, torrent.Category, mediainfo.Type)
		return false
	}

	// 比对年份
	if mediainfo.Year != "" {
		if mediainfo.Type == types.MediaTypeTV {
			// 剧集年份，每季的年份可能不同，没年份时不比较年份（很多剧集种子不带年份）
			if torrentMeta.Year != "" {
				matchYear := false
				for _, year := range mediainfo.SeasonYears {
					if torrentMeta.Year == strconv.Itoa(year) {
						matchYear = true
						break
					}
				}
				if !matchYear {
					logger.Debugf("%s - %s 年份不匹�?%v", torrent.SiteName, torrent.Title, mediainfo.SeasonYears)
					return false
				}
			}
		} else {
			// 电影年份，上下浮�?年，没年份时不通过
			if torrentMeta.Year == "" {
				logger.Debugf("%s - %s 年份不匹�?%s", torrent.SiteName, torrent.Title, mediainfo.Year)
				return false
			}

			mediainfoYear, _ := strconv.Atoi(mediainfo.Year)
			torrentMetaYear, _ := strconv.Atoi(torrentMeta.Year)

			if torrentMetaYear != mediainfoYear-1 && torrentMetaYear != mediainfoYear && torrentMetaYear != mediainfoYear+1 {
				logger.Debugf("%s - %s 年份不匹�?%s", torrent.SiteName, torrent.Title, mediainfo.Year)
				return false
			}
		}
	}

	// 比对标题和原语种标题
	for title := range metaNames {
		if mediaTitles[title] {
			logger.Infof("%s 通过标题匹配到资源：%s - %s", mediainfo.Title, torrent.SiteName, torrent.Title)
			return true
		}
	}

	// 比对别名和译�?	if len(mediaNames) > 0 {
		for name := range metaNames {
			if mediaNames[name] {
				logger.Infof("%s 通过别名或译名匹配到资源�?s - %s", mediainfo.Title, torrent.SiteName, torrent.Title)
				return true
			}
		}
	}

	// 标题拆分
	if torrentMeta.OrgString != "" {
		// 只拆分出标题中的非英文单词进行匹配，英文单词容易误匹配（带空格的多个单词组合除外�?		re := regexp.MustCompile(`[\s/【�?\[\]\-]+`)
		titles := re.Split(torrentMeta.OrgString, -1)
		clearedTitles := make([]string, 0)
		for _, title := range titles {
			if title != "" && !utils.IsEnglishWord(title) {
				clearedTitles = append(clearedTitles, utils.ClearUpper(title))
			}
		}

		// 在标题中判断是否存在标题、原语种标题
		for title := range mediaTitles {
			for _, clearedTitle := range clearedTitles {
				if title == clearedTitle {
					logger.Infof("%s 通过标题匹配到资源：%s - %s", mediainfo.Title, torrent.SiteName, torrent.Title)
					return true
				}
			}
		}
	}

	// 在副标题中（非英文单词）判断是否存在标题、原语种标题、别名、译�?	if torrent.Description != "" {
		re := regexp.MustCompile(`[\s/【】|]+`)
		subtitles := re.Split(torrent.Description, -1)
		clearedSubtitles := make(map[string]bool)
		for _, subtitle := range subtitles {
			if subtitle != "" && !utils.IsEnglishWord(subtitle) {
				clearedSubtitles[utils.ClearUpper(subtitle)] = true
			}
		}

		match := false
		// 检查媒体标�?		for title := range mediaTitles {
			if clearedSubtitles[title] {
				match = true
				break
			}
		}

		// 检查媒体别�?		if !match {
			for name := range mediaNames {
				if clearedSubtitles[name] {
					match = true
					break
				}
			}
		}

		if match {
			logger.Infof("%s 通过副标题匹配到资源�?s - %s，副标题�?s", mediainfo.Title, torrent.SiteName, torrent.Title, torrent.Description)
			return true
		}
	}

	// 未匹�?	logger.Debugf("%s - %s 标题不匹配，识别名称�?v", torrent.SiteName, torrent.Title, metaNames)
	return false
}

// FilterTorrent 检查种子是否匹配订阅过滤规�?func (t *TorrentHelper) FilterTorrent(torrentInfo *types.TorrentInfo, filterParams map[string]string) bool {
	if len(filterParams) == 0 {
		return true
	}

	// 匹配内容
	content := fmt.Sprintf("%s %s %s %s",
		torrentInfo.Title,
		torrentInfo.Description,
		strings.Join(torrentInfo.Labels, " "),
		torrentInfo.VolumeFactor)

	// 包含
	if include, exists := filterParams["include"]; exists && include != "" {
		matched, _ := regexp.MatchString(include, content)
		if !matched {
			logger.Infof("%s 不匹配包含规�?%s", content, include)
			return false
		}
	}

	// 排除
	if exclude, exists := filterParams["exclude"]; exists && exclude != "" {
		matched, _ := regexp.MatchString(exclude, content)
		if matched {
			logger.Infof("%s 匹配排除规则 %s", content, exclude)
			return false
		}
	}

	// 质量
	if quality, exists := filterParams["quality"]; exists && quality != "" {
		matched, _ := regexp.MatchString(quality, torrentInfo.Title)
		if !matched {
			logger.Infof("%s 不匹配质量规�?%s", torrentInfo.Title, quality)
			return false
		}
	}

	// 分辨�?	if resolution, exists := filterParams["resolution"]; exists && resolution != "" {
		matched, _ := regexp.MatchString(resolution, torrentInfo.Title)
		if !matched {
			logger.Infof("%s 不匹配分辨率规则 %s", torrentInfo.Title, resolution)
			return false
		}
	}

	// 特效
	if effect, exists := filterParams["effect"]; exists && effect != "" {
		matched, _ := regexp.MatchString(effect, torrentInfo.Title)
		if !matched {
			logger.Infof("%s 不匹配特效规�?%s", torrentInfo.Title, effect)
			return false
		}
	}

	// 大小
	if sizeRange, exists := filterParams["size"]; exists && sizeRange != "" {
		if strings.Contains(sizeRange, "-") {
			// 区间
			parts := strings.Split(sizeRange, "-")
			if len(parts) == 2 {
				sizeMin, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				sizeMax, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err1 == nil && err2 == nil {
					sizeMin *= 1024 * 1024
					sizeMax *= 1024 * 1024
					if torrentInfo.Size < sizeMin || torrentInfo.Size > sizeMax {
						return false
					}
				}
			}
		} else if strings.HasPrefix(sizeRange, ">") {
			// 大于
			sizeMin, err := strconv.ParseFloat(strings.TrimSpace(sizeRange[1:]), 64)
			if err == nil {
				sizeMin *= 1024 * 1024
				if torrentInfo.Size < sizeMin {
					return false
				}
			}
		} else if strings.HasPrefix(sizeRange, "<") {
			// 小于
			sizeMax, err := strconv.ParseFloat(strings.TrimSpace(sizeRange[1:]), 64)
			if err == nil {
				sizeMax *= 1024 * 1024
				if torrentInfo.Size > sizeMax {
					return false
				}
			}
		}
	}

	return true
}

// MatchSeasonEpisodes 判断种子是否匹配季集�?func (t *TorrentHelper) MatchSeasonEpisodes(torrent *types.TorrentInfo, meta *utils.MetaBase, seasonEpisodes map[int][]int) bool {
	// 匹配�?	seasons := make([]int, 0, len(seasonEpisodes))
	for s := range seasonEpisodes {
		seasons = append(seasons, s)
	}

	// 种子�?	torrentSeasons := meta.SeasonList
	if len(torrentSeasons) == 0 {
		// 按第一季处�?		torrentSeasons = []int{1}
	}

	// 种子�?	torrentEpisodes := meta.EpisodeList

	// 检查种子季是否在过滤季�?	validSeason := false
	for _, ts := range torrentSeasons {
		for _, s := range seasons {
			if ts == s {
				validSeason = true
				break
			}
		}
		if validSeason {
			break
		}
	}

	if !validSeason {
		// 种子季不在过滤季�?		logger.Debugf("种子 %s - %s 包含�?%v 不是需要的�?%v", torrent.SiteName, torrent.Title, torrentSeasons, seasons)
		return false
	}

	if len(torrentEpisodes) == 0 {
		// 整季按匹配处�?		return true
	}

	if len(torrentSeasons) == 1 {
		needEpisodes, exists := seasonEpisodes[torrentSeasons[0]]
		if exists && len(needEpisodes) > 0 {
			// 检查是否有交集
			hasIntersection := false
			for _, te := range torrentEpisodes {
				for _, ne := range needEpisodes {
					if te == ne {
						hasIntersection = true
						break
					}
				}
				if hasIntersection {
					break
				}
			}

			if !hasIntersection {
				// 单季集没有交集的不要
				logger.Debugf("种子 %s - %s �?%v 没有需要的集：%v", torrent.SiteName, torrent.Title, torrentEpisodes, needEpisodes)
				return false
			}
		}
	}

	return true
}

// md5Hash 计算字符串的MD5哈希�?func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
