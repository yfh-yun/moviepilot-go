package indexer

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	"github.com/PuerkitoBio/goquery"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// SiteSpiderImpl 站点爬虫实现
type SiteSpiderImpl struct {
	// 基本配置
	keyword  *string
	cat      *string
	mtype    models.MediaType
	page     int
	referer  *string
	
	// 索引器配�?	indexerID    int
	indexerName  string
	search       map[string]interface{}
	batch        map[string]interface{}
	browse       map[string]interface{}
	category     map[string]interface{}
	list         map[string]interface{}
	fields       map[string]interface{}
	domain       string
	resultNum    int
	timeout      int
	ua           string
	proxies      *string
	proxyServer  *string
	cookie       string
	
	// 爬取状�?	isError           bool
	torrentsInfo      map[string]interface{}
	torrentsInfoArray []map[string]interface{}
}

// NewSiteSpiderImpl 创建站点爬虫实例
func NewSiteSpiderImpl(indexer map[string]interface{}, keyword *string, mtype models.MediaType, cat *string, page int, referer *string) *SiteSpiderImpl {
	spider := &SiteSpiderImpl{
		keyword:  keyword,
		cat:      cat,
		mtype:    mtype,
		page:     page,
		referer:  referer,
		resultNum: 100,
		timeout:   15,
		isError:  false,
		torrentsInfo: make(map[string]interface{}),
		torrentsInfoArray: make([]map[string]interface{}, 0),
	}
	
	if indexer != nil {
		// 基本信息
		if id, ok := indexer["id"].(int); ok {
			spider.indexerID = id
		}
		if name, ok := indexer["name"].(string); ok {
			spider.indexerName = name
		}
		
		// 配置信息
		if search, ok := indexer["search"].(map[string]interface{}); ok {
			spider.search = search
		}
		if batch, ok := indexer["batch"].(map[string]interface{}); ok {
			spider.batch = batch
		}
		if browse, ok := indexer["browse"].(map[string]interface{}); ok {
			spider.browse = browse
		}
		if category, ok := indexer["category"].(map[string]interface{}); ok {
			spider.category = category
		}
		
		// 种子列表和字段配�?		if torrents, ok := indexer["torrents"].(map[string]interface{}); ok {
			if list, ok := torrents["list"].(map[string]interface{}); ok {
				spider.list = list
			}
			if fields, ok := torrents["fields"].(map[string]interface{}); ok {
				spider.fields = fields
			}
		}
		
		// 如果没有关键字且有浏览配置，使用浏览配置
		if keyword == nil && spider.browse != nil {
			if list, ok := spider.browse["list"].(map[string]interface{}); ok {
				spider.list = list
			}
			if fields, ok := spider.browse["fields"].(map[string]interface{}); ok {
				spider.fields = fields
			}
		}
		
		// 域名和其它配�?		if domain, ok := indexer["domain"].(string); ok {
			spider.domain = domain
			if spider.domain != "" && !strings.HasSuffix(spider.domain, "/") {
				spider.domain = spider.domain + "/"
			}
		}
		
		if resultNum, ok := indexer["result_num"].(int); ok {
			spider.resultNum = resultNum
		}
		if timeout, ok := indexer["timeout"].(int); ok {
			spider.timeout = timeout
		}
		if ua, ok := indexer["ua"].(string); ok {
			spider.ua = ua
		}
		if cookie, ok := indexer["cookie"].(string); ok {
			spider.cookie = cookie
		}
		
		// 代理配置
		if proxy, ok := indexer["proxy"].(bool); ok && proxy {
			// TODO: 设置代理配置
		}
	}
	
	return spider
}

