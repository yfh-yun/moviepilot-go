package db

import (
	"fmt"
	"time"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/db/models"
	"moviepilot-go/pkg/models"
)

// SiteOper 站点管理
type SiteOper struct {
	DB *gorm.DB
}

// NewSiteOper 创建站点管理实例
func NewSiteOper(db *gorm.DB) *SiteOper {
	return &SiteOper{
		DB: db,
	}
}

// Add 新增站点
func (s *SiteOper) Add(siteData map[string]interface{}) (bool, string) {
	// 从map中提取domain字段
	domain, _ := siteData["domain"].(string)
	
	// 检查站点是否已存在
	siteModel := &models.Site{}
	existingSite, err := siteModel.GetByDomain(s.DB, domain)
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Sprintf("查询站点失败: %v", err)
	}
	
	if existingSite != nil {
		return false, "站点已存在"
	}
	
	// 创建新站点
	site := &models.Site{}
	
	// 从map中提取有效字段
	if val, ok := siteData["name"]; ok {
		if str, ok := val.(string); ok {
			site.Name = str
		}
	}
	if val, ok := siteData["domain"]; ok {
		if str, ok := val.(string); ok {
			site.Domain = str
		}
	}
	if val, ok := siteData["url"]; ok {
		if str, ok := val.(string); ok {
			site.URL = str
		}
	}
	if val, ok := siteData["pri"]; ok {
		if num, ok := val.(int); ok {
			site.Pri = num
		} else if floatVal, ok := val.(float64); ok {
			site.Pri = int(floatVal)
		}
	}
	if val, ok := siteData["rss"]; ok {
		if str, ok := val.(string); ok {
			site.RSS = str
		}
	}
	if val, ok := siteData["cookie"]; ok {
		if str, ok := val.(string); ok {
			site.Cookie = str
		}
	}
	if val, ok := siteData["ua"]; ok {
		if str, ok := val.(string); ok {
			site.UA = str
		}
	}
	if val, ok := siteData["apikey"]; ok {
		if str, ok := val.(string); ok {
			site.APIKey = str
		}
	}
	if val, ok := siteData["token"]; ok {
		if str, ok := val.(string); ok {
			site.Token = str
		}
	}
	if val, ok := siteData["proxy"]; ok {
		if num, ok := val.(int); ok {
			site.Proxy = num
		} else if floatVal, ok := val.(float64); ok {
			site.Proxy = int(floatVal)
		}
	}
	if val, ok := siteData["filter"]; ok {
		if str, ok := val.(string); ok {
			site.Filter = str
		}
	}
	if val, ok := siteData["render"]; ok {
		if num, ok := val.(int); ok {
			site.Render = num
		} else if floatVal, ok := val.(float64); ok {
			site.Render = int(floatVal)
		}
	}
	if val, ok := siteData["public"]; ok {
		if num, ok := val.(int); ok {
			site.Public = num
		} else if floatVal, ok := val.(float64); ok {
			site.Public = int(floatVal)
		}
	}
	if val, ok := siteData["note"]; ok {
		if note, ok := val.(map[string]interface{}); ok {
			site.Note = note
		}
	}
	if val, ok := siteData["limit_interval"]; ok {
		if num, ok := val.(int); ok {
			site.LimitInterval = num
		} else if floatVal, ok := val.(float64); ok {
			site.LimitInterval = int(floatVal)
		}
	}
	if val, ok := siteData["limit_count"]; ok {
		if num, ok := val.(int); ok {
			site.LimitCount = num
		} else if floatVal, ok := val.(float64); ok {
			site.LimitCount = int(floatVal)
		}
	}
	if val, ok := siteData["limit_seconds"]; ok {
		if num, ok := val.(int); ok {
			site.LimitSeconds = num
		} else if floatVal, ok := val.(float64); ok {
			site.LimitSeconds = int(floatVal)
		}
	}
	if val, ok := siteData["timeout"]; ok {
		if num, ok := val.(int); ok {
			site.Timeout = num
		} else if floatVal, ok := val.(float64); ok {
			site.Timeout = int(floatVal)
		}
	}
	if val, ok := siteData["is_active"]; ok {
		if b, ok := val.(bool); ok {
			site.IsActive = b
		}
	}
	if val, ok := siteData["lst_mod_date"]; ok {
		if str, ok := val.(string); ok {
			site.LstModDate = str
		}
	}
	if val, ok := siteData["downloader"]; ok {
		if str, ok := val.(string); ok {
			site.Downloader = str
		}
	}
	
	// 创建站点
	err = s.DB.Create(site).Error
	if err != nil {
		return false, fmt.Sprintf("新增站点失败: %v", err)
	}
	
	return true, "新增站点成功"
}

