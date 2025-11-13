package indexer

import (
	"fmt"
	"time"
	
	"moviepilot-go/internal/core"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// IndexerModule 索引模块
type IndexerModule struct {
	siteSchemas []interface{} // 站点解析器列�?}

// InitModule 模块初始�?func (im *IndexerModule) InitModule() error {
	// 加载模块
	// TODO: 实现模块加载逻辑
	// im.siteSchemas = core.ModuleHelper.Load(
	// 	"app.modules.indexer.parser",
	// 	func(obj interface{}) bool {
	// 		if schemaGetter, ok := obj.(interface{ GetSchema() interface{} }); ok {
	// 			return schemaGetter.GetSchema() != nil
	// 		}
	// 		return false
	// 	},
	// )
	return nil
}

// GetName 获取模块名称
func (im *IndexerModule) GetName() string {
	return "站点索引"
}

// GetType 获取模块类型
func (im *IndexerModule) GetType() models.ModuleType {
	return models.ModuleTypeIndexer
}

// GetSubType 获取模块子类�?func (im *IndexerModule) GetSubType() interface{} {
	return models.OtherModulesTypeIndexer
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (im *IndexerModule) GetPriority() int {
	return 0
}

// Stop 停止模块
func (im *IndexerModule) Stop() error {
	return nil
}

// Test 测试模块连接�?func (im *IndexerModule) Test() (bool, string) {
	// TODO: 实现站点获取逻辑
	// sites := core.SitesHelper.GetIndexers()
	// if len(sites) == 0 {
	// 	return false, "未配置站点或未通过用户认证"
	// }
	return true, ""
}

// InitSetting 模块开关设�?func (im *IndexerModule) InitSetting() (string, interface{}) {
	return "", nil
}

// searchCheck 检查是否可以执行搜�?func (im *IndexerModule) searchCheck(site map[string]interface{}, searchWord *string) bool {
	// 可能为关键字或ttxxxx
	if searchWord != nil && 
		site["language"] == "en" && 
		utils.StringUtils.IsChinese(*searchWord) {
		// 不支持中�?		// TODO: 实现日志记录
		// core.Logger.Warn(fmt.Sprintf("%v 不支持中文搜�?, site["name"]))
		fmt.Printf("%v 不支持中文搜索\n", site["name"])
		return false
	}

	// 站点流控
	// TODO: 实现站点流控检�?	// domain := utils.StringUtils.GetURLDomain(site["domain"].(string))
	// state, msg := core.SitesHelper.Check(domain)
	// if state {
	// 	core.Logger.Warn(msg)
	// 	return false
	// }

	return true
}

// clearSearchText 清理搜索文本
func (im *IndexerModule) clearSearchText(text *string) *string {
	if text == nil {
		return text
	}
	// 去除特殊字符和多余空�?	result := utils.StringUtils.Clear(*text, " ", true)
	return &result
}

// indexerStatistic 索引器统�?func (im *IndexerModule) indexerStatistic(site map[string]interface{}, errorFlag bool, seconds int) {
	// TODO: 实现站点统计
	// domain := utils.StringUtils.GetURLDomain(site["domain"].(string))
	// if errorFlag {
	// 	core.SiteOper.Fail(domain)
	// } else {
	// 	core.SiteOper.Success(domain, seconds)
	// }
}