// getSearchURL 获取搜索URL
func (s *SiteSpiderImpl) getSearchURL() string {
	// 种子搜索相对路径
	var torrentspath string
	if s.search != nil {
		if paths, ok := s.search["paths"].([]interface{}); ok {
			if len(paths) == 1 {
				if pathMap, ok := paths[0].(map[string]interface{}); ok {
					if path, ok := pathMap["path"].(string); ok {
						torrentspath = path
					}
				}
			} else {
				for _, path := range paths {
					if pathMap, ok := path.(map[string]interface{}); ok {
						if pathType, ok := pathMap["type"].(string); ok {
							if pathType == "all" && s.mtype == "" {
								if path, ok := pathMap["path"].(string); ok {
									torrentspath = path
									break
								}
							} else if pathType == "movie" && s.mtype == models.MediaTypeMovie {
								if path, ok := pathMap["path"].(string); ok {
									torrentspath = path
									break
								}
							} else if pathType == "tv" && s.mtype == models.MediaTypeTV {
								if path, ok := pathMap["path"].(string); ok {
									torrentspath = path
									break
								}
							}
						}
					}
				}
			}
		}
	}
	
	// 精确搜索
	if s.keyword != nil {
		var searchWord string
		var searchMode string
		
		// 检查是否为批量搜索
		if keywordList, ok := interface{}(*s.keyword).([]string); ok {
			// 批量查询
			if s.batch != nil {
				delimiter := " "
				if d, ok := s.batch["delimiter"].(string); ok {
					delimiter = d
				}
				spaceReplace := " "
				if sr, ok := s.batch["space_replace"].(string); ok {
					spaceReplace = sr
				}
				
				var keywords []string
				for _, k := range keywordList {
					keywords = append(keywords, strings.ReplaceAll(k, " ", spaceReplace))
				}
				searchWord = strings.Join(keywords, delimiter)
			} else {
				searchWord = strings.Join(keywordList, " ")
			}
			// 查询模式：或
			searchMode = "1"
		} else {
			// 单个查询
			searchWord = *s.keyword
			// 查询模式�?			searchMode = "0"
		}
		
		// 构建搜索URL
		if s.search != nil {
			if indexerParams, ok := s.search["params"].(map[string]interface{}); ok {
				// 复制参数
				params := make(map[string]string)
				params["search_mode"] = searchMode
				params["search_area"] = "0"
				params["page"] = fmt.Sprintf("%d", s.page)
				params["notnewword"] = "1"
				
				// search_area�?表示支持imdbid搜索
				if searchArea, ok := indexerParams["search_area"]; ok && searchArea != nil {
					if s.keyword == nil || !strings.HasPrefix(*s.keyword, "tt") {
						// 支持imdbid搜索，但关键字不是imdbid时，不启用imdbid搜索
						// 不添加search_area参数
					} else {
						if sa, ok := searchArea.(string); ok {
							params["search_area"] = sa
						} else if sa, ok := searchArea.(int); ok {
							params["search_area"] = fmt.Sprintf("%d", sa)
						} else if sa, ok := searchArea.(float64); ok {
							params["search_area"] = fmt.Sprintf("%.0f", sa)
						}
					}
				}
				
				// 额外参数
				for key, value := range indexerParams {
					if key != "search_area" {
						if strValue, ok := value.(string); ok {
							// 简单的字符串替换，模拟Python中的.format(**inputs_dict)
							strValue = strings.ReplaceAll(strValue, "{keyword}", searchWord)
							params[key] = strValue
						} else if intValue, ok := value.(int); ok {
							params[key] = fmt.Sprintf("%d", intValue)
						} else if floatValue, ok := value.(float64); ok {
							params[key] = fmt.Sprintf("%.0f", floatValue)
						} else {
							params[key] = fmt.Sprintf("%v", value)
						}
					}
				}
				
				// 分类条件
				if s.category != nil {
					var cats []interface{}
					if s.mtype == models.MediaTypeTV {
						if tv, ok := s.category["tv"].([]interface{}); ok {
							cats = tv
						}
					} else if s.mtype == models.MediaTypeMovie {
						if movie, ok := s.category["movie"].([]interface{}); ok {
							cats = movie
						}
					} else {
						if movie, ok := s.category["movie"].([]interface{}); ok {
							cats = append(cats, movie...)
						}
						if tv, ok := s.category["tv"].([]interface{}); ok {
							cats = append(cats, tv...)
						}
					}
					
					allowedCats := make(map[string]bool)
					if s.cat != nil {
						for _, c := range strings.Split(*s.cat, ",") {
							allowedCats[strings.TrimSpace(c)] = true
						}
					}
					
					for _, cat := range cats {
						if catMap, ok := cat.(map[string]interface{}); ok {
							if catID, ok := catMap["id"].(string); ok {
								if len(allowedCats) == 0 || allowedCats[catID] {
									if field, ok := s.category["field"].(string); ok {
										if delimiter, ok := s.category["delimiter"].(string); ok {
											if val, exists := params[field]; exists {
												params[field] = val + delimiter + catID
											} else {
												params[field] = catID
											}
										}
									} else {
										params[fmt.Sprintf("cat%s", catID)] = "1"
									}
								}
							}
						}
					}
				}
				
				// 构建URL
				baseURL := s.domain + torrentspath
				query := url.Values{}
				for k, v := range params {
					query.Set(k, v)
				}
				return baseURL + "?" + query.Encode()
			} else {
				// 无额外参�?				inputsDict := map[string]string{
					"keyword": url.QueryEscape(searchWord),
					"page":    fmt.Sprintf("%d", s.page),
				}
				
				// 替换路径中的占位�?				path := torrentspath
				for k, v := range inputsDict {
					placeholder := fmt.Sprintf("{%s}", k)
					path = strings.ReplaceAll(path, placeholder, v)
				}
				
				return s.domain + path
			}
		}
	} else {
		// 列表浏览
		inputsDict := map[string]string{
			"page":    fmt.Sprintf("%d", s.page),
			"keyword": "",
		}
		
		var path string
		if s.browse != nil {
			// 有单独浏览路�?			if p, ok := s.browse["path"].(string); ok {
				path = p
			}
			
			if start, ok := s.browse["start"].(int); ok {
				startPage := start + s.page
				inputsDict["page"] = fmt.Sprintf("%d", startPage)
			}
		} else if s.page > 0 {
			path = torrentspath + fmt.Sprintf("?page=%d", s.page)
		} else {
			path = torrentspath
		}
		
		// 替换路径中的占位�?		for k, v := range inputsDict {
			placeholder := fmt.Sprintf("{%s}", k)
			path = strings.ReplaceAll(path, placeholder, v)
		}
		
		return s.domain + path
	}
	
	return ""
}

