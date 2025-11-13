package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"moviepilot-go/internal/core/context"
	"moviepilot-go/internal/core/metainfo"
	"moviepilot-go/internal/helper/rule"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils"
)

// FilterModule 过滤器模�?type FilterModule struct {
	parser     *RuleParser
	media      *context.MediaInfo
	rulehelper *rule.RuleHelper
	ruleSet    map[string]map[string]interface{}
}

// NewFilterModule 创建过滤器模块实�?func NewFilterModule() *FilterModule {
	f := &FilterModule{
		rulehelper: rule.NewRuleHelper(),
	}
	
	// 初始化内置规则集
	f.ruleSet = map[string]map[string]interface{}{
		// 蓝光原盘
		"BLU": {
			"include": []string{`(?i)(\bBlu-?Ray\b.*\b(?:VC-?1|AVC|MPEG-?2)\b|\b(?:UHD|4K|2160p)\b(?:.*Blu-?Ray)?.*\b(?:HEVC|H\.?265)\b|\bBlu-?Ray\b.*\b(?:UHD|4K|2160p)\b.*\b(?:HEVC|H\.?265)\b|\b(?:COMPLETE|FULL)\b.*\b(?:(?:UHD|4K|2160p)\b.*)?Blu-?Ray\b|\b(BD25|BD50|BD66|BD100|BDMV|MiniBD)\b)`},
			"exclude": []string{`(?i)(\b[XH]\.?264\b|\b[XH]\.?265\b|\bWEB-?DL\b|\bWEB-?RIP\b|\bHDTV(?:RIP)?\b|\bREMUX\b|\bBDRip\b|\bBRRip\b|\bHDRip\b|\bENCODE\b|\b(?<!WEB-|HDTV)RIP\b)`},
		},
		// 4K
		"4K": {
			"include": []string{"4k|2160p|x2160"},
			"exclude": []string{},
		},
		// 1080P
		"1080P": {
			"include": []string{"1080[pi]|x1080"},
			"exclude": []string{},
		},
		// 720P
		"720P": {
			"include": []string{"720[pi]|x720"},
			"exclude": []string{},
		},
		// 中字
		"CNSUB": {
			"include": []string{
				`[中国國繁简](/|\s|\\|\|)?[繁简英粤]|[英简繁](/|\s|\\|\|)?[中繁简]` +
					`|繁體|简体|[中国國][字配]|国语|國語|中文|中字|简日|繁日|简繁|繁体` +
					`|([\s,.-\[])(CHT|CHS|cht|chs)(|[\s,.-\]])`,
			},
			"exclude": []string{},
			"tmdb": map[string]string{
				"original_language": "zh,cn",
			},
		},
		// 官种
		"GZ": {
			"include": []string{"官方", "官种", "官组"},
			"match":   []string{"labels"},
		},
		// 特效字幕
		"SPECSUB": {
			"include": []string{"特效"},
			"exclude": []string{},
		},
		// BluRay
		"BLURAY": {
			"include": []string{`Blu-?Ray`},
			"exclude": []string{},
		},
		// UHD
		"UHD": {
			"include": []string{"UHD|UltraHD"},
			"exclude": []string{},
		},
		// H265
		"H265": {
			"include": []string{`[Hx].?265|HEVC`},
			"exclude": []string{},
		},
		// H264
		"H264": {
			"include": []string{`[Hx].?264|AVC`},
			"exclude": []string{},
		},
		// 杜比视界
		"DOLBY": {
			"include": []string{"Dolby[\\s.]+Vision|DOVI|[\\s.]+DV[\\s.]+|杜比视界"},
			"exclude": []string{},
		},
		// 杜比全景�?		"ATMOS": {
			"include": []string{"Dolby[\\s.+]+Atmos|Atmos|杜比全景[声聲]"},
			"exclude": []string{},
		},
		// HDR
		"HDR": {
			"include": []string{"[\\s.]+HDR[\\s.]+|HDR10|HDR10\\+"},
			"exclude": []string{},
		},
		// SDR
		"SDR": {
			"include": []string{"[\\s.]+SDR[\\s.]+"},
			"exclude": []string{},
		},
		// 重编�?		"REMUX": {
			"include": []string{"REMUX"},
			"exclude": []string{},
		},
		// WEB-DL
		"WEBDL": {
			"include": []string{"WEB-?DL|WEB-?RIP"},
			"exclude": []string{},
		},
		// 免费
		"FREE": {
			"downloadvolumefactor": 0,
		},
		// 国语配音
		"CNVOI": {
			"include": []string{`[国國][语語]配音|[国國]配|[国國][语語]`},
			"exclude": []string{},
			"tmdb": map[string]string{
				"original_language": "zh",
			},
		},
		// 粤语配音
		"HKVOI": {
			"include": []string{"粤语配音|粤语"},
			"exclude": []string{},
		},
		// 60FPS
		"60FPS": {
			"include": []string{"60fps|60�?},
			"exclude": []string{},
		},
		// 3D
		"3D": {
			"include": []string{"3D"},
			"exclude": []string{},
		},
	}
	
	return f
}

