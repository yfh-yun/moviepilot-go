package download

import (
	"context"
	"sort"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/dto"
)

// BatchDownloader 批量下载器
type BatchDownloader struct {
	downloadService *DownloadService
	logger          *zap.Logger
}

// NewBatchDownloader 创建批量下载器
func NewBatchDownloader(downloadService *DownloadService, logger *zap.Logger) *BatchDownloader {
	return &BatchDownloader{
		downloadService: downloadService,
		logger:          logger,
	}
}

// BatchDownload 根据缺失数据，从种子列表中组合择优下载
// 返回：已经下载的资源列表、剩余未下载到的剧集
func (b *BatchDownloader) BatchDownload(ctx context.Context, torrentList []*dto.Context, needEpisodes []int, downloader string) ([]*dto.Context, []int, error) {
	b.logger.Info("开始批量下载",
		zap.Int("torrent_count", len(torrentList)),
		zap.Ints("need_episodes", needEpisodes),
	)

	if len(torrentList) == 0 || len(needEpisodes) == 0 {
		b.logger.Info("没有需要下载的内容")
		return []*dto.Context{}, needEpisodes, nil
	}

	// 已下载的资源列表
	downloaded := make([]*dto.Context, 0)
	// 剩余未下载的剧集
	remaining := make([]int, len(needEpisodes))
	copy(remaining, needEpisodes)

	// 按优先级排序种子列表
	sortedTorrents := b.sortTorrentsByPriority(torrentList)

	// 遍历种子，尝试匹配需要的剧集
	for _, torrent := range sortedTorrents {
		if len(remaining) == 0 {
			break
		}

		// 获取种子包含的剧集
		torrentEpisodes := b.getTorrentEpisodes(torrent)
		if len(torrentEpisodes) == 0 {
			continue
		}

		// 检查是否有需要的剧集
		matchedEpisodes := b.findMatchedEpisodes(torrentEpisodes, remaining)
		if len(matchedEpisodes) == 0 {
			continue
		}

		b.logger.Info("找到匹配的种子",
			zap.String("title", torrent.TorrentInfo.Title),
			zap.Ints("matched_episodes", matchedEpisodes),
		)

		// 下载种子
		err := b.downloadService.DownloadSingleLegacy(ctx, torrent, "", nil, matchedEpisodes, downloader)
		if err != nil {
			b.logger.Error("下载种子失败",
				zap.String("title", torrent.TorrentInfo.Title),
				zap.Error(err),
			)
			continue
		}

		// 添加到已下载列表
		downloaded = append(downloaded, torrent)

		// 从剩余列表中移除已下载的剧集
		remaining = b.removeEpisodes(remaining, matchedEpisodes)

		b.logger.Info("种子下载成功",
			zap.String("title", torrent.TorrentInfo.Title),
			zap.Ints("remaining_episodes", remaining),
		)
	}

	b.logger.Info("批量下载完成",
		zap.Int("downloaded_count", len(downloaded)),
		zap.Int("remaining_count", len(remaining)),
	)

	return downloaded, remaining, nil
}

// sortTorrentsByPriority 按优先级排序种子
func (b *BatchDownloader) sortTorrentsByPriority(torrents []*dto.Context) []*dto.Context {
	sorted := make([]*dto.Context, len(torrents))
	copy(sorted, torrents)

	sort.Slice(sorted, func(i, j int) bool {
		// 优先级规则：
		// 1. 站点优先级
		// 2. 做种数
		// 3. 大小（越小越好）

		ti, tj := sorted[i].TorrentInfo, sorted[j].TorrentInfo

		// 站点优先级
		if ti.SiteOrder != tj.SiteOrder {
			return ti.SiteOrder < tj.SiteOrder
		}

		// 做种数
		if ti.Seeders != tj.Seeders {
			return ti.Seeders > tj.Seeders
		}

		// 大小
		return ti.Size < tj.Size
	})

	return sorted
}

// getTorrentEpisodes 获取种子包含的剧集
func (b *BatchDownloader) getTorrentEpisodes(torrent *dto.Context) []int {
	if torrent.MetaInfo == nil {
		return []int{}
	}

	episodes := make([]int, 0)

	// 如果有明确的集列表
	if len(torrent.MetaInfo.EpisodeList) > 0 {
		return torrent.MetaInfo.EpisodeList
	}

	// 如果有开始集和结束集
	if torrent.MetaInfo.BeginEpisode != nil && torrent.MetaInfo.EndEpisode != nil {
		for i := *torrent.MetaInfo.BeginEpisode; i <= *torrent.MetaInfo.EndEpisode; i++ {
			episodes = append(episodes, i)
		}
		return episodes
	}

	// 如果只有开始集
	if torrent.MetaInfo.BeginEpisode != nil {
		episodes = append(episodes, *torrent.MetaInfo.BeginEpisode)
		return episodes
	}

	return episodes
}

// findMatchedEpisodes 查找匹配的剧集
func (b *BatchDownloader) findMatchedEpisodes(torrentEpisodes []int, needEpisodes []int) []int {
	matched := make([]int, 0)

	for _, te := range torrentEpisodes {
		for _, ne := range needEpisodes {
			if te == ne {
				matched = append(matched, te)
				break
			}
		}
	}

	return matched
}

// removeEpisodes 从列表中移除指定的剧集
func (b *BatchDownloader) removeEpisodes(episodes []int, toRemove []int) []int {
	result := make([]int, 0)

	for _, ep := range episodes {
		found := false
		for _, rm := range toRemove {
			if ep == rm {
				found = true
				break
			}
		}
		if !found {
			result = append(result, ep)
		}
	}

	return result
}

// SelectBestTorrent 从种子列表中选择最佳种子
func (b *BatchDownloader) SelectBestTorrent(torrents []*dto.Context) *dto.Context {
	if len(torrents) == 0 {
		return nil
	}

	sorted := b.sortTorrentsByPriority(torrents)
	return sorted[0]
}