// GetTorrents 开始请�?func (s *SiteSpiderImpl) GetTorrents() []map[string]interface{} {
	if s.search == nil || s.domain == "" {
		return []map[string]interface{}{}
	}
	
	// 获取搜索URL
	searchURL := s.getSearchURL()
	
	// TODO: 实际的HTTP请求和HTML解析
	// 这里应该使用utils.HttpUtils发送请求并解析返回的HTML
	// 暂时返回空结�?	
	return []map[string]interface{}{}
}

// parse 解析整个页面
func (s *SiteSpiderImpl) parse(htmlText string) []map[string]interface{} {
	if htmlText == "" {
		s.isError = true
		return []map[string]interface{}{}
	}
	
	// 清空旧结�?	s.torrentsInfoArray = []map[string]interface{}{}
	
	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		s.isError = true
		return []map[string]interface{}{}
	}
	
	// 种子筛选器
	torrentsSelector := ""
	if s.list != nil {
		if selector, ok := s.list["selector"].(string); ok {
			torrentsSelector = selector
		}
	}
	
	// 遍历种子HTML列表
	doc.Find(torrentsSelector).EachWithBreak(func(i int, selection *goquery.Selection) bool {
		if i >= s.resultNum {
			return false // 停止遍历
		}
		
		// 获取种子信息
		torrentInfo := s.getInfo(selection)
		if len(torrentInfo) > 0 {
			s.torrentsInfoArray = append(s.torrentsInfoArray, torrentInfo)
		}
		
		return true
	})
	
	// 返回数组的副�?	result := make([]map[string]interface{}, len(s.torrentsInfoArray))
	copy(result, s.torrentsInfoArray)
	
	return result
}

