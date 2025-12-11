package indexer

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Aggregator 聚合搜索器
// 支持同时从多个索引器搜索并聚合结果
type Aggregator struct {
	factory *Factory
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewAggregator 创建聚合搜索器
func NewAggregator(factory *Factory) *Aggregator {
	return &Aggregator{
		factory: factory,
		logger:  logger.GetLogger(),
	}
}

// Search 聚合搜索
// 从所有已注册的索引器并发搜索，合并结果
func (a *Aggregator) Search(ctx context.Context, opts SearchOptions) ([]*Torrent, error) {
	a.mu.RLock()
	clients := a.factory.GetAll()
	a.mu.RUnlock()

	if len(clients) == 0 {
		a.logger.Warn("没有可用的索引器")
		return nil, fmt.Errorf("no indexers available")
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []*Torrent, len(clients))
	errorsChan := make(chan error, len(clients))

	// 并发搜索所有索引器
	for _, client := range clients {
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()

			a.logger.Info("开始搜索",
				zap.String("indexer", c.Name()),
				zap.String("query", opts.Query))

			results, err := c.Search(ctx, opts)
			if err != nil {
				a.logger.Error("搜索失败",
					zap.String("indexer", c.Name()),
					zap.Error(err))
				errorsChan <- fmt.Errorf("indexer %s: %w", c.Name(), err)
				return
			}

			a.logger.Info("搜索完成",
				zap.String("indexer", c.Name()),
				zap.Int("count", len(results)))

			resultsChan <- results
		}(client)
	}

	// 等待所有搜索完成
	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// 收集所有结果
	var allTorrents []*Torrent
	for results := range resultsChan {
		allTorrents = append(allTorrents, results...)
	}

	// 收集错误（仅记录，不影响结果）
	var errs []error
	for err := range errorsChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		a.logger.Warn("部分索引器搜索失败",
			zap.Int("failed_count", len(errs)),
			zap.Int("total_count", len(clients)))
	}

	// 去重（基于 Link）
	allTorrents = a.deduplicateTorrents(allTorrents)

	// 排序（按做种数降序）
	a.sortTorrentsBySeeders(allTorrents)

	// 应用限制
	if opts.Limit > 0 && len(allTorrents) > opts.Limit {
		allTorrents = allTorrents[:opts.Limit]
	}

	a.logger.Info("聚合搜索完成",
		zap.Int("total_results", len(allTorrents)),
		zap.Int("indexers_count", len(clients)))

	return allTorrents, nil
}

// SearchByIndexer 从指定索引器搜索
func (a *Aggregator) SearchByIndexer(ctx context.Context, indexerName string, opts SearchOptions) ([]*Torrent, error) {
	client, ok := a.factory.Get(indexerName)
	if !ok {
		return nil, fmt.Errorf("indexer not found: %s", indexerName)
	}

	a.logger.Info("搜索指定索引器",
		zap.String("indexer", indexerName),
		zap.String("query", opts.Query))

	results, err := client.Search(ctx, opts)
	if err != nil {
		a.logger.Error("搜索失败",
			zap.String("indexer", indexerName),
			zap.Error(err))
		return nil, err
	}

	a.logger.Info("搜索完成",
		zap.String("indexer", indexerName),
		zap.Int("count", len(results)))

	return results, nil
}

// TestAllIndexers 测试所有索引器的连接
func (a *Aggregator) TestAllIndexers(ctx context.Context) map[string]error {
	a.mu.RLock()
	clients := a.factory.GetAll()
	a.mu.RUnlock()

	results := make(map[string]error)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()

			err := c.TestConnection(ctx)
			mu.Lock()
			results[c.Name()] = err
			mu.Unlock()

			if err != nil {
				a.logger.Error("索引器连接测试失败",
					zap.String("indexer", c.Name()),
					zap.Error(err))
			} else {
				a.logger.Info("索引器连接测试成功",
					zap.String("indexer", c.Name()))
			}
		}(client)
	}

	wg.Wait()
	return results
}

// GetAllCapabilities 获取所有索引器的能力
func (a *Aggregator) GetAllCapabilities(ctx context.Context) map[string]*Capabilities {
	a.mu.RLock()
	clients := a.factory.GetAll()
	a.mu.RUnlock()

	results := make(map[string]*Capabilities)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()

			caps, err := c.GetCapabilities(ctx)
			if err != nil {
				a.logger.Error("获取索引器能力失败",
					zap.String("indexer", c.Name()),
					zap.Error(err))
				return
			}

			mu.Lock()
			results[c.Name()] = caps
			mu.Unlock()
		}(client)
	}

	wg.Wait()
	return results
}

// deduplicateTorrents 去重种子（基于 Link）
func (a *Aggregator) deduplicateTorrents(torrents []*Torrent) []*Torrent {
	seen := make(map[string]bool)
	unique := make([]*Torrent, 0, len(torrents))

	for _, t := range torrents {
		if t.Link == "" {
			continue
		}
		if !seen[t.Link] {
			seen[t.Link] = true
			unique = append(unique, t)
		}
	}

	return unique
}

// sortTorrentsBySeeders 按做种数降序排序
func (a *Aggregator) sortTorrentsBySeeders(torrents []*Torrent) {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Seeders > torrents[j].Seeders
	})
}

// FilterTorrents 过滤种子
func (a *Aggregator) FilterTorrents(torrents []*Torrent, filter func(*Torrent) bool) []*Torrent {
	filtered := make([]*Torrent, 0, len(torrents))
	for _, t := range torrents {
		if filter(t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
