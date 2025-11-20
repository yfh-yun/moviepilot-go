package repositories

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// downloadRepository 下载历史仓储实现
type downloadRepository struct {
	db *gorm.DB
}

// NewDownloadRepository 创建下载历史仓储
func NewDownloadRepository(db *gorm.DB) interfaces.DownloadRepository {
	return &downloadRepository{db: db}
}

// Create 创建下载历史记录
func (r *downloadRepository) Create(history *model.DownloadHistory) error {
	return r.db.Create(history).Error
}

// Update 更新下载历史记录
func (r *downloadRepository) Update(history *model.DownloadHistory) error {
	return r.db.Save(history).Error
}

// Delete 删除下载历史记录
func (r *downloadRepository) Delete(id uint) error {
	return r.db.Delete(&model.DownloadHistory{}, id).Error
}

// GetByID 根据ID获取下载历史
func (r *downloadRepository) GetByID(id uint) (*model.DownloadHistory, error) {
	var history model.DownloadHistory
	err := r.db.First(&history, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByHash 根据下载Hash获取下载历史
func (r *downloadRepository) GetByHash(downloadHash string) (*model.DownloadHistory, error) {
	var history model.DownloadHistory
	err := r.db.Where("download_hash = ?", downloadHash).Order("date DESC").First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByPath 根据路径获取下载历史
func (r *downloadRepository) GetByPath(path string) (*model.DownloadHistory, error) {
	var history model.DownloadHistory
	err := r.db.Where("path = ?", path).First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByMediaID 根据媒体ID获取下载历史
func (r *downloadRepository) GetByMediaID(tmdbID *int, doubanID *string) ([]*model.DownloadHistory, error) {
	var histories []*model.DownloadHistory
	query := r.db

	if tmdbID != nil {
		query = query.Where("tmdbid = ?", *tmdbID)
	} else if doubanID != nil {
		query = query.Where("doubanid = ?", *doubanID)
	}

	err := query.Find(&histories).Error
	return histories, err
}

// GetLast 获取最后的下载记录
func (r *downloadRepository) GetLast(mtype, title, year, season, episode *string, tmdbID *int) ([]*model.DownloadHistory, error) {
	var histories []*model.DownloadHistory
	query := r.db.Order("id DESC")

	if tmdbID != nil && mtype != nil {
		query = query.Where("tmdbid = ? AND type = ?", *tmdbID, *mtype)
		if season != nil && episode != nil {
			query = query.Where("seasons = ? AND episodes = ?", *season, *episode)
		} else if season != nil {
			query = query.Where("seasons = ?", *season)
		}
	} else if title != nil && year != nil {
		query = query.Where("title = ? AND year = ?", *title, *year)
		if season != nil && episode != nil {
			query = query.Where("seasons = ? AND episodes = ?", *season, *episode)
		} else if season != nil {
			query = query.Where("seasons = ?", *season)
		}
	} else {
		return []*model.DownloadHistory{}, nil
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByPage 分页获取下载历史
func (r *downloadRepository) ListByPage(page, count int) ([]*model.DownloadHistory, int64, error) {
	var histories []*model.DownloadHistory
	var total int64

	err := r.db.Model(&model.DownloadHistory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * count
	err = r.db.Offset(offset).Limit(count).Find(&histories).Error
	return histories, total, err
}

// ListByUserDate 根据用户和时间获取下载历史
func (r *downloadRepository) ListByUserDate(date string, username string) ([]*model.DownloadHistory, error) {
	var histories []*model.DownloadHistory
	query := r.db.Where("date < ?", date).Order("id DESC")

	if username != "" {
		query = query.Where("username = ?", username)
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByDate 根据日期获取下载历史
func (r *downloadRepository) ListByDate(date string, mtype string, tmdbID string, seasons *string) ([]*model.DownloadHistory, error) {
	var histories []*model.DownloadHistory
	query := r.db.Where("date > ? AND type = ? AND tmdbid = ?", date, mtype, tmdbID).Order("id DESC")

	if seasons != nil {
		query = query.Where("seasons = ?", *seasons)
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByType 根据类型获取下载历史
func (r *downloadRepository) ListByType(mtype string, days int) ([]*model.DownloadHistory, error) {
	var histories []*model.DownloadHistory
	since := time.Now().AddDate(0, 0, -days)

	err := r.db.Where("type = ? AND date >= ?", mtype, since).Find(&histories).Error
	return histories, err
}

// Count 统计下载历史数量
func (r *downloadRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.DownloadHistory{}).Count(&count).Error
	return count, err
}

// CountByUser 统计用户下载数量
func (r *downloadRepository) CountByUser(username string) (int64, error) {
	var count int64
	err := r.db.Model(&model.DownloadHistory{}).Where("username = ?", username).Count(&count).Error
	return count, err
}

// downloadFilesRepository 下载文件仓储实现
type downloadFilesRepository struct {
	db *gorm.DB
}

// NewDownloadFilesRepository 创建下载文件仓储
func NewDownloadFilesRepository(db *gorm.DB) interfaces.DownloadFilesRepository {
	return &downloadFilesRepository{db: db}
}

// Create 创建下载文件记录
func (r *downloadFilesRepository) Create(files *model.DownloadFiles) error {
	return r.db.Create(files).Error
}

// Update 更新下载文件记录
func (r *downloadFilesRepository) Update(files *model.DownloadFiles) error {
	return r.db.Save(files).Error
}

// Delete 删除下载文件记录
func (r *downloadFilesRepository) Delete(id uint) error {
	return r.db.Delete(&model.DownloadFiles{}, id).Error
}

// GetByHash 根据下载Hash获取文件列表
func (r *downloadFilesRepository) GetByHash(downloadHash string, state *int) ([]*model.DownloadFiles, error) {
	var files []*model.DownloadFiles
	query := r.db.Where("download_hash = ?", downloadHash)

	if state != nil {
		query = query.Where("state = ?", *state)
	}

	err := query.Find(&files).Error
	return files, err
}

// GetByFullPath 根据完整路径获取文件
func (r *downloadFilesRepository) GetByFullPath(fullPath string, allFiles bool) ([]*model.DownloadFiles, error) {
	var files []*model.DownloadFiles
	query := r.db.Where("fullpath = ?", fullPath).Order("id DESC")

	if !allFiles {
		var file model.DownloadFiles
		err := query.First(&file).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return []*model.DownloadFiles{}, nil
			}
			return nil, err
		}
		return []*model.DownloadFiles{&file}, nil
	}

	err := query.Find(&files).Error
	return files, err
}

// GetBySavePath 根据保存路径获取文件列表
func (r *downloadFilesRepository) GetBySavePath(savePath string) ([]*model.DownloadFiles, error) {
	var files []*model.DownloadFiles
	err := r.db.Where("savepath = ?", savePath).Find(&files).Error
	return files, err
}

// DeleteByFullPath 根据完整路径删除文件
func (r *downloadFilesRepository) DeleteByFullPath(fullPath string) error {
	return r.db.Model(&model.DownloadFiles{}).Where("fullpath = ? AND state = ?", fullPath, 1).Update("state", 0).Error
}

// UpdateState 更新文件状态
func (r *downloadFilesRepository) UpdateState(id uint, state int) error {
	return r.db.Model(&model.DownloadFiles{}).Where("id = ?", id).Update("state", state).Error
}

// Count 统计文件数量
func (r *downloadFilesRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.DownloadFiles{}).Count(&count).Error
	return count, err
}

// ==================== DownloadRepository 扩展方法 ====================

// BatchCreate 批量创建下载历史
func (r *downloadRepository) BatchCreate(histories []*model.DownloadHistory) error {
	if len(histories) == 0 {
		return nil
	}

	batchSize := 100
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(histories); i += batchSize {
			end := i + batchSize
			if end > len(histories) {
				end = len(histories)
			}

			batch := histories[i:end]
			if err := tx.CreateInBatches(batch, batchSize).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AddFiles 添加下载文件记录
func (r *downloadRepository) AddFiles(fileItems []map[string]interface{}) error {
	if len(fileItems) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, fileItem := range fileItems {
			downloadFile := &model.DownloadFiles{}
			
			// 将map转换为struct
			if err := mapToStruct(fileItem, downloadFile); err != nil {
				return fmt.Errorf("failed to convert file item: %w", err)
			}

			if err := tx.Create(downloadFile).Error; err != nil {
				return fmt.Errorf("failed to create download file: %w", err)
			}
		}
		return nil
	})
}

// GetFilesByHash 根据Hash获取下载文件
func (r *downloadRepository) GetFilesByHash(downloadHash string, state *int) ([]*model.DownloadFiles, error) {
	filesRepo := NewDownloadFilesRepository(r.db)
	return filesRepo.GetByHash(downloadHash, state)
}

// GetFileByFullPath 根据完整路径获取下载文件
func (r *downloadRepository) GetFileByFullPath(fullpath string) (*model.DownloadFiles, error) {
	filesRepo := NewDownloadFilesRepository(r.db)
	files, err := filesRepo.GetByFullPath(fullpath, false)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	return files[0], nil
}

// GetFilesByFullPath 根据完整路径获取所有下载文件
func (r *downloadRepository) GetFilesByFullPath(fullpath string) ([]*model.DownloadFiles, error) {
	filesRepo := NewDownloadFilesRepository(r.db)
	return filesRepo.GetByFullPath(fullpath, true)
}

// GetFilesBySavePath 根据保存路径获取下载文件
func (r *downloadRepository) GetFilesBySavePath(savepath string) ([]*model.DownloadFiles, error) {
	filesRepo := NewDownloadFilesRepository(r.db)
	return filesRepo.GetBySavePath(savepath)
}

// TruncateFiles 清空下载文件记录
func (r *downloadRepository) TruncateFiles() error {
	return r.db.Exec("DELETE FROM download_files").Error
}

// UpdateFileState 更新文件状态
func (r *downloadRepository) UpdateFileState(id uint, state int) error {
	filesRepo := NewDownloadFilesRepository(r.db)
	return filesRepo.UpdateState(id, state)
}

// BatchDeleteFiles 批量删除文件记录
func (r *downloadRepository) BatchDeleteFiles(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.Where("id IN ?", ids).Delete(&model.DownloadFiles{}).Error
}

// ==================== 辅助函数 ====================

// mapToStruct 将map转换为struct
func mapToStruct(data map[string]interface{}, target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		
		// 优先使用json标签，否则使用字段名
		mapKey := jsonTag
		if mapKey == "" || mapKey == "-" {
			mapKey = fieldName
		}

		if value, exists := data[mapKey]; exists {
			fieldValue := targetValue.Field(i)
			if fieldValue.CanSet() {
				// 处理gorm column标签
				gormTag := field.Tag.Get("gorm")
				if gormTag != "" {
					// 提取column名称
					for _, tag := range []string{"column:", "size:", "type:", "index", "not null"} {
						if len(tag) > 0 && tag[0] != ':' {
							columnName := tag
							if columnValue, exists := data[columnName]; exists {
								value = columnValue
								break
							}
						}
					}
				}

				// 类型转换
				switch fieldValue.Kind() {
				case reflect.String:
					if str, ok := value.(string); ok {
						fieldValue.SetString(str)
					}
				case reflect.Int, reflect.Int64:
					if num, ok := value.(float64); ok {
						fieldValue.SetInt(int64(num))
					} else if num, ok := value.(int); ok {
						fieldValue.SetInt(int64(num))
					}
				case reflect.Float64:
					if num, ok := value.(float64); ok {
						fieldValue.SetFloat(num)
					}
				case reflect.Bool:
					if b, ok := value.(bool); ok {
						fieldValue.SetBool(b)
					}
				case reflect.Ptr:
					if value != nil {
						// 处理指针类型
						switch fieldValue.Type().Elem().Kind() {
						case reflect.String:
							if str, ok := value.(string); ok {
								ptr := str
								fieldValue.Set(reflect.ValueOf(&ptr))
							}
						case reflect.Int:
							if num, ok := value.(float64); ok {
								intVal := int(num)
								fieldValue.Set(reflect.ValueOf(&intVal))
							} else if num, ok := value.(int); ok {
								fieldValue.Set(reflect.ValueOf(&num))
							}
						case reflect.Struct:
							if fieldValue.Type().Elem().String() == "time.Time" {
								if str, ok := value.(string); ok {
									if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
										fieldValue.Set(reflect.ValueOf(&t))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// ==================== DownloadFilesRepository 构造函数 ====================
