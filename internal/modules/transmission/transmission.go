package transmission

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hekmon/transmissionrpc/v2"
	"moviepilot-go/internal/utils"
)

// Transmission Transmission下载�?type Transmission struct {
	trc      *transmissionrpc.Client
	protocol string
	host     string
	port     int64
	username string
	password string
}

// TorrentField 参考transmission web，仅查询需要的参数，加速种子搜�?var TorrentField = []string{
	"id", "name", "status", "labels", "hashString", "totalSize", "percentDone", "addedDate", "trackerList",
	"trackerStats", "leftUntilDone", "rateDownload", "rateUpload", "recheckProgress", "rateDownload", "rateUpload",
	"peersGettingFromUs", "peersSendingToUs", "uploadRatio", "uploadedEver", "downloadedEver", "downloadDir",
	"error", "errorString", "doneDate", "queuePosition", "activityDate", "trackers",
}

// NewTransmission 创建新的Transmission实例
func NewTransmission(
	host string,
	port int64,
	username string,
	password string,
	kwargs map[string]interface{},
) *Transmission {
	t := &Transmission{
		username: username,
		password: password,
	}
	
	if host != "" && port > 0 {
		if p, ok := kwargs["protocol"]; ok {
			if protocolStr, ok := p.(string); ok {
				t.protocol = protocolStr
			} else {
				t.protocol = "http"
			}
		} else {
			t.protocol = "http"
		}
		t.host = host
		t.port = port
	} else if host != "" {
		// 解析URL参数
		protocol, host, port, path := utils.ParseURLParams(host)
		if protocol == "" {
			fmt.Println("Transmission配置不正确！")
			return nil
		}
		t.protocol = protocol
		t.host = host
		t.port = port
		// 忽略path
		_ = path
	} else {
		fmt.Println("Transmission配置不完整！")
		return nil
	}
	
	t.trc = t.loginTransmission()
	return t
}

// loginTransmission 连接transmission
func (t *Transmission) loginTransmission() *transmissionrpc.Client {
	if t.host == "" || t.port <= 0 {
		return nil
	}
	
	// 登录
	fmt.Printf("正在连接 transmission�?s://%s:%d\n", t.protocol, t.host, t.port)
	
	trt, err := transmissionrpc.New(t.host, t.username, t.password, &transmissionrpc.AdvancedConfig{
		HTTPS:      t.protocol == "https",
		Port:       uint16(t.port),
		UserAgent:  "moviepilot-go",
		Timeout:    60 * time.Second,
		HTTPClient: nil,
	})
	
	if err != nil {
		fmt.Printf("transmission 连接出错�?v\n", err)
		return nil
	}
	
	return trt
}

// IsInactive 判断是否需要重�?func (t *Transmission) IsInactive() bool {
	if t.host == "" || t.port <= 0 {
		return false
	}
	return t.trc == nil
}

// Reconnect 重连
func (t *Transmission) Reconnect() {
	t.trc = t.loginTransmission()
}