// parseResult 解析搜索结果�?TorrentInfo 对象
func (im *IndexerModule) parseResult(site map[string]interface{}, resultArray []map[string]interface{}, seconds int) []models.TorrentInfo {
	if len(resultArray) == 0 {
		// TODO: 实现日志记录
		// core.Logger.Warn(fmt.Sprintf("%v 未搜索到数据，耗时 %d �?, site["name"], seconds))
		fmt.Printf("%v 未搜索到数据，耗时 %d 秒\n", site["name"], seconds)
		return []models.TorrentInfo{}
	}
	
	// core.Logger.Info(fmt.Sprintf("%v 搜索完成，耗时 %d 秒，返回数据�?d", site["name"], seconds, len(resultArray)))
	fmt.Printf("%v 搜索完成，耗时 %d 秒，返回数据�?d\n", site["name"], seconds, len(resultArray))
	
	results := make([]models.TorrentInfo, len(resultArray))
	for i, result := range resultArray {
		torrentInfo := models.TorrentInfo{
			// 初始化基本字�?		}
		
		// 将result中的字段复制到torrentInfo�?		for k, v := range result {
			switch k {
			case "title":
				if val, ok := v.(string); ok {
					torrentInfo.Title = val
				}
			case "description":
				if val, ok := v.(string); ok {
					torrentInfo.Description = val
				}
			case "imdbid":
				if val, ok := v.(string); ok {
					torrentInfo.Imdbid = val
				}
			case "enclosure":
				if val, ok := v.(string); ok {
					torrentInfo.Enclosure = val
				}
			case "page_url":
				if val, ok := v.(string); ok {
					torrentInfo.PageUrl = val
				}
			case "size":
				if val, ok := v.(string); ok {
					// TODO: 转换为float64
					// torrentInfo.Size = val
				}
			case "seeders":
				if val, ok := v.(int); ok {
					torrentInfo.Seeders = val
				}
			case "peers":
				if val, ok := v.(int); ok {
					torrentInfo.Peers = val
				}
			case "grabs":
				if val, ok := v.(int); ok {
					torrentInfo.Grabs = val
				}
			case "pubdate":
				if val, ok := v.(string); ok {
					torrentInfo.Pubdate = val
				}
			case "date_elapsed":
				if val, ok := v.(string); ok {
					torrentInfo.DateElapsed = val
				}
			case "freedate":
				if val, ok := v.(string); ok {
					torrentInfo.Freedate = val
				}
			case "uploadvolumefactor":
				if val, ok := v.(int); ok {
					torrentInfo.Uploadvolumefactor = float64(val)
				}
			case "downloadvolumefactor":
				if val, ok := v.(int); ok {
					torrentInfo.Downloadvolumefactor = float64(val)
				}
			case "hit_and_run":
				if val, ok := v.(bool); ok {
					torrentInfo.HitAndRun = val
				}
			case "labels":
				// TODO: 处理标签数组
			case "pri_order":
				if val, ok := v.(int); ok {
					torrentInfo.PriOrder = val
				}
			case "volume_factor":
				if val, ok := v.(string); ok {
					torrentInfo.VolumeFactor = val
				}
			case "freedate_diff":
				if val, ok := v.(string); ok {
					torrentInfo.FreedateDiff = val
				}
			}
		}
		
		// 设置站点相关信息
		if siteID, ok := site["id"].(int); ok {
			torrentInfo.Site = siteID
		}
		if siteName, ok := site["name"].(string); ok {
			torrentInfo.SiteName = siteName
		}
		if siteCookie, ok := site["cookie"].(string); ok {
			torrentInfo.SiteCookie = siteCookie
		}
		if siteUA, ok := site["ua"].(string); ok {
			torrentInfo.SiteUa = siteUA
		}
		if siteProxy, ok := site["proxy"].(bool); ok {
			torrentInfo.SiteProxy = siteProxy
		}
		if sitePri, ok := site["pri"].(int); ok {
			torrentInfo.SiteOrder = sitePri
		}
		if siteDownloader, ok := site["downloader"].(string); ok {
			torrentInfo.SiteDownloader = siteDownloader
		}
		
		results[i] = torrentInfo
	}
	
	return results
}

// SearchTorrents 搜索一个站�?func (im *IndexerModule) SearchTorrents(site map[string]interface{},
	keyword *string,
	mtype models.MediaType,
	cat *string,
	page int) []models.TorrentInfo {
	
	// 索引结果
	var result []map[string]interface{}
	// 开始计�?	startTime := time.Now()
	// 错误标志
	errorFlag := false

	// 检查是否可以执行搜�?	if !im.searchCheck(site, keyword) {
		return []models.TorrentInfo{}
	}

	// 去除搜索关键字中的特殊字�?	searchWord := im.clearSearchText(keyword)

	// 开始搜�?	// 根据不同站点解析器进行搜�?	parser, hasParser := site["parser"]
	if hasParser {
		switch parser {
		case "TNodeSpider":
			errorFlag, result = SpiderTNodeSpider(site, searchWord, page)
		case "TorrentLeech":
			errorFlag, result = SpiderTorrentLeech(site, searchWord, page)
		case "mTorrent":
			errorFlag, result = SpiderMTorrent(site, searchWord, mtype, page)
		case "Yema":
			errorFlag, result = SpiderYema(site, searchWord, mtype, page)
		case "Haidan":
			errorFlag, result = SpiderHaidan(site, searchWord, mtype)
		case "HDDolby":
			errorFlag, result = SpiderHDDolby(site, searchWord, mtype, page)
		default:
			errorFlag, result = im.spiderSearch(searchWord, site, mtype, cat, page)
		}
	} else {
		errorFlag, result = im.spiderSearch(searchWord, site, mtype, cat, page)
	}

	// 索引花费的时�?	seconds := int(time.Since(startTime).Seconds())

	// 统计索引情况
	im.indexerStatistic(site, errorFlag, seconds)

	// 返回结果
	return im.parseResult(site, result, seconds)
}

