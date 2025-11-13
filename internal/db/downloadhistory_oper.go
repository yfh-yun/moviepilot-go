package db

import (
	"moviepilot-go/internal/db/models"
	"moviepilot-go/pkg/models"
	
	"gorm.io/gorm"
)

// DownloadHistoryOper 下载历史管理
type DownloadHistoryOper struct {
	DB *gorm.DB
}

// NewDownloadHistoryOper 创建下载历史管理实例
func NewDownloadHistoryOper(db *gorm.DB) *DownloadHistoryOper {
	return &DownloadHistoryOper{
		DB: db,
	}
}

// GetByPath 按路径查询下载记录
func (d *DownloadHistoryOper) GetByPath(path string) (*models.DownloadHistory, error) {
	dh := &models.DownloadHistory{}
	err := d.DB.Where("path = ?", path).First(dh).Error
	return dh, err
}

// GetByHash 按Hash查询下载记录
func (d *DownloadHistoryOper) GetByHash(downloadHash string) (*models.DownloadHistory, error) {
	history := &models.DownloadHistory{}
	err := d.DB.Where("download_hash = ?", downloadHash).Order("date DESC").First(history).Error
	return history, err
}

// GetByMediaid 按媒体ID查询下载记录
func (d *DownloadHistoryOper) GetByMediaid(tmdbid int, doubanid string) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.GetByMediaID(d.DB, tmdbid, doubanid)
}

// Add 新增下载历史
func (d *DownloadHistoryOper) Add(downloadHistory *models.DownloadHistory) error {
	return d.DB.Create(downloadHistory).Error
}

// AddFiles 新增下载历史文件
func (d *DownloadHistoryOper) AddFiles(fileItems []map[string]interface{}) error {
	for _, fileItem := range fileItems {
		downloadFile := &models.DownloadFiles{
			Downloader:    getString(fileItem, "downloader"),
			DownloadHash:  getString(fileItem, "download_hash"),
			Fullpath:      getString(fileItem, "fullpath"),
			Savepath:      getString(fileItem, "savepath"),
			Filepath:      getString(fileItem, "filepath"),
			Torrentname:   getString(fileItem, "torrentname"),
			State:         getInt(fileItem, "state", 1),
		}
		
		if err := d.DB.Create(downloadFile).Error; err != nil {
			return err
		}
	}
	return nil
}

// TruncateFiles 清空下载历史文件记录
func (d *DownloadHistoryOper) TruncateFiles() error {
	return d.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.DownloadFiles{}).Error
}

// GetFilesByHash 按Hash查询下载文件记录
func (d *DownloadHistoryOper) GetFilesByHash(downloadHash string, state *int) ([]models.DownloadFiles, error) {
	downloadFiles := &models.DownloadFiles{}
	return downloadFiles.GetByHash(d.DB, downloadHash, state)
}

// GetFileByFullpath 按fullpath查询下载文件记录
func (d *DownloadHistoryOper) GetFileByFullpath(fullpath string) (*models.DownloadFiles, error) {
	files, err := d.GetFilesByFullpath(fullpath)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	return &files[0], nil
}

// GetFilesByFullpath 按fullpath查询下载文件记录
func (d *DownloadHistoryOper) GetFilesByFullpath(fullpath string) ([]models.DownloadFiles, error) {
	downloadFiles := &models.DownloadFiles{}
	return downloadFiles.GetByFullpath(d.DB, fullpath, true)
}

// GetFilesBySavepath 按savepath查询下载文件记录
func (d *DownloadHistoryOper) GetFilesBySavepath(fullpath string) ([]models.DownloadFiles, error) {
	downloadFiles := &models.DownloadFiles{}
	return downloadFiles.GetBySavepath(d.DB, fullpath)
}

// DeleteFileByFullpath 按fullpath删除下载文件记录
func (d *DownloadHistoryOper) DeleteFileByFullpath(fullpath string) error {
	downloadFiles := &models.DownloadFiles{}
	return downloadFiles.DeleteByFullpath(d.DB, fullpath)
}

// GetHashByFullpath 按fullpath查询下载文件记录hash
func (d *DownloadHistoryOper) GetHashByFullpath(fullpath string) string {
	fileinfo, err := d.GetFileByFullpath(fullpath)
	if err != nil || fileinfo == nil {
		return ""
	}
	return fileinfo.DownloadHash
}

// ListByPage 分页查询下载历史
func (d *DownloadHistoryOper) ListByPage(page int, count int) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.ListByPage(d.DB, page, count)
}

// Truncate 清空下载记录
func (d *DownloadHistoryOper) Truncate() error {
	return d.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.DownloadHistory{}).Error
}

// GetLastBy 按类型、标题、年份、季集查询下载记录
func (d *DownloadHistoryOper) GetLastBy(mtype, title, year, season, episode string, tmdbid int) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.GetLastBy(d.DB, mtype, title, year, season, episode, tmdbid)
}

// ListByUserDate 查询某用户某时间之前的下载历史
func (d *DownloadHistoryOper) ListByUserDate(date string, username string) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.ListByUserDate(d.DB, date, username)
}

// ListByDate 查询某时间之后的下载历史
func (d *DownloadHistoryOper) ListByDate(date string, mtype string, tmdbid string, seasons string) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.ListByDate(d.DB, date, mtype, tmdbid, seasons)
}

// ListByType 获取指定类型的下载历史
func (d *DownloadHistoryOper) ListByType(mtype string, days int) ([]models.DownloadHistory, error) {
	downloadHistory := &models.DownloadHistory{}
	return downloadHistory.ListByType(d.DB, mtype, days)
}

// DeleteHistory 删除下载记录
func (d *DownloadHistoryOper) DeleteHistory(historyid uint) error {
	return d.DB.Delete(&models.DownloadHistory{}, historyid).Error
}

// DeleteDownloadfile 删除下载文件记录
func (d *DownloadHistoryOper) DeleteDownloadfile(downloadfileid uint) error {
	return d.DB.Delete(&models.DownloadFiles{}, downloadfileid).Error
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(int); ok {
			return num
		} else if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultVal
}