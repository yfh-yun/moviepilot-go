package media

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// ReleaseGroupConfigProvider 制作组配置提供者接口
type ReleaseGroupConfigProvider interface {
	GetCustomReleaseGroups() ([]string, error)
}

// ReleaseGroupsMatcher 制作组匹配器
type ReleaseGroupsMatcher struct {
	releaseGroups  string
	customProvider ReleaseGroupConfigProvider
	cache          interface{} // 缓存后端，实际类型为cache.CacheBackend
	once           sync.Once
}

// 内置组映射
var releaseGroupMap = map[string][]string{
	"0ff":          {`FF(?:(?:A|WE)B|CD|E(?:DU|B)|TV)`},
	"1pt":          {},
	"52pt":         {},
	"audiences":    {`Audies`, `AD(?:Audio|E(?:book|)|Music|Web)`},
	"azusa":        {},
	"beitai":       {`BeiTai`},
	"btschool":     {`Bts(?:CHOOL|HD|PAD|TV)`, `Zone`},
	"carpt":        {`CarPT`},
	"chdbits":      {`CHD(?:Bits|PAD|(?:|HK)TV|WEB|)`, `StBOX`, `OneHD`, `Lee`, `xiaopie`},
	"discfan":      {},
	"dragonhd":     {},
	"eastgame":     {`(?:(?:iNT|(?:HALFC|Mini(?:S|H|FH)D))-|)TLF`},
	"filelist":     {},
	"gainbound":    {`(?:DG|GBWE)B`},
	"hares":        {`Hares(?:(?:M|T)V|Web|)`},
	"hd4fans":      {},
	"hdarea":       {`HDA(?:pad|rea|TV)`, `EPiC`},
	"hdatmos":      {},
	"hdbd":         {},
	"hdchina":      {`HDC(?:hina|TV|)`, `k9611`, `tudou`, `iHD`},
	"hddolby":      {`D(?:ream|BTV)`, `(?:HD|QHstudI)o`},
	"hdfans":       {`beAst(?:TV|)`},
	"hdhome":       {`HDH(?:ome|Pad|TV|WEB|)`},
	"hdpt":         {`HDPT(?:Web|)`},
	"hdsky":        {`HDS(?:ky|TV|Pad|WEB|)`, `AQLJ`},
	"hdtime":       {},
	"HDU":          {},
	"hdvideo":      {},
	"hdzone":       {`HDZ(?:one|)`},
	"hhanclub":     {`HHWEB`},
	"hitpt":        {},
	"htpt":         {`HTPT`},
	"iptorrents":   {},
	"joyhd":        {},
	"keepfrds":     {`FRDS`, `Yumi`, `cXcY`},
	"lemonhd":      {`L(?:eague(?:(?:C|H)D|(?:M|T)V|NF|WEB)|HD)`, `i18n`, `CiNT`},
	"mteam":        {`MTeam(?:TV|)`, `MPAD`, `MWeb`},
	"nanyangpt":    {},
	"nicept":       {},
	"oshen":        {},
	"ourbits":      {`Our(?:Bits|TV)`, `FLTTH`, `Ao`, `PbK`, `MGs`, `iLove(?:HD|TV)`},
	"piggo":        {`PiGo(?:NF|(?:H|WE)B)`},
	"ptchina":      {},
	"pterclub":     {`PTer(?:DIY|Game|(?:M|T)V|WEB|)`},
	"pthome":       {`PTH(?:Audio|eBook|music|ome|tv|WEB|)`},
	"ptmsg":        {},
	"ptsbao":       {`PTsbao`, `OPS`, `F(?:Fans(?:AIeNcE|BD|D(?:VD|IY)|TV|WEB)|HDMv)`, `SGXT`},
	"pttime":       {},
	"putao":        {`PuTao`},
	"soulvoice":    {},
	"springsunday": {`CMCT(?:V|)`},
	"sharkpt":      {`Shark(?:WEB|DIY|TV|MV|)`},
	"tccf":         {},
	"tjupt":        {`TJUPT`},
	"totheglory":   {`TTG`, `WiKi`, `NGB`, `DoA`, `(?:ARi|ExRE)N`},
	"U2":           {},
	"ultrahd":      {},
	"others": {`B(?:MDru|eyondHD|TN)`, `C(?:fandora|trlhd|MRG)`, `DON`, `EVO`, `FLUX`, `HONE(?:yG|)`,
		`N(?:oGroup|T(?:b|G))`, `PandaMoon`, `SMURF`, `T(?:EPES|aengoo|rollHD )`},
	"anime": {`ANi`, `HYSUB`, `KTXP`, `LoliHouse`, `MCE`, `Nekomoe kissaten`, `SweetSub`, `MingY`,
		`(?:Lilith|NC)-Raws`, `织梦字幕组`, `枫叶字幕组`, `猎户手抄部`, `喵萌奶茶屋`, `漫猫字幕社`,
		`霜庭云花Sub`, `北宇治字幕组`, `氢气烤肉架`, `云歌字幕组`, `萌樱字幕组`, `极影字幕社`,
		`悠哈璃羽字幕社`,
		`❀拨雪寻春❀`, `沸羊羊(?:制作|字幕组)`, `(?:桜|樱)都字幕组`},
	"forge": {`FROG(?:E|Web|)`},
	"ubits": {`UB(?:its|WEB|TV)`},
}