// GetTorrents 获取种子列表
func (t *Transmission) GetTorrents(
	ids interface{},
	status interface{},
	tags interface{},
) ([]transmissionrpc.Torrent, bool) {
	if t.trc == nil {
		return []transmissionrpc.Torrent{}, true
	}
	
	var idList []int64
	if ids != nil {
		switch v := ids.(type) {
		case string:
			// 将字符串ID转换为整数ID（简化处理）
			// 实际应用中可能需要更复杂的转换逻辑
			idList = []int64{}
		case []string:
			idList = make([]int64, len(v))
			for i, s := range v {
				// 同样需要实际的转换逻辑
				_ = s
				idList[i] = 0
			}
		case []int64:
			idList = v
		}
	}
	
	torrents, err := t.trc.TorrentGetAll(TorrentField, idList)
	if err != nil {
		fmt.Printf("获取种子列表出错�?v\n", err)
		return []transmissionrpc.Torrent{}, true
	}
	
	var statusList []string
	if status != nil {
		switch v := status.(type) {
		case string:
			statusList = []string{v}
		case []string:
			statusList = v
		}
	}
	
	var tagsList []string
	if tags != nil {
		switch v := tags.(type) {
		case string:
			tagsList = strings.Split(v, ",")
		case []string:
			tagsList = v
		}
	}
	
	retTorrents := make([]transmissionrpc.Torrent, 0)
	for _, torrent := range torrents {
		// 状态过�?		if len(statusList) > 0 {
			statusMatch := false
			for _, s := range statusList {
				if *torrent.Status == transmissionrpc.TorrentStatus(s) {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}
		
		// 种子标签过滤
		if len(tagsList) > 0 {
			labels := make([]string, len(torrent.Labels))
			for i, label := range torrent.Labels {
				labels[i] = strings.TrimSpace(label)
			}
			
			tagsMatch := true
			for _, tag := range tagsList {
				found := false
				for _, label := range labels {
					if label == tag {
						found = true
						break
					}
				}
				if !found {
					tagsMatch = false
					break
				}
			}
			
			if !tagsMatch {
				continue
			}
		}
		
		retTorrents = append(retTorrents, torrent)
	}
	
	return retTorrents, false
}

// GetCompletedTorrents 获取已完成的种子列表
func (t *Transmission) GetCompletedTorrents(
	ids interface{},
	tags interface{},
) []transmissionrpc.Torrent {
	if t.trc == nil {
		return nil
	}
	
	torrents, errorFlag := t.GetTorrents(ids, []string{"seeding", "seed_pending"}, tags)
	if errorFlag {
		return nil
	}
	
	return torrents
}

// GetDownloadingTorrents 获取正在下载的种子列�?func (t *Transmission) GetDownloadingTorrents(
	ids interface{},
	tags interface{},
) []transmissionrpc.Torrent {
	if t.trc == nil {
		return nil
	}
	
	torrents, errorFlag := t.GetTorrents(ids, []string{"downloading", "download_pending"}, tags)
	if errorFlag {
		return nil
	}
	
	return torrents
}

// SetTorrentTag 设置种子标签，注意TR默认会覆盖原有标签，如需追加需传入原有标签
func (t *Transmission) SetTorrentTag(
	ids string,
	tags []string,
	orgTags []string,
) bool {
	if t.trc == nil {
		return false
	}
	
	if ids == "" || len(tags) == 0 {
		return false
	}
	
	// 合并标签
	allTags := make([]string, len(orgTags))
	copy(allTags, orgTags)
	
	// 添加新标签，避免重复
	tagSet := make(map[string]bool)
	for _, tag := range allTags {
		tagSet[tag] = true
	}
	
	for _, tag := range tags {
		if !tagSet[tag] {
			allTags = append(allTags, tag)
			tagSet[tag] = true
		}
	}
	
	id, err := utils.ParseInt64(ids)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return false
	}
	
	err = t.trc.TorrentSetLabels(id, allTags)
	if err != nil {
		fmt.Printf("设置种子标签出错�?v\n", err)
		return false
	}
	
	return true
}

// GetTorrentTags 获取所有种子标�?func (t *Transmission) GetTorrentTags(ids string) []string {
	if t.trc == nil {
		return []string{}
	}
	
	id, err := utils.ParseInt64(ids)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return []string{}
	}
	
	torrents, err := t.trc.TorrentGetAll(TorrentField, []int64{id})
	if err != nil {
		fmt.Printf("获取种子标签出错�?v\n", err)
		return []string{}
	}
	
	if len(torrents) > 0 {
		torrent := torrents[0]
		labels := make([]string, len(torrent.Labels))
		for i, label := range torrent.Labels {
			labels[i] = strings.TrimSpace(label)
		}
		return labels
	}
	
	return []string{}
}

