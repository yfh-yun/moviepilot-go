package chain

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// TransferChain 文件整理处理链
type TransferChain struct {
	cache           *cache.Cache
	logger          *logger.Logger
	transferRepo    *repository.TransferHistoryRepository
	mediaService    *service.MediaService
	transferService *service.TransferService
}

// NewTransferChain 创建文件整理处理链实例
func NewTransferChain(cache *cache.Cache, logger *logger.Logger, transferRepo *repository.TransferHistoryRepository, mediaService *service.MediaService) *TransferChain {
	return &TransferChain{
		cache:           cache,
		logger:          logger,
		transferRepo:    transferRepo,
		mediaService:    mediaService,
		transferService: service.NewTransferService(transferRepo, mediaService, logger),
	}
}

// ProcessTransfer 处理文件整理
func (c *TransferChain) ProcessTransfer(ctx context.Context, task model.TransferTask) (*model.TransferInfo, error) {
	c.logger.Info("处理文件整理", "filePath", task.FileItem.Path)

	// 验证任务参数
	if err := c.validateTransferTask(task); err != nil {
		return nil, err
	}

	// 识别媒体信息（如果未提供）
	if task.MediaInfo == nil {
		mediaInfo, err := c.identifyMedia(ctx, task)
		if err != nil {
			return nil, err
		}
		task.MediaInfo = mediaInfo
	}

	// 执行整理
	transferInfo, err := c.transferService.ProcessTransfer(ctx, task)
	if err != nil {
		c.logger.Error("文件整理失败", "error", err)
		return nil, err
	}

	c.logger.Info("文件整理完成", "sourcePath", task.FileItem.Path, "targetPath", transferInfo.TargetPath)
	return transferInfo, nil
}

// validateTransferTask 验证整理任务参数
func (c *TransferChain) validateTransferTask(task model.TransferTask) error {
	if task.FileItem.Path == "" {
		return errors.New("文件路径不能为空")
	}

	if task.TransferType == "" {
		task.TransferType = "copy" // 默认复制模式
	}

	return nil
}

// identifyMedia 识别媒体信息
func (c *TransferChain) identifyMedia(ctx context.Context, task model.TransferTask) (*model.MediaInfo, error) {
	c.logger.Info("识别媒体信息", "fileName", filepath.Base(task.FileItem.Path))

	// 从文件名提取元数据
	metaInfo := c.extractMetaFromFileName(task.FileItem.Path)

	// 识别媒体信息
	mediaInfo, err := c.mediaService.IdentifyMedia(ctx, metaInfo)
	if err != nil {
		return nil, err
	}

	if mediaInfo == nil {
		return nil, errors.New("无法识别媒体信息")
	}

	c.logger.Info("媒体信息识别成功", "title", mediaInfo.Title)
	return mediaInfo, nil
}