// Get 查询单个站点
func (s *SiteOper) Get(sid uint) (*models.Site, error) {
	var site models.Site
	err := s.DB.First(&site, sid).Error
	return &site, err
}

// List 获取站点列表
func (s *SiteOper) List() ([]models.Site, error) {
	var sites []models.Site
	err := s.DB.Find(&sites).Error
	return sites, err
}

// ListOrderByPri 获取站点列表（按优先级排序）
func (s *SiteOper) ListOrderByPri() ([]models.Site, error) {
	siteModel := &models.Site{}
	return siteModel.ListOrderByPri(s.DB)
}

// ListActive 按状态获取站点列表（仅获取启用的站点）
func (s *SiteOper) ListActive() ([]models.Site, error) {
	siteModel := &models.Site{}
	return siteModel.GetActives(s.DB)
}

// Delete 删除站点
func (s *SiteOper) Delete(sid uint) error {
	return s.DB.Delete(&models.Site{}, sid).Error
}

// Update 更新站点
func (s *SiteOper) Update(sid uint, payload map[string]interface{}) (*models.Site, error) {
	site, err := s.Get(sid)
	if err != nil {
		return nil, err
	}
	
	// 更新站点信息
	err = s.DB.Model(site).Updates(payload).Error
	return site, err
}

// GetByDomain 按域名获取站点
func (s *SiteOper) GetByDomain(domain string) (*models.Site, error) {
	siteModel := &models.Site{}
	return siteModel.GetByDomain(s.DB, domain)
}

// GetDomainsByIds 按ID获取站点域名
func (s *SiteOper) GetDomainsByIds(ids []int) ([]string, error) {
	siteModel := &models.Site{}
	return siteModel.GetDomainsByIds(s.DB, ids)
}

// Exists 判断站点是否存在
func (s *SiteOper) Exists(domain string) bool {
	_, err := s.GetByDomain(domain)
	return err == nil
}

// UpdateCookie 更新站点Cookie
func (s *SiteOper) UpdateCookie(domain string, cookies string) (bool, string) {
	site, err := s.GetByDomain(domain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "站点不存在"
		}
		return false, fmt.Sprintf("查询站点失败: %v", err)
	}
	
	err = s.DB.Model(site).Update("cookie", cookies).Error
	if err != nil {
		return false, fmt.Sprintf("更新站点Cookie失败: %v", err)
	}
	
	return true, "更新站点Cookie成功"
}

// UpdateRSS 更新站点RSS地址
func (s *SiteOper) UpdateRSS(domain string, rss string) (bool, string) {
	site, err := s.GetByDomain(domain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "站点不存在"
		}
		return false, fmt.Sprintf("查询站点失败: %v", err)
	}
	
	err = s.DB.Model(site).Update("rss", rss).Error
	if err != nil {
		return false, fmt.Sprintf("更新站点RSS地址失败: %v", err)
	}
	
	return true, "更新站点RSS地址成功"
}

