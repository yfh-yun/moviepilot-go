package models

import (
	"time"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/db"
	"moviepilot-go/pkg/models"
)

// DownloadHistory 下载历史记录模型
type DownloadHistory struct {
	models.DownloadHistory
}

// DownloadFiles 下载文件记录模型
type DownloadFiles struct {
	models.DownloadFiles
}

// GetByHash 根据Hash查询下载记录
func (dh *DownloadHistory) GetByHash(db *gorm.DB, downloadHash string) (*DownloadHistory, error) {
	var history DownloadHistory
	err := db.Where("download_hash = ?", downloadHash).Order("date DESC").First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// GetByMediaID 根据媒体ID查询下载记录
func (dh *DownloadHistory) GetByMediaID(db *gorm.DB, tmdbid int, doubanid string) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	
	if tmdbid > 0 {
		err := db.Where("tmdbid = ?", tmdbid).Find(&histories).Error
		return histories, err
	} else if doubanid != "" {
		err := db.Where("doubanid = ?", doubanid).Find(&histories).Error
		return histories, err
	}
	
	return histories, nil
}

// ListByPage 分页查询下载历史
func (dh *DownloadHistory) ListByPage(db *gorm.DB, page int, count int) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	offset := (page - 1) * count
	err := db.Offset(offset).Limit(count).Find(&histories).Error
	return histories, err
}

// GetByPath 根据路径查询下载记录
func (dh *DownloadHistory) GetByPath(db *gorm.DB, path string) (*DownloadHistory, error) {
	var history DownloadHistory
	err := db.Where("path = ?", path).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// GetLastBy 根据类型、标题、年份、季集查询下载记录
func (dh *DownloadHistory) GetLastBy(db *gorm.DB, mtype, title, year, season, episode string, tmdbid int) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	
	// TMDBID + 类型
	if tmdbid > 0 && mtype != "" {
		// 电视剧某季某集
		if season != "" && episode != "" {
			err := db.Where("tmdbid = ? AND type = ? AND seasons = ? AND episodes = ?", tmdbid, mtype, season, episode).
				Order("id DESC").Find(&histories).Error
			return histories, err
		// 电视剧某季
		} else if season != "" {
			err := db.Where("tmdbid = ? AND type = ? AND seasons = ?", tmdbid, mtype, season).
				Order("id DESC").Find(&histories).Error
			return histories, err
		// 电视剧所有季集/电影
		} else {
			err := db.Where("tmdbid = ? AND type = ?", tmdbid, mtype).
				Order("id DESC").Find(&histories).Error
			return histories, err
		}
	// 标题 + 年份
	} else if title != "" && year != "" {
		// 电视剧某季某集
		if season != "" && episode != "" {
			err := db.Where("title = ? AND year = ? AND seasons = ? AND episodes = ?", title, year, season, episode).
				Order("id DESC").Find(&histories).Error
			return histories, err
		// 电视剧某季
		} else if season != "" {
			err := db.Where("title = ? AND year = ? AND seasons = ?", title, year, season).
				Order("id DESC").Find(&histories).Error
			return histories, err
		// 电视剧所有季集/电影
		} else {
			err := db.Where("title = ? AND year = ?", title, year).
				Order("id DESC").Find(&histories).Error
			return histories, err
		}
	}
	
	return histories, nil
}

// ListByUserDate 查询某用户某时间之后的下载历史
func (dh *DownloadHistory) ListByUserDate(db *gorm.DB, date string, username string) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	
	if username != "" {
		err := db.Where("date < ? AND username = ?", date, username).
			Order("id DESC").Find(&histories).Error
		return histories, err
	} else {
		err := db.Where("date < ?", date).
			Order("id DESC").Find(&histories).Error
		return histories, err
	}
}

// ListByDate 查询某时间之后的下载历史
func (dh *DownloadHistory) ListByDate(db *gorm.DB, date string, mtype string, tmdbid string, seasons string) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	
	if seasons != "" {
		err := db.Where("date > ? AND type = ? AND tmdbid = ? AND seasons = ?", date, mtype, tmdbid, seasons).
			Order("id DESC").Find(&histories).Error
		return histories, err
	} else {
		err := db.Where("date > ? AND type = ? AND tmdbid = ?", date, mtype, tmdbid).
			Order("id DESC").Find(&histories).Error
		return histories, err
	}
}

// ListByType 获取指定类型的下载历史
func (dh *DownloadHistory) ListByType(db *gorm.DB, mtype string, days int) ([]DownloadHistory, error) {
	var histories []DownloadHistory
	// 计算指定天数前的时间
	cutoffTime := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	
	err := db.Where("type = ? AND date >= ?", mtype, cutoffTime).Find(&histories).Error
	return histories, err
}

// GetByHash 根据Hash查询下载文件记录
func (df *DownloadFiles) GetByHash(db *gorm.DB, downloadHash string, state *int) ([]DownloadFiles, error) {
	var files []DownloadFiles
	query := db.Where("download_hash = ?", downloadHash)
	
	if state != nil {
		query = query.Where("state = ?", *state)
	}
	
	err := query.Find(&files).Error
	return files, err
}

// GetByFullpath 根据完整路径查询下载文件记录
func (df *DownloadFiles) GetByFullpath(db *gorm.DB, fullpath string, allFiles bool) ([]DownloadFiles, error) {
	var files []DownloadFiles
	query := db.Where("fullpath = ?", fullpath).Order("id DESC")
	
	if !allFiles {
		var file DownloadFiles
		err := query.First(&file).Error
		if err != nil {
			return nil, err
		}
		return []DownloadFiles{file}, nil
	}
	
	err := query.Find(&files).Error
	return files, err
}

// GetBySavepath 根据保存路径查询下载文件记录
func (df *DownloadFiles) GetBySavepath(db *gorm.DB, savepath string) ([]DownloadFiles, error) {
	var files []DownloadFiles
	err := db.Where("savepath = ?", savepath).Find(&files).Error
	return files, err
}

// DeleteByFullpath 根据完整路径删除下载文件记录（标记为已删除）
func (df *DownloadFiles) DeleteByFullpath(db *gorm.DB, fullpath string) error {
	return db.Model(&DownloadFiles{}).Where("fullpath = ? AND state = 1", fullpath).
		Update("state", 0).Error
}