// getInfo 解析单条种子数据
func (s *SiteSpiderImpl) getInfo(torrent *goquery.Selection) map[string]interface{} {
	// 每次调用时重新初始化，避免数据累�?	s.torrentsInfo = make(map[string]interface{})
	
	// 标题
	s.getTitle(torrent)
	
	// 描述
	s.getDescription(torrent)
	
	// 详情页面
	s.getDetail(torrent)
	
	// 下载链接
	s.getDownload(torrent)
	
	// 完成�?	s.getGrabs(torrent)
	
	// 下载�?	s.getLeechers(torrent)
	
	// 做种�?	s.getSeeders(torrent)
	
	// 大小
	s.getSize(torrent)
	
	// IMDBID
	s.getIMDbID(torrent)
	
	// 下载系数
	s.getDownloadVolumeFactor(torrent)
	
	// 上传系数
	s.getUploadVolumeFactor(torrent)
	
	// 发布时间
	s.getPubDate(torrent)
	
	// 已发布时�?	s.getDateElapsed(torrent)
	
	// 免费截止时间
	s.getFreeDate(torrent)
	
	// 标签
	s.getLabels(torrent)
	
	// HR
	s.getHitAndRun(torrent)
	
	// 分类
	s.getCategory(torrent)
	
	// 返回当前种子信息的副�?	result := make(map[string]interface{})
	for k, v := range s.torrentsInfo {
		result[k] = v
	}
	
	return result
}

// getTitle 获取标题
func (s *SiteSpiderImpl) getTitle(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// title default text
	if _, exists := s.fields["title"]; !exists {
		return
	}
	
	selector := s.fields["title"].(map[string]interface{})
	if _, exists := selector["selector"]; exists {
		s.torrentsInfo["title"] = s.safeQuery(torrent, selector)
	} else if _, exists := selector["text"]; exists {
		renderDict := make(map[string]string)
		if _, exists := s.fields["title_default"]; exists {
			titleDefaultSelector := s.fields["title_default"].(map[string]interface{})
			titleDefault := s.safeQuery(torrent, titleDefaultSelector)
			renderDict["title_default"] = titleDefault
		}
		if _, exists := s.fields["title_optional"]; exists {
			titleOptionalSelector := s.fields["title_optional"].(map[string]interface{})
			titleOptional := s.safeQuery(torrent, titleOptionalSelector)
			renderDict["title_optional"] = titleOptional
		}
		
		// 简单的模板替换
		text := selector["text"].(string)
		for k, v := range renderDict {
			placeholder := fmt.Sprintf("{%s}", k)
			text = strings.ReplaceAll(text, placeholder, v)
		}
		s.torrentsInfo["title"] = s.filterText(text, selector)
	} else {
		s.torrentsInfo["title"] = s.filterText(s.torrentsInfo["title"].(string), selector)
	}
}

// getDescription 获取描述
func (s *SiteSpiderImpl) getDescription(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// description text
	if _, exists := s.fields["description"]; !exists {
		return
	}
	
	selector := s.fields["description"].(map[string]interface{})
	if _, exists := selector["selector"]; exists || _, exists := selector["selectors"]; exists {
		// 对于selectors情况，需要特殊处理selector_config
		descSelector := make(map[string]interface{})
		for k, v := range selector {
			descSelector[k] = v
		}
		
		if _, exists := selector["selectors"]; exists && _, exists2 := selector["selector"]; !exists2 {
			descSelector["selector"] = selector["selectors"]
		}
		s.torrentsInfo["description"] = s.safeQuery(torrent, descSelector)
	} else if _, exists := selector["text"]; exists {
		renderDict := make(map[string]string)
		if _, exists := s.fields["tags"]; exists {
			tagsSelector := s.fields["tags"].(map[string]interface{})
			tag := s.safeQuery(torrent, tagsSelector)
			renderDict["tags"] = tag
		}
		if _, exists := s.fields["subject"]; exists {
			subjectSelector := s.fields["subject"].(map[string]interface{})
			subject := s.safeQuery(torrent, subjectSelector)
			renderDict["subject"] = subject
		}
		if _, exists := s.fields["description_free_forever"]; exists {
			descriptionFreeForeverSelector := s.fields["description_free_forever"].(map[string]interface{})
			descriptionFreeForever := s.safeQuery(torrent, descriptionFreeForeverSelector)
			renderDict["description_free_forever"] = descriptionFreeForever
		}
		if _, exists := s.fields["description_normal"]; exists {
			descriptionNormalSelector := s.fields["description_normal"].(map[string]interface{})
			descriptionNormal := s.safeQuery(torrent, descriptionNormalSelector)
			renderDict["description_normal"] = descriptionNormal
		}
		
		// 简单的模板替换
		text := selector["text"].(string)
		for k, v := range renderDict {
			placeholder := fmt.Sprintf("{%s}", k)
			text = strings.ReplaceAll(text, placeholder, v)
		}
		s.torrentsInfo["description"] = s.filterText(text, selector)
	} else {
		s.torrentsInfo["description"] = s.filterText(s.torrentsInfo["description"].(string), selector)
	}
}