// InitModule 初始化模�?func (f *FilterModule) InitModule() {
	f.parser = GetRuleParser()
	f.initCustomRules()
}

// initCustomRules 加载用户自定义规�?func (f *FilterModule) initCustomRules() {
	// 加载用户自定义规则，如跟内置规则冲突，以用户自定义规则为�?	customRules := f.rulehelper.GetCustomRules()
	for _, rule := range customRules {
		logger.Infof("加载自定义规�?%s - %s", rule.ID, rule.Name)
		f.ruleSet[rule.ID] = rule.ToMap()
	}
}

// GetName 获取模块名称
func (f *FilterModule) GetName() string {
	return "过滤�?
}

// GetType 获取模块类型
func (f *FilterModule) GetType() types.ModuleType {
	return types.ModuleTypeOther
}

// GetSubtype 获取模块子类�?func (f *FilterModule) GetSubtype() types.OtherModulesType {
	return types.OtherModulesTypeFilter
}

// GetPriority 获取模块优先�?func (f *FilterModule) GetPriority() int {
	// 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?	return 4
}

// Stop 停止模块
func (f *FilterModule) Stop() {
	// 空实�?}

// Test 测试模块
func (f *FilterModule) Test() {
	// 空实�?}

// InitSetting 初始化设�?func (f *FilterModule) InitSetting() (string, interface{}) {
	// 空实�?	return "", nil
}

// FilterTorrents 过滤种子资源
func (f *FilterModule) FilterTorrents(ruleGroups []string,
	torrentList []context.TorrentInfo,
	mediainfo *context.MediaInfo) []context.TorrentInfo {
	
	// 过滤种子资源
	if len(ruleGroups) == 0 {
		return torrentList
	}
	
	f.media = mediainfo
	// 重新加载自定义规�?	f.initCustomRules()
	// 查询规则表详�?	groups := f.rulehelper.GetRuleGroupByMedia(mediainfo, ruleGroups)
	if len(groups) > 0 {
		for _, group := range groups {
			// 过滤种子
			torrentList = f.filterTorrents(
				group.RuleString,
				group.Name,
				torrentList,
			)
		}
	}
	return torrentList
}

// filterTorrents 过滤种子
func (f *FilterModule) filterTorrents(ruleString string, ruleName string,
	torrentList []context.TorrentInfo) []context.TorrentInfo {
	
	// 返回种子列表
	var retTorrents []context.TorrentInfo
	for _, torrent := range torrentList {
		// 能命中优先级的才返回
		if f.getOrder(&torrent, ruleString) == nil {
			logger.Debugf("种子 %s - %s %s 不匹�?%s 过滤规则",
				torrent.SiteName, torrent.Title, torrent.Description, ruleName)
			continue
		}
		retTorrents = append(retTorrents, torrent)
	}
	
	return retTorrents
}

// getOrder 获取种子匹配的规则优先级
func (f *FilterModule) getOrder(torrent *context.TorrentInfo, ruleStr string) *context.TorrentInfo {
	// 获取种子匹配的规则优先级，值越大越优先，未匹配时返回None
	// 多级规则
	ruleGroups := strings.Split(ruleStr, ">")
	// 优先�?	resOrder := 100
	// 是否匹配
	matched := false
	
	for _, ruleGroup := range ruleGroups {
		// 解析规则�?		parsedGroup := f.parser.Parse(strings.TrimSpace(ruleGroup))
		if len(parsedGroup) > 0 {
			if f.matchGroup(torrent, parsedGroup[0]) {
				// 出现匹配时中�?				matched = true
				logger.Debugf("种子 %s - %s 优先级为 %d",
					torrent.SiteName, torrent.Title, 100-resOrder+1)
				torrent.PriOrder = resOrder
				break
			}
		}
		// 优先级降低，继续匹配
		resOrder--
	}
	
	if !matched {
		return nil
	}
	return torrent
}

// matchGroup 判断种子是否匹配规则�?func (f *FilterModule) matchGroup(torrent *context.TorrentInfo, ruleGroup interface{}) bool {
	// 判断种子是否匹配规则�?	switch v := ruleGroup.(type) {
	case string:
		// 不是列表，说明是规则名称
		return f.matchRule(torrent, v)
	case []interface{}:
		if len(v) == 1 {
			// 只有一个规则项
			return f.matchGroup(torrent, v[0])
		} else if len(v) >= 2 {
			if str, ok := v[0].(string); ok && str == "not" {
				// 非操�?				var subGroup []interface{}
				if len(v) > 1 {
					subGroup = v[1:]
				}
				return !f.matchGroup(torrent, subGroup)
			} else if len(v) >= 3 {
				if op, ok := v[1].(string); ok {
					switch op {
					case "and":
						// 与操�?						return f.matchGroup(torrent, v[0]) && f.matchGroup(torrent, v[2])
					case "or":
						// 或操�?						return f.matchGroup(torrent, v[0]) || f.matchGroup(torrent, v[2])
					}
				}
			}
		}
		// 处理特殊的表达式格式
		if len(v) >= 3 {
			// 检查是否是 [expr, "and"/"or", expr] 格式
			if op, ok := v[1].(string); ok && (op == "and" || op == "or") {
				left := v[0]
				right := v[2]
				if op == "and" {
					return f.matchGroup(torrent, left) && f.matchGroup(torrent, right)
				} else {
					return f.matchGroup(torrent, left) || f.matchGroup(torrent, right)
				}
			}
		}
	}
	return false
}