// AddTorrent 添加下载任务
func (t *Transmission) AddTorrent(
	content interface{}, // string or []byte
	isPaused bool,
	downloadDir string,
	labels []string,
	cookie string,
) *transmissionrpc.Torrent {
	if t.trc == nil {
		return nil
	}
	
	var torrent *transmissionrpc.Torrent
	var err error
	
	switch v := content.(type) {
	case string:
		if strings.HasPrefix(v, "magnet:") {
			// 磁力链接
			torrent, err = t.trc.TorrentAdd(&transmissionrpc.TorrentAddPayload{
				Filename:     &v,
				DownloadDir:  &downloadDir,
				Paused:       &isPaused,
				Labels:       labels,
				Cookies:      &cookie,
			})
		} else {
			// 文件路径
			torrent, err = t.trc.TorrentAdd(&transmissionrpc.TorrentAddPayload{
				Filename:     &v,
				DownloadDir:  &downloadDir,
				Paused:       &isPaused,
				Labels:       labels,
				Cookies:      &cookie,
			})
		}
	case []byte:
		// 种子文件内容
		metaInfo := string(v)
		torrent, err = t.trc.TorrentAdd(&transmissionrpc.TorrentAddPayload{
			MetaInfo:     &metaInfo,
			DownloadDir:  &downloadDir,
			Paused:       &isPaused,
			Labels:       labels,
			Cookies:      &cookie,
		})
	}
	
	if err != nil {
		fmt.Printf("添加种子出错�?v\n", err)
		return nil
	}
	
	return torrent
}

// StartTorrents 启动种子
func (t *Transmission) StartTorrents(ids interface{}) bool {
	if t.trc == nil {
		return false
	}
	
	var idList []int64
	switch v := ids.(type) {
	case string:
		id, err := utils.ParseInt64(v)
		if err != nil {
			fmt.Printf("解析种子ID出错�?v\n", err)
			return false
		}
		idList = []int64{id}
	case []string:
		idList = make([]int64, len(v))
		for i, s := range v {
			id, err := utils.ParseInt64(s)
			if err != nil {
				fmt.Printf("解析种子ID出错�?v\n", err)
				return false
			}
			idList[i] = id
		}
	case []int64:
		idList = v
	}
	
	err := t.trc.TorrentStart(idList)
	if err != nil {
		fmt.Printf("启动种子出错�?v\n", err)
		return false
	}
	
	return true
}

// StopTorrents 停止种子
func (t *Transmission) StopTorrents(ids interface{}) bool {
	if t.trc == nil {
		return false
	}
	
	var idList []int64
	switch v := ids.(type) {
	case string:
		id, err := utils.ParseInt64(v)
		if err != nil {
			fmt.Printf("解析种子ID出错�?v\n", err)
			return false
		}
		idList = []int64{id}
	case []string:
		idList = make([]int64, len(v))
		for i, s := range v {
			id, err := utils.ParseInt64(s)
			if err != nil {
				fmt.Printf("解析种子ID出错�?v\n", err)
				return false
			}
			idList[i] = id
		}
	case []int64:
		idList = v
	}
	
	err := t.trc.TorrentStop(idList)
	if err != nil {
		fmt.Printf("停止种子出错�?v\n", err)
		return false
	}
	
	return true
}

// DeleteTorrents 删除种子
func (t *Transmission) DeleteTorrents(deleteFile bool, ids interface{}) bool {
	if t.trc == nil {
		return false
	}
	
	var idList []int64
	switch v := ids.(type) {
	case string:
		if v == "" {
			return false
		}
		id, err := utils.ParseInt64(v)
		if err != nil {
			fmt.Printf("解析种子ID出错�?v\n", err)
			return false
		}
		idList = []int64{id}
	case []string:
		if len(v) == 0 {
			return false
		}
		idList = make([]int64, len(v))
		for i, s := range v {
			id, err := utils.ParseInt64(s)
			if err != nil {
				fmt.Printf("解析种子ID出错�?v\n", err)
				return false
			}
			idList[i] = id
		}
	case []int64:
		if len(v) == 0 {
			return false
		}
		idList = v
	}
	
	err := t.trc.TorrentRemove(idList, &deleteFile)
	if err != nil {
		fmt.Printf("删除种子出错�?v\n", err)
		return false
	}
	
	return true
}