// getDetail 获取详情页面
func (s *SiteSpiderImpl) getDetail(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// details page text
	if _, exists := s.fields["details"]; !exists {
		return
	}
	
	selector := s.fields["details"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	detailLink := s.filterText(item, selector)
	
	if detailLink != "" {
		if !strings.HasPrefix(detailLink, "http") {
			if strings.HasPrefix(detailLink, "//") {
				parts := strings.Split(s.domain, ":")
				s.torrentsInfo["page_url"] = parts[0] + ":" + detailLink
			} else if strings.HasPrefix(detailLink, "/") {
				s.torrentsInfo["page_url"] = s.domain + detailLink[1:]
			} else {
				s.torrentsInfo["page_url"] = s.domain + detailLink
			}
		} else {
			s.torrentsInfo["page_url"] = detailLink
		}
	}
}

// getDownload 获取下载链接
func (s *SiteSpiderImpl) getDownload(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// download link text
	if _, exists := s.fields["download"]; !exists {
		return
	}
	
	selector := s.fields["download"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	downloadLink := s.filterText(item, selector)
	
	if downloadLink != "" {
		if !strings.HasPrefix(downloadLink, "http") && !strings.HasPrefix(downloadLink, "magnet") {
			// TODO: 实现StringUtils.get_url_netloc功能
			// _scheme, _domain := StringUtils.get_url_netloc(s.domain)
			// 暂时使用简单处�?			_domain := s.domain
			if strings.Contains(_domain, downloadLink) {
				if strings.HasPrefix(downloadLink, "/") {
					// s.torrentsInfo["enclosure"] = f"{_scheme}:{downloadLink}"
					s.torrentsInfo["enclosure"] = downloadLink
				} else {
					// s.torrentsInfo["enclosure"] = f"{_scheme}://{downloadLink}"
					s.torrentsInfo["enclosure"] = downloadLink
				}
			} else {
				if strings.HasPrefix(downloadLink, "/") {
					s.torrentsInfo["enclosure"] = s.domain + downloadLink[1:]
				} else {
					s.torrentsInfo["enclosure"] = s.domain + downloadLink
				}
			}
		} else {
			s.torrentsInfo["enclosure"] = downloadLink
		}
	}
}

// getIMDbID 获取IMDb ID
func (s *SiteSpiderImpl) getIMDbID(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// imdbid
	if _, exists := s.fields["imdbid"]; !exists {
		return
	}
	
	selector := s.fields["imdbid"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	s.torrentsInfo["imdbid"] = s.filterText(item, selector)
}

// getSize 获取种子大小
func (s *SiteSpiderImpl) getSize(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent size int
	if _, exists := s.fields["size"]; !exists {
		return
	}
	
	selector := s.fields["size"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	if item != "" {
		sizeVal := strings.ReplaceAll(item, "\n", "")
		sizeVal = strings.TrimSpace(sizeVal)
		sizeVal = s.filterText(sizeVal, selector)
		// TODO: 实现StringUtils.num_filesize功能
		// s.torrentsInfo["size"] = StringUtils.num_filesize(sizeVal)
		s.torrentsInfo["size"] = sizeVal
	} else {
		s.torrentsInfo["size"] = 0
	}
}

// getLeechers 获取下载�?func (s *SiteSpiderImpl) getLeechers(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent leechers int
	if _, exists := s.fields["leechers"]; !exists {
		return
	}
	
	selector := s.fields["leechers"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	if item != "" {
		parts := strings.Split(item, "/")
		peersVal := strings.ReplaceAll(parts[0], ",", "")
		peersVal = s.filterText(peersVal, selector)
		if peersVal != "" {
			if val, err := strconv.Atoi(peersVal); err == nil {
				s.torrentsInfo["peers"] = val
			} else {
				s.torrentsInfo["peers"] = 0
			}
		} else {
			s.torrentsInfo["peers"] = 0
		}
	} else {
		s.torrentsInfo["peers"] = 0
	}
}

// getSeeders 获取做种�?func (s *SiteSpiderImpl) getSeeders(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent seeders int
	if _, exists := s.fields["seeders"]; !exists {
		return
	}
	
	selector := s.fields["seeders"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	if item != "" {
		parts := strings.Split(item, "/")
		seedersVal := strings.ReplaceAll(parts[0], ",", "")
		seedersVal = s.filterText(seedersVal, selector)
		if seedersVal != "" {
			if val, err := strconv.Atoi(seedersVal); err == nil {
				s.torrentsInfo["seeders"] = val
			} else {
				s.torrentsInfo["seeders"] = 0
			}
		} else {
			s.torrentsInfo["seeders"] = 0
		}
	} else {
		s.torrentsInfo["seeders"] = 0
	}
}

// getGrabs 获取完成�?func (s *SiteSpiderImpl) getGrabs(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent grabs int
	if _, exists := s.fields["grabs"]; !exists {
		return
	}
	
	selector := s.fields["grabs"].(map[string]interface{})
	item := s.safeQuery(torrent, selector)
	if item != "" {
		parts := strings.Split(item, "/")
		grabsVal := strings.ReplaceAll(parts[0], ",", "")
		grabsVal = s.filterText(grabsVal, selector)
		if grabsVal != "" {
			if val, err := strconv.Atoi(grabsVal); err == nil {
				s.torrentsInfo["grabs"] = val
			} else {
				s.torrentsInfo["grabs"] = 0
			}
		} else {
			s.torrentsInfo["grabs"] = 0
		}
	} else {
		s.torrentsInfo["grabs"] = 0
	}
}

// getPubDate 获取发布时间
func (s *SiteSpiderImpl) getPubDate(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent pubdate yyyy-mm-dd hh:mm:ss
	if _, exists := s.fields["date_added"]; !exists {
		return
	}
	
	selector := s.fields["date_added"].(map[string]interface{})
	pubdateStr := s.safeQuery(torrent, selector)
	if pubdateStr != "" {
		pubdateStr = strings.ReplaceAll(pubdateStr, "\n", " ")
		pubdateStr = strings.TrimSpace(pubdateStr)
	}
	s.torrentsInfo["pubdate"] = s.filterText(pubdateStr, selector)
}

// getDateElapsed 获取已发布时�?func (s *SiteSpiderImpl) getDateElapsed(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// torrent date elapsed text
	if _, exists := s.fields["date_elapsed"]; !exists {
		return
	}
	
	selector := s.fields["date_elapsed"].(map[string]interface{})
	dateElapsed := s.safeQuery(torrent, selector)
	s.torrentsInfo["date_elapsed"] = s.filterText(dateElapsed, selector)
}

// getDownloadVolumeFactor 获取下载系数
func (s *SiteSpiderImpl) getDownloadVolumeFactor(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	selector, exists := s.fields["downloadvolumefactor"]
	if !exists {
		return
	}
	
	selectorMap := selector.(map[string]interface{})
	s.torrentsInfo["downloadvolumefactor"] = 1
	
	if _, exists := selectorMap["case"]; exists {
		// TODO: 实现case逻辑
	} else if _, exists := selectorMap["selector"]; exists {
		item := s.safeQuery(torrent, selectorMap)
		if item != "" {
			// 使用正则表达式查找数�?			re := regexp.MustCompile(`(\d+\.?\d*)`)
			matches := re.FindStringSubmatch(item)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					s.torrentsInfo["downloadvolumefactor"] = val
				}
			}
		}
	}
}

