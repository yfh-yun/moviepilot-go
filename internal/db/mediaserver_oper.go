package db

import (
	"fmt"
	"strconv"
	
	"moviepilot-go/internal/db/models"
	"moviepilot-go/pkg/models"
	
	"gorm.io/gorm"
)

// MediaServerOper 媒体服务器数据管理
type MediaServerOper struct {
	DB *gorm.DB
}

// NewMediaServerOper 创建媒体服务器数据管理实例
func NewMediaServerOper(db *gorm.DB) *MediaServerOper {
	return &MediaServerOper{
		DB: db,
	}
}

// Add 新增媒体服务器数据
func (m *MediaServerOper) Add(itemData map[string]interface{}) (bool, error) {
	// MediaServerItem中没有的属性剔除
	item := &models.MediaServerItem{}
	
	// 从map中提取有效字段
	if val, ok := itemData["server"]; ok {
		if str, ok := val.(string); ok {
			item.Server = str
		}
	}
	if val, ok := itemData["library"]; ok {
		if str, ok := val.(string); ok {
			item.Library = str
		}
	}
	if val, ok := itemData["item_id"]; ok {
		if str, ok := val.(string); ok {
			item.ItemID = str
		}
	}
	if val, ok := itemData["item_type"]; ok {
		if str, ok := val.(string); ok {
			item.ItemType = str
		}
	}
	if val, ok := itemData["title"]; ok {
		if str, ok := val.(string); ok {
			item.Title = str
		}
	}
	if val, ok := itemData["original_title"]; ok {
		if str, ok := val.(string); ok {
			item.OriginalTitle = str
		}
	}
	if val, ok := itemData["year"]; ok {
		if str, ok := val.(string); ok {
			item.Year = str
		}
	}
	if val, ok := itemData["tmdbid"]; ok {
		if num, ok := val.(int); ok {
			item.TmdbID = num
		} else if floatVal, ok := val.(float64); ok {
			item.TmdbID = int(floatVal)
		}
	}
	if val, ok := itemData["imdbid"]; ok {
		if str, ok := val.(string); ok {
			item.ImdbID = str
		}
	}
	if val, ok := itemData["tvdbid"]; ok {
		if str, ok := val.(string); ok {
			item.TvdbID = str
		}
	}
	if val, ok := itemData["path"]; ok {
		if str, ok := val.(string); ok {
			item.Path = str
		}
	}
	if val, ok := itemData["seasoninfo"]; ok {
		if seasoninfo, ok := val.(map[string]interface{}); ok {
			item.Seasoninfo = seasoninfo
		}
	}
	if val, ok := itemData["note"]; ok {
		if note, ok := val.(map[string]interface{}); ok {
			item.Note = note
		}
	}
	if val, ok := itemData["lst_mod_date"]; ok {
		if str, ok := val.(string); ok {
			item.LstModDate = str
		}
	}
	
	// 检查是否已存在
	existingItem := &models.MediaServerItem{}
	err := m.DB.Where("item_id = ?", item.ItemID).First(existingItem).Error
	if err != nil {
		// 如果找不到记录，则创建新记录
		if err == gorm.ErrRecordNotFound {
			err = m.DB.Create(item).Error
			return err == nil, err
		}
		// 其他数据库错误
		return false, err
	}
	// 记录已存在
	return false, nil
}

// Empty 清空媒体服务器数据
func (m *MediaServerOper) Empty(server *string) error {
	mediaServerItem := &models.MediaServerItem{}
	return mediaServerItem.Empty(m.DB, server)
}

// Exists 判断媒体服务器数据是否存在
func (m *MediaServerOper) Exists(kwargs map[string]interface{}) (*models.MediaServerItem, error) {
	var item *models.MediaServerItem
	var err error
	
	// 根据tmdbid查找
	if tmdbidVal, ok := kwargs["tmdbid"]; ok {
		var tmdbid int
		if num, ok := tmdbidVal.(int); ok {
			tmdbid = num
		} else if floatVal, ok := tmdbidVal.(float64); ok {
			tmdbid = int(floatVal)
		}
		
		if tmdbid > 0 {
			mtype := ""
			if mtypeVal, ok := kwargs["mtype"]; ok {
				if str, ok := mtypeVal.(string); ok {
					mtype = str
				}
			}
			
			mediaServerItem := &models.MediaServerItem{}
			item, err = mediaServerItem.ExistByTmdbID(m.DB, tmdbid, mtype)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, nil
				}
				return nil, err
			}
		}
	} else if titleVal, ok := kwargs["title"]; ok {
		// 根据标题、类型、年份查找
		if title, ok := titleVal.(string); ok {
			mtype := ""
			if mtypeVal, ok := kwargs["mtype"]; ok {
				if str, ok := mtypeVal.(string); ok {
					mtype = str
				}
			}
			
			year := ""
			if yearVal, ok := kwargs["year"]; ok {
				if str, ok := yearVal.(string); ok {
					year = str
				}
			}
			
			mediaServerItem := &models.MediaServerItem{}
			item, err = mediaServerItem.ExistsByTitle(m.DB, title, mtype, year)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, nil
				}
				return nil, err
			}
		}
	} else {
		// 没有提供有效的查询条件
		return nil, nil
	}
	
	if item == nil {
		return nil, nil
	}
	
	// 检查季是否存在
	if seasonVal, ok := kwargs["season"]; ok {
		if season, ok := seasonVal.(int); ok {
			if item.Seasoninfo == nil {
				return nil, nil
			}
			
			// 检查season是否在seasoninfo中存在
			seasonKey := fmt.Sprintf("%d", season)
			if _, exists := item.Seasoninfo[seasonKey]; !exists {
				return nil, nil
			}
		}
	}
	
	return item, nil
}

// GetItemID 获取媒体服务器数据ID
func (m *MediaServerOper) GetItemID(kwargs map[string]interface{}) (*string, error) {
	item, err := m.Exists(kwargs)
	if err != nil || item == nil {
		return nil, err
	}
	
	itemID := item.ItemID
	return &itemID, nil
}