// UpdateUserdata 更新站点用户数据
func (s *SiteOper) UpdateUserdata(domain string, name string, payload map[string]interface{}) (bool, string) {
	// 当前系统日期
	currentDay := time.Now().Format("2006-01-02")
	currentTime := time.Now().Format("15:04:05")
	
	// 更新payload
	payload["domain"] = domain
	payload["name"] = name
	payload["updated_day"] = currentDay
	payload["updated_time"] = currentTime
	
	// 处理错误信息
	if _, exists := payload["err_msg"]; !exists {
		payload["err_msg"] = ""
	}
	
	// 按站点+天判断是否存在数据
	siteUserDataModel := &models.SiteUserData{}
	var workdate = currentDay
	siteUserDatas, err := siteUserDataModel.GetByDomain(s.DB, domain, &workdate, nil)
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Sprintf("查询站点用户数据失败: %v", err)
	}
	
	if len(siteUserDatas) > 0 {
		// 存在则更新
		if _, exists := payload["err_msg"]; !exists || payload["err_msg"] == "" {
			err = s.DB.Model(&siteUserDatas[0]).Updates(payload).Error
			if err != nil {
				return false, fmt.Sprintf("更新站点用户数据失败: %v", err)
			}
		}
	} else {
		// 不存在则插入
		siteUserData := &models.SiteUserData{}
		
		// 从payload中提取字段
		if val, ok := payload["domain"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.Domain = str
			}
		}
		if val, ok := payload["name"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.Name = str
			}
		}
		if val, ok := payload["username"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.Username = str
			}
		}
		if val, ok := payload["userid"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.UserID = str
			}
		}
		if val, ok := payload["user_level"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.UserLevel = str
			}
		}
		if val, ok := payload["join_at"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.JoinAt = str
			}
		}
		if val, ok := payload["bonus"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Bonus = num
			}
		}
		if val, ok := payload["upload"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Upload = num
			}
		}
		if val, ok := payload["download"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Download = num
			}
		}
		if val, ok := payload["ratio"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Ratio = num
			}
		}
		if val, ok := payload["seeding"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Seeding = num
			}
		}
		if val, ok := payload["leeching"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.Leeching = num
			}
		}
		if val, ok := payload["seeding_size"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.SeedingSize = num
			}
		}
		if val, ok := payload["leeching_size"]; ok {
			if num, ok := val.(float64); ok {
				siteUserData.LeechingSize = num
			}
		}
		if val, ok := payload["seeding_info"]; ok {
			if info, ok := val.([]interface{}); ok {
				siteUserData.SeedingInfo = info
			}
		}
		if val, ok := payload["message_unread"]; ok {
			if num, ok := val.(int); ok {
				siteUserData.MessageUnread = num
			} else if floatVal, ok := val.(float64); ok {
				siteUserData.MessageUnread = int(floatVal)
			}
		}
		if val, ok := payload["message_unread_contents"]; ok {
			if contents, ok := val.([]interface{}); ok {
				siteUserData.MessageUnreadContents = contents
			}
		}
		if val, ok := payload["err_msg"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.ErrMsg = str
			}
		}
		if val, ok := payload["updated_day"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.UpdatedDay = str
			}
		}
		if val, ok := payload["updated_time"]; ok {
			if str, ok := val.(string); ok {
				siteUserData.UpdatedTime = str
			}
		}
		
		err = s.DB.Create(siteUserData).Error
		if err != nil {
			return false, fmt.Sprintf("创建站点用户数据失败: %v", err)
		}
	}
	
	return true, "更新站点用户数据成功"
}

// GetUserdata 获取站点用户数据
func (s *SiteOper) GetUserdata() ([]models.SiteUserData, error) {
	var siteUserDataList []models.SiteUserData
	err := s.DB.Find(&siteUserDataList).Error
	return siteUserDataList, err
}

// GetUserdataByDomain 获取站点用户数据
func (s *SiteOper) GetUserdataByDomain(domain string, workdate *string) ([]models.SiteUserData, error) {
	siteUserDataModel := &models.SiteUserData{}
	return siteUserDataModel.GetByDomain(s.DB, domain, workdate, nil)
}

// GetUserdataByDate 获取站点用户数据
func (s *SiteOper) GetUserdataByDate(date string) ([]models.SiteUserData, error) {
	siteUserDataModel := &models.SiteUserData{}
	return siteUserDataModel.GetByDate(s.DB, date)
}

// GetUserdataLatest 获取站点最新数据
func (s *SiteOper) GetUserdataLatest() ([]models.SiteUserData, error) {
	siteUserDataModel := &models.SiteUserData{}
	return siteUserDataModel.GetLatest(s.DB)
}

// GetIconByDomain 按域名获取站点图标
func (s *SiteOper) GetIconByDomain(domain string) (*models.SiteIcon, error) {
	siteIconModel := &models.SiteIcon{}
	return siteIconModel.GetByDomain(s.DB, domain)
}