// getUploadVolumeFactor 获取上传系数
func (s *SiteSpiderImpl) getUploadVolumeFactor(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	selector, exists := s.fields["uploadvolumefactor"]
	if !exists {
		return
	}
	
	selectorMap := selector.(map[string]interface{})
	s.torrentsInfo["uploadvolumefactor"] = 1
	
	if _, exists := selectorMap["case"]; exists {
		// TODO: 实现case逻辑
	} else if _, exists := selectorMap["selector"]; exists {
		item := s.safeQuery(torrent, selectorMap)
		if item != "" {
			// 使用正则表达式查找数�?			re := regexp.MustCompile(`(\d+\.?\d*)`)
			matches := re.FindStringSubmatch(item)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					s.torrentsInfo["uploadvolumefactor"] = val
				}
			}
		}
	}
}

// getLabels 获取标签
func (s *SiteSpiderImpl) getLabels(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// labels ['label1', 'label2']
	if _, exists := s.fields["labels"]; !exists {
		return
	}
	
	selector := s.fields["labels"].(map[string]interface{})
	if _, exists := selector["selector"]; !exists {
		s.torrentsInfo["labels"] = []string{}
		return
	}
	
	// TODO: 实现标签提取逻辑
	s.torrentsInfo["labels"] = []string{}
}