// spiderSearch 根据关键字搜索单个站�?func (im *IndexerModule) spiderSearch(searchWord *string,
	indexer map[string]interface{},
	mtype models.MediaType,
	cat *string,
	page int) (bool, []map[string]interface{}) {
	
	spider := NewSiteSpider(indexer, searchWord, mtype, cat, page)
	
	// 执行搜索
	result := spider.GetTorrents()
	
	// TODO: 判断是否有错�?	errorFlag := false
	
	return errorFlag, result
}

// RefreshTorrents 获取站点最新一页的种子
func (im *IndexerModule) RefreshTorrents(site map[string]interface{},
	keyword *string,
	cat *string,
	page int) []models.TorrentInfo {
	return im.SearchTorrents(site, keyword, "", cat, page)
}

// RefreshUserdata 刷新站点的用户数�?func (im *IndexerModule) RefreshUserdata(site map[string]interface{}) *models.SiteUserData {
	
	getSiteObj := func() SiteParserBase {
		for _, siteSchema := range im.siteSchemas {
			if schemaGetter, ok := siteSchema.(interface{ GetSchema() interface{} }); ok {
				if schemaStr, ok := schemaGetter.GetSchema().(string); ok {
					if schemaStr == site["schema"] {
						// 创建站点解析器实�?						if siteParserCreator, ok := siteSchema.(interface{
							NewSiteParser(siteName, url, siteCookie, apiKey, token, ua string, proxy bool) SiteParserBase
						}); ok {
							return siteParserCreator.NewSiteParser(
								site["name"].(string),
								site["url"].(string),
								site["cookie"].(string),
								site["apikey"].(string),
								site["token"].(string),
								site["ua"].(string),
								site["proxy"].(bool),
							)
						}
					}
				}
			}
		}
		return nil
	}

	siteObj := getSiteObj()
	if siteObj == nil {
		if public, ok := site["public"]; !ok || !public.(bool) {
			// core.Logger.Warn(fmt.Sprintf("站点 %v 未找到站点解析器，schema�?v", site["name"], site["schema"]))
			fmt.Printf("站点 %v 未找到站点解析器，schema�?v\n", site["name"], site["schema"])
		}
		return nil
	}

	// 获取用户数据
	defer func() {
		siteObj.Clear()
	}()

	// core.Logger.Info(fmt.Sprintf("站点 %v 开始以 %v 模型解析数据...", site["name"], site["schema"]))
	fmt.Printf("站点 %v 开始以 %v 模型解析数据...\n", site["name"], site["schema"])
	
	// TODO: 实现站点解析
	// siteObj.Parse()
	
	// core.Logger.Debug(fmt.Sprintf("站点 %v 数据解析完成", site["name"]))
	fmt.Printf("站点 %v 数据解析完成\n", site["name"])

	userData := models.NewSiteUserData()
	// TODO: 设置用户数据
	// userData.Domain = utils.StringUtils.GetURLDomain(site["url"].(string))
	// userData.UserID = siteObj.GetUserID()
	// userData.Username = siteObj.GetUsername()
	// userData.UserLevel = siteObj.GetUserLevel()
	// userData.JoinAt = siteObj.GetJoinAt()
	// userData.Upload = siteObj.GetUpload()
	// userData.Download = siteObj.GetDownload()
	// userData.Ratio = siteObj.GetRatio()
	// userData.Bonus = siteObj.GetBonus()
	// userData.Seeding = siteObj.GetSeeding()
	// userData.SeedingSize = siteObj.GetSeedingSize()
	
	// seedingInfo := siteObj.GetSeedingInfo()
	// if seedingInfo != nil {
	// 	userData.SeedingInfo = make([]interface{}, len(seedingInfo))
	// 	copy(userData.SeedingInfo, seedingInfo)
	// }
	
	// userData.Leeching = siteObj.GetLeeching()
	// userData.LeechingSize = siteObj.GetLeechingSize()
	// userData.MessageUnread = siteObj.GetMessageUnread()
	
	// messageUnreadContents := siteObj.GetMessageUnreadContents()
	// if messageUnreadContents != nil {
	// 	userData.MessageUnreadContents = make([]interface{}, len(messageUnreadContents))
	// 	copy(userData.MessageUnreadContents, messageUnreadContents)
	// }
	
	userData.UpdatedDay = time.Now().Format("2006-01-02")
	// userData.ErrMsg = siteObj.GetErrMsg()

	return userData
}