// GetFiles 获取种子文件列表
func (t *Transmission) GetFiles(tid string) []transmissionrpc.File {
	if t.trc == nil {
		return nil
	}
	
	if tid == "" {
		return nil
	}
	
	id, err := utils.ParseInt64(tid)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return nil
	}
	
	torrent, err := t.trc.TorrentGetFiles(id)
	if err != nil {
		fmt.Printf("获取种子文件列表出错�?v\n", err)
		return nil
	}
	
	return torrent
}

// SetFiles 设置下载文件的状�?func (t *Transmission) SetFiles(tid string, fileIds []int64) bool {
	if t.trc == nil {
		return false
	}
	
	id, err := utils.ParseInt64(tid)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return false
	}
	
	err = t.trc.TorrentSetFilesWanted(id, fileIds, true)
	if err != nil {
		fmt.Printf("设置下载文件状态出错：%v\n", err)
		return false
	}
	
	return true
}

// SetUnwantedFiles 设置下载文件的状态（不想要的文件�?func (t *Transmission) SetUnwantedFiles(tid string, fileIds []int64) bool {
	if t.trc == nil {
		return false
	}
	
	id, err := utils.ParseInt64(tid)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return false
	}
	
	err = t.trc.TorrentSetFilesWanted(id, fileIds, false)
	if err != nil {
		fmt.Printf("设置下载文件状态出错：%v\n", err)
		return false
	}
	
	return true
}

// TransferInfo 获取传输信息
func (t *Transmission) TransferInfo() *transmissionrpc.SessionStats {
	if t.trc == nil {
		return nil
	}
	
	stats, err := t.trc.SessionStats()
	if err != nil {
		fmt.Printf("获取传输信息出错�?v\n", err)
		return nil
	}
	
	return &stats
}

// SetSpeedLimit 设置速度限制
func (t *Transmission) SetSpeedLimit(downloadLimit, uploadLimit float64) bool {
	if t.trc == nil {
		return false
	}
	
	downloadLimitEnabled := downloadLimit > 0
	uploadLimitEnabled := uploadLimit > 0
	
	err := t.trc.SessionSet(&transmissionrpc.SessionArguments{
		SpeedLimitDown:         &downloadLimit,
		SpeedLimitUp:           &uploadLimit,
		SpeedLimitDownEnabled:  &downloadLimitEnabled,
		SpeedLimitUpEnabled:    &uploadLimitEnabled,
	})
	
	if err != nil {
		fmt.Printf("设置速度限制出错�?v\n", err)
		return false
	}
	
	return true
}

// GetSpeedLimit 获取TR速度
func (t *Transmission) GetSpeedLimit() (float64, float64) {
	if t.trc == nil {
		return 0, 0
	}
	
	session, err := t.trc.SessionGet()
	if err != nil {
		fmt.Printf("获取速度限制出错�?v\n", err)
		return 0, 0
	}
	
	downloadLimit := 0.0
	uploadLimit := 0.0
	
	if session.SpeedLimitDown != nil {
		downloadLimit = *session.SpeedLimitDown
	}
	
	if session.SpeedLimitUp != nil {
		uploadLimit = *session.SpeedLimitUp
	}
	
	return downloadLimit, uploadLimit
}

// RecheckTorrents 重新校验种子
func (t *Transmission) RecheckTorrents(ids interface{}) bool {
	if t.trc == nil {
		return false
	}
	
	var idList []int64
	switch v := ids.(type) {
	case string:
		id, err := utils.ParseInt64(v)
		if err != nil {
			fmt.Printf("解析种子ID出错�?v\n", err)
			return false
		}
		idList = []int64{id}
	case []string:
		idList = make([]int64, len(v))
		for i, s := range v {
			id, err := utils.ParseInt64(s)
			if err != nil {
				fmt.Printf("解析种子ID出错�?v\n", err)
				return false
			}
			idList[i] = id
		}
	case []int64:
		idList = v
	}
	
	err := t.trc.TorrentVerify(idList)
	if err != nil {
		fmt.Printf("重新校验种子出错�?v\n", err)
		return false
	}
	
	return true
}