// getFreeDate 获取免费截止时间
func (s *SiteSpiderImpl) getFreeDate(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// free date yyyy-mm-dd hh:mm:ss
	if _, exists := s.fields["freedate"]; !exists {
		return
	}
	
	selector := s.fields["freedate"].(map[string]interface{})
	freedate := s.safeQuery(torrent, selector)
	s.torrentsInfo["freedate"] = s.filterText(freedate, selector)
}

// getHitAndRun 获取HR标记
func (s *SiteSpiderImpl) getHitAndRun(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// hitandrun True/False
	if _, exists := s.fields["hr"]; !exists {
		return
	}
	
	selector := s.fields["hr"].(map[string]interface{})
	// TODO: 实现HR检测逻辑
	s.torrentsInfo["hit_and_run"] = false
}

// getCategory 获取分类
func (s *SiteSpiderImpl) getCategory(torrent *goquery.Selection) {
	if s.fields == nil {
		return
	}
	
	// category 电影/电视�?	if _, exists := s.fields["category"]; !exists {
		return
	}
	
	selector := s.fields["category"].(map[string]interface{})
	categoryValue := s.safeQuery(torrent, selector)
	categoryValue = s.filterText(categoryValue, selector)
	
	if categoryValue != "" && s.category != nil {
		var tvCats []string
		var movieCats []string
		
		if tv, ok := s.category["tv"].([]interface{}); ok {
			for _, cat := range tv {
				if catMap, ok := cat.(map[string]interface{}); ok {
					if id, ok := catMap["id"].(string); ok {
						tvCats = append(tvCats, id)
					}
				}
			}
		}
		
		if movie, ok := s.category["movie"].([]interface{}); ok {
			for _, cat := range movie {
				if catMap, ok := cat.(map[string]interface{}); ok {
					if id, ok := catMap["id"].(string); ok {
						movieCats = append(movieCats, id)
					}
				}
			}
		}
		
		isTVCat := false
		isMovieCat := false
		
		for _, cat := range tvCats {
			if cat == categoryValue {
				isTVCat = true
				break
			}
		}
		
		for _, cat := range movieCats {
			if cat == categoryValue {
				isMovieCat = true
				break
			}
		}
		
		if isTVCat && !isMovieCat {
			s.torrentsInfo["category"] = string(models.MediaTypeTV)
		} else if isMovieCat {
			s.torrentsInfo["category"] = string(models.MediaTypeMovie)
		} else {
			s.torrentsInfo["category"] = string(models.MediaTypeUnknown)
		}
	} else {
		s.torrentsInfo["category"] = string(models.MediaTypeUnknown)
	}
}

// safeQuery 安全地执行查询并自动清理资源
func (s *SiteSpiderImpl) safeQuery(torrent *goquery.Selection, selectorConfig map[string]interface{}) string {
	if selectorConfig == nil {
		return ""
	}
	
	if selector, exists := selectorConfig["selector"]; !exists || selector == nil {
		return ""
	}
	
	selectorStr := selector.(string)
	// 在当前选择器中查找元素
	selected := torrent.Find(selectorStr)
	
	// 移除元素
	s.remove(selected, selectorConfig)
	
	// 获取属性或文本
	items := s.attributeOrText(selected, selectorConfig)
	
	// 获取索引
	return s.index(items, selectorConfig)
}