// matchRule 判断种子是否匹配规则�?func (f *FilterModule) matchRule(torrent *context.TorrentInfo, ruleName string) bool {
	// 判断种子是否匹配规则�?	rule, exists := f.ruleSet[ruleName]
	if !exists {
		// 规则不存�?		logger.Debugf("规则 %s 不存�?, ruleName)
		return false
	}
	
	// TMDB规则
	tmdb, hasTMDB := rule["tmdb"].(map[string]string)
	// 符合TMDB规则的直接返回True，即不过�?	if hasTMDB && f.matchTMDB(tmdb) {
		logger.Debugf("种子 %s - %s 符合 %s 的TMDB规则，匹配成�?,
			torrent.SiteName, torrent.Title, ruleName)
		return true
	}
	
	// 匹配项：标题、副标题、标�?	content := fmt.Sprintf("%s %s %s", torrent.Title, torrent.Description, strings.Join(torrent.Labels, " "))
	
	// 只匹配指定关键字
	var matchContent []string
	if matches, ok := rule["match"].([]string); ok {
		for _, match := range matches {
			switch match {
			case "title":
				matchContent = append(matchContent, torrent.Title)
			case "description":
				matchContent = append(matchContent, torrent.Description)
			case "labels":
				matchContent = append(matchContent, torrent.Labels...)
			}
		}
	}
	
	if len(matchContent) > 0 {
		content = strings.Join(matchContent, " ")
	}
	
	// 包含规则�?	var includes []string
	if inc, ok := rule["include"].([]string); ok {
		includes = inc
	} else if incStr, ok := rule["include"].(string); ok {
		includes = []string{incStr}
	}
	
	// 排除规则�?	var excludes []string
	if exc, ok := rule["exclude"].([]string); ok {
		excludes = exc
	} else if excStr, ok := rule["exclude"].(string); ok {
		excludes = []string{excStr}
	}
	
	// 大小范围规则�?	sizeRange := ""
	if sr, ok := rule["size_range"].(string); ok {
		sizeRange = sr
	}
	
	// 做种人数规则�?	var seeders int
	if s, ok := rule["seeders"].(int); ok {
		seeders = s
	} else if sStr, ok := rule["seeders"].(string); ok {
		seeders, _ = strconv.Atoi(sStr)
	}
	
	// FREE规则
	var downloadvolumefactor *float64
	if dvf, ok := rule["downloadvolumefactor"].(int); ok {
		dvfFloat := float64(dvf)
		downloadvolumefactor = &dvfFloat
	} else if dvf, ok := rule["downloadvolumefactor"].(float64); ok {
		downloadvolumefactor = &dvf
	}
	
	// 发布时间规则
	pubdate := ""
	if pd, ok := rule["publish_time"].(string); ok {
		pubdate = pd
	}
	
	if len(includes) > 0 {
		matchFound := false
		for _, include := range includes {
			matched, _ := regexp.MatchString(include, content)
			if matched {
				matchFound = true
				break
			}
		}
		if !matchFound {
			// 未发现任何包含项
			logger.Debugf("种子 %s - %s 不包含任何项 %v", torrent.SiteName, torrent.Title, includes)
			return false
		}
	}
	
	for _, exclude := range excludes {
		matched, _ := regexp.MatchString(exclude, content)
		if matched {
			// 发现排除�?			logger.Debugf("种子 %s - %s 包含 %s", torrent.SiteName, torrent.Title, exclude)
			return false
		}
	}
	
	if sizeRange != "" {
		if !f.matchSize(torrent, sizeRange) {
			// 大小范围不匹�?			stringUtils := utils.NewStringUtils()
			logger.Debugf("种子 %s - %s 大小 %s 不在范围 %sMB",
				torrent.SiteName, torrent.Title,
				stringUtils.StrFilesize(int64(torrent.Size)), sizeRange)
			return false
		}
	}
	
	if seeders > 0 {
		if torrent.Seeders < seeders {
			// 做种人数不匹�?			logger.Debugf("种子 %s - %s 做种人数 %d 小于 %d",
				torrent.SiteName, torrent.Title, torrent.Seeders, seeders)
			return false
		}
	}
	
	if downloadvolumefactor != nil {
		if torrent.DownloadVolumeFactor != *downloadvolumefactor {
			// FREE规则不匹�?			logger.Debugf("种子 %s - %s FREE�?%.2f 不是 %.2f",
				torrent.SiteName, torrent.Title, torrent.DownloadVolumeFactor, *downloadvolumefactor)
			return false
		}
	}
	
	if pubdate != "" {
		// 种子发布时间
		pubMinutes := torrent.PubMinutes()
		// 发布时间规则
		pubTimes := strings.Split(pubdate, "-")
		if len(pubTimes) == 1 {
			// 发布时间小于规则
			pubTime, _ := strconv.ParseFloat(pubTimes[0], 64)
			if float64(pubMinutes) < pubTime {
				logger.Debugf("种子 %s - %s 发布时间 %d 小于 %.2f",
					torrent.SiteName, torrent.Title, pubMinutes, pubTime)
				return false
			}
		} else if len(pubTimes) == 2 {
			// 区间
			minTime, _ := strconv.ParseFloat(pubTimes[0], 64)
			maxTime, _ := strconv.ParseFloat(pubTimes[1], 64)
			if !(minTime <= float64(pubMinutes) && float64(pubMinutes) <= maxTime) {
				logger.Debugf("种子 %s - %s 发布时间 %d 不在 %.2f-%.2f 时间区间",
					torrent.SiteName, torrent.Title, pubMinutes, minTime, maxTime)
				return false
			}
		}
	}
	
	return true
}