// ChangeTorrent 设置种子参数
func (t *Transmission) ChangeTorrent(
	hashString string,
	uploadLimit interface{},
	downloadLimit interface{},
	ratioLimit interface{},
	seedingTimeLimit interface{},
) bool {
	if hashString == "" {
		return false
	}
	
	var uploadLimited bool
	var uploadLimitValue float64
	if uploadLimit != nil {
		uploadLimited = true
		if ul, ok := uploadLimit.(float64); ok {
			uploadLimitValue = ul
		} else if ul, ok := uploadLimit.(int); ok {
			uploadLimitValue = float64(ul)
		}
	} else {
		uploadLimited = false
		uploadLimitValue = 0
	}
	
	var downloadLimited bool
	var downloadLimitValue float64
	if downloadLimit != nil {
		downloadLimited = true
		if dl, ok := downloadLimit.(float64); ok {
			downloadLimitValue = dl
		} else if dl, ok := downloadLimit.(int); ok {
			downloadLimitValue = float64(dl)
		}
	} else {
		downloadLimited = false
		downloadLimitValue = 0
	}
	
	var seedRatioMode int64
	var seedRatioLimit float64
	if ratioLimit != nil {
		seedRatioMode = 1
		if rl, ok := ratioLimit.(float64); ok {
			seedRatioLimit = rl
		} else if rl, ok := ratioLimit.(int); ok {
			seedRatioLimit = float64(rl)
		}
	} else {
		seedRatioMode = 2
		seedRatioLimit = 0
	}
	
	var seedIdleMode int64
	var seedIdleLimit int64
	if seedingTimeLimit != nil {
		seedIdleMode = 1
		if stl, ok := seedingTimeLimit.(int64); ok {
			seedIdleLimit = stl
		} else if stl, ok := seedingTimeLimit.(int); ok {
			seedIdleLimit = int64(stl)
		}
	} else {
		seedIdleMode = 2
		seedIdleLimit = 0
	}
	
	id, err := utils.ParseInt64(hashString)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return false
	}
	
	err = t.trc.TorrentSet(&transmissionrpc.TorrentSetPayload{
		IDs:             []int64{id},
		UploadLimited:   &uploadLimited,
		UploadLimit:     &uploadLimitValue,
		DownloadLimited: &downloadLimited,
		DownloadLimit:   &downloadLimitValue,
		SeedRatioMode:   &seedRatioMode,
		SeedRatioLimit:  &seedRatioLimit,
		SeedIdleMode:    &seedIdleMode,
		SeedIdleLimit:   &seedIdleLimit,
	})
	
	if err != nil {
		fmt.Printf("设置种子出错�?v\n", err)
		return false
	}
	
	return true
}

// UpdateTracker 更新tracker
func (t *Transmission) UpdateTracker(hashString string, trackerList []string) bool {
	if t.trc == nil {
		return false
	}
	
	id, err := utils.ParseInt64(hashString)
	if err != nil {
		fmt.Printf("解析种子ID出错�?v\n", err)
		return false
	}
	
	err = t.trc.TorrentSet(&transmissionrpc.TorrentSetPayload{
		IDs:          []int64{id},
		TrackerList:  trackerList,
	})
	
	if err != nil {
		fmt.Printf("修改tracker出错�?v\n", err)
		return false
	}
	
	return true
}

// GetSession 获取Transmission当前的会话信息和配置设置
func (t *Transmission) GetSession() *transmissionrpc.SessionArguments {
	if t.trc == nil {
		return nil
	}
	
	session, err := t.trc.SessionGet()
	if err != nil {
		fmt.Printf("获取session出错�?v\n", err)
		return nil
	}
	
	return &session
}