// UpdateIcon 更新站点图标
func (s *SiteOper) UpdateIcon(name string, domain string, iconURL string, iconBase64 string) bool {
	// 处理Base64图标数据
	if iconBase64 != "" {
		iconBase64 = fmt.Sprintf("data:image/ico;base64,%s", iconBase64)
	}
	
	// 获取现有图标
	siteIcon, err := s.GetIconByDomain(domain)
	if err != nil && err != gorm.ErrRecordNotFound {
		return false
	}
	
	if siteIcon == nil {
		// 创建新图标记录
		newSiteIcon := &models.SiteIcon{
			Name:   name,
			Domain: domain,
			URL:    iconURL,
			Base64: iconBase64,
		}
		err = s.DB.Create(newSiteIcon).Error
		return err == nil
	} else if iconBase64 != "" {
		// 更新现有图标记录
		err = s.DB.Model(siteIcon).Updates(map[string]interface{}{
			"url":    iconURL,
			"base64": iconBase64,
		}).Error
		return err == nil
	}
	
	return true
}

// Success 站点访问成功
func (s *SiteOper) Success(domain string, seconds *int) error {
	lstDate := time.Now().Format("2006-01-02 15:04:05")
	
	// 获取现有统计记录
	siteStatisticModel := &models.SiteStatistic{}
	siteStatistic, err := siteStatisticModel.GetByDomain(s.DB, domain)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	
	if siteStatistic != nil {
		// 使用深复制确保 note 是全新的字典对象
		note := make(map[string]interface{})
		if siteStatistic.Note != nil {
			for k, v := range siteStatistic.Note {
				note[k] = v
			}
		}
		
		var avgSeconds *int
		if seconds != nil {
			note[lstDate] = *seconds
			avgTimes := len(note)
			if avgTimes > 10 {
				// 保留最新的10条记录
				// 在Go中，map是无序的，这里简化处理，实际应该用其他数据结构
				// 此处仅做功能实现，不完全等同于Python版本的逻辑
				if len(note) > 10 {
					// 简化处理，删除一些旧记录
					count := 0
					for k := range note {
						if count >= 10 {
							delete(note, k)
						}
						count++
					}
				}
			}
			
			// 计算平均时间
			total := 0
			for _, v := range note {
				if val, ok := v.(int); ok {
					total += val
				} else if val, ok := v.(float64); ok {
					total += int(val)
				}
			}
			avg := total / len(note)
			avgSeconds = &avg
		}
		
		// 更新记录
		updates := map[string]interface{}{
			"success":       siteStatistic.Success + 1,
			"lst_state":     0,
			"lst_mod_date":  lstDate,
			"note":          note,
		}
		
		if avgSeconds != nil {
			updates["seconds"] = *avgSeconds
		} else {
			updates["seconds"] = siteStatistic.Seconds
		}
		
		err = s.DB.Model(siteStatistic).Updates(updates).Error
	} else {
		note := make(map[string]interface{})
		if seconds != nil {
			note[lstDate] = *seconds
		}
		
		newSiteStatistic := &models.SiteStatistic{
			Domain:      domain,
			Success:     1,
			Fail:        0,
			Seconds:     1,
			LstState:    0,
			LstModDate:  lstDate,
			Note:        note,
		}
		
		if seconds != nil {
			newSiteStatistic.Seconds = *seconds
		}
		
		err = s.DB.Create(newSiteStatistic).Error
	}
	
	return err
}

// Fail 站点访问失败
func (s *SiteOper) Fail(domain string) error {
	lstDate := time.Now().Format("2006-01-02 15:04:05")
	
	// 获取现有统计记录
	siteStatisticModel := &models.SiteStatistic{}
	siteStatistic, err := siteStatisticModel.GetByDomain(s.DB, domain)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	
	if siteStatistic != nil {
		// 更新记录
		err = s.DB.Model(siteStatistic).Updates(map[string]interface{}{
			"fail":         siteStatistic.Fail + 1,
			"lst_state":    1,
			"lst_mod_date": lstDate,
		}).Error
	} else {
		newSiteStatistic := &models.SiteStatistic{
			Domain:      domain,
			Success:     0,
			Fail:        1,
			LstState:    1,
			LstModDate:  lstDate,
		}
		err = s.DB.Create(newSiteStatistic).Error
	}
	
	return err
}