// matchTMDB 判断种子是否匹配TMDB规则
func (f *FilterModule) matchTMDB(tmdb map[string]string) bool {
	// 判断种子是否匹配TMDB规则
	if f.media == nil {
		return false
	}
	
	for attr, value := range tmdb {
		if value == "" {
			continue
		}
		
		// 获取media信息的�?		infoValue := f.getMediaValue(attr)
		if infoValue == "" {
			// 没有该值，不匹�?			return false
		}
		
		var infoValues []string
		if attr == "production_countries" {
			// 国家信息
			// 注意：这里简化处理，实际应根据具体结构处�?			infoValues = []string{strings.ToUpper(infoValue)}
		} else {
			// media信息转化为数�?			infoValues = []string{strings.ToUpper(infoValue)}
		}
		
		// 过滤值转化为数组
		var values []string
		if strings.Contains(value, ",") {
			for _, val := range strings.Split(value, ",") {
				val = strings.TrimSpace(val)
				if val != "" {
					values = append(values, strings.ToUpper(val))
				}
			}
		} else {
			values = []string{strings.ToUpper(value)}
		}
		
		// 检查是否有交集
		if !f.hasIntersection(values, infoValues) {
			return false
		}
	}
	
	return true
}

// getMediaValue 获取媒体信息�?func (f *FilterModule) getMediaValue(key string) string {
	if f.media == nil {
		return ""
	}
	
	switch key {
	case "original_language":
		return f.media.OriginalLanguage
	case "production_countries":
		// 简化处�?		if len(f.media.ProductionCountries) > 0 {
			return f.media.ProductionCountries[0]
		}
		return ""
	default:
		return ""
	}
}