// extractMetaFromFileName 从文件名提取元数据
func (c *TransferChain) extractMetaFromFileName(filePath string) *model.MetaInfo {
	fileName := filepath.Base(filePath)
	
	// 使用正则表达式解析文件名
	metaInfo := &model.MetaInfo{}
	
	// 基础模式匹配
	patterns := []string{
		// S01E01, S1E1格式
		`^(.*?)[. _\-]+S(\d{1,2})E(\d{1,2})([. _\-].*)?$`,
		// S01, S1格式（整季）
		`^(.*?)[. _\-]+S(\d{1,2})([. _\-].*)?$`,
		// E01, E1格式（仅集数）
		`^(.*?)[. _\-]+E(\d{1,2})([. _\-].*)?$`,
		// 01x02格式
		`^(.*?)[. _\-]+(\d{1,2})x(\d{1,2})([. _\-].*)?$`,
		// 年份格式
		`^(.*?)[. _\-]+(\d{4})([. _\-].*)?$`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(fileName)
		if len(matches) > 0 {
			// 提取标题（去除特殊字符）
			title := regexp.MustCompile(`[._\-]+`).ReplaceAllString(matches[1], " ")
			title = strings.TrimSpace(title)
			metaInfo.Title = title
			
			// 提取季集信息
			if len(matches) > 2 && matches[2] != "" {
				if season, err := strconv.Atoi(matches[2]); err == nil {
					metaInfo.Season = season
				}
			}
			
			if len(matches) > 3 && matches[3] != "" {
				if episode, err := strconv.Atoi(matches[3]); err == nil {
					metaInfo.Episode = episode
				} else if strings.Contains(matches[3], "E") {
					// 处理E01格式
					episodeStr := regexp.MustCompile(`E(\d+)`).FindStringSubmatch(matches[3])
					if len(episodeStr) > 1 {
						if episode, err := strconv.Atoi(episodeStr[1]); err == nil {
							metaInfo.Episode = episode
						}
					}
				}
			}
			
			break
		}
	}
	
	// 如果没有匹配到标题，使用文件名作为标题
	if metaInfo.Title == "" {
		title := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		title = regexp.MustCompile(`[._\-]+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)
		metaInfo.Title = title
	}
	
	return metaInfo
}

// RecommendName 获取重命名后的名称
func (c *TransferChain) RecommendName(meta *model.MetaInfo, mediaInfo *model.MediaInfo) (string, error) {
	c.logger.Info("生成推荐名称", "mediaTitle", mediaInfo.Title, "originalTitle", meta.Title)
	
	// 构建目录结构
	var pathParts []string
	
	// 添加媒体根目录（根据类型）
	switch mediaInfo.Type {
	case "movie":
		pathParts = append(pathParts, "Movies")
	case "tv":
		pathParts = append(pathParts, "TV Shows")
	default:
		pathParts = append(pathParts, "Media")
	}
	
	// 添加年份标题
	yearTitle := fmt.Sprintf("%s %d", mediaInfo.Title, mediaInfo.Year)
	pathParts = append(pathParts, yearTitle)
	
	// 根据媒体类型构建文件名
	var fileName string
	
	switch mediaInfo.Type {
	case "movie":
		// 电影格式：标题 (年份) [质量]
		fileName = fmt.Sprintf("%s (%d)", mediaInfo.Title, mediaInfo.Year)
		if meta.Resolution != "" {
			fileName += fmt.Sprintf(" [%s]", meta.Resolution)
		}
		if meta.Source != "" {
			fileName += fmt.Sprintf(" [%s]", meta.Source)
		}
		
	case "tv":
		// 剧集目录格式：标题 (年份)
		showDir := fmt.Sprintf("%s (%d)", mediaInfo.Title, mediaInfo.Year)
		
		// 剧集文件格式：标题 (年份) - S01E01
		seasonStr := ""
		if meta.Season > 0 {
			seasonStr = fmt.Sprintf("S%02d", meta.Season)
		}
		
		episodeStr := ""
		if meta.Episode > 0 {
			episodeStr = fmt.Sprintf("E%02d", meta.Episode)
		}
		
		if seasonStr != "" && episodeStr != "" {
			fileName = fmt.Sprintf("%s (%d) - %s%s", mediaInfo.Title, mediaInfo.Year, seasonStr, episodeStr)
		} else if seasonStr != "" {
			fileName = fmt.Sprintf("%s (%d) - %s", mediaInfo.Title, mediaInfo.Year, seasonStr)
		} else {
			fileName = fmt.Sprintf("%s (%d)", mediaInfo.Title, mediaInfo.Year)
		}
		
		// 对于剧集，文件应该在季目录下
		if meta.Season > 0 {
			pathParts = append(pathParts, showDir)
			seasonDir := fmt.Sprintf("Season %02d", meta.Season)
			pathParts = append(pathParts, seasonDir)
		} else {
			pathParts = append(pathParts, showDir)
		}
		
	default:
		fileName = meta.Title
	}
	
	// 添加扩展名（如果原始文件有）
	ext := filepath.Ext(meta.FileName)
	if ext != "" {
		fileName += ext
	}
	
	// 构建完整路径
	var fullPath string
	if len(pathParts) > 0 {
		fullPath = filepath.Join(pathParts...)
		if mediaInfo.Type != "tv" || meta.Season == 0 {
			// 对于电影和没有季信息的剧集，文件在目录下
			fullPath = filepath.Join(fullPath, fileName)
		} else {
			// 对于有季信息的剧集，已经在季目录下，直接使用文件名
			fullPath = filepath.Join(fullPath, fileName)
		}
	} else {
		fullPath = fileName
	}
	
	// 根据请求类型返回路径或文件名
	if meta.Type == "dir" {
		// 返回目录路径
		if mediaInfo.Type == "tv" && meta.Season > 0 {
			// 返回季目录
			return filepath.Dir(fullPath), nil
		}
		// 返回媒体目录
		return filepath.Dir(fullPath), nil
	}
	
	// 返回完整路径
	return fullPath, nil
}

// GetRecommendName 获取推荐名称（用于API接口）
func (c *TransferChain) GetRecommendName(ctx context.Context, filePath string, fileType string) (*RecommendNameResult, error) {
	// 从文件路径提取元数据
	metaInfo := c.extractMetaFromFileName(filePath)
	metaInfo.FileName = filepath.Base(filePath)
	metaInfo.Type = fileType
	
	// 识别媒体信息
	mediaInfo, err := c.mediaService.IdentifyMedia(ctx, metaInfo)
	if err != nil {
		return nil, fmt.Errorf("识别媒体信息失败: %w", err)
	}
	
	if mediaInfo == nil {
		return nil, fmt.Errorf("未识别到媒体信息")
	}
	
	// 生成推荐名称
	recommendPath, err := c.RecommendName(metaInfo, mediaInfo)
	if err != nil {
		return nil, fmt.Errorf("生成推荐名称失败: %w", err)
	}
	
	// 根据文件类型返回结果
	var newName string
	if fileType == "dir" {
		// 返回目录名
		if mediaInfo.Type == "tv" && metaInfo.Season > 0 {
			// 返回季目录名
			parts := strings.Split(recommendPath, string(filepath.Separator))
			if len(parts) >= 2 {
				newName = parts[len(parts)-2] // 倒数第二个是季目录
			} else {
				newName = filepath.Base(recommendPath)
			}
		} else {
			// 返回媒体目录名
			newName = filepath.Base(filepath.Dir(recommendPath))
		}
	} else {
		// 返回文件名
		newName = filepath.Base(recommendPath)
	}
	
	return &RecommendNameResult{
		Success:     true,
		Name:        newName,
		FullPath:    recommendPath,
		MediaInfo:   mediaInfo,
		MetaInfo:    metaInfo,
	}, nil
}

// RecommendNameResult 推荐名称结果
type RecommendNameResult struct {
	Success   bool               `json:"success"`
	Name      string             `json:"name"`
	FullPath  string             `json:"full_path,omitempty"`
	MediaInfo *model.MediaInfo   `json:"media_info,omitempty"`
	MetaInfo  *model.MetaInfo    `json:"meta_info,omitempty"`
	Error     string             `json:"error,omitempty"`
}

	// 这里可以实现复杂的文件名解析逻辑
	// 暂时简单实现，实际应该使用正则表达式等

	return &model.MetaInfo{
		FileName: fileName,
		// 更多元数据提取逻辑...
	}
}

// GetTransferHistory 获取整理历史记录
func (c *TransferChain) GetTransferHistory(ctx context.Context, page, pageSize int) ([]*model.TransferHistory, int64, error) {
	c.logger.Info("获取整理历史记录", "page", page, "pageSize", pageSize)

	history, total, err := c.transferRepo.GetTransferHistory(ctx, page, pageSize)
	if err != nil {
		c.logger.Error("获取整理历史记录失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("获取整理历史记录成功", "count", len(history))
	return history, total, nil
}

// GetTransferHistoryBySource 根据源文件路径获取整理历史
func (c *TransferChain) GetTransferHistoryBySource(ctx context.Context, sourcePath string) (*model.TransferHistory, error) {
	c.logger.Info("根据源文件路径查询整理历史", "sourcePath", sourcePath)

	history, err := c.transferRepo.GetTransferHistoryBySource(ctx, sourcePath)
	if err != nil {
		c.logger.Error("查询整理历史失败", "error", err)
		return nil, err
	}

	return history, nil
}

// GetTransferHistoryByMedia 根据媒体信息获取整理历史
func (c *TransferChain) GetTransferHistoryByMedia(ctx context.Context, tmdbID int, mediaType model.MediaType) ([]*model.TransferHistory, error) {
	c.logger.Info("根据媒体信息查询整理历史", "tmdbID", tmdbID, "mediaType", mediaType)

	history, err := c.transferRepo.GetTransferHistoryByMedia(ctx, tmdbID, mediaType)
	if err != nil {
		c.logger.Error("查询整理历史失败", "error", err)
		return nil, err
	}

	c.logger.Info("查询整理历史成功", "count", len(history))
	return history, nil
}

// RedoTransfer 重新整理
func (c *TransferChain) RedoTransfer(ctx context.Context, historyID int64, newMediaID string, mediaType model.MediaType) (*model.TransferInfo, error) {
	c.logger.Info("重新整理", "historyID", historyID, "newMediaID", newMediaID)

	// 获取历史记录
	history, err := c.transferRepo.GetTransferHistoryByID(ctx, historyID)
	if err != nil {
		return nil, err
	}

	if history == nil {
		return nil, errors.New("整理记录不存在")
	}

	// 创建新的整理任务
	task := model.TransferTask{
		FileItem: model.FileItem{
			Path: history.SourcePath,
		},
		TransferType: history.Mode,
	}

	// 如果提供了新的媒体ID，使用新的媒体信息
	if newMediaID != "" {
		// 查询新的媒体信息
		mediaInfo, err := c.mediaService.GetMediaInfoByID(ctx, newMediaID, mediaType)
		if err != nil {
			return nil, err
		}
		task.MediaInfo = mediaInfo
	}

	// 执行整理
	return c.ProcessTransfer(ctx, task)
}

// GetTransferStats 获取整理统计信息
func (c *TransferChain) GetTransferStats(ctx context.Context) (*model.TransferStats, error) {
	c.logger.Info("获取整理统计信息")

	stats, err := c.transferRepo.GetTransferStats(ctx)
	if err != nil {
		c.logger.Error("获取整理统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// CleanTransferHistory 清理整理历史记录
func (c *TransferChain) CleanTransferHistory(ctx context.Context, days int) error {
	c.logger.Info("清理整理历史记录", "days", days)

	err := c.transferRepo.CleanTransferHistory(ctx, days)
	if err != nil {
		c.logger.Error("清理整理历史记录失败", "error", err)
		return err
	}

	c.logger.Info("清理整理历史记录完成")
	return nil
}
