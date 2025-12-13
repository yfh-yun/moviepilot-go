// Package services MoviePilot业务服务层
package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/logger"
)

// MediaScanner 媒体库扫描服务接口
type MediaScanner interface {
	// ScanLibrary 扫描媒体库
	ScanLibrary(ctx context.Context, paths []string, userID uint) error
	// ScanSingleFile 扫描单个文件
	ScanSingleFile(ctx context.Context, filePath string, userID uint) (*models.MediaFile, error)
	// UpdateLibrary 更新媒体库
	UpdateLibrary(ctx context.Context, userID uint) error
	// IdentifyFile 识别媒体文件
	IdentifyFile(ctx context.Context, file *models.MediaFile) (*models.MediaItem, *models.MediaVersion, error)
}

// MediaScannerImpl 媒体库扫描服务实现
type MediaScannerImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMediaScanner 创建媒体库扫描服务实例
func NewMediaScanner(db *gorm.DB) MediaScanner {
	return &MediaScannerImpl{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// ScanLibrary 扫描媒体库
func (s *MediaScannerImpl) ScanLibrary(ctx context.Context, paths []string, userID uint) error {
	// 记录扫描开始时间
	start := time.Now()
	s.logger.Info("开始扫描媒体库",
		zap.Strings("paths", paths),
		zap.Uint("user_id", userID),
	)

	// 扫描每个路径
	for _, path := range paths {
		if err := s.scanPath(ctx, path, userID); err != nil {
			s.logger.Error("扫描路径失败",
				zap.String("path", path),
				zap.Error(err),
			)
			continue
		}
	}

	// 记录扫描结束时间
	duration := time.Since(start)
	s.logger.Info("媒体库扫描完成",
		zap.Duration("duration", duration),
		zap.Uint("user_id", userID),
	)

	return nil
}

// scanPath 扫描单个路径
func (s *MediaScannerImpl) scanPath(ctx context.Context, path string, userID uint) error {
	// 遍历目录
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			s.logger.Error("访问文件失败",
				zap.String("path", filePath),
				zap.Error(err),
			)
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			// 跳过隐藏目录
			if filepath.Base(filePath) == "." || filepath.Base(filePath) == ".." ||
				strings.HasPrefix(filepath.Base(filePath), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(filePath))
		if !isMediaFile(ext) {
			return nil
		}

		// 扫描文件
		_, err = s.ScanSingleFile(ctx, filePath, userID)
		if err != nil {
			s.logger.Error("扫描文件失败",
				zap.String("file_path", filePath),
				zap.Error(err),
			)
		}

		return nil
	})
}

// ScanSingleFile 扫描单个文件
func (s *MediaScannerImpl) ScanSingleFile(ctx context.Context, filePath string, userID uint) (*models.MediaFile, error) {
	// 检查文件是否已经存在
	var existingFile models.MediaFile
	if err := s.db.Where("file_path = ?", filePath).First(&existingFile).Error; err == nil {
		// 文件已存在，更新信息
		s.logger.Debug("文件已存在，更新信息",
			zap.String("file_path", filePath),
		)
		return s.updateFile(ctx, &existingFile)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// 创建媒体文件记录
	file := &models.MediaFile{
		FilePath:    filePath,
		FileName:    fileInfo.Name(),
		FileSize:    fileInfo.Size(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Processed:   false,
		InLibrary:   false,
		IsDuplicate: false,
	}

	// 保存文件记录
	if err := s.db.Create(file).Error; err != nil {
		s.logger.Error("保存媒体文件失败",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return nil, err
	}

	// 识别文件
	mediaItem, mediaVersion, err := s.IdentifyFile(ctx, file)
	if err != nil {
		s.logger.Warn("识别媒体文件失败",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return file, nil
	}

	// 关联文件
	if mediaItem != nil {
		file.MediaID = &mediaItem.ID
		file.InLibrary = true
	}
	if mediaVersion != nil {
		file.VersionID = &mediaVersion.ID
	}

	// 更新文件状态
	file.Processed = true
	if err := s.db.Save(file).Error; err != nil {
		s.logger.Error("更新媒体文件失败",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
	}

	// 更新媒体项状态
	if mediaItem != nil {
		mediaItem.InLibrary = true
		if err := s.db.Save(mediaItem).Error; err != nil {
			s.logger.Error("更新媒体项失败",
				zap.String("title", mediaItem.Title),
				zap.Error(err),
			)
		}
	}

	// 更新媒体版本状态
	if mediaVersion != nil {
		mediaVersion.InLibrary = true
		if err := s.db.Save(mediaVersion).Error; err != nil {
			s.logger.Error("更新媒体版本失败",
				zap.String("title", mediaVersion.Title),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("扫描单个文件完成",
		zap.String("file_path", filePath),
		zap.String("file_name", file.FileName),
	)

	return file, nil
}

// UpdateLibrary 更新媒体库
func (s *MediaScannerImpl) UpdateLibrary(ctx context.Context, userID uint) error {
	// 记录更新开始时间
	start := time.Now()
	s.logger.Info("开始更新媒体库",
		zap.Uint("user_id", userID),
	)

	// 获取所有媒体文件
	var files []models.MediaFile
	if err := s.db.Where("in_library = ?", true).Find(&files).Error; err != nil {
		s.logger.Error("获取媒体文件失败",
			zap.Error(err),
		)
		return err
	}

	// 更新每个文件
	for _, file := range files {
		// 检查文件是否存在
		if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
			// 文件不存在，标记为已删除
			file.InLibrary = false
			if err := s.db.Save(&file).Error; err != nil {
				s.logger.Error("更新媒体文件状态失败",
					zap.String("file_path", file.FilePath),
					zap.Error(err),
				)
			}
		}
	}

	// 记录更新结束时间
	duration := time.Since(start)
	s.logger.Info("媒体库更新完成",
		zap.Duration("duration", duration),
		zap.Uint("user_id", userID),
	)

	return nil
}

// IdentifyFile 识别媒体文件
func (s *MediaScannerImpl) IdentifyFile(ctx context.Context, file *models.MediaFile) (*models.MediaItem, *models.MediaVersion, error) {
	// 这里实现媒体文件识别逻辑
	// 1. 解析文件名获取媒体信息
	// 2. 查询TMDB等API获取元数据
	// 3. 创建或更新媒体项和版本

	// TODO: 实现媒体文件识别逻辑
	s.logger.Info("识别媒体文件",
		zap.String("file_path", file.FilePath),
	)

	return nil, nil, nil
}

// updateFile 更新媒体文件信息
func (s *MediaScannerImpl) updateFile(ctx context.Context, file *models.MediaFile) (*models.MediaFile, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		return nil, err
	}

	// 更新文件信息
	file.FileSize = fileInfo.Size()
	file.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(file).Error; err != nil {
		return nil, err
	}

	return file, nil
}

// isMediaFile 检查是否为媒体文件
func isMediaFile(ext string) bool {
	// 支持的媒体文件扩展名
	mediaExts := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".wmv":  true,
		".flv":  true,
		".mov":  true,
		".mpg":  true,
		".mpeg": true,
		".ts":   true,
		".webm": true,
		".m4v":  true,
		".iso":  true,
	}

	return mediaExts[ext]
}