// hasIntersection 检查两个字符串切片是否有交�?func (f *FilterModule) hasIntersection(a, b []string) bool {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}
	
	for _, item := range b {
		if set[item] {
			return true
		}
	}
	return false
}

// matchSize 判断种子是否匹配大小范围
func (f *FilterModule) matchSize(torrent *context.TorrentInfo, sizeRange string) bool {
	// 判断种子是否匹配大小范围（MB），剧集拆分为每集大�?	if sizeRange == "" {
		return true
	}
	
	// 集数
	meta := metainfo.NewMetaInfo(torrent.Title, &torrent.Description)
	episodeCount := meta.TotalEpisode
	if episodeCount == 0 {
		episodeCount = 1
	}
	
	// 每集大小
	torrentSize := float64(torrent.Size) / float64(episodeCount)
	
	// 大小范围
	sizeRange = strings.TrimSpace(sizeRange)
	if strings.Contains(sizeRange, "-") {
		// 区间
		parts := strings.Split(sizeRange, "-")
		if len(parts) == 2 {
			sizeMin, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			sizeMax, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			sizeMin *= 1024 * 1024
			sizeMax *= 1024 * 1024
			return sizeMin <= torrentSize && torrentSize <= sizeMax
		}
	} else if strings.HasPrefix(sizeRange, ">") {
		// 大于
		sizeMin, _ := strconv.ParseFloat(strings.TrimSpace(sizeRange[1:]), 64)
		sizeMin *= 1024 * 1024
		return torrentSize >= sizeMin
	} else if strings.HasPrefix(sizeRange, "<") {
		// 小于
		sizeMax, _ := strconv.ParseFloat(strings.TrimSpace(sizeRange[1:]), 64)
		sizeMax *= 1024 * 1024
		return torrentSize <= sizeMax
	}
	
	return false
}

// GetInstance 获取过滤器模块实�?func GetInstance() modules.Module {
	return NewFilterModule()
}