// NewReleaseGroupsMatcher 创建新的ReleaseGroupsMatcher实例
// 支持两种调用方式：
// 1. NewReleaseGroupsMatcher(provider, cache) - 完整参数
// 2. NewReleaseGroupsMatcher(provider) - 只有配置提供者，无缓存
func NewReleaseGroupsMatcher(params ...interface{}) *ReleaseGroupsMatcher {
	var provider ReleaseGroupConfigProvider
	var cache interface{}

	// 解析参数
	switch len(params) {
	case 2:
		if p, ok := params[0].(ReleaseGroupConfigProvider); ok {
			provider = p
		}
		cache = params[1]
	case 1:
		if p, ok := params[0].(ReleaseGroupConfigProvider); ok {
			provider = p
		}
	}

	matcher := &ReleaseGroupsMatcher{
		customProvider: provider,
		cache:          cache,
	}
	matcher.init()
	return matcher
}

// init 初始化ReleaseGroupsMatcher
func (rm *ReleaseGroupsMatcher) init() {
	rm.once.Do(func() {
		var releaseGroups []string
		for _, siteGroups := range releaseGroupMap {
			for _, group := range siteGroups {
				releaseGroups = append(releaseGroups, group)
			}
		}
		rm.releaseGroups = strings.Join(releaseGroups, "|")
	})
}

// Match 匹配制作组/字幕组
// title: 资源标题或文件名
// groups: 制作组/字幕组正则表达式
// return: 匹配结果，多个组用@分隔
func (rm *ReleaseGroupsMatcher) Match(title, groups string) string {
	if title == "" {
		return ""
	}

	// 生成缓存键
	cacheKey := "releasegroup:" + title + ":" + groups
	region := "releasegroup"

	// 尝试从缓存获取结果
	if rm.cache != nil {
		// 使用类型断言调用缓存后端的Get方法
		type cacheBackend interface {
			Get(key string, region string) (interface{}, bool, error)
		}
		if cb, ok := rm.cache.(cacheBackend); ok {
			cachedValue, hit, err := cb.Get(cacheKey, region)
			if err == nil && hit {
				if result, ok := cachedValue.(string); ok {
					return result
				}
			}
		}
	}

	// 如果没有提供groups，则使用内置组和自定义组
	if groups == "" {
		groups = rm.releaseGroups
		// 从配置获取自定义组
		if rm.customProvider != nil {
			customGroups, err := rm.customProvider.GetCustomReleaseGroups()
			if err == nil && len(customGroups) > 0 {
				customGroupsStr := strings.Join(customGroups, "|")
				groups = groups + "|" + customGroupsStr
			}
		}
	}

	// 构建正则表达式 - 移除后顾断言，使用Go支持的语法
	title = title + " "
	// 简化的正则表达式，匹配常见的制作组/字幕组格式
	// 匹配以 [-@[￡【&] 开头，以 @.][】& 结尾的组名
	// 使用非捕获组和字符类
	pattern := `[-@\[￡【&]((?:` + groups + `))[@.\s\S\]\[】&]`
	re := regexp.MustCompile(`(?i)` + pattern) // (?i) 表示忽略大小写

	// 查找所有匹配项
	matches := re.FindAllStringSubmatch(title, -1)
	uniqueGroups := make(map[string]bool)
	var result []string

	for _, match := range matches {
		if len(match) > 1 {
			group := match[1] // 提取第一个捕获组，即实际的组名
			if group != "" && !uniqueGroups[group] {
				uniqueGroups[group] = true
				result = append(result, group)
			}
		}
	}

	// 生成结果
	resultStr := strings.Join(result, "@")

	// 缓存结果
	if rm.cache != nil {
		// 使用类型断言调用缓存后端的Set方法
		type cacheBackend interface {
			Set(key string, value interface{}, ttl time.Duration, region string, opts ...interface{}) error
		}
		if cb, ok := rm.cache.(cacheBackend); ok {
			cb.Set(cacheKey, resultStr, 24*time.Hour, region) // 缓存24小时
		}
	}

	return resultStr
}
