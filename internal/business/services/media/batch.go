package media

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/utils/workerpool"
)

// BatchIdentifier 批量识别器
type BatchIdentifier struct {
	service     Service
	batchSize   int
	concurrency int
	logger      *zap.Logger
}

// BatchConfig 批量配置
type BatchConfig struct {
	BatchSize   int // 每批处理的文件数
	Concurrency int // 并发数
}

// NewBatchIdentifier 创建批量识别器
func NewBatchIdentifier(service Service, config BatchConfig, logger *zap.Logger) *BatchIdentifier {
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 5
	}

	return &BatchIdentifier{
		service:     service,
		batchSize:   config.BatchSize,
		concurrency: config.Concurrency,
		logger:      logger,
	}
}

// IdentifyBatch 批量识别文件
func (b *BatchIdentifier) IdentifyBatch(ctx context.Context, files []FileItem, opts IdentifyOptions) ([]database.Media, error) {
	if len(files) == 0 {
		return []database.Media{}, nil
	}

	if b.logger != nil {
		b.logger.Info("starting batch identification",
			zap.Int("total_files", len(files)),
			zap.Int("batch_size", b.batchSize),
			zap.Int("concurrency", b.concurrency))
	}

	var (
		mu     sync.Mutex
		medias []database.Media
	)

	// 创建 Goroutine 池
	pool := workerpool.New(b.concurrency)
	defer pool.Wait()

	// 分批处理
	for i := 0; i < len(files); i += b.batchSize {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return medias, ctx.Err()
		default:
		}

		end := i + b.batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		batchNum := i/b.batchSize + 1

		// 提交批次任务
		pool.Submit(func() {
			if b.logger != nil {
				b.logger.Debug("processing batch",
					zap.Int("batch_num", batchNum),
					zap.Int("batch_size", len(batch)))
			}

			// 识别当前批次
			results, err := b.service.Identify(batch, opts)
			if err != nil {
				if b.logger != nil {
					b.logger.Warn("batch identification failed",
						zap.Int("batch_num", batchNum),
						zap.Error(err))
				}
				return
			}

			// 添加到结果集
			mu.Lock()
			medias = append(medias, results...)
			mu.Unlock()

			if b.logger != nil {
				b.logger.Debug("batch completed",
					zap.Int("batch_num", batchNum),
					zap.Int("results", len(results)))
			}
		})
	}

	if b.logger != nil {
		b.logger.Info("batch identification completed",
			zap.Int("total_results", len(medias)))
	}

	return medias, nil
}

// IdentifyBatchWithProgress 带进度的批量识别
func (b *BatchIdentifier) IdentifyBatchWithProgress(
	ctx context.Context,
	files []FileItem,
	opts IdentifyOptions,
	progressChan chan<- int,
) ([]database.Media, error) {
	if len(files) == 0 {
		return []database.Media{}, nil
	}

	var (
		mu        sync.Mutex
		medias    []database.Media
		processed int
	)

	pool := workerpool.New(b.concurrency)
	defer pool.Wait()

	totalBatches := (len(files) + b.batchSize - 1) / b.batchSize

	for i := 0; i < len(files); i += b.batchSize {
		select {
		case <-ctx.Done():
			return medias, ctx.Err()
		default:
		}

		end := i + b.batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		batchNum := i/b.batchSize + 1

		pool.Submit(func() {
			results, err := b.service.Identify(batch, opts)
			if err != nil {
				if b.logger != nil {
					b.logger.Warn("batch failed", zap.Int("batch", batchNum), zap.Error(err))
				}
			}

			mu.Lock()
			if err == nil {
				medias = append(medias, results...)
			}
			processed++
			mu.Unlock()

			// 报告进度
			if progressChan != nil {
				select {
				case progressChan <- (processed * 100) / totalBatches:
				default:
				}
			}
		})
	}

	return medias, nil
}