// remove 移除元素
func (s *SiteSpiderImpl) remove(item *goquery.Selection, selector map[string]interface{}) {
	if selector == nil {
		return
	}
	
	if remove, exists := selector["remove"]; exists {
		removeStr := remove.(string)
		removelist := strings.Split(removeStr, ", ")
		for _, v := range removelist {
			item.RemoveFiltered(v)
		}
	}
}

// attributeOrText 获取属性或文本
func (s *SiteSpiderImpl) attributeOrText(item *goquery.Selection, selector map[string]interface{}) []string {
	if selector == nil {
		text := item.Text()
		return []string{text}
	}
	
	if item.Length() == 0 {
		return []string{}
	}
	
	if attribute, exists := selector["attribute"]; exists {
		attr := attribute.(string)
		var items []string
		item.Each(func(i int, selection *goquery.Selection) {
			if val, exists := selection.Attr(attr); exists {
				items = append(items, val)
			}
		})
		return items
	} else {
		var items []string
		item.Each(func(i int, selection *goquery.Selection) {
			items = append(items, selection.Text())
		})
		return items
	}
}

// index 获取索引
func (s *SiteSpiderImpl) index(items []string, selector map[string]interface{}) string {
	if len(items) == 0 {
		return ""
	}
	
	if selector != nil {
		if contents, exists := selector["contents"]; exists {
			contentsIdx := int(contents.(float64))
			if len(items) > contentsIdx {
				lines := strings.Split(items[0], "\n")
				if len(lines) > contentsIdx {
					return lines[contentsIdx]
				}
			}
		} else if index, exists := selector["index"]; exists {
			indexIdx := int(index.(float64))
			if len(items) > indexIdx {
				return items[indexIdx]
			}
		} else {
			return items[0]
		}
	} else {
		return items[0]
	}
	
	return ""
}

// filterText 对文本进行处�?func (s *SiteSpiderImpl) filterText(text string, filters map[string]interface{}) string {
	if text == "" {
		return text
	}
	
	if filters == nil {
		return text
	}
	
	if filtersList, exists := filters["filters"]; exists {
		if filterArray, ok := filtersList.([]interface{}); ok {
			for _, filterItem := range filterArray {
				if filterMap, ok := filterItem.(map[string]interface{}); ok {
					if text == "" {
						break
					}
					
					methodName, exists := filterMap["name"].(string)
					if !exists {
						continue
					}
					
					switch methodName {
					case "re_search":
						if args, exists := filterMap["args"].([]interface{}); exists && len(args) >= 2 {
							if pattern, ok := args[0].(string); ok {
								if group, ok := args[len(args)-1].(int); ok {
									re := regexp.MustCompile(pattern)
									matches := re.FindStringSubmatch(text)
									if len(matches) > group {
										text = matches[group]
									}
								}
							}
						}
					case "split":
						if args, exists := filterMap["args"].([]interface{}); exists && len(args) >= 2 {
							if delimiter, ok := args[0].(string); ok {
								if index, ok := args[len(args)-1].(int); ok {
									parts := strings.Split(text, delimiter)
									if len(parts) > index {
										text = parts[index]
									}
								}
							}
						}
					case "replace":
						if args, exists := filterMap["args"].([]interface{}); exists && len(args) >= 2 {
							if old, ok := args[0].(string); ok {
								if new, ok := args[len(args)-1].(string); ok {
									text = strings.ReplaceAll(text, old, new)
								}
							}
						}
					case "dateparse":
						if args, exists := filterMap["args"].(string); exists {
							text = strings.ReplaceAll(text, "\n", " ")
							text = strings.TrimSpace(text)
							// TODO: 实现日期解析
							// text = datetime.datetime.strptime(text, r"%s" % args)
						}
					case "strip":
						text = strings.TrimSpace(text)
					case "appendleft":
						if args, exists := filterMap["args"].(string); exists {
							text = args + text
						}
					case "querystring":
						if args, exists := filterMap["args"].(string); exists {
							// 解析URL并提取查询参�?							parsedURL, err := url.Parse(text)
							if err == nil {
								queryParams := parsedURL.Query()
								if val := queryParams.Get(args); val != "" {
									text = val
								}
							}
						}
					}
				}
			}
		}
	}
	
	return strings.TrimSpace(text)
}
